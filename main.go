package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// Args defines the command line arguments for the transcription tool
type Args struct {
	File        string  `arg:"positional,required" help:"Path to the audio file to transcribe"`
	Model       string  `arg:"--model" default:"gpt-4o-transcribe" help:"OpenAI model to use for transcription"`
	Language    string  `arg:"--language" help:"Language of the audio file (optional)"`
	Prompt      string  `arg:"--prompt" help:"Optional text to guide the model's style or continue a previous audio segment"`
	Format      string  `arg:"--format" default:"text" help:"Output format: text, srt, verbose_json, or vtt"`
	OutputDir   string  `arg:"--output-dir,-o" help:"Directory to save the transcription output (defaults to current directory)"`
	OutputExt   string  `arg:"--output-ext" help:"Extension for the output file (defaults to .txt for text, or appropriate extension for other formats)"`
	APIKey      string  `arg:"--api-key" env:"OPENAI_API_KEY" help:"OpenAI API key (can also be set via OPENAI_API_KEY environment variable)"`
	Temperature float64 `arg:"--temperature" default:"0" help:"Sampling temperature between 0 and 1 (higher is more random)"`
}

func printHeader() {
	fmt.Println("  Pindar - Audio Transcription CLI")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printParameters(args Args, audioFile string) {
	fmt.Println("\n  Transcription Parameters:")
	fmt.Printf("   File:        %s\n", audioFile)
	fmt.Printf("   Model:       %s\n", args.Model)
	if args.Language != "" {
		fmt.Printf("   Language:    %s\n", args.Language)
	} else {
		fmt.Printf("   Language:    auto-detect\n")
	}
	fmt.Printf("   Format:      %s\n", args.Format)
	if args.Temperature != 0 {
		fmt.Printf("   Temperature: %.1f\n", args.Temperature)
	}
	if args.Prompt != "" {
		fmt.Printf("   Prompt:      %s\n", args.Prompt)
	}
	fmt.Println()
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		return strings.ToLower(ext[1:]) // Remove the dot and convert to lowercase
	}
	return ""
}

func isFormatSupported(ext string) bool {
	supportedFormats := map[string]bool{
		"flac": true, "mp3": true, "mp4": true, "mpeg": true, "mpga": true,
		"m4a": true, "ogg": true, "wav": true, "webm": true,
	}
	return supportedFormats[strings.ToLower(ext)]
}

func convertToMP4(inputPath string) (string, error) {
	// Create a temporary directory for the converted file
	tmpDir, err := os.MkdirTemp("", "pindar_convert")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Generate output file path
	baseName := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputPath := filepath.Join(tmpDir, nameWithoutExt+"_converted.mp4")

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg is required for audio format conversion but was not found in PATH. Please install ffmpeg")
	}

	// Run ffmpeg conversion with hidden output
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:a", "aac", "-b:a", "128k", "-y", outputPath)

	// Capture output to hide it
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, stderr.String())
	}

	return outputPath, nil
}

func main() {
	var args Args
	arg.MustParse(&args)

	printHeader()

	// Get API key using priority order: CLI arg → env var → config file → prompt user
	apiKey, err := getAPIKey(args.APIKey)
	if err != nil {
		fmt.Printf(" Error getting API key: %v\n", err)
		os.Exit(1)
	}

	// Create OpenAI client
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	// Check if format is supported, convert if necessary
	originalFile := args.File
	ext := getFileExtension(args.File)
	if !isFormatSupported(ext) {
		fmt.Printf(" Converting .%s to .mp4 format...\n", ext)
		convertedFile, err := convertToMP4(args.File)
		if err != nil {
			fmt.Printf(" Error converting audio file: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(convertedFile) // Clean up converted file
		args.File = convertedFile
	}

	// Print transcription parameters
	printParameters(args, originalFile)

	// Validate the audio file
	file, err := os.Open(args.File)
	if err != nil {
		fmt.Printf(" Error opening audio file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Start transcription
	fmt.Println(" Starting transcription...")

	// Reset file pointer to beginning
	file.Seek(0, 0)

	// Create the transcription params with required parameters
	params := openai.AudioTranscriptionNewParams{
		File:  file,
		Model: openai.AudioModel(args.Model),
	}

	if args.Language != "" {
		params.Language = param.NewOpt(args.Language)
	}

	if args.Prompt != "" {
		params.Prompt = param.NewOpt(args.Prompt)
	}

	// Set response format - always use JSON to avoid plain text parsing issues
	// We'll handle the user's desired format in post-processing
	params.ResponseFormat = openai.AudioResponseFormatJSON

	if args.Temperature != 0 {
		params.Temperature = param.NewOpt(args.Temperature)
	}

	// Create a context for the request
	ctx := context.Background()

	// Send the transcription request
	response, err := client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		// Handle specific error cases gracefully
		errStr := err.Error()

		if strings.Contains(errStr, "longer than 1500 seconds") || strings.Contains(errStr, "maximum for this model") {
			fmt.Printf("❌ Audio file too long: The audio duration exceeds the 25-minute limit for this model.\n")
			fmt.Printf("💡 Suggestions:\n")
			fmt.Printf("   • Split the audio into shorter segments (< 25 minutes each)\n")
			fmt.Printf("   • Use audio editing software to create multiple files\n")
			fmt.Printf("   • Consider using a different transcription service for longer files\n")
		} else if strings.Contains(errStr, "invalid_api_key") || strings.Contains(errStr, "Incorrect API key") {
			fmt.Printf("❌ API Key Error: Invalid or missing OpenAI API key.\n")
			fmt.Printf("💡 Please check your API key and try again.\n")
		} else if strings.Contains(errStr, "quota") || strings.Contains(errStr, "rate_limit") {
			fmt.Printf("❌ Rate Limit/Quota Error: API usage limit reached.\n")
			fmt.Printf("💡 Please wait a moment and try again, or check your OpenAI account billing.\n")
		} else {
			fmt.Printf("❌ Error calling OpenAI API: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Transcription completed successfully!")

	// Handle response - we always get JSON format from API to avoid parsing issues
	var transcriptionText string

	switch args.Format {
	case "text", "":
		// User wants plain text - just use the text field
		transcriptionText = response.Text
	case "verbose_json":
		// User wants verbose JSON - we need to note that we're using standard JSON
		// since we forced JSON format, this is what we get
		transcriptionText = response.Text
	case "srt", "vtt":
		// For SRT and VTT, we only get plain text from the API
		// The user would need to use a different service for timestamp formatting
		// For now, return the text with a note
		transcriptionText = response.Text
		fmt.Printf("⚠️  Note: SRT/VTT formats require timestamps. Using text output instead.\n")
	default:
		transcriptionText = response.Text
	}

	// Determine output file path
	outputFile := ""
	if args.OutputDir != "" || args.OutputExt != "" {
		outputFile = determineOutputFileName(args, originalFile)
	}

	// Print response to stdout or save to file
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(transcriptionText), 0644)
		if err != nil {
			fmt.Printf("❌ Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("💾 Transcription saved to: %s\n", outputFile)
	} else {
		// Output to stdout with nice formatting
		fmt.Println("\n📝 Transcription:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s\n", transcriptionText)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}
}

func determineOutputFileName(args Args, originalFile string) string {
	base := filepath.Base(originalFile)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]

	var outputExt string
	if args.OutputExt != "" {
		outputExt = args.OutputExt
		if !strings.HasPrefix(outputExt, ".") {
			outputExt = "." + outputExt
		}
	} else {
		// Default extensions based on format
		switch args.Format {
		case "srt":
			outputExt = ".srt"
		case "vtt":
			outputExt = ".vtt"
		case "verbose_json":
			outputExt = ".json"
		default:
			outputExt = ".txt"
		}
	}

	return filepath.Join(args.OutputDir, nameWithoutExt+outputExt)
}

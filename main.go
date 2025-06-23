package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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

// supportedFormats lists the audio formats supported by OpenAI API
var supportedFormats = map[string]bool{
	"flac": true,
	"m4a":  true,
	"mp3":  true,
	"mp4":  true,
	"mpeg": true,
	"mpga": true,
	"oga":  true,
	"ogg":  true,
	"wav":  true,
	"webm": true,
}

// getFileExtension returns the file extension without the dot
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		return strings.ToLower(ext[1:]) // Remove the dot and convert to lowercase
	}
	return ""
}

// isFormatSupported checks if the file format is supported by OpenAI API
func isFormatSupported(filename string) bool {
	ext := getFileExtension(filename)
	return supportedFormats[ext]
}

// convertToMP4 converts an audio file to MP4 format using ffmpeg
func convertToMP4(inputPath string) (string, error) {
	// Create a temporary file for the converted audio
	tmpDir := os.TempDir()
	baseName := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputPath := filepath.Join(tmpDir, nameWithoutExt+"_converted.mp4")

	fmt.Printf("Converting %s to MP4 format...\n", inputPath)

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w. Please install ffmpeg to convert unsupported audio formats", err)
	}

	// Run ffmpeg conversion
	// Using libfdk_aac for better quality, fallback to aac if not available
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:a", "aac", "-b:a", "128k", "-y", outputPath)

	// Capture stderr for error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, stderr.String())
	}

	fmt.Printf("Successfully converted to: %s\n", outputPath)
	return outputPath, nil
}

func main() {
	var args Args
	arg.MustParse(&args)

	// Get API key using priority order: CLI arg → env var → config file → prompt user
	apiKey, err := getAPIKey(args.APIKey)
	if err != nil {
		fmt.Printf("Error getting API key: %v\n", err)
		os.Exit(1)
	}

	// Set up OpenAI client
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	// Check if the file format is supported and convert if necessary
	audioFilePath := args.File
	var tempFilePath string

	if !isFormatSupported(args.File) {
		// Convert the file to MP4 format
		convertedPath, err := convertToMP4(args.File)
		if err != nil {
			fmt.Printf("Error converting file format: %v\n", err)
			os.Exit(1)
		}
		audioFilePath = convertedPath
		tempFilePath = convertedPath // Remember to clean up later
	}

	// Clean up temporary file when done
	if tempFilePath != "" {
		defer func() {
			if err := os.Remove(tempFilePath); err != nil {
				fmt.Printf("Warning: Failed to clean up temporary file %s: %v\n", tempFilePath, err)
			}
		}()
	}

	// Validate the audio file
	file, err := os.Open(audioFilePath)
	if err != nil {
		fmt.Printf("Error opening audio file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		os.Exit(1)
	}

	if fileInfo.Size() > 25*1024*1024 {
		fmt.Println("Error: File size exceeds 25MB limit for OpenAI API")
		os.Exit(1)
	}

	// Reset file pointer to beginning
	file.Seek(0, io.SeekStart)

	// Create the transcription params with required parameters
	params := openai.AudioTranscriptionNewParams{
		File:  file,
		Model: args.Model,
	}

	// Create a context for the request
	ctx := context.Background()

	// Send the transcription request
	response, err := client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		fmt.Printf("Error during transcription: %v\n", err)
		os.Exit(1)
	}

	// Determine output filename
	outputFileName := determineOutputFileName(args)

	// Save or print output
	if args.OutputDir != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(args.OutputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		outputPath := filepath.Join(args.OutputDir, outputFileName)
		if err := os.WriteFile(outputPath, []byte(response.Text), 0644); err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Transcription saved to: %s\n", outputPath)
	} else {
		// Print to stdout
		fmt.Println(response.Text)
	}
}

// determineOutputFileName generates an appropriate output filename based on input file and args
func determineOutputFileName(args Args) string {
	base := filepath.Base(args.File)
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

	return nameWithoutExt + outputExt
}

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Args defines the command line arguments for the transcription tool
type Args struct {
	File        string  `arg:"positional,required" help:"Path to the audio file to transcribe"`
	Model       string  `arg:"--model" default:"whisper-1" help:"OpenAI model to use for transcription"`
	Language    string  `arg:"--language" help:"Language of the audio file (optional)"`
	Prompt      string  `arg:"--prompt" help:"Optional text to guide the model's style or continue a previous audio segment"`
	Format      string  `arg:"--format" default:"text" help:"Output format: text, srt, verbose_json, or vtt"`
	OutputDir   string  `arg:"--output-dir,-o" help:"Directory to save the transcription output (defaults to current directory)"`
	OutputExt   string  `arg:"--output-ext" help:"Extension for the output file (defaults to .txt for text, or appropriate extension for other formats)"`
	APIKey      string  `arg:"--api-key" env:"OPENAI_API_KEY" help:"OpenAI API key (can also be set via OPENAI_API_KEY environment variable)"`
	Temperature float64 `arg:"--temperature" default:"0" help:"Sampling temperature between 0 and 1 (higher is more random)"`
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

	// Validate the audio file
	file, err := os.Open(args.File)
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

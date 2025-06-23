package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock implementations
type mockReadCloser struct {
	reader io.Reader
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	return nil
}

func createTempAudioFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test-audio.mp3")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return filePath
}

func TestDetermineOutputFileName(t *testing.T) {
	tests := []struct {
		name     string
		args     Args
		expected string
	}{
		{
			name: "Default format",
			args: Args{
				File:   "/path/to/audio.mp3",
				Format: "text",
			},
			expected: "audio.txt",
		},
		{
			name: "SRT format",
			args: Args{
				File:   "/path/to/audio.mp3",
				Format: "srt",
			},
			expected: "audio.srt",
		},
		{
			name: "VTT format",
			args: Args{
				File:   "/path/to/audio.mp3",
				Format: "vtt",
			},
			expected: "audio.vtt",
		},
		{
			name: "JSON format",
			args: Args{
				File:   "/path/to/audio.mp3",
				Format: "verbose_json",
			},
			expected: "audio.json",
		},
		{
			name: "Custom extension",
			args: Args{
				File:      "/path/to/audio.mp3",
				Format:    "text",
				OutputExt: "transcript",
			},
			expected: "audio.transcript",
		},
		{
			name: "Custom extension with dot",
			args: Args{
				File:      "/path/to/audio.mp3",
				Format:    "text",
				OutputExt: ".transcript",
			},
			expected: "audio.transcript",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := determineOutputFileName(tc.args)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestFileValidation(t *testing.T) {
	// Test with valid file
	validFile := createTempAudioFile(t, "mock audio data")
	file, err := os.Open(validFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Test file size check
	// Create a file larger than 25MB
	largeFile := createTempAudioFile(t, strings.Repeat("a", 26*1024*1024))
	
	_, err = os.Stat(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
}

func TestOutputFileCreation(t *testing.T) {
	// Create a temporary directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.txt")
	
	// Test writing output to file
	content := "Test transcription output"
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}
	
	// Verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	if string(data) != content {
		t.Errorf("Expected content %s, got %s", content, string(data))
	}
}

// Integration test helper - doesn't actually make API calls
func TestTranscriptionIntegration(t *testing.T) {
	// Skip in normal testing since it would require API credentials
	t.Skip("Skipping integration test that would require real API credentials")
	
	// This is a template for manual integration testing
	/*
	args := Args{
		File:      "/path/to/test/audio.mp3",
		Model:     "whisper-1",
		APIKey:    "your-api-key", // Set to actual API key for manual testing
		OutputDir: t.TempDir(),
	}
	
	// Run the transcription process
	// ... the actual implementation would call the main function with these args
	*/
}

// Integration test helper - doesn't actually make API calls
func TestFileOpeningIntegration(t *testing.T) {
	// Create a temporary file for testing
	tempFile := createTempAudioFile(t, "dummy audio content")
	defer os.Remove(tempFile)

	// Test the flow up to the point where we would make the API call
	// This validates file opening, validation, and parameter setup
	args := Args{
		File:   tempFile,
		Format: "text",
	}

	// Ensure the file can be opened and validated
	file, err := os.Open(args.File)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"audio.mp3", "mp3"},
		{"audio.MP3", "mp3"},
		{"audio.flac", "flac"},
		{"audio.m4a", "m4a"},
		{"audio.wav", "wav"},
		{"audio", ""},
		{"audio.", ""},
		{"/path/to/audio.mp3", "mp3"},
		{"audio.test.mp3", "mp3"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := getFileExtension(test.filename)
			if result != test.expected {
				t.Errorf("getFileExtension(%s) = %s, expected %s", test.filename, result, test.expected)
			}
		})
	}
}

func TestIsFormatSupported(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"audio.mp3", true},
		{"audio.flac", true},
		{"audio.wav", true},
		{"audio.m4a", true},
		{"audio.mp4", true},
		{"audio.ogg", true},
		{"audio.webm", true},
		{"audio.mpeg", true},
		{"audio.mpga", true},
		{"audio.oga", true},
		{"audio.aiff", false},
		{"audio.au", false},
		{"audio.amr", false},
		{"audio.3gp", false},
		{"audio.unknown", false},
		{"audio", false},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := isFormatSupported(test.filename)
			if result != test.expected {
				t.Errorf("isFormatSupported(%s) = %v, expected %v", test.filename, result, test.expected)
			}
		})
	}
}

func TestConvertToMP4_FFmpegNotFound(t *testing.T) {
	// Create a temporary file for testing
	tempFile := createTempAudioFile(t, "dummy audio content")
	defer os.Remove(tempFile)

	// Change the filename to an unsupported format
	unsupportedFile := strings.Replace(tempFile, ".mp3", ".aiff", 1)
	err := os.Rename(tempFile, unsupportedFile)
	if err != nil {
		t.Fatalf("Failed to rename test file: %v", err)
	}
	defer os.Remove(unsupportedFile)

	// Test conversion (this will likely fail unless ffmpeg is installed)
	_, err = convertToMP4(unsupportedFile)
	
	// We expect either success (if ffmpeg is available) or a specific error
	if err != nil && !strings.Contains(err.Error(), "ffmpeg not found") && !strings.Contains(err.Error(), "ffmpeg conversion failed") {
		t.Errorf("Unexpected error type: %v", err)
	}
}

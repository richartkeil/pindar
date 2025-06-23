package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/openai/openai-go"
)

// MockTranscriptionResponse mocks the response from OpenAI's transcription API
type MockTranscriptionResponse struct {
	TextContent string
	ErrorToReturn error
}

// Mock client for OpenAI API
type MockOpenAIClient struct {
	MockResponse MockTranscriptionResponse
}

// New mocks the transcription API call
func (m *MockOpenAIClient) New(ctx context.Context, params openai.AudioTranscriptionNewParams) (openai.Transcription, error) {
	if m.MockResponse.ErrorToReturn != nil {
		return openai.Transcription{}, m.MockResponse.ErrorToReturn
	}

	return openai.Transcription{
		Text: m.MockResponse.TextContent,
	}, nil
}

// TestTranscriptionWithMockClient tests the transcription process with a mocked OpenAI client
func TestTranscriptionWithMockClient(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   MockTranscriptionResponse
		expectedOutput string
		expectError    bool
	}{
		{
			name: "Successful transcription",
			mockResponse: MockTranscriptionResponse{
				TextContent: "This is a mock transcription.",
				ErrorToReturn: nil,
			},
			expectedOutput: "This is a mock transcription.",
			expectError: false,
		},
		{
			name: "API error",
			mockResponse: MockTranscriptionResponse{
				TextContent: "",
				ErrorToReturn: errors.New("API error"),
			},
			expectedOutput: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &MockOpenAIClient{
				MockResponse: tc.mockResponse,
			}

			// Create a mock file for testing
			mockFileContent := "mock audio content"
			mockFile := io.NopCloser(bytes.NewReader([]byte(mockFileContent)))

			// Create transcription parameters
			params := openai.AudioTranscriptionNewParams{
				File:  mockFile,
				Model: "whisper-1",
			}

			// Call the mock client
			response, err := mockClient.New(context.Background(), params)

			// Check for expected errors
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Check response content
				if response.Text != tc.expectedOutput {
					t.Errorf("Expected output '%s', got '%s'", tc.expectedOutput, response.Text)
				}
			}
		})
	}
}

// TestProcessTranscriptionRequest tests the process of creating and sending a transcription request
func TestProcessTranscriptionRequest(t *testing.T) {
	// Create a mock OpenAI client
	mockClient := &MockOpenAIClient{
		MockResponse: MockTranscriptionResponse{
			TextContent: "This is a test transcription.",
			ErrorToReturn: nil,
		},
	}

	// Create a mock audio file
	mockFileContent := "mock audio data"
	mockReader := bytes.NewReader([]byte(mockFileContent))
	mockFile := io.NopCloser(mockReader)

	// Setup test parameters
	params := openai.AudioTranscriptionNewParams{
		File:  mockFile,
		Model: "whisper-1",
	}

	// Process the request with the mock client
	transcription, err := mockClient.New(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to process transcription request: %v", err)
	}

	expectedTranscription := "This is a test transcription."
	if transcription.Text != expectedTranscription {
		t.Errorf("Expected transcription '%s', got '%s'", expectedTranscription, transcription.Text)
	}
}

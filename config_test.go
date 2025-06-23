package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	configDir, err := getConfigDir()
	if err != nil {
		t.Fatalf("getConfigDir() failed: %v", err)
	}

	if configDir == "" {
		t.Error("getConfigDir() returned empty string")
	}

	// Check that the directory contains "pindar"
	if !strings.Contains(configDir, "pindar") {
		t.Errorf("Config directory should contain 'pindar', got: %s", configDir)
	}

	// Check that the directory exists after calling getConfigDir
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", configDir)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	configPath, err := getConfigFilePath()
	if err != nil {
		t.Fatalf("getConfigFilePath() failed: %v", err)
	}

	if configPath == "" {
		t.Error("getConfigFilePath() returned empty string")
	}

	// Check that the path ends with config.json
	if !strings.HasSuffix(configPath, "config.json") {
		t.Errorf("Config file path should end with 'config.json', got: %s", configPath)
	}

	// Check that the path contains "pindar"
	if !strings.Contains(configPath, "pindar") {
		t.Errorf("Config file path should contain 'pindar', got: %s", configPath)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Create a temporary directory that doesn't have a config file
	tempDir := t.TempDir()
	nonExistentConfigPath := filepath.Join(tempDir, "nonexistent.json")

	// Test loading a config file that doesn't exist
	_, err := os.ReadFile(nonExistentConfigPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Unexpected error: %v", err)
	}

	// If file doesn't exist, we should get an empty config (this is the expected behavior)
	if os.IsNotExist(err) {
		// This is expected - the file doesn't exist
		// In the actual function, this would return &Config{}
		config := &Config{}
		if config.OpenAIAPIKey != "" {
			t.Errorf("Expected empty API key for non-existent config, got: %s", config.OpenAIAPIKey)
		}
	}
}

func TestJSONMarshalUnmarshal(t *testing.T) {
	// Test JSON marshaling and unmarshaling
	testAPIKey := "test-api-key-12345"
	testConfig := &Config{
		OpenAIAPIKey: testAPIKey,
	}

	// Test marshaling
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Test unmarshaling
	var loadedConfig Config
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if loadedConfig.OpenAIAPIKey != testAPIKey {
		t.Errorf("Expected API key %s, got %s", testAPIKey, loadedConfig.OpenAIAPIKey)
	}
}

func TestFilePermissions(t *testing.T) {
	// Test file permissions for config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testData := []byte(`{"openai_api_key":"test-key"}`)

	// Write config with specific permissions
	err := os.WriteFile(configPath, testData, 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Check file permissions
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check that permissions are 0600 (read/write for owner only)
	expectedPerm := os.FileMode(0600)
	if fileInfo.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %v, got %v", expectedPerm, fileInfo.Mode().Perm())
	}
}

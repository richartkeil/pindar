package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Config represents the application configuration
type Config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
}

// getConfigDir returns the platform-specific configuration directory
func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	
	appConfigDir := filepath.Join(configDir, "pindar")
	
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return appConfigDir, nil
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// loadConfig loads configuration from the config file
func loadConfig() (*Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	
	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return &config, nil
}

// saveConfig saves configuration to the config file
func saveConfig(config *Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// promptForAPIKey prompts the user to enter their OpenAI API key
func promptForAPIKey() (string, error) {
	fmt.Print("OpenAI API key not found. Please enter your OpenAI API key: ")
	
	// Use term.ReadPassword for secure input (doesn't echo to terminal)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		// Fallback to regular input if term.ReadPassword fails
		fmt.Print("\nFalling back to regular input: ")
		reader := bufio.NewReader(os.Stdin)
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read API key: %w", err)
		}
		return strings.TrimSpace(apiKey), nil
	}
	
	fmt.Println() // Add newline after password input
	apiKey := strings.TrimSpace(string(bytePassword))
	
	if apiKey == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}
	
	return apiKey, nil
}

// getAPIKey retrieves the API key using the priority order:
// 1. CLI argument
// 2. Environment variable
// 3. Config file
// 4. Prompt user and save to config
func getAPIKey(cliAPIKey string) (string, error) {
	// 1. CLI argument has highest priority
	if cliAPIKey != "" {
		return cliAPIKey, nil
	}
	
	// 2. Check environment variable
	if envAPIKey := os.Getenv("OPENAI_API_KEY"); envAPIKey != "" {
		return envAPIKey, nil
	}
	
	// 3. Check config file
	config, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	
	if config.OpenAIAPIKey != "" {
		return config.OpenAIAPIKey, nil
	}
	
	// 4. Prompt user and save to config
	fmt.Println("No OpenAI API key found in arguments, environment, or config file.")
	apiKey, err := promptForAPIKey()
	if err != nil {
		return "", err
	}
	
	// Save the API key to config
	config.OpenAIAPIKey = apiKey
	if err := saveConfig(config); err != nil {
		fmt.Printf("Warning: Failed to save API key to config file: %v\n", err)
		fmt.Println("You may need to provide the API key again next time.")
	} else {
		configPath, _ := getConfigFilePath()
		fmt.Printf("API key saved to: %s\n", configPath)
	}
	
	return apiKey, nil
}

// Package config handles loading and parsing the nlch configuration file.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderConfig holds configuration for a single provider.
type ProviderConfig struct {
	Key          string `yaml:"key,omitempty"`
	DefaultModel string `yaml:"default_model,omitempty"`
	URL          string `yaml:"url,omitempty"`
}

// Config holds the overall nlch configuration.
type Config struct {
	DefaultProvider string                    `yaml:"default_provider"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
}

// GetProviders returns the providers configuration
func (c *Config) GetProviders() map[string]ProviderConfig {
	return c.Providers
}

// Load reads the config from ~/.config/nlch/config.yaml,
// or falls back to ./nlch.yaml in the current directory if not found or broken.
// Returns a combined error only if both fail.
func Load() (*Config, error) {
	var pathsTried []string
	var errorsTried []string

	// Try user config first
	home, err := os.UserHomeDir()
	if err == nil {
		userPath := filepath.Join(home, ".config", "nlch", "config.yaml")
		pathsTried = append(pathsTried, userPath)
		data, err := os.ReadFile(userPath)
		if err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return &cfg, nil
			} else {
				errorsTried = append(errorsTried, fmt.Sprintf("%s: %v", userPath, err))
			}
		} else {
			errorsTried = append(errorsTried, fmt.Sprintf("%s: %v", userPath, err))
		}
	}

	// Fallback to project config
	projPath := "nlch.yaml"
	pathsTried = append(pathsTried, projPath)
	data, err := os.ReadFile(projPath)
	if err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return &cfg, nil
		} else {
			errorsTried = append(errorsTried, fmt.Sprintf("%s: %v", projPath, err))
		}
	} else {
		errorsTried = append(errorsTried, fmt.Sprintf("%s: %v", projPath, err))
	}

	return nil, fmt.Errorf("could not load config from any of: %v\nErrors:\n%s", pathsTried, strings.Join(errorsTried, "\n"))
}

// LoadOrCreate attempts to load the config, and if no user config exists, runs the initial setup
func LoadOrCreate() (*Config, error) {
	// Check if user config file exists first
	userPath, err := GetUserConfigPath()
	if err != nil {
		return nil, err
	}

	// If user config file doesn't exist, run first-time setup
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		fmt.Println("No configuration found. Let's set up nlch for the first time!")
		return CreateInitialConfig()
	}

	// If user config exists, try to load from any available location
	cfg, err := Load()
	if err == nil {
		return cfg, nil
	}

	// If file exists but couldn't be loaded, return the original error
	return nil, fmt.Errorf("configuration file exists but could not be loaded: %v", err)
}

// GetUserConfigPath returns the user's configuration file path
func GetUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nlch", "config.yaml"), nil
}

// CreateInitialConfig prompts the user for provider information and creates a config file
func CreateInitialConfig() (*Config, error) {
	fmt.Println("\n🎉 Welcome to nlch! Let's set up your configuration.")
	fmt.Println("nlch supports multiple AI providers. Let's configure one to get started.")
	fmt.Println()

	// Available providers
	providers := map[string]ProviderInfo{
		"1": {"openrouter", "OpenRouter (supports many models)", "https://openrouter.ai", "sk-or-v1-"},
		"2": {"anthropic", "Anthropic Claude", "https://console.anthropic.com", "sk-ant-"},
		"3": {"openai", "OpenAI GPT", "https://platform.openai.com", "sk-"},
		"4": {"gemini", "Google Gemini", "https://aistudio.google.com", ""},
		"5": {"ollama", "Ollama (local)", "https://ollama.ai", ""},
	}

	// Display options
	fmt.Println("Available providers:")
	for key, provider := range providers {
		fmt.Printf("  %s. %s\n", key, provider.Name)
	}

	// Get user choice
	reader := bufio.NewReader(os.Stdin)
	var selectedProvider ProviderInfo
	for {
		fmt.Print("\nSelect a provider (1-5): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if provider, ok := providers[choice]; ok {
			selectedProvider = provider
			break
		}
		fmt.Println("Invalid choice. Please select 1-5.")
	}

	fmt.Printf("\nYou selected: %s\n", selectedProvider.Name)

	// Get API key
	var apiKey string
	if selectedProvider.Key != "ollama" {
		fmt.Printf("You'll need an API key from: %s\n", selectedProvider.Website)
		if selectedProvider.KeyPrefix != "" {
			fmt.Printf("API keys typically start with: %s\n", selectedProvider.KeyPrefix)
		}

		for {
			fmt.Print("Enter your API key: ")
			apiKey, _ = reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKey)

			if apiKey != "" {
				break
			}
			fmt.Println("API key cannot be empty.")
		}
	}

	// Get URL for Ollama
	var url string
	if selectedProvider.Key == "ollama" {
		fmt.Print("Enter Ollama URL (press Enter for http://localhost:11434): ")
		url, _ = reader.ReadString('\n')
		url = strings.TrimSpace(url)
		if url == "" {
			url = "http://localhost:11434"
		}
	}

	// Get default model
	fmt.Print("Enter default model (press Enter for provider default): ")
	defaultModel, _ := reader.ReadString('\n')
	defaultModel = strings.TrimSpace(defaultModel)

	// Set provider-specific default models
	if defaultModel == "" {
		switch selectedProvider.Key {
		case "openrouter":
			defaultModel = "openai/gpt-4o-mini"
		case "anthropic":
			defaultModel = "claude-3-5-sonnet-20241022"
		case "openai":
			defaultModel = "gpt-4o-mini"
		case "gemini":
			defaultModel = "gemini-1.5-flash"
		case "ollama":
			defaultModel = "llama3.2"
		}
	}

	// Create config
	config := &Config{
		DefaultProvider: selectedProvider.Key,
		Providers: map[string]ProviderConfig{
			selectedProvider.Key: {
				Key:          apiKey,
				DefaultModel: defaultModel,
				URL:          url,
			},
		},
	}

	// Save config
	err := SaveConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Printf("\n✅ Configuration saved successfully!\n")
	fmt.Printf("Provider: %s\n", selectedProvider.Name)
	fmt.Printf("Default model: %s\n", defaultModel)
	fmt.Println("\nYou can now use nlch. Try: nlch \"list files in this directory\"")

	return config, nil
}

// ProviderInfo holds information about available providers
type ProviderInfo struct {
	Key       string
	Name      string
	Website   string
	KeyPrefix string
}

// SaveConfig saves the configuration to the user's config file
func SaveConfig(config *Config) error {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}

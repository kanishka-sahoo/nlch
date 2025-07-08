// Package config handles loading and parsing the nlch configuration file.
package config

import (
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

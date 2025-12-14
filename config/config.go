package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the complete application configuration
type Config struct {
	Watch  []string   `yaml:"watch"`
	Ignore []string   `yaml:"ignore"`
	Logs   LogsConfig `yaml:"logs"`
	LLM    LLMConfig  `yaml:"llm"`
}

// LogsConfig holds log scraping configuration
type LogsConfig struct {
	File  string `yaml:"file"`
	Lines int    `yaml:"lines"`
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider     string `yaml:"provider"`
	Model        string `yaml:"model"`
	APIKey       string `yaml:"api_key"`
	SystemPrompt string `yaml:"system_prompt"`
}

// Load loads configuration from .glimpse.yaml in the current directory
func Load() (*Config, error) {
	// Default configuration
	config := &Config{
		Watch: []string{
			"./*.go",
			"./internal/**/*.go",
			"./pkg/**/*.go",
		},
		Ignore: []string{"*_test.go"},
		Logs: LogsConfig{
			File:  "./tmp/server.log",
			Lines: 50,
		},
		LLM: LLMConfig{
			Provider:     "openai",
			Model:        "gpt-4o",
			SystemPrompt: "You are a Principal Go Engineer. Review strictly for bugs, perf, and slog context.",
		},
	}

	// Try to load from file
	configPath := filepath.Join(".", ".glimpse.yaml")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Get API key from environment if not in config
	if config.LLM.APIKey == "" {
		switch config.LLM.Provider {
		case "openai":
			config.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
		case "gemini":
			config.LLM.APIKey = os.Getenv("GEMINI_API_KEY")
		case "zai":
			config.LLM.APIKey = os.Getenv("ZAI_API_KEY")
		}
	}

	return config, nil
}

// GetDebounceDuration returns debounce duration for file changes
func (c *Config) GetDebounceDuration() time.Duration {
	return 2 * time.Second // Increased to prevent multiple LLM calls
}
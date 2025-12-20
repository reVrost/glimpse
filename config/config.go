package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/revrost/glimpse/ui"
	"github.com/revrost/glimpse/styles"
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

// getGlobalConfigPath returns the path to the global config file following XDG convention
func getGlobalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	// Check for XDG_CONFIG_HOME first
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, ".glimpse.yaml")
	}
	
	// Fall back to ~/.config
	return filepath.Join(home, ".config", ".glimpse.yaml")
}

// ensureGlobalConfigDir creates the global config directory if it doesn't exist
func ensureGlobalConfigDir() error {
	path := getGlobalConfigPath()
	if path == "" {
		return fmt.Errorf("could not determine home directory")
	}
	
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create global config directory: %w", err)
		}
	}
	return nil
}

// SaveGlobal saves the config to the global config file
func (c *Config) SaveGlobal() error {
	if err := ensureGlobalConfigDir(); err != nil {
		return err
	}
	
	path := getGlobalConfigPath()
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	
	return nil
}

// Load loads configuration with fallback: local -> global -> defaults
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
			Provider:     "",  // Empty default to trigger prompting
			Model:        "",  // Empty default to trigger prompting
			SystemPrompt: "You are a Principal Go Engineer. Review strictly for bugs, perf, and slog context.",
		},
	}

	// Try to load from local file first
	configPath := filepath.Join(".", ".glimpse.yaml")
	loadedFromLocal := false
	
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read local config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse local config file: %w", err)
		}
		loadedFromLocal = true
	}
	
	// If not found locally, try global config
	if !loadedFromLocal {
		globalPath := getGlobalConfigPath()
		if globalPath != "" {
			if _, err := os.Stat(globalPath); err == nil {
				data, err := os.ReadFile(globalPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read global config file: %w", err)
				}

				if err := yaml.Unmarshal(data, config); err != nil {
					return nil, fmt.Errorf("failed to parse global config file: %w", err)
				}
			}
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
		case "claude":
			config.LLM.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	return config, nil
}

// PromptAndSaveProvider prompts the user to select a provider and saves it to global config
func PromptAndSaveProvider() error {
	// Prompt for provider selection
	provider, err := ui.PromptProvider()
	if err != nil {
		return fmt.Errorf("provider selection failed: %w", err)
	}
	
	// Prompt for model selection
	model, err := ui.PromptModel(provider)
	if err != nil {
		return fmt.Errorf("model selection failed: %w", err)
	}
	
	// Show API key help
	ui.ShowAPIKeyHelp(provider)
	
	// Create config with selected provider and model
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
			Provider:     provider,
			Model:        model,
			SystemPrompt: "You are a Principal Go Engineer. Review strictly for bugs, perf, and slog context.",
		},
	}
	
	// Save to global config
	if err := config.SaveGlobal(); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}
	
	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Saved %s:%s to global config", provider, model)))
	return nil
}

// GetDebounceDuration returns debounce duration for file changes
func (c *Config) GetDebounceDuration() time.Duration {
	return 2 * time.Second // Increased to prevent multiple LLM calls
}
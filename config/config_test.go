package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Remove any existing config file
	os.Remove(".glimpse.yaml")
	
	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	
	// Check default values
	assert.Equal(t, []string{"./internal/**/*.go", "./pkg/**/*.go"}, config.Watch)
	assert.Equal(t, []string{"*_test.go"}, config.Ignore)
	assert.Equal(t, "./tmp/server.log", config.Logs.File)
	assert.Equal(t, 50, config.Logs.Lines)
	assert.Equal(t, "openai", config.LLM.Provider)
	assert.Equal(t, "gpt-4o", config.LLM.Model)
}

func TestLoadCustomConfig(t *testing.T) {
	// Create a custom config file
	configContent := `
watch:
  - "./src/**/*.go"
  - "./lib/**/*.go"
ignore:
  - "*_generated.go"
logs:
  file: "./debug.log"
  lines: 100
llm:
  provider: "gemini"
  model: "gemini-pro"
  system_prompt: "Custom prompt"
`
	err := os.WriteFile(".glimpse.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(".glimpse.yaml")
	
	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	
	// Check custom values
	assert.Equal(t, []string{"./src/**/*.go", "./lib/**/*.go"}, config.Watch)
	assert.Equal(t, []string{"*_generated.go"}, config.Ignore)
	assert.Equal(t, "./debug.log", config.Logs.File)
	assert.Equal(t, 100, config.Logs.Lines)
	assert.Equal(t, "gemini", config.LLM.Provider)
	assert.Equal(t, "gemini-pro", config.LLM.Model)
	assert.Equal(t, "Custom prompt", config.LLM.SystemPrompt)
}

func TestGetDebounceDuration(t *testing.T) {
	config := &Config{}
	duration := config.GetDebounceDuration()
	assert.Equal(t, "500ms", duration.String())
}
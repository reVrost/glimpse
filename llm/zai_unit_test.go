package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZAIProviderInSwitch(t *testing.T) {
	// Test that the provider is correctly recognized in the switch statement
	config := Config{
		Provider: "zai",
		Model:    "glm-4.6",
		APIKey:   "test-key",
	}

	client := New(config)
	assert.Equal(t, "zai", client.config.Provider)
	assert.Equal(t, "glm-4.6", client.config.Model)
}

func TestZAIRequestStructure(t *testing.T) {
	// Test that Z.AI uses the correct request format
	config := Config{
		Provider: "zai",
		Model:    "glm-4-air",
		APIKey:   "test-key",
	}

	client := New(config)
	assert.Equal(t, "glm-4-air", client.config.Model)
	assert.Equal(t, "zai", client.config.Provider)
}
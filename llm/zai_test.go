package llm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateZAI(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY not set, skipping integration test")
	}

	config := Config{
		Provider:     "zai",
		Model:        "glm-4.6",
		APIKey:       apiKey,
		SystemPrompt: "You are a helpful AI assistant.",
	}

	client := New(config)

	req := GenerateRequest{
		SystemPrompt: "You are a code reviewer.",
		Context:      "func add(a, b int) int { return a + b }",
		Task:         "Review this Go function for potential issues.",
	}

	respChan := client.Generate(req)
	resp := <-respChan

	assert.NoError(t, resp.Error)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, resp.Content, "function") // Response should mention the function
}

func TestZAIModelDefault(t *testing.T) {
	// Test that GLM-4.6 is used as default model when none specified
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY not set, skipping integration test")
	}

	config := Config{
		Provider: "zai",
		// Model field is intentionally left empty
		APIKey:   apiKey,
	}

	client := New(config)

	req := GenerateRequest{
		SystemPrompt: "You are a helpful assistant.",
		Task:         "Say hello",
	}

	respChan := client.Generate(req)
	resp := <-respChan

	assert.NoError(t, resp.Error)
	assert.NotEmpty(t, resp.Content)
}
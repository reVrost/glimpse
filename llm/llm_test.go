package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := Config{
		Provider:     "openai",
		Model:        "gpt-4",
		APIKey:       "test-key",
		SystemPrompt: "Test prompt",
	}
	
	client := New(config)
	assert.NotNil(t, client)
	assert.Equal(t, config.Provider, client.config.Provider)
	assert.Equal(t, config.Model, client.config.Model)
	assert.Equal(t, config.APIKey, client.config.APIKey)
	assert.Equal(t, config.SystemPrompt, client.config.SystemPrompt)
}

func TestGenerateOpenAIRequest(t *testing.T) {
	client := New(Config{
		Provider:     "openai",
		Model:        "gpt-4",
		APIKey:       "invalid-key",
		SystemPrompt: "Test prompt",
	})
	
	req := GenerateRequest{
		SystemPrompt: "You are a test assistant",
		Context:      "Test context",
		Task:         "Test task",
	}
	
	// This will fail due to invalid API key, but tests the request format
	respChan := client.Generate(req)
	resp := <-respChan
	
	assert.Error(t, resp.Error)
	assert.Empty(t, resp.Content)
	assert.Contains(t, resp.Error.Error(), "API key") // Check for API key error
}

func TestGenerateUnsupportedProvider(t *testing.T) {
	client := New(Config{
		Provider:     "unsupported",
		Model:        "test-model",
		APIKey:       "test-key",
		SystemPrompt: "Test prompt",
	})
	
	req := GenerateRequest{
		SystemPrompt: "You are a test assistant",
		Context:      "Test context",
		Task:         "Test task",
	}
	
	// This should fail due to unsupported provider
	respChan := client.Generate(req)
	resp := <-respChan
	
	assert.Error(t, resp.Error)
	assert.Empty(t, resp.Content)
	assert.Contains(t, resp.Error.Error(), "unsupported provider")
}

func TestGenerateClaudeRequest(t *testing.T) {
	client := New(Config{
		Provider:     "claude",
		Model:        "claude-3-5-sonnet-20241022",
		APIKey:       "invalid-key",
		SystemPrompt: "Test prompt",
	})
	
	req := GenerateRequest{
		SystemPrompt: "You are a test assistant",
		Context:      "Test context",
		Task:         "Test task",
	}
	
	// This will fail due to invalid API key, but tests the request format
	respChan := client.Generate(req)
	resp := <-respChan
	
	assert.Error(t, resp.Error)
	assert.Empty(t, resp.Content)
	// Check for API key error or authentication error
	errorMsg := resp.Error.Error()
	containsAuthError := assert.Contains(t, errorMsg, "x-api-key") || 
		assert.Contains(t, errorMsg, "API key") || 
		assert.Contains(t, errorMsg, "authentication")
	assert.True(t, containsAuthError, "Expected API key or authentication error, got: %s", errorMsg)
}
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/glimpse/styles"
	"github.com/revrost/glimpse/ui"
)

// Config holds the LLM configuration
type Config struct {
	Provider     string
	Model        string
	APIKey       string
	SystemPrompt string
}

// Client represents an LLM client
type Client struct {
	config Config
	client *http.Client
}

// New creates a new LLM client instance
func New(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
	}
}

// GenerateRequest represents a request to the LLM
type GenerateRequest struct {
	SystemPrompt string
	Context      string
	Task         string
	Stream       bool // Enable streaming output to stdout
}

// GenerateResponse represents the response from the LLM
type GenerateResponse struct {
	Content string
	Error   error
}

// Generate sends a prompt to the LLM and returns the response
func (c *Client) Generate(req GenerateRequest) <-chan GenerateResponse {
	respChan := make(chan GenerateResponse, 1)

	go func() {
		defer close(respChan)

		// Start loading animation if not streaming
		var spinnerChan chan bool
		var loadingText string
		if !req.Stream {
			_, err := ui.NewMarkdownRenderer()
			if err != nil {
				respChan <- GenerateResponse{
					Content: "",
					Error:   fmt.Errorf("failed to initialize markdown renderer: %w", err),
				}
				return
			}

			loadingText = fmt.Sprintf("Analyzing with %s (%s)...", c.config.Provider, c.config.Model)
			spinner := ui.NewSpinner(loadingText)

			spinnerChan = make(chan bool, 1)

			go func() {
				for {
					select {
					case <-spinnerChan:
						return
					case <-time.After(100 * time.Millisecond):
						fmt.Printf("\r%s", styles.Spinner.Render(spinner.Tick()))
					}
				}
			}()
		}

		// Make the API call
		var content string
		var err error
		switch c.config.Provider {
		case "openai":
			content, err = c.generateOpenAI(req)
		case "gemini":
			content, err = c.generateGemini(req)
		case "zai":
			content, err = c.generateZAI(req)
		case "claude":
			content, err = c.generateClaude(req)
		default:
			err = fmt.Errorf("unsupported provider: %s", c.config.Provider)
		}

		// Stop spinner if we started one
		if spinnerChan != nil {
			spinnerChan <- true
			fmt.Printf("\r%s\n", strings.Repeat(" ", len(loadingText)+20)) // Clear spinner line
		}

		// Render content with markdown if successful and not streaming
		if err == nil && !req.Stream {
			markdownRenderer, _ := ui.NewMarkdownRenderer()
			content = markdownRenderer.RenderResponse(content)
		}

		respChan <- GenerateResponse{
			Content: content,
			Error:   err,
		}
	}()

	return respChan
}

// generateOpenAI handles OpenAI API requests
func (c *Client) generateOpenAI(req GenerateRequest) (string, error) {
	// OpenAI API request structure
	type openAIMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type openAIRequest struct {
		Model    string          `json:"model"`
		Messages []openAIMessage `json:"messages"`
		Stream   bool            `json:"stream"`
	}

	type openAIResponse struct {
		Choices []struct {
			Message openAIMessage `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	// Build messages array
	messages := []openAIMessage{
		{Role: "system", Content: req.SystemPrompt},
		{Role: "user", Content: req.Context + "\n\n" + req.Task},
	}

	// Create request
	payload := openAIRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// generateZAI handles Z.AI API requests
func (c *Client) generateZAI(req GenerateRequest) (string, error) {
	// Z.AI API request structure (compatible with OpenAI format)
	type zaiMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type zaiRequest struct {
		Model       string       `json:"model"`
		Messages    []zaiMessage `json:"messages"`
		Temperature float64      `json:"temperature"`
		Stream      bool         `json:"stream"`
	}

	type zaiResponse struct {
		Choices []struct {
			Message zaiMessage `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	// Build messages array
	messages := []zaiMessage{
		{Role: "system", Content: req.SystemPrompt},
		{Role: "user", Content: req.Context + "\n\n" + req.Task},
	}

	// Create request - using GLM-4.6 as default model if not specified
	model := c.config.Model
	if model == "" {
		model = "glm-4.6"
	}

	payload := zaiRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 1.0,
		Stream:      false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Z.AI API
	httpReq, err := http.NewRequest("POST", "https://api.z.ai/api/coding/paas/v4/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept-Language", "en-US,en")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var zaiResp zaiResponse
	if err := json.Unmarshal(respBody, &zaiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if zaiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", zaiResp.Error.Message)
	}

	if len(zaiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return zaiResp.Choices[0].Message.Content, nil
}

// generateClaude handles Anthropic Claude API requests
func (c *Client) generateClaude(req GenerateRequest) (string, error) {
	// Claude API request structure
	type claudeMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type claudeRequest struct {
		Model     string         `json:"model"`
		MaxTokens int           `json:"max_tokens"`
		Messages  []claudeMessage `json:"messages"`
		System    string        `json:"system,omitempty"`
	}

	type claudeResponse struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model      string `json:"model"`
		StopReason string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence,omitempty"`
		Error      *struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	// Build messages array (Claude doesn't use system role in messages)
	messages := []claudeMessage{
		{Role: "user", Content: req.Context + "\n\n" + req.Task},
	}

	// Create request - using claude-3-5-sonnet as default model if not specified
	model := c.config.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	payload := claudeRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  messages,
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		payload.System = req.SystemPrompt
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Claude API
	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return claudeResp.Content[0].Text, nil
}

// generateGemini handles Google Gemini API requests
func (c *Client) generateGemini(_ GenerateRequest) (string, error) {
	// TODO: Implement Gemini API integration
	return "", fmt.Errorf("Gemini provider not yet implemented")
}

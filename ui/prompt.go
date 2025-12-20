package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/revrost/glimpse/styles"
)

// PromptProvider prompts the user to select an LLM provider
func PromptProvider() (string, error) {
	fmt.Println(styles.CreateHeader("Select LLM Provider"))
	fmt.Println(Separator(60))
	
	fmt.Println("Available providers:")
	fmt.Printf("  1) OpenAI (GPT-4o, GPT-3.5-turbo)\n")
	fmt.Printf("  2) Z.AI (GLM-4.6)\n")
	fmt.Printf("  3) Claude (Claude-3.5-Sonnet)\n")
	fmt.Printf("  4) Gemini (Coming Soon)\n")
	fmt.Println(Separator(60))
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter provider number (1-4): ")
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	
	input = strings.TrimSpace(input)
	
	switch input {
	case "1", "openai":
		return "openai", nil
	case "2", "zai":
		return "zai", nil
	case "3", "claude":
		return "claude", nil
	case "4", "gemini":
		return "", fmt.Errorf("Gemini provider is not yet implemented")
	default:
		return "", fmt.Errorf("invalid selection: %s", input)
	}
}

// PromptModel prompts the user to select a model for the given provider
func PromptModel(provider string) (string, error) {
	fmt.Println(styles.CreateHeader(fmt.Sprintf("Select %s Model", strings.ToUpper(provider))))
	fmt.Println(Separator(60))
	
	reader := bufio.NewReader(os.Stdin)
	
	switch provider {
	case "openai":
		fmt.Println("Available models:")
		fmt.Printf("  1) gpt-4o (recommended)\n")
		fmt.Printf("  2) gpt-4-turbo\n")
		fmt.Printf("  3) gpt-3.5-turbo\n")
		fmt.Println(Separator(60))
		fmt.Print("Enter model number (1-3) or custom model name: ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			return "gpt-4o", nil
		case "2":
			return "gpt-4-turbo", nil
		case "3":
			return "gpt-3.5-turbo", nil
		default:
			if input != "" {
				return input, nil
			}
		}
		
	case "zai":
		fmt.Println("Available models:")
		fmt.Printf("  1) glm-4.6 (recommended)\n")
		fmt.Printf("  2) glm-4\n")
		fmt.Printf("  3) glm-3-turbo\n")
		fmt.Println(Separator(60))
		fmt.Print("Enter model number (1-3) or custom model name: ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			return "glm-4.6", nil
		case "2":
			return "glm-4", nil
		case "3":
			return "glm-3-turbo", nil
		default:
			if input != "" {
				return input, nil
			}
		}
		
	case "claude":
		fmt.Println("Available models:")
		fmt.Printf("  1) claude-3-5-sonnet-20241022 (recommended)\n")
		fmt.Printf("  2) claude-3-opus-20240229\n")
		fmt.Printf("  3) claude-3-sonnet-20240229\n")
		fmt.Println(Separator(60))
		fmt.Print("Enter model number (1-3) or custom model name: ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			return "claude-3-5-sonnet-20241022", nil
		case "2":
			return "claude-3-opus-20240229", nil
		case "3":
			return "claude-3-sonnet-20240229", nil
		default:
			if input != "" {
				return input, nil
			}
		}
	}
	
	// Default fallback
	return "", fmt.Errorf("no model selected")
}

// ShowAPIKeyHelp shows helpful information about setting up API keys
func ShowAPIKeyHelp(provider string) {
	fmt.Println(styles.CreateHeader("API Key Setup"))
	fmt.Println(Separator(60))
	
	switch provider {
	case "openai":
		fmt.Printf("To use OpenAI, you need to set your API key:\n\n")
		fmt.Printf("  export OPENAI_API_KEY=\"your-api-key-here\"\n\n")
		fmt.Printf("Or add it to your shell profile (~/.zshrc, ~/.bashrc, etc.)\n\n")
		fmt.Printf("Get your API key from: https://platform.openai.com/api-keys\n")
		
	case "zai":
		fmt.Printf("To use Z.AI, you need to set your API key:\n\n")
		fmt.Printf("  export ZAI_API_KEY=\"your-api-key-here\"\n\n")
		fmt.Printf("Or add it to your shell profile (~/.zshrc, ~/.bashrc, etc.)\n\n")
		fmt.Printf("Get your API key from: https://z.ai\n")
		
	case "claude":
		fmt.Printf("To use Claude (Anthropic), you need to set your API key:\n\n")
		fmt.Printf("  export ANTHROPIC_API_KEY=\"your-api-key-here\"\n\n")
		fmt.Printf("Or add it to your shell profile (~/.zshrc, ~/.bashrc, etc.)\n\n")
		fmt.Printf("Get your API key from: https://console.anthropic.com/\n")
	}
	
	fmt.Println(Separator(60))
	fmt.Println(styles.Info.Render("After setting your API key, restart Glimpse to use the new configuration."))
}
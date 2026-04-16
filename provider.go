package llmprovider

import (
	"fmt"
	"os"
)

// Provider defaults. Override by setting the appropriate env var.
const (
	DefaultAnthropicModel = "claude-haiku-4-5-20251001"
	DefaultOpenAIModel    = "gpt-4o-mini"
	DefaultQwenModel      = "qwen3-max"
	DefaultQwenBaseURL    = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
	DefaultOpenAIBaseURL  = "https://api.openai.com/v1"
)

// NewFromEnv constructs an LLMClient based on environment variables.
//
// Set LLM_PROVIDER to select a provider:
//   - "anthropic" (default) — requires CLAUDE_API_KEY
//   - "openai"              — requires OPENAI_API_KEY
//   - "qwen"                — requires QWEN_API_KEY (optionally QWEN_BASE_URL)
//
// Set LLM_MODEL to override the default model for the chosen provider.
//
// Example:
//
//	os.Setenv("LLM_PROVIDER", "qwen")
//	os.Setenv("QWEN_API_KEY", "sk-...")
//	client, err := llmprovider.NewFromEnv()
//	text, usage, err := client.Call(ctx, "Hello", 100)
func NewFromEnv() (LLMClient, error) {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "anthropic"
	}
	modelOverride := os.Getenv("LLM_MODEL")

	switch provider {
	case "anthropic":
		apiKey := os.Getenv("CLAUDE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("CLAUDE_API_KEY required for anthropic provider")
		}
		model := DefaultAnthropicModel
		if modelOverride != "" {
			model = modelOverride
		}
		return NewClient(apiKey, model), nil

	case "qwen":
		apiKey := os.Getenv("QWEN_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("QWEN_API_KEY required for qwen provider")
		}
		model := DefaultQwenModel
		if modelOverride != "" {
			model = modelOverride
		}
		baseURL := os.Getenv("QWEN_BASE_URL")
		if baseURL == "" {
			baseURL = DefaultQwenBaseURL
		}
		return NewOpenAIClient(apiKey, model, baseURL), nil

	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY required for openai provider")
		}
		model := DefaultOpenAIModel
		if modelOverride != "" {
			model = modelOverride
		}
		return NewOpenAIClient(apiKey, model, DefaultOpenAIBaseURL), nil

	default:
		return nil, fmt.Errorf("unknown LLM_PROVIDER %q (supported: anthropic, openai, qwen)", provider)
	}
}

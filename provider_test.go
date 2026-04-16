package llmprovider

import (
	"os"
	"strings"
	"testing"
)

func TestCleanJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain text", "hello", "hello"},
		{"trimmed whitespace", "  {\"k\":1}  ", "{\"k\":1}"},
		{"json fence", "```json\n{\"k\":1}\n```", "{\"k\":1}"},
		{"plain fence", "```\n{\"k\":1}\n```", "{\"k\":1}"},
		{"fence without newlines", "```{\"k\":1}```", "{\"k\":1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanJSON(tt.in); got != tt.want {
				t.Errorf("CleanJSON(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewFromEnv_DefaultsToAnthropic(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("CLAUDE_API_KEY", "sk-test")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(*Client); !ok {
		t.Errorf("expected *Client (Anthropic), got %T", c)
	}
}

func TestNewFromEnv_Qwen(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "qwen")
	t.Setenv("QWEN_API_KEY", "sk-test")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(*OpenAIClient); !ok {
		t.Errorf("expected *OpenAIClient, got %T", c)
	}
}

func TestNewFromEnv_OpenAI(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "openai")
	t.Setenv("OPENAI_API_KEY", "sk-test")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(*OpenAIClient); !ok {
		t.Errorf("expected *OpenAIClient, got %T", c)
	}
}

func TestNewFromEnv_MissingKey(t *testing.T) {
	// Clear all keys so we hit the "required" paths
	for _, key := range []string{"CLAUDE_API_KEY", "QWEN_API_KEY", "OPENAI_API_KEY"} {
		_ = os.Unsetenv(key)
	}

	for _, provider := range []string{"anthropic", "qwen", "openai"} {
		t.Run(provider, func(t *testing.T) {
			t.Setenv("LLM_PROVIDER", provider)
			_, err := NewFromEnv()
			if err == nil {
				t.Fatal("expected error for missing API key, got nil")
			}
			if !strings.Contains(err.Error(), "required") {
				t.Errorf("error should mention 'required', got: %v", err)
			}
		})
	}
}

func TestNewFromEnv_UnknownProvider(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "not-a-real-provider")
	_, err := NewFromEnv()
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown', got: %v", err)
	}
}

func TestNewFromEnv_ModelOverride(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "anthropic")
	t.Setenv("CLAUDE_API_KEY", "sk-test")
	t.Setenv("LLM_MODEL", "claude-sonnet-4-6")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	client := c.(*Client)
	if client.model != "claude-sonnet-4-6" {
		t.Errorf("expected model override, got %q", client.model)
	}
}

func TestInterfaceSatisfaction(t *testing.T) {
	// Compile-time check that concrete types satisfy LLMClient.
	var _ LLMClient = (*Client)(nil)
	var _ LLMClient = (*OpenAIClient)(nil)
}

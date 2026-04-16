// Package llmprovider is a provider-agnostic Go client for large language models.
//
// One interface (LLMClient), three implementations:
//   - Anthropic Claude via the native Messages API
//   - Any OpenAI-compatible endpoint (OpenAI, Alibaba Qwen/DashScope, Azure OpenAI, etc.)
//
// Pick a provider at runtime via environment variables with NewFromEnv(), or
// construct a client directly with NewClient / NewOpenAIClient.
package llmprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LLMClient is the single interface every provider satisfies.
// Implementations must be safe for concurrent use by multiple goroutines.
type LLMClient interface {
	Call(ctx context.Context, prompt string, maxTokens int) (string, Usage, error)
}

// Usage reports token consumption for a single API call.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Client is an LLMClient for Anthropic's Messages API.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

// NewClient creates an Anthropic client for the given model.
// Timeouts default to 30 seconds; use NewClientWithHTTP to override.
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// NewClientWithHTTP creates an Anthropic client with a custom http.Client
// (useful for configuring timeouts, transport, proxies, or tracing).
func NewClientWithHTTP(apiKey, model string, httpClient *http.Client) *Client {
	return &Client{apiKey: apiKey, model: model, http: httpClient}
}

// Call sends a single user prompt and returns the raw text response.
// Markdown code fences around JSON responses are stripped automatically.
func (c *Client) Call(ctx context.Context, prompt string, maxTokens int) (string, Usage, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", Usage{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		snippet := string(respBody)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return "", Usage{}, fmt.Errorf("api returned %d: %s", resp.StatusCode, snippet)
	}

	var cr anthropicResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return "", Usage{}, fmt.Errorf("parse response: %w", err)
	}

	if len(cr.Content) == 0 || cr.Content[0].Type != "text" {
		return "", Usage{}, fmt.Errorf("no text content in response")
	}

	text := CleanJSON(cr.Content[0].Text)
	usage := Usage{InputTokens: cr.Usage.InputTokens, OutputTokens: cr.Usage.OutputTokens}
	return text, usage, nil
}

// CleanJSON strips markdown code fences and trims whitespace from a response.
// Useful when you asked the model for JSON and it wrapped the output in
// ```json ... ``` fences despite your instructions.
func CleanJSON(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

// --- internal request/response types ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

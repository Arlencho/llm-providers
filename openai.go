package llmprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient is an LLMClient for any OpenAI-compatible Chat Completions API.
//
// Works with:
//   - OpenAI (https://api.openai.com/v1)
//   - Alibaba Qwen / DashScope (https://dashscope-intl.aliyuncs.com/compatible-mode/v1)
//   - Azure OpenAI (your deployment URL)
//   - Groq, Together, OpenRouter, and any other OpenAI-compatible gateway
type OpenAIClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewOpenAIClient creates a client for an OpenAI-compatible endpoint.
// baseURL should NOT include a trailing "/chat/completions".
func NewOpenAIClient(apiKey, model, baseURL string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// NewOpenAIClientWithHTTP creates a client with a custom http.Client.
func NewOpenAIClientWithHTTP(apiKey, model, baseURL string, httpClient *http.Client) *OpenAIClient {
	return &OpenAIClient{apiKey: apiKey, model: model, baseURL: baseURL, http: httpClient}
}

// Call sends a single user prompt and returns the raw text response.
func (c *OpenAIClient) Call(ctx context.Context, prompt string, maxTokens int) (string, Usage, error) {
	reqBody := oaiRequest{
		Model:     c.model,
		Messages:  []oaiMessage{{Role: "user", Content: prompt}},
		MaxTokens: maxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", Usage{}, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var cr oaiResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return "", Usage{}, fmt.Errorf("parse response: %w", err)
	}

	if len(cr.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("no choices in response")
	}

	text := CleanJSON(cr.Choices[0].Message.Content)
	usage := Usage{
		InputTokens:  cr.Usage.PromptTokens,
		OutputTokens: cr.Usage.CompletionTokens,
	}
	return text, usage, nil
}

// --- OpenAI-compatible request/response types ---

type oaiRequest struct {
	Model     string       `json:"model"`
	Messages  []oaiMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
}

type oaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

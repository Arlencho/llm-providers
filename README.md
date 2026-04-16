# llm-providers

[![Go Reference](https://pkg.go.dev/badge/github.com/arlencho/llm-providers.svg)](https://pkg.go.dev/github.com/arlencho/llm-providers)
[![Go Report Card](https://goreportcard.com/badge/github.com/arlencho/llm-providers)](https://goreportcard.com/report/github.com/arlencho/llm-providers)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A small, provider-agnostic Go client for large language models.

One interface, three providers:

- **Anthropic Claude** (native Messages API)
- **OpenAI** and any **OpenAI-compatible** endpoint — Alibaba Qwen/DashScope, Azure OpenAI, Groq, Together, OpenRouter

Swap between them by changing one environment variable. No vendor lock-in, no SDK bloat, no runtime dependencies beyond the Go standard library.

## Why

I was paying a lot for one LLM provider on a production pipeline. Halfway through a month the provider's free credits ran out. I wanted to switch to a cheaper one for a few weeks while waiting for the next billing cycle, but switching required editing a dozen call sites across the codebase.

The fix was to hide the provider behind a single interface. That interface is this library.

## Install

```bash
go get github.com/arlencho/llm-providers
```

## Quick start

Pick a provider with environment variables:

```go
package main

import (
    "context"
    "fmt"
    "log"

    llm "github.com/arlencho/llm-providers"
)

func main() {
    // LLM_PROVIDER=anthropic + CLAUDE_API_KEY
    // LLM_PROVIDER=openai    + OPENAI_API_KEY
    // LLM_PROVIDER=qwen      + QWEN_API_KEY
    client, err := llm.NewFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    text, usage, err := client.Call(context.Background(), "Say hi in one word.", 50)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(text)
    fmt.Printf("in=%d out=%d\n", usage.InputTokens, usage.OutputTokens)
}
```

Or construct a client directly:

```go
// Anthropic Claude
c := llm.NewClient(os.Getenv("CLAUDE_API_KEY"), "claude-haiku-4-5-20251001")

// OpenAI
c := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")

// Alibaba Qwen via DashScope
c := llm.NewOpenAIClient(
    os.Getenv("QWEN_API_KEY"),
    "qwen3-max",
    "https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
)
```

## The interface

```go
type LLMClient interface {
    Call(ctx context.Context, prompt string, maxTokens int) (string, Usage, error)
}
```

That's the whole surface area. If you need structured output, ask for JSON in your prompt and use `llm.CleanJSON(text)` to strip markdown code fences that models sometimes add anyway.

## What this library is *not*

- **Not a full SDK.** No streaming, no tool calling, no vision, no multi-turn chat. Just "send a prompt, get text back."
- **Not a benchmarking harness.** Swap providers for cost/availability reasons — benchmark your own prompts.
- **Not production-grade for every use case.** It's ~300 lines of Go. Read it. Fork it. Adapt it.

If you need all the features of each provider's official SDK, use those SDKs. If you need the 20% of features that cover 80% of LLM use cases and care about swap-ability, this is for you.

## Supported environment variables

| Variable | Purpose |
|---|---|
| `LLM_PROVIDER` | `anthropic` (default), `openai`, or `qwen` |
| `LLM_MODEL` | Override the default model for the chosen provider |
| `CLAUDE_API_KEY` | Anthropic API key (when `LLM_PROVIDER=anthropic`) |
| `OPENAI_API_KEY` | OpenAI API key |
| `QWEN_API_KEY` | Alibaba DashScope API key |
| `QWEN_BASE_URL` | Override the DashScope endpoint (for different regions) |

## Tests

```bash
go test ./...
```

Tests cover the provider factory, JSON cleanup, and interface satisfaction. The actual HTTP calls aren't mocked — if you want to exercise them, set real API keys and write an integration test.

## License

MIT © Arlen Rios

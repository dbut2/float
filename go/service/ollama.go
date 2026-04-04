package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	client *api.Client
	model  string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	u, _ := url.Parse(baseURL)
	httpClient := &http.Client{Timeout: 120 * time.Second}
	return &OllamaClient{
		client: api.NewClient(u, httpClient),
		model:  model,
	}
}

func (c *OllamaClient) Model() string {
	return c.model
}

var classificationSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"bucket_name": {"type": "string"},
		"confidence": {"type": "number"},
		"reasoning": {"type": "string"}
	},
	"required": ["bucket_name", "confidence", "reasoning"]
}`)

func (c *OllamaClient) Classify(ctx context.Context, prompt string) (ClassificationResult, error) {
	stream := false
	req := &api.ChatRequest{
		Model: c.model,
		Messages: []api.Message{
			{Role: "user", Content: prompt},
		},
		Format:  classificationSchema,
		Stream:  &stream,
		Options: map[string]any{"temperature": 0},
	}

	var result ClassificationResult
	var lastResp api.ChatResponse

	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		lastResp = resp
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("ollama chat: %w", err)
	}

	if err := json.Unmarshal([]byte(lastResp.Message.Content), &result); err != nil {
		return result, fmt.Errorf("parse classification: %w", err)
	}

	return result, nil
}

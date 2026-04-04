package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	client openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

func (c *OpenAIClient) Model() string {
	return c.model
}

func (c *OpenAIClient) Classify(ctx context.Context, prompt string) (ClassificationResult, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bucket_name": map[string]any{"type": "string"},
			"confidence":  map[string]any{"type": "number"},
			"reasoning":   map[string]any{"type": "string"},
		},
		"required":             []string{"bucket_name", "confidence", "reasoning"},
		"additionalProperties": false,
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return ClassificationResult{}, fmt.Errorf("marshal schema: %w", err)
	}

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "classification",
					Schema: json.RawMessage(schemaBytes),
					Strict: openai.Bool(true),
				},
			},
		},
		Temperature: openai.Float(0),
	})
	if err != nil {
		return ClassificationResult{}, fmt.Errorf("openai chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ClassificationResult{}, fmt.Errorf("no choices in response")
	}

	var result ClassificationResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return ClassificationResult{}, fmt.Errorf("parse classification: %w", err)
	}

	return result, nil
}

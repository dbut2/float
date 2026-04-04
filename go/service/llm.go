package service

import "context"

type ClassificationResult struct {
	BucketName string  `json:"bucket_name"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

type LLMClient interface {
	Classify(ctx context.Context, prompt string) (ClassificationResult, error)
	Model() string
}

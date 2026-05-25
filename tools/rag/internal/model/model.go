package model

import "context"

type Message struct {
	Role    string
	Content string
}

type GenerateRequest struct {
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature float64
	Metadata    map[string]string
}

type GenerateResponse struct {
	Text         string
	Model        string
	InputTokens  int
	OutputTokens int
}

type Client interface {
	Generate(context.Context, GenerateRequest) (GenerateResponse, error)
	Name() string
}

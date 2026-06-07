package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
)

const (
	defaultBaseURL = "https://api.deepseek.com"
	defaultModel   = "deepseek-v4-flash"
)

type Client struct {
	apiKey     string
	baseURL    string
	modelName  string
	httpClient *http.Client
}

type requestBody struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Messages    []openAIMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	Thinking    *thinkingConfig `json:"thinking,omitempty"`
}

type thinkingConfig struct {
	Type            string `json:"type"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewFromEnv() (*Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("deepseek credentials are not set; configure DEEPSEEK_API_KEY")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	modelName := strings.TrimSpace(os.Getenv("DEEPSEEK_MODEL"))
	if modelName == "" {
		modelName = defaultModel
	}
	return &Client{
		apiKey:    apiKey,
		baseURL:   baseURL,
		modelName: modelName,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

func (c *Client) Name() string {
	return "deepseek:" + c.modelName
}

func (c *Client) Generate(ctx context.Context, req model.GenerateRequest) (model.GenerateResponse, error) {
	if c == nil {
		return model.GenerateResponse{}, fmt.Errorf("deepseek client is nil")
	}
	payload := requestBody{
		Model:       c.modelName,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		System:      req.System,
	}
	// Enable thinking for deepseek-v4 models
	if strings.HasPrefix(c.modelName, "deepseek-v4") {
		thinking := "enabled"
		payload.Thinking = &thinkingConfig{
			Type:            thinking,
			ReasoningEffort: "high",
		}
	}
	for _, msg := range req.Messages {
		payload.Messages = append(payload.Messages, openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	if payload.MaxTokens <= 0 {
		payload.MaxTokens = 2048
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("marshal deepseek request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("create deepseek request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("call deepseek: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("read deepseek response: %w", err)
	}
	var parsed responseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return model.GenerateResponse{}, fmt.Errorf("decode deepseek response: %w", err)
	}
	if resp.StatusCode >= 300 {
		if parsed.Error != nil {
			return model.GenerateResponse{}, fmt.Errorf("deepseek %s: %s", parsed.Error.Type, parsed.Error.Message)
		}
		return model.GenerateResponse{}, fmt.Errorf("deepseek request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	if len(parsed.Choices) == 0 {
		return model.GenerateResponse{}, fmt.Errorf("deepseek response has no choices")
	}
	return model.GenerateResponse{
		Text:         parsed.Choices[0].Message.Content,
		Model:        parsed.Model,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
	}, nil
}

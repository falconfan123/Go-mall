package anthropic

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
	defaultBaseURL = "https://api.anthropic.com"
	defaultVersion = "2023-06-01"
	defaultModel   = "claude-sonnet-4-20250514"
)

type Client struct {
	authMode   string
	authValue  string
	baseURL    string
	version    string
	modelName  string
	httpClient *http.Client
}

const (
	authModeAPIKey = "x-api-key"
	authModeBearer = "bearer"
)

type requestBody struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewFromEnv() (*Client, error) {
	authToken := strings.TrimSpace(os.Getenv("ANTHROPIC_AUTH_TOKEN"))
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	authMode := ""
	authValue := ""
	switch {
	case authToken != "":
		authMode = authModeBearer
		authValue = authToken
	case apiKey != "":
		authMode = authModeAPIKey
		authValue = apiKey
	default:
		return nil, fmt.Errorf("anthropic credentials are not set; configure ANTHROPIC_AUTH_TOKEN or ANTHROPIC_API_KEY (ANTHROPIC_AUTH_TOKEN takes precedence)")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	version := strings.TrimSpace(os.Getenv("ANTHROPIC_VERSION"))
	if version == "" {
		version = defaultVersion
	}
	modelName := strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL"))
	if modelName == "" {
		modelName = defaultModel
	}
	return &Client{
		authMode:  authMode,
		authValue: authValue,
		baseURL:   baseURL,
		version:   version,
		modelName: modelName,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

func (c *Client) Name() string {
	return "anthropic:" + c.modelName
}

func (c *Client) Generate(ctx context.Context, req model.GenerateRequest) (model.GenerateResponse, error) {
	if c == nil {
		return model.GenerateResponse{}, fmt.Errorf("anthropic client is nil")
	}
	payload := requestBody{
		Model:       c.modelName,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		System:      req.System,
		Metadata:    req.Metadata,
	}
	for _, msg := range req.Messages {
		payload.Messages = append(payload.Messages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	if payload.MaxTokens <= 0 {
		payload.MaxTokens = 2048
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("marshal anthropic request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("create anthropic request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	switch c.authMode {
	case authModeBearer:
		httpReq.Header.Set("Authorization", "Bearer "+c.authValue)
	case authModeAPIKey:
		httpReq.Header.Set("x-api-key", c.authValue)
	default:
		return model.GenerateResponse{}, fmt.Errorf("anthropic auth mode is invalid: %q", c.authMode)
	}
	httpReq.Header.Set("anthropic-version", c.version)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return model.GenerateResponse{}, fmt.Errorf("read anthropic response: %w", err)
	}
	var parsed responseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return model.GenerateResponse{}, fmt.Errorf("decode anthropic response: %w", err)
	}
	if resp.StatusCode >= 300 {
		if parsed.Error != nil {
			return model.GenerateResponse{}, fmt.Errorf("anthropic %s: %s", parsed.Error.Type, parsed.Error.Message)
		}
		return model.GenerateResponse{}, fmt.Errorf("anthropic request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var parts []string
	for _, item := range parsed.Content {
		if item.Type == "text" && item.Text != "" {
			parts = append(parts, item.Text)
		}
	}
	return model.GenerateResponse{
		Text:         strings.Join(parts, "\n"),
		Model:        parsed.Model,
		InputTokens:  parsed.Usage.InputTokens,
		OutputTokens: parsed.Usage.OutputTokens,
	}, nil
}

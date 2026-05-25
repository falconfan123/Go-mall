package anthropic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
)

func TestNewFromEnvRequiresCredentials(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := NewFromEnv()
	if err == nil {
		t.Fatal("NewFromEnv() error = nil, want error")
	}
	for _, part := range []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_API_KEY",
		"takes precedence",
	} {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("NewFromEnv() error = %q, want substring %q", err.Error(), part)
		}
	}
}

func TestGenerateUsesBearerTokenWhenConfigured(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "token-123")
	t.Setenv("ANTHROPIC_API_KEY", "")

	reqPath, authHeader, apiKeyHeader := exerciseGenerate(t, "")

	if reqPath != "/v1/messages" {
		t.Fatalf("request path = %q, want /v1/messages", reqPath)
	}
	if authHeader != "Bearer token-123" {
		t.Fatalf("Authorization header = %q, want bearer token", authHeader)
	}
	if apiKeyHeader != "" {
		t.Fatalf("x-api-key header = %q, want empty", apiKeyHeader)
	}
}

func TestGenerateUsesAPIKeyWhenTokenMissing(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "key-123")

	_, authHeader, apiKeyHeader := exerciseGenerate(t, "")

	if authHeader != "" {
		t.Fatalf("Authorization header = %q, want empty", authHeader)
	}
	if apiKeyHeader != "key-123" {
		t.Fatalf("x-api-key header = %q, want API key", apiKeyHeader)
	}
}

func TestGeneratePrefersBearerTokenOverAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "token-123")
	t.Setenv("ANTHROPIC_API_KEY", "key-123")

	_, authHeader, apiKeyHeader := exerciseGenerate(t, "")

	if authHeader != "Bearer token-123" {
		t.Fatalf("Authorization header = %q, want bearer token", authHeader)
	}
	if apiKeyHeader != "" {
		t.Fatalf("x-api-key header = %q, want empty when token is set", apiKeyHeader)
	}
}

func TestGenerateUsesConfiguredBaseURL(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "token-123")
	t.Setenv("ANTHROPIC_API_KEY", "")

	reqPath, _, _ := exerciseGenerate(t, "/gateway/")

	if reqPath != "/gateway/v1/messages" {
		t.Fatalf("request path = %q, want /gateway/v1/messages", reqPath)
	}
}

func exerciseGenerate(t *testing.T, basePath string) (string, string, string) {
	t.Helper()

	var reqPath string
	var authHeader string
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath = r.URL.Path
		authHeader = r.Header.Get("Authorization")
		apiKeyHeader = r.Header.Get("x-api-key")
		if got := r.Header.Get("anthropic-version"); got != defaultVersion {
			t.Fatalf("anthropic-version = %q, want %q", got, defaultVersion)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"test-model","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":2}}`))
	}))
	t.Cleanup(server.Close)

	t.Setenv("ANTHROPIC_BASE_URL", server.URL+basePath)
	t.Setenv("ANTHROPIC_MODEL", "test-model")
	t.Setenv("ANTHROPIC_VERSION", "")

	client, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv() error = %v", err)
	}
	_, err = client.Generate(context.Background(), model.GenerateRequest{
		Messages: []model.Message{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	return reqPath, authHeader, apiKeyHeader
}

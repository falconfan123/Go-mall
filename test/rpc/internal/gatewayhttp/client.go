package gatewayhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
)

const defaultGatewayBaseURL = "http://127.0.0.1:8888"

type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultGatewayBaseURL,
		httpClient: &http.Client{
			Timeout: testenv.Timeout(),
		},
		headers: make(http.Header),
	}
}

func (c *Client) clone() *Client {
	headers := make(http.Header, len(c.headers))
	for k, values := range c.headers {
		copyValues := append([]string(nil), values...)
		headers[k] = copyValues
	}

	return &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		headers:    headers,
	}
}

func (c *Client) WithTokens(shortToken, longToken string) *Client {
	next := c.clone()
	if shortToken != "" {
		next.headers.Set("Short-Token", shortToken)
	}
	if longToken != "" {
		next.headers.Set("Long-Token", longToken)
	}
	return next
}

func (c *Client) WithUserID(userID uint32) *Client {
	next := c.clone()
	next.headers.Set("user_id", fmt.Sprintf("%d", userID))
	next.headers.Set("X-User-Id", fmt.Sprintf("%d", userID))
	return next
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query map[string]string, body any, out any) (*http.Response, []byte, error) {
	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, nil, err
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path

	values := endpoint.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	endpoint.RawQuery = values.Encode()

	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), payload)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return resp, data, err
		}
	}

	return resp, data, nil
}

func RequireStatusOK(t *testing.T, resp *http.Response, body []byte) {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected http status %d body=%s", resp.StatusCode, string(body))
	}
}

func Eventually(t *testing.T, timeout time.Duration, fn func() error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := fn(); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(300 * time.Millisecond)
	}
	if lastErr != nil {
		t.Fatal(lastErr)
	}
}

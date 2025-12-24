package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type Client struct {
	baseURL *url.URL
	apiKey  string
	model   string
	http    *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func NewClient(baseURL, apiKey, model string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	return &Client{
		baseURL: parsed,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	payload := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.2,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal chat completion: %w", err)
	}
	endpoint := c.buildURL("/v1/chat/completions")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm api error: %s", string(payload))
	}

	var decoded ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return decoded.Choices[0].Message.Content, nil
}

func (c *Client) buildURL(p string) string {
	cpy := *c.baseURL
	cpy.Path = path.Join(c.baseURL.Path, p)
	return cpy.String()
}

// Package llm provides an LLM client interface and an Ollama implementation for text translation.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"translator-service/pkg/models"
)

// Translator defines the contract for text and localization translation services.
type Translator interface {
	TranslateSingle(ctx context.Context, source, target, text string) (string, error)
	TranslateLocaleFile(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error)
	ModelName() string
}

// Client implements Translator using a local Ollama REST API endpoint.
type Client struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// Ensure Client implements Translator.
var _ Translator = (*Client)(nil)

// NewClient creates a new Ollama LLM client with the specified base URL, model name, and timeout.
func NewClient(baseURL, model string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ModelName returns the identifier of the configured model.
func (c *Client) ModelName() string {
	return c.Model
}

// TranslateSingle translates an individual text message, instructing the model to respond in JSON.
func (c *Client) TranslateSingle(ctx context.Context, source, target, text string) (string, error) {
	prompt := fmt.Sprintf(`You are a professional software localization translator.
Translate the following software UI text from %s to %s.
Return ONLY valid JSON matching this schema: {"translation": "..."}

Text: %s`, source, target, text)

	payload := map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"think":  false,
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	rawJSON, err := c.sendGenerateRequest(ctx, payload)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Translation string `json:"translation"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse translation response JSON: %w (raw response: %s)", err, rawJSON)
	}

	return parsed.Translation, nil
}

// TranslateLocaleFile translates localization entries directly, preserving keys and placeholders.
func (c *Client) TranslateLocaleFile(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error) {
	if len(items) == 0 {
		return []models.LocaleItem{}, nil
	}

	inputJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input items: %w", err)
	}

	prompt := fmt.Sprintf(`You are an automated software localization engine.
Translate the "text" field of each item from %s to %s.
CRITICAL RULES:
1. Do NOT modify the "id" fields under any circumstance.
2. Return ONLY a valid JSON object matching: {"texts": [{"id": "...", "text": "..."}]}
3. Maintain technical placeholders like {0}, {{name}}, %%s, <tag> intact.
4. Output must contain the exact same number of items as the input.

Input list:
%s`, source, target, string(inputJSON))

	payload := map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"think":  false,
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	rawJSON, err := c.sendGenerateRequest(ctx, payload)
	if err != nil {
		return nil, err
	}

	var parsed models.LocaleFileResponse
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse locale file JSON: %w (raw response: %s)", err, rawJSON)
	}

	return parsed.Texts, nil
}

func (c *Client) sendGenerateRequest(ctx context.Context, payload map[string]interface{}) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respSnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respSnippet))
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ImageResult represents the result of an image generation request.
type ImageResult struct {
	URL       string
	LocalPath string
	Revised   string
}

// ImageOptions contains optional parameters for image generation.
type ImageOptions struct {
	Size    string
	Style   string
	Quality string
}

// ImageProvider abstracts image generation backends (OpenAI DALL-E, etc.).
type ImageProvider interface {
	Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error)
	Name() string
}

// --- OpenAI Image Provider ---

// OpenAIImageProvider implements ImageProvider using the OpenAI Images API.
type OpenAIImageProvider struct {
	apiKey     string
	apiBase    string
	model      string
	httpClient *http.Client
}

// NewOpenAIImageProvider creates an OpenAIImageProvider with the given configuration.
func NewOpenAIImageProvider(apiKey, apiBase, model string) *OpenAIImageProvider {
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "dall-e-3"
	}
	return &OpenAIImageProvider{
		apiKey:  apiKey,
		apiBase: apiBase,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OpenAIImageProvider) Name() string { return "openai" }

func (p *OpenAIImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	reqBody := map[string]any{
		"model":           p.model,
		"prompt":          prompt,
		"n":               1,
		"size":            opts.Size,
		"style":           opts.Style,
		"quality":         opts.Quality,
		"response_format": "url",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/images/generations", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Data []struct {
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no image returned from API")
	}

	return &ImageResult{
		URL:     apiResp.Data[0].URL,
		Revised: apiResp.Data[0].RevisedPrompt,
	}, nil
}

package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// --- Fal.ai Image Provider ---

// FalImageProvider implements ImageProvider using the Fal.ai API.
type FalImageProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewFalImageProvider creates a FalImageProvider with the given configuration.
func NewFalImageProvider(apiKey, model string) *FalImageProvider {
	if model == "" {
		model = "fal-ai/flux/dev"
	}
	return &FalImageProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *FalImageProvider) Name() string { return "fal" }

func (p *FalImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	imageSize := "square_hd"
	switch opts.Size {
	case "1792x1024":
		imageSize = "landscape_16_9"
	case "1024x1792":
		imageSize = "portrait_16_9"
	}

	reqBody := map[string]any{
		"prompt":                 prompt,
		"image_size":             imageSize,
		"num_images":             1,
		"enable_safety_checker":  false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "https://fal.run/" + p.model
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fal.ai API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
		Prompt string `json:"prompt"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Images) == 0 {
		return nil, fmt.Errorf("no image returned from fal.ai API")
	}

	return &ImageResult{
		URL: apiResp.Images[0].URL,
	}, nil
}

// --- Replicate Image Provider ---

// ReplicateImageProvider implements ImageProvider using the Replicate API.
type ReplicateImageProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewReplicateImageProvider creates a ReplicateImageProvider with the given configuration.
func NewReplicateImageProvider(apiKey, model string) *ReplicateImageProvider {
	if model == "" {
		model = "black-forest-labs/flux-schnell"
	}
	return &ReplicateImageProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *ReplicateImageProvider) Name() string { return "replicate" }

func (p *ReplicateImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	reqBody := map[string]any{
		"input": map[string]any{
			"prompt":         prompt,
			"aspect_ratio":   "1:1",
			"output_format":  "png",
			"output_quality": 90,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	createURL := "https://api.replicate.com/v1/models/" + p.model + "/predictions"
	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("replicate API returned status %d: %s", resp.StatusCode, string(body))
	}

	var prediction struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		URLs   struct {
			Get string `json:"get"`
		} `json:"urls"`
		Output []string `json:"output"`
		Error  string   `json:"error"`
	}

	if err := json.Unmarshal(body, &prediction); err != nil {
		return nil, fmt.Errorf("failed to parse prediction response: %w", err)
	}

	// Poll until succeeded or failed
	pollURL := prediction.URLs.Get
	if pollURL == "" {
		pollURL = "https://api.replicate.com/v1/predictions/" + prediction.ID
	}

	for attempt := 0; attempt < 60; attempt++ {
		if prediction.Status == "succeeded" {
			break
		}
		if prediction.Status == "failed" || prediction.Status == "canceled" {
			return nil, fmt.Errorf("replicate prediction %s: %s", prediction.Status, prediction.Error)
		}

		time.Sleep(2 * time.Second)

		pollReq, err := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create poll request: %w", err)
		}
		pollReq.Header.Set("Authorization", "Token "+p.apiKey)

		pollResp, err := p.httpClient.Do(pollReq)
		if err != nil {
			return nil, fmt.Errorf("poll request failed: %w", err)
		}
		pollBody, err := io.ReadAll(io.LimitReader(pollResp.Body, 1*1024*1024))
		pollResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read poll response: %w", err)
		}

		if err := json.Unmarshal(pollBody, &prediction); err != nil {
			return nil, fmt.Errorf("failed to parse poll response: %w", err)
		}
	}

	if prediction.Status != "succeeded" {
		return nil, fmt.Errorf("replicate prediction did not complete in time (status: %s)", prediction.Status)
	}

	if len(prediction.Output) == 0 {
		return nil, fmt.Errorf("no output from replicate prediction")
	}

	return &ImageResult{
		URL: prediction.Output[0],
	}, nil
}

// --- Together.ai Image Provider ---

// TogetherImageProvider implements ImageProvider using the Together.ai API.
type TogetherImageProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewTogetherImageProvider creates a TogetherImageProvider with the given configuration.
func NewTogetherImageProvider(apiKey, model string) *TogetherImageProvider {
	if model == "" {
		model = "black-forest-labs/FLUX.1-schnell-Free"
	}
	return &TogetherImageProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *TogetherImageProvider) Name() string { return "together" }

func (p *TogetherImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	width, height := 1024, 1024
	switch opts.Size {
	case "1792x1024":
		width, height = 1792, 1024
	case "1024x1792":
		width, height = 1024, 1792
	}

	reqBody := map[string]any{
		"model":           p.model,
		"prompt":          prompt,
		"width":           width,
		"height":          height,
		"steps":           4,
		"n":               1,
		"response_format": "url",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.together.xyz/v1/images/generations", bytes.NewReader(bodyBytes))
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("together.ai API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no image returned from together.ai API")
	}

	return &ImageResult{
		URL: apiResp.Data[0].URL,
	}, nil
}

// --- Google Imagen Provider ---

// GoogleImageProvider implements ImageProvider using the Google Generative Language API.
type GoogleImageProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGoogleImageProvider creates a GoogleImageProvider with the given configuration.
func NewGoogleImageProvider(apiKey, model string) *GoogleImageProvider {
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}
	return &GoogleImageProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *GoogleImageProvider) Name() string { return "google" }

func (p *GoogleImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"responseModalities": []string{"TEXT", "IMAGE"},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/" + p.model + ":generateContent?key=" + p.apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text       string `json:"text"`
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Google API")
	}

	for _, part := range apiResp.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			imgData, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 image data: %w", err)
			}

			// Write to a temp file; image.go will pick up LocalPath
			ext := ".png"
			if part.InlineData.MimeType == "image/jpeg" {
				ext = ".jpg"
			}
			tmpFile, err := os.CreateTemp("", "google-image-*"+ext)
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file: %w", err)
			}
			tmpPath := tmpFile.Name()
			if _, err := tmpFile.Write(imgData); err != nil {
				tmpFile.Close()
				os.Remove(tmpPath)
				return nil, fmt.Errorf("failed to write temp image: %w", err)
			}
			tmpFile.Close()

			return &ImageResult{
				LocalPath: tmpPath,
			}, nil
		}
	}

	return nil, fmt.Errorf("no image data found in Google API response")
}

// googleImageTempToWorkspace moves a Google provider temp file into the workspace save dir.
// Called from image.go Execute() when result.LocalPath is set but result.URL is empty.
func googleImageTempToWorkspace(result *ImageResult, saveDir, filename string) error {
	if result.LocalPath == "" || result.URL != "" {
		return nil
	}
	destPath := filepath.Join(saveDir, filename)
	data, err := os.ReadFile(result.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to read temp image: %w", err)
	}
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write image to workspace: %w", err)
	}
	os.Remove(result.LocalPath)
	result.LocalPath = destPath
	return nil
}

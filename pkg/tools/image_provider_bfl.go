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

// BFLImageProvider generates images via Black Forest Labs (api.bfl.ml) using
// an async two-step API: submit → poll until Ready.
type BFLImageProvider struct {
	apiKey string
	model  string
}

func NewBFLImageProvider(apiKey, model string) *BFLImageProvider {
	return &BFLImageProvider{apiKey: apiKey, model: model}
}

func (p *BFLImageProvider) Name() string { return "bfl" }

func (p *BFLImageProvider) Generate(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error) {
	// Step 1: submit the generation request
	body, _ := json.Marshal(map[string]any{
		"prompt": prompt,
		"width":  1024,
		"height": 1024,
	})
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://api.bfl.ml/v1/%s", p.model), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("BFL: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-key", p.apiKey)

	resp, err := imageHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("BFL: submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("BFL: submit %d: %s", resp.StatusCode, b)
	}

	var submitResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&submitResp); err != nil || submitResp.ID == "" {
		return nil, fmt.Errorf("BFL: unexpected submit response")
	}

	// Step 2: poll until Ready (max ~2 minutes)
	for range 60 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		pollReq, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("https://api.bfl.ml/v1/get_result?id=%s", submitResp.ID), nil)
		if err != nil {
			return nil, fmt.Errorf("BFL: create poll request: %w", err)
		}
		pollReq.Header.Set("x-key", p.apiKey)

		pollResp, err := imageHTTPClient.Do(pollReq)
		if err != nil {
			return nil, fmt.Errorf("BFL: poll: %w", err)
		}

		var result struct {
			Status string `json:"status"`
			Result struct {
				Sample string `json:"sample"`
			} `json:"result"`
		}
		json.NewDecoder(pollResp.Body).Decode(&result)
		pollResp.Body.Close()

		if result.Status == "Ready" && result.Result.Sample != "" {
			return &ImageResult{URL: result.Result.Sample}, nil
		}
	}
	return nil, fmt.Errorf("BFL: generation timed out after 2 minutes")
}

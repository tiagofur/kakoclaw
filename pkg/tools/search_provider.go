package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SearchResult represents a single web search result.
type SearchResult struct {
	Title       string
	URL         string
	Description string
}

// SearchProvider abstracts web search backends (Brave, SearXNG, etc.).
type SearchProvider interface {
	Search(ctx context.Context, query string, count int) ([]SearchResult, error)
	Name() string
}

// --- Brave Search Provider ---

// BraveSearchProvider implements SearchProvider using the Brave Search API.
type BraveSearchProvider struct {
	apiKey string
}

// NewBraveSearchProvider creates a BraveSearchProvider with the given API key.
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey}
}

func (b *BraveSearchProvider) Name() string { return "brave" }

func (b *BraveSearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(query), count)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", b.apiKey)

	resp, err := webSearchHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Web.Results))
	for _, r := range searchResp.Web.Results {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
		})
	}
	return results, nil
}

// --- SearXNG Search Provider ---

// SearXNGSearchProvider implements SearchProvider using a self-hosted SearXNG instance.
type SearXNGSearchProvider struct {
	baseURL string
}

// NewSearXNGSearchProvider creates a SearXNGSearchProvider with the given instance URL.
func NewSearXNGSearchProvider(baseURL string) *SearXNGSearchProvider {
	return &SearXNGSearchProvider{baseURL: baseURL}
}

func (s *SearXNGSearchProvider) Name() string { return "searxng" }

func (s *SearXNGSearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("%s/search?q=%s&format=json&categories=general&pageno=1",
		s.baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := webSearchHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Results))
	for i, r := range searchResp.Results {
		if i >= count {
			break
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}
	return results, nil
}

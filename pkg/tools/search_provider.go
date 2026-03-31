package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sipeed/makoclaw/pkg/logger"
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
	apiKey  string
	baseURL string
}

// NewBraveSearchProvider creates a BraveSearchProvider with the given API key.
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey, baseURL: "https://api.search.brave.com"}
}

// NewBraveSearchProviderWithBaseURL creates a BraveSearchProvider with a custom base URL,
// useful for testing with httptest.NewServer.
func NewBraveSearchProviderWithBaseURL(apiKey, baseURL string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/")}
}

func (b *BraveSearchProvider) Name() string { return "brave" }

func (b *BraveSearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("%s/res/v1/web/search?q=%s&count=%d",
		b.baseURL, url.QueryEscape(query), count)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
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

// --- Tavily Search Provider ---

// TavilySearchProvider implements SearchProvider using the Tavily API.
type TavilySearchProvider struct {
	apiKey  string
	baseURL string
}

// NewTavilySearchProvider creates a TavilySearchProvider with the given API key.
func NewTavilySearchProvider(apiKey string) *TavilySearchProvider {
	return &TavilySearchProvider{apiKey: apiKey, baseURL: "https://api.tavily.com"}
}

// NewTavilySearchProviderWithBaseURL creates a TavilySearchProvider with a custom base URL,
// useful for testing with httptest.NewServer.
func NewTavilySearchProviderWithBaseURL(apiKey, baseURL string) *TavilySearchProvider {
	return &TavilySearchProvider{apiKey: apiKey, baseURL: baseURL}
}

func (t *TavilySearchProvider) Name() string { return "tavily" }

func (t *TavilySearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	type tavilyRequest struct {
		APIKey      string `json:"api_key"`
		Query       string `json:"query"`
		MaxResults  int    `json:"max_results"`
		SearchDepth string `json:"search_depth"`
	}
	type tavilyResponse struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	reqBody := tavilyRequest{
		APIKey:      t.apiKey,
		Query:       query,
		MaxResults:  count,
		SearchDepth: "basic",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("tavily: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/search", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("tavily: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := webSearchHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("tavily: failed to read response body: %w", err)
	}
	var result tavilyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tavily: failed to decode response: %w", err)
	}

	results := make([]SearchResult, 0, len(result.Results))
	for _, r := range result.Results {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}
	return results, nil
}

// --- Fallback Search Provider ---

// FallbackSearchProvider tries providers in order, returning the first successful non-empty result.
type FallbackSearchProvider struct {
	providers []SearchProvider
}

// NewFallbackSearchProvider creates a FallbackSearchProvider that tries each provider in order.
func NewFallbackSearchProvider(providers ...SearchProvider) *FallbackSearchProvider {
	return &FallbackSearchProvider{providers: providers}
}

func (f *FallbackSearchProvider) Name() string { return "fallback" }

func (f *FallbackSearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	var lastErr error
	for _, p := range f.providers {
		results, err := p.Search(ctx, query, count)
		if err != nil {
			logger.WarnCF("tool", "search provider failed, trying next",
				map[string]interface{}{"provider": p.Name(), "error": err.Error()})
			lastErr = err
			continue
		}
		if len(results) == 0 {
			logger.WarnCF("tool", "search provider returned no results, trying next",
				map[string]interface{}{"provider": p.Name()})
			continue
		}
		return results, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("all search providers failed; last error: %w", lastErr)
	}
	return []SearchResult{}, nil
}

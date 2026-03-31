package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewWebSearchTool(t *testing.T) {
	tests := []struct {
		name           string
		provider       SearchProvider
		maxResults     int
		wantMaxResults int
		wantProvider   bool
	}{
		{
			name:           "valid maxResults with provider",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     3,
			wantMaxResults: 3,
			wantProvider:   true,
		},
		{
			name:           "zero maxResults defaults to 5",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     0,
			wantMaxResults: 5,
			wantProvider:   true,
		},
		{
			name:           "negative maxResults defaults to 5",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     -1,
			wantMaxResults: 5,
			wantProvider:   true,
		},
		{
			name:           "maxResults above 10 defaults to 5",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     11,
			wantMaxResults: 5,
			wantProvider:   true,
		},
		{
			name:           "maxResults exactly 10 is valid",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     10,
			wantMaxResults: 10,
			wantProvider:   true,
		},
		{
			name:           "maxResults exactly 1 is valid",
			provider:       NewBraveSearchProvider("test-key"),
			maxResults:     1,
			wantMaxResults: 1,
			wantProvider:   true,
		},
		{
			name:           "nil provider is allowed at construction",
			provider:       nil,
			maxResults:     5,
			wantMaxResults: 5,
			wantProvider:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWebSearchTool(tt.provider, tt.maxResults)
			if tool == nil {
				t.Fatal("expected non-nil tool")
			}
			if tool.maxResults != tt.wantMaxResults {
				t.Errorf("maxResults = %d, want %d", tool.maxResults, tt.wantMaxResults)
			}
			if tt.wantProvider && tool.provider == nil {
				t.Error("expected non-nil provider")
			}
			if !tt.wantProvider && tool.provider != nil {
				t.Error("expected nil provider")
			}
		})
	}
}

func TestWebSearchToolName(t *testing.T) {
	tool := NewWebSearchTool(nil, 5)
	if got := tool.Name(); got != "web_search" {
		t.Errorf("Name() = %q, want %q", got, "web_search")
	}
}

func TestWebSearchToolExecuteNoProvider(t *testing.T) {
	tool := NewWebSearchTool(nil, 5)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"query": "test query",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "No search provider configured") {
		t.Errorf("expected error about no provider configured, got %q", result)
	}
}

func TestWebSearchToolExecuteMissingQuery(t *testing.T) {
	tool := NewWebSearchTool(NewBraveSearchProvider("test-key"), 5)
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Errorf("expected 'query is required' error, got %v", err)
	}
}

func TestWebSearchToolExecuteWrongQueryType(t *testing.T) {
	tool := NewWebSearchTool(NewBraveSearchProvider("test-key"), 5)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"query": 12345,
	})
	if err == nil {
		t.Fatal("expected error for non-string query")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Errorf("expected 'query is required' error, got %v", err)
	}
}

func TestNewWebFetchTool(t *testing.T) {
	tests := []struct {
		name         string
		maxChars     int
		wantMaxChars int
	}{
		{
			name:         "valid maxChars",
			maxChars:     10000,
			wantMaxChars: 10000,
		},
		{
			name:         "zero maxChars defaults to 50000",
			maxChars:     0,
			wantMaxChars: 50000,
		},
		{
			name:         "negative maxChars defaults to 50000",
			maxChars:     -1,
			wantMaxChars: 50000,
		},
		{
			name:         "maxChars 1 is valid",
			maxChars:     1,
			wantMaxChars: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWebFetchTool(tt.maxChars)
			if tool == nil {
				t.Fatal("expected non-nil tool")
			}
			if tool.maxChars != tt.wantMaxChars {
				t.Errorf("maxChars = %d, want %d", tool.maxChars, tt.wantMaxChars)
			}
		})
	}
}

func TestWebFetchToolName(t *testing.T) {
	tool := NewWebFetchTool(50000)
	if got := tool.Name(); got != "web_fetch" {
		t.Errorf("Name() = %q, want %q", got, "web_fetch")
	}
}

func TestWebFetchToolExecuteMissingURL(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if !strings.Contains(err.Error(), "url is required") {
		t.Errorf("expected 'url is required' error, got %v", err)
	}
}

func TestWebFetchToolExecuteWrongURLType(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"url": 12345,
	})
	if err == nil {
		t.Fatal("expected error for non-string URL")
	}
	if !strings.Contains(err.Error(), "url is required") {
		t.Errorf("expected 'url is required' error, got %v", err)
	}
}

func TestWebFetchToolExecuteNonHTTPURL(t *testing.T) {
	tool := NewWebFetchTool(50000)

	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{
			name:    "ftp scheme",
			url:     "ftp://example.com/file",
			wantErr: "only http/https URLs are allowed",
		},
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: "only http/https URLs are allowed",
		},
		{
			name:    "javascript scheme",
			url:     "javascript:alert(1)",
			wantErr: "only http/https URLs are allowed",
		},
		{
			name:    "data scheme",
			url:     "data:text/html,<h1>test</h1>",
			wantErr: "only http/https URLs are allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(context.Background(), map[string]interface{}{
				"url": tt.url,
			})
			if err == nil {
				t.Fatal("expected error for non-HTTP URL")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestWebFetchToolExecuteMissingHost(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"url": "http://",
	})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
	if !strings.Contains(err.Error(), "missing domain") {
		t.Errorf("expected 'missing domain' error, got %v", err)
	}
}

func TestWebFetchToolExtractText(t *testing.T) {
	tool := NewWebFetchTool(50000)

	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "removes script tags and content",
			input:    "<p>Hello</p><script>alert('xss')</script><p>World</p>",
			contains: []string{"Hello", "World"},
			excludes: []string{"<script>", "alert", "</script>"},
		},
		{
			name:     "removes style tags and content",
			input:    "<p>Hello</p><style>.red { color: red; }</style><p>World</p>",
			contains: []string{"Hello", "World"},
			excludes: []string{"<style>", "color", "</style>"},
		},
		{
			name:     "removes HTML tags",
			input:    "<div class='main'><h1>Title</h1><p>Paragraph</p></div>",
			contains: []string{"Title", "Paragraph"},
			excludes: []string{"<div", "<h1>", "</h1>", "<p>", "</p>", "</div>"},
		},
		{
			name:     "collapses whitespace",
			input:    "<p>Hello    World</p>",
			contains: []string{"Hello", "World"},
			excludes: []string{},
		},
		{
			name:     "empty input",
			input:    "",
			contains: []string{},
			excludes: []string{},
		},
		{
			name:     "plain text passthrough",
			input:    "Just plain text content",
			contains: []string{"Just plain text content"},
			excludes: []string{},
		},
		{
			name:     "nested script tags",
			input:    "<script type='text/javascript'>var x = '<script>nested</script>';</script><p>Safe</p>",
			contains: []string{"Safe"},
			excludes: []string{"var x"},
		},
		{
			name:     "multiple scripts and styles",
			input:    "<script>a()</script><style>b{}</style><p>Content</p><script>c()</script>",
			contains: []string{"Content"},
			excludes: []string{"a()", "b{}", "c()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := tool.extractText(tt.input, "https://example.com")
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got %q", s, result)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(result, s) {
					t.Errorf("expected result NOT to contain %q, got %q", s, result)
				}
			}
		})
	}
}

func TestWebFetchToolExecuteValidHTTPSchemes(t *testing.T) {
	tool := NewWebFetchTool(50000)

	// http and https should not fail on scheme validation.
	// They will fail on the actual request (no server), but
	// we verify the scheme check passes by checking the error
	// message does NOT mention scheme restrictions.
	tests := []struct {
		name string
		url  string
	}{
		{"http scheme", "http://example.invalid.test"},
		{"https scheme", "https://example.invalid.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(context.Background(), map[string]interface{}{
				"url": tt.url,
			})
			// We expect an error from the network request, not from URL validation
			if err != nil && strings.Contains(err.Error(), "only http/https URLs are allowed") {
				t.Errorf("valid scheme %q should not be rejected", tt.url)
			}
			if err != nil && strings.Contains(err.Error(), "missing domain") {
				t.Errorf("valid URL %q should not fail domain check", tt.url)
			}
		})
	}
}

func TestWebSearchToolParameters(t *testing.T) {
	tool := NewWebSearchTool(nil, 5)
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	if _, ok := props["query"]; !ok {
		t.Error("expected 'query' in parameters properties")
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("expected required to be a string slice")
	}

	found := false
	for _, r := range required {
		if r == "query" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'query' to be in required list")
	}
}

func TestWebFetchToolParameters(t *testing.T) {
	tool := NewWebFetchTool(50000)
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	if _, ok := props["url"]; !ok {
		t.Error("expected 'url' in parameters properties")
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("expected required to be a string slice")
	}

	found := false
	for _, r := range required {
		if r == "url" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'url' to be in required list")
	}
}

func TestSearXNGSearchProvider(t *testing.T) {
	// Mock SearXNG server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		q := r.URL.Query().Get("q")
		if q == "" {
			t.Error("expected query parameter 'q'")
		}

		format := r.URL.Query().Get("format")
		if format != "json" {
			t.Errorf("expected format=json, got %q", format)
		}

		resp := map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"title":   "SearXNG Result 1",
					"url":     "https://example.com/1",
					"content": "Description of result 1",
				},
				{
					"title":   "SearXNG Result 2",
					"url":     "https://example.com/2",
					"content": "Description of result 2",
				},
				{
					"title":   "SearXNG Result 3",
					"url":     "https://example.com/3",
					"content": "Description of result 3",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewSearXNGSearchProvider(server.URL)

	if provider.Name() != "searxng" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "searxng")
	}

	results, err := provider.Search(context.Background(), "test query", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Title != "SearXNG Result 1" {
		t.Errorf("results[0].Title = %q, want %q", results[0].Title, "SearXNG Result 1")
	}
	if results[0].URL != "https://example.com/1" {
		t.Errorf("results[0].URL = %q, want %q", results[0].URL, "https://example.com/1")
	}
	if results[0].Description != "Description of result 1" {
		t.Errorf("results[0].Description = %q, want %q", results[0].Description, "Description of result 1")
	}

	if results[1].Title != "SearXNG Result 2" {
		t.Errorf("results[1].Title = %q, want %q", results[1].Title, "SearXNG Result 2")
	}
}

func TestBraveSearchProvider(t *testing.T) {
	// Mock Brave Search API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Subscription-Token")
		if token != "test-brave-key" {
			t.Errorf("expected X-Subscription-Token = %q, got %q", "test-brave-key", token)
		}

		resp := map[string]interface{}{
			"web": map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"title":       "Brave Result 1",
						"url":         "https://brave.com/1",
						"description": "Brave description 1",
					},
					{
						"title":       "Brave Result 2",
						"url":         "https://brave.com/2",
						"description": "Brave description 2",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the search URL by creating a provider and pointing it at the mock.
	// Since BraveSearchProvider hardcodes the URL, we test via the full tool flow
	// using a custom provider that hits our mock server.
	provider := &BraveSearchProvider{apiKey: "test-brave-key"}

	if provider.Name() != "brave" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "brave")
	}

	// We can't easily redirect the hardcoded Brave URL, so we test the mock
	// by temporarily swapping the HTTP client to route to our test server.
	// NOTE: not safe for t.Parallel() due to global client swap.
	originalClient := webSearchHTTPClient
	webSearchHTTPClient = server.Client()
	t.Cleanup(func() { webSearchHTTPClient = originalClient })

	// The Brave provider uses the hardcoded URL, so this will fail with a
	// connection error to api.search.brave.com. Instead, let's test the
	// SearXNG provider more thoroughly and just verify Brave provider
	// construction and name.
	// For a true integration test of Brave parsing, we'd need to mock at
	// the transport level. Here we verify the struct and interface compliance.

	var _ SearchProvider = provider // compile-time interface check
}

func TestBraveSearchProvider_WithBaseURL(t *testing.T) {
	// Mock server returning valid Brave-shaped response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Subscription-Token") == "" {
			t.Error("expected X-Subscription-Token header")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"web":{"results":[{"title":"Test Title","url":"https://example.com","description":"Test desc"}]}}`)
	}))
	defer server.Close()

	provider := NewBraveSearchProviderWithBaseURL("test-api-key", server.URL)
	results, err := provider.Search(context.Background(), "test query", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Test Title" {
		t.Errorf("Title = %q, want %q", results[0].Title, "Test Title")
	}
	if results[0].URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", results[0].URL, "https://example.com")
	}
}

func TestWebFetchToolExtractTextReadability(t *testing.T) {
	tool := NewWebFetchTool(50000)

	html := `<!DOCTYPE html>
<html>
<head><title>Test Article</title></head>
<body>
<nav>Navigation links here</nav>
<article>
<h1>Main Article Title</h1>
<p>This is the first paragraph of a meaningful article that has enough content
for the readability algorithm to detect it as the main content of the page.
It needs to be substantial enough to pass the content length threshold.</p>
<p>This is the second paragraph with more content. The readability library
works by analyzing the DOM structure and finding the most likely content
block. It considers paragraph density, text length, and various heuristics
to determine what is the main article content versus navigation, sidebars,
and other non-content elements.</p>
<p>A third paragraph ensures we have enough text density for readability
to confidently identify this as the main content area of the page. Without
sufficient content, the library might fall back or return empty results.</p>
</article>
<footer>Footer content here</footer>
</body>
</html>`

	result, method := tool.extractText(html, "https://example.com/article")

	if !strings.Contains(result, "Main Article Title") {
		t.Errorf("expected result to contain article title, got %q", result)
	}
	if !strings.Contains(result, "first paragraph") {
		t.Errorf("expected result to contain article content, got %q", result)
	}

	// The method should be either "readability" or "regex-fallback"
	if method != "readability" && method != "regex-fallback" {
		t.Errorf("unexpected extraction method: %q", method)
	}
}

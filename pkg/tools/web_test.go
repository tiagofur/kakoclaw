package tools

import (
	"context"
	"strings"
	"testing"
)

func TestNewWebSearchTool(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		maxResults     int
		wantMaxResults int
	}{
		{
			name:           "valid maxResults",
			apiKey:         "test-key",
			maxResults:     3,
			wantMaxResults: 3,
		},
		{
			name:           "zero maxResults defaults to 5",
			apiKey:         "test-key",
			maxResults:     0,
			wantMaxResults: 5,
		},
		{
			name:           "negative maxResults defaults to 5",
			apiKey:         "test-key",
			maxResults:     -1,
			wantMaxResults: 5,
		},
		{
			name:           "maxResults above 10 defaults to 5",
			apiKey:         "test-key",
			maxResults:     11,
			wantMaxResults: 5,
		},
		{
			name:           "maxResults exactly 10 is valid",
			apiKey:         "test-key",
			maxResults:     10,
			wantMaxResults: 10,
		},
		{
			name:           "maxResults exactly 1 is valid",
			apiKey:         "test-key",
			maxResults:     1,
			wantMaxResults: 1,
		},
		{
			name:           "empty api key is allowed at construction",
			apiKey:         "",
			maxResults:     5,
			wantMaxResults: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWebSearchTool(tt.apiKey, tt.maxResults)
			if tool == nil {
				t.Fatal("expected non-nil tool")
			}
			if tool.maxResults != tt.wantMaxResults {
				t.Errorf("maxResults = %d, want %d", tool.maxResults, tt.wantMaxResults)
			}
			if tool.apiKey != tt.apiKey {
				t.Errorf("apiKey = %q, want %q", tool.apiKey, tt.apiKey)
			}
		})
	}
}

func TestWebSearchToolName(t *testing.T) {
	tool := NewWebSearchTool("key", 5)
	if got := tool.Name(); got != "web_search" {
		t.Errorf("Name() = %q, want %q", got, "web_search")
	}
}

func TestWebSearchToolExecuteNoAPIKey(t *testing.T) {
	tool := NewWebSearchTool("", 5)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"query": "test query",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "BRAVE_API_KEY") {
		t.Errorf("expected error about BRAVE_API_KEY, got %q", result)
	}
}

func TestWebSearchToolExecuteMissingQuery(t *testing.T) {
	tool := NewWebSearchTool("test-key", 5)
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Errorf("expected 'query is required' error, got %v", err)
	}
}

func TestWebSearchToolExecuteWrongQueryType(t *testing.T) {
	tool := NewWebSearchTool("test-key", 5)
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
			contains: []string{"Hello World"},
			excludes: []string{"Hello    World"},
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
			result := tool.extractText(tt.input)
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
	tool := NewWebSearchTool("key", 5)
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

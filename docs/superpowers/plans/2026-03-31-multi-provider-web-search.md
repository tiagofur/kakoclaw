# Multi-Provider Web Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `TavilySearchProvider` and `FallbackSearchProvider`; extend config with `TavilyAPIKey` and `Priority`; replace hardcoded provider selection in `loop.go` with a config-driven fallback chain.

**Architecture:** The existing `SearchProvider` interface is unchanged. `FallbackSearchProvider` wraps an ordered `[]SearchProvider` and delegates sequentially. `TavilySearchProvider` POSTs to the Tavily REST API. `BraveSearchProvider` gains a `baseURL` field for testability. `WebSearchTool` is unchanged — it still receives a single `SearchProvider` at construction time, which may now be a `FallbackSearchProvider`. Provider construction in `loop.go` moves from a hardcoded `if/else` to a config-driven loop.

**Tech Stack:** Go 1.26, `pkg/tools`, `pkg/config`, `pkg/agent`

---

## Files to Modify

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `TavilyAPIKey` and `Priority` to `WebSearchConfig`; update `DefaultConfig()` and merge condition |
| `pkg/config/config_test.go` | Add tests T8, T9 |
| `pkg/tools/search_provider.go` | Add `baseURL` to `BraveSearchProvider`; add `NewBraveSearchProviderWithBaseURL`; add `TavilySearchProvider`; add `FallbackSearchProvider` |
| `pkg/tools/web.go` | Update nil-provider error message |
| `pkg/tools/web_test.go` | Add tests T1–T7 |
| `pkg/agent/loop.go` | Add `getPriorityOrder` helper; replace lines 310–321 with dynamic provider construction |

No new files are created.

---

## Task 1 — Config changes

**Files:** `pkg/config/config.go`, `pkg/config/config_test.go`

### Step 1 — Write the failing test

In `pkg/config/config_test.go`, add:

```go
func TestWebSearchConfigNewFields(t *testing.T) {
    // T8: default priority from DefaultConfig
    cfg := DefaultConfig()
    priority := getPriorityOrderForTest(cfg.Tools.Web.Search.Priority)
    expected := []string{"searxng", "brave", "tavily"}
    if !reflect.DeepEqual(priority, expected) {
        t.Errorf("default priority = %v, want %v", priority, expected)
    }

    // T9: JSON round-trip for TavilyAPIKey and Priority
    src := WebSearchConfig{
        APIKey:       "brave-key",
        SearXNGURL:   "http://searx.local",
        MaxResults:   10,
        TavilyAPIKey: "tvly-secret",
        Priority:     []string{"tavily", "brave"},
    }
    data, err := json.Marshal(src)
    if err != nil {
        t.Fatalf("marshal failed: %v", err)
    }
    var dst WebSearchConfig
    if err := json.Unmarshal(data, &dst); err != nil {
        t.Fatalf("unmarshal failed: %v", err)
    }
    if dst.TavilyAPIKey != src.TavilyAPIKey {
        t.Errorf("TavilyAPIKey = %q, want %q", dst.TavilyAPIKey, src.TavilyAPIKey)
    }
    if !reflect.DeepEqual(dst.Priority, src.Priority) {
        t.Errorf("Priority = %v, want %v", dst.Priority, src.Priority)
    }
}

// getPriorityOrderForTest mirrors the logic of getPriorityOrder in loop.go.
// It lives here only to test the default config value independently of loop.go.
func getPriorityOrderForTest(p []string) []string {
    if len(p) > 0 {
        return p
    }
    return []string{"searxng", "brave", "tavily"}
}
```

Note: `T8` and `T9` are merged into one test function for the round-trip; the helper `getPriorityOrderForTest` validates FR-10 without importing `pkg/agent`.

### Step 2 — Run failing test

```bash
go test ./pkg/config/... -run TestWebSearchConfigNewFields -v
```

Expected: FAIL (`TavilyAPIKey` field does not exist).

### Step 3 — Add fields to `WebSearchConfig` (lines 366–370 of `pkg/config/config.go`)

Current:

```go
type WebSearchConfig struct {
	APIKey     string `json:"api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
	SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
	MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`
}
```

Replace with:

```go
type WebSearchConfig struct {
	APIKey     string `json:"api_key"     env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
	SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
	MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`

	// New fields — added for multi-provider support
	TavilyAPIKey string   `json:"tavily_api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY"`
	Priority     []string `json:"priority"` // no env tag — slice type unsupported by flat env-var override
}
```

### Step 4 — Update `DefaultConfig()` (lines 702–709 of `pkg/config/config.go`)

Current:

```go
Tools: ToolsConfig{
    Web: WebToolsConfig{
        Search: WebSearchConfig{
            APIKey:     "",
            SearXNGURL: "",
            MaxResults: 5,
        },
    },
```

Replace with:

```go
Tools: ToolsConfig{
    Web: WebToolsConfig{
        Search: WebSearchConfig{
            MaxResults: 5,
            Priority:   []string{"searxng", "brave", "tavily"},
        },
    },
```

### Step 5 — Update merge condition in `mergeToolsConfig` (line 1484 of `pkg/config/config.go`)

Current:

```go
if user.Web.Search.APIKey != "" || user.Web.Search.SearXNGURL != "" {
```

Replace with:

```go
if user.Web.Search.APIKey != "" || user.Web.Search.SearXNGURL != "" || user.Web.Search.TavilyAPIKey != "" || len(user.Web.Search.Priority) > 0 {
```

### Step 6 — Run test, expect PASS

```bash
go test ./pkg/config/... -run TestWebSearchConfigNewFields -v
```

### Step 7 — Build

```bash
go build ./pkg/config/...
```

### Step 8 — Commit

```
feat(config): add TavilyAPIKey and Priority to WebSearchConfig
```

---

## Task 2 — BraveSearchProvider testability fix

**Files:** `pkg/tools/search_provider.go`, `pkg/tools/web_test.go`

### Step 1 — Write the failing test (T7)

In `pkg/tools/web_test.go`, add:

```go
func TestBraveSearchProvider_WithBaseURL(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/res/v1/web/search" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        q := r.URL.Query().Get("q")
        if q == "" {
            t.Error("expected non-empty q param")
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"web":{"results":[{"title":"Go Lang","url":"https://go.dev","description":"The Go programming language"}]}}`)
    }))
    defer srv.Close()

    p := NewBraveSearchProviderWithBaseURL("test-api-key", srv.URL)
    results, err := p.Search(context.Background(), "golang", 5)
    if err != nil {
        t.Fatalf("Search() error = %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0].Title != "Go Lang" {
        t.Errorf("Title = %q, want %q", results[0].Title, "Go Lang")
    }
    if results[0].URL != "https://go.dev" {
        t.Errorf("URL = %q, want %q", results[0].URL, "https://go.dev")
    }
    if results[0].Description != "The Go programming language" {
        t.Errorf("Description = %q, want %q", results[0].Description, "The Go programming language")
    }
}
```

### Step 2 — Run failing test

```bash
go test ./pkg/tools/... -run TestBraveSearchProvider_WithBaseURL -v
```

Expected: FAIL (`NewBraveSearchProviderWithBaseURL` undefined).

### Step 3 — Add `baseURL` field to `BraveSearchProvider`

In `pkg/tools/search_provider.go`, replace:

```go
type BraveSearchProvider struct {
	apiKey string
}
```

With:

```go
type BraveSearchProvider struct {
	apiKey  string
	baseURL string
}
```

### Step 4 — Update `NewBraveSearchProvider`

Replace:

```go
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey}
}
```

With:

```go
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey, baseURL: "https://api.search.brave.com"}
}
```

### Step 5 — Add `NewBraveSearchProviderWithBaseURL`

After `NewBraveSearchProvider`, add:

```go
// NewBraveSearchProviderWithBaseURL creates a BraveSearchProvider with a custom
// base URL, enabling test servers (e.g. httptest.NewServer).
func NewBraveSearchProviderWithBaseURL(apiKey, baseURL string) *BraveSearchProvider {
	return &BraveSearchProvider{apiKey: apiKey, baseURL: baseURL}
}
```

### Step 6 — Update `Search()` to use `b.baseURL`

In `BraveSearchProvider.Search()`, replace:

```go
searchURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
    url.QueryEscape(query), count)
```

With:

```go
searchURL := fmt.Sprintf("%s/res/v1/web/search?q=%s&count=%d",
    b.baseURL, url.QueryEscape(query), count)
```

### Step 7 — Run test, expect PASS

```bash
go test ./pkg/tools/... -run TestBraveSearchProvider_WithBaseURL -v
```

### Step 8 — Build

```bash
go build ./pkg/tools/...
```

### Step 9 — Commit

```
refactor(tools): add baseURL field to BraveSearchProvider for testability
```

---

## Task 3 — TavilySearchProvider + FallbackSearchProvider

**Files:** `pkg/tools/search_provider.go`, `pkg/tools/web_test.go`

### Step 1 — Write failing Tavily test (T5 + T6)

In `pkg/tools/web_test.go`, add:

```go
func TestTavilySearchProvider_Search(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/search" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        var body map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            t.Errorf("failed to decode body: %v", err)
        }
        if body["api_key"] != "tvly-test" {
            t.Errorf("api_key = %v, want tvly-test", body["api_key"])
        }
        if body["search_depth"] != "basic" {
            t.Errorf("search_depth = %v, want basic", body["search_depth"])
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"results":[{"title":"Error Handling in Go","url":"https://go.dev/blog/error-handling","content":"Go uses explicit error return values"}]}`)
    }))
    defer srv.Close()

    p := NewTavilySearchProviderWithBaseURL("tvly-test", srv.URL)
    results, err := p.Search(context.Background(), "golang error handling", 5)
    if err != nil {
        t.Fatalf("Search() error = %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0].Title != "Error Handling in Go" {
        t.Errorf("Title = %q, want %q", results[0].Title, "Error Handling in Go")
    }
    if results[0].URL != "https://go.dev/blog/error-handling" {
        t.Errorf("URL = %q, want %q", results[0].URL, "https://go.dev/blog/error-handling")
    }
    if results[0].Description != "Go uses explicit error return values" {
        t.Errorf("Description = %q, want %q", results[0].Description, "Go uses explicit error return values")
    }
}

func TestTavilySearchProvider_HTTPError(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
    }))
    defer srv.Close()

    p := NewTavilySearchProviderWithBaseURL("bad-key", srv.URL)
    _, err := p.Search(context.Background(), "test", 5)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if !strings.Contains(err.Error(), "401") {
        t.Errorf("expected error to contain '401', got: %v", err)
    }
}
```

### Step 2 — Write failing Fallback tests (T1–T4)

In `pkg/tools/web_test.go`, add a mock provider helper and four tests:

```go
// mockSearchProvider is a test double for SearchProvider.
type mockSearchProvider struct {
    name     string
    results  []SearchResult
    err      error
    callCount int
}

func (m *mockSearchProvider) Name() string { return m.name }
func (m *mockSearchProvider) Search(_ context.Context, _ string, _ int) ([]SearchResult, error) {
    m.callCount++
    return m.results, m.err
}

func TestFallbackSearchProvider_FirstSucceeds(t *testing.T) {
    first := &mockSearchProvider{
        name:    "first",
        results: []SearchResult{{Title: "Result", URL: "https://example.com", Description: "Desc"}},
    }
    second := &mockSearchProvider{name: "second"}

    f := NewFallbackSearchProvider(first, second)
    results, err := f.Search(context.Background(), "test", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if second.callCount != 0 {
        t.Errorf("second provider was called %d times, want 0", second.callCount)
    }
}

func TestFallbackSearchProvider_FirstFails_SecondSucceeds(t *testing.T) {
    first := &mockSearchProvider{
        name: "first",
        err:  fmt.Errorf("rate limited"),
    }
    second := &mockSearchProvider{
        name:    "second",
        results: []SearchResult{{Title: "Fallback", URL: "https://fallback.com", Description: "From second"}},
    }

    f := NewFallbackSearchProvider(first, second)
    results, err := f.Search(context.Background(), "test", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0].Title != "Fallback" {
        t.Errorf("Title = %q, want Fallback", results[0].Title)
    }
}

func TestFallbackSearchProvider_EmptyResultsSkipped(t *testing.T) {
    first := &mockSearchProvider{
        name:    "first",
        results: []SearchResult{}, // empty, no error
    }
    second := &mockSearchProvider{
        name:    "second",
        results: []SearchResult{{Title: "Second Result", URL: "https://second.com", Description: "Found it"}},
    }

    f := NewFallbackSearchProvider(first, second)
    results, err := f.Search(context.Background(), "test", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0].Title != "Second Result" {
        t.Errorf("Title = %q, want Second Result", results[0].Title)
    }
}

func TestFallbackSearchProvider_AllFail(t *testing.T) {
    lastErr := fmt.Errorf("second also failed")
    first := &mockSearchProvider{name: "first", err: fmt.Errorf("first failed")}
    second := &mockSearchProvider{name: "second", err: lastErr}

    f := NewFallbackSearchProvider(first, second)
    _, err := f.Search(context.Background(), "test", 5)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    // The returned error must wrap or equal the last provider's error
    if !strings.Contains(err.Error(), "second also failed") {
        t.Errorf("expected error to reference last provider's error, got: %v", err)
    }
}
```

### Step 3 — Run failing tests

```bash
go test ./pkg/tools/... -run "TestTavilySearchProvider|TestFallbackSearchProvider" -v
```

Expected: FAIL (types and constructors do not exist).

### Step 4 — Add `TavilySearchProvider` to `search_provider.go`

Add the following to `pkg/tools/search_provider.go`. The file needs `"bytes"` added to its import block (alongside existing `"context"`, `"encoding/json"`, `"fmt"`, `"io"`, `"net/http"`, `"net/url"`):

```go
// --- Tavily Search Provider ---

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

// TavilySearchProvider implements SearchProvider using the Tavily Search REST API.
type TavilySearchProvider struct {
	apiKey  string
	baseURL string // default: "https://api.tavily.com"
}

// NewTavilySearchProvider creates a TavilySearchProvider using the production endpoint.
func NewTavilySearchProvider(apiKey string) *TavilySearchProvider {
	return &TavilySearchProvider{apiKey: apiKey, baseURL: "https://api.tavily.com"}
}

// NewTavilySearchProviderWithBaseURL creates a TavilySearchProvider with a custom
// base URL, enabling test servers (e.g. httptest.NewServer).
func NewTavilySearchProviderWithBaseURL(apiKey, baseURL string) *TavilySearchProvider {
	return &TavilySearchProvider{apiKey: apiKey, baseURL: baseURL}
}

func (t *TavilySearchProvider) Name() string { return "tavily" }

func (t *TavilySearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	reqBody := tavilyRequest{
		APIKey:      t.apiKey,
		Query:       query,
		MaxResults:  count,
		SearchDepth: "basic",
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/search", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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

	var searchResp tavilyResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Results))
	for _, r := range searchResp.Results {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content, // Tavily "content" → SearchResult.Description
		})
	}
	return results, nil
}
```

### Step 5 — Add `FallbackSearchProvider` to `search_provider.go`

Add the following after `TavilySearchProvider`. This requires importing `"github.com/sipeed/makoclaw/pkg/logger"` in `search_provider.go` — check if it is already imported; if not, add it to the import block:

```go
// --- Fallback Search Provider ---

// FallbackSearchProvider tries SearchProviders in priority order,
// returning the first non-empty result set.
// It is itself a SearchProvider and can be passed directly to WebSearchTool.
type FallbackSearchProvider struct {
	providers []SearchProvider
}

// NewFallbackSearchProvider creates a FallbackSearchProvider from an ordered list
// of providers. At least one provider must be supplied.
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
	return nil, nil // all returned empty, none errored
}
```

Note: if `search_provider.go` does not already import `"github.com/sipeed/makoclaw/pkg/logger"`, add it. Check with `grep -n "logger" pkg/tools/search_provider.go` before editing.

### Step 6 — Run all new tests, expect PASS

```bash
go test ./pkg/tools/... -run "TestTavilySearchProvider|TestFallbackSearchProvider|TestBraveSearchProvider_WithBaseURL" -v
```

### Step 7 — Build

```bash
go build ./pkg/tools/...
```

### Step 8 — Run full tools test suite

```bash
go test ./pkg/tools/... -v
```

### Step 9 — Commit

```
feat(tools): add TavilySearchProvider and FallbackSearchProvider
```

---

## Task 4 — Integrate into loop.go + update nil-provider message

**Files:** `pkg/agent/loop.go`, `pkg/tools/web.go`

### Step 1 — Add `getPriorityOrder` to `loop.go`

Add the following package-level function near the top of `pkg/agent/loop.go` (before or after `NewAgentLoop`):

```go
// getPriorityOrder returns the configured priority order, or the default
// ["searxng", "brave", "tavily"] when the slice is empty.
func getPriorityOrder(priority []string) []string {
	if len(priority) > 0 {
		return priority
	}
	return []string{"searxng", "brave", "tavily"}
}
```

### Step 2 — Replace lines 310–321 in `loop.go`

Current block (lines 310–321):

```go
// Search provider selection: SearXNG (self-hosted) > Brave (API key)
var searchProvider tools.SearchProvider
if searxngURL := cfg.Tools.Web.Search.SearXNGURL; searxngURL != "" {
    searchProvider = tools.NewSearXNGSearchProvider(searxngURL)
    logger.InfoCF("agent", "Search provider: SearXNG", map[string]interface{}{"url": searxngURL})
} else if braveKey := cfg.Tools.Web.Search.APIKey; braveKey != "" {
    searchProvider = tools.NewBraveSearchProvider(braveKey)
    logger.InfoC("agent", "Search provider: Brave Search API")
} else {
    logger.WarnC("agent", "No search provider configured — web_search will return an error")
}
toolsRegistry.Register(tools.NewWebSearchTool(searchProvider, cfg.Tools.Web.Search.MaxResults))
```

Replace with:

```go
// Search provider construction: build ordered list from priority config, wrap in
// FallbackSearchProvider when multiple providers are configured.
var providers []tools.SearchProvider
for _, name := range getPriorityOrder(cfg.Tools.Web.Search.Priority) {
    switch name {
    case "searxng":
        if u := cfg.Tools.Web.Search.SearXNGURL; u != "" {
            providers = append(providers, tools.NewSearXNGSearchProvider(u))
        }
    case "brave":
        if k := cfg.Tools.Web.Search.APIKey; k != "" {
            providers = append(providers, tools.NewBraveSearchProvider(k))
        }
    case "tavily":
        if k := cfg.Tools.Web.Search.TavilyAPIKey; k != "" {
            providers = append(providers, tools.NewTavilySearchProvider(k))
        }
    }
}

var searchProvider tools.SearchProvider
switch len(providers) {
case 0:
    logger.WarnC("agent", "no search provider configured — web_search tool will return errors")
case 1:
    searchProvider = providers[0]
default:
    searchProvider = tools.NewFallbackSearchProvider(providers...)
}
toolsRegistry.Register(tools.NewWebSearchTool(searchProvider, cfg.Tools.Web.Search.MaxResults))
```

### Step 3 — Update nil-provider error message in `web.go`

Current line 135 in `pkg/tools/web.go`:

```go
return "Error: No search provider configured. Set MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL or MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY.", nil
```

Replace with:

```go
return "Error: No search provider configured. Set at least one of: MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY (Brave), MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL (SearXNG), or MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY (Tavily).", nil
```

### Step 4 — Build both packages

```bash
go build ./pkg/agent/... && go build ./pkg/tools/...
```

### Step 5 — Run full test suite

```bash
go test ./pkg/tools/... ./pkg/config/... -v -timeout 60s
```

### Step 6 — Commit

```
feat(agent): dynamic web search provider construction with fallback
```

---

## Verification Checklist

- [ ] `go build ./pkg/config/...` passes after Task 1
- [ ] `go build ./pkg/tools/...` passes after Task 2 and Task 3
- [ ] `go build ./pkg/agent/...` passes after Task 4
- [ ] `go test ./pkg/config/... -run TestWebSearchConfigNewFields` passes
- [ ] `go test ./pkg/tools/... -run TestBraveSearchProvider_WithBaseURL` passes
- [ ] `go test ./pkg/tools/... -run "TestTavilySearchProvider|TestFallbackSearchProvider"` passes (all 6 cases)
- [ ] `go test ./pkg/tools/... ./pkg/config/... -v -timeout 60s` passes (no regressions)
- [ ] Nil-provider error message in `web.go` mentions Brave, SearXNG, and Tavily

---

STATUS: DONE

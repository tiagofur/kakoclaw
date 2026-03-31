# Design: Multi-Provider Web Search

**Change:** `multi-provider-web-search`
**Date:** 2026-03-31
**Status:** Draft
**Spec:** `docs/superpowers/specs/2026-03-31-multi-provider-web-search-design.md`

---

## 1. Architecture Overview

The existing `SearchProvider` interface (`Search`, `Name`) is preserved without modification. A new `FallbackSearchProvider` implements that same interface, wrapping an ordered `[]SearchProvider` slice and delegating to each in sequence until one returns a non-empty result. `WebSearchTool` is unchanged — it continues to receive a single `SearchProvider` at construction time, which may now be a `FallbackSearchProvider` when multiple providers are configured. No new abstractions beyond `FallbackSearchProvider` and `TavilySearchProvider` are introduced. All new types live in the existing `pkg/tools/search_provider.go` file; provider construction logic moves from a hardcoded `if/else` chain to a config-driven loop in `pkg/agent/loop.go`.

---

## 2. `FallbackSearchProvider` Algorithm

```go
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
    return nil, nil // all returned empty
}
```

**Behavior summary:**

- Error from a provider → log warning, record `lastErr`, advance to next.
- Empty (non-nil, zero-length) result from a provider → log warning, advance to next (no error recorded).
- First non-empty result → return immediately without calling subsequent providers.
- All failed with at least one error → return `fmt.Errorf("all search providers failed; last error: %w", lastErr)`.
- All returned empty, none errored → return `nil, nil`.

---

## 3. `TavilySearchProvider` Implementation

```go
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
            Description: r.Content, // Tavily "content" maps to SearchResult.Description
        })
    }
    return results, nil
}
```

**Notes:**
- Uses the package-level `webSearchHTTPClient` (defined in `pkg/tools/web.go`) — no new HTTP clients.
- `search_depth` is always hardcoded to `"basic"` — not configurable.
- Error format (`"search API returned status %d: %s"`) matches the pattern used by `BraveSearchProvider` and `SearXNGSearchProvider`.
- `bytes.NewReader` requires importing `"bytes"` alongside the existing imports in `search_provider.go`.

---

## 4. `BraveSearchProvider` Update

The only change is replacing the hardcoded URL string literal with the `baseURL` field.

**Updated struct:**

```go
// BraveSearchProvider implements SearchProvider using the Brave Search API.
type BraveSearchProvider struct {
    apiKey  string
    baseURL string
}
```

**Updated constructors:**

```go
// NewBraveSearchProvider creates a BraveSearchProvider pointed at the production
// Brave Search API (https://api.search.brave.com).
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
    return &BraveSearchProvider{apiKey: apiKey, baseURL: "https://api.search.brave.com"}
}

// NewBraveSearchProviderWithBaseURL creates a BraveSearchProvider with a custom
// base URL, enabling test servers.
func NewBraveSearchProviderWithBaseURL(apiKey, baseURL string) *BraveSearchProvider {
    return &BraveSearchProvider{apiKey: apiKey, baseURL: baseURL}
}
```

**Updated `Search` method (URL construction line only):**

Before:
```go
searchURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
    url.QueryEscape(query), count)
```

After:
```go
searchURL := fmt.Sprintf("%s/res/v1/web/search?q=%s&count=%d",
    b.baseURL, url.QueryEscape(query), count)
```

All other logic in `Search` is unchanged.

---

## 5. Config Changes

### Current `WebSearchConfig` struct (lines 366–370 of `pkg/config/config.go`):

```go
type WebSearchConfig struct {
    APIKey     string `json:"api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
    SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
    MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`
}
```

### Updated `WebSearchConfig` struct (two fields added):

```go
type WebSearchConfig struct {
    APIKey     string `json:"api_key"     env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
    SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
    MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`

    // New fields
    TavilyAPIKey string   `json:"tavily_api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY"`
    Priority     []string `json:"priority"`
}
```

`Priority` has no `env` tag — slice-typed fields are not supported by the existing flat env-var override mechanism.

### `DefaultConfig()` change (current block, lines ~704–708):

```go
// Before
Search: WebSearchConfig{
    APIKey:     "",
    SearXNGURL: "",
    MaxResults: 5,
},
```

```go
// After
Search: WebSearchConfig{
    MaxResults: 5,
    Priority:   []string{"searxng", "brave", "tavily"},
},
```

`APIKey`, `SearXNGURL`, and `TavilyAPIKey` remain zero-value (empty string) in defaults — users supply their own credentials.

### Config merge behavior (`mergeToolsConfig`, line ~1484):

The existing merge condition:

```go
if user.Web.Search.APIKey != "" || user.Web.Search.SearXNGURL != "" {
    merged.Web.Search = user.Web.Search
} else {
    merged.Web.Search = global.Web.Search
}
```

Must be updated to also consider `TavilyAPIKey` (per-field merge so user's Tavily key merges independently from global Brave/SearXNG config):

```go
if user.Web.Search.APIKey != "" || user.Web.Search.SearXNGURL != "" || user.Web.Search.TavilyAPIKey != "" {
    merged.Web.Search = user.Web.Search
} else {
    merged.Web.Search = global.Web.Search
}
```

---

## 6. `loop.go` Construction (Exact Code)

### Current lines 310–321 (to be replaced):

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

### Replacement:

```go
// getPriorityOrder returns the configured priority order, or the default
// ["searxng", "brave", "tavily"] when cfg is empty.
func getPriorityOrder(priority []string) []string {
    if len(priority) > 0 {
        return priority
    }
    return []string{"searxng", "brave", "tavily"}
}
```

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

`getPriorityOrder` is a package-level function placed at the top of the file or adjacent to `NewAgentLoop`. It is not a method.

Providers with no config credential or URL are silently skipped during the loop — no log, no error. The warning for zero providers is emitted only after the loop completes.

---

## 7. Decision Log

| Decision | Choice | Why |
|----------|--------|-----|
| Where to place `FallbackSearchProvider` | `search_provider.go` | Same file as other providers; no new files per NFR-5 |
| Where to place `TavilySearchProvider` | `search_provider.go` | Same file as other providers; no new files per NFR-5 |
| DuckDuckGo | Out of scope | No viable structured-results API |
| Tavily `search_depth` | Hardcode `"basic"` | `"advanced"` adds cost; no user scenario requires it; keeps config minimal |
| Provider attribution in output | None (log only) | Changing `WebSearchTool.Execute` output format would break existing skills that parse results |
| Config merge strategy | Per-field (check `TavilyAPIKey`) | User's Tavily key merges independently from global Brave/SearXNG config; consistent with existing merge pattern |
| `getPriorityOrder` placement | Package-level function in `loop.go` | Keeps it close to its only call site; testable without the full agent loop |
| Skip condition logging | Silent skip | Providers with no config are expected absent; logging would be noisy for single-provider setups |
| 0/1/N provider handling | Explicit `switch len(providers)` | Avoids wrapping a single provider in `FallbackSearchProvider`, keeping the common case fast and observable |

---

## 8. Non-Goals

- No new files. All changes are additions or modifications to existing files.
- No changes to the `SearchResult` struct (`Title`, `URL`, `Description` fields are untouched).
- No changes to `WebSearchTool.Execute` formatting or result-building logic (beyond updating the nil-provider error message to name all three providers per FR-14).
- No caching. `FallbackSearchProvider` is stateless; each `Search` call is fully independent.
- No DuckDuckGo. No viable structured-results API exists for it.

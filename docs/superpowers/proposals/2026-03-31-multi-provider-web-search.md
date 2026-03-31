# Proposal: Multi-Provider Web Search

**Change:** `multi-provider-web-search`
**Date:** 2026-03-31
**Status:** Proposed

---

## 1. Problem Statement

Today, `WebSearchTool` is backed by a hardcoded selection in `pkg/agent/loop.go` (lines 310–321): if a SearXNG URL is configured it wins; otherwise Brave is used if its API key is present; otherwise the tool is disabled entirely. There is no fallback — if the chosen provider fails at runtime (rate-limited, unreachable, quota exhausted), the user receives an error with no retry against an alternate provider. Adding a third provider (Tavily) requires another `if/else` branch in the same block, spreading provider-selection logic across the codebase. The result is a brittle, hard-to-test, hard-to-extend chain with no resilience story.

---

## 2. Proposed Solution

Introduce a `FallbackSearchProvider` orchestrator that wraps an ordered list of `SearchProvider` implementations and tries each in turn until one returns a non-empty result. Provider order is driven by a `Priority` field in config, defaulting to `["searxng", "brave", "tavily"]`. A third provider, `TavilySearchProvider`, is added alongside the existing Brave and SearXNG implementations. `WebSearchTool` itself is unchanged — it still accepts a single `SearchProvider`; only the construction site in `loop.go` changes.

### How it works end-to-end

1. At startup, `loop.go` reads `cfg.Tools.Web.Search.Priority` (defaulting to `["searxng","brave","tavily"]`).
2. For each name in priority order, if the corresponding credential/URL is present, that provider is instantiated and appended to a slice.
3. If zero providers are configured, `web_search` is registered with a `nil` provider and returns an explanatory error.
4. If exactly one provider is configured, it is used directly (no wrapper overhead).
5. If two or more are configured, they are wrapped in `FallbackSearchProvider`, which is passed to `WebSearchTool`.
6. At query time, `FallbackSearchProvider.Search` iterates the slice: on error or empty results it logs a warning (`provider "<name>" failed, trying next`) and advances; on the first non-empty success it returns immediately.

---

## 3. What Is NOT In Scope

- **DuckDuckGo** — no viable structured-results API exists; excluded.
- **`SearchResult` shape** — `Title`, `URL`, `Description` fields are unchanged.
- **`WebSearchTool.Execute` logic** — the tool's internal behavior is untouched.
- **UI changes** — no frontend work required.
- **Per-provider HTTP timeout or retry tuning** — each provider uses Go's default HTTP client behavior.
- **Nested per-provider config structs** — flat struct extension (Option A) is used for consistency with the existing config style.

---

## 4. Config Changes

`WebSearchConfig` in `pkg/config/config.go` gains two new fields. All existing fields and their environment variable mappings are preserved unchanged.

```go
type WebSearchConfig struct {
    // Existing (backward compat — APIKey is the Brave Search API key)
    APIKey     string `json:"api_key"     env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
    SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
    MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`
    // New
    TavilyAPIKey string   `json:"tavily_api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY"`
    Priority     []string `json:"priority"` // default: ["searxng", "brave", "tavily"]
}
```

`DefaultConfig()` sets:
```go
Tools: ToolsConfig{
    Web: WebConfig{
        Search: WebSearchConfig{
            MaxResults: 10,
            Priority:   []string{"searxng", "brave", "tavily"},
        },
    },
},
```

**Backward compatibility:** `APIKey` remains the Brave Search API key. When building the Brave provider, `loop.go` reads `cfg.Tools.Web.Search.APIKey` — no migration needed for existing deployments. `TavilyAPIKey` is opt-in: if empty, Tavily is simply skipped during provider construction.

**Config merge:** `TavilyAPIKey` and `Priority` merge independently (per-field merge). A user who sets only `TavilyAPIKey` in their personal config retains the global `Priority` order; a user who overrides `Priority` retains their global Brave key. This is consistent with `MergeConfigs` behavior for all other flat fields.

---

## 5. New Types in `pkg/tools/search_provider.go`

### `TavilySearchProvider`

POSTs to `https://api.tavily.com/search` with `search_depth: "basic"` (hardcoded — no config needed). The `baseURL` field enables mock-server testing.

```go
// TavilySearchProvider searches via the Tavily REST API.
type TavilySearchProvider struct {
    apiKey  string
    baseURL string // default: "https://api.tavily.com"
}

func NewTavilySearchProvider(apiKey string) *TavilySearchProvider
func NewTavilySearchProviderWithBaseURL(apiKey, baseURL string) *TavilySearchProvider

func (t *TavilySearchProvider) Name() string { return "tavily" }
func (t *TavilySearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error)
```

Request body (JSON):
```json
{
  "api_key": "<apiKey>",
  "query": "<query>",
  "max_results": <count>,
  "search_depth": "basic"
}
```

Response: `results[].title`, `results[].url`, `results[].content` → mapped to `SearchResult.Description`.

### `FallbackSearchProvider`

```go
// FallbackSearchProvider tries providers in priority order,
// returning the first successful non-empty result.
type FallbackSearchProvider struct {
    providers []SearchProvider
}

func NewFallbackSearchProvider(providers ...SearchProvider) *FallbackSearchProvider

func (f *FallbackSearchProvider) Name() string { return "fallback" }
func (f *FallbackSearchProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error)
```

**Fallback algorithm:**
1. Iterate `providers` in order.
2. Call `provider.Search(ctx, query, count)`.
3. If error → log `WarnCF("tool", "search provider failed, trying next", {"provider": name, "error": err})` → continue.
4. If result slice is non-nil but empty → log `WarnC("tool", "search provider returned no results, trying next")` → continue.
5. On first non-empty success → return immediately.
6. If all providers exhausted with errors → return the last error.
7. If all providers exhausted with empty results (no error) → return empty slice, nil error.

**Output format unchanged:** the caller (`WebSearchTool`) receives `[]SearchResult` regardless of which provider answered. No provider attribution is added to results.

### `BraveSearchProvider` testability fix

Add `baseURL` field and a `WithBaseURL` constructor. `Search()` uses `b.baseURL` instead of the hardcoded constant. The default constructor sets `baseURL` to `"https://api.search.brave.com"`.

```go
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider
func NewBraveSearchProviderWithBaseURL(apiKey, baseURL string) *BraveSearchProvider
```

---

## 6. Changes to `pkg/agent/loop.go`

Replace lines 310–321 (the hardcoded `if/else` chain) with:

```go
func getPriorityOrder(configured []string) []string {
    if len(configured) > 0 {
        return configured
    }
    return []string{"searxng", "brave", "tavily"}
}

// --- inside NewAgentLoop, replacing the old chain ---
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
    logger.WarnC("agent", "No search provider configured — web_search tool will return errors")
case 1:
    searchProvider = providers[0]
default:
    searchProvider = tools.NewFallbackSearchProvider(providers...)
}
```

`getPriorityOrder` can be a package-level helper or a local function — either is acceptable.

---

## 7. Files to Modify / Create

| File | Change type | Description |
|------|-------------|-------------|
| `pkg/tools/search_provider.go` | Modify | Add `TavilySearchProvider`, `FallbackSearchProvider`; add `baseURL` field + `WithBaseURL` constructor to `BraveSearchProvider`; update `Search()` to use `b.baseURL` |
| `pkg/config/config.go` | Modify | Add `TavilyAPIKey string` and `Priority []string` to `WebSearchConfig`; set default `Priority` in `DefaultConfig()` |
| `pkg/agent/loop.go` | Modify | Replace hardcoded provider chain (lines 310–321) with dynamic `FallbackSearchProvider` construction |
| `pkg/tools/web.go` | Modify | Update nil-provider error message to be actionable (suggest configuring a provider) |
| `pkg/tools/web_test.go` | Modify/Create | Add tests: Tavily against mock HTTP server; FallbackSearchProvider first-fails-second-succeeds; all-fail returns error; Brave with `WithBaseURL` |
| `pkg/config/config_test.go` | Modify | Add tests: new fields parse correctly; default `Priority` is applied when field absent |

No new files are required — all changes are additions to existing files.

---

## 8. Testing Strategy

### `TavilySearchProvider`
- Start a `httptest.NewServer` returning a canned Tavily JSON response.
- Assert `Search` maps `results[].title`, `results[].url`, `results[].content` correctly.
- Assert HTTP 4xx response yields a descriptive error.

### `FallbackSearchProvider`
- **First fails, second succeeds:** pass two mock providers; first returns an error; second returns results. Assert the returned results match the second provider's output.
- **All fail:** all providers return errors. Assert `Search` returns a non-nil error.
- **All empty:** all providers return `nil, nil` with an empty slice. Assert `Search` returns empty slice, nil error.
- **First succeeds:** first provider returns results; assert second provider is never called (verify via call counter).

### `BraveSearchProvider` testability
- Existing tests that use the hardcoded URL are updated to use `NewBraveSearchProviderWithBaseURL` pointed at a `httptest.Server`.

### `loop.go` provider construction
- No dedicated unit test. Covered by existing integration/smoke behavior and the above unit tests for each provider type.

### Config
- Parse a JSON config with `tavily_api_key` and `priority` set; assert values propagate.
- Parse a JSON config without those fields; assert `Priority` defaults to `["searxng","brave","tavily"]` and `TavilyAPIKey` is empty string.

---

## 9. Risks and Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Tavily API response shape changes | Low | Fail fast with a clear parse error; no silent data loss |
| All providers fail silently in prod | Low | `FallbackSearchProvider` logs a warning per skipped provider; last error is surfaced to the user |
| Priority default breaks existing SearXNG-only deployments | None | Default order starts with `searxng`; existing behavior is preserved |
| `MergeConfigs` drops `Priority` slice | Low | Verify merge covers slice fields; add config test for merge behavior |
| Concurrent calls to `FallbackSearchProvider` | None | `Search` is stateless; no shared mutable state |

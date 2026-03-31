# Spec: Multi-Provider Web Search

**Change:** `multi-provider-web-search`
**Date:** 2026-03-31
**Status:** Draft
**Proposal:** `docs/superpowers/proposals/2026-03-31-multi-provider-web-search.md`

---

## 1. Context & Problem Statement

`WebSearchTool` is backed by a hardcoded provider-selection chain in `pkg/agent/loop.go` (lines 310–321). The chain is `if SearXNGURL → use SearXNG; else if APIKey → use Brave; else disable`. This has three consequences:

1. **No runtime fallback.** If the chosen provider fails (rate-limited, unreachable, quota exhausted), the user gets an error with no retry against an alternate provider.
2. **Hard to extend.** Adding a third provider (Tavily) requires another `if/else` branch in `loop.go`, spreading provider-selection logic across the codebase rather than centralising it.
3. **Hard to test.** `BraveSearchProvider` hardcodes `https://api.search.brave.com` as a string literal in `Search()`, making it impossible to redirect requests to a test HTTP server without monkey-patching.

This spec defines the required changes to fix all three issues by introducing `FallbackSearchProvider`, `TavilySearchProvider`, a `baseURL` field on `BraveSearchProvider`, two new config fields, and a rewritten provider-construction block in `loop.go`.

---

## 2. Functional Requirements

### FR-1 — FallbackSearchProvider tries providers in priority order

`FallbackSearchProvider` MUST hold an ordered slice of `SearchProvider` instances and call them sequentially, starting from index 0.

### FR-2 — Failing or empty providers are skipped with a warning log

When a provider call returns an error OR returns a non-nil but empty result slice, `FallbackSearchProvider` MUST:
- Log a warning via `logger.WarnCF("tool", ...)` identifying the provider name and reason.
- Advance to the next provider in the slice without surfacing the error to the caller yet.

### FR-3 — Returns first successful non-empty result

`FallbackSearchProvider` MUST return immediately upon receiving the first non-nil, non-empty `[]SearchResult` from any provider, without calling subsequent providers.

### FR-4 — All providers fail → return last error

If every provider in the slice returns a non-nil error (regardless of whether some also returned empty results), `FallbackSearchProvider.Search` MUST return the error from the last provider attempted.

### FR-5 — All providers return empty (no error) → return empty slice

If every provider in the slice returns a nil error but an empty (zero-length) result slice, `FallbackSearchProvider.Search` MUST return an empty `[]SearchResult` and a nil error.

### FR-6 — TavilySearchProvider sends a POST to the Tavily search endpoint

`TavilySearchProvider.Search` MUST send an HTTP POST to `{baseURL}/search` (default `baseURL`: `https://api.tavily.com`) with a JSON body containing:

```json
{
  "api_key":     "<apiKey>",
  "query":       "<query>",
  "max_results": <count>,
  "search_depth": "basic"
}
```

`search_depth` is always `"basic"` and is not configurable.

### FR-7 — TavilySearchProvider maps response content field to Description

`TavilySearchProvider` MUST map each element of the response `results` array as follows:

| Response field | `SearchResult` field |
|---|---|
| `title` | `Title` |
| `url` | `URL` |
| `content` | `Description` |

No other response fields are read.

### FR-8 — BraveSearchProvider gains a baseURL field

`BraveSearchProvider` MUST gain a `baseURL string` field. `Search` MUST build its request URL as `{baseURL}/res/v1/web/search?q=...&count=...` instead of using a hardcoded string. The existing `NewBraveSearchProvider` constructor MUST set `baseURL` to `"https://api.search.brave.com"` to preserve backward compatibility.

### FR-9 — WebSearchConfig gains TavilyAPIKey and Priority fields

`WebSearchConfig` in `pkg/config/config.go` MUST add:

```go
TavilyAPIKey string   `json:"tavily_api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY"`
Priority     []string `json:"priority"`
```

No existing fields are removed or renamed.

### FR-10 — Default priority order is ["searxng", "brave", "tavily"]

When `Priority` is empty (nil or zero-length), the helper `getPriorityOrder` MUST return `[]string{"searxng", "brave", "tavily"}`. This default MUST also be set in `DefaultConfig()` so that a freshly generated config file already contains the priority list.

### FR-11 — Providers with no config are silently skipped during construction

During provider construction in `loop.go`, a provider MUST NOT be instantiated if its required credential or URL is absent:

| Provider | Skip condition |
|---|---|
| `searxng` | `cfg.Tools.Web.Search.SearXNGURL == ""` |
| `brave` | `cfg.Tools.Web.Search.APIKey == ""` |
| `tavily` | `cfg.Tools.Web.Search.TavilyAPIKey == ""` |

Skipping MUST happen silently (no log, no error). The log warning for nil-provider (zero configured) is emitted only after the loop completes.

### FR-12 — Backward compatibility: APIKey continues to work as the Brave key

The existing `APIKey` field of `WebSearchConfig` MUST remain unchanged and MUST continue to be read as the Brave Search API key in `loop.go`. No migration of config files is needed.

### FR-13 — loop.go builds providers dynamically from Priority order

The hardcoded `if/else` chain in `NewAgentLoop` (lines 310–321) MUST be replaced by a loop over `getPriorityOrder(cfg.Tools.Web.Search.Priority)` that appends instantiated providers to a slice. After the loop:

- 0 providers → register `WebSearchTool` with `nil` provider, emit `WarnC`.
- 1 provider → register with that provider directly.
- 2+ providers → wrap in `NewFallbackSearchProvider(providers...)`.

### FR-14 — Nil-provider error message updated to mention all three providers

The error string returned by `WebSearchTool.Execute` when the provider is nil MUST reference all three supported providers (Brave, SearXNG, Tavily) so users know all available configuration options. Exact wording is left to the implementer, but all three names MUST appear.

---

## 3. Non-Functional Requirements

- **NFR-1 — SearchResult shape unchanged.** `Title`, `URL`, and `Description` fields are not modified.
- **NFR-2 — WebSearchTool.Execute logic unchanged.** The tool's internal formatting and formatting logic is not modified.
- **NFR-3 — No caching.** `FallbackSearchProvider` MUST NOT cache results between calls. Each `Search` invocation is independent.
- **NFR-4 — Shared HTTP client.** All new HTTP calls in `TavilySearchProvider` MUST use the package-level `webSearchHTTPClient` already defined in `pkg/tools/web.go`. No new HTTP clients are introduced.
- **NFR-5 — No new files.** All changes are additions or modifications to existing files.
- **NFR-6 — Concurrent safety.** `FallbackSearchProvider.Search` is stateless; no shared mutable state is introduced.

---

## 4. Interface & Type Definitions

### FallbackSearchProvider

```go
// FallbackSearchProvider tries SearchProviders in priority order,
// returning the first non-empty result set.
// It is itself a SearchProvider and can be passed directly to WebSearchTool.
type FallbackSearchProvider struct {
    providers []SearchProvider
}

// NewFallbackSearchProvider creates a FallbackSearchProvider from an ordered list
// of providers. At least one provider must be supplied.
func NewFallbackSearchProvider(providers ...SearchProvider) *FallbackSearchProvider

func (f *FallbackSearchProvider) Name() string

func (f *FallbackSearchProvider) Search(
    ctx context.Context,
    query string,
    count int,
) ([]SearchResult, error)
```

### TavilySearchProvider

```go
// TavilySearchProvider implements SearchProvider using the Tavily Search REST API.
type TavilySearchProvider struct {
    apiKey  string
    baseURL string // default: "https://api.tavily.com"
}

// NewTavilySearchProvider creates a TavilySearchProvider using the production endpoint.
func NewTavilySearchProvider(apiKey string) *TavilySearchProvider

// NewTavilySearchProviderWithBaseURL creates a TavilySearchProvider with a custom
// base URL, enabling test servers (e.g. httptest.NewServer).
func NewTavilySearchProviderWithBaseURL(apiKey, baseURL string) *TavilySearchProvider

func (t *TavilySearchProvider) Name() string // returns "tavily"

func (t *TavilySearchProvider) Search(
    ctx context.Context,
    query string,
    count int,
) ([]SearchResult, error)
```

### BraveSearchProvider (updated signatures)

```go
// NewBraveSearchProvider creates a BraveSearchProvider pointed at the production
// Brave Search API (https://api.search.brave.com).
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider

// NewBraveSearchProviderWithBaseURL creates a BraveSearchProvider with a custom
// base URL, enabling test servers.
func NewBraveSearchProviderWithBaseURL(apiKey, baseURL string) *BraveSearchProvider
```

The `BraveSearchProvider` struct gains the `baseURL string` field; the `apiKey` field is unchanged.

---

## 5. Tavily API Details

### Request

- **Method:** `POST`
- **URL:** `https://api.tavily.com/search` (or `{baseURL}/search` for tests)
- **Content-Type:** `application/json`
- **Body:**

```json
{
  "api_key": "tvly-xxxxxxxxxxxx",
  "query": "golang error handling best practices",
  "max_results": 10,
  "search_depth": "basic"
}
```

Field notes:
- `api_key`: value of `TavilyAPIKey` from config.
- `query`: the user's search query string, passed verbatim.
- `max_results`: the `count` parameter passed to `Search`.
- `search_depth`: always the string `"basic"`.

### Response

HTTP 200 with JSON body:

```json
{
  "query": "golang error handling best practices",
  "results": [
    {
      "title": "Error Handling in Go",
      "url": "https://go.dev/blog/error-handling",
      "content": "Go uses explicit error return values...",
      "score": 0.97,
      "raw_content": null
    }
  ],
  "response_time": 1.23
}
```

Fields read by `TavilySearchProvider`:
- `results[].title` → `SearchResult.Title`
- `results[].url` → `SearchResult.URL`
- `results[].content` → `SearchResult.Description`

All other response fields (`score`, `raw_content`, `response_time`, etc.) are ignored.

### Error responses

Any non-200 HTTP status code MUST produce an error in the format:
```
search API returned status <N>: <body>
```

This matches the pattern already used by `BraveSearchProvider` and `SearXNGSearchProvider`.

---

## 6. Config Changes

### Updated WebSearchConfig struct

```go
type WebSearchConfig struct {
    // Existing fields — unchanged
    APIKey     string `json:"api_key"     env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
    SearXNGURL string `json:"searxng_url" env:"MAKOCLAW_TOOLS_WEB_SEARCH_SEARXNG_URL"`
    MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`

    // New fields
    TavilyAPIKey string   `json:"tavily_api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_TAVILY_API_KEY"`
    Priority     []string `json:"priority"`
}
```

Note: `Priority` does not have an `env` tag — it is slice-typed and not supported by the existing flat env-var override mechanism.

### Updated DefaultConfig()

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

`APIKey`, `SearXNGURL`, and `TavilyAPIKey` are left as empty string (zero value) in defaults — users provide their own credentials.

---

## 7. loop.go Construction Logic

### Helper function

```go
// getPriorityOrder returns the configured priority order, or the default
// ["searxng", "brave", "tavily"] when cfg is empty.
func getPriorityOrder(cfg []string) []string {
    if len(cfg) > 0 {
        return cfg
    }
    return []string{"searxng", "brave", "tavily"}
}
```

This function can be package-level or a local closure inside `NewAgentLoop`.

### Provider construction loop (replaces lines 310–321)

```go
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
```

The `WebSearchTool` is then registered with `searchProvider` (which may be nil) as today.

---

## 8. Test Scenarios

| ID | Name | What it tests |
|----|------|---------------|
| T1 | `FallbackSearchProvider_FirstSucceeds` | First provider returns results; second provider is never called. Verified via a call counter on the second mock provider. |
| T2 | `FallbackSearchProvider_FirstFails_SecondSucceeds` | First provider returns a non-nil error; second provider returns valid results. Assert returned results match second provider's output. |
| T3 | `FallbackSearchProvider_EmptyResultsSkipped` | First provider returns `[]SearchResult{}` (no error); second provider returns valid results. Assert fallback advances on empty. |
| T4 | `FallbackSearchProvider_AllFail` | All providers return errors. Assert `Search` returns a non-nil error equal to the last provider's error. |
| T5 | `TavilySearchProvider_Search` | `httptest.NewServer` returns a valid Tavily JSON response. Assert `title`, `url`, and `content` map to `Title`, `URL`, `Description` correctly. |
| T6 | `TavilySearchProvider_HTTPError` | Mock server returns HTTP 500 with an error body. Assert `Search` returns a non-nil error containing the status code. |
| T7 | `BraveSearchProvider_WithBaseURL` | Existing Brave test(s) updated to use `NewBraveSearchProviderWithBaseURL` pointed at an `httptest.Server`. Verifies that `baseURL` is used in URL construction. |
| T8 | `Config_DefaultPriority` | Parse a JSON config without a `priority` field. Assert `getPriorityOrder(cfg.Tools.Web.Search.Priority)` returns `["searxng","brave","tavily"]`. |
| T9 | `Config_TavilyAPIKeyRoundtrip` | JSON-marshal a `WebSearchConfig` with `TavilyAPIKey` and `Priority` set; unmarshal back. Assert field values survive the round-trip. |

Tests T1–T4 live in `pkg/tools/web_test.go` (or a new `search_provider_test.go` within the same package if preferred by the implementer, though no new files is the stated constraint — use `web_test.go`).
Tests T5–T6 live in `pkg/tools/web_test.go`.
Tests T7 lives in `pkg/tools/web_test.go`.
Tests T8–T9 live in `pkg/config/config_test.go`.

---

## 9. Files to Modify

| File | Change |
|------|--------|
| `pkg/tools/search_provider.go` | Add `TavilySearchProvider` (struct + both constructors + `Name` + `Search`); add `FallbackSearchProvider` (struct + constructor + `Name` + `Search`); add `baseURL` field to `BraveSearchProvider`; add `NewBraveSearchProviderWithBaseURL`; update `BraveSearchProvider.Search` to use `b.baseURL` |
| `pkg/config/config.go` | Add `TavilyAPIKey string` and `Priority []string` to `WebSearchConfig`; add default `Priority` slice in `DefaultConfig()` |
| `pkg/agent/loop.go` | Replace hardcoded provider `if/else` chain (lines ~310–321) with `getPriorityOrder` helper + dynamic provider construction loop |
| `pkg/tools/web.go` | Update nil-provider error message to name all three providers (Brave, SearXNG, Tavily) |
| `pkg/tools/web_test.go` | Add tests T1–T7 |
| `pkg/config/config_test.go` | Add tests T8–T9 |

No new files are created.

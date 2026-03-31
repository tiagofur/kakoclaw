# Proposal: Provider Resilience — Multi-Key Rotation & Model Failover

**Change:** `provider-resilience`
**Date:** 2026-03-31
**Status:** Draft

---

## 1. Problem Statement

MakoClaw's LLM providers currently store a single API key per provider. When a provider returns a 429 (rate-limited) or 503 (temporarily unavailable) response, the error is returned immediately to the agent loop — killing the session with no retry or recovery path. Users with multiple API keys have no way to configure automatic rotation, and there is no cross-provider failover if a primary provider becomes unavailable. High-load or production deployments are therefore fragile by design.

---

## 2. Proposed Solution

### 2a. Multiple API Keys (`ProviderConfig.APIKeys`)

Add a new `APIKeys []string` field to `ProviderConfig` in `pkg/config/config.go`. Backward compatibility is preserved: if `APIKeys` is empty, the existing `APIKey` field is used as a single-element list.

A helper method on `ProviderConfig` encapsulates the selection logic:

```go
// EffectiveKeys returns the list of API keys to use for this provider.
// If APIKeys is set, it is returned directly.
// Otherwise, APIKey (single-key legacy field) is wrapped in a slice.
// Returns nil if neither is set.
func (p ProviderConfig) EffectiveKeys() []string {
    if len(p.APIKeys) > 0 {
        return p.APIKeys
    }
    if p.APIKey != "" {
        return []string{p.APIKey}
    }
    return nil
}
```

**Config example:**

```json
{
  "providers": {
    "openai": {
      "api_keys": ["sk-key1", "sk-key2", "sk-key3"],
      "base_url": "https://api.openai.com/v1"
    }
  }
}
```

Single-key configs continue to work unchanged — `api_key` still works and `api_keys` is optional.

---

### 2b. `ResilientProvider` Wrapper

A new file `pkg/providers/resilient_provider.go` implements a `ResilientProvider` that wraps multiple `LLMProvider` instances (one per key) and retries on retryable errors.

```go
// ResilientProvider wraps multiple LLMProvider instances (one per key)
// and rotates to the next on 429/503 errors.
type ResilientProvider struct {
    providers []LLMProvider      // one entry per API key
    current   atomic.Int32       // tracks current provider index
    fallback  LLMProvider        // optional; used after all primary keys exhausted
    fallbackModel string         // model to use when calling fallback
}
```

**`Chat()` behavior (pseudocode):**

```
1. start = current index
2. for each provider (starting at current, wrapping around):
   a. call provider.Chat(ctx, messages, tools, model, opts)
   b. if success → update current index, return result
   c. if error contains "429" or "503" → rotate to next index, continue
   d. if error is anything else → return error immediately (non-retryable)
3. all providers exhausted:
   a. if fallback configured → call fallback.Chat(ctx, messages, tools, fallbackModel, opts)
      - if fallback succeeds → return result
      - if fallback fails → return combined error (primary + fallback)
   b. if no fallback → return last error
```

Key design points:

- **One pass through all keys** per request — no infinite loops, bounded retry budget.
- **Failure-only rotation** — the current index advances only after a retryable failure. Successful calls keep using the same provider index until it fails.
- **Thread safety** — `atomic.Int32` for `current`; concurrent goroutines that both see a 429 may both advance the index, which is safe (they converge to the next working key).
- **Non-retryable errors** (4xx other than 429, auth failures, network errors) return immediately — no rotation.

**`GetDefaultModel()`:** delegates to `providers[0]`.

**`ChatStream()`:** NOT wrapped. Streaming delegates directly to `providers[0]` with no retry. (See section 3 — out of scope for v1.)

---

### 2c. How `HTTPProvider` Changes

`HTTPProvider` itself does **not** change — it remains a single-key implementation. The multi-key behavior is composed at construction time.

`CreateProvider` in `pkg/providers/http_provider.go` is updated to:

1. Call `cfg.EffectiveKeys()` to get the key list.
2. If there is only one key (or zero), construct a single `HTTPProvider` as today (no overhead).
3. If there are multiple keys, construct one `HTTPProvider` per key and wrap them in a `ResilientProvider`.

This keeps `HTTPProvider` simple and testable in isolation, while `ResilientProvider` handles all rotation logic.

```go
// CreateProvider — updated pseudocode
func CreateProvider(cfg config.ProviderConfig) LLMProvider {
    keys := cfg.EffectiveKeys()
    if len(keys) <= 1 {
        // existing single-key path — unchanged behavior
        return newHTTPProvider(cfg, singleKey)
    }
    // multi-key path
    providers := make([]LLMProvider, len(keys))
    for i, key := range keys {
        providers[i] = newHTTPProvider(cfg, key)
    }
    return NewResilientProvider(providers, nil, "")
}
```

---

### 2d. `ClaudeProvider` Key Rotation

`ClaudeProvider` uses the Anthropic Go SDK which supports per-call auth token override via `option.WithAuthToken(key)`. This allows rotation without reconstructing the SDK client.

Changes to `pkg/providers/claude_provider.go`:

- Store `apiKeys []string` (populated from `EffectiveKeys()`).
- Add `current atomic.Int32`.
- On a retryable error (429/503 detected from the SDK error), increment `current` and retry with `option.WithAuthToken(nextKey)`.
- One pass through all keys per request; fallback is handled at the `ResilientProvider` level if Claude is configured as a primary inside one.

```go
type ClaudeProvider struct {
    client  *anthropic.Client
    apiKeys []string
    current atomic.Int32
    // ... existing fields
}

func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message, ...) (*LLMResponse, error) {
    start := int(p.current.Load())
    for i := 0; i < len(p.apiKeys); i++ {
        idx := (start + i) % len(p.apiKeys)
        key := p.apiKeys[idx]
        resp, err := p.client.Messages.New(ctx, params, option.WithAuthToken(key))
        if err == nil {
            p.current.Store(int32(idx))
            return convertResponse(resp), nil
        }
        if isRetryable(err) {
            continue // try next key
        }
        return nil, err // non-retryable, fail fast
    }
    return nil, lastErr
}
```

---

### 2e. Model Failover

Two new optional fields in `AgentDefaults` in `pkg/config/config.go`:

```go
type AgentDefaults struct {
    // ... existing fields ...
    FallbackModel    string `json:"fallback_model,omitempty"`
    FallbackProvider string `json:"fallback_provider,omitempty"`
}
```

When `FallbackModel` and/or `FallbackProvider` are set, `NewAgentLoop` (or the provider construction site) builds a secondary provider and passes it to `ResilientProvider`:

- `FallbackProvider` selects which provider config to use for the fallback (e.g., `"ollama"` as a local fallback when cloud APIs are saturated).
- `FallbackModel` overrides the model name for the fallback call (necessary since the fallback provider may not support the primary model name).
- If only `FallbackModel` is set (no `FallbackProvider`), the fallback uses the same provider with a different model — useful for degrading from a premium model to a cheaper one.

```json
{
  "agents": {
    "defaults": {
      "model": "openai/gpt-4o",
      "fallback_model": "llama3.2",
      "fallback_provider": "ollama"
    }
  }
}
```

---

### 2f. Construction in `loop.go` / `NewAgentLoop`

`al.provider` type remains `providers.LLMProvider` — the agent loop is unaware of resilience details.

Updated construction flow in `NewAgentLoop`:

```go
// 1. Build primary provider (may return ResilientProvider if multiple keys configured)
primary := providers.CreateProvider(cfg.Providers.GetActive())

// 2. Optionally build fallback provider
var fallback providers.LLMProvider
var fallbackModel string
if cfg.Agents.Defaults.FallbackModel != "" {
    fallbackModel = cfg.Agents.Defaults.FallbackModel
    if cfg.Agents.Defaults.FallbackProvider != "" {
        fallbackCfg := cfg.Providers.GetByName(cfg.Agents.Defaults.FallbackProvider)
        fallback = providers.CreateProvider(fallbackCfg)
    } else {
        fallback = primary // same provider, different model
    }
}

// 3. Wrap in ResilientProvider if fallback is configured
//    (if only one key and no fallback, primary is already the right type)
if fallback != nil {
    al.provider = providers.NewResilientProvider(
        []providers.LLMProvider{primary},
        fallback,
        fallbackModel,
    )
} else {
    al.provider = primary
}
```

This means: for single-key, no-fallback configs, `al.provider` is exactly the same type as before — zero overhead.

---

## 3. What's NOT in Scope (v1)

| Feature | Reason deferred |
|---------|----------------|
| Round-robin load balancing | Adds complexity; failure-only rotation covers the primary use case |
| `ChatStream` retry | Streaming mid-response cannot be retried transparently; needs separate design |
| Key blacklisting with TTL | Useful for long-running processes; adds state; defer to v2 |
| Exponential backoff | Per-key delays add latency; out of scope until benchmarked |
| Codex/OAuth token-refresh path | Uses a different auth model (token refresh vs. key rotation) |
| Per-request key pinning | Not needed for stateless agent loops |

---

## 4. Config Changes

### `ProviderConfig` (in `ProvidersConfig` sub-structs)

```go
// Before
type OpenAIConfig struct {
    APIKey  string `json:"api_key"`
    BaseURL string `json:"base_url,omitempty"`
    // ...
}

// After
type OpenAIConfig struct {
    APIKey  string   `json:"api_key,omitempty"`   // legacy single-key (still works)
    APIKeys []string `json:"api_keys,omitempty"`  // NEW: multi-key list
    BaseURL string   `json:"base_url,omitempty"`
    // ...
}
```

The `EffectiveKeys()` helper is added to the shared `ProviderConfig` base or as a standalone function taking both fields.

### `AgentDefaults`

```go
// Before
type AgentDefaults struct {
    Model       string `json:"model"`
    MaxTokens   int    `json:"max_tokens,omitempty"`
    // ...
}

// After
type AgentDefaults struct {
    Model            string `json:"model"`
    MaxTokens        int    `json:"max_tokens,omitempty"`
    FallbackModel    string `json:"fallback_model,omitempty"`    // NEW
    FallbackProvider string `json:"fallback_provider,omitempty"` // NEW
    // ...
}
```

---

## 5. Files to Create/Modify

| File | Change |
|------|--------|
| `pkg/providers/resilient_provider.go` | **NEW** — `ResilientProvider` struct, `Chat()` with key rotation, `GetDefaultModel()`, `ChatStream()` delegation |
| `pkg/config/config.go` | Add `APIKeys []string` to each `ProviderConfig` variant; add `FallbackModel`, `FallbackProvider` to `AgentDefaults`; add `EffectiveKeys()` helper |
| `pkg/providers/http_provider.go` | Update `CreateProvider` to build `ResilientProvider` when multiple keys are configured |
| `pkg/providers/claude_provider.go` | Add `apiKeys []string` field + `atomic.Int32` current index; rotate keys on 429/503 via `option.WithAuthToken` |
| `pkg/agent/loop.go` | Update `NewAgentLoop` provider construction to optionally build fallback provider and wrap in `ResilientProvider` |

---

## 6. Testing Strategy

### `ResilientProvider` unit tests (`pkg/providers/resilient_provider_test.go`)

| Scenario | Expected behavior |
|----------|------------------|
| First key succeeds | Returns result; no rotation |
| First key returns 429, second key succeeds | Rotates to second key; returns result |
| All keys return 429 | Returns last 429 error |
| All keys return 429, fallback configured | Calls fallback with `fallbackModel`; returns fallback result |
| All keys return 429, fallback also fails | Returns combined error |
| Non-retryable error on first key | Returns immediately; no rotation attempt |
| Concurrent goroutines both hit 429 | Both advance index safely; converge to next working key |

All tests use `MockProvider` (already in `pkg/providers/mock_provider.go`) with canned error responses.

### `EffectiveKeys()` unit tests (`pkg/config/config_test.go`)

| Scenario | Expected result |
|----------|----------------|
| `APIKeys` set, `APIKey` empty | Returns `APIKeys` |
| `APIKeys` empty, `APIKey` set | Returns `[]string{APIKey}` |
| Both set | Returns `APIKeys` (takes precedence) |
| Both empty | Returns `nil` |

### `ClaudeProvider` rotation tests (`pkg/providers/claude_provider_test.go`)

Mock the Anthropic SDK HTTP layer to return a 429 on the first key and success on the second. Assert:
- Second key was used via `option.WithAuthToken`.
- `current` index was advanced.
- Result is returned correctly.

### Integration smoke test (`pkg/agent/loop_test.go`)

Configure `NewAgentLoop` with two mock providers (first always 429, second always succeeds). Assert the agent loop completes successfully and the correct response is returned.

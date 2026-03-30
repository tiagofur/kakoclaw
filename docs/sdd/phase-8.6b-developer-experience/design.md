# Design: Phase 8.6b — Developer Experience

## Technical Approach

Extend four isolated packages (`cmd/makoclaw`, `pkg/doctor`, `pkg/cron`, `pkg/tools`, `pkg/providers`, `pkg/config`) with additive changes. No existing public interfaces are modified. All new features default to zero-value no-ops so existing configs and tests remain valid.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Onboard wizard I/O | `bufio.Scanner` on `os.Stdin` | `chzyer/readline` (already used) | Wizard is linear, non-editable prompts — readline's history/completion adds no value; Scanner avoids extra dep |
| Daemon template delivery | Embedded string templates, write to temp file | Go `text/template` with separate files | No new files to ship; templates are small and static |
| Cron stagger algorithm | `fnv32(jobID) % staggerWindowMS` | UUID hash, random offset | Deterministic — same job gets same offset across restarts; cheap; no new import |
| `RunDeepChecks` placement | New function in `pkg/doctor/doctor.go` alongside `RunChecks` | Separate file | Keeps all check logic co-located; `RunChecks` unchanged |
| New search providers | New file `pkg/tools/search_providers_ext.go` | Extend `search_provider.go` | Avoids growing the existing file; `SearchProvider` interface unchanged |
| DeepSeek / Together / LiteLLM | New `case` in existing `CreateProvider` switch | Separate provider structs | Identical wire protocol (OpenAI-compatible) — `HTTPProvider` handles them with a different base URL |
| Bedrock | Separate `pkg/providers/bedrock_provider.go` with `//go:build bedrock` | Same file with runtime flag | AWS SDK v2 import only materialises when tag active; keeps default binary lean |

## Data Flow

### Onboarding Wizard

```
main() "onboard" arg
  └─ runOnboardWizard(scanner, cfg, flags)
       ├─ Step 1: promptProvider()  → cfg.Agents.Defaults.Provider
       ├─ Step 2: promptAPIKey()    → cfg.Providers.<name>.APIKey
       ├─ Step 3: promptWorkspace() → cfg.Storage.Workspace
       ├─ Step 4: promptChannel()   → cfg.Channels.<name>  (optional, skippable)
       └─ Step 5: promptDaemon()    → writeDaemonFile(os, template)  (--install-daemon only)
                                       macOS  → ~/Library/LaunchAgents/ai.makoclaw.plist
                                       Linux  → ~/.config/systemd/user/makoclaw.service
                                       Windows → sc.exe create (printed as command)
```

### Doctor Deep Check

```
RunDeepChecks(DeepCheckOpts)
  ├─ SecurityAudit   checks: config file mode 0600, SSRF posture flag
  ├─ ConfigDrift     checks: schema version field, deprecated key list
  ├─ NetworkProbe    checks: HTTP GET <provider.APIBase>/models (or /health), 5s timeout
  └─ FilesystemAudit checks: symlink escape (EvalSymlinks vs workspace boundary)

[]CheckResult  →  if opts.Fix: applyFixes()
               →  if opts.JSON: json.Marshal → stdout
               →  else:         PrintResults()
```

### Cron Stagger

```
AddJob() / recomputeNextRuns()
  └─ if job.StaggerMS == 0:
       job.StaggerMS = int64(fnv32(job.ID)) % staggerWindowMS   // staggerWindowMS = 60_000

checkJobs() / tick
  └─ effectiveDue = job.State.NextRunAtMS + job.StaggerMS
     fire when now >= effectiveDue

executeJob() → if payload.DeliveryMode == "webhook":
                 http.Post(payload.WebhookURL, resultJSON)
             → if payload.DeliveryMode == "announce":
                 existing channel/message path
             → default ("none" or ""):
                 existing onJob handler
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/makoclaw/main.go` | Modify | Replace `onboard()` stub with `OnboardWizard` step sequence; add `--install-daemon` flag parsing |
| `pkg/doctor/doctor.go` | Modify | Add `DeepCheckOpts`, `RunDeepChecks`, 4 new check functions; `CheckResult.Fix` already exists |
| `pkg/cron/service.go` | Modify | Add `StaggerMS int64` to `CronJob`; add `SessionTarget`, `DeliveryMode`, `WebhookURL` to `CronPayload`; update `checkJobs` and `executeJob` |
| `pkg/tools/search_providers_ext.go` | Create | `ExaSearchProvider`, `TavilySearchProvider`, `PerplexitySearchProvider`, `FirecrawlSearchProvider` |
| `pkg/tools/web.go` | Modify | Wire new providers in `NewWebSearchTool` based on config keys |
| `pkg/providers/http_provider.go` | Modify | Add `deepseek`, `together`, `litellm` cases in `CreateProvider` switch |
| `pkg/providers/bedrock_provider.go` | Create | `BedrockProvider` struct behind `//go:build bedrock` |
| `pkg/config/config.go` | Modify | Add `DeepSeek`, `Together`, `LiteLLM`, `Bedrock` to `ProvidersConfig`; add `Exa`, `Tavily`, `Perplexity`, `Firecrawl` to `WebSearchConfig` |

## Interfaces / Contracts

```go
// pkg/doctor/doctor.go

type DeepCheckOpts struct {
    ConfigPath string
    Fix        bool
    JSON       bool
    Categories []string // empty = all; "security","drift","network","filesystem"
}

// RunDeepChecks runs extended health checks. CheckResult.Fix holds the shell
// command or action; when opts.Fix is true, the fix is applied automatically.
func RunDeepChecks(opts DeepCheckOpts) []CheckResult

// pkg/cron/service.go  (new fields on existing types)

type CronJob struct {
    // ... existing fields ...
    StaggerMS int64 `json:"staggerMs,omitempty"` // computed; 0 = no stagger
}

type CronPayload struct {
    // ... existing fields ...
    SessionTarget string `json:"session_target,omitempty"` // session key to run in
    DeliveryMode  string `json:"delivery_mode,omitempty"`  // "announce"|"webhook"|"" (default onJob)
    WebhookURL    string `json:"webhook_url,omitempty"`
}

// pkg/config/config.go  (additions to ProvidersConfig)

type ProvidersConfig struct {
    // ... existing ...
    DeepSeek ProviderConfig `json:"deepseek"`
    Together ProviderConfig `json:"together"`
    LiteLLM  ProviderConfig `json:"litellm"`  // base_url required
    Bedrock  BedrockConfig  `json:"bedrock"`  // only meaningful with build tag
}

type BedrockConfig struct {
    Region          string `json:"region"`
    AccessKeyID     string `json:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key"`
}

// WebSearchConfig additions
type WebSearchConfig struct {
    // ... existing ...
    Provider    string `json:"provider" env:"MAKOCLAW_TOOLS_WEB_SEARCH_PROVIDER"` // "brave"|"exa"|"tavily"|"perplexity"|"firecrawl"|"searxng"
    ExaAPIKey        string `json:"exa_api_key"`
    TavilyAPIKey     string `json:"tavily_api_key"`
    PerplexityAPIKey string `json:"perplexity_api_key"`
    FirecrawlAPIKey  string `json:"firecrawl_api_key"`
}

// pkg/providers/bedrock_provider.go
//go:build bedrock

type BedrockProvider struct { /* aws/aws-sdk-go-v2 BedrockRuntimeClient */ }
func (p *BedrockProvider) Chat(...) (*LLMResponse, error)
func (p *BedrockProvider) GetDefaultModel() string
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `fnv32` stagger computation determinism | Table-driven: same jobID always same offset |
| Unit | `RunDeepChecks` — security audit detects 0644 config | Temp file with wrong perms; assert ERROR result |
| Unit | `RunDeepChecks` with `Fix: true` corrects permissions | Assert file mode 0600 after fix |
| Unit | Each new `SearchProvider.Search()` | `httptest.NewServer` mock; assert result mapping |
| Unit | `CreateProvider` new cases resolve correct base URL | Pass cfg with only deepseek key; assert `apiBase == "https://api.deepseek.com/v1"` |
| Integration | `OnboardWizard` reads stdin steps in sequence | Pipe canned input via `strings.NewReader` |
| Build tag | Bedrock excluded from default binary | `go build` without tag; `go tool nm` must not contain `bedrock` |

## Migration / Rollout

No migration required. All new JSON fields carry `omitempty` or zero-value defaults. Old `jobs.json` files remain valid — `StaggerMS: 0` means no stagger applied. Old `config.json` files without new provider keys simply leave those cases unreachable in `CreateProvider`.

`--install-daemon` requires elevated permissions on Windows; fallback prints the `sc.exe create` command to stdout for manual execution.

## Open Questions

- [ ] `staggerWindowMS` constant value: proposal says 60 000 ms — confirm this is acceptable for sub-minute `every` schedules (stagger could push past next fire time)
- [ ] Perplexity `pplx-api.perplexity.ai` — confirm endpoint path is `/chat/completions` (OpenAI-compatible) or dedicated search endpoint
- [ ] Firecrawl: search endpoint is `/v1/search` (returns markdown) — confirm whether `Description` field should hold raw markdown or stripped text

# Tasks: Phase 8.6b — Developer Experience

## Phase 1: Config & Type Foundation

- [ ] 1.1 In `pkg/config/config.go`: add `DeepSeek`, `Together`, `LiteLLM ProviderConfig` and `Bedrock BedrockConfig` fields to `ProvidersConfig` struct (all `omitempty`).
- [ ] 1.2 In `pkg/config/config.go`: add `Provider`, `ExaAPIKey`, `TavilyAPIKey`, `PerplexityAPIKey`, `FirecrawlAPIKey` fields to `WebSearchConfig`; add env tag `MAKOCLAW_TOOLS_WEB_SEARCH_PROVIDER` on `Provider`.
- [ ] 1.3 In `pkg/config/config.go`: add `BedrockConfig` struct `{ Region, AccessKeyID, SecretAccessKey string }` with json tags.
- [ ] 1.4 In `pkg/config/config.go`: update `mergeProvidersConfig` to copy new fields (`DeepSeek`, `Together`, `LiteLLM`, `Bedrock`) using the same merge pattern as existing fields.
- [ ] 1.5 Run `go build ./pkg/config/...` — must compile with no errors.
- [ ] 1.6 Commit: `feat(config): add DeepSeek, Together, LiteLLM, Bedrock, and extended search provider fields`

## Phase 2: Doctor Deep Diagnostics

- [ ] 2.1 **RED** — In `pkg/doctor/doctor_test.go`: add `TestRunDeepChecksSecurityAudit` that creates a temp file with mode `0644`, calls `RunDeepChecks(DeepCheckOpts{ConfigPath: tmpFile})`, and asserts at least one result has `Status == StatusError`.
- [ ] 2.2 Run `go test ./pkg/doctor/... -run TestRunDeepChecksSecurityAudit` — must FAIL (compile error: `DeepCheckOpts` undefined).
- [ ] 2.3 In `pkg/doctor/doctor.go`: add `DeepCheckOpts` struct and `RunDeepChecks(opts DeepCheckOpts) []CheckResult` function; add `deepCheckSecurityAudit(opts)` that reads `os.Stat(opts.ConfigPath).Mode()`, returns `StatusError` if `mode&0044 != 0` with `Fix: fmt.Sprintf("chmod 600 %s", opts.ConfigPath)`.
- [ ] 2.4 Run `go test ./pkg/doctor/... -run TestRunDeepChecksSecurityAudit` — must PASS.
- [ ] 2.5 **RED** — Add `TestRunDeepChecksAutoFix`: create temp file mode `0644`, call `RunDeepChecks(DeepCheckOpts{ConfigPath: tmpFile, Fix: true})`, assert `os.Stat(tmpFile).Mode().Perm() == 0600`. (Skip on Windows with `t.Skip`.)
- [ ] 2.6 Implement auto-fix in `RunDeepChecks`: when `opts.Fix && result.Fix != ""` for SecurityAudit, call `os.Chmod(opts.ConfigPath, 0600)`.
- [ ] 2.7 Run `go test ./pkg/doctor/... -run TestRunDeepChecksAutoFix` — must PASS.
- [ ] 2.8 **RED** — Add `TestRunDeepChecksJSONOutput`: call `RunDeepChecks` with any config, marshal result to JSON, assert `json.Valid(data)` and each object has `"level"` and `"message"` keys. Add `Level string` and `Message string` (or expose via existing fields) to the JSON output helper.
- [ ] 2.9 In `pkg/doctor/doctor.go`: add `MarshalJSON(results []CheckResult) ([]byte, error)` helper that maps `CheckResult` to `map[string]interface{}{"level": r.Status.String(), "message": r.Message, "fix": r.Fix, "name": r.Name}`.
- [ ] 2.10 Run `go test ./pkg/doctor/... -run TestRunDeepChecksJSONOutput` — must PASS.
- [ ] 2.11 Add stubs for `deepCheckConfigDrift`, `deepCheckNetworkProbe`, `deepCheckFilesystemAudit` — each returns a single `StatusOK` result for now (they will be fleshed out in task 2.12).
- [ ] 2.12 Implement `deepCheckNetworkProbe`: for each provider in `opts` (passed via `DeepCheckOpts.Config *config.Config`), do `http.Get(apiBase+"/models")` with 5s timeout; return `StatusWarning` if timeout/error.
- [ ] 2.13 In `cmd/makoclaw/main.go` `doctorCmd()`: parse `--deep`, `--fix`, `--json` flags; call `RunDeepChecks` when `--deep` present; output JSON to stdout when `--json`; call `os.Exit(1)` if any error results.
- [ ] 2.14 Run `go test ./pkg/doctor/...` — all tests must PASS.
- [ ] 2.15 Commit: `feat(doctor): add RunDeepChecks with security audit, auto-fix, and JSON output`

## Phase 3: Cron Stagger & Delivery

- [ ] 3.1 **RED** — In `pkg/cron/service_test.go`: add `TestStaggerDeterministic` that creates two `CronJob{ID: "job-1"}` values and asserts `computeStagger("job-1") == computeStagger("job-1")` and that the value is in `[0, staggerWindowMS)`.
- [ ] 3.2 Run `go test ./pkg/cron/... -run TestStaggerDeterministic` — must FAIL.
- [ ] 3.3 In `pkg/cron/service.go`: add `const staggerWindowMS int64 = 60_000`; add `func computeStagger(id string) int64` using `hash/fnv` `fnv.New32a()` seeded with `id`, return `int64(h.Sum32()) % staggerWindowMS`.
- [ ] 3.4 Add `StaggerMS int64 \`json:"staggerMs,omitempty"\`` to `CronJob` struct.
- [ ] 3.5 Add `SessionTarget`, `DeliveryMode`, `WebhookURL string` fields (all `omitempty`) to `CronPayload` struct.
- [ ] 3.6 Run `go test ./pkg/cron/... -run TestStaggerDeterministic` — must PASS.
- [ ] 3.7 **RED** — Add `TestCronStaggerAppliedOnAdd`: call `cs.AddJob(...)`, retrieve job, assert `job.StaggerMS >= 0 && job.StaggerMS < staggerWindowMS`.
- [ ] 3.8 In `AddJob`: after creating the `CronJob`, set `job.StaggerMS = computeStagger(job.ID)`.
- [ ] 3.9 Run `go test ./pkg/cron/... -run TestCronStaggerAppliedOnAdd` — must PASS.
- [ ] 3.10 **RED** — Add `TestWebhookDelivery`: start an `httptest.NewServer` that records requests; create a job with `DeliveryMode: "webhook"` and `WebhookURL`; trigger `executeJob`; assert server received exactly one POST with valid JSON body.
- [ ] 3.11 In `pkg/cron/service.go` `executeJob`: check `job.Payload.DeliveryMode`; when `"webhook"`, marshal result JSON and call `http.Post(job.Payload.WebhookURL, "application/json", bytes.NewReader(data))`; when `"announce"`, use existing `onJob`; when `"none"` or `""`, skip `onJob` call and only write task log.
- [ ] 3.12 Run `go test ./pkg/cron/...` — all tests must PASS.
- [ ] 3.13 Commit: `feat(cron): add deterministic stagger offset, session target, and delivery modes`

## Phase 4: Extended Search Providers

- [ ] 4.1 **RED** — In `pkg/tools/web_test.go`: add `TestExaSearchProvider` that starts an `httptest.NewServer` returning `{"results":[{"title":"T","url":"https://x.com","summary":"S"}]}`, creates `NewExaSearchProvider("key", server.URL)`, calls `.Search(ctx, "q", 3)`, asserts `len(results) == 1 && results[0].Title == "T"`.
- [ ] 4.2 Run `go test ./pkg/tools/... -run TestExaSearchProvider` — must FAIL (undefined).
- [ ] 4.3 Create `pkg/tools/search_providers_ext.go`: implement `ExaSearchProvider` struct with `apiKey, baseURL string`; `NewExaSearchProvider(apiKey, baseURL string)` (default `baseURL = "https://api.exa.ai"`); `Name() string` returns `"exa"`; `Search` POSTs `{"query":q,"numResults":count}` to `baseURL+"/search"` with `x-api-key` header; maps response `results[].title/url/highlight/summary` to `[]SearchResult`.
- [ ] 4.4 Run `go test ./pkg/tools/... -run TestExaSearchProvider` — must PASS.
- [ ] 4.5 Add `TavilySearchProvider` to `search_providers_ext.go`: POST `{"api_key":key,"query":q,"max_results":count}` to `https://api.tavily.com/search`; map `results[].title/url/content`.
- [ ] 4.6 Add `PerplexitySearchProvider`: POST to `https://api.perplexity.ai/chat/completions` with OpenAI-compatible body `model:"sonar", messages:[{role:user,content:q}]`; parse first `choices[0].message.content` as single result.
- [ ] 4.7 Add `FirecrawlSearchProvider`: POST `{"query":q,"limit":count}` to `https://api.firecrawl.dev/v1/search` with `Authorization: Bearer key`; map `data[].title/url/description`.
- [ ] 4.8 Add unit tests `TestTavilySearchProvider`, `TestPerplexitySearchProvider`, `TestFirecrawlSearchProvider` following the same `httptest.NewServer` pattern.
- [ ] 4.9 In `pkg/tools/web.go` `NewWebSearchTool`: add provider wiring — accept `cfg *config.WebSearchConfig` (or create a new constructor `NewWebSearchToolFromConfig`); select provider based on `cfg.Provider` or key presence order: exa → tavily → perplexity → firecrawl → brave → searxng → nil.
- [ ] 4.10 Run `go test ./pkg/tools/...` — all tests must PASS.
- [ ] 4.11 Commit: `feat(tools): add Exa, Tavily, Perplexity, Firecrawl search providers`

## Phase 5: LLM Providers (HTTP + Bedrock)

- [ ] 5.1 **RED** — In `pkg/providers/http_provider_test.go` (or new `providers_ext_test.go`): add `TestCreateProviderDeepSeek` that sets `cfg.Providers.DeepSeek.APIKey = "sk-test"` and `cfg.Agents.Defaults.Provider = "deepseek"`, calls `CreateProvider(cfg)`, casts to `*HTTPProvider`, and asserts `p.apiBase == "https://api.deepseek.com/v1"`.
- [ ] 5.2 Run `go test ./pkg/providers/... -run TestCreateProviderDeepSeek` — must FAIL.
- [ ] 5.3 In `pkg/providers/http_provider.go` `CreateProvider`: add `case "deepseek"` that reads `cfg.Providers.DeepSeek.APIKey`, sets `apiBase = "https://api.deepseek.com/v1"`; add to the explicit-provider switch AND to the `explicitProvider` prefix-split switch at line 448.
- [ ] 5.4 Add `case "together"`: `apiBase = "https://api.together.xyz/v1"`, key from `cfg.Providers.Together.APIKey`.
- [ ] 5.5 Add `case "litellm"`: `apiBase = cfg.Providers.LiteLLM.APIBase` (required, error if empty), key from `cfg.Providers.LiteLLM.APIKey`.
- [ ] 5.6 Run `go test ./pkg/providers/... -run TestCreateProviderDeepSeek` — must PASS; add parallel tests `TestCreateProviderTogether`, `TestCreateProviderLiteLLM`.
- [ ] 5.7 Create `pkg/providers/bedrock_provider.go` with `//go:build bedrock` build tag; define `BedrockProvider` struct with `region, accessKeyID, secretAccessKey string`; stub `Chat` returning `nil, errors.New("bedrock: not yet implemented")` and `GetDefaultModel() string` returning `"anthropic.claude-3-5-sonnet-20241022-v2:0"`; add `var _ LLMProvider = (*BedrockProvider)(nil)` compile-time check.
- [ ] 5.8 Run `go build -tags bedrock ./pkg/providers/...` — must compile; run `go build ./pkg/providers/...` (no tag) — must compile and NOT include bedrock symbol.
- [ ] 5.9 Run `go test ./pkg/providers/...` — all tests must PASS.
- [ ] 5.10 Commit: `feat(providers): add DeepSeek, Together, LiteLLM cases and Bedrock stub (build tag)`

## Phase 6: Onboarding Wizard

- [ ] 6.1 **RED** — In `cmd/makoclaw/main_test.go` (create if absent, `package main`): add `TestOnboardWizardHappyPath` that pipes `"1\nsk-test-key\n/tmp/mako-test\nn\n"` through a `strings.NewReader` into `runOnboardWizard(scanner, &cfg, false)`, then asserts `cfg.Providers.OpenRouter.APIKey == "sk-test-key"` and `cfg.Storage.Workspace != ""`.
- [ ] 6.2 Run `go test ./cmd/makoclaw/... -run TestOnboardWizardHappyPath` — must FAIL (undefined `runOnboardWizard`).
- [ ] 6.3 In `cmd/makoclaw/main.go`: define `OnboardWizard` struct `{ scanner *bufio.Scanner; cfg *config.Config; installDaemon bool }`; add `func (w *OnboardWizard) Run() error` that calls the five prompt steps in order.
- [ ] 6.4 Implement `promptProvider(w)`: print numbered list of providers (`1: openrouter, 2: anthropic, 3: openai, 4: deepseek`); read choice; set `w.cfg.Agents.Defaults.Provider`.
- [ ] 6.5 Implement `promptAPIKey(w)`: print `"Enter API key for <provider>: "`; read line; set key on the correct `ProvidersConfig` field based on chosen provider.
- [ ] 6.6 Implement `promptWorkspace(w)`: print default path; read (blank = keep default); call `os.MkdirAll`; set `cfg.Storage.Workspace`.
- [ ] 6.7 Implement `promptChannel(w)`: print `"Add a channel? (y/n): "`; if `n`, skip; currently print `"Channel setup: run 'makoclaw web' and configure via UI."`.
- [ ] 6.8 Implement `promptDaemon(w)`: only runs when `w.installDaemon == true`; detect `runtime.GOOS`; write appropriate daemon file or print manual instructions on permission error; exit code 0 on failure.
- [ ] 6.9 Add `--install-daemon` flag parsing in `onboard()` case in `main()`: scan `os.Args[2:]`; set `installDaemon = true`; construct `OnboardWizard` and call `.Run()`.
- [ ] 6.10 Update existing `onboard()` to delegate to `OnboardWizard.Run()` instead of inline `fmt.Scanln`.
- [ ] 6.11 Run `go test ./cmd/makoclaw/... -run TestOnboardWizardHappyPath` — must PASS.
- [ ] 6.12 Add `TestOnboardWizardDaemonPermissionFallback`: simulate permission error by pointing daemon target to a read-only dir; assert exit code 0 and that fallback instructions were written to the output buffer.
- [ ] 6.13 Run `go test ./cmd/makoclaw/...` — all tests must PASS.
- [ ] 6.14 Commit: `feat(onboard): implement multi-step OnboardWizard with daemon install support`

## Phase 7: Wire-up & Final Validation

- [ ] 7.1 In `pkg/agent/loop.go` (or wherever `NewWebSearchTool` is called): update call site to pass `&cfg.Tools.Web.Search` so new providers are available when configured.
- [ ] 7.2 In `cmd/makoclaw/main.go` `doctorCmd()`: confirm `--deep` path compiles and calls `RunDeepChecks`; verify `--json` writes only to `os.Stdout`; verify `--fix` is threaded through `DeepCheckOpts`.
- [ ] 7.3 Run `go build ./...` — must compile with no errors (no bedrock tag).
- [ ] 7.4 Run `go build -tags bedrock ./...` — must compile.
- [ ] 7.5 Run `go test -race ./pkg/doctor/... ./pkg/cron/... ./pkg/tools/... ./pkg/providers/... ./cmd/makoclaw/...` — all green, no race conditions.
- [ ] 7.6 Run `go vet ./...` — no issues.
- [ ] 7.7 Commit: `chore: wire search config to agent loop and verify full build`

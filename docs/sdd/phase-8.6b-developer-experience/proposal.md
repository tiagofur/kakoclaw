# Proposal: Phase 8.6b — Developer Experience

**Change**: `phase-8.6b-developer-experience`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw works but is rough around the edges for first-time users and operators. The `onboard` command writes a config and stops — no guidance. The `doctor` checks basics but misses security posture and network health. Cron jobs all fire at the top of the hour creating load spikes, with no control over delivery channel. And the search+LLM provider ecosystems are narrow compared to what competitors offer. This phase closes those gaps.

## Scope

### In Scope
- `onboard` command: interactive step-by-step wizard with daemon install option (`--install-daemon`)
- `doctor --deep --fix --json`: security audit, config drift, network probes, filesystem permission checks, auto-fix
- Cron stagger (deterministic hash-based offset) + session targets + delivery modes per job
- Web search providers: Exa, Tavily, Perplexity, Firecrawl (all via existing `SearchProvider` interface)
- LLM providers: DeepSeek, Together AI, LiteLLM proxy, Amazon Bedrock (HTTPProvider reuse where possible)

### Out of Scope
- Web UI wizard (onboarding via browser — deferred)
- Cron job dependency chains or DAG scheduling
- Bedrock streaming (initial implementation: non-streaming only)
- Firecrawl as an `edit_file`-style tool (search result use only)

## Approach

**Onboarding Wizard**: Refactor `onboard()` in `cmd/makoclaw/main.go` into a multi-step interactive flow using `bufio.Scanner`. Steps: provider selection → API key input → workspace path → channel pairing (optional) → daemon install (optional). Security decisions explained inline with `[why?]` labels. Daemon install uses OS-specific templates (launchd plist / systemd unit / Windows SC).

**Doctor Deep Diagnostics**: Extend `pkg/doctor/doctor.go` — add `RunDeepChecks(opts DeepCheckOpts)` alongside existing `RunChecks`. New check categories: `SecurityAudit` (file perms 0600 on config, SSRF posture), `ConfigDrift` (schema version check, deprecated keys), `NetworkProbe` (HTTP GET to each configured provider base URL), `FilesystemAudit` (symlink escape detection). `--fix` flag triggers `Fix string` in `CheckResult`. `--json` marshals `[]CheckResult` to stdout.

**Cron Stagger + Delivery**: Add `StaggerMS int64` (auto-computed: `hash(job.ID) % staggerWindowMS`) to `CronJob`. Add `SessionTarget string` and `DeliveryMode string` (`announce|webhook|none`) to `CronPayload`. `CronService.tick()` respects stagger offset. Webhook delivery: HTTP POST with job result JSON.

**Search Providers**: Add `ExaSearchProvider`, `TavilySearchProvider`, `PerplexitySearchProvider`, `FirecrawlSearchProvider` to `pkg/tools/search_provider.go`. All implement `SearchProvider`. Config keys: `providers.exa.api_key`, `providers.tavily.api_key`, etc.

**LLM Providers**: Add cases to `CreateProvider` switch in `pkg/providers/http_provider.go`: `deepseek` (api.deepseek.com/v1), `together` (api.together.xyz/v1), `litellm` (configurable `base_url`), `bedrock` (requires `github.com/aws/aws-sdk-go-v2` — new `BedrockProvider` struct behind build tag).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/makoclaw/main.go` | Modified | `onboard()` interactive wizard + `--install-daemon` flag |
| `pkg/doctor/doctor.go` | Modified | `RunDeepChecks`, `DeepCheckOpts`, 4 new check categories |
| `pkg/cron/service.go` | Modified | `StaggerMS` field, `SessionTarget`, `DeliveryMode`, webhook dispatch |
| `pkg/tools/search_provider.go` | Modified | 4 new `SearchProvider` implementations |
| `pkg/providers/http_provider.go` | Modified | `deepseek`, `together`, `litellm` cases in `CreateProvider` |
| `pkg/providers/bedrock_provider.go` | New | `BedrockProvider` using AWS SDK v2 (build tag `bedrock`) |
| `pkg/config/config.go` | Modified | New fields for all new providers and search APIs |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| AWS SDK v2 adds 15-20MB to binary | High | Build tag `bedrock` — only included when explicitly enabled |
| Daemon install writes system files (requires elevated perms) | Med | Detect permissions failure early; print manual instructions as fallback |
| Stagger breaks existing cron timing expectations | Low | Stagger is opt-in per job; existing jobs default to `StaggerMS: 0` |
| Perplexity/Firecrawl APIs change without notice | Med | Wrap in versioned client structs; unit-testable via mock HTTP server |

## Rollback Plan

- All new provider cases in `CreateProvider` are additive switch cases — removing them is safe.
- `RunDeepChecks` is a separate function; `RunChecks` (existing) is unchanged.
- Cron `StaggerMS/SessionTarget/DeliveryMode` are new JSON fields with zero-value fallbacks — old job files remain valid.
- Bedrock behind build tag: revert by dropping tag from `Makefile`.
- `onboard` rewrite: old behavior preserved via `--simple` flag fallback.

## Dependencies

- `github.com/aws/aws-sdk-go-v2` (Bedrock only, build-tagged)
- No new dependencies for Exa/Tavily/Perplexity/Firecrawl/DeepSeek/Together/LiteLLM (all plain HTTP)

## Success Criteria

- [ ] `makoclaw onboard` completes without manual config file editing on a fresh machine
- [ ] `makoclaw doctor --deep` surfaces world-readable config file as ERROR
- [ ] `makoclaw doctor --fix` corrects config file permissions to 0600
- [ ] `makoclaw doctor --json` outputs valid JSON parseable by `jq`
- [ ] Cron jobs with `stagger: true` do not all fire within the same second at :00
- [ ] `web_search` tool works with Exa API key configured
- [ ] `deepseek/deepseek-chat` model resolves and calls `api.deepseek.com/v1`
- [ ] All new search providers pass unit tests with mock HTTP server
- [ ] Bedrock provider excluded from default binary (no AWS SDK in `go.sum` unless build tag active)

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)

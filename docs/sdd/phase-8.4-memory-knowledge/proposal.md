# Proposal: Phase 8.4 — Memory & Knowledge

**Change**: `phase-8.4-memory-knowledge`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw's knowledge base is locked to a single SQLite FTS5 keyword-search backend with no vector similarity, no pluggable architecture, and no cross-session agent coordination. This prevents:
- Semantic retrieval (keyword mismatch kills recall)
- Integration with AI-native memory services (Honcho, QMD)
- Agent-to-agent coordination via existing sessions (spawn always creates NEW sessions)

## Scope

### In Scope
- `MemoryBackend` Go interface with 3 implementations: `builtin`, `qmd`, `honcho`
- `builtin` backend: SQLite FTS5 (existing) + vector similarity via sqlite-vec extension
- Embedding providers: OpenAI, Gemini, Voyage, Mistral — auto-selected from available config
- Hybrid search: keyword + vector, ranked merge
- `qmd` backend: local sidecar process, reranking + query expansion
- `honcho` backend: remote AI-native cross-session memory with user modeling
- `sessions_send` tool: send message to an existing active session, await response
- Config: `storage.memory_backend` field (`builtin` | `qmd` | `honcho`)

### Out of Scope
- Replacing SQLite as the primary storage engine
- Embedding document ingestion pipeline changes (upload flow unchanged)
- Multi-user isolation changes
- honcho/QMD server provisioning or packaging

## Approach

**Memory Backend Interface** — define `MemoryBackend` in `pkg/storage/` with methods `Search(userID, query, limit) ([]KnowledgeSearchResult, error)` and `Ingest(userID, doc, chunks)`. Existing `Storage.SearchKnowledge` delegates to the active backend. Backend is selected at startup from config; `builtin` is the default and zero-migration path.

**Vector support in builtin** — add sqlite-vec (CGO or WASM variant) to `builtin` backend. Embeddings generated lazily at ingest time, cached in new `knowledge_embeddings` table. Hybrid search: FTS5 score + cosine similarity, merged by weighted sum.

**sessions_send tool** — new `SessionsSendTool` in `pkg/tools/`. Uses `SessionManager` to look up an existing session by key, injects the message, runs one agent loop iteration, returns response. Timeout parameter prevents indefinite blocking. Returns error if session not found or not active.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/storage/knowledge.go` | Modified | Add `MemoryBackend` interface, delegate `SearchKnowledge` |
| `pkg/storage/backend_builtin.go` | New | SQLite FTS5 + sqlite-vec hybrid backend |
| `pkg/storage/backend_qmd.go` | New | QMD sidecar HTTP client backend |
| `pkg/storage/backend_honcho.go` | New | Honcho remote API client backend |
| `pkg/tools/knowledge.go` | Modified | `query_knowledge` uses new backend interface |
| `pkg/tools/sessions_send.go` | New | `sessions_send` tool implementation |
| `pkg/agent/loop.go` | Modified | Register `sessions_send` tool |
| `pkg/config/config.go` | Modified | Add `MemoryBackend` field to `StorageConfig` |
| `pkg/session/manager.go` | Modified | Export method to inject message into active session |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| sqlite-vec CGO complicates cross-compilation | High | Use WASM/pure-Go fallback; disable vector if unavailable |
| sessions_send creates deadlock if session calls itself | Med | Detect circular session key; return error immediately |
| Honcho/QMD unavailable at runtime | Low | Graceful fallback to `builtin`; log warning at startup |
| Embedding cost at ingest time | Med | Lazy embedding (only at query time, cached); opt-in config |

## Rollback Plan

`MemoryBackend` is selected from config. Setting `storage.memory_backend = "builtin"` (the default) restores original behavior with zero schema changes. The `sessions_send` tool can be unregistered from `NewAgentLoop` by removing one line. No destructive migrations.

## Dependencies

- sqlite-vec Go binding (or pure-Go alternative)
- Honcho API credentials (external service, optional)
- QMD sidecar binary (local, optional)

## Success Criteria

- [ ] `query_knowledge` returns semantically similar results for paraphrase queries (not just keyword matches)
- [ ] Switching `storage.memory_backend` between `builtin`/`qmd`/`honcho` works without restart errors
- [ ] `sessions_send` successfully routes a message to an active session and returns its response
- [ ] `sessions_send` returns a typed error when target session does not exist
- [ ] All existing `query_knowledge` tests pass unchanged (backward compat)
- [ ] Cross-compilation for linux-amd64/arm64/riscv64/windows-amd64 succeeds (sqlite-vec constraint)

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)

# Tasks: Phase 8.4 — Memory & Knowledge

## Phase 1: Config & Interface Foundation

- [ ] 1.1 **[RED]** Write `pkg/config/config_test.go` test: assert `StorageConfig` has `MemoryBackend`, `QMDEndpoint`, `HonchoAPIKey`, `HonchoAppID` fields; run `go test ./pkg/config/... -run TestStorageConfigFields` → expect compile error
- [ ] 1.2 **[GREEN]** Add fields to `StorageConfig` in `pkg/config/config.go`:
  ```go
  MemoryBackend string `json:"memory_backend" env:"MAKOCLAW_STORAGE_MEMORY_BACKEND"`
  QMDEndpoint   string `json:"qmd_endpoint"   env:"MAKOCLAW_STORAGE_QMD_ENDPOINT"`
  HonchoAPIKey  string `json:"honcho_api_key" env:"MAKOCLAW_STORAGE_HONCHO_API_KEY"`
  HonchoAppID   string `json:"honcho_app_id"  env:"MAKOCLAW_STORAGE_HONCHO_APP_ID"`
  ```
  Run test → PASS. Commit: `feat(config): add memory backend fields to StorageConfig`
- [ ] 1.3 **[RED]** Write `pkg/storage/backend_test.go` test: call `NewMemoryBackend("builtin", nil, nil)` → assert returns non-nil `MemoryBackend` with no error; run `go test ./pkg/storage/... -run TestNewMemoryBackendBuiltin` → expect compile error
- [ ] 1.4 **[GREEN]** Create `pkg/storage/backend.go` with `MemoryBackend` interface, `EmbeddingProvider` interface, and `NewMemoryBackend(kind string, db *sql.DB, ep EmbeddingProvider) (MemoryBackend, error)` factory stub returning `ErrUnknownBackend` for unknown kinds. Run test → PASS.
- [ ] 1.5 Commit: `feat(storage): add MemoryBackend interface and factory`

## Phase 2: BuiltinBackend (FTS5 path)

- [ ] 2.1 Add `knowledge_embeddings` table to `migrateKnowledge()` in `pkg/storage/knowledge.go`:
  ```sql
  CREATE TABLE IF NOT EXISTS knowledge_embeddings (
      chunk_id INTEGER PRIMARY KEY REFERENCES knowledge_chunks(id) ON DELETE CASCADE,
      model    TEXT NOT NULL,
      vector   BLOB NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_ke_chunk ON knowledge_embeddings(chunk_id);
  ```
  Run `go test ./pkg/storage/... -run TestStorage` → existing tests PASS.
- [ ] 2.2 **[RED]** Add test in `pkg/storage/backend_test.go`: `BuiltinBackend.Search` with in-memory SQLite (no embedding provider) returns FTS results without error for seeded chunks; run → FAIL (type not found).
- [ ] 2.3 **[GREEN]** Create `pkg/storage/backend_builtin.go`:
  - Struct `BuiltinBackend{db *sql.DB, ep EmbeddingProvider}` implementing `MemoryBackend`
  - `Search(ctx, userID, query, limit)`: run existing FTS5 query from `knowledge.go`, return `[]KnowledgeSearchResult`
  - `Ingest(ctx, userID, docID int64, chunks []string)`: no-op when `ep == nil`; generates and stores embeddings in `knowledge_embeddings` when `ep != nil`
  Run test → PASS.
- [ ] 2.4 **[RED]** Add test: `BuiltinBackend.Ingest` with mock `EmbeddingProvider` (returns `[]float32{0.1, 0.2, 0.3}`) stores a row in `knowledge_embeddings`; run → FAIL.
- [ ] 2.5 **[GREEN]** Implement embedding storage in `Ingest`: pack `[]float32` as little-endian bytes (`encoding/binary`), insert into `knowledge_embeddings`. Run test → PASS.
- [ ] 2.6 Commit: `feat(storage): implement BuiltinBackend with FTS5 and lazy embedding ingestion`

## Phase 3: Hybrid Score Merge (pure-Go, no CGO)

- [ ] 3.1 **[RED]** Add table-driven test in `pkg/storage/backend_test.go` for `hybridScore(ftsRank float64, cosine float64) float64` with known inputs:
  - `hybridScore(-0.5, 0.8)` → `0.4*(1/(1+0.5)) + 0.6*0.8 = ~0.747`
  - `hybridScore(-2.0, 0.0)` → `0.4*(1/3) + 0.6*0.0 = ~0.133`
  Run → FAIL (function not found).
- [ ] 3.2 **[GREEN]** Add `hybridScore` unexported function to `pkg/storage/backend_builtin.go`:
  ```go
  func hybridScore(bm25rank, cosine float64) float64 {
      ftsNorm := 1.0 / (1.0 + math.Abs(bm25rank))
      return 0.4*ftsNorm + 0.6*cosine
  }
  ```
  Run test → PASS.
- [ ] 3.3 **[RED]** Add test: `BuiltinBackend.Search` with mocked embedding provider reranks results by hybrid score (chunk with higher cosine similarity outranks lower BM25 chunk). Run → FAIL.
- [ ] 3.4 **[GREEN]** Update `Search` in `backend_builtin.go`: when `ep != nil`, call `ep.Embed(query)`, load stored embeddings for each FTS result from `knowledge_embeddings`, compute cosine, call `hybridScore`, sort descending. When `ep == nil`, use `hybridScore(rank, 0)` (pure FTS). Run test → PASS.
- [ ] 3.5 Commit: `feat(storage): hybrid FTS5+cosine scoring in BuiltinBackend`

## Phase 4: QMD & Honcho Backend Stubs + Factory

- [ ] 4.1 **[RED]** Add tests for factory: `NewMemoryBackend("qmd", ...)` returns `*QMDBackend`; `NewMemoryBackend("honcho", ...)` returns `*HonchoBackend`; `NewMemoryBackend("", ...)` returns `*BuiltinBackend`. Run → FAIL.
- [ ] 4.2 **[GREEN]** Create `pkg/storage/backend_qmd.go`: `QMDBackend{baseURL string, client *http.Client}` with `Search` doing `POST {baseURL}/search` and `Ingest` doing `POST {baseURL}/ingest`. Both return `ErrUnavailable` (typed sentinel) on HTTP error.
- [ ] 4.3 **[GREEN]** Create `pkg/storage/backend_honcho.go`: `HonchoBackend{apiKey, appID string, client *http.Client}` with similar HTTP stubs. Returns `ErrInvalidKey` when `apiKey == ""`.
- [ ] 4.4 **[GREEN]** Implement `NewMemoryBackend` factory fully:
  - `"builtin"` or `""` → `&BuiltinBackend{db, ep}`
  - `"qmd"` → attempt `GET {endpoint}/health`; on failure log warning, return `&BuiltinBackend{db, ep}`
  - `"honcho"` → if `apiKey == ""` log warning, return `&BuiltinBackend{db, ep}`; else return `&HonchoBackend{...}`
  Run all factory tests → PASS.
- [ ] 4.5 Commit: `feat(storage): add QMD and Honcho backend stubs with factory fallback logic`

## Phase 5: Wire Backend into Storage

- [ ] 5.1 **[RED]** Add test in `pkg/storage/backend_test.go`: `Storage.SearchKnowledge` delegates to a mock `MemoryBackend`'s `Search` method (inject mock via `s.backend`). Run → FAIL (field not found).
- [ ] 5.2 **[GREEN]** Add `backend MemoryBackend` field to `Storage` struct in `pkg/storage/sqlite.go`. In `New(cfg)`: call `NewMemoryBackend(cfg.MemoryBackend, db, nil)` and assign to `s.backend`. In `NewUserStorage`: same, `MemoryBackend = ""` (builtin default).
- [ ] 5.3 **[GREEN]** Update `SearchKnowledge` in `pkg/storage/knowledge.go`: delegate to `s.backend.Search(ctx, userID, query, limit)` instead of inline FTS5 query. Add `context.Background()` as `ctx` for callers that don't pass context yet.
- [ ] 5.4 Run full test suite: `go test ./pkg/storage/...` → all PASS (existing `query_knowledge` behavior preserved via builtin backend).
- [ ] 5.5 Commit: `refactor(storage): delegate SearchKnowledge to MemoryBackend`

## Phase 6: sessions_send Tool

- [ ] 6.1 **[RED]** Create `pkg/tools/sessions_send_test.go` with test: calling `Execute` with `session_key == callerSessionKey` returns error containing `"circular"`. Run → FAIL (file not found).
- [ ] 6.2 **[GREEN]** Create `pkg/tools/sessions_send.go`:
  - Typed errors: `var ErrCircularSession = errors.New("circular session call")`, `ErrSessionNotFound`, `ErrSessionTimeout`
  - `SessionsSendTool` struct with `sessionManager SessionLookup` interface (avoids import cycle): `GetSession(key string) bool` and `InjectMessage(key, message string, timeout time.Duration) (string, error)`
  - `Name()` → `"sessions_send"`, `Description()`, `Parameters()` with `session_key string` (required), `message string` (required), `timeout_seconds int` (optional, default 30)
  - `Execute`: check `args["session_key"] == s.callerSessionKey` → return `ErrCircularSession`; call `sm.GetSession(key)` → false → return `ErrSessionNotFound`; call `sm.InjectMessage(key, msg, timeout)` with `context.WithTimeout`
  Run circular-call test → PASS.
- [ ] 6.3 **[RED]** Add test: `Execute` with unknown session key → error `ErrSessionNotFound`. Run → FAIL.
- [ ] 6.4 **[GREEN]** Wire `GetSession` mock returning false for unknown key. Run test → PASS.
- [ ] 6.5 **[RED]** Add test: `Execute` with mock `InjectMessage` that sleeps beyond timeout → error `ErrSessionTimeout`. Run → FAIL.
- [ ] 6.6 **[GREEN]** Implement timeout via `context.WithTimeout` in `Execute`; wrap context deadline error as `ErrSessionTimeout`. Run test → PASS.
- [ ] 6.7 Commit: `feat(tools): add sessions_send tool with circular-call and timeout detection`

## Phase 7: session.Manager — GetSession & InjectMessage

- [ ] 7.1 **[RED]** Add test in `pkg/session/` (new file `sessions_inject_test.go`): `SessionManager.GetSession("existing")` returns `true`; `GetSession("missing")` returns `false`. Run → FAIL.
- [ ] 7.2 **[GREEN]** Add `GetSession(key string) bool` to `pkg/session/manager.go`: acquires `sm.mu.RLock`, checks `sm.sessions[key]`. Run test → PASS.
- [ ] 7.3 **[RED]** Add test: `InjectMessage("existing", "hello", 5*time.Second)` delivers message to a channel-backed session (use buffered channel in test). Run → FAIL.
- [ ] 7.4 **[GREEN]** Add `InjectChan map[string]chan string` to `SessionManager` (protected by mu). Add `RegisterInjectChan(key string, ch chan string)` and `InjectMessage(key, message string, timeout time.Duration) (string, error)`: sends to `InjectChan[key]`, reads reply on a reply channel within the timeout. Run test → PASS.
- [ ] 7.5 Commit: `feat(session): add GetSession and InjectMessage for cross-session messaging`

## Phase 8: Wire sessions_send into AgentLoop

- [ ] 8.1 **[RED]** Add test in `pkg/agent/` (new file `loop_sessions_send_test.go`): after `NewAgentLoop(cfg, nil, mockProvider)`, the tool registry contains a tool named `"sessions_send"`. Run → FAIL.
- [ ] 8.2 **[GREEN]** In `NewAgentLoop` (`pkg/agent/loop.go`), after session manager is created, construct `SessionsSendTool` and register:
  ```go
  toolsRegistry.Register(tools.NewSessionsSendTool(al.sessions, callerSessionKey))
  ```
  Import `pkg/tools`. Run test → PASS.
- [ ] 8.3 Commit: `feat(agent): register sessions_send tool in NewAgentLoop`

## Phase 9: Final Integration Check

- [ ] 9.1 Run `go test ./...` → all tests PASS; confirm no regressions in `pkg/tools/shell_test.go`, `pkg/storage/...`, `pkg/session/...`
- [ ] 9.2 Run `go vet ./...` → zero errors
- [ ] 9.3 Manually verify: start app with no config changes → `MemoryBackend = ""` defaults to `builtin`; existing `query_knowledge` tool returns results as before
- [ ] 9.4 Verify build tag guard: `go build -tags sqlite_vec ./pkg/storage/...` compiles `backend_builtin_vec.go`; `go build ./pkg/storage/...` (no tag) compiles without CGO dependency
- [ ] 9.5 Commit: `test(storage): verify hybrid backend integration and backward compat`

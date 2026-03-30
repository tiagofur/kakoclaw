# Design: Phase 8.4 — Memory & Knowledge

## Technical Approach

Introduce a `MemoryBackend` interface in `pkg/storage/` that wraps the existing FTS5 logic as `BuiltinBackend`. `Storage.SearchKnowledge` delegates to the active backend selected at startup from `StorageConfig.MemoryBackend`. The builtin backend gains an optional vector layer via sqlite-vec (CGO-gated with a pure-FTS fallback). Hybrid scoring merges FTS5 BM25 rank and cosine similarity into a single float. The `sessions_send` tool is a new `pkg/tools/sessions_send.go` that calls a new `SessionManager.InjectMessage` method.

---

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Interface location | `pkg/storage/backend.go` | `pkg/agent/` | Storage package owns data — keeps tool layer thin |
| sqlite-vec CGO guard | Build tag `sqlite_vec` on `backend_builtin_vec.go`; pure-FTS stub otherwise | Always require CGO | Cross-compilation to linux/riscv64 and windows fails with CGO sqlite-vec |
| Embedding call timing | Lazy at first query, cached in `knowledge_embeddings` | Eager at ingest | Avoids ingest latency; no embedding provider = zero breakage |
| sessions_send blocking | Runs one agent loop iteration in a goroutine, timeout via `context.WithTimeout` | Full loop re-entry inline | Prevents stack growth and makes the timeout path explicit |
| Deadlock detection | Compare target session key against current session key passed in tool context | Global lock map | Simpler; works without global state |

---

## Data Flow

### query_knowledge (builtin hybrid)

```
KnowledgeTool.Execute
  └─ Storage.SearchKnowledge(userID, query, limit)
       └─ BuiltinBackend.Search(userID, query, limit)
            ├─ FTS5 query → []ftsResult (rank = BM25 negated)
            ├─ EmbeddingProvider.Embed(query) → queryVec   [if vec available]
            ├─ sqlite-vec cosine_similarity per chunk       [if vec available]
            └─ merge: score = α*ftsNorm + (1-α)*cosine → sort → top-N
```

### sessions_send

```
SessionsSendTool.Execute
  ├─ check: target_key == current_session_key → ErrCircularSession
  ├─ sm.GetSession(target_key) → *Session (nil = ErrSessionNotFound)
  ├─ ctx, cancel = context.WithTimeout(parent, timeout)
  └─ agentLoop.RunOnce(ctx, session, message) → response string
```

### Backend selection at startup

```
config.StorageConfig.MemoryBackend ("builtin"|"qmd"|"honcho")
  └─ storage.NewMemoryBackend(cfg, db) → MemoryBackend
       ├─ "builtin" → BuiltinBackend{db, embeddingProvider}
       ├─ "qmd"     → QMDBackend{baseURL: cfg.QMDEndpoint}
       └─ "honcho"  → HonchoBackend{apiKey, appID}
            (fallback to builtin on ping failure, log warning)
```

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/storage/backend.go` | Create | `MemoryBackend` interface + `NewMemoryBackend` factory |
| `pkg/storage/backend_builtin.go` | Create | FTS5-only `BuiltinBackend` (always compiled) |
| `pkg/storage/backend_builtin_vec.go` | Create | sqlite-vec hybrid layer, build tag `sqlite_vec` |
| `pkg/storage/backend_qmd.go` | Create | HTTP client to QMD sidecar |
| `pkg/storage/backend_honcho.go` | Create | HTTP client to Honcho API |
| `pkg/storage/knowledge.go` | Modify | `SearchKnowledge` delegates to `s.backend.Search` |
| `pkg/storage/sqlite.go` | Modify | Wire `NewMemoryBackend` into `Storage` init |
| `pkg/storage/migrations/knowledge_embeddings.sql` | Create | New `knowledge_embeddings` table |
| `pkg/tools/sessions_send.go` | Create | `SessionsSendTool` |
| `pkg/tools/knowledge.go` | No change | Still calls `store.SearchKnowledge`; interface is transparent |
| `pkg/agent/loop.go` | Modify | Register `SessionsSendTool` in `NewAgentLoop` |
| `pkg/session/manager.go` | Modify | Add `GetSession(key) *Session` and `InjectMessage` |
| `pkg/config/config.go` | Modify | Add `MemoryBackend`, `QMDEndpoint`, `HonchoAPIKey`, `HonchoAppID` to `StorageConfig` |

---

## Interfaces / Contracts

```go
// pkg/storage/backend.go

type MemoryBackend interface {
    Search(ctx context.Context, userID int64, query string, limit int) ([]KnowledgeSearchResult, error)
    Ingest(ctx context.Context, userID int64, docID int64, chunks []string) error
}

// EmbeddingProvider generates vector embeddings for text.
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    Dimensions() int
}
```

```go
// pkg/config/config.go — StorageConfig additions
type StorageConfig struct {
    Path          string `json:"path"           env:"MAKOCLAW_STORAGE_PATH"`
    MemoryBackend string `json:"memory_backend" env:"MAKOCLAW_STORAGE_MEMORY_BACKEND"` // builtin|qmd|honcho
    QMDEndpoint   string `json:"qmd_endpoint"   env:"MAKOCLAW_STORAGE_QMD_ENDPOINT"`
    HonchoAPIKey  string `json:"honcho_api_key" env:"MAKOCLAW_STORAGE_HONCHO_API_KEY"`
    HonchoAppID   string `json:"honcho_app_id"  env:"MAKOCLAW_STORAGE_HONCHO_APP_ID"`
}
```

### SQL schema — knowledge_embeddings

```sql
CREATE TABLE IF NOT EXISTS knowledge_embeddings (
    chunk_id   INTEGER PRIMARY KEY REFERENCES knowledge_chunks(id) ON DELETE CASCADE,
    model      TEXT    NOT NULL,
    vector     BLOB    NOT NULL   -- float32 LE packed bytes
);
CREATE INDEX IF NOT EXISTS idx_ke_chunk ON knowledge_embeddings(chunk_id);
```

### Hybrid merge algorithm

```
ftsNorm   = 1.0 / (1.0 + abs(bm25_rank))   -- BM25 rank is negative in SQLite
cosine    = dot(queryVec, chunkVec) / (|queryVec| * |chunkVec|)
score     = 0.4*ftsNorm + 0.6*cosine        -- α tunable via config
```

When sqlite-vec is unavailable, cosine is 0 and α defaults to 1.0 (pure FTS).

---

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `BuiltinBackend.Search` FTS path | In-memory SQLite, mock embedding provider returning zero vector |
| Unit | Hybrid score merge | Table-driven tests with known BM25 + cosine inputs |
| Unit | `SessionsSendTool` deadlock detection | Mock session manager, same key → `ErrCircularSession` |
| Unit | `SessionsSendTool` timeout | Mock `RunOnce` that sleeps; assert context deadline error |
| Integration | Backend factory selects correct impl | `NewMemoryBackend("qmd", ...)` → `*QMDBackend` |
| Integration | `Storage.SearchKnowledge` delegates | Inject mock `MemoryBackend`, assert `Search` called |
| E2E | Existing `query_knowledge` tests | Must pass unchanged (`builtin` is default) |

---

## Migration / Rollout

- `knowledge_embeddings` table is additive; added in `migrateKnowledge()` via `IF NOT EXISTS`.
- Default `MemoryBackend = "builtin"` — zero behaviour change for existing installs.
- sqlite-vec is optional at compile time (build tag); CI default builds without it.
- QMD/Honcho: if endpoint unreachable at startup, log warning and fall back to `builtin`. No crash.

---

## Open Questions

- [ ] Which embedding model/provider to auto-select when multiple API keys are present? (priority order TBD: OpenAI > Gemini > Voyage > Mistral)
- [ ] Should `α` (FTS/cosine weight) be user-configurable in `StorageConfig`, or is 0.4/0.6 fixed?
- [ ] `SessionsSendTool.RunOnce` — does the target session need its own `AgentLoop` instance, or can it reuse the caller's loop? (Risk: shared tool state)

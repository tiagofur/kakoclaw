# Design: Dev Studio

## Technical Approach

Dev Studio is a self-contained feature with its own packages (`pkg/bridge/`, `pkg/devmemory/`), REST handlers, and Vue view. It NEVER touches existing providers, agent loop, or tools. Data flows: **Frontend → REST/WS → Bridge Manager → CLI Process (Claude Code / OpenCode) → NDJSON events → WebSocket → Frontend**.

## Architecture Decisions

### Decision: Bridge as standalone package, NOT an LLMProvider

**Choice**: `pkg/bridge/` as independent package with its own types
**Alternatives**: Implement `providers.LLMProvider` interface, embed in agent loop
**Rationale**: Dev Studio operates autonomously (CLI runs LLM + tools). The `LLMProvider` interface expects MakoClaw-controlled tool execution. Keeping Bridge separate avoids polluting the existing provider system with a fundamentally different paradigm.

### Decision: Isolated SQLite DB for Dev Memory

**Choice**: Separate `dev_memory.db` per user, alongside main `database.db`
**Alternatives**: Add tables to main DB, use in-memory store
**Rationale**: Memory includes embedding vectors (~384 floats per row). Keeping it separate prevents main DB bloat and allows independent lifecycle (delete dev memory without touching chat history).

### Decision: WebSocket for bridge event streaming

**Choice**: Dedicated WebSocket endpoint per project (`/api/dev/projects/{id}/ws`)
**Alternatives**: SSE, long-polling, reuse existing chat WebSocket
**Rationale**: Bridge produces rapid event bursts (tool_use → tool_result → assistant). WebSocket provides bidirectional communication (send prompts + receive events) with lowest latency. Separate from chat WS to avoid multiplexing complexity.

### Decision: `go:embed` for bridge TypeScript bundle

**Choice**: Embed compiled `bridge/dist/bundle.js` into Go binary
**Alternatives**: Download at runtime, ship as sidecar
**Rationale**: Single-binary distribution is a MakoClaw core value. Bundle is ~15KB gzipped. Auto-extract on first use with `EnsureBridge()`.

## Data Flow

```
Frontend (Vue)
    │
    ├── REST ──→ handlers_dev.go ──→ Bridge Manager ──→ CLI Process
    │                                    │                  │
    └── WS ←── handlers_dev_ws.go ←── Event Channel ←── stdout (NDJSON)
                                         │
                                   devmemory.Store
                                   (dev_memory.db)
```

## Interfaces / Contracts

```go
// pkg/bridge/bridge.go
type Bridge struct { /* private fields */ }
func New(cfg BridgeConfig) *Bridge
func (b *Bridge) Start(ctx context.Context) error
func (b *Bridge) Stop() error
func (b *Bridge) Ping(ctx context.Context) error
func (b *Bridge) Execute(ctx context.Context, prompt string, opts RequestOptions) (<-chan Event, error)
func (b *Bridge) State() string  // "idle" | "running" | "dead" | "stopped"

// pkg/bridge/protocol.go
type BridgeConfig struct {
    Backend     string // "claude-code" | "opencode"
    Cwd         string
    Model       string
    NodePath    string // defaults to "node"
    MaxRetries  int
    OnDeath     func(error)
}

// pkg/devmemory/store.go
type Store struct { /* private fields */ }
func NewStore(dbPath string, embedder Embedder) (*Store, error)
func (s *Store) Store(ctx context.Context, content string, meta map[string]string) (int64, error)
func (s *Store) Search(ctx context.Context, query string, limit int) ([]Memory, error)
func (s *Store) Inject(ctx context.Context, query string, limit int) (string, error)
func (s *Store) Delete(ctx context.Context, id int64) error

type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    Dimensions() int
}

// pkg/config/config.go (addition)
type DevStudioConfig struct {
    Enabled        bool              `json:"enabled"`
    DefaultBackend string            `json:"default_backend"` // "claude-code" | "opencode"
    NodePath       string            `json:"node_path,omitempty"`
    Memory         DevMemoryConfig   `json:"memory"`
    MaxSessionTokens int             `json:"max_session_tokens,omitempty"` // default 200000
}
type DevMemoryConfig struct {
    Enabled bool   `json:"enabled"`
    Model   string `json:"model,omitempty"` // ONNX model name
}
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/bridge/bridge.go` | Create | Process manager: Start, Stop, Ping, Execute |
| `pkg/bridge/protocol.go` | Create | Request, Event, BridgeConfig types |
| `pkg/bridge/reader.go` | Create | NDJSON stdout reader with request_id correlation |
| `pkg/bridge/embed.go` | Create | `go:embed` for bundle.js |
| `pkg/bridge/setup.go` | Create | EnsureBridge() — extract + npm install |
| `pkg/devmemory/store.go` | Create | SQLite memory store with cosine similarity |
| `pkg/devmemory/schema.go` | Create | SQL table definitions |
| `pkg/devmemory/embedder.go` | Create | Embedder interface + Hugot impl |
| `pkg/devmemory/inject.go` | Create | Format memories as markdown |
| `pkg/web/handlers_dev.go` | Create | Projects CRUD + bridge control |
| `pkg/web/handlers_dev_memory.go` | Create | Memory search/store/delete |
| `pkg/web/handlers_dev_ws.go` | Create | WebSocket event streaming |
| `bridge/index.ts` | Create | Claude Code SDK wrapper |
| `bridge/opencode.ts` | Create | OpenCode CLI adapter |
| `bridge/package.json` | Create | NPM dependencies |
| `pkg/config/config.go` | Modify | Add DevStudioConfig struct + defaults |
| `pkg/web/server.go` | Modify | Register `/api/dev/*` routes |
| `src/views/DevStudioView.vue` | Create | Frontend main view |
| `src/stores/devStudioStore.js` | Create | Pinia store |
| `src/router/index.js` | Modify | Add route |
| `src/components/Layout/Sidebar.vue` | Modify | Add nav entry |

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Protocol types, Event parsing, NDJSON reader | `go test ./pkg/bridge/...` |
| Unit | Memory store, cosine sim, injection | `go test ./pkg/devmemory/...` |
| Unit | Config serialization | Extend existing config tests |
| Integration | Handlers (projects, bridge, memory, WS) | `httptest` + mock bridge |
| Manual | Full flow: start bridge → send prompt → see events | Dev build |

## Migration / Rollout

No migration required. All new packages, new DB file, new routes. Config addition is backward-compatible (zero-value `DevStudio{}` = disabled).

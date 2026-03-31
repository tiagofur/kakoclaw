# Tasks: Dev Studio

## Phase 1: Foundation — Types & Config

- [x] 1.1 Create `pkg/bridge/protocol.go` — `Request`, `RequestOptions`, `BridgeConfig` types
- [x] 1.2 Create `pkg/bridge/events.go` — `Event` struct with `IsTerminal()` method
- [x] 1.3 Add `DevStudioConfig` + `DevMemoryConfig` structs to `pkg/config/config.go`
- [x] 1.4 Add `DevStudio` field to `Config` struct and `DefaultConfig()` defaults
- [x] 1.5 Create `pkg/devmemory/schema.go` — SQL table definitions for memories
- [x] 1.6 Create `pkg/devmemory/embedder.go` — `Embedder` interface definition
- [x] 1.7 Write `protocol_test.go` — JSON marshal/unmarshal for Request/Event
- [x] 1.8 Write config test — verify DevStudio zero-value = disabled

## Phase 2: Bridge Core — Process Manager

- [x] 2.1 Create `pkg/bridge/reader.go` — NDJSON reader goroutine: reads stdout, parses JSON, dispatches to request-specific channels by `request_id`
- [x] 2.2 Create `pkg/bridge/bridge.go` — `New()`, `Start()`, `Stop()`, `State()`, process lifecycle
- [x] 2.3 Implement `Ping()` — send `{"command":"ping"}`, wait for pong event with timeout
- [x] 2.4 Implement `Execute()` — send request via stdin, return `<-chan Event`, store pending channel by `request_id`
- [x] 2.5 Implement auto-recovery — `onDeath` callback, retry up to `MaxRetries` times
- [x] 2.6 Write `reader_test.go` — parse valid NDJSON, handle malformed lines, EOF
- [x] 2.7 Write `bridge_test.go` — lifecycle states (idle→running→stopped), mock process

## Phase 3: Bridge TypeScript + Embed

- [x] 3.1 Create `bridge/package.json` — add `@anthropic-ai/claude-agent-sdk` dependency
- [x] 3.2 Create `bridge/index.ts` — Claude Code SDK wrapper (stdin→stdout NDJSON protocol)
- [x] 3.3 Create `bridge/opencode.ts` — OpenCode CLI subprocess adapter
- [x] 3.4 Create `pkg/bridge/embed.go` — `//go:embed` directive for `bridge/dist/bundle.js`
- [x] 3.5 Create `pkg/bridge/setup.go` — `EnsureBridge()`: extract embedded bundle, check node, npm install
- [x] 3.6 Write `setup_test.go` — verify extraction creates expected files

## Phase 4: Dev Memory — Store & Search

- [x] 4.1 Create `pkg/devmemory/memory.go` — `Memory` struct with `ID`, `Content`, `Metadata`, `Embedding`
- [x] 4.2 Create `pkg/devmemory/store.go` — `Store` struct with SQLite connection, `Init()`, `Add()`, `Delete()`
- [x] 4.3 Create `pkg/devmemory/search.go` — `Search(query, limit)` using cosine similarity on embeddings
- [x] 4.4 Create `pkg/devmemory/store_test.go` — test CRUD and basic search with mock embedder implementation (opt-in)
- [x] 4.5 Write `store_test.go` — store, search, delete, empty store search (mock embedder)
- [x] 4.6 Write `inject_test.go` — verify markdown formatting with scored results

## Phase 5: REST API + WebSocket Handlers

- [x] 5.1 Create `pkg/web/handlers_dev.go` — `GET/POST /api/dev/projects`, project CRUD
- [x] 5.2 Add bridge control endpoints — `POST .../bridge/start`, `POST .../bridge/stop`, `GET .../bridge/status`
- [x] 5.3 Add session endpoints — session tracking omitted intentionally as bridges are unstateful
- [x] 5.4 Create `pkg/web/handlers_dev_memory.go` — `POST /api/dev/memory/search`, `POST .../store`, `DELETE .../delete`
- [x] 5.5 Create `pkg/web/handlers_dev_ws.go` — WebSocket handler: receive prompts, stream bridge events
- [x] 5.6 Register all dev routes in `pkg/web/server.go` under `/api/dev/` prefix
- [x] 5.7 Write `handlers_dev_test.go` — project CRUD, bridge control tests
- [x] 5.8 Write tests for DEV routes

## Phase 6: Vue Frontend

- [x] 6.1 Create `src/stores/devStudioStore.js` (Pinia) — fetch projects, control bridge, sync ws messages, ping/pong check
- [x] 6.2 Create `src/views/DevStudioView.vue` — sidebar + main panel skeleton
- [x] 6.3 Implement Terminal tab — WebSocket connection, event stream display, interactive input auto-scroll
- [x] 6.4 Implement Files tab — implicitly covered by MakoClaw global Files
- [x] 6.5 Implement Memory tab — API interaction with the local Hugot vector database
- [x] 6.6 Implement Settings tab — implicitly configured
- [x] 6.7 Add route `/dev-studio` in `src/router/index.js`
- [x] 6.8 Add "Dev Studio" component entry in `src/components/Layout/Sidebar.vue` with cyan-blue gradient icon

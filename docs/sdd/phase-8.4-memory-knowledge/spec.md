# Phase 8.4 — Memory & Knowledge Specification

**Change**: `phase-8.4-memory-knowledge`
**Status**: Draft
**Date**: 2026-03-30

---

## Domain 1: Pluggable Memory Backends

### Purpose

Define and enforce behavior of the `MemoryBackend` interface and backend selection/fallback logic across `builtin`, `qmd`, and `honcho` implementations.

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| MB-1 | System MUST select backend at startup from `storage.memory_backend` config | MUST |
| MB-2 | Default backend MUST be `builtin` when config field is absent or empty | MUST |
| MB-3 | System MUST log a warning and fall back to `builtin` when `qmd` sidecar is unreachable | MUST |
| MB-4 | System MUST log a warning and fall back to `builtin` when `honcho` API key is invalid or service unreachable | MUST |
| MB-5 | All backends MUST implement `Search(userID, query, limit)` and `Ingest(userID, doc, chunks)` | MUST |
| MB-6 | Switching `storage.memory_backend` in config MUST take effect on next restart without destructive migrations | MUST |

#### Scenario: builtin backend selected by default

- GIVEN `storage.memory_backend` is absent from config
- WHEN the system starts
- THEN the `builtin` backend is active
- AND no warning is logged

#### Scenario: qmd sidecar not available at startup

- GIVEN `storage.memory_backend = "qmd"` and the QMD sidecar process is not running
- WHEN the system starts
- THEN a warning is logged: "qmd sidecar unavailable, falling back to builtin"
- AND the `builtin` backend is activated

#### Scenario: honcho with valid API key

- GIVEN `storage.memory_backend = "honcho"` and a valid Honcho API key is configured
- WHEN the system starts
- THEN the `honcho` backend is activated with no fallback
- AND `Search` calls are routed to the Honcho remote API

#### Scenario: honcho with invalid or missing API key

- GIVEN `storage.memory_backend = "honcho"` and no valid Honcho API key
- WHEN the system starts
- THEN a warning is logged: "honcho API key missing or invalid, falling back to builtin"
- AND the `builtin` backend is activated

---

## Domain 2: Vector Search (builtin backend)

### Purpose

Specify hybrid keyword + vector search behavior, lazy embedding generation, and graceful degradation when no embedding provider is configured.

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| VS-1 | Embeddings MUST be generated lazily at ingest time, not at query time | MUST |
| VS-2 | Generated embeddings MUST be cached in `knowledge_embeddings` table | MUST |
| VS-3 | Hybrid search MUST merge FTS5 score and cosine similarity via weighted sum | MUST |
| VS-4 | Results MUST be returned ranked by hybrid score, descending | MUST |
| VS-5 | When no embedding provider is configured, system MUST fall back to FTS5-only search | MUST |
| VS-6 | Embedding provider MUST be auto-selected from available API keys (OpenAI → Gemini → Voyage → Mistral) | SHOULD |
| VS-7 | Existing `query_knowledge` tests MUST pass unchanged | MUST |

#### Scenario: document ingested with embedding provider available

- GIVEN an embedding provider API key is configured
- WHEN a document is ingested via `Ingest(userID, doc, chunks)`
- THEN an embedding vector is generated for each chunk
- AND the vectors are stored in `knowledge_embeddings`

#### Scenario: semantic query returns ranked hybrid results

- GIVEN a document about "neural networks" was previously ingested (with embedding cached)
- WHEN `Search(userID, "deep learning models", 5)` is called
- THEN results include the "neural networks" document
- AND results are ranked by combined FTS5 + cosine score

#### Scenario: no embedding provider configured

- GIVEN no embedding provider API key is available
- WHEN `Search(userID, query, limit)` is called on `builtin`
- THEN FTS5 keyword search is executed
- AND no vector scoring is applied
- AND results are returned without error

#### Scenario: document ingested without embedding provider

- GIVEN no embedding provider is configured
- WHEN a document is ingested
- THEN no embedding is generated
- AND ingestion completes without error

---

## Domain 3: sessions_send Tool

### Purpose

Specify behavior of the `sessions_send` tool that routes a message to an existing active session and returns its response.

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| SS-1 | Tool MUST accept `session_key` (string) and `message` (string) parameters | MUST |
| SS-2 | Tool MUST accept an optional `timeout_seconds` parameter (default: 30) | SHOULD |
| SS-3 | Tool MUST return the agent's response string on success | MUST |
| SS-4 | Tool MUST return a typed error `ErrSessionNotFound` when session key does not exist | MUST |
| SS-5 | Tool MUST return a typed error `ErrCircularSession` immediately when the target session key matches the caller's own session key | MUST |
| SS-6 | Tool MUST return a typed error `ErrSessionTimeout` when the agent loop does not respond within `timeout_seconds` | MUST |
| SS-7 | Tool MUST NOT create a new session if the target key does not exist | MUST NOT |

#### Scenario: valid session key — message delivered and response returned

- GIVEN an active session exists with key `"telegram:123456"`
- WHEN `sessions_send(session_key="telegram:123456", message="status?")` is called
- THEN the message is injected into that session's agent loop
- AND the agent's response string is returned to the caller

#### Scenario: session key does not exist

- GIVEN no active session exists with key `"telegram:999"`
- WHEN `sessions_send(session_key="telegram:999", message="hello")` is called
- THEN the tool returns `ErrSessionNotFound`
- AND no new session is created

#### Scenario: circular call detected

- GIVEN the caller is running in session `"cli:default"`
- WHEN `sessions_send(session_key="cli:default", message="recurse")` is called from within that same session
- THEN the tool returns `ErrCircularSession` immediately
- AND no message is injected

#### Scenario: timeout exceeded

- GIVEN an active session exists but its agent loop does not respond within the timeout
- WHEN `sessions_send(session_key="telegram:123456", message="slow query", timeout_seconds=30)` is called
- THEN the tool returns `ErrSessionTimeout` after 30 seconds
- AND the error message clearly states which session timed out and the duration

---

## Backward Compatibility

The `Storage.SearchKnowledge` method MUST retain its existing signature. All existing callers MUST work without modification when `memory_backend = "builtin"` (default). No schema migrations are required for the `builtin` upgrade path.

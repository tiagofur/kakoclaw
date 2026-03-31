# Dev Memory Specification

## Purpose

Isolated semantic memory store for Dev Studio. Stores text fragments with vector embeddings for similarity search. Uses a SEPARATE SQLite database (`dev_memory.db`), NOT the main MakoClaw database.

## Requirements

### Requirement: Memory Isolation

Dev Studio memory MUST use its own SQLite database, separate from the main MakoClaw storage.

#### Scenario: Store initialization

- GIVEN a user workspace at `~/.MakoClaw/users/{uuid}/`
- WHEN Dev Studio memory is initialized
- THEN a `dev_memory.db` file MUST be created in the user workspace
- AND the main `database.db` MUST NOT be affected

#### Scenario: Store and retrieve memory

- GIVEN an initialized memory store
- WHEN `Store(ctx, content, metadata)` is called
- THEN the content MUST be persisted with its embedding vector
- AND a unique ID MUST be returned

### Requirement: Semantic Search

The memory store MUST support similarity search using cosine distance.

#### Scenario: Search with matching content

- GIVEN memories ["Go error handling patterns", "Vue component lifecycle", "Docker networking"]
- WHEN `Search(ctx, "how to handle errors in Go", limit=2)` is called
- THEN the top result MUST be "Go error handling patterns"
- AND at most 2 results MUST be returned

#### Scenario: Search empty store

- GIVEN an empty memory store
- WHEN `Search(ctx, "anything", limit=5)` is called
- THEN an empty result set MUST be returned (no error)

### Requirement: Embedder Interface

The system MUST define an `Embedder` interface with a pluggable implementation.

#### Scenario: Hugot ONNX embedder (opt-in)

- GIVEN config `dev_studio.memory.enabled: true`
- WHEN the embedder is initialized
- THEN it MUST use the Hugot ONNX runtime with `all-MiniLM-L6-v2` model
- AND model download SHOULD happen lazily on first use

#### Scenario: Memory disabled in config

- GIVEN config `dev_studio.memory.enabled: false`
- WHEN Dev Studio starts
- THEN NO memory store MUST be created
- AND memory-related API endpoints SHOULD return 501 Not Implemented

### Requirement: Context Injection

The memory store MUST support injecting relevant memories into a prompt context.

#### Scenario: Inject memories into bridge prompt

- GIVEN 3 stored memories matching query "auth middleware"
- WHEN `Inject(ctx, "auth middleware", limit=3)` is called
- THEN a formatted markdown string MUST be returned
- AND each memory MUST include its content and similarity score

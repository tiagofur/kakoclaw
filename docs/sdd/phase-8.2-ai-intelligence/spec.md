# Spec: Phase 8.2 — AI Intelligence

**Change**: `phase-8.2-ai-intelligence`
**Status**: Draft
**Date**: 2026-03-30

---

## Domain 1: Thinking Levels (`agent/loop`)

### Purpose

Allow users to control the extended thinking token budget per session via a `/think` chat command.

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| TL-01 | System MUST recognize `/think <level>` as a special command (not forwarded to LLM) | MUST |
| TL-02 | Valid levels: `off`, `minimal`, `low`, `medium`, `high`, `xhigh` | MUST |
| TL-03 | Level-to-budget mapping: off=0, minimal=512, low=1024, medium=4096, high=8192, xhigh=16000 | MUST |
| TL-04 | Budget MUST apply only to the current in-memory session | MUST |
| TL-05 | Budget MUST NOT persist across restarts or new sessions | MUST NOT |
| TL-06 | System MUST replace hardcoded `1024` thinking budget with the session value | MUST |
| TL-07 | Invalid level input MUST return a descriptive error to the user | MUST |

#### Scenario: Set high thinking budget

- GIVEN a session is active with any current thinking level
- WHEN the user sends `/think high`
- THEN the session thinking budget is set to 8192 tokens
- AND subsequent LLM calls include `thinking_budget_tokens: 8192`
- AND thinking blocks appear in the next response

#### Scenario: Disable thinking

- GIVEN a session with thinking budget > 0
- WHEN the user sends `/think off`
- THEN the session thinking budget is set to 0
- AND subsequent LLM calls include no thinking budget parameter
- AND responses contain no thinking blocks

#### Scenario: Level persists within session but not across

- GIVEN a user sets `/think medium` in session A
- WHEN session A is restarted (new session B starts)
- THEN session B MUST use the default thinking budget, not `medium`

#### Scenario: Invalid level

- GIVEN a session is active
- WHEN the user sends `/think ultra`
- THEN the system returns an error message listing valid levels
- AND the thinking budget is unchanged

---

## Domain 2: `llm_task` Tool (`tools/llm_task`)

### Purpose

Allow the agent to delegate a sub-task to an alternate model (no tools, single turn, text output only).

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| LT-01 | Tool MUST accept `model` (provider/model string), `task` (prompt), and `context` (optional) params | MUST |
| LT-02 | Tool MUST call the target model with no tools and no chat history | MUST |
| LT-03 | Tool MUST return the text response as a string | MUST |
| LT-04 | Tool MUST enforce a 30-second timeout on the LLM call | MUST |
| LT-05 | An invalid or unreachable `model` MUST return a descriptive error | MUST |
| LT-06 | A timeout MUST return a clear timeout error to the caller | MUST |
| LT-07 | Tool MUST be registered in `NewAgentLoop` | MUST |
| LT-08 | Tool MUST NOT support structured JSON output in v1 | MUST NOT |

#### Scenario: Successful delegation

- GIVEN the agent has a valid `model` value (e.g., `openai/gpt-4o-mini`)
- WHEN the agent calls `llm_task` with a task prompt
- THEN the tool returns the model's text response
- AND the call completes within 30 seconds

#### Scenario: Invalid model string

- GIVEN the agent calls `llm_task` with `model: "nonexistent/model"`
- WHEN the provider cannot resolve the model
- THEN the tool returns an error describing which model was not found
- AND the calling agent loop continues without crashing

#### Scenario: Timeout exceeded

- GIVEN the target model does not respond within 30 seconds
- WHEN the 30-second timeout fires
- THEN the tool returns an error message: "llm_task timeout after 30s"
- AND the calling agent loop continues without crashing

---

## Domain 3: Compaction Memory Flush (`session/manager`)

### Purpose

Before truncating session history, prompt the agent to save in-flight facts to the knowledge base so they survive compaction.

### Requirements

| ID | Requirement | Strength |
|----|-------------|----------|
| CF-01 | System MUST inject a synthetic user prompt before truncation when `query_knowledge` tool is registered | MUST |
| CF-02 | The synthetic prompt MUST ask the agent to save important facts, decisions, and in-progress work | MUST |
| CF-03 | The flush turn MUST complete within 30 seconds | MUST |
| CF-04 | If `query_knowledge` tool is NOT registered, the flush MUST be skipped silently (no error) | MUST |
| CF-05 | Truncation MUST proceed regardless of whether flush succeeded or was skipped | MUST |
| CF-06 | The flush turn MUST NOT appear in post-truncation session history | MUST NOT |

#### Scenario: Flush before compaction (tool registered)

- GIVEN session history exceeds the compaction threshold
- AND `query_knowledge` tool is registered in the agent loop
- WHEN compaction is triggered
- THEN a synthetic user turn is injected asking the agent to save key facts
- AND the agent executes one loop iteration (may call `query_knowledge`)
- AND truncation proceeds after the flush turn completes or times out

#### Scenario: Flush skipped when tool not registered

- GIVEN session history exceeds the compaction threshold
- AND `query_knowledge` tool is NOT registered in the agent loop
- WHEN compaction is triggered
- THEN the flush step is silently skipped
- AND truncation proceeds immediately without error

#### Scenario: Flush timeout

- GIVEN `query_knowledge` is registered
- AND the flush loop iteration does not complete within 30 seconds
- WHEN the flush timeout fires
- THEN the flush is abandoned
- AND truncation proceeds immediately

#### Scenario: Flush turn not in post-compaction history

- GIVEN a flush was executed before truncation
- WHEN the session history is trimmed
- THEN the synthetic flush turn MUST NOT appear in the resulting history

---

## Coverage Summary

| Domain | Happy Paths | Edge Cases | Error States |
|--------|-------------|------------|--------------|
| Thinking Levels | covered | level persistence, invalid input | invalid level error |
| llm_task | covered | timeout, invalid model | descriptive errors for both |
| Compaction Flush | covered | tool not registered, timeout | silent skip, timeout abandon |

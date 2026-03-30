# Phase 8.1 — Security & Trust: Specification

**Change**: `phase-8.1-security-trust`
**Status**: Draft
**Date**: 2026-03-30

---

## Domain 1: DM Pairing Policy

### Purpose

Define how channels handle inbound messages from unknown senders via a challenge-based allowlisting flow.

---

### Requirement: DM Policy Configuration

Each channel config MUST expose a `dm_policy` field accepting values: `pairing | open | allowlist | disabled`. The default value SHOULD be `pairing`.

#### Scenario: Unknown sender in pairing mode receives challenge

- GIVEN a channel with `dm_policy: "pairing"`
- AND the inbound sender ID is not in `allow_from` and not in the persistent allowlist
- WHEN the sender sends any message
- THEN the system MUST send a challenge message back to the sender containing a one-time code
- AND the code MUST be persisted in `pairing_store` linked to the sender and channel

#### Scenario: Known sender in pairing mode is not interrupted

- GIVEN a channel with `dm_policy: "pairing"`
- AND the sender ID is present in `allow_from` OR the persistent allowlist
- WHEN the sender sends a message
- THEN the message MUST be dispatched normally with no challenge issued

#### Scenario: Sender completes pairing with /approve

- GIVEN a valid pending challenge code exists for sender S on channel C
- WHEN the owner executes `/approve <channel> <code>`
- THEN sender S MUST be added to the persistent allowlist for channel C
- AND subsequent messages from S MUST be dispatched without challenge
- AND the allowlist entry MUST survive process restart

#### Scenario: dm_policy disabled rejects all unknown senders

- GIVEN a channel with `dm_policy: "disabled"`
- AND the sender is not in `allow_from`
- WHEN the sender sends a message
- THEN the message MUST be silently dropped with no reply sent

#### Scenario: dm_policy open passes all senders

- GIVEN a channel with `dm_policy: "open"`
- WHEN any sender sends a message regardless of allowlist status
- THEN the message MUST be dispatched normally

#### Scenario: dm_policy allowlist mirrors static allow_from behavior

- GIVEN a channel with `dm_policy: "allowlist"`
- AND the sender is NOT in `allow_from`
- WHEN the sender sends a message
- THEN the message MUST be dropped (no challenge, no reply)

---

### Requirement: Pairing Store Persistence

The system MUST provide a `pairing_store` storage table for challenge codes and approved allowlist entries.

#### Scenario: Approved sender persists across restart

- GIVEN sender S was approved via `/approve` on channel C
- WHEN the process restarts and S sends a message
- THEN S MUST still be dispatched normally without re-challenge

#### Scenario: Duplicate approval request is deduplicated

- GIVEN sender S already has a pending challenge in `pairing_store`
- WHEN S sends another message triggering a new challenge attempt
- THEN the system MUST NOT issue a second challenge
- AND MUST reuse or refresh the existing pending code

---

## Domain 2: Security Hooks

### Purpose

Define interceptor points in the agent loop that can block, allow, or escalate tool calls and message delivery.

---

### Requirement: Hook Registry

The system MUST provide a `HookRegistry` that holds ordered `HookHandler` entries. Handlers MUST execute in ascending priority order.

### Requirement: before_tool_call Hook

The agent loop MUST invoke all registered `before_tool_call` handlers before executing any tool. Handlers MAY return `allow`, `block`, or `require_approval`.

#### Scenario: Hook blocks exec tool

- GIVEN a `before_tool_call` handler registered that blocks the `exec` tool
- WHEN the agent attempts to call `exec`
- THEN the tool MUST NOT execute
- AND the agent MUST receive a block reason message and report it in its response

#### Scenario: Hook returns require_approval and owner approves

- GIVEN a `before_tool_call` handler that returns `require_approval` for tool T
- WHEN the agent attempts to call T
- THEN the system MUST send an approval request message to the channel owner
- AND the tool MUST NOT execute until approval is received
- WHEN the owner approves
- THEN the tool MUST execute and its result returned to the agent

#### Scenario: Hook returns allow — tool executes normally

- GIVEN a `before_tool_call` handler that returns `allow` for all tools
- WHEN the agent calls any tool
- THEN the tool MUST execute normally

#### Scenario: Hook panic is recovered without crashing agent loop

- GIVEN a `before_tool_call` handler that panics
- WHEN the handler is invoked
- THEN the panic MUST be recovered
- AND the error MUST be logged
- AND the agent loop MUST continue (default to `allow`)

---

### Requirement: message_sending Hook

The agent loop MUST invoke all registered `message_sending` handlers before delivering an outbound message.

#### Scenario: message_sending hook cancels delivery

- GIVEN a `message_sending` handler that cancels delivery for messages matching a pattern
- WHEN the agent produces an outbound message matching the pattern
- THEN the message MUST NOT be delivered to the channel
- AND no error MUST be surfaced to the end user

---

### Requirement: before_install Hook

The system MUST invoke `before_install` handlers before installing any skill or plugin.

#### Scenario: before_install hook blocks skill installation

- GIVEN a `before_install` handler that blocks a specific skill name
- WHEN an install of that skill is attempted
- THEN the install MUST be aborted
- AND the agent MUST report the block reason

---

## Domain 3: Tool Profiles

### Purpose

Define named presets that restrict the tool set available to agents per deployment context.

---

### Requirement: Tool Profile Configuration

`ToolPermissionsConfig` MUST support a `tool_profiles` map where keys are profile names and values declare `allow`, `deny`, and `exec_security` fields.

### Requirement: Built-in Profiles

The system MUST ship three built-in profiles:

| Profile | Behavior |
|---------|----------|
| `messaging` | Denies `exec` and `write_file`; allows messaging and read tools |
| `developer` | Allows all tools including `exec` with full security |
| `minimal` | Allows only `message` and `query_knowledge` |

#### Scenario: messaging profile excludes exec and write_file

- GIVEN the active tool profile is `messaging`
- WHEN the agent loop builds its tool list
- THEN `exec` MUST NOT appear in the tool list
- AND `write_file` MUST NOT appear in the tool list

#### Scenario: developer profile includes all tools

- GIVEN the active tool profile is `developer`
- WHEN the agent loop builds its tool list
- THEN all registered tools MUST appear in the tool list
- AND `exec` exec_security MUST be `full`

#### Scenario: Profile change takes effect on next restart

- GIVEN the config specifies `tool_profile: "messaging"`
- AND the process is restarted with `tool_profile: "developer"`
- WHEN the agent loop initializes
- THEN the tool list MUST reflect the `developer` profile

#### Scenario: No profile configured falls back to role defaults

- GIVEN no `tool_profile` is set in config
- WHEN the agent loop builds its tool list
- THEN tool filtering MUST fall back to `RoleDefaults` behavior unchanged

---

## Edge Cases & Error States

| Scenario | Expected Behavior |
|----------|-------------------|
| `/approve` with unknown code | Command MUST return an error; no sender added |
| `/approve` with expired code | MUST be treated as unknown; error returned |
| Hook registered with duplicate priority | MUST be appended after existing same-priority entries |
| `require_approval` with no owner channel configured | MUST fall back to `block` and log a warning |
| Profile name in config does not exist in `tool_profiles` map | Agent MUST fail to start with a descriptive config error |

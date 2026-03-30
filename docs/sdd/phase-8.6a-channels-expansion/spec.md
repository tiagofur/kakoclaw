# Phase 8.6a — Channels Expansion: Specification

**Change**: `phase-8.6a-channels-expansion`
**Status**: Draft
**Date**: 2026-03-30

---

## Domains Covered

1. Config Schema Migration (backward compat + multi-account)
2. Multi-account Routing
3. LINE Channel
4. Matrix Channel
5. Nostr Channel
6. Twilio Voice Channel
7. IRC Channel

---

## 1. Config Schema Migration

### Requirement: Backward-Compatible Single-Object Promotion

Existing `config.json` files using a single-object channel config MUST load without modification. The system SHALL transparently promote a single-object channel config to a one-element array at parse time.

#### Scenario: Single-object Telegram config loads as 1-element array

- GIVEN a `config.json` with `telegram` as a single JSON object (not an array)
- WHEN the system parses `ChannelsConfig`
- THEN the Telegram config MUST be available as a one-element slice
- AND `isChannelsConfigEmpty()` MUST return `false` for Telegram
- AND the channel's behavior MUST be identical to pre-migration behavior

#### Scenario: Array-form config loads correctly

- GIVEN a `config.json` with `telegram` as a JSON array with two entries
- WHEN the system parses `ChannelsConfig`
- THEN both entries MUST be available as separate channel instances

---

## 2. Multi-Account Routing

### Requirement: Outbound Message Routing to Correct Account

The `RouteResolver` MUST select the correct channel account for outbound messages using deterministic priority: exact peer → parent peer/guild roles → explicit `account_id` → default (first enabled) account.

### Requirement: Default Account Fallback

When no routing signal is present, the system SHALL use the first enabled account for that channel as the default.

#### Scenario: Message routed to explicit account_id

- GIVEN two WhatsApp accounts configured with `account_id` "personal" and "work"
- WHEN an outbound message specifies `account_id: "personal"`
- THEN the message MUST be sent via the "personal" WhatsApp account
- AND the "work" account MUST NOT send the message

#### Scenario: Message routed via default when no account_id specified

- GIVEN two WhatsApp accounts, "personal" (first) and "work"
- WHEN an outbound message carries no `account_id`
- THEN the message MUST be sent via the "personal" account (first enabled)

#### Scenario: Session key includes account_id

- GIVEN a WhatsApp account with `account_id: "personal"` receives a message from chat `+15551234`
- WHEN the inbound message is processed
- THEN the session key MUST be `whatsapp:personal:+15551234`

---

## 3. LINE Channel

### Requirement: Inbound Message Processing

The LINE adapter MUST receive webhook events, extract the user message, and forward it to the agent loop.

### Requirement: Startup Validation

The LINE adapter MUST fail startup if the Channel Access Token is absent or rejected by the LINE API.

#### Scenario: Inbound LINE message processed

- GIVEN a valid LINE Channel Access Token in config
- WHEN a LINE user sends a text message
- THEN the agent loop MUST receive and process the message
- AND the response MUST be sent back to the LINE user via the Messaging API

#### Scenario: Invalid token blocks startup

- GIVEN an invalid or missing LINE Channel Access Token
- WHEN the system starts the LINE channel adapter
- THEN the adapter MUST return an error and MUST NOT register with the channel manager
- AND an error MUST be logged at the `channel` component

---

## 4. Matrix Channel

### Requirement: Room Message Processing

The Matrix adapter MUST join configured rooms and process incoming text events.

### Requirement: Homeserver Connectivity

If the homeserver is unreachable at startup, the adapter MUST retry with exponential backoff before failing.

#### Scenario: Inbound Matrix room message processed

- GIVEN a valid Matrix access token and reachable homeserver
- WHEN a user posts a text message in a joined room
- THEN the agent MUST process the message
- AND the response MUST be sent as a `m.room.message` event in the same room

#### Scenario: Homeserver unreachable — retry with backoff

- GIVEN a Matrix homeserver URL that is unreachable
- WHEN the adapter attempts to connect at startup
- THEN the adapter MUST retry with exponential backoff (minimum 3 attempts)
- AND after exhausting retries, MUST log an error and stop the adapter gracefully

---

## 5. Nostr Channel

### Requirement: NIP-04 DM Decryption and Processing

The Nostr adapter MUST decrypt NIP-04 DMs addressed to the configured public key, process them through the agent, and respond with an encrypted NIP-04 DM.

### Requirement: Invalid nsec Blocks Startup

If the configured private key (`nsec`) is invalid, the adapter MUST NOT start.

#### Scenario: Inbound NIP-04 DM processed and response encrypted

- GIVEN a valid `nsec` and connected relay
- WHEN a user sends a NIP-04 encrypted DM to the agent's public key
- THEN the adapter MUST decrypt the message
- AND the agent MUST process it
- AND the response MUST be sent as a NIP-04 encrypted DM to the sender

#### Scenario: Invalid nsec blocks startup

- GIVEN an `nsec` value that fails key parsing
- WHEN the system starts the Nostr adapter
- THEN the adapter MUST return an error and MUST NOT connect to any relay
- AND an error MUST be logged indicating invalid private key

---

## 6. Twilio Voice Channel

### Requirement: Inbound Call TwiML Webhook

The Twilio Voice adapter MUST expose an HTTP webhook endpoint that returns TwiML for inbound calls. The TwiML MUST instruct Twilio to play agent-generated text as audio via TTS.

### Requirement: Twilio Signature Validation

Incoming webhook requests MUST be validated via `X-Twilio-Signature`. Requests failing validation MUST be rejected with HTTP 403.

#### Scenario: Inbound call processed via TwiML

- GIVEN a valid Twilio Account SID and Auth Token in config
- WHEN an inbound call triggers the webhook endpoint
- THEN the adapter MUST pass the caller's speech/input to the agent loop
- AND return a TwiML `<Say>` response with the agent's reply
- AND Twilio MUST play the audio to the caller

#### Scenario: Invalid Twilio signature rejected

- GIVEN a webhook request with a missing or incorrect `X-Twilio-Signature` header
- WHEN the adapter receives the request
- THEN the adapter MUST respond with HTTP 403
- AND MUST NOT process the request through the agent

---

## 7. IRC Channel

### Requirement: Channel Message Processing

The IRC adapter MUST connect to the configured IRC network, join configured channels, and process text messages directed at the bot (by prefix or direct message).

### Requirement: Automatic Reconnection on Kick/Ban

If the bot is kicked from a channel, it MUST attempt to rejoin automatically. If banned, it MUST log the ban and not attempt rejoin.

#### Scenario: IRC channel message processed

- GIVEN a connected IRC bot in channel `#support`
- WHEN a user sends `botname: help me`
- THEN the agent MUST process the message
- AND the response MUST be sent to `#support` as a PRIVMSG

#### Scenario: Bot kicked — automatic rejoin

- GIVEN the bot is connected in channel `#support`
- WHEN the bot receives a KICK event for its own nick
- THEN the adapter MUST attempt to rejoin `#support` after a short delay (SHOULD use exponential backoff)

#### Scenario: Bot banned — no rejoin

- GIVEN the bot attempts to join `#support`
- WHEN the server returns a 474 (banned from channel) response
- THEN the adapter MUST NOT retry joining that channel
- AND MUST log a warning indicating the ban

---

## Coverage Summary

| Domain | Requirements | Happy Path | Edge Cases | Error States |
|--------|-------------|------------|------------|--------------|
| Config Migration | 2 | covered | covered | covered |
| Multi-account Routing | 2 | covered | covered | — |
| LINE | 2 | covered | — | covered |
| Matrix | 2 | covered | — | covered |
| Nostr | 2 | covered | — | covered |
| Twilio Voice | 2 | covered | — | covered |
| IRC | 2 | covered | covered | covered |

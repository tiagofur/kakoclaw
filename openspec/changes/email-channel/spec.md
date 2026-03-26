# Email Channel Specification

## Purpose

Defines the requirements for a two-way email communication channel using IMAP polling (inbound) and SMTP (outbound) with RFC 2822 threading, integrated as MakoClaw's 10th channel.

## Requirements

### Requirement: IMAP Polling Loop

The system MUST poll the configured IMAP mailbox at `poll_interval_seconds` (default 60s) for unseen messages with UID greater than the last-seen UID.

The system MUST persist the last-seen UID to `<workspace>/email_last_uid` so polling survives restarts.

The system MUST use TLS (IMAPS, port 993) by default. The system MAY allow `insecure_skip_verify` for development environments.

#### Scenario: New email arrives and is processed

- GIVEN the email channel is running and polling IMAP every 60s
- WHEN a new unseen email arrives from an allowed sender
- THEN the system fetches the email, publishes an `InboundMessage` to the bus with `Channel: "email"`, `SenderID` set to the sender address, and `SessionKey: "email:<thread-root-message-id>"`
- AND the email is marked as read (if `mark_as_read` is true)
- AND `last_uid` is updated to the new UID

#### Scenario: Poll cycle with no new emails

- GIVEN the IMAP mailbox has no unseen messages with UID > last-seen
- WHEN the poll timer fires
- THEN the system connects, checks, disconnects, and takes no further action

### Requirement: SMTP Reply with Threading

The system MUST send outbound replies via SMTP with proper threading headers: `Message-ID` (generated as `<uuid@makoclaw>`), `In-Reply-To` (last inbound Message-ID), and `References` (full chain).

The system MUST prefix the subject with `Re:` if not already present.

The system MUST send replies as `text/plain`.

#### Scenario: Agent replies to an email thread

- GIVEN an inbound email was processed with `Message-ID: <abc@sender.com>` and `Subject: Help with setup`
- WHEN the agent produces a response
- THEN an SMTP email is sent with `In-Reply-To: <abc@sender.com>`, `References: <abc@sender.com>`, and `Subject: Re: Help with setup`

#### Scenario: Reply within an existing thread

- GIVEN an inbound email has `References: <root@x.com> <mid@x.com>` and `Message-ID: <latest@x.com>`
- WHEN the agent replies
- THEN `References` includes the full chain plus the new Message-ID
- AND `In-Reply-To` is set to `<latest@x.com>`

### Requirement: Allow-List Filtering

The system MUST reject emails from senders not in `allow_from`. Filtering MUST use exact email address matching via `BaseChannel.IsAllowed()`.

The system SHOULD log rejected emails at debug level with the sender address.

#### Scenario: Email from unauthorized sender

- GIVEN `allow_from` is `["boss@company.com"]`
- WHEN an email arrives from `stranger@evil.com`
- THEN the email is silently ignored (not published to bus)
- AND a debug log is emitted: sender not in allow-list

#### Scenario: Email from allowed sender

- GIVEN `allow_from` is `["boss@company.com"]`
- WHEN an email arrives from `boss@company.com`
- THEN the email is processed normally

### Requirement: Session Key from Thread

The system MUST derive the session key as `email:<thread-root-message-id>` where the thread root is the first entry in the `References` header. If `References` is empty, the email's own `Message-ID` MUST be used as thread root.

#### Scenario: Multiple emails in same thread share session

- GIVEN email A has `Message-ID: <root@x.com>` (no References)
- AND email B has `References: <root@x.com>` and `Message-ID: <b@x.com>`
- WHEN both are processed
- THEN both produce `SessionKey: "email:<root@x.com>"`

#### Scenario: New conversation creates new session

- GIVEN an email arrives with no `References` and no `In-Reply-To`
- WHEN processed
- THEN `SessionKey` is `"email:<its-own-message-id>"`

### Requirement: Email Body Extraction

The system MUST prefer the `text/plain` MIME part. If only `text/html` is available, the system MUST strip HTML tags and convert to readable plaintext.

The system MUST detect attachments and append markers: `[attachment: filename.ext (size)]`.

The system SHOULD enforce a configurable `max_email_size_mb` (default 10). Emails exceeding this MUST be skipped with a warning log.

#### Scenario: HTML-only email

- GIVEN an inbound email has only a `text/html` part
- WHEN processed
- THEN HTML tags are stripped and the resulting plaintext is used as `Content`

#### Scenario: Email with attachments

- GIVEN an email has body text and two PDF attachments
- WHEN processed
- THEN `Content` contains the body text followed by `[attachment: report.pdf (2MB)]` and `[attachment: data.pdf (500KB)]`

#### Scenario: Oversized email skipped

- GIVEN `max_email_size_mb` is 10
- WHEN an email of 15MB arrives
- THEN the email is skipped and a warning log is emitted

### Requirement: Connection Resilience

The system MUST handle IMAP connection failures gracefully. On error, the system MUST log the error, wait for the next poll interval, and retry. The system SHOULD use exponential backoff (capped at 5 minutes) for consecutive failures.

#### Scenario: IMAP server temporarily unreachable

- GIVEN the IMAP server is down
- WHEN a poll cycle attempts to connect
- THEN the error is logged and the system retries on the next interval
- AND consecutive failures increase the backoff interval

#### Scenario: IMAP server recovers

- GIVEN the system has been in backoff due to connection errors
- WHEN a poll cycle succeeds
- THEN the backoff resets to the configured `poll_interval_seconds`

### Requirement: Email Channel Configuration

The system MUST accept an `EmailChannelConfig` in `ChannelsConfig` with these fields:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | Yes | false | Enable the channel |
| `imap_host` | string | Yes | — | IMAP server hostname |
| `imap_port` | int | No | 993 | IMAP server port |
| `smtp_host` | string | Yes | — | SMTP server hostname |
| `smtp_port` | int | No | 587 | SMTP server port |
| `username` | string | Yes | — | Auth username (both IMAP and SMTP) |
| `password` | string | Yes | — | Auth password / app-specific password |
| `from` | string | No | username | Sender address for outbound |
| `allow_from` | FlexibleStringSlice | No | [] | Allowed sender addresses |
| `poll_interval_seconds` | int | No | 60 | Seconds between IMAP polls |
| `mailbox` | string | No | "INBOX" | IMAP mailbox to monitor |
| `mark_as_read` | bool | No | true | Mark fetched emails as seen |
| `max_email_size_mb` | int | No | 10 | Skip emails larger than this |
| `insecure_skip_verify` | bool | No | false | Skip TLS cert verification |

#### Scenario: Minimal valid configuration

- GIVEN config has `enabled: true`, `imap_host`, `smtp_host`, `username`, and `password`
- WHEN the channel manager initializes
- THEN the email channel starts with defaults for all optional fields

#### Scenario: Channel disabled by default

- GIVEN no email channel config is present
- WHEN the system starts
- THEN no email channel is initialized and no IMAP connections are made

### Requirement: Channel Manager Integration

The system MUST register `EmailChannel` in `manager.initChannels()` following the existing pattern. The system MUST add email to `RestartChannel()` and include `!c.Email.Enabled` in `isChannelsConfigEmpty()`.

#### Scenario: Email channel registered on startup

- GIVEN email channel config is enabled with valid IMAP/SMTP settings
- WHEN `initChannels()` runs
- THEN an `EmailChannel` instance is created, user resolver applied, and stored in `m.channels["email"]`

### Requirement: Multi-User Isolation

The system MUST work in both single-user and multi-user deployments. In multi-user mode, each user's email channel config comes from their personal `config.json`. UID persistence MUST be per-user (stored in user workspace).

#### Scenario: Two users with different email configs

- GIVEN user A has email channel pointing to `alice@x.com` and user B to `bob@y.com`
- WHEN both are active
- THEN each polls their own IMAP inbox independently with separate UID tracking

## Security Considerations

- Credentials (`password`) MUST NOT be logged at any level.
- The `allow_from` list acts as a security boundary; an empty list with the channel enabled SHOULD log a warning (open to all senders).
- `insecure_skip_verify` SHOULD log a warning when enabled.

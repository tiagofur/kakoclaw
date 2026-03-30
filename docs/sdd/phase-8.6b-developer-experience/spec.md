# Spec: Phase 8.6b — Developer Experience

**Change**: `phase-8.6b-developer-experience`
**Status**: Draft
**Date**: 2026-03-30

---

## Domain: Onboarding Wizard

### Requirement: Interactive setup completes without manual config editing

On a fresh machine, `makoclaw onboard` MUST guide the user through all required configuration steps interactively and write a valid config file without requiring the user to manually edit any file.

#### Scenario: Happy path — fresh machine

- GIVEN no `~/.MakoClaw/config.json` exists
- WHEN the user runs `makoclaw onboard`
- THEN the wizard prompts for: provider selection, API key, workspace path
- AND writes a valid `~/.MakoClaw/config.json` on completion
- AND prints a confirmation message with the config path

#### Scenario: Daemon install on macOS

- GIVEN the user runs `makoclaw onboard --install-daemon` on macOS
- WHEN the wizard completes the provider/key/workspace steps
- THEN a launchd `.plist` file is written to `~/Library/LaunchAgents/`
- AND the output includes the `launchctl load` command to activate it

#### Scenario: Insufficient permissions for daemon install

- GIVEN the user runs `makoclaw onboard --install-daemon`
- AND the process lacks write permission to the daemon target directory
- WHEN the wizard attempts to write the daemon file
- THEN it prints step-by-step manual instructions for installing the daemon
- AND exits with code 0 (non-fatal)

---

## Domain: Doctor Deep Diagnostics

### Requirement: World-readable config detected as ERROR

`makoclaw doctor --deep` MUST check config file permissions and report any file readable by group or others as an ERROR-level finding.

#### Scenario: Config file is world-readable

- GIVEN `~/.MakoClaw/config.json` has permissions `0644`
- WHEN the user runs `makoclaw doctor --deep`
- THEN the output contains an ERROR entry for the config file permission
- AND the entry includes the current permission value and expected value (`0600`)

#### Scenario: Auto-fix with --fix flag

- GIVEN `~/.MakoClaw/config.json` has permissions `0644`
- WHEN the user runs `makoclaw doctor --deep --fix`
- THEN the config file permissions are changed to `0600`
- AND the output reports the fix as applied

#### Scenario: JSON output is valid

- GIVEN any system state
- WHEN the user runs `makoclaw doctor --deep --json`
- THEN stdout contains only valid JSON (parseable by `jq .`)
- AND the JSON is an array of objects each containing at minimum `level` and `message` fields

#### Scenario: Provider endpoint timeout

- GIVEN a configured provider whose base URL is unreachable (timeout > 5s)
- WHEN the user runs `makoclaw doctor --deep`
- THEN the output contains a WARNING entry for that provider
- AND the entry includes the provider name and the timeout duration

---

## Domain: Cron Stagger and Delivery

### Requirement: Jobs stagger deterministically — no simultaneous firing

When 100+ jobs have `stagger: true`, they MUST NOT all fire within the same second. Stagger offset MUST be deterministic (same job ID always gets the same offset).

#### Scenario: 100 jobs with stagger enabled

- GIVEN 100 cron jobs all scheduled for `0 * * * *` (top of hour) with `stagger: true`
- WHEN the scheduler fires at `:00`
- THEN no more than 1 job fires within any given 1-second window
- AND each job's actual fire time equals its scheduled time plus `hash(job.ID) % staggerWindowMS`

#### Scenario: Webhook delivery mode

- GIVEN a cron job with `delivery_mode: webhook` and a `webhook_url` configured
- WHEN the job completes
- THEN the scheduler performs an HTTP POST to `webhook_url`
- AND the request body is valid JSON containing the job result

#### Scenario: Silent delivery mode

- GIVEN a cron job with `delivery_mode: none`
- WHEN the job completes
- THEN no message is sent to any channel or webhook
- AND the job result is only written to the task log

---

## Domain: Exa Search Provider

### Requirement: Exa search returns semantic results when API key is configured

The `web_search` tool MUST use Exa when `providers.exa.api_key` is set and the user invokes `web_search` without specifying a provider override.

#### Scenario: Exa API key present

- GIVEN `providers.exa.api_key` is set in config
- WHEN the agent calls `web_search` with a query
- THEN the request is sent to the Exa API endpoint
- AND semantic search results are returned to the agent

#### Scenario: Exa API key absent — fallback error

- GIVEN `providers.exa.api_key` is NOT set
- AND `providers.brave.api_key` is also NOT set
- WHEN the agent calls `web_search`
- THEN the tool returns an error message
- AND the error message suggests configuring `providers.exa.api_key` or `providers.brave.api_key`

---

## Domain: DeepSeek LLM Provider

### Requirement: DeepSeek routes to its official endpoint

When the model string is `deepseek/*` or the provider is `deepseek`, requests MUST be sent to `api.deepseek.com/v1`.

#### Scenario: Standard DeepSeek model

- GIVEN `providers.deepseek.api_key` is set
- WHEN an agent is configured with model `deepseek/deepseek-chat`
- THEN `CreateProvider` selects the DeepSeek case
- AND the HTTP request is sent to `https://api.deepseek.com/v1/chat/completions`

#### Scenario: LiteLLM proxy with custom base_url

- GIVEN a provider config with `type: litellm` and a `base_url` set to a proxy address
- WHEN `CreateProvider` is called
- THEN the HTTP request is sent to `{base_url}/chat/completions`
- AND the `api_key` from config is used as the Bearer token

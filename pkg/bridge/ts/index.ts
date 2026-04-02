import { createInterface } from "node:readline";
import { readFile } from "node:fs/promises";
import { existsSync, realpathSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { homedir } from "node:os";
import { join, dirname, resolve } from "node:path";
import { query } from "@anthropic-ai/claude-agent-sdk";

// ── Types ────────────────────────────────────────────────────────────────────

interface MCPServerConfig {
  command?: string;
  args?: string[];
  env?: Record<string, string>;
  type?: string;
  url?: string;
  id?: string;
}

interface RequestOptions {
  model?: string;
  cwd?: string;
  system_prompt?: string;
  prompt_injection?: string;
  resume?: string;
  max_turns?: number;
  permission_mode?: string;
  mcp_servers?: Record<string, MCPServerConfig>;
  allowed_tools?: string[];
  continue?: boolean;
  agents?: Record<string, unknown>;
  no_user_settings?: boolean;
  disabled_tools?: string[];
}

// ── Cloud MCP (claude.ai) ───────────────────────────────────────────────────

interface CloudMCPServer {
  type: "claudeai-proxy";
  url: string;
  id: string;
}

interface CloudMCPCache {
  servers: Record<string, CloudMCPServer>;
  expiresAt: number;
}

const CLOUD_MCP_CACHE_TTL_MS = 5 * 60 * 1000;
const CLOUD_MCP_API_TIMEOUT_MS = 5000;
const CLOUD_MCP_PROXY_BASE = "https://mcp-proxy.anthropic.com/v1/mcp";
const CLOUD_MCP_API_URL = "https://api.anthropic.com/v1/mcp_servers?limit=1000";
const CLOUD_MCP_BETA_HEADER = "mcp-servers-2025-12-04";

let cloudMcpCache: CloudMCPCache | null = null;

async function loadCloudMCPs(): Promise<Record<string, CloudMCPServer>> {
  if (cloudMcpCache && Date.now() < cloudMcpCache.expiresAt) {
    return cloudMcpCache.servers;
  }
  try {
    const credsPath = join(homedir(), ".claude", ".credentials.json");
    const raw = await readFile(credsPath, "utf8");
    const creds = JSON.parse(raw);
    const oauth = creds?.claudeAiOauth;
    if (!oauth?.accessToken) return {};
    if (!oauth.scopes?.includes("user:mcp_servers")) return {};
    if (oauth.expiresAt && Date.now() > oauth.expiresAt) return {};

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), CLOUD_MCP_API_TIMEOUT_MS);
    const resp = await fetch(CLOUD_MCP_API_URL, {
      headers: {
        Authorization: `Bearer ${oauth.accessToken}`,
        "Content-Type": "application/json",
        "anthropic-beta": CLOUD_MCP_BETA_HEADER,
        "anthropic-version": "2023-06-01",
      },
      signal: controller.signal,
    });
    clearTimeout(timeout);
    if (!resp.ok) return {};

    const body = (await resp.json()) as { data: Array<{ id: string; display_name: string }> };
    const servers: Record<string, CloudMCPServer> = {};
    for (const srv of body.data) {
      const name = `claude.ai ${srv.display_name}`;
      servers[name] = { type: "claudeai-proxy", url: `${CLOUD_MCP_PROXY_BASE}/${srv.id}`, id: srv.id };
    }
    cloudMcpCache = { servers, expiresAt: Date.now() + CLOUD_MCP_CACHE_TTL_MS };
    return servers;
  } catch {
    return {};
  }
}

interface Request {
  command: string;
  prompt: string;
  request_id?: string;
  options?: RequestOptions;
}

interface OutEvent {
  event: string;
  request_id?: string;
  [key: string]: unknown;
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function emit(obj: OutEvent): void {
  process.stdout.write(JSON.stringify(obj) + "\n");
}

function log(msg: string): void {
  process.stderr.write(`[bridge] ${msg}\n`);
}

// ── Claude Code Executable Resolution ──────────────────────────────────────
// The Agent SDK uses pathToClaudeCodeExecutable to find Claude Code.
// It supports both native binaries and cli.js files (auto-detected by extension).
// When the bundle runs outside the Claude Code package directory, we resolve
// the executable path ourselves.

let cachedExePath: string | null | undefined;

function findClaudeCodeExecutable(): string | null {
  if (cachedExePath !== undefined) return cachedExePath;

  // Method 1: Find the 'claude' binary via which (preferred — uses the
  // installed native binary which has its own auth and is usually up to date)
  try {
    const claudeBin = execFileSync("which", ["claude"], { encoding: "utf8" }).trim();
    if (claudeBin && existsSync(claudeBin)) {
      cachedExePath = realpathSync(claudeBin);
      return cachedExePath;
    }
  } catch { /* which not available or claude not found */ }

  // Method 2: Check common global npm paths for cli.js fallback
  const globalPrefixes = [
    "/opt/node22/lib/node_modules",
    "/usr/local/lib/node_modules",
    "/usr/lib/node_modules",
    join(homedir(), ".npm/lib/node_modules"),
    join(homedir(), "node_modules"),
  ];
  for (const prefix of globalPrefixes) {
    const candidate = join(prefix, "@anthropic-ai/claude-code/cli.js");
    if (existsSync(candidate)) { cachedExePath = candidate; return cachedExePath; }
  }

  cachedExePath = null;
  return null;
}

async function buildSDKOptions(opts: RequestOptions | undefined) {
  if (!opts) return {};
  const sdkOpts: Record<string, unknown> = {};
  if (opts.model) sdkOpts.model = opts.model;
  if (opts.cwd) sdkOpts.cwd = opts.cwd;
  if (opts.system_prompt) sdkOpts.systemPrompt = opts.system_prompt;
  if (opts.resume) sdkOpts.resume = opts.resume;
  if (opts.continue) sdkOpts.continue = opts.continue;
  if (opts.agents && Object.keys(opts.agents).length > 0) sdkOpts.agents = opts.agents;
  if (opts.max_turns) sdkOpts.maxTurns = opts.max_turns;
  if (opts.permission_mode) {
    sdkOpts.permissionMode = opts.permission_mode;
    if (opts.permission_mode === "bypassPermissions") sdkOpts.allowDangerouslySkipPermissions = true;
  }
  if (opts.allowed_tools) sdkOpts.allowedTools = opts.allowed_tools;
  if (opts.no_user_settings) sdkOpts.settingSources = [];
  else sdkOpts.settingSources = ["user", "project", "local"];

  const cloudServers = opts.no_user_settings ? {} : await loadCloudMCPs();
  const agentServers = opts.mcp_servers ?? {};
  const merged = { ...cloudServers, ...agentServers };
  if (Object.keys(merged).length > 0) sdkOpts.mcpServers = merged;
  if (opts.disabled_tools && opts.disabled_tools.length > 0) sdkOpts.disallowedTools = opts.disabled_tools;

  // Resolve pathToClaudeCodeExecutable so the SDK can find Claude Code
  // even when the bundle runs outside the Claude Code package directory.
  const exePath = findClaudeCodeExecutable();
  if (exePath) {
    sdkOpts.pathToClaudeCodeExecutable = exePath;
    log(`Resolved Claude Code executable: ${exePath}`);
  } else {
    log("WARNING: Could not find Claude Code executable — query() may fail");
  }

  // Prevent "cannot be launched inside another Claude Code session" error.
  // When the bridge runs inside a Claude Code session (e.g. Dev Studio launched
  // from within Claude Code), the CLAUDECODE env var triggers nested session
  // detection. We strip it so the spawned Claude Code runs independently.
  if (!sdkOpts.env) {
    const cleanEnv = { ...process.env };
    delete cleanEnv.CLAUDECODE;
    delete cleanEnv.CLAUDE_CODE_SESSION;
    sdkOpts.env = cleanEnv;
  }

  return sdkOpts;
}

function extractText(content: unknown): string {
  if (!Array.isArray(content)) return "";
  return content
    .filter((block: unknown) => typeof block === "object" && block !== null && "type" in block && (block as Record<string, unknown>).type === "text" && "text" in block)
    .map((block: unknown) => (block as Record<string, string>).text)
    .join("");
}

async function handleQuery(req: Request): Promise<void> {
  const reqId = req.request_id || "";
  const emitReq = (obj: OutEvent) => emit({ ...obj, request_id: reqId });
  const sdkOptions = await buildSDKOptions(req.options);

  // Prepend semantic memory injection from devmemory if provided by the Go handler
  const effectivePrompt = req.options?.prompt_injection
    ? `${req.options.prompt_injection}\n\n---\n\n${req.prompt}`
    : req.prompt;

  log(`query start — rid=${reqId} model=${sdkOptions.model ?? "default"} prompt="${effectivePrompt.slice(0, 80)}..."`);
  const timeoutMs = 10 * 60 * 1000;
  const timeout = setTimeout(() => {
    emitReq({ event: "error", message: "query timeout: no result after 10 minutes" });
  }, timeoutMs);

  try {
    const stream = query({ prompt: effectivePrompt, options: sdkOptions as Parameters<typeof query>[0]["options"] });
    for await (const message of stream) {
      const msg = message as Record<string, unknown>;
      const msgType = msg.type as string | undefined;

      switch (msgType) {
        case "system":
          emitReq({ event: "system", session_id: msg.session_id as string, tools: msg.tools as string[], model: msg.model as string });
          break;
        case "assistant":
          const inner = msg.message as Record<string, unknown> | undefined;
          if (inner?.content && Array.isArray(inner.content)) {
            const text = extractText(inner.content);
            if (text) emitReq({ event: "assistant", text });
            for (const block of inner.content as Record<string, unknown>[]) {
              if (block.type === "tool_use") emitReq({ event: "tool_use", id: block.id as string, name: block.name as string, input: block.input as Record<string, unknown> });
            }
          }
          break;
        case "tool_use_summary":
          emitReq({ event: "tool_result", content: msg.summary as string });
          break;
        case "result":
          const subtype = msg.subtype as string | undefined;
          if (subtype === "success") emitReq({ event: "result", content: msg.result as string, cost_usd: msg.total_cost_usd as number, session_id: msg.session_id as string, duration_ms: msg.duration_ms as number, num_turns: msg.num_turns as number });
          else emitReq({ event: "error", message: (msg.errors as string[] | undefined)?.join("; ") ?? `result error: ${subtype}`, subtype: subtype ?? "unknown" });
          break;
      }
    }
  } catch (err: unknown) {
    const errMsg = err instanceof Error ? err.message : String(err);
    emitReq({ event: "error", message: errMsg });
  } finally {
    clearTimeout(timeout);
  }
}

async function handleRequest(line: string): Promise<void> {
  let req: Request;
  try { req = JSON.parse(line) as Request; } catch { emit({ event: "error", message: `invalid JSON: ${line.slice(0, 200)}` }); return; }
  if (!req.command) { emit({ event: "error", request_id: req.request_id || "", message: "missing 'command'" }); return; }
  const reqId = req.request_id || "";
  switch (req.command) {
    case "query":
      if (!req.prompt) { emit({ event: "error", request_id: reqId, message: "missing 'prompt'" }); return; }
      await handleQuery(req);
      break;
    case "ping":
      emit({ event: "pong", request_id: reqId });
      break;
    default:
      emit({ event: "error", request_id: reqId, message: `unknown command: ${req.command}` });
  }
}

function main(): void {
  const rl = createInterface({ input: process.stdin, terminal: false });
  rl.on("line", (line: string) => { const trimmed = line.trim(); if (trimmed) handleRequest(trimmed).catch((err: unknown) => emit({ event: "error", message: `internal bridge error: ${err instanceof Error ? err.message : String(err)}` })); });
  rl.on("close", () => process.exit(0));
  process.on("uncaughtException", (err: Error) => { emit({ event: "error", message: `uncaught exception: ${err.message}` }); process.exit(1); });
}

main();

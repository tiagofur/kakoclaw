import { createInterface } from "node:readline";
import { spawn } from "node:child_process";

// ── Types ────────────────────────────────────────────────────────────────────
interface RequestOptions {
  cwd?: string;
  model?: string;
  system_prompt?: string;
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
  process.stderr.write(`[bridge:opencode] ${msg}\n`);
}

async function handleQuery(req: Request): Promise<void> {
  const reqId = req.request_id || "";
  const emitReq = (obj: OutEvent) => emit({ ...obj, request_id: reqId });
  const cwd = req.options?.cwd || process.cwd();

  log(`query start — rid=${reqId} prompt="${req.prompt.slice(0, 80)}..."`);

  // Start OpenCode CLI as subprocess for this specific query
  const args = ["-p", req.prompt, "--non-interactive"];
  if (req.options?.model) {
      args.push("-m", req.options.model);
  }

  const child = spawn("opencode", args, {
      cwd,
      stdio: ["ignore", "pipe", "pipe"]
  });

  emitReq({ event: "system", session_id: `opencode-${reqId}`, tools: [], model: req.options?.model || "opencode-default" });

  if (child.stdout) {
      const rl = createInterface({ input: child.stdout, terminal: false });
      rl.on("line", (line) => {
          if (line.trim()) {
              // Echo stdout as assistant text
              emitReq({ event: "assistant", text: line + "\n" });
          }
      });
  }

  if (child.stderr) {
      const rl = createInterface({ input: child.stderr, terminal: false });
      rl.on("line", (line) => {
          if (line.trim()) {
              log(`subprocess stderr: ${line}`);
          }
      });
  }

  return new Promise<void>((resolve) => {
      child.on("close", (code) => {
          if (code === 0) {
              emitReq({ event: "result", content: "done", duration_ms: 0, cost_usd: 0, num_turns: 1, session_id: `opencode-${reqId}` });
          } else {
              emitReq({ event: "error", message: `opencode exited with code ${code}` });
          }
          resolve();
      });
  });
}

async function handleRequest(line: string): Promise<void> {
  let req: Request;
  try { req = JSON.parse(line) as Request; } catch { emit({ event: "error", message: `invalid JSON` }); return; }
  const reqId = req.request_id || "";
  if (!req.command) { emit({ event: "error", request_id: reqId, message: "missing 'command'" }); return; }
  
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
}

main();

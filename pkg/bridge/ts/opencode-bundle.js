// opencode.ts
import { createInterface } from "node:readline";
import { spawn } from "node:child_process";
var QUERY_TIMEOUT_MS = 10 * 60 * 1e3;
var AUTH_KEYWORDS = ["authentication", "login", "unauthorized", "unauthenticated", "not logged in", "api key", "auth"];
function isAuthError(msg) {
  const lower = msg.toLowerCase();
  return AUTH_KEYWORDS.some((kw) => lower.includes(kw));
}
function emit(obj) {
  process.stdout.write(JSON.stringify(obj) + "\n");
}
function log(msg) {
  process.stderr.write(`[bridge:opencode] ${msg}
`);
}
async function handleQuery(req) {
  const reqId = req.request_id || "";
  const emitReq = (obj) => emit({ ...obj, request_id: reqId });
  const cwd = req.options?.cwd || process.cwd();
  let effectivePrompt = req.prompt;
  if (req.options?.prompt_injection) {
    effectivePrompt = `${req.options.prompt_injection}

---

${effectivePrompt}`;
  }
  if (req.options?.system_prompt) {
    effectivePrompt = `SYSTEM: ${req.options.system_prompt}

---

${effectivePrompt}`;
  }
  log(`query start \u2014 rid=${reqId} prompt="${effectivePrompt.slice(0, 80)}..."`);
  const args = ["-p", effectivePrompt, "--non-interactive"];
  if (req.options?.model) {
    args.push("-m", req.options.model);
  }
  const child = spawn("opencode", args, {
    cwd,
    stdio: ["ignore", "pipe", "pipe"]
  });
  emitReq({ event: "system", session_id: `opencode-${reqId}`, tools: [], model: req.options?.model || "opencode-default" });
  const stderrLines = [];
  if (child.stdout) {
    const rl = createInterface({ input: child.stdout, terminal: false });
    rl.on("line", (line) => {
      if (line.trim()) {
        emitReq({ event: "assistant", text: line + "\n" });
      }
    });
  }
  if (child.stderr) {
    const rl = createInterface({ input: child.stderr, terminal: false });
    rl.on("line", (line) => {
      if (line.trim()) {
        log(`subprocess stderr: ${line}`);
        stderrLines.push(line);
      }
    });
  }
  const timeoutHandle = setTimeout(() => {
    log("query timeout \u2014 killing child process");
    child.kill("SIGTERM");
    emitReq({ event: "error", message: "query timeout: no result after 10 minutes" });
  }, QUERY_TIMEOUT_MS);
  return new Promise((resolve) => {
    child.on("close", (code) => {
      clearTimeout(timeoutHandle);
      if (code === 0) {
        emitReq({ event: "result", content: "done", duration_ms: 0, cost_usd: 0, num_turns: 1, session_id: `opencode-${reqId}` });
      } else {
        const stderrSummary = stderrLines.slice(-10).join(" | ");
        let errMsg = `opencode exited with code ${code}`;
        if (stderrSummary) errMsg += `
Process output: ${stderrSummary}`;
        if (isAuthError(errMsg)) errMsg += "\n\nHint: Run `opencode auth login` in your terminal to authenticate OpenCode.";
        emitReq({ event: "error", message: errMsg });
      }
      resolve();
    });
  });
}
async function handleRequest(line) {
  let req;
  try {
    req = JSON.parse(line);
  } catch {
    emit({ event: "error", message: `invalid JSON` });
    return;
  }
  const reqId = req.request_id || "";
  if (!req.command) {
    emit({ event: "error", request_id: reqId, message: "missing 'command'" });
    return;
  }
  switch (req.command) {
    case "query":
      if (!req.prompt) {
        emit({ event: "error", request_id: reqId, message: "missing 'prompt'" });
        return;
      }
      await handleQuery(req);
      break;
    case "ping":
      emit({ event: "pong", request_id: reqId });
      break;
    default:
      emit({ event: "error", request_id: reqId, message: `unknown command: ${req.command}` });
  }
}
function main() {
  const rl = createInterface({ input: process.stdin, terminal: false });
  rl.on("line", (line) => {
    const trimmed = line.trim();
    if (trimmed) handleRequest(trimmed).catch((err) => emit({ event: "error", message: `internal bridge error: ${err instanceof Error ? err.message : String(err)}` }));
  });
  rl.on("close", () => process.exit(0));
}
main();

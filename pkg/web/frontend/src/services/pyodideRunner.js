/**
 * Pyodide Runner Service
 * Lazy-loads Pyodide from CDN and executes Python code in the browser.
 */

const PYODIDE_CDN = 'https://cdn.jsdelivr.net/pyodide/v0.27.5/full/'
const EXECUTION_TIMEOUT = 30000 // 30 seconds

let pyodideInstance = null
let loadingPromise = null

/** Current status: 'idle' | 'loading' | 'ready' | 'running' | 'error' */
let status = 'idle'

/** Status change listeners */
const listeners = new Set()

function setStatus(newStatus) {
  status = newStatus
  listeners.forEach(fn => fn(status))
}

/**
 * Subscribe to status changes.
 * @param {function} fn - Callback receiving the new status string.
 * @returns {function} Unsubscribe function.
 */
export function onStatusChange(fn) {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

/** @returns {'idle'|'loading'|'ready'|'running'|'error'} */
export function getStatus() {
  return status
}

/**
 * Lazy-load the Pyodide runtime (singleton).
 * First call downloads ~15 MB of WASM; subsequent calls return immediately.
 */
export async function ensureLoaded() {
  if (pyodideInstance) return pyodideInstance

  if (loadingPromise) return loadingPromise

  setStatus('loading')

  loadingPromise = (async () => {
    try {
      // Dynamically import the Pyodide loader from CDN
      const { loadPyodide } = await import(/* @vite-ignore */ `${PYODIDE_CDN}pyodide.mjs`)
      pyodideInstance = await loadPyodide({
        indexURL: PYODIDE_CDN
      })
      setStatus('ready')
      return pyodideInstance
    } catch (err) {
      loadingPromise = null
      setStatus('error')
      throw err
    }
  })()

  return loadingPromise
}

/**
 * Run Python code and capture stdout/stderr.
 * @param {string} code - Python source code to execute.
 * @returns {Promise<{output: string, error: string, duration: number}>}
 */
export async function runPython(code) {
  const pyodide = await ensureLoaded()

  setStatus('running')
  const start = performance.now()

  // Redirect stdout and stderr to StringIO so we can capture them
  pyodide.runPython(`
import sys, io
__stdout_capture = io.StringIO()
__stderr_capture = io.StringIO()
sys.stdout = __stdout_capture
sys.stderr = __stderr_capture
`)

  let output = ''
  let error = ''

  try {
    // Wrap execution in a timeout
    const result = await Promise.race([
      (async () => {
        try {
          pyodide.runPython(code)
        } catch (pyErr) {
          // Python runtime errors (SyntaxError, NameError, etc.)
          error = String(pyErr)
        }
      })(),
      new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Execution timed out (30s limit)')), EXECUTION_TIMEOUT)
      )
    ])

    output = pyodide.runPython('__stdout_capture.getvalue()')
    if (!error) {
      error = pyodide.runPython('__stderr_capture.getvalue()')
    }
  } catch (timeoutErr) {
    error = timeoutErr.message
  } finally {
    // Restore original stdout/stderr
    try {
      pyodide.runPython(`
import sys
sys.stdout = sys.__stdout__
sys.stderr = sys.__stderr__
`)
    } catch (_) { /* best effort */ }

    setStatus('ready')
  }

  const duration = Math.round(performance.now() - start)
  return { output, error, duration }
}

export default { ensureLoaded, runPython, getStatus, onStatusChange }

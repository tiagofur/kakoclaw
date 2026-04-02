import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

const MAX_WS_FAILURES = 3

export const useDevStudioStore = defineStore('devStudio', () => {
  const projects = ref([])
  const currentProject = ref(null)
  const bridgeStatus = ref('idle')
  const terminalHistory = ref([])
  const searchResults = ref([])
  const usingHttpFallback = ref(false)
  const projectsRoot = ref('')
  const bridgeError = ref('')
  const bridgeBackend = ref('')
  const sessionId = ref('')
  const sessionTokens = ref(0)
  const sessionCostUsd = ref(0)
  const tokenLimit = ref(0)
  const numTurns = ref(0)

  let ws = null
  let wsFailures = 0

  function applySessionStats(stats = {}) {
    sessionId.value = stats.session_id || ''
    sessionTokens.value = Number(stats.tokens_used || 0)
    sessionCostUsd.value = Number(stats.cost_usd || 0)
    tokenLimit.value = Number(stats.token_limit || 0)
    numTurns.value = Number(stats.num_turns || 0)
  }

  async function fetchSessionStats() {
    try {
      const { data } = await api.get('/dev/session/stats')
      applySessionStats(data)
    } catch (e) {
      console.error('Failed to fetch session stats', e)
    }
  }

  function processTerminalEvent(msg) {
    if (!msg || msg.type === 'ping') return

    if (msg.type === 'result') {
      sessionTokens.value += Number(msg.tokens_used || msg.num_turns || 0)
      sessionCostUsd.value += Number(msg.cost_usd || 0)
      numTurns.value += Number(msg.num_turns || msg.tokens_used || 0)
      if (msg.session_id) {
        sessionId.value = msg.session_id
      }
    }

    if (msg.type === 'session_reset') {
      sessionTokens.value = 0
      sessionCostUsd.value = 0
      numTurns.value = 0
      sessionId.value = msg.new_session_id || ''
      terminalHistory.value.push({
        type: 'session_reset',
        message: msg.message || 'Session auto-reset: token limit reached'
      })
      return
    }

    terminalHistory.value.push(msg)
  }

  async function fetchProjects() {
    try {
      const { data } = await api.get('/dev/projects')
      projects.value = data.projects || []
      projectsRoot.value = data.projects_dir || data.root || ''
    } catch (e) {
      console.error('Failed to fetch projects', e)
      terminalHistory.value.push({ type: 'error', message: 'Failed to load projects. Is Dev Studio enabled in config?' })
    }
  }

  async function createProject(name, gitInit = true) {
    try {
      await api.post('/dev/projects', { name, git_init: gitInit })
      await fetchProjects()
    } catch (e) {
      console.error('Failed to create project', e)
      const msg = e.response?.data || e.message || 'Unknown error creating project'
      throw new Error(msg)
    }
  }

  async function startBridge(projectDir, backend) {
    bridgeError.value = ''
    try {
      const body = { project_dir: projectDir }
      if (backend) body.backend = backend
      const { data } = await api.post('/dev/bridge/start', body)
      currentProject.value = projectDir
      bridgeStatus.value = 'running'
      bridgeBackend.value = data?.backend || ''
      wsFailures = 0
      usingHttpFallback.value = false
      connectTerminal()
      await fetchSessionStats()
      // Show system message confirming active project and backend
      const backendLabel = bridgeBackend.value === 'opencode' ? 'OpenCode' : 'Claude Code'
      terminalHistory.value.push({
        type: 'system',
        message: `Bridge started | Backend: ${backendLabel} | Working in: ${projectDir}`
      })
    } catch (e) {
      console.error('Failed to start bridge', e)
      // Extract detailed error message from JSON response
      const respData = e.response?.data
      let msg = 'Unknown error starting bridge'
      if (typeof respData === 'object' && respData?.error) {
        msg = respData.error
      } else if (typeof respData === 'string') {
        msg = respData
      } else if (e.message) {
        msg = e.message
      }
      bridgeError.value = msg
      terminalHistory.value.push({ type: 'error', message: `Bridge start failed: ${msg}` })
    }
  }

  async function changeBackend(newBackend) {
    try {
      await api.post('/me/config/update', { dev_studio: { default_backend: newBackend } })
      bridgeBackend.value = newBackend
      const backendLabel = newBackend === 'opencode' ? 'OpenCode' : 'Claude Code'
      if (bridgeStatus.value === 'running' && currentProject.value) {
        terminalHistory.value.push({ type: 'system', message: `Switching backend to ${backendLabel}...` })
        await stopBridge()
        await startBridge(currentProject.value, newBackend)
      } else {
        terminalHistory.value.push({ type: 'system', message: `Backend changed to ${backendLabel}. Will apply on next bridge start.` })
      }
    } catch (e) {
      console.error('Failed to change backend', e)
      terminalHistory.value.push({ type: 'error', message: `Failed to change backend: ${e.message}` })
    }
  }

  async function stopBridge() {
    try {
      await api.post('/dev/bridge/stop')
      bridgeStatus.value = 'stopped'
      bridgeBackend.value = ''
      usingHttpFallback.value = false
      wsFailures = 0
      if (ws) {
        ws.close()
        ws = null
      }
      applySessionStats()
    } catch (e) {
      console.error('Failed to stop bridge', e)
    }
  }

  async function checkStatus() {
    try {
      const { data } = await api.get('/dev/bridge/status')
      bridgeStatus.value = data.status
      if (data.backend) {
        bridgeBackend.value = data.backend
      }
      // Restore last active project from state.json (returned by backend)
      if (data.project_dir && !currentProject.value) {
        currentProject.value = data.project_dir
      }
      await fetchSessionStats()
      if (data.status === 'running' && !ws && !usingHttpFallback.value) {
        connectTerminal()
      }
    } catch (e) {
      // Status check can fail if dev studio is disabled — that's expected.
      // Keep status as 'idle' and let the user see the hint in the terminal.
      console.error('Failed to check status', e)
    }
  }

  function connectTerminal() {
    if (ws || usingHttpFallback.value) return
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const port = window.location.port ? ':' + window.location.port : ''
    const token = localStorage.getItem('auth.token')

    ws = new WebSocket(`${protocol}//${window.location.hostname}${port}/ws/dev/terminal?token=${token}`)

    ws.onopen = () => {
      wsFailures = 0
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        processTerminalEvent(msg)
      } catch (e) {
        console.error('Failed to parse WS msg', e)
      }
    }

    ws.onerror = () => {
      wsFailures++
    }

    ws.onclose = () => {
      ws = null
      if (wsFailures >= MAX_WS_FAILURES) {
        usingHttpFallback.value = true
        // Bridge is still running — keep status so the input stays enabled
      }
      // Don't change bridgeStatus to 'stopped' on WS close —
      // the bridge process may still be alive. The WS is just the
      // transport layer; losing it doesn't mean the bridge died.
      // The user can still use HTTP fallback or reconnect.
    }
  }

  async function sendPromptViaHttp(message) {
    const token = localStorage.getItem('auth.token')
    let resp
    try {
      resp = await fetch('/api/v1/dev/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ message, session_id: sessionId.value || undefined })
      })
    } catch (e) {
      terminalHistory.value.push({ type: 'error', message: 'HTTP fallback request failed: ' + e.message })
      return
    }
    if (!resp.ok) {
      const errText = await resp.text().catch(() => `${resp.status}`)
      terminalHistory.value.push({ type: 'error', message: `HTTP error: ${errText}` })
      return
    }
    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buf = ''
    let streamDone = false
    while (!streamDone) {
      const { done, value } = await reader.read()
      if (done) {
        streamDone = true
        break
      }
      buf += decoder.decode(value, { stream: true })
      const lines = buf.split('\n')
      buf = lines.pop() // keep incomplete last line
      for (const line of lines) {
        if (!line.trim()) continue
        try {
          const msg = JSON.parse(line)
          processTerminalEvent(msg)
        } catch (e) {
          // Ignore parse errors for incomplete JSON lines in the stream
        }
      }
    }
  }

  function sendPrompt(prompt) {
    terminalHistory.value.push({ type: 'user', message: prompt })

    // If bridge is not running, try HTTP fallback anyway — the backend
    // will auto-start the bridge if possible.
    if (usingHttpFallback.value || (!ws || ws.readyState !== WebSocket.OPEN)) {
      sendPromptViaHttp(prompt)
      return
    }
    ws.send(JSON.stringify({ type: 'prompt', message: prompt, session_id: sessionId.value || undefined }))
  }

  async function searchMemory(query, limit = 5) {
    try {
      const { data } = await api.post('/dev/memory/search', { query, limit })
      searchResults.value = data.results || []
    } catch (e) {
      console.error('Failed to search memory', e)
    }
  }

  async function fetchHistory(projectName) {
    try {
      const sessionID = `dev_studio_${projectName}`
      const { data } = await api.get(`/chat/sessions/${sessionID}`)
      if (data && data.messages) {
        terminalHistory.value = data.messages.map(m => ({
          type: m.role === 'user' ? 'user' : 'assistant',
          message: m.content
        }))
      }
    } catch (e) {
      // Session might not exist yet, that's fine
      terminalHistory.value = []
    }
  }

  return {
    projects,
    projectsRoot,
    currentProject,
    bridgeStatus,
    terminalHistory,
    searchResults,
    usingHttpFallback,
    bridgeError,
    bridgeBackend,
    sessionId,
    sessionTokens,
    sessionCostUsd,
    tokenLimit,
    numTurns,
    fetchSessionStats,
    processTerminalEvent,
    fetchProjects,
    createProject,
    startBridge,
    stopBridge,
    checkStatus,
    sendPrompt,
    searchMemory,
    fetchHistory,
    changeBackend
  }
})

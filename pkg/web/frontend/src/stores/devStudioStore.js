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

  let ws = null
  let wsFailures = 0

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

  async function startBridge(projectDir) {
    bridgeError.value = ''
    try {
      await api.post('/dev/bridge/start', { project_dir: projectDir })
      currentProject.value = projectDir
      bridgeStatus.value = 'running'
      wsFailures = 0
      usingHttpFallback.value = false
      connectTerminal()
    } catch (e) {
      console.error('Failed to start bridge', e)
      const msg = e.response?.data || e.message || 'Unknown error starting bridge'
      bridgeError.value = typeof msg === 'string' ? msg : JSON.stringify(msg)
      terminalHistory.value.push({ type: 'error', message: `Bridge start failed: ${bridgeError.value}` })
    }
  }

  async function stopBridge() {
    try {
      await api.post('/dev/bridge/stop')
      bridgeStatus.value = 'stopped'
      usingHttpFallback.value = false
      wsFailures = 0
      if (ws) {
        ws.close()
        ws = null
      }
    } catch (e) {
      console.error('Failed to stop bridge', e)
    }
  }

  async function checkStatus() {
    try {
      const { data } = await api.get('/dev/bridge/status')
      bridgeStatus.value = data.status
      // Restore last active project from state.json (returned by backend)
      if (data.project_dir && !currentProject.value) {
        currentProject.value = data.project_dir
      }
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
        if (msg.type !== 'ping') {
          terminalHistory.value.push(msg)
        }
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
        body: JSON.stringify({ message })
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
          if (msg.type !== 'ping') terminalHistory.value.push(msg)
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
    ws.send(JSON.stringify({ type: 'prompt', message: prompt }))
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
    fetchProjects,
    createProject,
    startBridge,
    stopBridge,
    checkStatus,
    sendPrompt,
    searchMemory,
    fetchHistory
  }
})

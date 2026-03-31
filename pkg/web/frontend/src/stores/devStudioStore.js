import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

const MAX_WS_FAILURES = 3

export const useDevStudioStore = defineStore('devStudio', () => {
  const projects = ref([])
  const currentProject = ref(null)
  const bridgeStatus = ref('idle')
  const terminalHistory = ref([])
  const searchResults = ref([])
  const usingHttpFallback = ref(false)

  let ws = null
  let wsFailures = 0

  async function fetchProjects() {
    try {
      const { data } = await axios.get('/api/v1/dev/projects')
      projects.value = data.projects || []
    } catch (e) {
      console.error('Failed to fetch projects', e)
    }
  }

  async function startBridge(projectDir) {
    try {
      await axios.post('/api/v1/dev/bridge/start', { project_dir: projectDir })
      currentProject.value = projectDir
      bridgeStatus.value = 'running'
      wsFailures = 0
      usingHttpFallback.value = false
      connectTerminal()
    } catch (e) {
      console.error('Failed to start bridge', e)
    }
  }

  async function stopBridge() {
    try {
      await axios.post('/api/v1/dev/bridge/stop')
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
      const { data } = await axios.get('/api/v1/dev/bridge/status')
      bridgeStatus.value = data.status
      if (data.status === 'running' && !ws && !usingHttpFallback.value) {
        connectTerminal()
      }
    } catch (e) {
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
      } else {
        bridgeStatus.value = 'stopped'
      }
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
      terminalHistory.value.push({ type: 'error', error: 'HTTP fallback request failed: ' + e.message })
      return
    }
    if (!resp.ok) {
      terminalHistory.value.push({ type: 'error', error: `HTTP fallback error: ${resp.status}` })
      return
    }
    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buf = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buf += decoder.decode(value, { stream: true })
      const lines = buf.split('\n')
      buf = lines.pop() // keep incomplete last line
      for (const line of lines) {
        if (!line.trim()) continue
        try {
          const msg = JSON.parse(line)
          if (msg.type !== 'ping') terminalHistory.value.push(msg)
        } catch {}
      }
    }
  }

  function sendPrompt(prompt) {
    terminalHistory.value.push({ type: 'user', message: prompt })
    if (usingHttpFallback.value) {
      sendPromptViaHttp(prompt)
      return
    }
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'prompt', message: prompt }))
    }
  }

  async function searchMemory(query, limit = 5) {
    try {
      const { data } = await axios.post('/api/v1/dev/memory/search', { query, limit })
      searchResults.value = data.results || []
    } catch (e) {
      console.error('Failed to search memory', e)
    }
  }

  async function fetchHistory(projectName) {
    try {
      const sessionID = `dev_studio_${projectName}`
      const { data } = await axios.get(`/api/v1/chat/sessions/${sessionID}`)
      if (data && data.messages) {
        terminalHistory.value = data.messages.map(m => ({
          type: m.role === 'user' ? 'user' : 'stdout',
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
    currentProject,
    bridgeStatus,
    terminalHistory,
    searchResults,
    usingHttpFallback,
    fetchProjects,
    startBridge,
    stopBridge,
    checkStatus,
    sendPrompt,
    searchMemory,
    fetchHistory
  }
})

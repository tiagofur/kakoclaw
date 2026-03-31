import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useDevStudioStore = defineStore('devStudio', () => {
  const projects = ref([])
  const currentProject = ref(null)
  const bridgeStatus = ref('idle')
  const terminalHistory = ref([])
  const searchResults = ref([])
  
  let ws = null

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
      connectTerminal()
    } catch (e) {
      console.error('Failed to start bridge', e)
    }
  }

  async function stopBridge() {
    try {
      await axios.post('/api/v1/dev/bridge/stop')
      bridgeStatus.value = 'stopped'
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
      if (data.status === 'running' && !ws) {
        connectTerminal()
      }
    } catch (e) {
      console.error('Failed to check status', e)
    }
  }

  function connectTerminal() {
    if (ws) return
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const port = window.location.port ? ':' + window.location.port : ''
    const token = localStorage.getItem('auth.token')
    
    ws = new WebSocket(`${protocol}//${window.location.hostname}${port}/ws/dev/terminal?token=${token}`)
    
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
    
    ws.onclose = () => {
      ws = null
      bridgeStatus.value = 'stopped'
    }
  }

  function sendPrompt(prompt) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      terminalHistory.value.push({ type: 'user', message: prompt })
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
    fetchProjects,
    startBridge,
    stopBridge,
    checkStatus,
    sendPrompt,
    searchMemory,
    fetchHistory
  }
})

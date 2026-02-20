import client from './api'

export default {
  // Skills
  fetchSkills: async () => {
    const response = await client.get('/skills')
    return response.data
  },

  fetchAvailableSkills: async () => {
    const response = await client.get('/skills', { params: { type: 'available' } })
    return response.data
  },

  viewSkill: async (name) => {
    const response = await client.get(`/skills/${encodeURIComponent(name)}`)
    return response.data
  },

  installSkill: async (repository) => {
    const response = await client.post('/skills/install', { repository }, {
      timeout: 120000 // 2 min for installation
    })
    return response.data
  },

  uninstallSkill: async (name) => {
    const response = await client.delete(`/skills/${encodeURIComponent(name)}`)
    return response.data
  },

  generateSkillDraft: async (payload) => {
    const response = await client.post('/skills/generate', payload, { timeout: 120000 })
    return response.data
  },

  createSkill: async (payload) => {
    const response = await client.post('/skills/create', payload, {
      timeout: 120000 // 2 min for skill creation
    })
    return response.data
  },

  // Cron
  fetchCronJobs: async (includeDisabled = true) => {
    const response = await client.get('/cron', { params: { include_disabled: includeDisabled ? 'true' : 'false' } })
    return response.data
  },

  createCronJob: async (job) => {
    const response = await client.post('/cron', job)
    return response.data
  },

  deleteCronJob: async (id) => {
    const response = await client.delete(`/cron/${id}`)
    return response.data
  },

  toggleCronJob: async (id, enabled) => {
    const response = await client.patch(`/cron/${id}`, { enabled })
    return response.data
  },

  updateCronJob: async (id, data) => {
    const response = await client.put(`/cron/${id}`, data)
    return response.data
  },

  runCronJob: async (id) => {
    const response = await client.post(`/cron/${id}/run`, {}, {
      timeout: 300000 // 5 min for manual job run
    })
    return response.data
  },

  updateConfig: async (config) => {
    const response = await client.post('/config', config)
    return response.data
  },

  // Channels
  fetchChannels: async () => {
    const response = await client.get('/channels')
    return response.data
  },

  // Config
  fetchConfig: async () => {
    const response = await client.get('/config')
    return response.data
  },

  // Files
  fetchFiles: async (path = '') => {
    const response = await client.get(`/files/${path}`)
    return response.data
  },

  downloadFile: (path = '') => {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/')
    const token = localStorage.getItem('auth.token')
    window.open(`/api/v1/files/${encodedPath}?download=true&token=${token}`, '_blank')
  },

  uploadFile: async (path, file) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await client.post(`/files/${path}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000
    })
    return response.data
  },

  // Export
  exportTasks: (format = 'json') => {
    const token = localStorage.getItem('auth.token')
    window.open(`/api/v1/export/tasks?format=${format}&token=${token}`, '_blank')
  },

  exportChat: (sessionId = '') => {
    const params = sessionId ? `session_id=${encodeURIComponent(sessionId)}&` : ''
    const token = localStorage.getItem('auth.token')
    window.open(`/api/v1/export/chat?${params}token=${token}`, '_blank')
  },

  // Import conversations (ChatGPT, Claude, KakoClaw formats)
  importChat: async (data, format = 'auto') => {
    const response = await client.post('/import/chat', { format, data }, {
      timeout: 120000 // 2 min for large imports
    })
    return response.data
  },

  // Fork/branch a conversation at a specific message
  forkChat: async (sessionId, messageId = 0) => {
    const response = await client.post('/chat/fork', {
      session_id: sessionId,
      message_id: messageId
    })
    return response.data
  },

  // Models
  fetchModels: async () => {
    const response = await client.get('/models')
    return response.data
  },

  // Voice transcription
  transcribeAudio: async (audioBlob) => {
    const formData = new FormData()
    formData.append('audio', audioBlob, 'recording.webm')
    const response = await client.post('/voice/transcribe', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000 // 60s for transcription
    })
    return response.data
  },

  // Knowledge Base (RAG)
  fetchKnowledgeDocs: async () => {
    const response = await client.get('/knowledge')
    return response.data
  },

  uploadKnowledgeDoc: async (file) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await client.post('/knowledge', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000
    })
    return response.data
  },

  deleteKnowledgeDoc: async (id) => {
    const response = await client.delete(`/knowledge/${id}`)
    return response.data
  },

  searchKnowledge: async (query) => {
    const response = await client.get('/knowledge/search', { params: { q: query } })
    return response.data
  },

  fetchKnowledgeChunks: async (id) => {
    const response = await client.get(`/knowledge/${id}/chunks`)
    return response.data
  },

  updateKnowledgeChunk: async (chunkId, content) => {
    const response = await client.put(`/knowledge/chunks/${chunkId}`, { content })
    return response.data
  },

  // MCP Servers
  fetchMCPServers: async () => {
    const response = await client.get('/mcp')
    return response.data
  },

  reconnectMCPServer: async (name) => {
    const response = await client.post(`/mcp/${encodeURIComponent(name)}/reconnect`, {}, {
      timeout: 35000 // 35s for MCP reconnect (server has 30s timeout)
    })
    return response.data
  },

  // Observability Metrics
  fetchMetrics: async () => {
    const response = await client.get('/metrics')
    return response.data
  },

  // Tools
  fetchTools: async () => {
    const response = await client.get('/tools')
    return response.data
  },

  // Prompt Templates (F7)
  fetchPrompts: async () => {
    const response = await client.get('/prompts')
    return response.data
  },
  createPrompt: async (prompt) => {
    const response = await client.post('/prompts', prompt)
    return response.data
  },
  updatePrompt: async (id, prompt) => {
    const response = await client.put(`/prompts/${id}`, prompt)
    return response.data
  },
  deletePrompt: async (id) => {
    const response = await client.delete(`/prompts/${id}`)
    return response.data
  },

  // Chat File Attachments (F9)
  uploadChatAttachment: async (file) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await client.post('/chat/attachments', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    return response.data
  }
}


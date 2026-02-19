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
    const response = await client.post('/skills/install', { repository })
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
    const response = await client.post('/skills/create', payload)
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
    const response = await client.post(`/cron/${id}/run`)
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

  // Export
  exportTasks: (format = 'json') => {
    window.open(`/api/v1/export/tasks?format=${format}`, '_blank')
  },

  exportChat: (sessionId = '') => {
    const params = sessionId ? `?session_id=${encodeURIComponent(sessionId)}` : ''
    window.open(`/api/v1/export/chat${params}`, '_blank')
  },

  // Import conversations (ChatGPT, Claude, PicoClaw formats)
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
  }
}

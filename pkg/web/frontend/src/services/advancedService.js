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

  generateSkillConfig: async (prompt) => {
    const response = await client.post('/skills/generate-config', { prompt }, { timeout: 60000 })
    return response.data
  },

  createSkill: async (payload) => {
    const response = await client.post('/skills/create', payload, {
      timeout: 120000 // 2 min for skill creation
    })
    return response.data
  },

  // Skill refinement
  refineSkillDraft: async (draft, feedback) => {
    const response = await client.post('/skills/refine', { draft, feedback }, { timeout: 120000 })
    return response.data
  },

  // Skill security scan
  scanSkill: async (content) => {
    const response = await client.post('/skills/scan', { content }, { timeout: 60000 })
    return response.data
  },

  // Skill analytics
  fetchSkillAnalytics: async () => {
    const response = await client.get('/skills/analytics')
    return response.data
  },

  // Marketplace
  fetchMarketplaceSkills: async ({ category = '', page = 1, sort = '', limit } = {}) => {
    const params = { category, page }
    if (sort) params.sort = sort
    if (limit) params.limit = limit
    const response = await client.get('/marketplace/skills', { params })
    return response.data
  },

  fetchMarketplaceSkillDetail: async (slug) => {
    const response = await client.get(`/marketplace/skills/${encodeURIComponent(slug)}`)
    return response.data
  },

  installMarketplaceSkill: async (slug) => {
    const response = await client.post(`/marketplace/skills/${encodeURIComponent(slug)}/install`, {}, {
      timeout: 30000
    })
    return response.data
  },

  forkSkill: async (slug) => {
    const response = await client.post(`/marketplace/skills/${encodeURIComponent(slug)}/fork`, {})
    return response.data
  },

  submitToMarketplace: async (payload) => {
    const response = await client.post('/marketplace/submit', payload, { timeout: 60000 })
    return response.data
  },

  fetchMySubmissions: async () => {
    const response = await client.get('/marketplace/submissions')
    return response.data
  },

  fetchMarketplaceCategories: async () => {
    const response = await client.get('/marketplace/categories')
    return response.data
  },

  fetchSecurityAlerts: async () => {
    const response = await client.get('/marketplace/security-alerts')
    return response.data
  },

  rateSkill: async (slug, rating, review = '') => {
    const response = await client.post(`/marketplace/skills/${encodeURIComponent(slug)}/rate`, { rating, review })
    return response.data
  },

  fetchSkillRatings: async (slug) => {
    const response = await client.get(`/marketplace/skills/${encodeURIComponent(slug)}/rate`)
    return response.data
  },

  // Bundles
  fetchMarketplaceBundles: async (page = 1) => {
    const response = await client.get('/marketplace/bundles', { params: { page } })
    return response.data
  },

  installBundle: async (slug) => {
    const response = await client.post(`/marketplace/bundles/${encodeURIComponent(slug)}/install`, {}, { timeout: 60000 })
    return response.data
  },

  submitBundle: async (payload) => {
    const response = await client.post('/marketplace/bundles', payload)
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

  // Update user-specific config (not global)
  updateUserConfig: async (config) => {
    const response = await client.post('/me/config/update', config)
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

  // Fetch user-specific config (merged with global)
  fetchUserConfig: async () => {
    const response = await client.get('/me/config')
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

  // Delete file or folder
  deleteFile: async (path) => {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/')
    const response = await client.delete(`/files/${encodedPath}`)
    return response.data
  },

  // Create folder
  createFolder: async (path) => {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/')
    const response = await client.put(`/files/${encodedPath}?mkdir=true`)
    return response.data
  },

  // Update file content
  updateFileContent: async (path, content) => {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/')
    const response = await client.put(`/files/${encodedPath}`, content, {
      headers: { 'Content-Type': 'text/plain' }
    })
    return response.data
  },

  // Generate soul/identity/user context with AI
  generateSoul: async (description) => {
    const response = await client.post('/soul/generate', { description })
    return response.data
  },

  // Rename file or folder
  renameFile: async (path, newName) => {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/')
    const response = await client.patch(`/files/${encodedPath}`, { new_name: newName })
    return response.data
  },

  // Search files
  searchFiles: async (path, query) => {
    const encodedPath = path ? path.split('/').map(encodeURIComponent).join('/') : ''
    const response = await client.get(`/files/${encodedPath}?search=${encodeURIComponent(query)}`)
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

  // Import conversations (ChatGPT, Claude, MakoClaw formats)
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

  // Specialists
  fetchSpecialists: async () => {
    const response = await client.get('/agents/specialists')
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

  // Add new MCP server
  addMCPServer: async (config) => {
    const response = await client.post('/mcp', config)
    return response.data
  },

  // Update MCP server
  updateMCPServer: async (name, config) => {
    const response = await client.put(`/mcp/${encodeURIComponent(name)}`, config)
    return response.data
  },

  // Delete MCP server
  deleteMCPServer: async (name) => {
    const response = await client.delete(`/mcp/${encodeURIComponent(name)}`)
    return response.data
  },

  // Test MCP server connection
  testMCPServer: async (name) => {
    const response = await client.post(`/mcp/${encodeURIComponent(name)}/test`, {}, {
      timeout: 35000 // 35s for MCP test (server has 30s timeout)
    })
    return response.data
  },

  // Reconnect all MCP servers
  reconnectAllMCPServers: async () => {
    const response = await client.post('/mcp/reconnect-all', {}, {
      timeout: 65000 // 65s for reconnect all (server has 60s timeout)
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
  },

  // Reports
  sendReportEmail: async (to, subject, body) => {
    const response = await client.post('/reports/email', { to, subject, body }, {
      timeout: 30000 // 30s for email send
    })
    return response.data
  },

  // AI Services
  fixJsonWithAI: async (content, type = 'generic') => {
    const response = await client.post('/ai/fix-json', { content, type }, {
      timeout: 45000 // 45s for AI processing
    })
    return response.data
  },

  createCronWithAI: async (prompt) => {
    const response = await client.post('/ai/create-cron', { prompt }, {
      timeout: 60000 // 60s for AI generation
    })
    return response.data
  }
}


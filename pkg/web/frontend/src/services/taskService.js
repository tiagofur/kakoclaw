import client from './api'

export default {
  fetchTasks: async (includeArchived = false) => {
    const params = includeArchived ? { include_archived: 'true' } : {}
    const response = await client.get('/tasks', { params })
    // The backend sends { "tasks": [...] }
    // Return the full object so callers can do data.tasks
    return response.data
  },

  createTask: async (title, description = '', status = 'backlog') => {
    const response = await client.post('/tasks', {
      title,
      description,
      status
    })
    return response.data
  },

  updateTask: async (id, updates) => {
    const response = await client.put(`/tasks/${id}`, updates)
    return response.data
  },

  deleteTask: async (id) => {
    await client.delete(`/tasks/${id}`)
  },

  updateTaskStatus: async (id, status) => {
    const response = await client.patch(`/tasks/${id}/status`, { status })
    return response.data
  },

  getTaskLogs: async (id) => {
    const response = await client.get(`/tasks/${id}/logs`)
    return response.data?.logs || []
  },

  archiveTask: async (id) => {
    const response = await client.post(`/tasks/${id}/archive`)
    return response.data
  },

  unarchiveTask: async (id) => {
    const response = await client.post(`/tasks/${id}/unarchive`)
    return response.data
  },

  // Chat session endpoints
  fetchChatSessions: async ({ archived, limit, offset } = {}) => {
    const params = {}
    if (archived !== undefined) params.archived = archived ? 'true' : 'false'
    if (limit) params.limit = limit
    if (offset) params.offset = offset
    const response = await client.get('/chat/sessions', { params })
    return response.data
  },

  fetchSessionMessages: async (sessionId) => {
    const response = await client.get(`/chat/sessions/${encodeURIComponent(sessionId)}`)
    return response.data
  },

  deleteSession: async (sessionId) => {
    const response = await client.delete(`/chat/sessions/${encodeURIComponent(sessionId)}`)
    return response.data
  },

  updateSession: async (sessionId, { title, archived } = {}) => {
    const payload = {}
    if (title !== undefined) payload.title = title
    if (archived !== undefined) payload.archived = archived
    const response = await client.patch(`/chat/sessions/${encodeURIComponent(sessionId)}`, payload)
    return response.data
  },

  searchMessages: async (query) => {
    const response = await client.get('/chat/search', { params: { q: query } })
    return response.data
  },

  searchTasks: async (query) => {
    const response = await client.get('/tasks/search', { params: { q: query } })
    return response.data
  }
}

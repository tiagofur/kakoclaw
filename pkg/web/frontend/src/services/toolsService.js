import client from './api'

export default {
  // Tool Permissions API
  fetchToolPermissions: async () => {
    const response = await client.get('/tools/permissions')
    return response.data
  },

  updateToolPermissions: async (permissions) => {
    const response = await client.put('/tools/permissions', permissions)
    return response.data
  },

  // User Tool Overrides API
  fetchUserTools: async (userId) => {
    const response = await client.get(`/users/${userId}/tools`)
    return response.data
  },

  updateUserTools: async (userId, tools) => {
    const response = await client.put(`/users/${userId}/tools`, { tools })
    return response.data
  },

  // Audit Log API
  fetchAuditLog: async (filters = {}) => {
    const params = {}
    if (filters.user_id) params.user_id = filters.user_id
    if (filters.tool) params.tool = filters.tool
    if (filters.since) params.since = filters.since
    if (filters.until) params.until = filters.until
    if (filters.success !== undefined) params.success = filters.success
    if (filters.limit) params.limit = filters.limit
    if (filters.offset) params.offset = filters.offset

    const response = await client.get('/tools/audit', { params })
    return response.data
  }
}

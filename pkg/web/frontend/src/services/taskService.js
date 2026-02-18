import client from './api'

export default {
  fetchTasks: async () => {
    const response = await client.get('/tasks')
    return response.data || []
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
    return response.data || []
  }
}

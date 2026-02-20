import client from './api'

export default {
  listUsers: async () => {
    const response = await client.get('/users')
    return response.data
  },

  createUser: async (username, password, role) => {
    const response = await client.post('/users', { username, password, role })
    return response.data
  },

  updateUser: async (id, password, role) => {
    // only send fields that are truthy
    const payload = {}
    if (password) payload.password = password
    if (role) payload.role = role
    
    const response = await client.put(`/users/${id}`, payload)
    return response.data
  },

  deleteUser: async (id) => {
    const response = await client.delete(`/users/${id}`)
    return response.data
  }
}

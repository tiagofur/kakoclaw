import client from './api'

export default {
  listOrders: async () => {
    const response = await client.get('/standing-orders')
    return response.data
  },

  createOrder: async (content, label = '') => {
    const response = await client.post('/standing-orders', { content, label })
    return response.data
  },

  updateOrder: async (id, data) => {
    const response = await client.put(`/standing-orders/${id}`, data)
    return response.data
  },

  deleteOrder: async (id) => {
    const response = await client.delete(`/standing-orders/${id}`)
    return response.data
  }
}

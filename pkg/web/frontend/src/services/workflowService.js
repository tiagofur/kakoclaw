import client from './api'

export default {
  // List all workflows
  fetchWorkflows: async () => {
    const response = await client.get('/workflows')
    return response.data
  },

  // Create a new workflow
  createWorkflow: async (data) => {
    const response = await client.post('/workflows', data)
    return response.data
  },

  // Get a single workflow
  getWorkflow: async (id) => {
    const response = await client.get(`/workflows/${id}`)
    return response.data
  },

  // Update a workflow
  updateWorkflow: async (id, data) => {
    const response = await client.put(`/workflows/${id}`, data)
    return response.data
  },

  // Delete a workflow
  deleteWorkflow: async (id) => {
    const response = await client.delete(`/workflows/${id}`)
    return response.data
  },

  // Run a workflow
  runWorkflow: async (id) => {
    const response = await client.post(`/workflows/${id}/run`, {}, {
      timeout: 300000 // 5 min for workflow execution
    })
    return response.data
  },

  // Get recent runs for a workflow
  getWorkflowRuns: async (id) => {
    const response = await client.get(`/workflows/${id}/runs`)
    return response.data
  },

  // Fetch available tool names (for tool step configuration)
  fetchTools: async () => {
    const response = await client.get('/tools')
    return response.data
  }
}

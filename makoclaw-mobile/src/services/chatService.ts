import api from './api'

export interface ChatSession {
    session_id: string
    title: string
    last_message: string
    updated_at: string
    archived: boolean
}

export default {
    fetchSessions: async (params?: { archived?: boolean; limit?: number; offset?: number }) => {
        const queryParams: Record<string, string> = {}
        if (params?.archived !== undefined) queryParams.archived = params.archived ? 'true' : 'false'
        if (params?.limit) queryParams.limit = String(params.limit)
        if (params?.offset) queryParams.offset = String(params.offset)
        const response = await api.get('/chat/sessions', { params: queryParams })
        return response.data
    },

    fetchSessionMessages: async (sessionId: string) => {
        const response = await api.get(`/chat/sessions/${encodeURIComponent(sessionId)}`)
        return response.data
    },

    deleteSession: async (sessionId: string) => {
        const response = await api.delete(`/chat/sessions/${encodeURIComponent(sessionId)}`)
        return response.data
    },

    updateSession: async (sessionId: string, updates: { title?: string; archived?: boolean }) => {
        const response = await api.patch(`/chat/sessions/${encodeURIComponent(sessionId)}`, updates)
        return response.data
    },

    searchMessages: async (query: string) => {
        const response = await api.get('/chat/search', { params: { q: query } })
        return response.data
    },

    fetchModels: async () => {
        const response = await api.get('/models')
        return response.data
    },

    fetchAgents: async () => {
        const response = await api.get('/agents')
        return response.data
    }
}

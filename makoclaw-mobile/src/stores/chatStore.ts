import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ChatMessage {
    id: number
    role: 'user' | 'assistant' | 'system'
    content: string
    timestamp: string
    streaming?: boolean
    agents?: string[]
    toolCalls?: any[]
    agentActivities?: AgentActivity[]
}

export interface AgentActivity {
    id: number
    agent: string
    status: 'delegating' | 'working' | 'complete' | 'fallback' | 'requesting_help'
    reason: string
    timestamp: string
    expanded: boolean
    confidence?: number
    toolsUsed?: string[]
}

export const useChatStore = defineStore('chat', () => {
    const messages = ref<ChatMessage[]>([])
    const isConnected = ref(false)
    const isLoading = ref(false)
    const isStreaming = ref(false)
    const streamingMessageId = ref<number | null>(null)
    const activeSessionId = ref<string | null>(null)
    const selectedModel = ref('')
    const currentModel = ref('')

    // Agent status
    const orchestratorStatus = ref<'idle' | 'analyzing' | 'delegating' | 'working' | 'complete'>('idle')
    const activeSpecialist = ref<string | null>(null)
    const delegationReason = ref('')

    const isAgentWorking = computed(() =>
        orchestratorStatus.value === 'working' || orchestratorStatus.value === 'delegating'
    )

    function addMessage(message: Partial<ChatMessage>) {
        messages.value.push({
            id: Date.now(),
            role: 'user',
            content: '',
            timestamp: new Date().toISOString(),
            ...message
        })
    }

    function startStreamingMessage(): number {
        const id = Date.now()
        messages.value.push({
            id,
            role: 'assistant',
            content: '',
            timestamp: new Date().toISOString(),
            streaming: true,
            agents: [],
            agentActivities: []
        })
        streamingMessageId.value = id
        isStreaming.value = true
        return id
    }

    function appendStreamToken(token: string) {
        if (!streamingMessageId.value) return
        const msg = messages.value.find(m => m.id === streamingMessageId.value)
        if (msg) {
            msg.content += token
        }
    }

    function endStreamingMessage(finalContent?: string, agents?: string[]) {
        if (streamingMessageId.value) {
            const msg = messages.value.find(m => m.id === streamingMessageId.value)
            if (msg) {
                if (finalContent) msg.content = finalContent
                msg.streaming = false
                if (agents && agents.length > 0) msg.agents = agents
                // Mark agent activities as complete
                if (msg.agentActivities) {
                    for (const act of msg.agentActivities) {
                        if (act.status === 'working' || act.status === 'delegating') {
                            act.status = 'complete'
                            act.expanded = false
                        }
                    }
                }
            }
        }
        streamingMessageId.value = null
        isStreaming.value = false
    }

    function addAgentEvent(event: any) {
        const persistableStatuses = ['delegating', 'working', 'complete', 'fallback', 'requesting_help']
        if (!persistableStatuses.includes(event.status)) return

        if (streamingMessageId.value) {
            const msg = messages.value.find(m => m.id === streamingMessageId.value)
            if (msg) {
                if (!msg.agentActivities) msg.agentActivities = []
                const agentName = event.specialistName || event.agent
                const existing = msg.agentActivities.find(a => a.agent === agentName)
                if (existing) {
                    existing.status = event.status
                    if (event.reason) existing.reason = event.reason
                    existing.timestamp = event.timestamp || new Date().toISOString()
                    existing.expanded = event.status === 'working' || event.status === 'delegating'
                } else {
                    msg.agentActivities.push({
                        id: Date.now() + Math.random(),
                        agent: agentName,
                        status: event.status,
                        reason: event.reason || '',
                        timestamp: event.timestamp || new Date().toISOString(),
                        expanded: event.status === 'working' || event.status === 'delegating'
                    })
                }
            }
        }
    }

    function addToolCall(toolCall: any) {
        if (!streamingMessageId.value) return
        const msg = messages.value.find(m => m.id === streamingMessageId.value)
        if (msg) {
            if (!msg.toolCalls) msg.toolCalls = []
            msg.toolCalls.push({
                ...toolCall,
                id: Date.now() + Math.random(),
                timestamp: new Date().toISOString()
            })
        }
    }

    function setMessages(newMessages: ChatMessage[]) {
        messages.value = newMessages
    }

    function clearMessages() {
        messages.value = []
        isStreaming.value = false
        streamingMessageId.value = null
    }

    function setConnected(connected: boolean) {
        isConnected.value = connected
    }

    function setLoading(loading: boolean) {
        isLoading.value = loading
    }

    function setActiveSessionId(sessionId: string | null) {
        activeSessionId.value = sessionId
    }

    function setAgentStatus(status: typeof orchestratorStatus.value, specialist?: string, reason?: string) {
        orchestratorStatus.value = status
        if (specialist) activeSpecialist.value = specialist
        if (reason) delegationReason.value = reason
    }

    function clearAgentStatus() {
        orchestratorStatus.value = 'idle'
        activeSpecialist.value = null
        delegationReason.value = ''
    }

    return {
        messages,
        isConnected,
        isLoading,
        isStreaming,
        streamingMessageId,
        activeSessionId,
        selectedModel,
        currentModel,
        orchestratorStatus,
        activeSpecialist,
        delegationReason,
        isAgentWorking,
        addMessage,
        startStreamingMessage,
        appendStreamToken,
        endStreamingMessage,
        addAgentEvent,
        addToolCall,
        setMessages,
        clearMessages,
        setConnected,
        setLoading,
        setActiveSessionId,
        setAgentStatus,
        clearAgentStatus
    }
})

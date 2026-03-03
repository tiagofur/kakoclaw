import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAgentsStore } from './agentsStore'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const isConnected = ref(false)
  const isLoading = ref(false)
  const globalIsLoading = ref(false) // Tracks loading state even when not viewing chat
  const isWorking = ref(false)        // True while agent is processing a request (persists across navigation)
  const activeSessionId = ref(null)   // Persisted active session ID across view navigation
  const pendingMessages = ref([])     // Messages received while ChatView was unmounted
  const ws = ref(null)
  const selectedModel = ref('')       // User-selected model override (empty = default)
  const currentModel = ref('')        // Default model from server config
  const availableProviders = ref([])  // Providers list from /api/v1/models
  const isStreaming = ref(false)      // True while receiving streaming tokens
  const streamingMessageId = ref(null) // ID of message currently being streamed
  const webSearchEnabled = ref(true)  // Whether web_search tool is available to LLM (legacy/shortcut)
  const availableTools = ref([])      // All tools available from backend
  const enabledTools = ref([])        // Tools currently enabled by user

  const orchestratorStatus = ref('idle') // 'idle', 'analyzing', 'delegating', 'working', 'complete'
  const currentAgent = ref(null)         // Agente actualmente activo
  const activeSpecialist = ref(null)     // Especialista activo cuando se delega
  const delegationReason = ref('')       // Por qué el orchestrator delegó
  const agentHistory = ref([])           // Historial de eventos de agentes
  const currentSpecialist = ref(null)    // Currently active specialist name (legacy)
  const isMultiAgent = ref(false)         // Whether multi-agent mode is active
  const specialistReports = ref([])      // Structured reports from specialists
  const lastSpecialistReport = ref(null) // Most recent specialist report for quick access
  const delegationChain = ref([])        // Current delegation chain (e.g. ['orchestrator', 'developer'])
  const activeDelegation = ref(null)     // Currently active delegation update
  const delegationHistory = ref([])      // History of completed delegations

  function addMessage(message) {
    messages.value.push({
      id: Date.now(),
      ...message,
      timestamp: new Date().toISOString()
    })
  }

  // Start a new streaming assistant message (empty content, to be filled by tokens)
  function startStreamingMessage() {
    const id = Date.now()
    messages.value.push({
      id,
      role: 'assistant',
      content: '',
      timestamp: new Date().toISOString(),
      streaming: true,
      agents: [], // Initialize empty agents array
      segments: [] // Initialize empty segments array
    })
    streamingMessageId.value = id
    isStreaming.value = true
    return id
  }

  // Append a token to the currently streaming message
  function appendStreamToken(token) {
    if (!streamingMessageId.value) return
    const msg = messages.value.find(m => m.id === streamingMessageId.value)
    if (msg) {
      msg.content += token
    }
  }

  // Finalize the streaming message (set final content, mark as not streaming)
  function endStreamingMessage(finalContent, agents = []) {
    if (streamingMessageId.value) {
      const msg = messages.value.find(m => m.id === streamingMessageId.value)
      if (msg) {
        // Use final content from server if provided (authoritative), otherwise keep accumulated
        if (finalContent) {
          msg.content = finalContent
        }
        msg.streaming = false
        // Assign agents if provided
        if (agents && agents.length > 0) {
          msg.agents = agents
        }
        // Snapshot agent activity onto the message for persistence
        if (agentHistory.value.length > 0 || specialistReports.value.length > 0) {
          msg.agentActivity = {
            hadMultiAgent: agentHistory.value.length > 1,
            history: [...agentHistory.value],
            reports: [...specialistReports.value],
            delegationChain: [...delegationChain.value],
            delegationHistory: [...delegationHistory.value]
          }
        }
      }
    }
    streamingMessageId.value = null
    isStreaming.value = false
  }

  // Add an agent event as an inline indicator in the chat stream
  function addAgentEvent(event) {
    const persistableStatuses = ['delegating', 'working', 'complete', 'fallback', 'requesting_help']
    if (!persistableStatuses.includes(event.status)) return

    const eventMsg = {
      id: Date.now() + Math.random(),
      type: 'agent_event',
      role: 'system',
      agent: event.agent,
      status: event.status,
      specialistName: event.specialistName,
      reason: event.reason,
      timestamp: event.timestamp || new Date().toISOString()
    }

    // Insert before the current streaming message for chronological order
    if (streamingMessageId.value) {
      const streamIdx = messages.value.findIndex(m => m.id === streamingMessageId.value)
      if (streamIdx !== -1) {
        messages.value.splice(streamIdx, 0, eventMsg)
        return
      }
    }
    messages.value.push(eventMsg)
  }

  // Set agents for a specific message (for loading from history or post-streaming)
  function setAgentsForMessage(messageId, agents) {
    const msg = messages.value.find(m => m.id === messageId)
    if (msg) {
      msg.agents = agents || []
    }
  }

  // Set agent status during execution
  function setAgentStatus(agent, status, specialistName = null, reason = '') {
    orchestratorStatus.value = status
    currentAgent.value = agent

    if (specialistName) {
      activeSpecialist.value = specialistName
    }

    if (reason) {
      delegationReason.value = reason
    }

    agentHistory.value.push({
      agent,
      status,
      specialistName,
      reason,
      timestamp: new Date().toISOString()
    })
  }

  // Clear agent status (call when execution completes)
  function clearAgentStatus() {
    orchestratorStatus.value = 'idle'
    currentAgent.value = null
    activeSpecialist.value = null
    delegationReason.value = ''
    agentHistory.value = []
    specialistReports.value = []
    lastSpecialistReport.value = null
    delegationChain.value = []
    activeDelegation.value = null
    delegationHistory.value = []
  }

  // Add a specialist report (from structured feedback)
  function addSpecialistReport(report) {
    specialistReports.value.push({
      ...report,
      receivedAt: new Date().toISOString()
    })
    lastSpecialistReport.value = report

    // Update orchestrator status based on report
    if (report.status === 'needs_help' && report.request_help) {
      orchestratorStatus.value = 'delegating'
      activeSpecialist.value = report.request_help
      delegationReason.value = report.help_context || `${report.specialist_name} requested help`
    } else if (report.status === 'complete') {
      // Keep current status, specialist finished
    }
  }

  // Update delegation chain from agent_status events
  function updateDelegationChain(data) {
    if (data.delegation_chain && data.delegation_chain.length > 0) {
      delegationChain.value = [...data.delegation_chain]
    }
  }

  // Update delegation progress from delegation_update events
  function updateDelegationProgress(update) {
    activeDelegation.value = {
      id: update.delegation_id,
      from: update.from,
      to: update.to,
      status: update.status,
      iteration: update.iteration,
      maxIterations: update.max_iterations,
      elapsedMs: update.elapsed_ms,
      timestamp: update.timestamp
    }

    if (update.status === 'complete' || update.status === 'error') {
      delegationHistory.value.push({ ...activeDelegation.value })
      activeDelegation.value = null
    }
  }

  // Set delegation summary from stream_end
  function setDelegationSummary(summary) {
    // Summary is stored on the message via endStreamingMessage
    // This is kept for explicit external calls
    if (!summary) return
  }

  // Add a content segment to the current streaming message
  function addContentSegment(segment) {
    if (!streamingMessageId.value) return
    const msg = messages.value.find(m => m.id === streamingMessageId.value)
    if (msg) {
      if (!msg.segments) msg.segments = []
      msg.segments.push({
        agent: segment.agent,
        content: segment.content,
        segmentId: segment.segment_id
      })
      // Actualizar lista de agentes
      if (!msg.agents.includes(segment.agent)) {
        msg.agents.push(segment.agent)
      }
    }
  }

  function addToolCall(toolCall) {
    if (!streamingMessageId.value) return
    const msg = messages.value.find(m => m.id === streamingMessageId.value)
    if (msg) {
      if (!msg.toolCalls) msg.toolCalls = []
      
      // Try to find an open tool call with the same name to update it
      const existingIdx = msg.toolCalls.findLastIndex(tc => tc.name === toolCall.name && tc.status === 'started')
      if (existingIdx !== -1 && toolCall.status !== 'started') {
        msg.toolCalls[existingIdx] = { ...msg.toolCalls[existingIdx], ...toolCall }
      } else {
        msg.toolCalls.push({
          ...toolCall,
          id: Date.now() + Math.random(),
          timestamp: new Date().toISOString()
        })
      }
    }
  }

  function setMessages(newMessages) {
    messages.value = newMessages
  }

  function clearMessages() {
    messages.value = []
    isStreaming.value = false
    streamingMessageId.value = null
  }

  function setActiveSessionId(sessionId) {
    activeSessionId.value = sessionId
  }

  function setIsWorking(working) {
    isWorking.value = working
    globalIsLoading.value = working
  }

  // Called by background listener (MainLayout) to queue messages when ChatView is not mounted
  function enqueuePendingMessage(message) {
    pendingMessages.value.push(message)
  }

  // Called by ChatView on mount to flush and process all queued messages
  function flushPendingMessages() {
    const flushed = [...pendingMessages.value]
    pendingMessages.value = []
    return flushed
  }

  function sendMessage(content, sessionId) {
    if (ws.value && ws.value.isConnected()) {
      ws.value.send({
        type: 'message',
        content,
        session_id: sessionId || 'web:chat:' + Date.now().toString(36),
        model: selectedModel.value || undefined,
        exclude_tools: availableTools.value.filter(tool => !enabledTools.value.includes(tool))
      })
      return true
    }
    return false
  }

  function setConnected(connected) {
    isConnected.value = connected
  }

  function setLoading(loading) {
    isLoading.value = loading
    // Also update global loading state
    globalIsLoading.value = loading
  }

  function setGlobalLoading(loading) {
    globalIsLoading.value = loading
    if (!loading) isWorking.value = false
  }

  function setWebSocket(websocket) {
    ws.value = websocket
  }

  function setSelectedModel(model) {
    selectedModel.value = model
  }

  function setWebSearchEnabled(enabled) {
    webSearchEnabled.value = enabled
    // Sync with tools list
    if (enabled && !enabledTools.value.includes('web_search')) {
      enabledTools.value.push('web_search')
    } else if (!enabled) {
      enabledTools.value = enabledTools.value.filter(t => t !== 'web_search')
    }
  }

  function setTools(tools) {
    availableTools.value = tools
    // Default all tools to enabled if not already set
    if (enabledTools.value.length === 0) {
      enabledTools.value = [...tools]
    }
    // Sync webSearchEnabled
    webSearchEnabled.value = enabledTools.value.includes('web_search')
  }

  function toggleTool(toolName) {
    if (enabledTools.value.includes(toolName)) {
      enabledTools.value = enabledTools.value.filter(t => t !== toolName)
    } else {
      enabledTools.value.push(toolName)
    }
    // Sync webSearchEnabled
    if (toolName === 'web_search') {
      webSearchEnabled.value = enabledTools.value.includes('web_search')
    }
  }

  function setModelsData(data) {
    const normalized = normalizeModelsData(data)
    currentModel.value = normalized.currentModel
    availableProviders.value = normalized.providers
    // If no model selected yet, use the server default
    if (!selectedModel.value && normalized.currentModel) {
      selectedModel.value = normalized.currentModel
    }
    if (selectedModel.value) {
      const stillAvailable = normalized.providers.some(provider =>
        provider.models.some(model => model.id === selectedModel.value)
      )
      if (!stillAvailable) {
        selectedModel.value = normalized.currentModel || ''
      }
    }
  }

  // Flat list of all available models across providers
  const allModels = computed(() => {
    const models = []
    for (const provider of (Array.isArray(availableProviders.value) ? availableProviders.value : [])) {
      if (!provider.enabled) continue
      const providerModels = Array.isArray(provider.models) ? provider.models : []
      for (const model of providerModels) {
        if (!model?.id) continue
        models.push({
          id: model.id,
          provider: provider.name || model.provider || 'unknown',
          label: `${model.id}`,
          isDefault: model.id === currentModel.value
        })
      }
    }
    return models
  })

  function setOrchestratorStatus(status) {
    orchestratorStatus.value = status
  }

  function setCurrentSpecialist(specialistName) {
    currentSpecialist.value = specialistName
  }

  function setIsMultiAgent(enabled) {
    isMultiAgent.value = enabled
  }

  function getSpecialistDisplay(name) {
    if (!name) return null
    const agentsStore = useAgentsStore()
    return {
      name,
      icon: agentsStore.getSpecialistIcon(name),
      color: agentsStore.getSpecialistColor(name),
      bgColor: agentsStore.getSpecialistBgColor(name)
    }
  }

  return {
    messages,
    isConnected,
    isLoading,
    globalIsLoading,
    isWorking,
    activeSessionId,
    pendingMessages,
    isStreaming,
    streamingMessageId,
    ws,
    selectedModel,
    currentModel,
    availableProviders,
    allModels,
    webSearchEnabled,
    orchestratorStatus,
    currentAgent,
    activeSpecialist,
    delegationReason,
    agentHistory,
    currentSpecialist,
    isMultiAgent,
    addMessage,
    startStreamingMessage,
    appendStreamToken,
    endStreamingMessage,
    setAgentsForMessage,
    setAgentStatus,
    clearAgentStatus,
    addAgentEvent,
    addContentSegment,
    addToolCall,
    setMessages,
    clearMessages,
    sendMessage,
    setConnected,
    setLoading,
    setGlobalLoading,
    setIsWorking,
    setActiveSessionId,
    enqueuePendingMessage,
    flushPendingMessages,
    setWebSocket,
    setSelectedModel,
    setWebSearchEnabled,
    setAvailableTools: setTools,
    toggleTool,
    availableTools,
    enabledTools,
    setModelsData,
    setOrchestratorStatus,
    setCurrentSpecialist,
    setIsMultiAgent,
    getSpecialistDisplay,
    specialistReports,
    lastSpecialistReport,
    addSpecialistReport,
    delegationChain,
    activeDelegation,
    delegationHistory,
    updateDelegationChain,
    updateDelegationProgress,
    setDelegationSummary
  }
})
  const normalizeModelsData = (data) => {
    const providersRaw = Array.isArray(data?.providers) ? data.providers : []
    const providers = providersRaw.map((provider) => {
      const modelsRaw = Array.isArray(provider?.models) ? provider.models : []
      return {
        name: provider?.name || 'unknown',
        enabled: provider?.enabled !== false,
        is_active: provider?.is_active === true,
        models: modelsRaw
          .filter((model) => model && typeof model.id === 'string' && model.id.trim() !== '')
          .map((model) => ({
            id: model.id,
            provider: model.provider || provider?.name || 'unknown'
          }))
      }
    })

    return {
      currentModel: typeof data?.current_model === 'string' ? data.current_model : '',
      providers
    }
  }

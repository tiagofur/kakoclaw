# Design: Agent Visibility Realtime

**Change:** `agent-visibility-realtime`
**Status:** Ready for Tasks
**Created:** 2026-03-25

---

## Technical Approach

Fase 1 es puramente frontend: se modifica el comportamiento de colapso de `ToolCallItem` (computed en lugar de mutación directa), se añade `agentName` al objeto tool call en `chatStore.addToolCall()`, y se refactoriza `TeamActivityPanel` de watcher-based a computed desde el store. Fase 2 añade `ChatStream()` al `ClaudeProvider` (actualmente ausente — solo tiene `Chat()`), extiende `StreamChunk` con `ThinkingDelta`, propaga via callback en `processOptions`, y crea `ThinkingBlock.vue`.

---

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| Expanded state en ToolCallItem | `computed` (derivado del status + msg.streaming) en lugar de mutación directa `tc.expanded = !tc.expanded` | Mantener mutación directa | El estado actual muta `tc.expanded` directo en el template — acoplamiento entre UI y store. Con computed + toggle manual local (`localExpanded`), la regla FR-2 funciona sin race conditions |
| TeamActivityPanel refactor | `computed(() => groupBy(streamingMsg.toolCalls, 'agentName'))` desde chatStore | Mantener watcher-based con props | El panel actual usa 4 watchers y estado local duplicado; al usar computed desde el store se elimina la desincronización documentada en proposal |
| agentName en tool calls | Asignado en `chatStore.addToolCall()` usando `currentAgent.value` | Asignar en ChatView al recibir el WS event | chatStore ya tiene `currentAgent` ref; centralizarlo en la acción mantiene la lógica de fallback en un solo lugar |
| ClaudeProvider ChatStream | Nuevo método `ChatStream()` en `claude_provider.go` | Reusar el Chat() no-streaming | ClaudeProvider actualmente NO implementa `StreamingLLMProvider` — solo tiene `Chat()`. El streaming general ya funciona via otros providers; Claude necesita su propio `ChatStream` para acceder a thinking blocks |
| ThinkingBlock gating | Flag `extended_thinking` en UserConfig + check server-side antes de emitir WS | Solo gating client-side | Spec FR-6 MUST: backend no debe emitir thinking_delta si flag es false. Doble gating (server + client) por seguridad |
| OnThinking callback | Nuevo campo en `processOptions` struct | Canal separado | Consistent con el patrón existente: `OnToken`, `OnTool`, `OnAgentStatus`, `OnContentSegment` ya están en `processOptions` |

---

## Data Flow

### Fase 1: Tool Call con agentName

```
WS event {type: 'tool_call', name, args, status}
    ↓
ChatView.handleMessage('tool_call')
    ↓
chatStore.addToolCall({name, args, status})
    ├── agentName = currentAgent.value || 'main'   ← NEW
    ├── expanded = (msg.streaming && status === 'started')  ← NEW
    └── push to msg.toolCalls[]
    ↓
MessageBubble re-render
    ↓
ToolCallItem (isExpanded computed — no muta tc.expanded directo)
```

### Fase 1: TeamActivityPanel desde store

```
chatStore.messages (reactivo)
    ↓
streamingMsg = computed(() => messages.find(m => m.id === streamingMessageId))
    ↓
TeamActivityPanel:
  agentToolCalls = computed(() =>
    groupBy(streamingMsg.value?.toolCalls || [], 'agentName')
  )
    ↓
render: por cada agentName → mini-lista de ToolCallItem in-flight
```

### Fase 2: Extended Thinking stream

```
Claude API (extended thinking enabled)
    ↓ anthropic SDK: block.Type === "thinking"
ClaudeProvider.ChatStream() → emite StreamChunk{ThinkingDelta: text}
    ↓
loop.go runLLMIterationStream():
  if chunk.ThinkingDelta != "" && opts.OnThinking != nil
    opts.OnThinking(chunk.ThinkingDelta)
    ↓
server.go handleChatWS():
  if session.ExtendedThinking
    ws.WriteJSON({type: "thinking_delta", content})
    ↓
ChatView.handleMessage('thinking_delta')
    ↓
chatStore.appendThinkingDelta(content)  → msg.thinkingBlocks[]
    ↓
MessageBubble → ThinkingBlock.vue (auto-open durante streaming)
```

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/web/frontend/src/stores/chatStore.js` | Modify | `addToolCall`: add `agentName` + `expanded` logic; new `appendThinkingDelta()` action; expose `streamingMessage` computed |
| `pkg/web/frontend/src/components/ToolCallItem.vue` | Modify | Reemplazar mutación directa `tc.expanded` por computed local + toggle; badges semánticos FR-5; recibir prop `msg` para regla FR-2 |
| `pkg/web/frontend/src/components/MessageBubble.vue` | Modify | Pasar `msg` prop a `ToolCallItem`; renderizar `ThinkingBlock` antes del contenido (Fase 2) |
| `pkg/web/frontend/src/components/Chat/TeamActivityPanel.vue` | Modify | Eliminar 4 watchers + estado local; agregar `useChatStore()`; computed `agentToolCalls` desde store; renderizar tool calls per-agent |
| `pkg/web/frontend/src/components/Chat/AgentActivityItem.vue` | Modify | Agregar sección "Tool Calls" filtrada por `tc.agentName === activity.agent` desde `msg.toolCalls` prop |
| `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue` | Modify | Mostrar `currentToolName` cuando agente tiene tool call en `started` |
| `pkg/web/frontend/src/views/ChatView.vue` | Modify | Handler `thinking_delta` WS → `chatStore.appendThinkingDelta()` (Fase 2) |
| `pkg/web/frontend/src/components/Chat/ThinkingBlock.vue` | Create | Nuevo componente: `<details>` animado, brain icon, contenido italic gris, badge agentName (Fase 2) |
| `pkg/web/frontend/src/components/Settings/ProfileSettingsTab.vue` | Modify | Toggle "Extended Thinking (Claude only)" + persist a `/api/v1/user/config` (Fase 2) |
| `pkg/providers/types.go` | Modify (F2) | Campo `ThinkingDelta string` en `StreamChunk` |
| `pkg/providers/claude_provider.go` | Modify (F2) | Implementar `ChatStream()` — actualmente solo existe `Chat()`; detectar bloques `thinking` del Anthropic stream |
| `pkg/agent/loop.go` | Modify (F2) | Campo `OnThinking func(string)` en `processOptions`; propagación en `runLLMIterationStream` |
| `pkg/web/server.go` | Modify (F2) | Activar `OnThinking` callback condicionado a `session.ExtendedThinking`; emitir WS `thinking_delta` |

---

## Interfaces / Contracts

### chatStore — nuevas acciones y state

```js
// State nuevo
currentAgent: ref(null)  // ya existe — confirmar que addToolCall lo lee

// addToolCall actualizado
function addToolCall(toolCall) {
  const msg = messages.value.find(m => m.id === streamingMessageId.value)
  if (!msg) return
  if (!msg.toolCalls) msg.toolCalls = []
  const existingIdx = msg.toolCalls.findLastIndex(
    tc => tc.name === toolCall.name && tc.status === 'started'
  )
  const agentName = currentAgent.value || 'main'  // NEW
  if (existingIdx !== -1 && toolCall.status !== 'started') {
    msg.toolCalls[existingIdx] = { ...msg.toolCalls[existingIdx], ...toolCall, agentName }
  } else {
    msg.toolCalls.push({
      ...toolCall,
      id: Date.now() + Math.random(),
      timestamp: new Date().toISOString(),
      agentName,                                   // NEW
      expanded: msg.streaming && toolCall.status === 'started'  // NEW
    })
  }
}

// Fase 2: nuevo
function appendThinkingDelta(content) {
  const msg = messages.value.find(m => m.id === streamingMessageId.value)
  if (!msg) return
  if (!msg.thinkingBlocks) msg.thinkingBlocks = []
  const lastBlock = msg.thinkingBlocks[msg.thinkingBlocks.length - 1]
  if (lastBlock && !lastBlock.finalized) {
    lastBlock.content += content  // acumular en bloque activo
  } else {
    msg.thinkingBlocks.push({
      id: `tb-${Date.now()}`,
      content,
      expanded: true,   // auto-open durante streaming
      finalized: false,
      manuallyToggled: false
    })
  }
}

// Fase 2: llamar en endStreamingMessage para cerrar ThinkingBlocks
// si !block.manuallyToggled → block.expanded = false
```

### ToolCallItem.vue — nuevo contrato de props

```js
// Props actualizadas
defineProps({
  tc: { type: Object, required: true },
  msg: { type: Object, required: true }  // NEW — necesario para regla FR-2
})

// Computed local (reemplaza mutación directa)
const localExpanded = ref(false)
const isExpanded = computed(() => {
  if (props.msg.streaming && props.tc.status === 'started') return true
  return localExpanded.value
})
// Reset localExpanded cuando status cambia de 'started'
watch(() => props.tc.status, (newStatus) => {
  if (newStatus !== 'started') localExpanded.value = false
})
function toggleExpanded() {
  if (props.msg.streaming && props.tc.status === 'started') {
    localExpanded.value = false  // manual collapse de auto-expanded
  } else {
    localExpanded.value = !localExpanded.value
  }
}
```

### StreamChunk — Fase 2

```go
type StreamChunk struct {
    Content        string     `json:"content"`
    ToolCalls      []ToolCall `json:"tool_calls,omitempty"`
    ThinkingDelta  string     `json:"thinking_delta,omitempty"` // NEW
    FinishReason   string     `json:"finish_reason"`
    Done           bool       `json:"done"`
    Usage          *UsageInfo `json:"usage,omitempty"`
    Error          string     `json:"error,omitempty"`
}
```

### processOptions — Fase 2

```go
type processOptions struct {
    // ... campos existentes ...
    OnThinking func(string)  // NEW — callback para extended thinking deltas
}
```

---

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (JS) | `chatStore.addToolCall` asigna `agentName` desde `currentAgent` | Vitest: set `currentAgent.value = 'developer'`, dispatch addToolCall, assert `tc.agentName === 'developer'` |
| Unit (JS) | `currentAgent` null → fallback a `'main'` | Vitest: currentAgent = null, dispatch, assert agentName = 'main' |
| Unit (JS) | `appendThinkingDelta` acumula en bloque activo | Vitest: llamar 3 veces, assert 1 bloque con contenido concatenado |
| Unit (JS) | `endStreamingMessage` colapsa ThinkingBlocks no toggled manualmente | Vitest: thinkingBlock.manuallyToggled = false → expanded = false tras endStreaming |
| Unit (Go) | `StreamChunk.ThinkingDelta` propagado via `OnThinking` | Go test: mock streamingProvider que emite ThinkingDelta, assert OnThinking called |
| Component | `ToolCallItem` auto-expande con status=started + msg.streaming | Vue Test Utils: mount con msg.streaming=true + tc.status='started', assert expanded |
| Component | `ToolCallItem` colapsa al pasar status a 'finished' | Vue Test Utils: cambiar tc.status a 'finished', assert collapsed |
| Component | `TeamActivityPanel` renderiza tool calls por agente | Vue Test Utils: store con 2 agentes + toolCalls, assert columnas por agente |
| E2E | Histórico: todos los tool calls colapsados | Playwright: cargar mensaje histórico, assert no expanded items (F1-AC-1) |
| E2E | Toggle Extended Thinking persiste en reload | Playwright: toggle ON → reload → assert ON (F2-AC-3) |

---

## Key Implementation Notes

### ToolCallItem: el bug actual

El template actual hace `@click="tc.expanded = !tc.expanded"` — muta el objeto del store directamente desde el template. En Vue 3 con Pinia reactivo esto funciona pero viola el principio de unidireccionalidad. El nuevo diseño usa `localExpanded` + `computed isExpanded` para separar estado derivado (FR-2) de estado manual (click).

### TeamActivityPanel: eliminar defineExpose({ reset })

El refactor elimina el método `reset()` expuesto. Los consumidores (ChatView) que llamen `teamPanelRef.reset()` deben migrar a `chatStore.clearAgentStatus()`.

### ClaudeProvider.ChatStream() — implementación crítica

`claude_provider.go` actualmente NO implementa `StreamingLLMProvider`. Para Fase 2 se necesita:
1. Agregar `ChatStream()` usando `p.client.Messages.NewStreaming()`
2. Dentro del stream, detectar `block.Type == "thinking"` del Anthropic SDK
3. Emitir `StreamChunk{ThinkingDelta: text}` para cada thinking delta

El patrón de `ChatStream()` puede seguir el de `ollama_provider.go` que sí lo implementa.

### Extended Thinking: UserConfig

`UserConfig` ya existe en `pkg/storage/user_storage.go`. Agregar campo `ExtendedThinking bool` y endpoint `PUT /api/v1/user/config` para persistirlo. El flag debe estar en `false` por default (spec FR-6).

---

## Migration / Rollout

**Fase 1:** Sin migración. Cambios de presentación pura. Revertir `ToolCallItem.vue` y `MessageBubble.vue` restaura comportamiento previo.

**Fase 2:** Flag `extended_thinking: false` por default — usuarios existentes no ven cambio. El nuevo campo en `UserConfig` necesita migración de schema en SQLite (ALTER TABLE o recreación). No hay datos pre-existentes de `thinkingBlocks` (in-session only, no persistidos).

---

## Open Questions

- [ ] ¿El campo `extended_thinking` en `UserConfig` requiere migración de schema con ALTER TABLE o se usa una columna JSON existente?
- [ ] ¿Hay ya un endpoint `PUT /api/v1/user/config` en `server.go` o hay que crearlo?
- [ ] El `AgentStatusIndicator.vue` actualmente muestra solo el nombre del agente — ¿se muestra también el tool name o solo en `TeamActivityPanel`?

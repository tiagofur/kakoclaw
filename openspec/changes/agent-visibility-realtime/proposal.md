# Proposal: Agent Visibility Realtime

**Change:** agent-visibility-realtime  
**Date:** 2026-03-25  
**Status:** proposed

---

## Intent

Los usuarios de MakoClaw no tienen visibilidad real de lo que el agente hace mientras procesa. Los tool calls llegan al frontend via WS pero se muestran siempre expandidos y de forma idéntica en streaming histórico, creando ruido visual. En multi-agent, el `TeamActivityPanel` acumula agentes sin asociarlos a sus tool calls específicos. La Fase 2 agrega opcionalmente el "pensamiento interno" de Claude extended thinking, un diferenciador competitivo que requiere opt-in explícito por el costo adicional de tokens que genera.

**Problema concreto del usuario:**
- Durante streaming: ve tool calls apareciendo y no sabe cuál está "en vuelo"
- En histórico: todos los tool calls están expandidos y son ruidosos
- En multi-agent: el panel no muestra "qué herramienta usa cada agente ahora"

---

## Scope

### In Scope — Fase 1 (Frontend-Only, sin cambios backend)

- **`MessageBubble.vue`**: tool calls colapsados por default en histórico (`!msg.streaming`), expandidos durante streaming activo (`msg.streaming && tc.status === 'started'`)
- **`ToolCallItem.vue`**: badge de estado en tiempo real (`executing...` / `done` / `error`) más visible
- **`AgentActivityItem.vue`**: sección "Tool Calls" dentro del expand, mostrando tool calls por agente (las tools ya están en `activity.toolsUsed`; agregar las in-flight desde el store)
- **`TeamActivityPanel.vue`**: columna per-agent con tool calls activos en tiempo real, consumiendo `chatStore.streamingMessageId` + `msg.toolCalls`
- **`chatStore.js`**: asociar cada `tool_call` WS event con el agente activo en ese momento (`currentAgent`) para alimentar `AgentActivityItem`

### In Scope — Fase 2 (Backend + Frontend, opt-in)

- **`pkg/providers/types.go`**: campo `ThinkingDelta string` en `StreamChunk`
- **`pkg/providers/claude_provider.go`**: capturar bloques `thinking` del stream de Anthropic Extended Thinking; emitir `ThinkingDelta` en el canal
- **`pkg/agent/loop.go`**: nuevo callback `OnThinking func(string)` en `processOptions`; llamar en el stream loop cuando `chunk.ThinkingDelta != ""`
- **`pkg/web/server.go`**: nuevo evento WS `thinking_delta` con `{content: string}` cuando `OnThinking` dispara; solo activo si la sesión tiene `extended_thinking: true`
- **`ThinkingBlock.vue`**: nuevo componente colapsable con animación de "pensar", icono brain, texto italic en gris
- **`MessageBubble.vue`**: renderizar `msg.thinkingBlocks[]` antes del content principal
- **`chatStore.js`**: `appendThinkingDelta()`, `msg.thinkingBlocks[]`
- **`SettingsView.vue` / `ProfileSettingsTab.vue`**: toggle opt-in "Extended Thinking (Claude only)" — persistido en user preferences vía `/api/v1/user/config`

### Out of Scope

- Mobile app (tiene implementación WS propia — cambio separado)
- Fix del callback chain backend specialist (eso es `multi-agent-orchestration-visibility`)
- Thinking para providers no-Claude (OpenAI, Ollama, HTTP)
- Parallel delegation / token budgets
- Persistencia de `thinkingBlocks` en SQLite (se muestra solo en sesión activa)

---

## Approach

### Fase 1 — Colapso inteligente de tool calls

**Regla de expansión:**
```
expanded = msg.streaming && tc.status === 'started'
```
- Durante streaming: tool call activo (`started`) → auto-expandido
- Tool call completado (`status !== 'started'`) en mensaje streaming → colapsado
- Mensaje histórico (`!msg.streaming`) → todos colapsados por default

**Asociación tool-call → agente:**  
En `chatStore.js`, al procesar el evento WS `tool_call`, leer `currentAgent.value` en ese momento y adjuntarlo al objeto tool call (`tc.agentName = currentAgent.value`). `AgentActivityItem` ya tiene la sección "Tools Used"; se enriquece con in-flight tools filtrando desde `msg.toolCalls` donde `tc.agentName === activity.agent`.

**TeamActivityPanel — tool calls por agente:**  
Cambiar de prop-based a computed desde el store: `const agentToolCalls = computed(() => groupBy(streamingMsg.toolCalls, 'agentName'))`. Renderizar mini-lista por agente debajo del status badge. No requiere nuevos eventos WS.

### Fase 2 — Extended Thinking stream

**Backend (Claude):**  
Anthropic SDK expone bloques `thinking` en el stream (`block.Type == "thinking"`). En `ChatStream()`, cuando se detecta un `thinking` delta, emitir `StreamChunk{ThinkingDelta: text}`. El `loop.go` los propaga via `OnThinking`. El WS handler emite `{"type":"thinking_delta","content":"..."}` solo si el usuario tiene `extended_thinking: true` en su config.

**Opt-in:**  
Flag `extended_thinking bool` en `UserConfig` (ya existe estructura en `pkg/storage/user_storage.go`). UI toggle en `ProfileSettingsTab.vue`. Cuando `true`, el WS handler activa el callback `OnThinking`.

**Frontend:**  
`ThinkingBlock.vue` muestra un `<details>` animado. Durante streaming: open y animando. Al completar: cerrado. `chatStore` acumula deltas en `msg.thinkingBlocks[{id, content}]`.

---

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/web/frontend/src/components/MessageBubble.vue` | Modified | Lógica de colapso + renderizar ThinkingBlock |
| `pkg/web/frontend/src/components/ToolCallItem.vue` | Modified | Badge estado más prominente, default collapsed |
| `pkg/web/frontend/src/components/Chat/AgentActivityItem.vue` | Modified | Tool calls in-flight por agente |
| `pkg/web/frontend/src/components/Chat/TeamActivityPanel.vue` | Modified | Computed desde store, tool calls por agente |
| `pkg/web/frontend/src/components/Chat/ThinkingBlock.vue` | New | Componente colapsable para extended thinking |
| `pkg/web/frontend/src/stores/chatStore.js` | Modified | `agentName` en tool calls, `appendThinkingDelta`, `thinkingBlocks` |
| `pkg/web/frontend/src/views/ChatView.vue` | Modified | Handler `thinking_delta` WS |
| `pkg/web/frontend/src/components/Settings/ProfileSettingsTab.vue` | Modified | Toggle opt-in extended thinking |
| `pkg/providers/types.go` | Modified (F2) | Campo `ThinkingDelta` en `StreamChunk` |
| `pkg/providers/claude_provider.go` | Modified (F2) | Capturar bloques thinking del stream |
| `pkg/agent/loop.go` | Modified (F2) | Callback `OnThinking`, propagación |
| `pkg/web/server.go` | Modified (F2) | Emitir evento WS `thinking_delta` |

---

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| TeamActivityPanel desacoplado del store pierde sync | Med | Refactorizar para leer directamente de `chatStore` en lugar de props; eliminar estado local duplicado |
| Extended thinking aumenta costo de tokens (~$3/Mtok thinking) | High | Opt-in explícito con advertencia de costo en la UI; flag `false` por default |
| Mobile app no recibe mejoras | Low | Documentar en out-of-scope; issue separado para mobile |
| `currentAgent` puede ser null cuando llega `tool_call` WS | Med | Fallback a `"main"` si `currentAgent` es null al momento del evento |
| Claude SDK breaking changes en tipos de thinking blocks | Low | Pinear versión SDK; cubrir con tests |

---

## Rollback Plan

- **Fase 1**: Los cambios son puramente de presentación. Revertir `MessageBubble.vue` y `ToolCallItem.vue` a la versión anterior restaura el comportamiento. Sin migraciones, sin cambios de schema.
- **Fase 2**: El flag `extended_thinking` es `false` por default. Desactivar el toggle en UI es suficiente. Si se necesita revertir el código: los cambios en `claude_provider.go` son aditivos (nuevo campo en `StreamChunk`); revertir `server.go` para no emitir el evento WS es suficiente para aislar el feature.

---

## Dependencies

- Fase 1: Ninguna. Usa eventos WS `tool_call` y `agent_status` ya existentes.
- Fase 2: Anthropic SDK con soporte Extended Thinking (`claude-3-7-sonnet` o superior). Verificar que el modelo configurado soporte `thinking` blocks.
- Fase 2 depende de Fase 1 completada (ThinkingBlock va en MessageBubble que se modifica en F1).

---

## Success Criteria

- [ ] **F1**: En un mensaje histórico con 5 tool calls, todos aparecen colapsados por default
- [ ] **F1**: Durante streaming, el tool call en estado `started` se expande automáticamente; los completados se colapsan
- [ ] **F1**: `TeamActivityPanel` muestra el nombre de la herramienta activa debajo del badge de cada agente durante multi-agent
- [ ] **F1**: `AgentActivityItem` en el expand de cada agente muestra sus tool calls con estado en tiempo real
- [ ] **F2**: Usuario con `extended_thinking: false` (default) no recibe eventos `thinking_delta` y no ve ThinkingBlock
- [ ] **F2**: Usuario con `extended_thinking: true` ve ThinkingBlock animado durante streaming, colapsado al finalizar
- [ ] **F2**: Toggle en Settings persiste en user config y sobrevive reload de página

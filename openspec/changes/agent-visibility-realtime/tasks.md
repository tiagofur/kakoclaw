# Tasks: Agent Visibility Realtime

## Phase 1 — Frontend Only (Quick Wins)

- [x] 1.1 `chatStore.js` — agregar `currentAgent: ref('main')`; modificar `addToolCall()` para asignar `agentName = currentAgent.value || 'main'` y `expanded = msg.streaming && status === 'started'`; agregar acción `updateCurrentAgent(agentName)`
- [x] 1.2 `ChatView.vue` — agregar handler para evento WS `agent_status`; llamar `chatStore.updateCurrentAgent(agent.specialist_name || agent.name)`
- [x] 1.3 `ToolCallItem.vue` — reemplazar mutación directa `tc.expanded = !tc.expanded` con `localExpanded: ref(false)` + computed `isExpanded = (msg.streaming && tc.status === 'started') || localExpanded.value`; agregar `watch` que resetea `localExpanded` cuando `status !== 'started'`; recibir prop `msg`; badges semánticos por status (yellow/green/red, FR-5)
- [x] 1.4 `MessageBubble.vue` — pasar prop `msg` a cada `<ToolCallItem>`; verificar que histórico (`!msg.streaming`) no auto-expande ningún tool call
- [x] 1.5 `AgentActivityItem.vue` — agregar sección "Tool Calls" en el expand; filtrar `msg.toolCalls` por `tc.agentName === activity.agent`; mostrar badge de status por tool call
- [x] 1.6 `TeamActivityPanel.vue` — eliminar 4 watchers y estado local; importar `useChatStore()`; agregar computed `agentToolCalls = computed(() => groupBy(streamingMsg.value?.toolCalls || [], 'agentName'))`; renderizar mini-lista de tool calls por agente debajo de cada badge; ocultar sección si agente tiene 0 tools activos
- [x] 1.7 `AgentStatusIndicator.vue` — mostrar nombre de herramienta activa (`status === 'started'`) debajo del agent name usando `chatStore.currentAgent` + lookup en `streamingMsg.toolCalls`

## Phase 2 — Backend + Frontend (Extended Thinking)

- [ ] 2.1 `pkg/providers/types.go` — agregar campo `ThinkingDelta string \`json:"thinking_delta,omitempty"\`` a `StreamChunk`
- [ ] 2.2 `pkg/agent/loop.go` — agregar campo `OnThinking func(string)` a `processOptions`; en `runLLMIterationStream()` llamar `opts.OnThinking(chunk.ThinkingDelta)` cuando ambos no son vacíos/nil
- [ ] 2.3 `pkg/providers/claude_provider.go` — implementar `ChatStream(ctx, messages, opts)` usando `p.client.Messages.NewStreaming()`; detectar `block.Type == "thinking"` del Anthropic SDK; emitir `StreamChunk{ThinkingDelta: text}` por cada thinking delta; seguir patrón de `ollama_provider.go`
- [ ] 2.4 `pkg/storage/user_storage.go` — agregar campo `ExtendedThinking bool` a `UserConfig` (default `false`); agregar migración de schema si es necesario (ALTER TABLE)
- [ ] 2.5 `pkg/web/server.go` — en `handleChatWS()`, activar callback `OnThinking` condicionado a `session.ExtendedThinking`; emitir WS event `{"type":"thinking_delta","content":"..."}` solo si flag activo y modelo soporta thinking; verificar o crear endpoint `PUT /api/v1/user/config`
- [ ] 2.6 `chatStore.js` — agregar acción `appendThinkingDelta(content)`; acumular en `msg.thinkingBlocks[{id, content, expanded, finalized, manuallyToggled}]`; en `endStreamingMessage()` colapsar bloques donde `!block.manuallyToggled`
- [ ] 2.7 `ChatView.vue` — agregar handler para evento WS `thinking_delta`; llamar `chatStore.appendThinkingDelta(event.content)`
- [ ] 2.8 `components/Chat/ThinkingBlock.vue` — crear componente nuevo: `<details>` animado, brain icon 🧠, badge `agentName`, contenido italic gris; computed `isExpanded = props.thinkingBlock.expanded || props.isStreaming`; CSS transitions (no JS timers); toggle manual que setea `block.manuallyToggled = true`
- [ ] 2.9 `MessageBubble.vue` — renderizar `msg.thinkingBlocks[]` antes del content principal usando `<ThinkingBlock>` por cada bloque; no renderizar si array vacío
- [ ] 2.10 `Settings/ProfileSettingsTab.vue` — agregar toggle "Extended Thinking (Claude only)"; descripción de costo adicional; persistir via `PUT /api/v1/user/config`; solo visible si modelo activo es Claude

## Phase 3 — Testing

- [ ] 3.1 `stores/__tests__/chatStore.spec.js` — `addToolCall()` asigna `agentName` desde `currentAgent`; fallback a `'main'` cuando null; `expanded = true` solo con `streaming && started`; `updateCurrentAgent()` actualiza state; `appendThinkingDelta()` acumula en bloque activo; `endStreamingMessage` colapsa bloques no toggled (FR specs: Edge-1, F1-AC-7, F2-AC-5)
- [x] 3.2 `components/__tests__/ToolCallItem.spec.js` — renderiza expandido con `streaming=true` + `status='started'`; colapsa al cambiar a `'finished'`; toggle manual funciona; badge semántico por status (F1-AC-2, F1-AC-3, F1-AC-4)
- [ ] 3.3 `components/__tests__/ThinkingBlock.spec.js` — auto-expande durante streaming; colapsa en `stream_end` si no toggled; toggle manual setea `manuallyToggled`; badge agentName visible; estilo correcto (F2-AC-2, F2-AC-4)
- [ ] 3.4 `e2e/agent-visibility.spec.js` — histórico: todos los tool calls colapsados (F1-AC-1); toggle extended thinking persiste en reload (F2-AC-3); multi-agent: TeamActivityPanel muestra tool calls por agente (F1-AC-5); default user: sin ThinkingBlock en DOM (F2-AC-1)

## Phase 4 — Documentation & Cleanup

- [ ] 4.1 `README.md` — agregar sección "Agent Visibility": smart collapse de tool calls + Extended Thinking opt-in
- [ ] 4.2 `CHANGELOG.md` — Added: real-time tool call visibility con smart collapse; Added: Extended Thinking visualization (Claude, opt-in); Improved: TeamActivityPanel con tool calls por agente

---

## Dependencies

| Tarea | Depende de |
|-------|------------|
| 1.4 | 1.3 — ToolCallItem debe aceptar prop `msg` antes de pasarla |
| 1.5 | 1.1 — `agentName` debe estar en tool calls antes de filtrar |
| 1.6 | 1.1 — `agentName` debe estar disponible para `groupBy` |
| 2.3 | 2.1 — `StreamChunk.ThinkingDelta` debe existir |
| 2.2 | 2.1 — `StreamChunk.ThinkingDelta` debe existir |
| 2.5 | 2.2, 2.4 — `OnThinking` y `ExtendedThinking` deben existir |
| 2.8 | 2.6 — `thinkingBlocks` debe estar en store antes de crear el componente |
| 2.9 | 2.8 — `ThinkingBlock.vue` debe existir antes de usarlo en `MessageBubble` |
| 3.2 | 1.3 — refactor de `ToolCallItem` debe estar completo |
| 3.4 | 1.6, 2.9 — UI completa antes de E2E |

## Implementation Order

**Batch 1 (Fase 1):** 1.1 → 1.2 → 1.3 → 1.4 → 1.5 → 1.6 → 1.7  
**Batch 2 (Fase 2 Backend):** 2.1 → 2.2 → 2.3 → 2.4 → 2.5  
**Batch 3 (Fase 2 Frontend):** 2.6 → 2.7 → 2.8 → 2.9 → 2.10  
**Batch 4 (Testing):** 3.1 → 3.2 → 3.3 → 3.4  
**Batch 5 (Docs):** 4.1 → 4.2  

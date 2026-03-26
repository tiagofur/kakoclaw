# Proposal: Multi-Agent Chat Unification

**Change:** `multi-agent-chat-unification`
**Date:** 2026-03-25
**Status:** Approved for Spec

---

## Intent

Los usuarios ven "chats fantasma" (`specialist_developer`, `specialist_researcher`) en el sidebar porque cada especialista crea su propia sesión en DB. Además, los reportes de especialistas están ocultos y el estado del orchestrator se borra prematuramente. El resultado es una UX fragmentada y confusa.

Este cambio unifica la experiencia: las delegaciones ocurren en el mismo chat, los reportes de especialistas son visibles como accordiones colapsables, y el orchestrator tiene un indicador visual de síntesis.

---

## Scope

### In Scope — Fase 1 (Quick Wins)

- Filtrar sesiones `specialist_*` en `ListSessionsForUser()` (`pkg/storage/chat.go`)
- Mostrar `msg.segments` como accordiones colapsables en `MessageBubble.vue`
- Emitir eventos WS `synthesis_start` / `synthesis_end` en `server.go`
- Diferir `clearAgentStatus()` hasta que el usuario navegue o inicie nuevo chat (`ChatView.vue`)

### In Scope — Fase 2 (Medium Term)

- Propagar `sessionKey` del usuario a `processSpecialistTask()` (`orchestrator.go`)
- Llamar `aggregateReports()` en el flujo estándar, no solo en `ProcessWithFeedbackLoop()` (`orchestrator.go`)

### Out of Scope

- Reestructurar el schema de la tabla `sessions` (Fase 2+)
- Migración de datos de sesiones existentes
- Refactor completo de sub-threads por agente (Approach 3 — visión a largo plazo)

---

## Approach

**Fase 1** usa Approach 2 + Approach 4 de la exploración (bajo riesgo, alto impacto visible):

1. **Filtro de lectura** en `ListSessionsForUser()`: `.Where("session_id NOT LIKE 'specialist_%'")` — una línea, sin writes.
2. **Accordiones en `MessageBubble.vue`**: Usar el patrón de `ToolCallItem` existente para renderizar `msg.segments` colapsados con label del especialista.
3. **Eventos WS de síntesis**: `server.go` emite `{"type": "synthesis_start"}` antes de la llamada final del orchestrator LLM y `{"type": "synthesis_end"}` al completar.
4. **`clearAgentStatus()` diferida**: Mover la llamada de `stream_end` a los hooks de navegación (`beforeRouteLeave`) y al inicio de `sendMessage()`.

**Fase 2** requiere pasar el `sessionKey` via `context.Context` usando una clave tipada, luego extraerlo en `processSpecialistTask()` y pasarlo al especialista. El especialista guarda en la misma sesión con metadata `specialist_name`.

---

## Affected Areas

| Área | Impacto | Descripción |
|------|---------|-------------|
| `pkg/storage/chat.go` | Modified | Filtrar `specialist_*` en `ListSessionsForUser()` |
| `pkg/web/server.go` | Modified | Emitir `synthesis_start` / `synthesis_end` |
| `pkg/agent/orchestrator.go` | Modified (Fase 2) | Propagar sessionKey; llamar `aggregateReports()` en flujo estándar |
| `pkg/agent/specialist.go` | Modified (Fase 2) | Recibir sessionKey del usuario en `ProcessWithSpeciality()` |
| `pkg/web/frontend/src/stores/chatStore.js` | Modified | Diferir `clearAgentStatus()`; manejar eventos synthesis |
| `pkg/web/frontend/src/views/ChatView.vue` | Modified | Reubicar `clearAgentStatus()`; mostrar indicator "synthesizing" |
| `pkg/web/frontend/src/components/MessageBubble.vue` | Modified | Accordiones colapsables para `msg.segments` |

---

## Risks

| Riesgo | Probabilidad | Mitigación |
|--------|-------------|------------|
| Filtro `specialist_*` oculta sesiones útiles | Baja | Las respuestas siguen visibles como accordiones en el chat principal |
| `clearAgentStatus()` tardía consume memoria en chats largos | Baja | Limpiar en navegación/nuevo-chat, no acumular entre mensajes |
| Eventos `synthesis_start/end` no manejados en todos los clientes | Muy baja | Frontend ignora tipos de evento desconocidos (handler defensivo) |
| Propagar sessionKey en Fase 2 contamina contexto del especialista | Media | Feature-flag para Fase 2; probar en staging antes de producción |

---

## Rollback Plan

- **Fase 1** (filtro + frontend): Revertir el filtro SQL y los cambios de `MessageBubble.vue` / `ChatView.vue`. Sin migración de datos requerida.
- **Fase 2** (sessionKey propagation): Revertir el parámetro en `processSpecialistTask()`. El fallback a `"specialist_{name}"` restaura el comportamiento previo.
- Ambas fases son additive — no hay breaking changes de schema ni API pública.

---

## Dependencies

- Fase 1: Ninguna. Cambios de lectura en storage + UI.
- Fase 2: Depende de que Fase 1 esté estable; requiere entender el impacto en contexto LLM del especialista antes de implementar.

---

## Success Criteria

### Fase 1

- [ ] Las sesiones `specialist_*` NO aparecen en el sidebar
- [ ] Los reportes de especialistas se muestran como accordiones colapsables en el chat principal
- [ ] El usuario puede expandir cada accordión para ver el reporte completo del especialista
- [ ] El `agentHistory` persiste hasta que el usuario navega o inicia un nuevo chat
- [ ] El frontend muestra "Orchestrator synthesizing..." al recibir `synthesis_start`

### Fase 2

- [ ] Los especialistas reciben el historial del usuario en su contexto LLM
- [ ] `aggregateReports()` se invoca en el flujo estándar de delegación
- [ ] El resumen unificado del orchestrator referencia correctamente las contribuciones de cada especialista

---

## Notes

- Non-breaking para deployments multi-usuario: el filtro es por `session_id` prefix, no por `user_id`.
- Fase 2 es un cambio de lógica de negocio, no de schema — no requiere migración de DB.

# Exploration: Multi-Agent Chat Unification

**Change:** `multi-agent-chat-unification`  
**Date:** 2026-03-25  
**Status:** Ready for Proposal

---

## Current State

### Flujo actual de multi-agentes (de extremo a extremo)

```
Usuario → ChatView (sendMessage) → WebSocket /ws/chat
         → server.go handleChatWS()
            → AgentManager.GetActiveAgent() → orchestrator.AgentLoop
            → ProcessDirectWithUserAndModelStream(sessionID="web:chat:xxx")
               → runAgentLoop → llm call → [tool: delegate_to_specialist]
                  → DelegationTool.Execute()
                     → processSpecialistTask(specialistName)
                        → specialist.ProcessWithSpeciality()
                           → ProcessDirect(sessionKey="specialist_{name}")  ← SESIÓN SEPARADA
                        → emitContentSegment() → WS "content_segment"
                        → emitSpecialistReport() → WS "specialist_report"
               → [orchestrator LLM produce respuesta final]
               → WS "stream_end" con content=synthesized_result
```

### Qué existe HOY y funciona

#### Backend (completo e implementado)

- **`pkg/agent/orchestrator.go`** — El `OrchestratorAgent` con `DelegationTool` que llama especialistas
- **`pkg/agent/specialist.go`** — `SpecialistAgent` con `ProcessWithSpeciality()` y `RequestColleagueTool`
- **`pkg/agent/swarm.go`** — `SwarmRunner` con modos `sequential`, `parallel`, `consensus`
- **`pkg/agent/manager.go`** — `AgentManager` que inicializa y selecciona el agente activo
- **`pkg/web/server.go`** — WS callbacks para `agent_status`, `content_segment`, `specialist_report`, `delegation_update`

#### Frontend (parcialmente implementado)

- **`chatStore.js`** — Maneja todos los eventos: `addAgentEvent`, `addContentSegment`, `addSpecialistReport`, `updateDelegationProgress`
- **`ChatView.vue`** — `handleMessage()` procesa los 8+ tipos de eventos WS
- **`MessageBubble.vue`** — Renderiza `agentActivities`, `segments`, `toolCalls`, `thinkingBlocks`

---

## Problemas Identificados

### Problema 1: CAUSA RAÍZ — Los especialistas crean sesiones separadas

**Archivo:** `pkg/agent/specialist.go`, línea 285

```go
// ProcessWithSpeciality
return sa.ProcessDirect(agentCtx, fullMessage, fmt.Sprintf("specialist_%s", sa.name))
```

El `ProcessDirect` usa `"specialist_{name}"` como `sessionKey`. Este sessionKey es la clave de sesión en la DB (`pkg/storage/chat.go`). **Cada especialista crea su propio registro de sesión** en la tabla `sessions` con `session_id = "specialist_developer"`, `"specialist_researcher"`, etc.

Cuando el frontend llama `GET /api/v1/chat/sessions`, estas sesiones de especialistas aparecen en la sidebar como chats separados. El usuario ve:
- `web:chat:abc123` — su conversación real
- `specialist_developer` — chat fantasma creado por el especialista
- `specialist_researcher` — otro chat fantasma

### Problema 2: El orchestrator sintetiza, pero el streaming está fragmentado

El flujo completo es:

1. El orchestrator recibe el mensaje del usuario → inicia streaming (`stream_start`)
2. Llama al especialista → emite `content_segment` con la respuesta del especialista
3. El orchestrator LLM toma el resultado del especialista como tool result
4. El orchestrator produce su propia respuesta final → tokens van al `stream` principal
5. `stream_end` lleva el contenido final del orchestrator

**El problema:** El `content_segment` del especialista **no se muestra visiblemente** en el chat — se almacena en `msg.segments` pero solo es visible colapsado. El usuario ve primero la respuesta del orchestrator (que puede ser breve/incompleta) y los detalles del especialista están ocultos bajo un toggle.

### Problema 3: El seguimiento del orchestrator está incompleto en el frontend

En `ChatView.vue` `handleMessage()`, cuando llega `stream_end`:

```js
chatStore.endStreamingMessage(message.content || '', message.agents || [])
chatStore.clearAgentStatus() // Limpia inmediatamente
```

El `clearAgentStatus()` borra `agentHistory`, `specialistReports`, `delegationChain` antes de que el usuario pueda ver el resumen. **El snapshot se guarda en `msg.agentActivity`, pero la UI nunca muestra este snapshot de forma prominente post-streaming.**

### Problema 4: Los especialistas NO reciben el sessionKey del chat principal

Cuando `processSpecialistTask()` llama a `specialist.ProcessWithSpeciality(ctx, task)`, el ctx no transporta el sessionKey del usuario (`"web:chat:xxx"`). El especialista usa su propio `"specialist_{name}"`. 

Esto significa:
- El historial del especialista está **aislado** del historial del chat principal
- El especialista no tiene contexto de la conversación previa del usuario
- Si hay múltiples delegaciones en la misma sesión, el especialista sí recuerda su propia historia (pero no la del usuario)

### Problema 5: No hay síntesis/resumen unificado al final

El `aggregateReports()` existe en `orchestrator.go` pero solo se llama desde `ProcessWithFeedbackLoop()`. El flujo estándar (`ProcessWithSpeciality` → tool call) **no invoca** `aggregateReports()`. La síntesis depende completamente del LLM del orchestrator en su respuesta final.

Cuando el LLM del orchestrator no genera un buen resumen (porque el resultado del especialista es largo), el usuario ve una respuesta incompleta.

---

## Affected Areas

- `pkg/agent/specialist.go` — `ProcessWithSpeciality()`: sessionKey hardcodeado como `"specialist_{name}"`
- `pkg/agent/orchestrator.go` — `processSpecialistTask()`: no pasa sessionKey del usuario; síntesis solo en `ProcessWithFeedbackLoop`
- `pkg/web/server.go` — `handleChatWS()`: no hay evento WS para "synthesis_start" / "synthesis_complete"
- `pkg/web/frontend/src/views/ChatView.vue` — `handleMessage()`: `clearAgentStatus()` prematura; no hay componente de resumen post-delegación
- `pkg/web/frontend/src/stores/chatStore.js` — `endStreamingMessage()`: borra agentActivity inmediatamente
- `pkg/web/frontend/src/components/MessageBubble.vue` — No muestra el resumen de multi-agent prominentemente
- `pkg/storage/chat.go` — `ListSessionsForUser()`: devuelve sesiones internas `specialist_*` al frontend

---

## Approaches

### Approach 1: Session Inheritance (Recomendado)

**Descripción:** Pasar el sessionKey del usuario al contexto de Go para que los especialistas guarden sus mensajes en el mismo session, pero en un "sub-thread" identificado con metadata.

- **Cambio backend:** Agregar `sessionKeyFromCtx(ctx)` helper. En `processSpecialistTask`, extraer el sessionKey del `TeamContext` y pasarlo al especialista. El especialista guarda en el mismo session con metadata `{"role": "specialist", "specialist_name": "developer"}`.
- **Cambio storage:** Filtrar sesiones internas (`specialist_*`) en `ListSessionsForUser()`.
- **Pros:** Los especialistas mantienen contexto del usuario; no hay chats fantasma; historial unificado
- **Cons:** Requiere cambios en el schema de storage para metadata; la historia de especialistas puede contaminar el contexto del orchestrator
- **Effort:** Medium

### Approach 2: Suppress + Enhance Display (Quick Win)

**Descripción:** No cambiar el backend de sessions — solo filtrar en la API y mejorar el frontend para mostrar mejor las respuestas de especialistas.

- **Cambio backend:** En `handleChatSessions`, filtrar sesiones cuyo `session_id` empiece por `specialist_` 
- **Cambio frontend:** Mostrar `msg.segments` prominentemente debajo de la respuesta del orchestrator con labels por especialista
- **Cambio frontend:** Agregar `SynthesisBlock` que muestre el resumen del orchestrator después de todas las contribuciones
- **Pros:** Bajo riesgo; preserva el aislamiento de memoria de especialistas; implementación rápida
- **Cons:** Los chats separados siguen existiendo en DB (solo ocultos); el contexto de especialista sigue aislado
- **Effort:** Low

### Approach 3: Full Unification (Ideal pero costoso)

**Descripción:** Refactorizar para que toda la ejecución multi-agente ocurra en el mismo "chat thread" con mensajes etiquetados por agente.

- Nuevo tipo de mensaje: `{ role: "specialist", agent_name: "developer", content: "..." }`
- `SessionKey` unificado propagado por todo el delegation chain
- Frontend renderiza mensajes por agente como sub-burbujas dentro del mismo chat
- **Pros:** Experiencia unificada ideal; historial compartido real; contexto completo para todos los agentes
- **Cons:** Cambios breaking en schema; requiere migración de DB; alto riesgo de regressions; +3 semanas de trabajo
- **Effort:** High

### Approach 4: Synthesis WS Event (Complemento a cualquier approach)

**Descripción:** Agregar evento WS `synthesis_start` / `synthesis_end` para que el frontend sepa cuándo el orchestrator está sintetizando, y mostrar un indicador visual + bloque de resumen.

- Backend emite `{"type": "synthesis_start", "agents_involved": ["developer"]}` antes de la respuesta final
- Frontend muestra "Orchestrator synthesizing..." y luego un `SynthesisBlock` con el resumen
- **Pros:** Mejora la UX del seguimiento; fácil de implementar; ortogonal al approach de sessions
- **Cons:** No resuelve el problema de sesiones separadas por sí solo
- **Effort:** Low

---

## Recommendation

**Implementar Approach 2 (Quick Win) + Approach 4 (Synthesis Events) en paralelo.**

**Razón:**
1. El filtro de sesiones es una línea de código y resuelve el 80% del problema visible (chats fantasma)
2. Mejorar el display de `segments` y agregar `SynthesisBlock` da la experiencia unificada sin riesgo
3. El evento WS `synthesis_start/end` da seguimiento claro al orchestrator

**Approach 1** (Session Inheritance) puede hacerse después como mejora incremental para que los especialistas tengan contexto del usuario.

**Approach 3** (Full Unification) es la visión a largo plazo pero requiere una rewrite más cuidadosa.

---

## Gap Analysis

| Funcionalidad | Backend | Frontend | Gap |
|---|---|---|---|
| Eventos WS para delegación | ✅ Completo | ✅ Recibido | Síntesis no notificada |
| Specialist content_segment | ✅ Emitido | ⚠️ Colapsado | Visibilidad insuficiente |
| Orchestrator synthesis | ⚠️ Solo en feedback loop | ❌ Sin bloque dedicado | Falta por completo |
| Session unification | ❌ Sessions separadas | ❌ Chats fantasma en sidebar | Problema crítico |
| Filtro sessions internas | ❌ No existe | N/A | Fácil de agregar |
| Seguimiento post-delegación | ⚠️ clearAgentStatus prematuro | ❌ Sin snapshot visible | Estado se pierde |
| Contexto usuario → especialista | ❌ No propagado | N/A | Especialistas sin historial |

---

## Risks

- **Filtrar `specialist_*` en API**: Riesgo bajo — es un filtro de lectura, no modifica datos
- **Cambiar sessionKey en especialistas**: Riesgo medio — puede afectar el contexto de LLM del especialista si de repente recibe historia del usuario que no esperaba
- **Agregar `synthesis_start/end`**: Riesgo muy bajo — nuevo evento WS, frontend lo ignora si no lo maneja
- **Modificar `endStreamingMessage`**: Riesgo medio — afecta el estado global de loading; necesita tests

---

## Key Files Summary

| Archivo | Relevancia |
|---|---|
| `pkg/agent/orchestrator.go` | Delegation tool, processSpecialistTask, aggregateReports |
| `pkg/agent/specialist.go` | ProcessWithSpeciality, sessionKey hardcodeado (L285) |
| `pkg/agent/manager.go` | AgentManager, GetActiveAgent, InitializeOrchestrator |
| `pkg/agent/swarm.go` | SwarmRunner, aggregación de resultados |
| `pkg/agent/loop.go` | ProcessDirect*, runAgentLoop, sessionKey propagation |
| `pkg/web/server.go` | handleChatWS, WS callbacks, ListSessions API |
| `pkg/storage/chat.go` | Session schema, ListSessionsForUser |
| `pkg/web/frontend/src/stores/chatStore.js` | addContentSegment, endStreamingMessage, clearAgentStatus |
| `pkg/web/frontend/src/views/ChatView.vue` | handleMessage, stream_end handling |
| `pkg/web/frontend/src/components/MessageBubble.vue` | Renderizado de segments, agentActivities |

---

## Ready for Proposal

**Sí.** Los problemas están claramente identificados con sus causas raíz y los archivos exactos a modificar. Se pueden proponer cambios concretos y bien delimitados.

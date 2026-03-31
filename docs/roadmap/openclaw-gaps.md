# Roadmap de Gaps vs OpenClaw

Documento de referencia para cerrar los gaps identificados en el análisis comparativo MakoClaw vs OpenClaw.
Cada gap está documentado como un cambio SDD listo para ejecutar con `/sdd:new <change-name>`.

**Última actualización:** 2026-03-31
**Estado del análisis:** `docs/superpowers/specs/2026-03-31-memory-flush-pre-compaction-design.md`

---

## Estado General

| # | Gap | Prioridad | Esfuerzo | Estado |
|---|-----|-----------|---------|--------|
| 1 | Memory semántica (vector/híbrida) | 🔴 Alta | Alto | Pendiente |
| 2 | Memory flush pre-compaction | 🔴 Alta | Bajo | ✅ **Cerrado** (2026-03-31) |
| 3 | Multi-provider web search | 🔴 Alta | Medio | ✅ **Cerrado** (2026-03-31) |
| 4 | API key rotation + model failover | 🔴 Alta | Medio | Pendiente |
| 5 | Browser control (CDP real) | 🔴 Alta | Alto | Pendiente |
| 6 | Hook system (event-driven) | 🟡 Media | Medio | Pendiente |
| 7 | DM Pairing para nuevos senders | 🟡 Media | Bajo | Pendiente |
| 8 | Voice TTS en respuestas | 🟡 Media | Medio | Pendiente |
| 9 | Session pruning | 🟡 Media | Bajo | ✅ **Cerrado** (2026-03-31) |
| 10 | Pluggable context engines | 🟡 Media | Medio | Pendiente |
| 11 | Auth profiles multi-OAuth | 🟡 Media | Medio | Pendiente |
| 12 | Tailscale integration | 🟢 Baja | Medio | Pendiente |
| 13 | Más canales (14 faltantes) | 🟢 Baja | Alto | Pendiente |
| 14 | Background tasks con delivery | 🟢 Baja | Bajo | Pendiente |
| 15 | Standing orders | 🟢 Baja | Bajo | Pendiente |

---

## 🔴 Alta Prioridad

---

### GAP-1: Memory Semántica (Vector + Keyword Search)

**SDD Change:** `memory-semantic-search`
**Comando:** `/sdd:new memory-semantic-search`

**Problema actual:**
`pkg/agent/memory.go` guarda y lee MEMORY.md como texto plano. El agente lo lee completo en cada sesión. Con memorias grandes, consume tokens innecesarios y no puede encontrar información específica.

**Lo que hace OpenClaw:**
- SQLite con FTS5 (full-text search nativo)
- Embeddings vectoriales (auto-detecta OpenAI / Gemini / Voyage keys)
- Búsqueda híbrida: keyword + vector ranking
- 3 backends swappeables: `memory-core` (local SQLite), `memory-qmd` (sidecar con reranking), `memory-honcho` (cloud)

**Lo que queremos construir:**
1. Indexar MEMORY.md y daily notes en SQLite con FTS5
2. Tool `memory_search` que busca por keyword (sin embeddings = cero dependencias extra)
3. Optionally: embeddings via el mismo provider LLM ya configurado (si tiene soporte)
4. Mantener compatibilidad total con el formato actual de archivos Markdown

**Archivos clave:**
- `pkg/agent/memory.go` — MemoryStore actual (solo filesystem)
- `pkg/storage/sqlite.go` — patrón de DB a seguir
- `pkg/tools/` — donde va el nuevo `memory_search` tool
- `pkg/agent/context.go` — ContextBuilder carga la memoria (a integrar búsqueda)

**Dependencias:** Ninguna.
**Desbloquea:** GAP-3 (si usamos el search para contexto de sesión).

---

### GAP-2: Memory Flush Pre-Compaction ✅ CERRADO

**Cerrado:** 2026-03-31
**Commits:** `c02cbf6`, `06417ce`, `81c3a7a`, `a740cc5`
**Spec:** `docs/superpowers/specs/2026-03-31-memory-flush-pre-compaction-design.md`

Antes de cada compaction, el agente ejecuta un turno silencioso con tools de memoria solamente. Configurable via `memory_flush_before_compaction` en `config.json` (default: `true`).

---

### GAP-3: Multi-Provider Web Search

**SDD Change:** `multi-provider-web-search`
**Comando:** `/sdd:new multi-provider-web-search`

**Problema actual:**
`pkg/tools/web.go` tiene `web_search` hardcodeado a Brave Search API. Si el usuario no tiene Brave key o Brave falla, el tool no funciona.

**Lo que hace OpenClaw:**
- Providers: Brave, Tavily, Perplexity, Google, DuckDuckGo
- Selección configurable por defecto o por skill/sesión
- Fallback automático si el provider primario falla

**Lo que queremos construir:**
1. Interface `WebSearchProvider` (similar a `LLMProvider`)
2. Implementaciones: `BraveProvider` (ya existe, refactorizar), `TavilyProvider`, `DuckDuckGoProvider` (sin key)
3. Config `tools.web_search.providers[]` con orden de prioridad y fallback
4. Factory que selecciona el provider disponible según keys configuradas

**Archivos clave:**
- `pkg/tools/web.go` — implementación actual a refactorizar
- `pkg/config/config.go` — `ToolsConfig` donde va la nueva config
- `pkg/providers/` — patrón de interface a seguir

**Dependencias:** Ninguna.

---

### GAP-4: API Key Rotation + Model Failover

**SDD Change:** `provider-resilience`
**Comando:** `/sdd:new provider-resilience`

**Problema actual:**
Cada provider tiene una sola API key. Si falla (rate limit, error 429), el agente muere. No hay retry con key alternativa ni fallback a otro modelo.

**Lo que hace OpenClaw:**
- `PROVIDER_API_KEYS` acepta múltiples keys separadas por coma
- Rota automáticamente al siguiente si la actual falla con 429/503
- Model failover: si el modelo primario no responde, usa el fallback configurado

**Lo que queremos construir:**
1. `ProvidersConfig` acepta `api_keys []string` además de `api_key string`
2. `HTTPProvider` / `ClaudeProvider` intentan keys en orden round-robin o en error
3. Config `agents.defaults.fallback_model` para failover de modelo
4. Lógica de retry en el provider layer, transparente al agent loop

**Archivos clave:**
- `pkg/providers/http_provider.go` — provider HTTP principal
- `pkg/providers/claude_provider.go` — Claude provider
- `pkg/config/config.go` — `ProvidersConfig` a extender

**Dependencias:** Ninguna.

---

### GAP-5: Browser Control (CDP Real)

**SDD Change:** `browser-control-cdp`
**Comando:** `/sdd:new browser-control-cdp`

**Problema actual:**
`pkg/tools/browser.go` está en etapa temprana. `web_fetch` solo hace HTTP GET pasivo — no puede hacer click, login, llenar formularios, ni interactuar con apps dinámicas.

**Lo que hace OpenClaw:**
- Chromium CDP via Puppeteer
- Comandos: navigate, click, fill, screenshot, scroll, wait
- Manejo de autenticación (cookies, localStorage)
- Uploads de archivos

**Lo que queremos construir:**
1. Instanciar Chrome/Chromium headless via `chromedp` (Go) o `rod`
2. Tool `browser` con actions: `navigate`, `click`, `fill`, `screenshot`, `get_text`
3. Session persistence de cookies entre llamadas
4. Timeout y cleanup automático del browser

**Archivos clave:**
- `pkg/tools/browser.go` — a reescribir/completar
- `pkg/agent/loop.go` — registro del tool

**Dependencias:** Ninguna (usa library externa como `chromedp` o `go-rod`).
**Nota:** Requiere Chromium instalado en el sistema — documentar como dependencia opcional.

---

## 🟡 Media Prioridad

---

### GAP-6: Hook System (Event-Driven)

**SDD Change:** `hook-system`
**Comando:** `/sdd:new hook-system`

**Problema actual:**
No hay forma de ejecutar lógica arbitraria cuando ocurren eventos del agente (nueva sesión, reset, stop, mensaje enviado). Las automatizaciones son solo cron (time-based).

**Lo que hace OpenClaw:**
- Hooks en archivos TypeScript/JS en `hooks/` del workspace
- Eventos: `/new`, `/reset`, `/stop`, `message.sent`, `session.start`, `session.end`
- El hook recibe el evento completo (mensaje, session key, canal)
- Puede modificar el comportamiento o ejecutar side effects

**Lo que queremos construir:**
1. Interface `Hook` con `EventType` y `Execute(ctx, event)`
2. `HookRegistry` que carga hooks desde `<workspace>/hooks/*.go` compilados en runtime, o más simple: `hooks/*.sh` scripts
3. Eventos iniciales: `session.new`, `session.reset`, `message.received`, `message.sent`
4. Config para enable/disable hooks por tipo

**Archivos clave:**
- `pkg/hooks/` — existe pero limitado (revisar)
- `pkg/agent/loop.go` — puntos de inyección de eventos
- `pkg/channels/` — eventos de mensajes

**Dependencias:** Ninguna.

---

### GAP-7: DM Pairing para Nuevos Senders

**SDD Change:** `dm-pairing`
**Comando:** `/sdd:new dm-pairing`

**Problema actual:**
Los canales usan allowlist estática. Un sender desconocido es rechazado directamente sin posibilidad de onboarding. No hay mecanismo para que nuevos usuarios se registren dinámicamente.

**Lo que hace OpenClaw:**
- Cuando llega un DM de sender desconocido, envía un código de emparejamiento de 4-6 caracteres
- El sender debe responder con el código para quedar habilitado
- Policy configurable: `pairing` (default), `open` (todos permitidos), `closed` (ninguno nuevo)
- Códigos expiran después de N minutos

**Lo que queremos construir:**
1. `PairingPolicy` enum en `BaseChannel`: `open`, `pairing`, `closed`
2. `PairingStore` en SQLite con `(channel, sender_id, code, expires_at, status)`
3. Cuando sender desconocido escribe: genera código, responde con instrucciones
4. Cuando sender responde el código correcto: lo agrega a la allowlist del usuario

**Archivos clave:**
- `pkg/channels/base.go` — BaseChannel, donde vive allowlist
- `pkg/storage/sqlite.go` — nueva tabla `pairing_codes`
- `pkg/config/config.go` — `pairing_policy` por canal

**Dependencias:** Ninguna.

---

### GAP-8: Voice TTS en Respuestas

**SDD Change:** `voice-tts-responses`
**Comando:** `/sdd:new voice-tts-responses`

**Problema actual:**
MakoClaw tiene STT (Groq transcription) para entrada de voz en Telegram, pero no hay TTS para las respuestas. El agente siempre responde texto.

**Lo que hace OpenClaw:**
- TTS via ElevenLabs (alta calidad) o system TTS como fallback
- Configurable por canal (Telegram puede recibir audio, Discord también)
- Talk Mode: sesión continua de voz

**Lo que queremos construir:**
1. Interface `TTSProvider` con `Synthesize(ctx, text, voice) ([]byte, error)`
2. Implementaciones: `OpenAITTSProvider` (ya tiene SDK), `SystemTTSProvider` (fallback)
3. Config `tools.tts.provider` y `tools.tts.voice`
4. En canales que soportan audio (Telegram, Discord), enviar respuesta como audio cuando está habilitado

**Archivos clave:**
- `pkg/voice/` — existe para STT, extender para TTS
- `pkg/channels/telegram.go` — envío de audio
- `pkg/channels/discord.go` — envío de audio
- `pkg/config/config.go` — config de TTS

**Dependencias:** Ninguna.

---

### GAP-9: Session Pruning

**SDD Change:** `session-pruning`
**Comando:** `/sdd:new session-pruning`

**Problema actual:**
Cuando la sesión supera el threshold, se summariza todo. No hay opción de reducir tokens recortando solo los tool outputs (que suelen ser los más grandes) sin perder el hilo de la conversación.

**Lo que hace OpenClaw:**
- Antes de summarizar, recorta tool outputs largos del historial
- Reemplaza el contenido del tool result con `[output truncado, N chars]`
- Configurable por threshold de tamaño por tool result

**Lo que queremos construir:**
1. Función `pruneToolOutputs(history []Message, maxOutputLen int) []Message`
2. Llamada antes de `maybeSummarize` en `loop.go`
3. Config `agents.defaults.max_tool_output_in_history` (default: 2000 chars)
4. Mensajes con `role: "tool"` que superen el límite se truncan con sufijo `[truncated]`

**Archivos clave:**
- `pkg/agent/loop.go` — `maybeSummarize` y `summarizeSession`
- `pkg/session/manager.go` — historial de mensajes

**Dependencias:** Ninguna.
**Nota:** Cambio quirúrgico, baja complejidad. Candidato a implementar junto con otro gap.

---

### GAP-10: Pluggable Context Engines

**SDD Change:** `pluggable-context-engines`
**Comando:** `/sdd:new pluggable-context-engines`

**Problema actual:**
La summarización en `summarizeBatch` está hardcodeada con un prompt fijo y usa el mismo provider/modelo que el agente. No se puede reemplazar ni customizar.

**Lo que hace OpenClaw:**
- Interface `ContextEngine` pluggable
- Se puede reemplazar por un modelo más barato (ej: Haiku para summarizar)
- Parámetros configurables: modelo, prompt de summarización, estrategia

**Lo que queremos construir:**
1. Config `agents.defaults.summarization_model` (permite usar modelo diferente al principal)
2. Config `agents.defaults.summarization_prompt` (prompt customizable)
3. Refactorizar `summarizeBatch` para usar esos parámetros

**Archivos clave:**
- `pkg/agent/loop.go` — `summarizeBatch`, `summarizeSession`
- `pkg/config/config.go` — `AgentDefaults` a extender

**Dependencias:** GAP-4 (provider resilience ayuda pero no es necesario).

---

### GAP-11: Auth Profiles Multi-OAuth

**SDD Change:** `auth-profiles`
**Comando:** `/sdd:new auth-profiles`

**Problema actual:**
Cada provider LLM tiene una sola config de OAuth/API key. Si el usuario quiere usar múltiples cuentas de un mismo provider (ej: dos cuentas de Anthropic), no hay soporte.

**Lo que hace OpenClaw:**
- `auth-profiles.json` con N perfiles por provider
- Rotación manual o automática entre perfiles
- Cooldown entre usos del mismo perfil

**Lo que queremos construir:**
1. `UserProvidersConfig` acepta `profiles []ProviderProfile` además de config base
2. Selección de perfil activo por usuario
3. API endpoints para CRUD de profiles en `pkg/web/providers_handler.go`

**Archivos clave:**
- `pkg/storage/user_providers.go` — storage de provider configs por usuario
- `pkg/config/config.go` — estructura de providers
- `pkg/web/providers_handler.go` — handlers de provider config

**Dependencias:** GAP-4 (key rotation es el siguiente paso natural).

---

## 🟢 Baja Prioridad

---

### GAP-12: Tailscale Integration

**SDD Change:** `tailscale-integration`
**Comando:** `/sdd:new tailscale-integration`

Acceso remoto seguro al Gateway via Tailscale. Auth mode `tailscale` que valida la identidad del cliente por headers de Tailscale.

**Archivos clave:** `pkg/web/server.go`, `pkg/web/auth.go`

---

### GAP-13: Más Canales

**SDD Change:** `channel-matrix`, `channel-teams`, `channel-google-chat`, etc.

Canales faltantes priorizados por demanda:
1. **Matrix** — protocolo abierto, alta demanda en comunidades tech
2. **Microsoft Teams** — enterprise
3. **Google Chat** — Google Workspace
4. **LINE / Zalo** — mercados asiáticos

Cada canal es un SDD separado. Seguir el patrón de `pkg/channels/base.go`.

---

### GAP-14: Background Tasks con Delivery

**SDD Change:** `cron-channel-delivery`
**Comando:** `/sdd:new cron-channel-delivery`

Cuando un cron job completa, notificar al canal de origen que lo disparó. Hoy los resultados de cron van al log pero no vuelven al usuario.

**Archivos clave:** `pkg/cron/`, `pkg/channels/manager.go`, `pkg/storage/sqlite.go`

---

### GAP-15: Standing Orders

**SDD Change:** `standing-orders`
**Comando:** `/sdd:new standing-orders`

Tool `standing_orders` para instrucciones persistentes que aplican a todas las sesiones del usuario. Se inyectan en el context builder como parte del system prompt.

**Archivos clave:** `pkg/agent/context.go`, `pkg/storage/sqlite.go` (nueva tabla), `pkg/tools/`

---

## Orden de Ejecución Recomendado

```
Onda 1 (quick wins, bajo esfuerzo):
  GAP-9  → session-pruning
  GAP-7  → dm-pairing
  GAP-15 → standing-orders
  GAP-14 → cron-channel-delivery

Onda 2 (impacto alto, esfuerzo medio):
  GAP-3  → multi-provider-web-search
  GAP-4  → provider-resilience
  GAP-8  → voice-tts-responses
  GAP-10 → pluggable-context-engines

Onda 3 (proyectos grandes):
  GAP-1  → memory-semantic-search
  GAP-5  → browser-control-cdp
  GAP-6  → hook-system

Onda 4 (nice-to-have):
  GAP-11 → auth-profiles
  GAP-12 → tailscale-integration
  GAP-13 → canales adicionales
```

---

## Cómo usar este documento

Para arrancar con cualquier gap:

```
/sdd:new <change-name>
```

Ejemplo:
```
/sdd:new session-pruning
/sdd:new multi-provider-web-search
/sdd:new provider-resilience
```

Cada SDD pasa por: `explore → propose → spec → design → tasks → apply → verify → archive`.

Para updates rápidos (gaps simples, 1-2 archivos), usar directamente:
```
/sdd:ff <change-name>
```

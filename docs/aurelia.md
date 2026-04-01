## Task Completion Summary

### Specialist Contributions

**tech_research_agent** (Confidence: 82%):
I've gathered extensive information from the core documentation. Let me now compile the complete report from everything I've read.

---

# 📋 REPORTE COMPLETO — Aurelia OS

## 1. Resumen del Proyecto

**Aurelia OS** es un sistema operativo de agentes autónomos escrito en Go. Es **Telegram-nativo**, potenciado por **Claude Code SDK**, y diseñado para ser ligero.

| Atributo      | Detalle                                                                                                                                     |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **Autor**     | `kocar` (GitHub: `github.com/kocar/aurelia`)                                                                                                |
| **Propósito** | Orquestar agentes de IA autónomos accesibles vía Telegram con persistencia, scheduling, multi-proyecto y sesión continua                    |
| **Filosofía** | No re-implementar lo que Claude Code ya hace — orquestarlo: añadir persistencia, scheduling, multi-proyecto y una interfaz Telegram natural |
| **Lenguaje**  | Go 1.25+ (backend) + TypeScript (Bridge)                                                                                                    |
| **Estado**    | Desarrollo activo en branch `feat/aurelia-os`                                                                                               |
| **Madurez**   | ~6,800 líneas Go + ~270 líneas TypeScript, 11 paquetes de tests verdes, suite de tests e2e                                                  |

**Capacidades principales:**

- Conversación natural vía Telegram (texto, fotos, voz, documentos)
- Coding autónomo — lee, escribe, edita archivos, ejecuta comandos
- Multi-proyecto con contextos aislados
- Ejecución asíncrona con respuestas en paralelo
- Continuidad de sesión con tracking de tokens y auto-reset
- Routing inteligente por clasificación LLM
- Scheduler persistente con delivery a Telegram
- Hereda setup de Claude Code (MCPs, skills, plugins, hooks)

---

## 2. Arquitectura General

### Estructura de Módulos

```
cmd/aurelia/              CLI entry point, onboarding, cron CLI, telegram CLI
internal/bridge/          Go ↔ Bridge client (long-lived, multiplexed, bundle embedded via go:embed)
internal/telegram/        Telegram I/O, async pipeline, progress, reactions
internal/session/         Session store and token tracking with auto-reset
internal/agents/          Agent registry (markdown definitions, LLM classification)
internal/persona/         Persona loader (IDENTITY / SOUL / USER)
internal/cron/            Persistent cron scheduler with Telegram delivery
internal/config/          App configuration (providers, Telegram, sessions)
internal/runtime/         Path resolver + instance bootstrap
pkg/stt/                  Speech-to-text (Groq Whisper)
bridge/                   TypeScript Bridge source (esbuild → bundle.js → go:embed)
```

### Patrones de Diseño Identificados

1. **Modular Monolith** — Módulos cohesionados en `internal/`, sin microservicios
2. **Bridge/Adapter Pattern** — TypeScript Bridge como adaptador entre Go y Claude SDK
3. **Declarative Configuration** — Agentes definidos en Markdown con YAML frontmatter
4. **Composition Root** — `cmd/aurelia/app.go` como punto de ensamblaje
5. **Pipeline Pattern** — Procesamiento de mensajes en etapas (extracción → routing → ejecución → respuesta)
6. **Repository Pattern** — SQLite como store persistente para sesiones, cron jobs
7. **Embed Pattern** — Bridge bundle.js embebido en binario Go via `go:embed`
8. **Dependency Injection** — Inyección por constructores, interfaces pequeñas
9. **NDJSON Protocol** — Comunicación Go↔Bridge via stdin/stdout con multiplexing por `request_id`

### Diagrama de Flujo de Datos

```
┌─────────┐     ┌───────────┐     ┌──────────┐     ┌──────────────┐
│  User   │────▶│ Telegram  │────▶│ Pipeline │────▶│ Agent Router │
└─────────┘     └───────────┘     └──────────┘     └──────┬───────┘
                                                          │
                                          ┌───────────────┼───────────────┐
                                          ▼               ▼               ▼
                                   ┌──────────┐   ┌────────────┐   ┌──────────┐
                                   │ Persona  │   │   Bridge   │   │  Cron    │
                                   │ Assembly │   │  (TS/SDK)  │   │ Scheduler│
                                   └──────────┘   └─────┬──────┘   └──────────┘
                                                       │
                                                       ▼
                                               ┌──────────────┐
                                               │ Claude Code  │
                                               │ SDK + Tools  │
                                               │ + MCPs       │
                                               └──────┬───────┘
                                                       │
                                                       ▼
                                               ┌──────────────┐
                                               │  Response    │
                                               │  → Telegram  │
                                               └──────────────┘
```

### Scope Separation (3 niveles)

1. **Repository** — Código fuente, tests, documentación
2. **Local Instance** (`~/.aurelia/`) — Config, SQLite, logs, personas, artifacts
3. **Target Project** — Codebase externo sobre el que actúa el agente

---

## 3. Tecnologías y Dependencias

### Go (Backend)

| Dependencia                          | Versión | Propósito                             |
| ------------------------------------ | ------- | ------------------------------------- |
| `gopkg.in/telebot.v3`                | v3.3.8  | Telegram Bot API                      |
| `modernc.org/sqlite`                 | v1.46.1 | SQLite driver (pure Go, no CGO)       |
| `github.com/robfig/cron/v3`          | v3.0.1  | Cron expression parsing y scheduling  |
| `github.com/knights-analytics/hugot` | v0.6.5  | HuggingFace embeddings (ONNX runtime) |
| `github.com/yuin/goldmark`           | v1.7.8  | Markdown parsing                      |
| `gopkg.in/yaml.v3`                   | v3.0.1  | YAML frontmatter parsing              |
| `github.com/google/uuid`             | v1.6.0  | UUID generation                       |
| `golang.org/x/term`                  | v0.41.0 | Terminal handling                     |

### TypeScript (Bridge)

- **Claude Agent SDK** (`@anthropic-ai/claude-agent-sdk`) — Wrapper del SDK de Anthropic
- **esbuild** — Bundling a `bundle.js` embebido en Go
- Node.js 18+ requerido en runtime

### Proveedores LLM

| Proveedor      | Modo                                             |
| -------------- | ------------------------------------------------ |
| **Anthropic**  | API key o suscripción (OAuth via `claude login`) |
| **Kimi**       | API key (endpoint compatible Anthropic)          |
| **OpenRouter** | API key (proxy multi-modelo)                     |
| **Z.ai**       | API key (GLM Coding Plan)                        |
| **Alibaba**    | API key (Qwen Coding Plan)                       |

### Storage

- **SQLite** (pure Go via modernc.org/sqlite) — Sesiones, cron jobs, estado persistente

### STT

- **Groq Whisper** — Transcripción de mensajes de voz

---

## 4. Módulos Detallados

### 4.1 `internal/bridge/` — Cliente Go ↔ Bridge

**Qué hace:** Gestiona la comunicación entre Go y el proceso TypeScript Bridge. El Bridge es un proceso long-lived que wrappea el Claude Agent SDK.

**Patrones clave:**

- **go:embed** — `bundle.js` compilado se embebida en el binario Go
- **NDJSON Protocol** — Comunicación via stdin/stdout con JSON delimitado por newlines
- **Request Multiplexing** — Múltiples requests concurrentes con `request_id` único

**Protocolo Bridge:**

Go → Bridge (stdin):

```json
{
  "command": "query",
  "request_id": "req-1",
  "prompt": "...",
  "options": {
    "model": "k2.5",
    "system_prompt": "...",
    "cwd": "/path",
    "permission_mode": "bypassPermissions"
  }
}
```

Bridge → Go (stdout):

```json
{"event":"system","request_id":"req-1","session_id":"abc-123","tools":["Read","Write"]}
{"event":"tool_use","request_id":"req-1","name":"Read","input":{"file_path":"src/main.go"}}
{"event":"assistant","request_id":"req-1","text":"The project has..."}
{"event":"result","request_id":"req-1","content":"...","cost_usd":0.12,"session_id":"abc-123"}
```

**Interacción:** Es el único camino hacia LLM — Go nunca llama APIs LLM directamente.

### 4.2 `internal/telegram/` — Interfaz Telegram

**Qué hace:** Recibe eventos de Telegram, adapta input (texto, fotos, voz, documentos), envía respuestas con reply-to, progress y reactions.

**Componentes identificados (por README):**

- Input pipeline asíncrono
- Progreso de herramientas en tiempo real
- Reacciones con emojis contextuales
- Handlers para comandos (`/start`, `/help`, `/cwd`, `/reset`, `/cron`, `/agents`)

**Regla arquitectónica:** Telegram es una capa de interfaz, NO una capa de dominio.

### 4.3 `internal/session/` — Sesiones y Tokens

**Qué hace:** Gestiona sesiones por chat con tracking de tokens y auto-reset.

**Capacidades:**

- Session ID management per chat (warm continue / cold resume)
- Token usage accumulation y cost tracking
- Auto-reset cuando se excede el umbral configurable (`max_session_tokens`)

**Patrón:** Repository pattern con SQLite como backing store.

### 4.4 `internal/agents/` — Registry de Agentes

**Qué hace:** Carga y gestiona agentes definidos como archivos Markdown con YAML frontmatter desde `~/.aurelia/agents/`.

**Formato de agente:**

```markdown
---
name: prospector
description: Busca leads e entra em contato
model: kimi-k2-thinking
schedule: "0 9 * * 1"
cwd: D:\projetos\crm
mcp_servers:
  google-places: { command: "npx google-places-mcp" }
allowed_tools: ["WebSearch", "WebFetch", "Bash"]
---

Voce eh um agente de prospeccao comercial.
```

**Campos:** `name`, `description`, `model` (override), `schedule` (cron), `cwd`, `mcp_servers`, `allowed_tools`

**Routing:** Clasificación LLM que enruta mensajes al agente especialista correcto.

**Interacción:** Agentes con `schedule` se registran automáticamente en `internal/cron/`.

### 4.5 `internal/persona/` — Personalidades

**Qué hace:** Resuelve y ensambla archivos de identidad para construir system prompts.

**Tres archivos en `~/.aurelia/memory/personas/`:**

- `IDENTITY.md` — Nombre, rol, reglas, personalidad
- `SOUL.md` — Tono, estilo, comportamiento
- `USER.md` — Información del usuario, preferencias

**Creación:** Automática via `/start` en Telegram (elige preset "Coder" o "Assistant").

### 4.6 `internal/cron/` — Scheduler Persistente

**Qué hace:** Scheduler persistente con ejecución via Bridge y delivery a Telegram.

**Componentes:**

- Store persistente en SQLite
- Polling cada 15 segundos
- Ejecución via Bridge (con Telegram plugin bloqueado para evitar bot incorrecto)
- Delivery de resultados via TelegramDelivery

**CLI:**

```bash
aurelia cron add "30 8 * * *" "pesquise noticias" --chat-id 123456
aurelia cron once "2026-03-22T09:00:00Z" "gere relatorio" --chat-id 123456
aurelia cron list
aurelia cron del <job-id>
```

### 4.7 `internal/config/` — Configuración

**Qué hace:** Carga y validación de configuración desde `~/.aurelia/config/app.json`.

**Campos principales:**

- `default_provider` / `default_model`
- `providers` (auth_mode, api_keys, base URLs auto-configuradas)
- `telegram_bot_token` / `telegram_allowed_user_ids`
- `stt_provider`, `max_iterations`, `max_session_tokens`

### 4.8 `internal/runtime/` — Runtime y Paths

**Qué hace:** Resolución de paths de instancia y bootstrap del runtime. Separa los 3 scopes (repo, local instance, target project).

### 4.9 `pkg/stt/` — Speech-to-Text

**Qué hace:** Transcripción de voz usando Groq Whisper. Paquete reusable a nivel de `pkg/`.

---

## 5. Sistema de Agentes — Análisis Profundo

### Diseño

Los agentes son **declarativos**, definidos en Markdown, no en código. Esto es una decisión arquitectónica clave:

- **Registry** carga todos los `.md` de `~/.aurelia/agents/`
- **Frontmatter YAML** define metadata (nombre, modelo, schedule, MCPs, tools)
- **Body Markdown** = system prompt del agente
- **Routing LLM** clasifica mensajes y enruta al agente correcto
- **Auto-scheduling** — agentes con campo `schedule` se registran en cron automáticamente

### Flujo de Delegación

```
Mensaje → Pipeline → Agent Router (LLM classification)
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    [General]     [Specialist A]   [Specialist B]
         │               │               │
         └───────────────┴───────────────┘
                         ▼
                   Persona Assembly
                   (IDENTITY + SOUL + agent prompt)
                         ▼
                     Bridge → Claude SDK
```

### SDK de Agentes

El "SDK" es implícito — cada agente hereda:

- **Modelo** (default o override por agente)
- **MCPs** (configurados por agente)
- **Tools permitidas** (`allowed_tools`)
- **Directorio de trabajo** (`cwd`)
- **Schedule** (opcional)

**Constraint actual:** No hay orquestación multi-agente aún (single agent por ejecución).

---

## 6. Sistema de Memoria

### Embeddings

- **HuggingFace embeddings** via `hugot` (ONNX runtime, Go puro)
- Usado para búsqueda semántica y recuperación de contexto

### Persistencia

- SQLite como store principal
- Schema orientado a sesiones y estado

### Decisiones de Diseño (del STYLE_GUIDE)

> "No resolver problemas de continuidad inflando el tamaño del prompt"
> "No reemplazar memoria determinista con comportamiento vector-first como default"

Esto indica un enfoque **híbrido**: memoria determinista (SQLite) como default, embeddings como complemento.

---

## 7. Sistema de Cron/Scheduler

### Arquitectura

```
┌────────────────┐     ┌───────────────┐     ┌──────────────┐
│  Cron Store    │────▶│  Scheduler    │────▶│   Bridge     │
│  (SQLite)     │     │  (15s poll)   │     │  Execution   │
└────────────────┘     └───────┬───────┘     └──────┬───────┘
                               │                    │
                               ▼                    ▼
                       ┌───────────────┐    ┌──────────────┐
                       │  Agent Config │    │   Result     │
                       │  + Persona    │    │   Delivery   │
                       └───────────────┘    └──────┬───────┘
                                                   ▼
                                           ┌──────────────┐
                                           │   Telegram   │
                                           └──────────────┘
```

### Características

- **Persistente** — Jobs sobreviven restarts (SQLite)
- **One-shot y recurrente** — `cron add` (recurrente) y `cron once` (una vez)
- **Telegram plugin bloqueado** durante ejecución de cron para evitar que el agente use el bot incorrecto
- **CLI completo** para gestión de jobs

---

## 8. Sistema de Personalidades/Personas

### Diseño Three-File

| Archivo       | Contenido                         | Propósito          |
| ------------- | --------------------------------- | ------------------ |
| `IDENTITY.md` | Nombre, rol, reglas, personalidad | Quién es el agente |
| `SOUL.md`     | Tono, estilo, comportamiento      | Cómo se expresa    |
| `USER.md`     | Info del usuario, preferencias    | A quién sirve      |

### Flujo de Assembly

1. Se carga la identidad base (IDENTITY)
2. Se mergea con el estilo (SOUL)
3. Se enriquece con contexto del usuario (USER)
4. Se combina con el prompt del agente específico
5. Se envía como system prompt al Bridge

### Presets

- **Coder** — Orientado a desarrollo
- **Assistant** — Orientado a asistencia general

---

## 9. Bridge Frontend-Backend

### Arquitectura del Bridge

```
┌──────────────────────────────────────────┐
│              Go Binary                    │
│  ┌─────────────────────────────────────┐ │
│  │ internal/bridge/                    │ │
│  │  - Start TS process (bundle.js)    │ │
│  │  - Write NDJSON to stdin           │ │
│  │  - Read NDJSON from stdout         │ │
│  │  - Multiplex by request_id         │ │
│  └──────────────┬──────────────────────┘ │
└─────────────────┼────────────────────────┘
                  │ stdin/stdout NDJSON
┌─────────────────▼────────────────────────┐
│         TypeScript Process               │
│  ┌─────────────────────────────────────┐ │
│  │ bridge/index.ts                     │ │
│  │  - Read requests from stdin         │ │
│  │  - Call @anthropic-ai/claude-agent  │ │
│  │  - Stream events back to stdout     │ │
│  │  - Handle concurrent requests       │ │
│  └─────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

### Build Pipeline

```bash
cd bridge && npx esbuild index.ts --bundle --platform=node --target=node18 --outfile=bundle.js --format=esm
cp bundle.js ../internal/bridge/bundle.js  # go:embed picks this up
```

### Eventos del Protocolo

| Evento      | Dirección   | Propósito                                           |
| ----------- | ----------- | --------------------------------------------------- |
| `system`    | Bridge → Go | Info del sistema (session_id, tools disponibles)    |
| `tool_use`  | Bridge → Go | Agente usando herramienta (progreso en tiempo real) |
| `assistant` | Bridge → Go | Texto generado por el asistente (streaming)         |
| `result`    | Bridge → Go | Resultado final (content, cost_usd, session_id)     |
| `error`     | Bridge → Go | Error en ejecución                                  |
| `pong`      | Bridge → Go | Health check response                               |

---

## 10. CI/CD y Calidad

### Herramientas Identificadas

Del README y estructura del repo:

- **`go vet`** — Linting estático
- **`go test ./... -short`** — Suite de tests (11 paquetes)
- **`air`** — Hot reload para desarrollo
- **`.github/`** — Workflows de GitHub Actions (presente, contenido no leído)
- **e2e/** — Tests end-to-end

### Convenciones de Testing (del STYLE_GUIDE)

- **Unit tests** para reglas deterministas pequeñas
- **Integration tests** para persistencia, orquestación, boundaries de providers
- **E2E tests** para flujos críticos de usuario
- **TDD preferido** para reglas de dominio y regresiones
- **Benchmarking medido** — claims de performance requieren datos reales

### Seguridad

- **No commitear secrets**
- **No commitear DBs locales, logs, artifacts**
- **No loggear secrets en plain text**
- **Config de ejemplo** en vez de config real
- **Sanitizar** outputs de benchmarks antes de publicar

---

## 11. Documentación

| Documento              | Contenido                                                              | Insight                            |
| ---------------------- | ---------------------------------------------------------------------- | ---------------------------------- |
| `README.md`            | Overview completo, setup, CLI, arquitectura                            | Producto maduro con thinking claro |
| `AGENTS.md`            | Reglas de workflow, protocolo bridge, formato de agentes               | Guía operativa para desarrollo     |
| `docs/ARCHITECTURE.md` | Shape arquitectónica, boundaries, rules, capabilities/constraints      | Arquitectura explícita y mantenida |
| `docs/STYLE_GUIDE.md`  | Convenciones Go, testing, DI, naming, decisiones que no deben regresar | Disciplina de ingeniería alta      |
| `docs/LEARNINGS.md`    | Lecciones operacionales (2 entradas)                                   | Cultura de documentar errores      |
| `SECURITY.md`          | Políticas de seguridad                                                 | —                                  |
| `CONTRIBUTING.md`      | Guía de contribución                                                   | —                                  |

### Decisión clave de documentación (LEARNINGS.md):

> "Las reglas y decisiones nuevas deben ir al documento canónico que posee esa preocupación, en vez de crear guía superpuesta en otros lados"

---

## 12. Estado Actual del Proyecto

### Madurez

| Métrica            | Valor                                           |
| ------------------ | ----------------------------------------------- |
| Líneas Go          | ~6,800                                          |
| Líneas TypeScript  | ~270                                            |
| Paquetes con tests | 11                                              |
| CI                 | GitHub Actions (workflows no leídos en detalle) |
| Branch activa      | `feat/aurelia-os`                               |
| Docs canónicos     | 4 documentos principales completos              |

### Fortalezas

1. **Arquitectura explícita** — Decisiones documentadas, boundaries claros
2. **Modular monolith** — Simple de entender, fácil de evolucionar
3. **Declarative agents** — Markdown con frontmatter, zero código para nuevos agentes
4. **Bridge pattern** — Separación limpia Go/TS, protocolo simple (NDJSON)
5. **Multi-provider** — 5 providers LLM soportados
6. **Pure Go SQLite** — Sin CGO, deployment simple
7. **Cultura de calidad** — TDD, benchmarking medido, LEARNINGS.md

### Constraints Actuales

1. **Bridge requiere Node.js** en PATH — no es binario standalone puro
2. **No hay orquestación multi-agente** — single agent por ejecución
3. **Telegram-only** — No hay otros canales de interfaz aún
4. **Memoria vectorial** presente pero no es el approach principal

### Lecciones Aprendidas (documentadas)

1. **Documentación canónica** — Evitar guía superpuesta; cada tipo de decisión va a su documento
2. **Secrets nunca en repo** — Solo example configs; secret expuesto = rotación obligatoria

---

## ⚠️ Limitaciones de este Reporte

Los siguientes elementos **no fueron leídos** por límite de iteraciones de herramientas:

- **`.github/workflows/`** — Workflows CI/CD completos
- **`internal/*/*.go`** — Código fuente Go de cada módulo (solo se leyeron docs)
- **`bridge/index.ts`** — Código fuente del Bridge TypeScript
- **`cmd/aurelia/*.go`** — Entry points y wiring
- **`e2e/`** — Tests end-to-end
- **`SECURITY.md`**, **`CONTRIBUTING.md`** — Contenido no leído

Para un análisis completo a nivel de código, se recomienda una sesión adicional que lea estos archivos.

---

_Reporte compilado por Hermes 🦅 — Del Olimpo a tu servicio._

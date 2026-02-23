# Reporte Completo: MakoClaw

## Índice
1. [¿Qué es MakoClaw?](#qué-es-MakoClaw)
2. [Propósito y Filosofía](#propósito-y-filosofía)
3. [Características Principales](#características-principales)
4. [Arquitectura del Sistema](#arquitectura-del-sistema)
5. [Funcionalidades Detalladas](#funcionalidades-detalladas)
6. [Estructura del Código](#estructura-del-código)
7. [Herramientas Disponibles](#herramientas-disponibles)
8. [Canales de Comunicación](#canales-de-comunicación)
9. [Proveedores de LLM](#proveedores-de-llm)
10. [Sistema de Skills](#sistema-de-skills)
11. [Posibles Mejoras](#posibles-mejoras)
12. [Nuevas Features Sugeridas](#nuevas-features-sugeridas)
13. [Optimizaciones de Código](#optimizaciones-de-código)
14. [Panel Web: Funciones y Estado](#panel-web-funciones-y-estado)
15. [Auditoría de Seguridad](#auditoría-de-seguridad)
16. [Auditoría de Lógica y Fiabilidad](#auditoría-de-lógica-y-fiabilidad)
17. [Auditoría del Frontend](#auditoría-del-frontend)
18. [Auditoría de Configuración y Despliegue](#auditoría-de-configuración-y-despliegue)
19. [Plan de Correcciones Prioritarias](#plan-de-correcciones-prioritarias)

---

## ¿Qué es MakoClaw?

**MakoClaw** es un asistente personal de IA ultraligero escrito en Go, inspirado en [nanobot](https://github.com/HKUDS/nanobot). Es una refactorización completa desde cero donde el propio agente de IA impulsó toda la migración arquitectónica y optimización de código.

### Estadísticas del Proyecto
- **Lenguaje**: Go (56 archivos, ~13,600 líneas de código)
- **Versión**: 0.1.0
- **Licencia**: MIT
- **Memoria**: <10MB RAM
- **Tiempo de arranque**: <1 segundo
- **Hardware mínimo**: $10 (placas Linux de bajo costo)

---

## Propósito y Filosofía

### Objetivo Principal
Proveer un asistente de IA eficiente que pueda ejecutarse en hardware mínimo, haciendo la inteligencia artificial accesible para todos, independientemente de sus recursos computacionales.

### Filosofía de Diseño
1. **Simplicidad sobre complejidad**: Código limpio y mantenible
2. **Rendimiento sobre features**: Priorizar velocidad y eficiencia
3. **Control y privacidad del usuario**: Datos locales, código abierto
4. **Operación transparente**: El usuario siempre sabe qué está haciendo
5. **Desarrollo impulsado por la comunidad**: Código abierto y colaborativo

### Comparativa con Otras Soluciones

| Característica | OpenClaw | NanoBot | **MakoClaw** |
|---------------|----------|---------|--------------|
| **Lenguaje** | TypeScript | Python | **Go** |
| **RAM** | >1GB | >100MB | **<10MB** |
| **Arranque** (0.8GHz) | >500s | >30s | **<1s** |
| **Costo Hardware** | Mac Mini $599 | Linux SBC ~$50 | **Cualquier Linux $10** |

---

## Características Principales

### 🪶 Ultra-Ligero
- **<10MB** de memoria RAM
- **99%** más pequeño que Clawdbot
- Binary único autocontenido

### 💰 Costo Mínimo
- Corre en hardware de **$10**
- **98%** más barato que Mac Mini
- Sin dependencias externas pesadas

### ⚡️ Velocidad
- Arranque en **1 segundo** incluso en CPU de 0.6GHz
- **400x** más rápido que alternativas
- Respuestas instantáneas

### 🌍 Portabilidad Real
- Binary único para RISC-V, ARM y x86
- Una compilación, cualquier plataforma
- Compatibilidad cross-platform

### 🤖 Bootstrapping con IA
- **95%** del core generado por agentes
- Refinamiento human-in-the-loop
- Implementación nativa en Go

---

## Arquitectura del Sistema

### Diagrama de Componentes

```
┌─────────────────────────────────────────────────────────────┐
│                        MakoClaw                              │
├─────────────────────────────────────────────────────────────┤
│  CLI (cmd/MakoClaw/main.go)                                 │
│  ├── onboard                                                │
│  ├── agent (modo interactivo/directo)                      │
│  ├── gateway (servidor multi-canal)                        │
│  ├── cron (tareas programadas)                             │
│  ├── skills (gestión de habilidades)                       │
│  ├── auth (autenticación OAuth)                            │
│  ├── status                                                │
│  └── migrate (migración desde OpenClaw)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Core Packages                          │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Agent      │  │   Config     │  │   Providers  │      │
│  │  (agent/)    │  │  (config/)   │  │ (providers/) │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Tools      │  │   Channels   │  │    Bus       │      │
│  │  (tools/)    │  │ (channels/)  │  │   (bus/)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Skills     │  │   Session    │  │    Cron      │      │
│  │  (skills/)   │  │ (session/)   │  │  (cron/)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Flujo de Datos

```
Usuario → Canal (Telegram/Discord/etc) → MessageBus → Agent Loop → LLM Provider
                                              ↓
                                         Tool Registry
                                              ↓
                              [Web Search] [File Ops] [Shell] [Subagent]
                                              ↓
                                    Respuesta → Canal → Usuario
```

---

## Funcionalidades Detalladas

### 1. Modo Agente (CLI)

#### Modo Directo
```bash
MakoClaw agent -m "¿Qué es 2+2?"
```

#### Modo Interactivo
```bash
MakoClaw agent
# Inicia chat interactivo con readline (historial, edición)
```

#### Características:
- Historial de comandos (100 líneas)
- Soporte para sesiones múltiples (`-s session_name`)
- Modo debug (`--debug`)
- Respuestas formateadas

### 2. Gateway Multi-Canal

Inicia un servidor que escucha múltiples canales simultáneamente:

```bash
MakoClaw gateway
```

Servicios que se inician:
- **Cron Service**: Ejecución de tareas programadas
- **Heartbeat Service**: Monitoreo de estado
- **Channel Manager**: Gestión de canales habilitados
- **Agent Loop**: Procesamiento de mensajes

### 3. Sistema de Autenticación

Soporta OAuth y tokens:

```bash
# Login con OAuth (flujo de navegador)
MakoClaw auth login --provider openai

# Login con device code (headless)
MakoClaw auth login --provider openai --device-code

# Login con token manual
MakoClaw auth login --provider anthropic

# Ver estado
MakoClaw auth status

# Logout
MakoClaw auth logout --provider openai
```

### 4. Tareas Programadas (Cron)

```bash
# Listar trabajos
MakoClaw cron list

# Agregar trabajo recurrente
MakoClaw cron add -n "recordatorio" -m "Revisar emails" -e 3600

# Agregar con expresión cron
MakoClaw cron add -n "daily" -m "Backup" -c "0 9 * * *"

# Eliminar trabajo
MakoClaw cron remove <job_id>

# Habilitar/Deshabilitar
MakoClaw cron enable <job_id>
MakoClaw cron disable <job_id>
```

### 5. Gestión de Skills

```bash
# Listar skills instalados
MakoClaw skills list

# Instalar skill desde GitHub
MakoClaw skills install sipeed/MakoClaw-skills/weather

# Buscar skills disponibles
MakoClaw skills search

# Ver detalles
MakoClaw skills show weather

# Eliminar skill
MakoClaw skills remove weather

# Instalar skills built-in
MakoClaw skills install-builtin
MakoClaw skills list-builtin
```

### 6. Migración desde OpenClaw

```bash
# Migración completa
MakoClaw migrate

# Solo configuración
MakoClaw migrate --config-only

# Solo workspace
MakoClaw migrate --workspace-only

# Simulación (sin cambios)
MakoClaw migrate --dry-run

# Forzar sin confirmación
MakoClaw migrate --force

# Sincronizar nuevamente
MakoClaw migrate --refresh
```

---

## Estructura del Código

### Organización de Paquetes

```
pkg/
├── agent/           # Core del agente y loop principal
│   ├── loop.go      # Lógica principal del agente
│   ├── context.go   # Builder de contexto
│   └── memory.go    # Gestión de memoria
├── auth/            # Autenticación OAuth y tokens
│   ├── oauth.go     # Flujos OAuth
│   ├── pkce.go      # PKCE para OAuth
│   ├── token.go     # Gestión de tokens
│   └── store.go     # Almacenamiento de credenciales
├── bus/             # Message bus interno
│   ├── bus.go       # Implementación del bus
│   └── types.go     # Tipos de mensajes
├── channels/        # Integraciones con mensajería
│   ├── telegram.go  # Bot de Telegram
│   ├── discord.go   # Bot de Discord
│   ├── slack.go     # Integración Slack
│   ├── whatsapp.go  # WhatsApp bridge
│   ├── feishu.go    # Feishu/Lark
│   ├── dingtalk.go  # DingTalk
│   ├── qq.go        # QQ
│   ├── maixcam.go   # MaixCAM
│   ├── manager.go   # Gestor de canales
│   └── base.go      # Interfaces base
├── config/          # Configuración
│   └── config.go    # Estructura y carga de config
├── cron/            # Tareas programadas
│   └── service.go   # Servicio cron
├── heartbeat/       # Monitoreo
│   └── service.go   # Heartbeat
├── logger/          # Logging estructurado
│   └── logger.go    # Logger con campos
├── migrate/         # Migración OpenClaw
│   ├── migrate.go   # Lógica de migración
│   ├── config.go    # Migración de config
│   └── workspace.go # Migración de workspace
├── providers/       # Proveedores de LLM
│   ├── types.go     # Interfaces comunes
│   ├── http_provider.go  # Provider HTTP genérico
│   ├── claude_provider.go   # Anthropic Claude
│   └── codex_provider.go    # OpenAI Codex
├── session/         # Gestión de sesiones
│   └── manager.go   # Manager de sesiones
├── skills/          # Sistema de skills
│   ├── loader.go    # Carga de skills
│   └── installer.go # Instalación de skills
├── tools/           # Herramientas del agente
│   ├── base.go      # Interface de tool
│   ├── registry.go  # Registro de tools
│   ├── filesystem.go # Operaciones de archivo
│   ├── edit.go      # Edición de archivos
│   ├── shell.go     # Ejecución shell
│   ├── web.go       # Búsqueda web
│   ├── message.go   # Envío de mensajes
│   ├── subagent.go  # Subagentes
│   ├── spawn.go     # Spawning de tareas
│   └── cron.go      # Tool de cron
├── utils/           # Utilidades
│   ├── string.go    # Utilidades de strings
│   └── media.go     # Procesamiento de media
└── voice/           # Transcripción de voz
    └── transcriber.go # Transcripción Groq
```

### Estructura del Workspace

```
~/.MakoClaw/
├── config.json          # Configuración principal
├── workspace/
│   ├── sessions/        # Historial de conversaciones
│   ├── memory/          # Memoria a largo plazo
│   │   └── MEMORY.md
│   ├── cron/            # Base de datos de tareas
│   ├── skills/          # Skills personalizados
│   ├── AGENTS.md        # Instrucciones del agente
│   ├── IDENTITY.md      # Identidad del agente
│   ├── SOUL.md          # Alma/personalidad
│   ├── TOOLS.md         # Descripción de tools
│   └── USER.md          # Preferencias del usuario
└── auth.json            # Credenciales OAuth
```

---

## Herramientas Disponibles

### 1. Operaciones de Archivos

#### `read_file`
Lee contenido de archivos con soporte para offsets y límites.

**Parámetros:**
- `file_path`: Ruta del archivo
- `offset`: Línea inicial (opcional)
- `limit`: Número de líneas (opcional)

**Seguridad:** Puede restringirse al workspace.

#### `write_file`
Escribe contenido en archivos.

**Parámetros:**
- `file_path`: Ruta del archivo
- `content`: Contenido a escribir

#### `append_file`
Agrega contenido al final de archivos.

#### `list_dir`
Lista directorios con información detallada.

**Parámetros:**
- `path`: Ruta del directorio
- `recursive`: Listado recursivo

#### `edit_file`
Edición precisa de archivos con búsqueda/reemplazo.

**Características:**
- Búsqueda por string exacto
- Preserva indentación
- Soporte para múltiples reemplazos
- Verificación de cambios

### 2. Ejecución Shell

#### `exec`
Ejecuta comandos shell.

**Parámetros:**
- `command`: Comando a ejecutar
- `timeout`: Timeout en segundos (opcional)

**Seguridad:** Puede restringirse al workspace.

### 3. Web y Búsqueda

#### `web_search`
Búsqueda web usando Brave Search API.

**Parámetros:**
- `query`: Término de búsqueda

**Nota:** Requiere API key de Brave (2000 consultas/mes gratis).

#### `web_fetch`
Obtiene contenido de URLs.

**Parámetros:**
- `url`: URL a obtener
- `format`: Formato de salida (markdown, text, html)
- `max_length`: Longitud máxima del contenido

### 4. Comunicación

#### `message`
Envía mensajes a través de canales.

**Parámetros:**
- `content`: Contenido del mensaje
- `channel`: Canal destino (opcional)
- `to`: Destinatario (opcional)

### 5. Subagentes

#### `spawn`
Crea subagentes para tareas paralelas.

**Casos de uso:**
- Procesamiento concurrente
- Tareas en segundo plano
- Múltiples contextos

### 6. Tareas Programadas

#### `schedule`
Programa tareas recurrentes.

**Soporta:**
- Intervalos ("every 10 minutes")
- Expresiones cron ("0 9 * * *")
- Recordatorios one-time

---

## Canales de Comunicación

### 1. Telegram (Recomendado)
- **Setup**: Fácil (solo token)
- **Features**: Mensajes de texto, voz (con Groq), imágenes
- **Costo**: Gratis

### 2. Discord
- **Setup**: Fácil (bot token + intents)
- **Features**: Mensajes en canales, DMs, threads
- **Costo**: Gratis

### 3. Slack
- **Setup**: Medio (bot token + app token)
- **Features**: Mensajes, threads, reacciones
- **Costo**: Gratis (con limitaciones)

### 4. QQ
- **Setup**: Fácil (AppID + AppSecret)
- **Features**: Mensajes grupales y privados
- **Costo**: Gratis

### 5. DingTalk
- **Setup**: Medio (credenciales de app)
- **Features**: Mensajes organizacionales
- **Costo**: Gratis

### 6. WhatsApp
- **Setup**: Complejo (requiere bridge)
- **Features**: Mensajes de texto
- **Costo**: Gratis (con bridge local)

### 7. Feishu/Lark
- **Setup**: Medio (app credentials)
- **Features**: Mensajes empresariales
- **Costo**: Gratis

### 8. MaixCAM
- **Setup**: Integración con hardware
- **Features**: Comunicación con dispositivos MaixCAM
- **Costo**: Hardware requerido

---

## Proveedores de LLM

### Soportados Actualmente

| Proveedor | Tipo | Transcripción Voz | Obtener API Key |
|-----------|------|-------------------|-----------------|
| **OpenRouter** | Múltiples modelos | ❌ | [openrouter.ai](https://openrouter.ai/keys) |
| **Zhipu** | GLM-4, etc. | ❌ | [bigmodel.cn](https://bigmodel.cn) |
| **Anthropic** | Claude | ❌ | [console.anthropic.com](https://console.anthropic.com) |
| **OpenAI** | GPT-4, etc. | ❌ | [platform.openai.com](https://platform.openai.com) |
| **Gemini** | Google | ❌ | [aistudio.google.com](https://aistudio.google.com) |
| **DeepSeek** | DeepSeek | ❌ | [platform.deepseek.com](https://platform.deepseek.com) |
| **Groq** | Llama, Mixtral | ✅ Whisper | [console.groq.com](https://console.groq.com) |
| **vLLM** | Local | ❌ | Auto-hospedado |
| **Nvidia** | NVIDIA models | ❌ | [build.nvidia.com](https://build.nvidia.com) |
| **Moonshot** | Kimi | ❌ | [platform.moonshot.cn](https://platform.moonshot.cn) |

### Características de Configuración

```json
{
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "api_base": "https://openrouter.ai/api/v1"
    },
    "groq": {
      "api_key": "gsk_xxx",
      "api_base": ""
    }
  }
}
```

### Autenticación

- **API Key**: Directa en configuración
- **OAuth**: Flujo de navegador para OpenAI
- **Device Code**: Para entornos headless
- **Token Manual**: Para Anthropic

---

## Sistema de Skills

### ¿Qué son los Skills?

Los skills son extensiones de conocimiento que guían al agente para tareas específicas. Son archivos markdown con metadatos YAML.

### Estructura de un Skill

```markdown
---
name: weather
description: Get current weather and forecasts
homepage: https://wttr.in/:help
metadata: {"requires":{"bins":["curl"]}}
---

# Weather Skill

Instrucciones detalladas aquí...
```

### Skills Built-in Disponibles

| Skill | Descripción | Requisitos |
|-------|-------------|------------|
| **weather** | Clima y pronósticos | curl |
| **github** | Interacción con GitHub | gh CLI |
| **tmux** | Gestión de sesiones tmux | tmux |
| **summarize** | Resumen de contenido | - |
| **skill-creator** | Crear nuevos skills | - |

### Instalación de Skills

**Desde GitHub:**
```bash
MakoClaw skills install usuario/repo/skill-name
```

**Instalación local:**
- Copiar a `~/.MakoClaw/workspace/skills/`

**Estructura:**
```
skills/
└── skill-name/
    └── SKILL.md
```

### Uso en el Agente

El agente automáticamente:
1. Carga todos los skills disponibles
2. Los incluye en el contexto del sistema
3. Sigue las instrucciones según la solicitud del usuario

---

## Posibles Mejoras

### 1. Performance

#### A. Compresión de Contexto
- **Problema**: Ventanas de contexto grandes consumen tokens
- **Solución**: Implementar compresión inteligente de historial
- **Implementación**: 
  ```go
  // Agregar compresión en context.go
  func (cb *ContextBuilder) CompressHistory(messages []Message) []Message
  ```

#### B. Caché de Respuestas
- **Problema**: Consultas repetidas consumen API calls
- **Solución**: Caché local con hash de consulta
- **Beneficio**: Reducción de costos y latencia

#### C. Lazy Loading de Skills
- **Problema**: Todos los skills se cargan al inicio
- **Solución**: Cargar solo cuando se detecte intención relacionada
- **Implementación**: Sistema de intenciones/keywords

### 2. Seguridad

#### A. Sandboxing de Shell
- **Problema**: Comandos shell tienen acceso completo
- **Solución**: 
  - Whitelist de comandos permitidos
  - Ejecución en contenedores (docker/podman)
  - Chroot para operaciones de filesystem

#### B. Rate Limiting
- **Problema**: Sin límites en consumo de API
- **Solución**:
  ```go
  type RateLimiter struct {
      requests map[string][]time.Time
      limits   map[string]int // por minuto
  }
  ```

#### C. Sanitización de Inputs
- **Mejorar validación de todos los inputs del usuario
- Prevenir prompt injection attacks
- Validar paths de archivo para directory traversal

### 3. UX/UI

#### A. Web Dashboard
- **Feature**: Panel web para configuración y monitoreo
- **Tecnología**: Go templates + HTMX o React
- **Funciones**:
  - Ver y editar configuración
  - Monitorear sesiones en tiempo real
  - Visualizar logs
  - Gestionar skills

#### B. Mejor CLI Experience
- **Spinner**: Mostrar progreso durante operaciones largas
- **Colores**: Mejorar output con colores y formatting
- **Autocompletion**: Completions para bash/zsh/fish
- **Sugerencias**: "Did you mean?" para comandos incorrectos

#### C. Notificaciones Nativas
- Soporte para notificaciones del sistema operativo
- Integración con `notify-send` (Linux), `osascript` (macOS), `toast` (Windows)

### 4. Testing

#### A. Test Coverage
- **Actual**: Mínimos tests existentes
- **Meta**: >80% coverage
- **Prioridad**:
  1. `pkg/tools/` - Herramientas críticas
  2. `pkg/providers/` - LLM providers
  3. `pkg/agent/` - Core del agente
  4. `pkg/channels/` - Integraciones

#### B. Tests de Integración
- Tests end-to-end para cada canal
- Mock servers para providers de LLM
- Tests de migración

#### C. Benchmarks
- Benchmarks de performance para:
  - Inicio del agente
  - Procesamiento de mensajes
  - Ejecución de tools
  - Uso de memoria

### 5. Documentación

#### A. Documentación de API
- Documentar todas las interfaces internas
- Generar docs con `godoc`
- Ejemplos de uso para cada paquete

#### B. Guías de Desarrollo
- Cómo crear un nuevo provider
- Cómo crear un nuevo canal
- Cómo crear un nuevo tool
- Cómo crear un skill

#### C. Documentación de Arquitectura
- Diagramas de flujo detallados
- Decisiones de diseño documentadas (ADRs)
- Guía de contribución

### 6. DevOps

#### A. CI/CD Pipeline
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: make test
      - run: make lint
      - run: make build-all
```

#### B. Releases Automatizados
- Versionado semántico automático
- Changelog generado automáticamente
- Binarios pre-compilados para todas las plataformas
- Docker images multi-arch

#### C. Monitoreo
- Métricas de uso (Prometheus)
- Health checks
- Alertas para errores críticos

### 7. Extensiones Core

#### A. Multi-Agent Sistema
- **Feature**: Múltiples agentes especializados
- **Ejemplo**: 
  - Agent de código
  - Agent de investigación  
  - Agent de comunicación
- **Implementación**: Orquestador que delega según intención

#### B. Memoria Vectorial
- **Problema**: Búsqueda en memoria es lineal
- **Solución**: Embeddings + vector DB (sqlite-vec, qdrant)
- **Beneficio**: Recuperación semántica de información

#### C. Plugin System
- **Feature**: Plugins compilados o WASM
- **Ventaja**: Extensiones sin modificar core
- **Seguridad**: WASM sandboxed

---

## Nuevas Features Sugeridas

### 1. Integraciones de Terceros

#### A. Control de Versiones
- **Git Tool**: Operaciones git avanzadas
  ```go
  type GitTool struct {
      workspace string
  }
  // git_log, git_diff, git_blame, git_branch, etc.
  ```

#### B. Gestión de Proyectos
- **Integración con**:
  - GitHub Issues/Projects
  - Jira
  - Trello
  - Linear
  - Notion

#### C. Cloud Providers
- **AWS**: CLI integrado, logs CloudWatch
- **GCP**: Operaciones gcloud
- **Azure**: Comandos az

#### D. Bases de Datos
- **SQL Tool**: Ejecutar queries SQL
  ```go
  type SQLTool struct {
      connections map[string]*sql.DB
  }
  ```
- Soporta: PostgreSQL, MySQL, SQLite

### 2. Capacidades de IA Avanzadas

#### A. Image Understanding
- **Feature**: Análisis de imágenes
- **Implementación**: Integración con GPT-4V, Gemini Pro Vision
- **Uso**: 
  ```
  Usuario: [imagen]
  Agente: Analiza la imagen y describe el contenido
  ```

#### B. Generación de Imágenes
- **Feature**: Crear imágenes desde descripción
- **Integración**: DALL-E, Midjourney API, Stable Diffusion
- **Comando**: `/imagen un paisaje montañoso al atardecer`

#### C. Text-to-Speech
- **Feature**: Respuestas habladas
- **Implementación**: Integración con ElevenLabs, Coqui TTS
- **Configuración**: 
  ```json
  "voice": {
    "enabled": true,
    "provider": "elevenlabs",
    "voice_id": "xxx"
  }
  ```

#### D. Code Execution Seguro
- **Feature**: Ejecutar código en sandbox
- **Implementación**: Firecracker microVMs o gVisor
- **Soporta**: Python, JavaScript, Go, Rust

### 3. Automatización Avanzada

#### A. Workflow Engine
- **Feature**: Flujos de trabajo definidos por el usuario
- **Formato**: YAML o JSON
- **Ejemplo**:
  ```yaml
  workflows:
    daily_report:
      trigger: cron("0 9 * * *")
      steps:
        - web_search: "noticias tech"
        - summarize: "Crear resumen"
        - send_email: "destinatario@email.com"
  ```

#### B. Conditional Logic
- **Feature**: Respuestas condicionales basadas en contexto
- **Ejemplo**:
  ```go
  if session.TimeSinceLastMessage() > 24*time.Hour {
      response += "¡Hola de nuevo! Han pasado 24h desde la última vez."
  }
  ```

#### C. Event Triggers
- **Triggers**:
  - Cambios en archivos (fsnotify)
  - Webhooks HTTP
  - Eventos de calendario
  - Notificaciones del sistema

### 4. Mejoras en Conversación

#### A. Contexto Multi-Sesión
- **Feature**: Compartir contexto entre sesiones
- **Implementación**: Memoria global del usuario
- **Uso**: Recordar preferencias entre diferentes chats

#### B. Personalidad Configurable
- **Archivo**: `PERSONALITY.md` en workspace
- **Configuración**:
  ```markdown
  ## Personalidad
  - Estilo: Formal/Casual/Profesional
  - Tono: Amigable/Directo/Sarcástico
  - Largo de respuesta: Conciso/Detallado
  ```

#### C. Proactive Suggestions
- **Feature**: Sugerencias proactivas basadas en contexto
- **Ejemplo**: Detectar que usuario está trabajando en proyecto X y sugerir comandos útiles

### 5. Herramientas de Productividad

#### A. Note Taking
- **Feature**: Tomar notas rápidas
- **Comandos**:
  ```
  /note Reunión con Juan sobre proyecto X
  /notes list
  /notes search "proyecto X"
  ```

#### B. Task Management
- **Integración**: Con todo.txt o tasks.json
- **Features**:
  - Crear tareas
  - Establecer prioridades
  - Fechas límite
  - Proyectos/contextos

#### C. Time Tracking
- **Feature**: Seguimiento de tiempo
- **Comandos**:
  ```
  /timer start "Trabajando en feature Y"
  /timer stop
  /timer report --week
  ```

### 6. Capacidades Colaborativas

#### A. Shared Workspaces
- **Feature**: Espacios de trabajo compartidos entre usuarios
- **Uso**: Equipos que comparten contexto y memoria

#### B. Threaded Conversations
- **Mejora**: Soporte completo para threads en Discord/Slack
- **Beneficio**: Mejor organización de conversaciones largas

#### C. Mention System
- **Feature**: Mencionar al agente en canales grupales
- **Implementación**: @MakoClaw comando aquí

### 7. Capacidades Offline

#### A. Local LLM Support
- **Feature**: Soporte mejorado para LLMs locales
- **Opciones**:
  - Ollama integration
  - llama.cpp
  - LocalAI
  - text-generation-webui

#### B. Offline Mode
- **Feature**: Funcionar sin conexión para tareas básicas
- **Capacidades**:
  - Historial local
  - Búsqueda en archivos locales
  - Ejecución de comandos
  - Skills que no requieren internet

#### C. Sync When Online
- **Feature**: Sincronizar cuando hay conexión
- **Implementación**: Cola de operaciones pendientes

### 8. Mejoras en Búsqueda

#### A. Multi-Search Provider
- **Soportar**:
  - Brave (actual)
  - SearXNG (self-hosted)
  - DuckDuckGo
  - Google Custom Search
  - Bing Search API

#### B. Search Aggregation
- **Feature**: Agregar resultados de múltiples fuentes
- **Ranking**: Score combinado de múltiples motores

#### C. Search History
- **Feature**: Historial de búsquedas con respuestas cacheadas
- **Beneficio**: Respuestas instantáneas para consultas repetidas

### 9. Internacionalización

#### A. Multi-language Support
- **Feature**: Soporte para múltiples idiomas
- **Implementación**:
  - i18n para mensajes del sistema
  - Detección automática de idioma
  - Skills traducidos

#### B. RTL Support
- **Feature**: Soporte para idiomas RTL (árabe, hebreo)
- **Implementación**: CSS/logica de rendering RTL

### 10. Mobile Experience

#### A. Mobile App
- **Feature**: App móvil nativa o PWA
- **Plataformas**: iOS, Android
- **Features**:
  - Push notifications
  - Widgets
  - Quick actions

#### B. SMS Integration
- **Feature**: Interactuar vía SMS
- **Implementación**: Twilio o similar

---

## Optimizaciones de Código

### 1. Refactoring Sugerido

#### A. Separar Responsabilidades
**Problema actual**: `pkg/agent/loop.go` es muy grande (636 líneas)

**Solución**:
```
pkg/agent/
├── loop.go              # Solo orquestación
├── iteration.go         # Lógica de iteración LLM
├── summarization.go     # Lógica de resumen
├── context_builder.go   # Construcción de contexto
└── state.go            # Gestión de estado
```

#### B. Interfaces Más Pequeñas
**Principio**: Interface Segregation

**Ejemplo**:
```go
// En lugar de una interfaz grande
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
    Validate(args map[string]interface{}) error
    Cleanup() error
    // ... más métodos
}

// Separar en interfaces especializadas
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

type ValidatableTool interface {
    Tool
    Validate(args map[string]interface{}) error
}

type CleanableTool interface {
    Tool
    Cleanup() error
}
```

### 2. Mejoras de Performance

#### A. Pool de Conexiones HTTP
**Actual**: Nueva conexión por cada request

**Mejora**:
```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

#### B. Concurrent Tool Execution
**Actual**: Tools se ejecutan secuencialmente

**Mejora**:
```go
func (r *ToolRegistry) ExecuteParallel(ctx context.Context, calls []ToolCall) []Result {
    var wg sync.WaitGroup
    results := make([]Result, len(calls))
    
    for i, call := range calls {
        wg.Add(1)
        go func(idx int, tc ToolCall) {
            defer wg.Done()
            result, err := r.Execute(ctx, tc.Name, tc.Arguments)
            results[idx] = Result{Result: result, Error: err}
        }(i, call)
    }
    
    wg.Wait()
    return results
}
```

#### C. Context Cancellation
**Mejora**: Mejor manejo de cancelación

```go
func (al *AgentLoop) runAgentLoop(ctx context.Context, opts processOptions) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    // Ahora todas las operaciones respetan el timeout
}
```

### 3. Manejo de Errores

#### A. Errores Tipados
**Implementación**:
```go
var (
    ErrToolNotFound = errors.New("tool not found")
    ErrInvalidArgs  = errors.New("invalid arguments")
    ErrTimeout      = errors.New("operation timed out")
    ErrProviderDown = errors.New("LLM provider unavailable")
)

type ToolError struct {
    Tool    string
    Wrapped error
}

func (e *ToolError) Error() string {
    return fmt.Sprintf("tool %s failed: %v", e.Tool, e.Wrapped)
}

func (e *ToolError) Unwrap() error {
    return e.Wrapped
}
```

#### B. Retry Logic
**Implementación**:
```go
func withRetry(ctx context.Context, maxRetries int, fn func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        if err = fn(); err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        backoff := time.Duration(i*i) * time.Second
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
            continue
        }
    }
    return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```

### 4. Logging Mejorado

#### A. Structured Logging Consistente
**Actual**: Mezcla de log styles

**Mejora**:
```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

type Field struct {
    Key   string
    Value interface{}
}

// Uso
logger.Info("message processed",
    Field{"channel", msg.Channel},
    Field{"duration_ms", duration.Milliseconds()},
    Field{"session_id", sessionID},
)
```

#### B. Log Levels Configurables por Componente
```json
{
  "logging": {
    "default": "info",
    "components": {
      "agent": "debug",
      "tools": "warn",
      "channels": "info"
    }
  }
}
```

### 5. Configuración

#### A. Validación de Config
**Implementación**:
```go
func (c *Config) Validate() error {
    var errs []error
    
    if c.Agents.Defaults.Model == "" {
        errs = append(errs, errors.New("model is required"))
    }
    
    if c.Providers.GetAPIKey() == "" {
        errs = append(errs, errors.New("at least one provider API key is required"))
    }
    
    if len(errs) > 0 {
        return &ValidationError{Errors: errs}
    }
    return nil
}
```

#### B. Hot Reload
**Feature**: Recargar config sin reiniciar

```go
func (c *Config) Watch(path string) {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(path)
    
    go func() {
        for event := range watcher.Events {
            if event.Op&fsnotify.Write == fsnotify.Write {
                c.Reload(path)
            }
        }
    }()
}
```

### 6. Testing

#### A. Table-Driven Tests
**Ejemplo**:
```go
func TestToolRegistry_Execute(t *testing.T) {
    tests := []struct {
        name        string
        toolName    string
        args        map[string]interface{}
        wantResult  string
        wantErr     bool
        errContains string
    }{
        {
            name:       "read existing file",
            toolName:   "read_file",
            args:       map[string]interface{}{"file_path": "/tmp/test.txt"},
            wantResult: "content",
            wantErr:    false,
        },
        {
            name:        "tool not found",
            toolName:    "nonexistent",
            args:        map[string]interface{}{},
            wantErr:     true,
            errContains: "not found",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

#### B. Mocks Automatizados
**Uso de mockgen**:
```go
//go:generate mockgen -source=pkg/providers/types.go -destination=pkg/providers/mock/provider_mock.go

type MockLLMProvider struct {
    ctrl     *gomock.Controller
    recorder *MockLLMProviderMockRecorder
}
```

#### C. Tests de Integración con Containers
```go
func TestPostgreSQLTool(t *testing.T) {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "test",
        },
    }
    
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    // ... run tests
}
```

### 7. Documentación de Código

#### A. GoDoc Completo
**Estándar**:
```go
// Package tools provides a registry and execution framework for AI agent tools.
//
// Tools are the primary way the agent interacts with external systems like
// file systems, web services, and command execution.
//
// Basic usage:
//
//     registry := tools.NewToolRegistry()
//     registry.Register(tools.NewReadFileTool(workspace))
//     result, err := registry.Execute(ctx, "read_file", args)
//
package tools

// Tool represents an executable capability of the agent.
// Each tool has a name, description, parameters, and execution logic.
type Tool interface {
    // Name returns the unique identifier for this tool.
    // Must be unique within a registry.
    Name() string
    
    // Description returns a human-readable description of what the tool does.
    // This is used by the LLM to understand when to use the tool.
    Description() string
    
    // Execute runs the tool with the provided arguments.
    // ctx can be used for cancellation and timeouts.
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

#### B. Ejemplos Ejecutables
```go
// ExampleToolRegistry_Execute muestra cómo ejecutar una herramienta.
func ExampleToolRegistry_Execute() {
    registry := NewToolRegistry()
    registry.Register(NewReadFileTool("/tmp", true))
    
    result, err := registry.Execute(context.Background(), "read_file", map[string]interface{}{
        "file_path": "/tmp/example.txt",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result)
    // Output: contenido del archivo
}
```

### 8. Seguridad

#### A. Content Security Policy
**Para web dashboard**:
```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        next.ServeHTTP(w, r)
    })
}
```

#### B. Secrets Management
**Mejora**: No almacenar secrets en texto plano
```go
type SecureConfig struct {
    APIKey secret.String `json:"-"` // No serializar
}

type secretString struct {
    value string
}

func (s *secretString) UnmarshalJSON(data []byte) error {
    // Desencriptar si está encriptado
    // O usar keyring del OS
}
```

---

## Conclusión

MakoClaw es una implementación impresionante de un asistente de IA ultraligero que demuestra que es posible tener funcionalidades avanzadas con un footprint mínimo. El código está bien estructurado y sigue buenas prácticas de Go.

Las principales fortalezas son:
1. **Eficiencia**: <10MB RAM, <1s arranque
2. **Arquitectura limpia**: Paquetes bien separados
3. **Extensibilidad**: Sistema de skills flexible
4. **Multi-plataforma**: Soporte para RISC-V, ARM, x86

Las áreas de mejora identificadas incluyen:
1. Mayor cobertura de tests
2. Mejoras de seguridad (sandboxing)
3. Dashboard web para administración
4. Sistema de plugins más robusto
5. Mejor manejo de errores y retries

El proyecto tiene un potencial enorme para crecer mientras mantiene su filosofía de simplicidad y eficiencia.

---

**Reporte generado el**: 12 de Febrero de 2026  
**Versión analizada**: MakoClaw v0.1.0  
**Líneas de código**: ~13,600  
**Archivos Go**: 56

---

## Panel Web: Funciones y Estado

### Funciones Implementadas

El panel web (`pkg/web/`) es una SPA embebida que permite operar MakoClaw desde el navegador.

| Función | Descripción | Estado |
|---------|-------------|--------|
| **Login JWT** | Autenticación con usuario/contraseña, token HMAC-SHA256 | ✅ Funcional |
| **Logout** | Limpieza de sesión y token | ✅ Funcional |
| **Cambio de contraseña** | Desde la UI con validación mínima (10 chars) | ✅ Funcional |
| **Chat en tiempo real** | WebSocket bidireccional con el agente IA | ✅ Funcional |
| **Tablero Kanban** | Visualización de tareas por columnas (backlog→done) | ✅ Funcional |
| **CRUD de tareas** | Crear, editar título, cambiar estado, eliminar | ✅ Funcional |
| **Vista de detalle** | Panel lateral con meta, resultado del bot y logs | ✅ Funcional |
| **Logs de tarea** | Historial de eventos por tarea | ✅ Funcional |
| **Acciones rápidas chat** | `/task list`, `/task run`, `/task move` | ✅ Funcional |
| **Temporizador de sesión** | Cuenta regresiva visible + warning <2min | ✅ Funcional |
| **Auto-logout** | Expiración de JWT con redirección a login | ✅ Funcional |
| **Filtros avanzados** | Por texto, estado, fecha | ✅ Funcional |
| **Ordenamiento** | Por fecha o título (asc/desc) | ✅ Funcional |
| **Badges de estado** | Colores diferenciados por estado de tarea | ✅ Funcional |
| **Selección de tarea** | Click para ver detalle con highlight visual | ✅ Funcional |
| **Protección XSS** | Función `esc()` para sanitizar renders | ⚠️ Parcial |
| **Task Worker** | Procesamiento automático de tareas todo→review | ✅ Funcional |
| **WebSocket Tasks** | Actualización en tiempo real del tablero | ✅ Funcional |
| **Password auto-gen** | Si no hay password configurado, genera uno aleatorio | ✅ Funcional |

### Flujo de Uso Típico

1. Arrancar con `MakoClaw web` o `MakoClaw gateway` (con web habilitado en config)
2. Abrir `http://127.0.0.1:18880` en navegador
3. Login con usuario/contraseña configurados (o el password auto-generado)
4. Crear tareas desde el formulario o chat (`/task create mi tarea`)
5. Las tareas en "todo" se procesan automáticamente por el agent loop
6. Ver resultado en el panel de detalle, mover estados manualmente si se desea

---

## Auditoría de Seguridad

> Auditoría realizada sobre `pkg/web/auth.go`, `pkg/web/server.go`, `pkg/web/tasks_store.go`

### ✅ Buenas Prácticas Detectadas

| Práctica | Ubicación | Detalle |
|----------|-----------|---------|
| bcrypt con DefaultCost | `auth.go:63,187` | Hashing seguro de contraseñas |
| Comparación constant-time | `auth.go:114` | Protección contra timing attacks |
| JWT secret aleatorio 32 bytes | `auth.go:67-74` | Entropía adecuada |
| HS256 hardcodeado | `auth.go:124` | Sin confusión de algoritmos |
| Queries parametrizadas | `tasks_store.go` (todas) | Sin inyección SQL |
| Rate limiting en login | `server.go:500-505` | 5 intentos/min |
| Web bind a 127.0.0.1 por defecto | `config.go:277` | Solo acceso local |

### 🔴 Problemas Críticos

#### 1. Token JWT en URL del WebSocket
- **Archivo**: `server.go:197`
- **Riesgo**: El token se pasa como query parameter (`?token=...`)
- **Impacto**: Se filtra en historial del navegador, logs del servidor, headers Referer
- **Fix**: Usar subprotocolo WebSocket o cookie httpOnly

#### 2. Sin Protección CSRF
- **Archivo**: `server.go` (global)
- **Riesgo**: Endpoints POST/PATCH/DELETE sin token CSRF
- **Impacto**: Ataques cross-site pueden ejecutar acciones autenticadas
- **Fix**: Añadir token CSRF en formularios o header `X-CSRF-Token`

#### 3. Sin Revocación de Tokens
- **Archivo**: `auth.go` (global)
- **Riesgo**: No hay blacklist de tokens; cambiar contraseña no invalida tokens existentes
- **Impacto**: Tokens comprometidos permanecen válidos hasta expiración
- **Fix**: Implementar token blacklist o rotar JWT secret al cambiar password

### 🟡 Problemas Moderados

#### 4. Headers de Seguridad HTTP Ausentes
- **Archivo**: `server.go:160-181`
- **Faltan**: `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, `Content-Security-Policy`
- **Fix**: Añadir middleware de security headers

#### 5. Rate Limiting Incompleto
- **Archivo**: `server.go`
- **Problema**: Solo `/api/v1/auth/login` tiene rate limit; falta en `/api/v1/auth/change-password`, `/api/v1/tasks`, `/ws/chat`
- **Fix**: Aplicar rate limiting global por IP

#### 6. IP Spoofing vía X-Forwarded-For
- **Archivo**: `server.go:561-572`
- **Problema**: Se confía en `X-Forwarded-For` sin validar si hay reverse proxy
- **Fix**: Solo usar header si se configura explícitamente un trusted proxy

#### 7. Validación de Origen Incompleta (WebSocket)
- **Archivo**: `server.go:206-216`
- **Problema**: Solo compara hostname, no esquema (`http` vs `https`)
- **Fix**: Validar origen completo incluyendo esquema

#### 8. Password Mínimo No Aplicado en Login Inicial
- **Archivo**: `auth.go:113-121`
- **Problema**: El password configurado en JSON no tiene validación de longitud mínima
- **Fix**: Validar longitud mínima en `LoadConfig()`

### 🟢 Sin Problemas

- **Inyección SQL**: Todas las queries usan placeholders `?` → Seguro
- **Secretos hardcodeados**: Solo placeholders en archivos de ejemplo → OK
- **Test credentials**: Solo en archivos `*_test.go` → Esperado

---

## Auditoría de Lógica y Fiabilidad

> Auditoría de lógica de negocio, concurrencia, manejo de errores y recursos

### 🔴 Problemas Críticos

#### 1. Colisión de IDs en Tareas
- **Archivos**: `tasks_store.go:220-221`, `tasks.go:96`
- **Problema**: `generateID()` usa `time.Now().UTC().Format(...)` — no es único bajo concurrencia
- **Comparación**: `pkg/cron/service.go:450-457` usa `crypto/rand` (correcto)
- **Impacto**: Tareas duplicadas, violación de PRIMARY KEY
- **Fix recomendado**:
```go
func generateID() string {
    b := make([]byte, 16)
    crypto_rand.Read(b)
    return hex.EncodeToString(b)
}
```

#### 2. Task Worker Se Bloquea en Errores
- **Archivo**: `server.go:626-659` (`processNextTodoTask`)
- **Problema**: Usa `return` en lugar de `continue` al fallar; una tarea con error bloquea todas las demás
- **Impacto**: Tareas quedan permanentemente en "in_progress" sin reintento
- **Fix**: Cambiar `return` a `continue` y añadir logging del error

#### 3. Race Condition en WebSocket Broadcast
- **Archivo**: `server.go:587-611` (`broadcastTaskEvent`)
- **Problema**: `conn.WriteMessage()` se ejecuta fuera del lock; gorilla/websocket requiere escrituras serializadas
- **Impacto**: Corrupción de protocolo WebSocket, mensajes perdidos, desconexiones
- **Fix**: Añadir mutex por conexión o usar canal de escritura

### 🟡 Problemas Moderados

#### 4. Errores Silenciados en Task Worker
- **Archivo**: `server.go:644-645`
- **Código**: `_ = s.tasks.update(...)` y `_ = s.tasks.addLog(...)`
- **Impacto**: Fallos de escritura no registrados; debugging imposible
- **Fix**: Loguear errores con `logger.WarnC`

#### 5. Nil Dereference en Chat Commands
- **Archivo**: `server.go:411` (`handleTaskChatCommand`)
- **Problema**: No verifica `s.tasks == nil` (sí se hace en `handleTasks` línea 226)
- **Fix**: Añadir guard `if s.tasks == nil { return ... }`

#### 6. Defer Antes de Error Check (WebSocket)
- **Archivo**: `server.go:377-378`
- **Problema**: `defer conn.Close()` puede ejecutarse con `conn` nil si `Upgrade` falla
- **Fix**: Mover `defer` después de la verificación de error

#### 7. Sin Límite de Conexiones WebSocket
- **Archivo**: `server.go:36, 587-611`
- **Problema**: Mapa `tasksClients` crece sin límite; un atacante puede abrir miles de conexiones
- **Fix**: Limitar conexiones máximas (ej: 100) con cleanup

#### 8. Sin Validación de Transiciones de Estado
- **Archivo**: `tasks_store.go:140-160`
- **Problema**: Permite transiciones inválidas (ej: "done" → "backlog")
- **Fix**: Implementar máquina de estados con transiciones permitidas

#### 9. Sin Pool de Conexiones SQLite Configurado
- **Archivo**: `tasks_store.go:31-34`
- **Problema**: No se configura `SetMaxOpenConns`/`SetMaxIdleConns`
- **Fix**: Añadir `db.SetMaxOpenConns(25)` y `db.SetMaxIdleConns(5)`

---

## Auditoría del Frontend

> Auditoría de `pkg/web/static/index.html` (SPA completa)

### 🔴 Problemas Críticos

#### 1. Función `esc()` Incompleta (XSS)
- **Línea**: 485
- **Código actual**: `return String(v || "").replace(/</g, "&lt;")`
- **Problema**: Solo escapa `<`; no escapa `&`, `>`, `"`, `'`
- **Impacto**: Vulnerable a XSS en atributos HTML
- **Fix**:
```javascript
function esc(v) {
    return String(v || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}
```

#### 2. Chat Sin Escapar (XSS vía WebSocket)
- **Líneas**: 307-310
- **Problema**: `appendChat(\`bot: ${payload.content}\`)` — contenido del bot insertado sin sanitizar
- **Impacto**: Si el bot devuelve HTML/JS malicioso, se ejecuta en el navegador
- **Fix**: Usar `esc()` corregido o `textContent`

#### 3. Token en localStorage
- **Líneas**: 174, 200, 245, 253, 293, 320
- **Problema**: JWT almacenado en localStorage es accesible por cualquier script XSS
- **Impacto**: Combinado con las vulnerabilidades XSS anteriores, permite robo de sesión
- **Fix ideal**: Usar cookie httpOnly; alternativa: sanitizar 100% de renders

### 🟡 Problemas Moderados

#### 4. Sin Manejo de Errores en Muchas Llamadas API
- **Líneas afectadas**: 356, 379, 401, 414, 428, 437
- **Problema**: `loadTasks()`, crear/editar/eliminar/obtener logs no manejan errores
- **Fix**: Añadir `catch` con notificación al usuario

#### 5. Sin Exponential Backoff en Reconexión WebSocket
- **Líneas**: 313, 331
- **Problema**: Reintentos fijos a 1500ms pueden saturar el servidor
- **Fix**: Implementar backoff exponencial con jitter

#### 6. Validación de Input Débil
- **Línea 262**: Cambio de password no valida longitud mínima en frontend
- **Línea 378**: Título de tarea solo comprueba no-vacío
- **Línea 411**: `prompt()` devuelve `null` al cancelar, no se maneja correctamente

### 🟢 Buenas Prácticas

- `apiFetch()` centralizado con manejo de 401 → auto-logout
- `textContent` usado para chat del usuario (seguro)
- Reconexión WebSocket solo si hay token activo
- Timer de sesión con cleanup al logout

---

## Auditoría de Configuración y Despliegue

### 🔴 Problemas Críticos

#### 1. Sin Validación de WebConfig
- **Archivo**: `config.go:170-177`
- **Problema**: Port puede ser negativo o >65535; password puede estar vacío; JWTExpiry sin validar formato
- **Fix**: Añadir validación en `LoadConfig()`:
```go
if cfg.Web.Port < 1 || cfg.Web.Port > 65535 { return error }
if cfg.Web.Enabled && cfg.Web.Password == "" { log warning }
```

#### 2. Sin Validación de Variables de Entorno
- **Archivo**: `config.go:309-351`
- **Problema**: Valores de env vars no se validan (puertos, hosts, API keys)
- **Fix**: Validar tras parseo

### 🟡 Problemas Moderados

#### 3. Gateway Binds a 0.0.0.0 por Defecto
- **Archivo**: `config.go:272`
- **Problema**: Expone el gateway a toda la red local
- **Recomendación**: Documentar claramente o cambiar default a 127.0.0.1

#### 4. Makefile Sin Targets de Seguridad
- **Problema**: No hay `make audit`, `make security`, ni `make gosec`
- **Recomendación**: Añadir `go vet ./...` y `govulncheck` al CI

### 🟢 Buenas Prácticas

| Práctica | Estado |
|----------|--------|
| Graceful shutdown con context | ✅ Excelente |
| Signal handling (SIGINT) | ✅ Correcto |
| Multi-service shutdown ordenado | ✅ Correcto |
| Dependencias Go actualizadas | ✅ Sin CVEs conocidos |
| Web bind a 127.0.0.1 por defecto | ✅ Seguro |
| config.example.json sin secretos reales | ✅ Solo placeholders |

---

## Plan de Correcciones Prioritarias

### Prioridad 0 — Críticas (hacer antes de pruebas internas)

| # | Problema | Archivo | Esfuerzo | Impacto |
|---|----------|---------|----------|---------|
| 1 | `generateID()` con colisiones | `tasks_store.go`, `tasks.go` | 10 min | Pérdida de datos |
| 2 | Task worker se bloquea | `server.go:626-659` | 10 min | Tareas atascadas |
| 3 | Race condition WebSocket | `server.go:587-611` | 20 min | Desconexiones |
| 4 | Función `esc()` incompleta | `index.html:485` | 5 min | XSS |
| 5 | Chat XSS vía WebSocket | `index.html:307-310` | 10 min | XSS |

### Prioridad 1 — Importantes (hacer para uso seguro)

| # | Problema | Archivo | Esfuerzo |
|---|----------|---------|----------|
| 6 | Nil check en chat commands | `server.go:411` | 5 min |
| 7 | Defer antes de error check | `server.go:377-378` | 5 min |
| 8 | Security headers HTTP | `server.go` middleware | 15 min |
| 9 | Validación de WebConfig | `config.go` | 20 min |
| 10 | Error handling en frontend | `index.html` múltiples | 20 min |

### Prioridad 2 — Mejoras de Robustez

| # | Problema | Archivo | Esfuerzo |
|---|----------|---------|----------|
| 11 | Rate limiting global | `server.go` | 30 min |
| 12 | Límite de conexiones WS | `server.go` | 15 min |
| 13 | Exponential backoff WS | `index.html` | 15 min |
| 14 | Token revocation | `auth.go` | 1 hora |
| 15 | CSRF protection | `server.go` | 30 min |

### Prioridad 3 — Mejoras Futuras

| # | Problema | Archivo | Esfuerzo |
|---|----------|---------|----------|
| 16 | httpOnly cookies | `auth.go` + `index.html` | 2 horas |
| 17 | Validación transiciones estado | `tasks_store.go` | 30 min |
| 18 | DB connection pool config | `tasks_store.go` | 5 min |
| 19 | Targets de seguridad en Makefile | `Makefile` | 15 min |
| 20 | Accesibilidad (ARIA, labels) | `index.html` | 1 hora |

---

### Resumen Ejecutivo

**Estado general**: La app es **funcional y utilizable para pruebas internas**, pero tiene **5 bugs críticos** que deben corregirse antes de uso con datos reales.

**Fortalezas**:
- Arquitectura limpia y modular
- Autenticación JWT con bcrypt bien implementada
- Queries SQL parametrizadas (sin inyección)
- Graceful shutdown correcto
- SPA funcional con todas las features planeadas

**Debilidades principales**:
- `generateID()` no es collision-safe (usar crypto/rand)
- Task worker se bloquea ante errores (cambiar return→continue)
- Sanitización XSS incompleta en frontend
- Sin CSRF ni security headers
- Sin revocación de tokens

**Recomendación**: Corregir los 5 items P0 (~55 min de trabajo) antes de cualquier prueba interna. Los items P1 (~65 min) son necesarios para uso seguro con datos sensibles.

---

**Auditoría de seguridad actualizada el**: Julio 2025  
**Alcance**: `pkg/web/`, `pkg/tools/tasks.go`, `pkg/config/config.go`, `cmd/MakoClaw/main.go`

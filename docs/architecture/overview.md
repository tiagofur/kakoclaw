# Visión General de la Arquitectura

**MakoClaw** está diseñado con una arquitectura modular y desacoplada que permite la extensibilidad manteniendo un footprint mínimo.

---

## 🎯 Principios de Diseño

### 1. **Simplicidad**
- Código limpio y fácil de entender
- Menor cantidad de abstracciones innecesarias
- Priorizar claridad sobre complejidad

### 2. **Modularidad**
- Componentes independientes y reutilizables
- Interfaces bien definidas
- Bajo acoplamiento entre módulos

### 3. **Eficiencia**
- Mínimo uso de recursos (&lt;10MB RAM)
- Inicio rápido (&lt;1 segundo)
- Operaciones no bloqueantes donde sea posible

### 4. **Extensibilidad**
- Sistema de plugins (skills)
- Tools fácilmente agregables
- Canales configurables
- Protocolo MCP para herramientas externas

---

## 🏗️ Arquitectura de Alto Nivel

```
┌─────────────────────────────────────────────────────────────┐
│                     MakoClaw Application                │
│                     (The Apex AI Agent)                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │     CLI     │    │   Web UI    │    │   Gateway   │     │
│  │   (cmd)     │    │  (server)   │    │  (server)   │     │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘     │
│         │                  │                  │            │
│         └──────────────────┼──────────────────┘            │
│                            │                                │
│                            ▼                                │
│              ┌─────────────────────────┐                   │
│              │      Message Bus        │                   │
│              │    (internal queue)     │                   │
│              └───────────┬─────────────┘                   │
│                          │                                  │
│                          ▼                                  │
│              ┌─────────────────────────┐                   │
│              │   Agent Manager       │                   │
│              │   (orchestrator)      │                   │
│              └───────────┬─────────────┘                   │
│                          │                                  │
│              ┌───────────┴───────────┐                     │
│              │                       │                     │
│              ▼                       ▼                     │
│    ┌─────────────────┐   ┌─────────────────┐              │
│    │   Agent Loop    │   │  Specialists   │              │
│    │ (message proc)  │   │  (multi-agent) │              │
│    └───────┬────────┘   └─────────────────┘              │
│            │                                              │
│            ▼                                              │
│    ┌─────────────────────────┐                             │
│    │   Tool Registry       │                             │
│    │  [Filesystem]   │                             │
│    │  [Web Search]   │                             │
│    │  [Shell Exec]   │                             │
│    │  [Task Manager] │                             │
│    │  [Knowledge]    │                             │
│    │  [Email]       │                             │
│    │  [Spawn]        │                             │
│    │  [MCP Tools]    │                             │
│    └─────────────────────────┘                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Componentes Principales

### 1. **CLI (cmd/MakoClaw)**

Interfaz de línea de comandos que coordina todos los comandos disponibles.

**Responsabilidades:**
- Parseo de argumentos
- Inicialización de componentes
- Manejo de errores de usuario
- Formateo de salida

**Comandos principales:**
- `onboard`: Inicialización
- `agent`: Interacción directa
- `gateway`: Servidor multi-canal
- `web`: Servidor web UI
- `cron`: Gestión de tareas programadas
- `skills`: Gestión de skills
- `auth`: Autenticación

### 2. **Message Bus (pkg/bus)**

Sistema de mensajería interno desacoplado.

**Características:**
- Cola thread-safe
- Buffering configurable
- Context cancellation
- Soporte para múltiples consumers
- WebSocket para actualizaciones en tiempo real

**Flujo de mensajes:**
1. Canales publican mensajes entrantes
2. Agent Manager consume y procesa
3. Respuestas se publican como salientes
4. Canales envían al usuario

### 3. **Agent Manager & Loop (pkg/agent)**

Núcleo del procesamiento de mensajes con soporte multi-agente.

**Componentes:**

#### **Agent Loop**
- Construcción de contexto
- Iteración con LLM
- Ejecución de tools
- Gestión de sesiones
- Resumen automático de historial

#### **Orchestrator**
- Delegación automática a specialists
- Evaluación de capacidades de cada specialist
- Retries automáticos
- Métricas de performance

#### **Specialists**
- Agentes especializados por dominio
- Configuración independiente de LLM
- Modelos diferentes por specialist
- Monitoreo individual

**Algoritmo de procesamiento:**
```
1. Recibir mensaje del bus
2. Si Orchestrator está habilitado:
   a. Analizar tarea
   b. Seleccionar specialist apropiado
   c. Delegar al specialist
3. Si no, procesar con agente por defecto
4. Construir contexto (historial + skills + system)
5. Llamar a LLM con tools disponibles
6. Mientras LLM solicite tools:
   a. Ejecutar tool
   b. Enviar resultado a LLM
7. Retornar respuesta final
8. Guardar en sesión
9. Trigger resumen si es necesario
```

### 4. **Tool Registry (pkg/tools)**

Registro y ejecución de herramientas.

**Diseño:**
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

**Tools disponibles:**

#### **Gestión de Archivos**
- `read_file`: Lectura de archivos
- `write_file`: Escritura de archivos
- `list_dir`: Listado de directorios
- `edit_file`: Edición asistida por LLM

#### **Web & Búsqueda**
- `web_search`: Búsqueda web (Brave Search)
- `web_fetch`: Obtención de URLs

#### **Ejecución**
- `exec`: Ejecución shell con validaciones de seguridad

#### **Agentes**
- `spawn`: Creación de subagentes
- `message`: Envío de mensajes entre canales

#### **Gestión**
- `task_manager`: CRUD de tareas (Kanban)
- `schedule`: Programación de tareas

#### **Conocimiento**
- `query_knowledge`: Búsqueda en base de documentos (RAG)

#### **Comunicación**
- `email`: Envío de emails

#### **Memoria**
- `memory`: Gestión de memoria a largo plazo

### 5. **MCP (Model Context Protocol) (pkg/mcp)**

Soporte para herramientas externas via protocolo MCP.

**Características:**
- Servidores MCP configurables
- Tools externos dinámicos
- Integración con herramientas de terceros
- Registro automático de tools MCP

### 6. **Providers (pkg/providers)**

Abstracción de proveedores de LLM.

**Implementaciones:**
- HTTP Provider (genérico)
- Claude Provider (Anthropic)
- Codex Provider (OpenAI)
- Ollama Provider (local)

**Características:**
- Interfaz común
- Manejo de errores consistente
- Soporte para tool calling
- Streaming responses

### 7. **Channels (pkg/channels)**

Integraciones con plataformas de mensajería.

**Canales soportados:**
- **Telegram**: Bot de mensajería
- **Discord**: Integración con servidores
- **Slack**: Canales de trabajo
- **WhatsApp**: Comunicación empresarial (bridge)
- **Signal**: Mensajería segura
- **QQ**: Comunicación en China
- **DingTalk**: Colaboración empresarial
- **Feishu**: Plataforma de productividad
- **MaixCam**: Hardware AI camera

**Patrón:**
- Cada canal implementa la interfaz `Channel`
- Manager coordina múltiples canales
- Soporte para transcripción de voz (Groq Whisper)
- User resolution para multi-usuario

### 8. **Session Manager (pkg/session)**

Gestión de historial de conversaciones.

**Características:**
- Persistencia en SQLite
- Resumen automático
- Gestión de ventana de contexto
- Múltiples sesiones concurrentes
- Archivo de sesiones

### 9. **Skills System (pkg/skills)**

Sistema de extensión basado en markdown.

**Estructura:**
```
skills/
└── skill-name/
    └── SKILL.md
```

**Formato:**
```yaml
---
name: skill-name
description: What this skill does
metadata: {"requires": {"bins": ["curl"]}}
---

# Skill Documentation
Instructions for agent...
```

### 10. **Workflow Engine (pkg/workflow)**

Motor de ejecución de workflows visuales.

**Características:**
- Ejecución secuencial de pasos
- Soporte para prompts y tools
- Variables dinámicas (`{{step.1.output}}`)
- Manejo de errores
- Logging de ejecución
- Historial de ejecuciones

### 11. **Task Management (pkg/storage)**

Gestión de tareas con soporte Kanban.

**Características:**
- 5 columnas: backlog, todo, in_progress, review, done
- CRUD de tareas
- Filtrado y búsqueda
- Archivado de tareas
- Multi-user support

### 12. **Knowledge Base (pkg/storage)**

Base de conocimientos con RAG (Retrieval-Augmented Generation).

**Características:**
- Upload de documentos múltiples formatos
- Chunking inteligente
- Full-text search (SQLite FTS5)
- Búsqueda semántica
- Gestión de documentos

### 13. **Cron Service (pkg/cron)**

Servicio de tareas programadas.

**Características:**
- Expresiones cron estándar
- Ejecución manual
- Timezone support
- Historial de ejecuciones
- Estado de jobs

### 14. **Authentication (pkg/auth)**

Sistema de autenticación multi-usuario.

**Características:**
- OAuth 2.0 con PKCE
- Almacenamiento seguro de tokens
- Refresh automático
- Multi-usuario
- Session management

---

## 🔄 Flujo de Datos Detallado

### Flujo de un Mensaje (Multi-Agent)

```
Usuario envía mensaje
        │
        ▼
┌───────────────┐
│    Canal      │ (Telegram/Discord/Web/etc)
│   (Webhook/   │
│    Polling)   │
└───────┬───────┘
        │
        │ Convierte a InboundMessage
        │ (resuelve user ID)
        ▼
┌───────────────┐
│   MessageBus  │
│   (publish)   │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Agent       │
│   Manager    │
└───────┬───────┘
        │
        │ Orchestrator habilitado?
        │ Yes              No
        ▼                  │
┌───────────────┐           │
│ Orchestrator  │           │
│ (delega a)   │           │
│ specialist    │           │
└───────┬───────┘           │
        │                   │
        └─────┬─────────────┘
              │
              ▼
    ┌─────────────────┐
    │   Agent Loop   │
    │  (consume)     │
    └───────┬────────┘
            │
            │ 1. Construye contexto
            │    - System prompt
            │    - Skills
            │    - Historial
            │    - Mensaje actual
            │
            │ 2. Llama a LLM
            │
            ▼
    ┌───────────────┐
    │  LLM Provider │
    │  (with tools) │
    └───────┬───────┘
            │
            │ Tool calls?
            ▼
    ┌───────────────┐     No     ┌───────────────┐
    │ Tool Registry │──────────▶ │  Respuesta    │
    │  (execute)    │            │   final       │
    └───────┬───────┘            └───────────────┘
            │ Yes
            ▼
    ┌───────────────┐
    │  Ejecutar     │
    │    Tools      │
    └───────────────┘
            │
            │ (loop hasta que LLM no pida más tools)
            └────────────────────┐
                                 │
            ◀────────────────────┘
            │
            ▼
    ┌───────────────┐
    │  MessageBus   │
    │   (publish    │
    │   outbound)   │
    └───────┬───────┘
            │
            ▼
    ┌───────────────┐
    │     Canal     │
    │   (envía al   │
    │    usuario)   │
    └───────────────┘
```

---

## 🎨 Patrones de Diseño

### 1. **Registry Pattern**

Usado en `ToolRegistry` para registrar y ejecutar tools dinámicamente.

```go
type ToolRegistry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}

func (r *ToolRegistry) Register(tool Tool)
func (r *ToolRegistry) Execute(name string, args map[string]interface{})
```

### 2. **Strategy Pattern**

Providers de LLM implementan la misma interfaz con diferentes estrategias.

```go
type LLMProvider interface {
    Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error)
    GetDefaultModel() string
}
```

### 3. **Observer Pattern**

MessageBus permite suscripción a eventos.

```go
type MessageBus struct {
    inbound  chan InboundMessage
    outbound chan OutboundMessage
}
```

### 4. **Builder Pattern**

ContextBuilder construye el contexto paso a paso.

```go
type ContextBuilder struct {
    workspace string
    tools     *ToolRegistry
}

func (cb *ContextBuilder) BuildMessages(history, summary, message, skills []string, channel, chatID string) []Message
```

### 5. **Factory Pattern**

Creación de providers basada en configuración.

```go
func CreateProvider(cfg *config.Config) (LLMProvider, error)
```

### 6. **Template Method**

Specialists heredan comportamiento base con variaciones.

---

## 📊 Diagrama de Dependencias

```
cmd/MakoClaw
├── pkg/agent
│   ├── pkg/bus
│   ├── pkg/providers
│   ├── pkg/session
│   ├── pkg/skills
│   ├── pkg/tools
│   └── pkg/auth
├── pkg/web
│   ├── pkg/agent
│   ├── pkg/workflow
│   ├── pkg/auth
│   ├── pkg/channels
│   └── pkg/mcp
├── pkg/channels
│   └── pkg/bus
├── pkg/config
├── pkg/cron
├── pkg/mcp
├── pkg/heartbeat
├── pkg/logger
├── pkg/migrate
└── pkg/voice

pkg/agent
├── pkg/bus
├── pkg/config
├── pkg/logger
├── pkg/providers
├── pkg/session
├── pkg/skills
├── pkg/tools
└── pkg/utils
```

---

## 🔒 Seguridad

### Aislamiento
- Tools pueden restringirse al workspace
- Validación de paths previene directory traversal
- Configuración sensible en archivos con permisos restrictivos

### Autenticación
- Soporte OAuth 2.0 con PKCE
- Almacenamiento seguro de tokens
- Refresh automático
- Multi-user con aislamiento

### Ejecución
- Timeouts en operaciones shell
- Context cancellation para operaciones largas
- Rate limiting implícito por diseño
- Deny patterns para comandos peligrosos

### Validación
- User resolution por canal
- Whitelist de usuarios por canal
- Validación de input en tools

---

## 📈 Escalabilidad

### Vertical
- Optimizado para hardware mínimo
- Uso eficiente de goroutines
- Minimización de allocations
- Pooling de conexiones

### Horizontal
- Stateless por diseño (con persistencia en disco)
- Múltiples instancias posibles
- Compartir workspace entre instancias
- Load balancing via Message Bus

---

## 🔄 Ciclos de Vida

### Ciclo de Vida del Agente
1. **Inicialización**: Carga config, registra tools, inicializa providers
2. **Running**: Procesa mensajes del bus
3. **Shutdown**: Guarda sesiones, cierra conexiones graceful

### Ciclo de Vida de un Mensaje
1. **Receive**: Canal recibe mensaje
2. **Queue**: Bus encola mensaje
3. **Process**: Agent Manager procesa
4. **Respond**: Respuesta se encola en bus
5. **Deliver**: Canal envía respuesta

---

## 🎯 Decisiones de Arquitectura Clave

### 1. **Go como Lenguaje**
- **Razón**: Eficiencia, binario único, excelente concurrency
- **Trade-off**: Menor ecosistema ML que Python

### 2. **Message Bus Interno**
- **Razón**: Desacoplamiento, testabilidad, flexibilidad
- **Trade-off**: Overhead de serialización mínimo

### 3. **Skills como Markdown**
- **Razón**: Fácil de crear, versionar, y entender
- **Trade-off**: Menos estructurado que código

### 4. **SQLite para Sesiones**
- **Razón**: Zero-config, portable, suficiente para el uso
- **Trade-off**: No escala a múltiples servidores fácilmente

### 5. **Tool Registry Dinámico**
- **Razón**: Extensibilidad en runtime
- **Trade-off**: Menor type safety en tiempo de compilación

### 6. **Multi-Agent Architecture**
- **Razón**: Especialización, mejor rendimiento, escalabilidad
- **Trade-off**: Complejidad de coordinación

### 7. **MCP Protocol**
- **Razón**: Interoperabilidad con herramientas de terceros
- **Trade-off**: Dependencia de implementaciones externas

---

## 🚀 Integraciones Externas

### LLM Providers
- OpenRouter (multi-modelo)
- Anthropic Claude
- OpenAI GPT
- Groq (rápido y gratis)
- Ollama (local)

### Web Search
- Brave Search API (2000 consultas/mes gratis)

### Voice
- Groq Whisper (transcripción)

---

Para más detalles sobre componentes específicos, ver:
- [Flujo de Datos](./data-flow.md)
- [Componentes Principales](./components.md)
- [Diagramas del Sistema](./diagrams.md)

---

<div align="center">

**🦈 MakoClaw — The Apex AI Agent**

Apex Efficiency. Infinite Possibilities.

</div>

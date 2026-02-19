# Visión General de la Arquitectura

KakoClaw está diseñado con una arquitectura modular y desacoplada que permite la extensibilidad manteniendo un footprint mínimo.

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
- Mínimo uso de recursos (<10MB RAM)
- Inicio rápido (<1 segundo)
- Operaciones no bloqueantes donde sea posible

### 4. **Extensibilidad**
- Sistema de plugins (skills)
- Tools fácilmente agregables
- Canales configurables

## 🏗️ Arquitectura de Alto Nivel

```
┌─────────────────────────────────────────────────────────────┐
│                      KakoClaw Application                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │     CLI     │    │   Gateway   │    │   Cron      │     │
│  │   (cmd)     │    │  (server)   │    │  (scheduler)│     │
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
│              │      Agent Loop         │                   │
│              │   (message processor)   │                   │
│              └───────────┬─────────────┘                   │
│                          │                                  │
│              ┌───────────┴───────────┐                     │
│              │                       │                     │
│              ▼                       ▼                     │
│    ┌─────────────────┐   ┌─────────────────┐              │
│    │   LLM Provider  │   │  Tool Registry  │              │
│    │  (OpenRouter,   │   │                 │              │
│    │   Claude, etc)  │   │  [Filesystem]   │              │
│    └─────────────────┘   │  [Web Search]   │              │
│                          │  [Shell Exec]   │              │
│                          │  [Spawn]        │              │
│                          └─────────────────┘              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 📦 Componentes Principales

### 1. **CLI (cmd/KakoClaw)**
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
- `cron`: Gestión de tareas
- `skills`: Gestión de skills
- `auth`: Autenticación

### 2. **Message Bus (pkg/bus)**
Sistema de mensajería interno desacoplado.

**Características:**
- Cola thread-safe
- Buffering configurable
- Context cancellation
- Soporte para múltiples consumers

**Flujo de mensajes:**
1. Canales publican mensajes entrantes
2. Agent Loop consume y procesa
3. Respuestas se publican como salientes
4. Canales envían al usuario

### 3. **Agent Loop (pkg/agent)**
Núcleo del procesamiento de mensajes.

**Responsabilidades:**
- Construcción de contexto
- Iteración con LLM
- Ejecución de tools
- Gestión de sesiones
- Resumen automático de historial

**Algoritmo de procesamiento:**
```
1. Recibir mensaje del bus
2. Construir contexto (historial + skills + system)
3. Llamar a LLM con tools disponibles
4. Mientras LLM solicite tools:
   a. Ejecutar tool
   b. Enviar resultado a LLM
5. Retornar respuesta final
6. Guardar en sesión
7. Trigger resumen si es necesario
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
- `read_file`: Lectura de archivos
- `write_file`: Escritura de archivos
- `list_dir`: Listado de directorios
- `exec`: Ejecución shell
- `web_search`: Búsqueda web
- `web_fetch`: Obtención de URLs
- `message`: Envío de mensajes
- `spawn`: Creación de subagentes
- `schedule`: Tareas programadas

### 5. **Providers (pkg/providers)**
Abstracción de proveedores de LLM.

**Implementaciones:**
- HTTP Provider (genérico)
- Claude Provider (Anthropic)
- Codex Provider (OpenAI)

**Características:**
- Interfaz común
- Manejo de errores consistente
- Soporte para tool calling

### 6. **Channels (pkg/channels)**
Integraciones con plataformas de mensajería.

**Canales soportados:**
- Telegram
- Discord
- Slack
- WhatsApp
- QQ
- DingTalk
- Feishu
- MaixCAM

**Patrón:**
- Cada canal implementa la interfaz `Channel`
- Manager coordina múltiples canales
- Soporte para transcripción de voz (Groq)

### 7. **Session Manager (pkg/session)**
Gestión de historial de conversaciones.

**Características:**
- Persistencia en disco
- Resumen automático
- Gestión de ventana de contexto
- Múltiples sesiones concurrentes

### 8. **Skills System (pkg/skills)**
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
Instructions for the agent...
```

## 🔄 Flujo de Datos Detallado

### Flujo de un Mensaje

```
Usuario envía mensaje
        │
        ▼
┌───────────────┐
│    Canal      │ (Telegram/Discord/etc)
│   (Webhook/   │
│    Polling)   │
└───────┬───────┘
        │
        │ Convierte a InboundMessage
        ▼
┌───────────────┐
│   MessageBus  │
│   (publish)   │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Agent Loop  │
│  (consume)    │
└───────┬───────┘
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

## 📊 Diagrama de Dependencias

```
cmd/KakoClaw
├── pkg/agent
│   ├── pkg/bus
│   ├── pkg/providers
│   ├── pkg/session
│   ├── pkg/tools
│   └── pkg/skills
├── pkg/auth
├── pkg/channels
│   └── pkg/bus
├── pkg/config
├── pkg/cron
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

## 🔒 Seguridad

### Aislamiento
- Tools pueden restringirse al workspace
- Validación de paths previene directory traversal
- Configuración sensible en archivos con permisos restrictivos

### Autenticación
- Soporte OAuth 2.0 con PKCE
- Almacenamiento seguro de tokens
- Refresh automático

### Ejecución
- Timeouts en operaciones shell
- Context cancellation para operaciones largas
- Rate limiting implícito por diseño

## 📈 Escalabilidad

### Vertical
- Optimizado para hardware mínimo
- Uso eficiente de goroutines
- Minimización de allocations

### Horizontal
- Stateless por diseño (con persistencia en disco)
- Múltiples instancias posibles
- Compartir workspace entre instancias

## 🔄 Ciclos de Vida

### Ciclo de Vida del Agente
1. **Inicialización**: Carga config, registra tools, inicializa provider
2. **Running**: Procesa mensajes del bus
3. **Shutdown**: Guarda sesiones, cierra conexiones graceful

### Ciclo de Vida de un Mensaje
1. **Receive**: Canal recibe mensaje
2. **Queue**: Bus encola mensaje
3. **Process**: Agent Loop procesa
4. **Respond**: Respuesta se encola en bus
5. **Deliver**: Canal envía respuesta

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

---

Para más detalles sobre componentes específicos, ver:
- [Flujo de Datos](./data-flow.md)
- [Componentes Principales](./components.md)
- [Diagramas del Sistema](./diagrams.md)

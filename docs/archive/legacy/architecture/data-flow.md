# Flujo de Datos en MakoClaw

Este documento describe en detalle cómo fluye la información a través del sistema MakoClaw.

## 🔄 Flujo General

```
Entrada → Parsing → Enqueue → Procesamiento → LLM → Tools → Salida
```

## 📥 Flujo de Entrada

### 1. Recepción por Canal

Cada canal tiene su propia forma de recibir mensajes:

#### Telegram
```
Telegram API → Webhook/Polling → TelegramChannel.ParseUpdate()
```

#### Discord
```
Discord Gateway → discordgo → DiscordChannel.messageHandler()
```

#### CLI
```
Usuario → stdin → bufio.Scanner → agentCmd()
```

### 2. Normalización del Mensaje

Todos los mensajes se convierten a una estructura común:

```go
type InboundMessage struct {
    Channel    string // telegram, discord, cli
    SenderID   string // ID del usuario
    ChatID     string // ID del chat/grupo
    Content    string // Contenido del mensaje
    SessionKey string // Identificador único de sesión
}
```

### 3. Validación y Filtrado

```
InboundMessage
    │
    ├──▶ ¿Canal habilitado? ──No──▶ Drop
    │
    ├──▶ ¿Usuario permitido? ──No──▶ Drop
    │
    └──▶ ¿Mensaje válido? ──No──▶ Drop
```

### 4. Enqueue al Bus

```go
// Thread-safe operation
msgBus.PublishInbound(msg)
```

## ⚙️ Procesamiento

### 1. Consumo del Bus

```go
for {
    select {
    case msg := <-msgBus.Inbound():
        // Procesar mensaje
    case <-ctx.Done():
        return
    }
}
```

### 2. Construcción del Contexto

El `ContextBuilder` construye el contexto completo:

```
Contexto Final = System Prompt + Skills + Memory + Historial + Mensaje Actual
```

**Paso a paso:**

```go
// 1. System Prompt base
messages := []Message{
    {Role: "system", Content: baseSystemPrompt},
}

// 2. Agregar skills disponibles
if skills := cb.loadSkills(); len(skills) > 0 {
    messages = append(messages, Message{
        Role: "system", 
        Content: fmt.Sprintf("Available skills:\n%s", strings.Join(skills, "\n")),
    })
}

// 3. Agregar memory (resumen previo)
if summary != "" {
    messages = append(messages, Message{
        Role: "system",
        Content: fmt.Sprintf("Previous conversation summary: %s", summary),
    })
}

// 4. Agregar historial (últimos N mensajes)
messages = append(messages, history...)

// 5. Agregar mensaje actual
messages = append(messages, Message{
    Role: "user",
    Content: content,
})
```

### 3. Definiciones de Tools

```go
toolDefs := toolRegistry.GetDefinitions()
// Convierte cada Tool a formato del provider
```

**Ejemplo de definición:**
```json
{
  "type": "function",
  "function": {
    "name": "read_file",
    "description": "Read content from a file",
    "parameters": {
      "type": "object",
      "properties": {
        "file_path": {
          "type": "string",
          "description": "Path to the file"
        },
        "offset": {
          "type": "integer",
          "description": "Line number to start from"
        },
        "limit": {
          "type": "integer",
          "description": "Number of lines to read"
        }
      },
      "required": ["file_path"]
    }
  }
}
```

### 4. Llamada al LLM

```go
response, err := provider.Chat(ctx, messages, toolDefs, model, options)
```

**Estructura de respuesta:**
```go
type LLMResponse struct {
    Content      string     // Respuesta en texto
    ToolCalls    []ToolCall // Tools solicitados
    FinishReason string     // stop, tool_calls, length
    Usage        *UsageInfo // Tokens usados
}
```

## 🛠️ Ejecución de Tools

### Caso 1: Respuesta Directa (sin tools)

```
Usuario: "¿Qué hora es?"
    │
    ▼
LLM Response: {
    Content: "Son las 3:45 PM",
    ToolCalls: []
}
    │
    ▼
Guardar en sesión
    │
    ▼
Enviar respuesta al usuario
```

### Caso 2: Ejecución de Tools

```
Usuario: "Lee el archivo config.json"
    │
    ▼
LLM Response: {
    Content: "",
    ToolCalls: [{
        Name: "read_file",
        Arguments: {"file_path": "config.json"}
    }]
}
    │
    ▼
Ejecutar read_file(config.json)
    │
    ▼
Resultado: "{...contenido...}"
    │
    ▼
Enviar resultado a LLM
    │
    ▼
LLM Response: {
    Content: "Aquí está el contenido del archivo...",
    ToolCalls: []
}
    │
    ▼
Guardar y enviar al usuario
```

### Caso 3: Múltiples Tools

```
Usuario: "Busca información sobre Go y crea un resumen"
    │
    ▼
LLM solicita: web_search("Go programming language")
    │
    ▼
Resultado de búsqueda
    │
    ▼
LLM solicita: write_file("/tmp/resumen.md", contenido)
    │
    ▼
Archivo creado
    │
    ▼
LLM responde: "He creado un resumen en /tmp/resumen.md"
```

## Formato de Mensajes en la Conversación

### Mensaje de Usuario
```json
{
  "role": "user",
  "content": "Lee el archivo README.md"
}
```

### Mensaje del Asistente (con tool calls)
```json
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_123",
      "type": "function",
      "function": {
        "name": "read_file",
        "arguments": "{\"file_path\": \"README.md\"}"
      }
    }
  ]
}
```

### Mensaje de Tool
```json
{
  "role": "tool",
  "tool_call_id": "call_123",
  "content": "Contenido del archivo README.md..."
}
```

### Mensaje del Asistente (final)
```json
{
  "role": "assistant",
  "content": "El README.md contiene información sobre..."
}
```

## 📤 Flujo de Salida

### 1. Publicación al Bus

```go
msgBus.PublishOutbound(OutboundMessage{
    Channel: msg.Channel,
    ChatID:  msg.ChatID,
    Content: response,
})
```

### 2. Entrega por Canal

Cada canal implementa su método de envío:

#### Telegram
```go
telegramBot.SendMessage(telego.SendMessageParams{
    ChatID: telego.ChatID{ID: chatID},
    Text:   content,
})
```

#### Discord
```go
session.ChannelMessageSend(channelID, content)
```

### 3. Manejo de Errores de Entrega

```
Intentar envío
    │
    ├──▶ Éxito ──▶ Done
    │
    └──▶ Error ──▶ Log error
              │
              ├──▶ Reintentar (max 3)
              │
              └──▶ Guardar en cola de dead letter
```

## 💾 Persistencia

### Guardado de Sesiones

```
Después de cada interacción:
    │
    ▼
SessionManager.Save(session)
    │
    ▼
JSON → ~/.MakoClaw/workspace/sessions/<session_key>.json
```

**Estructura del archivo:**
```json
{
  "key": "telegram:123456",
  "messages": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ],
  "summary": "Resumen de la conversación...",
  "created_at": "2026-02-12T10:00:00Z",
  "updated_at": "2026-02-12T10:30:00Z"
}
```

### Resumen Automático

Cuando el historial crece:

```
Historial > 20 mensajes O tokens > 75% del límite
    │
    ▼
Trigger resumen (async)
    │
    ▼
Enviar historial a LLM: "Resume esta conversación"
    │
    ▼
Guardar resumen, truncar historial a últimos 4 mensajes
```

## 🔄 Flujos Especiales

### Tareas Programadas (Cron)

```
Cron Service (cada minuto)
    │
    ▼
Revisar trabajos programados
    │
    ▼
¿Job debe ejecutarse? ──Sí──▶ Crear mensaje de sistema
    │                              │
    No                             ▼
    │                         Enqueue al Bus
    ▼                              │
Esperar próximo minuto           ▼
                              Agent Loop procesa
                                   │
                                   ▼
                              Enviar respuesta
```

### Subagentes (Spawn)

```
Agent principal
    │
    ▼
Solicita spawn("tarea paralela")
    │
    ▼
Crear nuevo Agent Loop
    │
    ▼
Ejecutar en goroutine separada
    │
    ├──▶ Procesa tarea ──▶ Retorna resultado
    │
    ▼
Resultado disponible para agente principal
```

### Mensajes de Sistema

```
Sistema externo (cron, heartbeat, etc.)
    │
    ▼
Crear InboundMessage con channel="system"
    │
    ▼
Agent Loop detecta channel="system"
    │
    ▼
Procesa con contexto especial
    │
    ▼
Envía respuesta al canal original (si aplica)
```

## 📊 Diagrama de Secuencia

```
Usuario    Canal    Bus    Agent    LLM    Tools
  │         │        │       │       │       │
  │──msg───▶│        │       │       │       │
  │         │──msg──▶│       │       │       │
  │         │        │──msg─▶│       │       │
  │         │        │       │───────▶│       │
  │         │        │       │       │       │
  │         │        │       │◀──────│       │
  │         │        │       │       │       │
  │         │        │       │───────▶│       │
  │         │        │       │       │       │
  │         │        │       │◀──────│       │
  │         │        │       │       │       │
  │         │        │       │───────────────▶│
  │         │        │       │       │       │
  │         │        │       │◀───────────────│
  │         │        │       │       │       │
  │         │        │◀──────│       │       │
  │         │◀───────│       │       │       │
  │◀────────│        │       │       │       │
  │         │        │       │       │       │
```

## 🔍 Logging del Flujo

Cada etapa genera logs estructurados:

```go
// Recepción
logger.InfoCF("channel", "Message received",
    map[string]interface{}{
        "channel": msg.Channel,
        "sender":  msg.SenderID,
    })

// Procesamiento
logger.InfoCF("agent", "Processing message",
    map[string]interface{}{
        "session": msg.SessionKey,
        "content_length": len(msg.Content),
    })

// LLM Call
logger.InfoCF("llm", "Calling provider",
    map[string]interface{}{
        "provider": "openrouter",
        "model":    model,
        "messages": len(messages),
    })

// Tool Execution
logger.InfoCF("tool", "Tool execution completed",
    map[string]interface{}{
        "tool":        name,
        "duration_ms": duration.Milliseconds(),
    })

// Respuesta
logger.InfoCF("agent", "Response sent",
    map[string]interface{}{
        "session":       msg.SessionKey,
        "response_length": len(response),
        "iterations":    iteration,
    })
```

## 🎯 Optimizaciones

### 1. Batch Processing
Múltiples mensajes pueden procesarse en batch si llegan simultáneamente.

### 2. Caché de Tool Definitions
Las definiciones de tools se generan una vez y se reutilizan.

### 3. Lazy Loading de Skills
Los skills se cargan una vez al inicio y se mantienen en memoria.

### 4. Connection Pooling
Las conexiones HTTP se reutilizan entre requests.

---

Para entender los componentes individuales, ver [Componentes Principales](./components.md).

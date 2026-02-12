# Tools API Reference

Documentación de referencia para el sistema de Tools de PicoClaw.

## Índice

- [Overview](#overview)
- [Tool Interface](#tool-interface)
- [ContextualTool Interface](#contextualtool-interface)
- [Tool Registry](#tool-registry)
- [Built-in Tools](#built-in-tools)
- [Creating Custom Tools](#creating-custom-tools)

## Overview

El sistema de Tools permite extender las capacidades del agente mediante herramientas ejecutables. Cada tool puede:

- Leer y escribir archivos
- Ejecutar comandos
- Consultar APIs externas
- Enviar mensajes
- Y mucho más

## Tool Interface

```go
type Tool interface {
    // Name retorna el nombre único del tool
    // Usado por el LLM para identificar el tool
    Name() string

    // Description retorna una descripción para el LLM
    // Debe explicar claramente qué hace el tool
    Description() string

    // Parameters define los parámetros aceptados
    // Debe seguir el formato JSON Schema
    Parameters() map[string]interface{}

    // Execute ejecuta el tool con los argumentos proporcionados
    // Retorna el resultado como string o error
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

### Ejemplo de Implementación

```go
package tools

import (
    "context"
    "fmt"
)

// EchoTool repite el mensaje proporcionado
type EchoTool struct{}

func NewEchoTool() *EchoTool {
    return &EchoTool{}
}

func (t *EchoTool) Name() string {
    return "echo"
}

func (t *EchoTool) Description() string {
    return "Repite el mensaje proporcionado. Útil para pruebas."
}

func (t *EchoTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "message": map[string]interface{}{
                "type":        "string",
                "description": "El mensaje a repetir",
            },
        },
        "required": []string{"message"},
    }
}

func (t *EchoTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    message, ok := args["message"].(string)
    if !ok {
        return "", fmt.Errorf("message debe ser string")
    }
    return message, nil
}
```

## ContextualTool Interface

Para tools que necesitan conocer el contexto del mensaje:

```go
type ContextualTool interface {
    Tool
    SetContext(channel, chatID string)
}
```

### Ejemplo de Uso

```go
type MessageTool struct {
    channel string
    chatID  string
}

func (t *MessageTool) SetContext(channel, chatID string) {
    t.channel = channel
    t.chatID = chatID
}

func (t *MessageTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Usar t.channel y t.chatID para enviar mensaje
}
```

## Tool Registry

El registro central de tools:

```go
type ToolRegistry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}

// Crear nuevo registro
registry := tools.NewToolRegistry()

// Registrar un tool
registry.Register(tool)

// Obtener un tool
tool, ok := registry.Get("nombre_del_tool")

// Ejecutar un tool
result, err := registry.Execute(ctx, "nombre_del_tool", args)

// Ejecutar con contexto (para ContextualTools)
result, err := registry.ExecuteWithContext(ctx, "tool", args, "telegram", "123456")

// Obtener todas las definiciones
// Útil para enviar al LLM
definitions := registry.GetDefinitions()

// Listar nombres de tools
names := registry.List()

// Obtener conteo
count := registry.Count()
```

## Built-in Tools

### 1. Read File Tool

```go
tool := tools.NewReadFileTool(workspace string, restrict bool)
```

**Parameters:**
```json
{
  "file_path": {
    "type": "string",
    "description": "Ruta del archivo a leer"
  },
  "offset": {
    "type": "integer",
    "description": "Línea inicial (0-based)"
  },
  "limit": {
    "type": "integer",
    "description": "Número máximo de líneas"
  }
}
```

**Example:**
```json
{
  "file_path": "config.json",
  "offset": 0,
  "limit": 50
}
```

**Returns:** Contenido del archivo como string

**Errors:**
- Archivo no existe
- Permisos insuficientes
- Path fuera del workspace (si restrict=true)

### 2. Write File Tool

```go
tool := tools.NewWriteFileTool(workspace string, restrict bool)
```

**Parameters:**
```json
{
  "file_path": {
    "type": "string",
    "description": "Ruta donde escribir"
  },
  "content": {
    "type": "string",
    "description": "Contenido a escribir"
  }
}
```

**Example:**
```json
{
  "file_path": "hello.txt",
  "content": "Hola Mundo"
}
```

**Returns:** Confirmación de éxito

**Notes:**
- Crea directorios si no existen
- Sobrescribe archivo si existe

### 3. List Directory Tool

```go
tool := tools.NewListDirTool(workspace string, restrict bool)
```

**Parameters:**
```json
{
  "path": {
    "type": "string",
    "description": "Ruta del directorio"
  },
  "recursive": {
    "type": "boolean",
    "description": "Listar recursivamente"
  }
}
```

**Returns:** Lista formateada de archivos y directorios

### 4. Exec Tool

```go
tool := tools.NewExecTool(workspace string, restrict bool)
```

**Parameters:**
```json
{
  "command": {
    "type": "string",
    "description": "Comando a ejecutar"
  },
  "timeout": {
    "type": "integer",
    "description": "Timeout en segundos (default: 30)"
  }
}
```

**Example:**
```json
{
  "command": "ls -la",
  "timeout": 10
}
```

**Returns:** Output del comando (stdout + stderr)

**Security:**
- Respeta `restrict_to_workspace`
- Timeout para evitar comandos infinitos

### 5. Web Search Tool

```go
tool := tools.NewWebSearchTool(apiKey string, maxResults int)
```

**Parameters:**
```json
{
  "query": {
    "type": "string",
    "description": "Término de búsqueda"
  }
}
```

**Example:**
```json
{
  "query": "golang best practices"
}
```

**Returns:** Resultados de búsqueda formateados

**Note:** Requiere API key de Brave Search

### 6. Web Fetch Tool

```go
tool := tools.NewWebFetchTool(maxLength int)
```

**Parameters:**
```json
{
  "url": {
    "type": "string",
    "description": "URL a obtener"
  },
  "format": {
    "type": "string",
    "enum": ["markdown", "text", "html"],
    "description": "Formato de salida"
  },
  "max_length": {
    "type": "integer",
    "description": "Longitud máxima del contenido"
  }
}
```

**Example:**
```json
{
  "url": "https://example.com",
  "format": "markdown",
  "max_length": 5000
}
```

**Returns:** Contenido de la URL en el formato especificado

### 7. Message Tool

```go
tool := tools.NewMessageTool()
```

**Parameters:**
```json
{
  "content": {
    "type": "string",
    "description": "Contenido del mensaje"
  },
  "channel": {
    "type": "string",
    "description": "Canal destino (opcional)"
  },
  "to": {
    "type": "string",
    "description": "Destinatario (opcional)"
  }
}
```

**Note:** Implementa `ContextualTool`

### 8. Spawn Tool

```go
tool := tools.NewSpawnTool(manager *SubagentManager)
```

**Parameters:**
```json
{
  "task": {
    "type": "string",
    "description": "Descripción de la tarea"
  },
  "context": {
    "type": "string",
    "description": "Contexto adicional"
  }
}
```

**Returns:** Resultado de la tarea ejecutada por subagente

### 9. Edit File Tool

```go
tool := tools.NewEditFileTool(workspace string, restrict bool)
```

**Parameters:**
```json
{
  "file_path": {
    "type": "string",
    "description": "Ruta del archivo"
  },
  "old_string": {
    "type": "string",
    "description": "Texto a buscar"
  },
  "new_string": {
    "type": "string",
    "description": "Texto de reemplazo"
  }
}
```

**Example:**
```json
{
  "file_path": "config.txt",
  "old_string": "version: 1.0",
  "new_string": "version: 2.0"
}
```

**Returns:** Confirmación del cambio

**Features:**
- Preserva indentación
- Soporta múltiples reemplazos con `replaceAll`
- Verificación de cambios

## Creating Custom Tools

### Paso 1: Definir la Estructura

```go
package tools

import (
    "context"
    "fmt"
)

type MyTool struct {
    config MyConfig
}

type MyConfig struct {
    Option1 string
    Option2 int
}

func NewMyTool(config MyConfig) *MyTool {
    return &MyTool{config: config}
}
```

### Paso 2: Implementar la Interface

```go
func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Describe qué hace este tool"
}

func (t *MyTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Descripción de param1",
            },
            "param2": map[string]interface{}{
                "type":        "integer",
                "description": "Descripción de param2",
            },
        },
        "required": []string{"param1"},
    }
}

func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Extraer argumentos
    param1, ok := args["param1"].(string)
    if !ok {
        return "", fmt.Errorf("param1 debe ser string")
    }

    param2 := 0
    if v, ok := args["param2"].(float64); ok {
        param2 = int(v)
    }

    // Ejecutar lógica
    result, err := t.doSomething(ctx, param1, param2)
    if err != nil {
        return "", fmt.Errorf("error ejecutando my_tool: %w", err)
    }

    return result, nil
}
```

### Paso 3: Registrar el Tool

```go
// En el agente o donde se inicializan los tools
func setupTools() *ToolRegistry {
    registry := NewToolRegistry()
    
    // Tools existentes...
    
    // Tu nuevo tool
    myTool := NewMyTool(MyConfig{
        Option1: "valor",
        Option2: 42,
    })
    registry.Register(myTool)
    
    return registry
}
```

### Paso 4: Testing

```go
func TestMyTool_Execute(t *testing.T) {
    tool := NewMyTool(MyConfig{Option1: "test"})
    
    tests := []struct {
        name    string
        args    map[string]interface{}
        want    string
        wantErr bool
    }{
        {
            name: "success",
            args: map[string]interface{}{
                "param1": "hello",
                "param2": 42.0,
            },
            want:    "resultado esperado",
            wantErr: false,
        },
        {
            name: "missing required",
            args: map[string]interface{}{
                "param2": 42.0,
            },
            want:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tool.Execute(context.Background(), tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Execute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Best Practices

### 1. Nombres Claros

```go
// ❌ Mal
func (t *Tool) Name() string { return "t" }

// ✅ Bien
func (t *CalculatorTool) Name() string { return "calculator" }
```

### 2. Descripciones Detalladas

```go
// ❌ Vago
func (t *Tool) Description() string {
    return "Hace cosas"
}

// ✅ Detallado
func (t *WeatherTool) Description() string {
    return "Obtiene el clima actual y pronóstico para una ubicación. " +
           "Usa servicios gratuitos sin necesidad de API key. " +
           "Soporta ciudades, códigos de aeropuerto, y coordenadas GPS."
}
```

### 3. Validación de Inputs

```go
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Validar tipo
    path, ok := args["file_path"].(string)
    if !ok {
        return "", fmt.Errorf("file_path debe ser string, recibido %T", args["file_path"])
    }
    
    // Validar no vacío
    if path == "" {
        return "", fmt.Errorf("file_path no puede estar vacío")
    }
    
    // Validar formato
    if strings.Contains(path, "..") {
        return "", fmt.Errorf("path contiene caracteres inválidos")
    }
    
    // ...
}
```

### 4. Manejo de Context

```go
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Verificar cancellation
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }
    
    // Operación con timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    result, err := t.doWork(ctx)
    // ...
}
```

### 5. Retornar Información Útil

```go
// ❌ Pobre
return "OK", nil

// ✅ Informativo
return fmt.Sprintf("Archivo %s creado exitosamente (%d bytes)", 
    filePath, len(content)), nil
```

## Schema JSON para LLM

Cada tool debe proveer su schema en formato JSON Schema:

```json
{
  "type": "function",
  "function": {
    "name": "tool_name",
    "description": "Descripción clara",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {
          "type": "string",
          "description": "Descripción del parámetro",
          "enum": ["opcion1", "opcion2"]
        },
        "param2": {
          "type": "integer",
          "description": "Número entero",
          "minimum": 0,
          "maximum": 100
        }
      },
      "required": ["param1"]
    }
  }
}
```

---

Para crear tu primer tool, ver [Crear un Nuevo Tool](./creating-tools.md).

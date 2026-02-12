# Planes de Implementación de Issues

Documento detallado con planes de implementación para las issues clasificadas como "Útiles".

---

## Issue #75 - Soporte para LLMs Locales (Ollama)

### Descripción
Implementar soporte para Ollama y otros LLMs locales que permitan ejecutar modelos sin conexión a internet.

### Por qué es útil
- **Privacidad**: Los datos nunca salen de la máquina
- **Costo**: Sin costos de API
- **Offline**: Funciona sin internet
- **Latencia**: Respuestas más rápidas en local

### Cómo implementar

#### 1. Crear Provider para Ollama

**Archivo:** `pkg/providers/ollama_provider.go`

```go
package providers

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type OllamaProvider struct {
    baseURL string
    client  *http.Client
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
    if baseURL == "" {
        baseURL = "http://localhost:11434"
    }
    return &OllamaProvider{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 120 * time.Second},
    }
}

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
    // Convertir mensajes al formato de Ollama
    ollamaMessages := make([]OllamaMessage, len(messages))
    for i, msg := range messages {
        ollamaMessages[i] = OllamaMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    
    reqBody := OllamaRequest{
        Model:    model,
        Messages: ollamaMessages,
        Stream:   false,
        Options: map[string]interface{}{
            "temperature": getOption(options, "temperature", 0.7),
            "num_predict": getOption(options, "max_tokens", 2048),
        },
    }
    
    // Hacer request a Ollama API
    resp, err := p.makeRequest(ctx, "/api/chat", reqBody)
    if err != nil {
        return nil, fmt.Errorf("ollama chat failed: %w", err)
    }
    
    return &LLMResponse{
        Content: resp.Message.Content,
        Usage: &UsageInfo{
            TotalTokens: resp.EvalCount + resp.PromptEvalCount,
        },
    }, nil
}

func (p *OllamaProvider) GetDefaultModel() string {
    return "llama3.2"
}
```

#### 2. Configuración

**Archivo:** `pkg/config/config.go`

Agregar a `ProvidersConfig`:
```go
type ProvidersConfig struct {
    // ... providers existentes ...
    Ollama ProviderConfig `json:"ollama"`
}
```

#### 3. Detección Automática

**Archivo:** `pkg/providers/provider_factory.go`

```go
func CreateProvider(cfg *config.Config) (LLMProvider, error) {
    // ... código existente ...
    
    // Verificar si es modelo local de Ollama
    if strings.HasPrefix(model, "ollama/") || cfg.Providers.Ollama.APIBase != "" {
        return NewOllamaProvider(cfg.Providers.Ollama.APIBase), nil
    }
    
    // ... resto del código ...
}
```

#### 4. Ejemplo de uso

**Configuración:**
```json
{
  "providers": {
    "ollama": {
      "api_base": "http://localhost:11434"
    }
  },
  "agents": {
    "defaults": {
      "model": "llama3.2"
    }
  }
}
```

### Testing

```bash
# Verificar Ollama está corriendo
curl http://localhost:11434/api/tags

# Probar provider
picoclaw agent -m "Hola desde Ollama"
```

---

## Issue #66 - Fix Variables de Entorno {{.Name}}

### Descripción
Las variables de entorno con `{{.Name}}` no se procesan correctamente porque caarlos0/env no soporta templates.

### Por qué es útil
- Permite configuración por environment variables
- Esencial para Docker y deployments
- Buena práctica de 12-factor apps

### Cómo implementar

#### 1. Solución: Pre-procesamiento Manual

**Archivo:** `pkg/config/config.go`

```go
func LoadConfig(path string) (*Config, error) {
    cfg := DefaultConfig()

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return cfg, nil
        }
        return nil, err
    }

    if err := json.Unmarshal(data, cfg); err != nil {
        return nil, err
    }

    // Parse environment variables manualmente para providers
    if err := parseProviderEnvVars(cfg); err != nil {
        return nil, err
    }

    return cfg, nil
}

func parseProviderEnvVars(cfg *Config) error {
    providers := map[string]*ProviderConfig{
        "anthropic":  &cfg.Providers.Anthropic,
        "openai":     &cfg.Providers.OpenAI,
        "openrouter": &cfg.Providers.OpenRouter,
        "groq":       &cfg.Providers.Groq,
        "zhipu":      &cfg.Providers.Zhipu,
        "gemini":     &cfg.Providers.Gemini,
        "ollama":     &cfg.Providers.Ollama,
    }

    for name, provider := range providers {
        prefix := fmt.Sprintf("PICOCLAW_PROVIDERS_%s_", strings.ToUpper(name))
        
        if apiKey := os.Getenv(prefix + "API_KEY"); apiKey != "" {
            provider.APIKey = apiKey
        }
        if apiBase := os.Getenv(prefix + "API_BASE"); apiBase != "" {
            provider.APIBase = apiBase
        }
        if authMethod := os.Getenv(prefix + "AUTH_METHOD"); authMethod != "" {
            provider.AuthMethod = authMethod
        }
    }

    return nil
}
```

#### 2. Remover Tags Problemáticos

**Archivo:** `pkg/config/config.go`

Cambiar:
```go
// DE:
type ProviderConfig struct {
    APIKey     string `json:"api_key" env:"PICOCLAW_PROVIDERS_{{.Name}}_API_KEY"`
}

// A:
type ProviderConfig struct {
    APIKey     string `json:"api_key"`
}
```

### Testing

```bash
# Set environment variables
export PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY="sk-ant-xxxxx"
export PICOCLAW_PROVIDERS_OPENAI_API_KEY="sk-xxxxx"

# Verificar que se aplican
picoclaw status
```

---

## Issue #39 - Comando `picoclaw doctor`

### Descripción
Comando para diagnosticar problemas de configuración y entorno.

### Por qué es útil
- Ayuda a usuarios a solucionar problemas
- Detecta configuraciones incorrectas
- Verifica dependencias
- Similar a `brew doctor` o `flutter doctor`

### Cómo implementar

#### 1. Crear Comando

**Archivo:** `cmd/picoclaw/main.go`

Agregar case:
```go
case "doctor":
    doctorCmd()
```

#### 2. Implementar Lógica

**Archivo:** `pkg/doctor/doctor.go` (nuevo)

```go
package doctor

import (
    "fmt"
    "os"
    "path/filepath"
)

type CheckResult struct {
    Name    string
    Status  Status
    Message string
    Fix     string
}

type Status int

const (
    StatusOK Status = iota
    StatusWarning
    StatusError
)

func RunChecks() []CheckResult {
    var results []CheckResult
    
    results = append(results, checkConfig())
    results = append(results, checkWorkspace())
    results = append(results, checkAPIKeys())
    results = append(results, checkProviders())
    results = append(results, checkPermissions())
    results = append(results, checkDependencies())
    
    return results
}

func checkConfig() CheckResult {
    configPath := getConfigPath()
    
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return CheckResult{
            Name:    "Configuration File",
            Status:  StatusError,
            Message: fmt.Sprintf("Config file not found at %s", configPath),
            Fix:     "Run 'picoclaw onboard' to initialize",
        }
    }
    
    // Verificar JSON válido
    data, err := os.ReadFile(configPath)
    if err != nil {
        return CheckResult{
            Name:    "Configuration File",
            Status:  StatusError,
            Message: "Cannot read config file",
            Fix:     "Check file permissions",
        }
    }
    
    var cfg map[string]interface{}
    if err := json.Unmarshal(data, &cfg); err != nil {
        return CheckResult{
            Name:    "Configuration File",
            Status:  StatusError,
            Message: "Config file is not valid JSON",
            Fix:     "Fix JSON syntax errors",
        }
    }
    
    return CheckResult{
        Name:    "Configuration File",
        Status:  StatusOK,
        Message: fmt.Sprintf("Found at %s", configPath),
    }
}

func checkAPIKeys() CheckResult {
    cfg, err := LoadConfig()
    if err != nil {
        return CheckResult{
            Name:    "API Keys",
            Status:  StatusWarning,
            Message: "Cannot load config to check API keys",
        }
    }
    
    hasKey := cfg.Providers.GetAPIKey() != ""
    
    if !hasKey {
        return CheckResult{
            Name:    "API Keys",
            Status:  StatusError,
            Message: "No API key configured for any provider",
            Fix:     "Add API key to config.json or set environment variable",
        }
    }
    
    return CheckResult{
        Name:    "API Keys",
        Status:  StatusOK,
        Message: "At least one API key is configured",
    }
}

// ... más checks ...
```

#### 3. Integrar con CLI

**Archivo:** `cmd/picoclaw/main.go`

```go
func doctorCmd() {
    fmt.Println("🦞 PicoClaw Doctor")
    fmt.Println("==================")
    fmt.Println()
    
    results := doctor.RunChecks()
    
    okCount := 0
    warningCount := 0
    errorCount := 0
    
    for _, result := range results {
        icon := "✓"
        if result.Status == doctor.StatusWarning {
            icon = "⚠"
            warningCount++
        } else if result.Status == doctor.StatusError {
            icon = "✗"
            errorCount++
        } else {
            okCount++
        }
        
        fmt.Printf("%s %s: %s\n", icon, result.Name, result.Message)
        if result.Fix != "" {
            fmt.Printf("   Fix: %s\n", result.Fix)
        }
        fmt.Println()
    }
    
    fmt.Println("==================")
    fmt.Printf("✓ %d OK, ⚠ %d warnings, ✗ %d errors\n", okCount, warningCount, errorCount)
    
    if errorCount > 0 {
        os.Exit(1)
    }
}
```

### Uso

```bash
picoclaw doctor

# Salida:
🦞 PicoClaw Doctor
==================

✓ Configuration File: Found at /home/user/.picoclaw/config.json

✓ Workspace: Found at /home/user/.picoclaw/workspace

✓ API Keys: At least one API key is configured

✓ Providers: 3 providers configured

⚠ Permissions: Workspace is readable by others
   Fix: Run 'chmod 700 ~/.picoclaw/workspace'

==================
✓ 4 OK, ⚠ 1 warnings, ✗ 0 errors
```

---

## Issue #63 - Gestionar Cronjobs desde Session

### Descripción
Permitir crear, listar, modificar y eliminar tareas programadas directamente desde la conversación con el agente.

### Por qué es útil
- No requiere acceso a CLI
- Más intuitivo para usuarios no técnicos
- Contexto conversacional

### Cómo implementar

#### 1. Crear Tool de Cron

**Archivo:** `pkg/tools/cron.go` (actualizar)

```go
func (t *CronTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "action": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"list", "add", "remove", "enable", "disable"},
                "description": "Acción a realizar: list, add, remove, enable, disable",
            },
            "name": map[string]interface{}{
                "type":        "string",
                "description": "Nombre del job (requerido para add, remove, enable, disable)",
            },
            "schedule": map[string]interface{}{
                "type":        "string",
                "description": "Expresión cron o 'every X minutes/hours/days'",
            },
            "message": map[string]interface{}{
                "type":        "string",
                "description": "Mensaje que el agente procesará (requerido para add)",
            },
        },
        "required": []string{"action"},
    }
}

func (t *CronTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    action, _ := args["action"].(string)
    
    switch action {
    case "list":
        return t.listJobs()
    case "add":
        return t.addJob(args)
    case "remove":
        return t.removeJob(args)
    case "enable", "disable":
        return t.toggleJob(args, action == "enable")
    default:
        return "", fmt.Errorf("acción desconocida: %s", action)
    }
}
```

#### 2. Mejorar Parsing de Schedule

```go
func parseSchedule(schedule string) (CronSchedule, error) {
    // Soporte para "every X minutes/hours/days"
    if strings.HasPrefix(schedule, "every ") {
        parts := strings.Fields(schedule)
        if len(parts) >= 3 {
            value, _ := strconv.Atoi(parts[1])
            unit := parts[2]
            
            var ms int64
            switch unit {
            case "minute", "minutes":
                ms = int64(value * 60 * 1000)
            case "hour", "hours":
                ms = int64(value * 60 * 60 * 1000)
            case "day", "days":
                ms = int64(value * 24 * 60 * 60 * 1000)
            }
            
            return CronSchedule{
                Kind:    "every",
                EveryMS: &ms,
            }, nil
        }
    }
    
    // Si no, asumir que es expresión cron
    return CronSchedule{
        Kind: "cron",
        Expr: schedule,
    }, nil
}
```

### Ejemplos de Uso

```bash
# Listar jobs
Usuario: Muestra mis tareas programadas
Agente: *usa cron tool con action=list*

# Agregar job
Usuario: Recuérdame cada día a las 9am revisar emails
Agente: *usa cron tool con action=add, schedule="0 9 * * *"*

# Agregar job simple
Usuario: Envíame un recordatorio cada 30 minutos
Agente: *usa cron tool con action=add, schedule="every 30 minutes"*

# Eliminar job
Usuario: Elimina la tarea "daily-reminder"
Agente: *usa cron tool con action=remove*
```

---

## Issue #46 - Mejoras en Configuración

### Descripción
Sugerencias varias para mejorar el sistema de configuración.

### Posibles mejoras

#### 1. Validación de Configuración

```go
func (c *Config) Validate() error {
    var errs []string
    
    // Validar modelo
    if c.Agents.Defaults.Model == "" {
        errs = append(errs, "model is required")
    }
    
    // Validar al menos un provider
    if c.Providers.GetAPIKey() == "" {
        errs = append(errs, "at least one provider API key is required")
    }
    
    // Validar workspace
    if c.Agents.Defaults.Workspace == "" {
        errs = append(errs, "workspace is required")
    }
    
    // Validar canales
    for name, channel := range c.Channels {
        if err := validateChannel(name, channel); err != nil {
            errs = append(errs, fmt.Sprintf("channel %s: %v", name, err))
        }
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("config validation failed:\n- %s", strings.Join(errs, "\n- "))
    }
    
    return nil
}
```

#### 2. Hot Reload

```go
func (c *Config) Watch(path string) error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    
    if err := watcher.Add(path); err != nil {
        return err
    }
    
    go func() {
        for event := range watcher.Events {
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Println("Config file changed, reloading...")
                if err := c.Reload(path); err != nil {
                    log.Printf("Failed to reload config: %v", err)
                }
            }
        }
    }()
    
    return nil
}
```

#### 3. Configuración por Perfiles

```json
{
  "profiles": {
    "default": {
      "model": "gpt-4",
      "max_tokens": 8192
    },
    "fast": {
      "model": "gpt-3.5-turbo",
      "max_tokens": 2048
    },
    "local": {
      "provider": "ollama",
      "model": "llama3.2"
    }
  }
}
```

---

## Issue #43 - Mejorar Asignación Model→Provider

### Descripción
Actualmente el mapeo de modelos a providers es confuso y automático.

### Solución

#### 1. Mapeo Explícito

```go
type ModelMapping struct {
    Model    string
    Provider string
}

var defaultModelMappings = []ModelMapping{
    {Model: "gpt-4", Provider: "openai"},
    {Model: "gpt-3.5-turbo", Provider: "openai"},
    {Model: "claude-3.5-sonnet", Provider: "anthropic"},
    {Model: "claude-3-opus", Provider: "anthropic"},
    {Model: "llama3.2", Provider: "ollama"},
    {Model: "gemini-pro", Provider: "gemini"},
}

func GetProviderForModel(model string) (string, error) {
    // Verificar prefijo explícito
    if strings.Contains(model, "/") {
        parts := strings.Split(model, "/")
        return parts[0], nil
    }
    
    // Buscar en mapeos
    for _, mapping := range defaultModelMappings {
        if mapping.Model == model {
            return mapping.Provider, nil
        }
    }
    
    // Fallback a OpenRouter para modelos desconocidos
    return "openrouter", nil
}
```

#### 2. Sintaxis con Prefijo

Permitir especificar provider explícitamente:

```json
{
  "agents": {
    "defaults": {
      "model": "openai/gpt-4"
    }
  }
}
```

---

## Issue #37 - Telegram: Enviar Mensajes Proactivos

### Descripción
El bot solo puede responder mensajes, no iniciar conversaciones.

### Implementación

#### 1. Modificar TelegramChannel

**Archivo:** `pkg/channels/telegram.go`

```go
func (c *TelegramChannel) SendProactive(chatID string, message string) error {
    id, err := strconv.ParseInt(chatID, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid chat ID: %w", err)
    }
    
    params := &telego.SendMessageParams{
        ChatID: telego.ChatID{ID: id},
        Text:   message,
    }
    
    _, err = c.bot.SendMessage(params)
    return err
}
```

#### 2. Tool de Mensaje Proactivo

```go
func (t *MessageTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    content := args["content"].(string)
    channel := args["channel"].(string)
    to := args["to"].(string)
    
    // Si no hay contexto actual, es mensaje proactivo
    if t.channel == "" && channel != "" && to != "" {
        return t.sendProactive(channel, to, content)
    }
    
    // Mensaje contextual normal
    return t.sendContextual(content)
}
```

### Uso

```bash
# Desde cron o tool
picoclaw agent -m "Envía mensaje a Telegram: user123 'Recordatorio: reunión en 5 min'"
```

---

## Issue #62 - Fix Telegram allow_from

### Descripción
El filtro `allow_from` no funciona cuando el usuario tiene username en lugar de ID numérico.

### Implementación

```go
func (c *TelegramChannel) isAuthorized(update telego.Update) bool {
    var userID int64
    var username string
    
    if update.Message != nil {
        userID = update.Message.From.ID
        username = update.Message.From.Username
    } else if update.CallbackQuery != nil {
        userID = update.CallbackQuery.From.ID
        username = update.CallbackQuery.From.Username
    }
    
    // Convertir allow_from a string para comparación
    userIDStr := strconv.FormatInt(userID, 10)
    
    for _, allowed := range c.allowFrom {
        // Comparar como ID numérico
        if allowed == userIDStr {
            return true
        }
        // Comparar como username (case insensitive)
        if strings.EqualFold(allowed, username) {
            return true
        }
        // Comparar con @ prefix
        if strings.EqualFold(allowed, "@"+username) {
            return true
        }
    }
    
    return false
}
```

---

## Issue #15 - Fix Build ARM 32-bit

### Descripción
Error de compilación en ARM 32-bit: `math.MaxInt64` causa overflow.

### Implementación

```go
// En lugar de math.MaxInt64
const MaxInt = int(^uint(0) >> 1)

// O usar int64 explícitamente
var maxValue int64 = math.MaxInt64
```

Buscar todos los usos de `math.MaxInt64` y reemplazar:

```bash
grep -r "math.MaxInt64" pkg/
```

---

## Issue #9 - Rate Limiters

### Implementación básica

```go
type RateLimiter struct {
    requests map[string][]time.Time
    limit    int
    window   time.Duration
    mu       sync.RWMutex
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    
    // Limpiar requests antiguos
    if requests, ok := rl.requests[key]; ok {
        var valid []time.Time
        for _, r := range requests {
            if now.Sub(r) < rl.window {
                valid = append(valid, r)
            }
        }
        rl.requests[key] = valid
    }
    
    // Verificar límite
    if len(rl.requests[key]) >= rl.limit {
        return false
    }
    
    // Agregar request
    rl.requests[key] = append(rl.requests[key], now)
    return true
}
```

---

*Para más detalles sobre implementación específica, consultar cada archivo mencionado.*

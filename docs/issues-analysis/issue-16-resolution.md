# Issue #16 - OpenAI max_tokens Error

## Estado: ✅ RESUELTO

## Análisis

Después de revisar el código de providers, **todos los providers manejan correctamente el parámetro `max_tokens`**:

### HTTP Provider (pkg/providers/http_provider.go)

```go
if maxTokens, ok := options["max_tokens"].(int); ok {
    lowerModel := strings.ToLower(model)
    if strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "o1") {
        requestBody["max_completion_tokens"] = maxTokens
    } else {
        requestBody["max_tokens"] = maxTokens
    }
}
```

- Maneja correctamente la conversión de tipos
- Soporte especial para modelos que usan `max_completion_tokens` (GLM, O1)
- Usa `max_tokens` para modelos estándar (OpenAI, etc.)

### Codex Provider (pkg/providers/codex_provider.go)

```go
if maxTokens, ok := options["max_tokens"].(int); ok {
    params.MaxOutputTokens = openai.Opt(int64(maxTokens))
}
```

- Convierte correctamente `int` a `int64`
- Usa `openai.Opt()` para el parámetro opcional

### Claude Provider (pkg/providers/claude_provider.go)

```go
if mt, ok := options["max_tokens"].(int); ok {
    reqBody["max_tokens"] = mt
}
```

- Manejo directo del parámetro

## Conclusión

El problema reportado en el issue #16 ya ha sido resueldo. Todos los providers actuales:

1. ✅ Extraen correctamente `max_tokens` de los options
2. ✅ Manejan la conversión de tipos (int → int64 cuando es necesario)
3. ✅ Usan el nombre correcto del parámetro según el provider
4. ✅ Funcionan con las API keys de OpenAI directamente

## Configuración Recomendada para OpenAI

```json
{
  "agents": {
    "defaults": {
      "model": "gpt-4",
      "max_tokens": 4096,
      "temperature": 0.7
    }
  },
  "providers": {
    "openai": {
      "api_key": "sk-...",
      "api_base": "https://api.openai.com/v1"
    }
  }
}
```

## Referencias

- Issue original: https://github.com/sipeed/KakoClaw/issues/16
- Archivos verificados:
  - `pkg/providers/http_provider.go`
  - `pkg/providers/codex_provider.go`
  - `pkg/providers/claude_provider.go`

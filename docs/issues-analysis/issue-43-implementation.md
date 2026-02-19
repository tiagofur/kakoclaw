# Issue #43 - Improved Model to Provider Mapping

## Estado: ✅ IMPLEMENTADO

## Mejora Implementada

### Sintaxis Explícita provider/model

Ahora puedes especificar explícitamente el provider usando la sintaxis `provider/model`:

```json
{
  "agents": {
    "defaults": {
      "model": "openai/gpt-4"
    }
  }
}
```

Esto fuerza a usar el provider OpenAI independientemente de otras configuraciones.

### Providers Soportados

- `openai/` - OpenAI (GPT-4, GPT-3.5, etc.)
- `anthropic/` - Anthropic (Claude models)
- `openrouter/` - OpenRouter (múltiples modelos)
- `groq/` - Groq (Llama, Mixtral, etc.)
- `zhipu/` - Zhipu (GLM models)
- `gemini/` - Google (Gemini models)
- `moonshot/` - Moonshot (Kimi models)
- `nvidia/` - NVIDIA NIM

## Ejemplos

### Configuración JSON

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-3-5-sonnet-20241022"
    }
  },
  "providers": {
    "anthropic": {
      "api_key": "sk-ant-xxxxx"
    }
  }
}
```

### Detección Automática (sin prefijo)

Si no usas prefijo, el provider se detecta automáticamente:

```json
{
  "agents": {
    "defaults": {
      "model": "gpt-4"
    }
  }
}
```

Auto-detección:
- `gpt-*` → openai
- `claude*` → anthropic
- `kimi*` → moonshot
- `gemini*` → gemini
- `glm*` → zhipu
- otros → openrouter

## Uso

### Ver Provider Asignado

```bash
KakoClaw status

# Output:
# Model: openai/gpt-4
# Provider: openai (model: gpt-4)
```

### Cambiar Provider

Simplemente cambia el prefijo en el modelo:

```bash
# Para usar GPT-4 via OpenRouter en lugar de OpenAI directo:
# Editar ~/.KakoClaw/config.json
{
  "agents": {
    "defaults": {
      "model": "openrouter/openai/gpt-4"
    }
  }
}
```

## Ventajas

1. **Control explícito**: Sabes exactamente qué provider se usa
2. **Flexibilidad**: Un mismo modelo puede usarse con diferentes providers
3. **Claridad**: Elimina ambigüedad en la configuración
4. **Backward compatible**: La detección automática sigue funcionando

## Referencias

- Implementación: `pkg/providers/http_provider.go`
- Tests: `pkg/providers/http_provider_test.go`
- Issue original: https://github.com/sipeed/KakoClaw/issues/43

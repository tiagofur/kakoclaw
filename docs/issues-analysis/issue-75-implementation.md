# Issue #75 - Ollama Support (Local LLMs)

## Estado: ✅ IMPLEMENTADO

## Descripción

Soporte completo para Ollama, permitiendo ejecutar modelos de lenguaje localmente sin necesidad de conexión a internet ni API keys.

## Configuración

### 1. Instalar Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows
# Descargar desde https://ollama.com/download/windows
```

### 2. Descargar un Modelo

```bash
ollama pull llama3.2
# o
ollama pull mistral
# o
ollama pull codellama
```

### 3. Configurar PicoClaw

Editar `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "model": "ollama/llama3.2"
    }
  },
  "providers": {
    "ollama": {
      "api_base": "http://localhost:11434"
    }
  }
}
```

O simplemente:

```json
{
  "agents": {
    "defaults": {
      "model": "llama3.2"
    }
  }
}
```

(Si Ollama está en el puerto por defecto, no necesitas configurar providers.ollama)

## Uso

```bash
picoclaw agent -m "Hola, ¿cómo estás?"

# O modo interactivo
picoclaw agent
```

## Modelos Populares

| Modelo | Descripción | Comando |
|--------|-------------|---------|
| llama3.2 | Meta Llama 3.2 (1B-3B) | `ollama pull llama3.2` |
| llama3.1 | Meta Llama 3.1 (8B+) | `ollama pull llama3.1` |
| mistral | Mistral 7B | `ollama pull mistral` |
| codellama | Code Llama | `ollama pull codellama` |
| phi3 | Microsoft Phi-3 | `ollama pull phi3` |
| gemma2 | Google Gemma 2 | `ollama pull gemma2` |
| qwen2.5 | Alibaba Qwen 2.5 | `ollama pull qwen2.5` |

## Ventajas

- ✅ **Privacidad**: Los datos nunca salen de tu máquina
- ✅ **Sin costos**: Sin API keys ni cuotas
- ✅ **Offline**: Funciona sin conexión a internet
- ✅ **Velocidad**: Sin latencia de red (depende de tu hardware)
- ✅ **Control**: Tú eliges qué modelos usar

## Limitaciones

- ⚠️ **Hardware**: Requiere suficiente RAM (mínimo 8GB recomendado)
- ⚠️ **Calidad**: Modelos locales son generalmente más pequeños que GPT-4/Claude
- ⚠️ **Velocidad**: Depende de tu CPU/GPU
- ⚠️ **Funciones avanzadas**: No soporta tool calling en todos los modelos

## Requisitos de Hardware

| Modelo | RAM Mínima | VRAM (GPU) |
|--------|-----------|-----------|
| llama3.2 (1B) | 4 GB | 2 GB |
| llama3.2 (3B) | 6 GB | 4 GB |
| llama3.1 (8B) | 12 GB | 8 GB |
| mistral (7B) | 12 GB | 8 GB |
| codellama (7B) | 12 GB | 8 GB |

## Verificación

```bash
# Verificar que Ollama está corriendo
curl http://localhost:11434/api/tags

# Verificar PicoClaw puede conectar
picoclaw status

# Probar conversación
picoclaw agent -m "Hola desde Ollama"
```

## Troubleshooting

### "Ollama API error"

Asegúrate de que Ollama está corriendo:
```bash
ollama serve
```

### "Model not found"

Descarga el modelo primero:
```bash
ollama pull llama3.2
```

### Respuestas lentas

- Usa modelos más pequeños (1B-3B parámetros)
- Usa GPU si está disponible
- Aumenta la memoria swap si es necesario

## Referencias

- Ollama: https://ollama.com
- Modelos disponibles: https://ollama.com/library
- Issue original: https://github.com/sipeed/picoclaw/issues/75

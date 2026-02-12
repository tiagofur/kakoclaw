# Solución de Problemas Comunes (FAQ)

Respuestas a las preguntas más frecuentes sobre PicoClaw.

## General

### ¿PicoClaw es gratis?

Sí, PicoClaw es software open source bajo licencia MIT. Sin embargo, necesitarás:
- **API keys de LLM**: Algunos providers tienen tiers gratuitos (OpenRouter, Groq)
- **Servidor**: Para correr el gateway (puede ser tu propia computadora)

### ¿Cuánto cuesta usar PicoClaw?

El costo depende del proveedor de LLM:
- **OpenRouter**: 200K tokens/mes gratis
- **Groq**: Tier gratuito disponible
- **Zhipu**: 200K tokens/mes gratis
- **Anthropic/OpenAI**: Solo pago

### ¿Funciona sin internet?

Parcialmente. PicoClaw puede funcionar con LLMs locales (vLLM, Ollama), pero algunas features como web search requieren conexión.

### ¿Qué tan seguro es?

- No almacenamos tus datos en servidores externos
- API keys se guardan localmente en tu máquina
- Puedes restringir operaciones al workspace
- Código open source auditable

## Instalación

### ¿Qué versión de Go necesito?

Go 1.21 o superior.

```bash
go version
# go version go1.21.0 linux/amd64
```

### ¿Cómo actualizo PicoClaw?

```bash
# Si compilaste desde fuente
cd picoclaw
git pull origin main
make build
make install

# Si usaste binario pre-compilado
# Descarga la última versión del release
```

### "command not found" después de instalar

```bash
# Verificar que está en PATH
which picoclaw

# Si no, agregar a PATH
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

## Configuración

### ¿Dónde está el archivo de configuración?

```
~/.picoclaw/config.json
```

Puedes ver la ubicación exacta con:
```bash
picoclaw status
```

### ¿Cómo cambio el modelo de LLM?

Edita `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-3.5-sonnet"
    }
  }
}
```

Modelos populares:
- `anthropic/claude-3.5-sonnet` (recomendado)
- `openai/gpt-4`
- `meta-llama/llama-3.1-70b`
- `google/gemini-pro`

### ¿Puedo usar múltiples proveedores?

Sí:

```json
{
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx"
    },
    "groq": {
      "api_key": "gsk-yyy"
    }
  },
  "agents": {
    "defaults": {
      "provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet"
    }
  }
}
```

### ¿Cómo configuro variables de entorno?

Todas las opciones de config.json pueden ser variables de entorno:

```bash
export PICOCLAW_AGENTS_DEFAULTS_MODEL="gpt-4"
export PICOCLAW_PROVIDERS_OPENROUTER_API_KEY="sk-or-v1-xxx"
export PICOCLAW_CHANNELS_TELEGRAM_TOKEN="123456:ABC..."
```

## Uso

### ¿Cómo limpio el historial de conversaciones?

```bash
# Eliminar todas las sesiones
rm -rf ~/.picoclaw/workspace/sessions/*

# O eliminar sesión específica
rm ~/.picoclaw/workspace/sessions/telegram:123456.json
```

### ¿Cómo uso sesiones diferentes?

```bash
# Sesión de trabajo
picoclaw agent -s trabajo

# Sesión personal
picoclaw agent -s personal

# Cada sesión tiene su propio historial
```

### ¿Puedo usar PicoClaw en scripts?

Sí:

```bash
#!/bin/bash

RESULT=$(picoclaw agent -m "Genera un nombre para este archivo: $1")
echo "$RESULT"
```

### ¿Cómo obtengo más información de debug?

```bash
# Modo debug
picoclaw agent --debug -m "test"

# O variable de entorno
PICOCLAW_DEBUG=1 picoclaw agent
```

## Proveedores LLM

### "No API key configured"

Configura al menos un proveedor en `~/.picoclaw/config.json`:

```json
{
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-TU-API-KEY"
    }
  }
}
```

### "API key invalid"

Verifica:
1. Que la API key es correcta
2. Que no ha expirado
3. Que tienes saldo/créditos disponibles
4. Que estás usando el formato correcto

### ¿Cuál proveedor recomiendas?

Para empezar:
1. **Groq** - Rápido y buen tier gratuito
2. **OpenRouter** - Acceso a múltiples modelos
3. **Zhipu** - Si estás en China

### Los modelos son lentos

Prueba:
- **Groq**: Optimizado para velocidad
- **Claude 3.5 Sonnet**: Balance velocidad/calidad
- Reducir `max_tokens` en la configuración

## Canales

### "Telegram bot not responding"

Verifica:
1. Token correcto en config.json
2. User ID en `allow_from`
3. Solo una instancia corriendo el bot
4. Bot tiene permisos necesarios

```bash
# Verificar
picoclaw status
# Debe mostrar Telegram API: ✓
```

### "Discord connection failed"

Verifica:
1. Token correcto
2. Intents habilitados en Discord Developer Portal
3. Bot tiene permisos en el servidor

### "Conflict: terminated by other getUpdates"

Tienes dos instancias corriendo el mismo bot de Telegram:

```bash
# Matar procesos anteriores
pkill -f "picoclaw gateway"

# Luego iniciar de nuevo
picoclaw gateway
```

### WhatsApp no funciona

WhatsApp requiere un bridge externo:
1. Instalar [whatsmeow](https://github.com/tulir/whatsmeow) o similar
2. Configurar `bridge_url` en config.json
3. Iniciar el bridge antes que PicoClaw

## Tools

### "web_search says API 配置问题"

La búsqueda web necesita API key de Brave:

1. Ve a https://brave.com/search/api
2. Regístrate (2000 consultas/mes gratis)
3. Agrega a config.json:

```json
{
  "tools": {
    "web": {
      "search": {
        "api_key": "BSA-TU-API-KEY",
        "max_results": 5
      }
    }
  }
}
```

### "tool not found"

El agente no reconoció el comando. Sé más específico:

```bash
# ❌ Vago
picoclaw agent -m "lee config"

# ✅ Específico
picoclaw agent -m "lee el archivo config.json usando read_file"
```

### "Error: path outside workspace"

El tool está restringido al workspace por seguridad:

```bash
# ❌ Fuera del workspace
picoclaw agent -m "lee /etc/passwd"

# ✅ Dentro del workspace
picoclaw agent -m "lee config.json"
```

Para deshabilitar restricción (no recomendado):

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

### Operaciones de archivos fallan

Verifica:
1. Permisos de lectura/escritura
2. Archivo existe
3. Directorio existe
4. Espacio en disco disponible

```bash
# Verificar permisos
ls -la ~/.picoclaw/workspace/

# Ver espacio
df -h
```

## Cron y Tareas

### Las tareas programadas no se ejecutan

Verifica:
1. Gateway está corriendo (`picoclaw gateway`)
2. Job está habilitado: `picoclaw cron list`
3. Formato de expresión cron es correcto
4. Hora del sistema es correcta: `date`

### "Failed to parse cron expression"

Formato correcto:
```bash
# Cada hora
picoclaw cron add -n "hourly" -m "test" -c "0 * * * *"

# Todos los días a las 9am
picoclaw cron add -n "daily" -m "test" -c "0 9 * * *"

# Cada 5 minutos
picoclaw cron add -n "frequent" -m "test" -e 300
```

## Skills

### "Skill not found"

```bash
# Verificar skills instalados
picoclaw skills list

# Instalar skill
picoclaw skills install sipeed/picoclaw-skills/weather

# Verificar archivo existe
ls ~/.picoclaw/workspace/skills/
```

### Los skills no se cargan

Verifica:
1. Estructura correcta: `skills/nombre-skill/SKILL.md`
2. Frontmatter YAML válido
3. Skill no está corrupto

### Crear skill personalizado

```bash
mkdir -p ~/.picoclaw/workspace/skills/mi-skill

cat > ~/.picoclaw/workspace/skills/mi-skill/SKILL.md << 'EOF'
---
name: mi-skill
description: Mi skill personalizado
---

# Mi Skill

Instrucciones para el agente...
EOF
```

## Performance

### PicoClaw usa mucha memoria

Optimizaciones:

```json
{
  "agents": {
    "defaults": {
      "max_tokens": 2048,
      "max_tool_iterations": 10
    }
  }
}
```

### Respuestas lentas

- Usa Groq como provider (más rápido)
- Reduce `max_tokens`
- Usa modelo más ligero
- Verifica conexión a internet

### Gateway consume mucho CPU

```bash
# Verificar logs
picoclaw gateway --debug 2>&1 | head -100

# Reducir frecuencia de polling en canales
```

## Errores Comunes

### "context deadline exceeded"

Timeout en operación. Soluciones:
- Aumentar timeout en config
- Verificar conexión a internet
- Reducir complejidad de la tarea

### "rate limit exceeded"

Has excedido el límite del proveedor LLM:
- Espera un momento
- Cambia a otro provider
- Actualiza a plan de pago

### "tool execution failed"

Revisa los logs:
```bash
picoclaw agent --debug -m "comando" 2>&1
```

### "failed to create provider"

Verifica:
1. API key configurada
2. Provider soportado
3. Configuración válida

## Debugging

### Habilitar logs detallados

```bash
# Debug completo
PICOCLAW_DEBUG=1 picoclaw agent --debug

# Logs a archivo
picoclaw gateway --debug 2>&1 | tee picoclaw.log
```

### Ver estado del sistema

```bash
picoclaw status
```

### Inspeccionar workspace

```bash
# Estructura
tree ~/.picoclaw/

# Config
cat ~/.picoclaw/config.json | jq .

# Sesiones
ls -la ~/.picoclaw/workspace/sessions/

# Logs (si existen)
tail -f ~/.picoclaw/workspace/picoclaw.log
```

### Test de componentes individuales

```bash
# Test de provider
curl -H "Authorization: Bearer TU-API-KEY" \
  https://openrouter.ai/api/v1/models

# Test de tool
picoclaw agent -m "ejecuta echo test"

# Test de canal (si aplica)
picoclaw agent -m "envía mensaje de prueba"
```

## Contribución

### ¿Cómo reporto un bug?

1. Busca en [issues existentes](https://github.com/sipeed/picoclaw/issues)
2. Si no existe, crea uno nuevo con:
   - Descripción clara
   - Pasos para reproducir
   - Entorno (OS, versión, etc.)
   - Logs relevantes

### ¿Cómo sugiero una feature?

Abre un issue con label `enhancement` describiendo:
- El problema que resuelve
- Cómo debería funcionar
- Casos de uso

### ¿Puedo contribuir código?

¡Sí! Ver [Guía de Contribución](../development/contributing.md).

## Recursos Adicionales

- [Documentación](../README.md)
- [GitHub Issues](https://github.com/sipeed/picoclaw/issues)
- [Discord](https://discord.gg/V4sAZ9XWpN)
- [Releases](https://github.com/sipeed/picoclaw/releases)

## Aún tienes problemas?

1. Revisa los [issues existentes](https://github.com/sipeed/picoclaw/issues)
2. Únete a nuestro [Discord](https://discord.gg/V4sAZ9XWpN)
3. Crea un nuevo issue con:
   - Título descriptivo
   - Descripción completa
   - Pasos para reproducir
   - Logs de error
   - Tu configuración (sin API keys)

---

¿No encuentras tu respuesta? Revisa la [documentación completa](../README.md) o pregunta en la comunidad.

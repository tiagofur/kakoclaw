# Configuración de Proveedor y Modelo por Usuario

## Resumen

<function_calls>
<invoke name="create_file">
<parameter name="content"># Configuración de Proveedor y Modelo por Usuario

## Resumen

✅ **Cada usuario puede configurar su propio LLM provider y modelo predeterminado**  
✅ **Las configuraciones son completamente independientes entre usuarios**  
✅ **Los API keys están aislados por usuario en la base de datos**  
✅ **La configuración global sirve solo como plantilla inicial**

## Arquitectura

```
Usuario Alice                      Usuario Bob
├── API Keys (DB)                  ├── API Keys (DB)
│   └── openai: sk-alice...         │   └── anthropic: sk-bob...
└── Config Override (archivo)      └── Config Override (archivo)
    └── agents:                        └── agents:
        defaults:                          defaults:
          provider: openai                   provider: anthropic
          model: gpt-4                       model: claude-sonnet-4
```

## Cómo Funciona

### 1. Configuración Global (Template)
La configuración en `~/.MakoClaw/config.json` define valores predeterminados:

```json
{
  "agents": {
    "defaults": {
      "provider": "zhipu",
      "model": "glm-4.7",
      "max_tokens": 8192,
      "temperature": 0.7
    }
  }
}
```

**Importante**: Esta es solo una plantilla. Cada usuario puede sobrescribirla completamente.

### 2. API Keys por Usuario (Base de Datos)
Tabla `user_providers_config` almacena las credenciales de cada usuario:

```sql
user_id | config (JSON)
--------+-------------------------------------------
123     | {"openai": {"api_key": "sk-alice..."}}
456     | {"anthropic": {"api_key": "sk-bob..."}}
```

### 3. Configuración por Usuario (Archivos)
Cada usuario tiene su archivo de overrides en `~/.makoclaw/users/<uuid>/config.json`:

```json
{
  "agents": {
    "defaults": {
      "provider": "openai",
      "model": "gpt-4"
    }
  }
}
```

### 4. Merge en Runtime
Cuando un agente se ejecuta para un usuario:

1. Cargar config global: `~/.MakoClaw/config.json`
2. Cargar config de usuario: `~/.makoclaw/users/<uuid>/config.json`
3. **Merge a nivel de sección**: Los valores del usuario sobreescriben los globales
4. Cargar API keys del usuario desde la DB
5. **Resultado**: Configuración personalizada para ese usuario

```go
// En pkg/config/config.go
func MergeConfigs(global, user *Config) *Config {
    // Si agentes.defaults está configurado en user config,
    // toda la sección reemplaza  
a la global
    if !isAgentsConfigEmpty(&user.Agents) {
        merged.Agents = user.Agents
    } else {
        merged.Agents = global.Agents
    }
    // ...
}
```

## API para Usuarios

### Ver Mi Configuración Actual
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:18880/api/v1/me/config
```

Respuesta (merged global + user overrides):
```json
{
  "config": {
    "agents": {
      "defaults": {
        "provider": "openai",
        "model": "gpt-4",
        "temperature": 0.7
      }
    }
  }
}
```

### Configurar Mi Provider/Model Predeterminado
```bash
curl -X POST http://localhost:18880/api/v1/me/config/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agents": {
      "defaults": {
        "provider": "anthropic",
        "model": "claude-sonnet-4"
      }
    }
  }'
```

### Configurar API Key para un Provider
```bash
curl -X PUT "http://localhost:18880/api/v1/me/providers/update?provider=anthropic" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "sk-ant-YOUR-KEY",
    "api_base": "https://api.anthropic.com"
  }'
```

### Ver Mis Providers Configurados
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:18880/api/v1/me/providers
```

Respuesta (API keys redactes):
```json
{
  "anthropic": {
    "api_key": "sk-a****-key",
    "api_base": "https://api.anthropic.com"
  },
  "openai": {
    "api_key": "",
    "api_base": ""
  }
}
```

## Interfaz Web (Settings)

### Tab "General" (Agents)
Los usuarios pueden seleccionar:
- **Default Provider**: Dropdown con todos los providers disponibles
- **Default Model**: Dropdown con todos los modelos disponibles
- **Temperature**, **Max Tokens**, **Max Iterations**

Al hacer clic en "Save Agent Settings":
- Se guarda en `~/.makoclaw/users/<uuid>/config.json`
- Solo afecta a ese usuario
- No modifica la configuración global

### Tab "Providers"
Los usuarios pueden:
- Ver lista de todos los providers (Anthropic, OpenAI, Groq, etc.)
- Configurar API Key, API Base, Proxy para cada uno
- Los API keys se guardan en la DB (`user_providers_config`)
- Cada usuario ve solo sus propios keys

## Flujo de Onboarding (Setup Wizard)

1. **Usuario se registra** → Recibe config global como template
2. **Usuario selecciona provider** → Elige entre OpenAI, Claude, etc.
3. **Usuario ingresa API key** → Se guarda en `user_providers_config` DB
4. **Usuario confirma modelo** → Se guarda en su config como default
5. **¡Listo!** → El agente usa su provider/model configurado

## Casos de Uso

### Empresa con Múltiples Usuarios

**Escenario**: 10 usuarios, cada uno con diferentes preferencias

- **Alice (Marketing)**: Usa OpenAI GPT-4 para contenido creativo
- **Bob (Dev)**: Usa Claude Sonnet 4.5 para programación
- **Carol (Data)**: Usa Groq Llama 3.2 para análisis rápido

**Ventajas**:
- ✅ Cada uno usa su propia API key (costos separados)
- ✅ Cada uno obtiene respuestas del modelo que prefiere
- ✅ No hay interferencia entre usuarios

### Desarrollo y Producción

**Escenario**: Mismo usuario, diferentes entornos

- **Usuario en Dev**: `provider: ollama, model: llama3:local`
- **Usuario en Prod**: `provider: openai, model: gpt-4`

**Cómo**: Simplemente cambiar la configuración en Settings antes de trabajar.

### Admin Cambia la Global

**Escenario**: Admin actualiza el provider global de "zhipu" a "openai"

**Efecto**:
- ✅ Los usuarios SIN override continúan usando sus configs
- ✅ Solo los nuevos usuarios ven "openai" por defecto
- ✅ Los usuarios existentes no se ven afectados

## Permisos

| Acción | Usuario Regular | Admin |
|--------|----------------|-------|
| Ver su propia config | ✅ | ✅ |
| Modificar su provider/model | ✅ | ✅ |
| Modificar config global | ❌ | ✅ |
| Ver API keys de otros usuarios | ❌ | ❌ |

## Seguridad

✅ ** Aislamiento de API Keys**: Cada usuario solo puede ver/modificar sus propios keys  
✅ **Redacción Automática**: Las API keys se muestran como `sk-a****-key` en GET  
✅ **Separación de Datos**: Provider configs en DB, agent configs en archivos  
✅ **Sin Contaminación**: Nuevos usuarios ven configs vacías, no keys globales  

## Testing

```bash
# Test de configuración independiente por usuario
/tmp/test_per_user_providers.sh

# Resultado esperado:
# ✅ Alice usa openai/gpt-4
# ✅ Bob usa anthropic/claude-sonnet-4
# ✅ Las configuraciones no interfieren
```

## Solución de Problemas

### "No veo cambios después de actualizar mi provider"
- Verifica que la actualización fue exitosa: `GET /api/v1/me/config`
- Recarga la página del frontend
- Verifica los logs: `tail -f /tmp/makoclaw.log`

### "Mi config usa el provider global en vez del mío"
- Confirma que tu override está guardado: `cat ~/.makoclaw/users/<uuid>/config.json`
- Verifica que tu config incluye `agents.defaults.provider`
- Si la sección `agents` está vacía, el merge usa la global

### "No puedo guardar mi API key"
- Verifica que el provider name es correcto (ej: "openai", not "OpenAI")
- Confirma que tienes un token JWT válido
- Revisa los logs del backend para errores de DB

## Endpoints Relacionados

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/api/v1/me/config` | GET | Ver config merged (global + user) |
| `/api/v1/me/config/update` | POST | Actualizar config de usuario |
| `/api/v1/me/providers` | GET | Ver providers configurados (keys redacted) |
| `/api/v1/me/providers/update?provider=<name>` | PUT | Actualizar API key de un provider |
| `/api/v1/config` | GET | Ver config global (todos) |
| `/api/v1/config` | POST | Modificar config global (**admin only**) |

## Archivos Relevantes

**Backend**:
- `pkg/web/handlers_user_config.go` - Handlers para config de usuario
- `pkg/config/config.go` - Lógica de merge
- `pkg/storage/user_providers.go` - CRUD de providers en DB
- `pkg/agent/loop.go` - Carga config merged para ejecutar agente

**Frontend**:
- `pkg/web/frontend/src/views/SettingsView.vue` - Vista principal de settings
- `pkg/web/frontend/src/components/Settings/AgentSettingsTab.vue` - Tab para provider/model
- `pkg/web/frontend/src/components/Settings/ProvidersSettingsTab.vue` - Tab para API keys
- `pkg/web/frontend/src/services/advancedService.js` - Llamadas a API

## Resumen Técnico

```
┌─────────────────────────────────────────────────────────┐
│  Cada Usuario Tiene Configuración Independiente         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  1. API Keys (DB)           → Aislados por user_id      │
│  2. Provider/Model Default  → En config.json por UUID   │
│  3. Runtime Merge           → Global + User Overrides   │
│  4. Sin Interferencia       → Cambios de otros no       │
│                                 afectan tu config         │
│                                                          │
│  ✅ Cada usuario elige su LLM                            │
│  ✅ Cada usuario usa su API key                          │
│  ✅ Configuraciones completamente aisladas               │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

**Versión**: 1.0  
**Fecha**: 2026-02-21  
**Estado**: ✅ Implementado y Probado

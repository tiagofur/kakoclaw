# Guía de Inicio Rápido

Bienvenido a KakoClaw. Esta guía te ayudará a configurar y ejecutar tu asistente de IA en menos de 5 minutos.

## ✅ Requisitos Previos

- **Sistema Operativo**: Linux, macOS, o Windows
- **Go**: Versión 1.21 o superior (solo para compilar desde fuente)
- **Hardware**: Cualquier computadora moderna (incluso Raspberry Pi o placas de $10)
- **Conexión a Internet**: Para comunicación con LLMs

## 🚀 Instalación

### Opción 1: Binario Pre-compilado (Recomendado)

```bash
# Descargar el binario para tu plataforma
# Linux x86_64
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-amd64

# Linux ARM64 (Raspberry Pi, etc)
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-arm64

# macOS
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-darwin-amd64

# Hacer ejecutable
chmod +x KakoClaw-linux-amd64

# Mover a tu PATH
sudo mv KakoClaw-linux-amd64 /usr/local/bin/KakoClaw
```

### Opción 2: Compilar desde Fuente

```bash
# Clonar el repositorio
git clone https://github.com/sipeed/KakoClaw.git
cd KakoClaw

# Compilar
make build

# Instalar
make install

# Verificar instalación
KakoClaw version
```

## ⚙️ Configuración Inicial

### Paso 1: Inicializar KakoClaw

```bash
KakoClaw onboard
```

Esto creará:
- `~/.KakoClaw/config.json` - Archivo de configuración
- `~/.KakoClaw/workspace/` - Directorio de trabajo
- Archivos base: `AGENTS.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`

### Paso 2: Obtener API Key

Elige un proveedor de LLM y obtén tu API key:

#### Opción A: OpenRouter (Recomendado - Múltiples modelos)
1. Ve a [openrouter.ai/keys](https://openrouter.ai/keys)
2. Crea una cuenta
3. Genera una API key
4. Tienes 200K tokens gratis por mes

#### Opción B: Zhipu (Para usuarios de China)
1. Ve a [bigmodel.cn](https://bigmodel.cn)
2. Crea cuenta y obtén API key
3. Tienes 200K tokens gratis por mes

#### Opción C: Anthropic (Claude)
1. Ve a [console.anthropic.com](https://console.anthropic.com)
2. Crea cuenta y obtén API key

#### Opción D: Groq (Rápido y gratis)
1. Ve a [console.groq.com](https://console.groq.com)
2. Crea cuenta y obtén API key
3. Incluye Whisper para transcripción de voz

### Paso 3: Configurar API Key

Edita `~/.KakoClaw/config.json`:

```bash
# Abrir con tu editor favorito
nano ~/.KakoClaw/config.json
```

Configuración básica:

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 8192,
      "temperature": 0.7
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-TU-API-KEY-AQUI"
    }
  }
}
```

## 💬 Primer Uso

### Modo Directo (Una sola pregunta)

```bash
KakoClaw agent -m "Hola, ¿qué puedes hacer?"
```

### Modo Interactivo (Chat continuo)

```bash
KakoClaw agent

🐸 Interactive mode (Ctrl+C to exit)

🐸 You: Hola

🐸 Hola! Soy KakoClaw, tu asistente de IA ultraligero. Puedo ayudarte con:
- Buscar información en la web
- Leer y escribir archivos
- Ejecutar comandos en tu sistema
- Programar tareas recurrentes
- Y mucho más...

¿En qué puedo ayudarte hoy?

🐸 You: 
```

## 🔍 Funciones Básicas

### 1. Búsqueda Web

```bash
# Necesitas configurar Brave Search API (opcional pero recomendado)
# Ve a https://brave.com/search/api - 2000 consultas/mes gratis

KakoClaw agent -m "Busca información sobre Go programming"
```

### 2. Operaciones con Archivos

```bash
# Crear un archivo
KakoClaw agent -m "Crea un archivo hello.txt con el contenido 'Hola Mundo'"

# Leer un archivo
KakoClaw agent -m "Lee el archivo hello.txt"

# Listar directorio
KakoClaw agent -m "Lista los archivos en el directorio actual"
```

### 3. Ejecución de Comandos

```bash
# Ejecutar comando shell
KakoClaw agent -m "Ejecuta el comando 'date'"

# Análisis de sistema
KakoClaw agent -m "Muestra el uso de disco con df -h"
```

### 4. Tareas Programadas

```bash
# Crear recordatorio
KakoClaw cron add -n "reunion" -m "Tienes una reunión en 10 minutos" -e 600

# Ver tareas programadas
KakoClaw cron list
```

## 🤖 Uso Avanzado

### Configurar Canales (Telegram Bot)

1. **Crear bot en Telegram:**
   - Busca @BotFather en Telegram
   - Envía `/newbot`
   - Sigue las instrucciones y copia el token

2. **Obtener tu User ID:**
   - Busca @userinfobot en Telegram
   - Copia tu ID numérico

3. **Configurar en KakoClaw:**

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
      "allow_from": ["123456789"]
    }
  }
}
```

4. **Iniciar Gateway:**

```bash
KakoClaw gateway

# Ahora puedes escribirle a tu bot en Telegram!
```

### Usar Skills

```bash
# Ver skills disponibles
KakoClaw skills list

# Instalar skill de clima
KakoClaw skills install sipeed/KakoClaw-skills/weather

# Usar el skill
KakoClaw agent -m "¿Cómo está el clima en Madrid?"
```

### Múltiples Sesiones

```bash
# Sesión de trabajo
KakoClaw agent -s trabajo

# Sesión personal
KakoClaw agent -s personal

# Cada sesión tiene su propio historial y contexto
```

## 📊 Ver Estado

```bash
# Ver configuración y estado
KakoClaw status

# Salida esperada:
🐸 KakoClaw Status

Config: /home/user/.KakoClaw/config.json ✓
Workspace: /home/user/.KakoClaw/workspace ✓
Model: anthropic/claude-3.5-sonnet
OpenRouter API: ✓
```

## 🐛 Solución de Problemas

### Error: "No API key configured"

**Solución:** Verifica que configuraste al menos un proveedor en `config.json`

### Error: "Tool not found"

**Solución:** Algunos comandos necesitan sintaxis específica. Intenta ser más explícito:
- En lugar de "lee config.json", di "lee el archivo config.json"

### Error de conexión con Telegram

**Solución:** Verifica que:
1. El token es correcto
2. Tu user ID está en `allow_from`
3. No hay otra instancia corriendo el mismo bot

### No funciona la búsqueda web

**Solución:** Configura Brave Search API:
1. Ve a https://brave.com/search/api
2. Obtén API key gratuita
3. Agrega a config.json:
```json
{
  "tools": {
    "web": {
      "search": {
        "api_key": "BSA...",
        "max_results": 5
      }
    }
  }
}
```

## 🎓 Siguientes Pasos

- 📖 Lee la [documentación completa](../README.md)
- 🛠️ Aprende a [crear tus propios skills](../development/creating-skills.md)
- 💻 Configura [múltiples canales](../guides/channels.md)
- ⚡ Optimiza tu [configuración de LLM](../guides/llm-providers.md)

## 💡 Tips

1. **Sé específico**: Cuanto más detallada sea tu pregunta, mejor será la respuesta
2. **Usa sesiones**: Separa contextos diferentes (trabajo, personal, proyectos)
3. **Experimenta**: Prueba diferentes modelos y temperaturas
4. **Revisa logs**: Usa `--debug` para ver qué está pasando detrás
5. **Mantén actualizado**: `git pull && make install` periódicamente

## 🆘 Ayuda

- **Documentación**: [docs/](../README.md)
- **Issues**: [GitHub Issues](https://github.com/sipeed/KakoClaw/issues)
- **Comunidad**: [Discord](https://discord.gg/V4sAZ9XWpN)

---

**¡Felicitaciones!** Ahora tienes KakoClaw funcionando. 🐸

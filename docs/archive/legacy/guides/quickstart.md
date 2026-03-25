# Guía de Inicio Rápido

Bienvenido a **MakoClaw** — La plataforma de agentes de IA de alto nivel. Esta guía te ayudará a configurar y ejecutar tu ecosistema de IA en menos de 5 minutos.

<div align="center">

**🦈 MakoClaw — The Apex AI Agent**

Ultrafast · 10MB RAM · $10 Hardware · Self-Bootstrapped

</div>

---

## ✅ Requisitos Previos

- **Sistema Operativo**: Linux, macOS, o Windows
- **Go**: Versión 1.21 o superior (solo para compilar desde fuente)
- **Hardware**: Cualquier computadora moderna (incluso Raspberry Pi o placas de $10)
- **Conexión a Internet**: Para comunicación con LLMs

---

## 🚀 Instalación

### Opción 1: Binario Pre-compilado (Recomendado)

```bash
# Descargar el binario para tu plataforma
# Linux x86_64
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-amd64

# Linux ARM64 (Raspberry Pi, etc)
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-arm64

# macOS
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-darwin-amd64

# Windows
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-windows-amd64.exe

# Hacer ejecutable (Linux/macOS)
chmod +x MakoClaw-linux-amd64

# Mover a tu PATH
sudo mv MakoClaw-linux-amd64 /usr/local/bin/MakoClaw
```

### Opción 2: Compilar desde Fuente

```bash
# Clonar el repositorio
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw

# Compilar
make build

# Instalar
make install

# Verificar instalación
MakoClaw version
```

---

## ⚙️ Configuración Inicial

### Paso 1: Inicializar MakoClaw

```bash
MakoClaw onboard
```

Esto creará:

- `~/.MakoClaw/config.json` — Archivo de configuración
- `~/.MakoClaw/workspace/` — Directorio de trabajo
- Archivos base: `AGENTS.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`

### Paso 2: Obtener API Key

Elige un proveedor de LLM y obtén tu API key:

#### Opción A: OpenRouter (Recomendado — Múltiples modelos)

1. Ve a [openrouter.ai/keys](https://openrouter.ai/keys)
2. Crea una cuenta
3. Genera una API key
4. Tienes 200K tokens gratis por mes

#### Opción B: Groq (Rápido y gratis)

1. Ve a [console.groq.com](https://console.groq.com)
2. Crea cuenta y obtén API key
3. Incluye Whisper para transcripción de voz

#### Opción C: Anthropic (Claude)

1. Ve a [console.anthropic.com](https://console.anthropic.com)
2. Crea cuenta y obtén API key

#### Opción D: OpenAI (GPT-4)

1. Ve a [platform.openai.com](https://platform.openai.com)
2. Crea cuenta y obtén API key

### Paso 3: Configurar API Key (Opcional)

**Tienes dos opciones:**

#### Opción A: Configuración vía Web UI (Recomendado)

Sáltate la configuración manual y ve directamente al "Panel Web" (más abajo). MakoClaw iniciará en **Modo Degradado** con el panel web disponible para configuración fácil a través del Setup Wizard.

#### Opción B: Configuración Manual

Edita `~/.MakoClaw/config.json`:

```bash
# Abrir con tu editor favorito
nano ~/.MakoClaw/config.json
```

Configuración básica:

```json
{
  "agents": {
    "defaults": {
      "provider": "openrouter",
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

### 💡 Modo Degradado (Sin API Keys)

MakoClaw puede iniciar **sin ninguna configuración de LLM** en **Modo Degradado**:

```bash
# Iniciar sin configurar API keys
MakoClaw web
```

**En Modo Degradado:**

- ✅ Panel web totalmente accesible
- ✅ Setup Wizard disponible para configuración fácil
- ✅ Todas las funciones estáticas funcionan (autenticación, configuración, etc.)
- ❌ Funciones de agente/AI deshabilitadas hasta configurar proveedor
- ❌ Tareas cron deshabilitadas

**Para habilitar todas las funciones:**

1. Accede al panel web en `http://localhost:18880`
2. Haz clic en "Configure Now" en el banner de advertencia
3. Sigue el Setup Wizard para configurar tu proveedor LLM
4. Reinicia MakoClaw para activar funciones del agente

📖 **Más información**: [Guía de Modo Degradado](degraded-mode.md)

---

## 💬 Primer Uso

### Modo Directo (Una sola pregunta)

> **Nota**: El modo agent requiere un proveedor LLM configurado. Si omitiste el paso 3, usa el Panel Web (más abajo) para configurar via el Setup Wizard.

```bash
MakoClaw agent -m "Hola, ¿qué puedes hacer?"
```

### Modo Interactivo (Chat continuo)

> **Nota**: Requiere proveedor LLM configurado.

```bash
MakoClaw agent

🦈 Interactive mode (Ctrl+C to exit)

🦈 You: Hola

🦈 Hola! Soy MakoClaw, tu plataforma de agentes de IA de alto nivel. Puedo ayudarte con:
- Búsqueda en la web
- Lectura y escritura de archivos
- Ejecución de comandos en tu sistema
- Programar tareas recurrentes
- Gestión de tareas con Kanban
- Creación de workflows visuales
- Y mucho más...

¿En qué puedo ayudarte hoy?

🦈 You:
```

### Panel Web (Interfaz gráfica completa)

> **Funciona sin API keys**: El panel web está disponible en Modo Degradado para configuración fácil.

```bash
# Iniciar servidor web
MakoClaw web

# O usar el gateway para canales también (requiere proveedor para funciones completas)
MakoClaw gateway

# Abrir http://localhost:18880 en tu navegador
```

El panel web incluye:

- 💬 Chat con historial (requiere proveedor LLM)
- 📋 Kanban Board para tareas
- 🔄 Visual Workflows (requiere proveedor para ejecución)
- 🤖 Multi-Agent System (requiere proveedor)
- 📁 Gestión de archivos
- 🧠 Base de conocimientos
- ⏰ Cron jobs (requiere proveedor)
- 📊 Métricas y reportes
- ⚙️ Setup Wizard (configuración sin editar JSON)

---

## 🔍 Funciones Básicas

### 1. Búsqueda Web

```bash
# Necesitas configurar Brave Search API (opcional pero recomendado)
# Ve a https://brave.com/search/api — 2000 consultas/mes gratis

MakoClaw agent -m "Busca información sobre Go programming"
```

### 2. Operaciones con Archivos

```bash
# Crear un archivo
MakoClaw agent -m "Crea un archivo hello.txt con el contenido 'Hola Mundo'"

# Leer un archivo
MakoClaw agent -m "Lee el archivo hello.txt"

# Listar directorio
MakoClaw agent -m "Lista los archivos en el directorio actual"

# Editar un archivo (asistido por IA)
MakoClaw agent -m "Edita el archivo config.json y cambia el modelo"
```

### 3. Ejecución de Comandos

```bash
# Ejecutar comando shell
MakoClaw agent -m "Ejecuta el comando 'date'"

# Análisis de sistema
MakoClaw agent -m "Muestra el uso de disco con df -h"

# Procesos en ejecución
MakoClaw agent -m "Lista los procesos que más CPU consumen"
```

### 4. Gestión de Tareas (Kanban)

```bash
# Crear una tarea
MakoClaw agent -m "Crea una tarea: 'Revisar PR del proyecto' con alta prioridad"

# Listar tareas
MakoClaw agent -m "Muestra todas las tareas en el Kanban"

# Actualizar estado
MakoClaw agent -m "Marca la tarea 'Revisar PR' como en progreso"
```

### 5. Tareas Programadas (Cron)

```bash
# Crear recordatorio (via panel web o comando)
# Panel: Cron → New Job
# O comando:
MakoClaw cron add -n "reunion" -m "Tienes una reunión en 10 minutos" -e 600

# Ver tareas programadas
MakoClaw cron list
```

### 6. Workflows (Automatización Multi-Paso)

Los **workflows** te permiten crear pipelines de automatización combinando prompts, herramientas y lógica condicional.

**Acceso:** Panel Web → http://localhost:18880 → Workflows

#### Tu Primer Workflow

1. **Iniciar servidor web:**

```bash
MakoClaw web
```

2. **Acceder al panel:**
   - Abre http://localhost:18880 en tu navegador
   - Inicia sesión con tus credenciales

3. **Crear workflow:**
   - Click en "Workflows" en el menú
   - Click "New Workflow"
   - Nombre: "Quick Research"
   - Descripción: "Busca y resume información"

4. **Agregar pasos:**

   **Paso 1 - Búsqueda Web:**
   - Tipo: Tool
   - Label: "Search web"
   - Tool Name: `web_search`
   - Args: `{"query": "Go programming best practices 2026"}`

   **Paso 2 - Resumen con IA:**
   - Tipo: Prompt
   - Label: "Summarize"
   - Message: `Resumir en 3 puntos: {{step.1.output}}`

5. **Guardar y ejecutar:**
   - Click "Save"
   - Click "Run"
   - Ver resultados en tiempo real

#### Ejemplos de Workflows Útiles

- **Code Review Bot**: Analiza código → Crea tarea si hay issues → Asigna al desarrollador
- **Test Analyzer**: Corre tests → Analiza fallas → Guarda reporte → Notifica equipo
- **Research Assistant**: Busca → Sintetiza → Guarda notas → Genera resumen
- **SEO Optimizer**: Lee artículo → Analiza keywords → Genera mejoras → Actualiza archivo
- **Deploy Bot**: Corre tests → Buiild → Deploy → Notifica resultado
- **Backup Bot**: Comprime archivos → Sube a cloud → Guarda log → Programa próximo

📚 **Más información:** [Guía completa de Workflows](../examples/workflows.md)

🎯 **Templates listos:** [workflow-templates.json](../examples/workflow-templates.json)

---

## 🤖 Uso Avanzado

### Configurar Canales (Telegram Bot)

1. **Crear bot en Telegram:**
   - Busca @BotFather en Telegram
   - Envía `/newbot`
   - Sigue las instrucciones y copia el token

2. **Obtener tu User ID:**
   - Busca @userinfobot en Telegram
   - Copia tu ID numérico

3. **Configurar en MakoClaw:**

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
MakoClaw gateway

# Ahora puedes escribirle a tu bot en Telegram!
```

### Otros Canales Disponibles

- **Discord**: Configura bot token en `config.json`
- **Slack**: Configura bot token en `config.json`
- **WhatsApp**: Requiere bridge (Go-WhatsApp)
- **Signal**: Configura en `config.json`
- **QQ**: Configura en `config.json`
- **DingTalk**: Configura en `config.json`
- **Feishu**: Configura en `config.json`
- **MaixCam**: Configura en `config.json`

Ver [Canales de Mensajería](./channels.md) para más detalles.

### Multi-Agent System

1. **Configurar Orchestrator:**

```json
{
  "agents": {
    "orchestrator": {
      "enabled": true,
      "provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet",
      "temperature": 0.7,
      "max_delegation_retries": 3
    }
  }
}
```

2. **Crear Specialist:**

- Panel: Agents → New Specialist
- Configura modelo, temperatura, y descripción
- El Orchestrator delegará tareas automáticamente

### Base de Conocimiento (RAG)

1. **Subir documentos:**

- Panel: Knowledge → Upload
- Soporta: PDF, TXT, MD, JSON, CSV, HTML, XML, YAML, LOG

2. **Buscar:**

- Panel: Knowledge → Search
- O usa el comando `query_knowledge`

```bash
MakoClaw agent -m "¿Qué dice mi documento sobre el proyecto?"
```

### Usar Skills

```bash
# Ver skills disponibles
MakoClaw skills list

# Instalar skill de clima
MakoClaw skills install sipeed/MakoClaw-skills/weather

# Usar el skill
MakoClaw agent -m "¿Cómo está el clima en Madrid?"
```

### Múltiples Sesiones

```bash
# Sesión de trabajo
MakoClaw agent -s trabajo

# Sesión personal
MakoClaw agent -s personal

# Cada sesión tiene su propio historial y contexto
```

---

## 📊 Ver Estado

```bash
# Ver configuración y estado
MakoClaw status

# Salida esperada:
🦈 MakoClaw Status

Config: /home/user/.MakoClaw/config.json ✓
Workspace: /home/user/.MakoClaw/workspace ✓
Model: anthropic/claude-3.5-sonnet
OpenRouter API: ✓
Agents: 1 active
Workflows: 3 defined
```

---

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

### Panel web no carga

**Solución:**

1. Verifica que el puerto 18880 no esté en uso
2. Usa `--port` para cambiar el puerto: `MakoClaw web --port 8080`
3. Verifica firewall

---

## 🎓 Siguientes Pasos

- 📖 Lee la [documentación completa](../README.md)
- 🛠️ Aprende a [crear tus propios skills](../development/creating-skills.md)
- 💻 Configura [múltiples canales](./channels.md)
- ⚡ Optimiza tu [configuración de LLM](./llm-providers.md)
- 🤖 Explora el [Multi-Agent System](../development/MULTI_AGENT_SETUP.md)
- 🔄 Crea [Workflows avanzados](../examples/workflows.md)

---

## 💡 Tips

1. **Sé específico**: Cuanto más detallada sea tu pregunta, mejor será la respuesta
2. **Usa sesiones**: Separa contextos diferentes (trabajo, personal, proyectos)
3. **Experimenta**: Prueba diferentes modelos y temperaturas
4. **Revisa logs**: Usa `--debug` para ver qué está pasando detrás
5. **Mantén actualizado**: `git pull && make install` periódicamente
6. **Usa workflows**: Automatiza tareas repetitivas
7. **Leverage Kanban**: Deja que la IA gestione tus tareas
8. **Crea especialistas**: Agentes especializados para dominios específicos

---

## 🆘 Ayuda

- **Documentación**: [docs/](../README.md)
- **Issues**: [GitHub Issues](https://github.com/sipeed/MakoClaw/issues)
- **Discord**: [Comunidad](https://discord.gg/V4sAZ9XWpN)
- **Discussions**: [GitHub Discussions](https://github.com/sipeed/MakoClaw/discussions)

---

<div align="center">

**¡Felicitaciones!** 🦈

Ahora tienes **MakoClaw** funcionando. El tiburón más rápido del océano de la IA.

</div>

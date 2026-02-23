# Ejemplos Prácticos

Colección de ejemplos prácticos de uso de MakoClaw.

## Índice

- [Ejemplos Básicos](#ejemplos-básicos)
- [Automatización](#automatización)
- [Integraciones](#integraciones)
- [Workflows Completos](#workflows-completos)

## Ejemplos Básicos

### 1. Hola Mundo

```bash
# Ejecutar comando simple
MakoClaw agent -m "¿Qué hora es?"

# Modo interactivo
MakoClaw agent
# Escribe: Hola, ¿cómo estás?
```

### 2. Operaciones de Archivos

```bash
# Crear archivo
MakoClaw agent -m "Crea un archivo notas.txt con ideas para el proyecto"

# Leer archivo
MakoClaw agent -m "Lee el archivo notas.txt"

# Editar archivo
MakoClaw agent -m "En el archivo notas.txt, cambia 'idea 1' por 'implementar API'"

# Listar directorio
MakoClaw agent -m "Lista todos los archivos Go en el directorio actual"
```

### 3. Búsqueda Web

```bash
# Buscar información
MakoClaw agent -m "Busca información sobre Rust vs Go performance"

# Obtener contenido de URL
MakoClaw agent -m "Obtén el contenido de https://golang.org/doc/effective_go"

# Resumen de noticias
MakoClaw agent -m "Busca noticias de tecnología de hoy y dame un resumen"
```

### 4. Ejecución de Comandos

```bash
# Análisis de sistema
MakoClaw agent -m "Muestra el uso de memoria con free -h"

# Procesos
MakoClaw agent -m "Lista los 10 procesos que más CPU usan"

# Red
MakoClaw agent -m "Muestra las interfaces de red con ip addr"
```

## Automatización

### 5. Tareas Programadas

```bash
# Recordatorio diario
MakoClaw cron add -n "daily-standup" -m "Es hora de la daily standup" -c "0 9 * * 1-5"

# Backup semanal
MakoClaw cron add -n "weekly-backup" -m "Realiza backup de /home/user/documents" -c "0 2 * * 0"

# Monitoreo cada hora
MakoClaw cron add -n "monitor-disk" -m "Verifica uso de disco y alerta si > 80%" -e 3600

# Ver tareas
MakoClaw cron list

# Desactivar temporalmente
MakoClaw cron disable daily-standup
```

### 6. Script de Inicio Automático

Crea `scripts/daily-tasks.sh`:

```bash
#!/bin/bash
# Script de tareas diarias automatizadas

# 1. Actualizar skills
MakoClaw skills install-builtin

# 2. Verificar configuración
MakoClaw status

# 3. Resumen del día
MakoClaw agent -m "Genera un resumen de las tareas programadas para hoy"
```

```bash
chmod +x scripts/daily-tasks.sh

# Agregar al crontab
crontab -e
# 0 8 * * * /home/user/scripts/daily-tasks.sh
```

### 7. Automatización de Proyectos

```bash
# Inicializar proyecto
MakoClaw agent -m "Crea la estructura de directorios para un proyecto Go: cmd/, pkg/, internal/, docs/"

# Generar código
MakoClaw agent -m "Genera un archivo main.go básico para una API REST"

# Configurar CI/CD
MakoClaw agent -m "Crea un archivo .github/workflows/ci.yml para Go"

# Documentar
MakoClaw agent -m "Crea un README.md con instrucciones de instalación"
```

## Integraciones

### 8. Telegram Bot Completo

**Configuración** (`config.json`):
```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

**Iniciar**:
```bash
# Terminal 1
MakoClaw gateway

# Ahora escribe a tu bot en Telegram
```

**Usos**:
```
# En Telegram
Usuario: Busca recetas de pasta
Bot: [resultados de búsqueda]

Usuario: Lee /tmp/log.txt
Bot: [contenido del archivo]

Usuario: Ejecuta uptime
Bot: 14:30:00 up 5 days, 2:15, 1 user, load average: 0.52, 0.58, 0.59
```

### 9. Discord Bot

**Configuración**:
```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "YOUR_DISCORD_BOT_TOKEN",
      "allow_from": ["YOUR_DISCORD_USER_ID"]
    }
  }
}
```

**Uso**:
```bash
MakoClaw gateway

# En Discord
!claw buscar información sobre Docker
!claw crear archivo docker-compose.yml con nginx
```

### 10. Slack Integration

**Configuración**:
```json
{
  "channels": {
    "slack": {
      "enabled": true,
      "bot_token": "xoxb-YOUR-TOKEN",
      "app_token": "xapp-YOUR-TOKEN",
      "allow_from": ["U123456"]
    }
  }
}
```

**Uso en Slack**:
```
@MakoClaw analiza los logs en /var/log/app.log
@MakoClaw genera un reporte del sistema
```

## Workflows Completos

### 11. Workflow de Desarrollo

**Objetivo**: Asistente de desarrollo completo

```bash
# 1. Crear sesión de trabajo
MakoClaw agent -s proyecto-x

# 2. Entender codebase
MakoClaw agent -m "Lee el archivo README.md y main.go y explica qué hace este proyecto"

# 3. Analizar código
MakoClaw agent -m "Encuentra bugs potenciales en pkg/tools/shell.go"

# 4. Generar tests
MakoClaw agent -m "Genera tests unitarios para pkg/utils/string.go"

# 5. Documentar
MakoClaw agent -m "Crea documentación API para la función ParseJSON"

# 6. Commit message
MakoClaw agent -m "Genera un buen mensaje de commit para estos cambios: [pegar git diff]"
```

### 12. Workflow de Investigación

**Objetivo**: Investigación y resumen de información

```bash
# Sesión de investigación
MakoClaw agent -s research-ia

# 1. Buscar fuentes
MakoClaw agent -m "Busca 5 fuentes sobre modelos de lenguaje grandes (LLMs)"

# 2. Obtener contenido
MakoClaw agent -m "Obtén el contenido de estas URLs: [urls]"

# 3. Analizar y sintetizar
MakoClaw agent -m "Resume los puntos clave de estos artículos"

# 4. Crear documento
MakoClaw agent -m "Crea un documento research.md con la síntesis de la investigación"

# 5. Generar bibliografía
MakoClaw agent -m "Genera una lista de referencias en formato APA"
```

### 13. Workflow de Sistema

**Objetivo**: Administración y monitoreo de sistema

```bash
# Sesión de administrador
MakoClaw agent -s sysadmin

# 1. Health check
MakoClaw agent -m "Verifica el estado del sistema: disco, memoria, CPU, servicios"

# 2. Análisis de logs
MakoClaw agent -m "Analiza /var/log/syslog y encuentra errores de las últimas 24h"

# 3. Optimización
MakoClaw agent -m "Encuentra archivos grandes (>100MB) que se pueden eliminar"

# 4. Seguridad
MakoClaw agent -m "Verifica los puertos abiertos y conexiones activas"

# 5. Reporte
MakoClaw agent -m "Genera un reporte de salud del sistema en /tmp/system-report.md"
```

### 14. Workflow de Contenido

**Objetivo**: Creación y gestión de contenido

```bash
# Sesión de contenido
MakoClaw agent -s content

# 1. Brainstorming
MakoClaw agent -m "Genera 10 ideas de artículos sobre programación en Go"

# 2. Outline
MakoClaw agent -m "Crea un outline para 'Introducción a Goroutines'"

# 3. Escritura
MakoClaw agent -m "Escribe el artículo completo basado en el outline"

# 4. Revisión
MakoClaw agent -m "Revisa este artículo y sugiere mejoras"

# 5. Publicación
MakoClaw agent -m "Convierte el artículo a formato Markdown con frontmatter"
```

### 15. Workflow de Aprendizaje

**Objetivo**: Aprender nuevos temas

```bash
# Sesión de aprendizaje
MakoClaw agent -s learning-rust

# 1. Introducción
MakoClaw agent -m "Dame una introducción a Rust para programadores de Go"

# 2. Comparación
MakoClaw agent -m "Compara el sistema de ownership de Rust con el garbage collector de Go"

# 3. Ejemplos
MakoClaw agent -m "Muestra 5 ejemplos de código Rust equivalentes a Go"

# 4. Ejercicios
MakoClaw agent -m "Crea 3 ejercicios prácticos para practicar Rust"

# 5. Recursos
MakoClaw agent -m "Busca los mejores recursos para aprender Rust"

# 6. Guardar progreso
MakoClaw agent -m "Crea un archivo learning-rust.md con notas y recursos"
```

## Scripts Avanzados

### 16. Script de Backup Inteligente

`scripts/smart-backup.sh`:

```bash
#!/bin/bash

WORKSPACE="/home/user/documents"
BACKUP_DIR="/backup/$(date +%Y%m%d)"

# Crear directorio de backup
mkdir -p "$BACKUP_DIR"

# Usar MakoClaw para decidir qué respaldar
MakoClaw agent -m "
Analiza $WORKSPACE y crea un script que:
1. Encuentre archivos modificados en la última semana
2. Excluya archivos temporales (*.tmp, .cache)
3. Copie los archivos a $BACKUP_DIR
4. Genere un resumen del backup
" > /tmp/backup-script.sh

# Ejecutar script generado
bash /tmp/backup-script.sh

# Notificar
MakoClaw agent -m "Backup completado en $BACKUP_DIR" 2>/dev/null || true
```

### 17. Script de Generación de Proyecto

`scripts/create-project.sh`:

```bash
#!/bin/bash

PROJECT_NAME=$1
PROJECT_TYPE=$2  # web, cli, api, lib

if [ -z "$PROJECT_NAME" ]; then
    echo "Uso: $0 <nombre-proyecto> [tipo]"
    exit 1
fi

mkdir -p "$PROJECT_NAME"
cd "$PROJECT_NAME"

# Generar estructura con MakoClaw
MakoClaw agent -m "
Genera una estructura de proyecto Go de tipo $PROJECT_TYPE llamado $PROJECT_NAME.
Debe incluir:
1. go.mod con el módulo github.com/user/$PROJECT_NAME
2. Estructura de directorios estándar
3. main.go con código inicial
4. README.md con instrucciones
5. .gitignore para Go
6. Makefile con targets básicos
" > /tmp/project-structure.txt

# Extraer y ejecutar comandos del output
cat /tmp/project-structure.txt | grep "^mkdir\|^cat\|^echo\|^touch" | bash

echo "Proyecto $PROJECT_NAME creado!"
```

### 18. Script de CI/CD Helper

`scripts/ci-helper.sh`:

```bash
#!/bin/bash

# Analizar resultado de tests
if [ -f "test-output.txt" ]; then
    MakoClaw agent -m "
    Analiza este output de tests y:
    1. Resume los fallos si los hay
    2. Identifica tests flaky
    3. Sugiere fixes
    
    $(cat test-output.txt)
    " > test-analysis.md
fi

# Analizar coverage
if [ -f "coverage.out" ]; then
    MakoClaw agent -m "
    Analiza el coverage report y:
    1. Identifica paquetes con baja cobertura (<70%)
    2. Sugiere qué funciones necesitan tests
    3. Genera tests para las funciones críticas sin coverage
    " > coverage-analysis.md
fi
```

## Tips y Trucos

### 19. Sesiones Persistentes

```bash
# Crear alias útiles en ~/.bashrc
alias pc='MakoClaw agent'
alias pc-work='MakoClaw agent -s work'
alias pc-personal='MakoClaw agent -s personal'
alias pc-debug='MakoClaw agent --debug'

# Usar
pc-work -m "Revisa el código del PR #123"
pc-personal -m "Organiza mis tareas del fin de semana"
```

### 20. Combinar con otros comandos

```bash
# Pipe de output
cat error.log | MakoClaw agent -m "Analiza estos errores"

# Procesar output
MakoClaw agent -m "Resume el output" < large-file.txt

# Usar en scripts
STATUS=$(MakoClaw agent -m "Verifica si el servicio nginx está corriendo" 2>&1)
```

### 21. Templates de Prompts

Guarda prompts reutilizables:

```bash
# En ~/.MakoClaw/prompts/
cat > ~/.MakoClaw/prompts/code-review.txt << 'EOF'
Realiza una revisión de código de los siguientes archivos:
- Busca bugs potenciales
- Identifica problemas de estilo
- Sugiere mejoras de performance
- Verifica manejo de errores
EOF

# Usar
MakoClaw agent -m "$(cat ~/.MakoClaw/prompts/code-review.txt)" -f archivo.go
```

---

Para más ejemplos avanzados, consulta la [documentación de Skills](../guides/skills.md).

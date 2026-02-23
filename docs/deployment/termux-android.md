# MakoClaw en Termux (Android)

Guía completa para instalar y ejecutar MakoClaw en Android usando Termux.

## ✅ Compatibilidad

- **Android**: 7.0+ (API 24+)
- **Arquitectura**: ARM64 (ARMv8), ARMv7
- **RAM**: Mínimo 2GB recomendado
- **Almacenamiento**: 100MB libres

## 📱 Instalación de Termux

### 1. Descargar Termux

**Opción A: F-Droid (Recomendado)**
- Descargar desde: https://f-droid.org/packages/com.termux/
- Versión actualizada y estable

**Opción B: GitHub Releases**
- https://github.com/termux/termux-app/releases

⚠️ **NO usar Google Play Store** - La versión está desactualizada

### 2. Configurar Termux

```bash
# Actualizar paquetes
pkg update && pkg upgrade -y

# Instalar dependencias necesarias
pkg install -y git golang make

# Opcional: Instalar herramientas útiles
pkg install -y nano vim curl wget
```

## 🚀 Instalación de MakoClaw

### Opción 1: Script Automático (Recomendado)

```bash
# Descargar e instalar
curl -fsSL https://raw.githubusercontent.com/sipeed/MakoClaw/main/scripts/install-termux.sh | bash

# O usando wget
wget -qO- https://raw.githubusercontent.com/sipeed/MakoClaw/main/scripts/install-termux.sh | bash
```

### Opción 2: Instalación Manual

```bash
# 1. Clonar repositorio
cd ~
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw

# 2. Compilar
make build

# 3. Instalar
make install

# 4. Verificar
MakoClaw version
```

## ⚙️ Configuración

### 1. Inicializar MakoClaw

```bash
MakoClaw onboard
```

### 2. Configurar API Key

Editar configuración:

```bash
nano ~/.MakoClaw/config.json
```

Configuración básica:

```json
{
  "agents": {
    "defaults": {
      "model": "ollama/llama3.2",
      "max_tokens": 2048
    }
  }
}
```

**Para usar con Ollama (recomendado en Android):**

```bash
# Instalar Ollama en Termux (requiere proot-distro)
pkg install proot-distro
proot-distro install alpine
proot-distro login alpine

# Dentro de Alpine
apk add ollama
ollama serve &
ollama pull llama3.2
```

## 🔧 Configuraciones Especiales para Android

### 1. Permisos de Almacenamiento

Para acceder a archivos del dispositivo:

```bash
# Dar permiso de almacenamiento a Termux
termux-setup-storage

# Ahora puedes acceder a:
# ~/storage/shared/      → Almacenamiento interno
# ~/storage/downloads/   → Descargas
# ~/storage/documents/   → Documentos
```

### 2. Ejecutar en Segundo Plano

```bash
# Instalar termux-services
pkg install termux-services

# Crear script de inicio
mkdir -p ~/.config/MakoClaw
cat > ~/.config/MakoClaw/start.sh << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
export PATH="$HOME/.local/bin:$PATH"
source ~/.bashrc
MakoClaw gateway > ~/MakoClaw.log 2>&1 &
echo "MakoClaw iniciado"
EOF

chmod +x ~/.config/MakoClaw/start.sh

# Ejecutar
~/.config/MakoClaw/start.sh
```

### 3. Widget de Inicio Rápido (Opcional)

```bash
# Instalar Termux:Widget desde F-Droid
# Crear atajo
mkdir -p ~/.shortcuts
cat > ~/.shortcuts/MakoClaw << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
termux-notification --title "MakoClaw" --content "Iniciando..."
MakoClaw agent -m "$1" 2>&1 | termux-notification --title "MakoClaw Respuesta" --content "-"
EOF
chmod +x ~/.shortcuts/MakoClaw
```

## 📲 Canales Recomendados para Android

### Telegram Bot (Más fácil)

1. Crear bot con @BotFather
2. Configurar en `~/.MakoClaw/config.json`:

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

3. Ejecutar:

```bash
MakoClaw gateway
```

### Ollama Local (Sin internet)

Ideal para usar completamente offline:

```bash
# En proot-distro (Alpine o Ubuntu)
ollama serve &
ollama pull llama3.2

# Configurar MakoClaw
# En ~/.MakoClaw/config.json:
{
  "agents": {
    "defaults": {
      "model": "llama3.2"
    }
  },
  "providers": {
    "ollama": {
      "api_base": "http://localhost:11434"
    }
  }
}
```

## 🎮 Uso Práctico

### Modo Interactivo

```bash
MakoClaw agent

# Dentro del chat:
# Hola, ¿qué puedes hacer?
# Ayúdame a organizar mis tareas
# Busca información sobre Go
```

### Comandos Útiles

```bash
# Ver estado
MakoClaw status

# Verificar configuración
MakoClaw doctor

# Usar sesión específica
MakoClaw agent -s android-session

# Ejecutar comando directo
MakoClaw agent -m "Lista archivos en Downloads"
```

### Scripts de Automatización

```bash
# Script de backup diario
cat > ~/backup.sh << 'EOF'
#!/bin/bash
cd ~/storage/shared/Documents
MakoClaw agent -m "Genera un resumen de los archivos modificados hoy" > ~/backup-report.txt
EOF
chmod +x ~/backup.sh

# Ejecutar con cron (si está disponible)
# O manualmente
~/backup.sh
```

## 🔋 Optimización para Android

### 1. Reducir Consumo de Batería

```json
{
  "agents": {
    "defaults": {
      "max_tokens": 1024,
      "max_tool_iterations": 5
    }
  }
}
```

### 2. Uso de Memoria

```bash
# Limpiar sesiones antiguas regularmente
rm -rf ~/.MakoClaw/workspace/sessions/*.json

# O automáticamente con cron
# (si está instalado en proot-distro)
```

### 3. Almacenamiento

```bash
# Ver espacio usado

du -sh ~/.MakoClaw/

# Limpiar logs antiguos
rm -f ~/.MakoClaw/workspace/*.log
```

## 🐛 Troubleshooting

### "Permission denied"

```bash
# Verificar permisos
ls -la ~/.local/bin/MakoClaw

# Corregir
chmod +x ~/.local/bin/MakoClaw
```

### "cannot find package"

```bash
# Limpiar cache de Go
go clean -cache

# Reintentar
make build
```

### "Out of memory"

```bash
# Usar modelos más pequeños
# En config.json:
{
  "agents": {
    "defaults": {
      "model": "llama3.2:1b"
    }
  }
}
```

### Termux se cierra al ejecutar

```bash
# Ejecutar con nohup
nohup MakoClaw gateway > ~/MakoClaw.log 2>&1 &

# O usar tmux
pkg install tmux
tmux new -s MakoClaw
MakoClaw gateway
# Ctrl+B, D para desconectar
```

## 🎯 Casos de Uso en Android

### 1. Asistente Personal
- Organizar tareas diarias
- Recordatorios
- Notas rápidas

### 2. Desarrollo Móvil
- Revisar código
- Generar snippets
- Documentar proyectos

### 3. Productividad
- Resumir artículos
- Traducir textos
- Organizar archivos

### 4. Aprendizaje
- Explicar conceptos
- Practicar idiomas
- Resolver dudas

## 📚 Recursos Adicionales

- **Termux Wiki**: https://wiki.termux.com
- **MakoClaw Docs**: https://github.com/sipeed/MakoClaw/tree/main/docs
- **Ollama en Termux**: https://github.com/ollama/ollama

## 🤝 Soporte

Para problemas específicos de Termux:
- GitHub Issues: https://github.com/sipeed/MakoClaw/issues
- Discord: https://discord.gg/V4sAZ9XWpN
- Termux Reddit: r/termux

---

**¡Listo para usar MakoClaw en tu Android! 🦈📱**

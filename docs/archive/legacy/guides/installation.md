# Guía de Instalación y Configuración

Esta guía cubre la instalación detallada de MakoClaw en diferentes sistemas operativos y configuraciones.

## 📋 Tabla de Contenidos

- [Requisitos del Sistema](#requisitos-del-sistema)
- [Instalación en Linux](#instalación-en-linux)
- [Instalación en macOS](#instalación-en-macos)
- [Instalación en Windows](#instalación-en-windows)
- [Instalación en ARM/RISC-V](#instalación-en-armrisc-v)
- [Configuración Post-Instalación](#configuración-post-instalación)
- [Actualización](#actualización)
- [Desinstalación](#desinstalación)

## Requisitos del Sistema

### Mínimos
- **CPU**: 0.6GHz (cualquier procesador moderno)
- **RAM**: 50MB disponibles
- **Disco**: 20MB para el binario + espacio para workspace
- **SO**: Linux kernel 3.2+, macOS 10.14+, Windows 10+

### Recomendados
- **CPU**: 1GHz+ dual core
- **RAM**: 100MB disponibles
- **Disco**: 100MB+
- **Red**: Conexión estable a Internet

## Instalación en Linux

### Método 1: Script de Instalación Automática

```bash
curl -fsSL https://raw.githubusercontent.com/sipeed/MakoClaw/main/install.sh | bash
```

### Método 2: Instalación Manual

#### AMD64 (x86_64)

```bash
# Descargar
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-amd64 -O MakoClaw

# Hacer ejecutable
chmod +x MakoClaw

# Mover a PATH
sudo mv MakoClaw /usr/local/bin/

# Verificar
MakoClaw version
```

#### ARM64 (AArch64)

```bash
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-arm64 -O MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/
```

#### ARM (32-bit)

```bash
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-armv7 -O MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/
```

#### RISC-V

```bash
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-riscv64 -O MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/
```

### Método 3: Compilar desde Fuente

#### Dependencias

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y git golang-go make

# Fedora/RHEL
sudo dnf install -y git golang make

# Arch Linux
sudo pacman -S git go make
```

#### Compilación

```bash
# Clonar
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw

# Compilar
make build

# Instalar
make install

# O instalar en ubicación personalizada
make install INSTALL_PREFIX=$HOME/.local
```

### Configuración de Shell

Agrega a tu `~/.bashrc` o `~/.zshrc`:

```bash
# Si instalaste en ~/.local
export PATH="$HOME/.local/bin:$PATH"

# Autocompletado (opcional)
eval "$(MakoClaw completion bash)"
```

## Instalación en macOS

### Método 1: Homebrew (Próximamente)

```bash
# No disponible aún
# brew install MakoClaw
```

### Método 2: Binario Directo

#### Intel (AMD64)

```bash
curl -L https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-darwin-amd64 -o MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/
```

#### Apple Silicon (ARM64)

```bash
curl -L https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-darwin-arm64 -o MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/
```

### Método 3: Compilar desde Fuente

```bash
# Instalar dependencias con Homebrew
brew install go git

# Clonar y compilar
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw
make build
make install
```

### Configuración de macOS

Agregar a `~/.zshrc`:

```bash
# PATH si es necesario
export PATH="$HOME/.local/bin:$PATH"

# Autocompletado
eval "$(MakoClaw completion zsh)"
```

## Instalación en Windows

### Método 1: Scoop (Recomendado)

```powershell
# Instalar Scoop si no lo tienes
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Instalar MakoClaw
scoop bucket add MakoClaw https://github.com/sipeed/MakoClaw-bucket
scoop install MakoClaw
```

### Método 2: Descarga Directa

```powershell
# Descargar
Invoke-WebRequest -Uri "https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-windows-amd64.exe" -OutFile "MakoClaw.exe"

# Mover a un directorio en PATH
# Ejemplo: C:\Tools
Move-Item MakoClaw.exe C:\Tools\

# Agregar C:\Tools al PATH del sistema si no está
```

### Método 3: Compilar desde Fuente

```powershell
# Instalar Go desde https://golang.org/dl/

# Clonar
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw

# Compilar
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o MakoClaw.exe ./cmd/MakoClaw

# El binario está listo para usar
```

### Configuración de PowerShell

Agregar a tu perfil de PowerShell (`$PROFILE`):

```powershell
# Autocompletado
Invoke-Expression (&MakoClaw completion powershell)
```

## Instalación en ARM/RISC-V

### Raspberry Pi

```bash
# Descargar versión ARM64
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-arm64 -O MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/

# Inicializar
MakoClaw onboard
```

### LicheeRV Nano ($10)

```bash
# Descargar versión RISC-V
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-riscv64 -O MakoClaw
chmod +x MakoClaw

# Mover a PATH local
mkdir -p ~/.local/bin
mv MakoClaw ~/.local/bin/

# Agregar a PATH
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc

# Inicializar
MakoClaw onboard
```

### MaixCAM

```bash
# En MaixCAM (ARM64)
curl -L https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-arm64 -o /usr/local/bin/MakoClaw
chmod +x /usr/local/bin/MakoClaw

# Configurar canal MaixCAM en config.json
# Luego iniciar
MakoClaw gateway
```

## Configuración Post-Instalación

### Paso 1: Inicialización

```bash
MakoClaw onboard
```

Crea la estructura:
```
~/.MakoClaw/
├── config.json
├── workspace/
│   ├── sessions/
│   ├── memory/
│   ├── cron/
│   └── skills/
└── auth.json
```

### Paso 2: Configuración Básica

Edita `~/.MakoClaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.MakoClaw/workspace",
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-TU_API_KEY"
    }
  },
  "gateway": {
    "host": "0.0.0.0",
    "port": 18790
  }
}
```

### Paso 3: Verificar Instalación

```bash
# Ver versión
MakoClaw version

# Ver estado
MakoClaw status

# Prueba básica
MakoClaw agent -m "Hola, ¿funcionas?"
```

### Paso 4: Configurar Variables de Entorno (Opcional)

```bash
# Agregar a ~/.bashrc o ~/.zshrc

# Configuración por defecto
export MakoClaw_AGENTS_DEFAULTS_MODEL="anthropic/claude-3.5-sonnet"

# API Keys (alternativa a config.json)
export MakoClaw_PROVIDERS_OPENROUTER_API_KEY="sk-or-v1-xxx"

# Directorio workspace personalizado
export MakoClaw_AGENTS_DEFAULTS_WORKSPACE="~/proyectos/MakoClaw"
```

## Configuración Avanzada

### Configuración con Environment Variables

Todas las opciones de config.json pueden usarse como variables de entorno:

```bash
# Sintaxis: MakoClaw_<SECCION>_<OPCION>
export MakoClaw_AGENTS_DEFAULTS_MODEL="gpt-4"
export MakoClaw_AGENTS_DEFAULTS_MAX_TOKENS="8192"
export MakoClaw_CHANNELS_TELEGRAM_ENABLED="true"
export MakoClaw_CHANNELS_TELEGRAM_TOKEN="123456:ABC..."
```

### Configuración para Múltiples Entornos

```bash
# Desarrollo
MakoClaw agent --config ~/.MakoClaw/config.dev.json

# Producción
MakoClaw agent --config ~/.MakoClaw/config.prod.json
```

### Configuración de Proxy

```json
{
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "proxy": "http://proxy.company.com:8080"
    }
  }
}
```

## Actualización

### Método 1: Script de Actualización

```bash
# Descargar última versión
MakoClaw update

# O manualmente
curl -fsSL https://raw.githubusercontent.com/sipeed/MakoClaw/main/install.sh | bash -s -- --update
```

### Método 2: Actualización Manual

```bash
# Backup de configuración
cp ~/.MakoClaw/config.json ~/.MakoClaw/config.json.backup

# Descargar nueva versión
wget https://github.com/sipeed/MakoClaw/releases/latest/download/MakoClaw-linux-amd64 -O MakoClaw
chmod +x MakoClaw
sudo mv MakoClaw /usr/local/bin/

# Verificar
MakoClaw version

# Restaurar config si es necesario
# cp ~/.MakoClaw/config.json.backup ~/.MakoClaw/config.json
```

### Método 3: Desde Fuente

```bash
cd MakoClaw
git pull origin main
make build
make install
```

## Desinstalación

### Desinstalación Completa

```bash
# Eliminar binario
sudo rm /usr/local/bin/MakoClaw

# Eliminar datos
rm -rf ~/.MakoClaw

# Eliminar autocompletado de shell
# Editar ~/.bashrc o ~/.zshrc y quitar líneas de MakoClaw
```

### Desinstalación con Make

```bash
cd MakoClaw
make uninstall
make uninstall-all  # Incluye workspace y configuración
```

## Verificación de la Instalación

Ejecuta este checklist:

```bash
# 1. Verificar binario
which MakoClaw
MakoClaw version

# 2. Verificar configuración
ls -la ~/.MakoClaw/
cat ~/.MakoClaw/config.json

# 3. Verificar workspace
ls -la ~/.MakoClaw/workspace/

# 4. Prueba funcional
MakoClaw agent -m "Di 'MakoClaw está funcionando correctamente'"

# 5. Verificar permisos
touch ~/.MakoClaw/workspace/test.txt
rm ~/.MakoClaw/workspace/test.txt
```

## Solución de Problemas de Instalación

### "command not found"

```bash
# Verificar PATH
echo $PATH

# Si ~/.local/bin no está en PATH
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

### "permission denied"

```bash
# Corregir permisos
chmod +x /usr/local/bin/MakoClaw

# O si instalaste sin sudo
sudo chown $(whoami) /usr/local/bin/MakoClaw
```

### Error de GLIBC

En sistemas antiguos, compilar desde fuente:

```bash
# Estático linking
CGO_ENABLED=0 go build -ldflags="-s -w" -o MakoClaw ./cmd/MakoClaw
```

### Problemas de Memoria en Dispositivos Embebidos

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

---

Para configurar proveedores LLM específicos, ver [Configuración de Proveedores LLM](./llm-providers.md).

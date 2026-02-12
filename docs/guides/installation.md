# Guía de Instalación y Configuración

Esta guía cubre la instalación detallada de PicoClaw en diferentes sistemas operativos y configuraciones.

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
curl -fsSL https://raw.githubusercontent.com/sipeed/picoclaw/main/install.sh | bash
```

### Método 2: Instalación Manual

#### AMD64 (x86_64)

```bash
# Descargar
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-amd64 -O picoclaw

# Hacer ejecutable
chmod +x picoclaw

# Mover a PATH
sudo mv picoclaw /usr/local/bin/

# Verificar
picoclaw version
```

#### ARM64 (AArch64)

```bash
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-arm64 -O picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/
```

#### ARM (32-bit)

```bash
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-armv7 -O picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/
```

#### RISC-V

```bash
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-riscv64 -O picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/
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
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

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
eval "$(picoclaw completion bash)"
```

## Instalación en macOS

### Método 1: Homebrew (Próximamente)

```bash
# No disponible aún
# brew install picoclaw
```

### Método 2: Binario Directo

#### Intel (AMD64)

```bash
curl -L https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-darwin-amd64 -o picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/
```

#### Apple Silicon (ARM64)

```bash
curl -L https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-darwin-arm64 -o picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/
```

### Método 3: Compilar desde Fuente

```bash
# Instalar dependencias con Homebrew
brew install go git

# Clonar y compilar
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw
make build
make install
```

### Configuración de macOS

Agregar a `~/.zshrc`:

```bash
# PATH si es necesario
export PATH="$HOME/.local/bin:$PATH"

# Autocompletado
eval "$(picoclaw completion zsh)"
```

## Instalación en Windows

### Método 1: Scoop (Recomendado)

```powershell
# Instalar Scoop si no lo tienes
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Instalar PicoClaw
scoop bucket add picoclaw https://github.com/sipeed/picoclaw-bucket
scoop install picoclaw
```

### Método 2: Descarga Directa

```powershell
# Descargar
Invoke-WebRequest -Uri "https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-windows-amd64.exe" -OutFile "picoclaw.exe"

# Mover a un directorio en PATH
# Ejemplo: C:\Tools
Move-Item picoclaw.exe C:\Tools\

# Agregar C:\Tools al PATH del sistema si no está
```

### Método 3: Compilar desde Fuente

```powershell
# Instalar Go desde https://golang.org/dl/

# Clonar
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# Compilar
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o picoclaw.exe ./cmd/picoclaw

# El binario está listo para usar
```

### Configuración de PowerShell

Agregar a tu perfil de PowerShell (`$PROFILE`):

```powershell
# Autocompletado
Invoke-Expression (&picoclaw completion powershell)
```

## Instalación en ARM/RISC-V

### Raspberry Pi

```bash
# Descargar versión ARM64
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-arm64 -O picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/

# Inicializar
picoclaw onboard
```

### LicheeRV Nano ($10)

```bash
# Descargar versión RISC-V
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-riscv64 -O picoclaw
chmod +x picoclaw

# Mover a PATH local
mkdir -p ~/.local/bin
mv picoclaw ~/.local/bin/

# Agregar a PATH
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc

# Inicializar
picoclaw onboard
```

### MaixCAM

```bash
# En MaixCAM (ARM64)
curl -L https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-arm64 -o /usr/local/bin/picoclaw
chmod +x /usr/local/bin/picoclaw

# Configurar canal MaixCAM en config.json
# Luego iniciar
picoclaw gateway
```

## Configuración Post-Instalación

### Paso 1: Inicialización

```bash
picoclaw onboard
```

Crea la estructura:
```
~/.picoclaw/
├── config.json
├── workspace/
│   ├── sessions/
│   ├── memory/
│   ├── cron/
│   └── skills/
└── auth.json
```

### Paso 2: Configuración Básica

Edita `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
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
picoclaw version

# Ver estado
picoclaw status

# Prueba básica
picoclaw agent -m "Hola, ¿funcionas?"
```

### Paso 4: Configurar Variables de Entorno (Opcional)

```bash
# Agregar a ~/.bashrc o ~/.zshrc

# Configuración por defecto
export PICOCLAW_AGENTS_DEFAULTS_MODEL="anthropic/claude-3.5-sonnet"

# API Keys (alternativa a config.json)
export PICOCLAW_PROVIDERS_OPENROUTER_API_KEY="sk-or-v1-xxx"

# Directorio workspace personalizado
export PICOCLAW_AGENTS_DEFAULTS_WORKSPACE="~/proyectos/picoclaw"
```

## Configuración Avanzada

### Configuración con Environment Variables

Todas las opciones de config.json pueden usarse como variables de entorno:

```bash
# Sintaxis: PICOCLAW_<SECCION>_<OPCION>
export PICOCLAW_AGENTS_DEFAULTS_MODEL="gpt-4"
export PICOCLAW_AGENTS_DEFAULTS_MAX_TOKENS="8192"
export PICOCLAW_CHANNELS_TELEGRAM_ENABLED="true"
export PICOCLAW_CHANNELS_TELEGRAM_TOKEN="123456:ABC..."
```

### Configuración para Múltiples Entornos

```bash
# Desarrollo
picoclaw agent --config ~/.picoclaw/config.dev.json

# Producción
picoclaw agent --config ~/.picoclaw/config.prod.json
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
picoclaw update

# O manualmente
curl -fsSL https://raw.githubusercontent.com/sipeed/picoclaw/main/install.sh | bash -s -- --update
```

### Método 2: Actualización Manual

```bash
# Backup de configuración
cp ~/.picoclaw/config.json ~/.picoclaw/config.json.backup

# Descargar nueva versión
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-amd64 -O picoclaw
chmod +x picoclaw
sudo mv picoclaw /usr/local/bin/

# Verificar
picoclaw version

# Restaurar config si es necesario
# cp ~/.picoclaw/config.json.backup ~/.picoclaw/config.json
```

### Método 3: Desde Fuente

```bash
cd picoclaw
git pull origin main
make build
make install
```

## Desinstalación

### Desinstalación Completa

```bash
# Eliminar binario
sudo rm /usr/local/bin/picoclaw

# Eliminar datos
rm -rf ~/.picoclaw

# Eliminar autocompletado de shell
# Editar ~/.bashrc o ~/.zshrc y quitar líneas de picoclaw
```

### Desinstalación con Make

```bash
cd picoclaw
make uninstall
make uninstall-all  # Incluye workspace y configuración
```

## Verificación de la Instalación

Ejecuta este checklist:

```bash
# 1. Verificar binario
which picoclaw
picoclaw version

# 2. Verificar configuración
ls -la ~/.picoclaw/
cat ~/.picoclaw/config.json

# 3. Verificar workspace
ls -la ~/.picoclaw/workspace/

# 4. Prueba funcional
picoclaw agent -m "Di 'PicoClaw está funcionando correctamente'"

# 5. Verificar permisos
touch ~/.picoclaw/workspace/test.txt
rm ~/.picoclaw/workspace/test.txt
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
chmod +x /usr/local/bin/picoclaw

# O si instalaste sin sudo
sudo chown $(whoami) /usr/local/bin/picoclaw
```

### Error de GLIBC

En sistemas antiguos, compilar desde fuente:

```bash
# Estático linking
CGO_ENABLED=0 go build -ldflags="-s -w" -o picoclaw ./cmd/picoclaw
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

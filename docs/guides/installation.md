# Guía de Instalación y Configuración

Esta guía cubre la instalación detallada de KakoClaw en diferentes sistemas operativos y configuraciones.

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
curl -fsSL https://raw.githubusercontent.com/sipeed/KakoClaw/main/install.sh | bash
```

### Método 2: Instalación Manual

#### AMD64 (x86_64)

```bash
# Descargar
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-amd64 -O KakoClaw

# Hacer ejecutable
chmod +x KakoClaw

# Mover a PATH
sudo mv KakoClaw /usr/local/bin/

# Verificar
KakoClaw version
```

#### ARM64 (AArch64)

```bash
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-arm64 -O KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/
```

#### ARM (32-bit)

```bash
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-armv7 -O KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/
```

#### RISC-V

```bash
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-riscv64 -O KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/
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
git clone https://github.com/sipeed/KakoClaw.git
cd KakoClaw

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
eval "$(KakoClaw completion bash)"
```

## Instalación en macOS

### Método 1: Homebrew (Próximamente)

```bash
# No disponible aún
# brew install KakoClaw
```

### Método 2: Binario Directo

#### Intel (AMD64)

```bash
curl -L https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-darwin-amd64 -o KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/
```

#### Apple Silicon (ARM64)

```bash
curl -L https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-darwin-arm64 -o KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/
```

### Método 3: Compilar desde Fuente

```bash
# Instalar dependencias con Homebrew
brew install go git

# Clonar y compilar
git clone https://github.com/sipeed/KakoClaw.git
cd KakoClaw
make build
make install
```

### Configuración de macOS

Agregar a `~/.zshrc`:

```bash
# PATH si es necesario
export PATH="$HOME/.local/bin:$PATH"

# Autocompletado
eval "$(KakoClaw completion zsh)"
```

## Instalación en Windows

### Método 1: Scoop (Recomendado)

```powershell
# Instalar Scoop si no lo tienes
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Instalar KakoClaw
scoop bucket add KakoClaw https://github.com/sipeed/KakoClaw-bucket
scoop install KakoClaw
```

### Método 2: Descarga Directa

```powershell
# Descargar
Invoke-WebRequest -Uri "https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-windows-amd64.exe" -OutFile "KakoClaw.exe"

# Mover a un directorio en PATH
# Ejemplo: C:\Tools
Move-Item KakoClaw.exe C:\Tools\

# Agregar C:\Tools al PATH del sistema si no está
```

### Método 3: Compilar desde Fuente

```powershell
# Instalar Go desde https://golang.org/dl/

# Clonar
git clone https://github.com/sipeed/KakoClaw.git
cd KakoClaw

# Compilar
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o KakoClaw.exe ./cmd/KakoClaw

# El binario está listo para usar
```

### Configuración de PowerShell

Agregar a tu perfil de PowerShell (`$PROFILE`):

```powershell
# Autocompletado
Invoke-Expression (&KakoClaw completion powershell)
```

## Instalación en ARM/RISC-V

### Raspberry Pi

```bash
# Descargar versión ARM64
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-arm64 -O KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/

# Inicializar
KakoClaw onboard
```

### LicheeRV Nano ($10)

```bash
# Descargar versión RISC-V
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-riscv64 -O KakoClaw
chmod +x KakoClaw

# Mover a PATH local
mkdir -p ~/.local/bin
mv KakoClaw ~/.local/bin/

# Agregar a PATH
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc

# Inicializar
KakoClaw onboard
```

### MaixCAM

```bash
# En MaixCAM (ARM64)
curl -L https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-arm64 -o /usr/local/bin/KakoClaw
chmod +x /usr/local/bin/KakoClaw

# Configurar canal MaixCAM en config.json
# Luego iniciar
KakoClaw gateway
```

## Configuración Post-Instalación

### Paso 1: Inicialización

```bash
KakoClaw onboard
```

Crea la estructura:
```
~/.KakoClaw/
├── config.json
├── workspace/
│   ├── sessions/
│   ├── memory/
│   ├── cron/
│   └── skills/
└── auth.json
```

### Paso 2: Configuración Básica

Edita `~/.KakoClaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.KakoClaw/workspace",
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
KakoClaw version

# Ver estado
KakoClaw status

# Prueba básica
KakoClaw agent -m "Hola, ¿funcionas?"
```

### Paso 4: Configurar Variables de Entorno (Opcional)

```bash
# Agregar a ~/.bashrc o ~/.zshrc

# Configuración por defecto
export KakoClaw_AGENTS_DEFAULTS_MODEL="anthropic/claude-3.5-sonnet"

# API Keys (alternativa a config.json)
export KakoClaw_PROVIDERS_OPENROUTER_API_KEY="sk-or-v1-xxx"

# Directorio workspace personalizado
export KakoClaw_AGENTS_DEFAULTS_WORKSPACE="~/proyectos/KakoClaw"
```

## Configuración Avanzada

### Configuración con Environment Variables

Todas las opciones de config.json pueden usarse como variables de entorno:

```bash
# Sintaxis: KakoClaw_<SECCION>_<OPCION>
export KakoClaw_AGENTS_DEFAULTS_MODEL="gpt-4"
export KakoClaw_AGENTS_DEFAULTS_MAX_TOKENS="8192"
export KakoClaw_CHANNELS_TELEGRAM_ENABLED="true"
export KakoClaw_CHANNELS_TELEGRAM_TOKEN="123456:ABC..."
```

### Configuración para Múltiples Entornos

```bash
# Desarrollo
KakoClaw agent --config ~/.KakoClaw/config.dev.json

# Producción
KakoClaw agent --config ~/.KakoClaw/config.prod.json
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
KakoClaw update

# O manualmente
curl -fsSL https://raw.githubusercontent.com/sipeed/KakoClaw/main/install.sh | bash -s -- --update
```

### Método 2: Actualización Manual

```bash
# Backup de configuración
cp ~/.KakoClaw/config.json ~/.KakoClaw/config.json.backup

# Descargar nueva versión
wget https://github.com/sipeed/KakoClaw/releases/latest/download/KakoClaw-linux-amd64 -O KakoClaw
chmod +x KakoClaw
sudo mv KakoClaw /usr/local/bin/

# Verificar
KakoClaw version

# Restaurar config si es necesario
# cp ~/.KakoClaw/config.json.backup ~/.KakoClaw/config.json
```

### Método 3: Desde Fuente

```bash
cd KakoClaw
git pull origin main
make build
make install
```

## Desinstalación

### Desinstalación Completa

```bash
# Eliminar binario
sudo rm /usr/local/bin/KakoClaw

# Eliminar datos
rm -rf ~/.KakoClaw

# Eliminar autocompletado de shell
# Editar ~/.bashrc o ~/.zshrc y quitar líneas de KakoClaw
```

### Desinstalación con Make

```bash
cd KakoClaw
make uninstall
make uninstall-all  # Incluye workspace y configuración
```

## Verificación de la Instalación

Ejecuta este checklist:

```bash
# 1. Verificar binario
which KakoClaw
KakoClaw version

# 2. Verificar configuración
ls -la ~/.KakoClaw/
cat ~/.KakoClaw/config.json

# 3. Verificar workspace
ls -la ~/.KakoClaw/workspace/

# 4. Prueba funcional
KakoClaw agent -m "Di 'KakoClaw está funcionando correctamente'"

# 5. Verificar permisos
touch ~/.KakoClaw/workspace/test.txt
rm ~/.KakoClaw/workspace/test.txt
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
chmod +x /usr/local/bin/KakoClaw

# O si instalaste sin sudo
sudo chown $(whoami) /usr/local/bin/KakoClaw
```

### Error de GLIBC

En sistemas antiguos, compilar desde fuente:

```bash
# Estático linking
CGO_ENABLED=0 go build -ldflags="-s -w" -o KakoClaw ./cmd/KakoClaw
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

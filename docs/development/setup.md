# Configuración del Entorno de Desarrollo

Guía completa para configurar tu entorno de desarrollo y contribuir a KakoClaw.

## 📋 Requisitos

### Software Necesario

- **Go**: Versión 1.21 o superior
- **Git**: Para control de versiones
- **Make**: Para automatizar tareas
- **Editor**: VS Code, GoLand, Vim, o tu preferido

### Opcional pero Recomendado

- **Docker**: Para tests de integración
- **golangci-lint**: Para linting
- **mockgen**: Para generar mocks
- **delve**: Para debugging

## 🚀 Setup Inicial

### 1. Fork y Clone

```bash
# Fork el repositorio en GitHub
# Luego clonar tu fork
git clone https://github.com/TU_USUARIO/KakoClaw.git
cd KakoClaw

# Agregar upstream
git remote add upstream https://github.com/sipeed/KakoClaw.git
```

### 2. Verificar Go

```bash
go version
# Debe mostrar go1.21 o superior

# Verificar GOPATH
go env GOPATH

# Verificar GOROOT
go env GOROOT
```

### 3. Instalar Dependencias

```bash
# Descargar módulos
go mod download

# O usando make
make deps
```

### 4. Compilar

```bash
# Compilar para desarrollo
go build -o KakoClaw-dev ./cmd/KakoClaw

# O con make
make build

# Verificar que compila
./KakoClaw-dev version
```

## 🛠️ Configuración de IDE

### VS Code

#### Extensiones Recomendadas

Instala estas extensiones:

```json
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "github.copilot",
    "usernamehw.errorlens",
    "streetsidesoftware.code-spell-checker"
  ]
}
```

#### Configuración (settings.json)

```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.vulncheck": "Imports",
  "go.buildOnSave": "workspace",
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "gopls": {
    "ui.diagnostic.annotations": {
      "bounds": true,
      "escape": true,
      "inline": true,
      "nil": true
    },
    "formatting.gofumpt": true
  },
  "go.diagnostic.vulncheck": "imports",
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true,
  "go.coverOnSingleTest": true
}
```

#### Tasks (tasks.json)

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "Lint",
      "type": "shell",
      "command": "make lint"
    },
    {
      "label": "Run",
      "type": "shell",
      "command": "./build/KakoClaw agent",
      "dependsOn": ["Build"]
    }
  ]
}
```

#### Launch (launch.json)

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Agent",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/KakoClaw",
      "args": ["agent", "-m", "Hola"],
      "env": {
        "KakoClaw_DEBUG": "true"
      }
    },
    {
      "name": "Launch Gateway",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/KakoClaw",
      "args": ["gateway", "--debug"]
    },
    {
      "name": "Test Current Package",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/${relativeFileDirname}",
      "showLog": true
    }
  ]
}
```

### GoLand / IntelliJ

#### Configuración

1. **Abrir proyecto**: File → Open → Seleccionar carpeta KakoClaw
2. **Go SDK**: Settings → Go → Go SDK → Seleccionar Go 1.21+
3. **Go Modules**: Settings → Go → Go Modules → Enable
4. **Linter**: Settings → Tools → Go Linter → Seleccionar golangci-lint

#### Run Configurations

**Agent:**
```
Type: Go Application
Package: github.com/sipeed/KakoClaw/cmd/KakoClaw
Arguments: agent -m "Hola"
```

**Gateway:**
```
Type: Go Application
Package: github.com/sipeed/KakoClaw/cmd/KakoClaw
Arguments: gateway --debug
```

**Tests:**
```
Type: Go Test
Package: github.com/sipeed/KakoClaw/pkg/...
```

### Vim / Neovim

#### Vim-Go

```vim
" Instalar vim-go
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }

" Configuración
let g:go_fmt_command = "gofumpt"
let g:go_metalinter_enabled = ['vet', 'golint', 'errcheck']
let g:go_metalinter_autosave = 1
let g:go_def_mode='gopls'
let g:go_info_mode='gopls'
```

#### LSP (Neovim)

```lua
-- gopls configuration
require('lspconfig').gopls.setup({
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
        shadow = true,
      },
      staticcheck = true,
      gofumpt = true,
    },
  },
})
```

## 🧪 Instalación de Herramientas

### golangci-lint

```bash
# Instalar
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verificar
golangci-lint version
```

### gofumpt (formateador estricto)

```bash
go install mvdan.cc/gofumpt@latest
```

### mockgen

```bash
go install github.com/golang/mock/mockgen@latest
```

### delve (debugger)

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### air (hot reload)

```bash
go install github.com/cosmtrek/air@latest
```

## 📝 Configuración de Desarrollo

### 1. Configuración de Git

```bash
# Configurar git
git config user.name "Tu Nombre"
git config user.email "tu@email.com"

# Hooks (opcional)
cp .git-hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

### 2. Archivo de Configuración de Desarrollo

```bash
# Crear config de desarrollo
cp config.example.json ~/.KakoClaw/config.dev.json

# Configurar para desarrollo
```

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/KakoClaw-dev-workspace",
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 4096,
      "temperature": 0.7,
      "max_tool_iterations": 10
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-DEV-KEY"
    }
  }
}
```

### 3. Scripts de Desarrollo

Crea `scripts/dev.sh`:

```bash
#!/bin/bash
# Script de desarrollo

export KakoClaw_CONFIG="$HOME/.KakoClaw/config.dev.json"

case "$1" in
  "build")
    go build -o KakoClaw-dev ./cmd/KakoClaw
    ;;
  "agent")
    ./KakoClaw-dev agent --config "$KakoClaw_CONFIG"
    ;;
  "gateway")
    ./KakoClaw-dev gateway --config "$KakoClaw_CONFIG" --debug
    ;;
  "test")
    go test -v -race ./...
    ;;
  "lint")
    golangci-lint run
    ;;
  *)
    echo "Uso: $0 {build|agent|gateway|test|lint}"
    ;;
esac
```

```bash
chmod +x scripts/dev.sh
```

### 4. Hot Reload con Air

Crea `.air.toml`:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["agent"]
  bin = "./tmp/KakoClaw"
  cmd = "go build -o ./tmp/KakoClaw ./cmd/KakoClaw"
  delay = 1000
  exclude_dir = ["assets", "tmp", "testdata", "docs"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
```

Uso:
```bash
air
```

## 🔄 Workflow de Desarrollo

### 1. Antes de Empezar

```bash
# Actualizar tu fork
git checkout main
git fetch upstream
git rebase upstream/main
git push origin main

# Crear branch para tu feature
git checkout -b feature/mi-nueva-feature
```

### 2. Durante el Desarrollo

```bash
# Compilar frecuentemente
go build ./cmd/KakoClaw

# Ejecutar tests
make test

# Linting
make lint

# Formatear código
go fmt ./...
```

### 3. Testing

```bash
# Tests unitarios
go test -v ./pkg/tools/...

# Tests con race detector
go test -race ./...

# Tests de integración
go test -v -tags=integration ./...

# Coverage
make test-coverage
```

### 4. Commit

```bash
# Stage cambios
git add .

# Commit con mensaje descriptivo
git commit -m "feat: agrega nueva funcionalidad X

- Implementa funcionalidad Y
- Agrega tests
- Actualiza documentación"

# Push a tu fork
git push origin feature/mi-nueva-feature
```

### 5. Pull Request

1. Ve a GitHub
2. Crea Pull Request desde tu branch
3. Completa el template
4. Espera revisión

## 🐛 Debugging

### Con Delve

```bash
# Debug modo agent
dlv debug ./cmd/KakoClaw -- agent -m "test"

# Debug modo gateway
dlv debug ./cmd/KakoClaw -- gateway --debug

# En la consola de delve:
(dlv) break main.main
(dlv) continue
(dlv) next
(dlv) print variable
(dlv) locals
```

### Con VS Code

1. Set breakpoints en el código
2. F5 para iniciar debugging
3. Usar el panel de debugging para:
   - Ver variables
   - Evaluar expresiones
   - Ver call stack
   - Step over/into/out

### Logs de Debug

```bash
# Habilitar debug logging
KakoClaw agent --debug

# O con variable de entorno
KakoClaw_DEBUG=1 KakoClaw agent
```

## 📊 Profiling

### CPU Profile

```bash
# Generar perfil
go build -o KakoClaw-profile ./cmd/KakoClaw
./KakoClaw-profile agent -cpuprofile=cpu.prof -m "test"

# Analizar
go tool pprof cpu.prof
(pprof) top
(pprof) web
```

### Memory Profile

```bash
./KakoClaw-profile agent -memprofile=mem.prof -m "test"
go tool pprof mem.prof
```

### Benchmarks

```bash
# Ejecutar benchmarks
go test -bench=. -benchmem ./pkg/...
```

## 🧹 Limpieza

```bash
# Limpiar builds
make clean

# Limpiar módulos
go clean -modcache

# Limpiar test cache
go clean -testcache
```

## 📚 Recursos Adicionales

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

---

Para contribuir código, ver la [Guía de Contribución](./contributing.md).

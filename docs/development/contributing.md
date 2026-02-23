# Guía de Contribución

¡Gracias por tu interés en contribuir a MakoClaw! Este documento proporciona las pautas y mejores prácticas para contribuir al proyecto.

## 🤝 Cómo Contribuir

Hay muchas formas de contribuir:

- 🐛 Reportar bugs
- 💡 Sugerir nuevas funcionalidades
- 📝 Mejorar documentación
- 🔧 Enviar pull requests
- 🧪 Escribir tests
- 📢 Compartir el proyecto

## 🚀 Primeros Pasos

### 1. Fork el Repositorio

```bash
# Haz fork en GitHub, luego:
git clone https://github.com/TU_USUARIO/MakoClaw.git
cd MakoClaw

# Configura upstream
git remote add upstream https://github.com/sipeed/MakoClaw.git
```

### 2. Configura tu Entorno

Ver [Configuración del Entorno](./setup.md) para instrucciones detalladas.

### 3. Verifica que Todo Funciona

```bash
make build
make test
MakoClaw version
```

## 📋 Guías de Contribución

### Reportar Bugs

Antes de reportar:
1. Busca en [issues existentes](https://github.com/sipeed/MakoClaw/issues)
2. Verifica que estás usando la última versión
3. Intenta reproducir en un entorno limpio

**Template de Bug Report:**

```markdown
**Descripción**
Descripción clara del bug.

**Para Reproducir**
Pasos para reproducir:
1. Ir a '...'
2. Ejecutar comando '...'
3. Ver error

**Comportamiento Esperado**
Qué debería pasar.

**Screenshots/Logs**
Si aplica, agrega logs o screenshots.

**Entorno:**
 - OS: [e.g., Ubuntu 22.04]
 - Version: [e.g., 0.1.0]
 - Go Version: [e.g., 1.21]
 - Hardware: [e.g., x86_64, ARM64]

**Configuración**
```json
// Tu config.json (sin API keys)
```
```

### Sugerir Features

**Template de Feature Request:**

```markdown
**¿Tu feature está relacionada con un problema?**
Descripción clara del problema. Ej: "Siempre me frustra cuando..."

**Describe la solución que te gustaría**
Descripción clara de lo que quieres que pase.

**Describe alternativas que has considerado**
Otras soluciones o features alternativos.

**Contexto adicional**
Cualquier otro contexto, screenshots, etc.
```

### Pull Requests

#### Antes de Enviar

1. **Tests**: Asegúrate de que todos los tests pasan
   ```bash
   make test
   ```

2. **Linting**: El código debe pasar el linter
   ```bash
   make lint
   ```

3. **Formato**: El código debe estar formateado
   ```bash
   go fmt ./...
   ```

4. **Documentación**: Actualiza la documentación si es necesario

5. **Commits**: Sigue las [convenciones de commits](#convenciones-de-commits)

#### Proceso de PR

1. Crea un branch desde `main`
   ```bash
   git checkout -b feature/nombre-descriptivo
   ```

2. Haz tus cambios con commits atómicos

3. Push a tu fork
   ```bash
   git push origin feature/nombre-descriptivo
   ```

4. Crea el PR en GitHub

5. Completa el template del PR

6. Espera revisión (generalmente 24-48h)

#### Template de PR

```markdown
## Descripción
Breve descripción de los cambios.

## Tipo de Cambio
- [ ] Bug fix
- [ ] Nueva feature
- [ ] Breaking change
- [ ] Documentación

## Checklist
- [ ] He testeado mis cambios
- [ ] He actualizado la documentación
- [ ] Mis cambios no rompen tests existentes
- [ ] He seguido las guías de estilo
- [ ] He agregado tests para nueva funcionalidad

## Screenshots/Logs
Si aplica, agrega evidencia visual.

## Issues Relacionados
Fixes #123
```

## 📝 Convenciones de Código

### Estilo Go

Seguimos las convenciones estándar de Go:

```go
// Bueno: Nombres descriptivos, comentarios en exports
package tools

// FileReader proporciona operaciones de lectura de archivos.
type FileReader struct {
    workspace string
    restrict  bool
}

// NewFileReader crea una nueva instancia de FileReader.
// El workspace define el directorio base para operaciones.
// Si restrict es true, las operaciones se limitan al workspace.
func NewFileReader(workspace string, restrict bool) *FileReader {
    return &FileReader{
        workspace: workspace,
        restrict:  restrict,
    }
}

// Read lee el contenido de un archivo.
// Retorna error si el archivo no existe o no se puede leer.
func (r *FileReader) Read(path string) (string, error) {
    // Implementación
}
```

### Estructura de Paquetes

```
pkg/
├── feature/           # Un paquete por feature
│   ├── module.go     # Archivo principal
│   ├── types.go      # Tipos y interfaces
│   └── module_test.go # Tests
```

### Nomenclatura

- **Paquetes**: Minúsculas, sin guiones bajos (`tools`, `not tools_lib`)
- **Funciones exportadas**: PascalCase (`ReadFile`, `not readFile`)
- **Funciones privadas**: camelCase (`readInternal`)
- **Variables**: camelCase (`filePath`)
- **Constantes**: UPPER_SNAKE_CASE o PascalCase si exportadas
- **Interfaces**: Nombres descriptivos, terminan en "-er" (`Reader`, `Writer`)

### Convenciones de Commits

Usamos [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Tipos:**
- `feat`: Nueva feature
- `fix`: Bug fix
- `docs`: Cambios en documentación
- `style`: Formato, punto y coma, etc. (no cambia código)
- `refactor`: Refactorización de código
- `perf`: Mejora de performance
- `test`: Agregar o arreglar tests
- `chore`: Tareas de build, dependencias, etc.

**Ejemplos:**

```bash
# Feature
feat(tools): agrega soporte para búsqueda web

# Bug fix
fix(agent): corrige race condition en session manager

# Documentación
docs(readme): actualiza instrucciones de instalación

# Refactor
refactor(config): simplifica parsing de configuración

# Con breaking change
feat(agent)!: cambia API de process message

BREAKING CHANGE: `ProcessDirect` ahora retorna `(string, error, int)`
```

## 🧪 Testing

### Tests Unitarios

```go
func TestNewReadFileTool(t *testing.T) {
    tests := []struct {
        name      string
        workspace string
        restrict  bool
        wantErr   bool
    }{
        {
            name:      "valid workspace",
            workspace: "/tmp/test",
            restrict:  true,
            wantErr:   false,
        },
        {
            name:      "empty workspace",
            workspace: "",
            restrict:  false,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewReadFileTool(tt.workspace, tt.restrict)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewReadFileTool() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got == nil {
                t.Error("NewReadFileTool() returned nil")
            }
        })
    }
}
```

### Tests de Integración

```go
//go:build integration

package tools_test

func TestExecTool_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    tool := NewExecTool("/tmp", true)
    result, err := tool.Execute(context.Background(), map[string]interface{}{
        "command": "echo hello",
    })

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != "hello\n" {
        t.Errorf("expected 'hello\\n', got %q", result)
    }
}
```

### Coverage

```bash
# Ejecutar con coverage
make test-coverage

# Ver reporte HTML
go tool cover -html=coverage.out
```

## 📖 Documentación

### Código

- Todos los exports deben tener comentarios
- Comentarios deben empezar con el nombre del elemento
- Ejemplos son bienvenidos

```go
// ReadFile lee el contenido de un archivo.
//
// El archivo debe existir dentro del workspace si restrict está habilitado.
// Retorna error si el archivo no existe o no se puede leer.
//
// Ejemplo:
//   content, err := tool.ReadFile("config.json")
//   if err != nil {
//       log.Fatal(err)
//   }
func (t *ReadFileTool) ReadFile(filePath string) (string, error)
```

### Documentación de Usuario

Para cambios que afectan a usuarios:

1. Actualizar `docs/guides/`
2. Actualizar `README.md` si es necesario
3. Agregar ejemplos en `docs/examples/`

### Changelog

Las entradas del changelog se generan automáticamente desde los commits.

## 🔒 Seguridad

### Reportar Vulnerabilidades

**NO** abras un issue público. En su lugar:

1. Email a: security@sipeed.com
2. Incluye descripción detallada
3. Incluye pasos de reproducción
4. Propón solución si la tienes

### Mejores Prácticas

- Nunca commitees API keys
- Usa variables de entorno para secrets
- Valida todos los inputs
- Sanitiza paths de archivos
- Usa context con timeout para operaciones externas

## 🎯 Áreas de Contribución Prioritarias

### Alta Prioridad

1. **Tests**: Aumentar cobertura de tests
2. **Documentación**: Mejorar docs de API y guías
3. **Canales**: Agregar soporte para más plataformas
4. **Providers**: Soporte para más LLM providers
5. **Performance**: Optimizaciones de memoria y CPU

### Media Prioridad

1. **Skills**: Crear más skills útiles
2. **Tools**: Nuevas herramientas
3. **UX**: Mejorar CLI experience
4. **Internacionalización**: Soporte multi-idioma

### Baja Prioridad (pero bienvenidas)

1. **Refactorización**: Mejorar código existente
2. **Dependencies**: Actualizar dependencias
3. **CI/CD**: Mejorar pipelines

## 🏷️ Labels de Issues

Usamos estos labels para organizar issues:

- `bug`: Algo no funciona
- `enhancement`: Nueva feature
- `documentation`: Docs
- `good first issue`: Para nuevos contribuidores
- `help wanted`: Necesitamos ayuda
- `priority/high`: Urgente
- `priority/medium`: Importante
- `priority/low`: Cuando haya tiempo

## 💬 Comunicación

### Canales

- **GitHub Issues**: Bugs y features
- **GitHub Discussions**: Preguntas y discusiones
- **Discord**: Chat en tiempo real
- **Email**: Contribuciones grandes o privadas

### Código de Conducta

- Sé respetuoso y constructivo
- Acepta críticas constructivas
- Enfócate en lo que es mejor para la comunidad
- Muestra empatía hacia otros

## 🎓 Recursos para Contribuidores

### Aprender Go

- [A Tour of Go](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go by Example](https://gobyexample.com/)

### Arquitectura

- [Documentación de Arquitectura](../architecture/overview.md)
- [Crear un Tool](./creating-tools.md)
- [Crear un Canal](./creating-channels.md)

### Testing

- [Testing en Go](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)

## ❓ FAQ para Contribuidores

### ¿Necesito permiso para contribuir?

No, solo abre un issue o PR. Para cambios grandes, mejor abrir un issue primero para discutir.

### ¿Puedo agregar una dependencia?

Sí, pero justifícala. Preferimos mantener dependencias mínimas.

### ¿Cómo se elige qué PRs se mergean?

Criterios:
1. Calidad del código
2. Tests incluidos
3. Documentación actualizada
4. No rompe compatibilidad (a menos que esté planeado)
5. Resuelve un problema real

### ¿Cuánto tiempo toma la revisión?

Generalmente 24-48 horas para PRs pequeños, 3-5 días para PRs grandes.

### ¿Puedo ser maintainer?

Con contribuciones consistentes y de calidad durante varios meses, sí.

## 🎉 Reconocimientos

Los contribuidores serán reconocidos en:

- Archivo `CONTRIBUTORS.md`
- Release notes
- Documentación
- Twitter de la comunidad

---

¡Gracias por contribuir! 🦈

Para preguntas, únete a nuestro [Discord](https://discord.gg/V4sAZ9XWpN).

# Documentación de KakoClaw

Bienvenido a la documentación oficial de KakoClaw - Tu asistente de IA ultraligero.

## 📚 Estructura de la Documentación

### 🏗️ [Arquitectura](./architecture/)
Documentación técnica sobre la estructura interna y diseño del sistema.

- [Visión General](./architecture/overview.md)
- [Flujo de Datos](./architecture/data-flow.md)
- [Componentes Principales](./architecture/components.md)
- [Diagramas del Sistema](./architecture/diagrams.md)

### 📖 [Guías de Usuario](./guides/)
Guías paso a paso para usuarios finales.

- [Guía de Inicio Rápido](./guides/quickstart.md)
- [Instalación y Configuración](./guides/installation.md)
- [Configuración de Proveedores LLM](./guides/llm-providers.md)
- [Canales de Mensajería](./guides/channels.md)
- [Uso del Agente CLI](./guides/agent-cli.md)
- [Tareas Programadas](./guides/cron-jobs.md)
- [Sistema de Skills](./guides/skills.md)
- [Configuracion de Email](./guides/email-setup.md)

### 💻 [Desarrollo](./development/)
Documentación para contribuidores y desarrolladores.

- [Configuración del Entorno](./development/setup.md)
- [Estructura del Proyecto](./development/project-structure.md)
- [Guía de Contribución](./development/contributing.md)
- [Crear un Nuevo Tool](./development/creating-tools.md)
- [Crear un Nuevo Canal](./development/creating-channels.md)
- [Crear un Nuevo Skill](./development/creating-skills.md)
- [Tests y Calidad](./development/testing.md)
- [Convenciones de Código](./development/code-conventions.md)

### 📋 [Referencia de API](./api-reference/)
Documentación de referencia de interfaces y APIs.

- [Tools API](./api-reference/tools.md)
- [Providers API](./api-reference/providers.md)
- [Channels API](./api-reference/channels.md)
- [Config API](./api-reference/config.md)
- [Agent API](./api-reference/agent.md)

### 🚀 [Despliegue](./deployment/)
Guías para desplegar KakoClaw en diferentes entornos.

- [Despliegue Local](./deployment/local.md)
- [Despliegue en Servidor](./deployment/server.md)
- [Docker](./deployment/docker.md)
- [Systemd Service](./deployment/systemd.md)
- [Placas ARM/RISC-V](./deployment/embedded.md)

### 🎯 [Ejemplos](./examples/)
Ejemplos prácticos y casos de uso.

- [Ejemplos Básicos](./examples/basic-examples.md)
- [Automatización de Tareas](./examples/automation.md)
- [Integraciones](./examples/integrations.md)
- [Workflows Completos](./examples/workflows.md)

### 🔧 [Solución de Problemas](./troubleshooting/)
Ayuda para resolver problemas comunes.

- [Problemas Comunes](./troubleshooting/common-issues.md)
- [Errores de Configuración](./troubleshooting/config-errors.md)
- [Problemas de Canales](./troubleshooting/channel-issues.md)
- [Debugging](./troubleshooting/debugging.md)
- [FAQ](./troubleshooting/faq.md)

### 📊 [Análisis de Issues](./issues-analysis/)
Análisis y clasificación de issues abiertas en GitHub.

- [Resumen Ejecutivo](./issues-analysis/summary.md) - Overview de todas las issues
- [Análisis Completo](./issues-analysis/README.md) - Clasificación detallada
- [Planes de Implementación](./issues-analysis/implementation-plans.md) - Guías para contribuir

## 🚀 Empezando

### Instalación Rápida

```bash
# Clonar el repositorio
git clone https://github.com/sipeed/KakoClaw.git
cd KakoClaw

# Compilar
make build

# Instalar
make install

# Inicializar configuración
KakoClaw onboard
```

### Primer Uso

```bash
# Configurar tu API key en ~/.KakoClaw/config.json

# Iniciar una conversación
KakoClaw agent -m "Hola, ¿qué puedes hacer?"

# O modo interactivo
KakoClaw agent
```

## 📊 Estadísticas del Proyecto

- **Lenguaje**: Go 1.21+
- **Líneas de código**: ~13,600
- **Archivos**: 56 archivos Go
- **Memoria**: <10MB RAM
- **Tiempo de arranque**: <1 segundo
- **Licencia**: MIT

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor lee nuestra [Guía de Contribución](./development/contributing.md) antes de enviar un PR.

## 💬 Comunidad

- GitHub Issues: [https://github.com/sipeed/KakoClaw/issues](https://github.com/sipeed/KakoClaw/issues)
- Discord: [https://discord.gg/V4sAZ9XWpN](https://discord.gg/V4sAZ9XWpN)

## 📄 Licencia

KakoClaw está licenciado bajo la Licencia MIT. Ver [LICENSE](../LICENSE) para más detalles.

---

**Versión de la documentación**: 1.0  
**Última actualización**: Febrero 2026

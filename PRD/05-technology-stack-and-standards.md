# PRD 05 - Technology Stack and Engineering Standards

## Stack oficial

### Backend

- Go como lenguaje principal.
- Arquitectura modular en `pkg/*`.
- API web en `pkg/web`.
- Persistencia SQLite para contexto local/multiusuario.

### Frontend

- Vue 3 + Vite.
- Pinia para estado.
- Tailwind CSS con tokens de tema.
- E2E con Playwright/Cypress segun modulo.

### Integraciones

- Proveedores LLM (OpenAI, Anthropic, Ollama y otros configurables).
- MCP client para tool ecosystems externos.
- Canales de mensajeria multiples (Telegram, Discord, Slack, etc.).

### DevOps y despliegue

- Docker y docker-compose para ambientes de runtime.
- Makefile como interfaz de tareas recurrentes.
- Scripts de soporte en `scripts/`.

## Standards de arquitectura

1. Cambios grandes deben pasar por flujo spec-driven.
2. Evitar acoplamiento circular entre paquetes.
3. Mantener contratos explicitos en handlers, tools y providers.
4. Observabilidad y logs estructurados por componente.

## Standards de calidad

- Tests unitarios para logica critica.
- Tests de integracion para flujos de usuario.
- Verificacion de permisos/aislamiento en cambios de auth, storage y tools.
- Documentacion actualizada como parte de cada feature.

## Standards de seguridad

- Principio de minimo privilegio en tools y endpoints.
- No exponer secretos en logs, errores ni respuestas API.
- Validar inputs en frontera (HTTP, tools, canales, mcp).
- Revisiones de seguridad en features que tocan auth, filesystem, exec o channels.

## Standards de versionado y release

- Conventional commits.
- Changelog legible por usuario y por tecnico.
- Release gates definidos en `08-kpis-quality-and-release-gates.md`.

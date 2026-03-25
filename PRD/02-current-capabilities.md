# PRD 02 - Current Capabilities

## Estado funcional actual (alto nivel)

### Core platform

- Motor de agente con loop de herramientas y providers LLM multiples.
- Arquitectura por canales con normalizacion de mensajes en bus.
- Herramientas internas y registro dinamico de tools.
- Soporte MCP para integrar herramientas externas.

### Multiusuario

- Estructura por usuario con UUID.
- Configuracion global + overrides por usuario.
- Storage por usuario para datos operativos.
- Workspaces separados por usuario.

### Frontend web

- Vistas para chat, tareas, workflows, agents, settings, cron, knowledge, mcp, metrics.
- Streaming en chat, seleccion de modelo, sesiones con rename/archive/delete.
- Design system documentado en `docs/development/FRONTEND_DESIGN_SYSTEM.md`.

### Productividad y automatizacion

- Task board y task logs.
- Cron jobs con schedule visual y timezone.
- Knowledge base con FTS5 para busqueda.
- API documentada via OpenAPI/Swagger.

## Brechas prioritarias

1. Consolidar PRD operativo unico (resuelto con este hub).
2. Endurecer y validar aislamiento multiusuario extremo a extremo.
3. Definir release gates de seguridad/calidad obligatorios.
4. Formalizar handoff para delegacion a humanos/agentes.
5. Ordenar roadmap por fases ejecutables y dependencias reales.

## Riesgos actuales

- Documentacion dispersa entre roadmap, prd historicos y reportes.
- Inconsistencias potenciales entre implementacion y docs legacy.
- Riesgo de cambios ad hoc sin chequeos de arquitectura.

## Acciones inmediatas

- Usar este hub como fuente de verdad para nuevas iniciativas.
- Ejecutar auditoria de aislamiento con pruebas de caja negra y caja blanca.
- Conectar tareas del backlog al playbook de delegacion.

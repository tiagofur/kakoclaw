# PRD 01 - Product Vision

## Vision

MakoClaw es una plataforma de agentes AI multiusuario, segura y operable, pensada para correr en hardware austero sin sacrificar capacidades enterprise.

## Mision

Entregar una plataforma donde cada usuario tenga su propio espacio aislado para trabajar con agentes, herramientas, memoria y canales, con una experiencia moderna y una arquitectura mantenible.

## Principios de producto

1. Secure by default: ningun acceso entre usuarios salvo permisos explicitos de administracion.
2. Isolation first: configuracion, datos y workspace separados por usuario.
3. Performance real: mantener ventajas de eficiencia (<10MB RAM y startup rapido cuando aplique).
4. Operabilidad: logs, auditoria, estados de salud y recovery claros.
5. UX coherente: misma calidad de interfaz en todas las vistas.
6. Arquitectura sin parches: cambios guiados por spec, diseno, tareas y validacion.

## Personas principales

- Admin de plataforma: configura instancia, seguridad, canales globales y politicas.
- Usuario operativo: usa chat, tareas, workflows, memoria y herramientas.
- Desarrollador/contributor: extiende backend, frontend, tools y canales.
- AI specialist/agent: ejecuta tareas delegadas con contexto y limites definidos.

## Objetivos estrategicos 2026

1. Multiusuario real en produccion con aislamiento verificable.
2. Plataforma lista para delegacion de trabajo entre equipos y agentes.
3. Frontend con sistema visual consistente y alta usabilidad.
4. Escalabilidad funcional: swarms, workflows, MCP, knowledge y observabilidad.
5. Seguridad y compliance basicos cubiertos para uso empresarial.

## Non-goals (por ahora)

- Multi-tenant distribuido entre regiones con replicacion activa-activa.
- Marketplace abierto sin validacion ni controles de seguridad.
- Soporte de features experimentales sin telemetria y release gates.

## Criterios de exito

- Cero filtraciones de datos entre usuarios en pruebas de aislamiento.
- Tiempo de onboarding de nuevo colaborador < 1 dia usando solo docs.
- Features nuevas implementadas siguiendo flujo spec->design->tasks->apply->verify.
- Disminucion de incidencias por regresiones de UX/UI y permisos.

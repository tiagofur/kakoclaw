# PRD 08 - KPIs, Quality, and Release Gates

## Objetivo

Definir como medimos si MakoClaw mejora de forma real y sostenible, y cuando un release puede salir sin comprometer seguridad ni calidad.

## KPIs de producto

### Seguridad y aislamiento

- Incidentes de acceso cruzado entre usuarios: objetivo = 0.
- Vulnerabilidades criticas abiertas: objetivo = 0.
- Cobertura de pruebas de aislamiento en flujos core: objetivo >= 90% de escenarios definidos.

### Calidad y estabilidad

- Regresiones criticas por release: objetivo = 0.
- MTTR de incidencias severas: objetivo en descenso continuo.
- Tasa de errores en endpoints criticos (auth, storage, chat): objetivo bajo y estable.

### Experiencia de usuario

- Tiempo de primer valor (onboarding a primer mensaje util).
- Tasa de exito en tareas principales (chat, tareas, workflows).
- Satisfaccion cualitativa de UI/UX por iteracion.

### Delivery

- Lead time de cambios (idea -> produccion).
- Ratio de tareas completadas con DoD completo.
- Porcentaje de cambios con docs actualizados.

## Release gates obligatorios

Un release no pasa si falla cualquiera de estos puntos:

1. Seguridad

- Sin findings criticos abiertos.
- Verificacion de aislamiento multiusuario en verde.
- Validacion de permisos y controles de acceso.

2. Calidad tecnica

- Tests relevantes en verde.
- Sin errores bloqueantes en flujos core.
- Validacion de migraciones y compatibilidad de datos.

3. Calidad UX/UI

- Cumplimiento de design system en cambios visuales.
- Estados vacio/carga/error/exito implementados.
- Verificacion desktop + mobile para vistas afectadas.

4. Operacion y observabilidad

- Logs y trazas suficientes para diagnostico.
- Alertas minimas activas para fallas graves.
- Runbook o pasos de rollback definidos para cambios sensibles.

5. Documentacion

- PRD y/o docs tecnicos actualizados.
- Changelog actualizado cuando corresponde.

## Cadencia de revision

- KPI review quincenal para equipo tecnico.
- Revision mensual de roadmap y prioridades.
- Retro por release para ajustar gates y procesos.

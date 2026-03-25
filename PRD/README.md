# MakoClaw Product Requirements Hub (PRD)

## Objetivo

Este directorio concentra la informacion de producto y arquitectura necesaria para ejecutar MakoClaw como un sistema de nivel produccion, con foco en:

- Multiusuario real con aislamiento de datos.
- Seguridad por defecto y trazabilidad.
- UX/UI consistente en todo el producto.
- Delegacion clara para equipos humanos y agentes.
- Evolucion sin "parches" ni deuda estructural.

## Como usar este hub

1. Empezar por `01-product-vision.md` para entender direccion y criterios de exito.
2. Revisar `02-current-capabilities.md` para saber que ya existe y que falta.
3. Usar `03-ux-ui-system.md` para toda decision de interfaz y experiencia.
4. Basarse en `04-multiuser-architecture-and-isolation.md` para cambios de backend, auth, datos y permisos.
5. Seguir `07-delivery-and-delegation-playbook.md` para ejecutar trabajo sin romper cohesion del sistema.

## Estructura

- `01-product-vision.md`
- `02-current-capabilities.md`
- `03-ux-ui-system.md`
- `04-multiuser-architecture-and-isolation.md`
- `05-technology-stack-and-standards.md`
- `06-roadmap-and-execution-phases.md`
- `07-delivery-and-delegation-playbook.md`
- `08-kpis-quality-and-release-gates.md`

## Fuentes consolidadas

Este hub consolida y referencia contenido ya existente en el repo:

- `docs/PRD.md`
- `docs/PRD-NEW-FEATURES.md`
- `docs/ROADMAP.md`
- `docs/development/FRONTEND_DESIGN_SYSTEM.md`
- `docs/architecture/overview.md`
- `docs/architecture/data-flow.md`
- `CLAUDE.md`
- `AGENTS.md`

Los documentos anteriores se mantienen como historial tecnico. Este hub define la version operativa actual para delegar y ejecutar.

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

### Core Platform
- `01-product-vision.md` - Visión, misión, principios, personas, objetivos estratégicos
- `02-current-capabilities.md` - Capacidades actuales del sistema
- `03-ux-ui-system.md` - Sistema de diseño UI/UX
- `04-multiuser-architecture-and-isolation.md` - Arquitectura multiusuario y aislamiento
- `05-technology-stack-and-standards.md` - Stack tecnológico y estándares de ingeniería
- `06-roadmap-and-execution-phases.md` - Roadmap y fases de ejecución
- `07-delivery-and-delegation-playbook.md` - Playbook para entrega y delegación
- `08-kpis-quality-and-release-gates.md` - KPIs, calidad y gates de release

### Mobile Apps
- `09-android-native-app.md` - App nativa Android (Kotlin, Jetpack Compose, Material3)
- `10-ios-native-app.md` - App nativa iOS (Swift, SwiftUI, Human Interface Guidelines)

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

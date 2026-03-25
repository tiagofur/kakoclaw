# PRD 03 - UX/UI System

## Objetivo UX/UI

Construir una experiencia consistente, rapida y entendible, donde cada pantalla se perciba parte del mismo sistema y no una suma de componentes desconectados.

## Principios UX

1. Claridad de estado: siempre mostrar que esta pasando (loading, streaming, success, error).
2. Flujo continuo: minimizar friccion entre chat, tareas, workflows y configuracion.
3. Feedback inmediato: acciones con resultado visible (toasts, estados, progreso).
4. Escalabilidad visual: la UI debe soportar crecimiento de features sin colapsar navegacion.
5. Mobile-first pragmatica: usable en mobile, potente en desktop.

## Sistema visual base

Fuente de referencia: `docs/development/FRONTEND_DESIGN_SYSTEM.md`

- Tokens semanticos de color (`makoclaw-*`).
- Componentes glass panel/cards/buttons/badges estandar.
- Gradientes por pantalla para identidad visual consistente.
- Capas z-index definidas para overlays, modales y toasts.
- Animaciones sutiles y funcionales.

## Patrones de interfaz obligatorios

- Header de pagina consistente con icono, titulo, subtitulo y accion principal.
- Vistas con empty state explicito y CTA claro.
- Formularios con validacion inline y mensajes accionables.
- Confirmaciones en operaciones destructivas.
- Persistencia de preferencias de UI (tema, filtros relevantes).

## Information Architecture (IA)

Areas nucleares del producto:

- Operacion: Chat, History, Tasks, Cron.
- Conocimiento y automatizacion: Knowledge, Workflows, MCP.
- Configuracion y gobierno: Settings, Agents, Channels, Providers.
- Observabilidad: Dashboard, Metrics, Reports.

## Criterios de calidad UX/UI

1. Consistencia: mismo lenguaje visual y de interaccion en todas las vistas.
2. Legibilidad: contraste, espaciado y jerarquia tipografica claros.
3. Accesibilidad minima: estados de foco, labels, touch targets >= 40px.
4. Rendimiento percibido: skeletons/transiciones sin bloquear acciones.
5. Error recovery: cada error tiene via de recuperacion visible.

## Definition of Done UX

Una feature visual se considera lista cuando:

- Respeta el design system.
- Tiene estados vacio/carga/error/exito.
- Funciona en desktop y mobile.
- No rompe patrones de navegacion global.
- Incluye evidencia visual (capturas o gif) en PR.

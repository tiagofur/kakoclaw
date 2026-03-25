# Mobile App Development Tasks

Este documento resume las tareas pendientes y completadas para la aplicación móvil **MakoClaw**.

## Fase 1 – Infraestructura (Fundación) ✅

- [x] Servicio API base (`src/services/api.ts`)
- [x] Servicio de autenticación (`src/services/authService.ts`)
- [x] Store de autenticación con Pinia (`src/stores/authStore.ts`)
- [x] Instalación de Pinia
- [x] Pantalla de Login (`src/views/LoginPage.vue`)
- [x] Pantalla de Settings (URL del servidor, tema)
- [x] Store de configuración (`src/stores/configStore.ts`)
- [x] Actualización del router con nuevas rutas

## Fase 2 – Chat Funcional ⏳

- [ ] Crear WebSocket service (`src/services/websocketService.ts`)
- [ ] Crear chat store (`src/stores/chatStore.ts`)
- [ ] Crear chat service (`src/services/chatService.ts`)
- [ ] Refactorizar `HomePage.vue` con datos reales del backend
- [ ] Implementar selector de agente en el header
- [ ] Implementar streaming de respuestas

## Fase 3 – Tasks y Navegación ⏳

- [ ] Crear task store (`src/stores/taskStore.ts`)
- [ ] Crear task service (`src/services/taskService.ts`)
- [ ] Refactorizar `KanbanPage.vue` con datos reales
- [ ] Implementar tab bar de navegación inferior
- [ ] Crear vista de historial de conversaciones

## Fase 4 – Features Exclusivas Mobile ⏳

- [ ] Captura con cámara (Capacitor Camera)
- [ ] Dictado por voz
- [ ] Notificaciones push
- [ ] Haptics contextuales

---

_Este documento se generó automáticamente para guiar el desarrollo._

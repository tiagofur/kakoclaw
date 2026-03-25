Reporte de Estado del MVP - MakoClaw
Resumen Ejecutivo
Preparación General del MVP: 82%

MakoClaw está muy cerca de estar listo para producción. La arquitectura es sólida, los componentes core están completos, y la mayoría de las funcionalidades están operativas.

Estado por Componente
Componente Completitud Estado
Agent Loop 95% ✅ Listo
Sistema de Tools 90% ✅ Listo
Providers (10) 100% ✅ Listo
Channels (9) 90% ✅ Listo
Storage SQLite 95% ✅ Listo
Web UI 75% ⚠️ Necesita polish
Configuración 85% ⚠️ Necesita solución
Seguridad 90% ⚠️ Config.json pendiente
Tests 80% ✅ 221 tests pasando
🚨 Tu Pregunta sobre Config.json
El Problema Actual

~/.MakoClaw/users/{uuid}/
├── config.json ← FUERA del workspace (API keys, tokens)
└── workspace/ ← SANDBOX del agente (restringido)
└── sessions/, skills/, etc.
Por qué está así: Seguridad. El agente está sandboxeado al workspace y no puede acceder a config.json que contiene API keys de 10+ providers y 9 channels.

El problema: El usuario dice "Agrega mi token de Telegram: 123..." y el agente no puede ayudar.

Solución Recomendada: Tool configure
Crear una herramienta dedicada que:

Solo permite escribir (no leer) valores sensibles
Whitelist estricta de campos permitidos
Redacta siempre los valores en la respuesta
Auditoría completa de todos los cambios
Ejemplo de uso:

Usuario: "Agrega mi token de Telegram: 123456:ABC-xyz..."
Agente: [llama configure(section="channels", provider="telegram", field="token", value="...")]
Agente: "✓ Actualizado channels.telegram.token a \*\*\*\*. Telegram configurado."

Usuario: "¿Cuál es mi token de Telegram?"
Agente: "No tengo acceso de lectura a valores sensibles. Esto protege tus credenciales."
Esfuerzo: 2-3 días

Bloqueadores Críticos para MVP
Prioridad Item Esfuerzo Impacto
🚨 CRÍTICO Tool configure 2-3 días Sin esto, el agente no puede ayudar con setup
⚠️ ALTO Knowledge Base UI 2-3 días Feature core sin interfaz de upload
⚠️ MEDIO Métricas con charts 1-2 días Dashboard básico funciona pero sin visualización
Features Completas ✅
Agent Loop - Tool calling, streaming, multi-iteración, sesiones
14 Tools - Filesystem, shell, web, email, spawn, tasks, knowledge
10 Providers - Claude, OpenAI, OpenRouter, Groq, Ollama, etc.
9 Channels - Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam
Multi-usuario - DBs aisladas, workspaces separados, permisos por rol
Web UI - 13/20 vistas completas (Chat, Tasks Kanban, Settings, Dashboard, etc.)
Skills System - 6 skills built-in + extensible
Auth - JWT + bcrypt + roles
MCP - Integración de tools externos
Features Incompletas ⚠️
Feature Estado Falta
Knowledge Base 70% UI de upload y búsqueda
Métricas Dashboard 60% Charts y visualización
Workflow Engine 50% Editor visual de nodos
Files View 30% Upload, preview, syntax
Reports 0% Todo
Tiempo Estimado hasta MVP Funcional
5-8 días de desarrollo enfocado

Días 1-3: Implementar tool configure
Días 3-5: Completar Knowledge Base UI
Días 5-6: Agregar charts a métricas
Días 6-8: Testing, documentación de deploy, polish
Mi Recomendación
¿Sugieres que comience con el tool configure?

Es el bloqueador más crítico porque afecta directamente la experiencia de usuario - sin él, MakoClaw no puede ayudar a los usuarios a configurar sus propios channels y providers, que es una de las promesas core de un "AI assistant".

¿Quieres que inicie /sdd:new configure-tool para planificar la implementación?

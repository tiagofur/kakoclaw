# Análisis de Issues de KakoClaw

Documento de análisis y clasificación de issues abiertas en el repositorio de KakoClaw.

**Fecha de análisis:** Febrero 2026  
**Total Issues abiertas:** 23  
**Total PRs abiertos:** 5  
**Repositorio:** https://github.com/sipeed/KakoClaw

---

## Resumen por Categoría

| Categoría | Issues | Útiles | No Útiles | Prioridad |
|-----------|--------|--------|-----------|-----------|
| **Providers LLM** | 6 | 5 | 1 | Alta |
| **Canales** | 5 | 5 | 0 | Media |
| **Configuración** | 4 | 3 | 1 | Alta |
| **Features** | 4 | 3 | 1 | Media |
| **Bug Fixes** | 2 | 2 | 0 | Alta |
| **Hardware** | 2 | 1 | 1 | Baja |

**Leyenda:**
- ✅ **Útil** - Issue que aporta valor real al proyecto
- ❌ **No Útil** - Issue poco relevante, spam o duplicado
- 🔴 **Prioridad Alta** - Crítico para el funcionamiento
- 🟡 **Prioridad Media** - Mejora importante
- 🟢 **Prioridad Baja** - Nice to have

---

## Issues por Categoría

### 1. Providers LLM (6 issues)

#### #75 - Support for local LLM; ollama?
- **Estado:** Open
- **Autor:** watrworld
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Feature Request
- **Descripción:** Solicita soporte para Ollama (LLMs locales)
- **Análisis:** Crítico para usuarios que quieren privacidad o trabajar offline. Muy solicitado en la comunidad.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-75)

#### #68 - [Feature request] Add open code and antigravity
- **Estado:** Open
- **Autor:** tuanlevi95
- **Útil:** ❌ NO
- **Prioridad:** 🟢 Baja
- **Tipo:** Feature Request
- **Descripción:** Solicita agregar "open code" y "antigravity" (servicios gratuitos)
- **Análisis:** Vago, no especifica qué servicios exactos ni cómo integrarlos. Parece spam o solicitud sin investigación previa.

#### #66 - KakoClaw_PROVIDERS_* env vars not applied
- **Estado:** Open
- **Autor:** binkbink168
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** Las variables de entorno con `{{.Name}}` no funcionan con caarlos0/env
- **Análisis:** Bug crítico que rompe la configuración por environment variables. Fácil de reproducir y fix.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-66)

#### #43 - Improve how models are assigned to providers
- **Estado:** Open
- **Autor:** vijaykarthiktk
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Enhancement
- **Descripción:** Mejorar la asignación automática de modelos a providers
- **Análisis:** Mejora UX importante. Actualmente el mapeo model→provider es confuso.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-43)

#### #17 - No way to explicitly select which provider is being used
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Enhancement
- **Descripción:** No hay forma de elegir explícitamente qué provider usar
- **Análisis:** Limitación real. Actualmente se infiere por la API key configurada.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-17)

#### #16 - OpenAI API key does not work - complains about max_tokens
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** Error con max_tokens al usar OpenAI
- **Análisis:** Bug que impide usar OpenAI directamente. Afecta a usuarios que prefieren OpenAI sobre OpenRouter.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-16)

---

### 2. Canales (5 issues)

#### #62 - BUG: Telegram allow_from with numeric user ID does not work when the user has a username
- **Estado:** Open
- **Autor:** ackness
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** El filtro allow_from no funciona correctamente cuando el usuario tiene username
- **Análisis:** Bug de seguridad/funcionalidad importante. Afecta a usuarios de Telegram.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-62)

#### #41 - Feat: Add Signal channel integration
- **Estado:** Open
- **Autor:** eti0
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Agregar soporte para Signal messenger
- **Análisis:** Signal es popular en usuarios de privacidad. Buena adición pero no crítica.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-41)

#### #37 - KakoClaw can't send messages on Telegram Gateway by itself
- **Estado:** Open
- **Autor:** shuantsu
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** El bot no puede enviar mensajes proactivos, solo responder
- **Análisis:** Limitación importante para casos de uso como recordatorios o alertas.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-37)

#### #36 - Telegram Gateway hangs on "Thinking..." after successful connection
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** El gateway se queda en "Thinking..." indefinidamente
- **Análisis:** Bug crítico que hace que Telegram sea inusable. Posible timeout o error no manejado.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-36)

#### #28 - Feat Request: LM Studio Easy Connect
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Integración fácil con LM Studio (LLM local)
- **Análisis:** Similar a #75 pero específico para LM Studio. Buena para usuarios que ya usan LM Studio.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-28)

---

### 3. Configuración (4 issues)

#### #46 - Configuration file modification suggestions
- **Estado:** Open
- **Autor:** lizhichao
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Enhancement
- **Descripción:** Sugerencias para mejorar el formato del archivo de configuración
- **Análisis:** Feedback valioso de usuario. Puede incluir validación, mejor estructura, etc.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-46)

#### #39 - Feature Request: Add `KakoClaw doctor` command
- **Estado:** Open
- **Autor:** vijaykarthiktk
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Comando para diagnosticar problemas de configuración
- **Análisis:** Muy útil para troubleshooting. Similar a `brew doctor` o `flutter doctor`.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-39)

#### #15 - Build fails on 32-bit ARM (linux/armv7): math.MaxInt64 overflow
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** Fallo de compilación en ARM 32-bit debido a overflow
- **Análisis:** Crítico para soporte de Raspberry Pi 32-bit y dispositivos embebidos.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-15)

#### #9 - Urgent need of rate limiters
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Necesidad de rate limiting para evitar abuso
- **Análisis:** Importante para producción y APIs con límites. Prevenir costos inesperados.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-9)

---

### 4. Features (4 issues)

#### #63 - [Feature Request] Manage cronjobs within session
- **Estado:** Open
- **Autor:** JokerQyou
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Gestionar cronjobs desde la sesión de chat, no solo CLI
- **Análisis:** Mejora UX. Permitir crear/editar tareas programadas conversacionalmente.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-63)

#### #61 - Implement file sending and receiving in chat
- **Estado:** Open
- **Autor:** vijaykarthiktk
- **Útil:** ✅ SÍ
- **Prioridad:** 🟡 Media
- **Tipo:** Feature Request
- **Descripción:** Permitir enviar y recibir archivos en los chats
- **Análisis:** Feature importante para compartir documentos, imágenes, etc.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-61)

#### #59 - OpenAI OAuth blank page
- **Estado:** Open
- **Autor:** AtefR
- **Útil:** ✅ SÍ
- **Prioridad:** 🔴 Alta
- **Tipo:** Bug
- **Descripción:** La página de OAuth de OpenAI aparece en blanco
- **Análisis:** Bug que impide usar autenticación OAuth con OpenAI.
- **Implementación:** Ver [implementation-plans.md](./implementation-plans.md#issue-59)

#### #11 - [RFC / Partnership] Building Water AI: The "Excel" of the Post-LLM Era
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ❌ NO
- **Prioridad:** 🟢 Baja
- **Tipo:** Spam/Promoción
- **Descripción:** Propuesta de "partnership" para otro producto
- **Análisis:** Spam o promoción de otro proyecto no relacionado. No es una issue real.

---

### 5. Hardware (2 issues)

#### #35 - Adjust to esp32
- **Estado:** Open
- **Autor:** (sin datos)
- **Útil:** ❌ NO
- **Prioridad:** 🟢 Baja
- **Tipo:** Feature Request
- **Descripción:** Adaptar para ESP32
- **Análisis:** ESP32 tiene recursos muy limitados (512KB RAM). KakoClaw requiere ~10MB. Imposible sin reescritura total.

#### #6 - Support for RISC-V
- **Estado:** Open  
- **Autor:** (sin datos)
- **Útil:** ✅ SÍ
- **Prioridad:** 🟢 Baja
- **Tipo:** Feature Request
- **Descripción:** Soporte para arquitectura RISC-V
- **Análisis:** Ya funciona en RISC-V según el README. Posiblemente documentación desactualizada.

---

## Pull Requests Abiertos

### #70 - feat: add Ollama search tools and update LLM providers
- **Autor:** instax-dutta
- **Estado:** PR Open
- **Descripción:** Agrega soporte para Ollama, NVIDIA NIM, Moonshot, y fix para 32-bit ARM
- **Relevancia:** Muy relevante, implementa varios issues (#75, #15)

### #67 - feat: added Docker Support
- **Autor:** fahadahmadansari111
- **Estado:** PR Open
- **Descripción:** Soporte completo para Docker con multi-stage build, docker-compose, CI/CD
- **Relevancia:** Excelente aportación para deployment

### #65 - feat: add Moonshot/Kimi and NVIDIA provider support
- **Autor:** siciyuan404
- **Estado:** PR Open
- **Descripción:** Agrega soporte para Moonshot y NVIDIA
- **Relevancia:** Buena, complementa #70

### #64 - [Sin datos completos]
- **Estado:** PR Open

### #60 - [Sin datos completos]
- **Estado:** PR Open

---

## Recomendaciones

### Issues Prioritarias (Alta prioridad)

1. **#66** - Bug de env vars (crítico)
2. **#62** - Bug de Telegram allow_from (seguridad)
3. **#36** - Telegram "Thinking..." hang (usabilidad)
4. **#16** - OpenAI max_tokens bug (funcionalidad)
5. **#15** - Build ARM 32-bit (compatibilidad)

### Issues Recomendadas para Contribuir

1. **#39** - `KakoClaw doctor` (fácil, buena primera contribución)
2. **#46** - Mejoras config (medio, mejora UX)
3. **#63** - Cronjobs en session (medio, feature útil)
4. **#75** - Soporte Ollama (difícil pero valioso)

### Issues a Cerrar/Revisar

1. **#68** - Vago, pedir más información o cerrar
2. **#11** - Spam, cerrar
3. **#35** - No factible, cerrar con explicación

---

## Próximos Pasos

Para cada issue marcada como "Útil", ver el documento [implementation-plans.md](./implementation-plans.md) donde se detalla:
- Qué implementar exactamente
- Cómo hacerlo paso a paso
- Archivos a modificar
- Ejemplos de código

---

*Documento generado automáticamente. Última actualización: Febrero 2026*

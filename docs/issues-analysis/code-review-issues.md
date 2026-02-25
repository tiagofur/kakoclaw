# Reporte de Code Review - Problemas de Seguridad, Concurrencia y Calidad

**Fecha:** Febrero 2026
**Branch:** `claude/code-review-improvements-ZwCI6`
**Revisión:** Auditoría completa del código fuente Go

---

## Resumen

Se realizó una auditoría exhaustiva del código, identificando problemas de seguridad, race conditions, fugas de recursos y manejo de errores. Se corrigieron **35+ problemas** en 4 commits. A continuación se documenta cada problema, su estado actual, y los problemas pendientes que aún requieren atención.

### Estadísticas

| Categoría | Corregidos | Pendientes |
|-----------|:----------:|:----------:|
| Seguridad Crítica | 12 | 4 |
| Race Conditions | 6 | 1 |
| Fugas de Recursos | 3 | 2 |
| Manejo de Errores | 10 | 3 |
| Calidad de Código | 4 | 5 |
| **Total** | **35** | **15** |

---

## PROBLEMAS CORREGIDOS

### Seguridad - Corregidos

#### 1. Inyección de cabeceras de email
- **Archivo:** `pkg/tools/email.go`
- **Severidad:** CRITICA
- **Problema:** Los campos `subject` y `to` no sanitizaban `\r\n`, permitiendo inyección de cabeceras SMTP.
- **Corrección:** Se sanitizan `\r` y `\n` en campos subject y to.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 2. Path traversal via symlinks en filesystem
- **Archivo:** `pkg/tools/filesystem.go`
- **Severidad:** CRITICA
- **Problema:** Sin resolución de symlinks, un atacante podía crear un enlace simbólico apuntando fuera del workspace.
- **Corrección:** Se usa `filepath.EvalSymlinks()` + validación `filepath.Rel()` para verificar que la ruta resuelta está dentro del workspace.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 3. Bypass de límite de directorio con rutas como `/workspace-hack`
- **Archivo:** `pkg/tools/filesystem.go`
- **Severidad:** CRITICA
- **Problema:** La validación `strings.HasPrefix` con la ruta del workspace permitía que rutas con prefijo similar (ej. `/workspace-hack`) pasaran la validación.
- **Corrección:** Se usa `filepath.Rel()` para validación correcta de contención.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 4. working_dir del shell no validado contra workspace
- **Archivo:** `pkg/tools/shell.go`
- **Severidad:** ALTA
- **Problema:** El parámetro `working_dir` de la herramienta `exec` no se validaba contra el workspace, permitiendo ejecución fuera de él.
- **Corrección:** Se valida `working_dir` con `filepath.Rel()` cuando `restrictToWorkspace` está activo.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 5. Timing attack en autenticación
- **Archivo:** `pkg/web/auth.go`
- **Severidad:** ALTA
- **Problema:** Se verificaba la contraseña antes de verificar si la cuenta estaba bloqueada, permitiendo inferir existencia de cuentas.
- **Corrección:** Se verifica el estado bloqueado antes de la comparación de contraseñas.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 6. Construcción insegura de JSON via concatenación de strings
- **Archivos:** `pkg/storage/setup_session.go`, `pkg/storage/central.go`
- **Severidad:** ALTA
- **Problema:** JSON construido por concatenación de strings, vulnerable a inyección.
- **Corrección:** Se reemplazó con `json.Marshal()` para serialización segura.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 7. Inyección de wildcards en consultas LIKE
- **Archivos:** `pkg/storage/chat.go`, `pkg/storage/task.go`
- **Severidad:** MEDIA
- **Problema:** Los caracteres `%` y `_` en queries de búsqueda no se escapaban, permitiendo wildcard injection.
- **Corrección:** Se escapan caracteres especiales de LIKE antes de usar en la consulta.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 8. Auto-asignación de rol admin en registro
- **Archivo:** `pkg/web/server.go`
- **Severidad:** CRITICA
- **Problema:** El endpoint de registro permitía que el usuario se asignara el rol "admin" en el payload.
- **Corrección:** Se fuerza `role = "user"` para todos los registros públicos.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 9. Path traversal en claves de sesión
- **Archivo:** `pkg/session/manager.go`
- **Severidad:** CRITICA
- **Problema:** La clave de sesión se usaba directamente como nombre de archivo sin sanitizar, permitiendo leer/escribir archivos arbitrarios via `../../`.
- **Corrección:** Se añade `sanitizeSessionKey()` que elimina separadores de ruta y `..`.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 10. Delete de knowledge destruía chunks antes de verificar ownership
- **Archivo:** `pkg/storage/knowledge.go`
- **Severidad:** CRITICA
- **Problema:** Al eliminar un documento, se borraban los chunks antes de verificar que el usuario fuera el dueño.
- **Corrección:** Se reordena para verificar ownership primero.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 11. WebSocket sin límite de tamaño de mensaje (DoS de memoria)
- **Archivo:** `pkg/web/server.go`
- **Severidad:** ALTA
- **Problema:** Conexiones WebSocket sin límite de lectura, permitiendo enviar mensajes de tamaño arbitrario.
- **Corrección:** Se añade `SetReadLimit(1 << 20)` (1 MB) en ambos endpoints WebSocket.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 12. Lectura ilimitada de respuestas HTTP en web_fetch/web_search
- **Archivos:** `pkg/tools/web.go`
- **Severidad:** ALTA
- **Problema:** Los cuerpos de respuesta HTTP se leían completos sin límite, permitiendo DoS por consumo de memoria.
- **Corrección:** Se añade `io.LimitReader` (10 MB para fetch, 5 MB para search).
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 13. Zip slip en importación de backups
- **Archivo:** `pkg/web/handlers_backup.go`
- **Severidad:** ALTA
- **Problema:** Entradas de archivos zip con rutas tipo `../../etc/passwd` podían escapar el directorio destino.
- **Corrección:** Se valida `f.Name` para `..` antes de la extracción.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 14. Exfiltración via symlinks en exportación de backup
- **Archivo:** `pkg/web/handlers_backup.go`
- **Severidad:** ALTA
- **Problema:** `filepath.Walk` seguía symlinks durante la exportación, permitiendo incluir archivos fuera del workspace.
- **Corrección:** Se omiten symlinks durante el walk.
- **Commit:** `65c7b55`
- **Estado:** ~~CORREGIDO~~

#### 15. Validación débil de email en registro
- **Archivo:** `pkg/web/auth.go`
- **Severidad:** MEDIA
- **Problema:** Solo se verificaba la presencia de `@` y `.`, no la conformidad RFC 5322.
- **Corrección:** Se reemplaza con `net/mail.ParseAddress` para validación completa.
- **Commit:** `65c7b55`
- **Estado:** ~~CORREGIDO~~

#### 16. Política de contraseña inconsistente
- **Archivos:** `pkg/web/auth.go`, `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** Longitud mínima de contraseña era 8 en registro pero 10 en cambio de contraseña.
- **Corrección:** Se estandariza a 8 caracteres en ambos endpoints.
- **Commit:** `65c7b55`
- **Estado:** ~~CORREGIDO~~

### Race Conditions - Corregidos

#### 17. Race condition en BaseChannel.running
- **Archivo:** `pkg/channels/base.go`
- **Severidad:** ALTA
- **Problema:** El campo `running` de tipo `bool` se accedía desde múltiples goroutines sin sincronización.
- **Corrección:** Se reemplaza con `atomic.Bool`.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 18. Panic en bus.go al enviar a canal cerrado
- **Archivo:** `pkg/bus/bus.go`
- **Severidad:** ALTA
- **Problema:** `Close()` cerraba los canales de Go, pero goroutines podían intentar enviar después del cierre, causando panic.
- **Corrección:** Se implementa patrón de cierre basado en señales en lugar de cerrar canales directamente.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 19. Race en MCP client entre readLoop y Close
- **Archivo:** `pkg/mcp/client.go`
- **Severidad:** ALTA
- **Problema:** `Close()` cerraba canales pendientes mientras `readLoop()` podía estar enviando a ellos.
- **Corrección:** `Close()` envía `nil` a los canales pendientes en lugar de cerrarlos, evitando send-on-closed-channel.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 20. Race en Specialist tool-swap sin mutex
- **Archivo:** `pkg/agent/specialist.go`
- **Severidad:** ALTA
- **Problema:** `ProcessWithSpeciality` reemplazaba herramientas del AgentLoop sin sincronización.
- **Corrección:** Se añade mutex para serializar las operaciones de intercambio de herramientas.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 21. Race en SpawnTool en campos de contexto
- **Archivo:** `pkg/tools/spawn.go`
- **Severidad:** ALTA
- **Problema:** `SetContext()` y `Execute()` accedían a `originChannel` y `originChatID` sin protección.
- **Corrección:** Se añade mutex para proteger los campos.
- **Commit:** `65c7b55`
- **Estado:** ~~CORREGIDO~~

#### 22. TOCTOU race en GetOrCreateForUser de sesiones
- **Archivo:** `pkg/session/manager.go`
- **Severidad:** ALTA
- **Problema:** Race condition entre verificar si una sesión existe y crearla.
- **Corrección:** Se añade doble verificación bajo write lock.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

### Fugas de Recursos - Corregidos

#### 23. Context leak en Slack channel
- **Archivo:** `pkg/channels/slack.go`
- **Severidad:** MEDIA
- **Problema:** `defer cancel()` dentro de un `for` loop no liberaba el contexto hasta que terminaba el loop completo.
- **Corrección:** Se reestructura para cancelar el contexto correctamente en cada iteración.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 24. QQ channel map eviction parcial causando leak
- **Archivo:** `pkg/channels/qq.go`
- **Severidad:** MEDIA
- **Problema:** La evicción parcial del mapa de sesiones no limpiaba todas las entradas expiradas.
- **Corrección:** Se reemplaza con reset completo del mapa.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

### Manejo de Errores - Corregidos

#### 25. Scanner.Err() no verificado en streaming HTTP
- **Archivo:** `pkg/providers/http_provider.go`
- **Severidad:** MEDIA
- **Problema:** Después de iterar con `scanner.Scan()`, no se verificaba `scanner.Err()` para detectar errores de lectura.
- **Corrección:** Se añade verificación de `scanner.Err()`.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 26. Scanner.Err() no verificado en Ollama streaming
- **Archivo:** `pkg/providers/ollama_provider.go`
- **Severidad:** MEDIA
- **Problema:** Igual que el anterior pero en el provider de Ollama.
- **Corrección:** Se añade campo `Error` a `StreamChunk` y verificación de `scanner.Err()`.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 27. Fallos de parseo JSON en SSE ignorados silenciosamente
- **Archivo:** `pkg/providers/http_provider.go`
- **Severidad:** BAJA
- **Problema:** Errores de parseo de chunks SSE se ignoraban sin log.
- **Corrección:** Se logean los fallos de parseo.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 28. Errores ignorados en importación de backup
- **Archivo:** `pkg/storage/backup.go`
- **Severidad:** MEDIA
- **Problema:** Errores de operaciones de base de datos en import se ignoraban.
- **Corrección:** Se registran errores con logger.WarnCF.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 29. Shell guard retornaba vacío en error de Abs
- **Archivo:** `pkg/tools/shell.go`
- **Severidad:** ALTA
- **Problema:** Si `filepath.Abs()` fallaba, `guardCommand` retornaba string vacío, permitiendo bypass de seguridad.
- **Corrección:** Se retorna error de bloqueo si `Abs()` falla.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 30. Session manager ignoraba errores de MkdirAll
- **Archivo:** `pkg/session/manager.go`
- **Severidad:** MEDIA
- **Problema:** Error al crear directorios de sesión se ignoraba.
- **Corrección:** Se propaga el error.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 31. OAuth io.ReadAll errores descartados
- **Archivo:** `pkg/auth/oauth.go`
- **Severidad:** MEDIA
- **Problema:** Errores de `io.ReadAll` en respuestas OAuth se ignoraban.
- **Corrección:** Se manejan los errores correctamente.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 32. Knowledge LastInsertId error no verificado
- **Archivo:** `pkg/storage/knowledge.go`
- **Severidad:** MEDIA
- **Problema:** El error de `LastInsertId()` se ignoraba, comprometiendo integridad de datos.
- **Corrección:** Se verifica el error.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

### Calidad de Código - Corregidos

#### 33. Type assertions inseguras en agent loop
- **Archivo:** `pkg/agent/loop.go`
- **Severidad:** MEDIA
- **Problema:** Type assertions sin verificación podían causar panic en runtime.
- **Corrección:** Se reemplazan con type assertions guardadas (`val, ok := ...`).
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 34. Acceso a array de mensajes vacío
- **Archivo:** `pkg/agent/loop.go`
- **Severidad:** MEDIA
- **Problema:** Se accedía a `messages[len(messages)-1]` sin verificar que el array no estuviera vacío.
- **Corrección:** Se añade verificación de longitud.
- **Commit:** `b845424`
- **Estado:** ~~CORREGIDO~~

#### 35. editFile pisaba permisos originales con 0644 hardcodeado
- **Archivo:** `pkg/tools/edit.go`
- **Severidad:** MEDIA
- **Problema:** Al escribir un archivo editado, siempre se usaba `0644`, ignorando los permisos originales.
- **Corrección:** Se preservan los bits de permiso originales del archivo.
- **Commit:** `dcbda1d`
- **Estado:** ~~CORREGIDO~~

#### 36. Sanitización de filenames en Content-Disposition
- **Archivo:** `pkg/web/handlers_advanced.go`, `pkg/web/handlers_backup.go`
- **Severidad:** MEDIA
- **Problema:** Filenames en headers HTTP no se sanitizaban, permitiendo inyección de headers.
- **Corrección:** Se sanitizan los nombres de archivo.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 37. Fallback JSON en workflow engine en fallo de marshal
- **Archivo:** `pkg/workflow/engine.go`
- **Severidad:** BAJA
- **Problema:** Si `json.Marshal` fallaba, no había fallback.
- **Corrección:** Se proporciona JSON de fallback en caso de error de serialización.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

#### 38. Escrituras redundantes racey en subagent goroutine
- **Archivo:** `pkg/tools/subagent.go`
- **Severidad:** BAJA
- **Problema:** La goroutine `runTask` escribía campos redundantes que ya se habían establecido, creando una race.
- **Corrección:** Se eliminan las escrituras redundantes.
- **Commit:** `5c9bdfe`
- **Estado:** ~~CORREGIDO~~

---

## PROBLEMAS PENDIENTES

### Seguridad - Pendientes

#### P1. Token JWT aceptado en query string
- **Archivo:** `pkg/web/server.go:543-548`
- **Severidad:** ALTA
- **Problema:** Se aceptan tokens JWT en parámetros de URL (`?token=...`) para endpoints `/ws/` y `/api/`. Los tokens en query strings se registran en logs de acceso, historial del navegador, y headers Referer, facilitando su exposición.
- **Recomendación:** Eliminar soporte de token por query string. Para WebSocket, enviar el token en el primer mensaje después de la conexión. Evaluar impacto en clientes existentes antes de eliminar.

#### P2. Allowlist de shell commands evadible con pipes/semicolons
- **Archivo:** `pkg/tools/shell.go:256-264`
- **Severidad:** ALTA
- **Problema:** `SetSafeCommandsForUser()` genera patrones regex `^\s*<cmd>\b` que solo validan el **inicio** del comando. Un usuario restringido puede evadir la restricción con: `echo safe; rm -rf /` o `ls | dangerous_cmd`.
- **Recomendación:** Implementar parsing real de comandos shell o prohibir pipes (`|`), semicolons (`;`), y operadores (`&&`, `||`) cuando se usa allowlist.

#### P3. WebSocket origin check siempre retorna true
- **Archivo:** `pkg/web/server.go:553-557`
- **Severidad:** MEDIA
- **Problema:** `checkWebSocketOrigin()` siempre retorna `true`, deshabilitando la protección CORS para WebSocket. Aunque se valida JWT, esto permite ataques CSRF de WebSocket si el token es obtenido por otro medio.
- **Recomendación:** Implementar validación de origen configurable, manteniendo JWT como segunda línea de defensa.

#### P4. CSP con 'unsafe-inline' para scripts y estilos
- **Archivo:** `pkg/web/server.go:453`
- **Severidad:** BAJA
- **Problema:** La política Content-Security-Policy incluye `'unsafe-inline'` tanto para scripts como para estilos, debilitando la protección contra XSS.
- **Recomendación:** Migrar a nonces o hashes para scripts inline. Para estilos, evaluar si es posible externalizar.

### Race Conditions - Pendientes

#### P5. MCP readLoop puede enviar a canal después de cleanup en sendRequest
- **Archivo:** `pkg/mcp/client.go:415-423`
- **Severidad:** MEDIA
- **Problema:** Hay un window entre `delete(c.pending, *resp.ID)` en readLoop (línea 418) y el `delete(c.pending, id)` en el defer de sendRequest (línea 335). Si readLoop encuentra la respuesta después de que sendRequest ha eliminado la entrada del mapa (por timeout), no hay race, pero si ambos ejecutan concurrentemente la lectura+delete, el channel send en línea 422 podría enviarse a un canal que nadie lee (buffered=1, así que no bloquea, pero el dato se pierde silenciosamente).
- **Recomendación:** Considerar no eliminar del mapa en readLoop, y dejar que solo sendRequest maneje el cleanup.

### Fugas de Recursos - Pendientes

#### P6. Goroutine leak en MCP client readLoop
- **Archivo:** `pkg/mcp/client.go:181, 391-425`
- **Severidad:** ALTA
- **Problema:** `readLoop()` se inicia como goroutine sin mecanismo de cancelación por contexto. Solo termina si la lectura del pipe retorna error. Si el proceso MCP queda colgado, la goroutine queda bloqueada indefinidamente en `ReadBytes()`.
- **Recomendación:** Pasar contexto a readLoop. Usar `context.Done()` con un select, o asegurar que `Close()` siempre cierra el pipe de stdout para desbloquear la lectura.

#### P7. Goroutines con context.Background() que no se cancelan
- **Archivos:** `pkg/web/server.go:3458-3466`, `pkg/agent/loop.go:1429-1432`, `pkg/observability/metrics.go:200-203`
- **Severidad:** MEDIA
- **Problema:** Múltiples goroutines se lanzan con `context.Background()` o sin contexto. No se cancelan cuando el servidor se apaga, potencialmente dejando operaciones huérfanas.
- **Recomendación:** Propagar el contexto del servidor/parent a las goroutines lanzadas. Usar `context.WithTimeout` para operaciones que no deban durar indefinidamente.

### Manejo de Errores - Pendientes

#### P8. rows.Err() no verificado en backup export
- **Archivo:** `pkg/storage/backup.go:108-117, 132-141, 156-167`
- **Severidad:** MEDIA
- **Problema:** Después de iterar `rows.Next()` en las tres consultas de exportación (sessions, messages, tasks), no se verifica `rows.Err()`. Si ocurre un error durante la iteración, los datos parciales se retornan sin indicación de que están incompletos.
- **Recomendación:** Añadir `if err := rows.Err(); err != nil { ... }` después de cada loop de iteración.

#### P9. Error de json.Unmarshal ignorado en import options
- **Archivo:** `pkg/web/handlers_backup.go:416`
- **Severidad:** MEDIA
- **Problema:** `json.Unmarshal([]byte(body), &importOptions)` ignora silenciosamente errores de parseo, causando que opciones inválidas se traten como defaults.
- **Recomendación:** Verificar el error y retornar `400 Bad Request` si el JSON es inválido.

#### P10. Errores silenciosos en queries de backup export
- **Archivo:** `pkg/storage/backup.go:108, 132, 156`
- **Severidad:** MEDIA
- **Problema:** Si las queries de exportación fallan (`err != nil`), se ignora silenciosamente y se retorna datos vacíos sin indicar error.
- **Recomendación:** Al menos logear los errores, o retornarlos al llamador.

### Calidad de Código - Pendientes

#### P11. Directorios temporales predecibles
- **Archivos:** `cmd/makoclaw/main.go:638`, `pkg/utils/media.go:67`, `pkg/channels/signal.go:262`
- **Severidad:** MEDIA
- **Problema:** Se usan rutas fijas en `/tmp/` (`makoclaw_history`, `makoclaw_media`, `makoclaw-signal`). En sistemas multi-usuario, otro usuario podría pre-crear estos directorios o symlinks maliciosos.
- **Recomendación:** Usar `os.MkdirTemp()` para directorios, y añadir componentes aleatorios a los nombres de archivo. Para el historial CLI, usar el directorio home del usuario.

#### P12. Sin recovery de panics en HTTP handlers
- **Archivos:** Todos los handlers en `pkg/web/handlers_*.go`
- **Severidad:** ALTA
- **Problema:** No hay middleware de recovery. Un panic en cualquier handler HTTP crashea todo el proceso del servidor web.
- **Recomendación:** Añadir un middleware `recoveryMiddleware` que capture panics con `defer recover()`, logee el stack trace, y retorne `500 Internal Server Error`.

#### P13. Sin rate limiting en endpoints API (excepto login)
- **Archivo:** `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** Solo el endpoint de login tiene rate limiting. Endpoints como `/api/v1/chat`, `/api/v1/tasks`, `/api/v1/knowledge` no tienen límite, permitiendo abuso.
- **Recomendación:** Implementar rate limiting global por usuario/IP para endpoints autenticados. Considerar un middleware basado en token bucket.

#### P14. Mensajes de error HTTP exponen detalles internos
- **Archivos:** `pkg/web/handlers_features.go` y otros handlers
- **Severidad:** BAJA
- **Problema:** Mensajes de error como `http.Error(w, "failed to create prompt: "+err.Error(), 500)` pueden contener detalles de implementación o rutas del sistema.
- **Recomendación:** Logear el error internamente y retornar un mensaje genérico al usuario. Usar JSON consistente para respuestas de error.

#### P15. Acumulación de mensajes sin límite en sesiones
- **Archivo:** `pkg/session/manager.go`
- **Severidad:** MEDIA
- **Problema:** Los mensajes de sesión se acumulan sin límite estricto. La sumarización existe pero se ejecuta asíncronamente y puede no ser suficiente para sesiones de muy larga duración.
- **Recomendación:** Implementar un hard limit de mensajes por sesión, truncando los más antiguos cuando se excede.

---

## Prioridades de Acción

### Inmediato (Seguridad)
1. **P2** - Fix de allowlist evadible (ALTA)
2. **P12** - Recovery middleware para panics (ALTA)
3. **P6** - Fix de goroutine leak en MCP (ALTA)

### Corto Plazo
4. **P1** - Migrar token JWT fuera de query string (ALTA)
5. **P8/P10** - Verificar rows.Err() y errores en backup (MEDIA)
6. **P13** - Rate limiting en API endpoints (MEDIA)

### Mediano Plazo
7. **P3** - WebSocket origin validation (MEDIA)
8. **P7** - Propagación de contexto a goroutines (MEDIA)
9. **P11** - Directorios temporales seguros (MEDIA)
10. **P5/P9** - Mejoras menores de manejo de errores (MEDIA)

### Bajo Prioridad
11. **P4** - Tightening de CSP (BAJA)
12. **P14** - Sanitización de mensajes de error HTTP (BAJA)
13. **P15** - Hard limit en mensajes de sesión (MEDIA-BAJA)

---

## Commits de Corrección

| Commit | Descripción | Archivos |
|--------|-------------|----------|
| `b845424` | Fix security bugs, race conditions, error handling | 18 archivos |
| `5c9bdfe` | Fix race conditions, resource leaks, error handling | 11 archivos |
| `dcbda1d` | Fix critical security vulnerabilities and data integrity | 6 archivos |
| `65c7b55` | Fix symlink exfiltration, spawn race, email validation | 4 archivos |

**Total:** 33 archivos modificados, 620 líneas añadidas, 209 eliminadas.

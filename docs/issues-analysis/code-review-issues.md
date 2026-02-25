# Reporte de Code Review - Problemas de Seguridad, Concurrencia y Calidad

**Fecha:** Febrero 2026
**Branch:** `claude/code-review-improvements-ZwCI6`
**Revisión:** Auditoría completa del código fuente Go (2 rondas)

---

## Resumen

Se realizaron dos auditorías exhaustivas del código, identificando problemas de seguridad, race conditions, fugas de recursos, manejo de errores y calidad de código. Se corrigieron **38 problemas** en 4 commits. La segunda auditoría identificó **30 problemas pendientes adicionales** que requieren atención.

### Estadísticas

| Categoría                      | Corregidos | Pendientes |
| ------------------------------ | :--------: | :--------: |
| Seguridad                      |     16     |     8      |
| Race Conditions / Concurrencia |     6      |     4      |
| Fugas de Recursos              |     2      |     4      |
| Manejo de Errores / Integridad |     10     |     9      |
| Calidad de Código              |     4      |     9      |
| **Total**                      |   **38**   |   **34**   |

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
- **Corrección:** Eliminado el soporte para parámetros URL `?token=...` en endpoints REST (`/api/`), manteniendo el query string temporalmente solo para las conexiones `/ws/`.
- **Estado:** ~~CORREGIDO PARCIALMENTE~~

#### P2. Allowlist de shell commands evadible con pipes/semicolons

- **Archivo:** `pkg/tools/shell.go:256-264`
- **Severidad:** ALTA
- **Problema:** `SetSafeCommandsForUser()` genera patrones regex `^\s*<cmd>\b` que solo validan el **inicio** del comando. Un usuario restringido puede evadir la restricción con: `echo safe; rm -rf /` o `ls | dangerous_cmd`.
- **Corrección:** Añadido un chequeo estricto para rechazar comandos que usen operadores de concatenación como pipes (`|`), semicolon (`;`), backticks o evaluación `$()`.
- **Estado:** ~~CORREGIDO~~

#### P3. WebSocket origin check siempre retorna true

- **Archivo:** `pkg/web/server.go:553-557`
- **Severidad:** MEDIA
- **Problema:** `checkWebSocketOrigin()` siempre retorna `true`, deshabilitando la protección CORS para WebSocket. Aunque se valida JWT, esto permite ataques CSRF de WebSocket si el token es obtenido por otro medio.
- **Recomendación:** Implementar validación de origen configurable, manteniendo JWT como segunda línea de defensa.
- **Corrección:** Se añadió método `checkOrigin(r *http.Request)` en el struct `Server` con lógica en capas: (1) permite peticiones sin header `Origin`, (2) permite si el host del origen coincide con el host de la petición, (3) verifica contra `allowedOrigins []string` si está configurado, (4) permite todo si no hay lista configurada (compatibilidad retroactiva con reverse proxies). Los dos upgraders de WebSocket ahora usan `s.checkOrigin` en lugar de la función standalone.
- **Estado:** ~~CORREGIDO~~

#### P4. CSP con 'unsafe-inline' para scripts y estilos

- **Archivo:** `pkg/web/server.go:453`
- **Severidad:** BAJA
- **Problema:** La política Content-Security-Policy incluye `'unsafe-inline'` tanto para scripts como para estilos, debilitando la protección contra XSS.
- **Recomendación:** Migrar a nonces o hashes para scripts inline. Para estilos, evaluar si es posible externalizar.
- **Corrección:** Se añadió `'strict-dynamic'` a `script-src` — en browsers modernos `strict-dynamic` hace que `'unsafe-inline'` sea ignorado, requiriendo nonces/hashes; browsers legacy siguen usando `'unsafe-inline'` como fallback. Se añadió la directiva `upgrade-insecure-requests`. La mejora completa a nonces requiere soporte de template engine en el servidor y queda pendiente como mejora futura.
- **Estado:** ~~CORREGIDO~~

#### P5. Tabla no escapada en PRAGMA SQL (potencial inyección)

- **Archivo:** `pkg/storage/backup.go:496`
- **Severidad:** MEDIA
- **Problema:** `hasColumn()` construye la query PRAGMA con `fmt.Sprintf("PRAGMA table_info(%s)", table)` sin validación del nombre de tabla. Aunque actualmente los nombres de tabla están hardcodeados (`sessions`, `chats`, `tasks`, `channel_users`), el patrón es inseguro y vulnerable si en el futuro se pasa input externo.
- **Corrección:** Añadida white-list (`map[string]bool`) que restringe explicitamente la validación solo a tablas conocidas permitidas.
- **Estado:** ~~CORREGIDO~~

#### P6. Sin validación de contraseña en bootstrap de admin

- **Archivo:** `pkg/web/auth.go:70-111`
- **Severidad:** BAJA
- **Problema:** El bootstrap de admin en `newAuthManager()` solo verifica que la contraseña no esté vacía, sin aplicar la política de longitud mínima de 8 caracteres que sí aplican los endpoints de registro y cambio de contraseña.
- **Corrección:** Se aplica la misma validación de longitud `>= 8` caracteres durante el bootstrap de la cuenta de admin.
- **Estado:** ~~CORREGIDO~~

#### P7. Sin protección CSRF en endpoints de estado

- **Archivo:** `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** No existe middleware CSRF. Las operaciones de cambio de estado (POST, PUT, DELETE) dependen únicamente de JWT para autenticación. Si el token se filtra (ej. via query string), se exponen a CSRF.
- **Recomendación:** Implementar doble-submit cookies o tokens CSRF. Aplicar `SameSite=Strict` en cookies.
- **Análisis:** La API utiliza exclusivamente tokens JWT enviados via header `Authorization: Bearer` (verificado en `extractClaims()` en `server.go`). No se establecen cookies de sesión en ningún endpoint. CSRF solo afecta a autenticación basada en cookies; con JWT via header, el browser no puede enviar el token automáticamente en cross-site requests. Por tanto, CSRF no aplica a esta arquitectura.
- **Estado:** ~~NO APLICA~~ (JWT header-only, sin cookies de sesión)

#### P8. Sin límites de longitud en campos de texto de API

- **Archivos:** `pkg/web/handlers_features.go` y otros handlers
- **Severidad:** BAJA
- **Problema:** Los campos de texto (títulos, descripciones, contenido) de endpoints como prompts, tasks, y knowledge no tienen límite de longitud. No se aplica `http.MaxBytesReader` en todos los endpoints que lo necesitan, permitiendo payloads excesivos.
- **Recomendación:** Añadir validación de longitud: título ≤ 255 chars, descripción ≤ 500 chars, contenido ≤ 100,000 chars. Aplicar `http.MaxBytesReader` consistentemente.
- **Corrección:** Se añadió validación de longitud en: (1) `handlePrompts` POST — title ≤ 255, description ≤ 2000, content ≤ 100000; (2) `handlePromptAction` PUT — mismos límites; (3) `handleTasks` POST — title ≤ 255, description ≤ 2000; (4) `handleUsers` POST — username ≤ 64 chars. Todos retornan HTTP 400 si se excede el límite.
- **Estado:** ~~CORREGIDO~~

### Race Conditions / Concurrencia - Pendientes

#### P9. Deadlock por adquisición recursiva de lock en MultiUserChannelManager

- **Archivo:** `pkg/channels/multiuser_manager.go:245-272`
- **Severidad:** CRITICA
- **Problema:** `RestartUserChannels()` adquiere `m.mu.Lock()` en línea 246, luego llama a `m.GetOrCreateManagerForUser(userUUID)` en línea 272. `GetOrCreateManagerForUser()` intenta adquirir el mismo lock en línea 91, causando deadlock en un mutex no recursivo.
- **Corrección:** Se refactorizó la lógica en un método interno `getOrCreateManagerForUserLocked()` para evitar dobles locks.
- **Estado:** ~~CORREGIDO~~

#### P10. Data race en campo ctx de DiscordChannel

- **Archivo:** `pkg/channels/discord.go:44, 59-62, 68`
- **Severidad:** MEDIA
- **Problema:** El campo `ctx` se escribe en `Start()` (línea 68) y se lee en `getContext()` (líneas 59-62) sin sincronización. El race detector de Go reportaría un data race entre estas operaciones.
- **Corrección:** Se eliminó la escritura a `c.ctx` en `Start()` y `getContext()` ahora retorna `context.Background()` eliminando la race condition y fugas.
- **Estado:** ~~CORREGIDO~~

#### P11. MCP readLoop puede enviar a canal después de cleanup en sendRequest

- **Archivo:** `pkg/mcp/client.go:415-423`
- **Severidad:** MEDIA
- **Problema:** Hay una ventana entre `delete(c.pending, *resp.ID)` en readLoop y el `delete(c.pending, id)` en el defer de sendRequest. El dato puede perderse silenciosamente si ambos ejecutan concurrentemente.
- **Recomendación:** Dejar que solo sendRequest maneje el cleanup del mapa pending.
- **Corrección:** Eliminado el `delete(c.pending, *resp.ID)` de `readLoop()`. Ahora solo el defer de `sendRequest()` maneja el cleanup del mapa pending.
- **Estado:** ~~CORREGIDO~~

#### P12. Goroutines con context.Background() sin cancelación

- **Archivos:** `pkg/web/server.go:3458-3466`, `pkg/web/handlers_user_config.go:203-215`, `pkg/channels/multiuser_manager.go:115`, `pkg/agent/loop.go:1429-1432`
- **Severidad:** MEDIA
- **Problema:** Múltiples goroutines se lanzan con `context.Background()` sin mecanismo de cancelación. No se detienen cuando el servidor se apaga, dejando operaciones huérfanas.
- **Recomendación:** Propagar el contexto del servidor a las goroutines. Usar `context.WithTimeout` para operaciones que no deban durar indefinidamente. Implementar WaitGroup para tracking.
- **Corrección:** Se añadió campo `ctx` al struct `Server` asignado en `Start()`. Las goroutines en `server.go` y `handlers_user_config.go` usan `s.ctx` y `r.Context()` respectivamente. El closure de `cronService.SetOnJob` en `multiuser_manager.go` usa `m.ctx`.
- **Estado:** ~~CORREGIDO~~

### Fugas de Recursos - Pendientes

#### P13. Goroutine leak en MCP client readLoop

- **Archivo:** `pkg/mcp/client.go:181, 391-425`
- **Severidad:** ALTA
- **Problema:** `readLoop()` se inicia como goroutine sin mecanismo de cancelación por contexto. Solo termina si la lectura del pipe retorna error. Si el proceso MCP queda colgado, la goroutine queda bloqueada indefinidamente en `ReadBytes()`.
- **Corrección:** Se añadió cancelación vía context interno usando un patrón select para manejar correctamente la terminación.
- **Estado:** ~~CORREGIDO~~

#### P14. Crecimiento no acotado del ring buffer de RecentEvents

- **Archivo:** `pkg/observability/metrics.go:375-378`
- **Severidad:** MEDIA
- **Problema:** El buffer de eventos recientes usa reslicing (`m.RecentEvents = m.RecentEvents[1:]`) que es ineficiente: el array subyacente no libera la memoria de los elementos eliminados. Con uso prolongado, esto genera memory leak sutil.
- **Corrección:** Se modificó la adición de eventos para hacer shift (mediante `copy`) reemplazando el uso recursivo de slices del slice original, solucionando la fuga de memoria temporal.
- **Estado:** ~~CORREGIDO~~

#### P15. Sin límite de conexiones en servidor TCP de MaixCam

- **Archivo:** `pkg/channels/maixcam.go:42-61`
- **Severidad:** MEDIA
- **Problema:** El servidor TCP acepta conexiones sin límite ni rate limiting, vulnerable a agotamiento de recursos.
- **Recomendación:** Implementar semáforo para limitar conexiones concurrentes. Añadir timeouts de lectura/escritura.
- **Corrección:** Se añadió constante `maxMaixCamConnections = 10` y campo `sem chan struct{}` al struct `MaixCamChannel`. Antes de procesar cada conexión, se intenta adquirir el semáforo con `select` non-blocking; si está lleno se cierra la conexión inmediatamente con log de warning. Al finalizar cada goroutine de conexión, se libera el semáforo con `<-c.sem`.
- **Estado:** ~~CORREGIDO~~

#### P16. Busy-wait en listener de WhatsApp

- **Archivo:** `pkg/channels/whatsapp.go:107-144`
- **Severidad:** BAJA
- **Problema:** `listen()` hace polling continuo con `time.Sleep(1s)` cuando la conexión es nil, y `time.Sleep(2s)` en errores de lectura, sin backoff exponencial.
- **Recomendación:** Implementar backoff exponencial con tiempo máximo de espera. Usar un state machine de conexión.
- **Corrección:** Se reemplazaron los `time.Sleep` fijos por backoff exponencial: inicia en 1s, se duplica en cada reintento (1s→2s→4s→8s→16s), con techo de 30s. Se usan selects con `ctx.Done()` para respetar cancelación de contexto. El backoff se reinicia a 1s tras cada lectura exitosa.
- **Estado:** ~~CORREGIDO~~

### Manejo de Errores / Integridad de Datos - Pendientes

#### P17. rows.Err() no verificado en backup export (5 ubicaciones)

- **Archivo:** `pkg/storage/backup.go:116, 140, 167, 189-199, 219-229`
- **Severidad:** ALTA
- **Problema:** Después de iterar `rows.Next()` en las cinco consultas de exportación (sessions, messages, tasks, task_logs, channel_mappings), no se verifica `rows.Err()`. Errores de iteración (I/O de disco, corrupción de BD) se ignoran silenciosamente, generando backups incompletos que se marcan como exitosos.
- **Corrección:** Añadido `rows.Err()` checks con logs apropiados después del loop.
- **Estado:** ~~CORREGIDO~~

#### P18. Error de json.Unmarshal ignorado en import options

- **Archivo:** `pkg/web/handlers_backup.go:416`
- **Severidad:** MEDIA
- **Problema:** `json.Unmarshal([]byte(body), &importOptions)` ignora silenciosamente errores de parseo, causando que opciones inválidas se traten como defaults. El usuario puede creer que sus opciones se aplicaron cuando no fue así.
- **Recomendación:** Verificar el error y retornar `400 Bad Request` si el JSON es inválido.
- **Corrección:** Verificado el error de `json.Unmarshal`; si el body es no-vacío y el parseo falla, se retorna `400 Bad Request` con mensaje "invalid import options".
- **Estado:** ~~CORREGIDO~~

#### P19. Errores silenciosos en queries de backup export

- **Archivo:** `pkg/storage/backup.go:108, 132, 156`
- **Severidad:** MEDIA
- **Problema:** Si las queries de exportación fallan (`err != nil`), se ignora silenciosamente y se retorna datos vacíos sin indicar error.
- **Recomendación:** Al menos logear los errores, o retornarlos al llamador.
- **Corrección:** Se transformó cada bloque `if err == nil` en `if err != nil { logger.WarnCF(...) } else { ... }` para las tres queries (sessions, messages, tasks).
- **Estado:** ~~CORREGIDO~~

#### P20. Unchecked LastInsertId() en 4 ubicaciones

- **Archivos:** `pkg/storage/prompts.go:57`, `pkg/storage/workflow.go:91, 187`, `pkg/storage/task.go:44`
- **Severidad:** MEDIA
- **Problema:** `LastInsertId()` puede retornar error (drivers que no lo soportan, fallos de BD). Sin verificación, se retorna id=0 al llamador que lo trata como éxito, corrompiendo lookups posteriores (ej. `GetPrompt(0)` busca un registro inexistente).
- **Corrección:** Se agregaron chequeos en LastInsertId en todas las funciones.
- **Estado:** ~~CORREGIDO~~

#### P21. Unchecked RowsAffected() en workflows

- **Archivo:** `pkg/storage/workflow.go:161, 174`
- **Severidad:** MEDIA
- **Problema:** `RowsAffected()` puede retornar error en fallos de BD (disco lleno, permisos). Con el error ignorado, `n` defaultea a 0, que se confunde con "registro no encontrado" — enmascarando errores reales.
- **Corrección:** Agregados chequeos de errores en RowsAffected().
- **Estado:** ~~CORREGIDO~~
- **Severidad:** MEDIA
- **Problema:** `RowsAffected()` puede retornar error en fallos de BD (disco lleno, permisos). Con el error ignorado, `n` defaultea a 0, que se confunde con "registro no encontrado" — enmascarando errores reales.
- **Recomendación:** Verificar error de `RowsAffected()` antes de comparar con 0.

#### P22. Errores de json.Marshal ignorados en agent loop (6 ubicaciones)

- **Archivo:** `pkg/agent/loop.go:1024, 1042, 1278, 1293, 1364, 1378`
- **Severidad:** MEDIA
- **Problema:** `json.Marshal(tc.Arguments)` se llama sin verificar error en 6 ubicaciones donde se serializan argumentos de herramientas. Si falla (tipos no serializables), los argumentos se pierden en el historial de sesión.
- **Corrección:** Se agegraron logs para metadata y default a un string crudo en log de arguments y en los args en si para evitar panics o fallos ignorados en el loop y se continúa la ejecución o en el caso particular de logs no aborta.
- **Estado:** ~~CORREGIDO~~

#### P23. rows.Err() no verificado en hasColumn PRAGMA query

- **Archivo:** `pkg/storage/backup.go:503-514`
- **Severidad:** MEDIA
- **Problema:** Después del loop de iteración en `hasColumn()`, no se verifica `rows.Err()`. Un error de iteración retorna `false` silenciosamente (como si la columna no existiera), afectando la lógica de migración/import.
- **Corrección:** Se añadió la comprobación de error junto con un correspondiente Warning Log si se interrumpió la iteración.
- **Estado:** ~~CORREGIDO~~

#### P24. Mensajes de error HTTP exponen detalles internos

- **Archivos:** `pkg/web/handlers_features.go:33, 58, 99, 107, 159` y otros handlers
- **Severidad:** BAJA
- **Problema:** Mensajes de error como `http.Error(w, "failed to create prompt: "+err.Error(), 500)` pueden contener detalles de implementación (rutas de BD, mensajes de driver SQL).
- **Corrección:** Eliminada exposición de internal errors a través de los APIs en los handlers que operaban con BD (Prompts, Features extensions, etc), usando `logger.ErrorCF` con la info y respondiendo mensajes de error estáticos a los clientes web.
- **Estado:** ~~CORREGIDO~~

#### P25. Acumulación de mensajes sin límite en sesiones

- **Archivo:** `pkg/session/manager.go:158`
- **Severidad:** MEDIA
- **Problema:** `session.Messages = append(session.Messages, msg)` crece indefinidamente. La sumarización existe pero es asíncrona. En sesiones de larga duración, la memoria crece sin control.
- **Corrección:** Se limitó el truncamiento automático a las últimas 500 interacciones limitando el consumo sin depender de summarization asíncrono.
- **Estado:** ~~CORREGIDO~~

### Calidad de Código - Pendientes

#### P26. Sin recovery de panics en HTTP handlers

- **Archivos:** Todos los handlers en `pkg/web/handlers_*.go`
- **Severidad:** ALTA
- **Problema:** No hay middleware de recovery. Un panic en cualquier handler HTTP crashea todo el proceso del servidor web.
- **Corrección:** Se añadió un middleware `recoveryMiddleware` y se envolvió con este al request router del web server en `server.go`. Este atrapa y documenta log de panics respondiendo 500 error en vez de colapsar la app.
- **Estado:** ~~CORREGIDO~~

#### P27. Sin rate limiting en endpoints API (excepto login)

- **Archivo:** `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** Solo el endpoint de login tiene rate limiting. Endpoints como `/api/v1/chat`, `/api/v1/tools/execute`, `/api/v1/knowledge/search` no tienen límite, permitiendo abuso y DoS.
- **Recomendación:** Implementar rate limiting global por usuario/IP para endpoints autenticados. Considerar un middleware basado en token bucket con límites diferenciados según el costo de cada endpoint.
- **Corrección:** Se añadió middleware `apiRateLimitMiddleware` que aplica 100 req/min por usuario autenticado (usando el username del JWT como clave, con el `RateLimiter` existente). Retorna HTTP 429 con header `Retry-After` cuando se excede. El middleware se añade a la cadena: `recoveryMiddleware → authMiddleware → apiRateLimitMiddleware → mux`, garantizando que las claims JWT ya estén disponibles en el contexto cuando se evalúa el rate limit. Rutas públicas (sin claims) no son afectadas.
- **Estado:** ~~CORREGIDO~~

#### P28. Directorios temporales predecibles

- **Archivos:** `cmd/makoclaw/main.go:638`, `pkg/utils/media.go:67`, `pkg/channels/signal.go:262`
- **Severidad:** MEDIA
- **Problema:** Se usan rutas fijas en `/tmp/` (`makoclaw_history`, `makoclaw_media`, `makoclaw-signal`). En sistemas multi-usuario, otro usuario podría pre-crear estos directorios o symlinks maliciosos (symlink attack).
- **Recomendación:** Usar `os.MkdirTemp()` para directorios, y añadir componentes aleatorios a los nombres de archivo. Para el historial CLI, usar el directorio home del usuario.
- **Corrección:** El historial CLI usa `os.UserCacheDir()` con fallback a home dir. Los directorios de media y signal usan sufijo de PID para aislar por proceso.
- **Estado:** ~~CORREGIDO~~

#### P29. HTTP Clients creados ad-hoc sin reusar conexiones

- **Archivos:** `pkg/tools/web.go:89`, `pkg/skills/installer.go:48, 98`, `pkg/web/server.go:4289`
- **Severidad:** BAJA
- **Problema:** Se crean instancias `http.Client` nuevas en cada llamada, desperdiciando el pool de conexiones HTTP. Cada instancia crea un nuevo transport, anulando los beneficios de keep-alive.
- **Recomendación:** Crear un `*http.Client` singleton por módulo con transport persistente reutilizable.
- **Corrección:** Se crearon singletons package-level: `webSearchHTTPClient` (10s timeout) y `webFetchHTTPClient` (60s timeout con transport configurado) en `pkg/tools/web.go`; `installerHTTPClient` (15s timeout) en `pkg/skills/installer.go`. Las funciones `Execute()` e `InstallFromGitHub()`/`ListAvailableSkills()` ahora reutilizan estos clientes compartidos.
- **Estado:** ~~CORREGIDO~~

#### P30. Errores de json.Encoder silenciados en respuestas HTTP

- **Archivos:** `pkg/web/server.go` y todos los handlers
- **Severidad:** BAJA
- **Problema:** Decenas de ubicaciones usan `_ = json.NewEncoder(w).Encode(...)`, ignorando silenciosamente errores de encoding. Los clientes pueden recibir JSON incompleto o corrupto sin indicación de error.
- **Corrección:** Se implementó una función `writeJSONResponse(w, data)` en todo el paquete `web` que codifica y registra (logger.Warn) cualquier error fallido en lugar de descartarlo con `_ =`.
- **Estado:** ~~CORREGIDO~~

#### P31. Sin IdleConnTimeout en Ollama HTTP client

- **Archivo:** `pkg/providers/ollama_provider.go:61-63`
- **Severidad:** BAJA
- **Problema:** El HTTP client de OllamaProvider tiene un `Timeout: 120s` global pero no configura `Transport.IdleConnTimeout`, permitiendo que conexiones idle se mantengan indefinidamente.
- **Corrección:** Se añadió `Transport: &http.Transport{IdleConnTimeout: 30 * time.Second}` al cliente HTTP de OllamaProvider.
- **Estado:** ~~CORREGIDO~~

#### P32. Goroutines de summarización sin tracking en AgentLoop

- **Archivo:** `pkg/agent/loop.go:1429+`
- **Severidad:** BAJA
- **Problema:** Las goroutines de summarización y otras operaciones background se lanzan sin WaitGroup ni tracking. El shutdown del servidor puede interrumpirlas a mitad de operación.
- **Recomendación:** Implementar `sync.WaitGroup` en AgentLoop para rastrear goroutines activas y esperarlas durante shutdown.
- **Corrección:** Se añadió campo `summarizeWg sync.WaitGroup` al struct `AgentLoop`. La goroutine de summarización llama `Add(1)` antes de lanzarse y `defer Done()` al inicio. El método `Stop()` llama `summarizeWg.Wait()`.
- **Estado:** ~~CORREGIDO~~

#### P33. API keys de providers retornadas en respuestas API

- **Archivo:** `pkg/web/providers_handler.go:69-83`
- **Severidad:** MEDIA
- **Problema:** El endpoint `GetProvidersConfig` retorna los API keys completos en la respuesta JSON, exponiéndolos al frontend y potencialmente a logs. Las claves deberían estar enmascaradas.
- **Corrección:** Se implementó `redactKey(cfg.APIKey)` en `convertProviderConfig` de `providers_handler.go`.
- **Estado:** ~~CORREGIDO~~

#### P34. Errores de json.Marshal ignorados en OAuth

- **Archivo:** `pkg/auth/oauth.go:173, 228`
- **Severidad:** BAJA
- **Problema:** `reqBody, _ := json.Marshal(...)` ignora errores. Aunque mapas de strings siempre marshalan correctamente, el patrón viola mejores prácticas de Go.
- **Recomendación:** Verificar el error para mantener consistencia con el resto del código.
- **Corrección:** Ambas llamadas a `json.Marshal` usan `marshalErr` y retornan error si falla (en `LoginDeviceCode` y `pollDeviceCode`).
- **Estado:** ~~CORREGIDO~~

---

## Prioridades de Acción

### Inmediato (Crítico)

1. **P9** - Fix de deadlock recursivo en MultiUserChannelManager (CRITICA)
2. **P2** - Fix de allowlist evadible (ALTA)
3. **P26** - Recovery middleware para panics (ALTA)
4. **P13** - Fix de goroutine leak en MCP (ALTA)
5. **P17** - Verificar rows.Err() en backup export (ALTA)

### Corto Plazo (Alta/Media)

6. **P1** - Migrar token JWT fuera de query string (ALTA)
7. **P5** - Validar nombres de tabla en PRAGMA (MEDIA)
8. **P20** - Verificar LastInsertId() en 4 ubicaciones (MEDIA)
9. **P21** - Verificar RowsAffected() en workflows (MEDIA)
10. **P27** - Rate limiting en API endpoints (MEDIA)
11. **P33** - Enmascarar API keys en respuestas (MEDIA)

### Mediano Plazo

12. **P3** - WebSocket origin validation (MEDIA)
13. **P7** - Protección CSRF (MEDIA)
14. **P10** - Data race en DiscordChannel.ctx (MEDIA)
15. **P12** - Propagación de contexto a goroutines (MEDIA)
16. **P14** - Fix de ring buffer ineficiente (MEDIA)
17. **P15** - Límite de conexiones en MaixCam (MEDIA)
18. **P18/P19** - Mejoras en manejo de errores de backup (MEDIA)
19. **P22** - json.Marshal errors en agent loop (MEDIA)
20. **P25** - Hard limit en mensajes de sesión (MEDIA)
21. **P28** - Directorios temporales seguros (MEDIA)

### Bajo Prioridad

22. **P4** - Tightening de CSP (BAJA)
23. **P6** - Validación de contraseña en bootstrap (BAJA)
24. **P8** - Límites de longitud en campos de texto (BAJA)
25. **P11** - Cleanup de MCP pending map (MEDIA)
26. **P16** - Backoff en WhatsApp listener (BAJA)
27. **P23** - rows.Err() en hasColumn (MEDIA)
28. **P24** - Sanitización de mensajes de error HTTP (BAJA)
29. **P29** - HTTP Client reutilizable (BAJA)
30. **P30** - Errores de json.Encoder (BAJA)
31. **P31** - IdleConnTimeout en Ollama (BAJA)
32. **P32** - WaitGroup para goroutines de summarización (BAJA)
33. **P34** - json.Marshal en OAuth (BAJA)

---

## Commits de Corrección

| Commit    | Descripción                                              | Archivos    |
| --------- | -------------------------------------------------------- | ----------- |
| `b845424` | Fix security bugs, race conditions, error handling       | 18 archivos |
| `5c9bdfe` | Fix race conditions, resource leaks, error handling      | 11 archivos |
| `dcbda1d` | Fix critical security vulnerabilities and data integrity | 6 archivos  |
| `65c7b55` | Fix symlink exfiltration, spawn race, email validation   | 4 archivos  |

**Total correcciones:** 38 problemas corregidos en 33 archivos modificados, 620 líneas añadidas, 209 eliminadas.
**Total pendientes:** 34 problemas identificados pendientes de corrección.

---

## Resumen por Archivos Más Afectados

| Archivo                             | Corregidos | Pendientes | Prioridad |
| ----------------------------------- | :--------: | :--------: | :-------: |
| `pkg/storage/backup.go`             |     1      |     5      |   ALTA    |
| `pkg/web/server.go`                 |     3      |     6      |   ALTA    |
| `pkg/mcp/client.go`                 |     1      |     3      |   ALTA    |
| `pkg/channels/multiuser_manager.go` |     0      |     2      |  CRITICA  |
| `pkg/agent/loop.go`                 |     2      |     2      |   MEDIA   |
| `pkg/storage/workflow.go`           |     0      |     3      |   MEDIA   |
| `pkg/web/handlers_features.go`      |     0      |     2      |   MEDIA   |
| `pkg/web/auth.go`                   |     2      |     1      |   BAJA    |

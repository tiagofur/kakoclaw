# Reporte de Code Review - Problemas de Seguridad, Concurrencia y Calidad

**Fecha:** Febrero 2026
**Branch:** `claude/code-review-improvements-ZwCI6`
**Revisión:** Auditoría completa del código fuente Go (2 rondas)

---

## Resumen

Se realizaron dos auditorías exhaustivas del código, identificando problemas de seguridad, race conditions, fugas de recursos, manejo de errores y calidad de código. Se corrigieron **38 problemas** en 4 commits. La segunda auditoría identificó **30 problemas pendientes adicionales** que requieren atención.

### Estadísticas

| Categoría | Corregidos | Pendientes |
|-----------|:----------:|:----------:|
| Seguridad | 16 | 8 |
| Race Conditions / Concurrencia | 6 | 4 |
| Fugas de Recursos | 2 | 4 |
| Manejo de Errores / Integridad | 10 | 9 |
| Calidad de Código | 4 | 9 |
| **Total** | **38** | **34** |

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

#### P5. Tabla no escapada en PRAGMA SQL (potencial inyección)
- **Archivo:** `pkg/storage/backup.go:496`
- **Severidad:** MEDIA
- **Problema:** `hasColumn()` construye la query PRAGMA con `fmt.Sprintf("PRAGMA table_info(%s)", table)` sin validación del nombre de tabla. Aunque actualmente los nombres de tabla están hardcodeados (`sessions`, `chats`, `tasks`, `channel_users`), el patrón es inseguro y vulnerable si en el futuro se pasa input externo.
- **Recomendación:** Implementar whitelist de tablas permitidas:
  ```go
  allowedTables := map[string]bool{"sessions": true, "chats": true, "tasks": true, "channel_users": true}
  if !allowedTables[table] { return false }
  ```

#### P6. Sin validación de contraseña en bootstrap de admin
- **Archivo:** `pkg/web/auth.go:70-111`
- **Severidad:** BAJA
- **Problema:** El bootstrap de admin en `newAuthManager()` solo verifica que la contraseña no esté vacía, sin aplicar la política de longitud mínima de 8 caracteres que sí aplican los endpoints de registro y cambio de contraseña.
- **Recomendación:** Validar `len(password) >= 8` en el bootstrap también.

#### P7. Sin protección CSRF en endpoints de estado
- **Archivo:** `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** No existe middleware CSRF. Las operaciones de cambio de estado (POST, PUT, DELETE) dependen únicamente de JWT para autenticación. Si el token se filtra (ej. via query string), se exponen a CSRF.
- **Recomendación:** Implementar doble-submit cookies o tokens CSRF. Aplicar `SameSite=Strict` en cookies.

#### P8. Sin límites de longitud en campos de texto de API
- **Archivos:** `pkg/web/handlers_features.go` y otros handlers
- **Severidad:** BAJA
- **Problema:** Los campos de texto (títulos, descripciones, contenido) de endpoints como prompts, tasks, y knowledge no tienen límite de longitud. No se aplica `http.MaxBytesReader` en todos los endpoints que lo necesitan, permitiendo payloads excesivos.
- **Recomendación:** Añadir validación de longitud: título ≤ 255 chars, descripción ≤ 500 chars, contenido ≤ 100,000 chars. Aplicar `http.MaxBytesReader` consistentemente.

### Race Conditions / Concurrencia - Pendientes

#### P9. Deadlock por adquisición recursiva de lock en MultiUserChannelManager
- **Archivo:** `pkg/channels/multiuser_manager.go:245-272`
- **Severidad:** CRITICA
- **Problema:** `RestartUserChannels()` adquiere `m.mu.Lock()` en línea 246, luego llama a `m.GetOrCreateManagerForUser(userUUID)` en línea 272. `GetOrCreateManagerForUser()` intenta adquirir el mismo lock en línea 91, causando deadlock en un mutex no recursivo.
- **Recomendación:** Crear un método interno `getOrCreateManagerForUserLocked()` que asuma que el lock ya está adquirido. Llamar al método interno desde `RestartUserChannels()` y mantener el método público con lock para uso externo.

#### P10. Data race en campo ctx de DiscordChannel
- **Archivo:** `pkg/channels/discord.go:44, 59-62, 68`
- **Severidad:** MEDIA
- **Problema:** El campo `ctx` se escribe en `Start()` (línea 68) y se lee en `getContext()` (líneas 59-62) sin sincronización. El race detector de Go reportaría un data race entre estas operaciones.
- **Recomendación:** Proteger `ctx` con `sync.RWMutex` o usar `atomic.Value` para el almacenamiento del contexto.

#### P11. MCP readLoop puede enviar a canal después de cleanup en sendRequest
- **Archivo:** `pkg/mcp/client.go:415-423`
- **Severidad:** MEDIA
- **Problema:** Hay una ventana entre `delete(c.pending, *resp.ID)` en readLoop y el `delete(c.pending, id)` en el defer de sendRequest. El dato puede perderse silenciosamente si ambos ejecutan concurrentemente.
- **Recomendación:** Dejar que solo sendRequest maneje el cleanup del mapa pending.

#### P12. Goroutines con context.Background() sin cancelación
- **Archivos:** `pkg/web/server.go:3458-3466`, `pkg/web/handlers_user_config.go:203-215`, `pkg/channels/multiuser_manager.go:115`, `pkg/agent/loop.go:1429-1432`
- **Severidad:** MEDIA
- **Problema:** Múltiples goroutines se lanzan con `context.Background()` sin mecanismo de cancelación. No se detienen cuando el servidor se apaga, dejando operaciones huérfanas.
- **Recomendación:** Propagar el contexto del servidor a las goroutines. Usar `context.WithTimeout` para operaciones que no deban durar indefinidamente. Implementar WaitGroup para tracking.

### Fugas de Recursos - Pendientes

#### P13. Goroutine leak en MCP client readLoop
- **Archivo:** `pkg/mcp/client.go:181, 391-425`
- **Severidad:** ALTA
- **Problema:** `readLoop()` se inicia como goroutine sin mecanismo de cancelación por contexto. Solo termina si la lectura del pipe retorna error. Si el proceso MCP queda colgado, la goroutine queda bloqueada indefinidamente en `ReadBytes()`.
- **Recomendación:** Añadir `stopChan` al Client struct. Usar `context.Done()` con select en readLoop, o asegurar que `Close()` siempre cierra el pipe de stdout para desbloquear la lectura.

#### P14. Crecimiento no acotado del ring buffer de RecentEvents
- **Archivo:** `pkg/observability/metrics.go:375-378`
- **Severidad:** MEDIA
- **Problema:** El buffer de eventos recientes usa reslicing (`m.RecentEvents = m.RecentEvents[1:]`) que es ineficiente: el array subyacente no libera la memoria de los elementos eliminados. Con uso prolongado, esto genera memory leak sutil.
- **Recomendación:** Implementar un buffer circular verdadero con array de tamaño fijo y tracking por índice.

#### P15. Sin límite de conexiones en servidor TCP de MaixCam
- **Archivo:** `pkg/channels/maixcam.go:42-61`
- **Severidad:** MEDIA
- **Problema:** El servidor TCP acepta conexiones sin límite ni rate limiting, vulnerable a agotamiento de recursos.
- **Recomendación:** Implementar semáforo para limitar conexiones concurrentes. Añadir timeouts de lectura/escritura.

#### P16. Busy-wait en listener de WhatsApp
- **Archivo:** `pkg/channels/whatsapp.go:107-144`
- **Severidad:** BAJA
- **Problema:** `listen()` hace polling continuo con `time.Sleep(1s)` cuando la conexión es nil, y `time.Sleep(2s)` en errores de lectura, sin backoff exponencial.
- **Recomendación:** Implementar backoff exponencial con tiempo máximo de espera. Usar un state machine de conexión.

### Manejo de Errores / Integridad de Datos - Pendientes

#### P17. rows.Err() no verificado en backup export (5 ubicaciones)
- **Archivo:** `pkg/storage/backup.go:116, 140, 167, 189-199, 219-229`
- **Severidad:** ALTA
- **Problema:** Después de iterar `rows.Next()` en las cinco consultas de exportación (sessions, messages, tasks, task_logs, channel_mappings), no se verifica `rows.Err()`. Errores de iteración (I/O de disco, corrupción de BD) se ignoran silenciosamente, generando backups incompletos que se marcan como exitosos.
- **Recomendación:** Añadir `if err := rows.Err(); err != nil { return data, fmt.Errorf("iterating X: %w", err) }` después de cada loop.

#### P18. Error de json.Unmarshal ignorado en import options
- **Archivo:** `pkg/web/handlers_backup.go:416`
- **Severidad:** MEDIA
- **Problema:** `json.Unmarshal([]byte(body), &importOptions)` ignora silenciosamente errores de parseo, causando que opciones inválidas se traten como defaults. El usuario puede creer que sus opciones se aplicaron cuando no fue así.
- **Recomendación:** Verificar el error y retornar `400 Bad Request` si el JSON es inválido.

#### P19. Errores silenciosos en queries de backup export
- **Archivo:** `pkg/storage/backup.go:108, 132, 156`
- **Severidad:** MEDIA
- **Problema:** Si las queries de exportación fallan (`err != nil`), se ignora silenciosamente y se retorna datos vacíos sin indicar error.
- **Recomendación:** Al menos logear los errores, o retornarlos al llamador.

#### P20. Unchecked LastInsertId() en 4 ubicaciones
- **Archivos:** `pkg/storage/prompts.go:57`, `pkg/storage/workflow.go:91, 187`, `pkg/storage/task.go:44`
- **Severidad:** MEDIA
- **Problema:** `LastInsertId()` puede retornar error (drivers que no lo soportan, fallos de BD). Sin verificación, se retorna id=0 al llamador que lo trata como éxito, corrompiendo lookups posteriores (ej. `GetPrompt(0)` busca un registro inexistente).
- **Recomendación:**
  ```go
  id, err := result.LastInsertId()
  if err != nil { return 0, fmt.Errorf("get insert id: %w", err) }
  ```

#### P21. Unchecked RowsAffected() en workflows
- **Archivo:** `pkg/storage/workflow.go:161, 174`
- **Severidad:** MEDIA
- **Problema:** `RowsAffected()` puede retornar error en fallos de BD (disco lleno, permisos). Con el error ignorado, `n` defaultea a 0, que se confunde con "registro no encontrado" — enmascarando errores reales.
- **Recomendación:** Verificar error de `RowsAffected()` antes de comparar con 0.

#### P22. Errores de json.Marshal ignorados en agent loop (6 ubicaciones)
- **Archivo:** `pkg/agent/loop.go:1024, 1042, 1278, 1293, 1364, 1378`
- **Severidad:** MEDIA
- **Problema:** `json.Marshal(tc.Arguments)` se llama sin verificar error en 6 ubicaciones donde se serializan argumentos de herramientas. Si falla (tipos no serializables), los argumentos se pierden en el historial de sesión.
- **Recomendación:** Verificar el error y usar `{}` como fallback:
  ```go
  argumentsJSON, err := json.Marshal(tc.Arguments)
  if err != nil {
      logger.WarnCF("agent", "failed to marshal tool arguments", ...)
      argumentsJSON = []byte("{}")
  }
  ```

#### P23. rows.Err() no verificado en hasColumn PRAGMA query
- **Archivo:** `pkg/storage/backup.go:503-514`
- **Severidad:** MEDIA
- **Problema:** Después del loop de iteración en `hasColumn()`, no se verifica `rows.Err()`. Un error de iteración retorna `false` silenciosamente (como si la columna no existiera), afectando la lógica de migración/import.
- **Recomendación:** Añadir `rows.Err()` check o al menos logear el error.

#### P24. Mensajes de error HTTP exponen detalles internos
- **Archivos:** `pkg/web/handlers_features.go:33, 58, 99, 107, 159` y otros handlers
- **Severidad:** BAJA
- **Problema:** Mensajes de error como `http.Error(w, "failed to create prompt: "+err.Error(), 500)` pueden contener detalles de implementación (rutas de BD, mensajes de driver SQL).
- **Recomendación:** Logear el error internamente y retornar un mensaje genérico al usuario.

#### P25. Acumulación de mensajes sin límite en sesiones
- **Archivo:** `pkg/session/manager.go:158`
- **Severidad:** MEDIA
- **Problema:** `session.Messages = append(session.Messages, msg)` crece indefinidamente. La sumarización existe pero es asíncrona. En sesiones de larga duración, la memoria crece sin control.
- **Recomendación:** Implementar truncamiento automático: `if len(session.Messages) > 500 { session.Messages = session.Messages[len(session.Messages)-500:] }`.

### Calidad de Código - Pendientes

#### P26. Sin recovery de panics en HTTP handlers
- **Archivos:** Todos los handlers en `pkg/web/handlers_*.go`
- **Severidad:** ALTA
- **Problema:** No hay middleware de recovery. Un panic en cualquier handler HTTP crashea todo el proceso del servidor web.
- **Recomendación:** Añadir un middleware `recoveryMiddleware` que capture panics con `defer recover()`, logee el stack trace, y retorne `500 Internal Server Error`.

#### P27. Sin rate limiting en endpoints API (excepto login)
- **Archivo:** `pkg/web/server.go`
- **Severidad:** MEDIA
- **Problema:** Solo el endpoint de login tiene rate limiting. Endpoints como `/api/v1/chat`, `/api/v1/tools/execute`, `/api/v1/knowledge/search` no tienen límite, permitiendo abuso y DoS.
- **Recomendación:** Implementar rate limiting global por usuario/IP para endpoints autenticados. Considerar un middleware basado en token bucket con límites diferenciados según el costo de cada endpoint.

#### P28. Directorios temporales predecibles
- **Archivos:** `cmd/makoclaw/main.go:638`, `pkg/utils/media.go:67`, `pkg/channels/signal.go:262`
- **Severidad:** MEDIA
- **Problema:** Se usan rutas fijas en `/tmp/` (`makoclaw_history`, `makoclaw_media`, `makoclaw-signal`). En sistemas multi-usuario, otro usuario podría pre-crear estos directorios o symlinks maliciosos (symlink attack).
- **Recomendación:** Usar `os.MkdirTemp()` para directorios, y añadir componentes aleatorios a los nombres de archivo. Para el historial CLI, usar el directorio home del usuario.

#### P29. HTTP Clients creados ad-hoc sin reusar conexiones
- **Archivos:** `pkg/tools/web.go:89`, `pkg/skills/installer.go:48, 98`, `pkg/web/server.go:4289`
- **Severidad:** BAJA
- **Problema:** Se crean instancias `http.Client` nuevas en cada llamada, desperdiciando el pool de conexiones HTTP. Cada instancia crea un nuevo transport, anulando los beneficios de keep-alive.
- **Recomendación:** Crear un `*http.Client` singleton por módulo con transport persistente reutilizable.

#### P30. Errores de json.Encoder silenciados en respuestas HTTP
- **Archivos:** `pkg/web/server.go` y todos los handlers
- **Severidad:** BAJA
- **Problema:** Decenas de ubicaciones usan `_ = json.NewEncoder(w).Encode(...)`, ignorando silenciosamente errores de encoding. Los clientes pueden recibir JSON incompleto o corrupto sin indicación de error.
- **Recomendación:** Al mínimo, logear los errores de encoding:
  ```go
  if err := json.NewEncoder(w).Encode(...); err != nil {
      logger.WarnCF("web", "failed to encode response", ...)
  }
  ```

#### P31. Sin IdleConnTimeout en Ollama HTTP client
- **Archivo:** `pkg/providers/ollama_provider.go:61-63`
- **Severidad:** BAJA
- **Problema:** El HTTP client de OllamaProvider tiene un `Timeout: 120s` global pero no configura `Transport.IdleConnTimeout`, permitiendo que conexiones idle se mantengan indefinidamente.
- **Recomendación:** Añadir `Transport: &http.Transport{IdleConnTimeout: 30 * time.Second}`.

#### P32. Goroutines de summarización sin tracking en AgentLoop
- **Archivo:** `pkg/agent/loop.go:1429+`
- **Severidad:** BAJA
- **Problema:** Las goroutines de summarización y otras operaciones background se lanzan sin WaitGroup ni tracking. El shutdown del servidor puede interrumpirlas a mitad de operación.
- **Recomendación:** Implementar `sync.WaitGroup` en AgentLoop para rastrear goroutines activas y esperarlas durante shutdown.

#### P33. API keys de providers retornadas en respuestas API
- **Archivo:** `pkg/web/providers_handler.go:69-83`
- **Severidad:** MEDIA
- **Problema:** El endpoint `GetProvidersConfig` retorna los API keys completos en la respuesta JSON, exponiéndolos al frontend y potencialmente a logs. Las claves deberían estar enmascaradas.
- **Recomendación:** Enmascarar API keys en respuestas (ej. `sk-...abc` mostrando solo los últimos 3 caracteres). Implementar un endpoint separado y protegido para obtener claves completas si es necesario.

#### P34. Errores de json.Marshal ignorados en OAuth
- **Archivo:** `pkg/auth/oauth.go:173, 228`
- **Severidad:** BAJA
- **Problema:** `reqBody, _ := json.Marshal(...)` ignora errores. Aunque mapas de strings siempre marshalan correctamente, el patrón viola mejores prácticas de Go.
- **Recomendación:** Verificar el error para mantener consistencia con el resto del código.

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

| Commit | Descripción | Archivos |
|--------|-------------|----------|
| `b845424` | Fix security bugs, race conditions, error handling | 18 archivos |
| `5c9bdfe` | Fix race conditions, resource leaks, error handling | 11 archivos |
| `dcbda1d` | Fix critical security vulnerabilities and data integrity | 6 archivos |
| `65c7b55` | Fix symlink exfiltration, spawn race, email validation | 4 archivos |

**Total correcciones:** 38 problemas corregidos en 33 archivos modificados, 620 líneas añadidas, 209 eliminadas.
**Total pendientes:** 34 problemas identificados pendientes de corrección.

---

## Resumen por Archivos Más Afectados

| Archivo | Corregidos | Pendientes | Prioridad |
|---------|:----------:|:----------:|:---------:|
| `pkg/storage/backup.go` | 1 | 5 | ALTA |
| `pkg/web/server.go` | 3 | 6 | ALTA |
| `pkg/mcp/client.go` | 1 | 3 | ALTA |
| `pkg/channels/multiuser_manager.go` | 0 | 2 | CRITICA |
| `pkg/agent/loop.go` | 2 | 2 | MEDIA |
| `pkg/storage/workflow.go` | 0 | 3 | MEDIA |
| `pkg/web/handlers_features.go` | 0 | 2 | MEDIA |
| `pkg/web/auth.go` | 2 | 1 | BAJA |

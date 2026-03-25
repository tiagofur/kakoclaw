# PRD 04 - Multiuser Architecture and Data Isolation

## Objetivo

Garantizar aislamiento total de datos, configuraciones y ejecucion entre usuarios para eliminar riesgos de fuga de informacion y habilitar uso empresarial seguro.

## Modelo de aislamiento

### 1) Identidad

- Cada usuario tiene UUID unico.
- Toda solicitud debe resolver identidad en backend antes de tocar datos.
- JWT solo habilita sesion; no reemplaza controles de autorizacion por recurso.

### 2) Datos y almacenamiento

- Datos operativos por usuario en storage dedicado.
- Historial de chat, tareas, knowledge y configuracion separados.
- Nada de consultas sin filtro de usuario.
- Migraciones con validacion de integridad por tenant/usuario.

### 3) Configuracion

- Capas: defaults del sistema -> config global -> override de usuario.
- Secretos por usuario (API keys, tokens de canal) en su propio ambito.
- Admin no debe acceder automaticamente a secretos de usuario.

### 4) Workspace y filesystem

- Operaciones de archivos confinadas al workspace del usuario.
- Validar path traversal y symlink escape en herramientas de filesystem/shell.
- Politicas por rol para tools riesgosas.

## Arquitectura de permisos

1. RBAC base: admin, operator, viewer (extensible).
2. Tool permissions por rol + overrides por usuario.
3. Canales permitidos por usuario/sender mapping.
4. Auditoria de operaciones sensibles (auth, config, exec, files, channels).

## Controles de seguridad minimos

- Input validation en todos los handlers.
- Sanitizacion y limites para comandos shell.
- Rate limiting en auth y endpoints sensibles.
- Logs estructurados sin exponer secretos.
- Rotacion de tokens y revocacion de sesiones.

## Plan de hardening por fases

### Fase A - Baseline verificable

- Matriz de acceso por recurso x rol x usuario.
- Tests automatizados de aislamiento (lectura, escritura, listados).
- Pruebas de regresion en endpoints y tools.

### Fase B - Endurecimiento

- Auditoria central de eventos criticos.
- Alertas por intentos de acceso cruzado.
- Politicas de expiracion y rotacion de credenciales.

### Fase C - Operacion enterprise

- Reportes periodicos de seguridad y aislamiento.
- Playbook de incident response.
- Evidencia de compliance operativo (interno).

## Requisitos no negociables

1. Ningun usuario puede leer datos de otro usuario.
2. Ningun usuario puede ejecutar acciones fuera de su workspace.
3. Ninguna consulta de storage se acepta sin contexto de usuario.
4. Toda operacion sensible deja traza auditable.

## Validacion tecnica requerida

- Unit tests de guardas y filtros de usuario.
- Integration tests de API multiusuario.
- Security tests de traversal, auth bypass y privilege escalation.
- Checklist manual de smoke para cada release.

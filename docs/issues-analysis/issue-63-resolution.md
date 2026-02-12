# Issue #63 - Manage Cronjobs Within Session

## Estado: ✅ RESUELTO

## Análisis

La funcionalidad solicitada en el issue #63 ya está **completamente implementada** en el codebase actual.

## Implementación Existente

### Tool: `cron` (pkg/tools/cron.go)

El agente tiene acceso a un tool `cron` que permite gestionar tareas programadas directamente desde la conversación:

**Acciones soportadas:**
- `add` - Agregar nueva tarea programada
- `list` - Listar todas las tareas
- `remove` - Eliminar una tarea
- `enable` - Habilitar una tarea
- `disable` - Deshabilitar una tarea

**Parámetros:**
- `action` - Acción a realizar (requerido)
- `message` - Mensaje de la tarea (para add)
- `at_seconds` - Segundos desde ahora para tarea única
- `every_seconds` - Intervalo en segundos para tareas recurrentes
- `cron_expr` - Expresión cron para horarios complejos
- `job_id` - ID de la tarea (para remove/enable/disable)

## Uso

### Desde el Gateway (Modo Recomendado)

```bash
picoclaw gateway
```

Luego en la conversación:

```
Usuario: Recuérdame revisar emails en 10 minutos
Agente: [usa tool cron con action=add, at_seconds=600]

Usuario: Muestra mis tareas programadas
Agente: [usa tool cron con action=list]

Usuario: Elimina la tarea de revisar emails
Agente: [usa tool cron con action=remove, job_id=xxx]
```

### Ejemplos de Prompts

**Recordatorios one-time:**
- "Recuérdame llamar a Juan en 5 minutos"
- "Avísame cuando pasen 30 minutos"

**Tareas recurrentes:**
- "Todos los días a las 9am revisa mis emails"
- "Cada 2 horas envíame un recordatorio de descanso"
- "Lunes y viernes a las 8am envía reporte semanal"

**Gestión:**
- "Muestra todas mis tareas programadas"
- "Elimina la tarea de revisar emails"
- "Deshabilita temporalmente el recordatorio diario"

## Arquitectura

```
Usuario → Canal → Agent Loop → Cron Tool
                                ↓
                         Cron Service
                                ↓
                         Job Storage (JSON)
                                ↓
                         Ejecución en horario
```

## Almacenamiento

Las tareas se guardan en:
```
~/.picoclaw/workspace/cron/jobs.json
```

## Limitaciones

- El tool `cron` **solo está disponible en modo gateway** (`picoclaw gateway`)
- No está disponible en modo agente directo (`picoclaw agent -m "..."`)
- Esto es por diseño, ya que las tareas programadas necesitan el servicio cron en ejecución

## Verificación

Para verificar que funciona:

```bash
1. picoclaw gateway
2. Enviar mensaje: "Recuérdame en 1 minuto que PicoClaw funciona"
3. Esperar 1 minuto
4. El bot debería enviar el recordatorio
```

## Referencias

- Implementación: `pkg/tools/cron.go`
- Servicio: `pkg/cron/service.go`
- Issue original: https://github.com/sipeed/picoclaw/issues/63

# Mejoras Pendientes en Documentación

**Generado**: 2026-03-04
**Estado**: Plan de mejora basado en auditoría completa

---

## Resumen Ejecutivo

| Métrica | Valor | Estado |
|---------|-------|--------|
| Archivos totales | 85 | Buena cobertura |
| Archivos faltantes | 26 | **Crítico** |
| Enlaces rotos | ~26+ | **Crítico** |
| Salud general | 7/10 | Necesita trabajo |

---

## 1. ARCHIVOS FALTANTES (Prioridad Alta)

### 1.1 API Reference (4 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `api-reference/providers.md` | Alta | Documentación de interfaz LLMProvider |
| `api-reference/channels.md` | Alta | Documentación de interfaz Channel |
| `api-reference/config.md` | Media | Referencia de estructura de configuración |
| `api-reference/agent.md` | Media | Referencia de API del AgentLoop |

### 1.2 Development (6 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `development/creating-tools.md` | **Crítica** | Guía para crear tools personalizadas |
| `development/creating-channels.md` | **Crítica** | Guía para crear canales nuevos |
| `development/creating-skills.md` | Alta | Guía del sistema de skills |
| `development/testing.md` | Alta | Estrategia y patrones de testing |
| `development/code-conventions.md` | Media | Guía de estilo de código |
| `development/project-structure.md` | Media | Organización de directorios |

### 1.3 Guides (4 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `guides/llm-providers.md` | Alta | Configuración detallada de providers |
| `guides/channels.md` | Alta | Instrucciones de setup por canal |
| `guides/agent-cli.md` | Media | Uso del agente en CLI |
| `guides/cron-jobs.md` | Media | Gestión de tareas programadas |

### 1.4 Deployment (4 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `deployment/local.md` | Media | Deployment local para desarrollo |
| `deployment/server.md` | Media | Deployment en VPS/servidor |
| `deployment/systemd.md` | Baja | Integración con systemd |
| `deployment/embedded.md` | Baja | Deployment en ARM/RISC-V |

### 1.5 Architecture (2 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `architecture/components.md` | Alta | Documentación de componentes individuales |
| `architecture/diagrams.md` | Media | Diagramas visuales del sistema |

### 1.6 Troubleshooting (4 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `troubleshooting/common-issues.md` | Media | Soluciones a problemas comunes |
| `troubleshooting/config-errors.md` | Media | Errores de configuración |
| `troubleshooting/channel-issues.md` | Baja | Problemas por canal |
| `troubleshooting/debugging.md` | Baja | Técnicas de debugging |

### 1.7 Examples (2 archivos)
| Archivo | Prioridad | Descripción |
|---------|-----------|-------------|
| `examples/automation.md` | Baja | Casos de uso de automatización |
| `examples/integrations.md` | Baja | Ejemplos de integraciones |

---

## 2. CONTENIDO DUPLICADO A CONSOLIDAR

### 2.1 Quickstart
- `guides/quickstart.md` (583 líneas) + `guides/QUICK_START_OVERVIEW.md`
- **Acción**: Fusionar en uno solo

### 2.2 Docker
- `deployment/docker.md` (71 líneas) + `deployment/DOCKER_DEPLOYMENT.md` (614 líneas)
- **Acción**: Mantener DOCKER_DEPLOYMENT.md, convertir docker.md en referencia rápida

### 2.3 Setup de Desarrollo
- `development/setup.md` + `development/DEVELOPMENT_SETUP_GUIDE.md`
- **Acción**: Fusionar en uno solo

### 2.4 Configuración
- `development/CONFIG_ARCHITECTURE.md`
- `development/CONFIG_PERMISSIONS.md`
- `development/CONFIG_PERMISSIONS_QUICK_REF.md`
- **Acción**: Consolidar en 1 guía completa + 1 quick reference

---

## 3. ARCHIVOS A ARCHIVAR

### 3.1 Artefactos de Sesión (mover a `archive/sessions/`)
```
development/PHASE3_FRONTEND_WIZARD.md
development/PHASE3_QUICK_INDEX.md
development/PHASE4_CHANNEL_ONBOARDING.md
development/PHASES_1_2_3_SUMMARY.md
development/PHASES_1_2_3_4_COMPLETE.md
development/IMPLEMENTATION_COMPLETE.md
development/IMPLEMENTATION_CONFIG_PERMISSIONS.md
development/ONBOARDING_IMPLEMENTATION_SUMMARY.md
```

### 3.2 Reportes Desactualizados
```
REPORTE_COMPLETO_PICACLAW.md → archive/legacy/
```

---

## 4. INCONSISTENCIAS DE FORMATO

### 4.1 Convención de Nombres
| Actual | Recomendado |
|--------|-------------|
| `UPPERCASE_FILE.md` | `lowercase-file.md` |
| `MixedCase.md` | `lowercase-file.md` |

**Política**: Usar `lowercase-with-hyphens.md` para todos los archivos nuevos.

### 4.2 Idioma
| Sección | Idioma Actual | Recomendación |
|---------|---------------|---------------|
| Guías de usuario | Español | Mantener español |
| API Reference | Inglés | Mantener inglés |
| Development | Mixto | Estandarizar a inglés |

### 4.3 Markdown
- Usar siempre especificador de lenguaje en code blocks: ` ```go `, ` ```json `
- Headers empezar con `#`, no saltar niveles
- Listas usar `-` consistentemente

---

## 5. ENLACES ROTOS (Arreglar en README.md)

```markdown
# Enlaces que apuntan a archivos que no existen:
- [Componentes Principales](./architecture/components.md)
- [Diagramas del Sistema](./architecture/diagrams.md)
- [Configuración de Proveedores LLM](./guides/llm-providers.md)
- [Crear un Nuevo Tool](./development/creating-tools.md)
- [Crear un Nuevo Canal](./development/creating-channels.md)
```

---

## 6. PLAN DE EJECUCIÓN

### Fase 1: Correcciones Críticas (1-2 días)
- [ ] Crear stubs para 26 archivos faltantes
- [ ] Arreglar todos los enlaces rotos en READMEs
- [ ] Actualizar fecha de versión a 2026-03-04

### Fase 2: Consolidación (2-3 días)
- [ ] Fusionar documentos duplicados (quickstart, docker, setup)
- [ ] Mover artefactos de sesión a `archive/`
- [ ] Consolidar docs de configuración

### Fase 3: Contenido Prioritario (1 semana)
- [ ] Escribir `development/creating-tools.md`
- [ ] Escribir `development/creating-channels.md`
- [ ] Escribir `api-reference/providers.md`
- [ ] Escribir `guides/channels.md`

### Fase 4: Completar (2 semanas)
- [ ] Completar todos los archivos faltantes
- [ ] Estandarizar formato markdown
- [ ] Agregar índice a `plans/`
- [ ] Revisar y actualizar contenido desactualizado

---

## 7. ARCHIVOS POR SECCIÓN - ESTADO

### API Reference
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `tools.md` | ✅ Completo | 699 |
| `workflows.md` | ✅ Completo | 807 |
| `providers.md` | ❌ Falta | — |
| `channels.md` | ❌ Falta | — |
| `config.md` | ❌ Falta | — |
| `agent.md` | ❌ Falta | — |

### Architecture
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `overview.md` | ✅ Excelente | 670 |
| `data-flow.md` | ✅ Bueno | 527 |
| `components.md` | ❌ Falta | — |
| `diagrams.md` | ❌ Falta | — |

### Guides
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `quickstart.md` | ✅ Excelente | 583 |
| `installation.md` | ✅ Excelente | 520 |
| `email-setup.md` | ✅ Completo | ~200 |
| `degraded-mode.md` | ✅ Bueno | ~150 |
| `onboarding-wizard.md` | 🟡 Parcial | ~200 |
| `llm-providers.md` | ❌ Falta | — |
| `channels.md` | ❌ Falta | — |
| `agent-cli.md` | ❌ Falta | — |
| `cron-jobs.md` | ❌ Falta | — |

### Development
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `setup.md` | ✅ Bueno | 584 |
| `MCP_SETUP_GUIDE.md` | ✅ Completo | 514 |
| `MULTI_AGENT_SETUP.md` | ✅ Completo | ~300 |
| `contributing.md` | 🟡 Parcial | ~200 |
| `AGENTS.md` | 🔴 Stub | 34 |
| `creating-tools.md` | ❌ Falta | — |
| `creating-channels.md` | ❌ Falta | — |
| `creating-skills.md` | ❌ Falta | — |
| `testing.md` | ❌ Falta | — |
| `code-conventions.md` | ❌ Falta | — |

### Deployment
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `DOCKER_DEPLOYMENT.md` | ✅ Excelente | 614 |
| `termux-android.md` | ✅ Completo | 196 |
| `docker.md` | 🟡 Duplicado | 71 |
| `local.md` | ❌ Falta | — |
| `server.md` | ❌ Falta | — |
| `systemd.md` | ❌ Falta | — |
| `embedded.md` | ❌ Falta | — |

### Examples
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `workflows.md` | ✅ Excelente | 1,174 |
| `basic-examples.md` | ✅ Bueno | 239 |
| `workflow-templates.json` | ✅ Completo | 22KB |
| `automation.md` | ❌ Falta | — |
| `integrations.md` | ❌ Falta | — |

### Troubleshooting
| Archivo | Estado | Líneas |
|---------|--------|--------|
| `faq.md` | ✅ Bueno | 549 |
| `workflows.md` | ✅ Bueno | 615 |
| `workflow-ui-guide.md` | ✅ Específico | 279 |
| `http-safety-guard-error.md` | ✅ Específico | 168 |
| `common-issues.md` | ❌ Falta | — |
| `config-errors.md` | ❌ Falta | — |
| `channel-issues.md` | ❌ Falta | — |
| `debugging.md` | ❌ Falta | — |

---

## 8. PRIORIDADES RECOMENDADAS

### Crítico (Esta semana)
1. `development/creating-tools.md` - Los desarrolladores lo necesitan
2. `development/creating-channels.md` - Los desarrolladores lo necesitan
3. Arreglar enlaces rotos en `docs/README.md`

### Alto (Próximas 2 semanas)
4. `api-reference/providers.md`
5. `guides/channels.md`
6. `architecture/components.md`
7. Consolidar documentos duplicados

### Medio (Este mes)
8. Completar guías faltantes
9. Archivar documentos de sesión
10. Estandarizar formato

### Bajo (Backlog)
11. Documentación de deployment adicional
12. Ejemplos de automatización e integraciones
13. Guías de troubleshooting detalladas

---

## Notas

- La documentación existente es de buena calidad
- El problema principal es **completitud**, no calidad
- Priorizar documentación que desbloquea a desarrolladores externos
- Los archivos de "plans/" están bien pero necesitan un índice

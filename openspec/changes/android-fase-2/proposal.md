# Proposal: Android Fase 2 - Feature Parity

## Intent

Completar los 11 features esqueletales de la app Android MakoClaw para lograr paridad funcional con la app web. La Fase 1 estableció la arquitectura y estructura base, pero los features actuales tienen solo 40-60% de completitud. Esta fase es crítica para entregar una app Android con las mismas capacidades que la versión web, permitiendo a los usuarios gestionar agentes AI, workflows, knowledge y metrics desde Android con la misma experiencia funcional.

## Scope

### In Scope
- **11 features esqueletales** (completar implementación):
  - feature-knowledge: RAG base con markdown rendering, CRUD completo
  - feature-workflows: Visual canvas editor + node-based automation
  - feature-metrics: Vico charts, time series, multi-metric
  - feature-cron: Custom cron selector visual (dial + lists)
  - feature-skills: Marketplace UI, install/uninstall, filters
  - feature-agents: Swarm visualizer, agent cards, status indicators
  - feature-memory: Context viewer, retention settings, search
  - feature-files: File manager, upload/download, permissions
  - feature-mcp: Protocol server list, enable/disable, config
  - feature-history: Historial con filtros (date, type, agent)
  - feature-reports: Report generation UI, templates, export
- **Core modules completos**:
  - core-database: DAOs para todas las nuevas entidades
  - core-datastore: Persistencia completa de settings y prefs
  - core-security: Secure storage para JWT y tokens sensibles
- **Componentes UI específicos**: Workflows canvas, Metrics charts, Cron selector

### Out Scope
- Nuevas features no existentes en la app web
- Refactorización mayor de código existente (salvo lo necesario para features)
- Performance optimization avanzada (solo evitar regresiones)
- Implementación de nuevas APIs backend (todas ya definidas)
- Features de colaboración multiusuario (no está en la web)

## Approach

**Parallel Implementation** (recomendado en exploración):

Implementar features en paralelo agrupados por prioridad técnica y valor de negocio. Core modules se completan primero como base para todos los features. Features de prioridad ALTA (knowledge, workflows, metrics, cron) son blockers principales y se implementan en tandas. Features MEDIA y BAJA pueden hacerse en paralelo entre sí.

**Orden de implementación por prioridad**:

**Prioridad ALTA** (críticos, alta complejidad):
1. Core modules completos → 2 días
2. feature-knowledge → 3 días (base para RAG)
3. feature-workflows → 5 días (HIGH-COMPLEXITY, canvas editor)
4. feature-metrics → 4 días (charts con Vico)
5. feature-cron → 3 días (selector visual custom)

**Prioridad MEDIA** (valor medio, complejidad moderada):
6. feature-skills → 3 días (marketplace UI)
7. feature-agents → 3 días (swarm visualizer)
8. feature-memory → 3 días (context viewer)

**Prioridad BAJA** (value add, baja complejidad):
9. feature-files → 2 días (file manager estándar)
10. feature-mcp → 2 días (list + toggle)
11. feature-history → 2 días (list + filtros)
12. feature-reports → 2 días (generation + export)

**Estrategia de paralelismo**:
- Dev 1: Core modules → feature-knowledge → feature-workflows → feature-skills
- Dev 2: feature-metrics → feature-cron → feature-agents → feature-memory
- Dev 3: feature-files → feature-mcp → feature-history → feature-reports

**Decisiones técnicas específicas**:
- **Workflows editor**: Canvas API de Compose (implementación custom, no librería)
- **Metrics charts**: Vico library (alternativa: PhilJay/MPAndroidChart si Vico tiene issues)
- **Cron selector**: Custom implementation con dial rotatorio + listas de selección
- **Markdown rendering**: Markwon (reutilizar del módulo de chat)
- **File upload**: Accompanist Permissions + ActivityResultContracts
- **Image preview**: Coil (ya usado en el proyecto)
- **Tests**: JUnit5 + MockK para unitarios, ComposeTestRule para UI tests

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `core-database/` | Modified | Agregar DAOs para entidades de features nuevos (knowledge, workflows, etc.) |
| `core-datastore/` | Modified | Completar persistencia de settings y preferences de features |
| `core-security/` | Modified | Implementar secure storage para JWT tokens en features que necesitan auth |
| `core-network/` | Modified | Agregar llamadas REST para los nuevos features (endpoints ya definidos) |
| `feature-knowledge/` | Modified | Completar UI, viewmodels, repository, local storage |
| `feature-workflows/` | Modified | Implementar visual canvas editor, node management, edge connections |
| `feature-metrics/` | Modified | Integrar Vico charts, time series data fetching, multi-metric toggle |
| `feature-cron/` | Modified | Implementar cron selector visual con dial y lists |
| `feature-skills/` | Modified | Marketplace UI, install/uninstall, category filters |
| `feature-agents/` | Modified | Swarm visualizer, agent cards, status indicators, logs viewer |
| `feature-memory/` | Modified | Context viewer, retention settings UI, search functionality |
| `feature-files/` | Modified | File manager, upload/download handlers, permission requests |
| `feature-mcp/` | Modified | Server list UI, enable/disable toggles, config screens |
| `feature-history/` | Modified | History list con filtros por fecha/tipo/agente |
| `feature-reports/` | Modified | Report generation UI, template selection, export options |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| **Workflows editor complejidad** | HIGH | Implementación incremental: primero nodos básicos, luego arrastrar, luego conectar edges. Prototipo funcional antes de UI polished. |
| **Vico library bugs/limitaciones** | MED | Evaluar Vico con POC antes de implementación completa. Tener MPAndroidChart como fallback (es estable, maduro). |
| **Core modules como dependencia** | MED | Completar core modules ANTES de features avanzados. Priorizar core-database DAOs y core-security secure storage. |
| **Paralelismo → conflictos merge** | MED | Branch por feature, code review obligatorio, CI/CD con tests antes de merge. Daily sync meeting. |
| **Performance con charts complejos** | LOW | Lazy loading, pagination en metrics data, virtualization en listas grandes. Profiling temprano. |
| **Cron selector UX confusa** | LOW | User testing temprano del diseño visual del dial + lists. Iterar con diseñador si es necesario. |

## Rollback Plan

Cada feature se implementa en branch separado (`feature/xxx-implement`). Si algo falla:

1. **Feature individual**: Revertir merge del feature branch, volver a base funcional antes del feature.
2. **Core module**: Revertir cambios al core, usar stub/mock temporal hasta arreglar.
3. **Library issue (Vico)**: Cambiar a MPAndroidChart como fallback plan, actualizar dependencias.
4. **Performance regression**: Revertir optimización específica, perfil de nuevo, iterar.

**Rollback por milestone**:
- Milestone 1: Si core modules fallan, revertir a base Fase 1, reevaluar enfoque.
- Milestone 2-4: Si feature falla, deshabilitar en feature flags, continuar con otros features en paralelo.

## Dependencies

- **No depende de otros cambios**: Es independiente del desarrollo de la web o backend.
- **APIs del backend**: Ya definidas y documentadas en `core-network/src/main/java/com/makoclaw/core/network/api/FeatureApis.kt` ✅
- **Core modules**: Deben completarse antes de features avanzados (knowledge, workflows, metrics)
- **Librerías externas**: Vico, Markwon, Accompanist, Coil (todas ya evaluadas o en uso)

## Success Criteria

- [ ] Todos los 11 features esqueletales completados y funcionales en Android
- [ ] Paridad funcional con la app web Vue (100% de funcionalidades críticas)
- [ ] Tests unitarios escritos para features nuevos (target: >70% coverage)
- [ ] UI tests críticos pasando (happy paths para features principales)
- [ ] Core modules (database, datastore, security) completados y probados
- [ ] Performance aceptable: no regresiones vs baseline de Fase 1, animaciones 60fps
- [ ] Zero blocking bugs, <10 bugs moderados en QA
- [ ] Documentación de componentes complejos (workflows editor, cron selector)

## Timeline & Resources

**Estimación**: 4-6 semanas con 2-3 developers

**Milestone 1 (Week 1-2)**: Core + Prioridad ALTA (Part 1)
- Completar core-database, core-datastore, core-security
- Implementar feature-knowledge

**Milestone 2 (Week 2-3)**: Prioridad ALTA (Part 2)
- Implementar feature-workflows (editor visual)
- Implementar feature-metrics (gráficos)
- Implementar feature-cron (selector visual)

**Milestone 3 (Week 3-4)**: Prioridad MEDIA
- Implementar feature-skills
- Implementar feature-agents
- Implementar feature-memory

**Milestone 4 (Week 4-5)**: Prioridad BAJA
- Implementar feature-files
- Implementar feature-mcp
- Implementar feature-history
- Implementar feature-reports

**Milestone 5 (Week 5-6)**: Polish & QA
- Tests completos (unitarios + UI)
- UI polish (animations, transitions, estados vacíos)
- Bug fixes
- Performance optimization (evitar regresiones)

**Resources**:
- **2-3 developers Android**: Seniors con experiencia en Compose + Clean Architecture
- **1 designer**: Para componentes complejos (workflows editor, cron selector UX)
- **1 QA**: Para testing manual de features nuevos, especialmente workflows editor

**Success Metrics**:
- Features completados: 11/11 ✅
- Tests coverage: >70%
- Bugs encontrados: <10 (moderados)
- Performance: no regresiones vs baseline Fase 1
- Paridad con web: 100% funcionalidades críticas

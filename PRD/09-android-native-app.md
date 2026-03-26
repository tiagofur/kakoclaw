# PRD 09 - Android Native App

## Visión

MakoClaw Android es una aplicación nativa de primera categoría que proporciona acceso completo a la plataforma MakoClaw, manteniendo la identidad visual de la app web mientras sigue fielmente los lineamientos de **Material Design 3 (Material You)** de Google. La app ofrece una experiencia nativa fluida, con funcionalidades móviles específicas y paridad funcional completa con la app web.

## Misión

Entregar una experiencia Android nativa que:

- Siga Material Design 3 (Material You) para consistencia con el ecosistema Android
- Mantenga la identidad visual de MakoClaw (glass morphism, gradientes, colores temáticos)
- Aproveche capacidades nativas de Android (notifications, share, shortcuts, widgets)
- Proporcione funcionalidades específicas móviles (haptics, gestures, camera, voice)
- Mantenga paridad funcional completa con la app web Vue

## Principios de diseño Android

1. **Material You First**: Seguir Material Design 3 sin compromisos
2. **Native Feel**: Uso de componentes nativos, animaciones fluidas, gestures
3. **Visual Consistency**: Alinearse con el design system de la web pero adaptado a Android
4. **Mobile-First**: Funcionalidades específicas para contexto móvil
5. **Performance**: Animaciones 60fps, carga optimizada, caching inteligente
6. **Accessibility**: Soporte completo de TalkBack, tamaño de texto dinámico, contraste

## Arquitectura técnica

### Stack tecnológico

- **Lenguaje**: Kotlin 1.9+
- **UI**: Jetpack Compose + Material3
- **Arquitectura**: Clean Architecture con MVVM
- **Inyección de dependencias**: Hilt
- **Navegación**: Navigation Compose (type-safe)
- **Networking**: Retrofit2 + OkHttp + WebSocket
- **Base de datos local**: Room + SQLite
- **Preferences**: DataStore (Protocol Buffers)
- **Serialización**: Kotlin Serialization (JSON)
- **Background tasks**: WorkManager
- **Coroutines**: kotlinx.coroutines
- **Flow**: Reactive streams para UI
- **Testing**: JUnit5, Compose UI Testing, MockK

### Estructura modular

```
makoclaw-android/
├── app/                    # Módulo principal
│   ├── MainActivity.kt
│   ├── MakoClawApp.kt
│   └── navigation/         # Navegación global
├── core/                   # Módulos core compartidos
│   ├── core-common/        # Utilidades, extensiones, constants
│   ├── core-model/         # Modelos de datos compartidos
│   ├── core-network/      # APIs, WebSocket, interceptors
│   ├── core-database/     # Room DAOs, entidades, migrations
│   ├── core-datastore/     # Preferences con DataStore
│   ├── core-security/      # Criptografía, JWT, secure storage
│   └── core-ui/            # Componentes UI compartidos, theme
└── feature/                # Módulos de funcionalidades independientes
    ├── feature-auth/       # Autenticación, onboarding, server config
    ├── feature-chat/       # Chat en tiempo real con streaming
    ├── feature-tasks/      # Kanban board con drag & drop
    ├── feature-dashboard/  # Dashboard principal con stats
    ├── feature-agents/     # Gestión de agentes AI
    ├── feature-history/    # Historial de sesiones
    ├── feature-knowledge/  # Base de conocimiento (RAG)
    ├── feature-memory/     # Memoria a largo plazo
    ├── feature-skills/     # Marketplace de skills
    ├── feature-workflows/  # Flujos visuales
    ├── feature-metrics/    # Métricas y analíticas
    ├── feature-cron/       # Tareas programadas
    ├── feature-files/      # Gestión de archivos
    ├── feature-mcp/        # Protocolo MCP
    └── feature-reports/    # Reportes
```

## Design system

### Alineación con la app web

La app Android mantiene consistencia visual con la app web Vue pero adaptada a Material Design 3:

| Aspecto Web Vue | Adaptación Android (Material3) |
| --------------- | ------------------------------ |
| Tokens `makoclaw-*` | Material3 color schemes + custom tokens |
| Glass morphism (backdrop-blur) | SurfaceContainer + Elevation + Blur effects |
| Gradient backgrounds | GradientSurface + SecondaryContainer |
| Cards (border-radius, border) | Card(surfaceColor, elevation) |
| Animaciones 300ms | Material motion system (emphasized, standard) |
| Typography scale | Material3 typography scale |
| Icons (@heroicons/vue) | Material Symbols (Outlined/Filled) |
| Empty states | Empty state pattern de Google |

### Paleta de colores (Material3)

La app usa Material3 Dynamic Colors con overrides para mantener identidad MakoClaw:

```kotlin
// Dark theme (predominado en la app)
val darkColorScheme = darkColorScheme(
    primary = Color(0xFF3B82F6),      // makoclaw-accent (Blue 500)
    onPrimary = Color(0xFFFFFFFF),
    primaryContainer = Color(0xFF1E3A8A),
    onPrimaryContainer = Color(0xFFDBEAFE),

    secondary = Color(0xFF8B5CF6),    // Purple 500
    onSecondary = Color(0xFFFFFFFF),
    secondaryContainer = Color(0xFF4C1D95),
    onSecondaryContainer = Color(0xFFE9D5FF),

    tertiary = Color(0xFF10B981),     // Emerald 500
    onTertiary = Color(0xFFFFFFFF),
    tertiaryContainer = Color(0xFF064E3B),
    onTertiaryContainer = Color(0xFFD1FAE5),

    background = Color(0xFF0F172A),   // Slate 900
    onBackground = Color(0xFFF8FAFC),
    surface = Color(0xFF1E293B),      // Slate 800
    onSurface = Color(0xFFF8FAFC),

    error = Color(0xFFEF4444),
    onError = Color(0xFFFFFFFF),
    errorContainer = Color(0xFF7F1D1D),
    onErrorContainer = Color(0xFFFEE2E2)
)

// Light theme
val lightColorScheme = lightColorScheme(
    primary = Color(0xFF3B82F6),
    onPrimary = Color(0xFFFFFFFF),
    primaryContainer = Color(0xFFDBEAFE),
    onPrimaryContainer = Color(0xFF1E3A8A),

    secondary = Color(0xFF8B5CF6),
    onSecondary = Color(0xFFFFFFFF),
    secondaryContainer = Color(0xFFE9D5FF),
    onSecondaryContainer = Color(0xFF4C1D95),

    tertiary = Color(0xFF10B981),
    onTertiary = Color(0xFFFFFFFF),
    tertiaryContainer = Color(0xFFD1FAE5),
    onTertiaryContainer = Color(0xFF064E3B),

    background = Color(0xFFF8FAFC),  // Slate 50
    onBackground = Color(0xFF0F172A),
    surface = Color(0xFFFFFFFF),
    onSurface = Color(0xFF0F172A),

    error = Color(0xFFEF4444),
    onError = Color(0xFFFFFFFF),
    errorContainer = Color(0xFFFEE2E2),
    onErrorContainer = Color(0xFF7F1D1D)
)
```

### Gradientes por pantalla (alineados con web)

Cada pantalla tiene un gradiente secundario sutil que mantiene identidad MakoClaw:

| Pantalla | Primary Gradient | Secondary Gradient |
| -------- | ---------------- | ------------------ |
| Dashboard | primary/15% | indigo/10% |
| Chat | primary/20% (blue) | purple/15% |
| Tasks | blue/15% | emerald/10% |
| Agents | lime/15% | green/10% |
| Knowledge | teal/15% | emerald/10% |
| Workflows | rose/15% | fuchsia/10% |
| Skills | purple/15% | pink/10% |
| Cron | cyan/15% | blue/10% |
| Files | indigo/15% | violet/10% |
| MCP | orange/15% | amber/10% |

### Componentes clave

#### Glass Card (inspirado en web)
```kotlin
@Composable
fun MakoClawCard(
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
    content: @Composable ColumnScope.() -> Unit
) {
    val cardColor = MaterialTheme.colorScheme.surface
        .copy(alpha = 0.6f)

    Card(
        onClick = onClick ?: {},
        modifier = modifier,
        colors = CardDefaults.cardColors(
            containerColor = cardColor
        ),
        elevation = CardDefaults.cardElevation(
            defaultElevation = 2.dp,
            pressedElevation = 4.dp
        )
    ) {
        content()
    }
}
```

#### Gradient Surface
```kotlin
@Composable
fun GradientSurface(
    modifier: Modifier = Modifier,
    content: @Composable BoxScope.() -> Unit
) {
    Box(
        modifier = modifier
    ) {
        // Gradient sutil de fondo
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.05f),
                            Color.Transparent,
                            MaterialTheme.colorScheme.secondary.copy(alpha = 0.05f)
                        )
                    )
                )
        )
        content()
    }
}
```

#### Glow Effect (para items destacados)
```kotlin
@Composable
fun GlowingItem(
    selected: Boolean,
    onClick: () -> Unit,
    content: @Composable RowScope.() -> Unit
) {
    val glowAlpha by animateFloatAsState(
        targetValue = if (selected) 0.3f else 0f,
        label = "glow"
    )

    Surface(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .drawWithContent {
                drawContent()
                if (selected) {
                    drawRect(
                        color = MaterialTheme.colorScheme.primary.copy(alpha = glowAlpha),
                        size = size,
                        style = Stroke(width = 2.dp.toPx())
                    )
                }
            }
    ) {
        content()
    }
}
```

## Navegación

### Bottom Navigation (principal)

```kotlin
sealed class BottomNavRoute(val route: String, val icon: ImageVector, val label: String) {
    data object Chat : BottomNavRoute("chat", Icons.Filled.Chat, "Chat")
    data object Tasks : BottomNavRoute("tasks", Icons.Filled.TaskAlt, "Tasks")
    data object Dashboard : BottomNavRoute("dashboard", Icons.Filled.Dashboard, "Home")
    data object Agents : BottomNavRoute("agents", Icons.Filled.Groups, "Agents")
}
```

### Drawer Navigation (secundario)

Organizado en secciones lógicas:

```kotlin
data class DrawerItem(
    val route: String,
    val icon: ImageVector,
    val label: String,
    val section: String? = null
)

private val drawerItems = listOf(
    // Conversations
    DrawerItem("history", Icons.Filled.History, "History", "Conversations"),
    DrawerItem("memory", Icons.Filled.Memory, "Memory", "Conversations"),

    // Knowledge
    DrawerItem("knowledge", Icons.Filled.LibraryBooks, "Knowledge Base", "Knowledge"),
    DrawerItem("skills", Icons.Filled.Extension, "Skills & Marketplace", "Knowledge"),

    // Automation
    DrawerItem("workflows", Icons.Filled.AccountTree, "Workflows", "Automation"),
    DrawerItem("cron", Icons.Filled.Schedule, "Scheduled Tasks", "Automation"),

    // System
    DrawerItem("metrics", Icons.Filled.Analytics, "Metrics", "System"),
    DrawerItem("files", Icons.Filled.Folder, "Files", "System"),
    DrawerItem("mcp", Icons.Filled.Hub, "MCP Servers", "System"),
    DrawerItem("reports", Icons.Filled.Assessment, "Reports", "System"),

    // Config
    DrawerItem("settings", Icons.Filled.Settings, "Settings", null)
)
```

## Estado actual de implementación

### Módulos completos (90%+)

#### ✅ feature-auth
- **Screens**: ServerConfig, Login, Signup, Onboarding
- **Funcionalidad**:
  - Configuración de URL de servidor
  - Login con email/password
  - Signup con validación
  - Onboarding wizard
  - Persistencia de credenciales (DataStore)
- **Estado**: 95% completo

#### ✅ feature-chat
- **Screens**: ChatScreen
- **Funcionalidad**:
  - Chat en tiempo real con WebSocket
  - Streaming de respuestas
  - Tool calls visuales
  - Agent events (delegación, especialistas)
  - Thinking block expandible
  - Cambio de sesión
  - Nueva sesión
  - Cancelar ejecución
  - Selección de modelo
  - Historial de mensajes
- **Estado**: 95% completo

#### ✅ feature-dashboard
- **Screens**: DashboardScreen
- **Funcionalidad**:
  - Stats de chats, tareas, agentes, LLM calls
  - Provider info
  - Health check del sistema
  - Pull to refresh
  - Welcome banner con estado
- **Estado**: 90% completo

#### ✅ feature-tasks
- **Screens**: TasksScreen, TaskCard, NewTaskBottomSheet
- **Funcionalidad**:
  - Kanban con tabs swipeables (HorizontalPager)
  - 5 columnas: Backlog, To Do, In Progress, Review, Done
  - Crear nueva tarea con modal
  - Search de tareas
  - Stats por columna
  - Drag & drop (pendiente)
- **Estado**: 85% completo

#### ✅ feature-settings
- **Screens**: SettingsScreen con 8 tabs
- **Funcionalidad**:
  - ProfileTab: usuario, cambio de contraseña, logout
  - ProvidersTab: lista de providers y modelos
  - ToolsTab: lista de herramientas
  - Placeholders: Agents, Channels, Audit, System, Soul
- **Estado**: 70% completo (tabs principales implementados)

### Módulos esqueletales (screens creados, funcionalidad limitada)

#### ⚠️ feature-agents
- **Screens**: AgentsScreen
- **Estado**: 60% (vista de lista creada, funcionalidad pendiente)

#### ⚠️ feature-history
- **Screens**: HistoryScreen
- **Estado**: 60% (vista de sesiones creada, funcionalidad pendiente)

#### ⚠️ feature-knowledge
- **Screens**: KnowledgeScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-memory
- **Screens**: MemoryScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-skills
- **Screens**: SkillsScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-workflows
- **Screens**: WorkflowsScreen
- **Estado**: 40% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-metrics
- **Screens**: MetricsScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-cron
- **Screens**: CronScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-files
- **Screens**: FilesScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-mcp
- **Screens**: McpScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

#### ⚠️ feature-reports
- **Screens**: ReportsScreen
- **Estado**: 50% (estructura creada, funcionalidad pendiente)

### Core modules

#### ✅ core-network
- **Apis**: Chat, Auth, Config, Task, Agent, Metrics, FeatureApis
- **WebSocket**: ChatWebSocketClient, TaskWebSocketClient
- **Interceptor**: AuthInterceptor para JWT
- **Estado**: 90% completo

#### ✅ core-model
- **Models**: Chat, Task, Agent, Knowledge, Provider, User
- **Estado**: 95% completo

#### ✅ core-ui
- **Components**: EmptyState, LoadingScreen
- **Theme**: MakoClawTheme (Material3)
- **Estado**: 80% completo

#### ⚠️ core-database
- **Estado**: 30% (estructura creada, pendiente implementación)

#### ⚠️ core-datastore
- **Estado**: 30% (estructura creada, pendiente implementación)

#### ⚠️ core-security
- **Estado**: 40% (estructura creada, pendiente implementación)

#### ⚠️ core-common
- **Estado**: 70% completo

## Funcionalidades faltantes

### Alta prioridad

1. **feature-tasks**: Drag & drop entre columnas de Kanban
2. **feature-settings**: Implementar tabs de Agents, Channels, Audit, System, Soul
3. **feature-agents**: Gestión completa de agentes (crear, editar, eliminar)
4. **feature-history**: Historial de sesiones con búsqueda y filtros
5. **feature-knowledge**: CRUD de documentos, upload, RAG search
6. **feature-memory**: Memoria a largo plazo con timeline y filtros
7. **feature-skills**: Marketplace con rating, installation, uninstallation
8. **feature-workflows**: Editor visual de flujos con nodes y connections
9. **feature-metrics**: Gráficos de métricas con filtros de tiempo
10. **feature-cron**: CRUD de cron jobs con selector visual de horarios
11. **feature-files**: Gestión de archivos con upload, preview, delete
12. **feature-mcp**: Gestión de servidores MCP
13. **feature-reports**: Generación de reportes con filtros y exportación

### Media prioridad

14. **core-database**: Implementar Room con DAOs para caché offline
15. **core-datastore**: Implementar DataStore para persistencia de settings
16. **core-security**: Implementar secure storage para tokens
17. **core-ui**: Componentes adicionales (SwipeToRefresh, PullToRefresh)
18. **feature-chat**: Attachments (imágenes, archivos)
19. **feature-chat**: Voice input (Speech Recognition)
20. **feature-tasks**: Filtros avanzados, tags, asignación
21. **feature-agents**: Swarm visualizer
22. **feature-workflows**: Executar flujos, ver logs en tiempo real

### Baja prioridad

23. Push notifications (Firebase Cloud Messaging)
24. Widgets (dashboard stats, quick actions)
25. Shortcuts (deep linking)
26. Share intent (compartir desde otras apps)
27. App shortcuts (long press on icon)
28. Picture-in-Picture (para chat)
29. Biometric auth (fingerprint, face unlock)
30. Haptic feedback contextual

## Roadmap de implementación

### Fase 1 - Core completado (2 semanas)
- [x] Arquitectura modular establecida
- [x] Navegación completa
- [x] Auth flow funcional
- [x] Chat con streaming
- [x] Dashboard con stats
- [x] Tasks Kanban básico
- [x] Settings básico

### Fase 2 - Feature parity (4 semanas)
- [ ] feature-agents: CRUD completo
- [ ] feature-history: Historial con search/filtros
- [ ] feature-knowledge: CRUD documentos
- [ ] feature-memory: Memoria con timeline
- [ ] feature-skills: Marketplace básico
- [ ] feature-workflows: Editor visual
- [ ] feature-metrics: Gráficos
- [ ] feature-cron: Cron jobs CRUD
- [ ] feature-files: Gestión archivos
- [ ] feature-mcp: Servidores MCP
- [ ] feature-reports: Reportes

### Fase 3 - Core modules completado (2 semanas)
- [ ] core-database: Room con caching
- [ ] core-datastore: Preferences persistidas
- [ ] core-security: Secure storage
- [ ] core-ui: Componentes adicionales

### Fase 4 - Mobile features (3 semanas)
- [ ] feature-chat: Attachments
- [ ] feature-chat: Voice input
- [ ] feature-tasks: Drag & drop
- [ ] feature-settings: Tabs faltantes
- [ ] Push notifications
- [ ] Widgets
- [ ] Shortcuts
- [ ] Share intent

### Fase 5 - Polish & QA (2 semanas)
- [ ] Tests unitarios (70% coverage)
- [ ] Tests UI (Compose UI Testing)
- [ ] Performance optimization
- [ ] Accessibility audit
- [ ] UI polish (animaciones, transiciones)
- [ ] Error handling robusto
- [ ] Loading states consistentes

## KPIs y calidad

### KPIs de performance

| Métrica | Objetivo | Actual |
| ------- | -------- | ------ |
| Cold start time | < 2s | TBD |
| Warm start time | < 500ms | TBD |
| Screen transition | < 300ms | TBD |
| API response (p50) | < 200ms | TBD |
| API response (p95) | < 1s | TBD |
| WebSocket latency | < 100ms | TBD |
| Memory usage | < 200MB | TBD |
| Battery impact | < 2%/hour | TBD |

### KPIs de calidad

| Métrica | Objetivo | Actual |
| ------- | -------- | ------ |
| Crash-free users | > 99.9% | TBD |
| ANR rate | < 0.1% | TBD |
| Test coverage | > 70% | TBD |
| Accessibility score | > 90% | TBD |
| Lighthouse score | > 90 | TBD |
| Store rating | > 4.5 | TBD |

### Definition of Done

Un feature se considera listo cuando:

- ✅ Funcionalidad implementada y testeada
- ✅ UI sigue Material Design 3
- ✅ Alineado con design system de la web
- ✅ Animaciones fluidas (60fps)
- ✅ Estados de carga/error/vacío
- ✅ Responsive en todos los tamaños de pantalla
- ✅ Accessibility (TalkBack, content descriptions)
- ✅ Dark/Light theme soportado
- ✅ Error handling robusto
- ✅ Tests unitarios escritos
- ✅ Documentación actualizada
- ✅ Code review aprobado

## Testing strategy

### Unit tests
- ViewModel lógica
- Use cases
- Repository layer
- Utils y extensions

### Integration tests
- API layer
- WebSocket connection
- Database operations
- DataStore operations

### UI tests
- Compose UI Testing
- Navegación
- User flows
- Componentes UI

### Performance tests
- Benchmarking
- Memory profiling
- Battery consumption

### Accessibility tests
- TalkBack navigation
- Contrast ratio
- Touch targets (min 48dp)
- Text scaling

## Deployment strategy

### Beta testing
- Firebase App Distribution para testing interno
- TestFlight (si hay iOS) para testing externo
- Crashlytics para crash reports
- Performance monitoring con Firebase Performance

### Release gates
1. Todos los tests pasando
2. Performance benchmarks OK
3. Accessibility audit OK
4. Security review OK
5. Manual testing en dispositivos principales
6. Beta testing con usuarios reales (1 semana)
7. Bug fixes y refinamientos
8. Release candidate (RC)
9. Final review y release

## Consideraciones especiales

### Material You (Dynamic Colors)
- Soportar Dynamic Colors para Android 12+
- Permitir overrides para mantener identidad MakoClaw
- Fallback a colores predefinidos en versiones anteriores

### Multi-window
- Soportar split screen
- Picture-in-Picture para chat

### Work Profiles
- Funcionamiento correcto en dispositivos con work profiles
- Data isolation apropiado

### Tablet support
- Layouts adaptables para tablets
- Two-pane layouts donde apropiado
- Drag & drop entre panes

### Android TV support (opcional)
- Navegación con D-pad
- Layouts adaptados para TV
- Focus states visibles

## Referencias

- [Material Design 3 Guidelines](https://m3.material.io/)
- [Jetpack Compose Documentation](https://developer.android.com/jetpack/compose)
- [Android App Quality Guidelines](https://developer.android.com/docs/quality-guidelines)
- [MakoClaw Web Design System](../frontend-design/SKILL.md)
- [MakoClaw PRD - UX/UI System](./03-ux-ui-system.md)

# PRD 10 - iOS Native App

## Visión

MakoClaw iOS es una aplicación nativa de primera categoría que proporciona acceso completo a la plataforma MakoClaw, manteniendo la identidad visual de la app web mientras sigue fielmente los lineamientos de **Human Interface Guidelines (HIG)** de Apple. La app ofrece una experiencia iOS nativa fluida, con funcionalidades móviles específicas y paridad funcional completa con la app web.

## Misión

Entregar una experiencia iOS nativa que:

- Siga Human Interface Guidelines sin compromisos
- Mantenga la identidad visual de MakoClaw (glass morphism, gradientes, colores temáticos)
- Aproveche capacidades nativas de iOS (notifications, share, shortcuts, widgets, CarPlay)
- Proporcione funcionalidades específicas de iOS (haptics, gestures, camera, voice, Pencil support)
- Mantenga paridad funcional completa con la app web Vue

## Principios de diseño iOS

1. **HIG First**: Seguir Human Interface Guidelines sin compromisos
2. **Apple Feel**: Uso de componentes nativos, animaciones fluidas, gestures (swipe, long press, force touch)
3. **Visual Consistency**: Alinearse con el design system de la web pero adaptado a iOS
4. **Mobile-First**: Funcionalidades específicas para contexto móvil
5. **Performance**: Animaciones 120fps (ProMotion), carga optimizada, caching inteligente
6. **Accessibility**: Soporte completo de VoiceOver, Dynamic Type, Reduce Motion, Larger Text
7. **Privacy-First**: App Tracking Transparency, secure storage, minimal permissions

## Arquitectura técnica

### Stack tecnológico

- **Lenguaje**: Swift 5.9+ / Swift 6 (con concurrency)
- **UI**: SwiftUI + UIKit (para componentes no disponibles en SwiftUI)
- **Arquitectura**: MVVM (Model-View-ViewModel) con Combine
- **Inyección de dependencias**: @Environment o manual DI container
- **Navegación**: NavigationStack / NavigationPath (type-safe)
- **Networking**: URLSession + Async/Await + Starscream (WebSocket)
- **Base de datos local**: SwiftData (iOS 17+) o CoreData
- **Preferences**: UserDefaults + @AppStorage
- **Serialización**: Codable (JSON)
- **Background tasks**: BackgroundTasks framework + BGTaskScheduler
- **Concurrency**: Async/Await, Actors, Tasks
- **Reactive streams**: Combine framework
- **Testing**: XCTest, Swift Testing (XCTest 2)

### Estructura modular

```
MakoClaw/
├── App/                    # Módulo principal (App entry point)
│   ├── MakoClawApp.swift
│   ├── AppDelegate.swift
│   └── SceneDelegate.swift
├── Core/                   # Módulos core compartidos
│   ├── CoreCommon/         # Utilidades, extensiones, constants
│   ├── CoreModel/          # Modelos de datos compartidos (Codable)
│   ├── CoreNetwork/        # APIs, WebSocket, URLProtocol
│   ├── CoreDatabase/       # SwiftData/CoreData stacks
│   ├── CoreDataStore/      # Preferences con @AppStorage
│   ├── CoreSecurity/       # Keychain, CryptoKit, JWT
│   └── CoreUI/             # Componentes UI compartidos, theme
└── Feature/                # Módulos de funcionalidades independientes
    ├── FeatureAuth/        # Autenticación, onboarding, server config
    ├── FeatureChat/        # Chat en tiempo real con streaming
    ├── FeatureTasks/       # Kanban board con drag & drop
    ├── FeatureDashboard/   # Dashboard principal con stats
    ├── FeatureAgents/      # Gestión de agentes AI
    ├── FeatureHistory/    # Historial de sesiones
    ├── FeatureKnowledge/   # Base de conocimiento (RAG)
    ├── FeatureMemory/      # Memoria a largo plazo
    ├── FeatureSkills/      # Marketplace de skills
    ├── FeatureWorkflows/   # Flujos visuales
    ├── FeatureMetrics/    # Métricas y analíticas
    ├── FeatureCron/        # Tareas programadas
    ├── FeatureFiles/       # Gestión de archivos
    ├── FeatureMCP/         # Protocolo MCP
    └── FeatureReports/     # Reportes
```

### Proyecto Xcode

```
MakoClaw.xcodeproj/
├── MakoClaw/
│   ├── App/
│   ├── Core/
│   ├── Feature/
│   ├── Resources/          # Assets.xcassets, Localizable.strings
│   ├── Preview Content/    # SwiftUI previews
│   └── Info.plist
├── MakoClawTests/
├── MakoClawUITests/
└── MakoClawPreviewTests/
```

## Design system

### Alineación con la app web

La app iOS mantiene consistencia visual con la app web Vue pero adaptada a Human Interface Guidelines:

| Aspecto Web Vue | Adaptación iOS (HIG) |
| --------------- | ------------------- |
| Tokens `makoclaw-*` | SF Symbols + Custom Colors |
| Glass morphism (backdrop-blur) | UIMaterialView + BlurEffect |
| Gradient backgrounds | LinearGradient + SwiftUI gradient |
| Cards (border-radius, border) | Card view with shadow and rounded corners |
| Animaciones 300ms | Spring animations, easing functions |
| Typography scale | Dynamic Type (scaled fonts) |
| Icons (@heroicons/vue) | SF Symbols (Apple's icon system) |
| Empty states | Empty state pattern de Apple |

### Paleta de colores (iOS)

La app usa colores del sistema con overrides para mantener identidad MakoClaw:

```swift
// Dark theme (predominado en la app)
extension Color {
    // Primary - MakoClaw Blue
    static let makoclawPrimary = Color(red: 0.23, green: 0.51, blue: 0.96) // #3B82F6

    // Secondary - Purple
    static let makoclawSecondary = Color(red: 0.55, green: 0.36, blue: 0.96) // #8B5CF6

    // Tertiary - Emerald
    static let makoclawTertiary = Color(red: 0.06, green: 0.73, blue: 0.51) // #10B981

    // Background - Dark
    static let makoclawBackground = Color(red: 0.06, green: 0.09, blue: 0.16) // #0F172A
    static let makoclawBackgroundSecondary = Color(red: 0.12, green: 0.16, blue: 0.24) // #1E293B

    // Semantic colors
    static let success = Color.green
    static let warning = Color.orange
    static let error = Color.red
    static let info = Color.blue
}

// Light theme
extension Color {
    static let makoclawLightPrimary = Color(red: 0.23, green: 0.51, blue: 0.96)
    static let makoclawLightBackground = Color(red: 0.97, green: 0.98, blue: 0.99) // #F8FAFC
    static let makoclawLightBackgroundSecondary = Color(red: 1.0, green: 1.0, blue: 1.0) // #FFFFFF
}
```

### Gradientes por pantalla (alineados con web)

Cada pantalla tiene un gradiente secundario sutil que mantiene identidad MakoClaw:

```swift
struct ScreenGradient: ViewModifier {
    let primary: Color
    let secondary: Color

    func body(content: Content) -> some View {
        ZStack {
            LinearGradient(
                colors: [primary.opacity(0.05), Color.clear, secondary.opacity(0.05)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            content
        }
    }
}

extension View {
    func screenGradient(primary: Color, secondary: Color) -> some View {
        modifier(ScreenGradient(primary: primary, secondary: secondary))
    }
}
```

### Tipografía (Dynamic Type)

```swift
extension Font {
    // Following Apple's typography scale
    static let makoclawLargeTitle = Font.largeTitle.weight(.bold)
    static let makoclawTitle = Font.title.weight(.bold)
    static let makoclawTitle2 = Font.title2.weight(.semibold)
    static let makoclawTitle3 = Font.title3.weight(.semibold)
    static let makoclawHeadline = Font.headline.weight(.semibold)
    static let makoclawBody = Font.body
    static let makoclawCallout = Font.callout
    static let makoclawSubheadline = Font.subheadline
    static let makoclawFootnote = Font.footnote
    static let makoclawCaption = Font.caption
    static let makoclawCaption2 = Font.caption2
}
```

### Componentes clave

#### Glass Card (inspirado en web y iOS)

```swift
struct MakoClawCard<Content: View>: View {
    let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        content
            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 16))
            .shadow(color: .black.opacity(0.1), radius: 8, x: 0, y: 4)
    }
}
```

#### SF Symbols mapping

| Web Icon | SF Symbol | Variants |
| -------- | --------- | ------- |
| Chat | `message` | message.fill, message.circle |
| Tasks | `checklist` | checkmark.circle, checkmark.circle.fill |
| Dashboard | `square.grid.2x2` | square.grid.3x2 |
| Agents | `person.2` | person.2.fill |
| History | `clock` | clock.fill |
| Memory | `brain` | brain.head.profile |
| Knowledge | `book` | book.fill, book.circle |
| Skills | `puzzlepiece` | puzzlepiece.fill |
| Workflows | `flowchart` | flowchart.fill |
| Metrics | `chart.bar` | chart.bar.fill |
| Cron | `bell` | bell.fill, bell.badge |
| Files | `folder` | folder.fill |
| MCP | `network` | network.fill |
| Reports | `doc.text` | doc.text.fill |
| Settings | `gearshape` | gearshape.fill |

## Navegación

### Tab Bar (principal)

```swift
enum MainTab: String, CaseIterable, Identifiable {
    case chat = "Chat"
    case tasks = "Tasks"
    case dashboard = "Home"
    case agents = "Agents"

    var id: String { rawValue }

    var icon: String {
        switch self {
        case .chat: return "message"
        case .tasks: return "checklist"
        case .dashboard: return "square.grid.2x2"
        case .agents: return "person.2"
        }
    }

    var filledIcon: String {
        icon + ".fill"
    }
}
```

### Sidebar / Navigation Stack (secundario)

Organizado en secciones lógicas:

```swift
struct NavigationSection: Identifiable {
    let id = UUID()
    let title: String
    let items: [NavigationItem]
}

struct NavigationItem: Identifiable {
    let id = UUID()
    let title: String
    let icon: String
    let destination: AnyHashable
}

let navigationSections = [
    NavigationSection(title: "Conversations", items: [
        NavigationItem(title: "History", icon: "clock", destination: HistoryView()),
        NavigationItem(title: "Memory", icon: "brain.head.profile", destination: MemoryView())
    ]),
    NavigationSection(title: "Knowledge", items: [
        NavigationItem(title: "Knowledge Base", icon: "book", destination: KnowledgeView()),
        NavigationItem(title: "Skills & Marketplace", icon: "puzzlepiece", destination: SkillsView())
    ]),
    NavigationSection(title: "Automation", items: [
        NavigationItem(title: "Workflows", icon: "flowchart", destination: WorkflowsView()),
        NavigationItem(title: "Scheduled Tasks", icon: "bell", destination: CronView())
    ]),
    NavigationSection(title: "System", items: [
        NavigationItem(title: "Metrics", icon: "chart.bar", destination: MetricsView()),
        NavigationItem(title: "Files", icon: "folder", destination: FilesView()),
        NavigationItem(title: "MCP Servers", icon: "network", destination: MCPView()),
        NavigationItem(title: "Reports", icon: "doc.text", destination: ReportsView()),
    ]),
    NavigationSection(title: nil, items: [
        NavigationItem(title: "Settings", icon: "gearshape", destination: SettingsView())
    ])
]
```

## Estado actual de implementación

⚠️ **Nota**: La app iOS aún no está implementada. Este PRD es una especificación completa para su desarrollo.

### Arquitectura base (0%)

- [ ] Configuración de proyecto Xcode
- [ ] Estructura modular (Core/Feature)
- [ ] MVVM setup con Combine
- [ ] NavigationStack type-safe
- [ ] Theme system (light/dark)
- [ ] Environment values (DI)

### Core modules (0%)

#### CoreNetwork (0%)
- [ ] URLSession wrapper con Async/Await
- [ ] WebSocket client (Starscream o native)
- [ ] API clients (Chat, Auth, Config, Task, Agent, Metrics)
- [ ] Request/Response models (Codable)
- [ ] Authentication interceptor
- [ ] Error handling

#### CoreModel (0%)
- [ ] Models (Chat, Task, Agent, Knowledge, Provider, User)
- [ ] Codable con proper decoding strategies
- [ ] Identifiable conequencia

#### CoreDatabase (0%)
- [ ] SwiftData setup (iOS 17+) or CoreData
- [ ] Entity definitions
- [ ] DAOs / Queries
- [ ] Migrations

#### CoreDataStore (0%)
- [ ] UserDefaults wrappers
- [ ] @AppStorage properties
- [ ] Preference keys

#### CoreSecurity (0%)
- [ ] Keychain wrapper
- [ ] JWT handling
- [ ] Secure storage

#### CoreUI (0%)
- [ ] Theme (light/dark)
- [ ] Colors
- [ ] Typography
- [ ] SF Symbols extensions
- [ ] Reusable components (Card, Button, etc.)
- [ ] Empty states
- [ ] Loading states

### Feature modules (0%)

#### FeatureAuth (0%)
- [ ] ServerConfigScreen
- [ ] LoginScreen
- [ ] SignupScreen
- [ ] OnboardingScreen
- [ ] AuthViewModel
- [ ] AuthRepository
- [ ] Form validation

#### FeatureChat (0%)
- [ ] ChatScreen con streaming
- [ ] MessageBubble (user/assistant)
- [ ] ThinkingBlock
- [ ] ToolCallView
- [ ] AgentEventView
- [ ] ChatViewModel
- [ ] WebSocket integration
- [ ] Session management
- [ ] Model selector
- [ ] Input area with keyboard handling

#### FeatureTasks (0%)
- [ ] TasksScreen con TabView
- [ ] Kanban columns (5: Backlog, To Do, In Progress, Review, Done)
- [ ] TaskCard
- [ ] NewTaskSheet
- [ ] TasksViewModel
- [ ] Drag & drop
- [ ] Search
- [ ] Filters

#### FeatureDashboard (0%)
- [ ] DashboardScreen
- [ ] Stats grid
- [ ] Provider info
- [ ] Health check banner
- [ ] DashboardViewModel
- [ ] Pull to refresh

#### FeatureAgents (0%)
- [ ] AgentsScreen (list)
- [ ] AgentDetailView
- [ ] SwarmVisualizer
- [ ] AgentsViewModel

#### FeatureHistory (0%)
- [ ] HistoryScreen (sessions list)
- [ ] SessionDetailView
- [ ] Search & filters
- [ ] HistoryViewModel

#### FeatureKnowledge (0%)
- [ ] KnowledgeScreen (docs list)
- [ ] DocumentDetailView
- [ ] UploadDocumentSheet
- [ ] RAG search
- [ ] KnowledgeViewModel

#### FeatureMemory (0%)
- [ ] MemoryScreen (timeline)
- [ ] MemoryEntryView
- [ ] Search & filters
- [ ] MemoryViewModel

#### FeatureSkills (0%)
- [ ] SkillsScreen (marketplace)
- [ ] SkillDetailView
- [ ] Rating widget
- [ ] Install/Uninstall
- [ ] SkillsViewModel

#### FeatureWorkflows (0%)
- [ ] WorkflowsScreen
- [ ] WorkflowEditorView (visual)
- [ ] ExecuteWorkflowView
- [ ] WorkflowsViewModel

#### FeatureMetrics (0%)
- [ ] MetricsScreen (charts)
- [ ] Filters (time range)
- [ ] MetricsViewModel

#### FeatureCron (0%)
- [ ] CronScreen
- [ ] CronJobDetailView
- [ ] Schedule selector
- [ ] CronViewModel

#### FeatureFiles (0%)
- [ ] FilesScreen
- [ ] FilePreviewView
- [ ] UploadFileSheet
- [ ] FilesViewModel

#### FeatureMCP (0%)
- [ ] McpScreen
- [ ] McpServerDetailView
- [ ] McpViewModel

#### FeatureReports (0%)
- [ ] ReportsScreen
- [ ] ReportDetailView
- [ ] GenerateReportView
- [ ] ReportsViewModel

#### FeatureSettings (0%)
- [ ] SettingsScreen con TabView
- [ ] ProfileTab
- [ ] ProvidersTab
- [ ] AgentsTab
- [ ] ChannelsTab
- [ ] ToolsTab
- [ ] AuditTab
- [ ] SystemTab
- [ ] SoulTab
- [ ] SettingsViewModel

## Roadmap de implementación

### Fase 0 - Setup inicial (1 semana)
- [ ] Configurar proyecto Xcode
- [ ] Setup arquitectura modular
- [ ] Implementar MVVM base
- [ ] Setup navigation
- [ ] Implementar theme system
- [ ] Setup DI container

### Fase 1 - Core completado (2 semanas)
- [ ] CoreNetwork: APIs, WebSocket, Auth interceptor
- [ ] CoreModel: Models (Codable)
- [ ] CoreUI: Theme, colors, components
- [ ] CoreDataStore: Preferences
- [ ] CoreSecurity: Keychain wrapper

### Fase 2 - FeatureAuth + FeatureChat (3 semanas)
- [ ] FeatureAuth: Login, Signup, ServerConfig, Onboarding
- [ ] FeatureChat: Chat completo con streaming, WebSocket, sessions
- [ ] Tests: Unit + UI para auth y chat

### Fase 3 - FeatureTasks + FeatureDashboard (2 semanas)
- [ ] FeatureTasks: Kanban con tabs, drag & drop
- [ ] FeatureDashboard: Stats, health check, refresh
- [ ] Tests: Unit + UI para tasks y dashboard

### Fase 4 - FeatureAgents + FeatureHistory + FeatureKnowledge (3 semanas)
- [ ] FeatureAgents: CRUD de agentes
- [ ] FeatureHistory: Historial de sesiones
- [ ] FeatureKnowledge: CRUD de documentos, RAG search
- [ ] Tests: Unit + UI

### Fase 5 - FeatureMemory + FeatureSkills + FeatureWorkflows (3 semanas)
- [ ] FeatureMemory: Memoria con timeline
- [ ] FeatureSkills: Marketplace
- [ ] FeatureWorkflows: Editor visual
- [ ] Tests: Unit + UI

### Fase 6 - FeatureMetrics + FeatureCron + FeatureFiles + FeatureMCP + FeatureReports (3 semanas)
- [ ] FeatureMetrics: Gráficos
- [ ] FeatureCron: Cron jobs CRUD
- [ ] FeatureFiles: Gestión archivos
- [ ] FeatureMCP: Servidores MCP
- [ ] FeatureReports: Reportes
- [ ] Tests: Unit + UI

### Fase 7 - FeatureSettings + CoreDatabase (2 semanas)
- [ ] FeatureSettings: 8 tabs completos
- [ ] CoreDatabase: SwiftData implementation
- [ ] Caching offline
- [ ] Tests: Unit + UI

### Fase 8 - iOS features específicos (3 semanas)
- [ ] Push notifications (APNs)
- [ ] Widgets (dashboard stats, quick actions)
- [ ] App Intents (shortcuts, Siri)
- [ ] Share extension
- [ ] Spotlight search
- [ ] Handoff
- [ ] Continuity Camera (upload desde iPhone)

### Fase 9 - Polish & QA (2 semanas)
- [ ] Tests completos (70% coverage)
- [ ] UI tests (XCUITest)
- [ ] Performance optimization
- [ ] Accessibility audit
- [ ] UI polish (animaciones, transiciones)
- [ ] Error handling robusto
- [ ] Loading states consistentes
- [ ] Dark/Light mode testing

### Fase 10 - Beta testing & Release (2 semanas)
- [ ] TestFlight beta (1 semana)
- [ ] Bug fixes y refinamientos
- [ ] App Store optimization (ASO)
- [ ] Screenshots y metadata
- [ ] Store listing
- [ ] Release

Total estimado: **~23 semanas (5-6 meses)**

## Funcionalidades específicas de iOS

### Haptic feedback

```swift
import UIKit

enum HapticFeedback {
    static func impact(_ style: UIImpactFeedbackGenerator.FeedbackStyle = .medium) {
        let generator = UIImpactFeedbackGenerator(style: style)
        generator.impactOccurred()
    }

    static func notification(_ type: UINotificationFeedbackGenerator.FeedbackType) {
        let generator = UINotificationFeedbackGenerator()
        generator.notificationOccurred(type)
    }

    static func selection() {
        let generator = UISelectionFeedbackGenerator()
        generator.selectionChanged()
    }
}
```

### Dynamic Type support

```swift
// All text should use Dynamic Type scaling
Text("Hello, World!")
    .font(.body)
    .dynamicTypeSize(...large)

// Or accessabilityDynamicType() for larger text support
```

### VoiceOver support

```swift
// All interactive elements must have accessibility labels
Button("Send") {
    // Action
}
.accessibilityLabel("Send message")
.accessibilityHint("Tap to send your message to MakoClaw")

// Images need descriptions
Image("user-avatar")
    .accessibilityLabel("User avatar")
```

### Share sheet

```swift
ShareLink(item: URL(string: "https://makoclaw.example.com/share")!) {
    Label("Share", systemImage: "square.and.arrow.up")
}
```

### App Intents (Siri shortcuts)

```swift
import AppIntents

struct QuickChatIntent: AppIntent {
    static var title: LocalizedStringResource = "Quick Chat"
    static var description = IntentDescription("Start a quick chat with MakoClaw")

    @MainActor
    func perform() async throws -> some IntentResult {
        // Navigate to chat
        return .result()
    }
}
```

### WidgetKit

```swift
import WidgetKit
import SwiftUI

struct MakoClawWidget: Widget {
    let kind: String = "MakoClawWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: Provider()) { entry in
            MakoClawWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("MakoClaw Stats")
        .description("Shows quick stats from your MakoClaw dashboard")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}
```

### Spotlight search

```swift
import CoreSpotlight

func indexChatSession(id: String, title: String, content: String) {
    let attributeSet = CSSearchableItemAttributeSet(itemContentType: "text")
    attributeSet.title = title
    attributeSet.contentDescription = content
    attributeSet.keywords = title.components(separatedBy: " ")

    let item = CSSearchableItem(
        uniqueIdentifier: id,
        domainIdentifier: "com.makoclaw.sessions",
        attributeSet: attributeSet
    )

    CSSearchableIndex.default().indexSearchableItems([item])
}
```

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
| Memory usage | < 250MB | TBD |
| Battery impact | < 2%/hour | TBD |
| App size | < 50MB | TBD |

### KPIs de calidad

| Métrica | Objetivo | Actual |
| ------- | -------- | ------ |
| Crash-free users | > 99.9% | TBD |
| Launch crashes | < 0.1% | TBD |
| Test coverage | > 70% | TBD |
| Accessibility score | > 90% | TBD |
| App Store rating | > 4.5 | TBD |
| Time to interactive | < 3s | TBD |

### Definition of Done

Un feature se considera listo cuando:

- ✅ Funcionalidad implementada y testeada
- ✅ UI sigue Human Interface Guidelines
- ✅ Alineado con design system de la web
- ✅ Animaciones fluidas (60-120fps en ProMotion)
- ✅ Estados de carga/error/vacío
- ✅ Responsive en todos los tamaños de pantalla (iPhone SE a Pro Max)
- ✅ Accessibility (VoiceOver, Dynamic Type, Reduce Motion)
- ✅ Dark/Light theme soportado
- ✅ Error handling robusto
- ✅ Tests unitarios escritos
- ✅ UI tests escritos
- ✅ Code review aprobado
- ✅ Documentation actualizada

## Testing strategy

### Unit tests
- ViewModel lógica
- Use cases
- Repository layer
- Utils y extensions
- Model validation

### Integration tests
- API layer
- WebSocket connection
- Database operations
- DataStore operations

### UI tests
- XCTest / XCUITest
- Navigation flows
- User interactions
- Componentes UI
- Accessibility testing

### Performance tests
- XCTest performance tests
- Time Profiler
- Memory Leaks
- Battery usage

### Accessibility tests
- VoiceOver navigation
- Dynamic Type scaling
- Contrast ratio (min 4.5:1)
- Touch targets (min 44pt)
- Reduce Motion

## Deployment strategy

### Beta testing
- TestFlight para beta testing
- Crashlytics para crash reports
- Performance monitoring con Firebase Performance o MetricKit
- Analytics para user insights

### Release gates
1. Todos los tests pasando (unit, integration, UI)
2. Performance benchmarks OK
3. Accessibility audit OK
4. Security review OK
5. Manual testing en dispositivos principales
6. TestFlight beta con usuarios reales (1 semana)
7. Bug fixes y refinamientos
8. Release candidate (RC)
9. Final review
10. App Store submission

### App Store listing

- **Icon**: SF Symbol-based custom icon, 1024x1024
- **Screenshots**: iPhone (6.7", 6.5", 5.5") y iPad (12.9")
- **Preview video**: 30s demo de features principales
- **Description**: Localizado en inglés y español
- **Keywords**: AI, assistant, chat, automation, tasks, workflows
- **Age rating**: 4+
- **Privacy**: App Privacy Report (data collection minimal)

## Consideraciones especiales

### Human Interface Guidelines (HIG)
- Seguir HIG sin compromisos
- Use SF Symbols para icons
- Standard UI components (Navigation Bar, Tab Bar, Modals, Sheets)
- Gestures estándar (swipe to delete, long press, force touch)
- Animaciones spring-based
- Typography con Dynamic Type

### Adaptive layouts
- iPhone (SE, 14, 14 Pro, 14 Pro Max, 15, 15 Pro Max)
- iPad (mini, 10.9", 11", 12.9")
- iPadOS-specific features (Stage Manager, Slide Over, Split View)
- Layouts adaptables para diferentes tamaños

### Multitasking & Windowing
- Soportar Slide Over y Split Over en iPad
- Scene configuration apropiada
- State restoration

### Focus & Screen Time
- Soportar modo Focus
- Integración con Screen Time
- Notifications respetar Focus modes

### Privacy & Security
- App Tracking Transparency (ATT)
- Privacy manifest
- Secure storage en Keychain
- Minimal permissions (solicitar solo cuando necesario)
- Data minimization

### App Clips (opcional)
- App Clip para quick actions (ej: start chat)
- NFC tags o QR codes para invocar App Clip
- Lightweight version (< 10MB)

### CarPlay (opcional)
- Dashboard stats en CarPlay
- Voice control con Siri
- Notificaciones en vehicle

## Referencias

- [Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)
- [Swift Concurrency](https://developer.apple.com/documentation/swift/concurrency)
- [App Store Review Guidelines](https://developer.apple.com/app-store/review/guidelines/)
- [MakoClaw Web Design System](../frontend-design/SKILL.md)
- [MakoClaw PRD - UX/UI System](./03-ux-ui-system.md)
- [MakoClaw PRD - Android Native App](./09-android-native-app.md)

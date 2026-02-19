# KakoClaw Web UI - Vue 3 Frontend

Modern, responsive web interface for KakoClaw built with Vue 3, Vite, and Tailwind CSS.

## Features

- **Responsive Design**: Desktop (two-pane layout) and Mobile (tab-based layout)
- **Sidebar Navigation**: VS Code-style navigation with collapsible sidebar
- **Real-time Chat**: WebSocket-based agent communication
- **Kanban Board**: 5-column task management with drag & drop
- **Dark Theme**: Professional dark mode optimized for long coding sessions
- **State Management**: Pinia stores for auth, chat, tasks, and UI preferences
- **Modern Stack**: Vue 3 Composition API, Vite, TailwindCSS, axios

## Development

### Prerequisites

- Node.js 18+ and npm
- Running KakoClaw backend on `http://localhost:8080`

### Local Development

```bash
cd pkg/web/frontend

# Install dependencies
npm install

# Start development server (http://localhost:5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Project Structure

```
src/
├── components/
│   ├── Auth/           # Login and password change components
│   ├── Layout/         # Sidebar and layout components
│   └── Tasks/          # Kanban, task modals, etc.
├── views/
│   ├── LoginPage.vue   # Login screen
│   ├── DashboardPage.vue # Main dashboard with sidebar
│   ├── ChatTab.vue     # Chat interface
│   └── TasksTab.vue    # Tasks/Kanban interface
├── stores/             # Pinia state management
│   ├── authStore.js    # Authentication state
│   ├── chatStore.js    # Chat messages state
│   ├── taskStore.js    # Tasks and filters state
│   └── uiStore.js      # UI preferences (theme, layout)
├── services/           # API and WebSocket services
│   ├── api.js          # Axios client
│   ├── authService.js  # Auth endpoints
│   ├── taskService.js  # Task endpoints
│   └── websocketService.js # WebSocket classes
├── router/
│   └── index.js        # Vue Router configuration
└── styles/
    └── globals.css     # Global Tailwind CSS
```

## API Integration

The frontend connects to the backend via:

### REST APIs
- `/api/v1/auth/login` - Authentication
- `/api/v1/auth/change-password` - Password change
- `/api/v1/tasks` - CRUD operations for tasks

### WebSockets
- `/ws/chat` - Real-time chat with agent
- `/ws/tasks` - Real-time task updates

**Note**: Development server proxies API/WebSocket calls to `http://localhost:8080` via Vite config.

## Build & Deployment

The frontend is built to `dist/` and embedded in the Go binary:

```bash
# From repository root
make build          # Builds frontend + Go binary
make build-frontend # Builds frontend only
make build-all      # Cross-platform builds
```

The Go server serves the bundled frontend from `pkg/web/dist/` via the `embed` directive.

## Theming

Colors are defined in `tailwind.config.js`:

```js
colors: {
  'KakoClaw': {
    'bg': '#0d1117',        // Main background
    'surface': '#161b22',   // Surface color
    'border': '#30363d',    // Border color
    'accent': '#007acc',    // Blue accent
    'text': '#e0e0e0',      // Primary text
    'text-secondary': '#8b949e', // Secondary text
    'success': '#3fb950',   // Success green
    'error': '#f85149',     // Error red
  }
}
```

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari 12+, Chrome Android)

## Performance

- **Vite**: Fast HMR during development, optimized production builds
- **Tree-shaking**: Unused dependencies are removed in production
- **Code-splitting**: Vue Router + async components for lazy loading
- **Compression**: All assets compressed in production build

Run Lighthouse for detailed performance metrics:

```bash
npm run build && npm run preview
# Open http://localhost:4173 in browser DevTools
```

## Troubleshooting

### Frontend not loading
- Ensure backend is running: `KakoClaw web` on port 8080
- Check browser console (F12) for errors
- Verify backend is accessible: `curl http://localhost:8080/api/v1/health`

### WebSocket connection fails
- Backend must be running on same host/port
- Check browser DevTools Network tab for WebSocket errors
- Verify firewall allows WebSocket connections

### Styling issues
- Clear browser cache: Ctrl+Shift+R (or Cmd+Shift+R)
- In dev mode, Vite HMR should auto-update CSS
- Production build must be re-run with `npm run build`

## Contributing

See [contributing guidelines](../../../docs/development/contributing.md)

# KakoClaw Frontend - Quick Start Guide

## Start Here

### 🏃 Quick Run (Production - Compiled)

```bash
# Build frontend + Go binary
make build

# Run production server
./build/KakoClaw web

# Open browser
open http://localhost:8080
```

**Note**: First time requires setup with credentials

---

## 💻 Development (Hot Reload)

### Terminal 1: Frontend Dev Server

```bash
cd pkg/web/frontend
npm install        # Once only
npm run dev        # http://localhost:5173

# Edit any file → auto-reload in browser
```

### Terminal 2: Backend API Server

```bash
# Ensure config.json has username/password set
docker compose up   # or your existing backend
# Must be on http:// localhost:8080
```

**Backend API**: http://localhost:8080
**Frontend Dev**: http://localhost:5173 (proxies API to backend)

---

## 📦 Build for Production

```bash
# One-liner from repo root
make build

# Or step by step
cd pkg/web/frontend && npm run build && cd ../..
make build

# Binary ready
./build/KakoClaw web
```

---

## 🗂️ Project Structure

```
pkg/web/frontend/
├── src/
│   ├── views/
│   │   ├── LoginPage.vue      ← Auth screen
│   │   ├── DashboardPage.vue  ← Main container
│   │   ├── ChatTab.vue        ← Chat UI (~85 lines)
│   │   └── TasksTab.vue       ← Kanban UI (~130 lines)
│   ├── components/
│   │   ├── Auth/
│   │   │   ├── LoginForm.vue
│   │   │   └── ChangePasswordModal.vue
│   │   ├── Layout/
│   │   │   └── Sidebar.vue    ← Navigation + theme + profile
│   │   └── Tasks/
│   │       ├── KanbanColumn.vue
│   │       ├── NewTaskModal.vue
│   │       └── TaskDetailsModal.vue
│   ├── stores/
│   │   ├── authStore.js       ← User, JWT, session
│   │   ├── chatStore.js       ← Messages, connection
│   │   ├── taskStore.js       ← Tasks, filters, sorting
│   │   └── uiStore.js         ← Theme, sidebar, active tab
│   └── services/
│       ├── api.js             ← Axios client + interceptors
│       ├── authService.js
│       ├── taskService.js
│       └── websocketService.js
├── router/
│   └── index.js               ← Routes: /login, /dashboard
├── styles/
│   └── globals.css            ← Tailwind base + custom
├── App.vue                    ← Root component
├── main.js                    ← Entry point
├── index.html                 ← HTML template
├── vite.config.js             ← Build configuration
├── tailwind.config.js         ← Design tokens
└── postcss.config.js
```

---

## 🎨 Design Your Feature

### Add New Page

```javascript
// src/views/MyPage.vue
<template>
  <Sidebar>
    <div class="p-4">
      <h1 class="text-2xl font-bold">My Feature</h1>
    </div>
  </Sidebar>
</template>

<script setup>
import Sidebar from '../components/Layout/Sidebar.vue'
</script>
```

### Add to Navigation

```javascript
// src/components/Layout/Sidebar.vue - add to nav items
<router-link
  to="/my-page"
  class="flex items-center gap-3 px-3 py-2 rounded hover:bg-KakoClaw-border"
>
  <svg class="w-5 h-5"><!-- icon --></svg>
  <span v-if="!sidebarCollapsed">My Page</span>
</router-link>
```

### Use API in Component

```javascript
import { ref, onMounted } from 'vue'
import taskService from '../services/taskService'

const tasks = ref([])

onMounted(async () => {
  try {
    tasks.value = await taskService.fetchTasks()
  } catch (error) {
    console.error('Failed to fetch:', error)
  }
})
```

### State Management (Pinia)

```javascript
// src/stores/myStore.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useMyStore = defineStore('my', () => {
  const data = ref([])
  const isLoading = ref(false)
  
  const count = computed(() => data.value.length)
  
  function add(item) { data.value.push(item) }
  
  return { data, isLoading, count, add }
})

// In component:
import { useMyStore } from '../stores/myStore'
const store = useMyStore()
// store.data, store.add(), store.count
```

---

## 🌐 Available Endpoints

### REST API

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/v1/health` | GET | ❌ | Server status |
| `/api/v1/auth/login` | POST | ❌ | Login |
| `/api/v1/auth/change-password` | POST | ✅ | Change password |
| `/api/v1/auth/me` | GET | ✅ | Current user |
| `/api/v1/tasks` | GET | ✅ | List tasks |
| `/api/v1/tasks` | POST | ✅ | Create task |
| `/api/v1/tasks/{id}` | PUT | ✅ | Update task |
| `/api/v1/tasks/{id}` | DELETE | ✅ | Delete task |
| `/api/v1/tasks/{id}/status` | PATCH | ✅ | Change status |
| `/api/v1/tasks/{id}/logs` | GET | ✅ | Task logs |

### WebSocket

| URL | Auth | Purpose |
|-----|------|---------|
| `/ws/chat` | ✅ | Chat with agent |
| `/ws/tasks` | ✅ | Task updates |

**Auth**: Token automatically sent in `Authorization: Bearer {token}` header

---

## 🎯 Common Tasks

### Login
```javascript
import authService from '../services/authService'

const { token } = await authService.login('admin', 'password123')
// Token stored in localStorage automatically
```

### Chat
```javascript
import { ChatWebSocket } from '../services/websocketService'
import { useChatStore } from '../stores/chatStore'

const ws = new ChatWebSocket()
await ws.connect()

ws.send({
  type: 'message',
  content: 'Hello'
})

ws.on('message', (msg) => {
  const chatStore = useChatStore()
  chatStore.addMessage(msg)
})
```

### Tasks
```javascript
const taskStore = useTaskStore()

// Fetch
const tasks = await taskService.fetchTasks()
taskStore.setTasks(tasks)

// Filter
taskStore.setFilter('status', 'in_progress')

// Sort
taskStore.setSortBy('recent')

// Access filtered data
console.log(taskStore.filteredTasks)
console.log(taskStore.tasksByStatus.todo)
```

---

## 🔧 Troubleshooting

### npm install fails
```bash
rm -rf node_modules package-lock.json
npm install
```

### Port 5173 already in use
```bash
# Kill the process
lsof -i :5173 | grep -v PID | awk '{print $2}' | xargs kill

# Or use different port
npm run dev -- --port 3000
```

### CORS/WebSocket errors
- Backend must be on `http://localhost:8080`
- Check `vite.config.js` proxy settings
- Browser console (F12) for actual error

### Components not updating
- Vue DevTools browser extension
- Check if state is properly defined in Pinia store
- Use `computed()` for derived values

### Build fails
```bash
# Clear cache
rm -rf dist node_modules package-lock.json
npm install
npm run build
```

---

## 📊 File Sizes

| File | Size | Gzip |
|------|------|------|
| vue.js | 93 KB | 35 KB |
| vendor.js | 36 KB | 14 KB |
| app.js | 40 KB | 10 KB |
| styles.css | 14 KB | 3.6 KB |
| **Total** | **~183 KB** | **~63 KB** |

---

## 🚀 Tips & Tricks

### Faster Development Loop
```bash
# Terminal 1: Watch only frontend changes
cd pkg/web/frontend && npm run dev

# Terminal 2: Watch GO changes (optional)
go run ./cmd/KakoClaw web &
```

### Debug Network Requests
```javascript
// In browser console
localStorage.setItem('DEBUG', 'api:*')
location.reload()
// Then check Network tab in DevTools
```

### Test Responsive Design
```bash
# Dev server supports this
# Press F12 in Chrome →  Device Toolbar (Ctrl+Shift+M)

# Or test production build
npm run build && npm run preview
# http://localhost:4173
```

### Component Isolation Testing
```bash
# Create a test component
src/components/TestComponent.vue

# Import in DashboardPage temporarily
<TestComponent />

# Run dev server to see live
npm run dev
```

---

## 📖 Useful Commands

```bash
# Frontend directory
cd pkg/web/frontend

# Install dependencies
npm install

# Development with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check (if TypeScript added later)
npm run type-check
```

---

## 🎓 Learning Resources

- [Vue 3 Docs](https://vuejs.org)
- [Vite Guide](https://vitejs.dev)
- [TailwindCSS](https://tailwindcss.com)
- [Pinia](https://pinia.vuejs.org)
- [Vue Router](https://router.vuejs.org)

---

## Next Steps

1. ✅ Frontend built and working
2. Run `npm run dev` for live development
3. Make changes to `.vue` files → see updates instantly
4. For production: `make build` → `./build/KakoClaw web`

**Questions?** Check `pkg/web/frontend/README.md` for detailed docs.

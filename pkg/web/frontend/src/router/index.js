import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import LoginPage from '../views/LoginPage.vue'
import MainLayout from '../components/Layout/MainLayout.vue'
import DashboardView from '../views/DashboardView.vue'
import ChatView from '../views/ChatView.vue'
import TasksView from '../views/TasksView.vue'
import HistoryView from '../views/HistoryView.vue'
import MemoryView from '../views/MemoryView.vue'
import ReportsView from '../views/ReportsView.vue'
import SkillsView from '../views/SkillsView.vue'
import CronView from '../views/CronView.vue'
import SettingsView from '../views/SettingsView.vue'
import FilesView from '../views/FilesView.vue'
import KnowledgeView from '../views/KnowledgeView.vue'
import MCPView from '../views/MCPView.vue'
import MetricsView from '../views/MetricsView.vue'
import WorkflowView from '../views/WorkflowView.vue'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: LoginPage,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'dashboard',
        component: DashboardView
      },
      {
        path: 'chat',
        name: 'chat',
        component: ChatView
      },
      {
        path: 'tasks',
        name: 'tasks',
        component: TasksView
      },
      {
        path: 'history',
        name: 'history',
        component: HistoryView
      },
      {
        path: 'memory',
        name: 'memory',
        component: MemoryView
      },
      {
        path: 'reports',
        name: 'reports',
        component: ReportsView
      },
      {
        path: 'skills',
        name: 'skills',
        component: SkillsView
      },
      {
        path: 'cron',
        name: 'cron',
        component: CronView
      },
      {
        path: 'settings',
        name: 'settings',
        component: SettingsView
      },
      {
        path: 'files',
        name: 'files',
        component: FilesView
      },
      {
        path: 'knowledge',
        name: 'knowledge',
        component: KnowledgeView
      },
      {
        path: 'mcp',
        name: 'mcp',
        component: MCPView
      },
      {
        path: 'metrics',
        name: 'metrics',
        component: MetricsView
      },
      {
        path: 'workflows',
        name: 'workflows',
        component: WorkflowView
      },
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  // Check if session is expired
  if (authStore.isAuthenticated && authStore.isSessionExpired) {
    authStore.logout()
  }

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.name === 'login' && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router

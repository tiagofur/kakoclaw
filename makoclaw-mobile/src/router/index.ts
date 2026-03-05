import { createRouter, createWebHistory } from '@ionic/vue-router'
import { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import HomePage from '../views/HomePage.vue'
import KanbanPage from '../views/KanbanPage.vue'
import LoginPage from '../views/LoginPage.vue'
import SettingsPage from '../views/SettingsPage.vue'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { requiresAuth: false }
  },
  {
    path: '/home',
    name: 'Home',
    component: HomePage,
    meta: { requiresAuth: true }
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: KanbanPage,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsPage,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Auth navigation guard
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Restore session from storage if needed
  if (!authStore.isAuthenticated) {
    await authStore.restoreSession()
  }

  const requiresAuth = to.meta.requiresAuth !== false
  const isAuthenticated = authStore.isAuthenticated && !authStore.isSessionExpired

  if (requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && isAuthenticated) {
    next('/home')
  } else {
    next()
  }
})

export default router

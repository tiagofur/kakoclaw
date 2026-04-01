import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useUIStore = defineStore('ui', () => {
  const theme = ref('dark')
  const sidebarCollapsed = ref(false)
  const activeTab = ref('chat')
  const canInstallPwa = ref(false)
  const devStudioEnabled = ref(false)
  const devStudioStatusLoaded = ref(false)
  let deferredPrompt = null

  // Capture the PWA install event
  if ('beforeinstallprompt' in window) {
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      deferredPrompt = e
      canInstallPwa.value = true
    })
  }

  async function installPwa() {
    if (!deferredPrompt) return
    deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    if (outcome === 'accepted') {
      console.log('User accepted the install prompt')
    } else {
      console.log('User dismissed the install prompt')
    }
    deferredPrompt = null
    canInstallPwa.value = false
  }

  function setTheme(newTheme) {
    theme.value = newTheme
    localStorage.setItem('ui.theme', newTheme)
    if (newTheme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  function toggleTheme() {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('ui.sidebarCollapsed', sidebarCollapsed.value)
  }

  function setActiveTab(tab) {
    activeTab.value = tab
  }

  function restoreUIPreferences() {
    const savedTheme = localStorage.getItem('ui.theme')
    const savedSidebarState = localStorage.getItem('ui.sidebarCollapsed')
    
    if (savedTheme) {
      setTheme(savedTheme)
    }
    
    if (savedSidebarState !== null) {
      sidebarCollapsed.value = JSON.parse(savedSidebarState)
    }
  }

  // Fetch Dev Studio enabled status from user config
  async function fetchDevStudioStatus() {
    try {
      const { data } = await api.get('/me/config')
      devStudioEnabled.value = !!data?.config?.dev_studio?.enabled
    } catch {
      devStudioEnabled.value = false
    } finally {
      devStudioStatusLoaded.value = true
    }
  }

  // Set Dev Studio status locally (called after toggle save)
  function setDevStudioEnabled(enabled) {
    devStudioEnabled.value = enabled
  }

  return {
    theme,
    sidebarCollapsed,
    activeTab,
    canInstallPwa,
    devStudioEnabled,
    devStudioStatusLoaded,
    setTheme,
    toggleTheme,
    toggleSidebar,
    setActiveTab,
    restoreUIPreferences,
    installPwa,
    fetchDevStudioStatus,
    setDevStudioEnabled
  }
})

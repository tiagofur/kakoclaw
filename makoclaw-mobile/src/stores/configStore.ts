import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Preferences } from '@capacitor/preferences'

const STORAGE_KEY = 'makoclaw.config'

export const useConfigStore = defineStore('config', () => {
    // User-configurable server URL (e.g. https://myserver.com or http://192.168.1.100:8080)
    const serverUrl = ref('http://localhost:8080')
    const theme = ref<'dark' | 'light' | 'system'>('system')
    const notificationsEnabled = ref(true)

    // Server status (read-only, fetched from API)
    const configured = ref(false)
    const degradedMode = ref(false)
    const activeProviders = ref<string[]>([])

    const isServerConfigured = computed(() => configured.value && !degradedMode.value)

    async function loadFromStorage() {
        try {
            const { value } = await Preferences.get({ key: STORAGE_KEY })
            if (value) {
                const data = JSON.parse(value)
                serverUrl.value = data.serverUrl || serverUrl.value
                theme.value = data.theme || theme.value
                notificationsEnabled.value = data.notificationsEnabled ?? true
            }
        } catch (e) {
            console.warn('Failed to load config from storage:', e)
        }
    }

    async function saveToStorage() {
        try {
            await Preferences.set({
                key: STORAGE_KEY,
                value: JSON.stringify({
                    serverUrl: serverUrl.value,
                    theme: theme.value,
                    notificationsEnabled: notificationsEnabled.value
                })
            })
        } catch (e) {
            console.warn('Failed to save config to storage:', e)
        }
    }

    async function setServerUrl(url: string) {
        // Normalize: remove trailing slash
        serverUrl.value = url.replace(/\/+$/, '')
        await saveToStorage()
    }

    async function setTheme(newTheme: 'dark' | 'light' | 'system') {
        theme.value = newTheme
        await saveToStorage()
    }

    async function setNotificationsEnabled(enabled: boolean) {
        notificationsEnabled.value = enabled
        await saveToStorage()
    }

    // Fetch server config status from the backend
    async function checkStatus() {
        try {
            // Import dynamically to avoid circular deps
            const api = (await import('../services/api')).default
            const response = await api.get('/config/status')
            configured.value = response.data.configured
            degradedMode.value = response.data.degradedMode
            activeProviders.value = response.data.activeProviders || []
            return response.data
        } catch (error) {
            console.error('Failed to check config status:', error)
            degradedMode.value = true
            configured.value = false
            throw error
        }
    }

    return {
        serverUrl,
        theme,
        notificationsEnabled,
        configured,
        degradedMode,
        activeProviders,
        isServerConfigured,
        loadFromStorage,
        saveToStorage,
        setServerUrl,
        setTheme,
        setNotificationsEnabled,
        checkStatus
    }
})

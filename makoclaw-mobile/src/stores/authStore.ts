import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Preferences } from '@capacitor/preferences'

const AUTH_KEYS = {
    user: 'makoclaw.auth.user',
    token: 'makoclaw.auth.token',
    expiry: 'makoclaw.auth.sessionExpiry'
}

export interface User {
    username: string
    role: string
}

export const useAuthStore = defineStore('auth', () => {
    const user = ref<User | null>(null)
    const token = ref<string | null>(null)
    const sessionExpiry = ref<string | null>(null)

    const isAuthenticated = computed(() => !!token.value)

    const isSessionExpired = computed(() => {
        if (!sessionExpiry.value) return false
        return new Date() > new Date(sessionExpiry.value)
    })

    async function setCredentials(newUser: User, newToken: string, expiryMinutes = 1440) {
        user.value = newUser
        token.value = newToken
        const expiry = new Date()
        expiry.setMinutes(expiry.getMinutes() + expiryMinutes)
        sessionExpiry.value = expiry.toISOString()

        // Persist to Capacitor secure storage
        await Promise.all([
            Preferences.set({ key: AUTH_KEYS.user, value: JSON.stringify(newUser) }),
            Preferences.set({ key: AUTH_KEYS.token, value: newToken }),
            Preferences.set({ key: AUTH_KEYS.expiry, value: sessionExpiry.value })
        ])
    }

    async function restoreSession() {
        try {
            const [savedUser, savedToken, savedExpiry] = await Promise.all([
                Preferences.get({ key: AUTH_KEYS.user }),
                Preferences.get({ key: AUTH_KEYS.token }),
                Preferences.get({ key: AUTH_KEYS.expiry })
            ])

            if (savedUser.value && savedToken.value && savedExpiry.value) {
                user.value = JSON.parse(savedUser.value)
                token.value = savedToken.value
                sessionExpiry.value = savedExpiry.value
            }
        } catch (e) {
            console.warn('Failed to restore auth session:', e)
        }
    }

    async function logout() {
        user.value = null
        token.value = null
        sessionExpiry.value = null
        await Promise.all([
            Preferences.remove({ key: AUTH_KEYS.user }),
            Preferences.remove({ key: AUTH_KEYS.token }),
            Preferences.remove({ key: AUTH_KEYS.expiry })
        ])
    }

    return {
        user,
        token,
        sessionExpiry,
        isAuthenticated,
        isSessionExpired,
        setCredentials,
        restoreSession,
        logout
    }
})

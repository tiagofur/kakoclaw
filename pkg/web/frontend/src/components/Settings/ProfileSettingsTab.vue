<template>
  <div class="space-y-6 max-w-2xl mx-auto animate-fadeIn">
    <div class="glass-panel rounded-2xl p-8">
      <h3 class="text-sm font-bold uppercase tracking-widest text-kakoclaw-text-secondary opacity-60 mb-8 flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        Profile Information
      </h3>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <div v-else-if="errorMessage" class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400 text-sm mb-6">
        {{ errorMessage }}
      </div>

      <div v-else class="space-y-6">
        <div v-if="successMessage" class="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4 text-emerald-400 text-sm mb-6">
          {{ successMessage }}
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Username</label>
          <input v-model="profile.username" type="text" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent backdrop-blur-sm transition-all" />
          <p class="text-xs text-kakoclaw-text-secondary mt-1 opacity-70">Changing your username will require you to log in again.</p>
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Email</label>
          <input v-model="profile.email" type="email" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent backdrop-blur-sm transition-all" placeholder="your.email@example.com" />
          <p class="text-xs text-kakoclaw-text-secondary mt-1 opacity-70">Used for password recovery and notifications.</p>
        </div>

        <div class="grid grid-cols-2 gap-4 pt-2">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-1 opacity-70">Role</label>
            <div class="bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm text-kakoclaw-text">
              <span class="px-2 py-0.5 text-[10px] font-bold uppercase rounded-full" :class="profile.role === 'admin' ? 'bg-teal-500/10 text-teal-400' : 'bg-kakoclaw-accent/10 text-kakoclaw-accent'">
                {{ profile.role }}
              </span>
            </div>
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-1 opacity-70">Member Since</label>
            <div class="bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm text-kakoclaw-text-secondary">
              {{ formatDate(profile.created_at) }}
            </div>
          </div>
        </div>

        <div class="pt-6 space-y-3">
          <button @click="saveProfile" :disabled="saving" class="w-full bg-kakoclaw-accent text-white py-3 rounded-xl font-bold hover:bg-kakoclaw-accent-hover transition-all shadow-lg shadow-kakoclaw-accent/20 flex items-center justify-center disabled:opacity-50 active:scale-95">
            <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
            Save Profile
          </button>

          <button @click="showChangePassword = true" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border text-kakoclaw-text py-3 rounded-xl font-medium hover:bg-kakoclaw-surface transition-all flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            Change Password
          </button>
        </div>
      </div>
    </div>

    <!-- Change Password Modal -->
    <Teleport to="body">
      <div v-if="showChangePassword" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="showChangePassword = false">
        <div class="glass-panel rounded-2xl p-8 max-w-md w-full shadow-2xl animate-fadeIn" @click.stop>
          <h3 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-emerald-500 bg-clip-text text-transparent mb-6">Change Password</h3>

          <div v-if="passwordError" class="bg-red-500/10 border border-red-500/30 rounded-xl p-3 text-red-400 text-sm mb-4">
            {{ passwordError }}
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-kakoclaw-text-secondary mb-2">Current Password</label>
              <input v-model="passwordForm.oldPassword" type="password" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent" />
            </div>
            <div>
              <label class="block text-xs font-medium text-kakoclaw-text-secondary mb-2">New Password</label>
              <input v-model="passwordForm.newPassword" type="password" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent" />
            </div>
            <div>
              <label class="block text-xs font-medium text-kakoclaw-text-secondary mb-2">Confirm New Password</label>
              <input v-model="passwordForm.confirmPassword" type="password" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent" />
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showChangePassword = false" class="flex-1 bg-kakoclaw-bg/40 border border-kakoclaw-border text-kakoclaw-text py-2.5 rounded-xl font-medium hover:bg-kakoclaw-surface transition-all">
              Cancel
            </button>
            <button @click="changePassword" :disabled="changingPassword" class="flex-1 bg-kakoclaw-accent text-white py-2.5 rounded-xl font-bold hover:bg-kakoclaw-accent-hover transition-all shadow-lg shadow-kakoclaw-accent/20 disabled:opacity-50 flex items-center justify-center">
              <span v-if="changingPassword" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
              Change
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../../stores/authStore'

const authStore = useAuthStore()

const profile = ref({
  username: '',
  email: '',
  role: '',
  created_at: ''
})

const loading = ref(true)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const showChangePassword = ref(false)
const changingPassword = ref(false)
const passwordError = ref('')
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const formatDate = (dateString) => {
  if (!dateString) return 'Unknown'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

const loadProfile = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch('/api/v1/auth/profile', {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    })

    if (!response.ok) {
      throw new Error('Failed to load profile')
    }

    const data = await response.json()
    profile.value = {
      username: data.username || '',
      email: data.email || '',
      role: data.role || '',
      created_at: data.created_at || ''
    }
  } catch (err) {
    errorMessage.value = err.message
  } finally {
    loading.value = false
  }
}

const saveProfile = async () => {
  saving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const response = await fetch('/api/v1/auth/profile', {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: profile.value.username,
        email: profile.value.email
      })
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'Failed to update profile')
    }

    const data = await response.json()
    successMessage.value = 'Profile updated successfully!'

    // If new token is provided (username changed), update it
    if (data.token) {
      authStore.setToken(data.token)
      authStore.user.username = profile.value.username
      setTimeout(() => {
        window.location.reload()
      }, 1500)
    }
  } catch (err) {
    errorMessage.value = err.message
  } finally {
    saving.value = false
  }
}

const changePassword = async () => {
  passwordError.value = ''

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordError.value = 'New passwords do not match'
    return
  }

  if (passwordForm.value.newPassword.length < 8) {
    passwordError.value = 'Password must be at least 8 characters long'
    return
  }

  changingPassword.value = true

  try {
    const response = await fetch('/api/v1/auth/change-password', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        old_password: passwordForm.value.oldPassword,
        new_password: passwordForm.value.newPassword
      })
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'Failed to change password')
    }

    showChangePassword.value = false
    successMessage.value = 'Password changed successfully!'
    passwordForm.value = { oldPassword: '', newPassword: '', confirmPassword: '' }
  } catch (err) {
    passwordError.value = err.message
  } finally {
    changingPassword.value = false
  }
}

onMounted(() => {
  loadProfile()
})
</script>

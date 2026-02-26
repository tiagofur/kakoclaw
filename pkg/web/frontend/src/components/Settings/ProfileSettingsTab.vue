<template>
  <div class="space-y-8 max-w-2xl mx-auto animate-fade-in-up">
    <!-- Profile Card -->
    <div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full group-hover:bg-makoclaw-accent/20 transition-all duration-1000"></div>
      
      <div class="relative z-10">
        <div class="flex items-center gap-4 mb-10">
           <div class="p-3 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-blue-600 shadow-lg shadow-makoclaw-accent/20">
             <IconProfile class="w-6 h-6 text-white" />
           </div>
           <div>
             <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Bio-Identity</h3>
             <p class="text-[10px] font-bold text-makoclaw-accent uppercase tracking-widest mt-0.5">Core Account Profile</p>
           </div>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <div class="w-10 h-10 border-4 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin"></div>
        </div>

        <div v-else class="space-y-8">
          <div v-if="errorMessage" class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-xs font-bold uppercase tracking-widest">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="p-4 bg-makoclaw-success/10 border border-makoclaw-success/20 rounded-2xl text-makoclaw-success text-xs font-bold uppercase tracking-widest">
            {{ successMessage }}
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Username</label>
              <input v-model="profile.username" type="text" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all" />
              <p class="text-[9px] font-medium text-makoclaw-text-secondary/40 italic ml-1">* Identity change requires re-authentication</p>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Communication Link</label>
              <input v-model="profile.email" type="email" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all" placeholder="commander@mako.net" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-6 p-4 rounded-3xl bg-makoclaw-bg/20 border border-makoclaw-border/30">
            <div>
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 block mb-1">Authorization</span>
              <span class="px-2 py-0.5 text-[10px] font-black uppercase rounded-lg bg-makoclaw-accent/10 text-makoclaw-accent border border-makoclaw-accent/20">
                {{ profile.role }}
              </span>
            </div>
            <div>
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 block mb-1">Joined Matrix</span>
              <span class="text-[11px] font-black text-makoclaw-text italic">{{ formatDate(profile.created_at) }}</span>
            </div>
          </div>

          <div class="pt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
            <button @click="saveProfile" :disabled="saving" class="px-8 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center disabled:opacity-50 active:scale-95 group">
              <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-3"></span>
              <IconSave v-else class="w-4 h-4 mr-3 group-hover:scale-110 transition-transform" />
              Update Identity
            </button>

            <button @click="showChangePassword = true" class="px-8 py-4 bg-makoclaw-surface/50 border-2 border-makoclaw-border/50 text-makoclaw-text rounded-2xl font-black uppercase tracking-widest hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center group active:scale-95">
              <IconLock class="w-4 h-4 mr-3 group-hover:scale-110 transition-transform" />
              Secure Key
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Security Information Card -->
    <div class="glass-panel rounded-[2rem] p-6 bg-blue-500/5 border border-blue-500/10">
       <div class="flex items-start gap-4">
          <div class="p-2 rounded-xl bg-blue-500/10 text-blue-400 mt-1"><IconInfo class="w-4 h-4"/></div>
          <div>
            <h4 class="text-[11px] font-black uppercase tracking-widest text-makoclaw-text italic">Protocol Sync</h4>
            <p class="text-xs font-medium text-makoclaw-text-secondary/60 mt-1 leading-relaxed">Your profile syncs across all connected neural nodes including Telegram and Discord specialists. Ensure your communication link is active for emergency recovery.</p>
          </div>
       </div>
    </div>

    <!-- Change Password Modal -->
    <Teleport to="body">
      <Transition name="modal">
      <div v-if="showChangePassword" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-xl p-4" @click.self="showChangePassword = false">
        <div class="glass-panel rounded-[2.5rem] p-10 max-w-md w-full shadow-2xl border border-makoclaw-border/50 relative overflow-hidden animate-zoom" @click.stop>
          <div class="absolute -top-12 -left-12 w-48 h-48 bg-blue-500/10 blur-[60px] rounded-full pointer-events-none"></div>

          <h3 class="text-2xl font-black bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent italic mb-8">Update Shield Key</h3>

          <div v-if="passwordError" class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-[11px] font-black uppercase tracking-widest mb-6">
            {{ passwordError }}
          </div>

          <div class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Current Key</label>
              <input v-model="passwordForm.oldPassword" type="password" class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">New Neural Hash</label>
              <input v-model="passwordForm.newPassword" type="password" class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Verify Hash</label>
              <input v-model="passwordForm.confirmPassword" type="password" class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all" />
            </div>
          </div>

          <div class="flex gap-4 mt-10 pt-8 border-t border-makoclaw-border/30">
            <button @click="showChangePassword = false" class="flex-1 px-6 py-4 text-xs font-black uppercase tracking-widest text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Abort</button>
            <button @click="changePassword" :disabled="changingPassword" class="flex-1 px-6 py-4 bg-makoclaw-accent text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 active:scale-95 disabled:opacity-30 flex items-center justify-center">
              <span v-if="changingPassword" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-3"></span>
              Secure
            </button>
          </div>
        </div>
      </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { useAuthStore } from '../../stores/authStore'

const authStore = useAuthStore()

// Micro Icons
const IconProfile = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' })]) }
const IconSave = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconLock = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 00-2 2zm10-10V7a4 4 0 00-8 0v4h8z' })]) }
const IconInfo = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' })]) }

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
  if (!dateString) return 'Unknown Cycle'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

const loadProfile = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await fetch('/api/v1/auth/profile', {
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    })
    if (!response.ok) throw new Error('Identity Retrieval Failed')
    const data = await response.json()
    profile.value = {
      username: data.username || '',
      email: data.email || '',
      role: data.role || '',
      created_at: data.created_at || ''
    }
  } catch (err) { errorMessage.value = err.message } finally { loading.value = false }
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
    if (!response.ok) throw new Error('Identity Update Denied')
    const data = await response.json()
    successMessage.value = 'Identity Re-indexed'
    if (data.token) {
      const updatedUser = { ...(authStore.user || {}), username: profile.value.username || (authStore.user?.username || '') }
      authStore.setCredentials(updatedUser, data.token)
      setTimeout(() => { window.location.reload() }, 1500)
    }
  } catch (err) { errorMessage.value = err.message } finally { saving.value = false }
}

const changePassword = async () => {
  passwordError.value = ''
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordError.value = 'Hash Mismatch'
    return
  }
  if (passwordForm.value.newPassword.length < 8) {
    passwordError.value = 'Complexity Requirement Failed'
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
    if (!response.ok) throw new Error('Security Key Update Failed')
    showChangePassword.value = false
    successMessage.value = 'Neural Hash Secured'
    passwordForm.value = { oldPassword: '', newPassword: '', confirmPassword: '' }
  } catch (err) { passwordError.value = err.message } finally { changingPassword.value = false }
}

onMounted(() => { loadProfile() })
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-fade-in-up {
  animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.animate-zoom {
  animation: zoom 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes zoom {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.modal-enter-active, .modal-leave-active {
  transition: opacity 0.3s ease;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
}
</style>

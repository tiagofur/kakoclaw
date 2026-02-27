<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Decorative background blobs -->
    <div class="absolute -top-24 -right-24 w-96 h-96 bg-makoclaw-accent/5 blur-[100px] rounded-full pointer-events-none animate-pulse" />
    <div
      class="absolute bottom-0 -left-24 w-72 h-72 bg-blue-500/5 blur-[80px] rounded-full pointer-events-none"
      style="animation: float 20s infinite ease-in-out"
    />

    <!-- Mobile Header/Nav -->
    <div class="lg:hidden flex-none glass-sticky z-30 p-4 border-b border-makoclaw-border/50">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-xl font-black bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent italic">
            Settings
          </h2>
          <p class="text-[10px] font-medium text-makoclaw-text-secondary uppercase tracking-widest">
            Mission Configuration
          </p>
        </div>
        <div class="flex gap-2">
          <button
            class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border active:scale-90 transition-transform"
            @click="loadData"
          >
            <ArrowPathIcon class="w-4 h-4 text-makoclaw-text-secondary" />
          </button>
        </div>
      </div>
      
      <!-- Mobile Tabs (Horizontal Scroll) -->
      <div class="flex gap-2 overflow-x-auto pb-2 scrollbar-none snap-x">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="flex-none px-3 py-1.5 rounded-xl text-xs font-medium tracking-wide transition-all snap-start"
          :class="[activeTab === tab.key 
            ? 'bg-makoclaw-accent text-white shadow-lg shadow-makoclaw-accent/20' 
            : 'bg-makoclaw-surface/50 text-makoclaw-text-secondary border border-makoclaw-border/50 hover:border-makoclaw-accent/30']"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <!-- Desktop Sidebar Navigation -->
      <aside class="hidden lg:flex flex-col w-72 flex-none border-r border-makoclaw-border/10 p-6 space-y-6 z-20 bg-makoclaw-bg/20">

        <nav class="flex-1 space-y-1.5 overflow-y-auto custom-scrollbar pr-2">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="w-full flex items-center justify-between p-3 rounded-2xl transition-all duration-300 group relative overflow-hidden active:scale-[0.98]"
            :class="activeTab === tab.key 
              ? 'bg-makoclaw-accent text-white shadow-xl shadow-makoclaw-accent/20' 
              : 'text-makoclaw-text-secondary hover:bg-white/5 hover:text-makoclaw-text'"
            @click="activeTab = tab.key"
          >
            <div class="flex items-center gap-3 z-10">
              <div
                class="w-9 h-9 rounded-xl flex items-center justify-center transition-all duration-500 group-hover:scale-110"
                :class="activeTab === tab.key ? 'bg-white/20' : 'bg-makoclaw-surface border border-makoclaw-border/10 shadow-inner'"
              >
                <component
                  :is="getTabIcon(tab.key)"
                  class="w-4.5 h-4.5"
                />
              </div>
              <div class="flex flex-col items-start translate-y-px">
                <span class="text-[11px] font-black uppercase tracking-wider">
                  {{ tab.label }}
                </span>
              </div>
            </div>
            <div
              v-if="activeTab === tab.key"
              class="w-1.5 h-1.5 rounded-full bg-white shadow-[0_0_10px_rgba(255,255,255,0.8)] z-10 animate-pulse"
            />
          </button>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="flex-1 min-w-0 glass-panel border-0 border-l border-makoclaw-border/10">
        <!-- Tab Content -->
        <div class="h-full overflow-y-auto scrollbar-hide">
          <div class="max-w-4xl mx-auto p-6 md:p-10 animate-fade-in">
            <header class="mb-10">
              <div class="flex items-center gap-3 text-makoclaw-accent mb-2">
                <component
                  :is="activeTabIcon"
                  class="w-5 h-5"
                />
                <span class="text-[10px] font-bold uppercase tracking-[0.2em] leading-none">
                  {{ activeTabLabel }}
                </span>
              </div>
              <h2 class="text-3xl font-black text-makoclaw-text tracking-tight uppercase italic">
                {{ activeTabLabel }}
              </h2>
              <p class="text-makoclaw-text-secondary mt-2 text-sm font-medium">
                {{ activeTabDescription }}
              </p>
            </header>

            <div class="relative min-h-[400px]">
              <transition
                enter-active-class="transition duration-300 ease-out"
                enter-from-class="opacity-0 translate-y-4"
                enter-to-class="opacity-100 translate-y-0"
                leave-active-class="transition duration-200 ease-in"
                leave-from-class="opacity-100 translate-y-0"
                leave-to-class="opacity-0 -translate-y-4"
                mode="out-in"
              >
                <component
                  :is="activeTabComponent"
                  :key="activeTab"
                  :agents="configData?.agents || { defaults: {}, orchestrator: {}, specialists: {} }"
                  :providers="configData?.providers || {}"
                  :providers-list="providersList"
                  :saving="saving"
                  @save="saveConfig"
                  @toggle="toggleChannel"
                  @config="openChannelConfig"
                />
              </transition>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- Modals -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showChannelModal || showUserModal || showBlockModal || showUnblockModal"
          class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-makoclaw-bg/80 backdrop-blur-xl"
        >
          <!-- Shared Modal Wrapper -->
          <div class="glass-panel rounded-2xl shadow-xl w-full max-w-lg border border-makoclaw-border/50 overflow-hidden animate-zoom">
            <!-- Channel Config Modal -->
            <div
              v-if="showChannelModal"
              class="p-8"
            >
              <div class="flex justify-between items-center mb-5">
                <h3 class="text-xl font-black text-makoclaw-text italic flex items-center gap-3">
                  <span
                    class="p-2 bg-makoclaw-bg border border-makoclaw-border rounded-xl scale-90"
                  >
                    {{ selectedChannel?.icon }}
                  </span>
                  Configure {{ selectedChannel?.name }}
                </h3>
                <button
                  class="p-2 rounded-full hover:bg-white/5 transition-colors"
                  @click="showChannelModal = false"
                >
                  <XMarkIcon class="w-6 h-6" />
                </button>
              </div>
              
              <div class="space-y-6">
                <div
                  v-if="['telegram', 'discord'].includes(selectedChannel?.id)"
                  class="space-y-6"
                >
                  <div class="space-y-2">
                    <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">
                      Bot Token / Secret Code
                    </label>
                    <input
                      v-model="channelForm.token"
                      type="password"
                      class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none"
                    >
                  </div>
                  <div class="space-y-2">
                    <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">
                      Authorized Access Nodes
                    </label>
                    <input
                      v-model="channelForm.allow_from"
                      type="text"
                      placeholder="ID-1, ID-2, @username"
                      class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none"
                    >
                  </div>
                </div>
              </div>
              <div class="flex justify-end gap-3 mt-12 pt-8 border-t border-makoclaw-border/30">
                <button
                  class="px-4 py-2 text-xs font-medium text-makoclaw-text-secondary hover:text-makoclaw-text"
                  @click="showChannelModal = false"
                >
                  Abort
                </button>
                <button
                  class="px-5 py-2.5 bg-makoclaw-accent text-white rounded-xl text-xs font-semibold shadow-lg shadow-makoclaw-accent/20 active:scale-95"
                  @click="saveChannelConfig"
                >
                  Commit & Restart
                </button>
              </div>
            </div>

            <!-- User Modal -->
            <div
              v-if="showUserModal"
              class="p-8"
            >
              <h3 class="text-lg font-semibold text-makoclaw-text mb-5">
                {{ userForm.id ? 'Modify Identity' : 'Register Identity' }}
              </h3>
              <div class="space-y-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">
                    Username
                  </label>
                  <input
                    v-model="userForm.username"
                    :disabled="!!userForm.id"
                    class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 font-black text-makoclaw-text focus:border-makoclaw-accent outline-none disabled:opacity-40"
                  >
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">
                    Security Key
                  </label>
                  <input
                    v-model="userForm.password"
                    type="password"
                    class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-makoclaw-text focus:border-makoclaw-accent outline-none"
                  >
                </div>
              </div>
              <div class="flex justify-end gap-3 mt-10">
                <button
                  class="px-6 py-3 text-xs font-bold text-makoclaw-text-secondary"
                  @click="showUserModal = false"
                >
                  Cancel
                </button>
                <button
                  class="px-5 py-2.5 bg-makoclaw-accent text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20"
                  @click="saveUser"
                >
                  Commit
                </button>
              </div>
            </div>
            
            <!-- Block Modal -->
            <div
              v-if="showBlockModal"
              class="p-8 text-center"
            >
              <div class="w-16 h-16 bg-red-500/10 text-red-500 rounded-full flex items-center justify-center mx-auto mb-6">
                <NoSymbolIcon class="w-8 h-8" />
              </div>
              <h3 class="text-xl font-black text-makoclaw-text mb-4">
                Deactivate Node?
              </h3>
              <p class="text-sm font-medium text-makoclaw-text-secondary mb-5">
                This will instantly sever all neural links for @{{ blockForm.user?.username }}.
              </p>
              <textarea
                v-model="blockForm.reason"
                placeholder="Neural Severance Reason..."
                class="w-full bg-makoclaw-bg border border-makoclaw-border rounded-xl p-3 text-sm text-makoclaw-text outline-none focus:border-red-500 min-h-[100px] mb-5"
              />
              <div class="flex gap-4">
                <button
                  class="flex-1 py-2.5 text-xs font-medium text-makoclaw-text-secondary"
                  @click="showBlockModal = false"
                >
                  Maintain Link
                </button>
                <button
                  class="flex-1 py-2.5 bg-red-500 text-white rounded-xl text-xs font-semibold shadow-lg shadow-red-500/20"
                  @click="blockUser"
                >
                  Sever Connection
                </button>
              </div>
            </div>

            <!-- Unblock Modal -->
            <div
              v-if="showUnblockModal"
              class="p-8 text-center"
            >
              <div class="w-16 h-16 bg-makoclaw-success/10 text-makoclaw-success rounded-full flex items-center justify-center mx-auto mb-6">
                <CheckCircleIcon class="w-8 h-8" />
              </div>
              <h3 class="text-xl font-black text-makoclaw-text mb-4">
                Restore Node Integration?
              </h3>
              <p class="text-sm font-medium text-makoclaw-text-secondary mb-5">
                Re-establishing connection for @{{ unblockForm.user?.username }}...
              </p>
              <div class="flex gap-4">
                <button
                  class="flex-1 py-2.5 text-xs font-medium text-makoclaw-text-secondary"
                  @click="showUnblockModal = false"
                >
                  Leave Offline
                </button>
                <button
                  class="flex-1 py-2.5 bg-makoclaw-success text-white rounded-xl text-xs font-semibold shadow-lg shadow-green-500/20"
                  @click="unblockUser"
                >
                  Activate Neural Link
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { 
  UserIcon, 
  ShieldCheckIcon, 
  BeakerIcon, 
  ChatBubbleLeftRightIcon, 
  Cog6ToothIcon,
  ArrowPathIcon,
  UserGroupIcon,
  DocumentMagnifyingGlassIcon,
  GlobeAltIcon,
  XMarkIcon,
  NoSymbolIcon,
  CheckCircleIcon
} from '@heroicons/vue/24/outline'

import ProfileSettingsTab from '../components/Settings/ProfileSettingsTab.vue'
import AgentSettingsTab from '../components/Settings/AgentSettingsTab.vue'
import ProvidersSettingsTab from '../components/Settings/ProvidersSettingsTab.vue'
import ChannelsSettingsTab from '../components/Settings/ChannelsSettingsTab.vue'
import ToolPermissionsTab from '../components/Settings/ToolPermissionsTab.vue'
import AuditLogTab from '../components/Settings/AuditLogTab.vue'

import advancedService from '../services/advancedService'
import usersService from '../services/usersService'
import { useToast } from '../composables/useToast'
import { useAuthStore } from '../stores/authStore'

const toast = useToast()
const authStore = useAuthStore()
const route = useRoute()
const loading = ref(true)
const saving = ref(false)
const configData = ref(null)
const providersList = ref([])
const activeTab = ref('profile')

// Tabs Logic
const tabs = computed(() => {
  const base = [
    { key: 'profile', label: 'Profile' },
    { key: 'agents', label: 'Agents' },
    { key: 'providers', label: 'Providers' },
    { key: 'channels', label: 'Channels' }
  ]
  if (authStore.user?.role === 'admin') {
    base.push({ key: 'users', label: 'Node Management' })
    base.push({ key: 'permissions', label: 'Safety Protocols' })
    base.push({ key: 'audit', label: 'Pulse Logs' })
  }
  base.push({ key: 'system', label: 'Core System' })
  return base
})

const getTabIcon = (key) => {
  const map = {
    profile: UserIcon,
    agents: BeakerIcon,
    providers: Cog6ToothIcon,
    channels: ChatBubbleLeftRightIcon,
    users: UserGroupIcon,
    permissions: ShieldCheckIcon,
    audit: DocumentMagnifyingGlassIcon,
    system: GlobeAltIcon
  }
  return map[key] || GlobeAltIcon
}

const activeTabLabel = computed(() => tabs.value.find(t => t.key === activeTab.value)?.label || 'Settings')

const activeTabDescription = computed(() => {
  const map = {
    profile: 'Manage your primary mission profile',
    agents: 'Configure neural specialist parameters',
    providers: 'Connect external knowledge vectors',
    channels: 'Manage communication protocols',
    users: 'Control node access and identities',
    permissions: 'Define tool safety boundaries',
    audit: 'Review mission activity logs',
    system: 'Update core system parameters'
  }
  return map[activeTab.value] || 'System configuration'
})

const activeTabIcon = computed(() => getTabIcon(activeTab.value))

const activeTabComponent = computed(() => {
  const map = {
    profile: ProfileSettingsTab,
    agents: AgentSettingsTab,
    providers: ProvidersSettingsTab,
    channels: ChannelsSettingsTab,
    users: null,
    permissions: ToolPermissionsTab,
    audit: AuditLogTab
  }
  return map[activeTab.value] || null
})

const setActiveTabFromRoute = () => {
  const requestedTab = String(route.query.tab || '')
  if (requestedTab && tabs.value.some((tab) => tab.key === requestedTab)) {
    activeTab.value = requestedTab
  }
}

// Data Fetching & State
const usersList = ref([])
const showUserModal = ref(false)
const userForm = ref({})
const savingUser = ref(false)
const showBlockModal = ref(false)
const showUnblockModal = ref(false)
const blockForm = ref({ user: null, reason: '' })
const unblockForm = ref({ user: null })
const showChannelModal = ref(false)
const selectedChannel = ref(null)
const channelForm = ref({})

const loadData = async () => {
  loading.value = true
  try {
    const promises = [advancedService.fetchUserConfig(), advancedService.fetchModels()]
    if (authStore.user?.role === 'admin') promises.push(usersService.listUsers())
    const results = await Promise.all(promises)
    configData.value = results[0].config || {}
    providersList.value = results[1].providers || []
    if (results[2]) usersList.value = results[2]
  } catch {
    toast.error('Sync failed')
  } finally {
    loading.value = false
  }
}

const saveConfig = async (payload) => {
  saving.value = true
  try {
    const hasSystemSettings = payload.web || payload.gateway || payload.storage
    const hasUserSettings = payload.tools || payload.channels || payload.agents || payload.providers
    if (hasSystemSettings && authStore.user?.role === 'admin') await advancedService.updateConfig(payload)
    if (hasUserSettings) await advancedService.updateUserConfig(payload)
    toast.success('Matrix updated')
    setTimeout(loadData, 500)
  } catch (err) {
    toast.error('Update rejected: ' + err.message)
  } finally {
    saving.value = false
  }
}

const saveUser = async () => {
  savingUser.value = true
  try {
    if (userForm.value.id) await usersService.updateUser(userForm.value.id, userForm.value.password, userForm.value.role)
    else await usersService.createUser(userForm.value.username, userForm.value.password, userForm.value.role)
    toast.success('Record saved')
    showUserModal.value = false
    loadData()
  } catch {
    toast.error('Save failed')
  } finally {
    savingUser.value = false
  }
}

const blockUser = async () => { 
  await usersService.blockUser(blockForm.value.user.id, blockForm.value.reason)
  showBlockModal.value = false
  loadData() 
}

const unblockUser = async () => { 
  await usersService.unblockUser(unblockForm.value.user.id)
  showUnblockModal.value = false
  loadData() 
}

const saveChannelConfig = async () => { 
  const payload = { channels: { [selectedChannel.value.id]: { ...channelForm.value, enabled: true } } }
  await saveConfig(payload)
  showChannelModal.value = false 
}

const openChannelConfig = (c) => { 
  selectedChannel.value = c
  channelForm.value = { token: '', allow_from: '' }
  showChannelModal.value = true 
}

const toggleChannel = async (id) => { 
  await saveConfig({ channels: { [id]: { enabled: !configData.value.channels[id]?.enabled } } }) 
}

onMounted(() => { 
  setActiveTabFromRoute()
  loadData() 
})

watch(() => route.query.tab, setActiveTabFromRoute)
</script>

<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-20px) translateX(10px); }
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-fade-in {
  animation: fade-in 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.scrollbar-none::-webkit-scrollbar { display: none; }
.scrollbar-none { -ms-overflow-style: none; scrollbar-width: none; }

.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(156, 163, 175, 0.1); border-radius: 10px; }

.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>

<template>
  <div class="h-full flex flex-col bg-picoclaw-bg">
    <!-- Header -->
    <header class="flex items-center justify-between px-8 py-6 border-b border-picoclaw-border bg-picoclaw-surface sticky top-0 z-10 shadow-sm">
      <div class="flex items-center space-x-4">
        <h1 class="text-2xl font-semibold text-picoclaw-text tracking-tight">Channels</h1>
        <div 
          class="px-3 py-1 rounded-full text-xs font-medium bg-picoclaw-bg border border-picoclaw-border text-picoclaw-text-secondary transition-colors"
        >
          Connected messaging platforms and their status
        </div>
      </div>
      
      <div class="flex items-center space-x-3">
        <button 
          @click="loadChannels" 
          class="p-2 text-picoclaw-text-secondary hover:text-picoclaw-text hover:bg-picoclaw-bg rounded-lg transition-colors border border-transparent hover:border-picoclaw-border"
          title="Refresh Status"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v3.276a1 1 0 01-2 0V14.907a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-8 space-y-8">
      <!-- Active Channels Banner -->
      <div v-if="activeChannels.length > 0" class="bg-gradient-to-r from-picoclaw-accent/10 to-transparent border-l-4 border-picoclaw-accent p-4 rounded-r-lg mb-6">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <span class="inline-flex items-center justify-center h-8 w-8 rounded-full bg-picoclaw-accent/20 text-picoclaw-accent">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
            </span>
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-picoclaw-accent">Active channels</h3>
            <div class="mt-1 text-sm text-picoclaw-text-secondary">
              <span v-for="(channel, index) in activeChannels" :key="channel" class="font-medium text-picoclaw-text">
                {{ formatChannelName(channel) }}<span v-if="index < activeChannels.length - 1">, </span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="loading && !channelsConfig" class="text-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-picoclaw-accent mx-auto"></div>
        <p class="mt-4 text-picoclaw-text-secondary">Loading channel status...</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <!-- Channel Cards -->
        <div 
          v-for="channel in availableChannels" 
          :key="channel.id"
          class="bg-picoclaw-surface rounded-xl border transition-all duration-200"
          :class="[
            isChannelActive(channel.id) 
              ? 'border-picoclaw-accent/50 shadow-md ring-1 ring-picoclaw-accent/20' 
              : 'border-picoclaw-border hover:border-picoclaw-border/80 hover:shadow-sm'
          ]"
        >
          <div class="p-6">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center space-x-3">
                <div 
                  class="w-10 h-10 rounded-lg flex items-center justify-center transition-colors px-2"
                  :class="isChannelActive(channel.id) ? 'bg-picoclaw-accent text-white' : 'bg-picoclaw-bg text-picoclaw-text-secondary'"
                  v-html="channel.icon"
                >
                </div>
                <div>
                  <h3 class="font-medium text-picoclaw-text">{{ channel.name }}</h3>
                  <p class="text-xs text-picoclaw-text-secondary">{{ isChannelActive(channel.id) ? 'Connected' : 'Not configured' }}</p>
                </div>
              </div>
              
              <!-- Toggle Switch -->
              <button 
                @click="openConfigModal(channel)"
                class="relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-picoclaw-accent"
                :class="isChannelActive(channel.id) ? 'bg-picoclaw-success' : 'bg-picoclaw-border'"
              >
                <span class="sr-only">Use setting</span>
                <span 
                  aria-hidden="true" 
                  class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                  :class="isChannelActive(channel.id) ? 'translate-x-5' : 'translate-x-0'"
                ></span>
              </button>
            </div>
            
            <p class="text-sm text-picoclaw-text-secondary mb-4 min-h-[40px]">{{ channel.description }}</p>
            
            <div class="flex justify-between items-center pt-4 border-t border-picoclaw-border/50">
              <button 
                @click="openConfigModal(channel)"
                class="text-xs font-medium transition-colors flex items-center"
                :class="isChannelActive(channel.id) ? 'text-picoclaw-accent hover:text-picoclaw-accent-hover' : 'text-picoclaw-text-secondary hover:text-picoclaw-text'"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
                </svg>
                Configure
              </button>
              
              <span v-if="isChannelActive(channel.id)" class="flex h-2 w-2 relative">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-picoclaw-success opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-picoclaw-success"></span>
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Empty State -->
      <div v-if="!loading && activeChannels.length === 0" class="text-center py-12 bg-picoclaw-surface rounded-xl border border-dashed border-picoclaw-border">
        <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-12 w-12 text-picoclaw-text-secondary/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-picoclaw-text">No active channels</h3>
        <p class="mt-1 text-sm text-picoclaw-text-secondary">Configure a channel to start interacting with your agent from other platforms.</p>
      </div>

    </div>

    <!-- Configuration Modal -->
    <div v-if="showConfigModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div class="bg-picoclaw-surface rounded-xl shadow-2xl w-full max-w-md border border-picoclaw-border transform transition-all">
        <div class="flex justify-between items-center p-6 border-b border-picoclaw-border">
          <h3 class="text-lg font-medium text-picoclaw-text flex items-center">
            <span class="w-5 h-5 mr-2 text-picoclaw-text-secondary" v-html="selectedChannel?.icon"></span>
            Configure {{ selectedChannel?.name }}
          </h3>
          <button @click="closeConfigModal" class="text-picoclaw-text-secondary hover:text-picoclaw-text transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between mb-4">
            <label class="text-sm font-medium text-picoclaw-text">Enable Channel</label>
            <button 
              @click="configForm.enabled = !configForm.enabled"
              class="relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-picoclaw-accent"
              :class="configForm.enabled ? 'bg-picoclaw-success' : 'bg-picoclaw-border'"
            >
              <span 
                aria-hidden="true" 
                class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                :class="configForm.enabled ? 'translate-x-5' : 'translate-x-0'"
              ></span>
            </button>
          </div>

          <!-- Telegram Config -->
          <div v-if="selectedChannel?.id === 'telegram'" class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">Bot Token</label>
              <input 
                v-model="configForm.token" 
                type="password"
                :placeholder="channelsConfig['telegram']?.configured ? '•••••••••••••••• (Configured)' : '123456:ABC-DEF1234...'"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
              />
              <p class="mt-1 text-xs text-picoclaw-text-secondary">Get this from @BotFather. Leave blank to keep existing.</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">Allowed IDs / Usernames (comma separated)</label>
              <input 
                v-model="configForm.allow_from" 
                type="text"
                placeholder="12345678,your_username"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
              />
              <p class="mt-1 text-xs text-picoclaw-text-secondary">Only users in this list can interact with the bot. Leave empty to allow everyone.</p>
            </div>
          </div>

          <!-- Discord Config -->
          <div v-if="selectedChannel?.id === 'discord'" class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">Bot Token</label>
              <input 
                v-model="configForm.token" 
                type="password"
                :placeholder="channelsConfig['discord']?.configured ? '•••••••••••••••• (Configured)' : 'OT...X.Y.Z'"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
              />
              <p class="mt-1 text-xs text-picoclaw-text-secondary">From Discord Developer Portal. Leave blank to keep existing.</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">Allowed Server/User IDs (comma separated)</label>
              <input 
                v-model="configForm.allow_from" 
                type="text"
                placeholder="9876543210"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
              />
            </div>
          </div>

          <!-- Slack Config -->
          <div v-if="selectedChannel?.id === 'slack'" class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">Bot Token (xoxb-...)</label>
              <input 
                v-model="configForm.bot_token" 
                type="password"
                placeholder="xoxb-..."
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
              />
            </div>
          </div>
          
           <!-- Generic Token Config (Default fallback) -->
           <div v-if="!['telegram', 'discord', 'slack'].includes(selectedChannel?.id)" class="space-y-4">
            <div class="p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
                <p class="text-xs text-yellow-500">UI configuration for this channel is limited. Please verify settings manually if advanced options are needed.</p>
            </div>
             <!-- Try to guess fields based on channel -->
             <div v-if="selectedChannel?.configFields" v-for="field in selectedChannel.configFields" :key="field.key">
                <label class="block text-xs font-medium text-picoclaw-text-secondary mb-1">{{ field.label }}</label>
                <input 
                    v-model="configForm[field.key]" 
                    :type="field.type || 'text'"
                    class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-picoclaw-text text-sm focus:outline-none focus:ring-2 focus:ring-picoclaw-accent/50"
                />
             </div>
          </div>

        </div>

        <div class="flex justify-end space-x-3 p-6 border-t border-picoclaw-border bg-picoclaw-bg/50 rounded-b-xl">
          <button 
            @click="closeConfigModal" 
            class="px-4 py-2 text-sm text-picoclaw-text-secondary hover:text-picoclaw-text transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="saveConfig" 
            :disabled="saving"
            class="px-4 py-2 text-sm bg-picoclaw-accent text-white rounded-lg hover:bg-picoclaw-accent-hover transition-colors disabled:opacity-50 flex items-center"
          >
            <span v-if="saving" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
            {{ saving ? 'Saving...' : 'Save & Restart' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const channelsStatus = ref({})
const activeChannels = ref([])
const showConfigModal = ref(false)
const selectedChannel = ref(null)
const configForm = ref({})
const saving = ref(false)

const chatIcon = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H8.25m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H12m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 0 1-2.555-.337A5.972 5.972 0 0 1 5.41 20.97a5.969 5.969 0 0 1-.474-.065 4.48 4.48 0 0 0 .978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25Z" /></svg>'
const hashIcon = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h13.5m-13.5 7.5h13.5m-3-10.5-3 15m-3-15-3 15" /></svg>'

// Available Channels Definition
const availableChannels = [
  { 
    id: 'telegram', 
    name: 'Telegram', 
    icon: chatIcon,
    description: 'Chat with your agent via Telegram bot.'
  },
  { 
    id: 'discord', 
    name: 'Discord', 
    icon: hashIcon,
    description: 'Connect your agent to a Discord server.'
  },
  { 
    id: 'slack', 
    name: 'Slack', 
    icon: hashIcon,
    description: 'Integrate into your Slack workspace.'
  },
  { 
    id: 'whatsapp', 
    name: 'WhatsApp', 
    icon: chatIcon,
    description: 'Connect via WhatsApp bridge.',
    configFields: [
        { key: 'bridge_url', label: 'Bridge URL' }
    ]
  },
  { 
    id: 'signal', 
    name: 'Signal', 
    icon: chatIcon,
    description: 'Secure messaging via Signal.',
    configFields: [
        { key: 'phone_number', label: 'Phone Number' }
    ]
  },
   {
    id: 'feishu',
    name: 'Feishu / Lark',
    icon: chatIcon,
    description: 'Enterprise collaboration platform.',
    configFields: [
        { key: 'app_id', label: 'App ID' },
        { key: 'app_secret', label: 'App Secret', type: 'password' }
    ]
   }
]

// Current Config from Backend
const channelsConfig = ref({})

const loadChannels = async () => {
    loading.value = true
    try {
        const statusData = await advancedService.fetchChannels()
        channelsStatus.value = statusData.channels || {}
        activeChannels.value = statusData.enabled || []
        
        // Fetch raw config for form population (using dedicated GET method)
        const configData = await advancedService.fetchConfig()
        if (configData && configData.config && configData.config.channels) {
            channelsConfig.value = configData.config.channels
        }
    } catch (err) {
        console.error("Failed to load channels", err)
        toast.error("Failed to load channel status")
    } finally {
        loading.value = false
    }
}

const isChannelActive = (id) => {
    return activeChannels.value.includes(id) || channelsStatus.value[id]?.running
}

const formatChannelName = (id) => {
    const ch = availableChannels.find(c => c.id === id)
    return ch ? ch.name : id.charAt(0).toUpperCase() + id.slice(1)
}

const openConfigModal = (channel) => {
    selectedChannel.value = channel
    
    // Get existing config if any
    const existing = channelsConfig.value[channel.id] || {}
    
    // Initialize form with existing values (redacted secrets will be empty)
    configForm.value = {
        enabled: isChannelActive(channel.id),
        ...existing
    }
    
    // Ensure tokens/secrets are empty in UI if just loaded (they are redacted to "" or configured:true anyway)
    // but the backend might return "" for secrets, so let's be safe.
    showConfigModal.value = true
}

const closeConfigModal = () => {
    showConfigModal.value = false
    selectedChannel.value = null
    configForm.value = {}
}

const saveConfig = async () => {
    if (!selectedChannel.value) return
    
    saving.value = true
    try {
        const channelId = selectedChannel.value.id
        
        // Prepare update payload
        const updates = {
            channels: {
                [channelId]: {
                    ...configForm.value
                }
            }
        }
        
        await advancedService.updateConfig(updates)
        
        toast.success(`${selectedChannel.value.name} configuration updated`)
        closeConfigModal()
        // Wait a moment for restart then reload status
        setTimeout(loadChannels, 1000)
        
    } catch (err) {
        const msg = err.response?.data?.error || err.message
        toast.error(`Failed to save config: ${msg}`)
    } finally {
        saving.value = false
    }
}

onMounted(() => {
    loadChannels()
})
</script>

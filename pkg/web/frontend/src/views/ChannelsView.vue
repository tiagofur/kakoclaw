<template>
  <div class="h-full flex flex-col bg-picoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-picoclaw-border bg-picoclaw-surface">
      <h2 class="text-xl font-bold bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent">Channels</h2>
      <p class="text-sm text-picoclaw-text-secondary mt-1">Connected messaging platforms and their status</p>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-picoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Enabled Channels Summary -->
        <div class="mb-6 px-4 py-3 rounded-lg border bg-picoclaw-surface border-picoclaw-border">
          <span class="font-medium">Active channels: </span>
          <span v-if="enabled.length > 0" class="text-emerald-400">{{ enabled.join(', ') }}</span>
          <span v-else class="text-picoclaw-text-secondary">None configured</span>
        </div>

        <!-- Channel Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="(info, name) in channelStatus"
            :key="name"
            class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5 relative overflow-hidden"
          >
            <!-- Status indicator -->
            <div
              class="absolute top-0 left-0 w-full h-1"
              :class="{
                'bg-emerald-500': info.running,
                'bg-yellow-500': info.enabled && !info.running,
                'bg-gray-600': !info.enabled
              }"
            ></div>

            <div class="flex items-center gap-3 mt-2">
              <div
                class="w-10 h-10 rounded-lg flex items-center justify-center text-lg font-bold"
                :class="channelColor(name)"
              >{{ channelIcon(name) }}</div>
              <div>
                <h3 class="font-semibold capitalize">{{ name }}</h3>
                <span class="text-xs"
                  :class="{
                    'text-emerald-400': info.running,
                    'text-yellow-400': info.enabled && !info.running,
                    'text-gray-500': !info.enabled
                  }"
                >{{ info.running ? 'Connected' : info.enabled ? 'Enabled (not connected)' : 'Disabled' }}</span>
              </div>
            </div>

            <div class="mt-3 text-xs text-picoclaw-text-secondary">
              <span v-if="info.enabled">Enabled in config</span>
              <span v-else>Not enabled in config.json</span>
            </div>
          </div>
        </div>

        <!-- Empty state if no channels at all -->
        <div v-if="Object.keys(channelStatus).length === 0" class="text-center py-12 text-picoclaw-text-secondary">
          <p class="text-lg">No channel data available</p>
          <p class="text-sm mt-2">Channel management requires the full gateway mode</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const channelStatus = ref({})
const enabled = ref([])

const channelColor = (name) => {
  const colors = {
    telegram: 'bg-blue-500/20 text-blue-400',
    discord: 'bg-indigo-500/20 text-indigo-400',
    slack: 'bg-purple-500/20 text-purple-400',
    whatsapp: 'bg-emerald-500/20 text-emerald-400',
    feishu: 'bg-cyan-500/20 text-cyan-400',
    dingtalk: 'bg-blue-500/20 text-blue-400',
    qq: 'bg-sky-500/20 text-sky-400',
    maixcam: 'bg-orange-500/20 text-orange-400',
    signal: 'bg-blue-600/20 text-blue-400'
  }
  return colors[name] || 'bg-gray-500/20 text-gray-400'
}

const channelIcon = (name) => {
  const icons = {
    telegram: 'TG',
    discord: 'DC',
    slack: 'SL',
    whatsapp: 'WA',
    feishu: 'FS',
    dingtalk: 'DT',
    qq: 'QQ',
    maixcam: 'MX',
    signal: 'SG'
  }
  return icons[name] || name.substring(0, 2).toUpperCase()
}

onMounted(async () => {
  try {
    const data = await advancedService.fetchChannels()
    channelStatus.value = data.channels || {}
    enabled.value = data.enabled || []
  } catch (err) {
    console.error('Failed to load channels:', err)
    toast.error('Failed to load channels')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
</style>

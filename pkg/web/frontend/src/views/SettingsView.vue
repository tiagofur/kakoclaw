<template>
  <div class="h-full flex flex-col bg-picoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-picoclaw-border bg-picoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent">Settings</h2>
        <p class="text-sm text-picoclaw-text-secondary mt-1">System configuration (read-only)</p>
      </div>
      <div class="flex bg-picoclaw-bg rounded-lg p-1 border border-picoclaw-border">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="px-3 py-1.5 rounded-md text-sm font-medium transition-all"
          :class="activeTab === tab.key ? 'bg-white dark:bg-gray-700 shadow-sm text-picoclaw-accent' : 'text-picoclaw-text-secondary hover:text-picoclaw-text'"
        >{{ tab.label }}</button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-picoclaw-accent"></div>
      </div>

      <template v-else-if="configData">
        <!-- Agent Settings -->
        <div v-if="activeTab === 'agents'" class="space-y-4">
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5">
            <h3 class="font-semibold mb-4">Agent Defaults</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="(val, key) in configData.agents?.defaults || {}" :key="key" class="flex justify-between items-center py-2 border-b border-picoclaw-border last:border-0">
                <span class="text-sm text-picoclaw-text-secondary">{{ formatKey(key) }}</span>
                <span class="text-sm font-mono">{{ val || '(empty)' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Providers -->
        <div v-if="activeTab === 'providers'" class="space-y-3">
          <div
            v-for="(info, name) in configData.providers || {}"
            :key="name"
            class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5"
          >
            <div class="flex items-center justify-between">
              <h3 class="font-semibold capitalize">{{ name }}</h3>
              <span
                class="px-2 py-0.5 text-xs rounded-full"
                :class="info.configured ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'"
              >{{ info.configured ? 'Configured' : 'Not configured' }}</span>
            </div>
            <div class="mt-3 text-sm space-y-1">
              <div class="flex justify-between">
                <span class="text-picoclaw-text-secondary">API Key</span>
                <span class="font-mono">{{ info.api_key || '(not set)' }}</span>
              </div>
              <div v-if="info.api_base" class="flex justify-between">
                <span class="text-picoclaw-text-secondary">API Base</span>
                <span class="font-mono text-xs">{{ info.api_base }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Channels Config -->
        <div v-if="activeTab === 'channels'" class="space-y-3">
          <div
            v-for="(info, name) in configData.channels || {}"
            :key="name"
            class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5"
          >
            <div class="flex items-center justify-between">
              <h3 class="font-semibold capitalize">{{ name }}</h3>
              <div class="flex gap-2">
                <span
                  class="px-2 py-0.5 text-xs rounded-full"
                  :class="info.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'"
                >{{ info.enabled ? 'Enabled' : 'Disabled' }}</span>
                <span v-if="info.configured !== undefined"
                  class="px-2 py-0.5 text-xs rounded-full"
                  :class="info.configured ? 'bg-blue-500/10 text-blue-400' : 'bg-yellow-500/10 text-yellow-400'"
                >{{ info.configured ? 'Configured' : 'Not configured' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Web & Gateway -->
        <div v-if="activeTab === 'system'" class="space-y-4">
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5">
            <h3 class="font-semibold mb-4">Web Server</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="(val, key) in configData.web || {}" :key="key" class="flex justify-between items-center py-2 border-b border-picoclaw-border last:border-0">
                <span class="text-sm text-picoclaw-text-secondary">{{ formatKey(key) }}</span>
                <span class="text-sm font-mono">{{ String(val) }}</span>
              </div>
            </div>
          </div>
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5">
            <h3 class="font-semibold mb-4">Gateway</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="(val, key) in configData.gateway || {}" :key="key" class="flex justify-between items-center py-2 border-b border-picoclaw-border last:border-0">
                <span class="text-sm text-picoclaw-text-secondary">{{ formatKey(key) }}</span>
                <span class="text-sm font-mono">{{ String(val) }}</span>
              </div>
            </div>
          </div>
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5">
            <h3 class="font-semibold mb-4">Storage</h3>
            <div v-for="(val, key) in configData.storage || {}" :key="key" class="flex justify-between items-center py-2">
              <span class="text-sm text-picoclaw-text-secondary">{{ formatKey(key) }}</span>
              <span class="text-sm font-mono">{{ val }}</span>
            </div>
          </div>
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5">
            <h3 class="font-semibold mb-4">Tools</h3>
            <div class="space-y-2">
              <div class="flex justify-between items-center py-2 border-b border-picoclaw-border">
                <span class="text-sm text-picoclaw-text-secondary">Web Search API Key</span>
                <span class="text-sm font-mono">{{ configData.tools?.web?.search?.api_key || '(not set)' }}</span>
              </div>
              <div class="flex justify-between items-center py-2 border-b border-picoclaw-border">
                <span class="text-sm text-picoclaw-text-secondary">Web Search Max Results</span>
                <span class="text-sm font-mono">{{ configData.tools?.web?.search?.max_results || 0 }}</span>
              </div>
              <div class="flex justify-between items-center py-2">
                <span class="text-sm text-picoclaw-text-secondary">Email Enabled</span>
                <span class="text-sm font-mono">{{ configData.tools?.email?.enabled ? 'Yes' : 'No' }}</span>
              </div>
            </div>
          </div>
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
const configData = ref(null)
const activeTab = ref('agents')

const tabs = [
  { key: 'agents', label: 'Agent' },
  { key: 'providers', label: 'Providers' },
  { key: 'channels', label: 'Channels' },
  { key: 'system', label: 'System' }
]

const formatKey = (key) => {
  return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

onMounted(async () => {
  try {
    const data = await advancedService.fetchConfig()
    configData.value = data.config || {}
  } catch (err) {
    console.error('Failed to load config:', err)
    toast.error('Failed to load configuration')
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

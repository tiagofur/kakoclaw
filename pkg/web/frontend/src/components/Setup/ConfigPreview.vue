<template>
  <div class="config-preview space-y-6">
    <div>
      <h3 class="text-xl font-bold text-white mb-4">Review Your Configuration</h3>
      <p class="text-slate-400 mb-6">
        Please review the settings below before we save them to your account
      </p>
    </div>

    <!-- Provider summary -->
    <div class="bg-slate-700/50 rounded-lg border border-slate-600 p-6 space-y-4">
      <div class="flex items-center justify-between mb-4 pb-4 border-b border-slate-600">
        <h4 class="text-lg font-bold text-white">🤖 AI Provider</h4>
        <button
          @click="$emit('edit', 1)"
          class="text-blue-400 hover:text-blue-300 text-sm font-medium"
        >
          Edit
        </button>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <p class="text-slate-400 text-sm">Provider Type</p>
          <p class="text-white font-bold capitalize">{{ config.provider.type }}</p>
        </div>
        <div>
          <p class="text-slate-400 text-sm">API Key</p>
          <p class="text-white font-mono text-sm">{{ maskedApiKey }}</p>
        </div>
        <div v-if="config.provider.model" class="col-span-2">
          <p class="text-slate-400 text-sm">Model</p>
          <p class="text-white font-mono text-sm">{{ config.provider.model }}</p>
        </div>
      </div>

      <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-3 mt-4">
        <p class="text-blue-300 text-sm">
          ✓ Provider configuration looks good
        </p>
      </div>
    </div>

    <!-- Channel summary -->
    <div class="bg-slate-700/50 rounded-lg border border-slate-600 p-6 space-y-4">
      <div class="flex items-center justify-between mb-4 pb-4 border-b border-slate-600">
        <h4 class="text-lg font-bold text-white">💬 Communication Channel</h4>
        <button
          @click="$emit('edit', 2)"
          class="text-blue-400 hover:text-blue-300 text-sm font-medium"
        >
          Edit
        </button>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <p class="text-slate-400 text-sm">Channel Type</p>
          <p class="text-white font-bold capitalize">{{ config.channel.type }}</p>
        </div>
        <div>
          <p class="text-slate-400 text-sm">Bot Token</p>
          <p class="text-white font-mono text-sm">{{ maskedBotToken }}</p>
        </div>
        <div v-if="config.channel.channelId">
          <p class="text-slate-400 text-sm">Channel ID</p>
          <p class="text-white font-mono text-sm">{{ config.channel.channelId }}</p>
        </div>
        <div v-if="config.channel.webhookUrl">
          <p class="text-slate-400 text-sm">Webhook</p>
          <p class="text-white font-mono text-sm text-xs truncate">{{ config.channel.webhookUrl }}</p>
        </div>
      </div>

      <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-3 mt-4">
        <p class="text-blue-300 text-sm">
          ✓ Channel configuration looks good
        </p>
      </div>
    </div>

    <!-- Security info -->
    <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4 space-y-2">
      <p class="text-blue-300 text-sm font-bold">🔒 Security reminder:</p>
      <ul class="text-blue-300/80 text-sm space-y-1">
        <li>• API keys and tokens are encrypted at rest</li>
        <li>• Only transmitted securely via HTTPS</li>
        <li>• Never logged or exposed in debug output</li>
        <li>• You can revoke access anytime in Settings</li>
      </ul>
    </div>

    <!-- Ready to save -->
    <div class="bg-gradient-to-r from-blue-500/20 to-blue-500/20 border border-blue-500/50 rounded-lg p-6 text-center">
      <p class="text-white text-lg font-bold mb-2">Ready to get started! 🚀</p>
      <p class="text-slate-300 text-sm">
        Click "Finish Setup" to save your configuration and start using makoclaw
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  config: {
    type: Object,
    required: true
  }
})

const maskedApiKey = computed(() => {
  const key = props.config.provider.apiKey
  if (!key) return '(not set)'
  if (key.length <= 8) return key
  return key.substring(0, 4) + '••••••••' + key.substring(key.length - 4)
})

const maskedBotToken = computed(() => {
  const token = props.config.channel.botToken
  if (!token) return '(not set)'
  if (token.length <= 8) return token
  return token.substring(0, 4) + '••••••••' + token.substring(token.length - 4)
})

defineEmits(['edit'])
</script>

<style scoped>
</style>



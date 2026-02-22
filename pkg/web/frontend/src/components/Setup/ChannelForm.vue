<template>
  <div class="channel-form space-y-6">
    <div>
      <h3 class="text-xl font-bold text-white mb-4">Connect a Communication Channel</h3>
      <p class="text-slate-400 mb-6">
        Choose where your agent will listen for messages and respond
      </p>

      <div class="grid grid-cols-1 gap-3 mb-6">
        <button
          v-for="channel in channels"
          :key="channel.id"
          @click="selectChannel(channel)"
          :class="[
            'p-4 rounded-lg border-2 transition-all text-left',
            selectedChannel?.id === channel.id
              ? 'border-blue-500 bg-blue-900/20'
              : 'border-slate-600 bg-slate-700/50 hover:border-slate-500'
          ]"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="font-bold text-white">{{ channel.name }}</p>
              <p class="text-sm text-slate-400">{{ channel.description }}</p>
            </div>
            <div
              :class="[
                'w-5 h-5 rounded border-2 flex items-center justify-center',
                selectedChannel?.id === channel.id
                  ? 'border-blue-500 bg-blue-500'
                  : 'border-slate-500'
              ]"
            >
              <span v-if="selectedChannel?.id === channel.id" class="text-white text-sm">✓</span>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- Quick Setup via Channel -->
    <div v-if="selectedChannel" class="bg-slate-700/30 border border-slate-600/50 rounded-lg p-6">
      <h4 class="text-white font-bold mb-3 flex items-center gap-2">
        <span class="text-lg">⚡</span> Quick Setup
      </h4>
      <p class="text-slate-400 text-sm mb-4">
        Use this to start setup directly from {{ selectedChannel.name }}
      </p>
      
      <div v-if="!setupToken" class="space-y-3">
        <button
          @click="generateSetupToken"
          :disabled="generatingToken"
          class="w-full px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white font-medium transition-colors"
        >
          {{ generatingToken ? '🔄 Generating...' : `📱 Generate Setup Link for ${selectedChannel.name}` }}
        </button>
        <p v-if="tokenError" class="text-red-400 text-xs">{{ tokenError }}</p>
      </div>

      <div v-else class="space-y-4">
        <QRCode :url="setupUrl" />
        
        <div class="flex gap-2">
          <button
            @click="copySetupUrl"
            class="flex-1 px-4 py-2 rounded-lg border border-blue-500 text-blue-400 hover:bg-blue-900/20 transition-colors font-medium text-sm"
          >
            📋 Copy Link
          </button>
          <button
            @click="resetSetupToken"
            class="flex-1 px-4 py-2 rounded-lg border border-slate-500 text-slate-300 hover:bg-slate-600/20 transition-colors font-medium text-sm"
          >
            ↺ Generate New
          </button>
        </div>

        <div v-if="copySuccess" class="text-emerald-400 text-xs text-center">
          ✓ Link copied to clipboard!
        </div>
      </div>
    </div>

    <!-- Setup guide for selected channel -->
    <ChannelSetupGuide v-if="selectedChannel" :channel="selectedChannel.id" />

    <!-- Bot Token input -->
    <div v-if="selectedChannel">
      <label class="block text-white font-medium mb-2">
        Bot Token
        <span class="text-red-400">*</span>
      </label>
      <input
        v-model="botToken"
        type="password"
        placeholder="Paste your bot token here"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
      />
      <p class="text-xs text-slate-400 mt-2">
        Your token is encrypted and stored securely
      </p>
    </div>

    <!-- Channel ID (if applicable) -->
    <div v-if="selectedChannel?.requiresChannelId">
      <label class="block text-white font-medium mb-2">
        Channel ID / Chat ID
        <span class="text-red-400">*</span>
      </label>
      <input
        v-model="channelId"
        type="text"
        placeholder="Enter the channel or chat ID"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
      />
      <p class="text-xs text-slate-400 mt-2">
        {{ selectedChannel?.channelIdHelp || 'This ID identifies where the agent will listen' }}
      </p>
    </div>

    <!-- Webhook URL (if applicable) -->
    <div v-if="selectedChannel?.requiresWebhook">
      <label class="block text-white font-medium mb-2">
        Webhook URL
        <span class="text-red-400">*</span>
      </label>
      <input
        v-model="webhookUrl"
        type="url"
        placeholder="https://your-domain.com/webhook"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
      />
      <p class="text-xs text-slate-400 mt-2">
        The public URL where {{ selectedChannel?.name }} will send events
      </p>
    </div>

    <!-- Test connection button -->
    <div v-if="selectedChannel && botToken" class="pt-4">
      <button
        @click="testConnection"
        :disabled="isTesting"
        class="w-full px-4 py-2 rounded-lg border border-emerald-500 text-emerald-400 hover:bg-emerald-900/20 disabled:opacity-50 transition-colors"
      >
        {{ isTesting ? '🔄 Testing...' : '✓ Test Connection' }}
      </button>
      <p v-if="testResult" :class="['text-xs mt-2', testResult.success ? 'text-emerald-400' : 'text-red-400']">
        {{ testResult.message }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import ChannelSetupGuide from './ChannelSetupGuide.vue'
import QRCode from '../QRCode.vue'

const botToken = ref('')
const channelId = ref('')
const webhookUrl = ref('')
const selectedChannel = ref(null)
const isTesting = ref(false)
const testResult = ref(null)
const setupToken = ref('')
const generatingToken = ref(false)
const tokenError = ref('')
const copySuccess = ref(false)

const setupUrl = computed(() => {
  if (!setupToken.value) return ''
  const baseUrl = window.location.origin
  return `${baseUrl}/onboarding?token=${setupToken.value}`
})

const channels = [
  {
    id: 'telegram',
    name: 'Telegram',
    description: 'Private messaging & groups',
    setupUrl: 'https://t.me/botfather',
    setupInstructions: 'Get your bot token from BotFather',
    requiresChannelId: true,
    channelIdHelp: 'Your personal Telegram chat ID (starts with - for groups)',
    steps: [
      'Open Telegram and search for @BotFather',
      'Start a chat and send /start',
      'Send /newbot and follow the prompts',
      'Copy the HTTP API token provided',
      'Find your chat ID: send /getid to @userinfobot',
      'Paste both values above'
    ],
    testUrl: '/api/v1/test-channel/telegram'
  },
  {
    id: 'discord',
    name: 'Discord',
    description: 'Servers, channels & community',
    setupUrl: 'https://discord.com/developers/applications',
    setupInstructions: 'Create a bot in Discord Developer Portal',
    requiresChannelId: false,
    steps: [
      'Go to Discord Developer Portal',
      'Click "New Application"',
      'Go to "Bot" and click "Add Bot"',
      'Copy the bot token',
      'Go to "OAuth2" → "URL Generator"',
      'Select "bot" and grant needed permissions',
      'Use the generated URL to invite the bot to your server'
    ],
    testUrl: '/api/v1/test-channel/discord'
  },
  {
    id: 'slack',
    name: 'Slack',
    description: 'Team workspace collaboration',
    setupUrl: 'https://api.slack.com/apps',
    setupInstructions: 'Create an app in Slack API workspace',
    requiresWebhook: true,
    steps: [
      'Go to Slack API apps page',
      'Click "Create New App" → "From scratch"',
      'Name your app and select workspace',
      'Enable "Socket Mode" (for webhooks)',
      'Under "Event Subscriptions", enable events',
      'Subscribe to message events',
      'Generate and copy your bot token'
    ],
    testUrl: '/api/v1/test-channel/slack'
  },
  {
    id: 'whatsapp',
    name: 'WhatsApp',
    description: 'WhatsApp Business API',
    setupUrl: 'https://www.twilio.com/console/sms/whatsapp-sandbox',
    setupInstructions: 'Set up WhatsApp via Twilio',
    requiresChannelId: true,
    channelIdHelp: 'Your WhatsApp business phone number',
    steps: [
      'Sign up for Twilio',
      'Navigate to WhatsApp Sandbox',
      'Follow the sandbox setup instructions',
      'Get your Account SID and Auth Token',
      'Save your WhatsApp number'
    ],
    testUrl: '/api/v1/test-channel/whatsapp'
  }
]

const selectChannel = (channel) => {
  selectedChannel.value = channel
  botToken.value = ''
  channelId.value = ''
  webhookUrl.value = ''
  testResult.value = null
  setupToken.value = ''
  tokenError.value = ''
}

const generateSetupToken = async () => {
  try {
    generatingToken.value = true
    tokenError.value = ''

    const response = await fetch('/api/v1/setup/initialize', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        channel: selectedChannel.value.id,
        sender_id: 'setup_wizard',
        metadata: {
          source: 'onboarding_wizard'
        }
      })
    })

    if (!response.ok) {
      throw new Error('Failed to generate setup token')
    }

    const data = await response.json()
    setupToken.value = data.token
  } catch (err) {
    tokenError.value = `Error: ${err.message}`
  } finally {
    generatingToken.value = false
  }
}

const resetSetupToken = () => {
  setupToken.value = ''
  tokenError.value = ''
  copySuccess.value = false
}

const copySetupUrl = () => {
  navigator.clipboard.writeText(setupUrl.value).then(() => {
    copySuccess.value = true
    setTimeout(() => {
      copySuccess.value = false
    }, 2000)
  })
}

const testConnection = async () => {
  try {
    isTesting.value = true
    testResult.value = null

    const response = await fetch(selectedChannel.value.testUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        botToken: botToken.value,
        channelId: channelId.value,
        webhookUrl: webhookUrl.value
      })
    })

    const data = await response.json()
    testResult.value = {
      success: response.ok,
      message: data.message || (response.ok ? 'Connection successful!' : 'Connection failed')
    }
  } catch (err) {
    testResult.value = {
      success: false,
      message: `Error: ${err.message}`
    }
  } finally {
    isTesting.value = false
  }
}

defineExpose({
  config: computed(() => ({
    type: selectedChannel.value?.id || '',
    botToken: botToken.value,
    channelId: channelId.value,
    webhookUrl: webhookUrl.value
  }))
})
</script>

<style scoped>
</style>

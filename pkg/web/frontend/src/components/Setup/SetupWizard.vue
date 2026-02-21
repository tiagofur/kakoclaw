<template>
  <div class="setup-wizard min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
    <div class="max-w-2xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-4xl font-bold text-white mb-2">Welcome to KakoClaw</h1>
        <p class="text-slate-300">Let's set up your AI agent in just a few steps</p>
      </div>

      <!-- Progress indicator -->
      <div class="mb-8">
        <div class="flex items-center justify-between mb-4">
          <div
            v-for="(step, index) in steps"
            :key="index"
            class="flex items-center"
          >
            <div
              :class="[
                'w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm transition-all',
                currentStep > index
                  ? 'bg-emerald-500 text-white'
                  : currentStep === index
                  ? 'bg-blue-500 text-white ring-4 ring-blue-300'
                  : 'bg-slate-700 text-slate-400'
              ]"
            >
              {{ currentStep > index ? '✓' : index + 1 }}
            </div>
            <span
              :class="[
                'ml-2 text-sm font-medium hidden sm:inline',
                currentStep >= index ? 'text-white' : 'text-slate-400'
              ]"
            >
              {{ step.label }}
            </span>
            <div
              v-if="index < steps.length - 1"
              :class="[
                'flex-1 h-1 mx-4 rounded-full transition-all',
                currentStep > index ? 'bg-emerald-500' : 'bg-slate-700'
              ]"
            />
          </div>
        </div>
      </div>

      <!-- Step content -->
      <div class="bg-slate-800 rounded-lg shadow-xl p-8 border border-slate-700">
        <!-- Welcome step -->
        <div v-if="currentStep === 0" class="space-y-6 animate-fade-in">
          <div class="text-center">
            <div class="text-6xl mb-4">🤖</div>
            <h2 class="text-2xl font-bold text-white mb-2">Ready to get started?</h2>
            <p class="text-slate-300">
              We'll help you configure your AI provider and communication channels in just a few steps.
            </p>
          </div>

          <div class="bg-slate-700/50 rounded-lg p-4 space-y-3">
            <div class="flex items-start">
              <span class="text-emerald-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">AI Provider</p>
                <p class="text-sm text-slate-400">OpenAI, Anthropic, or other LLM</p>
              </div>
            </div>
            <div class="flex items-start">
              <span class="text-emerald-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">Communication Channel</p>
                <p class="text-sm text-slate-400">Telegram, Discord, Slack, etc.</p>
              </div>
            </div>
            <div class="flex items-start">
              <span class="text-emerald-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">Review & Deploy</p>
                <p class="text-sm text-slate-400">Preview your setup before saving</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Provider step -->
        <ProviderForm
          v-if="currentStep === 1"
          v-model="config.provider"
          @update:model-value="onProviderChange"
        />

        <!-- Channel step -->
        <ChannelForm
          v-if="currentStep === 2"
          v-model="config.channel"
          :provider="config.provider"
          @update:model-value="onChannelChange"
        />

        <!-- Preview step -->
        <ConfigPreview
          v-if="currentStep === 3"
          :config="config"
          @edit="handleEdit"
        />

        <!-- Success step -->
        <div v-if="currentStep === 4" class="text-center space-y-6 animate-fade-in">
          <div class="text-6xl mb-4">🎉</div>
          <h2 class="text-2xl font-bold text-white">All Set!</h2>
          <p class="text-slate-300">
            Your KakoClaw agent is configured and ready to use.
          </p>

          <div class="bg-emerald-900/20 border border-emerald-500/30 rounded-lg p-4 text-emerald-300 text-sm">
            <p class="font-medium mb-2">What's next?</p>
            <ul class="space-y-1 text-left">
              <li>• Open your configured channel and send a test message</li>
              <li>• Monitor agent responses in the Chat view</li>
              <li>• Customize settings in the Settings panel</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- Navigation buttons -->
      <div class="mt-8 flex justify-between">
        <button
          v-if="currentStep > 0"
          @click="previousStep"
          class="px-6 py-2 rounded-lg border border-slate-600 text-white hover:bg-slate-700 transition-colors"
        >
          ← Back
        </button>
        <div v-else />

        <button
          v-if="currentStep < steps.length - 1"
          @click="nextStep"
          :disabled="!isStepValid"
          :class="[
            'px-8 py-2 rounded-lg font-medium transition-colors',
            isStepValid
              ? 'bg-blue-600 text-white hover:bg-blue-700'
              : 'bg-slate-700 text-slate-500 cursor-not-allowed'
          ]"
        >
          Next →
        </button>
        <button
          v-else-if="currentStep === steps.length - 1"
          @click="finishSetup"
          :disabled="isSaving"
          class="px-8 py-2 rounded-lg font-medium bg-emerald-600 text-white hover:bg-emerald-700 transition-colors disabled:opacity-50"
        >
          {{ isSaving ? 'Saving...' : 'Finish Setup' }}
        </button>
        <button
          v-else
          @click="goToDashboard"
          class="px-8 py-2 rounded-lg font-medium bg-emerald-600 text-white hover:bg-emerald-700 transition-colors"
        >
          Go to Dashboard →
        </button>
      </div>

      <!-- Error message -->
      <div
        v-if="error"
        class="mt-6 p-4 rounded-lg bg-red-900/20 border border-red-500/30 text-red-300 text-sm"
      >
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/authStore'
import ProviderForm from './ProviderForm.vue'
import ChannelForm from './ChannelForm.vue'
import ConfigPreview from './ConfigPreview.vue'

const router = useRouter()
const authStore = useAuthStore()

const currentStep = ref(0)
const isSaving = ref(false)
const error = ref('')

const steps = [
  { label: 'Welcome' },
  { label: 'AI Provider' },
  { label: 'Channel' },
  { label: 'Preview' },
  { label: 'Success' }
]

const config = ref({
  provider: {
    type: '', // 'openai', 'anthropic', 'groq', etc.
    apiKey: '',
    model: ''
  },
  channel: {
    type: '', // 'telegram', 'discord', 'slack', etc.
    botToken: '',
    channelId: '',
    webhookUrl: ''
  }
})

const isStepValid = computed(() => {
  switch (currentStep.value) {
    case 0: // Welcome
      return true
    case 1: // Provider
      return config.value.provider.type && config.value.provider.apiKey
    case 2: // Channel
      return config.value.channel.type && config.value.channel.botToken
    case 3: // Preview
      return true
    default:
      return false
  }
})

const nextStep = () => {
  if (isStepValid.value && currentStep.value < steps.length - 1) {
    currentStep.value++
    error.value = ''
  }
}

const previousStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
    error.value = ''
  }
}

const onProviderChange = (provider) => {
  config.value.provider = provider
}

const onChannelChange = (channel) => {
  config.value.channel = channel
}

const handleEdit = (step) => {
  currentStep.value = step
}

const finishSetup = async () => {
  try {
    isSaving.value = true
    error.value = ''

    // Save provider config
    if (config.value.provider.type) {
      const providerResponse = await fetch('/api/v1/users/me/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          providers: {
            [config.value.provider.type]: {
              apiKey: config.value.provider.apiKey,
              model: config.value.provider.model || undefined
            }
          }
        })
      })

      if (!providerResponse.ok) {
        throw new Error('Failed to save provider configuration')
      }
    }

    // Save channel config
    if (config.value.channel.type) {
      const channelResponse = await fetch('/api/v1/users/me/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          channels: {
            [config.value.channel.type]: {
              botToken: config.value.channel.botToken,
              channelId: config.value.channel.channelId,
              webhookUrl: config.value.channel.webhookUrl
            }
          }
        })
      })

      if (!channelResponse.ok) {
        throw new Error('Failed to save channel configuration')
      }
    }

    currentStep.value = 4 // Success step
  } catch (err) {
    error.value = err.message || 'Failed to save configuration'
    console.error('Setup error:', err)
  } finally {
    isSaving.value = false
  }
}

const goToDashboard = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 0.3s ease-out;
}
</style>

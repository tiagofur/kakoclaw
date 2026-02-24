<template>
  <div class="setup-wizard min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
    <div class="max-w-2xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-4xl font-bold text-white mb-2">Welcome to makoclaw</h1>
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
                  ? 'bg-blue-500 text-white'
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
                currentStep > index ? 'bg-blue-500' : 'bg-slate-700'
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
              <span class="text-blue-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">AI Provider</p>
                <p class="text-sm text-slate-400">OpenAI, Anthropic, or other LLM</p>
              </div>
            </div>
            <div class="flex items-start">
              <span class="text-blue-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">Workspace Setup</p>
                <p class="text-sm text-slate-400">Skills and example files (optional)</p>
              </div>
            </div>
            <div class="flex items-start">
              <span class="text-blue-400 mr-3">✓</span>
              <div>
                <p class="text-white font-medium">Communication Channel</p>
                <p class="text-sm text-slate-400">Telegram, Discord, Slack, etc. (optional)</p>
              </div>
            </div>
            <div class="flex items-start">
              <span class="text-blue-400 mr-3">✓</span>
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
          ref="providerFormRef"
          v-model="config.provider"
          @update:model-value="onProviderChange"
        />

        <!-- Workspace step -->
        <WorkspaceSetupForm
          v-if="currentStep === 2"
          v-model="config.workspace"
          @update:model-value="onWorkspaceChange"
        />

        <!-- Channel step -->
        <ChannelForm
          v-if="currentStep === 3"
          v-model="config.channel"
          :provider="config.provider"
          @update:model-value="onChannelChange"
        />

        <!-- Preview step -->
        <ConfigPreview
          v-if="currentStep === 4"
          :config="config"
          @edit="handleEdit"
        />

        <!-- Success step -->
        <div v-if="currentStep === 5" class="text-center space-y-6 animate-fade-in">
          <div class="text-6xl mb-4">🎉</div>
          <h2 class="text-2xl font-bold text-white">All Set!</h2>
          <p class="text-slate-300">
            Your makoclaw agent is configured and ready to use.
          </p>

          <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4 text-blue-300 text-sm">
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

        <div class="flex gap-3">
          <!-- Skip button for optional steps (Workspace & Channel) -->
          <button
            v-if="canSkipCurrentStep"
            @click="skipStep"
            class="px-6 py-2 rounded-lg border border-slate-600 text-slate-300 hover:bg-slate-700 transition-colors"
          >
            Skip
          </button>

          <!-- Close button on Welcome step (Step 0) -->
          <button
            v-if="currentStep === 0"
            @click="closeWizard"
            class="px-6 py-2 rounded-lg border border-red-600/50 text-red-300 hover:bg-red-900/20 transition-colors"
          >
            Close
          </button>

          <!-- Configure Later button (available after provider setup until preview) -->
          <button
            v-if="currentStep > 0 && currentStep < steps.length - 1"
            @click="configureLater"
            :disabled="isSaving"
            class="px-6 py-2 rounded-lg border border-slate-500 text-slate-300 hover:bg-slate-700 transition-colors disabled:opacity-50"
          >
            {{ isSaving ? 'Saving...' : 'Configure Later' }}
          </button>

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
            class="px-8 py-2 rounded-lg font-medium bg-blue-600 text-white hover:bg-blue-700 transition-colors disabled:opacity-50"
          >
            {{ isSaving ? 'Saving...' : 'Finish Setup' }}
          </button>
        </div>
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
import { useOnboardingStore } from '../../stores/onboardingStore'
import { useConfigStore } from '../../stores/configStore'
import ProviderForm from './ProviderForm.vue'
import ChannelForm from './ChannelForm.vue'
import ConfigPreview from './ConfigPreview.vue'
import WorkspaceSetupForm from './WorkspaceSetupForm.vue'

const router = useRouter()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const configStore = useConfigStore()

const currentStep = ref(0)
const isSaving = ref(false)
const error = ref('')
const providerFormRef = ref(null)

// Store the provider config locally when it's selected
const providerConfig = ref({
  type: '',
  apiKey: '',
  model: '',
  apiBase: ''
})

const steps = [
  { label: 'Welcome' },
  { label: 'AI Provider' },
  { label: 'Workspace' },
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
  workspace: {
    skills: [],
    exampleFiles: false
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
      // Check if provider form is valid using the ref
      return providerFormRef.value?.isValid
    case 2: // Workspace (optional)
      return true
    case 3: // Channel (optional)
      return true
    case 4: // Preview
      return true
    default:
      return false
  }
})

const canSkipCurrentStep = computed(() => {
  return currentStep.value === 2 || currentStep.value === 3
})

const skipStep = () => {
  if (canSkipCurrentStep.value) {
    nextStep()
  }
}

const closeWizard = () => {
  console.log('🚪 Closing wizard without saving')
  router.push('/dashboard')
}

const nextStep = () => {
  if (isStepValid.value && currentStep.value < steps.length - 1) {
    // Capture provider config from the form before advancing
    if (currentStep.value === 1 && providerFormRef.value) {
      const configFromRef = providerFormRef.value.config
      console.log('📸 Captured provider config:', configFromRef)
      providerConfig.value = {
        type: configFromRef?.type || '',
        apiKey: configFromRef?.apiKey || '',
        model: configFromRef?.model || '',
        apiBase: configFromRef?.apiBase || ''
      }
    }
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

const onWorkspaceChange = (workspace) => {
  config.value.workspace = workspace
}

const onChannelChange = (channel) => {
  config.value.channel = channel
}

const handleEdit = (step) => {
  currentStep.value = step
}

const getFallbackModelForProvider = (providerType) => {
  const fallbackModels = {
    openrouter: 'openai/gpt-4o-mini',
    openai: 'gpt-4o-mini',
    anthropic: 'claude-3-5-haiku-20241022',
    groq: 'llama-3.1-8b-instant',
    zhipu: 'glm-4-7',
    gemini: 'gemini-2.0-flash-exp',
    nvidia: 'meta/llama-3.1-8b-instruct',
    moonshot: 'moonshot-v1-8k',
    ollama: 'llama3.1'
  }
  return fallbackModels[providerType] || 'gpt-4o-mini'
}

const resolveProviderModel = (providerCfg) => {
  const model = providerCfg?.model?.trim()
  if (model) {
    return model
  }
  return getFallbackModelForProvider(providerCfg?.type)
}

const buildChannelPayload = (channelCfg) => {
  const channelType = channelCfg?.type
  if (!channelType) {
    return null
  }

  const baseConfig = { enabled: true }
  if (channelType === 'telegram') {
    return {
      [channelType]: {
        ...baseConfig,
        token: channelCfg.botToken || '',
        allow_from: channelCfg.channelId ? [String(channelCfg.channelId)] : []
      }
    }
  }

  if (channelType === 'discord') {
    return {
      [channelType]: {
        ...baseConfig,
        token: channelCfg.botToken || ''
      }
    }
  }

  if (channelType === 'slack') {
    return {
      [channelType]: {
        ...baseConfig,
        bot_token: channelCfg.botToken || ''
      }
    }
  }

  if (channelType === 'whatsapp') {
    return {
      [channelType]: {
        ...baseConfig,
        bridge_url: channelCfg.webhookUrl || ''
      }
    }
  }

  return {
    [channelType]: {
      ...baseConfig,
      token: channelCfg.botToken || '',
      webhook_url: channelCfg.webhookUrl || ''
    }
  }
}

const finishSetup = async () => {
  try {
    isSaving.value = true
    error.value = ''

    console.log('🚀 Starting setup completion...')

    // Get the provider config (either from stored or from ref as fallback)
    let finalProviderConfig = providerConfig.value
    
    // If config is empty, try to read from ref again
    if (!finalProviderConfig.type || !finalProviderConfig.apiKey) {
      const configFromRef = providerFormRef.value?.config
      console.log('⚠️ Config was empty, re-reading from ref...', configFromRef)
      if (configFromRef && configFromRef.type && configFromRef.apiKey) {
        finalProviderConfig = {
          type: configFromRef.type || '',
          apiKey: configFromRef.apiKey || '',
          model: configFromRef.model || '',
          apiBase: configFromRef.apiBase || ''
        }
      }
    }

    console.log('📦 Final provider config:', finalProviderConfig)

    if (!finalProviderConfig.type || !finalProviderConfig.apiKey) {
      throw new Error(`Provider configuration incomplete: type="${finalProviderConfig.type}", apiKey="${finalProviderConfig.apiKey ? 'present' : 'missing'}"`)
    }

    // Save provider config using the main config update endpoint
    console.log('📝 Saving provider config:', finalProviderConfig.type)
    
    const providerResponse = await fetch('/api/v1/me/config/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authStore.token}`
      },
      body: JSON.stringify({
        agents: {
          defaults: {
            provider: finalProviderConfig.type,
            model: resolveProviderModel(finalProviderConfig)
          }
        },
        providers: {
          [finalProviderConfig.type]: {
            api_key: finalProviderConfig.apiKey,
            api_base: finalProviderConfig.apiBase || '',
            models: resolveProviderModel(finalProviderConfig)
              ? [resolveProviderModel(finalProviderConfig)]
              : []
          }
        }
      })
    })

    if (!providerResponse.ok) {
      const errorData = await providerResponse.text()
      console.error('❌ Provider config response:', providerResponse.status, errorData)
      throw new Error(`Failed to save provider configuration (${providerResponse.status})`)
    }
    
    console.log('✅ Provider config saved')

    // Save workspace config (install skills)
    if (config.value.workspace.skills && config.value.workspace.skills.length > 0) {
      console.log('📝 Initializing workspace with skills:', config.value.workspace.skills)
      
      const workspaceResponse = await fetch('/api/v1/me/workspace/init', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          skills: config.value.workspace.skills,
          exampleFiles: config.value.workspace.exampleFiles || false
        })
      })

      if (!workspaceResponse.ok) {
        const errorData = await workspaceResponse.text()
        console.error('❌ Workspace response:', workspaceResponse.status, errorData)
        throw new Error(`Failed to initialize workspace (${workspaceResponse.status})`)
      }
      
      console.log('✅ Workspace initialized')
    }

    // Save channel config
    if (config.value.channel.type) {
      console.log('📝 Saving channel config:', config.value.channel.type)
      const channelPayload = buildChannelPayload(config.value.channel)
      
      const channelResponse = await fetch('/api/v1/me/config/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          channels: channelPayload
        })
      })

      if (!channelResponse.ok) {
        const errorData = await channelResponse.text()
        console.error('❌ Channel response:', channelResponse.status, errorData)
        throw new Error(`Failed to save channel configuration (${channelResponse.status})`)
      }
      
      console.log('✅ Channel config saved')
    }

    // Mark onboarding as complete
    console.log('📝 Marking onboarding as complete...')
    const onboardingCompleted = await onboardingStore.completeOnboarding()
    
    if (!onboardingCompleted) {
      throw new Error('Failed to mark onboarding as complete')
    }
    
    console.log('✅ Onboarding marked as complete')

    // Refresh config status so degraded banner updates immediately
    await configStore.checkStatus()

    currentStep.value = 5 // Success step
    console.log('🎉 Setup completed successfully!')
    
    // Redirect to dashboard after 2 seconds
    setTimeout(() => {
      console.log('📍 Redirecting to dashboard...')
      router.push('/dashboard')
    }, 2000)
  } catch (err) {
    error.value = err.message || 'Failed to save configuration'
    console.error('❌ Setup error:', err)
  } finally {
    isSaving.value = false
  }
}

const configureLater = async () => {
  isSaving.value = true
  try {
    console.log('📝 Saving current configuration and proceeding to dashboard...')
    console.log('📦 Current step:', currentStep.value)
    
    // Capture provider config if we're still on that step
    if (currentStep.value === 1 && providerFormRef.value) {
      const configFromRef = providerFormRef.value.config
      providerConfig.value = {
        type: configFromRef?.type || '',
        apiKey: configFromRef?.apiKey || '',
        model: configFromRef?.model || '',
        apiBase: configFromRef?.apiBase || ''
      }
    }

    // Save provider configuration (always required)
    if (providerConfig.value.type && providerConfig.value.apiKey) {
      console.log('📝 Saving provider config:', providerConfig.value.type)
      const providerResponse = await fetch('/api/v1/me/config/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          agents: {
            defaults: {
              provider: providerConfig.value.type,
              model: resolveProviderModel(providerConfig.value)
            }
          },
          providers: {
            [providerConfig.value.type]: {
              api_key: providerConfig.value.apiKey,
              api_base: providerConfig.value.apiBase || '',
              models: resolveProviderModel(providerConfig.value)
                ? [resolveProviderModel(providerConfig.value)]
                : []
            }
          }
        })
      })

      if (!providerResponse.ok) {
        const errorData = await providerResponse.text()
        console.error('❌ Provider response:', providerResponse.status, errorData)
        throw new Error(`Failed to save provider configuration (${providerResponse.status})`)
      }
      console.log('✅ Provider config saved')
    }

    // Save workspace config only if it has skills selected
    if (config.value.workspace.skills && config.value.workspace.skills.length > 0) {
      console.log('📝 Initializing workspace with skills:', config.value.workspace.skills)
      const workspaceResponse = await fetch('/api/v1/me/workspace/init', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          skills: config.value.workspace.skills,
          exampleFiles: config.value.workspace.exampleFiles || false
        })
      })

      if (!workspaceResponse.ok) {
        const errorData = await workspaceResponse.text()
        console.error('❌ Workspace response:', workspaceResponse.status, errorData)
        throw new Error(`Failed to initialize workspace (${workspaceResponse.status})`)
      }
      console.log('✅ Workspace initialized')
    }

    // Save channel config only if it has a type selected
    if (config.value.channel.type) {
      console.log('📝 Saving channel config:', config.value.channel.type)
      const channelPayload = buildChannelPayload(config.value.channel)
      const channelResponse = await fetch('/api/v1/me/config/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          channels: channelPayload
        })
      })

      if (!channelResponse.ok) {
        const errorData = await channelResponse.text()
        console.error('❌ Channel response:', channelResponse.status, errorData)
        throw new Error(`Failed to save channel configuration (${channelResponse.status})`)
      }
      console.log('✅ Channel config saved')
    }

    console.log('✅ All configured settings saved')
  await configStore.checkStatus()
    console.log('📍 Redirecting to dashboard...')
    
    // Brief delay then redirect
    setTimeout(() => {
      router.push('/dashboard')
    }, 800)
  } catch (err) {
    error.value = err.message || 'Failed to save configuration'
    console.error('❌ Save error:', err)
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



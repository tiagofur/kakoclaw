<template>
  <div class="provider-form space-y-6">
    <!-- Loading state -->
    <div v-if="loading" class="text-center py-8">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto"></div>
      <p class="text-slate-400 mt-4">Loading provider options...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="bg-red-900/20 border border-red-500/30 rounded-lg p-4">
      <p class="text-red-300">{{ error }}</p>
      <button @click="$router.go(0)" class="mt-3 text-sm text-blue-400 hover:text-blue-300">
        Reload Page
      </button>
    </div>

    <!-- Provider selection -->
    <div v-else>
      <h3 class="text-xl font-bold text-white mb-4">Choose Your AI Provider</h3>
      <p class="text-slate-400 mb-6">
        Select an AI provider that will power your agent's intelligence
      </p>

      <div class="grid grid-cols-1 gap-3 mb-6">
        <button
          v-for="provider in providers"
          :key="provider.id"
          @click="selectProvider(provider)"
          :class="[
            'p-4 rounded-lg border-2 transition-all text-left',
            selectedProvider?.id === provider.id
              ? 'border-blue-500 bg-blue-900/20'
              : 'border-slate-600 bg-slate-700/50 hover:border-slate-500'
          ]"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="font-bold text-white">
                <span v-if="provider.icon">{{ provider.icon }} </span>
                {{ provider.name }}
                <span v-if="provider.freeTier" class="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded ml-2">FREE</span>
              </p>
              <p class="text-sm text-slate-400">{{ provider.description }}</p>
            </div>
            <div
              :class="[
                'w-5 h-5 rounded border-2 flex items-center justify-center flex-shrink-0',
                selectedProvider?.id === provider.id
                  ? 'border-blue-500 bg-blue-500'
                  : 'border-slate-500'
              ]"
            >
              <span v-if="selectedProvider?.id === provider.id" class="text-white text-sm">✓</span>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- API Key input -->
    <div v-if="selectedProvider">
      <label class="block text-white font-medium mb-2">
        API Key
        <span class="text-red-400">*</span>
      </label>
      <input
        v-model="apiKey"
        type="password"
        placeholder="Enter your API key"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
      />
      <p class="text-xs text-slate-400 mt-2">
        Your API key is encrypted and stored securely. 
        <a
          :href="selectedProvider.docsUrl"
          target="_blank"
          rel="noopener"
          class="text-blue-400 hover:text-blue-300"
        >
          Get your API key →
        </a>
      </p>
    </div>

    <!-- Model selection (if applicable) -->
    <div v-if="selectedProvider">
      <label class="block text-white font-medium mb-2">
        Model
        <span class="text-red-400">*</span>
      </label>
      
      <!-- Predefined models dropdown (if available) -->
      <select
        v-if="selectedProvider.models && selectedProvider.models.length > 0"
        v-model="selectedModel"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white focus:outline-none focus:border-blue-500 transition-colors mb-3"
      >
        <option value="">Select a model</option>
        <option v-for="model in selectedProvider.models" :key="model" :value="model">
          {{ model }}
        </option>
        <option value="__custom__">✏️ Enter custom model...</option>
      </select>
      
      <!-- Custom model input -->
      <div v-if="!selectedProvider.models || selectedProvider.models.length === 0 || selectedModel === '__custom__'">
        <input
          v-model="customModel"
          type="text"
          placeholder="e.g., openrouter/free, gpt-4o, claude-3-5-sonnet"
          class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
        />
        <p class="text-xs text-slate-400 mt-2">
          Enter the exact model identifier from your provider's documentation
        </p>
      </div>
      
      <p v-if="selectedProvider.models && selectedProvider.models.length > 0 && selectedModel !== '__custom__'" class="text-xs text-slate-400 mt-2">
        Different models have different capabilities and costs. Select "Enter custom model..." to use a model not listed.
      </p>
    </div>

    <!-- Provider info box -->
    <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4">
      <p class="text-blue-300 text-sm">
        <span class="font-bold">💡 Tip:</span>
        {{ selectedProvider?.tip || 'Choose a provider that best fits your needs and budget.' }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import axios from 'axios'

const apiKey = ref('')
const selectedModel = ref('')
const customModel = ref('')
const selectedProvider = ref(null)
const providers = ref([])
const loading = ref(true)
const error = ref(null)

// Load providers from API on mount
onMounted(async () => {
  try {
    loading.value = true
    const response = await axios.get('/api/v1/providers/catalog')
    providers.value = response.data.providers.map(p => ({
      id: p.id,
      name: p.name,
      description: p.description,
      icon: p.icon,
      models: p.models,
      requiresAPIKey: p.requires_api_key,
      requiresAPIBase: p.requires_api_base,
      defaultAPIBase: p.default_api_base,
      docsUrl: p.docs_url,
      apiKeyUrl: p.api_key_url,
      tip: p.tip,
      freeTier: p.free_tier
    }))
  } catch (err) {
    console.error('Failed to load providers:', err)
    error.value = 'Failed to load provider options. Please refresh the page.'
  } finally {
    loading.value = false
  }
})

const selectProvider = (provider) => {
  selectedProvider.value = provider
  selectedModel.value = ''
  customModel.value = ''
  apiKey.value = ''
}

const getModel = () => {
  if (selectedModel.value === '__custom__' || !selectedProvider.value?.models || selectedProvider.value.models.length === 0) {
    return customModel.value
  } else {
    return selectedModel.value || ''
  }
}

const isValid = computed(() => {
  if (!selectedProvider.value || !apiKey.value) return false
  const model = getModel()
  return model && model.trim().length > 0
})

defineExpose({
  isValid,
  config: computed(() => {
    return {
      type: selectedProvider.value?.id || '',
      apiKey: apiKey.value,
      model: getModel(),
      apiBase: selectedProvider.value?.defaultAPIBase || ''
    }
  })
})
</script>

<style scoped>
</style>

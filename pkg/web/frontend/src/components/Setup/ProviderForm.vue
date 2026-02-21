<template>
  <div class="provider-form space-y-6">
    <div>
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
              <p class="font-bold text-white">{{ provider.name }}</p>
              <p class="text-sm text-slate-400">{{ provider.description }}</p>
            </div>
            <div
              :class="[
                'w-5 h-5 rounded border-2 flex items-center justify-center',
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
    <div v-if="selectedProvider?.models">
      <label class="block text-white font-medium mb-2">
        Model
        <span class="text-red-400">*</span>
      </label>
      <select
        v-model="selectedModel"
        class="w-full px-4 py-2 rounded-lg bg-slate-700 border border-slate-600 text-white focus:outline-none focus:border-blue-500 transition-colors"
      >
        <option value="">Select a model</option>
        <option v-for="model in selectedProvider.models" :key="model" :value="model">
          {{ model }}
        </option>
      </select>
      <p class="text-xs text-slate-400 mt-2">
        Different models have different capabilities and costs
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
import { ref, computed } from 'vue'

const apiKey = ref('')
const selectedModel = ref('')
const selectedProvider = ref(null)

const providers = [
  {
    id: 'openai',
    name: 'OpenAI',
    description: 'GPT-4, GPT-3.5 - Most capable models',
    models: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo'],
    docsUrl: 'https://platform.openai.com/api-keys',
    tip: 'OpenAI provides the most advanced models including GPT-4. Excellent for complex reasoning.'
  },
  {
    id: 'anthropic',
    name: 'Anthropic',
    description: 'Claude - Safe and reliable',
    models: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
    docsUrl: 'https://console.anthropic.com',
    tip: 'Claude models are known for safety, reliability, and great instruction-following.'
  },
  {
    id: 'groq',
    name: 'Groq',
    description: 'Fast inference - Great for real-time',
    models: ['mixtral-8x7b-32768', 'llama2-70b-4096'],
    docsUrl: 'https://console.groq.com',
    tip: 'Groq offers extremely fast inference. Perfect for low-latency applications.'
  },
  {
    id: 'together',
    name: 'Together AI',
    description: 'Multiple open-source models',
    models: ['meta-llama/Llama-2-70b', 'mistralai/Mistral-7B'],
    docsUrl: 'https://www.together.ai',
    tip: 'Access to various open-source models with flexible pricing.'
  },
  {
    id: 'huggingface',
    name: 'Hugging Face',
    description: 'Open-source models via inference API',
    models: [],
    docsUrl: 'https://huggingface.co/inference-api',
    tip: 'Great for open-source models. Good for privacy-conscious users.'
  }
]

const selectProvider = (provider) => {
  selectedProvider.value = provider
  selectedModel.value = ''
  apiKey.value = ''
}

defineExpose({
  config: computed(() => ({
    type: selectedProvider.value?.id || '',
    apiKey: apiKey.value,
    model: selectedModel.value || selectedProvider.value?.models?.[0] || ''
  }))
})
</script>

<style scoped>
</style>

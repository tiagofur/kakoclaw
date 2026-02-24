<template>
  <!--
    SetupView - Degraded Mode Quick Setup
    
    This view is shown when:
    - LLM provider is not configured (degraded mode)
    - User clicks "Configure Now" button in DegradedModeBanner
    - Routes to /setup
    
    Features:
    - Quick 2-step provider configuration
    - Step 1: Choose provider (OpenAI, Anthropic, Groq, Ollama, etc.)
    - Step 2: Configure API key and model
    - Does NOT mark onboarding complete or setup workspace
    
    Difference from OnboardingView:
    - OnboardingView is for first-time users (/onboarding) with full 6-step wizard
    - SetupView is for degraded mode quick fix when provider is missing
  -->
  <div class="setup-view">
    <div class="setup-container">
      <div class="setup-header">
        <h1>🤖 Configure LLM Provider</h1>
        <p>Choose and configure your AI model provider to get started</p>
      </div>

      <div v-if="!showSuccess" class="setup-wizard">
        <!-- Step 1: Choose Provider -->
        <div class="wizard-step" :class="{ active: step === 1 }">
          <h2>Step 1: Choose Provider</h2>
          
          <!-- Loading State -->
          <div v-if="loading" class="loading-state">
            <div class="spinner"></div>
            <p>Loading provider options...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="error-state">
            <div class="error-icon">⚠️</div>
            <p>{{ error }}</p>
            <button @click="window.location.reload()" class="btn-secondary">Reload Page</button>
          </div>

          <!-- Provider Grid -->
          <div v-else class="provider-grid">
            <div
              v-for="provider in providers"
              :key="provider.id"
              class="provider-card"
              :class="{ selected: selectedProvider === provider.id }"
              @click="selectProvider(provider.id)"
            >
              <div class="provider-icon">
                <span v-if="provider.icon">{{ provider.icon }}</span>
                {{ provider.name }}
              </div>
              <p>{{ provider.description }}</p>
              <div class="provider-badges">
                <span v-if="provider.apiKeyRequired" class="badge">API Key Required</span>
                <span v-else class="badge local">Local/Free</span>
                <span v-if="provider.freeTier" class="badge free">FREE</span>
              </div>
            </div>
          </div>
          
          <button
            @click="step = 2"
            :disabled="!selectedProvider"
            class="btn-primary"
          >
            Next →
          </button>
        </div>

        <!-- Step 2: Configure Provider -->
        <div class="wizard-step" :class="{ active: step === 2 }">
          <button @click="step = 1" class="btn-back">← Back</button>
          <h2>Step 2: Configure {{ getProviderName() }}</h2>

          <form @submit.prevent="saveConfiguration" class="config-form">
            <div class="form-group">
              <label>Model</label>
              <select v-model="config.model" @change="handleModelChange" required>
                <option value="">Select a model...</option>
                <option
                  v-for="model in getProviderModels()"
                  :key="model.id"
                  :value="model.id"
                >
                  {{ model.name }}
                </option>
                <option value="__custom__">✏️ Enter custom model...</option>
              </select>
            </div>

            <!-- Custom model input -->
            <div v-if="config.model === '__custom__'" class="form-group">
              <label>Custom Model ID</label>
              <input
                v-model="customModel"
                type="text"
                placeholder="e.g., openrouter/free, deepseek-chat, qwen/qwen-2-7b-instruct:free"
                required
              />
              <small>Enter the exact model identifier from your provider's documentation</small>
            </div>

            <div v-if="requiresApiKey()" class="form-group">
              <label>API Key</label>
              <input
                v-model="config.apiKey"
                type="password"
                placeholder="Enter your API key"
                required
              />
              <small>Get your API key from {{ getProviderSettings().apiKeyUrl }}</small>
            </div>

            <div v-if="requiresApiBase()" class="form-group">
              <label>API Base URL (Optional)</label>
              <input
                v-model="config.apiBase"
                type="url"
                :placeholder="getProviderSettings().defaultApiBase"
              />
            </div>

            <div class="form-actions">
              <button
                type="button"
                @click="validateConfiguration"
                :disabled="validating || !canValidate()"
                class="btn-secondary"
              >
                {{ validating ? 'Validating...' : 'Test Connection' }}
              </button>
              <button
                type="submit"
                :disabled="saving"
                class="btn-primary"
              >
                {{ saving ? 'Saving...' : 'Save Configuration' }}
              </button>
            </div>

            <div v-if="validationMessage" class="validation-message" :class="validationClass">
              {{ validationMessage }}
            </div>

            <div v-if="error" class="error-message">
              {{ error }}
            </div>
          </form>
        </div>
      </div>

      <!-- Success State -->
      <div v-else class="success-state">
        <div class="success-icon">✅</div>
        <h2>Configuration Saved!</h2>
        <p>Your LLM provider has been configured successfully.</p>
        <p class="restart-notice">
          ⚠️ Please restart the application to apply the changes.
        </p>
        <button @click="goToDashboard" class="btn-primary">
          Go to Dashboard
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfigStore } from '../stores/configStore'
import axios from 'axios'

const router = useRouter()
const configStore = useConfigStore()

const step = ref(1)
const selectedProvider = ref('')
const customModel = ref('')
const config = ref({
  model: '',
  apiKey: '',
  apiBase: ''
})

const validating = ref(false)
const saving = ref(false)
const validationMessage = ref('')
const validationClass = ref('')
const error = ref('')
const showSuccess = ref(false)
const providers = ref([])
const loading = ref(true)

// Load providers from API on mount
onMounted(async () => {
  try {
    loading.value = true
    const response = await axios.get('/api/v1/providers/catalog')
    providers.value = response.data.providers.map(p => ({
      id: p.id,
      name: p.name,
      icon: p.icon || '🤖',
      description: p.description,
      apiKeyRequired: p.requires_api_key,
      apiBaseRequired: p.requires_api_base,
      models: p.models.map(m => ({ id: m, name: m })),
      defaultApiBase: p.default_api_base || '',
      apiKeyUrl: p.api_key_url || p.docs_url,
      freeTier: p.free_tier
    }))
  } catch (err) {
    console.error('Failed to load providers:', err)
    error.value = 'Failed to load provider options. Please refresh the page.'
  } finally {
    loading.value = false
  }
})

function selectProvider(providerId) {
  selectedProvider.value = providerId
  customModel.value = ''
  config.value = {
    model: '',
    apiKey: '',
    apiBase: ''
  }
}

function handleModelChange() {
  if (config.value.model !== '__custom__') {
    customModel.value = ''
  }
}

function getProviderName() {
  const provider = providers.find(p => p.id === selectedProvider.value)
  return provider ? provider.name : ''
}

function getProviderModels() {
  const provider = providers.find(p => p.id === selectedProvider.value)
  return provider ? provider.models : []
}

function getProviderSettings() {
  const provider = providers.find(p => p.id === selectedProvider.value)
  return provider || {}
}

function requiresApiKey() {
  const provider = providers.find(p => p.id === selectedProvider.value)
  return provider ? provider.apiKeyRequired : false
}

function requiresApiBase() {
  return selectedProvider.value === 'ollama'
}

function canValidate() {
  if (!selectedProvider.value) return false
  
  // Check if model is selected (either from dropdown or custom)
  if (config.value.model === '__custom__') {
    if (!customModel.value) return false
  } else if (!config.value.model) {
    return false
  }
  
  if (requiresApiKey() && !config.value.apiKey) return false
  return true
}

async function validateConfiguration() {
  if (!canValidate()) return

  validating.value = true
  validationMessage.value = ''
  error.value = ''

  try {
    const result = await configStore.validateProvider({
      provider: selectedProvider.value,
      api_key: config.value.apiKey,
      api_base: config.value.apiBase || getProviderSettings().defaultApiBase
    })

    if (result.valid) {
      validationMessage.value = '✅ Connection successful!'
      validationClass.value = 'success'
    } else {
      validationMessage.value = '❌ ' + (result.error || 'Validation failed')
      validationClass.value = 'error'
    }
  } catch (err) {
    validationMessage.value = '❌ Validation failed: ' + err.message
    validationClass.value = 'error'
  } finally {
    validating.value = false
  }
}

async function saveConfiguration() {
  if (!canValidate()) return

  saving.value = true
  error.value = ''

  try {
    // Determine final model value
    let finalModel = config.value.model
    if (config.value.model === '__custom__') {
      finalModel = customModel.value
    }

    const payload = {
      provider: selectedProvider.value,
      model: finalModel,
      api_key: config.value.apiKey
    }

    if (config.value.apiBase) {
      payload.api_base = config.value.apiBase
    } else if (getProviderSettings().defaultApiBase) {
      payload.api_base = getProviderSettings().defaultApiBase
    }

    await configStore.updateProvider(payload)
    showSuccess.value = true
  } catch (err) {
    error.value = 'Failed to save configuration: ' + err.message
  } finally {
    saving.value = false
  }
}

function goToDashboard() {
  router.push('/')
}
</script>

<style scoped>
.setup-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.setup-container {
  max-width: 900px;
  width: 100%;
  background: white;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.setup-header {
  text-align: center;
  margin-bottom: 40px;
}

.setup-header h1 {
  font-size: 32px;
  margin: 0 0 12px 0;
  color: #333;
}

.setup-header p {
  font-size: 16px;
  color: #666;
  margin: 0;
}

.wizard-step {
  display: none;
}

.wizard-step.active {
  display: block;
}

.wizard-step h2 {
  font-size: 24px;
  margin: 0 0 24px 0;
  color: #333;
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.provider-card {
  padding: 20px;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.provider-card:hover {
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.provider-card.selected {
  border-color: #667eea;
  background: #f0f4ff;
}

.provider-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.provider-card h3 {
  font-size: 18px;
  margin: 0 0 8px 0;
  color: #333;
}

.provider-card p {
  font-size: 14px;
  color: #666;
  margin: 0 0 12px 0;
}

.badge {
  display: inline-block;
  padding: 4px 12px;
  background: #ffc107;
  color: #000;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  margin: 2px;
}

.badge.local {
  background: #4caf50;
  color: white;
}

.badge.free {
  background: #00bcd4;
  color: white;
}

.provider-badges {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 4px;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e0e0e0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-state p,
.error-state p {
  font-size: 16px;
  color: #666;
  margin: 0 0 16px 0;
}

.error-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.config-form {
  margin-top: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

.form-group small {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #666;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.btn-primary,
.btn-secondary,
.btn-back {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f5f5f5;
  color: #333;
}

.btn-secondary:hover:not(:disabled) {
  background: #e0e0e0;
}

.btn-back {
  background: none;
  color: #667eea;
  padding: 8px;
  margin-bottom: 16px;
}

.btn-back:hover {
  background: #f0f4ff;
}

.validation-message {
  margin-top: 16px;
  padding: 12px;
  border-radius: 8px;
  font-weight: 500;
}

.validation-message.success {
  background: #d4edda;
  color: #155724;
}

.validation-message.error {
  background: #f8d7da;
  color: #721c24;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #f8d7da;
  color: #721c24;
  border-radius: 8px;
}

.success-state {
  text-align: center;
  padding: 40px 20px;
}

.success-icon {
  font-size: 72px;
  margin-bottom: 24px;
}

.success-state h2 {
  font-size: 28px;
  margin: 0 0 16px 0;
  color: #333;
}

.success-state p {
  font-size: 16px;
  color: #666;
  margin: 0 0 12px 0;
}

.restart-notice {
  background: #fff3cd;
  color: #856404;
  padding: 12px;
  border-radius: 8px;
  margin: 24px 0;
}

@media (max-width: 768px) {
  .setup-container {
    padding: 24px;
  }

  .provider-grid {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }
}
</style>

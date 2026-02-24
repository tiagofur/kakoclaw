<template>
  <div v-if="configStore.needsConfiguration() && !dismissed" class="degraded-banner">
    <div class="banner-content">
      <div class="banner-icon">⚠️</div>
      <div class="banner-text">
        <h3>LLM Provider Not Configured</h3>
        <p>{{ configStore.reason || 'Configure your LLM provider to enable AI features.' }}</p>
      </div>
      <button @click="showSetupWizard" class="setup-button">
        Configure Now
      </button>
      <button @click="dismiss" class="dismiss-button" v-if="dismissible">
        ×
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useConfigStore } from '../stores/configStore'
import { useRouter } from 'vue-router'

const props = defineProps({
  dismissible: {
    type: Boolean,
    default: false
  }
})

const configStore = useConfigStore()
const router = useRouter()
const dismissed = ref(false)

function showSetupWizard() {
  router.push('/setup')
}

function dismiss() {
  dismissed.value = true
}
</script>

<style scoped>
.degraded-banner {
  background: linear-gradient(135deg, #fff3cd 0%, #ffeaa7 100%);
  border-bottom: 2px solid #ffc107;
  padding: 16px 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 1000;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 16px;
  max-width: 1200px;
  margin: 0 auto;
}

.banner-icon {
  font-size: 32px;
  flex-shrink: 0;
}

.banner-text {
  flex: 1;
}

.banner-text h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #856404;
}

.banner-text p {
  margin: 0;
  font-size: 14px;
  color: #856404;
  opacity: 0.9;
}

.setup-button {
  padding: 10px 24px;
  background: #ffc107;
  color: #000;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.setup-button:hover {
  background: #ffb300;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.setup-button:active {
  transform: translateY(0);
}

.dismiss-button {
  background: none;
  border: none;
  font-size: 24px;
  color: #856404;
  cursor: pointer;
  padding: 0 8px;
  line-height: 1;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.dismiss-button:hover {
  opacity: 1;
}

@media (max-width: 768px) {
  .banner-content {
    flex-wrap: wrap;
    gap: 12px;
  }

  .banner-icon {
    font-size: 24px;
  }

  .banner-text h3 {
    font-size: 16px;
  }

  .banner-text p {
    font-size: 13px;
  }

  .setup-button {
    width: 100%;
    flex-basis: 100%;
  }
}
</style>

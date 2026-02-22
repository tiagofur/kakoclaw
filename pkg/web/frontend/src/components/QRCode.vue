<template>
  <div class="qr-container">
    <div ref="qrElement" class="qr-code"></div>
    <p class="qr-subtitle">Scan with your phone to continue setup</p>
    <p class="qr-url-display">
      Or visit: <a :href="url" target="_blank">{{ urlText }}</a>
    </p>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import QRCode from 'qrcode'

const props = defineProps({
  url: {
    type: String,
    required: true,
  },
})

const qrElement = ref(null)
const urlText = ref('')

const generateQR = async () => {
  try {
    if (qrElement.value) {
      // Clear previous QR code
      qrElement.value.innerHTML = ''
      
      // Generate QR code
      await QRCode.toCanvas(qrElement.value, props.url, {
        width: 300,
        margin: 2,
        color: {
          dark: '#1e293b', // slate-900
          light: '#f8fafc', // slate-50
        },
      })

      // Extract URL display text
      try {
        const urlObj = new URL(props.url)
        urlText.value = `${urlObj.hostname}${urlObj.pathname}...`
      } catch {
        urlText.value = props.url.substring(0, 50) + '...'
      }
    }
  } catch (err) {
    console.error('QR Code generation failed:', err)
  }
}

onMounted(() => {
  generateQR()
})

// Regenerate QR code when URL changes
watch(
  () => props.url,
  () => {
    generateQR()
  }
)
</script>

<style scoped>
.qr-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.03) 0%, rgba(15, 23, 42, 0.03) 100%);
  border-radius: 0.75rem;
  border: 1px solid rgba(100, 116, 139, 0.2);
}

.qr-code {
  padding: 1rem;
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.qr-subtitle {
  margin: 0;
  font-size: 0.875rem;
  color: #64748b;
  font-weight: 500;
}

.qr-url-display {
  margin: 0;
  font-size: 0.75rem;
  color: #94a3b8;
  text-align: center;
  word-break: break-all;
}

.qr-url-display a {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.qr-url-display a:hover {
  color: #2563eb;
  text-decoration: underline;
}
</style>

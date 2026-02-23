import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/globals.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

// Register PWA service worker
const isLocalhost =
  window.location.hostname === 'localhost' ||
  window.location.hostname === '127.0.0.1'

if ('serviceWorker' in navigator && !isLocalhost) {
  import('virtual:pwa-register').then(({ registerSW }) => {
    const updateSW = registerSW({
      onNeedRefresh() {
        if (confirm('A new version of MakoClaw is available. Reload to update?')) {
          updateSW(true)
        }
      },
      onOfflineReady() {
        console.log('MakoClaw is ready to work offline')
      },
      onRegisteredSW(swUrl, registration) {
        // Check for updates every 60 minutes
        if (registration) {
          setInterval(() => {
            registration.update()
          }, 60 * 60 * 1000)
        }
      }
    })
  })
}

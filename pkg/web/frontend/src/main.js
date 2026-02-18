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
if ('serviceWorker' in navigator) {
  import('virtual:pwa-register').then(({ registerSW }) => {
    const updateSW = registerSW({
      onNeedRefresh() {
        if (confirm('A new version of PicoClaw is available. Reload to update?')) {
          updateSW(true)
        }
      },
      onOfflineReady() {
        console.log('PicoClaw is ready to work offline')
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

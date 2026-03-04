import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/globals.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

// NOTE: PWA service worker disabled — vite-plugin-pwa incompatible with Node.js v24

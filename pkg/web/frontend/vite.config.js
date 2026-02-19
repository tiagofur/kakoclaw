import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'prompt',
      includeAssets: ['favicon.svg'],
      manifest: {
        name: 'KakoClaw',
        short_name: 'KakoClaw',
        description: 'AI Agent Dashboard — Chat, Tasks, Knowledge Base, MCP & more',
        theme_color: '#6366f1',
        background_color: '#0f172a',
        display: 'standalone',
        orientation: 'any',
        scope: '/',
        start_url: '/',
        categories: ['productivity', 'utilities'],
        icons: [
          {
            src: 'favicon.svg',
            sizes: 'any',
            type: 'image/svg+xml',
            purpose: 'any'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,png,woff2}'],
        navigateFallback: 'index.html',
        navigateFallbackDenylist: [/^\/api\//, /^\/ws\//],
        runtimeCaching: [
          {
            urlPattern: /^\/api\/v1\/models$/,
            handler: 'StaleWhileRevalidate',
            options: {
              cacheName: 'api-models',
              expiration: { maxEntries: 1, maxAgeSeconds: 300 }
            }
          },
          {
            urlPattern: /^\/api\/v1\/health$/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-health',
              expiration: { maxEntries: 1, maxAgeSeconds: 60 }
            }
          }
        ]
      }
    })
  ],
  root: __dirname,
  base: '/',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    minify: 'terser',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'vue': ['vue', 'vue-router', 'pinia'],
          'vendor': ['axios']
        }
      }
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: false
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        rewriteWsOrigin: true
      }
    }
  }
})

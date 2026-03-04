import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// NOTE: vite-plugin-pwa disabled — incompatible with Node.js v24 (minipass/path-scurry issue)
// import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
  ],
  root: __dirname,
  base: '/',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    minify: 'esbuild',
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
        target: 'http://localhost:18880',
        changeOrigin: true,
        ws: false
      },
      '/ws': {
        target: 'ws://localhost:18880',
        ws: true,
        rewriteWsOrigin: true
      }
    }
  }
})

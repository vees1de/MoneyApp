import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
      },
      '/healthz': {
        target: process.env.VITE_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
      },
      '/readyz': {
        target: process.env.VITE_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
      },
      '/openapi.yaml': {
        target: process.env.VITE_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
      },
      '/swagger': {
        target: process.env.VITE_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
})

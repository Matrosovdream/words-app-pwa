import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: true,
    port: 5173,
    watch: {
      // Polling is more reliable inside Docker volumes on macOS
      usePolling: true,
      interval: 300,
    },
    proxy: {
      '/api': process.env.VITE_API_TARGET || 'http://localhost:8080',
    },
  },
})

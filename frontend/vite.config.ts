import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    // Proxy API calls to Go server during local dev (avoid CORS issues)
    proxy: {
      '/auth': 'http://localhost:8080',
      '/projects': 'http://localhost:8080',
      '/tasks': 'http://localhost:8080',
    },
  },
})

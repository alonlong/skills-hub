import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

const backendPort = process.env.BACKEND_HTTP_PORT || '3001'
const frontendPort = Number(process.env.FRONTEND_PORT) || 3002
const apiOrigin = `http://127.0.0.1:${backendPort}`

/** Same proxy for dev + preview so /api and /oauth2 always reach the Go server. */
const devProxy = {
  '/api': {
    target: apiOrigin,
    changeOrigin: true,
  },
  '/oauth2': {
    target: apiOrigin,
    changeOrigin: true,
  },
} as const

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    exclude: ['**/node_modules/**'],
  },
  server: {
    port: frontendPort,
    strictPort: true,
    watch: {
      usePolling: true,
      interval: 150,
    },
    proxy: devProxy,
  },
  preview: {
    port: frontendPort,
    strictPort: true,
    proxy: devProxy,
  },
})

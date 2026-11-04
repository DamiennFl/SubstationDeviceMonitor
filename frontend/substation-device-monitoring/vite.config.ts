import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],

  // Define forwarding proxy, so requests from frontend go to localhost (for now) via /api.
  // Frontend code can now just define /api/<something>, instead of having to define the target path.
  // Note: This is purely for development, in prod Vite builds frontend without vite.config.ts
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  }
})
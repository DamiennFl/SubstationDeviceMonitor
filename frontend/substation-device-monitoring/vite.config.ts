import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],

  // Define forwarding proxy, so requests from frontend go to localhost (for now) via /api.
  // Frontend code can now just define /api, instead of having to define the target path.
  // This means during deployment this can be updated to the api path instead of localhost.
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  }
})
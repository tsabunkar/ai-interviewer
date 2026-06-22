import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/question': 'http://localhost:8080',
      '/answer': 'http://localhost:8080',
      '/results': 'http://localhost:8080',
      '/leaderboard': 'http://localhost:8080',
      '/stats': 'http://localhost:8080',
    },
  },
})

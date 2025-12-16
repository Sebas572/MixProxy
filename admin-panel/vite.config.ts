import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import config from '../proxy.config.json'

// https://vite.dev/config/
export default defineConfig({
  server: {
    allowedHosts: [
      `${config.subdomain_admin_panel}.${config.hostname}`,
    ],
  },
  plugins: [
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
  ],
})

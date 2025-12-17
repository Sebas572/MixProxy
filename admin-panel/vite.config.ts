import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import config from '../proxy.config.json'
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(() => ({
  server: {
    host: "::",
    port: 5173,
    allowedHosts: [
      `${config.subdomain_admin_panel}.${config.hostname}`,
    ]
  },
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));

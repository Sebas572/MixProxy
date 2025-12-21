import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import config from "../.config/proxy.config.json";
import path from "path";

console.log(`${config.subdomain_admin_panel}.${config.hostname}`)

// https://vitejs.dev/config/
export default defineConfig(() => ({
  server: {
    host: "::",
    port: 5173,
    allowedHosts: [
      `${config.subdomain_admin_panel}.${config.hostname}`,
    ]
  },
  preview: {
    host: "::",
    port: 4173,
    allowedHosts: [
      `${config.subdomain_admin_panel}.${config.hostname}`,
      config.subdomain_admin_panel,
      "localhost",
      "127.0.0.1"
    ]
  },
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));

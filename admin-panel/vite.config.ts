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
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));

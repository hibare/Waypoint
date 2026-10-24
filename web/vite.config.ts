import path from "node:path";
import fs from "node:fs";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  server: (() => {
    const keyPath = path.resolve(import.meta.dirname, "../certs/key.pem");
    const certPath = path.resolve(import.meta.dirname, "../certs/cert.pem");

    if (fs.existsSync(keyPath) && fs.existsSync(certPath)) {
      return {
        https: {
          key: fs.readFileSync(keyPath),
          cert: fs.readFileSync(certPath),
        },
        proxy: {
          "^/api/v1": {
            target: "https://localhost:5000",
            changeOrigin: true,
            secure: false,
          },
        },
      };
    } else {
      return {
        proxy: {
          "^/api/v1": {
            target: "http://localhost:5000",
            changeOrigin: true,
            secure: false,
          },
        },
      };
    }
  })(),
});

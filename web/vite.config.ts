/// <reference types="vitest/config" />
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// The Go server binds 127.0.0.1:8080 by default; the dev server proxies /api
// (including the SSE endpoints) to it so the frontend talks to one origin.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  // Build straight into the Go package that embeds it (-tags embed), so a
  // single `make web-build` feeds the embedded binary.
  build: { outDir: "../internal/web/dist", emptyOutDir: true },
  server: {
    proxy: {
      // ws: true so the exec WebSocket at /api/containers/:id/exec proxies too.
      "/api": { target: "http://127.0.0.1:8080", changeOrigin: true, ws: true },
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
  },
});

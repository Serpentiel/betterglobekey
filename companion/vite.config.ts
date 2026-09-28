import { resolve } from 'node:path'

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// Tauri serves the dev server under `tauri dev` and bundles out/ for release;
// both are wired in src-tauri/tauri.conf.json.
export default defineConfig({
  root: resolve(__dirname, 'src/renderer'),
  plugins: [react()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
  },
  build: {
    outDir: resolve(__dirname, 'out'),
    emptyOutDir: true,
    // The webview is the system WebKit: macOS 13, the oldest supported, ships
    // Safari 16. This lowers CSS nesting (16.5+) and prefixes backdrop-filter.
    target: 'safari16',
  },
})

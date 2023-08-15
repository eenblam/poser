// vite.config.ts

import { defineConfig } from 'vite'
import { resolve } from 'path'
import react from '@vitejs/plugin-react-swc'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    manifest: "manifest.json",
    rollupOptions: {
      input: {
        home: resolve(__dirname, 'home.html'),
        app: resolve(__dirname, 'app.html'),
      },
    },
  },
})

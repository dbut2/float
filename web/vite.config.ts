import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  envDir: '..',
  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        'firebase-messaging-sw': resolve(__dirname, 'src/firebase-messaging-sw.ts'),
      },
      output: {
        entryFileNames: (chunkInfo) => {
          if (chunkInfo.name === 'firebase-messaging-sw') return 'firebase-messaging-sw.js'
          return 'assets/[name]-[hash].js'
        },
      },
    },
  },
})

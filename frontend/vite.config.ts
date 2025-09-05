import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import path from 'path'

export default defineConfig(({ mode }) => ({
  plugins: [
    svelte(),
    nodePolyfills({
      // Включаем полифилы для Buffer и других Node.js API
      include: ['buffer', 'crypto'],
      // Можно также добавить глобальный Buffer
      globals: {
        Buffer: true,
      }
    })
  ],
  base: mode === 'production' ? '/ht/' : '/ht_front_dev/',
  server: {
    port: 5173,
    strictPort: true,
    host: true,
    hmr: {
      port: 5173,
      host: 'lenichev.space'
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      'buffer': 'buffer'
    }
  }
}))
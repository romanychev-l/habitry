import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import path from 'path'

export default defineConfig(({ mode }) => ({
  plugins: [svelte()],
  base: mode === 'production' ? '/ht_front_dev/' : '/ht/',
  server: {
    port: 5173,
    strictPort: true,
    hmr: false
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  }
}))
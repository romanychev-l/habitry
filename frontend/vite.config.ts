// import { defineConfig } from 'vite'
// import { svelte } from '@sveltejs/vite-plugin-svelte'

// // https://vite.dev/config/
// export default defineConfig({
//   plugins: [svelte()],
// })

import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import path from 'path'

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  base: './' // Важно для Telegram Mini App
})
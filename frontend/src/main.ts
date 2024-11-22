import './app.css'
import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'

console.log('🚀 Starting app initialization...')

const target = document.getElementById('app')
if (!target) throw new Error('Element #app not found')

// Для Svelte 5 используем mount вместо new
const app = mount(App, { target });
console.log('✅ App initialized:', app)

initTelegram()
console.log('📱 Telegram initialized')

export default app
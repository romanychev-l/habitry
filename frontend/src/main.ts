import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'
import { init as initTelegramSDK, isTMA } from '@telegram-apps/sdk'
import './i18n/i18n'

console.log('🚀 Starting app initialization...')

// Инициализируем Telegram SDK
initTelegramSDK()
console.log('📱 Telegram SDK initializedd')

if (await isTMA()) {
    console.log('It\'s Telegram Mini Apps');
} else {
    console.log('It\'s not Telegram Mini Apps');
}

const target = document.getElementById('app')
if (!target) throw new Error('Element #app not found')

const app = mount(App, { target });
console.log('✅ App initialized:', app)

initTelegram()
console.log('📱 Telegram initialized')

export default app
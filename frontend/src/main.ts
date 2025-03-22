import './app.css'
import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'
import { init as initTelegramSDK } from '@telegram-apps/sdk'
import './i18n/i18n'
import { showTelegramOrCustomAlert } from './stores/alert';

console.log('🚀 Starting app initialization...')

// Инициализируем Telegram SDK
initTelegramSDK()
console.log('📱 Telegram SDK initializedd')

const target = document.getElementById('app')
if (!target) throw new Error('Element #app not found')

// Переопределяем стандартную функцию alert
const originalAlert = window.alert;
window.alert = function(message) {
  showTelegramOrCustomAlert('Уведомление', message);
};

const app = mount(App, { target });
console.log('✅ App initialized:', app)

initTelegram()
console.log('📱 Telegram initialized')

export default app
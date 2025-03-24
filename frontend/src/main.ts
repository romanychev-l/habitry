import './app.css'
import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'
import { init as initTelegramSDK } from '@telegram-apps/sdk'
import './i18n/i18n'
import { showTelegramOrCustomAlert } from './stores/alert';
import { get } from 'svelte/store';
import { _ } from 'svelte-i18n';
import { initTelegramWebAppStore } from './stores/telegram';

console.log('🚀 Starting app initialization...')

// Инициализируем стор для Telegram WebApp - делаем это один раз при старте
console.log('Initializing Telegram WebApp Store...');
const webAppInitialized = initTelegramWebAppStore();
console.log('Telegram WebApp Store initialized:', webAppInitialized);

// Инициализируем Telegram WebApp и SDK
initTelegram()
initTelegramSDK()
console.log('📱 Telegram SDK and WebApp initialization started')

const target = document.getElementById('app')
if (!target) throw new Error('Element #app not found')

// Переопределяем стандартную функцию alert
const originalAlert = window.alert;
window.alert = function(message) {
  showTelegramOrCustomAlert(get(_)('alerts.notification'), message);
};

// Монтируем приложение после инициализации Telegram
const app = mount(App, { target: target });
console.log('✅ App initialized:', app)

export default app
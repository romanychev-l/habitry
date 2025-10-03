import './app.css'
import { mount } from 'svelte';
import App from './App.svelte'
import { init } from '@telegram-apps/sdk-svelte'
import TelegramAnalytics from '@telegram-apps/analytics'
import './i18n/i18n'
import { _ } from 'svelte-i18n';

console.log('🚀 Starting app initialization...')

// Инициализируем Telegram SDK
init({
  acceptCustomStyles: true
});

// Инициализируем Telegram Analytics
const analyticsToken = import.meta.env.VITE_ANALYTICS_TOKEN;
if (analyticsToken) {
  TelegramAnalytics.init({
    token: analyticsToken, 
    appName: 'habitry', 
  });
  console.log('📊 Telegram Analytics initialized');
} else {
  console.warn('⚠️ Analytics token not found, skipping initialization');
}

// Монтируем приложение
mount(App, {
  target: document.getElementById('app') as HTMLElement,
});

console.log('✅ App mounted successfully')
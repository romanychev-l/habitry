import './app.css'
import { mount } from 'svelte';
import App from './App.svelte'
import { init } from '@telegram-apps/sdk-svelte'
import './i18n/i18n'
import { _ } from 'svelte-i18n';

console.log('🚀 Starting app initialization...')

// Инициализируем Telegram SDK
init({
  acceptCustomStyles: true
});

// Монтируем приложение
// Аналитика будет инициализирована в App.svelte после инициализации темы
mount(App, {
  target: document.getElementById('app') as HTMLElement,
});

console.log('✅ App mounted successfully')
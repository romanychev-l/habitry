import './app.css'
import { Buffer } from 'buffer'
// Глобальный полифил Buffer для использования в TON SDK
if (typeof window !== 'undefined') {
  window.Buffer = Buffer
}

import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'
import { init as initTelegramSDK } from '@telegram-apps/sdk'
import { isTMA, mockTelegramEnv, parseInitData } from '@telegram-apps/sdk'
import './i18n/i18n'

console.log('🚀 Starting app initialization...')

async function initializeApp() {
  try {
    // Инициализируем Telegram SDK
    initTelegramSDK()
    console.log('📱 Telegram SDK initializedd')
    
    if (await isTMA()) {
      console.log('It\'s Telegram Mini Apps');
    } else {
      console.log("Not in Telegram Mini Apps environment");
      const initDataRaw = new URLSearchParams([
        ['user', JSON.stringify({
          id: 99281932,
          first_name: 'Andrew',
          last_name: 'Rogue',
          username: 'rogue',
          language_code: 'en',
          is_premium: true,
          allows_write_to_pm: true,
        })],
        ['hash', '89d6079ad6762351f38c6dbbc41bb53048019256a9443988af7a48bcad16ba31'],
        ['auth_date', '1716922846'],
        ['signature', 'abc'],
        ['start_param', 'debug'],
        ['chat_type', 'sender'],
        ['chat_instance', '8428209589180549439'],
      ]).toString();
      
      mockTelegramEnv({
        themeParams: {
          accentTextColor: '#6ab2f2',
          bgColor: '#17212b',
          buttonColor: '#5288c1',
          buttonTextColor: '#ffffff',
          destructiveTextColor: '#ec3942',
          headerBgColor: '#17212b',
          hintColor: '#708499',
          linkColor: '#6ab3f3',
          secondaryBgColor: '#232e3c',
          sectionBgColor: '#17212b',
          sectionHeaderTextColor: '#6ab3f3',
          subtitleTextColor: '#708499',
          textColor: '#f5f5f5',
        },
        initData: parseInitData(initDataRaw),
        initDataRaw,
        version: '7.2',
        platform: 'tdesktop',
      });
    }
    
    // Монтируем приложение
    const app = new App({
      target: document.getElementById('app') as HTMLElement,
    })
    
    console.log('✅ App mounted successfully')
  } catch (error) {
    console.error('❌ Error during app initialization:', error)
  }
}

// Запускаем инициализацию
initializeApp()

export default App
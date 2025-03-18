import { mount } from 'svelte';
import App from './App.svelte'
import { initTelegram } from './utils/telegram'
import { init as initTelegramSDK } from '@telegram-apps/sdk'
import { isTMA, mockTelegramEnv, parseInitData } from '@telegram-apps/sdk'
import './i18n/i18n'

console.log('🚀 Starting app initialization...')

// Инициализируем Telegram SDK
initTelegramSDK()
console.log('📱 Telegram SDK initializedd')

if (await isTMA()) {
    console.log('It\'s Telegram Mini Apps');
} else {
    
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

const target = document.getElementById('app')
if (!target) throw new Error('Element #app not found')

const app = mount(App, { target });
console.log('✅ App initialized:', app)

initTelegram()
console.log('📱 Telegram initialized')

export default app
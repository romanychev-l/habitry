import { writable } from 'svelte/store';

// Стор для хранения данных WebApp
export const telegramWebApp = writable<{
  initData: string | null;
  initDataUnsafe: any | null;
  colorScheme: string | null;
  ready: boolean;
}>({
  initData: null,
  initDataUnsafe: null,
  colorScheme: null,
  ready: false
});

// Функция для обновления данных WebApp
export function updateTelegramWebApp() {
  const webapp = window.Telegram?.WebApp;
  
  if (!webapp) {
    return;
  }
  
  // Проверяем текущие данные в сторе
  telegramWebApp.update(data => {
    // Обновляем только если данные еще не установлены
    if (data.ready && data.initData && data.initDataUnsafe) {
      return data; // Данные уже есть, ничего не меняем
    }
    
    return {
      initData: webapp.initData || null,
      initDataUnsafe: webapp.initDataUnsafe || null,
      colorScheme: webapp.colorScheme || null,
      ready: true
    };
  });
}

// Проверяем состояние WebApp каждые 500мс (максимум 10 попыток)
export function initTelegramWebAppStore() {
  let attempts = 0;
  const maxAttempts = 10;
  
  const checkWebApp = () => {
    if (window.Telegram?.WebApp) {
      updateTelegramWebApp();
      return true;
    }
    
    attempts++;
    if (attempts < maxAttempts) {
      setTimeout(checkWebApp, 500);
    } else {
      console.error('Не удалось инициализировать Telegram WebApp после', maxAttempts, 'попыток');
    }
    return false;
  };
  
  return checkWebApp();
} 
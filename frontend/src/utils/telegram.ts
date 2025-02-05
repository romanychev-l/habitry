import { user } from '../stores/user';
import { api } from './api';

export const initTelegram = () => {
    // Проверяем, что window.Telegram.WebApp существует
    if (!window.Telegram?.WebApp) {
      console.error('Telegram WebApp is not initialized');
      return;
    }
  
    const webapp = window.Telegram.WebApp;
  
    // Получаем данные пользователя
    const userData = webapp.initDataUnsafe?.user;
    if (userData) {
      // Исправляем URL фотографии, заменяя экранированные слэши на обычные
      const photoUrl = userData.photo_url?.replace(/\\\//g, '/');
      
      user.set({
        id: userData.id,
        firstName: userData.first_name,
        username: userData.username,
        languageCode: userData.language_code,
        photoUrl: photoUrl
      });
    }
    console.log('user', userData);
  
    // Сообщаем что приложение готово
    webapp.ready();
  
    // Расширяем на весь экран
    webapp.expand();
    webapp.disableVerticalSwipes();
    // webapp.requestFullscreen();
  
    // Устанавливаем тему
    document.documentElement.classList.add(webapp.colorScheme);
  
    // Устанавливаем цвета из темы Telegram
    const root = document.documentElement;
    root.style.setProperty('--tg-theme-bg-color', webapp.backgroundColor);
    root.style.setProperty('--tg-theme-text-color', webapp.textColor);
    root.style.setProperty('--tg-theme-button-color', webapp.buttonColor);
    root.style.setProperty('--tg-theme-button-text-color', webapp.buttonTextColor);
    root.style.setProperty('--tg-theme-secondary-bg-color', webapp.secondaryBackgroundColor);
};

export async function openTelegramInvoice(amount: number) {
    if (!window.Telegram?.WebApp) {
        console.error('Telegram WebApp is not available');
        return;
    }

    try {
        const data = await api.createInvoice(amount);
        console.log('Invoice data:', data);
        
        if (!data.url) {
            throw new Error('No invoice URL in response');
        }
        
        window.Telegram.WebApp.openInvoice(data.url, (status: string) => {
            if (status === 'paid') {
                console.log('Оплата прошла успешно');
            } else if (status === 'failed') {
                console.log('Ошибка оплаты');
            } else if (status === 'cancelled') {
                console.log('Оплата отменена');
            }
        });
    } catch (error) {
        console.error('Ошибка при создании инвойса:', error);
    }
}
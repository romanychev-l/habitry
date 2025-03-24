import { user } from '../stores/user';
import { api } from './api';
import { invoice } from '@telegram-apps/sdk';
import { updateTelegramWebApp } from '../stores/telegram';

export const initTelegram = () => {
    // Проверяем, что window.Telegram.WebApp существует
    if (!window.Telegram?.WebApp) {
      console.error('Telegram WebApp is not initialized');
      return;
    }
  
    const webapp = window.Telegram.WebApp;
    
    // Обновляем стор с данными WebApp
    updateTelegramWebApp();
  
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
  
    console.log(webapp);
};

export async function openTelegramInvoice(starsAmount: number) {
    try {
        console.log('Creating invoice for', starsAmount, 'Stars');
        const data = await api.createInvoice(starsAmount);
        console.log('Invoice data:', data);
        
        if (!data.url) {
            throw new Error('No invoice URL in response');
        }

        if (!invoice.isSupported()) {
            console.error('Invoices are not supported in this version of Telegram');
            return;
        }

        const status = await invoice.open(data.url, 'url');
        console.log('Payment status:', status);
        
        switch (status) {
            case 'paid':
                console.log('Оплата прошла успешно');
                window.location.reload();
                break;
            case 'failed':
                console.log('Ошибка оплаты');
                break;
            case 'cancelled':
                console.log('Оплата отменена');
                window.location.reload();
                break;
            case 'pending':
                console.log('Оплата в процессе');
                break;
            default:
                console.log('Неизвестный статус оплаты:', status);
        }
    } catch (error) {
        console.error('Ошибка при создании инвойса:', error);
    }
}
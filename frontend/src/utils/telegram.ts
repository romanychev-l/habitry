import { user } from '../stores/user';

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
      user.set({
        firstName: userData.first_name,
        lastName: userData.last_name,
        username: userData.username
      });
    }
  
    // Сообщаем что приложение готово
    webapp.ready();
  
    // Расширяем на весь экран
    webapp.expand();
  
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
        const response = await fetch(`https://c1aec5b435247ce3066498060ecc3ada.serveo.net/create-invoice?amount=${amount}`);
        console.log(response);
        const data = await response.json();
        console.log(data);
        
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
import { addMessages, init, getLocaleFromNavigator } from 'svelte-i18n';

import ru from './locales/ru.json';
import en from './locales/en.json';

addMessages('ru', ru);
addMessages('en', en);

// Получаем язык из Telegram WebApp или используем язык браузера
const telegramLanguage = window.Telegram?.WebApp?.initDataUnsafe?.user?.language_code;
const initialLocale = telegramLanguage || getLocaleFromNavigator();

init({
  fallbackLocale: 'ru',
  initialLocale: initialLocale,
}); 
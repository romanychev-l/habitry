import { addMessages, init, getLocaleFromNavigator } from 'svelte-i18n';

import ru from './locales/ru.json';
import en from './locales/en.json';

addMessages('ru', ru);
addMessages('en', en);

init({
  fallbackLocale: 'ru',
  initialLocale: getLocaleFromNavigator(),
}); 
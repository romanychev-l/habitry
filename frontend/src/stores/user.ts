import { writable } from 'svelte/store';
import type { User } from '../types';

// Тип для данных пользователя, получаемых от Telegram WebApp
interface TelegramUser {
  id: number;
  firstName?: string;
  username?: string;
  languageCode?: string;
  photoUrl?: string;
}

// Store для данных пользователя из Telegram
export const user = writable<TelegramUser | null>(null);
// Store для баланса пользователя из нашего бэкенда
export const balance = writable<number>(0);
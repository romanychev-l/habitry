import { writable } from 'svelte/store';
import type { User } from '../types';

// Тип для данных пользователя, получаемых от Telegram WebApp
interface TelegramUser {
  id: number;
  first_name: string;
  username?: string;
  language_code?: string;
  photo_url?: string;
}

// Store для данных пользователя из Telegram
export const user = writable<TelegramUser | null>(null);
// Store для баланса пользователя из нашего бэкенда
export const balance = writable<number>(0);
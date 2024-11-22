import { writable } from 'svelte/store';

interface TelegramUser {
  firstName: string;
  lastName?: string;
  username?: string;
}

export const user = writable<TelegramUser | null>(null);
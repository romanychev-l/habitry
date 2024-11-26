import { writable } from 'svelte/store';

interface TelegramUser {
  id: number;
  firstName: string;
  username?: string;
}

export const user = writable<TelegramUser | null>(null);
import { writable } from 'svelte/store';
import type { Habit, ViewMode } from '../types';

export const habits = writable<Habit[]>([]);
export const viewMode = writable<ViewMode>('card');

export const addHabit = (habit: Omit<Habit, '_id' | 'created_at' | 'last_click_date' | 'streak' | 'score'>) => {
  habits.update(items => [...items, {
    ...habit,
    _id: crypto.randomUUID(),
    created_at: new Date().toISOString(),
    last_click_date: null,
    streak: 0,
    score: 0
  }]);
};

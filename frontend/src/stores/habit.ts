import { writable } from 'svelte/store';
import type { HabitWithStats, ViewMode } from '../types';

export const habits = writable<HabitWithStats[]>([]);
export const viewMode = writable<ViewMode>('card');

export const addHabit = (habit: Omit<HabitWithStats['habit'], '_id' | 'created_at'>) => {
  habits.update(items => [...items, {
    habit: {
      ...habit,
      _id: crypto.randomUUID(),
      created_at: new Date().toISOString()
    },
    last_click_date: null,
    streak: 0,
    score: 0
  }]);
};

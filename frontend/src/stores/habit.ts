import { writable } from 'svelte/store';
import type { Habit, ViewMode } from '../types';

export const habits = writable<Habit[]>([]);
export const viewMode = writable<ViewMode>('card');

export const addHabit = (habit: Omit<Habit, 'id' | 'createdAt'>) => {
  habits.update(items => [...items, {
    ...habit,
    id: crypto.randomUUID(),
    createdAt: new Date()
    },
  ]);
};

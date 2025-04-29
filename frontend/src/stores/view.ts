import { writable } from 'svelte/store';

export const isListView = writable(false);
export const displayScore = writable(false);

// Инициализируем значения из localStorage при загрузке
if (typeof localStorage !== 'undefined') {
  isListView.set(localStorage.getItem('isListView') === 'true');
  displayScore.set(localStorage.getItem('displayScore') === 'true');
}

// Подписываемся на изменения и сохраняем в localStorage
isListView.subscribe(value => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('isListView', value.toString());
  }
});

displayScore.subscribe(value => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('displayScore', value.toString());
  }
}); 
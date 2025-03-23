import { writable } from 'svelte/store';
import { _ } from 'svelte-i18n';
import { get } from 'svelte/store';

type AlertState = {
  visible: boolean;
  title: string;
  message: string;
  showConfirm: boolean;
  confirmText: string;
  cancelText: string;
  onConfirm?: () => void;
};

// Создаем хранилище с начальным состоянием
const initialState: AlertState = {
  visible: false,
  title: '',
  message: '',
  showConfirm: false,
  confirmText: '',
  cancelText: '',
  onConfirm: undefined
};

// Создаем хранилище Svelte
export const alertStore = writable<AlertState>(initialState);

// Функции для управления алертами
export function showAlert(title: string, message: string) {
  alertStore.set({
    ...initialState,
    visible: true,
    title,
    message
  });
}

export function showConfirm(title: string, message: string, onConfirm: () => void, confirmText = get(_)('alerts.confirm'), cancelText = get(_)('alerts.cancel')) {
  alertStore.set({
    visible: true,
    title,
    message,
    showConfirm: true,
    confirmText,
    cancelText,
    onConfirm
  });
}

export function hideAlert() {
  alertStore.update(state => ({ ...state, visible: false }));
}

// Функция для отображения алерта, всегда использует кастомный алерт
export function showTelegramOrCustomAlert(title: string, message: string) {
  // Всегда используем наш кастомный алерт
  showAlert(title, message);
}

// Функция для отображения подтверждения, всегда использует кастомный алерт
export function showTelegramOrCustomConfirm(
  title: string, 
  message: string, 
  onConfirm: () => void, 
  confirmText = get(_)('alerts.confirm'), 
  cancelText = get(_)('alerts.cancel')
) {
  // Всегда используем наш кастомный алерт с подтверждением
  showConfirm(title, message, onConfirm, confirmText, cancelText);
} 
import { initDataRaw } from '@telegram-apps/sdk-svelte';

const API_URL = import.meta.env.VITE_API_URL;

interface RequestOptions extends RequestInit {
    params?: Record<string, string>;
}

async function request(endpoint: string, options: RequestOptions = {}) {
    const { params, ...fetchOptions } = options;
    
    // Формируем URL с параметрами
    const url = new URL(`${API_URL}${endpoint}`);
    if (params) {
        Object.entries(params).forEach(([key, value]) => {
            url.searchParams.append(key, value);
        });
    }

    // Добавляем заголовки
    const headers = new Headers(fetchOptions.headers);
    
    // Добавляем timezone в заголовки
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    console.log('timezone', timezone);
    headers.set('X-Timezone', timezone);
    
    // Используем правильный формат авторизации для Telegram Mini Apps
    if (initDataRaw()) {
        headers.set('Authorization', `tma ${initDataRaw()}`);
    } else {
        console.warn('No Telegram authentication data available');
    }
    headers.set('Content-Type', 'application/json');

    console.log(`Выполняем запрос: ${url.toString()}, метод: ${fetchOptions.method || 'GET'}`);
    if (fetchOptions.body) {
        console.log('Тело запроса:', fetchOptions.body);
    }

    try {
        // Выполняем запрос
        const response = await fetch(url.toString(), {
            ...fetchOptions,
            headers
        });

        // Проверяем статус ответа
        if (!response.ok) {
            console.error(`Ошибка HTTP: ${response.status} ${response.statusText}`);
            let errorText = '';
            try {
                errorText = await response.text();
                console.error('Тело ответа с ошибкой:', errorText);
            } catch (textError) {
                console.error('Не удалось прочитать тело ответа с ошибкой:', textError);
            }
            throw new Error(`${response.status}: ${errorText}`);
        }

        // Если есть тело ответа, парсим его как JSON
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            const jsonResponse = await response.json();
            console.log('Успешный ответ (JSON):', jsonResponse);
            return jsonResponse;
        }

        const textResponse = await response.text();
        console.log('Успешный ответ (текст):', textResponse);
        return textResponse;
    } catch (error) {
        console.error('Ошибка при выполнении запроса:', error);
        if (error instanceof TypeError && error.message.includes('Network request failed')) {
            console.error('Сетевая ошибка. Проверьте соединение и доступность сервера.');
        }
        throw error;
    }
}

// Регистрация TON-депозита
async function registerTonDeposit(data: { 
    transaction_id: string; 
    amount: number; 
    will_amount: number; 
    wallet_address: string; 
    telegram_id: number;
}) {
    return request('/api/ton/deposit', { 
        method: 'POST', 
        body: JSON.stringify({
            ...data,
            currency: 'ton',
            payment_type: 'deposit'
        }) 
    });
}

// Регистрация USDT-депозита
async function registerUsdtDeposit(data: {
  transaction_id: string;
  amount: number;
  will_amount: number;
  wallet_address: string;
  usdt_master_address: string;
}) {
  console.log('registerUsdtDeposit вызван с данными:', data);
  try {
    // Проверка валидности данных перед отправкой
    if (!data.transaction_id) {
      console.error('Отсутствует transaction_id');
      throw new Error('Отсутствует transaction_id');
    }
    
    if (data.amount <= 0) {
      console.error('Некорректная сумма USDT:', data.amount);
      throw new Error('Некорректная сумма USDT');
    }
    
    if (!data.wallet_address) {
      console.error('Отсутствует адрес кошелька');
      throw new Error('Отсутствует адрес кошелька');
    }
    
    if (!data.usdt_master_address) {
      console.error('Отсутствует адрес мастер-контракта USDT');
      throw new Error('Отсутствует адрес мастер-контракта USDT');
    }
    
    console.log('Данные проверены, отправляем запрос на /api/ton/usdt-deposit');
    const result = await request('/api/ton/usdt-deposit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...data,
        currency: 'usdt',
        payment_type: 'deposit'
      }),
    });
    
    console.log('Успешно зарегистрирован USDT-депозит:', result);
    return result;
  } catch (error) {
    console.error('Ошибка при регистрации USDT-депозита:', error);
    throw error;
  }
}

// Проверка статуса USDT-транзакции
async function checkUsdtTransaction(transactionId: string) {
  return request('/api/ton/check-usdt-transaction', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ transaction_id: transactionId }),
  });
}

// Регистрация запроса на вывод токенов
async function registerWithdrawal(data: {
  transaction_id: string;
  will_amount: number;
  amount: number;
  wallet_address: string;
}) {
  console.log('registerWithdrawal вызван с данными:', data);
  try {
    // Проверка валидности данных перед отправкой
    if (!data.transaction_id) {
      console.error('Отсутствует transaction_id');
      throw new Error('Отсутствует transaction_id');
    }
    
    if (data.will_amount <= 0) {
      console.error('Некорректная сумма WILL:', data.will_amount);
      throw new Error('Некорректная сумма WILL');
    }
    
    if (!data.wallet_address) {
      console.error('Отсутствует адрес кошелька');
      throw new Error('Отсутствует адрес кошелька');
    }
    
    console.log('Данные проверены, отправляем запрос на /api/ton/withdraw');
    const result = await request('/api/ton/withdraw', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...data,
        currency: 'usdt',
        payment_type: 'withdraw'
      }),
    });
    
    console.log('Успешно зарегистрирован запрос на вывод:', result);
    return result;
  } catch (error) {
    console.error('Ошибка при регистрации запроса на вывод:', error);
    throw error;
  }
}

// API методы
export const api = {
    // Пользователь
    getUser: async (data: {
        photo_url?: string;
    }) => {
        return request('/api/user', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },
    
    updateLastVisit: async () => {
        return request('/api/user/last_visit', {
            method: 'PUT'
        });
    },
    
    getUserSettings: () =>
        request('/api/user/settings'),
    
    updateUserSettings: (data: {
        notifications_enabled: boolean;
        notification_time: string;
    }) =>
        request('/api/user/settings', { method: 'PUT', body: JSON.stringify(data) }),
    
    getUserProfile: (username: string) =>
        request('/api/user/profile', { params: { username } }),
    
    // Привычки
    createHabit: (data: any) =>
        request('/api/habit/create', { method: 'POST', body: JSON.stringify(data) }),
    
    updateHabit: (data: any) =>
        request('/api/habit/click', { method: 'PUT', body: JSON.stringify(data) }),
    
    editHabit: (data: any) =>
        request('/api/habit/edit', { method: 'PUT', body: JSON.stringify(data) }),
    
    deleteHabit: (data: any) =>
        request('/api/habit/delete', { method: 'DELETE', body: JSON.stringify(data) }),
    
    undoHabit: (data: any) =>
        request('/api/habit/undo', { method: 'PUT', body: JSON.stringify(data) }),
    
    joinHabit: (data: any) =>
        request('/api/habit/join', { method: 'POST', body: JSON.stringify(data) }),
    
    getHabitFollowers: (habitId: string) =>
        request('/api/habit/followers', { params: { habit_id: habitId } }),
    
    getHabitActivity: (habitId: string) =>
        request('/api/habit/activity', { params: { habit_id: habitId } }),
    
    unfollowHabit: (data: any) =>
        request('/api/habit/unfollow', { method: 'POST', body: JSON.stringify(data) }),

    // Архив
    archiveHabit: (data: { _id: string }) =>
        request('/api/habit/archive', { method: 'PUT', body: JSON.stringify(data) }),
    unarchiveHabit: (data: { _id: string }) =>
        request('/api/habit/unarchive', { method: 'PUT', body: JSON.stringify(data) }),
    getArchivedHabits: () =>
        request('/api/habit/archived'),

    getLeaderboard: () => 
        request('/api/leaderboard'),
    
    // TON-транзакции
    registerTonDeposit,
    
    checkTonTransaction: (transactionId: string, telegramId: number) =>
        request('/api/ton/transaction', { 
            params: { transaction_id: transactionId, telegram_id: telegramId.toString() } 
        }),
    
    // Инвойсы
    createInvoice: (amount: number) =>
        request('/api/invoice', { params: { amount: amount.toString() } }),

    // Новые методы
    registerUsdtDeposit,
    checkUsdtTransaction,
    registerWithdrawal,
    createPing: async (data: {
      follower_id: number;
      follower_username: string;
      habit_id: string;
      habit_title: string;
      sender_id: number;
      sender_username: string;
    }) => {
      try {
        return request('/api/pings/create', {
          method: 'POST',
          body: JSON.stringify(data)
        });
      } catch (error) {
        console.error('Error creating ping:', error);
        throw error;
      }
    },
    subscribeToFollowerHabit: async (data: {
      current_user_habit_id: string;
      target_user_habit_id: string;
    }) => {
      return request('/api/habit/subscribe', {
        method: 'POST',
        body: JSON.stringify(data)
      });
    },
}; 
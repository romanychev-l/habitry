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

    // Получаем данные инициализации из Telegram
    const initData = window.Telegram?.WebApp?.initData;
    console.log('Telegram initData:', initData);
    
    // Добавляем заголовки
    const headers = new Headers(fetchOptions.headers);
    if (initData) {
        // Отправляем initData как есть, без дополнительного кодирования
        headers.set('X-Telegram-Data', initData);
        console.log('Setting X-Telegram-Data header:', initData);
    } else {
        console.warn('No Telegram initData available');
    }
    headers.set('Content-Type', 'application/json');

    // Выполняем запрос
    const response = await fetch(url.toString(), {
        ...fetchOptions,
        headers
    });

    // Проверяем статус ответа
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`${response.status}: ${error}`);
    }

    // Если есть тело ответа, парсим его как JSON
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
        return response.json();
    }

    return response.text();
}

// API методы
export const api = {
    // Пользователь
    getUser: (telegramId: number, timezone: string) => 
        request('/api/user', { params: { telegram_id: telegramId.toString(), timezone } }),
    
    createUser: (userData: any) => 
        request('/api/user', { method: 'POST', body: JSON.stringify(userData) }),
    
    updateLastVisit: (data: { telegram_id: number; timezone: string }) =>
        request('/api/user/last-visit', { method: 'PUT', body: JSON.stringify(data) }),
    
    getUserSettings: (telegramId: number) =>
        request('/api/user/settings', { params: { telegram_id: telegramId.toString() } }),
    
    updateUserSettings: (data: any) =>
        request('/api/user/settings', { method: 'PUT', body: JSON.stringify(data) }),
    
    getUserProfile: (username: string) =>
        request('/api/user/profile', { params: { username } }),
    
    // Привычки
    createHabit: (data: any) =>
        request('/api/habit', { method: 'POST', body: JSON.stringify(data) }),
    
    updateHabit: (data: any) =>
        request('/api/habit/click', { method: 'PUT', body: JSON.stringify(data) }),
    
    editHabit: (data: any) =>
        request('/api/habit/edit', { method: 'PUT', body: JSON.stringify(data) }),
    
    deleteHabit: (data: { telegram_id: number; habit_id: string }) =>
        request('/api/habit/delete', { method: 'DELETE', body: JSON.stringify(data) }),
    
    undoHabit: (data: any) =>
        request('/api/habit/undo', { method: 'PUT', body: JSON.stringify(data) }),
    
    joinHabit: (data: any) =>
        request('/api/habit/join', { method: 'POST', body: JSON.stringify(data) }),
    
    getHabitFollowers: (habitId: string, telegramId: number) =>
        request('/api/habit/followers', { params: { habit_id: habitId, telegram_id: telegramId.toString() } }),
    
    getHabitProgress: (habitId: string, telegramId: number) =>
        request('/api/habit/progress', { params: { habit_id: habitId, telegram_id: telegramId.toString() } }),
    
    getHabitActivity: (habitId: string, telegramId: number) =>
        request('/api/habit/activity', { params: { habit_id: habitId, telegram_id: telegramId.toString() } }),
    
    unfollowHabit: (data: any) =>
        request('/api/habit/unfollow', { method: 'POST', body: JSON.stringify(data) }),
    
    // Инвойсы
    createInvoice: (amount: number) =>
        request('/api/invoice', { params: { amount: amount.toString() } })
}; 
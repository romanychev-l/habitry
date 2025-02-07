export interface FollowerInfo {
    _id: string;
    telegram_id: number;
    title: string;
    last_click_date: string | null;
    streak: number;
    score: number;
}

export interface Habit {
    _id: string;
    telegram_id: number;
    title: string;
    want_to_become?: string;
    days: number[];
    is_one_time: boolean;
    is_auto: boolean;
    created_at: string;
    last_click_date: string | null;
    streak: number;
    score: number;
    stake: number;
    followers: FollowerInfo[];
}

export type ViewMode = 'card' | 'list';

export interface User {
    _id: string;
    telegram_id: number;
    username: string;
    first_name: string;
    language_code: string;
    photo_url: string;
    created_at: string;
    balance: number;
    last_visit: string;
    timezone: string;
    notifications_enabled: boolean;
    notification_time: string;
    habits: Habit[];
}

export interface HabitProgress {
    total_followers: number;
    completed_today: number;
    progress: number;
}

export interface HabitRequest {
    telegram_id: number;
    habit: Habit;
}

export interface Habit {
    _id: string;
    title: string;
    want_to_become?: string;
    creator_id: number;
    created_at: string;
    days: number[];
    is_one_time: boolean;
}

export interface HabitFollower {
    _id: string;
    telegram_id: number;
    habit_id: string;
    last_click_date: string | null;
    streak: number;
    score: number;
}

export type ViewMode = 'card' | 'list';

export interface HabitWithStats {
    habit: Habit;
    last_click_date: string | null;
    streak: number;
    score: number;
}

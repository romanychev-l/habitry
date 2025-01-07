export interface Habit {
    _id: string;
    title: string;
    want_to_become?: string;
    participants: {
        telegram_id: number;
        last_click_date: string | null;
        streak: number;
        score: number;
    }[];
}

export type ViewMode = 'card' | 'list';

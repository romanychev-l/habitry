export interface Habit {
    id: string;
    title: string;
    days: number[];
    createdAt: Date;
}

export type ViewMode = 'card' | 'list';

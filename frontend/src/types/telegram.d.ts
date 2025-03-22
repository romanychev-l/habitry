interface TelegramUser {
    id: number;
    first_name: string;
    last_name?: string;
    username?: string;
    language_code?: string;
    photo_url?: string;
}

interface TelegramWebApp {
    initData: string;
    initDataUnsafe: {
        query_id?: string;
        user?: TelegramUser;
        auth_date?: number;
        hash?: string;
        start_param?: string;
    };
    colorScheme: 'light' | 'dark';
    backgroundColor: string;
    textColor: string;
    buttonColor: string;
    buttonTextColor: string;
    secondaryBackgroundColor: string;
    ready: () => void;
    expand: () => void;
    disableVerticalSwipes: () => void;
    requestFullscreen: () => void;
    close: () => void;
    openInvoice: (url: string, callback: (status: string) => void) => void;
    openLink: (url: string) => void;
    shareUrl: (url: string) => void;
    MainButton: {
      setText: (text: string) => void;
      show: () => void;
      hide: () => void;
      onClick: (callback: () => void) => void;
    };
    share: (url: string) => void;
    showAlert: (message: string) => void;
    showPopup?: (params: {
      title: string;
      message: string;
      buttons: Array<{ type: string }>;
    }) => void;
    showAlert?: (message: string, callback?: () => void) => void;
    showConfirm?: (message: string, callback: (confirmed: boolean) => void) => void;
}

interface Window {
    Telegram?: {
        WebApp: TelegramWebApp;
    };
}

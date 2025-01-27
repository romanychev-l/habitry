interface TelegramWebApp {
    ready: () => void;
    expand: () => void;
    disableVerticalSwipes: () => void;
    requestFullscreen: () => void;
    close: () => void;
    backgroundColor: string;
    textColor: string;
    buttonColor: string;
    buttonTextColor: string;
    secondaryBackgroundColor: string;
    colorScheme: 'light' | 'dark';
    initDataUnsafe: {
      start_param?: string;
      user?: {
        id: number;
        first_name: string;
        last_name?: string;
        username?: string;
        language_code?: string;
        photo_url?: string;
      };
    };
    openInvoice: (url: string, callback: (status: string) => void) => void;
  }
  
  interface Window {
    Telegram?: {
      WebApp: TelegramWebApp;
  };
}

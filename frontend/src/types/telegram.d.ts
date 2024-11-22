interface TelegramWebApp {
    ready: () => void;
    expand: () => void;
    close: () => void;
    backgroundColor: string;
    textColor: string;
    buttonColor: string;
    buttonTextColor: string;
    secondaryBackgroundColor: string;
    colorScheme: 'light' | 'dark';
  }
  
  interface Window {
    Telegram?: {
      WebApp: TelegramWebApp;
  };
}

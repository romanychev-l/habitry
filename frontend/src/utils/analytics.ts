/**
 * Google Analytics integration for Telegram Mini App
 * Ensures proper user tracking with Telegram ID
 */

declare global {
  interface Window {
    gtag?: (...args: any[]) => void;
    dataLayer?: any[];
  }
}

/**
 * Initialize Google Analytics with Telegram user ID
 * IMPORTANT: This should be called only ONCE in the entire application
 * @param gaId - Google Analytics measurement ID (e.g., 'G-K6D736VS2T')
 * @param telegramUserId - Telegram user ID from initData
 */
export function initGoogleAnalytics(gaId: string, telegramUserId: number): void {
  try {
    // 1. Инициализируем dataLayer
    window.dataLayer = window.dataLayer || [];
    window.gtag = function() {
      window.dataLayer!.push(arguments);
    };

    // 2. Динамически загружаем gtag.js скрипт
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://www.googletagmanager.com/gtag/js?id=${gaId}`;
    document.head.appendChild(script);

    // 3. Инициализируем gtag с текущей датой
    window.gtag('js', new Date());

    // 4. Вызываем config с Telegram user ID
    // Это помогает Google дедуплицировать сессии, когда Telegram очищает данные
    window.gtag('config', gaId, {
      userId: telegramUserId.toString(),
      send_page_view: true,
    });

    console.log('📊 Google Analytics initialized with Telegram user ID:', telegramUserId);
  } catch (error) {
    console.error('❌ Error initializing Google Analytics:', error);
  }
}

/**
 * Track custom event in Google Analytics
 * @param eventName - Name of the event
 * @param eventParams - Additional event parameters
 */
export function trackEvent(eventName: string, eventParams?: Record<string, any>): void {
  if (!window.gtag) {
    console.warn('⚠️ gtag is not available');
    return;
  }

  try {
    window.gtag('event', eventName, eventParams);
    console.log('📊 Event tracked:', eventName, eventParams);
  } catch (error) {
    console.error('❌ Error tracking event:', error);
  }
}


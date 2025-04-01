import { TonConnectUI, THEME, type Wallet } from '@tonconnect/ui';
import { themeParams } from '@telegram-apps/sdk-svelte';

// Манифест для подключения к TON
const manifestUrl = 'https://romanychev-l.github.io/habitry_public/manifest.json';

// Глобальный экземпляр TonConnectUI
let tonConnectInstance: TonConnectUI | null = null;
let isInitializing = false;

// Функция для получения экземпляра TonConnectUI
export function getTonConnect(): TonConnectUI {
  if (!tonConnectInstance && !isInitializing) {
    isInitializing = true;
    try {
      console.log('Создаем новый экземпляр TonConnectUI');
      
      // Проверяем наличие элемента для кнопки
      const buttonElement = document.getElementById('ton-connect');
      if (!buttonElement) {
        console.warn('Элемент с id "ton-connect" не найден, создаем временный элемент');
        const tempElement = document.createElement('div');
        tempElement.id = 'ton-connect';
        tempElement.style.display = 'none';
        document.body.appendChild(tempElement);
      }
      
      // Определяем тему на основе backgroundColor
      const isDarkTheme = themeParams.backgroundColor() === '#000000';
      
      // Создаем экземпляр с базовыми настройками
      tonConnectInstance = new TonConnectUI({
        manifestUrl: manifestUrl,
        buttonRootId: 'ton-connect',
        uiPreferences: {
          theme: isDarkTheme ? THEME.DARK : THEME.LIGHT
        }
      });
      
      // Добавляем обработчик ошибок
      tonConnectInstance.onStatusChange((wallet: Wallet | null) => {
        if (!wallet) {
          console.log('Кошелек отключен или произошла ошибка подключения');
          isInitializing = false;
        }
      });
      
      console.log('TonConnectUI instance created successfully');
    } catch (error) {
      console.error('Error creating TonConnectUI instance:', error);
      isInitializing = false;
      throw error;
    }
    isInitializing = false;
  }
  return tonConnectInstance!;
}

// Функция для проверки, был ли уже создан экземпляр
export function isTonConnectInitialized(): boolean {
  return !!tonConnectInstance;
}

// Функция для подписки на изменения состояния кошелька
export function subscribeToWalletChanges(callback: (wallet: Wallet | null) => void): () => void {
  try {
    const tonConnect = getTonConnect();
    console.log('Подписываемся на изменения кошелька');
    const unsubscribe = tonConnect.onStatusChange(callback);
    
    // Проверяем начальное состояние
    const wallet = tonConnect.wallet;
    if (wallet) {
      console.log('Кошелек уже подключен:', wallet.account.address);
      callback(wallet);
    } else {
      console.log('Кошелек не подключен');
    }
    
    return unsubscribe;
  } catch (error) {
    console.error('Ошибка при подписке на изменения кошелька:', error);
    return () => {}; // Возвращаем пустую функцию в случае ошибки
  }
}

// Функция для отключения кошелька
export function disconnectWallet(): void {
  if (tonConnectInstance?.wallet) {
    console.log('Отключаем кошелек');
    tonConnectInstance.disconnect();
  }
}

// Функция для безопасной отправки транзакции с дополнительной обработкой ошибок
export async function sendTonTransaction(transaction: any): Promise<any> {
  try {
    console.log('Подготовка к отправке транзакции:', JSON.stringify(transaction));
    
    // Проверяем корректность транзакции
    if (!transaction || !transaction.messages || !transaction.messages.length) {
      throw new Error('Некорректный формат транзакции');
    }
    
    // Проверяем, что сумма транзакции является строкой и содержит только цифры
    const amount = transaction.messages[0].amount;
    if (typeof amount !== 'string' || !/^\d+$/.test(amount)) {
      console.error('Некорректный формат суммы:', amount);
      // Исправляем формат суммы, если он некорректный
      transaction.messages[0].amount = String(Math.floor(Number(amount)));
      console.log('Исправленный формат суммы:', transaction.messages[0].amount);
    }
    
    // Проверяем формат адреса
    const address = transaction.messages[0].address;
    if (typeof address !== 'string') {
      console.error('Некорректный формат адреса:', address);
      throw new Error('Некорректный формат адреса. Адрес должен быть строкой.');
    }
    
    // Проверяем формат payload
    const payload = transaction.messages[0].payload;
    if (payload && typeof payload !== 'string') {
      console.error('Некорректный формат payload:', payload);
      throw new Error('Некорректный формат payload. Payload должен быть строкой.');
    }
    
    const tonConnect = getTonConnect();
    
    // Проверяем подключение кошелька
    if (!tonConnect.wallet) {
      throw new Error('Кошелек не подключен');
    }
    
    // Отправляем транзакцию
    console.log('Отправляем транзакцию через TonConnect');
    
    // Оборачиваем вызов в try-catch для более детальной обработки ошибок
    try {
      const result = await tonConnect.sendTransaction(transaction);
      console.log('Транзакция успешно отправлена:', result);
      console.log('boc:', result.boc);
      return result;
    } catch (sendError: any) {
      // Логируем детальную информацию об ошибке
      console.error('Ошибка при отправке транзакции через TonConnect:', sendError);
      console.error('Тип ошибки:', typeof sendError);
      console.error('Сообщение ошибки:', sendError.message);
      console.error('Стек ошибки:', sendError.stack);
      console.error('Детали транзакции:', JSON.stringify(transaction, null, 2));
      
      // Улучшенная обработка ошибок
      if (sendError.message && sendError.message.includes('NullPointerException')) {
        throw new Error('Ошибка в приложении кошелька. Пожалуйста, попробуйте использовать другой кошелек или обновите приложение.');
      }
      
      if (sendError.message && sendError.message.includes('TON_CONNECT_SDK_ERROR')) {
        throw new Error('Ошибка SDK TonConnect. Пожалуйста, проверьте формат транзакции и попробуйте снова.');
      }
      
      throw sendError;
    }
  } catch (error) {
    console.error('Ошибка при отправке транзакции:', error);
    throw error;
  }
}

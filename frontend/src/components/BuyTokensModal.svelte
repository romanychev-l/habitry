<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { getTonConnect, subscribeToWalletChanges, sendTonTransaction } from '../utils/tonConnect';
  import { api } from '../utils/api';
  import type { Wallet } from '@tonconnect/ui';
  import { beginCell, Address, toNano } from '@ton/core'
  import { TonClient, JettonMaster } from '@ton/ton';

  // Используем уже объявленные глобальные типы без declare global
  
  export let telegramId: number;
  
  const dispatch = createEventDispatcher();
  let tokensAmount = 100;
  const EXCHANGE_RATE = 10; // 1 Stars = 10 WILL
  const TON_EXCHANGE_RATE = 100; // 1 TON = 100 WILL
  const USDT_EXCHANGE_RATE = 1000; // 1 USDT = 1000 WILL (изменено)

  // Изменяем тип paymentMethod, добавляя 'usdt'
  let paymentMethod: 'stars' | 'ton' | 'usdt' = 'usdt';
  // Добавляем новую переменную для переключения между режимами покупки и вывода
  let modalMode: 'buy' | 'withdraw' = 'buy';
  // Добавим переменную для суммы вывода
  let withdrawAmount = 100;
  let walletConnected = false;
  let walletAddress = '';
  let unsubscribe: (() => void) | null = null; 
  let isProcessing = false;
  let transactionError = '';
  
  // Добавляем переменную для хранения баланса пользователя
  let userBalance = 0;
  let isLoadingBalance = false;

  onMount(async () => {
    // Подписываемся на изменения состояния кошелька
    unsubscribe = subscribeToWalletChanges((wallet: Wallet | null) => {
      if (wallet) {
        walletConnected = true;
        walletAddress = wallet.account.address;
        console.log('Кошелек подключен:', walletAddress);
      } else {
        walletConnected = false;
        walletAddress = '';
        console.log('Кошелек отключен');
      }
    });
    
    // Загружаем баланс пользователя при открытии модального окна
    await loadUserBalance();
  });

  onDestroy(() => {
    // Отписываемся от событий
    if (unsubscribe) {
      unsubscribe();
      unsubscribe = null;
    }
  });

  // Функция для загрузки баланса пользователя
  async function loadUserBalance() {
    if (!telegramId) return;
    
    try {
      isLoadingBalance = true;
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      const userData = await api.getUser(telegramId, timezone);
      userBalance = userData.balance || 0;
      console.log('Баланс пользователя загружен:', userBalance);
    } catch (error) {
      console.error('Ошибка при загрузке баланса пользователя:', error);
    } finally {
      isLoadingBalance = false;
    }
  }

  function calculateStars(tokens: number): number {
    return Math.ceil(tokens / EXCHANGE_RATE);
  }

  function calculateTon(tokens: number): number {
    return parseFloat((tokens / TON_EXCHANGE_RATE).toFixed(2));
  }

  function calculateUsdt(tokens: number): number {
    return parseFloat((tokens / USDT_EXCHANGE_RATE).toFixed(2));
  }

  async function handleBuy() {
    if (paymentMethod === 'stars') {
      dispatch('buy', {
        starsAmount: calculateStars(tokensAmount),
        paymentMethod
      });
    } else if (paymentMethod === 'ton') {
      await handleTonPayment();
    } else if (paymentMethod === 'usdt') {
      await handleUsdtPayment();
    }
  }
  
  // Новая функция для обработки вывода токенов
  async function handleWithdraw() {
    if (!walletConnected) {
      console.error('Кошелек не подключен');
      transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
      showTelegramPopup('Ошибка', 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.');
      return;
    }
    
    // Проверяем баланс
    if (withdrawAmount > userBalance) {
      console.error('Недостаточно средств для вывода');
      transactionError = `Недостаточно средств для вывода. Ваш баланс: ${userBalance} WILL`;
      showTelegramPopup('Ошибка', `Недостаточно средств для вывода. Ваш баланс: ${userBalance} WILL`);
      return;
    }
    
    try {
      isProcessing = true;
      transactionError = '';
      
      // Создаем уникальный идентификатор транзакции
      const transactionId = `W${Date.now()}${Math.random().toString(36).substring(2, 6)}`;
      
      const usdtAmount = parseFloat((withdrawAmount / USDT_EXCHANGE_RATE).toFixed(2));
      console.log(`Запрос на вывод: ${withdrawAmount} WILL = ${usdtAmount} USDT`);
      
      try {
        // Используем новый метод API для регистрации вывода
        const response = await api.registerWithdrawal({
          transaction_id: transactionId,
          will_amount: withdrawAmount,
          amount: usdtAmount,
          wallet_address: walletAddress,
          telegram_id: telegramId
        });
        
        console.log('Ответ сервера:', response);
        
        // Показываем уведомление об успешной обработке запроса
        showTelegramPopup(
          'Запрос на вывод отправлен',
          `Ваш запрос на вывод ${withdrawAmount} WILL (${usdtAmount} USDT) отправлен на обработку. Средства поступят на ваш кошелек в течение 24 часов.`
        );
        
        // Обновляем баланс пользователя
        await loadUserBalance();
        
        // Закрываем модальное окно
        dispatch('close');
      } catch (apiError: any) {
        console.error('Ошибка при регистрации запроса на вывод:', apiError);
        transactionError = apiError.message || 'Ошибка при регистрации запроса на вывод';
        showTelegramPopup('Ошибка', transactionError);
      }
    } catch (error: any) {
      console.error('Ошибка при обработке запроса на вывод:', error);
      transactionError = error.message || 'Произошла неизвестная ошибка';
      showTelegramPopup('Ошибка', transactionError);
    } finally {
      isProcessing = false;
    }
  }

  // Удобная функция для показа уведомлений
  function showTelegramPopup(title: string, message: string) {
    try {
      // @ts-ignore
      if (window.Telegram?.WebApp?.showPopup) {
        // @ts-ignore
        window.Telegram.WebApp.showPopup({
          title: title,
          message: message,
          buttons: [{ type: 'close' }]
        });
      }
    } catch (error) {
      console.warn('Не удалось показать уведомление через Telegram WebApp API:', error);
    }
  }

  // Функция для получения адреса Jetton-кошелька
  async function getJettonWalletAddress(userAddress: string, masterAddress: string) {
    try {
      const client = new TonClient({
        endpoint: 'https://toncenter.com/api/v2/jsonRPC',
      });

      const jettonMasterAddress = Address.parse(masterAddress);
      const userWalletAddress = Address.parse(userAddress);

      const jettonMaster = client.open(JettonMaster.create(jettonMasterAddress));
      const jettonWalletAddress = await jettonMaster.getWalletAddress(userWalletAddress);
      
      console.log('Получен адрес Jetton-кошелька:', jettonWalletAddress.toString());
      return jettonWalletAddress;
    } catch (error) {
      console.error('Ошибка при получении адреса Jetton-кошелька:', error);
      throw error;
    }
  }

  // Функция для обработки USDT-платежей
  async function handleUsdtPayment() {
    if (!walletConnected) {
      console.error('Кошелек не подключен');
      transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
      return;
    }

    try {
      isProcessing = true;
      transactionError = '';
      
      const usdtAmount = calculateUsdt(tokensAmount);
      
      // Адрес мастер-контракта USDT в сети TON
      const rawUsdtMasterAddress = import.meta.env.VITE_USDT_MASTER_ADDRESS;
      
      // Преобразуем строковый адрес мастер-контракта в объект Address
      let usdtMasterAddress;
      try {
        usdtMasterAddress = Address.parse(rawUsdtMasterAddress);
        console.log('Адрес мастер-контракта USDT успешно преобразован:', usdtMasterAddress.toString());
      } catch (addrError) {
        console.error('Ошибка при преобразовании адреса мастер-контракта USDT:', addrError);
        transactionError = `Недействительный адрес мастер-контракта USDT: ${rawUsdtMasterAddress}`;
        isProcessing = false;
        return;
      }
      
      // Адрес кошелька приложения
      const rawAppWalletAddress = import.meta.env.VITE_TON_WALLET_ADDRESS;
      
      // Преобразуем строковый адрес в объект Address
      let appWalletAddress;
      try {
        appWalletAddress = Address.parse(rawAppWalletAddress);
        console.log('Адрес кошелька приложения успешно преобразован:', appWalletAddress.toString());
      } catch (addrError) {
        console.error('Ошибка при преобразовании адреса приложения:', addrError);
        transactionError = `Недействительный адрес кошелька приложения: ${rawAppWalletAddress}`;
        isProcessing = false;
        return;
      }

      // Получаем адрес Jetton-кошелька для пользователя
      const jettonWalletAddress = await getJettonWalletAddress(walletAddress, rawUsdtMasterAddress);
      console.log('Адрес Jetton-кошелька пользователя:', jettonWalletAddress.toString());
      
      // Создаем уникальный идентификатор транзакции
      const transactionId = `${Date.now()}${Math.random().toString(36).substring(2, 6)}`;

      // Подробное логирование адресов
      console.log('Адрес отправителя:', walletAddress);
      console.log('Адрес отправителя 2:', Address.parse(walletAddress).toString());
      console.log('Тип адреса отправителя:', typeof walletAddress);
      console.log('Формат адреса отправителя:', walletAddress.startsWith('EQ') ? 'user-friendly (EQ)' : walletAddress.startsWith('0:') ? 'raw (0:)' : 'неизвестный');
      console.log('Адрес мастер-контракта USDT:', rawUsdtMasterAddress);
      console.log('Тип адреса мастер-контракта:', typeof rawUsdtMasterAddress);
      console.log('Формат адреса мастер-контракта:', rawUsdtMasterAddress.startsWith('EQ') ? 'user-friendly (EQ)' : rawUsdtMasterAddress.startsWith('0:') ? 'raw (0:)' : 'неизвестный');
      console.log('Адрес кошелька приложения:', rawAppWalletAddress);
      console.log('Тип адреса приложения:', typeof rawAppWalletAddress);
      console.log('Формат адреса приложения:', rawAppWalletAddress.startsWith('EQ') ? 'user-friendly (EQ)' : rawAppWalletAddress.startsWith('0:') ? 'raw (0:)' : 'неизвестный');

      // Создаем комментарий для идентификации транзакции
      const commentPayload = beginCell()
        .storeUint(0, 32) // 32 нулевых бита указывают на текстовый комментарий
        .storeStringTail(transactionId)
        .endCell();

      // Создаем сообщение для трансфера Jetton
      // console.log('TO NANO', toNano("0.05").toString())
      const msg = {
        address: jettonWalletAddress.toString(), // адрес Jetton-кошелька пользователя
        amount: toNano("0.05").toString(), // увеличиваем сумму для покрытия газа
        payload: beginCell()
          .storeUint(0xf8a7ea5, 32) // op transfer
          .storeUint(0, 64) // query_id
          .storeCoins(BigInt(Math.floor(usdtAmount * 1_000_000))) // amount
          .storeAddress(appWalletAddress) // destination
          .storeAddress(Address.parse(walletAddress)) // response destination (возвращаем на адрес отправителя)
          .storeBit(false) // custom payload
          .storeCoins(1) // forward_ton_amount = 1 nanoton для уведомления
          .storeBit(true) // forward_payload в виде reference
          .storeRef(commentPayload) // комментарий с ID транзакции
          .endCell()
          .toBoc()
          .toString("base64")
      };

      console.log('Подготовка USDT транзакции:', {
        ...msg,
        decodedPayload: {
          op: '0xf8a7ea5',
          queryId: '0',
          amount: usdtAmount,
          destination: appWalletAddress.toString(),
          responseDestination: walletAddress.toString(),
          forwardTonAmount: '0.000000001 TON',
          comment: transactionId
        }
      });
      
      // Отправляем транзакцию
      const transaction = {
        validUntil: Math.floor(Date.now() / 1000) + 360, // 6 минут на выполнение
        messages: [msg]
      };
      
      // Добавляем подробное логирование
      console.log('Отправляем USDT транзакцию. Полные данные:', JSON.stringify(transaction));
      console.log('Адрес получателя (строка):', msg.address);
      console.log('Тип адреса:', typeof msg.address);
      console.log('Сумма (строка):', msg.amount);
      console.log('Тип суммы:', typeof msg.amount);
      console.log('Payload (строка):', msg.payload);
      console.log('Тип payload:', typeof msg.payload);
      
      try {
        // Проверка подключения кошелька через TonConnect
        const tonConnect = getTonConnect();
        if (!tonConnect.wallet) {
          console.error('Кошелек не подключен');
          transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
          showTelegramPopup('Ошибка кошелька', transactionError);
          isProcessing = false;
          return;
        }

        // Проверяем, что адрес мастер-контракта USDT корректный
        if (!usdtMasterAddress.toString().startsWith('EQ')) {
          console.error('Некорректный адрес мастер-контракта USDT');
          transactionError = 'Некорректный адрес мастер-контракта USDT. Пожалуйста, проверьте конфигурацию.';
          showTelegramPopup('Ошибка конфигурации', transactionError);
          isProcessing = false;
          return;
        }

        // Проверяем, что адрес кошелька приложения корректный
        if (!appWalletAddress.toString().startsWith('EQ')) {
          console.error('Некорректный адрес кошелька приложения');
          transactionError = 'Некорректный адрес кошелька приложения. Пожалуйста, проверьте конфигурацию.';
          showTelegramPopup('Ошибка конфигурации', transactionError);
          isProcessing = false;
          return;
        }
        
        // Используем существующую функцию для отправки транзакции
        const result = await sendTonTransaction(transaction);
        console.log('USDT транзакция отправлена:', result);
        
        // Регистрируем транзакцию на бэкенде
        try {
          console.log('Начинаем регистрацию USDT транзакции на бэкенде, данные:', {
            transaction_id: transactionId,
            amount: usdtAmount,
            will_amount: tokensAmount,
            wallet_address: walletAddress,
            telegram_id: telegramId,
            usdt_master_address: rawUsdtMasterAddress
          });
          
          const responseData = await api.registerUsdtDeposit({
            transaction_id: transactionId,
            amount: usdtAmount,
            will_amount: tokensAmount,
            wallet_address: walletAddress,
            telegram_id: telegramId,
            usdt_master_address: rawUsdtMasterAddress // Используем исходный строковый адрес
          });
          
          console.log('USDT транзакция зарегистрирована:', responseData);
          
          // Запускаем проверку статуса
          startUsdtTransactionStatusCheck(transactionId);
          
          // Закрываем модальное окно и сообщаем об успешной отправке
          dispatch('usdt-transaction-sent', {
            transactionId,
            amount: usdtAmount,
            willAmount: tokensAmount
          });
          
          // Закрываем модальное окно
          dispatch('close');
        } catch (apiError: any) {
          console.error('Ошибка при регистрации USDT транзакции:', apiError);
          console.error('Тип ошибки:', typeof apiError);
          console.error('Сообщение ошибки:', apiError.message);
          console.error('Стек ошибки:', apiError.stack);
          
          // Проверяем, является ли ошибка объектом ответа
          if (apiError.response) {
            console.error('Ответ сервера:', apiError.response);
            console.error('Статус ответа:', apiError.response.status);
            console.error('Тело ответа:', apiError.response.data);
          }
          
          // Если это ошибка сети, логируем дополнительную информацию
          if (apiError.request) {
            console.error('Запрос был отправлен, но ответ не получен');
            console.error('Детали запроса:', apiError.request);
          }
          
          transactionError = apiError.message || 'Ошибка при регистрации транзакции на сервере';
          
          // Показываем уведомление об ошибке
          showTelegramPopup('Ошибка при регистрации транзакции', transactionError);
        }
      } catch (sendError: any) {
        console.error('Ошибка при отправке USDT транзакции:', sendError);
        transactionError = sendError.message || 'Произошла ошибка при отправке USDT транзакции';
        showTelegramPopup('Ошибка при отправке USDT транзакции', transactionError);
      }
    } catch (error) {
      console.error('Ошибка при обработке платежа USDT:', error);
      transactionError = error instanceof Error ? error.message : 'Неизвестная ошибка';
    } finally {
      isProcessing = false;
    }
  }

  function startUsdtTransactionStatusCheck(transactionId: string) {
    // Сохраняем ID транзакции в localStorage для последующих проверок
    localStorage.setItem('last_usdt_tx', transactionId);
    
    // Запускаем проверку статуса через 30 секунд
    setTimeout(() => {
      checkUsdtTransactionStatus(transactionId);
    }, 30000);
    
    // Отправляем уведомление пользователю
    showTelegramPopup(
      'USDT транзакция отправлена',
      'Ваша USDT транзакция отправлена в блокчейн TON. Обработка может занять несколько минут.'
    );
  }

  async function checkUsdtTransactionStatus(transactionId: string) {
    try {
      const data = await api.checkUsdtTransaction(transactionId, telegramId);
      console.log('Статус USDT транзакции:', data);
      
      if (data.tx_status === 'completed') {
        // Транзакция успешно обработана
        showTelegramPopup(
          'Транзакция подтверждена',
          `На ваш счет начислено ${data.will_amount} WILL токенов`
        );
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkUsdtTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        showTelegramPopup(
          'Транзакция не удалась',
          'Не удалось обработать вашу USDT транзакцию. Пожалуйста, попробуйте еще раз.'
        );
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      }
    } catch (error) {
      console.error('Ошибка при проверке статуса USDT транзакции:', error);
    }
  }

  async function handleTonPayment() {
    if (!walletConnected) {
      console.error('Кошелек не подключен');
      transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
      return;
    }

    try {
      isProcessing = true;
      transactionError = '';
      
      const tonAmount = calculateTon(tokensAmount);
      
      // Получаем адрес кошелька приложения из переменных окружения или конфигурации
      const appWalletAddress = import.meta.env.VITE_TON_WALLET_ADDRESS;
      
      // Создаем уникальный идентификатор транзакции (только цифры и буквы)
      const transactionId = `${Date.now()}${Math.random().toString(36).substring(2, 6)}`;

      const body = beginCell()
        .storeUint(0, 32) // write 32 zero bits to indicate that a text comment will follow
        .storeStringTail(`${transactionId}`) // write our text comment
        .endCell();
      
      // Максимально упрощенная транзакция для тестирования
      const transaction = {
        validUntil: Math.floor(Date.now() / 1000) + 360, // 5 минут на выполнение
        messages: [
          {
            address: appWalletAddress,
            amount: String(Math.floor(tonAmount * 1_000_000_000)), // Конвертируем в наноТОНы
            payload: body.toBoc().toString("base64") // Изменил payload на body - так работает в TON
          }
        ]
      };

      console.log('Подготовка транзакции:', transaction);
      
      try {
        // Проверка подключения кошелька через TonConnect
        const tonConnect = getTonConnect();
        if (!tonConnect.wallet) {
          console.error('Кошелек не подключен');
          transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
          showTelegramPopup('Ошибка кошелька', transactionError);
          isProcessing = false;
          return;
        }
        
        // Используем существующую функцию для отправки транзакции
        const result = await sendTonTransaction(transaction);
        console.log('Транзакция отправлена:', result);
        
        // Отправляем информацию о транзакции на бэкенд через api.ts
        try {
          const responseData = await api.registerTonDeposit({
            transaction_id: transactionId,
            amount: tonAmount,
            will_amount: tokensAmount,
            wallet_address: walletAddress,
            telegram_id: telegramId
          });
          
          console.log('Транзакция зарегистрирована:', responseData);
          
          // Запускаем проверку статуса транзакции
          startTransactionStatusCheck(transactionId);
          
          // Закрываем модальное окно и сообщаем об успешной отправке
          dispatch('ton-transaction-sent', {
            transactionId,
            amount: tonAmount,
            willAmount: tokensAmount
          });
          
          // Закрываем модальное окно
          dispatch('close');
        } catch (apiError: any) {
          console.error('Ошибка при регистрации транзакции:', apiError);
          transactionError = apiError.message || 'Ошибка при регистрации транзакции на сервере';
          
          // Показываем уведомление об ошибке
          showTelegramPopup('Ошибка при регистрации транзакции', transactionError);
        }
      } catch (sendError: any) {
        console.error('Ошибка при отправке транзакции:', sendError);
        
        // Проверяем, содержит ли ошибка NullPointerException
        if (sendError.message && sendError.message.includes('NullPointerException')) {
          transactionError = 'Ошибка в приложении кошелька. Пожалуйста, попробуйте использовать другой кошелек или обновите приложение.';
        } else {
          transactionError = sendError.message || 'Произошла ошибка при отправке транзакции';
        }
        
        // Показываем уведомление об ошибке
        showTelegramPopup('Ошибка при отправке транзакции', transactionError);
      }
    } catch (error) {
      console.error('Ошибка при обработке платежа TON:', error);
      transactionError = error instanceof Error ? error.message : 'Неизвестная ошибка';
    } finally {
      isProcessing = false;
    }
  }
  
  // Функция для проверки статуса транзакции
  function startTransactionStatusCheck(transactionId: string) {
    // Сохраняем ID транзакции в localStorage для последующих проверок
    localStorage.setItem('last_ton_tx', transactionId);
    
    // Запускаем проверку статуса через 30 секунд
    setTimeout(() => {
      checkTransactionStatus(transactionId);
    }, 30000);
    
    // Отправляем уведомление пользователю
    showTelegramPopup(
      'Транзакция отправлена',
      'Ваша транзакция отправлена в блокчейн TON. Обработка может занять несколько минут.'
    );
  }
  
  // Функция для проверки статуса транзакции
  async function checkTransactionStatus(transactionId: string) {
    try {
      const data = await api.checkTonTransaction(transactionId, telegramId);
      console.log('Статус транзакции:', data);
      
      if (data.tx_status === 'completed') {
        // Транзакция успешно обработана
        showTelegramPopup(
          'Транзакция подтверждена',
          `На ваш счет начислено ${data.will_amount} WILL токенов`
        );
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        showTelegramPopup(
          'Транзакция не удалась',
          'Не удалось обработать вашу транзакцию. Пожалуйста, попробуйте еще раз.'
        );
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      }
    } catch (error) {
      console.error('Ошибка при проверке статуса транзакции:', error);
    }
  }
</script>

<div class="wrapper">
  <div 
    class="overlay" 
    role="button"
    tabindex="0"
    on:click={() => dispatch('close')}
    on:keydown={e => e.key === 'Enter' && dispatch('close')}
  ></div>
  <div class="modal-container">
    <div class="modal">
      <div class="content">
        <div class="mode-selector">
          <button 
            class:active={modalMode === 'buy'} 
            on:click={() => modalMode = 'buy'}
          >
            Купить WILL
          </button>
          <button 
            class:active={modalMode === 'withdraw'} 
            on:click={() => modalMode = 'withdraw'}
          >
            Вывести WILL
          </button>
        </div>

        {#if modalMode === 'buy'}
          <!-- Контент для покупки WILL -->
          <div class="ton-info">
            {#if !walletConnected}
              <p class="wallet-status">Для оплаты необходимо подключить кошелек на главном экране</p>
            {:else}
              <p class="wallet-status">Кошелек подключен: {walletAddress.slice(0, 8)}...{walletAddress.slice(-6)}</p>
            {/if}
          </div>

          <div class="info-block">
            <div class="exchange-rate">
              <span class="label">Курс обмена</span>
              <span class="value">
                1 USDT = {USDT_EXCHANGE_RATE} WILL
              </span>
            </div>

            <div class="input-group">
              <label for="tokens-amount">Количество WILL</label>
              <input
                type="number"
                id="tokens-amount"
                bind:value={tokensAmount}
                min="10"
                step="10"
                placeholder="Введите количество WILL"
              />
            </div>

            <div class="summary">
              <span class="label">К оплате</span>
              <span class="value">
                {calculateUsdt(tokensAmount)} USDT
              </span>
            </div>
          </div>

          {#if transactionError}
            <div class="error-message">
              {transactionError}
            </div>
          {/if}

          <div class="footer">
            <button 
              class="action-btn" 
              on:click={handleBuy}
              disabled={tokensAmount < 10 || !walletConnected || isProcessing}
            >
              {#if isProcessing}
                Обработка...
              {:else}
                Купить
              {/if}
            </button>
          </div>
        {:else}
          <!-- Контент для вывода WILL -->
          <div class="ton-info">
            {#if !walletConnected}
              <p class="wallet-status">Для вывода необходимо подключить кошелек на главном экране</p>
            {:else}
              <p class="wallet-status">Кошелек подключен: {walletAddress.slice(0, 8)}...{walletAddress.slice(-6)}</p>
            {/if}
            
            <!-- Отображаем текущий баланс -->
            <p class="balance-status">
              {#if isLoadingBalance}
                Загрузка баланса...
              {:else}
                Ваш баланс: <strong>{userBalance} WILL</strong>
              {/if}
            </p>
          </div>
          
          <div class="info-block">
            <div class="exchange-rate">
              <span class="label">Курс обмена</span>
              <span class="value">
                1000 WILL = 1 USDT
              </span>
            </div>
            
            <div class="input-group">
              <label for="withdraw-amount">Количество WILL для вывода</label>
              <input
                type="number"
                id="withdraw-amount"
                bind:value={withdrawAmount}
                min="10"
                step="10"
                max={userBalance}
                placeholder="Введите количество WILL для вывода"
              />
            </div>
            
            <div class="summary">
              <span class="label">Вы получите</span>
              <span class="value">
                {(withdrawAmount / USDT_EXCHANGE_RATE).toFixed(2)} USDT
              </span>
            </div>
          </div>
          
          {#if transactionError}
            <div class="error-message">
              {transactionError}
            </div>
          {/if}
          
          <div class="footer">
            <button 
              class="action-btn" 
              on:click={handleWithdraw}
              disabled={withdrawAmount < 10 || withdrawAmount > userBalance || !walletConnected || isProcessing}
            >
              {#if isProcessing}
                Обработка...
              {:else}
                Вывести
              {/if}
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .wrapper {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    z-index: 1000;
  }

  .overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
  }

  .modal-container {
    position: relative;
    width: 100%;
    z-index: 1;
  }

  .modal {
    width: 100%;
    background: #F9F8F3;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    max-height: 90vh;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .content {
    padding: 24px 16px;
  }

  .info-block {
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 16px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .exchange-rate {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .exchange-rate .label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  .exchange-rate .value {
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-group label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  input[type="number"] {
    width: 100%;
    padding: 12px;
    border: 2px solid var(--tg-theme-bg-color);
    border-radius: 12px;
    font-size: 16px;
    background: var(--tg-theme-bg-color);
    color: var(--tg-theme-text-color);
    box-sizing: border-box;
  }

  input[type="number"]:focus {
    outline: none;
    border-color: #00D5A0;
  }

  .summary {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 12px;
    border-top: 1px solid var(--tg-theme-bg-color);
  }

  .summary .label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  .summary .value {
    font-size: 24px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .footer {
    position: sticky;
    bottom: 0;
    background: inherit;
    z-index: 2;
    padding: 12px 16px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
  }

  /* Общий стиль для кнопок действий (купить/вывести) */
  .action-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: #00D5A0;
    color: white;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) input[type="number"] {
    color: #ffffff;
    background: var(--tg-theme-bg-color);
    border-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .info-block {
    background: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) .summary {
    border-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) input::placeholder {
    color: rgba(255, 255, 255, 0.6) !important;
  }

  :global([data-theme="dark"]) .label,
  :global([data-theme="dark"]) .value,
  :global([data-theme="dark"]) label {
    color: #ffffff;
  }

  :global([data-theme="dark"]) .exchange-rate .label {
    color: #ffffff;
    opacity: 0.7;
  }

  :global([data-theme="dark"]) .exchange-rate .value {
    color: #ffffff;
  }

  :global([data-theme="dark"]) .summary .label {
    color: #ffffff;
    opacity: 0.7;
  }

  :global([data-theme="dark"]) .summary .value {
    color: #ffffff;
  }

  :global([data-theme="dark"]) .input-group label {
    color: #ffffff;
    opacity: 0.7;
  }

  .ton-info {
    margin-bottom: 16px;
    padding: 12px;
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
  }

  .wallet-status {
    margin: 0;
    font-size: 14px;
    color: var(--tg-theme-text-color);
    text-align: center;
  }

  :global([data-theme="dark"]) .wallet-status {
    color: #ffffff;
  }

  .error-message {
    margin-top: 16px;
    padding: 12px;
    background: rgba(255, 0, 0, 0.1);
    border-radius: 12px;
    color: #e74c3c;
    font-size: 14px;
    text-align: center;
  }

  :global([data-theme="dark"]) .ton-info {
    background: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) .error-message {
    color: #ff6b6b;
  }

  /* Стили для переключателя режимов */
  .mode-selector {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 20px;
  }
  
  .mode-selector button {
    flex: 1;
    padding: 12px;
    border: none;
    border-radius: 12px;
    background-color: var(--tg-theme-button-color, #50B4F3);
    color: var(--tg-theme-button-text-color, #ffffff);
    font-size: 16px;
    cursor: pointer;
    opacity: 0.7;
    transition: all 0.2s ease;
  }
  
  .mode-selector button.active {
    opacity: 1;
    font-weight: bold;
  }
   
  /* Добавляем стили для темной темы для кнопок режимов */
  :global([data-theme="dark"]) .mode-selector button {
    color: #ffffff;
  }
  
  /* Добавляем стили для темной темы для кнопки действия */
  :global([data-theme="dark"]) .action-btn {
    color: #ffffff;
  }

  /* Добавляем стили для отображения баланса */
  .balance-status {
    margin: 10px 0 0;
    font-size: 14px;
    color: var(--tg-theme-text-color);
    text-align: center;
  }
  
  .balance-status strong {
    font-weight: 600;
  }
  
  :global([data-theme="dark"]) .balance-status {
    color: #ffffff;
  }
</style>

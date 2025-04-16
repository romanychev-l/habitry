<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { getTonConnect, subscribeToWalletChanges, sendTonTransaction } from '../utils/tonConnect';
  import { api } from '../utils/api';
  import type { Wallet } from '@tonconnect/ui';
  import { beginCell, Address, toNano } from '@ton/core'
  import { TonClient, JettonMaster } from '@ton/ton';
  import { popup } from '@telegram-apps/sdk-svelte';
  import InstructionsModal from './InstructionsModal.svelte';

  // Константы с ссылками на инструкции
  const TON_SPACE_GUIDE = 'https://walletru.helpscoutdocs.com/article/84-chto-takoe-ton-space';
  const P2P_GUIDE = 'https://walletru.helpscoutdocs.com/article/74-znakomstvo-s-r2r-marketom';
  const USDT_GUIDE = 'https://walletru.helpscoutdocs.com/article/60-znakomstvo-s-wallet';

  // Предотвращаем скролл на основной странице
  function disableBodyScroll() {
    document.body.style.overflow = 'hidden';
  }
  
  function enableBodyScroll() {
    document.body.style.overflow = '';
  }

  // Используем уже объявленные глобальные типы без declare global
  
  export let telegramId: number;
  export let show = false;
  
  const dispatch = createEventDispatcher();
  let tokensAmount = 100;
  const EXCHANGE_RATE = 10; // 1 Stars = 10 WILL
  const TON_EXCHANGE_RATE = 100; // 1 TON = 100 WILL
  const USDT_EXCHANGE_RATE = 1000; // 1 USDT = 1000 WILL (изменено)
  const MIN_WITHDRAW_AMOUNT = 500; // Минимальная сумма вывода в WILL

  // Изменяем тип paymentMethod, добавляя 'usdt'
  let paymentMethod: 'stars' | 'ton' | 'usdt' = 'usdt';
  // Добавляем новую переменную для переключения между режимами покупки и вывода
  let modalMode: 'buy' | 'withdraw' = 'buy';
  // Добавим переменную для суммы вывода
  let withdrawAmount = 100;
  let walletConnected = false;
  let walletAddress = '';
  let walletAddressFriendly = '';
  let unsubscribe: (() => void) | null = null; 
  let isProcessing = false;
  let transactionError = '';
  
  // Добавляем переменную для хранения баланса пользователя
  let userBalance = 0;
  let isLoadingBalance = false;

  let showInstructions = false;

  onMount(async () => {
    // Подписываемся на изменения состояния кошелька
    unsubscribe = subscribeToWalletChanges((wallet: Wallet | null) => {
      if (wallet) {
        walletConnected = true;
        walletAddress = wallet.account.address;
        walletAddressFriendly = Address.parse(walletAddress).toString({ bounceable: false });
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
    enableBodyScroll();
  });

  // Добавляем обработчик для управления скроллом при изменении видимости модального окна
  $: if (show) {
    disableBodyScroll();
  } else {
    enableBodyScroll();
  }

  // Функция для загрузки баланса пользователя
  async function loadUserBalance() {
    if (!telegramId) return;
    
    try {
      isLoadingBalance = true;
      const userData = await api.getUser({});
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
      await popup.open({
        title: $_('alerts.error'),
        message: $_('alerts.wallet_not_connected'),
        buttons: [{ id: 'close', type: 'close' }]
      });
      return;
    }
    
    // Проверяем минимальную сумму для вывода
    if (withdrawAmount < MIN_WITHDRAW_AMOUNT) {
      console.error('Сумма меньше минимальной для вывода');
      transactionError = $_('payment.min_withdraw_amount', { values: { amount: MIN_WITHDRAW_AMOUNT } });
      await popup.open({
        title: $_('alerts.error'),
        message: $_('payment.min_withdraw_amount', { values: { amount: MIN_WITHDRAW_AMOUNT } }),
        buttons: [{ id: 'close', type: 'close' }]
      });
      return;
    }
    
    // Проверяем баланс
    if (withdrawAmount > userBalance) {
      console.error('Недостаточно средств для вывода');
      transactionError = $_('alerts.insufficient_funds', { values: { balance: userBalance } });
      await popup.open({
        title: $_('alerts.error'),
        message: $_('alerts.insufficient_funds', { values: { balance: userBalance } }),
        buttons: [{ id: 'close', type: 'close' }]
      });
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
          wallet_address: walletAddress
        });
        
        console.log('Ответ сервера:', response);
        
        // Показываем уведомление об успешной обработке запроса
        await popup.open({
          title: $_('alerts.withdrawal_request_sent'),
          message: $_('payment.withdrawal_request_message', { values: { will_amount: withdrawAmount, usdt_amount: usdtAmount } }),
          buttons: [{ id: 'close', type: 'close' }]
        });
        
        // Обновляем баланс пользователя
        await loadUserBalance();
        
        // Закрываем модальное окно
        dispatch('close');
      } catch (apiError: any) {
        console.error('Ошибка при регистрации запроса на вывод:', apiError);
        transactionError = apiError.message || 'Ошибка при регистрации запроса на вывод';
        await popup.open({
          title: $_('alerts.error'),
          message: $_('payment.withdrawal_error'),
          buttons: [{ id: 'close', type: 'close' }]
        });
      }
    } catch (error: any) {
      console.error('Ошибка при обработке запроса на вывод:', error);
      transactionError = $_('payment.unknown_error');
      await popup.open({
        title: $_('alerts.error'),
        message: $_('payment.unknown_error'),
        buttons: [{ id: 'close', type: 'close' }]
      });
    } finally {
      isProcessing = false;
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
      await popup.open({
        title: $_('alerts.error'),
        message: $_('alerts.wallet_not_connected'),
        buttons: [{ id: 'close', type: 'close' }]
      });
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
        transactionError = $_('payment.invalid_usdt_master_address');
        isProcessing = false;
        await popup.open({
          title: $_('alerts.error'),
          message: $_('payment.invalid_usdt_master_address'),
          buttons: [{ id: 'close', type: 'close' }]
        });
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
        transactionError = $_('payment.invalid_app_wallet_address');
        isProcessing = false;
        await popup.open({
          title: $_('alerts.error'),
          message: $_('payment.invalid_app_wallet_address'),
          buttons: [{ id: 'close', type: 'close' }]
        });
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
          await popup.open({
            title: $_('alerts.error'),
            message: $_('payment.wallet_not_connected'),
            buttons: [{ id: 'close', type: 'close' }]
          });
          isProcessing = false;
          return;
        }

        // Проверяем, что адрес мастер-контракта USDT корректный
        if (!usdtMasterAddress.toString().startsWith('EQ')) {
          console.error('Некорректный адрес мастер-контракта USDT');
          transactionError = $_('payment.invalid_usdt_master_address');
          await popup.open({
            title: $_('alerts.error'),
            message: $_('payment.invalid_usdt_master_address'),
            buttons: [{ id: 'close', type: 'close' }]
          });
          isProcessing = false;
          return;
        }

        // Проверяем, что адрес кошелька приложения корректный
        if (!appWalletAddress.toString().startsWith('EQ')) {
          console.error('Некорректный адрес кошелька приложения');
          transactionError = $_('payment.invalid_app_wallet_address');
          await popup.open({
            title: $_('alerts.error'),
            message: $_('payment.invalid_app_wallet_address'),
            buttons: [{ id: 'close', type: 'close' }]
          });
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
          
          transactionError = $_('payment.withdrawal_error');
          
          // Показываем уведомление об ошибке
          await popup.open({
            title: $_('alerts.error'),
            message: $_('payment.withdrawal_error'),
            buttons: [{ id: 'close', type: 'close' }]
          });
        }
      } catch (error) {
        console.error('Ошибка при отправке USDT транзакции:', error);
        transactionError = $_('payment.transaction_send_error');
        await popup.open({
          title: $_('alerts.error'),
          message: $_('payment.transaction_send_error'),
          buttons: [{ id: 'close', type: 'close' }]
        });
      }
    } catch (error) {
      console.error('Ошибка при обработке платежа USDT:', error);
      transactionError = error instanceof Error ? error.message : 'Неизвестная ошибка';
      await popup.open({
        title: $_('alerts.error'),
        message: transactionError,
        buttons: [{ id: 'close', type: 'close' }]
      });
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
    popup.open({
      title: $_('alerts.transaction_sent'),
      message: $_('alerts.ton_transaction_sent')
    });
  }

  async function checkUsdtTransactionStatus(transactionId: string) {
    try {
      console.log('Проверяем статус USDT транзакции:', transactionId);
      const data = await api.checkUsdtTransaction(transactionId);
      console.log('Статус USDT транзакции:', data);
      
      if (data.tx_status === 'completed') {
        // Транзакция успешно обработана
        popup.open({
          title: $_('alerts.transaction_confirmed'),
          message: `На ваш счет начислено ${data.will_amount} WILL токенов`
        });
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkUsdtTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        popup.open({
          title: $_('alerts.transaction_failed'),
          message: 'Не удалось обработать вашу транзакцию. Пожалуйста, попробуйте еще раз.'
        });
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      }
    } catch (error: any) {
      console.error('Ошибка при проверке статуса USDT транзакции:', error);
      
      // Если транзакция не найдена, очищаем localStorage
      if (error.response?.data?.error === 'transaction not found') {
        localStorage.removeItem('last_usdt_tx');
        console.log('Транзакция не найдена, очищаем localStorage');
      }
    }
  }

  async function handleTonPayment() {
    if (!walletConnected) {
      console.error('Кошелек не подключен');
      transactionError = 'Кошелек не подключен. Пожалуйста, подключите кошелек на главном экране.';
      await popup.open({
        title: $_('alerts.error'),
        message: transactionError,
        buttons: [{ id: 'close', type: 'close' }]
      });
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
          await popup.open({
            title: $_('alerts.error'),
            message: transactionError,
            buttons: [{ id: 'close', type: 'close' }]
          });
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
          await popup.open({
            title: $_('alerts.error'),
            message: transactionError,
            buttons: [{ id: 'close', type: 'close' }]
          });
        }
      } catch (error) {
        console.error('Ошибка при отправке транзакции:', error);
        
        if (error instanceof Error && error.message.includes('NullPointerException')) {
          transactionError = 'Ошибка в приложении кошелька. Пожалуйста, попробуйте использовать другой кошелек или обновите приложение.';
        } else {
          transactionError = error instanceof Error ? error.message : 'Произошла ошибка при отправке транзакции';
        }
        
        // Показываем уведомление об ошибке
        await popup.open({
          title: $_('alerts.error'),
          message: transactionError,
          buttons: [{ id: 'close', type: 'close' }]
        });
      }
    } catch (error) {
      console.error('Ошибка при обработке платежа TON:', error);
      transactionError = error instanceof Error ? error.message : 'Неизвестная ошибка';
      await popup.open({
        title: $_('alerts.error'),
        message: transactionError,
        buttons: [{ id: 'close', type: 'close' }]
      });
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
    popup.open({
      title: $_('alerts.transaction_sent'),
      message: $_('alerts.ton_transaction_sent')
    });
  }
  
  // Функция для проверки статуса транзакции
  async function checkTransactionStatus(transactionId: string) {
    try {
      const response = await fetch(`https://lenichev.site/api/wallet/transactions/${transactionId}`);
      const data = await response.json();
      
      if (data.tx_status === 'success') {
        // Транзакция успешна
        popup.open({
          title: $_('payment.success_title'),
          message: $_('payment.success_message')
        });
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        popup.open({
          title: $_('payment.transaction_failed'),
          message: $_('payment.transaction_failed_message')
        });
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      }
    } catch (error) {
      console.error('Ошибка при проверке статуса транзакции:', error);
    }
  }

  // Заменяем функцию openInstructions на:
  function openInstructions() {
    showInstructions = true;
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
            {$_('payment.buy')} WILL
          </button>
          <button 
            class:active={modalMode === 'withdraw'} 
            on:click={() => modalMode = 'withdraw'}
          >
            {$_('payment.withdraw')} WILL
          </button>
        </div>

        {#if modalMode === 'buy'}
          <!-- Контент для покупки WILL -->
          <div class="ton-info">
            {#if !walletConnected}
              <p class="wallet-status">{$_('alerts.wallet_not_connected')}</p>
            {:else}
              <p class="wallet-status">
                {$_('payment.wallet')} {$_('payment.connected')}: 
                {walletAddressFriendly.slice(0, 8)}...{walletAddressFriendly.slice(-6)}
              </p>            
            {/if}
          </div>

          <div class="info-block">
            <div class="exchange-rate">
              <span class="label">{$_('payment.exchange_rate')}</span>
              <span class="value">
                1 USDT = {USDT_EXCHANGE_RATE} WILL
              </span>
            </div>

            <div class="input-group">
              <label for="tokens-amount">{$_('payment.amount')} WILL</label>
              <input
                type="number"
                id="tokens-amount"
                bind:value={tokensAmount}
                min="10"
                step="10"
                placeholder={$_('payment.enter_amount')}
              />
            </div>

            <div class="summary">
              <span class="label">{$_('payment.price')}</span>
              <span class="value">
                {calculateUsdt(tokensAmount)} USDT
              </span>
            </div>

            <button class="instructions-btn" on:click={openInstructions}>
              {$_('payment.how_to_buy')}
            </button>
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
                {$_('payment.processing')}
              {:else}
                {$_('payment.buy')}
              {/if}
            </button>
          </div>
        {:else}
          <!-- Контент для вывода WILL -->
          <div class="ton-info">
            {#if !walletConnected}
              <p class="wallet-status">{$_('alerts.wallet_not_connected')}</p>
            {:else}
              <p class="wallet-status">{$_('payment.wallet')} {$_('payment.connected')}: {walletAddressFriendly.slice(0, 8)}...{walletAddressFriendly.slice(-6)}</p>
            {/if}
          </div>
          
          <div class="info-block">
            <div class="exchange-rate">
              <span class="label">{$_('payment.exchange_rate')}</span>
              <span class="value">
                1000 WILL = 1 USDT
              </span>
            </div>
            
            <div class="input-group">
              <label for="withdraw-amount">{$_('payment.withdraw_amount')}</label>
              <div class="input-with-max">
                <input
                  type="number"
                  id="withdraw-amount"
                  bind:value={withdrawAmount}
                  min={MIN_WITHDRAW_AMOUNT}
                  step="10"
                  max={userBalance}
                  placeholder={$_('payment.enter_withdraw_amount')}
                />
                <button 
                  class="max-btn" 
                  on:click={() => withdrawAmount = userBalance}
                >
                  max
                </button>
              </div>
            </div>
            
            <div class="summary">
              <span class="label">{$_('payment.will_receive')}</span>
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
                {$_('payment.processing')}
              {:else}
                {$_('payment.withdraw')}
              {/if}
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<InstructionsModal 
  show={showInstructions} 
  on:close={() => showInstructions = false} 
/>

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
    background: var(--tg-theme-bg-color);
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    max-height: 90vh;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .content {
    padding: 24px 16px 4px 16px;
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
    border-color: var(--tg-theme-button-color);
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
  }

  /* Общий стиль для кнопок действий (купить/вывести) */
  .action-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-button-color);
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

  /* Стили для кнопки max и контейнера ввода */
  .input-with-max {
    position: relative;
    display: flex;
    width: 100%;
  }

  .max-btn {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color, #ffffff);
    border: none;
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 12px;
    cursor: pointer;
    font-weight: 600;
    z-index: 2;
  }

  .max-btn:hover {
    opacity: 0.9;
  }

  :global([data-theme="dark"]) .max-btn {
    color: #ffffff;
  }

  .instructions-btn {
    width: 100%;
    padding: 12px;
    margin-top: 16px;
    border: 2px solid var(--tg-theme-button-color);
    border-radius: 12px;
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .instructions-btn:hover {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  :global([data-theme="dark"]) .instructions-btn {
    color: #ffffff;
    border-color: #ffffff;
  }

  :global([data-theme="dark"]) .instructions-btn:hover {
    background: #ffffff;
    color: var(--tg-theme-bg-color);
  }
</style>

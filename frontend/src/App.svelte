<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import HabitLinkModal from './components/HabitLinkModal.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import OnboardingModal from './components/OnboardingModal.svelte';
  import UserProfilePage from './components/UserProfilePage.svelte';
  import BuyTokensModal from './components/BuyTokensModal.svelte';
  import { user, balance } from './stores/user';
  import { isListView } from './stores/view';
  import { openTelegramInvoice } from './utils/telegram';
  import { _ } from 'svelte-i18n';
  import { habits } from './stores/habit';
  import type { Habit } from './types';
  import { api } from './utils/api';
  import { onMount, onDestroy } from 'svelte';
  import { subscribeToWalletChanges } from './utils/tonConnect';
  import type { Wallet } from '@tonconnect/ui';
  import { popup, initData, themeParams } from '@telegram-apps/sdk-svelte';
  
  // Инициализируем значение из localStorage
  $isListView = localStorage.getItem('isListView') === 'true';
  let showModal = false;
  let showHabitLinkModal = false;
  let showSettings = false;
  let showOnboarding = false;
  let showBuyTokens = false;
  let sharedHabitId = '';
  let sharedByTelegramId = '';
  let showUserProfile = false;
  let profileUsername = '';
  let isDarkTheme = false;
  let isInitialized = false;
  
  // Переменные для TON Connect
  let walletConnected = false;
  let walletAddress = '';
  let unsubscribeTonConnect: (() => void) | null = null;

  initData.restore();
  
  onMount(() => {
    // Инициализируем Telegram WebApp
    initTelegramWebApp();
    
    // Инициализируем TON Connect
    initTonConnect();
    checkPendingTonTransactions();
  });
  
  // Функция для инициализации Telegram WebApp
  async function initTelegramWebApp() {
    console.log('Инициализация Telegram WebApp через @telegram-apps/sdk-svelte...');

    try {
      // Получаем данные пользователя через initData
      const userData = initData.user();
      console.log('userData', userData);
      if (userData) {
        // Исправляем URL фотографии, заменяя экранированные слэши на обычные
        const photoUrl = userData.photo_url?.replace(/\\\//g, '/');
        
        user.set({
          id: userData.id,
          first_name: userData.first_name,
          username: userData.username,
          language_code: userData.language_code,
          photo_url: photoUrl
        });
        console.log('Пользователь:', userData);
      } else {
        console.warn('Данные пользователя недоступны');
      }
    
      // Инициализируем и привязываем параметры темы
      if (themeParams.mount.isAvailable()) {
        try {
          await themeParams.mount();
          themeParams.bindCssVars();
          isDarkTheme = themeParams.backgroundColor() === '#000000';
          console.log('Тема установлена:', isDarkTheme ? 'dark' : 'light');
        } catch (err) {
          console.error('Ошибка при инициализации темы:', err);
        }
      }
      
      console.log('Telegram WebApp успешно инициализирован');

      // Проверяем параметры запуска после инициализации
      handleStartParam();
    } catch (error) {
      console.error('Ошибка при инициализации Telegram WebApp:', error);
    }
  }
  
  function initTonConnect() {
    try {
      console.log('Инициализация TON Connect...');
      
      // Создаем элемент для кнопки TON Connect, если его еще нет
      if (!document.getElementById('ton-connect')) {
        console.log('Создание элемента ton-connect...');
        const tonConnectElement = document.createElement('div');
        tonConnectElement.id = 'ton-connect';
        tonConnectElement.style.display = 'inline-block';
        
        const container = document.getElementById('ton-connect-container');
        if (container) {
          container.appendChild(tonConnectElement);
          console.log('Элемент ton-connect успешно добавлен в контейнер');
        } else {
          console.error('Контейнер для TON Connect не найден, создаем новый...');
          const newContainer = document.createElement('div');
          newContainer.id = 'ton-connect-container';
          newContainer.className = 'ton-connect-wrapper';
          newContainer.appendChild(tonConnectElement);
          document.querySelector('header')?.appendChild(newContainer);
        }
      } else {
        console.log('Элемент ton-connect уже существует');
      }
      
      // Подписываемся на изменения состояния кошелька
      unsubscribeTonConnect = subscribeToWalletChanges((wallet: Wallet | null) => {
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
    } catch (error) {
      console.error('Ошибка при инициализации TON Connect:', error);
    }
  }
  
  onDestroy(() => {
    // Отписываемся от событий TON Connect
    if (unsubscribeTonConnect) {
      unsubscribeTonConnect();
      unsubscribeTonConnect = null;
    }
  });
  
  // Функция для проверки незавершенных TON транзакций
  async function checkPendingTonTransactions() {
    // const lastTonTx = localStorage.getItem('last_ton_tx');
    // if (lastTonTx) {
    //   console.log('Найдена незавершенная TON транзакция:', lastTonTx);
    //   checkTransactionStatus(lastTonTx);
    // }
    
    // Добавляем проверку USDT-транзакций
    const lastUsdtTx = localStorage.getItem('last_usdt_tx');
    if (lastUsdtTx) {
      console.log('Найдена незавершенная USDT транзакция:', lastUsdtTx);
      checkUsdtTransactionStatus(lastUsdtTx);
    }
  }
  
  // Функция для проверки статуса транзакции
  async function checkTransactionStatus(transactionId: string) {
    try {
      const data = await api.checkTonTransaction(transactionId, $user?.id || 0);
      console.log('Статус транзакции:', data);
      
      if (data.tx_status === 'completed') {
        // Транзакция успешно обработана
        try {
          await popup.open({
            title: $_('alerts.transaction_confirmed'),
            message: $_('alerts.transaction_confirmed_message', { values: { amount: data.will_amount } }),
            buttons: [{ id: 'close', type: 'close' }]
          });
        } catch (error) {
          console.warn('Telegram WebApp API недоступен:', error);
        }
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        try {
          await popup.open({
            title: $_('alerts.transaction_failed'),
            message: $_('alerts.transaction_failed_message'),
            buttons: [{ id: 'close', type: 'close' }]
          });
        } catch (error) {
          console.warn('Telegram WebApp API недоступен:', error);
        }
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_ton_tx');
      }
    } catch (error) {
      console.error('Ошибка при проверке статуса транзакции:', error);
    }
  }
  
  // Функция для проверки статуса USDT-транзакции
  async function checkUsdtTransactionStatus(transactionId: string) {
    try {
      const data = await api.checkUsdtTransaction(transactionId);
      console.log('Статус USDT транзакции:', data);
      
      if (data.status === 'completed') {
        // Транзакция успешно обработана
        try {
          await popup.open({
            title: $_('alerts.transaction_confirmed'),
            message: $_('alerts.transaction_confirmed_message', { values: { amount: data.will_amount } }),
            buttons: [{ id: 'close', type: 'close' }]
          });
        } catch (error) {
          console.warn('Telegram WebApp API недоступен:', error);
        }
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      } else if (data.status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkUsdtTransactionStatus(transactionId);
        }, 60000);
      } else if (data.status === 'failed') {
        // Транзакция не удалась
        try {
          await popup.open({
            title: $_('alerts.transaction_failed'),
            message: $_('alerts.usdt_transaction_failed_message'),
            buttons: [{ id: 'close', type: 'close' }]
          });
        } catch (error) {
          console.warn('Telegram WebApp API недоступен:', error);
        }
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      }
    } catch (error) {
      console.error('Ошибка при проверке статуса USDT транзакции:', error);
    }
  }
  
  async function handleStartParam() {
    try {
      // Получаем параметры через initData
      const startParam = initData.startParam();
      console.log('Start param:', startParam);
      
      if (startParam) {
        if (startParam.startsWith('habit_')) {
          const [_, habitId, sharedByUserId] = startParam.split('_');
          console.log('Показываем окно выбора привычки:', { habitId, sharedByUserId });
          
          sharedHabitId = habitId;
          sharedByTelegramId = sharedByUserId;
          showHabitLinkModal = true;
        } else if (startParam.startsWith('profile_')) {
          const username = startParam.slice(8);
          console.log('Показываем профиль пользователя:', username);
          
          profileUsername = username;
          showUserProfile = true;
        }
      }
    } catch (error) {
      console.error('Ошибка при обработке start param:', error);
    }
  }

  // Добавляем реактивный блок для обработки параметров запуска
  $: {
    if (initData.startParam()) {
      console.log('TelegramWebApp доступен после инициализации, параметры запуска уже обработаны в initTelegramWebApp');
    }
  }

  async function handleHabitLink(event: CustomEvent) {
    console.log('handleHabitLink in App.svelte', event.detail);
    try {
      if (!$user?.id) return;
      
      const data = await api.joinHabit({
        telegram_id: $user.id,
        habit_id: event.detail.habitId,
        shared_by_telegram_id: sharedByTelegramId,
        shared_by_habit_id: sharedHabitId
      });

      habits.update(currentHabits => data.habits || []);
      showHabitLinkModal = false;
    } catch (error) {
      console.error('Ошибка при присоединении к привычке:', error);
      popup.open({
        title: $_('alerts.error'),
        message: $_('alerts.habit_join_error'),
        buttons: [{ id: 'close', type: 'close' }]
      });
    }
  }

  async function initializeUser() {
    if (isInitialized) return;
    
    try {
      console.log('initializeUser', $user);
      console.log('Telegram WebApp (from window):', initData.user());
      
      // Обрабатываем параметры запуска перед проверкой telegramId
      await handleStartParam();
      
      const telegramId = $user?.id;
      if (!telegramId) return;
      
      // Явно получаем часовой пояс пользователя
      const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      console.log('User timezone:', userTimezone);
      
      try {
        // Отправляем все данные пользователя при каждом запросе
        const data = await api.getUser({
          photo_url: $user.photo_url
        });
        
        habits.update(currentHabits => data.habits || []);
        console.log('Setting balance from API response:', data.balance);
        balance.set(data.balance);
        console.log('Balance after set:', $balance);
        
        // Получаем текущую дату в часовом поясе пользователя
        const now = new Date();
        const userDate = now.toLocaleString('en-US', { timeZone: userTimezone }).split(',')[0];
        console.log('User local date:', userDate, 'in timezone', userTimezone);
        console.log('Checking dates:', { last_visit: data.last_visit, userDate, balance: data.balance });
        
        if (data.last_visit !== userDate && data.balance > 0) {
          console.log('Showing invoice and updating last_visit');
          // openTelegramInvoice(data.balance);
        }
      } catch (error) {
        console.error('Ошибка при инициаизации пользователя:', error);
      }
      
      isInitialized = true;
    } catch (error) {
      console.error('Ошибка при инициаизации пользователя:', error);
    }
  }

  $: if ($user) {
    initializeUser();
  }

  async function handleNewHabit(event: { detail: any }) {
    try {
      const telegramId = $user?.id;
      if (!telegramId) {
        console.error('Отсутствует telegramId');
        return;
      }

      console.log('Sending habit data:', event.detail);
      const habitData = {
        title: event.detail.title,
        want_to_become: event.detail.want_to_become,
        days: event.detail.days,
        is_one_time: event.detail.is_one_time,
        is_auto: event.detail.is_auto,
        stake: event.detail.stake
      };

      console.log('Request payload:', JSON.stringify(habitData));

      const newHabit = await api.createHabit(habitData);
      console.log('Response:', newHabit);
      
      habits.update(currentHabits => [...currentHabits, newHabit]);
      showModal = false;
    } catch (error) {
      console.error('Ошибка при создании привычки:', error);
      popup.open({
        title: $_('alerts.error'),
        message: $_('alerts.habit_create_error'),
        buttons: [{ id: 'close', type: 'close' }]
      });
    }
  }

  $: {
    localStorage.setItem('isListView', $isListView.toString());
  }

  $: {
    document.documentElement.setAttribute('data-theme', isDarkTheme ? 'dark' : 'light');
  }

  // Отслеживаем изменения баланса
  $: {
    if ($balance !== undefined && $balance !== null) {
      console.log('Balance updated:', $balance);
    }
  }

  $: habitsList = $habits as Habit[];
  // ($habits as Habit[]).sort((a, b) => {
  //   // Сначала проверяем выполненность за сегодня
  //   const today = new Date().toISOString().split('T')[0];
  //   const aCompletedToday = a.last_click_date == today;
  //   const bCompletedToday = b.last_click_date == today;
    
  //   if (aCompletedToday === bCompletedToday) {
  //     // Если статус выполнения одинаковый, сохраняем исходный порядок
  //     return 0;
  //   }
  //   // Невыполненные привычки идут вверх (false перед true)
  //   return aCompletedToday ? 1 : -1;
  // });

  function handleOnboardingFinish() {
    showOnboarding = false;
    if (!showHabitLinkModal) {
      showModal = true;
    }
  }

  function handleOnboardingSkip() {
    showOnboarding = false;
  }

  function handleTonTransactionSent(event: CustomEvent) {
    const { transactionId } = event.detail;
    console.log('Транзакция TON отправлена:', transactionId);
    
    // Запускаем проверку статуса транзакции
    setTimeout(() => checkTransactionStatus(transactionId), 30000);
  }

  // Добавляем обработчик для USDT транзакций
  function handleUsdtTransactionSent(event: CustomEvent) {
    const { transactionId } = event.detail;
    console.log('Транзакция USDT отправлена:', transactionId);
    
    // Запускаем проверку статуса транзакции
    setTimeout(() => checkUsdtTransactionStatus(transactionId), 30000);
  }
</script>

<main>
  <header>
    <div class="user-info">
      {#if $user}
        <button 
          class="profile-button"
          on:click={() => showSettings = true}
        >
          {#if $user.photo_url}
            <img 
              src={$user.photo_url} 
              alt="Profile" 
              class="profile-photo"
            />
          {:else}
            <div class="profile-placeholder">
              {$user.first_name?.[0] || $user.username?.[0] || '?'}
            </div>
          {/if}
        </button>
      {/if}
    </div>

    <div id="ton-connect-container" class="ton-connect-wrapper"></div>
    
    <div class="balance-container">
      {#if $user}
        <div class="balance">
          {#if $balance !== undefined && $balance !== null}
            {$balance} WILL
          {:else}
            {console.log('Balance is undefined or null:', $balance)}
            0 WILL
          {/if}
        </div>
        <button class="add-balance-button" on:click={() => {
          console.log('Opening buy tokens modal');
          showBuyTokens = true;
        }}>
          +
        </button>
      {/if}
    </div>
  </header>

  <div class="habit-container" class:list-view={$isListView}>
    {#if $user}
      {#each habitsList as habit}
        <HabitCard 
          {habit}
          telegramId={$user.id} 
        />
      {/each}
    {/if}
  </div>

  <button 
    class="add-button"
    on:click={() => showModal = true}
  >
    +
  </button>

  {#if showModal}
    <NewHabitModal 
      show={showModal}
      on:save={handleNewHabit}
      on:close={() => showModal = false}
    />
  {/if}

  {#if showHabitLinkModal}
    <HabitLinkModal
      show={showHabitLinkModal}
      habits={habitsList}
      {sharedHabitId}
      {sharedByTelegramId}
      on:select={handleHabitLink}
      on:close={() => showHabitLinkModal = false}
    />
  {/if}

  {#if showUserProfile && profileUsername}
    <UserProfilePage 
      username={profileUsername} 
      on:back={() => showUserProfile = false} 
    />
  {:else if showSettings}
    <SettingsPage on:back={() => showSettings = false} />
  {:else}
    {#if showOnboarding}
      <OnboardingModal 
        on:finish={handleOnboardingFinish}
        on:skip={handleOnboardingSkip}
        isSharedHabit={showHabitLinkModal}
      />
    {/if}
  {/if}

  {#if showBuyTokens}
    <BuyTokensModal 
      show={showBuyTokens}
      telegramId={$user?.id || 0}
      on:close={() => showBuyTokens = false}
      on:buy={event => {
        showBuyTokens = false;
        if (event.detail.paymentMethod === 'stars') {
          openTelegramInvoice(event.detail.starsAmount);
          console.log('after openTelegramInvoice');
        }
      }}
      on:ton-transaction-sent={handleTonTransactionSent}
      on:usdt-transaction-sent={handleUsdtTransactionSent}
    />
  {/if}

  <div class="creator-container">
    <a href="https://t.me/romanychev" target="_blank" class="creator-link">
      {$_('creator.by')}
    </a>
  </div>
</main>

<style>
  main {
    padding: 20px 8px;
    padding-bottom: 80px;
    max-width: 800px;
    margin: 0 auto;
    box-sizing: border-box;
    min-height: 100vh;
    background-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .habit-container {
    color: white;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 4px 16px 8px 16px;
  }

  .habit-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
    box-sizing: border-box;
    margin-bottom: 20px; /* Уменьшаем отступ снизу для контейнера с привычками */
  }

  .habit-container.list-view {
    gap: 0;
    padding: 0;
  }

  .habit-container :global(.habit-card) {
    width: 100%;
    max-width: 280px;
    aspect-ratio: 1;
    margin: 0 auto;
    box-sizing: border-box;
    position: relative;
    z-index: 1;
  }

  .habit-container.list-view :global(.habit-card) {
    width: 100%;
    max-width: 100%;
    aspect-ratio: auto;
    height: 60px;
    border-radius: 12px;
    margin: 0;
    box-sizing: border-box;
  }

  .add-button {
    position: fixed;
    bottom: 20px;
    right: 20px;
    width: 40px;
    height: 40px;
    border: none;
    background: var(--tg-theme-button-color);;
    color: white;
    font-size: 24px;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    z-index: 2;
    /* padding-bottom: 4px; */
  }

  .user-info {
    display: flex;
    align-items: center;
  }

  .profile-button {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .profile-photo {
    width: 40px;
    height: 40px;
    object-fit: cover;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
    transform: scale(1.1);
  }

  .profile-placeholder {
    width: 40px;
    height: 40px;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 500;
    text-transform: uppercase;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  .balance-container {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .balance {
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .add-balance-button {
    width: 40px;
    height: 40px;
    border: none;
    background: var(--tg-theme-hint-color);
    color: white;
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
    /* padding-bottom: 4px; */
  }

  /* :global([data-theme="dark"]) .balance {
    color: white;
  } */

  /* .payment-button {
    position: fixed;
    bottom: 20px;
    left: 20px;
    padding: 12px 24px;
    border-radius: 28px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 16px;
    font-weight: 500;
  } */

  :global(body) {
    margin: 0;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  :global([data-theme="dark"]) .add-balance-button {
    background: var(--tg-theme-hint-color);
  }

  .ton-connect-wrapper {
    display: flex;
    align-items: center;
  }

  :global(#ton-connect) {
    display: inline-block !important;
  }

  :global(#ton-connect button) {
    height: 40px !important;
    border-radius: 12px !important;
    padding: 0 12px !important;
    transition: all 0.2s ease !important;
  }

  .creator-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 8px 0; /* Уменьшаем отступы сверху и снизу */
  }

  .creator-link {
    color: var(--tg-theme-hint-color);
    text-decoration: none;
    font-size: 14px;
    opacity: 0.7;
    transition: opacity 0.2s ease;
  }

  .creator-link:hover {
    opacity: 1;
  }
</style>
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
  let isDarkTheme = window.Telegram?.WebApp?.colorScheme === 'dark';
  let isInitialized = false;
  
  // Переменные для TON Connect
  let walletConnected = false;
  let walletAddress = '';
  let unsubscribeTonConnect: (() => void) | null = null;
  
  onMount(() => {
    // Инициализируем TON Connect
    initTonConnect();
    checkPendingTonTransactions();
  });
  
  function initTonConnect() {
    try {
      // Создаем элемент для кнопки TON Connect, если его еще нет
      if (!document.getElementById('ton-connect')) {
        const tonConnectElement = document.createElement('div');
        tonConnectElement.id = 'ton-connect';
        tonConnectElement.style.display = 'inline-block';
        const container = document.getElementById('ton-connect-container');
        if (container) {
          container.appendChild(tonConnectElement);
        } else {
          console.error('Контейнер для TON Connect не найден');
          return;
        }
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
    const lastTonTx = localStorage.getItem('last_ton_tx');
    if (lastTonTx) {
      console.log('Найдена незавершенная TON транзакция:', lastTonTx);
      checkTransactionStatus(lastTonTx);
    }
    
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
          // @ts-ignore - игнорируем проблемы с типизацией
          if (window.Telegram && window.Telegram.WebApp && typeof window.Telegram.WebApp.showPopup === 'function') {
            // @ts-ignore
            window.Telegram.WebApp.showPopup({
              title: 'Транзакция подтверждена',
              message: `На ваш счет начислено ${data.will_amount} WILL токенов`,
              buttons: [{ type: 'close' }]
            });
          }
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
          // @ts-ignore - игнорируем проблемы с типизацией
          if (window.Telegram && window.Telegram.WebApp && typeof window.Telegram.WebApp.showPopup === 'function') {
            // @ts-ignore
            window.Telegram.WebApp.showPopup({
              title: 'Транзакция не удалась',
              message: 'Не удалось обработать вашу транзакцию. Пожалуйста, попробуйте еще раз.',
              buttons: [{ type: 'close' }]
            });
          }
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
      const data = await api.checkUsdtTransaction(transactionId, $user?.id || 0);
      console.log('Статус USDT транзакции:', data);
      
      if (data.tx_status === 'completed') {
        // Транзакция успешно обработана
        try {
          // @ts-ignore
          if (window.Telegram?.WebApp?.showPopup) {
            // @ts-ignore
            window.Telegram.WebApp.showPopup({
              title: 'Транзакция подтверждена',
              message: `На ваш счет начислено ${data.will_amount} WILL токенов`,
              buttons: [{ type: 'close' }]
            });
          }
        } catch (error) {
          console.warn('Telegram WebApp API недоступен:', error);
        }
        
        // Удаляем ID транзакции из localStorage
        localStorage.removeItem('last_usdt_tx');
      } else if (data.tx_status === 'pending') {
        // Транзакция все еще в обработке, проверим еще раз через минуту
        setTimeout(() => {
          checkUsdtTransactionStatus(transactionId);
        }, 60000);
      } else if (data.tx_status === 'failed') {
        // Транзакция не удалась
        try {
          // @ts-ignore
          if (window.Telegram?.WebApp?.showPopup) {
            // @ts-ignore
            window.Telegram.WebApp.showPopup({
              title: 'Транзакция не удалась',
              message: 'Не удалось обработать вашу USDT транзакцию. Пожалуйста, попробуйте еще раз.',
              buttons: [{ type: 'close' }]
            });
          }
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
      const startParam = window.Telegram?.WebApp?.initDataUnsafe?.start_param;
      console.log('Start param:', startParam);
      
      if (startParam) {
        if (startParam.startsWith('habit_')) {
          const [_, habitId, sharedByUserId] = startParam.split('_');
          console.log('Показываем окно выбора привычки:', { habitId, sharedByUserId });
          
          sharedHabitId = habitId;
          sharedByTelegramId = sharedByUserId;
          showHabitLinkModal = true;
        } else if (startParam.startsWith('profile_')) {
          const username = startParam.split('_')[1];
          console.log('Показываем профиль пользователя:', username);
          
          profileUsername = username;
          showUserProfile = true;
        }
      }
    } catch (error) {
      console.error('Ошибка при обработке start param:', error);
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
    }
  }

  async function initializeUser() {
    if (isInitialized) return;
    
    try {
      console.log('initializeUser', $user);
      console.log('Telegram WebApp:', window.Telegram?.WebApp);
      console.log('Telegram initData:', window.Telegram?.WebApp?.initData);
      console.log('Telegram initDataUnsafe:', window.Telegram?.WebApp?.initDataUnsafe);
      
      const telegramId = $user?.id;
      if (!telegramId) return;
      
      await handleStartParam();

      const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      
      try {
        const data = await api.getUser(telegramId, userTimezone);
        habits.update(currentHabits => data.habits || []);
        console.log('Setting balance from API response:', data.balance);
        balance.set(data.balance);
        console.log('Balance after set:', $balance);
        
        const now = new Date();
        const userDate = now.toLocaleString('en-US', { timeZone: userTimezone }).split(',')[0];
        console.log('Checking dates:', { last_visit: data.last_visit, userDate, balance: data.balance });
        
        if (data.last_visit !== userDate && data.balance > 0) {
          console.log('Showing invoice and updating last_visit');
          // openTelegramInvoice(data.balance);
          
          try {
            console.log('Updating last_visit', telegramId, userTimezone);
            await api.updateLastVisit({
              telegram_id: telegramId,
              timezone: userTimezone
            });
            console.log('Last visit updated successfully');
          } catch (error) {
            console.error('Error updating last_visit:', error);
          }
        }
      } catch (error) {
        console.log('error user', error);
        console.log('error type:', typeof error);
        console.log('error message:', error instanceof Error ? error.message : 'not an Error instance');
        if (error instanceof Error && (error.message.includes('404') || error.message.startsWith('404:'))) {
          console.log('create user');
          showOnboarding = true;
          await api.createUser({
            telegram_id: telegramId,
            first_name: $user.firstName,
            username: $user.username,
            language_code: $user.languageCode,
            photo_url: $user.photoUrl,
            timezone: userTimezone,
            balance: 1000
          });
          balance.set(1000);
          habits.update(currentHabits => []);
        } else {
          throw error;
        }
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
        telegram_id: telegramId,
        habit: {
          title: event.detail.title,
          want_to_become: event.detail.want_to_become,
          days: event.detail.days,
          is_one_time: event.detail.is_one_time,
          is_auto: event.detail.is_auto,
          stake: event.detail.stake
        }
      };

      console.log('Request payload:', JSON.stringify(habitData));

      const data = await api.createHabit(habitData);
      console.log('Response:', data);
      
      if (data.habit) {
        habits.update(currentHabits => [...currentHabits, data.habit]);
        showModal = false;
      } else {
        console.error('No habit in response');
        throw new Error('No habit in response');
      }
    } catch (error) {
      console.error('Ошибка при создании привычки:', error);
      alert($_('habits.errors.create'));
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
    {#if $user}
      <div class="user-info">
        <button 
          class="profile-button"
          on:click={() => showSettings = true}
        >
          {#if $user.photoUrl}
            <img 
              src={$user.photoUrl} 
              alt="Profile" 
              class="profile-photo"
            />
          {:else}
            <div class="profile-placeholder">
              {$user.firstName?.[0] || $user.username?.[0] || '?'}
            </div>
          {/if}
        </button>
      </div>

      <div class="right-section">
        <div id="ton-connect-container" class="ton-connect-wrapper"></div>
        
        <div class="balance-container">
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
        </div>
      </div>
    {/if}
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
      on:save={handleNewHabit}
      on:close={() => showModal = false}
    />
  {/if}

  {#if showHabitLinkModal}
    <HabitLinkModal
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
      telegramId={$user?.id || 0}
      on:close={() => showBuyTokens = false}
      on:buy={event => {
        showBuyTokens = false;
        if (event.detail.paymentMethod === 'stars') {
          openTelegramInvoice(event.detail.starsAmount);
          console.log('after openTelegramInvoice');
        }
        // Закомментируем обработку TON
        /* 
        else {
          console.log('TON transaction sent:', event.detail);
        }
        */
      }}
      on:ton-transaction-sent={handleTonTransactionSent}
      on:usdt-transaction-sent={handleUsdtTransactionSent}
    />
  {/if}
</main>

<style>
  main {
    padding: 20px 8px;
    padding-bottom: 80px;
    max-width: 800px;
    margin: 0 auto;
    box-sizing: border-box;
    min-height: 100vh;
    background-color: #F9F8F3;
  }

  :global([data-theme="dark"]) main {
    background-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) body {
    background-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .habit-container {
    color: white;
  }

  :global([data-theme="dark"]) .toggle-label {
    color: white;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 12px 16px 8px 16px;
  }

  .view-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .toggle-label {
    color: var(--tg-theme-text-color);
    font-size: 14px;
  }

  .switch {
    position: relative;
    display: inline-block;
    width: 40px;
    height: 20px;
  }

  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #ccc;
    transition: .3s;
    border-radius: 20px;
  }

  .slider:before {
    position: absolute;
    content: "";
    height: 16px;
    width: 16px;
    left: 2px;
    bottom: 2px;
    background-color: white;
    transition: .3s;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: #00D5A0;
  }

  input:checked + .slider:before {
    transform: translateX(20px);
  }

  .habit-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
    box-sizing: border-box;
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
    background: #00D5A0;
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
    border: 2px solid var(--tg-theme-button-color);
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
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

  :global([data-theme="dark"]) .balance {
    color: white;
  }

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

  :global([data-theme="dark"]) .add-button {
    background: #00D5A0;
  }

  :global([data-theme="dark"]) .add-balance-button {
    background: var(--tg-theme-hint-color);
  }

  .right-section {
    display: flex;
    align-items: center;
    gap: 12px;
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
</style>
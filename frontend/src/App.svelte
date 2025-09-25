<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import HabitLinkModal from './components/HabitLinkModal.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import OnboardingModal from './components/OnboardingModal.svelte';
  import UserProfilePage from './components/UserProfilePage.svelte';
  import LeaderboardModal from './components/LeaderboardModal.svelte';
  import ArchivedHabitsModal from './components/ArchivedHabitsModal.svelte';
  // import BuyTokensModal from './components/BuyTokensModal.svelte';
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
  import { popup, initData, themeParams, swipeBehavior, viewport } from '@telegram-apps/sdk-svelte';
  import plusIcon from './assets/plus.svg'; // Import the SVG
  import trophyIcon from './assets/trophy.svg';
  
  // Инициализируем значение из localStorage
  $isListView = localStorage.getItem('isListView') === 'true';
  let showModal = false;
  let showHabitLinkModal = false;
  let showSettings = false;
  let showOnboarding = false;
  let showBuyTokens = false;
  let showLeaderboard = false;
  let showArchived = false;
  let sharedHabitId = '';
  let sharedByTelegramId = '';
  let showUserProfile = false;
  let profileUsername = '';
  let cameFromLeaderboard = false;
  let isDarkTheme = false;
  let isInitialized = false;
  let isHabitCardModalOpen = false;

  $: isAnyModalOpen = showModal || showHabitLinkModal || showSettings || showOnboarding || showUserProfile || showLeaderboard || showArchived || isHabitCardModalOpen;
  
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
      console.log('Проверка доступности themeParams.mount:', themeParams.mount.isAvailable());
      if (themeParams.mount.isAvailable()) {
        try {
          console.log('Попытка монтирования темы...');
          await themeParams.mount();
          console.log('Тема успешно смонтирована');
          
          console.log('Привязка CSS переменных...');
          themeParams.bindCssVars();
          console.log('CSS переменные привязаны');
          
          const bgColor = themeParams.backgroundColor();
          console.log('Цвет фона темы:', bgColor);
          isDarkTheme = bgColor === '#000000';
          console.log('Тема установлена:', isDarkTheme ? 'dark' : 'light');
          
          // Проверяем значения CSS переменных
          console.log('CSS переменные темы:', {
            bgColor: getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-bg-color'),
            textColor: getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-text-color'),
            buttonColor: getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-button-color'),
            buttonTextColor: getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-button-text-color'),
            secondaryBgColor: getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-secondary-bg-color')
          });
        } catch (err) {
          console.error('Ошибка при инициализации темы:', err);
        }
      } else {
        console.warn('themeParams.mount недоступен');
      }
      
      console.log('Telegram WebApp успешно инициализирован');

      // Проверяем параметры запуска после инициализации
      handleStartParam();

      // Управление поведением свайпа
      console.log('Проверка доступности swipeBehavior.mount:', swipeBehavior.mount.isAvailable());
      if (swipeBehavior.mount.isAvailable()) {
        try {
          console.log('Попытка монтирования swipeBehavior...');
          await swipeBehavior.mount();
          console.log('swipeBehavior успешно смонтирован');

          console.log('Проверка доступности swipeBehavior.disableVertical:', swipeBehavior.disableVertical.isAvailable());
          if (swipeBehavior.disableVertical.isAvailable()) {
            swipeBehavior.disableVertical();
            console.log('Вертикальный свайп отключен');
          } else {
            console.warn('swipeBehavior.disableVertical недоступен');
          }
        } catch (err) {
          console.error('Ошибка при настройке swipeBehavior:', err);
        }
      } else {
        console.warn('swipeBehavior.mount недоступен');
      }

      // Управление viewport
      console.log('Проверка доступности viewport.mount:', viewport.mount.isAvailable());
      if (viewport.expand.isAvailable()) {
        viewport.expand();
      }
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
          cameFromLeaderboard = false;
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

  function handleUserSelectFromLeaderboard(event: CustomEvent) {
    profileUsername = event.detail.username;
    showUserProfile = true;
    showLeaderboard = false;
    cameFromLeaderboard = true;
  }

  function handleUserProfileBack() {
    showUserProfile = false;
    profileUsername = '';
    if (cameFromLeaderboard) {
      showLeaderboard = true;
      cameFromLeaderboard = false;
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

  $: habitsList = ($habits as Habit[]).filter(h => !h.archived);
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
        <!--
        <button class="add-balance-button" on:click={() => {
          console.log('Opening buy tokens modal');
          showBuyTokens = true;
        }}>
          <img src={plusIcon} alt="Add Balance" />
        </button>
        -->
      {/if}
    </div>
  </header>

  <div class="habit-container" class:list-view={$isListView}>
    {#if $user}
      {#each habitsList as habit}
        <HabitCard 
          {habit}
          telegramId={$user.id} 
          on:modalOpened={() => isHabitCardModalOpen = true}
          on:modalClosed={() => isHabitCardModalOpen = false}
        />
      {/each}
      <div class="archive-entry">
        <div 
          class="archive-entry-inner"
          role="button"
          tabindex="0"
          on:click={() => showArchived = true}
          on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && (showArchived = true)}
        >
          <span class="archive-emoji">📦</span>
          <span class="archive-title">{$_('habits.archived.title')}</span>
        </div>
      </div>
    {/if}
  </div>

  <button 
    class="leaderboard-button"
    class:behind-modal={isAnyModalOpen}
    on:click={() => showLeaderboard = true}
  >
    <img src={trophyIcon} alt="Leaderboard" />
  </button>


  <button 
    class="add-button"
    class:behind-modal={isAnyModalOpen}
    on:click={() => showModal = true}
  >
    <img src={plusIcon} alt="Add Habit" />
  </button>

  {#if showLeaderboard}
    <LeaderboardModal 
      show={showLeaderboard}
      on:close={() => showLeaderboard = false}
      on:userselect={handleUserSelectFromLeaderboard}
    />
  {/if}

  {#if showArchived}
    <ArchivedHabitsModal
      show={showArchived}
      on:close={() => showArchived = false}
      on:unarchived={(e) => {
        // Добавляем восстановленную привычку в основной список
        const restored = e.detail.habit as Habit;
        habits.update(current => {
          const exists = current.some(h => h._id === restored._id);
          return exists ? current.map(h => h._id === restored._id ? restored : h) : [...current, restored];
        });
      }}
    />
  {/if}

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
      on:back={handleUserProfileBack} 
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

  <!-- {#if showBuyTokens}
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
  {/if} -->

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

  .leaderboard-button {
    position: fixed;
    bottom: 20px;
    left: 20px;
    background-color: var(--tg-theme-button-color);
    color: var(--button-text-color);
    border: none;
    width: 40px;
    height: 40px;
    display: flex;
    justify-content: center;
    align-items: center;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    transition: background-color 0.2s;
    z-index: 1000;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  .archive-entry {
    width: 100%;
    display: flex;
    justify-content: center;
    margin-top: 8px;
  }
  .archive-entry-inner {
    width: auto;
    min-width: 140px;
    max-width: 220px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 10px 14px;
    border-radius: 16px;
    background: var(--tg-theme-secondary-bg-color);
    color: var(--tg-theme-text-color);
    cursor: pointer;
    user-select: none;
  }
  .archive-emoji { font-size: 18px; }
  .archive-title { font-weight: 600; }

  .leaderboard-button.behind-modal {
    z-index: 1;
  }

  .leaderboard-button img {
    width: 24px;
    height: 24px;
    filter: brightness(0) invert(1);
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
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    z-index: 2;
  }

  .add-button.behind-modal {
    z-index: 1;
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
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  .add-button img,
  .add-balance-button img {
    width: 80%;
    height: 80%;
    filter: brightness(0) invert(1);
  }

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
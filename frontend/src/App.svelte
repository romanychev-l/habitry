<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import HabitLinkModal from './components/HabitLinkModal.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import OnboardingModal from './components/OnboardingModal.svelte';
  import UserProfilePage from './components/UserProfilePage.svelte';
  import { user } from './stores/user';
  import { isListView } from './stores/view';
  import { openTelegramInvoice } from './utils/telegram';
  import { _ } from 'svelte-i18n';
  import { habits } from './stores/habit';
  import type { Habit } from './types';
  
  // Инициализируем значение из localStorage
  $isListView = localStorage.getItem('isListView') === 'true';
  let showModal = false;
  let showHabitLinkModal = false;
  let showSettings = false;
  let showOnboarding = false;
  let sharedHabitId = '';
  let sharedByTelegramId = '';
  let showUserProfile = false;
  let profileUsername = '';
  const API_URL = import.meta.env.VITE_API_URL;
  let isDarkTheme = window.Telegram?.WebApp?.colorScheme === 'dark';
  let isInitialized = false;
  
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
      
      const response = await fetch(`${API_URL}/habit/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          telegram_id: $user.id,
          habit_id: event.detail.habitId,
          shared_by_telegram_id: sharedByTelegramId,
          shared_by_habit_id: sharedHabitId
        })
      });

      if (!response.ok) {
        throw new Error('Ошибка при присоединении к привычке');
      }
      
      const data = await response.json();
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
      const telegramId = $user?.id;
      if (!telegramId) return;
      
      await handleStartParam();

      const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      const response = await fetch(`${API_URL}/user?telegram_id=${telegramId}&timezone=${userTimezone}`);
      
      if (response.status === 404) {
        console.log('create user');
        showOnboarding = true;
        const createResponse = await fetch(`${API_URL}/user`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            telegram_id: telegramId,
            first_name: $user.firstName,
            username: $user.username,
            language_code: $user.languageCode,
            photo_url: $user.photoUrl,
            timezone: userTimezone
          })
        });
        if (!createResponse.ok) {
          throw new Error($_('habits.errors.user_create'));
        }

        habits.update(currentHabits => []);
      } else {
        const data = await response.json();
        habits.update(currentHabits => data.habits || []);
        
        const now = new Date();
        const userDate = now.toLocaleString('en-US', { timeZone: userTimezone }).split(',')[0];
        console.log('Checking dates:', { last_visit: data.last_visit, userDate, credit: data.credit });
        
        if (data.last_visit !== userDate && data.credit > 0) {
          console.log('Showing invoice and updating last_visit');
          openTelegramInvoice(data.credit);
          
          try {
            const visitResponse = await fetch(`${API_URL}/user/visit`, {
              method: 'PUT',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                telegram_id: telegramId,
                timezone: userTimezone
              })
            });
            
            if (!visitResponse.ok) {
              console.error('Failed to update last_visit:', await visitResponse.text());
            } else {
              console.log('Last visit updated successfully');
            }
          } catch (error) {
            console.error('Error updating last_visit:', error);
          }
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
          is_one_time: event.detail.is_one_time
        }
      };

      console.log('Request payload:', JSON.stringify(habitData));

      const response = await fetch(`${API_URL}/habit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(habitData)
      });

      console.log('Response status:', response.status);
      const responseText = await response.text();
      console.log('Response text:', responseText);

      if (!response.ok) {
        throw new Error($_('habits.errors.create'));
      }

      const data = JSON.parse(responseText);
      console.log('Parsed response:', data);
      
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
    {/if}
    <div class="view-toggle">
      <span class="toggle-label">{$_('habits.compact_view')}</span>
      <label class="switch">
        <input type="checkbox" bind:checked={$isListView}>
        <span class="slider"></span>
      </label>
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
    padding: 8px 16px;
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
    width: 56px;
    height: 56px;
    border-radius: 50%;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 24px;
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
    color: white;
  }
</style>
<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import { user } from './stores/user';
  import { openTelegramInvoice } from './utils/telegram';
  
  let isListView = localStorage.getItem('isListView') === 'true';
  let showModal = false;
  let habits: any[] = [];
  const API_URL = import.meta.env.VITE_API_URL;
  
  async function initializeUser() {
    try {
      console.log('initializeUser', $user);
      const telegramId = $user?.id;
      if (!telegramId) return;

      const response = await fetch(`${API_URL}/user?telegram_id=${telegramId}`);
      console.log('response', response);
      console.log('response.ok', response.status);
      // console.log('response.json', await response.json());
      console.log('telegramId', telegramId);
      if (response.status === 404) {
        console.log('create user');
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
            photo_url: $user.photoUrl
          })
        });
        console.log('createResponse', createResponse);
        if (!createResponse.ok) {
          throw new Error('Ошибка при создании пользователя');
        }

        habits = [];
      } else {
        console.log('user already exists');
        const data = await response.json();
        habits = data.habits || [];
        
        const today = new Date().toISOString().split('T')[0];
        console.log('today', today);
        console.log('data', data);
        console.log('data.last_visit', data.last_visit);
        if (data.last_visit !== today && data.credit > 0) {
          openTelegramInvoice(data.credit);
        }
      }
    } catch (error) {
      console.error('Ошибка при инициализации пользователя:', error);
    }
  }

  $: if ($user) {
    initializeUser();
  }

  async function handleNewHabit(event: { detail: any }) {
    console.log('handleNewHabit', event);
    try {
      const telegramId = $user?.id;
      if (!telegramId) return;

      const newHabit = {
        id: Date.now().toString(),
        score: 0,
        streak: 0,
        ...event.detail
      };
      console.log('newHabit', newHabit);
      const response = await fetch(`${API_URL}/habit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          telegram_id: telegramId,
          habit: newHabit
        })
      });
      console.log('response', response);
      if (!response.ok) {
        throw new Error('Ошибка при сохранении привычки');
      }

      habits = [...habits, newHabit];
      showModal = false;
    } catch (error) {
      console.error('Ошибка при создании привычки:', error);
      alert('Не удалось создать привычку. Попробуйте еще раз.');
    }
  }

  function handlePayment() {
    openTelegramInvoice(1);
  }

  $: {
    localStorage.setItem('isListView', isListView.toString());
  }
</script>

<main>
  <header>
    {#if $user}
      <div class="user-info">
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
      </div>
    {/if}
    <div class="view-toggle">
      <span class="toggle-label">Compact View</span>
      <label class="switch">
        <input type="checkbox" bind:checked={isListView}>
        <span class="slider"></span>
      </label>
    </div>
  </header>

  <div class="habit-container" class:list-view={isListView}>
    {#if $user}
      {#each habits as habit}
        <HabitCard {habit} telegramId={$user.id} />
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

  <!-- <button 
    class="payment-button"
    on:click={handlePayment}
  >
    💎 Премиум
  </button> -->
</main>

<style>
  main {
    padding: 20px;
    padding-bottom: 80px;
    max-width: 800px;
    margin: 0 auto;
    box-sizing: border-box;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 8px 0;
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
    background-color: green;
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
    margin: 4px 0;
    padding: 8px 16px;
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

  .profile-photo {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid var(--tg-theme-button-color);
  }

  .profile-placeholder {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 500;
    text-transform: uppercase;
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
</style>
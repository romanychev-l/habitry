<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import { user } from './stores/user';
  import { openTelegramInvoice } from './utils/telegram';
  
  let isListView = false;
  let showModal = false;
  let habits: any[] = [];

  async function initializeUser() {
    try {
      console.log('initializeUser', $user);
      const telegramId = $user?.id;
      if (!telegramId) return;

      const response = await fetch(`https://lenichev.site/ht_back/user?telegram_id=${telegramId}`);
      console.log('response', response);
      console.log('telegramId', telegramId);
      if (response.status === 404) {
        const createResponse = await fetch('https://lenichev.site/ht_back/user', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            telegram_id: telegramId,
            first_name: $user.firstName,
            username: $user.username,
          })
        });

        if (!createResponse.ok) {
          throw new Error('Ошибка при создании пользователя');
        }

        habits = [];
      } else {
        console.log('user already exists');
        const data = await response.json();
        habits = data.habits || [];
      }
    } catch (error) {
      console.error('Ошибка при инициализации пользователя:', error);
    }
  }

  $: if ($user) {
    initializeUser();
  }

  async function handleNewHabit(event: { detail: any }) {
    try {
      const telegramId = $user?.id;
      if (!telegramId) return;

      const newHabit = {
        id: Date.now().toString(),
        score: 0,
        streak: 0,
        goal: 'Ежедневно',
        ...event.detail
      };

      const response = await fetch('https://lenichev.site/ht_back/habits', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          telegram_id: telegramId,
          habit: newHabit
        })
      });

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
</script>

<main>
  <header>
    {#if $user}
      <div class="user-info">
        <span class="greeting">Привет, {$user.username}! 👋</span>
      </div>
    {/if}
    <button class="view-toggle" on:click={() => isListView = !isListView}>
      {isListView ? '📱' : '📋'}
    </button>
  </header>

  <div class="habit-container" class:list-view={isListView}>
    {#each habits as habit}
      <HabitCard {habit} />
    {/each}
  </div>

  <button 
    class="add-button"
    on:click={() => showModal = true}
  >
    +
  </button>

  {#if showModal}
    <NewHabitModal on:save={handleNewHabit} />
  {/if}

  <button 
    class="payment-button"
    on:click={handlePayment}
  >
    💎 Премиум
  </button>
</main>

<style>
  main {
    padding: 20px;
    padding-bottom: 80px;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 8px 0;
  }

  .view-toggle {
    background: none;
    border: none;
    font-size: 24px;
    padding: 8px;
  }

  .habit-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .habit-container.list-view :global(.habit-card) {
    border-radius: 12px;
    height: 25px;
    margin: 4px 0;
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
    font-size: 18px;
    color: var(--tg-theme-text-color);
  }

  .greeting {
    font-weight: 500;
  }

  .payment-button {
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
  }
</style>
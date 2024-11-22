<script lang="ts">
  import HabitCard from './components/HabitCard.svelte';
  import NewHabitModal from './components/NewHabitModal.svelte';
  import { user } from './stores/user';
  import { openTelegramInvoice } from './utils/telegram';
  
  let isListView = false;
  let showModal = false;
  let habits = [
    { id: '1', title: 'Медитация', days: [1, 3, 5], goal: 'Ежедневно', score: 0, streak: 0 },
    { id: '2', title: 'Бег', days: [0, 2, 4], goal: 'Ежедневно', score: 0, streak: 0 },
  ];

  function handleNewHabit(event: CustomEvent) {
    const newHabit = {
      id: Date.now().toString(),
      score: 0,
      streak: 0,
      goal: 'Ежедневно',
      ...event.detail
    };
    habits = [...habits, newHabit];
    showModal = false;
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
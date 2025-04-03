<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import type { Habit } from '../types';
  import { popup } from '@telegram-apps/sdk-svelte';

  const dispatch = createEventDispatcher();
  
  export let habits: Habit[] = [];
  export let sharedHabitId: string;
  export let sharedByTelegramId: string;
  export let show = false;

  function handleHabitSelect(habit_id: string) {
    dispatch('select', {
      habitId: habit_id,
      sharedHabitId,
      sharedByTelegramId
    });
  }

  function handleClose() {
    dispatch('close');
  }

  // Предотвращаем скролл на основной странице
  function disableBodyScroll() {
    document.body.style.overflow = 'hidden';
  }
  
  function enableBodyScroll() {
    document.body.style.overflow = '';
  }

  onMount(() => {
    if (show) {
      disableBodyScroll();
    }
  });

  onDestroy(() => {
    enableBodyScroll();
  });

  // Добавляем обработчик для управления скроллом при изменении видимости модального окна
  $: if (show) {
    disableBodyScroll();
  } else {
    enableBodyScroll();
  }
</script>

<div 
  class="overlay" 
  on:click={handleClose}
  on:keydown={(e) => e.key === 'Escape' && handleClose()}
  role="button"
  tabindex="0"
>
  <section 
    class="modal" 
    role="dialog"
  >
    <div class="header">
      <h2>{$_('habits.link.title')}</h2>
    </div>

    <div class="content">
      <p class="description">{$_('habits.link.description')}</p>
      
      <div class="habits-list">
        <button 
          class="habit-item create-button"
          on:click={() => handleHabitSelect(sharedHabitId)}
        >
          <h3>{$_('habits.create_new')}</h3>
        </button>

        {#each habits as habit}
          <button 
            class="habit-item"
            on:click={() => handleHabitSelect(habit._id)}
          >
            <h3>{habit.title}</h3>
            {#if habit.want_to_become}
              <p class="want-to-become">{habit.want_to_become}</p>
            {/if}
          </button>
        {/each}
      </div>
    </div>
  </section>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    z-index: 1000;
  }

  .modal {
    width: 100%;
    background: var(--tg-theme-bg-color, #F9F8F3);
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  @supports (-webkit-touch-callout: none) {
    .overlay {
      position: absolute;
      height: 100vh;
      min-height: -webkit-fill-available;
    }

    .modal:focus-within {
      transform: translateY(-35vh);
    }
  }

  .header {
    padding: 32px 16px 16px 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
    flex-shrink: 0;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .content {
    padding: 24px;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .description {
    margin: 0 0 24px 0;
    text-align: center;
    color: var(--tg-theme-hint-color);
  }

  .habits-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .habit-item {
    width: 100%;
    padding: 16px;
    border: 2px solid var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    background: none;
    text-align: left;
    transition: border-color 0.2s;
  }

  .habit-item:active {
    border-color: var(--tg-theme-button-color);
  }

  .habit-item h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .want-to-become {
    margin: 8px 0 0 0;
    font-size: 14px;
    color: var(--tg-theme-hint-color);
  }

  .create-button {
    background: var(--tg-theme-button-color, #00D5A0);
    border: none;
  }

  .create-button h3 {
    color: var(--tg-theme-button-text-color, white);
    text-align: center;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .modal * {
    color: white !important;
  }

  :global([data-theme="dark"]) .create-button h3 {
    color: var(--tg-theme-button-text-color) !important;
  }

  :global([data-theme="dark"]) .want-to-become {
    color: rgba(255, 255, 255, 0.6) !important;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.6) !important;
  }
</style> 
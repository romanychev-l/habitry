<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import type { Habit } from '../types';

  const dispatch = createEventDispatcher();
  const BOT_USERNAME = import.meta.env.VITE_BOT_USERNAME;
  export let habit: Habit;

  function handleShare() {
    const telegram = window.Telegram?.WebApp;
    const userId = telegram?.initDataUnsafe?.user?.id || '';
    const baseUrl = `https://t.me/${BOT_USERNAME}/app`;
    const startAppParam = `startapp=habit_${habit._id}_${userId}`;
    const appUrl = `${baseUrl}?${startAppParam}`;
    const shareText = `\n${$_('habits.join_habit.0')} ${habit.title} ${$_('habits.join_habit.1')} ${habit.streak} ${$_('habits.join_habit.2')}`;
    
    const url = `https://t.me/share/url?url=${encodeURIComponent(appUrl)}&text=${encodeURIComponent(shareText)}`;
    window.open(url, '_blank');
    dispatch('close');
  }
</script>

<div 
  class="dialog-overlay" 
  on:click|stopPropagation={() => dispatch('close')}
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  role="button"
  tabindex="0"
>
  <div class="dialog">
    <div class="dialog-header">
      <h2>{habit.title}</h2>
    </div>
    <div class="dialog-content">
      <button 
        class="dialog-button share"
        on:click={handleShare}
      >
        {$_('habits.share')}
      </button>
      <button 
        class="dialog-button edit"
        on:click={() => {
          dispatch('close');
          dispatch('showEditModal');
        }}
      >
        {$_('habits.edit')}
      </button>
      <button 
        class="dialog-button delete"
        on:click={() => {
          dispatch('close');
          dispatch('showDeleteConfirm');
        }}
      >
        {$_('habits.delete')}
      </button>
    </div>
  </div>
</div>

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    z-index: 1000;
  }

  .dialog {
    width: 100%;
    background: #F9F8F3;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
  }

  .dialog-header {
    padding: 32px 16px 16px 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }

  .dialog-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .dialog-content {
    padding: 24px;
  }

  .dialog-button {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    font-size: 16px;
    font-weight: 500;
    text-align: center;
  }

  .dialog-button.edit {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    margin-bottom: 12px;
  }

  .dialog-button.share {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    margin-bottom: 12px;
  }

  .dialog-button.delete {
    background: #ff3b30;
    color: white;
  }

  :global([data-theme="dark"]) .dialog {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .dialog * {
    color: white !important;
  }
</style> 
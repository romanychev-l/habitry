<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();
</script>

<div 
  class="dialog-overlay"
  on:click|stopPropagation={() => dispatch('close')}
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  role="button"
  tabindex="0"
  transition:fade={{ duration: 200 }}
>
  <div class="dialog" transition:fly={{ y: 500, duration: 300, opacity: 1 }}>
    <div class="dialog-header">
      <h2>{$_('habits.modals.delete_title')}</h2>
    </div>
    <div class="dialog-content">
      <div class="dialog-buttons">
        <button 
          class="dialog-button cancel"
          on:click={() => dispatch('close')}
        >
          {$_('habits.modals.delete_cancel')}
        </button>
        <button 
          class="dialog-button delete"
          on:click={() => dispatch('delete')}
        >
          {$_('habits.modals.delete_confirm')}
        </button>
      </div>
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
    background: var(--tg-theme-bg-color);
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

  .dialog-buttons {
    display: flex;
    gap: 12px;
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

  .dialog-button.cancel {
    background: var(--tg-theme-secondary-bg-color);
    color: var(--tg-theme-text-color);
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
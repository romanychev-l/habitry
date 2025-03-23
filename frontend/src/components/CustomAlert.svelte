<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';

  // Props
  export let title: string = '';
  export let message: string = '';
  export let showConfirm: boolean = false;
  export let confirmText: string = $_('alerts.confirm');
  export let cancelText: string = $_('alerts.cancel');

  // Event dispatcher для закрытия и подтверждения
  const dispatch = createEventDispatcher();

  // Действия
  function handleClose() {
    dispatch('close');
  }

  function handleConfirm() {
    dispatch('confirm');
    handleClose();
  }

  function handleOverlayClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      handleClose();
    }
  }
</script>

<div 
  class="overlay" 
  on:click={handleOverlayClick}
  on:keydown={(e) => e.key === 'Escape' && handleClose()}
  role="button"
  tabindex="0"
>
  <div class="alert-dialog">
    <div class="alert-header">
      <h2>{title}</h2>
    </div>
    <div class="alert-content">
      <p>{message}</p>
    </div>
    <div class="alert-actions">
      {#if showConfirm}
        <button class="action-button cancel" on:click={handleClose}>
          {cancelText}
        </button>
        <button class="action-button confirm" on:click={handleConfirm}>
          {confirmText}
        </button>
      {:else}
        <button class="action-button single" on:click={handleClose}>
          {$_('alerts.ok')}
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100dvh;
    z-index: 1000;
  }

  .alert-dialog {
    width: 85%;
    max-width: 350px;
    background: var(--tg-theme-bg-color, #F9F8F3);
    border-radius: 16px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
    overflow: hidden;
  }

  .alert-header {
    padding: 20px 16px 12px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }

  .alert-header h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .alert-content {
    padding: 16px;
    text-align: center;
  }

  .alert-content p {
    margin: 0;
    color: var(--tg-theme-text-color);
    opacity: 0.8;
    font-size: 16px;
    line-height: 1.4;
  }

  .alert-actions {
    padding: 0 16px 16px;
    display: flex;
    gap: 8px;
  }

  .action-button {
    flex: 1;
    padding: 12px;
    border-radius: 12px;
    border: none;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .action-button.single {
    background: var(--tg-theme-button-color, #00D5A0);
    color: var(--tg-theme-button-text-color, white);
  }

  .action-button.confirm {
    background: var(--tg-theme-button-color, #00D5A0);
    color: var(--tg-theme-button-text-color, white);
  }

  .action-button.cancel {
    background: var(--tg-theme-secondary-bg-color);
    color: var(--tg-theme-text-color);
  }

  :global([data-theme="dark"]) .alert-dialog {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .alert-dialog h2 {
    color: white !important;
  }

  :global([data-theme="dark"]) .alert-content p {
    color: white !important;
    opacity: 0.9;
  }

  :global([data-theme="dark"]) .action-button.cancel {
    color: white !important;
  }

  :global([data-theme="dark"]) .action-button.single {
    color: white !important;
  }
</style> 
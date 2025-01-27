<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import { user } from '../stores/user';

  const dispatch = createEventDispatcher();
  const API_URL = import.meta.env.VITE_API_URL;

  let notificationsEnabled = false;
  let notificationTime = "";
  let isSaving = false;
  let saveMessage = "";

  // Загружаем настройки при инициализации
  async function loadSettings() {
    try {
      if (!$user?.id) return;
      
      const response = await fetch(`${API_URL}/user/settings?telegram_id=${$user.id}`);
      if (response.ok) {
        const data = await response.json();
        notificationsEnabled = data.notifications_enabled || false;
        notificationTime = data.notification_time || "";
      }
    } catch (error) {
      console.error('Error loading settings:', error);
    }
  }

  // Сохраняем настройки
  async function saveSettings() {
    try {
      if (!$user?.id) return;
      
      isSaving = true;
      saveMessage = "";

      const response = await fetch(`${API_URL}/user/settings`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          telegram_id: $user.id,
          notifications_enabled: notificationsEnabled,
          notification_time: notificationTime
        })
      });

      if (response.ok) {
        saveMessage = $_('settings.saved');
        setTimeout(() => {
          saveMessage = "";
        }, 3000);
      } else {
        throw new Error($_('settings.error'));
      }
    } catch (error) {
      console.error('Error saving settings:', error);
      saveMessage = $_('settings.error');
    } finally {
      isSaving = false;
    }
  }

  function handleBack() {
    dispatch('back');
  }

  // Загружаем настройки при монтировании компонента
  loadSettings();
</script>

<div class="settings-page">
  <header>
    <button class="back-button" on:click={handleBack}>
      ←
    </button>
    <h1>{$_('settings.title')}</h1>
  </header>

  <div class="settings-content">
    <section class="settings-section">
      <h2>{$_('settings.notifications')}</h2>
      
      <div class="setting-item">
        <label class="switch">
          <input 
            type="checkbox" 
            bind:checked={notificationsEnabled}
            on:change={saveSettings}
          >
          <span class="slider"></span>
        </label>
        <span class="setting-label">{$_('settings.notifications_enabled')}</span>
      </div>

      {#if notificationsEnabled}
        <div class="setting-item">
          <input 
            type="time" 
            class="time-input"
            bind:value={notificationTime}
            on:change={saveSettings}
            placeholder={$_('settings.notification_time_placeholder')}
          >
          <span class="setting-label">{$_('settings.notification_time')}</span>
        </div>
      {/if}

      {#if saveMessage}
        <div class="save-message" class:error={saveMessage === $_('settings.error')}>
          {saveMessage}
        </div>
      {/if}
    </section>
  </div>
</div>

<style>
  .settings-page {
    position: fixed;
    inset: 0;
    background-color: #F9F8F3;
    z-index: 1000;
    display: flex;
    flex-direction: column;
  }

  :global([data-theme="dark"]) .settings-page {
    background-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .settings-content {
    color: white;
  }

  :global([data-theme="dark"]) .setting-label {
    color: white;
  }

  :global([data-theme="dark"]) h1 {
    color: white;
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .back-button {
    color: white;
  }

  header {
    display: flex;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
  }

  .back-button {
    background: none;
    border: none;
    font-size: 24px;
    padding: 8px 16px;
    margin-right: 8px;
    cursor: pointer;
    color: var(--tg-theme-text-color);
  }

  h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .settings-content {
    flex: 1;
    padding: 24px;
    overflow-y: auto;
  }

  .settings-section {
    margin-bottom: 32px;
  }

  .settings-section h2 {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 16px 0;
    color: var(--tg-theme-text-color);
  }

  .setting-item {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
    padding: 12px;
    background: var(--tg-theme-secondary-bg-color, rgba(0, 0, 0, 0.05));
    border-radius: 12px;
  }

  .setting-label {
    margin-left: 12px;
    color: var(--tg-theme-text-color);
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
    background-color: var(--tg-theme-button-color, #00D5A0);
  }

  input:checked + .slider:before {
    transform: translateX(20px);
  }

  .time-input {
    background: var(--tg-theme-bg-color);
    border: 1px solid var(--tg-theme-hint-color);
    padding: 8px 12px;
    border-radius: 8px;
    color: var(--tg-theme-text-color);
    font-size: 16px;
  }

  :global([data-theme="dark"]) .time-input {
    background: var(--tg-theme-secondary-bg-color);
    border-color: var(--tg-theme-hint-color);
    color: var(--tg-theme-text-color);
  }

  :global([data-theme="dark"]) .time-input::-webkit-calendar-picker-indicator {
    filter: invert(1);
  }

  .save-message {
    margin-top: 16px;
    padding: 12px;
    border-radius: 8px;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    text-align: center;
  }

  .save-message.error {
    background: #ff4d4d;
  }
</style> 
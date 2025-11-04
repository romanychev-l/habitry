<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import { user } from '../stores/user';
  import { isListView, displayScore } from '../stores/view';
  import { api } from '../utils/api';
  import { fade } from 'svelte/transition';

  const dispatch = createEventDispatcher();
  const API_URL = import.meta.env.VITE_API_URL;
  const BOT_USERNAME = import.meta.env.VITE_BOT_USERNAME;

  let notificationsEnabled = false;
  let notificationTime = "";
  let isSaving = false;
  let saveMessage = "";
  let saveTimeout: ReturnType<typeof setTimeout>;
  let showReferralInfo = false;

  // Action для отслеживания кликов вне элемента
  function clickOutside(node: HTMLElement, handler: () => void) {
      const handleClick = (event: MouseEvent) => {
          if (node && !node.contains(event.target as Node) && !event.defaultPrevented) {
              handler();
          }
      };

      document.addEventListener('click', handleClick, true);

      return {
          destroy() {
              document.removeEventListener('click', handleClick, true);
          }
      };
  }
  
  // Загружаем настройки при инициализации
  async function loadSettings() {
    try {
      if (!$user?.id) return;
      
      const data = await api.getUserSettings();
      notificationsEnabled = data.notifications_enabled || false;
      notificationTime = data.notification_time || "";
    } catch (error) {
      console.error('Error loading settings:', error);
    }
  }

  // Сохраняем настройки с debounce
  async function saveNotificationSettings() {
    try {
      if (!$user?.id) return;

      // Очищаем предыдущий таймер
      if (saveTimeout) {
        clearTimeout(saveTimeout);
      }

      // Устанавливаем новый таймер
      saveTimeout = setTimeout(async () => {
        isSaving = true;
        saveMessage = "";

        console.log('Сохраняем настройки уведомлений:', {
          telegram_id: $user.id,
          notifications_enabled: notificationsEnabled,
          notification_time: notificationTime,
        });

        await api.updateUserSettings({
          notifications_enabled: notificationsEnabled,
          notification_time: notificationTime
        });

        saveMessage = $_('settings.saved');
        setTimeout(() => {
          saveMessage = "";
        }, 3000);
        
        isSaving = false;
      }, 500); // Задержка 500мс
    } catch (error) {
      console.error('Error saving settings:', error);
      saveMessage = $_('settings.error');
      isSaving = false;
    }
  }

  function handleShareProfile() {
    if ($user?.username) {
      const baseUrl = `https://t.me/${BOT_USERNAME}/app`;
      const startAppParam = `startapp=profile_${$user.username}`;
      const appUrl = `${baseUrl}?${startAppParam}`;
      const shareText = $_('settings.share_profile_text');
      
      const url = `https://t.me/share/url?url=${encodeURIComponent(appUrl)}&text=${encodeURIComponent(shareText)}`;
      window.open(url, '_blank');
    }
  }

  function handleShareReferral() {
    if ($user?.username) {
      const baseUrl = `https://t.me/${BOT_USERNAME}/app`;
      const startAppParam = `startapp=ref_${$user.username}`;
      const appUrl = `${baseUrl}?${startAppParam}`;
      const shareText = $_('settings.share_referral_text');
      
      const url = `https://t.me/share/url?url=${encodeURIComponent(appUrl)}&text=${encodeURIComponent(shareText)}`;
      window.open(url, '_blank');
    }
  }

  function handleJoinChannel() {
    window.open('https://t.me/habitry_channel', '_blank');
  }

  function handleJoinCommunity() {
    window.open('https://t.me/+xxqVvxQiFyYzNjgy', '_blank');
  }

  function handleStartOnboarding() {
    dispatch('startOnboarding');
  }

  // Загружаем настройки при монтировании компонента
  loadSettings();
</script>

<div class="settings-page">
  <div class="settings-content">
    <section class="settings-section">
      <h2>{$_('settings.preferences')}</h2>
      
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.notifications_enabled')}</span>
          </div>
          <label class="switch">
            <input type="checkbox" bind:checked={notificationsEnabled} on:change={saveNotificationSettings}>
            <span class="slider"></span>
          </label>
        </div>

        {#if notificationsEnabled}
          <div class="setting-item">
            <div class="setting-label">
              <span>{$_('settings.notification_time')}</span>
            </div>
            <input 
              type="time" 
              bind:value={notificationTime}
              class="time-input"
              on:change={saveNotificationSettings}
              placeholder={$_('settings.notification_time_placeholder')}
            />
          </div>
        {/if}
      </div>

      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('habits.compact_view')}</span>
          </div>
          <label class="switch">
            <input type="checkbox" bind:checked={$isListView}>
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.display_score')}</span>
          </div>
          <label class="switch">
            <input type="checkbox" bind:checked={$displayScore}>
            <span class="slider"></span>
          </label>
        </div>
      </div>
    </section>

    <section class="settings-section">
      <div class="section-header">
        <h2>{$_('settings.referral_program')}</h2>
        <div class="info-container">
          <button class="info-button" on:click|stopPropagation={() => showReferralInfo = !showReferralInfo}>?</button>
          {#if showReferralInfo}
            <div class="tooltip" use:clickOutside={() => showReferralInfo = false} transition:fade>
              {$_('settings.referral_info_text')}
            </div>
          {/if}
        </div>
      </div>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.share_profile_and_referral')}</span>
          </div>
          <button 
            class="share-button" 
            on:click={handleShareProfile}
            disabled={!$user?.username}
          >
            {$_('settings.share')}
          </button>
        </div>
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.share_referral_only')}</span>
          </div>
          <button 
            class="share-button" 
            on:click={handleShareReferral}
            disabled={!$user?.username}
          >
            {$_('settings.share')}
          </button>
        </div>
      </div>
      {#if !$user?.username}
        <p class="warning">{$_('settings.username_required')}</p>
      {/if}
    </section>

    <section class="settings-section">
      <h2>{$_('settings.community')}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.telegram_channel')}</span>
            <small>t.me/habitry_channel</small>
          </div>
          <button 
            class="join-button" 
            on:click={handleJoinChannel}
          >
            {$_('settings.join')}
          </button>
        </div>
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.community_chat')}</span>
            <small>t.me/+xxqVvxQiFyYzNjgy</small>
          </div>
          <button 
            class="join-button" 
            on:click={handleJoinCommunity}
          >
            {$_('settings.join')}
          </button>
        </div>
      </div>
    </section>

    <section class="settings-section">
      <h2>{$_('settings.other')}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-label">
            <span>{$_('settings.restart_onboarding')}</span>
          </div>
          <button 
            class="action-button" 
            on:click={handleStartOnboarding}
          >
            {$_('settings.start')}
          </button>
        </div>
      </div>
    </section>

    {#if saveMessage}
      <div class="save-message" class:error={saveMessage === $_('settings.error')}>
        {saveMessage}
      </div>
    {/if}
  </div>
</div>

<style>
  .settings-page {
    position: fixed;
    inset: 0;
    background-color: var(--tg-theme-bg-color);
    z-index: 1000;
    display: flex;
    flex-direction: column;
  }

  .settings-content {
    flex: 1;
    padding: 24px;
    overflow-y: auto;
    color: var(--tg-theme-text-color);
  }

  .setting-label {
    color: var(--tg-theme-text-color);
    font-size: 14px;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 16px 0;
    color: var(--tg-theme-text-color);
  }

  .settings-section {
    margin-bottom: 32px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .section-header h2 {
    margin: 0;
  }

  .info-container {
    position: relative;
    display: flex;
    align-items: center;
  }

  .info-button {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--tg-theme-hint-color);
    color: var(--tg-theme-bg-color);
    font-size: 12px;
    font-weight: bold;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .tooltip {
    position: absolute;
    right: 28px;
    top: 50%;
    transform: translateY(-50%);
    width: 250px;
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 8px;
    padding: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 1001;
    font-size: 14px;
    color: var(--tg-theme-text-color);
    line-height: 1.4;
    border: 1px solid var(--tg-theme-bg-color);
  }

  .settings-group {
    margin-bottom: 16px;
  }

  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    padding: 12px;
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
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
    background-color: var(--tg-theme-button-color);
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
    width: 70px;
  }

  .time-input::-webkit-calendar-picker-indicator {
    filter: invert(0.5);
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

  .share-button {
    background-color: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
  }

  .share-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .warning {
    color: #e53935;
    font-size: 12px;
    margin-top: 4px;
  }

  .join-button {
    background-color: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
  }

  .action-button {
    background-color: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
  }

  .setting-label small {
    display: block;
    color: var(--tg-theme-hint-color);
    font-size: 12px;
    margin-top: 2px;
  }
</style> 
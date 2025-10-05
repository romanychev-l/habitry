<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import type { Habit } from '../types';
  import { initData } from '@telegram-apps/sdk-svelte';
  import { shareStory } from '@telegram-apps/sdk';
  import { generateHabitStoryImage, uploadBase64ToImgbb, getGradientForHabitId } from '../utils/storyShare';

  const dispatch = createEventDispatcher();
  const BOT_USERNAME = import.meta.env.VITE_BOT_USERNAME;
  export let habit: Habit;

  function handleShare() {
    const userId = initData.user()?.id || '';
    const baseUrl = `https://t.me/${BOT_USERNAME}/app`;
    const startAppParam = `startapp=habit_${habit._id}_${userId}`;
    const appUrl = `${baseUrl}?${startAppParam}`;
    const shareText = `\n${$_('habits.join_habit.0')} ${habit.title} ${$_('habits.join_habit.1')} ${habit.streak} ${$_('habits.join_habit.2')}`;
    
    const url = `https://t.me/share/url?url=${encodeURIComponent(appUrl)}&text=${encodeURIComponent(shareText)}`;
    window.open(url, '_blank');
    dispatch('close');
  }

  async function handleShareStory() {
    const userId = initData.user()?.id || '';
    const baseUrl = `https://t.me/${BOT_USERNAME}/app`;
    const startAppParam = `startapp=habit_${habit._id}_${userId}`;
    const appUrl = `${baseUrl}?${startAppParam}`;
    const botHandle = `@${BOT_USERNAME}`;
    const statsLine = `${$_('habits.join_habit.0')} ${$_('habits.join_habit.1')} ${habit.streak} ${$_('habits.join_habit.2')}`;
    const text = `🔥 ${habit.title}\n📅 ${statsLine}\n🤖 ${botHandle}\n${appUrl}`;

    try {
      const dataUrl = await generateHabitStoryImage({
        habitTitle: habit.title,
        streakDays: habit.streak || 0,
        wantToBecomeText: habit.want_to_become,
        wantLabelText: $_('habits.want_to_become'),
        gradientCss: getGradientForHabitId(habit._id),
      });

      const IMGBB_KEY = import.meta.env.VITE_IMGBB_API_KEY;
      if (!IMGBB_KEY) {
        throw new Error('VITE_IMGBB_API_KEY is not set');
      }
      const uploaded = await uploadBase64ToImgbb(dataUrl, IMGBB_KEY, {
        name: `habit_${habit._id}_${Date.now()}`,
        expirationSeconds: 60 * 60,
      });

      if (shareStory.isAvailable()) {
        shareStory(uploaded.url, {
          // caption убираем, только widgetLink
          widgetLink: {
            url: appUrl,
            name: botHandle
          }
        });
      } 
    } catch (e) {
      console.error('Error sharing story', e);
    } finally {
      dispatch('close');
    }
  }
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
      <h2>{habit.title}</h2>
    </div>
    <div class="dialog-content">
      <button 
        class="dialog-button share"
        on:click={handleShare}
      >
        {$_('habits.share_in_chat')}
      </button>
      <button 
        class="dialog-button share"
        on:click={handleShareStory}
      >
        {$_('habits.share_in_story')}
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
        class="dialog-button archive"
        on:click={() => {
          dispatch('close');
          dispatch('showArchiveConfirm');
        }}
      >
        {$_('habits.archive') || 'Архивировать'}
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
    background: var(--tg-theme-button-color);;
    color: var(--tg-theme-button-text-color);
    margin-bottom: 12px;
  }

  .dialog-button.share {
    background: var(--tg-theme-button-color);;
    color: var(--tg-theme-button-text-color);
    margin-bottom: 12px;
  }

  .dialog-button.delete {
    background: #ff3b30;
    color: white;
  }

  .dialog-button.archive {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    margin-bottom: 12px;
  }

  :global([data-theme="dark"]) .dialog {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .dialog * {
    color: white !important;
  }
</style> 
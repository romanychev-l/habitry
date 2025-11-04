<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';
  import HabitCard from '../components/habits/HabitCard.svelte';
  import type { Habit } from '../types';
  import { api } from '../utils/api';
  import { initData } from '@tma.js/sdk-svelte';
  
  const dispatch = createEventDispatcher();
  export let username: string;
  
  interface UserProfile {
    telegram_id: string;
    username: string;
    first_name: string;
    photo_url: string;
    habits: Habit[];
  }
  
  let userProfile: UserProfile | null = null;
  let error: string | null = null;
  
  async function loadUserProfile() {
    try {
      const data = await api.getUserProfile(username);
      console.log('Received data:', data);
      userProfile = {
        ...data,
        habits: data.habits || []
      };
      console.log('Transformed userProfile:', userProfile);
    } catch (err: unknown) {
      error = err instanceof Error ? err.message : 'Unknown error';
    }
  }
  
  $: if (username) {
    loadUserProfile();
  }
</script>

<div class="profile-page">
  {#if error}
    <div class="error">{error}</div>
  {:else if userProfile}
    <div class="profile-content">
      <div class="profile-header">
        {#if userProfile.photo_url}
          <img src={userProfile.photo_url} alt="Profile" class="profile-photo" />
        {:else}
          <div class="profile-placeholder">
            {userProfile.first_name?.[0] || userProfile.username?.[0] || '?'}
          </div>
        {/if}
        <h2>{userProfile.first_name || userProfile.username}</h2>
      </div>
      
      <div class="habits-list">
        <h3>{$_('profile.habits')}</h3>
        {#if userProfile.habits?.length > 0}
          <div class="habit-container">
            {#each userProfile.habits as habit}
              <HabitCard 
                habit={habit} 
                telegramId={parseInt(userProfile.telegram_id)} 
                readonly={userProfile.telegram_id !== initData.user()?.id?.toString()}
              />
            {/each}
          </div>
        {:else}
          <p class="no-habits">{$_('profile.no_habits')}</p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="loading">{$_('profile.loading')}</div>
  {/if}
</div>

<style>
  .profile-page {
    position: fixed;
    inset: 0;
    background-color: var(--tg-theme-bg-color);
    z-index: 1000;
    display: flex;
    flex-direction: column;
  }
  
  .profile-content {
    flex: 1;
    padding: 24px 0;
    overflow-y: auto;
    /* -webkit-overflow-scrolling: touch; */ /* Temporarily disabled to test iOS rendering */
  }
  
  .profile-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 2rem;
    padding: 0 24px;
  }
  
  .profile-photo {
    width: 64px;
    height: 64px;
    object-fit: cover;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }
  
  .profile-placeholder {
    width: 64px;
    height: 64px;
    background: var(--tg-theme-hint-color);
    color: var(--tg-theme-button-text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    font-weight: 500;
    text-transform: uppercase;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }
  
  h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }
  
  h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }
  
  .habits-list {
    margin-top: 2rem;
  }
  
  .habits-list h3 {
    padding: 0 24px;
  }
  
  .habit-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
    box-sizing: border-box;
    padding: 0 24px;
  }
  
  .habit-container :global(.habit-card) {
    width: 100%;
    max-width: 280px;
    aspect-ratio: 1;
    margin: 0 auto;
    box-sizing: border-box;
    background: white;
  }
  
  @media (max-width: 600px) {
    .habit-container {
      padding: 0 16px;
    }
    
    .habits-list h3 {
      padding: 0 16px;
    }
  }
  
  .no-habits {
    text-align: center;
    color: var(--tg-theme-hint-color);
    padding: 2rem;
  }
  
  .error {
    color: #ff3b30;
    padding: 1rem;
    text-align: center;
  }
  
  .loading {
    text-align: center;
    padding: 2rem;
    color: var(--tg-theme-hint-color);
  }

  :global([data-theme="dark"]) .profile-page {
    background-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .error {
    color: #ff453a;
  }

  :global([data-theme="dark"]) h2,
  :global([data-theme="dark"]) h3 {
    color: white;
  }

  :global([data-theme="dark"]) .habit-container :global(.habit-card) {
    background: white;
  }

  :global([data-theme="dark"]) .habit-container :global(.habit-card h3) {
    color: #333;
  }

  :global([data-theme="dark"]) .habit-container :global(.habit-card .want-to-become) {
    color: #333;
  }

  :global([data-theme="dark"]) .habit-container :global(.habit-card .more-button),
  :global([data-theme="dark"]) .habit-container :global(.habit-card .more-list-view-button),
  :global([data-theme="dark"]) .habit-container :global(.habit-card .undo-button) {
    color: #333;
  }
</style> 
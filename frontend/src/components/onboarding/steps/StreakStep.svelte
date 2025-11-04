<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { scale } from 'svelte/transition';
  import { elasticOut } from 'svelte/easing';
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  export let streakDays = 0;
  export let isSharedHabit = false;
  export let friendsAdded = 0;
  
  function incrementStreak() {
    if (streakDays < 7) {
      streakDays++;
      if (streakDays === 7) {
        dispatch('complete');
      }
      dispatch('update', { streakDays });
    }
  }
  
  function addFriend() {
    if (friendsAdded < 2) {
      friendsAdded++;
      if (friendsAdded === 2) {
        dispatch('complete');
      }
      dispatch('update', { friendsAdded });
    }
  }
</script>

{#if !isSharedHabit}
  <!-- Интерактив со стриком -->
  <div class="illustration-container">
    <div class="glow-effect streak-glow"></div>
    <div class="streak-display">
      <div class="flame-icon">🔥</div>
      <div class="streak-number">{streakDays}</div>
    </div>
  </div>
  <h2>{$_('onboarding.interaction.title')}</h2>
  <p class="description">{$_('onboarding.interaction.description')}</p>
  
  <div class="days-grid">
    {#each Array(7) as _, i}
      <button 
        class="day-cell"
        class:active={i < streakDays}
        on:click={incrementStreak}
        disabled={i < streakDays}
      >
        {#if i < streakDays}
          <span class="checkmark" in:scale={{ duration: 300, easing: elasticOut }}>✓</span>
        {:else}
          <span class="day-number">{i + 1}</span>
        {/if}
      </button>
    {/each}
  </div>

  {#if streakDays >= 7}
    <div class="success-badge" in:scale={{ duration: 400, easing: elasticOut }}>
      <span class="badge-icon">🏆</span>
      <span class="badge-text">{$_('onboarding.interaction.streak_achieved')}</span>
    </div>
  {/if}
{:else}
  <!-- Для shared habit: социальный аспект -->
  <div class="illustration-container">
    <div class="glow-effect social-glow"></div>
    <div class="friends-demo">
      <div class="friend-avatar" class:added={friendsAdded > 0}>👤</div>
      <div class="friend-avatar" class:added={friendsAdded > 1}>👤</div>
    </div>
  </div>
  <h2>{$_('onboarding.interaction.title')}</h2>
  <p class="description">{$_('onboarding.interaction.description')}</p>
  
  <button 
    class="add-friend-btn" 
    on:click={addFriend}
    disabled={friendsAdded >= 2}
  >
    {friendsAdded >= 2 ? '✨ Команда собрана!' : `➕ Добавить друга (${friendsAdded}/2)`}
  </button>
{/if}

<style>
  .illustration-container {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 8px;
    min-height: 120px;
  }

  .glow-effect {
    position: absolute;
    width: 120px;
    height: 120px;
    border-radius: 50%;
    filter: blur(20px);
    animation: glow 2s ease-in-out infinite;
  }

  .glow-effect.streak-glow {
    background: radial-gradient(circle, rgba(249, 115, 22, 0.4) 0%, transparent 70%);
  }

  .glow-effect.social-glow {
    background: radial-gradient(circle, rgba(34, 197, 94, 0.4) 0%, transparent 70%);
  }

  @keyframes glow {
    0%, 100% { transform: scale(1); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 0.8; }
  }

  h2 {
    margin: 0;
    font-size: 28px;
    font-weight: 700;
    color: var(--tg-theme-text-color);
    letter-spacing: -0.5px;
  }

  .description {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
    max-width: 300px;
  }

  .streak-display {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    position: relative;
    z-index: 1;
  }

  .flame-icon {
    font-size: 64px;
    animation: flicker 1.5s ease-in-out infinite;
  }

  @keyframes flicker {
    0%, 100% { transform: scale(1); filter: brightness(1); }
    50% { transform: scale(1.1); filter: brightness(1.2); }
  }

  .streak-number {
    font-size: 48px;
    font-weight: 800;
    background: linear-gradient(135deg, #f97316 0%, #dc2626 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
    margin-top: 12px;
  }

  .day-cell {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 2px solid rgba(0, 0, 0, 0.1);
    background: var(--tg-theme-secondary-bg-color, #fff);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .day-cell:not(:disabled):active {
    transform: scale(0.9);
  }

  .day-cell.active {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
    border-color: #10b981;
    color: white;
  }

  .checkmark {
    font-size: 20px;
  }

  .day-number {
    color: var(--tg-theme-hint-color);
  }

  .success-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.2) 0%, rgba(59, 130, 246, 0.2) 100%);
    border-radius: 16px;
    border: 2px solid rgba(34, 197, 94, 0.3);
    margin-top: 8px;
  }

  .badge-icon {
    font-size: 24px;
  }

  .badge-text {
    font-size: 15px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  /* Friends demo */
  .friends-demo {
    display: flex;
    gap: 16px;
    position: relative;
    z-index: 1;
  }

  .friend-avatar {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: var(--tg-theme-secondary-bg-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 40px;
    opacity: 0.3;
    transition: all 0.3s ease;
    border: 3px solid rgba(0, 0, 0, 0.1);
  }

  .friend-avatar.added {
    opacity: 1;
    border-color: rgba(34, 197, 94, 0.5);
    animation: friend-added 0.5s ease;
  }

  @keyframes friend-added {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.15); }
  }

  .add-friend-btn {
    padding: 14px 24px;
    border-radius: 14px;
    border: none;
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-top: 12px;
  }

  .add-friend-btn:not(:disabled):active {
    transform: scale(0.95);
  }

  .add-friend-btn:disabled {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.7);
  }

  :global([data-theme="dark"]) .day-cell {
    background: #2a2a2a;
    border-color: rgba(255, 255, 255, 0.1);
  }
</style>


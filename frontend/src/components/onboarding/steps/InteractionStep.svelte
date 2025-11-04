<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { scale } from 'svelte/transition';
  import { elasticOut } from 'svelte/easing';
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  export let habitCompleted = false;
  export let holdProgress = 0;
  export let isHolding = false;
  
  let holdInterval: number | null = null;
  
  function startHold() {
    if (habitCompleted) return;
    isHolding = true;
    holdProgress = 0;
    
    holdInterval = window.setInterval(() => {
      holdProgress += 2;
      if (holdProgress >= 100) {
        completeHabit();
      }
    }, 20);
    
    dispatch('holding', { isHolding: true, holdProgress });
  }
  
  function stopHold() {
    isHolding = false;
    if (holdInterval !== null) {
      clearInterval(holdInterval);
      holdInterval = null;
    }
    if (!habitCompleted) {
      holdProgress = 0;
    }
    dispatch('holding', { isHolding: false, holdProgress: habitCompleted ? 100 : 0 });
  }
  
  function completeHabit() {
    stopHold();
    habitCompleted = true;
    holdProgress = 100;
    dispatch('complete');
  }
</script>

<div class="illustration-container">
  <div class="glow-effect pulse"></div>
  <div class="illustration">💪</div>
</div>
<h2>{$_('onboarding.social.title')}</h2>
<p class="description">{$_('onboarding.social.description')}</p>

<button 
  class="habit-demo" 
  class:completed={habitCompleted}
  class:holding={isHolding}
  on:mousedown={startHold}
  on:mouseup={stopHold}
  on:mouseleave={stopHold}
  on:touchstart={startHold}
  on:touchend={stopHold}
  on:touchcancel={stopHold}
  disabled={habitCompleted}
>
  <div class="habit-demo-content">
    <div class="habit-icon">{habitCompleted ? '✨' : '💪'}</div>
    <div class="habit-text">
      {habitCompleted ? $_('onboarding.social.great_job') : $_('onboarding.social.hold_to_complete')}
    </div>
  </div>
  {#if holdProgress > 0 && !habitCompleted}
    <div class="progress-ring" style="--progress: {holdProgress}"></div>
  {/if}
</button>

{#if habitCompleted}
  <div class="success-badge" in:scale={{ duration: 400, easing: elasticOut }}>
    <span class="badge-icon">🎉</span>
    <span class="badge-text">{$_('onboarding.social.success_message')}</span>
  </div>
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

  .glow-effect.pulse {
    background: radial-gradient(circle, rgba(59, 130, 246, 0.4) 0%, transparent 70%);
  }

  @keyframes glow {
    0%, 100% { transform: scale(1); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 0.8; }
  }

  .illustration {
    font-size: 80px;
    position: relative;
    z-index: 1;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
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

  .habit-demo {
    position: relative;
    width: 100%;
    max-width: 300px;
    height: 160px;
    margin-top: 20px;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(147, 51, 234, 0.15) 100%);
    border: none;
    cursor: pointer;
    transition: all 0.3s ease;
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.1));
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
  }

  .habit-demo:not(:disabled):active {
    transform: scale(0.97);
  }

  .habit-demo.holding {
    transform: scale(0.97);
    filter: drop-shadow(0 4px 12px rgba(59, 130, 246, 0.3));
  }

  .habit-demo.completed {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.25) 0%, rgba(59, 130, 246, 0.25) 100%);
    animation: success-pulse 0.6s ease;
    filter: drop-shadow(0 4px 16px rgba(34, 197, 94, 0.4));
  }

  @keyframes success-pulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.05); }
  }

  .habit-demo-content {
    position: absolute;
    inset: 0;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 24px 32px;
  }

  .habit-icon {
    font-size: 48px;
    line-height: 1;
  }

  .habit-text {
    font-size: 15px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
    text-align: center;
    line-height: 1.3;
  }

  .progress-ring {
    position: absolute;
    inset: -2px;
    background: conic-gradient(
      rgba(59, 130, 246, 0.7) calc(var(--progress) * 1%), 
      transparent calc(var(--progress) * 1%)
    );
    pointer-events: none;
    transition: background 0.05s linear;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    filter: blur(1px);
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

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.7);
  }
</style>


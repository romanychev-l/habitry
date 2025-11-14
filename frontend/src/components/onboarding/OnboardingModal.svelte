<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { scale, fly } from 'svelte/transition';
  import { quintOut, elasticOut } from 'svelte/easing';
  
  import ConfettiEffect from './shared/ConfettiEffect.svelte';
  import StarsBackground from './shared/StarsBackground.svelte';
  import ProgressBar from './shared/ProgressBar.svelte';
  import WelcomeStep from './steps/WelcomeStep.svelte';
  import InteractionStep from './steps/InteractionStep.svelte';
  import StreakStep from './steps/StreakStep.svelte';
  import StatsStep from './steps/StatsStep.svelte';
  import CommunityStep from './steps/CommunityStep.svelte';
  import FinalStep from './steps/FinalStep.svelte';

  const dispatch = createEventDispatcher();
  
  export let isSharedHabit = false;
  
  let currentStep = 0;
  $: totalSteps = isSharedHabit ? 3 : 6;
  let showConfetti = false;
  
  // Состояние для интерактивных шагов
  let habitCompleted = false;
  let holdProgress = 0;
  let isHolding = false;
  let streakDays = 0;
  let friendsAdded = 0;
  let mounted = false;
  let direction = 1; // 1 для движения вперёд, -1 для движения назад

  // Предотвращаем скролл на основной странице
  function disableBodyScroll() {
    document.body.style.overflow = 'hidden';
    document.body.style.position = 'fixed';
    document.body.style.width = '100%';
  }

  function enableBodyScroll() {
    document.body.style.overflow = '';
    document.body.style.position = '';
    document.body.style.width = '';
  }

  onMount(() => {
    mounted = true;
    disableBodyScroll();
  });

  onDestroy(() => {
    enableBodyScroll();
  });

  function nextStep() {
    if (currentStep < totalSteps - 1) {
      // Проверяем, требуется ли интерактив на текущем шаге
      if (currentStep === 1 && !habitCompleted) return;
      if (currentStep === 2 && !isSharedHabit && streakDays < 7) return;
      if (currentStep === 2 && isSharedHabit && friendsAdded < 2) return;
      
      direction = 1; // Движение вперёд
      
      // Сброс состояний перед переходом на следующий шаг
      habitCompleted = false;
      holdProgress = 0;
      streakDays = 0;
      friendsAdded = 0;
      currentStep++;
    } else {
      triggerConfetti();
      setTimeout(() => dispatch('finish'), 1500);
    }
  }

  function skipOnboarding() {
    dispatch('skip');
  }

  function previousStep() {
    if (currentStep > 0) {
      direction = -1; // Движение назад
      
      // Сброс состояний перед возвратом
      habitCompleted = false;
      holdProgress = 0;
      streakDays = 0;
      friendsAdded = 0;
      currentStep--;
    }
  }

  function triggerConfetti() {
    showConfetti = true;
    setTimeout(() => showConfetti = false, 3000);
  }

  // Обработчики для InteractionStep
  function handleInteractionComplete() {
    habitCompleted = true;
  }

  function handleHolding(event: CustomEvent) {
    isHolding = event.detail.isHolding;
    holdProgress = event.detail.holdProgress;
  }

  // Обработчики для StreakStep
  function handleStreakComplete() {
    // Завершение интерактива
  }

  function handleStreakUpdate(event: CustomEvent) {
    if (event.detail.streakDays !== undefined) {
      streakDays = event.detail.streakDays;
    }
    if (event.detail.friendsAdded !== undefined) {
      friendsAdded = event.detail.friendsAdded;
    }
  }

  $: canProceed = 
    currentStep === 0 || 
    currentStep === 3 ||
    currentStep === 4 ||
    currentStep === 5 ||
    (currentStep === 1 && habitCompleted) ||
    (currentStep === 2 && !isSharedHabit && streakDays >= 7) ||
    (currentStep === 2 && isSharedHabit && friendsAdded >= 2);
</script>

<div class="fullscreen-container">
  <StarsBackground />
  <ProgressBar {totalSteps} {currentStep} />
  
  <button class="close-btn" on:click={skipOnboarding}>
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
  </button>

  <div class="content">
    {#key currentStep}
      <div class="step"
        in:fly={{ x: direction * 300, duration: 400, easing: quintOut }}
        out:fly={{ x: -direction * 300, duration: 400, easing: quintOut }}
      >
        {#if currentStep === 0}
          <WelcomeStep />
        {:else if currentStep === 1}
          <InteractionStep 
            bind:habitCompleted
            bind:holdProgress
            bind:isHolding
            on:complete={handleInteractionComplete}
            on:holding={handleHolding}
          />
        {:else if currentStep === 2}
          <StreakStep 
            bind:streakDays
            bind:friendsAdded
            {isSharedHabit}
            on:complete={handleStreakComplete}
            on:update={handleStreakUpdate}
          />
        {:else if currentStep === 3 && !isSharedHabit}
          <StatsStep />
        {:else if currentStep === 4 && !isSharedHabit}
          <CommunityStep />
        {:else if (currentStep === 5 && !isSharedHabit) || (currentStep === 3 && isSharedHabit)}
          <FinalStep />
        {/if}
      </div>
    {/key}
  </div>
  
  <div class="footer">
    {#if currentStep > 0}
      <button class="back-btn" on:click={previousStep}>
        ← {$_('onboarding.back')}
      </button>
    {/if}
    <button 
      class="next-btn" 
      class:pulse-btn={canProceed}
      class:final-btn={currentStep === totalSteps - 1}
      on:click={nextStep}
      disabled={!canProceed}
    >
      <span class="btn-text">
        {currentStep === totalSteps - 1 ? $_('onboarding.start.action') : $_('onboarding.next')}
      </span>
      {#if currentStep === totalSteps - 1}
        <span class="btn-icon">🚀</span>
      {/if}
    </button>
  </div>
  
  <ConfettiEffect show={showConfetti} />
</div>

<style>
  .fullscreen-container {
    position: fixed;
    inset: 0;
    display: flex;
    flex-direction: column;
    height: 100dvh;
    z-index: 1000;
    background: var(--tg-theme-bg-color, #F9F8F3);
    overflow: hidden;
  }

  .close-btn {
    position: absolute;
    top: 12px;
    right: 12px;
    z-index: 1001;
    background: rgba(0,0,0,0.1);
    color: var(--tg-theme-text-color, #000);
    border: none;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0;
  }
  
  :global([data-theme="dark"]) .close-btn {
    background: rgba(255,255,255,0.1);
    color: #fff;
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    overflow: hidden;
    position: relative;
    padding: 60px 24px 24px;
  }

  .step {
    position: absolute;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 20px;
    width: calc(100% - 48px);
    height: 100%;
  }
  
  .footer {
    padding: 16px 24px 20px;
    display: flex;
    gap: 12px;
    width: 100%;
    box-sizing: border-box;
    z-index: 10;
  }

  .back-btn {
    flex: 1;
    padding: 16px;
    border-radius: 14px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 15px;
    font-weight: 600;
    transition: all 0.2s ease;
    cursor: pointer;
  }

  .back-btn:active {
    transform: scale(0.97);
  }

  .next-btn {
    flex: 2;
    padding: 16px;
    border-radius: 14px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    position: relative;
    overflow: hidden;
  }
  
  .next-btn:not(:disabled):active {
    transform: scale(0.97);
  }

  .next-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .pulse-btn {
    animation: button-pulse 1.5s ease-in-out infinite;
  }

  @keyframes button-pulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(147, 51, 234, 0.7); }
    50% { box-shadow: 0 0 0 8px rgba(147, 51, 234, 0); }
  }

  .final-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    animation: rainbow-shift 3s ease infinite;
    background-size: 200% 200%;
  }

  @keyframes rainbow-shift {
    0%, 100% { background-position: 0% 50%; }
    50% { background-position: 100% 50%; }
  }

  .btn-icon {
    font-size: 20px;
  }
  
  :global([data-theme="dark"]) .back-btn {
    color: white;
    border-color: white;
  }

  :global([data-theme="dark"]) .next-btn {
    background: white;
    color: #1a1a1a;
  }

  :global([data-theme="dark"]) .final-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }
</style>


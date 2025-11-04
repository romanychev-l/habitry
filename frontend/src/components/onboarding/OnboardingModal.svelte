<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { fade, scale, fly } from 'svelte/transition';
  import { cubicOut, elasticOut } from 'svelte/easing';
  
  import ConfettiEffect from './shared/ConfettiEffect.svelte';
  import StarsBackground from './shared/StarsBackground.svelte';
  import ProgressDots from './shared/ProgressDots.svelte';
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

  onMount(() => {
    mounted = true;
  });

  function nextStep() {
    if (currentStep < totalSteps - 1) {
      // Проверяем, требуется ли интерактив на текущем шаге
      if (currentStep === 1 && !habitCompleted) return;
      if (currentStep === 2 && !isSharedHabit && streakDays < 7) return;
      if (currentStep === 2 && isSharedHabit && friendsAdded < 2) return;
      
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

<div class="wrapper">
  <div 
    class="overlay" 
    role="button"
    tabindex="-1"
    on:click={skipOnboarding}
    on:keydown={(e) => e.key === 'Escape' && skipOnboarding()}
    transition:fade={{ duration: 300 }}
  ></div>
  
  <ConfettiEffect show={showConfetti} />

  <div class="modal-container" transition:scale={{ duration: 400, start: 0.9, easing: elasticOut }}>
    <div class="modal">
      <StarsBackground />

      <div class="content">
        {#key currentStep}
          <div class="step" 
            in:fly={{ y: 50, duration: 500, delay: 100, easing: cubicOut }}
            out:fly={{ y: -50, duration: 300, easing: cubicOut }}
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

        <ProgressDots {totalSteps} {currentStep} />
      </div>

      <div class="footer">
        <button class="skip-btn" on:click={skipOnboarding}>
          {$_('onboarding.skip')}
        </button>
        <button 
          class="next-btn" 
          class:pulse-btn={canProceed && currentStep < totalSteps - 1}
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
    </div>
  </div>
</div>

<style>
  .wrapper {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100dvh;
    z-index: 1000;
  }

  .overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
  }

  .modal-container {
    position: relative;
    width: 100%;
    max-width: 360px;
    margin: 0 16px;
    z-index: 1;
  }

  .modal {
    position: relative;
    width: 100%;
    background: var(--tg-theme-bg-color, #F9F8F3);
    border-radius: 28px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    overflow: hidden;
  }

  .content {
    position: relative;
    padding: 40px 24px 32px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    min-height: 480px;
  }

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    width: 100%;
  }

  .footer {
    padding: 16px 24px 20px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
    display: flex;
    gap: 12px;
  }

  .skip-btn {
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

  .skip-btn:active {
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

  /* Dark theme adjustments */
  :global([data-theme="dark"]) .modal {
    background: #1a1a1a;
  }

  :global([data-theme="dark"]) .skip-btn {
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


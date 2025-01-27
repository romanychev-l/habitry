<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';

  const dispatch = createEventDispatcher();
  
  export let isSharedHabit = false;
  
  let currentStep = 0;
  $: totalSteps = isSharedHabit ? 3 : 4;

  function nextStep() {
    if (currentStep < totalSteps - 1) {
      currentStep++;
    } else {
      dispatch('finish');
    }
  }

  function skipOnboarding() {
    dispatch('skip');
  }
</script>

<div class="wrapper">
  <div class="overlay" on:click={skipOnboarding}></div>
  <div class="modal-container">
    <div class="modal">
      <div class="content">
        {#if currentStep === 0}
          <div class="step">
            <div class="illustration">👋</div>
            <h2>{$_('onboarding.welcome.title')}</h2>
            <p>{$_('onboarding.welcome.description')}</p>
          </div>
        {:else if currentStep === 1}
          <div class="step">
            <div class="illustration">🤝</div>
            <h2>{$_('onboarding.social.title')}</h2>
            <p>{$_('onboarding.social.description')}</p>
          </div>
        {:else if currentStep === 2}
          <div class="step">
            <div class="illustration">✨</div>
            <h2>{$_('onboarding.interaction.title')}</h2>
            <p>{$_('onboarding.interaction.description')}</p>
          </div>
        {:else if currentStep === 3 && !isSharedHabit}
          <div class="step">
            <div class="illustration">🎯</div>
            <h2>{$_('onboarding.start.title')}</h2>
            <p>{$_('onboarding.start.description')}</p>
          </div>
        {/if}

        <div class="progress">
          {#each Array(totalSteps) as _, i}
            <div class="dot" class:active={i === currentStep}></div>
          {/each}
        </div>
      </div>

      <div class="footer">
        <button class="skip-btn" on:click={skipOnboarding}>
          {$_('onboarding.skip')}
        </button>
        <button class="next-btn" on:click={nextStep}>
          {currentStep === totalSteps - 1 ? $_('onboarding.start.action') : $_('onboarding.next')}
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
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
  }

  .modal-container {
    position: relative;
    width: 100%;
    max-width: 320px;
    margin: 0 16px;
    z-index: 1;
  }

  .modal {
    width: 100%;
    background: #F9F8F3;
    border-radius: 24px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
    overflow: hidden;
  }

  .content {
    padding: 32px 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .illustration {
    font-size: 64px;
    margin-bottom: 8px;
  }

  h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  p {
    margin: 0;
    font-size: 16px;
    line-height: 1.5;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
  }

  .progress {
    display: flex;
    gap: 8px;
    margin-top: 32px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--tg-theme-secondary-bg-color);
    transition: all 0.2s;
  }

  .dot.active {
    background: var(--tg-theme-button-color);
    transform: scale(1.2);
  }

  .footer {
    padding: 16px 24px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
    display: flex;
    gap: 12px;
  }

  .skip-btn {
    flex: 1;
    padding: 14px;
    border-radius: 12px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 16px;
    font-weight: 500;
  }

  .next-btn {
    flex: 2;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 16px;
    font-weight: 500;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) p {
    color: rgba(255, 255, 255, 0.6);
  }

  :global([data-theme="dark"]) .skip-btn {
    color: white;
    border-color: white;
  }

  :global([data-theme="dark"]) .next-btn {
    background: white;
    color: var(--tg-theme-bg-color);
  }
</style> 
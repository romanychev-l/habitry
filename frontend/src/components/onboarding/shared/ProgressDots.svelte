<script lang="ts">
  export let totalSteps: number;
  export let currentStep: number;
</script>

<div class="progress">
  {#each Array(totalSteps) as _, i}
    <div 
      class="dot" 
      class:active={i === currentStep}
      class:completed={i < currentStep}
    ></div>
  {/each}
</div>

<style>
  .progress {
    display: flex;
    gap: 10px;
    margin-top: auto;
    padding-top: 24px;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--tg-theme-secondary-bg-color);
    transition: all 0.3s ease;
  }

  .dot.completed {
    background: rgba(34, 197, 94, 0.6);
    transform: scale(0.8);
  }

  .dot.active {
    background: var(--tg-theme-button-color);
    transform: scale(1.3);
    box-shadow: 0 0 12px rgba(147, 51, 234, 0.6);
  }
</style>


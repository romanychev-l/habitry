<script lang="ts">
  export let totalSteps: number;
  export let currentStep: number;
</script>

<div class="progress-container">
  {#each Array(totalSteps) as _, i}
    <div class="segment">
      <div class="progress" style:width={i < currentStep ? '100%' : '0%'} />
    </div>
  {/each}
</div>

<style>
  .progress-container {
    position: absolute;
    top: 20px;
    left: 16px;
    right: 16px;
    display: grid;
    grid-auto-columns: 1fr;
    grid-auto-flow: column;
    gap: 4px;
    height: 3px;
    z-index: 10;
  }

  .segment {
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.2);
    border-radius: 2px;
    overflow: hidden;
  }
  
  :global([data-theme="dark"]) .segment {
     background-color: rgba(255, 255, 255, 0.2);
  }

  .progress {
    width: 0%;
    height: 100%;
    background-color: var(--tg-theme-button-color, #3B82F6);
    border-radius: 2px;
    transition: width 0.3s ease-in-out;
  }
  
  :global([data-theme="dark"]) .progress {
    background-color: white;
  }
</style>

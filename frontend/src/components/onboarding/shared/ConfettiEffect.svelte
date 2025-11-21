<script lang="ts">
  import { fade } from 'svelte/transition';
  
  export let show = false;
</script>

{#if show}
  <div class="confetti-container" transition:fade={{ duration: 500 }}>
    {#each Array(50) as _, i}
      <div 
        class="confetti" 
        style="
          left: {Math.random() * 100}%;
          animation-delay: {Math.random() * 0.5}s;
          animation-duration: {2 + Math.random() * 2}s;
          background: hsl({Math.random() * 360}, 70%, 60%);
        "
      ></div>
    {/each}
  </div>
{/if}

<style>
  .confetti-container {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 2;
    overflow: hidden;
  }

  .confetti {
    position: absolute;
    width: 10px;
    height: 10px;
    top: -10px;
    animation: fall linear forwards;
    transform-origin: center;
  }

  @keyframes fall {
    to {
      transform: translateY(100vh) rotate(360deg);
      opacity: 0;
    }
  }
</style>


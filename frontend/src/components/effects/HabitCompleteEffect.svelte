<script lang="ts">
  import { onMount } from 'svelte';
  export let show = false;

  type Particle = {
    x: number;
    y: number;
    scale: number;
    duration: number;
  };

  let particles: Particle[] = [];
  onMount(() => {
    particles = Array(30).fill(0).map(() => ({
      x: (Math.random() - 0.5) * 500,
      y: (Math.random() - 0.5) * 500,
      scale: 0.5 + Math.random() * 1,
      duration: 0.5 + Math.random() * 1,
    }));
  });
</script>

{#if show}
<div class="effect-container">
  {#each particles as p, i}
    <div 
      class="particle" 
      style="
        --x: {p.x}px; 
        --y: {p.y}px; 
        --scale: {p.scale};
        --duration: {p.duration}s;
        --delay: {Math.random() * 0.2}s;
      "
    ></div>
  {/each}
</div>
{/if}

<style>
  .effect-container {
    position: fixed;
    top: 50%;
    left: 50%;
    width: 1px;
    height: 1px;
    z-index: 9999;
    pointer-events: none;
  }

  .particle {
    position: absolute;
    left: 0;
    top: 0;
    width: 8px;
    height: 8px;
    background: var(--tg-theme-button-color, #764ba2);
    border-radius: 50%;
    opacity: 0;
    animation: burst var(--duration) ease-out var(--delay) forwards;
    box-shadow: 0 0 10px var(--tg-theme-button-color), 0 0 20px var(--tg-theme-button-color);
  }

  @keyframes burst {
    0% {
      transform: translate(0, 0) scale(0);
      opacity: 1;
    }
    50% {
      opacity: 1;
    }
    100% {
      transform: translate(var(--x), var(--y)) scale(var(--scale));
      opacity: 0;
    }
  }
</style>

<div class="stars-bg">
  {#each Array(20) as _, i}
    <div 
      class="star" 
      style="
        left: {Math.random() * 100}%;
        top: {Math.random() * 100}%;
        animation-delay: {Math.random() * 3}s;
      "
    ></div>
  {/each}
</div>

<style>
  .stars-bg {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .star {
    position: absolute;
    width: 2px;
    height: 2px;
    background: white;
    border-radius: 50%;
    animation: twinkle 3s infinite;
    opacity: 0;
  }

  @keyframes twinkle {
    0%, 100% { opacity: 0; }
    50% { opacity: 0.8; }
  }
</style>


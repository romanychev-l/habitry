<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { fade, fly } from 'svelte/transition';

  export let show = false;

  const dispatch = createEventDispatcher();
  
  const TON_SPACE_GUIDE = 'https://walletru.helpscoutdocs.com/article/84-chto-takoe-ton-space';
  const P2P_GUIDE = 'https://walletru.helpscoutdocs.com/article/74-znakomstvo-s-r2r-marketom';
  const USDT_GUIDE = 'https://walletru.helpscoutdocs.com/article/60-znakomstvo-s-wallet';

  function openGuide(url: string) {
    window.open(url, '_blank');
  }
</script>

{#if show}
<div class="wrapper">
  <div 
    class="overlay" 
    role="button"
    tabindex="0"
    on:click={() => dispatch('close')}
    on:keydown={e => e.key === 'Enter' && dispatch('close')}
    transition:fade={{ duration: 200 }}
  ></div>
  
  <div class="modal-container" transition:fly={{ y: 500, duration: 300, opacity: 1 }}>
    <div class="modal">
      <div class="header">
        <h2>{$_('payment.instructions_title')}</h2>
        <button 
          class="close-btn" 
          on:click={() => dispatch('close')}
        >
          ✕
        </button>
      </div>

      <div class="content">
        <div class="step">
          <h3>{$_('payment.ton_space_guide')}</h3>
          <p>{$_('payment.ton_space_description')}</p>
          <button class="guide-btn" on:click={() => openGuide(TON_SPACE_GUIDE)}>
            {$_('payment.open_guide')}
          </button>
        </div>

        <div class="step">
          <h3>{$_('payment.p2p_guide')}</h3>
          <p>{$_('payment.p2p_description')}</p>
          <button class="guide-btn" on:click={() => openGuide(P2P_GUIDE)}>
            {$_('payment.open_guide')}
          </button>
        </div>

        <div class="step">
          <h3>{$_('payment.usdt_guide')}</h3>
          <p>{$_('payment.usdt_description')}</p>
          <button class="guide-btn" on:click={() => openGuide(USDT_GUIDE)}>
            {$_('payment.open_guide')}
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
{/if}

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
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    z-index: 1;
  }

  .modal {
    background: var(--tg-theme-bg-color);
    border-radius: 16px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
  }

  .header h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 20px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
    cursor: pointer;
    padding: 4px;
  }

  .close-btn:hover {
    opacity: 1;
  }

  .content {
    padding: 20px;
  }

  .step {
    padding: 16px;
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    margin-bottom: 16px;
  }

  .step:last-child {
    margin-bottom: 0;
  }

  .step h3 {
    margin: 0 0 8px 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .step p {
    margin: 0 0 16px 0;
    font-size: 14px;
    line-height: 1.4;
    color: var(--tg-theme-text-color);
    opacity: 0.8;
  }

  .guide-btn {
    width: 100%;
    padding: 12px;
    border: 2px solid var(--tg-theme-button-color);
    border-radius: 8px;
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .guide-btn:hover {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  :global([data-theme="dark"]) .guide-btn {
    color: #ffffff;
    border-color: #ffffff;
  }

  :global([data-theme="dark"]) .guide-btn:hover {
    background: #ffffff;
    color: var(--tg-theme-bg-color);
  }
</style> 
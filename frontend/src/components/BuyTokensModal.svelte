<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { onMount } from 'svelte';
  // import { TonConnectProvider } from '@/tonconnect';
  import * as TON_CONNECT_UI from '@tonconnect/ui';

  // const manifestUrl = 'https://lenichev.site/tonconnect-manifest.json';
  const manifestUrl = 'https://romanychev-l.github.io/habitry_public/manifest.json';
  // export let url = '';

  const tonConnectUI = new TON_CONNECT_UI.TonConnectUI({
        manifestUrl: manifestUrl,
        buttonRootId: 'ton-connect'
    });

  async function connectToWallet() {
      const connectedWallet = await tonConnectUI.connectWallet();
      // Do something with connectedWallet if needed
      console.log(connectedWallet);
  }

  // Call the function
  connectToWallet().catch(error => {
      console.error("Error connecting to wallet:", error);
  });
  
  
  const dispatch = createEventDispatcher();
  let tokensAmount = 100;
  const EXCHANGE_RATE = 10; // 1 Stars = 10 WILL

  // Добавляем переключатель метода оплаты
  let paymentMethod: 'stars' | 'ton' = 'stars';

  function calculateStars(tokens: number): number {
    return Math.ceil(tokens / EXCHANGE_RATE);
  }

  function handleBuy() {
    dispatch('buy', {
      starsAmount: calculateStars(tokensAmount),
      paymentMethod
    });
  }
</script>

<div class="wrapper">
  <div 
    class="overlay" 
    role="button"
    tabindex="0"
    on:click={() => dispatch('close')}
    on:keydown={e => e.key === 'Enter' && dispatch('close')}
  ></div>
  <div class="modal-container">
    <div class="modal">
      <div class="header">
        <h2>Купить WILL</h2>
      </div>

      <div class="content">
        <div class="payment-method-selector">
          <button 
            class:active={paymentMethod === 'stars'} 
            on:click={() => paymentMethod = 'stars'}
          >
            Telegram Stars
          </button>
          <button 
            class:active={paymentMethod === 'ton'} 
            on:click={() => paymentMethod = 'ton'}
          >
            TON Connect
          </button>
        </div>

        <!-- {#if paymentMethod === 'ton'}
          {#if walletConnected}
            <p>Кошелёк подключён: {address}</p>
          {:else}
            <button on:click={connectWallet}>Подключить кошелёк TON</button>
          {/if}
        {/if} -->
        
        <!-- {#if paymentMethod === 'ton'}
          <TonConnectProvider {manifestUrl}>
            <div class="ton-connect-container">
              <div id="ton-connect"></div>
            </div>
          </TonConnectProvider>
        {/if} -->

        <div id="ton-connect">Connect</div>

        <div class="info-block">
          <div class="exchange-rate">
            <span class="label">Курс обмена</span>
            <span class="value">1 Stars = {EXCHANGE_RATE} WILL</span>
          </div>

          <div class="input-group">
            <label for="tokens-amount">Количество WILL</label>
            <input
              type="number"
              id="tokens-amount"
              bind:value={tokensAmount}
              min="10"
              step="10"
              placeholder="Введите количество WILL"
            />
          </div>

          <div class="summary">
            <span class="label">К оплате</span>
            <span class="value">{calculateStars(tokensAmount)} Stars</span>
          </div>
        </div>
      </div>

      <div class="footer">
        <button 
          class="buy-btn" 
          on:click={handleBuy}
          disabled={tokensAmount < 10}
        >
          Купить
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
    align-items: flex-end;
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
    z-index: 1;
  }

  .modal {
    width: 100%;
    background: #F9F8F3;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    max-height: 90vh;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .header {
    position: sticky;
    top: 0;
    background: inherit;
    z-index: 2;
    padding: 24px 16px 12px 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .content {
    padding: 24px 16px;
  }

  .info-block {
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 16px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .exchange-rate {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .exchange-rate .label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  .exchange-rate .value {
    font-size: 18px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-group label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  input[type="number"] {
    width: 100%;
    padding: 12px;
    border: 2px solid var(--tg-theme-bg-color);
    border-radius: 12px;
    font-size: 16px;
    background: var(--tg-theme-bg-color);
    color: var(--tg-theme-text-color);
    box-sizing: border-box;
  }

  input[type="number"]:focus {
    outline: none;
    border-color: #00D5A0;
  }

  .summary {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 12px;
    border-top: 1px solid var(--tg-theme-bg-color);
  }

  .summary .label {
    font-size: 14px;
    color: var(--tg-theme-text-color);
    opacity: 0.7;
  }

  .summary .value {
    font-size: 24px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .footer {
    position: sticky;
    bottom: 0;
    background: inherit;
    z-index: 2;
    padding: 12px 16px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
  }

  .buy-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: #00D5A0;
    color: white;
    font-size: 16px;
    font-weight: 500;
  }

  .buy-btn:disabled {
    opacity: 0.6;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) input[type="number"] {
    color: white;
  }

  :global([data-theme="dark"]) .info-block {
    background: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) input[type="number"] {
    background: var(--tg-theme-bg-color);
    border-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .summary {
    border-color: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) input::placeholder {
    color: rgba(255, 255, 255, 0.6) !important;
  }

  :global([data-theme="dark"]) h2,
  :global([data-theme="dark"]) .label,
  :global([data-theme="dark"]) .value,
  :global([data-theme="dark"]) label {
    color: white !important;
  }

  .payment-method-selector {
    display: flex;
    gap: 8px;
    margin-bottom: 16px;
  }

  .payment-method-selector button {
    flex: 1;
    padding: 12px;
    border-radius: 12px;
    border: 2px solid var(--tg-theme-secondary-bg-color);
    background: var(--tg-theme-bg-color);
    color: var(--tg-theme-text-color);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .payment-method-selector button.active {
    background: #00D5A0;
    border-color: #00D5A0;
    color: white;
  }

  :global([data-theme="dark"]) .payment-method-selector button {
    background: var(--tg-theme-bg-color);
    border-color: rgba(255, 255, 255, 0.1);
    color: white;
  }

  :global([data-theme="dark"]) .payment-method-selector button.active {
    background: #00D5A0;
    border-color: #00D5A0;
  }

  .ton-connect-container {
    margin-bottom: 16px;
    display: flex;
    justify-content: center;
  }

  :global(#ton-connect) {
    width: 100%;
  }
</style> 
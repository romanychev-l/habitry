<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';
  
  const dispatch = createEventDispatcher();
  let title = '';
  let selectedDays = new Set();
  let isOneTime = false;
  let wantToBecome = '';
  let isDarkTheme = window.Telegram?.WebApp?.colorScheme === 'dark';

  function toggleDay(index: number) {
    if (selectedDays.has(index)) {
      selectedDays.delete(index);
    } else {
      selectedDays.add(index);
    }
    selectedDays = selectedDays;
  }

  function handleSave() {
    if (!title.trim()) {
        alert($_('habits.errors.title_required'));
        return;
    }

    dispatch('save', {
        title,
        want_to_become: wantToBecome,
        days: Array.from(selectedDays),
        is_one_time: isOneTime
    });
  }

  function handleOverlayClick(event: MouseEvent) {
    // Закрываем только если клик был именно по оверлею
    if (event.target === event.currentTarget) {
      dispatch('close');
    }
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
        <h2>{$_('habits.add')}</h2>
      </div>

      <div class="content">
        <input 
          type="text" 
          bind:value={title} 
          placeholder={$_('habits.title')}
          autofocus
          class="input-field"
        />
        
        <input 
          type="text" 
          bind:value={wantToBecome} 
          placeholder={$_('habits.want_to_become')}
          class="input-field"
        />
        
        <!-- <button class="type-selector" on:click={() => isOneTime = !isOneTime}>
          <span>{$_('habits.one_time')}</span>
          <div class="switch">
            <span class="slider" class:checked={isOneTime}></span>
          </div>
        </button> -->
        
        {#if !isOneTime}
          <div class="days-wrapper">
            <div class="days-selector">
              {#each [0, 1, 2, 3, 4, 5, 6] as day}
                <button 
                  class:selected={selectedDays.has(day)}
                  on:click={() => toggleDay(day)}
                >
                  {$_(`days.${['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'][day]}`)}
                </button>
              {/each}
              <button 
                class:selected={selectedDays.size === 7}
                on:click={() => {
                  if (selectedDays.size === 7) {
                    selectedDays = new Set();
                  } else {
                    selectedDays = new Set([0, 1, 2, 3, 4, 5, 6]);
                  }
                }}
              >
                All
              </button>
            </div>
          </div>
        {/if}
      </div>

      <div class="footer">
        <button 
          class="save-btn" 
          on:click={handleSave}
          disabled={!title || (!isOneTime && selectedDays.size === 0)}
        >
          {$_('habits.save')}
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
  }

  @supports (-webkit-touch-callout: none) {
    .wrapper {
      position: absolute;
      height: 100vh;
      min-height: -webkit-fill-available;
    }

    .modal-container:focus-within {
      transform: translateY(-35vh);
      transition: transform 0.3s ease-out;
    }
  }

  .header {
    padding: 32px 16px 16px 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .content {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .type-selector {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    background: none;
    border: none;
    color: var(--tg-theme-text-color);
    padding: 0;
    font-size: 16px;
    margin-top: 8px;
  }

  .daily-habit-btn {
    border-radius: 12px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-weight: 500;
    text-align: left;
    font-size: 14px;
  }

  .daily-habit-btn:active {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  .days-wrapper {
    display: flex;
    gap: 8px;
    align-items: start;
  }

  .days-selector {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
    flex: 1;
  }

  .all-btn {
    aspect-ratio: 1;
    border-radius: 12px;
    border: 2px solid #00D5A0;
    background: var(--tg-theme-secondary-bg-color);
    font-weight: 500;
    font-size: 14px;
    padding: 0;
    width: 48px;
    height: 48px;
  }

  .all-btn.selected {
    border-color: #00D5A0;
    background: #00D5A0;
    color: var(--tg-theme-button-text-color);
  }

  .days-selector button {
    aspect-ratio: 1;
    border-radius: 12px;
    border: 2px solid #00D5A0;
    background: var(--tg-theme-secondary-bg-color);
    font-weight: 500;
    font-size: 14px;
    padding: 0;
    width: 100%;
    height: 48px;
  }

  .days-selector button.selected {
    border-color: #00D5A0;
    background: #00D5A0;
    color: white;
  }

  input[type="text"] {
    border: 2px solid var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    background: transparent;
    font-size: 16px;
    padding: 14px 0px;
    margin: 0;
    width: 100%;
  }

  input[type="text"]:focus {
    outline: none;
    border-color: var(--tg-theme-button-color);
  }

  .switch {
    display: flex;
    align-items: center;
    margin-right: 4px;
  }

  .slider {
    position: relative;
    width: 36px;
    height: 20px;
    background: #ccc;
    border-radius: 10px;
    margin-left: 12px;
  }

  .slider:before {
    content: "";
    position: absolute;
    height: 16px;
    width: 16px;
    left: 2px;
    bottom: 2px;
    background: white;
    border-radius: 50%;
    transition: transform 0.2s;
  }

  /* .slider.checked {
    background: #00D5A0;
  }

  .slider.checked:before {
    transform: translateX(16px);
  } */

  .footer {
    padding: 16px 20px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
  }

  .save-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: #00D5A0;
    color: white;
    font-size: 16px;
    font-weight: 500;
  }

  .save-btn:disabled {
    opacity: 0.6;
  }

  .input-field {
    margin-bottom: 16px;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) input[type="text"] {
    color: white;
  }

  :global([data-theme="dark"]) .type-selector {
    color: white;
  }

  :global([data-theme="dark"]) .days-selector button {
    color: white;
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .modal * {
    color: white !important;
  }

  :global([data-theme="dark"]) input::placeholder {
    color: rgba(255, 255, 255, 0.6) !important;
  }

  :global([data-theme="dark"]) .days-selector button:not(.selected) {
    color: white !important;
  }
</style>
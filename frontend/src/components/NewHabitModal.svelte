<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { _ } from 'svelte-i18n';
  
  const dispatch = createEventDispatcher();
  let title = '';
  let selectedDays = new Set();
  let isOneTime = false;

  function toggleDay(index: number) {
    if (selectedDays.has(index)) {
      selectedDays.delete(index);
    } else {
      selectedDays.add(index);
    }
    selectedDays = selectedDays;
  }

  function handleSubmit() {
    if (isOneTime) {
      const today = new Date();
      let dayIndex = today.getDay() - 1;
      if (dayIndex === -1) dayIndex = 6;
      
      dispatch('save', {
        title,
        days: [dayIndex],
        is_one_time: true
      });
    } else {
      dispatch('save', {
        title,
        days: Array.from(selectedDays),
        is_one_time: false
      });
    }
  }

  function handleOverlayClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      dispatch('close');
    }
  }
</script>

<div 
  class="overlay" 
  on:click={handleOverlayClick}
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')} 
  role="button"
  tabindex="0"
>
  <div class="modal">
    <div class="header">
      <h2>{$_('habits.add')}</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>✕</button>
    </div>

    <div class="content">
      <div class="form-group">
        <input 
          type="text" 
          bind:value={title} 
          placeholder={$_('habits.title')}
        />
      </div>
      
      <div class="form-group">
        <div class="type-selector">
          <span class="label">{$_('habits.one_time')}</span>
          <label class="switch">
            <input 
              type="checkbox" 
              bind:checked={isOneTime}
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>
      
      {#if !isOneTime}
        <div class="form-group">
          <button 
            class="daily-habit-btn" 
            on:click={() => {
              selectedDays = new Set([0, 1, 2, 3, 4, 5, 6]);
            }}
          >
            {$_('habits.every_day')}
          </button>
        </div>

        <div class="form-group">
          <div class="days-selector">
            {#each [0, 1, 2, 3, 4, 5, 6] as day}
              <button 
                class:selected={selectedDays.has(day)}
                on:click={() => toggleDay(day)}
              >
                {$_(`days.${['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'][day]}`)}
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <div class="footer">
      <button 
        class="save-btn" 
        on:click={handleSubmit}
        disabled={!title || (!isOneTime && selectedDays.size === 0)}
      >
        {$_('habits.save')}
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    z-index: 1000;
  }

  .modal {
    position: relative;
    width: 100%;
    background: var(--tg-theme-bg-color);
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    overflow: hidden;
    transform: translateY(0);
    transition: transform 0.3s ease;
  }

  /* iOS-специфичные стили */
  @supports (-webkit-touch-callout: none) {
    .overlay {
        position: absolute;
        height: 100vh;
        min-height: -webkit-fill-available;
    }

    .modal:focus-within {
        transform: translateY(-35vh);
    }
  }

  .header {
    padding: 16px 20px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .content {
    padding: 20px;
  }

  .footer {
    padding: 16px 20px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: var(--tg-theme-text-color);
    opacity: 0.6;
    padding: 8px;
    cursor: pointer;
  }

  .form-group {
    margin-bottom: 24px;
    width: 100%;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  .type-selector {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 0 4px;
  }

  .switch {
    display: flex;
    align-items: center;
  }

  .label {
    font-size: 16px;
    color: var(--tg-theme-text-color);
  }

  .daily-habit-btn {
    width: 100%;
    padding: 12px 16px;
    border-radius: 12px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    text-align: left;
  }

  .daily-habit-btn:active {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  .days-selector {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
  }

  .days-selector button {
    aspect-ratio: 1;
    border-radius: 12px;
    border: 2px solid transparent;
    background: var(--tg-theme-secondary-bg-color);
    cursor: pointer;
    transition: all 0.2s ease;
    font-weight: 500;
    font-size: 14px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .days-selector button.selected {
    border-color: var(--tg-theme-button-color);
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  .save-btn {
    width: 100%;
    padding: 14px 16px;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.2s ease;
    text-align: left;
  }

  .save-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  input[type="text"] {
    width: 100%;
    padding: 12px 16px;
    border: 2px solid var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    font-size: 16px;
    background: transparent;
    color: var(--tg-theme-text-color);
    transition: all 0.2s ease;
    box-sizing: border-box;
  }

  input[type="text"]:focus {
    outline: none;
    border-color: var(--tg-theme-button-color);
  }

  .slider {
    position: relative;
    width: 44px;
    height: 24px;
    background: var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    transition: all 0.2s ease;
    margin-left: 12px;
  }

  .slider:before {
    content: "";
    position: absolute;
    height: 20px;
    width: 20px;
    left: 2px;
    bottom: 2px;
    background: white;
    border-radius: 50%;
    transition: all 0.2s ease;
  }

  input:checked + .slider {
    background: var(--tg-theme-button-color);
  }

  input:checked + .slider:before {
    transform: translateX(20px);
  }
</style>
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
    <h2>{$_('habits.add')}</h2>
    <input 
      type="text" 
      bind:value={title} 
      placeholder={$_('habits.title')}
    />
    
    <div class="type-selector">
      <label>
        <input 
          type="checkbox" 
          bind:checked={isOneTime}
        />
        {$_('habits.one_time')}
      </label>
    </div>
    
    {#if !isOneTime}
      <button 
        class="daily-habit-btn" 
        on:click={() => {
          selectedDays = new Set([0, 1, 2, 3, 4, 5, 6]);
        }}
      >
        {$_('habits.every_day')}
      </button>

      <div class="days-selector">
        <button 
          class:selected={selectedDays.has(0)}
          on:click={() => toggleDay(0)}
        >
          {$_('days.monday')}
        </button>
        <button 
          class:selected={selectedDays.has(1)}
          on:click={() => toggleDay(1)}
        >
          {$_('days.tuesday')}
        </button>
        <button 
          class:selected={selectedDays.has(2)}
          on:click={() => toggleDay(2)}
        >
          {$_('days.wednesday')}
        </button>
        <button 
          class:selected={selectedDays.has(3)}
          on:click={() => toggleDay(3)}
        >
          {$_('days.thursday')}
        </button>
        <button 
          class:selected={selectedDays.has(4)}
          on:click={() => toggleDay(4)}
        >
          {$_('days.friday')}
        </button>
        <button 
          class:selected={selectedDays.has(5)}
          on:click={() => toggleDay(5)}
        >
          {$_('days.saturday')}
        </button>
        <button 
          class:selected={selectedDays.has(6)}
          on:click={() => toggleDay(6)}
        >
          {$_('days.sunday')}
        </button>
      </div>
    {/if}

    <button 
      class="save-btn" 
      on:click={handleSubmit}
      disabled={!title || (!isOneTime && selectedDays.size === 0)}
    >
      {$_('habits.save')}
    </button>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.3);
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    overflow-y: auto;
  }

  .modal {
    position: relative;
    width: 100%;
    background: var(--tg-theme-bg-color);
    padding: 24px;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
  }

  h2 {
    margin: 0 0 20px 0;
    font-size: 24px;
    font-weight: 600;
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
  }

  input[type="text"]:focus {
    outline: none;
    border-color: var(--tg-theme-button-color);
  }

  .type-selector {
    margin: 24px 0;
  }

  .type-selector label {
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
    font-size: 16px;
  }

  .type-selector input[type="checkbox"] {
    width: 22px;
    height: 22px;
    border-radius: 6px;
  }

  .days-selector {
    display: flex;
    gap: 8px;
    margin: 20px 0;
    justify-content: space-between;
  }

  .days-selector button {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 2px solid transparent;
    background: var(--tg-theme-secondary-bg-color);
    cursor: pointer;
    transition: all 0.2s ease;
    font-weight: 500;
    font-size: 15px;
  }

  .days-selector button.selected {
    border-color: var(--tg-theme-button-color);
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  .daily-habit-btn {
    width: 100%;
    padding: 12px;
    border-radius: 12px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .daily-habit-btn:active {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }

  .save-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 16px;
    font-weight: 500;
    margin-top: 24px;
    cursor: pointer;
    transition: opacity 0.2s ease;
  }

  .save-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
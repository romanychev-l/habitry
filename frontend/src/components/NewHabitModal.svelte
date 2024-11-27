<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  let title = '';
  let selectedDays = new Set();
  const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

  function toggleDay(index: number) {
    if (selectedDays.has(index)) {
      selectedDays.delete(index);
    } else {
      selectedDays.add(index);
    }
    selectedDays = selectedDays; // триггер обновления
  }

  function handleSubmit() {
    dispatch('save', {
      title,
      days: Array.from(selectedDays)
    });
  }

  function handleOverlayClick(event: MouseEvent) {
    // Проверяем, что клик был именно по оверлею, а не по модальному окну
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
    <h2>Новая привычка</h2>
    <input 
      type="text" 
      bind:value={title} 
      placeholder="Название привычки"
    />
    
    <div class="days-selector">
      {#each weekDays as day, i}
        <button 
          class:selected={selectedDays.has(i)}
          on:click={() => toggleDay(i)}
        >
          {day}
        </button>
      {/each}
    </div>

    <button 
      class="save-btn" 
      on:click={handleSubmit}
      disabled={!title || selectedDays.size === 0}
    >
      Сохранить
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
  }

  .modal {
    position: relative;
    width: 100%;
    background: var(--tg-theme-bg-color);
    padding: 20px;
    border-radius: 20px 20px 0 0;
    box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  }

  .days-selector {
    display: flex;
    gap: 8px;
    margin: 20px 0;
  }

  .days-selector button {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 2px solid transparent;
    background: var(--tg-theme-secondary-bg-color);
    cursor: pointer;
    transition: all 0.2s ease;
    font-weight: 500;
  }

  .days-selector button.selected {
    border-color: var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    transform: scale(1.05);
  }

  .days-selector button:active {
    transform: scale(0.95);
  }

  .save-btn {
    width: 100%;
    padding: 12px;
    border-radius: 10px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
  }
</style>
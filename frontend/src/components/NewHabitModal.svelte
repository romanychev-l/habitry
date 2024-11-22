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
</script>

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

<style>
  .modal {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
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
    border: none;
    background: var(--tg-theme-secondary-bg-color);
  }

  .days-selector button.selected {
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
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
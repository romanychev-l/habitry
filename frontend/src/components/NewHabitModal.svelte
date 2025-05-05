<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { fade } from 'svelte/transition';
  import type { Habit } from '../types';
  import { themeParams, popup } from '@telegram-apps/sdk-svelte';
  import { habits } from '../stores/habit';
  import { user } from '../stores/user';
  // import { showTelegramOrCustomAlert } from '../stores/alert';
  
  const dispatch = createEventDispatcher();
  export let habit: Habit | null = null;
  export let show = false;
  let title = habit?.title || '';
  let selectedDays = new Set(habit?.days || []);
  let isOneTime = habit?.is_one_time || false;
  let wantToBecome = habit?.want_to_become || '';
  let isDarkTheme = themeParams.backgroundColor() === '#000000';
  let isAuto = habit?.is_auto || false;
  let stake = '';
  let titleInput: HTMLInputElement;
  let modalHeight: number;
  let contentWrapper: HTMLDivElement;
  let showTooltip = false;
  let showWantToBecomeTooltip = false;
  let showStakeTooltip = false;
  let showDaysTooltip = false;

  // Ошибки валидации
  let titleError = '';
  let wantToBecomeError = '';
  const MAX_LENGTH = 34;

  // Предотвращаем скролл на основной странице
  function disableBodyScroll() {
    document.body.style.overflow = 'hidden';
  }
  
  function enableBodyScroll() {
    document.body.style.overflow = '';
  }

  function updateModalHeight() {
    const vh = window.visualViewport?.height || window.innerHeight;
    document.documentElement.style.setProperty('--vh', `${vh}px`);
    if (contentWrapper) {
      contentWrapper.style.height = `${vh}px`;
    }
  }

  // Action для отслеживания кликов вне элемента
  function clickOutside(node: HTMLElement, handler: () => void) {
      const handleClick = (event: MouseEvent) => {
          if (node && !node.contains(event.target as Node) && !event.defaultPrevented) {
              handler();
          }
      };

      document.addEventListener('click', handleClick, true);

      return {
          destroy() {
              document.removeEventListener('click', handleClick, true);
          }
      };
  }

  onMount(() => {
    titleInput?.focus();
    if (habit?.stake) {
      stake = habit.stake.toString();
    }
    
    window.visualViewport?.addEventListener('resize', updateModalHeight);
    window.visualViewport?.addEventListener('scroll', updateModalHeight);
    window.addEventListener('resize', updateModalHeight);
    updateModalHeight();
  });

  onDestroy(() => {
    window.visualViewport?.removeEventListener('resize', updateModalHeight);
    window.visualViewport?.removeEventListener('scroll', updateModalHeight);
    window.removeEventListener('resize', updateModalHeight);
    enableBodyScroll();
  });

  // Добавляем обработчик для управления скроллом при изменении видимости модального окна
  $: if (show) {
    disableBodyScroll();
  } else {
    enableBodyScroll();
  }

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
        popup.open({
          title: $_('alerts.error'),
          message: $_('habits.errors.title_required'),
          buttons: [{ id: 'close', type: 'close' }]
        });
        return;
    }

    const habitData = {
      title,
      want_to_become: wantToBecome,
      days: Array.from(selectedDays),
      is_one_time: isOneTime,
      is_auto: isAuto,
      stake: parseInt(stake) || 0
    };
    dispatch('save', habitData);
  }

  function handleOverlayClick(event: MouseEvent) {
    // Закрываем только если клик был именно по оверлею
    // Also check if any tooltip is open, if so, close it instead of closing the modal
    if (showTooltip || showWantToBecomeTooltip || showStakeTooltip || showDaysTooltip) {
       showTooltip = false;
       showWantToBecomeTooltip = false;
       showStakeTooltip = false;
       showDaysTooltip = false;
       return; // Prevent modal close if a tooltip was closed
    }

    if (event.target === event.currentTarget) {
      dispatch('close');
    }
  }

  function toggleTooltip() {
    showTooltip = !showTooltip;
  }

  function toggleWantToBecomeTooltip() {
    showWantToBecomeTooltip = !showWantToBecomeTooltip;
  }

  function toggleStakeTooltip() {
    showStakeTooltip = !showStakeTooltip;
  }

  function toggleDaysTooltip() {
    showDaysTooltip = !showDaysTooltip;
  }

  // Реактивная валидация длины полей
  $: {
    if (title.length > MAX_LENGTH) {
      titleError = $_('habits.errors.max_length', { values: { max: MAX_LENGTH } }) || `Максимум ${MAX_LENGTH} символов`;
    } else {
      titleError = '';
    }
  }

  $: {
    if (wantToBecome.length > MAX_LENGTH) {
      wantToBecomeError = $_('habits.errors.max_length', { values: { max: MAX_LENGTH } }) || `Максимум ${MAX_LENGTH} символов`;
    } else {
      wantToBecomeError = '';
    }
  }
</script>

<div class="wrapper" bind:this={contentWrapper}>
  <div 
    class="overlay" 
    role="button"
    tabindex="0"
    on:click={handleOverlayClick}
    on:keydown={e => e.key === 'Enter' && dispatch('close')}
  ></div>
  <div class="modal-container">
    <div class="modal">
      <div class="header">
        <h2>{habit ? $_('habits.edit') : $_('habits.add')}</h2>
      </div>

      <div class="content">
        <div class="form-control">
          <label class="label" for="habit-title">
            <span class="label-text">{$_('habits.title')}</span>
          </label>
          <input 
            type="text" 
            id="habit-title"
            bind:value={title}
            bind:this={titleInput}
            placeholder="{$_('habits.title_placeholder') || 'Например: Медитация'}"
            class="input"
          />
          {#if titleError}
            <p class="error-message">{titleError}</p>
          {/if}
        </div>
        
        <div class="form-control">
          <label class="label" for="want-to-become">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.want_to_become')}</span>
              <button class="info-button" on:click|stopPropagation={toggleWantToBecomeTooltip}>?</button>
              {#if showWantToBecomeTooltip}
                <div class="tooltip" use:clickOutside={() => showWantToBecomeTooltip = false} transition:fade>
                  {$_('habits.want_to_become_tooltip')}
                </div>
              {/if}
            </div>
          </label>
          <input
            type="text"
            id="want-to-become"
            bind:value={wantToBecome}
            class="input"
            placeholder={$_('habits.want_to_become_placeholder')}
          />
          {#if wantToBecomeError}
            <p class="error-message">{wantToBecomeError}</p>
          {/if}
        </div>
        
        <div class="form-control">
          <label class="label" for="stake">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.stake')}</span>
              <button class="info-button" on:click|stopPropagation={toggleStakeTooltip}>?</button>
              {#if showStakeTooltip}
                <div class="tooltip" use:clickOutside={() => showStakeTooltip = false} transition:fade>
                  {$_('habits.stake_tooltip')}
                </div>
              {/if}
            </div>
          </label>
          <input
            type="number"
            id="stake"
            bind:value={stake}
            min="0"
            placeholder="0"
            class="input"
          />
        </div>
        
        <div class="form-control">
          <label class="label" for="auto-habit">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.auto_habit')}</span>
              <button class="info-button" on:click|stopPropagation={toggleTooltip}>?</button>
              {#if showTooltip}
                <div class="tooltip" use:clickOutside={() => showTooltip = false} transition:fade>
                  {$_('habits.auto_habit_tooltip')}
                </div>
              {/if}
            </div>
          </label>
          <div class="checkbox-container">
            <input type="checkbox" id="auto-habit" class="checkbox" bind:checked={isAuto} />
          </div>
        </div>
        
        {#if !isOneTime}
          <div class="label">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.days')}</span>
              <button class="info-button" on:click|stopPropagation={toggleDaysTooltip}>?</button>
              {#if showDaysTooltip}
                <div class="tooltip" use:clickOutside={() => showDaysTooltip = false} transition:fade>
                  {$_('habits.days_tooltip')}
                </div>
              {/if}
            </div>
          </div>
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
                class="day-tag all"
                class:selected={selectedDays.size === 7}
                on:click={() => {
                  if (selectedDays.size === 7) {
                    selectedDays = new Set();
                  } else {
                    selectedDays = new Set([0, 1, 2, 3, 4, 5, 6]);
                  }
                }}
              >
                {$_('days.all')}
              </button>
            </div>
          </div>
        {/if}
      </div>

      <div class="footer">
        <button 
          class="save-btn" 
          on:click={handleSave}
          disabled={!title || !!titleError || !!wantToBecomeError || (!isOneTime && selectedDays.size === 0)}
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
    justify-content: center;
    z-index: 1000;
    height: 100%;
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
    display: flex;
    flex-direction: column;
    max-height: 90%;
  }

  .modal {
    width: 100%;
    background: var(--tg-theme-bg-color);
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    display: flex;
    flex-direction: column;
    height: auto;
    max-height: 90vh;
    overflow: hidden;
  }

  @supports (-webkit-touch-callout: none) {
    .wrapper {
      position: fixed;
      height: var(--vh, 100%);
    }
  }

  .header {
    position: sticky;
    top: 0;
    background: inherit;
    z-index: 2;
    padding: 24px 24px 12px 24px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .content {
    padding: 16px 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    flex: 1;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .form-control {
    width: 100%;
  }

  .days-wrapper {
    display: flex;
    gap: 8px;
    align-items: start;
    width: 100%;
  }

  .days-selector {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
    flex: 1;
  }

  .days-selector button {
    aspect-ratio: 1;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-secondary-bg-color);
    font-weight: 500;
    font-size: 14px;
    padding: 0;
    width: 100%;
    height: 48px;
    color: var(--tg-theme-hint-color);
  }

  .days-selector button.selected {
    background: var(--tg-theme-button-color);
    color: white;
  }

  input[type="text"], 
  input[type="number"],
  .input {
    border: 2px solid var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    background: transparent;
    font-size: 16px;
    padding: 14px 12px;
    margin: 8px 0 0 0;
    width: 100%;
    box-sizing: border-box;
    color: var(--tg-theme-text-color);
  }

  input[type="text"]:focus, 
  input[type="number"]:focus,
  .input:focus {
    outline: none;
    border-color: var(--tg-theme-button-color);
  }

  /* Убираем стрелки у input number */
  input[type="number"]::-webkit-inner-spin-button,
  input[type="number"]::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  input[type="number"] {
    -moz-appearance: textfield;
  }

  /* .slider.checked {
    background: #00D5A0;
  }

  .slider.checked:before {
    transform: translateX(16px);
  } */

  .footer {
    position: sticky;
    bottom: 0;
    background: inherit;
    z-index: 2;
    padding: 12px 24px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
    margin-top: auto;
  }

  .save-btn {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    background: var(--tg-theme-button-color);
    color: white;
    font-size: 16px;
    font-weight: 500;
  }

  .save-btn:disabled {
    opacity: 0.6;
  }

  .auto-habit-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .label-with-info {
    display: flex;
    align-items: center;
    position: relative;
  }

  .info-button {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--tg-theme-button-color);
    color: white;
    font-size: 12px;
    font-weight: bold;
    border: none;
    margin-left: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .tooltip {
    position: absolute;
    /* Adjusted positioning for better fit */
    left: 28px; /* Move slightly to the right of the '?' button */
    top: 50%;
    transform: translateY(-50%);
    width: 250px; /* Keep width */
    background: var(--tg-theme-bg-color); /* Adjusted background for theme */
    border-radius: 8px;
    padding: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 1001;
    font-size: 14px;
    color: var(--tg-theme-text-color); /* Adjusted color for theme */
    line-height: 1.4;
    border: 1px solid var(--tg-theme-secondary-bg-color); /* Added border for visibility */
  }

  .checkbox-container {
    margin-top: 8px;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .days-selector button:not(.selected) {
    color: var(--tg-theme-hint-color) !important;
    background-color: var(--tg-theme-secondary-bg-color);
  }

  :global([data-theme="dark"]) .days-selector button.selected {
    color: white !important;
    background-color: var(--tg-theme-button-color);
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  /* Removed broad rule with !important to avoid potential conflicts */
  /* :global([data-theme="dark"]) .modal * {
    color: white !important;
  } */

  /* Use more specific rules for dark theme text */
   :global([data-theme="dark"]) .modal .label-text,
   :global([data-theme="dark"]) .modal .input,
   :global([data-theme="dark"]) .modal .days-selector button { /* Apply to other text elements if needed */
      color: var(--tg-theme-text-color) !important; /* Ensure text color is set */
   }

  /* Keep important for specific overrides where necessary */
  :global([data-theme="dark"]) .modal h2,
  :global([data-theme="dark"]) .modal .save-btn,
  :global([data-theme="dark"]) .info-button {
     color: white !important;
  }


  :global([data-theme="dark"]) input::placeholder {
    color: rgba(255, 255, 255, 0.6) !important;
  }

  :global([data-theme="dark"]) .tooltip {
    background: var(--tg-theme-secondary-bg-color); /* Darker background for tooltip */
    color: var(--tg-theme-text-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    border-color: var(--tg-theme-bg-color); /* Adjust border for dark theme */
  }

  .error-message {
    color: var(--tg-theme-destructive-text-color) !important; /* Added !important back */
    font-size: 12px;
    margin-top: 4px;
    margin-bottom: 0;
  }
</style>
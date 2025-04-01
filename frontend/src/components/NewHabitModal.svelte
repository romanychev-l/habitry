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

  function updateModalHeight() {
    const vh = window.visualViewport?.height || window.innerHeight;
    document.documentElement.style.setProperty('--vh', `${vh}px`);
    if (contentWrapper) {
      contentWrapper.style.height = `${vh}px`;
    }
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
  });

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
</script>

<div class="wrapper" bind:this={contentWrapper}>
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
        </div>
        
        <div class="form-control">
          <label class="label" for="want-to-become">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.want_to_become')}</span>
              <button class="info-button" on:click|stopPropagation={toggleWantToBecomeTooltip}>?</button>
              {#if showWantToBecomeTooltip}
                <div class="tooltip" transition:fade>
                  {$_('habits.want_to_become_tooltip')}
                  <button class="tooltip-close" on:click|stopPropagation={toggleWantToBecomeTooltip}>×</button>
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
        </div>
        
        <div class="form-control">
          <label class="label" for="stake">
            <div class="label-with-info">
              <span class="label-text">{$_('habits.stake')}</span>
              <button class="info-button" on:click|stopPropagation={toggleStakeTooltip}>?</button>
              {#if showStakeTooltip}
                <div class="tooltip" transition:fade>
                  {$_('habits.stake_tooltip')}
                  <button class="tooltip-close" on:click|stopPropagation={toggleStakeTooltip}>×</button>
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
                <div class="tooltip" transition:fade>
                  {$_('habits.auto_habit_tooltip')}
                  <button class="tooltip-close" on:click|stopPropagation={toggleTooltip}>×</button>
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
                <div class="tooltip" transition:fade>
                  {$_('habits.days_tooltip')}
                  <button class="tooltip-close" on:click|stopPropagation={toggleDaysTooltip}>×</button>
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
    border: 2px solid var(--tg-theme-button-color);
    background: var(--tg-theme-bg-color);
    font-weight: 500;
    font-size: 14px;
    padding: 0;
    width: 100%;
    height: 48px;
    color: var(--tg-theme-text-color);
  }

  .days-selector button.selected {
    border-color: var(--tg-theme-button-color);
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
    left: 50px;
    top: 50%;
    transform: translateY(-50%);
    width: 250px;
    background: white;
    border-radius: 8px;
    padding: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 1001;
    font-size: 14px;
    color: #333;
    line-height: 1.4;
  }

  .tooltip-close {
    position: absolute;
    top: 8px;
    right: 8px;
    background: none;
    border: none;
    font-size: 16px;
    cursor: pointer;
    color: #999;
  }

  .checkbox-container {
    margin-top: 8px;
  }

  :global([data-theme="dark"]) .modal {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .days-selector button:not(.selected) {
    color: white !important;
  }

  :global([data-theme="dark"]) .days-selector button {
    color: white !important;
    background-color: var(--tg-theme-secondary-bg-color);
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

  :global([data-theme="dark"]) .tooltip {
    background: var(--tg-theme-bg-color);
    color: var(--tg-theme-text-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }
</style>
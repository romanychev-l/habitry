<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { createEventDispatcher, onMount } from 'svelte';
  import type { Habit } from '../types';
  import { api } from '../utils/api';
  import { initData } from '@telegram-apps/sdk-svelte';
  import DeleteConfirmModal from './DeleteConfirmModal.svelte';
  import UnarchiveConfirmModal from './UnarchiveConfirmModal.svelte';

  const dispatch = createEventDispatcher();

  export let show: boolean = false;
  export let telegramId: number | undefined = undefined;
  let archivedHabits: Habit[] = [];
  let isLoading = false;
  let showUnarchiveConfirm = false;
  let habitToUnarchive: Habit | null = null;
  let showDeleteConfirm = false;
  let habitToDelete: Habit | null = null;

  onMount(loadArchived);

  async function loadArchived() {
    try {
      isLoading = true;
      const data = await api.getArchivedHabits();
      archivedHabits = data || [];
    } catch (e) {
      console.error('Failed to load archived habits', e);
    } finally {
      isLoading = false;
    }
  }

  async function handleUnarchive(habit: Habit) {
    try {
      const updated = await api.unarchiveHabit({ _id: habit._id });
      archivedHabits = archivedHabits.filter(h => h._id !== habit._id);
      dispatch('unarchived', { habit: updated });
    } catch (e) {
      console.error('Failed to unarchive habit', e);
    }
  }

  function openUnarchive(h: Habit) {
    habitToUnarchive = h;
    showUnarchiveConfirm = true;
  }

  function openDelete(h: Habit) {
    habitToDelete = h;
    showDeleteConfirm = true;
  }

  async function handleDeleteConfirmed() {
    if (!habitToDelete) return;
    try {
      const idToDelete = habitToDelete._id; // сохраняем до await, чтобы избежать гонки
      const ownerId = telegramId ?? initData.user()?.id ?? 0;
      await api.deleteHabit({ telegram_id: ownerId, habit_id: idToDelete });
      archivedHabits = archivedHabits.filter(h => h._id !== idToDelete);
      habitToDelete = null;
      showDeleteConfirm = false;
    } catch (e) {
      console.error('Failed to delete archived habit', e);
    }
  }
</script>

{#if show}
<div 
  class="dialog-overlay"
  on:click|stopPropagation={() => dispatch('close')}
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  role="button"
  tabindex="0"
>
  <div class="dialog" role="dialog" aria-modal="true" tabindex="-1" on:click|stopPropagation on:keydown|stopPropagation>
    <div class="dialog-header">
      <h2>{$_('habits.archived.title') || 'Архив'}</h2>
    </div>
    <div class="dialog-content">
      {#if isLoading}
        <div class="loader">{$_('loading') || 'Загрузка...'}</div>
      {:else if archivedHabits.length === 0}
        <div class="empty">{$_('habits.archived.empty') || 'Архив пуст'}</div>
      {:else}
        <ul class="archived-list">
          {#each archivedHabits as h}
            <li class="archived-item">
              <div class="info">
                <div class="title">{h.title}</div>
                {#if h.want_to_become}
                  <div class="subtitle">{h.want_to_become}</div>
                {/if}
              </div>
              <div class="actions">
                <button class="restore" type="button" on:click={() => openUnarchive(h)} on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && openUnarchive(h)}>
                  {$_('habits.archived.restore') || 'Вернуть'}
                </button>
                <button class="delete" type="button" on:click={() => openDelete(h)} on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && openDelete(h)}>
                  {$_('habits.delete')}
                </button>
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
</div>
{#if showUnarchiveConfirm && habitToUnarchive}
  <UnarchiveConfirmModal
    on:close={() => { showUnarchiveConfirm = false; habitToUnarchive = null; }}
    on:unarchive={async () => {
      if (!habitToUnarchive) return;
      await handleUnarchive(habitToUnarchive);
      showUnarchiveConfirm = false;
      habitToUnarchive = null;
    }}
  />
{/if}
{#if showDeleteConfirm && habitToDelete}
  <DeleteConfirmModal
    on:close={() => { showDeleteConfirm = false; habitToDelete = null; }}
    on:delete={handleDeleteConfirmed}
  />
{/if}
{/if}

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: flex-end;
    height: 100dvh;
    z-index: 1000;
  }

  .dialog {
    width: 100%;
    background: var(--tg-theme-bg-color);
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    max-height: 80vh;
    overflow: auto;
  }

  .dialog-header {
    padding: 24px 16px 8px 16px;
    border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    text-align: center;
  }
  .dialog-header h2 { margin: 0; font-size: 18px; font-weight: 600; }

  .dialog-content { padding: 16px; }

  .empty { opacity: 0.7; text-align: center; padding: 24px 0; }
  .loader { opacity: 0.7; text-align: center; padding: 24px 0; }

  .archived-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px; }
  .archived-item { display: flex; justify-content: space-between; align-items: center; padding: 12px; border-radius: 12px; background: var(--tg-theme-secondary-bg-color); }
  .archived-item .info { display: flex; flex-direction: column; gap: 4px; }
  .archived-item .title { font-weight: 600; }
  .archived-item .subtitle { opacity: 0.7; font-size: 13px; }
  .actions { display: flex; align-items: center; gap: 8px; }
  .restore { background: var(--tg-theme-button-color); color: var(--tg-theme-button-text-color); border: none; border-radius: 10px; padding: 8px 12px; cursor: pointer; }
  .delete { background: #ff3b30; color: #fff; border: none; border-radius: 10px; padding: 8px 12px; cursor: pointer; }

  :global([data-theme="dark"]) .dialog { background: var(--tg-theme-bg-color); }
  :global([data-theme="dark"]) .dialog * { color: white !important; }
</style>



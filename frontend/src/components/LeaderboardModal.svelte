<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { api } from '../utils/api';

  export let show: boolean;

  const dispatch = createEventDispatcher();

  type Leader = {
    rank: number;
    username: string;
    first_name: string;
    photo_url: string;
    balance: number;
  }

  let leaders: Leader[] = [];
  let isLoading = true;

  // Предотвращаем скролл на основной странице
  function disableBodyScroll() {
    document.body.style.overflow = 'hidden';
  }

  function enableBodyScroll() {
    document.body.style.overflow = '';
  }

  $: if (show) {
    disableBodyScroll();
  } else {
    enableBodyScroll();
  }

  // Очищаем стили при размонтировании компонента
  onDestroy(() => {
    enableBodyScroll();
  });

  onMount(async () => {
    if (show) {
      try {
        const data = await api.getLeaderboard();
        leaders = data;
      } catch (error) {
        console.error('Failed to fetch leaderboard:', error);
        // Тут можно показать сообщение об ошибке
      } finally {
        isLoading = false;
      }
    }
  });

  function closeModal() {
    dispatch('close');
  }

  function selectLeader(leader: Leader) {
    dispatch('userselect', { username: leader.username });
  }
</script>

{#if show}
  <div class="modal-overlay" on:click={closeModal} role="presentation">
    <div class="modal-content" on:click|stopPropagation role="dialog" aria-modal="true">
      <h2 class="modal-title">{$_('leaderboard.title', { default: 'Leaderboard' })}</h2>
      
      {#if isLoading}
        <p>{$_('leaderboard.loading', { default: 'Loading...' })}</p>
      {:else if leaders.length === 0}
        <p>{$_('leaderboard.empty', { default: 'No leaders yet. Be the first!' })}</p>
      {:else}
        <div class="leaderboard-list">
          {#each leaders as leader}
            <div 
              class="leader-item" 
              on:click={() => selectLeader(leader)} 
              on:keydown={(e) => e.key === 'Enter' && selectLeader(leader)}
              role="button" 
              tabindex="0"
            >
              <div class="leader-rank">{leader.rank}</div>
              <div class="leader-info">
                {#if leader.photo_url}
                  <img src={leader.photo_url} alt={leader.first_name} class="leader-photo" />
                {:else}
                  <div class="leader-placeholder">
                    {leader.first_name?.[0] || leader.username?.[0] || '?'}
                  </div>
                {/if}
                <span class="leader-name">{leader.first_name || leader.username}</span>
              </div>
              <div class="leader-balance">{leader.balance} WILL</div>
            </div>
          {/each}
        </div>
      {/if}

    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 2000;
  }

  .modal-content {
    background: var(--tg-theme-bg-color);
    padding: 24px;
    border-radius: 16px;
    width: 90%;
    max-width: 400px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.1);
    animation: slide-up 0.3s ease-out;
    color: var(--tg-theme-text-color);
  }

  .modal-title {
    font-size: 20px;
    font-weight: 600;
    margin: 0 0 20px 0;
    text-align: center;
  }

  .leaderboard-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .leader-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background-color: var(--tg-theme-secondary-bg-color);
    border-radius: 12px;
    transition: background-color 0.2s ease;
    cursor: pointer;
  }
  
  .leader-item:hover {
    background-color: var(--tg-theme-hint-color);
  }

  .leader-rank {
    font-weight: 600;
    font-size: 16px;
    min-width: 20px;
    text-align: center;
  }

  .leader-info {
    display: flex;
    align-items: center;
    gap: 12px;
    /* Allow info to take up remaining space */
    flex-grow: 1;
    /* Prevent shrinking */
    flex-shrink: 1;
    /* Hide overflow */
    overflow: hidden;
  }
  
  .leader-photo {
    width: 40px;
    height: 40px;
    object-fit: cover;
    flex-shrink: 0;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
  }
  
  .leader-placeholder {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 500;
    text-transform: uppercase;
    flex-shrink: 0;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
  }
  
  .leader-name {
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .leader-balance {
    font-weight: 600;
    color: var(--tg-theme-link-color);
    white-space: nowrap;
    padding-left: 12px;
  }

  @keyframes slide-up {
    from {
      transform: translateY(30px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
</style>

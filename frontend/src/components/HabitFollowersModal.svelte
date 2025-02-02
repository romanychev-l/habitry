<script lang="ts">
    import { _ } from 'svelte-i18n';
    import type { Habit } from '../types';
    import { createEventDispatcher } from 'svelte';
    
    const dispatch = createEventDispatcher();
    
    export let show = false;
    export let habit: Habit;
    export let telegramId: number;
    export let initialFollowers: Array<{ username: string; telegram_id: number }> = [];
    
    let followers = initialFollowers;
    let loading = false;
    let error = '';
    let showUnfollowConfirm = false;
    let selectedFollower: { username: string; telegram_id: number } | null = null;
    
    const API_URL = import.meta.env.VITE_API_URL;
    
    async function loadFollowers() {
        if (initialFollowers.length > 0) {
            followers = initialFollowers;
            return;
        }

        try {
            loading = true;
            const response = await fetch(`${API_URL}/habit/followers?habit_id=${habit._id}`);
            if (!response.ok) {
                throw new Error($_('habits.errors.load_followers'));
            }
            followers = await response.json();
        } catch (err: any) {
            error = err.message || $_('habits.errors.load_followers');
            console.error('Error loading followers:', err);
        } finally {
            loading = false;
        }
    }
    
    async function unfollowHabit() {
        if (!selectedFollower) return;
        
        try {
            const response = await fetch(`${API_URL}/habit/unfollow`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegram_id: telegramId,
                    habit_id: habit._id,
                    unfollow_id: selectedFollower.telegram_id
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.unfollow'));
            }
            
            const currentFollower = selectedFollower;
            followers = followers.filter(f => f.telegram_id !== currentFollower.telegram_id);
            dispatch('followersUpdated', { followers });
            showUnfollowConfirm = false;
            selectedFollower = null;
        } catch (err: any) {
            error = err.message || $_('habits.errors.unfollow');
            console.error('Error unfollowing habit:', err);
        }
    }
    
    function handleUnfollowClick(follower: { username: string; telegram_id: number }) {
        selectedFollower = follower;
        showUnfollowConfirm = true;
    }
    
    function handleClose() {
        show = false;
        dispatch('close');
    }
    
    function handleOverlayClick(event: MouseEvent) {
        if (event.target === event.currentTarget) {
            handleClose();
        }
    }
    
    $: if (show) {
        loadFollowers();
    }
</script>

{#if show}
    <div 
        class="dialog-overlay" 
        on:click|stopPropagation={handleOverlayClick}
        on:keydown={(e) => e.key === 'Escape' && handleClose()}
        role="button"
        tabindex="0"
    >
        <div class="dialog">
            <div class="dialog-header">
                <h2>{$_('habits.followers_list')}</h2>
            </div>
            
            <div class="dialog-content">
                {#if loading}
                    <div class="loading">{$_('common.loading')}</div>
                {:else if error}
                    <div class="error">{error}</div>
                {:else if followers.length === 0}
                    <div class="empty">{$_('habits.no_followers')}</div>
                {:else}
                    <ul class="followers-list">
                        {#each followers as follower}
                            <li class="follower-item">
                                <a 
                                    href="https://t.me/{follower.username}" 
                                    target="_blank" 
                                    rel="noopener noreferrer"
                                    class="username"
                                >
                                    @{follower.username}
                                </a>
                                {#if follower.telegram_id !== habit.creator_id}
                                    <button 
                                        class="unfollow-button"
                                        on:click={() => handleUnfollowClick(follower)}
                                    >
                                        {$_('habits.unfollow')}
                                    </button>
                                {/if}
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        </div>
    </div>
{/if}

{#if showUnfollowConfirm}
    <div 
        class="dialog-overlay" 
        on:click|stopPropagation={handleOverlayClick}
        on:keydown={(e) => e.key === 'Escape' && handleClose()}
        role="button"
        tabindex="0"
    >
        <div class="dialog">
            <div class="dialog-header">
                <h2>{$_('habits.confirm_unfollow')}</h2>
            </div>
            <div class="dialog-content">
                <p class="confirm-text">{$_('habits.unfollow_user', { values: { username: selectedFollower?.username } })}</p>
                <div class="button-group">
                    <button class="dialog-button cancel" on:click={() => showUnfollowConfirm = false}>
                        {$_('common.cancel')}
                    </button>
                    <button class="dialog-button delete" on:click={unfollowHabit}>
                        {$_('common.confirm')}
                    </button>
                </div>
            </div>
        </div>
    </div>
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
        background: #F9F8F3;
        border-radius: 24px 24px 0 0;
        box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
    }

    .dialog-header {
        padding: 32px 16px 16px 16px;
        border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
        text-align: center;
    }

    .dialog-header h2 {
        margin: 0;
        font-size: 20px;
        font-weight: 600;
    }

    .dialog-content {
        padding: 24px;
    }

    .followers-list {
        list-style: none;
        padding: 0;
        margin: 0;
    }
    
    .follower-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px;
        border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    }
    
    .username {
        font-size: 16px;
        text-decoration: none;
        color: var(--tg-theme-text-color);
    }
    
    .username:hover {
        text-decoration: underline;
    }
    
    .unfollow-button {
        padding: 8px 16px;
        border-radius: 12px;
        background: #ff3b30;
        color: white;
        border: none;
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
    }

    .button-group {
        display: flex;
        flex-direction: column;
        gap: 12px;
        margin-top: 24px;
    }

    .dialog-button {
        width: 100%;
        padding: 14px;
        border-radius: 12px;
        border: none;
        font-size: 16px;
        font-weight: 500;
        text-align: center;
        cursor: pointer;
    }

    .dialog-button.cancel {
        background: var(--tg-theme-button-color);
        color: var(--tg-theme-button-text-color);
    }

    .dialog-button.delete {
        background: #ff3b30;
        color: white;
    }

    .confirm-text {
        margin-bottom: 0;
        text-align: center;
        font-size: 16px;
    }
    
    .loading, .error, .empty {
        text-align: center;
        padding: 20px;
    }
    
    .error {
        color: #ff3b30;
    }

    :global([data-theme="dark"]) .dialog {
        background: var(--tg-theme-bg-color);
    }

    :global([data-theme="dark"]) .dialog * {
        color: white !important;
    }
</style> 
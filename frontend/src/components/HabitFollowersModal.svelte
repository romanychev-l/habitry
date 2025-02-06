<script lang="ts">
    import { _ } from 'svelte-i18n';
    import type { Habit } from '../types';
    import { habits } from '../stores/habit';
    import { createEventDispatcher } from 'svelte';
    import ActivityHeatmap from './ActivityHeatmap.svelte';
    import { api } from '../utils/api';
    
    const dispatch = createEventDispatcher();
    
    export let show = false;
    export let habit: Habit;
    export let telegramId: number;
    export let initialFollowers: Array<{ username: string; telegram_id: number }> | null = null;
    
    let followers: Array<{ username: string; telegram_id: number }> = [];
    let loading = false;
    let error = '';
    let success = '';
    let showUnfollowConfirm = false;
    let selectedFollower: { username: string; telegram_id: number } | null = null;
    let activityData: { date: string; count: number }[] = [];
    
    const API_URL = import.meta.env.VITE_API_URL;
    
    async function loadFollowers() {
        try {
            loading = true;
            console.log('Loading followers for habit:', habit._id, 'telegramId:', telegramId);
            if (initialFollowers && initialFollowers.length > 0) {
                followers = initialFollowers.map(f => ({ ...f }));
                console.log('Using initial followers:', followers);
            } else {
                const data = await api.getHabitFollowers(habit._id, telegramId);
                console.log('Received followers data:', data);
                followers = Array.isArray(data) ? data : [];
                console.log('Processed followers:', followers);
            }
        } catch (err: any) {
            console.error('Error loading followers:', err);
            error = err.message || $_('habits.errors.load_followers');
            followers = [];
        } finally {
            loading = false;
        }
    }

    async function loadActivityData() {
        console.log("Loading activity data for habit:", habit);
        try {
            const data = await api.getHabitActivity(habit._id, telegramId);
            console.log("Activity data response:", data);
            activityData = [...data];
        } catch (err) {
            console.error('Error loading activity data:', err);
            activityData = [];
        }
    }
    
    async function unfollowHabit() {
        if (!selectedFollower) return;
        
        error = '';
        success = '';
        
        const requestData = {
            habit_id: habit._id,
            unfollow_id: selectedFollower.telegram_id
        };
        
        console.log('Отправляем запрос на отписку:', requestData);
        
        try {
            await api.unfollowHabit(requestData);
            console.log('Успешно отписались');
            
            // Обновляем список подписчиков
            const data = await api.getHabitFollowers(habit._id, telegramId);
            dispatch('followersUpdated', { followers: data });
            
            // Закрываем модальное окно
            showUnfollowConfirm = false;
            selectedFollower = null;
            success = $_('habits.unfollow_success');
        } catch (error) {
            console.error('Error unfollowing:', error);
            error = $_('habits.errors.unfollow');
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
        loadActivityData();
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
                <div class="activity-section">
                    <h3>{$_('habits.activity')}</h3>
                    <ActivityHeatmap data={activityData} />
                </div>
                
                <div class="followers-section">
                    <h3>{$_('habits.followers_list')}</h3>
                    {#if loading}
                        <div class="loading">{$_('common.loading')}</div>
                    {:else if error}
                        <div class="error">{error}</div>
                    {:else if success}
                        <div class="success">{success}</div>
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
                                    <button 
                                        class="unfollow-button"
                                        on:click={() => handleUnfollowClick(follower)}
                                    >
                                        {$_('habits.unfollow')}
                                    </button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                </div>
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
        padding: 24px 16px 16px 16px;
        border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
        text-align: center;
    }

    .dialog-header h2 {
        margin: 0;
        font-size: 20px;
        font-weight: 600;
    }

    .dialog-content {
        padding: 16px 24px;
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
        text-align: center;
        padding: 12px;
        margin-bottom: 16px;
    }

    .success {
        color: #34c759;
        text-align: center;
        padding: 12px;
        margin-bottom: 16px;
    }

    :global([data-theme="dark"]) .dialog {
        background: var(--tg-theme-bg-color);
    }

    :global([data-theme="dark"]) .dialog * {
        color: white !important;
    }

    .activity-section {
        margin-bottom: 1rem;
        padding: 1rem;
        background-color: var(--background-secondary);
        border-radius: 8px;
    }
    
    .activity-section h3 {
        margin-top: 0;
        margin-bottom: 0.5rem;
        color: var(--text-primary);
    }
    
    .followers-section {
        padding: 1rem;
        background-color: var(--background-secondary);
        border-radius: 8px;
    }
    
    .followers-section h3 {
        margin-top: 0;
        margin-bottom: 0.5rem;
        color: var(--text-primary);
    }
</style> 
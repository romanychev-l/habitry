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
    export let initialFollowers: Array<{ username: string; telegram_id: number; first_name?: string; photo_url?: string; completed_today?: boolean }> | null = null;
    
    let followers: Array<{ 
        username: string; 
        telegram_id: number; 
        first_name?: string; 
        photo_url?: string;
        is_mutual?: boolean;
        completed_today?: boolean;
    }> = [];
    let loading = false;
    let error = '';
    let success = '';
    let showUnfollowConfirm = false;
    let selectedFollower: { username: string; telegram_id: number; first_name?: string; photo_url?: string } | null = null;
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
    
    function handleUnfollowClick(follower: { username: string; telegram_id: number; first_name?: string; photo_url?: string }) {
        selectedFollower = follower;
        showUnfollowConfirm = true;
    }
    
    function handlePingClick(follower: { username: string; telegram_id: number; first_name?: string; photo_url?: string }) {
        // Сразу отправляем запрос на создание пинга без подтверждения
        try {
            api.createPing({
                follower_id: follower.telegram_id,
                follower_username: follower.username,
                habit_id: habit._id,
                habit_title: habit.title,
                sender_id: telegramId,
                sender_username: window.Telegram?.WebApp?.initDataUnsafe?.user?.username || ""
            })
            .then(() => {
                // Показываем временное сообщение об успехе
                success = $_('habits.ping_success', { values: { username: follower.username } });
                
                // Также показываем всплывающее уведомление через Telegram.WebApp если доступно
                if (window.Telegram?.WebApp?.showPopup) {
                    window.Telegram.WebApp.showPopup({
                        title: $_('habits.ping_sent_title'),
                        message: $_('habits.ping_sent_message', { values: { username: follower.username } }),
                        buttons: [{ type: "ok" }]
                    });
                }
                
                // Скрываем сообщение через 3 секунды
                const username = follower.username;
                setTimeout(() => {
                    if (success === $_('habits.ping_success', { values: { username } })) {
                        success = '';
                    }
                }, 3000);
            })
            .catch((error: Error) => {
                console.error('Error creating ping:', error);
                if (window.Telegram && window.Telegram.WebApp) {
                    window.Telegram.WebApp.showAlert($_('habits.errors.ping'));
                }
            });
        } catch (error) {
            console.error('Error sending ping:', error);
            if (window.Telegram && window.Telegram.WebApp) {
                window.Telegram.WebApp.showAlert($_('habits.errors.ping'));
            }
        }
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
                                    <div class="follower-info">
                                        {#if follower.photo_url}
                                            <img 
                                                src={follower.photo_url} 
                                                alt={follower.username} 
                                                class="follower-avatar" 
                                            />
                                        {:else}
                                            <div class="follower-avatar-placeholder">
                                                {follower.first_name?.[0] || follower.username?.[0] || '?'}
                                            </div>
                                        {/if}
                                        <div class="follower-details">
                                            <span class="follower-name">{follower.first_name || follower.username}</span>
                                            <a 
                                                href="https://t.me/{follower.username}" 
                                                target="_blank" 
                                                rel="noopener noreferrer"
                                                class="username"
                                            >
                                                @{follower.username}
                                            </a>
                                        </div>
                                    </div>
                                    <div class="follower-actions">
                                        {#if follower.is_mutual}
                                            {#if !follower.completed_today}
                                                <button 
                                                    class="ping-button"
                                                    on:click={() => handlePingClick(follower)}
                                                    title={$_('habits.ping_follower')}
                                                >
                                                    🔔
                                                </button>
                                            {:else}
                                                <span class="completed-icon" title={$_('habits.completed_today')}>
                                                    ✅
                                                </span>
                                            {/if}
                                        {:else if follower.completed_today}
                                            <span class="completed-icon" title={$_('habits.completed_today')}>
                                                ✅
                                            </span>
                                        {:else}
                                            <span class="not-completed-icon" title={$_('habits.not_completed_today')}>
                                                ❌
                                            </span>
                                        {/if}
                                        <button 
                                            class="unfollow-button"
                                            on:click={() => handleUnfollowClick(follower)}
                                            title={$_('habits.unfollow')}
                                        >
                                            🗑️
                                        </button>
                                    </div>
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

    .activity-section {
        margin-bottom: 1rem;
        padding: 1rem 0;
        background-color: var(--background-secondary);
        border-radius: 8px;
    }
    
    .activity-section h3 {
        margin: 0 1rem 0.5rem 1rem;
        color: var(--text-primary);
    }
    
    .followers-section {
        padding: 1rem 0;
        background-color: var(--background-secondary);
        border-radius: 8px;
    }
    
    .followers-section h3 {
        margin: 0 1rem 0.5rem 1rem;
        color: var(--text-primary);
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
        padding: 12px 0;
        border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
        margin: 0 1rem;
    }
    
    .follower-item:first-child {
        padding-top: 0;
    }
    
    .follower-item:last-child {
        padding-bottom: 0;
        border-bottom: none;
    }
    
    .follower-info {
        display: flex;
        align-items: center;
        gap: 12px;
    }
    
    .follower-avatar {
        width: 40px;
        height: 40px;
        object-fit: cover;
        mask: url('/src/assets/squircley.svg') no-repeat center / contain;
        -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    }
    
    .follower-avatar-placeholder {
        width: 40px;
        height: 40px;
        background: var(--tg-theme-button-color);
        color: var(--tg-theme-button-text-color);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        font-weight: 500;
        text-transform: uppercase;
        mask: url('/src/assets/squircley.svg') no-repeat center / contain;
        -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    }
    
    .follower-details {
        display: flex;
        flex-direction: column;
    }
    
    .follower-name {
        font-size: 16px;
        font-weight: 500;
        color: var(--tg-theme-text-color);
    }
    
    .username {
        font-size: 14px;
        text-decoration: none;
        color: var(--tg-theme-hint-color, #999);
    }
    
    .username:hover {
        text-decoration: underline;
    }
    
    .follower-actions {
        display: flex;
        gap: 8px;
        align-items: center;
    }
    
    .completed-icon {
        display: flex;
        justify-content: center;
        align-items: center;
        width: 32px;
        height: 32px;
        font-size: 16px;
        border-radius: 8px;
    }
    
    .mutual-icon, .one-way-icon, .completed-icon {
        font-size: 20px;
    }
    
    .not-completed-icon {
        display: flex;
        justify-content: center;
        align-items: center;
        width: 32px;
        height: 32px;
        font-size: 16px;
        border-radius: 8px;
        color: #ff3b30;
    }
    
    .unfollow-button {
        padding: 0;
        width: 32px;
        height: 32px;
        border-radius: 8px;
        background: #ff3b30;
        color: white;
        border: none;
        cursor: pointer;
        font-size: 16px;
        font-weight: 500;
        display: flex;
        align-items: center;
        justify-content: center;
        line-height: 1;
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

    .ping-button {
        display: flex;
        justify-content: center;
        align-items: center;
        background: #007aff;
        color: white;
        width: 32px;
        height: 32px;
        border: none;
        border-radius: 8px;
        font-size: 16px;
        cursor: pointer;
    }

    :global([data-theme="dark"]) .dialog {
        background: var(--tg-theme-bg-color);
    }

    :global([data-theme="dark"]) .dialog * {
        color: white !important;
    }
</style> 
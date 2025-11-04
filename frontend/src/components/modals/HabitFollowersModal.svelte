<script lang="ts">
    import { _ } from 'svelte-i18n';
    import type { Habit } from '../../types';
    import { createEventDispatcher, onMount, onDestroy } from 'svelte';
    import { fade, fly } from 'svelte/transition';
    import ActivityHeatmap from '../habits/ActivityHeatmap.svelte';
    import { api } from '../../utils/api';
    import { user } from '../../stores/user';
    import { popup } from '@tma.js/sdk-svelte';
    
    const dispatch = createEventDispatcher();
    
    export let show = false;
    export let habit: Habit;
    export let telegramId: number;
    
    type FollowerDetail = {
        _id: string;
        telegram_id: number;
        username: string;
        first_name?: string;
        photo_url?: string;
        title: string;
        last_click_date: string;
        streak: number;
        score: number;
        completed_today: boolean;
        currentUserFollowsThisUser: boolean;
        thisUserFollowsCurrentUser: boolean;
    };

    let followers: Array<FollowerDetail> = [];
    let loading = false;
    let error = '';
    let showUnfollowConfirm = false;
    let selectedFollowerForUnfollow: FollowerDetail | null = null;
    
    let activityData: { date: string; count: number }[] = [];
    
    function getCurrentDate() {
        const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
        const date = new Date();
        const localDate = new Date(date.toLocaleString('en-US', { timeZone: userTimezone }));
        
        const year = localDate.getFullYear();
        const month = String(localDate.getMonth() + 1).padStart(2, '0');
        const day = String(localDate.getDate()).padStart(2, '0');
        
        return `${year}-${month}-${day}`;
    }
    
    // Предотвращаем скролл на основной странице
    function disableBodyScroll() {
        document.body.style.overflow = 'hidden';
    }
    
    function enableBodyScroll() {
        document.body.style.overflow = '';
    }
    
    $: if (show) {
        loadFollowers();
        loadActivityData();
        disableBodyScroll();
    } else {
        enableBodyScroll();
    }
    
    // Очищаем стили при размонтировании компонента
    onDestroy(() => {
        enableBodyScroll();
    });
    
    async function loadFollowers() {
        try {
            loading = true;
            console.log('Loading followers for habit:', habit._id, 'telegramId:', telegramId);
            
            const data = await api.getHabitFollowers(habit._id) as Array<FollowerDetail>;
            console.log('Received followers data:', data);
            
            if (!Array.isArray(data)) {
                console.warn('Setting followers to empty array because data is not an array');
                followers = [];
            } else {
                followers = data;
                console.log('Processed followers:', followers);
            }
            
            dispatch('followersUpdated', { followers });
            
            error = '';
        } catch (err: unknown) {
            console.error('Error loading followers:', err);
            error = err instanceof Error ? err.message : $_('habits.errors.load_followers');
            
            followers = [];
        } finally {
            loading = false;
        }
    }

    async function loadActivityData() {
        console.log("Loading activity data for habit:", habit);
        try {
            const rawActivityData = await api.getHabitActivity(habit._id);

            activityData = rawActivityData.map((item: { date: string; done: boolean }) => ({
                date: item.date,
                count: item.done ? 1 : 0
            }));
            console.log("Activity data response:", activityData);
        } catch (err) {
            console.error('Error loading activity data:', err);
            activityData = [];
        }
    }
    
    async function unfollowHabit() {
        if (!selectedFollowerForUnfollow) return;
        
        error = '';
        
        const requestData = {
            habit_id: habit._id,
            unfollow_id: selectedFollowerForUnfollow.telegram_id
        };
        
        console.log('Отправляем запрос на отписку:', requestData);
        
        try {
            await api.unfollowHabit({
                habit_id: habit._id,
                unfollow_id: selectedFollowerForUnfollow.telegram_id
            });

            console.log('Успешно отписались');
            await loadFollowers();
            showUnfollowConfirm = false;
            selectedFollowerForUnfollow = null;
            
            await popup.show({
                title: $_('alerts.success'),
                message: $_('habits.unfollow_success'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        } catch (errorMsg) {
            console.error('Error unfollowing:', errorMsg);
            error = $_('habits.errors.unfollow');
        }
    }
    
    function handleUnfollowClick(follower: FollowerDetail) {
        selectedFollowerForUnfollow = follower;
        showUnfollowConfirm = true;
    }

    async function handleSubscribeClick(targetHabit: FollowerDetail) {
        if (!targetHabit) return;
        error = '';
        loading = true;

        try {
            await api.subscribeToFollowerHabit({
                current_user_habit_id: habit._id,
                target_user_habit_id: targetHabit._id
            });

            await popup.show({
                title: $_('alerts.success'),
                message: $_('habits.follow_success', { values: { username: targetHabit.username } }),
                buttons: [{ id: 'close', type: 'close' }]
            });
            await loadFollowers();
        } catch (err) {
            console.error('Error subscribing:', err);
            error = $_('habits.errors.follow');
            popup.show({
                title: $_('alerts.error'),
                message: error,
                buttons: [{ id: 'close', type: 'close' }]
            });
        } finally {
            loading = false;
        }
    }
    
    function handlePingClick(follower: FollowerDetail) {
        const today = getCurrentDate();
        if (habit.last_click_date !== today) {
            popup.show({
                title: $_('alerts.warning'),
                message: $_('habits.complete_before_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        try {
            api.createPing({
                follower_id: follower.telegram_id,
                follower_username: follower.username,
                habit_id: habit._id,
                habit_title: habit.title,
                sender_id: telegramId,
                sender_username: $user?.username || ""
            })
            .then(async () => {
                // Показываем стандартный алерт
                await popup.show({
                    title: $_('alerts.success'),
                    message: $_('habits.ping_sent_message', { values: { username: follower.username } }),
                    buttons: [{ id: 'close', type: 'close' }]
                });
                
                // Обновляем список подписчиков
                try {
                    await loadFollowers();
                } catch (err) {
                    console.error('Error reloading followers after ping:', err);
                }
            })
            .catch(async (error: Error) => {
                console.error('Error creating ping:', error);
                await popup.show({
                    title: $_('alerts.error'),
                    message: $_('habits.errors.ping'),
                    buttons: [{ id: 'close', type: 'close' }]
                });
            });
        } catch (error) {
            console.error('Error sending ping:', error);
            popup.show({
                title: $_('alerts.error'),
                message: $_('habits.errors.ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        }
    }
    
    function handlePingAllClick() {
        const today = getCurrentDate();
        if (habit.last_click_date !== today) {
            popup.show({
                title: $_('alerts.warning'),
                message: $_('habits.complete_before_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        const followersToPing = followers.filter(f => 
            f.currentUserFollowsThisUser && 
            f.thisUserFollowsCurrentUser && 
            !f.completed_today
        );
        
        if (followersToPing.length === 0) {
            popup.show({
                title: $_('alerts.info'),
                message: $_('habits.no_followers_to_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        let successCount = 0;
        let pingPromises: Promise<any>[] = [];
        
        // Отправляем пинг каждому подписчику и собираем промисы
        followersToPing.forEach(follower => {
            const pingPromise = api.createPing({
                follower_id: follower.telegram_id,
                follower_username: follower.username,
                habit_id: habit._id,
                habit_title: habit.title,
                sender_id: telegramId,
                sender_username: $user?.username || ""
            })
            .then(() => {
                successCount++;
            })
            .catch((error: Error) => {
                console.error('Error creating ping:', error);
            });
            
            pingPromises.push(pingPromise);
        });
        
        // Ждем завершения всех запросов на пинг
        Promise.all(pingPromises)
            .then(async () => {
                // Если были успешные пинги, показываем сообщение об успехе
                if (successCount > 0) {
                    await popup.show({
                        title: $_('alerts.success'),
                        message: $_('habits.ping_all_sent_message', { values: { count: successCount } }),
                        buttons: [{ id: 'close', type: 'close' }]
                    });
                }
            });
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
    
    function handleDialogScroll(event: Event) {
        event.stopPropagation();
    }
    
    onMount(() => {
        if (show) {
            loadFollowers();
            loadActivityData();
        }
    });

    $: usersIFollow = followers.filter(f => f.currentUserFollowsThisUser);
    $: usersFollowingMeNotMutual = followers.filter(f => f.thisUserFollowsCurrentUser && !f.currentUserFollowsThisUser);
</script>

{#if show}
    <div 
        class="dialog-overlay" 
        on:click|stopPropagation={handleOverlayClick}
        on:keydown={(e) => e.key === 'Escape' && handleClose()}
        role="button"
        tabindex="0"
        transition:fade={{ duration: 200 }}
    >
        <div class="dialog" transition:fly={{ y: 500, duration: 300, opacity: 1 }}>
            <div class="dialog-header">
                <h2>{$_('habits.followers_management')}</h2>
            </div>
            
            <div 
                class="dialog-content" 
                on:scroll|stopPropagation={handleDialogScroll}
            >
                <div class="activity-section">
                    <h3>{$_('habits.activity')}</h3>
                    <ActivityHeatmap {habit} data={activityData} />
                </div>
                
                <button 
                    class="ping-all-button" 
                    on:click={handlePingAllClick}
                    title={$_('habits.ping_all_inactive_mutual')}
                    disabled={followers.filter(f => f.currentUserFollowsThisUser && f.thisUserFollowsCurrentUser && !f.completed_today).length === 0}
                >
                    🔔 {$_('habits.ping_all_inactive_mutual')}
                </button>
                
                {#if error}
                    <div class="error">{error}</div>
                {/if}

                <div class="followers-section">
                    <h3>{$_('habits.i_follow')}</h3>
                    {#if loading}
                        <div class="skeleton-list">
                            {#each [1, 2, 3] as _}
                                <div class="skeleton-item">
                                    <div class="skeleton-avatar"></div>
                                    <div class="skeleton-text">
                                        <div class="skeleton-line"></div>
                                        <div class="skeleton-line short"></div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {:else if usersIFollow.length === 0}
                        <div class="empty">{$_('habits.no_one_i_follow')}</div>
                    {:else if usersIFollow.length > 0}
                        <ul class="followers-list">
                            {#each usersIFollow as follower (follower._id)}
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
                                        {#if follower.thisUserFollowsCurrentUser}
                                            {#if !follower.completed_today}
                                                <button 
                                                    class="ping-button"
                                                    on:click={() => handlePingClick(follower)}
                                                    title={$_('habits.ping_follower')}
                                                >
                                                    🔔
                                                </button>
                                            {:else if follower.completed_today}
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

                <div class="followers-section section-spacing">
                    <h3>{$_('habits.following_me')}</h3>
                    {#if loading}
                        <div class="skeleton-list">
                            {#each [1, 2] as _}
                                <div class="skeleton-item">
                                    <div class="skeleton-avatar"></div>
                                    <div class="skeleton-text">
                                        <div class="skeleton-line"></div>
                                        <div class="skeleton-line short"></div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {:else if usersFollowingMeNotMutual.length === 0}
                        <div class="empty">{$_('habits.no_one_following_me_yet')}</div>
                    {:else if usersFollowingMeNotMutual.length > 0}
                        <ul class="followers-list">
                            {#each usersFollowingMeNotMutual as follower (follower._id)}
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
                                        {#if follower.completed_today}
                                            <span class="completed-icon" title={$_('habits.completed_today')}>
                                                ✅
                                            </span>
                                        {:else}
                                             <span class="not-completed-icon" title={$_('habits.not_completed_today')}>
                                                ❌
                                            </span>
                                        {/if}
                                        <button 
                                            class="subscribe-button"
                                            on:click={() => handleSubscribeClick(follower)}
                                            title={$_('habits.subscribe_to_follower', { values: { username: follower.username }})}
                                        >
                                            ➕
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
        on:click|stopPropagation={(e) => { if (e.target === e.currentTarget) showUnfollowConfirm = false; }}
        on:keydown={(e) => e.key === 'Escape' && (showUnfollowConfirm = false)}
        role="button"
        tabindex="0"
        transition:fade={{ duration: 200 }}
    >
        <div class="dialog" transition:fly={{ y: 500, duration: 300, opacity: 1 }}>
            <div class="dialog-header">
                <h2>{$_('habits.confirm_unfollow')}</h2>
            </div>
            <div class="dialog-content">
                <p class="confirm-text">{$_('habits.unfollow_user', { values: { username: selectedFollowerForUnfollow?.username } })}</p>
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
        background: var(--tg-theme-bg-color);
        border-radius: 24px 24px 0 0;
        box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
        max-height: 90vh;
        min-height: 400px;
        display: flex;
        flex-direction: column;
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
        overflow-y: auto;
        flex: 1;
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

    .subscribe-button {
        padding: 0;
        width: 32px;
        height: 32px;
        border-radius: 8px;
        background: var(--tg-theme-button-color);
        color: var(--tg-theme-button-text-color);
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
    
    .empty {
        text-align: center;
        padding: 60px 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 120px;
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
        background: var(--tg-theme-button-color);
        color: white;
        width: 32px;
        height: 32px;
        border: none;
        border-radius: 8px;
        font-size: 16px;
        cursor: pointer;
    }

    .ping-all-button {
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--tg-theme-button-color);
        color: white;
        border: none;
        border-radius: 12px;
        padding: 12px 16px;
        font-size: 16px;
        font-weight: 500;
        cursor: pointer;
        margin: 12px 0;
        width: 100%;
        gap: 8px;
    }
    
    .ping-all-button:active {
        opacity: 0.8;
    }

    .section-spacing {
        margin-top: 1.5rem;
    }

    /* Skeleton loaders */
    .skeleton-list {
        padding: 0 1rem;
    }

    .skeleton-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 0;
        border-bottom: 1px solid var(--tg-theme-secondary-bg-color);
    }

    .skeleton-item:last-child {
        border-bottom: none;
    }

    .skeleton-avatar {
        width: 40px;
        height: 40px;
        border-radius: 12px;
        background: linear-gradient(90deg, 
            var(--tg-theme-secondary-bg-color) 0%, 
            var(--tg-theme-hint-color, rgba(0,0,0,0.1)) 50%, 
            var(--tg-theme-secondary-bg-color) 100%);
        background-size: 200% 100%;
        animation: skeleton-loading 1.5s ease-in-out infinite;
        flex-shrink: 0;
    }

    .skeleton-text {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .skeleton-line {
        height: 14px;
        background: linear-gradient(90deg, 
            var(--tg-theme-secondary-bg-color) 0%, 
            var(--tg-theme-hint-color, rgba(0,0,0,0.1)) 50%, 
            var(--tg-theme-secondary-bg-color) 100%);
        background-size: 200% 100%;
        animation: skeleton-loading 1.5s ease-in-out infinite;
        border-radius: 4px;
        width: 60%;
    }

    .skeleton-line.short {
        width: 40%;
        height: 12px;
    }

    @keyframes skeleton-loading {
        0% {
            background-position: 200% 0;
        }
        100% {
            background-position: -200% 0;
        }
    }

    :global([data-theme="dark"]) .dialog {
        background: var(--tg-theme-bg-color);
    }

    :global([data-theme="dark"]) .dialog * {
        color: white !important;
    }

    :global([data-theme="dark"]) .skeleton-avatar,
    :global([data-theme="dark"]) .skeleton-line {
        background: linear-gradient(90deg, 
            rgba(255,255,255,0.05) 0%, 
            rgba(255,255,255,0.1) 50%, 
            rgba(255,255,255,0.05) 100%);
        background-size: 200% 100%;
    }
</style> 
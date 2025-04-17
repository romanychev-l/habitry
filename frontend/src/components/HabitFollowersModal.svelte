Кн<script lang="ts">
    import { _ } from 'svelte-i18n';
    import type { Habit } from '../types';
    import { habits } from '../stores/habit';
    import { createEventDispatcher, onMount, onDestroy } from 'svelte';
    import ActivityHeatmap from './ActivityHeatmap.svelte';
    import { api } from '../utils/api';
    // import { showTelegramOrCustomAlert } from '../stores/alert';
    import { user } from '../stores/user';
    import { popup, initData } from '@telegram-apps/sdk-svelte';
    
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
    let showUnfollowConfirm = false;
    let selectedFollower: { username: string; telegram_id: number; first_name?: string; photo_url?: string } | null = null;
    let activityData: { date: string; count: number }[] = [];
    
    const API_URL = import.meta.env.VITE_API_URL;
    
    // Функция для получения текущей даты с учетом часового пояса
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
            
            // Получаем данные от API
            const data = await api.getHabitFollowers(habit._id);
            console.log('Received followers data:', data);
            
            // Проверяем и обрабатываем данные
            if (!Array.isArray(data)) {
                console.warn('Setting followers to empty array because data is not an array');
                followers = [];
            } else {
                followers = data;
                console.log('Processed followers:', followers);
            }
            
            // Отправляем событие обновления в родительский компонент
            dispatch('followersUpdated', { followers });
            
            error = ''; // Сбрасываем ошибку если запрос успешен
        } catch (err: unknown) {
            console.error('Error loading followers:', err);
            error = err instanceof Error ? err.message : $_('habits.errors.load_followers');
            
            // В случае ошибки, если есть initialFollowers, используем его
            if (initialFollowers && initialFollowers.length > 0) {
                followers = initialFollowers.map(f => ({ ...f }));
            } else {
                followers = [];
            }
        } finally {
            loading = false;
        }
    }

    async function loadActivityData() {
        console.log("Loading activity data for habit:", habit);
        try {
            const data = await api.getHabitActivity(habit._id);
            console.log("Activity data response:", data);
            activityData = Array.isArray(data) ? data : [];
        } catch (err) {
            console.error('Error loading activity data:', err);
            activityData = [];
        }
    }
    
    async function unfollowHabit() {
        if (!selectedFollower) return;
        
        error = '';
        
        const requestData = {
            habit_id: habit._id,
            unfollow_id: selectedFollower.telegram_id
        };
        
        console.log('Отправляем запрос на отписку:', requestData);
        
        try {
            await api.unfollowHabit(requestData);
            console.log('Успешно отписались');
            
            // Обновляем список подписчиков
            await loadFollowers();
            
            // Закрываем модальное окно подтверждения
            showUnfollowConfirm = false;
            selectedFollower = null;
            
            // Показываем алерт
            await popup.open({
                title: $_('alerts.success'),
                message: $_('habits.unfollow_success'),
                buttons: [{ id: 'close', type: 'close' }]
            });
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
        // Проверяем, выполнил ли пользователь привычку сегодня
        const today = getCurrentDate();
        if (habit.last_click_date !== today) {
            // Показываем стандартный алерт с предупреждением
            popup.open({
                title: $_('alerts.warning'),
                message: $_('habits.complete_before_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        // Сразу отправляем запрос на создание пинга без подтверждения
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
                await popup.open({
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
                await popup.open({
                    title: $_('alerts.error'),
                    message: $_('habits.errors.ping'),
                    buttons: [{ id: 'close', type: 'close' }]
                });
            });
        } catch (error) {
            console.error('Error sending ping:', error);
            popup.open({
                title: $_('alerts.error'),
                message: $_('habits.errors.ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        }
    }
    
    // Новая функция для отправки пинга всем подписчикам, которые не выполнили привычку
    function handlePingAllClick() {
        // Проверяем, выполнил ли пользователь привычку сегодня
        const today = getCurrentDate();
        if (habit.last_click_date !== today) {
            // Показываем стандартный алерт с предупреждением
            popup.open({
                title: $_('alerts.warning'),
                message: $_('habits.complete_before_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        // Фильтруем только взаимных подписчиков, которые не выполнили привычку сегодня
        const followersToPing = followers.filter(f => f.is_mutual && !f.completed_today);
        
        if (followersToPing.length === 0) {
            // Показываем алерт, если нет подписчиков для пинга
            popup.open({
                title: $_('alerts.info'),
                message: $_('habits.no_followers_to_ping'),
                buttons: [{ id: 'close', type: 'close' }]
            });
            return;
        }
        
        // Счетчик успешных пингов
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
                    await popup.open({
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
    
    // Добавляем обработчик инициализации
    onMount(() => {
        // if (window.Telegram?.WebApp?.ready) {
        //     // Свойство ready - это булев флаг, а не метод
        // }
        
        // Устанавливаем начальное состояние из initialFollowers, если они доступны
        if (initialFollowers && initialFollowers.length > 0 && show) {
            followers = initialFollowers.map(f => ({ ...f }));
            console.log('Set initial followers state:', followers);
        }
    });
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
            
            <div 
                class="dialog-content" 
                on:scroll|stopPropagation={handleDialogScroll}
            >
                <div class="activity-section">
                    <h3>{$_('habits.activity')}</h3>
                    <ActivityHeatmap data={activityData} />
                </div>
                
                <!-- Кнопка "Пингануть всех" между секциями -->
                <button 
                    class="ping-all-button" 
                    on:click={handlePingAllClick}
                    title={$_('habits.ping_all_inactive')}
                >
                    🔔 {$_('habits.ping_all_inactive')}
                </button>
                
                <div class="followers-section">
                    <h3>{$_('habits.followers_list')}</h3>
                    
                    {#if error}
                        <div class="error">{error}</div>
                    {/if}
                    
                    {#if loading}
                        <div class="loading">{$_('common.loading')}</div>
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
        background: var(--tg-theme-bg-color);
        border-radius: 24px 24px 0 0;
        box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
        max-height: 90vh;
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

    :global([data-theme="dark"]) .dialog {
        background: var(--tg-theme-bg-color);
    }

    :global([data-theme="dark"]) .dialog * {
        color: white !important;
    }
</style> 
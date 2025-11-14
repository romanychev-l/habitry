<script lang="ts">
    import { _ } from 'svelte-i18n';
    import { isListView, displayScore } from '../../stores/view';
    import { habits } from '../../stores/habit';
import { user, balance } from '../../stores/user';
    import { createEventDispatcher } from 'svelte';
    import HabitActionsModal from '../modals/HabitActionsModal.svelte';
    import DeleteConfirmModal from '../modals/DeleteConfirmModal.svelte';
    import ArchiveConfirmModal from '../modals/ArchiveConfirmModal.svelte';
    import NewHabitModal from '../modals/NewHabitModal.svelte';
    import HabitFollowersModal from '../modals/HabitFollowersModal.svelte';
    import HabitLinkModal from '../modals/HabitLinkModal.svelte';
    import type { Habit as HabitType } from '../../types';
    import { onMount } from 'svelte';
    import { api } from '../../utils/api';
    import { popup, initData, hapticFeedback } from '@tma.js/sdk-svelte';
    import { gradients } from '../../utils/gradients'; // Import gradients
    import SolarUndoLeftRoundLinear from '../icons/SolarUndoLeftRoundLinear.svelte'; // Импортируем иконку
    import SolarMenuDotsBold from '../icons/SolarMenuDotsBold.svelte'; // Импортируем новую иконку меню

    const dispatch = createEventDispatcher();

    export let habit: HabitType & { progress: number };
    export let telegramId: number;
    export let readonly: boolean = false;
    export let closeModalsSignal: number = 0;
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    let clickTimeout: ReturnType<typeof setTimeout> | undefined;

    let showFollowersModal = false;
    let showLinkModal = false;
    let showActions = false;
    let showDeleteConfirm = false;
    let showEditModal = false;
    let showArchiveConfirm = false;

    $: isAnyModalOpen = showFollowersModal || showLinkModal || showActions || showDeleteConfirm || showEditModal || showArchiveConfirm;

    $: {
        if (isAnyModalOpen) {
            dispatch('modalOpened');
        } else {
            dispatch('modalClosed');
        }
    }

    // Закрытие всех модалок по внешнему сигналу (например, по нативной кнопке Назад)
    $: if (closeModalsSignal) {
        showFollowersModal = false;
        showLinkModal = false;
        showActions = false;
        showDeleteConfirm = false;
        showEditModal = false;
        showArchiveConfirm = false;
    }

    let pressStartTime: number;
    let startY: number;
    let isSwiping = false;
    let preloadedFollowers: Array<{ username: string; telegram_id: number }> = [];
    const API_URL = import.meta.env.VITE_API_URL;
    
    function handleClick(event: MouseEvent) {
        // Проверяем, не является ли цель клика кнопкой more или undo
        const target = event.target as HTMLElement;
        if (target.closest('.more-button') || target.closest('.more-list-view-button') || target.closest('.undo-button')) {
            return;
        }

        if (readonly) {
            if (telegramId === initData.user()?.id) {
                popup.show({
                    title: $_('alerts.warning'),
                    message: $_('habits.errors.follow_self'),
                    buttons: [{ id: 'close', type: 'close' }]
                });
                return;
            }
            showLinkModal = true;
        } else {
            showFollowersModal = true;
        }
    }

    function handlePointerDown(event: PointerEvent) {
        // Проверяем, не является ли цель клика кнопкой more или undo
        const target = event.target as HTMLElement;
        if (target.closest('.more-button') || target.closest('.more-list-view-button') || target.closest('.undo-button')) {
            return;
        }

        if (!readonly) {
            pressStartTime = Date.now();
            startY = event.clientY;
            isSwiping = false;
            isPressed = true;
            isPressTimeout = setTimeout(async () => {
                try {
                    if (hapticFeedback.impactOccurred.isAvailable()) {
                        hapticFeedback.impactOccurred('medium');
                    }
                    const data = await updateHabitOnServer();
                    if (data.habit && hapticFeedback.impactOccurred.isAvailable()) {
                        hapticFeedback.impactOccurred('medium');
                    }
                } catch (error) {
                    // Ошибка уже обработана в updateHabitOnServer
                } finally {
                    isPressed = false;
                }
            }, 800);
        }
    }

    function handlePointerMove(event: PointerEvent) {
        if (isPressed) {
            const deltaY = Math.abs(event.clientY - startY);
            if (deltaY > 10) {
                isSwiping = true;
                clearTimeout(isPressTimeout);
                isPressed = false;
            }
        }
    }

    function handlePointerUp(event: PointerEvent) {
        clearTimeout(isPressTimeout);
        isPressed = false;
    }

    let completed = false;
    
    // Функция для получения текущей даты с учетом часового пояса
    function getCurrentDate() {
        // Для тестирования - раскомментируйте нужную строку
        // return '2024-03-20'; // Вчера
        // return '2024-03-21'; // Сегодня
        // return '2024-03-22'; // Завтра
        
        const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
        const date = new Date();
        const localDate = new Date(date.toLocaleString('en-US', { timeZone: userTimezone }));
        
        const year = localDate.getFullYear();
        const month = String(localDate.getMonth() + 1).padStart(2, '0');
        const day = String(localDate.getDate()).padStart(2, '0');
        
        return `${year}-${month}-${day}`;
    }
    
    // Загружаем данные при монтировании компонента
    onMount(loadFollowers);

    // Обновляем состояние при изменении habit
    $: {
        if (habit) {
            const today = getCurrentDate();
            completed = habit.last_click_date === today;
            console.log('today', today);
            console.log('habit.last_click_date', habit.last_click_date);
            loadFollowers(); // Обновляем список подписчиков при изменении привычки
        }
    }
    
    let isAnimating = false;
    
    async function updateHabitOnServer() {
        try {
            console.log('Отправляем запрос на обновление привычки:', {
                _id: habit._id
            });
            
            const updatedHabit = await api.updateHabit({
                _id: habit._id
            });
            
            console.log('Получен ответ от сервера:', updatedHabit);
            // Оптимистично обновляем баланс пользователя (+1 WILL)
            balance.update(b => (b ?? 0) + 1);
            
            isAnimating = true;
            // Сбрасываем флаг анимации после её завершения
            setTimeout(() => {
                isAnimating = false;
            }, 800);
            
            console.log('Обновляем store habits. Текущее состояние:', $habits);
            
            // Обновляем store habits для пересортировки
            habits.update(currentHabits => {
                const updatedHabits = currentHabits.map(h => 
                    h._id === updatedHabit._id ? updatedHabit : h
                );
                console.log('Новое состояние store:', updatedHabits);
                return updatedHabits;
            });
            
            return updatedHabit;
        } catch (error) {
            console.error('Ошибка при обновлении привычки:', error);
            throw error;
        }
    }
    
    async function handleUndo() {
        try {
            const updatedHabit = await api.undoHabit({
                _id: habit._id
            });
            
            isAnimating = true;
            setTimeout(() => {
                isAnimating = false;
            }, 800);

            completed = false;

            // Оптимистично обновляем баланс пользователя (-1 WILL)
            balance.update(b => Math.max(0, (b ?? 0) - 1));

            console.log('handleUndo in HabitCard.svelte', updatedHabit);

            habits.update(currentHabits => {
                const updatedHabits = currentHabits.map(h => 
                    h._id === updatedHabit._id ? updatedHabit : h
                );
                return updatedHabits;
            });
        } catch (error) {
            console.error('Error undoing habit click:', error);
            await popup.show({
                title: 'Ошибка',
                message: $_('habits.errors.undo'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        }
    }
    
    async function handleDelete() {
        try {
            const data = await api.deleteHabit({
                telegram_id: telegramId,
                habit_id: habit._id
            });
            
            // Обновляем store вместо перезагрузки страницы
            habits.update(currentHabits => 
                currentHabits.filter(h => h._id !== habit._id)
            );
            showDeleteConfirm = false;
        } catch (error) {
            console.error('Error deleting habit:', error);
            await popup.show({
                title: 'Ошибка',
                message: $_('habits.errors.delete'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        }
    }

    // Helper function to calculate a hash from a string
    function simpleHash(str: string): number {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            const char = str.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash; // Convert to 32bit integer
        }
        return Math.abs(hash);
    }

    // Select a gradient based on the habit ID hash
    $: gradientIndex = simpleHash(habit._id) % gradients.length;
    $: gradientStyle = gradients[gradientIndex];

    async function handleEdit(event: CustomEvent) {
        try {
            console.log('EditHabit event.detail:', event.detail);
            console.log('Original habit:', habit);
            
            const habitData = {
                _id: habit._id,
                title: event.detail.title,
                want_to_become: event.detail.want_to_become,
                days: event.detail.days,
                is_one_time: event.detail.is_one_time,
                is_auto: event.detail.is_auto,
                stake: event.detail.stake,
                created_at: habit.created_at,
                last_click_date: habit.last_click_date,
                streak: habit.streak,
                score: habit.score
            };
            
            console.log('HabitData being sent to server:', habitData);

            const updatedHabit = await api.editHabit(habitData);
            console.log('Response from server:', updatedHabit);
            
            habits.update(currentHabits => 
                currentHabits.map(h => h._id === habit._id ? updatedHabit : h)
            );
            
            showEditModal = false;
        } catch (error) {
            if (error instanceof Error && error.message.includes('403')) {
                alert($_('habits.errors.edit_forbidden'));
                return;
            }
            console.error('Error:', error);
            alert($_('habits.errors.update'));
        }
    }

    async function loadFollowers() {
        try {
            const data = await api.getHabitFollowers(habit._id);
            preloadedFollowers = data;
        } catch (error) {
            console.error('Error loading followers:', error);
        }
    }

    function handleFollowersUpdated(event: CustomEvent) {
        preloadedFollowers = event.detail.followers;
    }

    async function handleHabitLink(event: CustomEvent<{habitId: string; sharedHabitId: string; sharedByTelegramId: string}>) {
        console.log('handleHabitLink in HabitCard.svelte', event.detail);
        try {
            const currentUserId = initData.user()?.id;
            if (!currentUserId) {
                throw new Error($_('habits.errors.link'));
            }

            const data = await api.joinHabit({
                telegram_id: currentUserId,
                habit_id: event.detail.habitId,
                shared_by_telegram_id: event.detail.sharedByTelegramId,
                shared_by_habit_id: event.detail.sharedHabitId
            });

            habits.update(currentHabits => data.habits || []);

            showLinkModal = false;
            await popup.show({
                title: 'Успех',
                message: $_('habits.link_success'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        } catch (error) {
            console.error('Error:', error);
            await popup.show({
                title: 'Ошибка',
                message: $_('habits.errors.link'),
                buttons: [{ id: 'close', type: 'close' }]
            });
        }
    }
</script>
  
<div class="habit-wrapper" style="--habit-gradient: {gradientStyle}; --progress: {habit.progress}">
  <div class="card-shadow">
    <div
      class="habit-card"
      class:pressed={isPressed}
      class:animating={isAnimating}
      class:list-view={$isListView}
      class:readonly={readonly}
      role="button"
      tabindex="0"
      on:click={handleClick}
      on:keydown={e => {
        if (e.key === 'Enter') {
          handleClick(new MouseEvent('click'));
        }
      }}
      on:pointerdown={handlePointerDown}
      on:pointermove={readonly ? null : handlePointerMove}
      on:pointerup={handlePointerUp}
      on:pointercancel={handlePointerUp}
      on:pointerleave={handlePointerUp}
      style="--habit-gradient: {gradientStyle}">
      <div class="content">
        <h3>{habit.title}</h3>
        
        {#if !$isListView && habit.want_to_become}
          <div class="want-to-become">
            <span class="label">{$_('habits.want_to_become')}</span>
            <span class="value">{habit.want_to_become}</span>
          </div>
        {/if}

        {#if completed && !readonly}
          <button 
            class="undo-button" 
            on:click|stopPropagation={handleUndo}
            style="bottom: 40px;"
          ><SolarUndoLeftRoundLinear style="width: 0.8em; height: 0.8em;" /></button>
        {/if}
      </div>

      {#if !readonly}
        <button 
          class={!$isListView ? 'more-button' : 'more-list-view-button'}
          on:click={() => showActions = true}
          style="top: 35px;"
        >
          <SolarMenuDotsBold style="width: 0.8em; height: 0.8em;" />
        </button>
      {/if}
    </div>
  </div>
  <div class="streak-shadow">
    <div class="streak-counter" style="--habit-gradient: {gradientStyle}">
      {$displayScore ? (habit.score || 0) : (habit.streak || 0)}
    </div>
  </div>
</div>

{#if showLinkModal}
  <HabitLinkModal
    habits={$habits}
    sharedHabitId={habit._id}
    sharedByTelegramId={telegramId.toString()}
    on:close={() => showLinkModal = false}
    on:select={handleHabitLink}
  />
{/if}

{#if showActions && !readonly}
  <HabitActionsModal 
    habit={habit}
    on:close={() => showActions = false}
    on:showDeleteConfirm={() => {
      showDeleteConfirm = true;
    }}
    on:showEditModal={() => {
      showEditModal = true;
    }}
    on:showArchiveConfirm={() => {
      showArchiveConfirm = true;
    }}
  />
{/if}

{#if showDeleteConfirm && !readonly}
  <DeleteConfirmModal 
    on:close={() => {
      showDeleteConfirm = false;
      showActions = false;
    }}
    on:delete={handleDelete}
  />
{/if}

{#if showArchiveConfirm && !readonly}
  <ArchiveConfirmModal 
    on:close={() => {
      showArchiveConfirm = false;
      showActions = false;
    }}
    on:archive={async () => {
      try {
        const updatedHabit = await api.archiveHabit({ _id: habit._id });
        // Удаляем из текущего списка (архивные не должны показываться)
        habits.update(current => current.filter(h => h._id !== habit._id));
        showArchiveConfirm = false;
        showActions = false;
      } catch (e) {
        console.error('Error archiving habit', e);
      }
    }}
  />
{/if}

{#if showEditModal && !readonly}
  <NewHabitModal
    habit={habit}
    on:close={() => {
      showEditModal = false;
      showActions = false;
    }}
    on:save={handleEdit}
  />
{/if}

{#if showFollowersModal && !readonly}
  <HabitFollowersModal
    show={showFollowersModal}
    habit={habit}
    telegramId={telegramId}
    on:close={() => showFollowersModal = false}
    on:followersUpdated={handleFollowersUpdated}
  />
{/if}

<style>
  .habit-wrapper {
    position: relative;
    width: 300px;
    aspect-ratio: 1;
    margin: 0 auto;
  }

  :global(.list-view) .habit-wrapper {
    width: 100%;
    aspect-ratio: unset;
    min-height: 70px;
    height: auto;
    margin: 8px auto;
    max-width: 800px;
    padding: 0 8px;
    box-sizing: border-box;
  }

  .card-shadow {
    width: calc(100% + 20px);
    height: calc(100% + 20px);
    margin: -10px;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.08));
  }

  .card-shadow:has(.habit-card.pressed) {
    filter: drop-shadow(0 4px 12px rgba(139, 92, 246, 0.3));
  }

  .habit-card {
    width: 100%;
    height: 100%;
    padding: 32px;
    position: relative;
    transition: background 0.8s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    user-select: none;
    -webkit-user-select: none;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    background: white;
    font-family: 'Manrope', sans-serif;
    transform: translateZ(0);
    backface-visibility: hidden;
  }

  .habit-card::before {
    content: '';
    position: absolute;
    left: 0;
    bottom: 0;
    width: 100%;
    height: calc(var(--progress) * 100%);
    background: var(--habit-gradient);
    transition: none;
    z-index: 0;
  }

  .habit-card.animating::before {
    transition: height 0.8s ease;
  }

  :global(.list-view) .habit-card {
    border-radius: 16px;
    padding: 12px 16px;
    mask: none !important;
    -webkit-mask: none !important;
    min-height: 85px;
    text-align: left;
    position: relative;
    isolation: isolate;
  }

  :global(.list-view) .habit-card::before {
    width: calc(var(--progress) * 100%);
    height: 100%;
    transition: none;
    z-index: 0;
    border-radius: inherit;
  }

  :global(.list-view) .habit-card.animating::before {
    transition: width 0.8s ease;
  }

  .streak-shadow {
    position: absolute;
    top: 15px;
    right: 5px;
    width: 60px;
    height: 60px;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.08));
    z-index: 1;
  }

  :global(.list-view) .streak-shadow {
    left: 14px;
    top: 3px;
    right: auto;
    /* top: 50%; */
    /* transform: translateY(-50%); */
    filter: none;
  }

  .streak-counter {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: var(--habit-gradient);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 24px;
    padding: 0;
    line-height: 0; /* Убираем влияние line-height на SVG */
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  :global(.list-view) .streak-counter {
    width: 60px;
    height: 60px;
  }

  /* Изменяем цвет счетчика стрика при полном выполнении */
  .habit-wrapper[style*="--progress: 1"] .streak-counter {
    background: white;
    color: black;
  }

  .habit-card h3 {
    position: relative;
    z-index: 1;
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #333;
    /* Ограничение двумя строками */
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    /* Отступы слева и справа */
    padding: 0 10px; 
    word-break: break-word; /* Перенос слов */
    line-height: 1.5; /* Немного увеличим межстрочный интервал */
  }

  :global(.list-view) .habit-card h3 {
    margin-right: 50px;
    margin-left: 65px;
    line-height: 1.2;
  }

  :global(.list-view) .habit-card[style*="--progress: 1"] h3 {
    color: white;
  }

  .undo-button {
    position: absolute;
    color: #333;
    bottom: 25px;
    left: 50%;
    transform: translateX(-50%);
    background: none;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    font-size: 28px;
    cursor: pointer;
    opacity: 0.8;
    z-index: 3;
  }

  :global(.list-view) .undo-button {
    position: absolute;
    top: 50%;
    right: 50px;
    left: auto;
    transform: translateY(-50%);
    font-size: 22px; 
  }

  .undo-button:hover {
    opacity: 1;
  }

  .want-to-become {
    margin-top: 16px;
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 8px;
    color: #333;
    position: relative;
    z-index: 1;
  }

  .want-to-become .label {
    font-size: 12px;
    opacity: 0.7;
    font-weight: 400;
  }

  .want-to-become .value {
    font-size: 20px;
    font-weight: 600;
    /* Ограничение двумя строками */
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    /* Отступы слева и справа */
    padding: 0 10px;
    word-break: break-word; /* Перенос слов */
    line-height: 1.5; /* Немного увеличим межстрочный интервал */
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  :global(.list-view) .content {
    width: 100%;
    text-align: left;
  }

  :global(.list-view) .want-to-become {
    display: none;
  }

  .more-button {
    position: absolute;
    top: 25px;
    left: 50%;
    transform: translateX(-50%);
    background: none;
    border: none;
    font-size: 30px;
    padding: 8px;
    cursor: pointer;
    opacity: 0.8;
    z-index: 3;
    color: #333;
  }

  .more-list-view-button {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    margin-top: 8px; /* Коррекция для визуального центрирования */
    right: 16px;
    background: none;
    border: none;
    font-size: 24px;
    padding: 8px;
    cursor: pointer;
    opacity: 0.8;
    z-index: 3;
    color: #333;
  }

  .more-button:hover,
  .more-list-view-button:hover {
    opacity: 1;
  }

  :global(.list-view) .habit-card[style*="--progress: 1"] .more-list-view-button {
    color: white;
  }

  .habit-card.readonly {
    cursor: pointer;
  }

  .habit-card.readonly:active {
    opacity: 0.8;
  }
</style>
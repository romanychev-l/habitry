<script lang="ts">
    import { _ } from 'svelte-i18n';
    import { isListView } from '../stores/view';
    import { habits } from '../stores/habit';
    import HabitActionsModal from './HabitActionsModal.svelte';
    import DeleteConfirmModal from './DeleteConfirmModal.svelte';
    import NewHabitModal from './NewHabitModal.svelte';
    import HabitFollowersModal from './HabitFollowersModal.svelte';
    import HabitLinkModal from './HabitLinkModal.svelte';
    import type { Habit } from '../types';
    import { onMount } from 'svelte';
    import { api } from '../utils/api';
    
    export let habit: Habit;
    export let telegramId: number;
    export let readonly: boolean = false;
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    let clickTimeout: ReturnType<typeof setTimeout> | undefined;
    let showFollowersModal = false;
    let showLinkModal = false;
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
                    if (navigator.vibrate) {
                        navigator.vibrate([50]);
                    }
                    const data = await updateHabitOnServer();
                    if (data.habit && navigator.vibrate) {
                        navigator.vibrate(50);
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

    let progress = 0;
    let completed = false;
    
    async function updateProgress() {
        progress = await calculateProgress();
    }
    
    // Загружаем данные при монтировании компонента
    onMount(loadFollowers);

    // Обновляем состояние при изменении habit
    $: {
        if (habit) {
            const today = new Date().toISOString().split('T')[0];
            completed = habit.last_click_date === today;
            updateProgress();
            loadFollowers(); // Обновляем список подписчиков при изменении привычки
        }
    }
    
    let isAnimating = false;
    
    async function updateHabitOnServer() {
        try {
            console.log('Отправляем запрос на обновление привычки:', {
                telegram_id: telegramId,
                habit: {
                    _id: habit._id
                }
            });
            
            const data = await api.updateHabit({
                telegram_id: telegramId,
                habit: {
                    _id: habit._id
                }
            });
            
            console.log('Получен ответ от сервера:', data);
            
            if (data.habit) {
                isAnimating = true;
                // Сбрасываем флаг анимации после её завершения
                setTimeout(() => {
                    isAnimating = false;
                }, 800);
                
                console.log('Обновляем store habits. Текущее состояние:', $habits);
                
                // Обновляем store habits для пересортировки
                habits.update(currentHabits => {
                    const updatedHabits = currentHabits.map(h => 
                        h._id === data.habit._id ? data.habit : h
                    );
                    console.log('Новое состояние store:', updatedHabits);
                    return updatedHabits;
                });
                
                // После обновления store пересчитываем прогресс
                await updateProgress();
            }
            return data;
        } catch (error) {
            console.error('Ошибка при обновлении привычки:', error);
            throw error;
        }
    }
    
    async function handleUndo() {
      console.log('handleUndo in HabitCard.svelte');
        try {
            const data = await api.undoHabit({
                telegram_id: telegramId,
                habit: {
                    _id: habit._id
                }
            });
            
            if (data.habit) {
                isAnimating = true;
                // Сбрасываем флаг анимации после её завершения
                setTimeout(() => {
                    isAnimating = false;
                }, 800);

                // Обновляем локальное состояние
                completed = false;

                console.log('handleUndo in HabitCard.svelte', data);

                // Обновляем store habits для пересортировки
                habits.update(currentHabits => {
                    const updatedHabits = currentHabits.map(h => 
                        h._id === data.habit._id ? data.habit : h
                    );
                    return updatedHabits;
                });
                
                // После обновления store пересчитываем прогресс
                await updateProgress();
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert($_('habits.errors.undo'));
        }
    }
    
    async function handleDelete() {
        try {
            const data = await api.deleteHabit({
                telegram_id: telegramId,
                habit_id: habit._id
            });
            
            // Перезагружаем страницу после успешного удаления
            window.location.reload();
        } catch (error) {
            if (error instanceof Error && error.message.includes('403')) {
                alert($_('habits.errors.delete_forbidden'));
                return;
            }
            console.error('Error:', error);
            alert($_('habits.errors.delete'));
        }
    }

    // Функция для генерации цвета на основе строки
    function stringToColor(str: string): string {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = str.charCodeAt(i) + ((hash << 5) - hash);
        }
        const h = Math.abs(hash % 360);
        return `hsl(${h}, 70%, 60%)`; // Используем HSL для сохранения яркости
    }

    // Получаем два цвета для градиента и мемоизируем их
    $: color1 = stringToColor(habit._id);
    $: color2 = stringToColor(habit._id.split('').reverse().join(''));
    $: gradientStyle = `linear-gradient(135deg, ${color1} 0%, ${color2} 100%)`;

    let showActions = false;
    let showDeleteConfirm = false;
    let showEditModal = false;

    // Добавляем функцию подсчета прогресса
    async function calculateProgress(): Promise<number> {
        console.log('calculateProgress', completed);
        
        try {
            const data = await api.getHabitProgress(habit._id, telegramId);
            console.log('Progress data:', data);
            return data.progress;
        } catch (error) {
            console.error('Error fetching progress:', error);
            return 0;
        }
    }

    async function handleEdit(event: CustomEvent) {
        try {
            const habitData = {
                telegram_id: telegramId,
                habit: {
                    ...habit,
                    title: event.detail.title,
                    want_to_become: event.detail.want_to_become,
                    days: event.detail.days,
                    is_one_time: event.detail.is_one_time,
                    is_auto: event.detail.is_auto,
                    stake: event.detail.stake
                }
            };

            const data = await api.editHabit(habitData);
            if (data.habit) {
                habits.update(currentHabits => 
                    currentHabits.map(h => h._id === habit._id ? data.habit : h)
                );
            }
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
            const data = await api.getHabitFollowers(habit._id, telegramId);
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
            const currentUserId = window.Telegram?.WebApp?.initDataUnsafe?.user?.id;
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
            window.Telegram?.WebApp?.showAlert($_('habits.link_success'));
        } catch (error) {
            console.error('Error:', error);
            window.Telegram?.WebApp?.showAlert($_('habits.errors.link'));
        }
    }
</script>
  
<div class="habit-wrapper" style="--habit-gradient: {gradientStyle}; --progress: {progress}">
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
          >&larr;</button>
        {/if}
      </div>

      {#if !readonly}
        <button 
          class={!$isListView ? 'more-button' : 'more-list-view-button'}
          on:click={() => showActions = true}
        >
          {!$isListView ? '…' : '⋮'}
        </button>
      {/if}
    </div>
  </div>
  <div class="streak-counter" style="--habit-gradient: {gradientStyle}">
    {habit.streak || 0}
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
      showActions = false;
      showDeleteConfirm = true;
    }}
    on:showEditModal={() => {
      showActions = false;
      showEditModal = true;
    }}
  />
{/if}

{#if showDeleteConfirm && !readonly}
  <DeleteConfirmModal 
    on:close={() => showDeleteConfirm = false}
    on:delete={handleDelete}
  />
{/if}

{#if showEditModal && !readonly}
  <NewHabitModal
    habit={habit}
    on:close={() => showEditModal = false}
    on:save={handleEdit}
  />
{/if}

{#if showFollowersModal && !readonly}
  <HabitFollowersModal
    show={showFollowersModal}
    habit={habit}
    telegramId={telegramId}
    initialFollowers={preloadedFollowers}
    on:close={() => showFollowersModal = false}
    on:followersUpdated={handleFollowersUpdated}
  />
{/if}

<style>
  .habit-wrapper {
    position: relative;
    width: 280px;
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
    width: 100%;
    height: 100%;
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

  .streak-counter {
    position: absolute;
    top: 5px;
    right: -5px;
    width: 60px;
    height: 60px;
    background: var(--habit-gradient);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 24px;
    z-index: 1;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  :global(.list-view) .streak-counter {
    left: 20px;
    right: auto;
    top: 50%;
    transform: translateY(-50%);
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
    font-weight: 700;
    color: #333;
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
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    background: none;
    border: none;
    font-size: 24px;
    padding: 8px;
    cursor: pointer;
    opacity: 0.8;
    z-index: 3;
  }

  :global(.list-view) .undo-button {
    position: absolute;
    top: 50%;
    right: 30px;
    left: auto;
    transform: translateY(-50%);
    font-size: 20px;
    margin: 0;
    padding: 8px;
    height: 24px;
    line-height: 24px;
    display: flex;
    align-items: center;
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
  }

  .want-to-become .value {
    font-size: 20px;
    font-weight: 700;
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
    top: 16px;
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
    right: 8px;
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
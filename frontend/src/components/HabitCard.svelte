<script lang="ts">
    import { _ } from 'svelte-i18n';
    import { isListView } from '../stores/view';
    import HabitActionsModal from './HabitActionsModal.svelte';
    import DeleteConfirmModal from './DeleteConfirmModal.svelte';
    import type { HabitWithStats } from '../types';
    
    export let habitWithStats: HabitWithStats;
    export let telegramId: number;
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    const API_URL = import.meta.env.API_URL;
    
    function handlePointerDown() {
        if (navigator.vibrate) {
            navigator.vibrate([100, 30, 100]);
        }
        
        isPressed = true;
        isPressTimeout = setTimeout(async () => {
            try {
                const data = await updateHabitOnServer();
                if (data.habit) {
                    habitWithStats = data.habit;
                    if (navigator.vibrate) {
                        navigator.vibrate(200);
                    }
                }
            } catch (error) {
                // Ошибка уже обработана в updateHabitOnServer
            } finally {
                isPressed = false;
            }
        }, 800);
    }

    function handlePointerUp() {
        clearTimeout(isPressTimeout);
        isPressed = false;
    }

    let progress = 0;
    let completed = false;
    
    async function updateProgress() {
        progress = await calculateProgress();
    }

    // Обновляем completed и прогресс при изменении habitWithStats
    $: {
        if (habitWithStats) {
            completed = isCompletedToday();
            console.log('completed', completed);
            updateProgress();
        }
    }
    
    function isCompletedToday(): boolean {
        if (!habitWithStats.last_click_date) return false;
        
        const today = new Date().toISOString().split('T')[0];
        const lastClick = habitWithStats.last_click_date.split('T')[0];
        
        return lastClick === today;
    }
    
    async function updateHabitOnServer() {
        try {
            const response = await fetch(`${API_URL}/habit/update`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegram_id: telegramId,
                    habit: {
                        _id: habitWithStats.habit._id
                    }
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.update'));
            }
            
            const data = await response.json();
            console.log('Update response:', data);
            
            if (data.habit) {
                habitWithStats = data.habit;
                console.log('HabitWithStats:', habitWithStats);
                // Обновляем прогресс после обновления привычки
                progress = await calculateProgress();
                console.log('progress', progress);
                
                if (navigator.vibrate) {
                    navigator.vibrate(200);
                }
            }
            
            return data;
        } catch (error) {
            console.error('Ошибка:', error);
            throw error;
        }
    }
    
    async function handleUndo() {
        try {
            const response = await fetch(`${API_URL}/habit/undo`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegram_id: telegramId,
                    habit: {
                        _id: habitWithStats.habit._id
                    }
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.undo'));
            }
            
            const data = await response.json();
            if (data.habit) {
                habitWithStats = { ...data.habit }; // Создаем новый объект для гарантированного обновления
                progress = 0; // Сразу сбрасываем прогресс
            }
        } catch (error) {
            console.error('Ошибка:', error);
        }
    }
    
    async function handleDelete() {
        try {
            const response = await fetch(`${API_URL}/habit/delete`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegram_id: telegramId,
                    habit_id: habitWithStats.habit._id
                })
            });
            
            if (!response.ok) {
                if (response.status === 403) {
                    alert($_('habits.errors.delete_forbidden'));
                    return;
                }
                throw new Error($_('habits.errors.delete'));
            }
            
            // Перезагружаем страницу после успешного удаления
            window.location.reload();
        } catch (error) {
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

    // Получаем два цвета для градиента
    const color1 = stringToColor(habitWithStats.habit._id);
    const color2 = stringToColor(habitWithStats.habit._id.split('').reverse().join(''));

    // Создаем строку градиента
    const gradientStyle = `linear-gradient(135deg, ${color1} 0%, ${color2} 100%)`;

    let showActions = false;
    let showDeleteConfirm = false;

    // Добавляем функцию подсчета прогресса
    async function calculateProgress(): Promise<number> {
        console.log('calculateProgress', completed);
        
        try {
            const response = await fetch(`${API_URL}/habit/progress?habit_id=${habitWithStats.habit._id}&telegram_id=${telegramId}`);
            if (!response.ok) {
                throw new Error('Failed to fetch progress');
            }
            
            const data = await response.json();
            console.log('Progress data:', data);
            return data.progress;
        } catch (error) {
            console.error('Error fetching progress:', error);
            return 0;
        }
    }
</script>
  
<div class="habit-wrapper" style="--habit-gradient: {gradientStyle}; --progress: {progress}">
  <div class="card-shadow">
    <div class="habit-card"
      class:pressed={isPressed}
      on:pointerdown={handlePointerDown}
      on:pointerup={handlePointerUp}
      on:pointerleave={handlePointerUp}>
      <div class="content">
        <h3>{habitWithStats.habit.title}</h3>
        
        {#if !$isListView && habitWithStats.habit.want_to_become}
          <div class="want-to-become">
            <span class="label">{$_('habits.want_to_become')}</span>
            <span class="value">{habitWithStats.habit.want_to_become}</span>
          </div>
        {/if}

        {#if completed}
          <button class="undo-button" on:click={handleUndo}>↩</button>
        {/if}
      </div>

      <button 
        class={!$isListView ? 'more-button' : 'more-list-view-button'}
        on:click={() => showActions = true}
      >
        {!$isListView ? '…' : '⋮'}
      </button>
    </div>
  </div>
  <div class="streak-counter" style="--progress: {progress}">
    {habitWithStats.streak || 0}
  </div>
</div>

{#if showActions}
  <HabitActionsModal 
    habit={habitWithStats.habit}
    on:close={() => showActions = false}
    on:showDeleteConfirm={() => {
      showActions = false;  // Закрываем окно действий
      showDeleteConfirm = true;  // Показываем окно подтверждения
    }}
  />
{/if}

{#if showDeleteConfirm}
  <DeleteConfirmModal 
    on:close={() => showDeleteConfirm = false}
    on:delete={handleDelete}
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
    transition: height 0.8s ease;
    z-index: -1;
  }

  :global(.list-view) .habit-card {
    border-radius: 16px;
    padding: 12px 16px;
    mask: none !important;
    -webkit-mask: none !important;
    min-height: 85px;
    text-align: left;
    overflow: hidden;
  }

  :global(.list-view) .habit-card::before {
    width: calc(var(--progress) * 100%);
    height: 100%;
    transition: width 0.8s ease;
    z-index: 0;
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

  .streak-counter[style*="--progress: 1"]{
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
    bottom: 47%;
    right: 30px;
    left: auto;
    transform: translateY(50%);
    font-size: 20px;
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
</style>
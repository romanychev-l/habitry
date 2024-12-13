<script lang="ts">
    import { _ } from 'svelte-i18n';
    import { isListView } from '../stores/view';
    
    export let habit: {
        id: string;
        title: string;
        streak: number;
        score: number;
        last_click_date?: string | null;
        want_to_become?: string;
    };
    
    export let telegramId: number;
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    const API_URL = import.meta.env.VITE_API_URL;
    // Делаем переменную реактивной с помощью $:
    $: completed = isCompletedToday();
    
    function isCompletedToday(): boolean {
        if (!habit.last_click_date) return false;
        
        const now = new Date();
        const todayStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
        const lastClick = habit.last_click_date.split('T')[0];
        
        return lastClick === todayStr;
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
                        id: habit.id,
                        title: habit.title
                    }
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.update'));
            }
            
            const data = await response.json();
            if (data.habit) {
                habit = { ...data.habit };
            }
            return data;
        } catch (error) {
            console.error('Ошибка:', error);
            throw error;
        }
    }
    
    function handlePointerDown() {
        if (navigator.vibrate) {
            navigator.vibrate([100, 30, 100]);
        }
        
        isPressed = true;
        isPressTimeout = setTimeout(async () => {
            try {
                await updateHabitOnServer();
                
                completed = isCompletedToday();
                
                if (navigator.vibrate) {
                    navigator.vibrate(200);
                }
            } catch (error) {
                // Ошибка уже обработа в updateHabitOnServer
            } finally {
                isPressed = false;
            }
        }, 800);
    }
    
    function handlePointerUp() {
        clearTimeout(isPressTimeout);
        isPressed = false;
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
                        id: habit.id,
                        title: habit.title
                    }
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.undo'));
            }
            
            const data = await response.json();
            if (data.habit) {
                habit = { ...data.habit };
                habit.last_click_date = null;
                completed = false;
            }
        } catch (error) {
            console.error('Ошибка:', error);
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
    const color1 = stringToColor(habit.id);
    const color2 = stringToColor(habit.id.split('').reverse().join(''));

    // Создаем строку градиента
    const gradientStyle = `linear-gradient(135deg, ${color1} 0%, ${color2} 100%)`;

    let showActions = false;
    let showDeleteConfirm = false;
    
    async function handleDelete() {
        try {
            const response = await fetch(`${API_URL}/habit/delete`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegram_id: telegramId,
                    habit_id: habit.id
                })
            });
            
            if (!response.ok) {
                throw new Error($_('habits.errors.delete'));
            }
            
            // Перезагружаем страницу после успешного удаления
            window.location.reload();
        } catch (error) {
            console.error('Error:', error);
            alert($_('habits.errors.delete'));
        }
    }
</script>
  
<div class="habit-wrapper" style="--habit-gradient: {gradientStyle}">
  <div class="card-shadow">
    <div class="habit-card"
      class:pressed={isPressed}
      class:completed={completed}
      on:pointerdown={handlePointerDown}
      on:pointerup={handlePointerUp}
      on:pointerleave={handlePointerUp}>
      <div class="content">
        <h3>{habit.title}</h3>
        
        {#if !$isListView && habit.want_to_become}
          <div class="want-to-become">
            <span class="label">{$_('habits.want_to_become')}</span>
            <span class="value">{habit.want_to_become}</span>
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
  <div class="streak-counter">
    {habit.streak}
  </div>
</div>

{#if showActions}
  <div 
    class="dialog-overlay" 
    on:click|stopPropagation={() => showActions = false}
    on:keydown={(e) => e.key === 'Escape' && (showActions = false)}
    role="button"
    tabindex="0"
  >
    <div class="dialog">
      <div class="dialog-header">
        <h2>{habit.title}</h2>
      </div>
      <div class="dialog-content">
        <button 
          class="dialog-button delete"
          on:click={() => {
            showActions = false;
            showDeleteConfirm = true;
          }}
        >
          {$_('habits.delete')}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDeleteConfirm}
  <div 
    class="dialog-overlay" 
    on:click|stopPropagation={() => showDeleteConfirm = false}
    on:keydown={(e) => e.key === 'Escape' && (showDeleteConfirm = false)}
    role="button"
    tabindex="0"
  >
    <div class="dialog">
      <div class="dialog-header">
        <h2>{$_('habits.modals.delete_title')}</h2>
      </div>
      <div class="dialog-content">
        <p>{$_('habits.modals.delete_message')}</p>
        <div class="dialog-buttons">
          <button 
            class="dialog-button cancel"
            on:click={() => showDeleteConfirm = false}
          >
            {$_('habits.cancel')}
          </button>
          <button 
            class="dialog-button delete"
            on:click={handleDelete}
          >
            {$_('habits.delete')}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .habit-wrapper {
    position: relative;
    width: 280px;
    aspect-ratio: 1;
    margin: 0 auto;
  }

  /* Обновляем стили для режима списка */
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

  /* Общие стили для card-shadow */
  .card-shadow {
    width: 100%;
    height: 100%;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.08));
  }

  .card-shadow:has(.habit-card.pressed),
  .card-shadow:has(.habit-card.completed) {
    filter: drop-shadow(0 4px 12px rgba(139, 92, 246, 0.3));
  }

  /* Убираем специальные стили теней для списка */
  :global(.list-view) .card-shadow {
    width: 100%;
    height: 100%;
  }

  /* Убираем дополнительные тени */
  .habit-card.pressed,
  .habit-card.completed {
    background: var(--habit-gradient);
  }

  .streak-counter {
    position: absolute;
    top: 5px;
    right: -5px;
    width: 60px;
    height: 60px;
    /* По умолчанию фиолетовый градиент */
    background: var(--habit-gradient);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 24px;
    box-shadow: none;
    z-index: 1;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  /* Изменяем положение streak в режиме списка */
  :global(.list-view) .streak-counter {
    position: absolute;
    left: 20px;
    top: 50%;
    transform: translateY(-50%);
    width: 60px;
    height: 60px;
    background: var(--habit-gradient);
    color: white;
    z-index: 2;
  }

  /* Если привычка выполнена - streak белый */
  .habit-wrapper:has(.habit-card.completed) .streak-counter {
    background: white;
    color: #8B5CF6;
  }

  .habit-card {
    width: 100%;
    height: 100%;
    background: white;
    padding: 32px;
    position: relative;
    transition: background 0.8s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    user-select: none;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
  }

  /* Обновляем стили карточки для режима списка */
  :global(.list-view) .habit-card {
    border-radius: 16px;
    padding: 12px 16px;
    mask: none !important;
    -webkit-mask: none !important;
    width: 100%;
    min-height: 85px;
    height: auto;
    background: white;
    color: #333;
    text-align: left;
    position: relative;
  }

  .habit-card.pressed,
  .habit-card.completed {
    background: var(--habit-gradient);
  }

  .habit-card.pressed h3,
  .habit-card.completed h3 {
    color: white;
  }

  h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
    color: #333;
  }

  /* Уменьшаем разер заголовка в режиме списка */
  :global(.list-view) h3 {
    font-size: 20px;
    white-space: normal;
    overflow: visible;
    text-overflow: unset;
    margin-right: 50px;
    margin-left: 65px;
    line-height: 1.2;
    color: #333;
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

  /* Обновляем стили для кнопки отмены в режиме списка */
  :global(.list-view) .undo-button {
    position: absolute;
    bottom: 47%;
    right: 30px;
    left: auto;
    transform: translateY(50%);
    font-size: 20px;
    padding: 8px;
    z-index: 3;
    color: inherit;
  }

  /* Убираем все hover и active эффекты */
  .undo-button:hover {
    opacity: 1;
  }

  /* Добавляем контейнер дя списка */
  :global(.list-view) {
    overflow-x: hidden;
    width: 100%;
    padding: 0 4px;
    display: flex;
    flex-direction: column;
  }

  /* Убираем тен для card-shadow в режиме списка */
  :global(.list-view) .card-shadow {
    width: 100%;
    height: 100%;
  }

  :global(.list-view) .card-shadow:has(.habit-card.completed) {
    filter: drop-shadow(0 4px 12px rgba(139, 92, 246, 0.3));
  }

  :global(.list-view) .card-shadow:has(.habit-card.pressed) {
    filter: drop-shadow(0 4px 12px rgba(139, 92, 246, 0.3));
  }

  /* Добавляем стили для completed состояния в режиме списка */
  :global(.list-view) .habit-card.completed {
    background: var(--habit-gradient);
    color: white;
  }

  :global(.list-view) .habit-card.completed .undo-button {
    color: white;
  }

  /* Убираем белый фон streak для completed состояния в режиме списка */
  :global(.list-view) .habit-wrapper:has(.habit-card.completed) .streak-counter {
    background: white;
    color: var(--habit-gradient);
  }

  /* Обновляем цвет текста для режима списка */
  :global(.list-view) .habit-card h3 {
    color: #333;
  }

  /* Обновляем стили для completed состояния */
  :global(.list-view) .habit-card.completed {
    background: var(--habit-gradient);
    color: white;
  }

  /* Обновляем streak для completed состояния */
  :global(.list-view) .habit-wrapper:has(.habit-card.completed) .streak-counter {
    background: white;
    color: #8B5CF6;
  }

  /* Обновляем цвет текста */
  :global(.list-view) .habit-card h3 {
    color: #333;
  }

  :global(.list-view) .habit-card.completed h3 {
    color: white;
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

  .want-to-become .label {
    font-size: 12px;
    opacity: 0.7;
  }

  .want-to-become .value {
    font-size: 20px;
    font-weight: 700;
  }

  :global(.list-view) .content {
    padding-left: 0;
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
    font-size: 24px;
    padding: 8px;
    cursor: pointer;
    opacity: 0.8;
    z-index: 3;
  }

  .more-list-view-button {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    right: 8px;
    background: none;
    border: none;
    color: black;
    font-size: 24px;
    padding: 8px;
    cursor: pointer;
    z-index: 3;
  }

  .hidden {
    display: none !important;
  }

  :global(.list-view) .habit-card.completed .more-list-view-button {
    color: white;
  }

  :global([data-theme="dark"]) .more-list-view-button {
    color: white;
  }

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

  @supports (-webkit-touch-callout: none) {
    .dialog-overlay {
      position: absolute;
      height: 100vh;
      min-height: -webkit-fill-available;
    }
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

  .dialog-button {
    width: 100%;
    padding: 14px;
    border-radius: 12px;
    border: none;
    font-size: 16px;
    font-weight: 500;
    text-align: center;
    margin-bottom: 12px;
  }

  .dialog-buttons {
    display: flex;
    gap: 12px;
    margin-top: 24px;
  }

  .dialog-button.cancel {
    background: var(--tg-theme-secondary-bg-color);
    color: var(--tg-theme-text-color);
  }

  .dialog-button.delete {
    background: #ff3b30;
    color: white;
  }

  :global([data-theme="dark"]) .dialog {
    background: var(--tg-theme-bg-color);
  }

  :global([data-theme="dark"]) .dialog * {
    color: white !important;
  }

  :global([data-theme="dark"]) .dialog-button.delete {
    color: white !important;
  }

  /* Стиль для кнопки в бычном режиме */
  .habit-card:not(:global(.list-view) *) .more-button {
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
  }

  /* Стиль для кнопки в режиме списка */
  :global(.list-view) .more-list-view-button {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    right: 8px;
    background: none;
    border: none;
    color: black;
    font-size: 24px;
    padding: 8px;
    cursor: pointer;
    z-index: 3;
  }
</style>
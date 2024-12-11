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
                // Ошибка уже обработа��а в updateHabitOnServer
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
        
        {#if !$isListView}
          {#if habit.want_to_become}
            <div class="want-to-become">
              <span class="label">{$_('habits.want_to_become')}</span>
              <span class="value">{habit.want_to_become}</span>
            </div>
          {/if}
        {/if}

        {#if completed}
          <button class="undo-button" on:click={handleUndo}>↩</button>
        {/if}
      </div>
    </div>
  </div>
  <div class="streak-counter">
    {habit.streak}
  </div>
</div>

<style>
  .habit-wrapper {
    position: relative;
    width: 280px;
    aspect-ratio: 1;
    margin: 0 auto;
  }

  /* Обновляем стили для режима списка */
  :global(.list-view) .habit-wrapper {
    width: calc(100% - 16px);
    aspect-ratio: unset;
    min-height: 80px;
    height: auto;
    margin: 4px auto;
  }

  .card-shadow {
    width: 100%;
    height: 100%;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.08));
  }

  .card-shadow:has(.habit-card.pressed),
  .card-shadow:has(.habit-card.completed) {
    filter: drop-shadow(0 4px 12px rgba(139, 92, 246, 0.3));
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
    box-shadow: 0 4px 8px rgba(139, 92, 246, 0.3);
    z-index: 1;
    mask: url('/src/assets/streak.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/streak.svg') no-repeat center / contain;
  }

  /* Изменяем положение streak в режиме спика */
  :global(.list-view) .streak-counter {
    position: absolute;
    left: 10px;
    top: 50%;
    transform: translateY(-50%);
    width: 60px;
    height: 60px;
    z-index: 2; /* Поднимаем streak над карточкой */
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
    border-radius: 100px;
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
    padding: 16px 32px 16px 84px;
    mask: none !important;
    -webkit-mask: none !important;
    width: 100%;
    min-height: 80px;
    height: auto;
    background: white;
    color: #333;
    text-align: left;
  }

  .habit-card.pressed,
  .habit-card.completed {
    background: var(--habit-gradient);
    box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
  }

  .habit-card.pressed h3,
  .habit-card.completed h3 {
    color: white;
  }

  h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
  }

  /* Уменьшаем размер заголовка в режиме списка */
  :global(.list-view) h3 {
    font-size: 20px;
    white-space: normal;
    overflow: visible;
    text-overflow: unset;
    margin-right: 40px;
    margin-left: 60px;
    line-height: 1.2;
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
    bottom: 50%;
    right: 16px;
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

  /* Добавляем контейнер для списка */
  :global(.list-view) {
    overflow-x: hidden;
    width: 100%;
    padding: 0 4px;
    display: flex;
    flex-direction: column;
  }

  /* Убираем тен для card-shadow в режиме списка */
  :global(.list-view) .card-shadow {
    background: transparent;
    filter: none;
  }

  /* Убираем тень для completed карточек в режиме списка */
  :global(.list-view) .card-shadow:has(.habit-card.completed) {
    filter: none;
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

  /* Обновляем стил�� для completed состояния */
  :global(.list-view) .habit-card.completed {
    background: var(--habit-gradient);
    color: white;
  }

  /* Обновляем streak в обычном состоянии */
  :global(.list-view) .streak-counter {
    position: absolute;
    left: 10px;
    top: 50%;
    transform: translateY(-50%);
    width: 60px;
    height: 60px;
    background: var(--habit-gradient);
    color: white;
    z-index: 2;
  }

  /* Обновляем streak для completed состояния */
  :global(.list-view) .habit-wrapper:has(.habit-card.completed) .streak-counter {
    background: white;
    color: #8B5CF6;
  }

  /* Добавляем отступ для текста, чтобы не пересекался со streak и кнопкой отмены */
  :global(.list-view) h3 {
    font-size: 20px;
    white-space: normal;
    overflow: visible;
    text-overflow: unset;
    margin-right: 40px;
    margin-left: 60px;
    line-height: 1.2;
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

  h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
  }

  .want-to-become .label {
    font-size: 12px;
    opacity: 0.7;
  }

  .want-to-become .value {
    font-size: 20px;
    font-weight: 700;
  }

  /* Обновляем стили для текста в режиме списка */
  :global(.list-view) h3 {
    font-size: 20px;
    white-space: normal;
    overflow: visible;
    text-overflow: unset;
    margin-right: 40px;
    margin-left: 65px;
    line-height: 1.2;
  }

  :global(.list-view) .content {
    padding-left: 0;
    width: 100%;
    text-align: left;
  }

  :global(.list-view) .want-to-become {
    display: none;
  }
</style>
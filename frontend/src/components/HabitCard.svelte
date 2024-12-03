<script lang="ts">
    export let habit: {
        id: string;
        title: string;
        streak: number;
        score: number;
        last_click_date?: string;
    };
    
    export let telegramId: number;
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    
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
            const response = await fetch('https://lenichev.site/ht_back/habit/update', {
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
                throw new Error('Ошибка при обновлении привычки');
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
    
    async function handleUndo() {
        try {
            const response = await fetch('https://lenichev.site/ht_back/habit/undo', {
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
                throw new Error('Ошибка при отмене привычки');
            }
            
            const data = await response.json();
            if (data.habit) {
                habit = { ...data.habit };
                completed = isCompletedToday();
            }
        } catch (error) {
            console.error('Ошибка:', error);
        }
    }
</script>
  
<div class="habit-card"
  class:pressed={isPressed}
  class:completed={completed}
  on:pointerdown={handlePointerDown}
  on:pointerup={handlePointerUp}
  on:pointerleave={handlePointerUp}>
  <div class="streak-counter">
    {habit.streak}
  </div>
  <h3>{habit.title}</h3>
  
  {#if completed}
    <button 
      class="undo-button"
      on:click|stopPropagation={handleUndo}
    >
      ↩️
    </button>
  {/if}
</div>

<style>
  .habit-card {
    width: 280px;
    aspect-ratio: 1;
    background: white;
    border-radius: 100px;
    padding: 32px;
    position: relative;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    transition: background 0.8s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    user-select: none;
    -webkit-user-select: none; /* Safari */
    -moz-user-select: none; /* Firefox */
    -ms-user-select: none; /* IE10+/Edge */
  }

  .habit-card.pressed,
  .habit-card.completed {
    background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%);
  }

  .habit-card.pressed h3,
  .habit-card.completed h3 {
    color: white;
  }

  .streak-counter {
    position: absolute;
    top: -10px;
    right: -10px;
    width: 60px;
    height: 60px;
    background: #8B5CF6;
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 24px;
    box-shadow: 0 4px 8px rgba(139, 92, 246, 0.3);
  }

  /* Добавляем стили для режима списка */
  :global(.list-view) .streak-counter {
    top: 50%;
    right: auto;
    left: 8px;
    transform: translateY(-50%);
    width: 30px;
    height: 30px;
    border-radius: 15px;
    font-size: 16px;
  }

  h3 {
    margin: 0;
    font-size: 24px;
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
    transition: all 0.2s ease;
  }

  /* Добавляем стили для кнопки отмены в режиме списка */
  :global(.list-view) .undo-button {
    bottom: 50%;
    left: auto;
    right: 8px;
    transform: translateY(50%);
    font-size: 16px;
    padding: 4px;
  }

  .undo-button:hover {
    opacity: 1;
    transform: translateX(-50%) scale(1.1);
  }

  .undo-button:active {
    transform: translateX(-50%) scale(0.9);
  }
</style>
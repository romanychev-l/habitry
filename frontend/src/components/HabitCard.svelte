<script lang="ts">
    export let habit: {
      title: string;
      streak: number;
    };
    
    let isPressed = false;
    let isPressTimeout: ReturnType<typeof setTimeout>;
    
    function handlePointerDown() {
        isPressed = true;
        isPressTimeout = setTimeout(() => {
            habit.streak += 1;
            isPressed = false;
        }, 800); // 800мс для удержания
    }
    
    function handlePointerUp() {
        clearTimeout(isPressTimeout);
        isPressed = false;
    }
</script>
  
  <div class="habit-card"
    class:pressed={isPressed}
    on:pointerdown={handlePointerDown}
    on:pointerup={handlePointerUp}
    on:pointerleave={handlePointerUp}>
    <div class="streak-counter">
      {habit.streak}
    </div>
    <h3>{habit.title}</h3>
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
  
    .habit-card.pressed {
      background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%);
    }
  
    .habit-card.pressed h3 {
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
  
    h3 {
      margin: 0;
      font-size: 24px;
      color: #333;
    }
  </style>

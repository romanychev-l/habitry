<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { fade, scale, fly, blur } from 'svelte/transition';
  import { cubicOut, elasticOut } from 'svelte/easing';

  const dispatch = createEventDispatcher();
  
  export let isSharedHabit = false;
  
  let currentStep = 0;
  $: totalSteps = isSharedHabit ? 3 : 6;
  let showConfetti = false;
  let interactionComplete = false;
  let holdProgress = 0;
  let isHolding = false;
  let holdInterval: number | null = null;
  let habitCompleted = false;
  let streakDays = 0;
  let friendsAdded = 0;
  let mounted = false;

  onMount(() => {
    mounted = true;
  });

  function nextStep() {
    if (currentStep < totalSteps - 1) {
      // Проверяем, требуется ли интерактив на текущем шаге
      if (currentStep === 1 && !habitCompleted) return; // Шаг с long press
      if (currentStep === 2 && streakDays < 7) return; // Шаг со стриком
      
      interactionComplete = false;
      holdProgress = 0;
      habitCompleted = false;
      streakDays = 0;
      friendsAdded = 0;
      currentStep++;
    } else {
      triggerConfetti();
      setTimeout(() => dispatch('finish'), 1500);
    }
  }

  function skipOnboarding() {
    dispatch('skip');
  }

  function triggerConfetti() {
    showConfetti = true;
    setTimeout(() => showConfetti = false, 3000);
  }

  // Интерактивные функции для long press
  function startHold() {
    if (habitCompleted) return;
    isHolding = true;
    holdProgress = 0;
    
    holdInterval = window.setInterval(() => {
      holdProgress += 2;
      if (holdProgress >= 100) {
        completeHabit();
      }
    }, 20);
  }

  function stopHold() {
    isHolding = false;
    if (holdInterval !== null) {
      clearInterval(holdInterval);
      holdInterval = null;
    }
    if (!habitCompleted) {
      holdProgress = 0;
    }
  }

  function completeHabit() {
    stopHold();
    habitCompleted = true;
    interactionComplete = true;
    holdProgress = 100;
  }

  function incrementStreak() {
    if (streakDays < 7) {
      streakDays++;
      if (streakDays === 7) {
        interactionComplete = true;
      }
    }
  }

  function addFriend() {
    if (friendsAdded < 2) {
      friendsAdded++;
      if (friendsAdded === 2) {
        interactionComplete = true;
      }
    }
  }

  $: canProceed = 
    currentStep === 0 || 
    currentStep === 3 ||
    currentStep === 4 ||
    currentStep === 5 ||
    (currentStep === 1 && habitCompleted) ||
    (currentStep === 2 && streakDays >= 7) ||
    (isSharedHabit && currentStep === 2 && friendsAdded >= 2);
</script>

<div class="wrapper">
  <div 
    class="overlay" 
    role="button"
    tabindex="-1"
    on:click={skipOnboarding}
    on:keydown={(e) => e.key === 'Escape' && skipOnboarding()}
    transition:fade={{ duration: 300 }}
  ></div>
  
  {#if showConfetti}
    <div class="confetti-container" transition:fade={{ duration: 500 }}>
      {#each Array(50) as _, i}
        <div 
          class="confetti" 
          style="
            left: {Math.random() * 100}%;
            animation-delay: {Math.random() * 0.5}s;
            animation-duration: {2 + Math.random() * 2}s;
            background: hsl({Math.random() * 360}, 70%, 60%);
          "
        ></div>
      {/each}
    </div>
  {/if}

  <div class="modal-container" transition:scale={{ duration: 400, start: 0.9, easing: elasticOut }}>
    <div class="modal">
      <div class="stars-bg">
        {#each Array(20) as _, i}
          <div 
            class="star" 
            style="
              left: {Math.random() * 100}%;
              top: {Math.random() * 100}%;
              animation-delay: {Math.random() * 3}s;
            "
          ></div>
        {/each}
      </div>

      <div class="content">
        {#key currentStep}
          <div class="step" 
            in:fly={{ y: 50, duration: 500, delay: 100, easing: cubicOut }}
            out:fly={{ y: -50, duration: 300, easing: cubicOut }}
          >
            {#if currentStep === 0}
              <!-- Шаг 1: Приветствие -->
              <div class="illustration-container">
                <div class="glow-effect"></div>
                <div class="illustration animated-wave">👋</div>
              </div>
              <h2 class="gradient-text">{$_('onboarding.welcome.title')}</h2>
              <p class="description">{$_('onboarding.welcome.description')}</p>
              <div class="quote">
                <div class="quote-icon">💫</div>
                <p class="quote-text">{$_('onboarding.welcome.quote')}</p>
                <p class="quote-author">{$_('onboarding.welcome.quote_author')}</p>
              </div>
              
            {:else if currentStep === 1}
              <!-- Шаг 2: Интерактив с long press -->
              <div class="illustration-container">
                <div class="glow-effect pulse"></div>
                <div class="illustration">💪</div>
              </div>
              <h2>{$_('onboarding.social.title')}</h2>
              <p class="description">{$_('onboarding.social.description')}</p>
              
              <button 
                class="habit-demo" 
                class:completed={habitCompleted}
                class:holding={isHolding}
                on:mousedown={startHold}
                on:mouseup={stopHold}
                on:mouseleave={stopHold}
                on:touchstart={startHold}
                on:touchend={stopHold}
                on:touchcancel={stopHold}
                disabled={habitCompleted}
              >
                <div class="habit-demo-content">
                  <div class="habit-icon">{habitCompleted ? '✨' : '💪'}</div>
                  <div class="habit-text">
                    {habitCompleted ? $_('onboarding.social.great_job') : $_('onboarding.social.hold_to_complete')}
                  </div>
                </div>
                {#if holdProgress > 0 && !habitCompleted}
                  <div class="progress-ring" style="--progress: {holdProgress}"></div>
                {/if}
              </button>
              
              {#if habitCompleted}
                <div class="success-badge" in:scale={{ duration: 400, easing: elasticOut }}>
                  <span class="badge-icon">🎉</span>
                  <span class="badge-text">{$_('onboarding.social.success_message')}</span>
                </div>
              {/if}

            {:else if currentStep === 2 && !isSharedHabit}
              <!-- Шаг 3: Интерактив со стриком -->
              <div class="illustration-container">
                <div class="glow-effect streak-glow"></div>
                <div class="streak-display">
                  <div class="flame-icon">🔥</div>
                  <div class="streak-number">{streakDays}</div>
                </div>
              </div>
              <h2>{$_('onboarding.interaction.title')}</h2>
              <p class="description">{$_('onboarding.interaction.description')}</p>
              
              <div class="days-grid">
                {#each Array(7) as _, i}
                  <button 
                    class="day-cell"
                    class:active={i < streakDays}
                    on:click={incrementStreak}
                    disabled={i < streakDays}
                  >
                    {#if i < streakDays}
                      <span class="checkmark" in:scale={{ duration: 300, easing: elasticOut }}>✓</span>
                    {:else}
                      <span class="day-number">{i + 1}</span>
                    {/if}
                  </button>
                {/each}
              </div>

              {#if streakDays >= 7}
                <div class="success-badge" in:scale={{ duration: 400, easing: elasticOut }}>
                  <span class="badge-icon">🏆</span>
                  <span class="badge-text">{$_('onboarding.interaction.streak_achieved')}</span>
                </div>
              {/if}

            {:else if currentStep === 2 && isSharedHabit}
              <!-- Для shared habit: социальный аспект -->
              <div class="illustration-container">
                <div class="glow-effect social-glow"></div>
                <div class="friends-demo">
                  <div class="friend-avatar" class:added={friendsAdded > 0}>👤</div>
                  <div class="friend-avatar" class:added={friendsAdded > 1}>👤</div>
                </div>
              </div>
              <h2>{$_('onboarding.interaction.title')}</h2>
              <p class="description">{$_('onboarding.interaction.description')}</p>
              
              <button 
                class="add-friend-btn" 
                on:click={addFriend}
                disabled={friendsAdded >= 2}
              >
                {friendsAdded >= 2 ? '✨ Команда собрана!' : `➕ Добавить друга (${friendsAdded}/2)`}
              </button>

            {:else if currentStep === 3 && !isSharedHabit}
              <!-- Шаг 4: Статистика -->
              <div class="illustration-container">
                <div class="glow-effect stats-glow"></div>
                <div class="illustration stats-icon">📊</div>
              </div>
              <h2>{$_('onboarding.stats.title')}</h2>
              <p class="description">{$_('onboarding.stats.description')}</p>
              
              <div class="stats-comparison">
                <div class="stat-bar">
                  <div class="stat-label">
                    <span class="stat-emoji">👥</span>
                    <span class="stat-name">{$_('onboarding.stats.with_friends')}</span>
                  </div>
                  <div class="stat-progress-container">
                    <div class="stat-progress with-friends" style="width: 95%"></div>
                    <span class="stat-value">95%</span>
                  </div>
                </div>
                
                <div class="stat-bar">
                  <div class="stat-label">
                    <span class="stat-emoji">🚶</span>
                    <span class="stat-name">{$_('onboarding.stats.alone')}</span>
                  </div>
                  <div class="stat-progress-container">
                    <div class="stat-progress alone" style="width: 33%"></div>
                    <span class="stat-value">33%</span>
                  </div>
                </div>
              </div>
              
              <div class="stats-note">
                <span class="stats-note-icon">💡</span>
                <span class="stats-note-text">{$_('onboarding.stats.success_rate')}</span>
              </div>

            {:else if currentStep === 4 && !isSharedHabit}
              <!-- Шаг 5: Социальный аспект -->
              <div class="illustration-container">
                <div class="glow-effect community-glow"></div>
                <div class="illustration community-icon">🤝</div>
              </div>
              <h2>{$_('onboarding.community.title')}</h2>
              <p class="description">{$_('onboarding.community.description')}</p>
              <div class="features-list">
                <div class="feature-item">
                  <span class="feature-icon">👥</span>
                  <span class="feature-text">{$_('onboarding.community.shared_habits')}</span>
                </div>
                <div class="feature-item">
                  <span class="feature-icon">🏆</span>
                  <span class="feature-text">{$_('onboarding.community.leaderboard')}</span>
                </div>
                <div class="feature-item">
                  <span class="feature-icon">💬</span>
                  <span class="feature-text">{$_('onboarding.community.friend_support')}</span>
                </div>
              </div>

            {:else if (currentStep === 5 && !isSharedHabit) || (currentStep === 3 && isSharedHabit)}
              <!-- Финальный шаг -->
              <div class="illustration-container final">
                <div class="glow-effect rainbow-glow"></div>
                <div class="illustration rocket">🚀</div>
              </div>
              <h2 class="gradient-text final-title">{$_('onboarding.start.title')}</h2>
              <p class="description final-description">{$_('onboarding.start.description')}</p>
              <div class="motivational-box">
                <div class="motivational-icon">✨</div>
                <p class="motivational-text">{$_('onboarding.start.motivation')}</p>
              </div>
            {/if}
          </div>
        {/key}

        <div class="progress">
          {#each Array(totalSteps) as _, i}
            <div 
              class="dot" 
              class:active={i === currentStep}
              class:completed={i < currentStep}
            ></div>
          {/each}
        </div>
      </div>

      <div class="footer">
        <button class="skip-btn" on:click={skipOnboarding}>
          {$_('onboarding.skip')}
        </button>
        <button 
          class="next-btn" 
          class:pulse-btn={canProceed && currentStep < totalSteps - 1}
          class:final-btn={currentStep === totalSteps - 1}
          on:click={nextStep}
          disabled={!canProceed}
        >
          <span class="btn-text">
            {currentStep === totalSteps - 1 ? $_('onboarding.start.action') : $_('onboarding.next')}
          </span>
          {#if currentStep === totalSteps - 1}
            <span class="btn-icon">🚀</span>
          {/if}
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .wrapper {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100dvh;
    z-index: 1000;
  }

  .overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
  }

  .confetti-container {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 2;
    overflow: hidden;
  }

  .confetti {
    position: absolute;
    width: 10px;
    height: 10px;
    top: -10px;
    animation: fall linear forwards;
    transform-origin: center;
  }

  @keyframes fall {
    to {
      transform: translateY(100vh) rotate(360deg);
      opacity: 0;
    }
  }

  .modal-container {
    position: relative;
    width: 100%;
    max-width: 360px;
    margin: 0 16px;
    z-index: 1;
  }

  .modal {
    position: relative;
    width: 100%;
    background: var(--tg-theme-bg-color, #F9F8F3);
    border-radius: 28px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    overflow: hidden;
  }

  .stars-bg {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .star {
    position: absolute;
    width: 2px;
    height: 2px;
    background: white;
    border-radius: 50%;
    animation: twinkle 3s infinite;
    opacity: 0;
  }

  @keyframes twinkle {
    0%, 100% { opacity: 0; }
    50% { opacity: 0.8; }
  }

  .content {
    position: relative;
    padding: 40px 24px 32px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    min-height: 480px;
  }

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    width: 100%;
  }

  .illustration-container {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 8px;
    min-height: 120px;
  }

  .glow-effect {
    position: absolute;
    width: 120px;
    height: 120px;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(147, 51, 234, 0.3) 0%, transparent 70%);
    filter: blur(20px);
    animation: glow 2s ease-in-out infinite;
  }

  .glow-effect.pulse {
    background: radial-gradient(circle, rgba(59, 130, 246, 0.4) 0%, transparent 70%);
  }

  .glow-effect.streak-glow {
    background: radial-gradient(circle, rgba(249, 115, 22, 0.4) 0%, transparent 70%);
  }

  .glow-effect.community-glow {
    background: radial-gradient(circle, rgba(34, 197, 94, 0.4) 0%, transparent 70%);
  }

  .glow-effect.stats-glow {
    background: radial-gradient(circle, rgba(168, 85, 247, 0.4) 0%, transparent 70%);
  }

  .glow-effect.rainbow-glow {
    background: radial-gradient(circle, 
      rgba(147, 51, 234, 0.3) 0%, 
      rgba(59, 130, 246, 0.3) 33%,
      rgba(34, 197, 94, 0.3) 66%,
      transparent 100%
    );
    animation: rainbow-glow 3s ease-in-out infinite;
  }

  @keyframes glow {
    0%, 100% { transform: scale(1); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 0.8; }
  }

  @keyframes rainbow-glow {
    0%, 100% { transform: scale(1) rotate(0deg); opacity: 0.6; }
    50% { transform: scale(1.2) rotate(180deg); opacity: 0.9; }
  }

  .illustration {
    font-size: 80px;
    position: relative;
    z-index: 1;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
  }

  .animated-wave {
    animation: wave 1s ease-in-out infinite;
  }

  @keyframes wave {
    0%, 100% { transform: rotate(0deg); }
    25% { transform: rotate(20deg); }
    75% { transform: rotate(-20deg); }
  }

  .rocket {
    animation: float 2s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
  }

  .community-icon {
    animation: pulse-scale 2s ease-in-out infinite;
  }

  @keyframes pulse-scale {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.1); }
  }

  h2 {
    margin: 0;
    font-size: 28px;
    font-weight: 700;
    color: var(--tg-theme-text-color);
    letter-spacing: -0.5px;
  }

  .gradient-text {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .final-title {
    font-size: 32px;
    animation: pulse-scale 2s ease-in-out infinite;
  }

  .description {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
    max-width: 300px;
  }

  .final-description {
    font-size: 18px;
    font-weight: 500;
  }

  .quote {
    background: linear-gradient(135deg, rgba(147, 51, 234, 0.1) 0%, rgba(79, 70, 229, 0.1) 100%);
    border-radius: 16px;
    padding: 20px;
    margin-top: 12px;
    border: 1px solid rgba(147, 51, 234, 0.2);
  }

  .quote-icon {
    font-size: 32px;
    margin-bottom: 8px;
  }

  .quote-text {
    font-size: 15px;
    font-style: italic;
    line-height: 1.6;
    color: var(--tg-theme-text-color);
    margin: 8px 0;
  }

  .quote-author {
    font-size: 13px;
    color: var(--tg-theme-hint-color);
    margin: 8px 0 0;
    font-weight: 600;
  }

  /* Интерактивная кнопка-привычка */
  .habit-demo {
    position: relative;
    width: 100%;
    max-width: 300px;
    height: 160px;
    margin-top: 20px;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(147, 51, 234, 0.15) 100%);
    border: none;
    cursor: pointer;
    transition: all 0.3s ease;
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.1));
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
  }

  .habit-demo:not(:disabled):active {
    transform: scale(0.97);
  }

  .habit-demo.holding {
    transform: scale(0.97);
    filter: drop-shadow(0 4px 12px rgba(59, 130, 246, 0.3));
  }

  .habit-demo.completed {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.25) 0%, rgba(59, 130, 246, 0.25) 100%);
    animation: success-pulse 0.6s ease;
    filter: drop-shadow(0 4px 16px rgba(34, 197, 94, 0.4));
  }

  @keyframes success-pulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.05); }
  }

  .habit-demo-content {
    position: absolute;
    inset: 0;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 24px 32px;
  }

  .habit-icon {
    font-size: 48px;
    line-height: 1;
  }

  .habit-text {
    font-size: 15px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
    text-align: center;
    line-height: 1.3;
  }

  .progress-ring {
    position: absolute;
    inset: -2px;
    background: conic-gradient(
      rgba(59, 130, 246, 0.7) calc(var(--progress) * 1%), 
      transparent calc(var(--progress) * 1%)
    );
    pointer-events: none;
    transition: background 0.05s linear;
    mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    -webkit-mask: url('/src/assets/squircley.svg') no-repeat center / contain;
    filter: blur(1px);
  }

  /* Streak display */
  .streak-display {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    position: relative;
    z-index: 1;
  }

  .flame-icon {
    font-size: 64px;
    animation: flicker 1.5s ease-in-out infinite;
  }

  @keyframes flicker {
    0%, 100% { transform: scale(1); filter: brightness(1); }
    50% { transform: scale(1.1); filter: brightness(1.2); }
  }

  .streak-number {
    font-size: 48px;
    font-weight: 800;
    background: linear-gradient(135deg, #f97316 0%, #dc2626 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
    margin-top: 12px;
  }

  .day-cell {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 2px solid rgba(0, 0, 0, 0.1);
    background: var(--tg-theme-secondary-bg-color, #fff);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .day-cell:not(:disabled):active {
    transform: scale(0.9);
  }

  .day-cell.active {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
    border-color: #10b981;
    color: white;
  }

  .checkmark {
    font-size: 20px;
  }

  .day-number {
    color: var(--tg-theme-hint-color);
  }

  /* Success badge */
  .success-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.2) 0%, rgba(59, 130, 246, 0.2) 100%);
    border-radius: 16px;
    border: 2px solid rgba(34, 197, 94, 0.3);
    margin-top: 8px;
  }

  .badge-icon {
    font-size: 24px;
  }

  .badge-text {
    font-size: 15px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  /* Friends demo */
  .friends-demo {
    display: flex;
    gap: 16px;
    position: relative;
    z-index: 1;
  }

  .friend-avatar {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: var(--tg-theme-secondary-bg-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 40px;
    opacity: 0.3;
    transition: all 0.3s ease;
    border: 3px solid rgba(0, 0, 0, 0.1);
  }

  .friend-avatar.added {
    opacity: 1;
    border-color: rgba(34, 197, 94, 0.5);
    animation: friend-added 0.5s ease;
  }

  @keyframes friend-added {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.15); }
  }

  .add-friend-btn {
    padding: 14px 24px;
    border-radius: 14px;
    border: none;
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-top: 12px;
  }

  .add-friend-btn:not(:disabled):active {
    transform: scale(0.95);
  }

  .add-friend-btn:disabled {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
  }

  /* Features list */
  .features-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 16px;
    width: 100%;
  }

  .feature-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--tg-theme-secondary-bg-color, rgba(0, 0, 0, 0.05));
    border-radius: 12px;
    transition: all 0.2s ease;
  }

  .feature-icon {
    font-size: 24px;
  }

  .feature-text {
    font-size: 15px;
    font-weight: 500;
    color: var(--tg-theme-text-color);
  }

  /* Stats comparison */
  .stats-comparison {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin-top: 20px;
    width: 100%;
  }

  .stat-bar {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .stat-label {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .stat-emoji {
    font-size: 20px;
  }

  .stat-name {
    font-size: 15px;
    font-weight: 600;
    color: var(--tg-theme-text-color);
  }

  .stat-progress-container {
    position: relative;
    height: 32px;
    background: var(--tg-theme-secondary-bg-color, rgba(0, 0, 0, 0.05));
    border-radius: 8px;
    overflow: hidden;
    display: flex;
    align-items: center;
  }

  .stat-progress {
    height: 100%;
    border-radius: 8px;
    transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
    animation: progress-fill 1.5s ease-out;
  }

  @keyframes progress-fill {
    from { width: 0 !important; }
  }

  .stat-progress.with-friends {
    background: linear-gradient(90deg, #22c55e 0%, #16a34a 100%);
    box-shadow: 0 0 20px rgba(34, 197, 94, 0.4);
  }

  .stat-progress.alone {
    background: linear-gradient(90deg, #94a3b8 0%, #64748b 100%);
  }

  .stat-value {
    position: absolute;
    right: 12px;
    font-size: 14px;
    font-weight: 700;
    color: white;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    z-index: 1;
  }

  .stats-note {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 16px;
    padding: 12px 16px;
    background: rgba(168, 85, 247, 0.1);
    border-radius: 12px;
    border: 1px solid rgba(168, 85, 247, 0.2);
  }

  .stats-note-icon {
    font-size: 20px;
  }

  .stats-note-text {
    font-size: 13px;
    font-weight: 500;
    color: var(--tg-theme-text-color);
  }

  .stats-icon {
    animation: stats-bounce 2s ease-in-out infinite;
  }

  @keyframes stats-bounce {
    0%, 100% { transform: translateY(0px) scale(1); }
    50% { transform: translateY(-8px) scale(1.05); }
  }

  /* Motivational box */
  .motivational-box {
    background: linear-gradient(135deg, rgba(147, 51, 234, 0.15) 0%, rgba(59, 130, 246, 0.15) 100%);
    border-radius: 20px;
    padding: 24px;
    margin-top: 16px;
    border: 2px solid rgba(147, 51, 234, 0.3);
    animation: pulse-glow 2s ease-in-out infinite;
  }

  @keyframes pulse-glow {
    0%, 100% { box-shadow: 0 0 0 rgba(147, 51, 234, 0.4); }
    50% { box-shadow: 0 0 20px rgba(147, 51, 234, 0.6); }
  }

  .motivational-icon {
    font-size: 40px;
    margin-bottom: 12px;
  }

  .motivational-text {
    font-size: 16px;
    font-weight: 600;
    line-height: 1.5;
    color: var(--tg-theme-text-color);
    margin: 0;
  }

  .progress {
    display: flex;
    gap: 10px;
    margin-top: auto;
    padding-top: 24px;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--tg-theme-secondary-bg-color);
    transition: all 0.3s ease;
  }

  .dot.completed {
    background: rgba(34, 197, 94, 0.6);
    transform: scale(0.8);
  }

  .dot.active {
    background: var(--tg-theme-button-color);
    transform: scale(1.3);
    box-shadow: 0 0 12px rgba(147, 51, 234, 0.6);
  }

  .footer {
    padding: 16px 24px 20px;
    border-top: 1px solid var(--tg-theme-secondary-bg-color);
    display: flex;
    gap: 12px;
  }

  .skip-btn {
    flex: 1;
    padding: 16px;
    border-radius: 14px;
    border: 2px solid var(--tg-theme-button-color);
    background: transparent;
    color: var(--tg-theme-button-color);
    font-size: 15px;
    font-weight: 600;
    transition: all 0.2s ease;
    cursor: pointer;
  }

  .skip-btn:active {
    transform: scale(0.97);
  }

  .next-btn {
    flex: 2;
    padding: 16px;
    border-radius: 14px;
    border: none;
    background: var(--tg-theme-button-color);
    color: var(--tg-theme-button-text-color);
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    position: relative;
    overflow: hidden;
  }

  .next-btn:not(:disabled):active {
    transform: scale(0.97);
  }

  .next-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .pulse-btn {
    animation: button-pulse 1.5s ease-in-out infinite;
  }

  @keyframes button-pulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(147, 51, 234, 0.7); }
    50% { box-shadow: 0 0 0 8px rgba(147, 51, 234, 0); }
  }

  .final-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    animation: rainbow-shift 3s ease infinite;
    background-size: 200% 200%;
  }

  @keyframes rainbow-shift {
    0%, 100% { background-position: 0% 50%; }
    50% { background-position: 100% 50%; }
  }

  .btn-icon {
    font-size: 20px;
  }

  /* Dark theme adjustments */
  :global([data-theme="dark"]) .modal {
    background: #1a1a1a;
  }

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description,
  :global([data-theme="dark"]) .quote-text {
    color: rgba(255, 255, 255, 0.7);
  }

  :global([data-theme="dark"]) .skip-btn {
    color: white;
    border-color: white;
  }

  :global([data-theme="dark"]) .next-btn {
    background: white;
    color: #1a1a1a;
  }

  :global([data-theme="dark"]) .final-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  :global([data-theme="dark"]) .day-cell {
    background: #2a2a2a;
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) .feature-item {
    background: rgba(255, 255, 255, 0.05);
  }

  :global([data-theme="dark"]) .stat-progress-container {
    background: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) .stats-note {
    background: rgba(168, 85, 247, 0.15);
    border-color: rgba(168, 85, 247, 0.3);
  }
</style> 
<script lang="ts">
  import { _ } from 'svelte-i18n';
</script>

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

<style>
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
    filter: blur(20px);
    animation: glow 2s ease-in-out infinite;
  }

  .glow-effect.stats-glow {
    background: radial-gradient(circle, rgba(168, 85, 247, 0.4) 0%, transparent 70%);
  }

  @keyframes glow {
    0%, 100% { transform: scale(1); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 0.8; }
  }

  .illustration {
    font-size: 80px;
    position: relative;
    z-index: 1;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
  }

  .stats-icon {
    animation: stats-bounce 2s ease-in-out infinite;
  }

  @keyframes stats-bounce {
    0%, 100% { transform: translateY(0px) scale(1); }
    50% { transform: translateY(-8px) scale(1.05); }
  }

  h2 {
    margin: 0;
    font-size: 28px;
    font-weight: 700;
    color: var(--tg-theme-text-color);
    letter-spacing: -0.5px;
  }

  .description {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
    max-width: 300px;
  }

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

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.7);
  }

  :global([data-theme="dark"]) .stat-progress-container {
    background: rgba(255, 255, 255, 0.1);
  }

  :global([data-theme="dark"]) .stats-note {
    background: rgba(168, 85, 247, 0.15);
    border-color: rgba(168, 85, 247, 0.3);
  }
</style>


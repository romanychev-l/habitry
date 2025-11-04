<script lang="ts">
  import { _ } from 'svelte-i18n';
</script>

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

  .glow-effect.community-glow {
    background: radial-gradient(circle, rgba(34, 197, 94, 0.4) 0%, transparent 70%);
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

  .description {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
    max-width: 300px;
  }

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

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.7);
  }

  :global([data-theme="dark"]) .feature-item {
    background: rgba(255, 255, 255, 0.05);
  }
</style>


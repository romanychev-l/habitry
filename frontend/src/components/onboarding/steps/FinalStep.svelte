<script lang="ts">
  import { _ } from 'svelte-i18n';
</script>

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

  .rocket {
    animation: float 2s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
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

  @keyframes pulse-scale {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.1); }
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

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description {
    color: rgba(255, 255, 255, 0.7);
  }
</style>


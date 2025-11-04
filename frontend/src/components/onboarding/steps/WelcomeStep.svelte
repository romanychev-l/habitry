<script lang="ts">
  import { _ } from 'svelte-i18n';
</script>

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
    background: radial-gradient(circle, rgba(147, 51, 234, 0.3) 0%, transparent 70%);
    filter: blur(20px);
    animation: glow 2s ease-in-out infinite;
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

  .animated-wave {
    animation: wave 1s ease-in-out infinite;
  }

  @keyframes wave {
    0%, 100% { transform: rotate(0deg); }
    25% { transform: rotate(20deg); }
    75% { transform: rotate(-20deg); }
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

  .description {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
    max-width: 300px;
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

  :global([data-theme="dark"]) h2 {
    color: white;
  }

  :global([data-theme="dark"]) .description,
  :global([data-theme="dark"]) .quote-text {
    color: rgba(255, 255, 255, 0.7);
  }
</style>


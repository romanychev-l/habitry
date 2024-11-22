<script lang="ts">
    import { onMount } from 'svelte';
  
    onMount(() => {
      // Дополнительная проверка на случай, если что-то пошло не так при инициализации
      if (window.Telegram?.WebApp) {
        window.Telegram.WebApp.ready();
        window.Telegram.WebApp.expand();
      }
    });
  </script>
  
  <div class="min-h-screen w-full bg-[var(--tg-theme-bg-color)] text-[var(--tg-theme-text-color)]">
  <slot />
</div>

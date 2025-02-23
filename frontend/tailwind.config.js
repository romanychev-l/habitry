/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        'tg': {
          'bg': 'var(--tg-theme-bg-color)',
          'text': 'var(--tg-theme-text-color)',
          'hint': 'var(--tg-theme-hint-color)',
          'link': 'var(--tg-theme-link-color)',
          'button': 'var(--tg-theme-button-color)',
          'button-text': 'var(--tg-theme-button-text-color)',
          'secondary-bg': 'var(--tg-theme-secondary-bg-color)',
        },
        'primary': '#00D5A0',
      },
      zIndex: {
        '1000': '1000',
      },
    },
  },
  plugins: [],
} 
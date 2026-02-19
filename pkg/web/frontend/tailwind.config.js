/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'kakoclaw': {
          'bg': 'rgb(var(--pc-bg) / <alpha-value>)',
          'surface': 'rgb(var(--pc-surface) / <alpha-value>)',
          'surface-hover': 'rgb(var(--pc-surface-hover) / <alpha-value>)',
          'border': 'rgb(var(--pc-border) / <alpha-value>)',
          'accent': 'rgb(var(--pc-accent) / <alpha-value>)',
          'accent-hover': 'rgb(var(--pc-accent-hover) / <alpha-value>)',
          'success': 'rgb(var(--pc-success) / <alpha-value>)',
          'warning': 'rgb(var(--pc-warning) / <alpha-value>)',
          'error': 'rgb(var(--pc-error) / <alpha-value>)',
          'text': 'rgb(var(--pc-text) / <alpha-value>)',
          'text-secondary': 'rgb(var(--pc-text-secondary) / <alpha-value>)',
        }
      },
      fontSize: {
        'xs': ['12px', '16px'],
        'sm': ['13px', '17px'],
        'base': ['14px', '20px'],
        'lg': ['16px', '24px'],
        'xl': ['20px', '28px'],
      }
    },
  },
  darkMode: 'class',
  plugins: [],
}

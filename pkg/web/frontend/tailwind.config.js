/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'makoclaw': {
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
      },
      zIndex: {
        'sticky': '20',
        'overlay-backdrop': '40',
        'sidebar': '45',
        'dropdown': '50',
        'modal': '60',
        'modal-nested': '70',
        'toast': '80',
        'banner': '90',
        'popover': '100',
      },
      keyframes: {
        slideDown: {
          from: { opacity: '0', transform: 'translateY(-100%)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        slideIn: {
          from: { opacity: '0', transform: 'translateX(-10px)' },
          to: { opacity: '1', transform: 'translateX(0)' },
        },
        scaleIn: {
          from: { opacity: '0', transform: 'scale(0.95)' },
          to: { opacity: '1', transform: 'scale(1)' },
        },
        fadeSlide: {
          from: { opacity: '0', transform: 'translateY(8px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        skeleton: {
          '0%': { backgroundPosition: '200% center' },
          '100%': { backgroundPosition: '-200% center' },
        },
        expandDown: {
          from: { opacity: '0', maxHeight: '0' },
          to: { opacity: '1', maxHeight: 'var(--expand-height, 500px)' },
        },
      },
      animation: {
        'slideDown': 'slideDown 0.3s ease-out forwards',
        'slideIn': 'slideIn 0.3s ease-out forwards',
        'scaleIn': 'scaleIn 0.2s ease-out forwards',
        'fadeSlide': 'fadeSlide 0.3s ease-out forwards',
        'skeleton': 'skeleton 1.5s ease-in-out infinite',
        'expandDown': 'expandDown 0.3s ease-out forwards',
      },
    },
  },
  darkMode: 'class',
  plugins: [],
}

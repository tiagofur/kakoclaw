/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'picoclaw': {
          'bg': '#0d1117',
          'surface': '#161b22',
          'border': '#30363d',
          'accent': '#007acc',
          'accent-hover': '#1f6feb',
          'success': '#3fb950',
          'warning': '#d29922',
          'error': '#f85149',
          'text': '#e0e0e0',
          'text-secondary': '#8b949e',
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

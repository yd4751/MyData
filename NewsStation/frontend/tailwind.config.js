/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'dark-bg': '#0a0e17',
        'dark-card': 'rgba(20, 30, 50, 0.8)',
        'accent-red': '#ef4444',
        'accent-blue': '#3b82f6',
      },
      backdropBlur: {
        xs: '2px',
      }
    },
  },
  plugins: [],
}
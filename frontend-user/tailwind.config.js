/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    screens: {
      sm: '375px',
      md: '768px',
      xl: '1280px',
    },
    extend: {
      colors: {
        paper: '#F3E6D4',
        card: '#FFF8EE',
        ink: '#2A2118',
        clay: '#C45C26',
        moss: '#3D6B4F',
        gold: '#E0A100',
        rose: '#B42318',
        line: '#E4D2B8',
      },
      fontFamily: {
        display: ['Fraunces', 'Source Serif 4', 'serif'],
        body: ['Noto Serif SC', 'Source Serif 4', 'Songti SC', 'serif'],
      },
      boxShadow: {
        warm: '0 12px 32px rgba(90, 50, 20, 0.08)',
        stamp: '0 4px 0 rgba(90, 50, 20, 0.12)',
      },
      borderRadius: {
        page: '22px',
      },
    },
  },
  plugins: [],
}

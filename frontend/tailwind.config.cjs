/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        varys: {
          bg: '#290137',
          surface: '#350645',
          primary: '#9333ea',
          secondary: '#f59e0b',
          muted: '#4b1263',
          border: '#5b21b6',
        }
      }
    },
  },
  plugins: [],
}
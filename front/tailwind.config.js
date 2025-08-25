/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}", // <-- scan React files
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}

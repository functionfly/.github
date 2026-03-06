/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'admin': {
          '50': '#f8f9fa',
          '100': '#f1f3f5',
          '200': '#e9ecef',
          '300': '#dee2e6',
          '400': '#ced4da',
          '500': '#adb5bd',
          '600': '#6c757d',
          '700': '#495057',
          '800': '#343a40',
          '900': '#212529',
        }
      },
      fontFamily: {
        'sans': ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
      }
    },
  },
  plugins: [],
}

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      colors: {
        // Primary brand colors
        'accent-primary': '#3b82f6', // Blue-500
        'accent-secondary': '#1e40af', // Blue-700

        // Background colors
        'bg-primary': '#0a0a0f', // Very dark blue-black
        'bg-secondary': '#111118', // Slightly lighter dark

        // Text colors
        'text-primary': '#ffffff',
        'text-secondary': '#94a3b8', // Slate-400

        // Border colors
        'border-primary': '#1e293b', // Slate-800
        'border-secondary': '#334155', // Slate-700
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'grid-pattern': "url(\"data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32' width='32' height='32' fill='none' stroke='rgb(71 85 105 / 0.1)'%3e%3cpath d='m0 .5h32m-32 8h32m-32 8h32m-32 8h32'/%3e%3c/svg%3e\")",
      },
    },
  },
  plugins: [],
};
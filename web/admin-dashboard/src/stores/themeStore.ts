import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Theme = 'light' | 'dark' | 'system';

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'system' as Theme,

      setTheme: (theme: Theme) => {
        set({ theme });
        applyTheme(theme);
      },

      toggleTheme: () => {
        const currentTheme = get().theme;
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        set({ theme: newTheme });
        applyTheme(newTheme);
      },
    }),
    {
      name: 'admin-theme-storage',
    }
  )
);

/**
 * Apply theme to the document
 */
function applyTheme(theme: Theme) {
  const root = document.documentElement;

  if (theme === 'system') {
    const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.classList.toggle('dark', systemDark);
  } else {
    root.classList.toggle('dark', theme === 'dark');
  }
}

// Initialize theme on load
if (typeof window !== 'undefined') {
  const stored = localStorage.getItem('admin-theme-storage');
  if (stored) {
    try {
      const { state } = JSON.parse(stored);
      if (state?.theme) {
        applyTheme(state.theme);
      }
    } catch {
      applyTheme('system');
    }
  }

  // Listen for system theme changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    const stored = localStorage.getItem('admin-theme-storage');
    if (stored) {
      try {
        const { state } = JSON.parse(stored);
        if (state?.theme === 'system') {
          applyTheme('system');
        }
      } catch {
        // Ignore
      }
    }
  });
}

export default useThemeStore;

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Theme = 'light' | 'dark' | 'system';

// Get system preference
const getSystemTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

// Get initial theme from localStorage or system preference
const getInitialTheme = (): Theme => {
  if (typeof window === 'undefined') return 'dark';
  const stored = localStorage.getItem('theme-storage');
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed.state?.theme) {
        return parsed.state.theme;
      }
    } catch {}
  }
  return 'system';
};

// Get resolved theme (resolves system to actual theme)
const getResolvedTheme = (theme: Theme): 'light' | 'dark' => {
  if (theme === 'system') {
    const systemTheme = getSystemTheme();
    return systemTheme === 'dark' ? 'dark' : 'light';
  }
  return theme;
};

// Apply theme-specific styles to elements
const applyThemeStyles = (theme: 'light' | 'dark') => {
  if (typeof window === 'undefined') return;

  // Apply styles to newsletter input
  const newsletterInput = document.querySelector('.newsletter-input') as HTMLInputElement;
  if (newsletterInput) {
    if (theme === 'light') {
      newsletterInput.style.backgroundColor = '#ffffff';
      newsletterInput.style.borderColor = 'rgba(0, 0, 0, 0.15)';
      newsletterInput.style.color = '#0f0f1a';
      newsletterInput.style.setProperty('--tw-ring-color', 'rgba(99, 102, 241, 0.2)');
    } else {
      newsletterInput.style.backgroundColor = '';
      newsletterInput.style.borderColor = '';
      newsletterInput.style.color = '';
      newsletterInput.style.removeProperty('--tw-ring-color');
    }
  }

  // Apply styles to social icons
  const socialIcons = document.querySelectorAll('.footer-enhanced .text-text-secondary');
  socialIcons.forEach((icon) => {
    const element = icon as HTMLElement;
    if (theme === 'light') {
      element.style.color = '#4a4a5a';
    } else {
      element.style.color = '';
    }
  });
};

interface ThemeState {
  theme: Theme;
  resolvedTheme: 'light' | 'dark';
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: getInitialTheme(),
      resolvedTheme: getResolvedTheme(getInitialTheme()),
      setTheme: (theme: Theme) => {
        const resolved = getResolvedTheme(theme);
        if (import.meta.env.DEV) {
          console.log('Setting theme:', theme, 'resolved:', resolved);
        }
        set({
          theme,
          resolvedTheme: resolved,
        });
        // Update the data-theme attribute
        if (typeof window !== 'undefined') {
          document.documentElement.setAttribute('data-theme', resolved);

          // Apply theme-specific styles after DOM is ready
          if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
              applyThemeStyles(resolved);
            });
          } else {
            requestAnimationFrame(() => applyThemeStyles(resolved));
          }
        }
      },
      toggleTheme: () => {
        const current = get().resolvedTheme;
        const nextTheme = current === 'dark' ? 'light' : 'dark';
        set({
          theme: nextTheme,
          resolvedTheme: nextTheme,
        });
        // Update the data-theme attribute and apply styles
        if (typeof window !== 'undefined') {
          document.documentElement.setAttribute('data-theme', nextTheme);
          // Apply styles after DOM is ready
          if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
              applyThemeStyles(nextTheme);
            });
          } else {
            requestAnimationFrame(() => applyThemeStyles(nextTheme));
          }
        }
      },
    }),
    {
      name: 'theme-storage',
      onRehydrateStorage: () => (state) => {
        // Listen for system theme changes
        if (typeof window !== 'undefined') {
          const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
          const handleChange = () => {
            const currentState = useThemeStore.getState();
            if (currentState.theme === 'system') {
              useThemeStore.setState({
                resolvedTheme: getSystemTheme(),
              });
              // Update the data-theme attribute
              document.documentElement.setAttribute('data-theme', getSystemTheme());
            }
          };
          mediaQuery.addEventListener('change', handleChange);

          // Set initial theme on the document
          if (state) {
            document.documentElement.setAttribute('data-theme', state.resolvedTheme);
            // Apply styles after DOM is ready
            if (document.readyState === 'loading') {
              document.addEventListener('DOMContentLoaded', () => {
                applyThemeStyles(state.resolvedTheme);
              });
            } else {
              // DOM is already ready
              requestAnimationFrame(() => applyThemeStyles(state.resolvedTheme));
            }
          }
        }
      },
    }
  )
);

import { create } from 'zustand';
import { initTheme, setTheme as sharedSetTheme, subscribe, type ThemeState } from '@functionfly/shared/theme';

export type Theme = 'light' | 'dark' | 'system';

const getSystemTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

const getResolvedTheme = (theme: Theme): 'light' | 'dark' => {
  return theme === 'system' ? getSystemTheme() : theme;
};

const applyThemeStyles = (theme: 'light' | 'dark') => {
  if (typeof window === 'undefined') return;

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

const applyThemeToDocument = (resolved: 'light' | 'dark') => {
  if (typeof window === 'undefined') return;
  document.documentElement.setAttribute('data-theme', resolved);
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => applyThemeStyles(resolved));
  } else {
    requestAnimationFrame(() => applyThemeStyles(resolved));
  }
};

interface ThemeStoreState {
  theme: Theme;
  resolvedTheme: 'light' | 'dark';
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeStoreState>()((set, get) => {
  let initialized = false;

  const handleExternalChange = (state: ThemeState) => {
    const resolved = getResolvedTheme(state.theme);
    set({ theme: state.theme, resolvedTheme: resolved });
    applyThemeToDocument(resolved);
  };

  const init = () => {
    if (initialized || typeof window === 'undefined') return;
    initialized = true;

    const initial = initTheme();
    applyThemeToDocument(initial.resolvedTheme);
    subscribe(handleExternalChange);
  };

  if (typeof window !== 'undefined') {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', init);
    } else {
      init();
    }
  }

  return {
    theme: 'system',
    resolvedTheme: 'dark',
    setTheme: (theme: Theme) => {
      const resolved = getResolvedTheme(theme);
      if (import.meta.env.DEV) {
        console.log('Setting theme:', theme, 'resolved:', resolved);
      }
      set({ theme, resolvedTheme: resolved });
      sharedSetTheme(theme);
      applyThemeToDocument(resolved);
    },
    toggleTheme: () => {
      const current = get().resolvedTheme;
      const nextTheme: Theme = current === 'dark' ? 'light' : 'dark';
      const resolved = nextTheme;
      set({ theme: nextTheme, resolvedTheme: resolved });
      sharedSetTheme(nextTheme);
      applyThemeToDocument(resolved);
    },
  };
});
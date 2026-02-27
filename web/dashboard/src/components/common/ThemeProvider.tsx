import { useEffect, useSyncExternalStore } from 'react';
import { useThemeStore } from '@/stores/themeStore';

// Store subscription helper
function subscribe(callback: () => void) {
  window.addEventListener('storage', callback);
  // Also listen for our own store changes
  useThemeStore.subscribe(callback);
  return () => {
    window.removeEventListener('storage', callback);
  };
}

function useStore<T>(selector: (state: ReturnType<typeof useThemeStore.getState>) => T): T {
  const state = useSyncExternalStore(
    subscribe,
    () => selector(useThemeStore.getState()),
    () => selector({ theme: 'dark', resolvedTheme: 'dark', setTheme: () => {}, toggleTheme: () => {} } as any)
  );
  return state;
}

interface ThemeProviderProps {
  children: React.ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  const resolvedTheme = useStore((state) => state.resolvedTheme);

  useEffect(() => {
    const root = document.documentElement;

    // Set the data-theme attribute with resolved theme
    root.setAttribute('data-theme', resolvedTheme);

    // Update meta theme-color for mobile browsers
    const metaThemeColor = document.querySelector('meta[name="theme-color"]');
    if (metaThemeColor) {
      metaThemeColor.setAttribute('content', resolvedTheme === 'dark' ? '#0a0a0f' : '#ffffff');
    }

    // Update body class for additional styling if needed
    document.body.classList.toggle('dark', resolvedTheme === 'dark');
    document.body.classList.toggle('light', resolvedTheme === 'light');
  }, [resolvedTheme]);

  return <>{children}</>;
}

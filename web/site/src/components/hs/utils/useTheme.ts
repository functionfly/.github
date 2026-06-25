import { useEffect, useState } from 'react';

export type Theme = 'light' | 'dark';

/**
 * Read the active theme from `data-theme` on the root html element.
 * Returns 'dark' as the safe default (matches existing Velocity Brand).
 */
export function useTheme(): Theme {
  const [theme, setTheme] = useState<Theme>('dark');

  useEffect(() => {
    if (typeof document === 'undefined') return;
    const read = () => {
      const attr = document.documentElement.getAttribute('data-theme');
      if (attr === 'light') setTheme('light');
      else if (attr === 'dark') setTheme('dark');
      else {
        // No data-theme set: fall back to system preference
        const prefersLight =
          typeof window !== 'undefined' &&
          window.matchMedia &&
          window.matchMedia('(prefers-color-scheme: light)').matches;
        setTheme(prefersLight ? 'light' : 'dark');
      }
    };

    read();

    // Watch for theme changes (set by header toggle)
    const observer = new MutationObserver(read);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['data-theme'],
    });

    // Watch for system preference changes
    const mq = window.matchMedia?.('(prefers-color-scheme: light)');
    const mqHandler = () => read();
    mq?.addEventListener('change', mqHandler);

    return () => {
      observer.disconnect();
      mq?.removeEventListener('change', mqHandler);
    };
  }, []);

  return theme;
}

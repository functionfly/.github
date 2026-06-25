import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from "react";

type Theme = "light" | "dark" | "system";
type ResolvedTheme = "light" | "dark";

interface ThemeState {
  mode: Theme;
  resolved: ResolvedTheme;
}

interface ThemeContextValue {
  theme: ResolvedTheme;
  toggle: () => void;
  setMode: (mode: Theme) => void;
}

const STORAGE_KEY = "ff-user-theme";

const getSystemTheme = (): ResolvedTheme => {
  if (typeof window === "undefined") return "dark";
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
};

const getStoredTheme = (): Theme => {
  if (typeof window === "undefined") return "system";
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as { mode: Theme };
      if (parsed.mode === "light" || parsed.mode === "dark" || parsed.mode === "system") {
        return parsed.mode;
      }
    }
  } catch {}
  return "system";
};

const getResolvedTheme = (theme: Theme): ResolvedTheme => {
  return theme === "system" ? getSystemTheme() : theme;
};

const applyTheme = (mode: Theme) => {
  if (typeof window === "undefined") return;
  const resolved = getResolvedTheme(mode);
  document.documentElement.setAttribute("data-theme", resolved);
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ mode }));
  } catch {}
  window.dispatchEvent(
    new CustomEvent<ThemeState>("ff-theme-change", {
      detail: { mode, resolved },
    }),
  );
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<ResolvedTheme>("dark");

  useEffect(() => {
    const stored = getStoredTheme();
    const resolved = getResolvedTheme(stored);
    setTheme(resolved);

    const handleStorage = (e: StorageEvent) => {
      if (e.key === STORAGE_KEY && e.newValue) {
        try {
          const parsed = JSON.parse(e.newValue) as { mode: Theme };
          setTheme(getResolvedTheme(parsed.mode));
        } catch {}
      }
    };

    const handleCustomEvent = (e: CustomEvent<ThemeState>) => {
      setTheme(e.detail.resolved);
    };

    window.addEventListener("storage", handleStorage);
    window.addEventListener("ff-theme-change", handleCustomEvent as EventListener);

    return () => {
      window.removeEventListener("storage", handleStorage);
      window.removeEventListener("ff-theme-change", handleCustomEvent as EventListener);
    };
  }, []);

  const toggle = useCallback(() => {
    const next: ResolvedTheme = theme === "dark" ? "light" : "dark";
    applyTheme(next);
    setTheme(next);
  }, [theme]);

  const setMode = useCallback((mode: Theme) => {
    applyTheme(mode);
    setTheme(getResolvedTheme(mode));
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, toggle, setMode }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    return { theme: "dark", toggle: () => {}, setMode: () => {} };
  }
  return context;
}

export { type Theme, type ResolvedTheme, getSystemTheme, getResolvedTheme, applyTheme };

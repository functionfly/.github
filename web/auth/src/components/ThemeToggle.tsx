import { useEffect, useState } from "react";

type Theme = "light" | "dark" | "system";
type ResolvedTheme = "light" | "dark";

interface ThemeState {
  mode: Theme;
  resolved: ResolvedTheme;
}

declare global {
  interface Window {
    ffTheme?: {
      init: () => void;
      get: () => ThemeState;
      set: (mode: Theme) => void;
      toggle: () => void;
      subscribe: (callback: (state: ThemeState) => void) => () => void;
    };
  }
}

const getSystemTheme = (): ResolvedTheme => {
  if (typeof window === "undefined") return "dark";
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
};

const getStoredTheme = (): Theme => {
  if (typeof window === "undefined") return "system";
  try {
    const stored = localStorage.getItem("ff-user-theme");
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

const toggleTheme = () => {
  if (typeof window === "undefined") return;
  const current = getResolvedTheme(getStoredTheme());
  const next: ResolvedTheme = current === "dark" ? "light" : "dark";

  document.documentElement.setAttribute("data-theme", next);

  try {
    localStorage.setItem("ff-user-theme", JSON.stringify({ mode: next }));
  } catch {}
};

export default function ThemeToggle() {
  const [theme, setTheme] = useState<ResolvedTheme>("dark");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const resolved = getResolvedTheme(getStoredTheme());
    setTheme(resolved);
    setMounted(true);

    const handleStorage = (e: StorageEvent) => {
      if (e.key === "ff-user-theme" && e.newValue) {
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

  if (!mounted) {
    return (
      <button
        className="theme-toggle"
        aria-label="Toggle theme"
        style={{
          width: 36,
          height: 36,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "transparent",
          border: "1px solid var(--ff-border-default)",
          borderRadius: "var(--ff-radius-md)",
          cursor: "pointer",
          color: "var(--ff-secondary-text)",
          transition: "all var(--ff-transition-fast)",
        }}
      >
        <span style={{ width: 18, height: 18 }} />
      </button>
    );
  }

  return (
    <button
      className="theme-toggle"
      onClick={toggleTheme}
      aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
      title={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
      style={{
        width: 36,
        height: 36,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "transparent",
        border: "1px solid var(--ff-border-default)",
        borderRadius: "var(--ff-radius-md)",
        cursor: "pointer",
        color: "var(--ff-secondary-text)",
        transition: "all var(--ff-transition-fast)",
      }}
    >
      {theme === "dark" ? (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
        </svg>
      ) : (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
        </svg>
      )}
    </button>
  );
}
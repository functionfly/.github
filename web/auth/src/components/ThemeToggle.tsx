import { useEffect, useState } from "react";

type Theme = "light" | "dark";

const STORAGE_KEY = "theme-storage";

function getSystemTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function getStoredTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed.state?.theme) {
        return parsed.state.theme === "system" ? getSystemTheme() : parsed.state.theme;
      }
    } catch {}
  }
  return "system" as Theme;
}

function getResolvedTheme(theme: "light" | "dark" | "system"): Theme {
  return theme === "system" ? getSystemTheme() : theme;
}

function toggleTheme() {
  const current = getResolvedTheme(getStoredTheme() as Theme);
  const next: Theme = current === "dark" ? "light" : "dark";

  document.documentElement.setAttribute("data-theme", next);

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    let currentState = { state: { theme: "system" } };
    if (stored) {
      currentState = JSON.parse(stored);
    }
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ ...currentState, state: { ...currentState.state, theme: next } })
    );
  } catch {}
}

export default function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>("dark");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const resolved = getResolvedTheme(getStoredTheme() as Theme);
    setTheme(resolved);
    setMounted(true);
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

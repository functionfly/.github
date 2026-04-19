import type { JSX } from "react";
import { useCallback, useEffect, useState } from "react";
import type { LinkProps, OutletProps } from "react-router-dom";
import { Link as RouterLink, Outlet as RouterOutlet } from "react-router-dom";
import "./Layout.css";

// Cast to work around React Router v7 + React 19 type incompatibility
const Link = RouterLink as (props: LinkProps) => JSX.Element;
const Outlet = RouterOutlet as (props: OutletProps) => JSX.Element;

// Inline SVG icons to avoid lucide-react dependency
const Icons = {
  BookOpen: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
      <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
    </svg>
  ),
  Code: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="16 18 22 12 16 6" />
      <polyline points="8 6 2 12 8 18" />
    </svg>
  ),
  Command: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M15 6v12a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3V6a3 3 0 1 0-3 3h12a3 3 0 1 0-3-3" />
    </svg>
  ),
  Github: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22" />
    </svg>
  ),
  Menu: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="4" x2="20" y1="12" y2="12" />
      <line x1="4" x2="20" y1="6" y2="6" />
      <line x1="4" x2="20" y1="18" y2="18" />
    </svg>
  ),
  Search: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="11" cy="11" r="8" />
      <path d="m21 21-4.3-4.3" />
    </svg>
  ),
  Twitter: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M22 4s-.7 2.1-2 3.4c1.6 10-9.4 17.3-18 11.6 2.2.1 4.4-.6 6-2C3 15.5.5 9.6 3 5c2.2 2.6 5.6 4.1 9 4-.9-4.2 4-6.6 7-3.8 1.1 0 3-1.2 3-1.2z" />
    </svg>
  ),
  X: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  ),
  Zap: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
    </svg>
  ),
  Terminal: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" x2="20" y1="19" y2="19" />
    </svg>
  ),
};

// Quick action shortcuts
const QUICK_ACTIONS = [
  { key: "s", label: "Search Docs", icon: Icons.Search },
  { key: "g", label: "GitHub", icon: Icons.Github },
  { key: "f", label: "Functions", icon: Icons.Code },
  { key: "a", label: "API Reference", icon: Icons.Terminal },
];

export default function Layout() {
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Command palette: Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setShowCommandPalette(true);
      }

      // Close on escape
      if (e.key === "Escape" && showCommandPalette) {
        setShowCommandPalette(false);
      }

      // Quick navigation
      if (e.metaKey || e.ctrlKey) {
        if (e.key === "g") {
          e.preventDefault();
          window.open("https://github.com/functionfly/", "_blank");
        }
      }
    },
    [showCommandPalette],
  );

  useEffect(() => {
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div className="layout">
      <header className="header bg-aviation-bg-primary/95 backdrop-blur-xl border-b border-aviation-border-panel">
        <div className="container header-content">
          <Link to="/" className="logo group">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-linear-to-br from-aviation-amber to-aviation-amber-glow group-hover:shadow-lg group-hover:shadow-aviation-amber/20 transition-shadow">
              <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
                <circle
                  cx="16"
                  cy="16"
                  r="14"
                  stroke="currentColor"
                  strokeWidth="2"
                  className="text-aviation-bg-primary"
                />
                <path
                  d="M10 16L14 20L22 12"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="text-aviation-bg-primary"
                />
              </svg>
            </div>
            <span className="font-bold text-aviation-text-primary">
              FunctionFly
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="nav hidden md:flex">
            <Link
              to="/"
              className="nav-link text-aviation-text-secondary hover:text-aviation-text-primary transition-colors flex items-center gap-2"
            >
              <Icons.BookOpen />
              Docs
            </Link>
            <a
              href="https://functionfly.com/registry"
              className="nav-link text-aviation-text-secondary hover:text-aviation-text-primary transition-colors flex items-center gap-2"
            >
              <Icons.Code />
              Functions
            </a>
            <a
              href="https://functionfly.com"
              className="nav-link text-aviation-text-secondary hover:text-aviation-text-primary transition-colors flex items-center gap-2"
            >
              <Icons.Zap />
              Home
            </a>
            <Link
              to="/api-reference"
              className="nav-link text-aviation-text-secondary hover:text-aviation-text-primary transition-colors flex items-center gap-2"
            >
              <Icons.Terminal />
              API Reference
            </Link>

            {/* Command Palette Trigger */}
            <button
              onClick={() => setShowCommandPalette(true)}
              className="flex items-center gap-2 px-3 py-1.5 ml-4 text-sm text-aviation-text-muted bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-all"
            >
              <Icons.Command />
              <span className="hidden lg:inline">Search</span>
              <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                ⌘K
              </kbd>
            </button>
          </nav>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            {mobileMenuOpen ? <Icons.X /> : <Icons.Menu />}
          </button>
        </div>
      </header>

      {/* Mobile Menu */}
      <div
        className={`md:hidden bg-aviation-bg-primary border-b border-aviation-border-panel overflow-hidden transition-all duration-300 ease-in-out ${
          mobileMenuOpen ? "max-h-96 opacity-100" : "max-h-0 opacity-0"
        }`}
      >
        <nav className="flex flex-col p-4 space-y-2">
          <Link
            to="/"
            onClick={() => setMobileMenuOpen(false)}
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.BookOpen />
            Docs
          </Link>
          <a
            href="https://functionfly.com/registry"
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.Code />
            Functions
          </a>
          <a
            href="https://functionfly.com"
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.Zap />
            Home
          </a>
          <Link
            to="/api-reference"
            onClick={() => setMobileMenuOpen(false)}
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.Terminal />
            API Reference
          </Link>
          <hr className="border-aviation-border-panel" />
          <a
            href="https://github.com/functionfly/functionfly"
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.Github />
            GitHub
          </a>
          <a
            href="https://twitter.com/functionfly"
            className="flex items-center gap-2 px-3 py-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument rounded-lg transition-colors"
          >
            <Icons.Twitter />
            Twitter
          </a>
        </nav>
      </div>

      <main className="main">
        <Outlet />
      </main>

      <footer className="footer bg-aviation-bg-secondary border-t border-aviation-border-panel">
        <div className="container py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-aviation-text-secondary text-sm">
              &copy; {new Date().getFullYear()} FunctionFly. Open source
              serverless functions.
            </p>
            <div className="flex items-center gap-4">
              <a
                href="https://github.com/functionfly/functionfly"
                className="text-aviation-text-secondary hover:text-aviation-text-primary transition-colors"
              >
                <Icons.Github />
              </a>
              <a
                href="https://twitter.com/functionfly"
                className="text-aviation-text-secondary hover:text-aviation-text-primary transition-colors"
              >
                <Icons.Twitter />
              </a>
            </div>
          </div>
        </div>
      </footer>

      {/* Command Palette Overlay */}
      {showCommandPalette && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-start justify-center pt-[20vh] animate-in fade-in duration-200"
          onClick={() => setShowCommandPalette(false)}
        >
          <div
            className="w-full max-w-2xl mx-4 bg-aviation-bg-primary border border-aviation-border-panel rounded-xl shadow-2xl overflow-hidden animate-in slide-in-from-top-5 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Search Input */}
            <div className="flex items-center gap-3 px-4 py-4 border-b border-aviation-border-panel">
              <Icons.Command />
              <input
                type="text"
                placeholder="Search documentation..."
                className="flex-1 text-base text-aviation-text-primary placeholder:text-aviation-text-dim bg-transparent focus:outline-none"
                autoFocus
              />
              <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-2 py-1 rounded">
                ESC
              </kbd>
            </div>

            {/* Quick Actions */}
            <div className="p-2">
              <p className="px-3 py-2 text-xs font-semibold text-aviation-text-muted uppercase tracking-wider">
                Quick Actions
              </p>
              <div className="space-y-1">
                {QUICK_ACTIONS.map((action) => (
                  <button
                    key={action.key}
                    onClick={() => {
                      setShowCommandPalette(false);
                      if (action.key === "s") {
                        // Focus search or navigate to search page
                        const searchInput = document.querySelector(
                          'input[type="search"]',
                        ) as HTMLInputElement;
                        if (searchInput) searchInput.focus();
                      }
                      if (action.key === "g") {
                        window.open(
                          "https://github.com/functionfly/functionfly",
                          "_blank",
                        );
                      }
                      if (action.key === "f") {
                        window.location.href =
                          "https://functionfly.com/registry";
                      }
                      if (action.key === "a") {
                        window.location.href = "/api-reference";
                      }
                    }}
                    className="w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <action.icon />
                      <span>{action.label}</span>
                    </div>
                    <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                      ⌘{action.key.toUpperCase()}
                    </kbd>
                  </button>
                ))}
              </div>
            </div>

            {/* Footer */}
            <div className="px-4 py-3 bg-aviation-bg-secondary border-t border-aviation-border-panel text-xs text-aviation-text-muted">
              <p className="flex items-center gap-2">
                <span>Use</span>
                <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">
                  ↑
                </kbd>
                <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">
                  ↓
                </kbd>
                <span>to navigate,</span>
                <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">
                  ↵
                </kbd>
                <span>to select</span>
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

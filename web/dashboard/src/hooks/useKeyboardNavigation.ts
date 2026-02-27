import { useEffect, useCallback } from 'react';
import { useLocation } from 'react-router-dom';
import { ROUTES } from '@/lib/constants';

interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  action: () => void;
  description: string;
}

interface KeyboardNavigationConfig {
  shortcuts?: KeyboardShortcut[];
  onArrowUp?: () => void;
  onArrowDown?: () => void;
  onArrowLeft?: () => void;
  onArrowRight?: () => void;
  onEnter?: () => void;
  onEscape?: () => void;
  onTab?: () => void;
  enabled?: boolean;
}

export function useKeyboardNavigation(config: KeyboardNavigationConfig = {}) {
  const {
    shortcuts = [],
    onArrowUp,
    onArrowDown,
    onArrowLeft,
    onArrowRight,
    onEnter,
    onEscape,
    onTab,
    enabled = true,
  } = config;

  const location = useLocation();

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!enabled) return;

    // Handle custom shortcuts first
    for (const shortcut of shortcuts) {
      if (
        e.key === shortcut.key &&
        !!e.ctrlKey === !!shortcut.ctrlKey &&
        !!e.metaKey === !!shortcut.metaKey &&
        !!e.shiftKey === !!shortcut.shiftKey &&
        !!e.altKey === !!shortcut.altKey
      ) {
        e.preventDefault();
        shortcut.action();
        return;
      }
    }

    // Handle arrow keys and special keys
    switch (e.key) {
      case 'ArrowUp':
        if (onArrowUp) {
          e.preventDefault();
          onArrowUp();
        }
        break;
      case 'ArrowDown':
        if (onArrowDown) {
          e.preventDefault();
          onArrowDown();
        }
        break;
      case 'ArrowLeft':
        if (onArrowLeft) {
          e.preventDefault();
          onArrowLeft();
        }
        break;
      case 'ArrowRight':
        if (onArrowRight) {
          e.preventDefault();
          onArrowRight();
        }
        break;
      case 'Enter':
        if (onEnter) {
          e.preventDefault();
          onEnter();
        }
        break;
      case 'Escape':
        if (onEscape) {
          e.preventDefault();
          onEscape();
        }
        break;
      case 'Tab':
        if (onTab) {
          e.preventDefault();
          onTab();
        }
        break;
    }
  }, [shortcuts, onArrowUp, onArrowDown, onArrowLeft, onArrowRight, onEnter, onEscape, onTab, enabled]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return {
    // Return default shortcuts that can be used globally
    defaultShortcuts: [
      {
        key: 'g',
        action: () => {
          // Go to dashboard - implement in component
        },
        description: 'Go to Dashboard',
      },
      {
        key: 'f',
        action: () => {
          // Focus functions - implement in component
        },
        description: 'Go to Functions',
      },
      {
        key: 'p',
        action: () => {
          // Go to providers - implement in component
        },
        description: 'Go to Providers',
      },
      {
        key: 'a',
        action: () => {
          // Go to analytics - implement in component
        },
        description: 'Go to Analytics',
      },
      {
        key: 's',
        action: () => {
          // Go to settings - implement in component
        },
        description: 'Go to Settings',
      },
    ].filter(shortcut => {
      // Only include shortcuts that make sense for current page
      switch (shortcut.key) {
        case 'g':
          return location.pathname !== ROUTES.DASHBOARD;
        case 'f':
          return location.pathname !== ROUTES.FUNCTIONS;
        case 'p':
          return location.pathname !== ROUTES.PROVIDERS;
        case 'a':
          return location.pathname !== ROUTES.ANALYTICS;
        case 's':
          return location.pathname !== ROUTES.SETTINGS;
        default:
          return true;
      }
    }),
  };
}
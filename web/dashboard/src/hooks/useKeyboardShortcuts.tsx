'use client';

import { useEffect, useCallback } from 'react';

interface KeyboardShortcutsOptions {
  onNext?: () => void;
  onPrevious?: () => void;
  onSelect?: () => void;
  onEscape?: () => void;
  onRefresh?: () => void;
  enabled?: boolean;
}

export function useKeyboardShortcuts({
  onNext,
  onPrevious,
  onSelect,
  onEscape,
  onRefresh,
  enabled = true,
}: KeyboardShortcutsOptions) {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (!enabled) return;

      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }

      switch (event.key) {
        case 'j':
        case 'ArrowDown':
          event.preventDefault();
          onNext?.();
          break;
        case 'k':
        case 'ArrowUp':
          event.preventDefault();
          onPrevious?.();
          break;
        case 'Enter':
        case ' ':
          event.preventDefault();
          onSelect?.();
          break;
        case 'Escape':
          event.preventDefault();
          onEscape?.();
          break;
        case 'r':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            onRefresh?.();
          }
          break;
      }
    },
    [enabled, onNext, onPrevious, onSelect, onEscape, onRefresh]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
}

export function KeyboardShortcutsHelp() {
  return (
    <div className="text-xs text-muted-foreground space-y-1">
      <p>
        <kbd className="px-1 py-0.5 rounded bg-muted font-mono text-[10px]">j</kbd> /{' '}
        <kbd className="px-1 py-0.5 rounded bg-muted font-mono text-[10px]">k</kbd> - Navigate
      </p>
      <p>
        <kbd className="px-1 py-0.5 rounded bg-muted font-mono text-[10px]">Enter</kbd> - Select
      </p>
      <p>
        <kbd className="px-1 py-0.5 rounded bg-muted font-mono text-[10px]">Esc</kbd> - Close/Cancel
      </p>
    </div>
  );
}

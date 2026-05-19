import { useEffect, useCallback } from "react";

export interface KeyboardShortcut {
  key: string;
  meta?: boolean;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  action: () => void;
  description: string;
}

export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        const metaMatch = shortcut.meta ? e.metaKey : !shortcut.meta;
        const ctrlMatch = shortcut.ctrl ? e.ctrlKey : !shortcut.ctrl;
        const shiftMatch = shortcut.shift ? e.shiftKey : !shortcut.shift;
        const altMatch = shortcut.alt ? e.altKey : !shortcut.alt;

        if (
          e.key.toLowerCase() === shortcut.key.toLowerCase() &&
          metaMatch &&
          ctrlMatch &&
          shiftMatch &&
          altMatch
        ) {
          e.preventDefault();
          shortcut.action();
          return;
        }
      }
    },
    [shortcuts]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);
}

export const DEFAULT_STUDIO_SHORTCUTS = {
  COMMAND_PALETTE: { key: "k", meta: true, description: "Open command palette" },
  SAVE: { key: "s", meta: true, description: "Save current file" },
  RUN: { key: "Enter", meta: true, description: "Run workflow" },
  FORMAT: { key: "f", meta: true, shift: true, description: "Format code" },
  UNDO: { key: "z", meta: true, description: "Undo" },
  REDO: { key: "z", meta: true, shift: true, description: "Redo" },
  TOGGLE_LEFT_PANEL: { key: "b", meta: true, description: "Toggle left panel" },
  TOGGLE_RIGHT_PANEL: { key: "j", meta: true, description: "Toggle right panel" },
  TOGGLE_BOTTOM_PANEL: { key: "i", meta: true, description: "Toggle bottom panel" },
  FOCUS_EDITOR: { key: "e", meta: true, description: "Focus editor" },
  CLOSE_PANEL: { key: "Escape", description: "Close/cancel" },
} as const;
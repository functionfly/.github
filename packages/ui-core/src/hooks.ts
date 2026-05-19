/**
 * @functionfly/ui-core
 * Utility hook for keyboard shortcuts
 */

import * as React from "react";

interface ShortcutHandler {
  key: string;
  modifiers?: { ctrl?: boolean; shift?: boolean; alt?: boolean; meta?: boolean };
  handler: (e: KeyboardEvent) => void;
}

export function useKeyboardShortcuts(shortcuts: ShortcutHandler[], deps: React.DependencyList = []) {
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        const { key, modifiers = {}, handler } = shortcut;

        const ctrlMatch = modifiers.ctrl ? e.ctrlKey || e.metaKey : !e.ctrlKey && !e.metaKey;
        const shiftMatch = modifiers.shift ? e.shiftKey : !e.shiftKey;
        const altMatch = modifiers.alt ? e.altKey : !e.altKey;

        if (e.key === key && ctrlMatch && shiftMatch && altMatch) {
          e.preventDefault();
          handler(e);
          return;
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [shortcuts, ...deps]);
}

// --- Animation hooks ---
export function useAnimationFrame(callback: (timestamp: number) => void) {
  const callbackRef = React.useRef(callback);
  callbackRef.current = callback;

  React.useEffect(() => {
    let frameId: number;
    const loop = (time: number) => {
      callbackRef.current(time);
      frameId = requestAnimationFrame(loop);
    };
    frameId = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(frameId);
  }, []);
}

export function useDebounce<T extends (...args: any[]) => any>(fn: T, delay: number): (...args: Parameters<T>) => void {
  const timer = React.useRef<ReturnType<typeof setTimeout> | null>(null);
  return React.useCallback(
    (...args: Parameters<T>) => {
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => {
        fn(...args);
        timer.current = null;
      }, delay);
    },
    [fn, delay]
  );
}

export function useLocalStorage<T>(key: string, initialValue: T): [T, (value: T) => void] {
  const [value, setValue] = React.useState<T>(() => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch {
      return initialValue;
    }
  });

  const setStoredValue = React.useCallback(
    (newValue: T) => {
      setValue(newValue);
      try {
        localStorage.setItem(key, JSON.stringify(newValue));
      } catch (err) {
        console.warn(`Failed to save ${key} to localStorage:`, err);
      }
    },
    [key]
  );

  return [value, setStoredValue];
}

// --- Canvas zoom/pan state management ---
export interface CanvasTransform {
  x: number;
  y: number;
  zoom: number;
}

export function useCanvasTransform(
  initialTransform: CanvasTransform = { x: 0, y: 0, zoom: 1 }
): [CanvasTransform, { setTransform: (t: CanvasTransform) => void; zoomIn: () => void; zoomOut: () => void; resetView: () => void; fitView: (nodes: { x: number; y: number; width: number; height: number }[]) => void }] {
  const [transform, setTransformState] = React.useState<CanvasTransform>(initialTransform);

  const setTransform = React.useCallback((newTransform: CanvasTransform | ((prev: CanvasTransform) => CanvasTransform)) => {
    setTransformState((prev) => {
      const next = typeof newTransform === "function" ? newTransform(prev) : newTransform;
      return { x: next.x, y: next.y, zoom: Math.max(0.1, Math.min(10, next.zoom)) };
    });
  }, []);

  const zoomIn = React.useCallback(() => {
    setTransformState((prev) => ({ ...prev, zoom: Math.min(prev.zoom * 1.2, 10) }));
  }, []);

  const zoomOut = React.useCallback(() => {
    setTransformState((prev) => ({ ...prev, zoom: Math.max(prev.zoom / 1.2, 0.1) }));
  }, []);

  const resetView = React.useCallback(() => {
    setTransformState(initialTransform);
  }, [initialTransform]);

  const fitView = React.useCallback((nodes: { x: number; y: number; width: number; height: number }[]) => {
    if (nodes.length === 0) return;
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    for (const n of nodes) {
      minX = Math.min(minX, n.x);
      minY = Math.min(minY, n.y);
      maxX = Math.max(maxX, n.x + n.width);
      maxY = Math.max(maxY, n.y + n.height);
    }
    const padding = 100;
    const width = maxX - minX + padding * 2;
    const height = maxY - minY + padding * 2;
    setTransformState({
      x: -(minX + width / 2),
      y: -(minY + height / 2),
      zoom: 1,
    });
  }, []);

  return [transform, { setTransform, zoomIn, zoomOut, resetView, fitView }];
}
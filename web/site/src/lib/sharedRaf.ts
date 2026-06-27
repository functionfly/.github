import { useEffect, useRef, useCallback } from 'react';

type Listener = (deltaSeconds: number) => void;

const listeners = new Set<Listener>();
let rafId: number | null = null;
let lastTimestamp: number | null = null;

function tick(now: number): void {
  if (lastTimestamp === null) {
    lastTimestamp = now;
  }
  const deltaSeconds = (now - lastTimestamp) / 1000;
  lastTimestamp = now;
  for (const listener of listeners) {
    listener(deltaSeconds);
  }
  if (listeners.size > 0) {
    rafId = requestAnimationFrame(tick);
  } else {
    rafId = null;
    lastTimestamp = null;
  }
}

function ensureRunning(): void {
  if (rafId === null && listeners.size > 0 && typeof window !== 'undefined') {
    rafId = requestAnimationFrame(tick);
  }
}

function subscribe(listener: Listener): () => void {
  listeners.add(listener);
  ensureRunning();
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0 && rafId !== null) {
      cancelAnimationFrame(rafId);
      rafId = null;
      lastTimestamp = null;
    }
  };
}

export function useSharedAnimationFrame(callback: (deltaSeconds: number) => void): void {
  const callbackRef = useRef(callback);
  callbackRef.current = callback;

  const stableCallback = useCallback((deltaSeconds: number) => {
    callbackRef.current(deltaSeconds);
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    return subscribe(stableCallback);
  }, [stableCallback]);
}

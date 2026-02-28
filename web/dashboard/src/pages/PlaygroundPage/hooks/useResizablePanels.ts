import { useState, useCallback, useRef, useEffect } from 'react';

interface PanelSizes {
  input: number;
  output: number;
}

const STORAGE_KEY = 'playground-panel-sizes';
const MIN_PANEL_SIZE = 20; // percent
const MAX_PANEL_SIZE = 80; // percent
const DEFAULT_SIZES: PanelSizes = { input: 50, output: 50 };

function loadSizes(): PanelSizes {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as PanelSizes;
      if (
        typeof parsed.input === 'number' &&
        typeof parsed.output === 'number' &&
        parsed.input >= MIN_PANEL_SIZE &&
        parsed.input <= MAX_PANEL_SIZE
      ) {
        return parsed;
      }
    }
  } catch {
    // ignore
  }
  return DEFAULT_SIZES;
}

function saveSizes(sizes: PanelSizes) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sizes));
  } catch {
    // ignore
  }
}

export function useResizablePanels() {
  const [sizes, setSizes] = useState<PanelSizes>(loadSizes);
  const isDragging = useRef(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const startX = useRef(0);
  const startSizes = useRef<PanelSizes>(DEFAULT_SIZES);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDragging.current = true;
    startX.current = e.clientX;
    startSizes.current = { ...sizes };
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }, [sizes]);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging.current || !containerRef.current) return;

    const containerWidth = containerRef.current.offsetWidth;
    const deltaX = e.clientX - startX.current;
    const deltaPercent = (deltaX / containerWidth) * 100;

    const newInputSize = Math.min(
      MAX_PANEL_SIZE,
      Math.max(MIN_PANEL_SIZE, startSizes.current.input + deltaPercent)
    );
    const newOutputSize = 100 - newInputSize;

    setSizes({ input: newInputSize, output: newOutputSize });
  }, []);

  const handleMouseUp = useCallback(() => {
    if (!isDragging.current) return;
    isDragging.current = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    saveSizes(sizes);
  }, [sizes]);

  // Save sizes when they change
  useEffect(() => {
    if (!isDragging.current) {
      saveSizes(sizes);
    }
  }, [sizes]);

  useEffect(() => {
    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [handleMouseMove, handleMouseUp]);

  const resetToEqual = useCallback(() => {
    setSizes(DEFAULT_SIZES);
    saveSizes(DEFAULT_SIZES);
  }, []);

  return {
    sizes,
    containerRef,
    handleMouseDown,
    resetToEqual,
    isDragging: isDragging.current,
  };
}

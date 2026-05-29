import { useMemo } from 'react';
import { useWindowSize } from './useWindowSize';

type Options = {
  /** Minimum rows per page */
  min?: number;
  /** Maximum rows per page */
  max?: number;
  /** Approximate row height in px */
  rowHeight?: number;
  /** Fixed chrome above the list (headers, search, etc.) */
  reservedHeight?: number;
};

/**
 * Derives a page size from viewport dimensions so lists stay readable without
 * rendering the entire catalog at once.
 */
export function useResponsivePageSize(options: Options = {}): number {
  const { min = 5, max = 20, rowHeight = 56, reservedHeight } = options;
  const [width, height] = useWindowSize();

  return useMemo(() => {
    const reserved =
      reservedHeight ?? (width < 640 ? 520 : width < 1024 ? 460 : width < 1280 ? 420 : 400);
    const available = Math.max(height - reserved, rowHeight * min);
    const rows = Math.floor(available / rowHeight);
    return Math.max(min, Math.min(rows, max));
  }, [width, height, min, max, rowHeight, reservedHeight]);
}

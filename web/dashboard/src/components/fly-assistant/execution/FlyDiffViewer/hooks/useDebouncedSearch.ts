/**
 * Debounced search hook
 */

import { useState, useEffect } from "react";

export function useDebouncedSearch(initialValue: string = "", delay: number = 300) {
  const [value, setValue] = useState(initialValue);
  const [debouncedValue, setDebouncedValue] = useState(initialValue);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return [debouncedValue, setValue] as const;
}
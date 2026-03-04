/**
 * Search functionality hook
 */

import { useMemo } from "react";
import type { DiffLine } from "../types";
import { highlightSearch } from "../utils/formatters";

export function useSearch(searchTerm: string, lines: DiffLine[]): DiffLine[] {
  return useMemo(() => {
    if (!searchTerm.trim()) return lines;

    return lines.map((line) => ({
      ...line,
      searchHighlights: highlightSearch(line.content, searchTerm),
    }));
  }, [lines, searchTerm]);
}
/**
 * Collapse/expand functionality hook
 */

import { useState, useEffect } from "react";
import type { DiffHunk } from "../types";

export function useCollapse(
  hunks: DiffHunk[],
  enableCollapse: boolean,
  maxLinesBeforeCollapse: number
): {
  collapsedHunks: Set<number>;
  toggleHunkCollapse: (hunkIndex: number) => void;
  expandAll: () => void;
  collapseAll: () => void;
} {
  const [collapsedHunks, setCollapsedHunks] = useState<Set<number>>(new Set());

  // Auto-collapse large hunks
  useEffect(() => {
    if (!enableCollapse) return;

    const newCollapsed = new Set<number>();
    hunks.forEach((hunk, index) => {
      if (hunk.lines.length > maxLinesBeforeCollapse) {
        newCollapsed.add(index);
      }
    });
    setCollapsedHunks(newCollapsed);
  }, [hunks, enableCollapse, maxLinesBeforeCollapse]);

  const toggleHunkCollapse = (hunkIndex: number) => {
    setCollapsedHunks((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(hunkIndex)) {
        newSet.delete(hunkIndex);
      } else {
        newSet.add(hunkIndex);
      }
      return newSet;
    });
  };

  const expandAll = () => {
    setCollapsedHunks(new Set());
  };

  const collapseAll = () => {
    setCollapsedHunks(new Set(hunks.map((_, index) => index)));
  };

  return { collapsedHunks, toggleHunkCollapse, expandAll, collapseAll };
}
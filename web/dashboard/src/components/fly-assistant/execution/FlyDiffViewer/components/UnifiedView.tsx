/**
 * UnifiedView component for inline diff display
 */

import React, { useRef, useMemo } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { ChevronDown, ChevronUp } from "lucide-react";
import { cn } from "@/lib/utils";
import type { DiffHunk, ThemeMode } from "../types";
import { DiffLineRow } from "./DiffLineRow";

interface UnifiedViewProps {
  hunks: DiffHunk[];
  showLineNumbers: boolean;
  enableSyntaxHighlighting?: boolean;
  language?: string;
  theme?: ThemeMode;
  searchTerm?: string;
  collapsedHunks?: Set<number>;
  onToggleCollapse?: (hunkIndex: number) => void;
  maxHeight?: number;
}

export const UnifiedView = React.memo<UnifiedViewProps>(({
  hunks,
  showLineNumbers,
  enableSyntaxHighlighting,
  language,
  theme,
  searchTerm,
  collapsedHunks = new Set(),
  onToggleCollapse,
  maxHeight
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  // Flatten hunks for virtualization
  const { flattenedLines } = useMemo(() => {
    const lines: { line: any; hunkIndex: number; lineIndex: number; isHeader?: boolean; headerContent?: string }[] = [];

    hunks.forEach((hunk, hunkIndex) => {
      if (collapsedHunks.has(hunkIndex)) {
        // Add collapsed placeholder
        lines.push({
          line: { content: `@@ ${hunk.lines.length} lines collapsed @@`, type: "unchanged" },
          hunkIndex,
          lineIndex: -1,
          isHeader: true,
          headerContent: `@@ -${hunk.oldRange.start},${hunk.oldRange.count} +${hunk.newRange.start},${hunk.newRange.count} @@ (${hunk.lines.length} lines collapsed)`,
        });
        return;
      }

      // Add hunk header
      lines.push({
        line: { content: "", type: "unchanged" },
        hunkIndex,
        lineIndex: -1,
        isHeader: true,
        headerContent: `@@ -${hunk.oldRange.start},${hunk.oldRange.count} +${hunk.newRange.start},${hunk.newRange.count} @@`,
      });

      // Add hunk lines
      hunk.lines.forEach((line, lineIndex) => {
        lines.push({ line, hunkIndex, lineIndex });
      });
    });

    return { flattenedLines: lines };
  }, [hunks, collapsedHunks]);

  // Virtualization for large diffs
  const virtualizer = useVirtualizer({
    count: flattenedLines.length,
    getScrollElement: () => containerRef.current,
    estimateSize: (index) => flattenedLines[index]?.isHeader ? 32 : 24, // Headers are taller
    overscan: 5,
  });

  return (
    <div ref={containerRef} className="overflow-auto" style={{ maxHeight }}>
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => {
          const item = flattenedLines[virtualItem.index];
          if (!item) return null;

          if (item.isHeader) {
            return (
              <div
                key={`header-${item.hunkIndex}`}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  height: `${virtualItem.size}px`,
                  transform: `translateY(${virtualItem.start}px)`,
                }}
                className={cn(
                  "px-4 py-1 bg-(--color-bg-tertiary) text-xs text-text-muted font-mono border-y border-(--color-border)",
                  "flex items-center justify-between cursor-pointer hover:bg-(--color-bg-secondary)"
                )}
                onClick={() => onToggleCollapse?.(item.hunkIndex)}
                role="button"
                tabIndex={0}
                aria-label={`Hunk ${item.hunkIndex + 1}: ${item.headerContent}`}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    onToggleCollapse?.(item.hunkIndex);
                  }
                }}
              >
                <span>{item.headerContent}</span>
                <div className="flex items-center gap-2">
                  {collapsedHunks.has(item.hunkIndex) ? (
                    <ChevronUp className="h-3 w-3" />
                  ) : (
                    <ChevronDown className="h-3 w-3" />
                  )}
                </div>
              </div>
            );
          }

          return (
            <div
              key={`${item.hunkIndex}-${item.lineIndex}`}
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                width: "100%",
                height: `${virtualItem.size}px`,
                transform: `translateY(${virtualItem.start}px)`,
              }}
            >
              <DiffLineRow
                line={item.line}
                showLineNumbers={showLineNumbers}
                enableSyntaxHighlighting={enableSyntaxHighlighting}
                language={language}
                theme={theme}
                searchTerm={searchTerm}
              />
            </div>
          );
        })}
      </div>
    </div>
  );
});

UnifiedView.displayName = "UnifiedView";
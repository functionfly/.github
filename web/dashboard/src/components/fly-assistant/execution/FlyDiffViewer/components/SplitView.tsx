/**
 * SplitView component for side-by-side diff display
 */

import React, { useRef, useMemo } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { DiffHunk, ThemeMode } from "../types";
import { DiffLineRow } from "./DiffLineRow";

interface SplitViewProps {
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

export const SplitView = React.memo<SplitViewProps>(({
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

  // Separate added and removed lines for side-by-side display
  const { leftLines, rightLines } = useMemo(() => {
    const left: any[] = [];
    const right: any[] = [];

    hunks.forEach((hunk) => {
      if (collapsedHunks.has(hunks.indexOf(hunk))) return;

      hunk.lines.forEach((line) => {
        if (line.type === "removed") {
          left.push(line);
          right.push({
            content: "",
            type: "unchanged",
          });
        } else if (line.type === "added") {
          left.push({
            content: "",
            type: "unchanged",
          });
          right.push(line);
        } else {
          left.push(line);
          right.push(line);
        }
      });
    });

    return { leftLines: left, rightLines: right };
  }, [hunks, collapsedHunks]);

  // Virtualization for large diffs
  const virtualizer = useVirtualizer({
    count: Math.max(leftLines.length, rightLines.length),
    getScrollElement: () => containerRef.current,
    estimateSize: () => 24, // Estimated line height
    overscan: 5,
  });

  return (
    <div ref={containerRef} className="flex flex-col sm:flex-row overflow-auto" style={{ maxHeight }}>
      {/* Left side (before) */}
      <div className="flex-1 border-b sm:border-b-0 sm:border-r border-(--color-border)">
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            position: "relative",
          }}
        >
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const line = leftLines[virtualItem.index];
            if (!line) return null;

            return (
              <div
                key={`left-${virtualItem.index}`}
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
                  line={{ ...line, type: line.content ? line.type : "unchanged" }}
                  showLineNumbers={showLineNumbers}
                  isSplit
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

      {/* Right side (after) */}
      <div className="flex-1">
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            position: "relative",
          }}
        >
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const line = rightLines[virtualItem.index];
            if (!line) return null;

            return (
              <div
                key={`right-${virtualItem.index}`}
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
                  line={{ ...line, type: line.content ? line.type : "unchanged" }}
                  showLineNumbers={showLineNumbers}
                  isSplit
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
    </div>
  );
});

SplitView.displayName = "SplitView";
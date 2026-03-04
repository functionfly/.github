/**
 * DiffLineRow component for rendering individual diff lines
 */

import React from "react";
import { cn } from "@/lib/utils";
import type { DiffLine } from "../types";
import { CHANGE_TYPE_STYLES } from "../constants";
import { createSyntaxHighlightedContent } from "../utils/syntaxHighlighting";
import { renderContentWithHighlights, renderWordDiffs } from "../utils/formatters";

interface DiffLineProps {
  line: DiffLine;
  showLineNumbers: boolean;
  isSplit?: boolean;
  enableSyntaxHighlighting?: boolean;
  language?: string;
  theme?: "light" | "dark";
  searchTerm?: string;
}

export const DiffLineRow = React.memo<DiffLineProps>(
  ({ line, showLineNumbers, isSplit = false, enableSyntaxHighlighting, language, theme = "dark", searchTerm }) => {
    const styles = CHANGE_TYPE_STYLES[line.type];

    // Render content with syntax highlighting or search highlights
    const renderContent = () => {
      if (line.wordDiffs) {
        return renderWordDiffs(line.wordDiffs);
      }

      if (enableSyntaxHighlighting && language) {
        return createSyntaxHighlightedContent(line.content, language, theme);
      }

      return renderContentWithHighlights(line.content, line.searchHighlights);
    };

    return (
      <div
        className={cn(
          "flex font-mono text-sm leading-relaxed",
          "border-l-2",
          styles.bg,
          styles.border,
          line.collapsed && "hidden"
        )}
        role="row"
        aria-label={`Line ${line.oldLineNumber || line.newLineNumber || "unknown"}: ${line.type} ${line.content.slice(0, 50)}...`}
      >
        {/* Line numbers */}
        {showLineNumbers && (
          <div className="flex shrink-0 select-none" role="cell">
            <span
              className={cn(
                "w-8 sm:w-12 text-right pr-2 sm:pr-3 text-text-muted text-xs sm:text-sm",
                line.oldLineNumber ? "opacity-100" : "opacity-30"
              )}
              aria-label={`Old line number ${line.oldLineNumber}`}
            >
              {line.oldLineNumber ?? ""}
            </span>
            {!isSplit && (
              <span
                className={cn(
                  "w-8 sm:w-12 text-right pr-2 sm:pr-3 text-text-muted text-xs sm:text-sm",
                  line.newLineNumber ? "opacity-100" : "opacity-30"
                )}
                aria-label={`New line number ${line.newLineNumber}`}
              >
                {line.newLineNumber ?? ""}
              </span>
            )}
          </div>
        )}

        {/* Content */}
        <div className="flex-1 flex items-center min-w-0 overflow-hidden" role="cell">
          <span
            className={cn(
              "shrink-0 w-4 text-center select-none",
              styles.text
            )}
            aria-label={`Change type: ${line.type}`}
          >
            {styles.prefix}
          </span>
          <div
            className={cn(
              "flex-1 whitespace-pre overflow-x-auto scrollbar-none",
              styles.text
            )}
          >
            {renderContent() || " "}
          </div>
        </div>
      </div>
    );
  }
);

DiffLineRow.displayName = "DiffLineRow";
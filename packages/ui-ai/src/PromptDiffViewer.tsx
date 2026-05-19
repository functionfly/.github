/**
 * @functionfly/ui-ai
 * Prompt Diff Viewer - Compare prompt versions
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Diff, Plus, Minus, ArrowLeft, ArrowRight, RotateCcw } from "lucide-react";

export interface DiffLine {
  type: "added" | "removed" | "unchanged";
  content: string;
  lineNumber?: number;
}

export interface PromptDiffViewerProps {
  original: string;
  modified: string;
  title?: string;
  onRestore?: () => void;
  onApply?: () => void;
  className?: string;
}

export function PromptDiffViewer({
  original,
  modified,
  title = "Prompt Comparison",
  onRestore,
  onApply,
  className,
}: PromptDiffViewerProps) {
  const [viewMode, setViewMode] = React.useState<"split" | "unified">("split");

  const diffLines = React.useMemo(() => {
    const originalLines = original.split("\n");
    const modifiedLines = modified.split("\n");
    const result: DiffLine[] = [];

    // Simple line-by-line comparison
    const maxLen = Math.max(originalLines.length, modifiedLines.length);
    for (let i = 0; i < maxLen; i++) {
      const origLine = originalLines[i];
      const modLine = modifiedLines[i];

      if (origLine === modLine) {
        result.push({ type: "unchanged", content: origLine ?? "", lineNumber: i + 1 });
      } else {
        if (origLine !== undefined) {
          result.push({ type: "removed", content: origLine, lineNumber: i + 1 });
        }
        if (modLine !== undefined) {
          result.push({ type: "added", content: modLine, lineNumber: i + 1 });
        }
      }
    }

    return result;
  }, [original, modified]);

  const stats = React.useMemo(() => {
    const added = diffLines.filter(l => l.type === "added").length;
    const removed = diffLines.filter(l => l.type === "removed").length;
    return { added, removed };
  }, [diffLines]);

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Diff className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">{title}</span>
          <div className="flex items-center gap-1 ml-2">
            <Badge variant="success" size="sm">+{stats.added}</Badge>
            <Badge variant="error" size="sm">-{stats.removed}</Badge>
          </div>
        </div>

        {/* View Mode Toggle */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode("split")}
            className={cn(
              "px-2 py-1 text-xs rounded transition-colors",
              viewMode === "split" ? "bg-brand-500/10 text-brand-500" : "text-text-muted hover:bg-bg-tertiary"
            )}
          >
            Split
          </button>
          <button
            onClick={() => setViewMode("unified")}
            className={cn(
              "px-2 py-1 text-xs rounded transition-colors",
              viewMode === "unified" ? "bg-brand-500/10 text-brand-500" : "text-text-muted hover:bg-bg-tertiary"
            )}
          >
            Unified
          </button>
        </div>
      </div>

      {/* Diff Content */}
      <div className="flex-1 overflow-auto">
        {viewMode === "split" ? (
          <div className="grid grid-cols-2 h-full">
            {/* Original */}
            <div className="border-r border-border-subtle">
              <div className="px-3 py-2 bg-bg-tertiary/50 border-b border-border-subtle text-[10px] font-medium text-text-muted uppercase tracking-wide">
                Original
              </div>
              <div className="font-mono text-xs">
                {diffLines.filter(l => l.type !== "added").map((line, i) => (
                  <div
                    key={i}
                    className={cn(
                      "flex px-3 py-0.5",
                      line.type === "removed" ? "bg-error/5 text-error" : "text-text-secondary"
                    )}
                  >
                    <span className="w-8 text-text-muted shrink-0 select-none">{line.lineNumber}</span>
                    <span className="flex-1">{line.content}</span>
                    {line.type === "removed" && <Minus className="size-3 text-error shrink-0 mt-0.5" />}
                  </div>
                ))}
              </div>
            </div>

            {/* Modified */}
            <div>
              <div className="px-3 py-2 bg-bg-tertiary/50 border-b border-border-subtle text-[10px] font-medium text-text-muted uppercase tracking-wide">
                Modified
              </div>
              <div className="font-mono text-xs">
                {diffLines.filter(l => l.type !== "removed").map((line, i) => (
                  <div
                    key={i}
                    className={cn(
                      "flex px-3 py-0.5",
                      line.type === "added" ? "bg-success/5 text-success" : "text-text-secondary"
                    )}
                  >
                    <span className="w-8 text-text-muted shrink-0 select-none">{line.lineNumber}</span>
                    <span className="flex-1">{line.content}</span>
                    {line.type === "added" && <Plus className="size-3 text-success shrink-0 mt-0.5" />}
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : (
          /* Unified View */
          <div className="font-mono text-xs">
            {diffLines.map((line, i) => (
              <div
                key={i}
                className={cn(
                  "flex px-3 py-0.5",
                  line.type === "added" ? "bg-success/5" : line.type === "removed" ? "bg-error/5" : ""
                )}
              >
                <span className="w-8 text-text-muted shrink-0 select-none">{line.lineNumber}</span>
                <span className="w-6 text-center shrink-0">
                  {line.type === "added" && <Plus className="size-3 text-success" />}
                  {line.type === "removed" && <Minus className="size-3 text-error" />}
                </span>
                <span className={cn(
                  "flex-1",
                  line.type === "added" ? "text-success" : line.type === "removed" ? "text-error" : "text-text-secondary"
                )}>
                  {line.content}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2 px-4 py-3 border-t border-border-subtle">
        <button
          onClick={onRestore}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors"
        >
          <RotateCcw className="size-3" />
          Restore Original
        </button>
        <button
          onClick={onApply}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors ml-auto"
        >
          Apply Changes
          <ArrowRight className="size-3" />
        </button>
      </div>
    </div>
  );
}

/**
 * Production-ready diff viewer component for replay/version insights
 */

import React, { useMemo, useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Copy,
  Check,
  Split,
  AlignLeft,
  FileCode,
  GitCompare,
  Search,
  X,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";

import type { FlyDiffViewerProps, DiffError, DiffStats } from "./types";
import { DEFAULT_PROPS } from "./constants";
import { computeDiff } from "./utils";
import { useCopyToClipboard, useDebouncedSearch, useCollapse, useSearch } from "./hooks";
import { SplitView, UnifiedView, LoadingState } from "./components";
import { formatFileSize } from "./utils/formatters";

/**
 * Main FlyDiffViewer component
 */
export const FlyDiffViewer = React.memo<FlyDiffViewerProps>(
  (props) => {
    // Merge props with defaults
    const {
      before,
      after,
      beforeLabel = DEFAULT_PROPS.beforeLabel,
      afterLabel = DEFAULT_PROPS.afterLabel,
      language = DEFAULT_PROPS.language,
      viewMode: initialViewMode = DEFAULT_PROPS.viewMode,
      showLineNumbers = DEFAULT_PROPS.showLineNumbers,
      fileName,
      className,
      showViewToggle = DEFAULT_PROPS.showViewToggle,
      maxHeight = DEFAULT_PROPS.maxHeight,
      enableWordDiffs = DEFAULT_PROPS.enableWordDiffs,
      enableSyntaxHighlighting = DEFAULT_PROPS.enableSyntaxHighlighting,
      theme = DEFAULT_PROPS.theme,
      enableCollapse = DEFAULT_PROPS.enableCollapse,
      maxLinesBeforeCollapse = DEFAULT_PROPS.maxLinesBeforeCollapse,
      enableSearch = DEFAULT_PROPS.enableSearch,
      diffOptions = {},
      onDiffComputed,
      onError,
      loadingComponent,
      errorComponent,
    } = props;

    // State management
    const [viewMode, setViewMode] = useState(initialViewMode);
    const [searchTerm, setSearchTerm] = useDebouncedSearch("", 300);
    const [copiedBefore, copyBefore] = useCopyToClipboard();
    const [copiedAfter, copyAfter] = useCopyToClipboard();
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<DiffError | null>(null);

    // Compute diff with error handling
    const diffResult = useMemo(() => {
      setIsLoading(true);
      setError(null);

      try {
        const result = computeDiff(before, after, {
          enableWordDiffs,
          ...diffOptions,
        });

        if (result.error) {
          setError(result.error);
          onError?.(result.error);
        } else {
          onDiffComputed?.(result.stats);
        }

        setIsLoading(false);
        return result;
      } catch (err) {
        const diffError: DiffError = {
          message: err instanceof Error ? err.message : "Unexpected error during diff computation",
          code: "DIFF_COMPUTATION_ERROR",
          recoverable: false,
        };
        setError(diffError);
        onError?.(diffError);
        setIsLoading(false);
        return { hunks: [], stats: { added: 0, removed: 0, modified: 0, total: 0, totalLines: 0 } };
      }
    }, [before, after, enableWordDiffs, diffOptions, onDiffComputed, onError]);

    const { hunks, stats } = diffResult;

    // Apply search to lines
    const searchedHunks = useMemo(() => {
      if (!searchTerm.trim()) return hunks;

      return hunks.map((hunk) => ({
        ...hunk,
        lines: hunk.lines.map((line) => ({
          ...line,
          searchHighlights: line.content.includes(searchTerm)
            ? [{ start: line.content.indexOf(searchTerm), end: line.content.indexOf(searchTerm) + searchTerm.length }]
            : undefined,
        })),
      }));
    }, [hunks, searchTerm]);

    // Collapse functionality
    const { collapsedHunks, toggleHunkCollapse, expandAll, collapseAll } = useCollapse(
      searchedHunks,
      enableCollapse,
      maxLinesBeforeCollapse
    );

    // File size information
    const fileSize = useMemo(() => ({
      before: formatFileSize(new Blob([before]).size),
      after: formatFileSize(new Blob([after]).size),
    }), [before, after]);

    // Loading state
    if (isLoading) {
      return loadingComponent || (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className={cn(
            "rounded-xl overflow-hidden",
            "bg-bg-primary",
            "border border-(--color-border)",
            "font-mono text-sm",
            className
          )}
        >
          <LoadingState />
        </motion.div>
      );
    }

    // Error state
    if (error) {
      return errorComponent || (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className={cn(
            "rounded-xl overflow-hidden",
            "bg-bg-primary",
            "border border-red-500/20",
            "font-mono text-sm",
            className
          )}
        >
          <div className="flex items-center justify-center p-8 text-red-400">
            <AlertTriangle className="h-6 w-6 mr-2" />
            <div>
              <p className="font-medium">Diff computation failed</p>
              <p className="text-sm text-red-300 mt-1">{error.message}</p>
              {error.recoverable && (
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3"
                  onClick={() => window.location.reload()}
                >
                  Retry
                </Button>
              )}
            </div>
          </div>
        </motion.div>
      );
    }

    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className={cn(
          "rounded-xl overflow-hidden",
          "bg-bg-primary",
          "border border-(--color-border)",
          "font-mono text-sm",
          className
        )}
        role="region"
        aria-label="Diff viewer"
      >
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between px-4 py-2 bg-(--color-bg-secondary) border-b border-(--color-border) gap-2">
          <div className="flex items-center gap-3 min-w-0">
            <FileCode className="h-4 w-4 text-(--color-brand-500) flex-shrink-0" />
            <div className="flex flex-col min-w-0">
              <span className="font-medium text-(--color-text-primary) truncate">
                {fileName || "Changes"}
              </span>
              <div className="flex items-center gap-2 text-xs text-text-muted">
                <span className="hidden sm:inline">{fileSize.before}</span>
                <span className="hidden sm:inline">→</span>
                <span>{fileSize.after}</span>
              </div>
            </div>
            <div className="flex items-center gap-2 flex-shrink-0">
              {stats.added > 0 && (
                <Badge
                  variant="success"
                  className="text-xs bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                >
                  +{stats.added}
                </Badge>
              )}
              {stats.removed > 0 && (
                <Badge
                  variant="error"
                  className="text-xs bg-red-500/10 text-red-400 border-red-500/20"
                >
                  -{stats.removed}
                </Badge>
              )}
              {stats.modified > 0 && (
                <Badge
                  variant="warning"
                  className="text-xs bg-amber-500/10 text-amber-400 border-amber-500/20"
                >
                  ~{stats.modified}
                </Badge>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2 flex-wrap sm:flex-nowrap">
            {/* Search */}
            {enableSearch && (
              <div className="relative order-1 sm:order-none">
                <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-text-muted" />
                <Input
                  type="text"
                  placeholder="Search..."
                  className="pl-7 h-7 w-24 sm:w-32 text-xs"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  aria-label="Search diff content"
                />
                {searchTerm && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute right-1 top-1/2 transform -translate-y-1/2 h-5 w-5"
                    onClick={() => setSearchTerm("")}
                    aria-label="Clear search"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                )}
              </div>
            )}

            {/* Mobile controls group */}
            <div className="flex items-center gap-1 order-2 sm:order-none">
              {/* Collapse controls */}
              {enableCollapse && searchedHunks.length > 0 && (
                <>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={expandAll}
                    title="Expand all hunks"
                    aria-label="Expand all collapsed hunks"
                  >
                    <ChevronDown className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={collapseAll}
                    title="Collapse all hunks"
                    aria-label="Collapse all hunks"
                  >
                    <ChevronUp className="h-3.5 w-3.5" />
                  </Button>
                </>
              )}

              {/* Copy buttons */}
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => copyBefore(before)}
                title={`Copy ${beforeLabel}`}
                aria-label={`Copy ${beforeLabel} content`}
              >
                {copiedBefore ? (
                  <Check className="h-3.5 w-3.5 text-emerald-500" />
                ) : (
                  <Copy className="h-3.5 w-3.5" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => copyAfter(after)}
                title={`Copy ${afterLabel}`}
                aria-label={`Copy ${afterLabel} content`}
              >
                {copiedAfter ? (
                  <Check className="h-3.5 w-3.5 text-emerald-500" />
                ) : (
                  <Copy className="h-3.5 w-3.5" />
                )}
              </Button>
            </div>

            {/* View toggle */}
            {showViewToggle && (
              <div className="flex items-center gap-1 border-l border-(--color-border) pl-2 order-3 sm:order-none">
                <Button
                  variant={viewMode === "split" ? "secondary" : "ghost"}
                  size="icon"
                  className="h-7 w-7"
                  onClick={() => setViewMode("split")}
                  title="Split view"
                  aria-label="Switch to split view"
                  aria-pressed={viewMode === "split"}
                >
                  <Split className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant={viewMode === "unified" ? "secondary" : "ghost"}
                  size="icon"
                  className="h-7 w-7"
                  onClick={() => setViewMode("unified")}
                  title="Unified view"
                  aria-label="Switch to unified view"
                  aria-pressed={viewMode === "unified"}
                >
                  <AlignLeft className="h-3.5 w-3.5" />
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Column headers for split view */}
        {viewMode === "split" && (
          <div className="flex border-b border-(--color-border)">
            <div className="flex-1 px-2 sm:px-4 py-1.5 text-xs text-text-secondary border-r border-(--color-border)">
              <span className="hidden sm:inline">{beforeLabel}</span>
              <span className="sm:hidden">Before</span>
            </div>
            <div className="flex-1 px-2 sm:px-4 py-1.5 text-xs text-text-secondary">
              <span className="hidden sm:inline">{afterLabel}</span>
              <span className="sm:hidden">After</span>
            </div>
          </div>
        )}

        {/* Diff content */}
        <div className="relative">
          {searchedHunks.length === 0 ? (
            <div className="p-4 sm:p-8 text-center text-text-secondary">
              <GitCompare className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No changes detected</p>
              {searchTerm && (
                <p className="text-xs mt-1">Try adjusting your search terms</p>
              )}
            </div>
          ) : viewMode === "split" ? (
            <SplitView
              hunks={searchedHunks}
              showLineNumbers={showLineNumbers}
              enableSyntaxHighlighting={enableSyntaxHighlighting}
              language={language}
              theme={theme}
              searchTerm={searchTerm}
              collapsedHunks={collapsedHunks}
              onToggleCollapse={toggleHunkCollapse}
              maxHeight={maxHeight}
            />
          ) : (
            <UnifiedView
              hunks={searchedHunks}
              showLineNumbers={showLineNumbers}
              enableSyntaxHighlighting={enableSyntaxHighlighting}
              language={language}
              theme={theme}
              searchTerm={searchTerm}
              collapsedHunks={collapsedHunks}
              onToggleCollapse={toggleHunkCollapse}
              maxHeight={maxHeight}
            />
          )}
        </div>
      </motion.div>
    );
  }
);

FlyDiffViewer.displayName = "FlyDiffViewer";
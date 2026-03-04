/**
 * Types and interfaces for the FlyDiffViewer component
 */

export type ChangeType = "added" | "removed" | "modified" | "unchanged";

/**
 * Word-level diff segment within a line
 */
export interface WordDiff {
  /** The text content */
  text: string;
  /** Type of change for this segment */
  type: ChangeType;
  /** Whether this segment is highlighted */
  highlighted?: boolean;
}

/**
 * Single line in a diff with word-level highlighting
 */
export interface DiffLine {
  /** Line number in the original file */
  oldLineNumber?: number;
  /** Line number in the new file */
  newLineNumber?: number;
  /** Line content */
  content: string;
  /** Type of change */
  type: ChangeType;
  /** Whether this line is part of a change block */
  isInChangeBlock?: boolean;
  /** Word-level diffs for modified lines */
  wordDiffs?: WordDiff[];
  /** Whether this line is collapsed/hidden */
  collapsed?: boolean;
  /** Search highlights */
  searchHighlights?: { start: number; end: number }[];
}

/**
 * A hunk of changes in the diff
 */
export interface DiffHunk {
  /** Old file range info */
  oldRange: { start: number; count: number };
  /** New file range info */
  newRange: { start: number; count: number };
  /** Lines in this hunk */
  lines: DiffLine[];
  /** Whether this hunk is collapsed */
  collapsed?: boolean;
  /** Hunk header context */
  context?: string;
}

/**
 * View mode for diff display
 */
export type DiffViewMode = "split" | "unified";

/**
 * Theme mode for syntax highlighting
 */
export type ThemeMode = "light" | "dark";

/**
 * Diff statistics
 */
export interface DiffStats {
  added: number;
  removed: number;
  modified: number;
  total: number;
  totalLines: number;
}

/**
 * Error information for diff processing
 */
export interface DiffError {
  message: string;
  code: string;
  recoverable: boolean;
}

/**
 * Props for the FlyDiffViewer component
 */
export interface FlyDiffViewerProps {
  /** Original content */
  before: string;
  /** Modified content */
  after: string;
  /** Label for the before version */
  beforeLabel?: string;
  /** Label for the after version */
  afterLabel?: string;
  /** Programming language for syntax highlighting */
  language?: string;
  /** Display mode */
  viewMode?: DiffViewMode;
  /** Show line numbers */
  showLineNumbers?: boolean;
  /** File or function name */
  fileName?: string;
  /** Custom className */
  className?: string;
  /** Whether to show the view mode toggle */
  showViewToggle?: boolean;
  /** Maximum height before scrolling */
  maxHeight?: number;
  /** Enable word-level diffs */
  enableWordDiffs?: boolean;
  /** Enable syntax highlighting */
  enableSyntaxHighlighting?: boolean;
  /** Theme mode for syntax highlighting */
  theme?: ThemeMode;
  /** Enable collapse/expand functionality */
  enableCollapse?: boolean;
  /** Maximum lines to show before collapsing */
  maxLinesBeforeCollapse?: number;
  /** Enable search functionality */
  enableSearch?: boolean;
  /** Custom diff algorithm options */
  diffOptions?: {
    timeout?: number;
    editCost?: number;
  };
  /** Callback when diff computation completes */
  onDiffComputed?: (stats: DiffStats) => void;
  /** Callback when an error occurs */
  onError?: (error: DiffError) => void;
  /** Loading component */
  loadingComponent?: React.ReactNode;
  /** Error component */
  errorComponent?: React.ReactNode;
}

/**
 * Props for the FlyDiffViewerWithBoundary component
 */
export interface FlyDiffViewerWithBoundaryProps extends FlyDiffViewerProps {
  enableErrorBoundary?: boolean;
}
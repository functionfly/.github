/**
 * Constants and configuration for the FlyDiffViewer component
 */

import type { ChangeType } from "./types";

/**
 * Change type styles for diff visualization
 */
export const CHANGE_TYPE_STYLES: Record<ChangeType, { bg: string; text: string; border: string; prefix: string }> = {
  added: {
    bg: "bg-emerald-500/10",
    text: "text-emerald-400",
    border: "border-l-emerald-500",
    prefix: "+",
  },
  removed: {
    bg: "bg-red-500/10",
    text: "text-red-400",
    border: "border-l-red-500",
    prefix: "-",
  },
  modified: {
    bg: "bg-amber-500/10",
    text: "text-amber-400",
    border: "border-l-amber-500",
    prefix: "~",
  },
  unchanged: {
    bg: "transparent",
    text: "text-text-secondary",
    border: "border-l-transparent",
    prefix: " ",
  },
};

/**
 * Supported programming languages for syntax highlighting
 */
export const SUPPORTED_LANGUAGES = new Set([
  "javascript", "typescript", "jsx", "tsx", "python", "java", "cpp", "c", "csharp",
  "php", "ruby", "go", "rust", "swift", "kotlin", "scala", "html", "css", "scss",
  "sass", "less", "json", "xml", "yaml", "yml", "markdown", "md", "sql", "bash",
  "shell", "dockerfile", "makefile", "toml", "ini", "diff", "log", "text"
]);

/**
 * Default props for the FlyDiffViewer component
 */
export const DEFAULT_PROPS = {
  beforeLabel: "Before",
  afterLabel: "After",
  language: "text",
  viewMode: "unified" as const,
  showLineNumbers: true,
  showViewToggle: true,
  maxHeight: 400,
  enableWordDiffs: false,
  enableSyntaxHighlighting: true,
  theme: "dark" as const,
  enableCollapse: true,
  maxLinesBeforeCollapse: 50,
  enableSearch: true,
  enableErrorBoundary: true,
} as const;
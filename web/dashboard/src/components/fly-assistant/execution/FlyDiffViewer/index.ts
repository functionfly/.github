/**
 * FlyDiffViewer - Production-ready diff viewer component
 */

// Main component
export { FlyDiffViewer } from "./FlyDiffViewer";

// Error boundary wrapper
export { DiffErrorBoundary } from "./ErrorBoundary";

// Wrapped component with error boundary
export { FlyDiffViewerWithBoundary } from "./FlyDiffViewerWithBoundary";

// Types and interfaces
export type {
  ChangeType,
  WordDiff,
  DiffLine,
  DiffHunk,
  DiffViewMode,
  ThemeMode,
  DiffStats,
  DiffError,
  FlyDiffViewerProps,
  FlyDiffViewerWithBoundaryProps,
} from "./types";

// Utilities
export * from "./utils";

// Hooks
export * from "./hooks";

// Components (for advanced usage)
export * from "./components";

// Constants
export { CHANGE_TYPE_STYLES, SUPPORTED_LANGUAGES, DEFAULT_PROPS } from "./constants";
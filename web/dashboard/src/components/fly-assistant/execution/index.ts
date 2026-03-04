/**
 * Execution Components (Phase 4 - Action & Execution)
 *
 * These components connect the chat to real actions and provide safety/visibility
 * for AI-suggested executions. They bridge the gap between AI recommendations
 * and actual system changes.
 *
 * @module fly-assistant/execution
 */

// ============================================================================
// FlyQuickActions - Dynamic AI-generated action buttons
// ============================================================================

export {
  FlyQuickActions,
} from "./FlyQuickActions";

export type {
  FlyQuickActionsProps,
  QuickAction,
  ActionVariant,
  ActionState,
  ActionIconName,
} from "./FlyQuickActions";

// ============================================================================
// FlyExecutionPreview - Safety preview before execution
// ============================================================================

export {
  FlyExecutionPreview,
} from "./FlyExecutionPreview";

export type {
  FlyExecutionPreviewProps,
  RiskLevel,
  AffectedFunction,
} from "./FlyExecutionPreview";

// ============================================================================
// FlyDiffViewer - Before/after comparison for code changes
// ============================================================================

export {
  FlyDiffViewer,
  FlyDiffViewerWithBoundary,
} from "./FlyDiffViewer";

export type {
  FlyDiffViewerProps,
  FlyDiffViewerWithBoundaryProps,
  DiffViewMode,
  ChangeType,
  DiffLine,
  DiffHunk,
  ThemeMode,
  DiffStats,
  DiffError,
} from "./FlyDiffViewer";

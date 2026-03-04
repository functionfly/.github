/**
 * FlyAssistant UI Components (Phase 5 - UX Polish)
 *
 * These components provide the polish that separates good from elite UX.
 * They enhance the user experience with micro-interactions, notifications,
 * and advanced power-user features.
 *
 * @module fly-assistant/ui
 */

// ============================================================================
// Notification Components
// ============================================================================

export {
  FlyNotificationDot,
} from "./FlyNotificationDot";

export type {
  FlyNotificationDotProps,
} from "./FlyNotificationDot";

// ============================================================================
// Micro-Suggestion Components
// ============================================================================

export {
  FlyAutoNudge,
} from "./FlyAutoNudge";

export type {
  FlyAutoNudgeProps,
} from "./FlyAutoNudge";

// ============================================================================
// Memory & History Components
// ============================================================================

export {
  FlyMemoryTimeline,
} from "./FlyMemoryTimeline";

export type {
  FlyMemoryTimelineProps,
  ConversationSummary,
} from "./FlyMemoryTimeline";

// ============================================================================
// Command Palette Components
// ============================================================================

export {
  FlyCommandPalette,
} from "./FlyCommandPalette";

export type {
  FlyCommandPaletteProps,
} from "./FlyCommandPalette";

// ============================================================================
// Mode Switcher Components
// ============================================================================

export {
  FlyModeSwitcher,
} from "./FlyModeSwitcher";

export type {
  FlyModeSwitcherProps,
  AssistantMode,
} from "./FlyModeSwitcher";

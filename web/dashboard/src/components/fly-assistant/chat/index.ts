/**
 * FlyAssistant Chat Components
 *
 * Internal message engine components for the FlyAssistant chat system.
 *
 * @module fly-assistant/chat
 */

// ============================================================================
// Chat Window (Virtualized Message List)
// ============================================================================

export { FlyChatWindow } from "./FlyChatWindow";
export type {
  FlyChatWindowProps,
} from "./FlyChatWindow";

// ============================================================================
// Message Component (Variants for different message types)
// ============================================================================

export { FlyMessage } from "./FlyMessage";
export type {
  FlyMessageProps,
  MessageVariant,
  MessageAction,
  MarketplaceSuggestion,
} from "./FlyMessage";

// ============================================================================
// Typing Indicator (Animated AI thinking state)
// ============================================================================

export {
  FlyTypingIndicator,
  FlyTypingIndicatorCompact,
  FlyTypingSkeleton,
} from "./FlyTypingIndicator";
export type {
  FlyTypingIndicatorProps,
  FlyTypingIndicatorCompactProps,
  FlyTypingSkeletonProps,
} from "./FlyTypingIndicator";

// ============================================================================
// Input Bar (Slash commands, attachments, voice)
// ============================================================================

export { FlyInputBar } from "./FlyInputBar";
export type {
  FlyInputBarProps,
  SlashCommand,
  Attachment,
} from "./FlyInputBar";

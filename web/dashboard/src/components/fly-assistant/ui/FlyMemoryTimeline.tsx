/**
 * FlyMemoryTimeline.tsx
 *
 * Shows past conversations related to the current function.
 * Provides a visual timeline with conversation nodes, grouped by date.
 *
 * @module fly-assistant/ui
 */

import React, { useMemo, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Clock, MessageSquare, Trash2, History } from "lucide-react";
import { cn } from "@/lib/utils";
import { formatDistanceToNow, isToday, isYesterday, format } from "date-fns";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface ConversationSummary {
  /** Unique conversation ID */
  id: string;
  /** Preview of last message (max 40 chars recommended) */
  preview: string;
  /** Timestamp of conversation */
  timestamp: Date;
  /** Number of messages in conversation */
  messageCount: number;
}

export interface FlyMemoryTimelineProps {
  /** Array of conversation summaries */
  conversations: ConversationSummary[];
  /** Current function context (optional filter) */
  currentFunction?: string;
  /** Callback when a conversation is selected */
  onSelectConversation: (id: string) => void;
  /** Callback to clear all history */
  onClearHistory?: () => void;
  /** Custom className */
  className?: string;
}

// ============================================================================
// Date Grouping Helper
// ============================================================================

interface GroupedConversations {
  today: ConversationSummary[];
  yesterday: ConversationSummary[];
  lastWeek: ConversationSummary[];
  older: ConversationSummary[];
}

function groupConversationsByDate(
  conversations: ConversationSummary[]
): GroupedConversations {
  const now = new Date();
  const oneWeekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  return conversations.reduce(
    (acc, conv) => {
      const date = new Date(conv.timestamp);

      if (isToday(date)) {
        acc.today.push(conv);
      } else if (isYesterday(date)) {
        acc.yesterday.push(conv);
      } else if (date > oneWeekAgo) {
        acc.lastWeek.push(conv);
      } else {
        acc.older.push(conv);
      }

      return acc;
    },
    {
      today: [],
      yesterday: [],
      lastWeek: [],
      older: [],
    } as GroupedConversations
  );
}

// ============================================================================
// Component
// ============================================================================

export const FlyMemoryTimeline: React.FC<FlyMemoryTimelineProps> = ({
  conversations,
  currentFunction,
  onSelectConversation,
  onClearHistory,
  className,
}) => {
  // Group conversations by date
  const grouped = useMemo(
    () => groupConversationsByDate(conversations),
    [conversations]
  );

  // Handle conversation selection
  const handleSelect = useCallback(
    (id: string) => {
      onSelectConversation(id);
    },
    [onSelectConversation]
  );

  // Empty state
  if (conversations.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className={cn(
          "flex flex-col items-center justify-center py-8 px-4",
          "text-[var(--color-text-muted)]",
          className
        )}
      >
        <History className="w-10 h-10 mb-3 opacity-50" aria-hidden="true" />
        <p className="text-sm text-center">No previous conversations</p>
        {currentFunction && (
          <p className="text-xs mt-1 text-center opacity-70">
            Start chatting with {currentFunction}
          </p>
        )}
      </motion.div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header with clear action */}
      {onClearHistory && (
        <div className="flex items-center justify-between px-1">
          <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
            History
          </span>
          <button
            onClick={onClearHistory}
            className={cn(
              "flex items-center gap-1 text-xs",
              "text-[var(--color-text-muted)] hover:text-red-400",
              "transition-colors focus:outline-none",
              "focus:ring-2 focus:ring-[var(--color-border-focus)] rounded px-1.5 py-0.5"
            )}
            aria-label="Clear conversation history"
          >
            <Trash2 className="w-3 h-3" aria-hidden="true" />
            Clear
          </button>
        </div>
      )}

      {/* Timeline */}
      <div className="relative">
        {/* Timeline line */}
        <div
          className="absolute left-[4px] top-2 bottom-2 w-[2px] bg-[var(--color-border)]"
          aria-hidden="true"
        />

        {/* Conversation groups */}
        <div className="space-y-5">
          {/* Today */}
          {grouped.today.length > 0 && (
            <TimelineGroup
              title="Today"
              conversations={grouped.today}
              onSelect={handleSelect}
            />
          )}

          {/* Yesterday */}
          {grouped.yesterday.length > 0 && (
            <TimelineGroup
              title="Yesterday"
              conversations={grouped.yesterday}
              onSelect={handleSelect}
            />
          )}

          {/* Last Week */}
          {grouped.lastWeek.length > 0 && (
            <TimelineGroup
              title="Last Week"
              conversations={grouped.lastWeek}
              onSelect={handleSelect}
            />
          )}

          {/* Older */}
          {grouped.older.length > 0 && (
            <TimelineGroup
              title="Older"
              conversations={grouped.older}
              onSelect={handleSelect}
              showFullDate
            />
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Timeline Group Component
// ============================================================================

interface TimelineGroupProps {
  title: string;
  conversations: ConversationSummary[];
  onSelect: (id: string) => void;
  showFullDate?: boolean;
}

const TimelineGroup: React.FC<TimelineGroupProps> = ({
  title,
  conversations,
  onSelect,
  showFullDate = false,
}) => {
  return (
    <div>
      {/* Group header */}
      <div className="flex items-center gap-3 mb-2">
        <div className="w-[10px] h-[10px] rounded-full bg-[var(--color-brand-500)] ring-2 ring-[var(--color-bg-primary)] z-10" />
        <span className="text-xs font-medium text-[var(--color-text-tertiary)]">
          {title}
        </span>
      </div>

      {/* Conversations in group */}
      <div className="space-y-1 pl-5">
        {conversations.map((conv, index) => (
          <motion.button
            key={conv.id}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.05 }}
            onClick={() => onSelect(conv.id)}
            className={cn(
              "w-full text-left p-2.5 rounded-lg",
              "border border-transparent",
              "bg-[var(--color-bg-secondary)]/50",
              "hover:bg-[var(--color-bg-secondary)]",
              "hover:border-[var(--color-brand-500)]/30",
              "transition-all duration-200",
              "focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-500)]/30"
            )}
            aria-label={`Conversation from ${format(
              new Date(conv.timestamp),
              "PPp"
            )}: ${conv.preview}`}
          >
            {/* Preview text */}
            <p className="text-sm text-[var(--color-text-primary)] truncate pr-2">
              {conv.preview.length > 40
                ? conv.preview.slice(0, 40) + "..."
                : conv.preview}
            </p>

            {/* Meta info */}
            <div className="flex items-center gap-3 mt-1.5">
              {/* Timestamp */}
              <span className="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
                <Clock className="w-3 h-3" aria-hidden="true" />
                {showFullDate
                  ? format(new Date(conv.timestamp), "MMM d")
                  : formatDistanceToNow(new Date(conv.timestamp), {
                      addSuffix: true,
                    })}
              </span>

              {/* Message count */}
              {conv.messageCount > 0 && (
                <span className="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
                  <MessageSquare className="w-3 h-3" aria-hidden="true" />
                  {conv.messageCount}
                </span>
              )}
            </div>
          </motion.button>
        ))}
      </div>
    </div>
  );
};

export default FlyMemoryTimeline;

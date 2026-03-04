/**
 * FlyChatWindow.tsx
 *
 * Virtualized scroll container for chat messages with auto-scroll logic
 * and replay state support for viewing historical conversations.
 */

import React, { useRef, useEffect, useCallback, useState } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { motion, AnimatePresence } from "framer-motion";
import { MessageSquare, History } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Message } from "../FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyChatWindowProps {
  /** Array of messages to display */
  messages: Message[];
  /** Whether messages are currently loading */
  isLoading?: boolean;
  /** Whether in replay mode (viewing historical conversation) */
  isReplayMode?: boolean;
  /** Callback when scroll reaches bottom */
  onScrollToBottom?: () => void;
  /** Custom empty state component */
  emptyState?: React.ReactNode;
  /** Custom className */
  className?: string;
  /** Render function for message items */
  renderMessage: (message: Message, index: number) => React.ReactNode;
  /** Render function for typing indicator */
  renderTypingIndicator?: () => React.ReactNode;
}

// ============================================================================
// Constants
// ============================================================================

const ESTIMATED_ITEM_HEIGHT = 80;
const OVERSCAN_COUNT = 5;
const SCROLL_THRESHOLD = 100; // Pixels from bottom to trigger auto-scroll

// ============================================================================
// Component
// ============================================================================

export const FlyChatWindow = React.forwardRef<HTMLDivElement, FlyChatWindowProps>(
  (
    {
      messages,
      isLoading = false,
      isReplayMode = false,
      onScrollToBottom,
      emptyState,
      className,
      renderMessage,
      renderTypingIndicator,
    },
    forwardedRef
  ) => {
    const parentRef = useRef<HTMLDivElement>(null);
    const [isUserScrolling, setIsUserScrolling] = useState(false);
    const [showScrollButton, setShowScrollButton] = useState(false);
    const scrollTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const lastMessageCountRef = useRef(messages.length);

    // Merge refs
    const setRefs = useCallback(
      (element: HTMLDivElement | null) => {
        (parentRef as React.MutableRefObject<HTMLDivElement | null>).current = element;
        if (typeof forwardedRef === "function") {
          forwardedRef(element);
        } else if (forwardedRef) {
          (forwardedRef as React.MutableRefObject<HTMLDivElement | null>).current = element;
        }
      },
      [forwardedRef]
    );

    // Virtualizer setup
    const virtualizer = useVirtualizer({
      count: messages.length,
      getScrollElement: () => parentRef.current,
      estimateSize: () => ESTIMATED_ITEM_HEIGHT,
      overscan: OVERSCAN_COUNT,
      measureElement: (element) => element.getBoundingClientRect().height,
    });

    const virtualItems = virtualizer.getVirtualItems();

    // Check if user is near bottom
    const isNearBottom = useCallback(() => {
      const container = parentRef.current;
      if (!container) return true;

      const { scrollTop, scrollHeight, clientHeight } = container;
      return scrollHeight - scrollTop - clientHeight < SCROLL_THRESHOLD;
    }, []);

    // Scroll to bottom
    const scrollToBottom = useCallback(
      (behavior: "auto" | "smooth" = "smooth") => {
        if (virtualizer) {
          virtualizer.scrollToIndex(messages.length - 1, {
            align: "end",
            behavior,
          });
        }
        onScrollToBottom?.();
      },
      [virtualizer, messages.length, onScrollToBottom]
    );

    // Handle scroll events
    const handleScroll = useCallback(() => {
      if (!isUserScrolling) {
        setIsUserScrolling(true);
      }

      // Clear existing timeout
      if (scrollTimeoutRef.current) {
        clearTimeout(scrollTimeoutRef.current);
      }

      // Show/hide scroll button based on position
      setShowScrollButton(!isNearBottom());

      // Reset user scrolling state after delay
      scrollTimeoutRef.current = setTimeout(() => {
        setIsUserScrolling(false);
      }, 150);
    }, [isUserScrolling, isNearBottom]);

    // Auto-scroll on new messages (if user hasn't scrolled up)
    useEffect(() => {
      if (messages.length > lastMessageCountRef.current) {
        const shouldAutoScroll = !isUserScrolling || isNearBottom();
        if (shouldAutoScroll) {
          // Use requestAnimationFrame for smooth scroll after render
          requestAnimationFrame(() => {
            scrollToBottom("smooth");
          });
        }
      }
      lastMessageCountRef.current = messages.length;
    }, [messages.length, isUserScrolling, isNearBottom, scrollToBottom]);

    // Scroll to bottom on initial load
    useEffect(() => {
      if (messages.length > 0) {
        scrollToBottom("auto");
      }
    }, []); // eslint-disable-line react-hooks/exhaustive-deps

    // Cleanup timeout on unmount
    useEffect(() => {
      return () => {
        if (scrollTimeoutRef.current) {
          clearTimeout(scrollTimeoutRef.current);
        }
      };
    }, []);

    // Empty state
    if (messages.length === 0 && !isLoading) {
      return (
        <div
          ref={setRefs}
          className={cn(
            "flex flex-col items-center justify-center h-full p-6",
            "text-center",
            className
          )}
        >
          {emptyState || <DefaultEmptyState />}
        </div>
      );
    }

    return (
      <div className={cn("relative flex flex-col h-full", className)}>
        {/* Replay mode indicator */}
        <AnimatePresence>
          {isReplayMode && (
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className={cn(
                "absolute top-0 left-0 right-0 z-10",
                "flex items-center justify-center gap-2",
                "py-2 px-4",
                "bg-amber-500/10 border-b border-amber-500/30",
                "text-amber-400 text-sm font-medium"
              )}
            >
              <History className="w-4 h-4" />
              <span>Viewing conversation history</span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Scrollable container */}
        <div
          ref={setRefs}
          onScroll={handleScroll}
          className={cn(
            "flex-1 overflow-y-auto overflow-x-hidden",
            "scroll-smooth",
            isReplayMode && "pt-10"
          )}
          role="log"
          aria-live="polite"
          aria-label="Chat messages"
        >
          <div
            style={{
              height: `${virtualizer.getTotalSize()}px`,
              width: "100%",
              position: "relative",
            }}
          >
            {virtualItems.map((virtualItem) => {
              const message = messages[virtualItem.index];
              return (
                <div
                  key={message.id}
                  data-index={virtualItem.index}
                  ref={virtualizer.measureElement}
                  style={{
                    position: "absolute",
                    top: 0,
                    left: 0,
                    width: "100%",
                    transform: `translateY(${virtualItem.start}px)`,
                  }}
                >
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{
                      duration: 0.3,
                      delay: Math.min(virtualItem.index * 0.05, 0.3),
                    }}
                  >
                    {renderMessage(message, virtualItem.index)}
                  </motion.div>
                </div>
              );
            })}
          </div>

          {/* Typing indicator */}
          {isLoading && renderTypingIndicator && (
            <div className="px-4 py-2">{renderTypingIndicator()}</div>
          )}
        </div>

        {/* Scroll to bottom button */}
        <AnimatePresence>
          {showScrollButton && (
            <motion.button
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              onClick={() => scrollToBottom("smooth")}
              className={cn(
                "absolute bottom-4 right-4",
                "flex items-center justify-center",
                "w-10 h-10",
                "bg-[var(--color-bg-tertiary)] hover:bg-[var(--color-bg-secondary)]",
                "border border-[var(--color-border)]",
                "rounded-full shadow-lg",
                "transition-colors duration-200",
                "z-20"
              )}
              aria-label="Scroll to bottom"
            >
              <svg
                className="w-5 h-5 text-[var(--color-text-primary)]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 14l-7 7m0 0l-7-7m7 7V3"
                />
              </svg>
            </motion.button>
          )}
        </AnimatePresence>
      </div>
    );
  }
);

FlyChatWindow.displayName = "FlyChatWindow";

// ============================================================================
// Default Empty State
// ============================================================================

function DefaultEmptyState() {
  return (
    <div className="flex flex-col items-center gap-4 text-[var(--color-text-secondary)]">
      <div
        className={cn(
          "flex items-center justify-center",
          "w-16 h-16",
          "rounded-2xl",
          "bg-[var(--color-bg-secondary)]",
          "border border-[var(--color-border)]"
        )}
      >
        <MessageSquare className="w-8 h-8 text-[var(--color-brand-500)]" />
      </div>
      <div className="space-y-1 text-center">
        <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
          Start a conversation
        </h3>
        <p className="text-sm max-w-xs">
          Ask me anything about your functions, deployments, or get help with
          optimization.
        </p>
      </div>
    </div>
  );
}

export default FlyChatWindow;

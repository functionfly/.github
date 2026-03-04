/**
 * FlyMessage.tsx
 *
 * Base message component with variants for user, AI, system insights,
 * marketplace suggestions, and warning alerts.
 */

import React, { useCallback, useState } from "react";
import { motion } from "framer-motion";
import ReactMarkdown from "react-markdown";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import remarkGfm from "remark-gfm";
import {
  User,
  Bot,
  Lightbulb,
  Store,
  AlertTriangle,
  Copy,
  Check,
  ThumbsUp,
  ThumbsDown,
  X,
  CheckCircle,
  Info,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { format } from "date-fns";

// ============================================================================
// Types & Interfaces
// ============================================================================

export type MessageVariant = "user" | "ai" | "system" | "marketplace" | "warning";

export interface MessageAction {
  id: string;
  label: string;
  variant?: "primary" | "secondary" | "danger" | "ghost";
  icon?: React.ReactNode;
}

export interface MarketplaceSuggestion {
  id: string;
  name: string;
  description: string;
  author: string;
  rating: number;
  downloads: number;
  tags: string[];
}

export interface FlyMessageProps {
  /** Message variant */
  variant: MessageVariant;
  /** Message content (markdown supported for AI variant) */
  content: string;
  /** Timestamp */
  timestamp?: Date | number;
  /** Available actions */
  actions?: MessageAction[];
  /** Callback when action is clicked */
  onAction?: (actionId: string) => void;
  /** Callback when feedback is given */
  onFeedback?: (type: "positive" | "negative") => void;
  /** Whether message is currently streaming */
  isStreaming?: boolean;
  /** For marketplace variant: suggestion data */
  suggestion?: MarketplaceSuggestion;
  /** Custom className */
  className?: string;
  /** User avatar URL (for user variant) */
  userAvatar?: string;
  /** AI avatar/logo URL */
  aiAvatar?: string;
}

// ============================================================================
// Component
// ============================================================================

export const FlyMessage = React.forwardRef<HTMLDivElement, FlyMessageProps>(
  (
    {
      variant,
      content,
      timestamp,
      actions,
      onAction,
      onFeedback,
      isStreaming = false,
      suggestion,
      className,
      userAvatar,
      aiAvatar,
    },
    ref
  ) => {
    const [copied, setCopied] = useState(false);
    const [feedbackGiven, setFeedbackGiven] = useState<"positive" | "negative" | null>(null);

    const handleCopy = useCallback(async () => {
      try {
        await navigator.clipboard.writeText(content);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (err) {
        console.error("Failed to copy:", err);
      }
    }, [content]);

    const handleFeedback = useCallback(
      (type: "positive" | "negative") => {
        setFeedbackGiven(type);
        onFeedback?.(type);
      },
      [onFeedback]
    );

    const formattedTime = timestamp
      ? format(new Date(timestamp), "h:mm a")
      : null;

    return (
      <div
        ref={ref}
        className={cn(
          "group flex gap-3 px-4 py-3",
          variant === "user" && "flex-row-reverse",
          className
        )}
        role="article"
        aria-label={`${variant} message`}
      >
        {/* Avatar */}
        <MessageAvatar variant={variant} userAvatar={userAvatar} aiAvatar={aiAvatar} />

        {/* Content Container */}
        <div
          className={cn(
            "flex flex-col max-w-[85%] sm:max-w-[75%]",
            variant === "user" && "items-end"
          )}
        >
          {/* Message Bubble */}
          <MessageBubble
            variant={variant}
            content={content}
            isStreaming={isStreaming}
            suggestion={suggestion}
          />

          {/* Footer: Actions & Metadata */}
          <div
            className={cn(
              "flex items-center gap-2 mt-1.5",
              variant === "user" ? "flex-row-reverse" : "flex-row"
            )}
          >
            {/* Timestamp */}
            {formattedTime && (
              <span className="text-xs text-[var(--color-text-tertiary)]">
                {formattedTime}
              </span>
            )}

            {/* Action Buttons */}
            {actions && actions.length > 0 && (
              <div className="flex items-center gap-1">
                {actions.map((action) => (
                  <motion.button
                    key={action.id}
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                    onClick={() => onAction?.(action.id)}
                    className={cn(
                      "flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-full transition-colors",
                      action.variant === "primary" &&
                        "bg-[var(--color-brand-500)] text-white hover:bg-[var(--color-brand-600)]",
                      action.variant === "secondary" &&
                        "bg-[var(--color-bg-secondary)] text-[var(--color-text-primary)] hover:bg-[var(--color-bg-tertiary)]",
                      action.variant === "danger" &&
                        "bg-red-500/10 text-red-400 hover:bg-red-500/20",
                      (!action.variant || action.variant === "ghost") &&
                        "text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-secondary)]"
                    )}
                  >
                    {action.icon && <span className="w-3 h-3">{action.icon}</span>}
                    {action.label}
                  </motion.button>
                ))}
              </div>
            )}

            {/* Copy Button (AI messages only) */}
            {(variant === "ai" || variant === "system") && !isStreaming && (
              <motion.button
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={handleCopy}
                className={cn(
                  "p-1.5 rounded-md",
                  "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]",
                  "hover:bg-[var(--color-bg-secondary)]",
                  "transition-colors"
                )}
                aria-label={copied ? "Copied" : "Copy message"}
              >
                {copied ? (
                  <Check className="w-3.5 h-3.5 text-green-400" />
                ) : (
                  <Copy className="w-3.5 h-3.5" />
                )}
              </motion.button>
            )}

            {/* Feedback Buttons (AI messages only) */}
            {(variant === "ai" || variant === "system") && !isStreaming && onFeedback && (
              <div className="flex items-center gap-0.5">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => handleFeedback("positive")}
                  className={cn(
                    "p-1.5 rounded-md transition-colors",
                    feedbackGiven === "positive"
                      ? "text-green-400 bg-green-400/10"
                      : "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-secondary)]"
                  )}
                  aria-label="Helpful"
                >
                  <ThumbsUp className="w-3.5 h-3.5" />
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => handleFeedback("negative")}
                  className={cn(
                    "p-1.5 rounded-md transition-colors",
                    feedbackGiven === "negative"
                      ? "text-red-400 bg-red-400/10"
                      : "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-secondary)]"
                  )}
                  aria-label="Not helpful"
                >
                  <ThumbsDown className="w-3.5 h-3.5" />
                </motion.button>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }
);

FlyMessage.displayName = "FlyMessage";

// ============================================================================
// Sub-components
// ============================================================================

function MessageAvatar({
  variant,
  userAvatar,
  aiAvatar,
}: {
  variant: MessageVariant;
  userAvatar?: string;
  aiAvatar?: string;
}) {
  const avatarClasses = cn(
    "flex-shrink-0 w-8 h-8 rounded-full",
    "flex items-center justify-center",
    "overflow-hidden"
  );

  switch (variant) {
    case "user":
      return (
        <div className={cn(avatarClasses, "bg-[var(--color-brand-500)]")}>
          {userAvatar ? (
            <img src={userAvatar} alt="User" className="w-full h-full object-cover" />
          ) : (
            <User className="w-4 h-4 text-white" />
          )}
        </div>
      );

    case "ai":
      return (
        <div
          className={cn(
            avatarClasses,
            "bg-[var(--color-bg-secondary)]",
            "border border-[var(--color-brand-500)]"
          )}
        >
          {aiAvatar ? (
            <img src={aiAvatar} alt="AI" className="w-full h-full object-cover" />
          ) : (
            <Bot className="w-4 h-4 text-[var(--color-brand-500)]" />
          )}
        </div>
      );

    case "system":
      return (
        <div
          className={cn(avatarClasses, "bg-amber-500/10 border border-amber-500/30")}
        >
          <Lightbulb className="w-4 h-4 text-amber-400" />
        </div>
      );

    case "marketplace":
      return (
        <div
          className={cn(
            avatarClasses,
            "bg-[var(--color-bg-secondary)]",
            "border border-[var(--color-brand-500)/30]"
          )}
        >
          <Store className="w-4 h-4 text-[var(--color-brand-500)]" />
        </div>
      );

    case "warning":
      return (
        <div className={cn(avatarClasses, "bg-red-500/10 border border-red-500/30")}>
          <AlertTriangle className="w-4 h-4 text-red-400" />
        </div>
      );

    default:
      return null;
  }
}

function MessageBubble({
  variant,
  content,
  isStreaming,
  suggestion,
}: {
  variant: MessageVariant;
  content: string;
  isStreaming: boolean;
  suggestion?: MarketplaceSuggestion;
}) {
  const bubbleClasses = cn("relative px-4 py-3 text-sm leading-relaxed");

  switch (variant) {
    case "user":
      return (
        <div
          className={cn(
            bubbleClasses,
            "bg-[var(--color-bg-tertiary)]",
            "text-[var(--color-text-primary)]",
            "rounded-[18px_18px_4px_18px]",
            "max-w-full whitespace-pre-wrap break-words"
          )}
        >
          {content}
        </div>
      );

    case "ai":
      return (
        <div
          className={cn(
            bubbleClasses,
            "bg-[var(--color-bg-secondary)]",
            "text-[var(--color-text-primary)]",
            "rounded-[4px_18px_18px_18px]",
            "border-l-[3px] border-[var(--color-brand-500)]",
            "max-w-full"
          )}
        >
          <MarkdownContent content={content} isStreaming={isStreaming} />
        </div>
      );

    case "system":
      return (
        <div
          className={cn(
            bubbleClasses,
            "bg-amber-500/10",
            "text-[var(--color-text-primary)]",
            "rounded-lg",
            "border border-amber-500/30",
            "max-w-full"
          )}
        >
          <div className="flex items-start gap-2">
            <Info className="w-4 h-4 text-amber-400 flex-shrink-0 mt-0.5" />
            <div className="prose prose-invert prose-sm max-w-none">
              <MarkdownContent content={content} isStreaming={isStreaming} />
            </div>
          </div>
        </div>
      );

    case "marketplace":
      return (
        <div
          className={cn(
            bubbleClasses,
            "bg-[var(--color-bg-secondary)]",
            "text-[var(--color-text-primary)]",
            "rounded-lg",
            "border border-[var(--color-brand-500)]/30",
            "max-w-full"
          )}
        >
          <div className="prose prose-invert prose-sm max-w-none mb-3">
            <MarkdownContent content={content} isStreaming={isStreaming} />
          </div>
          {suggestion && <MarketplaceCard suggestion={suggestion} />}
        </div>
      );

    case "warning":
      return (
        <div
          className={cn(
            bubbleClasses,
            "bg-red-500/10",
            "text-[var(--color-text-primary)]",
            "rounded-lg",
            "border border-red-500/30",
            "max-w-full"
          )}
        >
          <div className="flex items-start gap-2">
            <AlertTriangle className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
            <span className="whitespace-pre-wrap break-words">{content}</span>
          </div>
        </div>
      );

    default:
      return null;
  }
}

function MarkdownContent({
  content,
  isStreaming,
}: {
  content: string;
  isStreaming: boolean;
}) {
  return (
    <div
      className={cn(
        "prose prose-invert prose-sm max-w-none",
        "prose-pre:bg-[var(--color-bg-tertiary)] prose-pre:border prose-pre:border-[var(--color-border)]",
        "prose-code:text-[var(--color-brand-400)] prose-code:bg-[var(--color-bg-tertiary)] prose-code:px-1 prose-code:py-0.5 prose-code:rounded",
        "prose-a:text-[var(--color-brand-400)] prose-a:no-underline hover:prose-a:underline",
        isStreaming && "animate-pulse"
      )}
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code({ className, children }) {
            const match = /language-(\w+)/.exec(className || "");
            const language = match ? match[1] : "";

            if (className && language) {
              return (
                <SyntaxHighlighter
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  style={vscDarkPlus as any}
                  language={language}
                  PreTag="div"
                  className="rounded-lg !bg-[var(--color-bg-tertiary)] !m-0"
                >
                  {String(children).replace(/\n$/, "")}
                </SyntaxHighlighter>
              );
            }

            return (
              <code
                className={cn(
                  "bg-[var(--color-bg-tertiary)] text-[var(--color-brand-400)]",
                  "px-1.5 py-0.5 rounded text-xs font-mono"
                )}
              >
                {children}
              </code>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

function MarketplaceCard({ suggestion }: { suggestion: MarketplaceSuggestion }) {
  return (
    <div
      className={cn(
        "mt-3 p-3 rounded-lg",
        "bg-[var(--color-bg-tertiary)]",
        "border border-[var(--color-border)]"
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <h4 className="font-semibold text-[var(--color-text-primary)] truncate">
            {suggestion.name}
          </h4>
          <p className="text-xs text-[var(--color-text-secondary)] mt-1 line-clamp-2">
            {suggestion.description}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3 mt-2 text-xs text-[var(--color-text-tertiary)]">
        <span>by {suggestion.author}</span>
        <span className="flex items-center gap-1">
          <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
          </svg>
          {suggestion.rating.toFixed(1)}
        </span>
        <span>{suggestion.downloads.toLocaleString()} downloads</span>
      </div>

      {suggestion.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-2">
          {suggestion.tags.map((tag) => (
            <span
              key={tag}
              className={cn(
                "px-2 py-0.5 text-xs rounded-full",
                "bg-[var(--color-bg-secondary)]",
                "text-[var(--color-text-tertiary)]"
              )}
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center gap-2 mt-3">
        <button
          className={cn(
            "flex-1 flex items-center justify-center gap-1.5",
            "px-3 py-1.5 text-xs font-medium",
            "bg-[var(--color-brand-500)] hover:bg-[var(--color-brand-600)]",
            "text-white rounded-md",
            "transition-colors"
          )}
        >
          <CheckCircle className="w-3.5 h-3.5" />
          Accept
        </button>
        <button
          className={cn(
            "flex items-center justify-center gap-1.5",
            "px-3 py-1.5 text-xs font-medium",
            "bg-[var(--color-bg-secondary)] hover:bg-[var(--color-bg-tertiary)]",
            "text-[var(--color-text-secondary)] rounded-md",
            "transition-colors"
          )}
        >
          <X className="w-3.5 h-3.5" />
          Dismiss
        </button>
      </div>
    </div>
  );
}

export default FlyMessage;

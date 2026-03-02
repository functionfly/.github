/**
 * ActivityComposer Component
 *
 * A simple form for posting status updates to the user's activity feed.
 * Includes text input with character limit and post button.
 *
 * @example
 * <ActivityComposer
 *   onSubmit={handlePost}
 *   maxLength={500}
 *   placeholder="What's on your mind?"
 * />
 */

import { useState, useCallback, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Send,
  Loader2,
  ImageIcon,
  Link2,
  Code2,
  Smile,
  X,
  AlertCircle,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface ActivityComposerProps {
  onSubmit: (content: string, metadata?: { type?: string; link?: string }) => Promise<void>;
  maxLength?: number;
  placeholder?: string;
  className?: string;
  avatarUrl?: string;
  userName?: string;
}

interface ComposerToolbarProps {
  onImageClick?: () => void;
  onLinkClick?: () => void;
  onCodeClick?: () => void;
  disabled?: boolean;
}

function ComposerToolbar({
  onImageClick,
  onLinkClick,
  onCodeClick,
  disabled,
}: ComposerToolbarProps) {
  const tools = [
    { icon: ImageIcon, label: "Add image", onClick: onImageClick },
    { icon: Link2, label: "Add link", onClick: onLinkClick },
    { icon: Code2, label: "Add code", onClick: onCodeClick },
  ];

  return (
    <TooltipProvider>
      <div className="flex items-center gap-1">
        {tools.map((tool) => (
          <Tooltip key={tool.label}>
            <TooltipTrigger asChild>
              <button
                type="button"
                onClick={tool.onClick}
                disabled={disabled}
                className={cn(
                  "p-2 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover transition-colors",
                  disabled && "opacity-50 cursor-not-allowed"
                )}
              >
                <tool.icon className="w-4 h-4" />
              </button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{tool.label}</p>
            </TooltipContent>
          </Tooltip>
        ))}
      </div>
    </TooltipProvider>
  );
}

export function ActivityComposer({
  onSubmit,
  maxLength = 500,
  placeholder = "Share what you're working on...",
  className,
  avatarUrl,
  userName,
}: ActivityComposerProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isFocused, setIsFocused] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const characterCount = content.length;
  const isOverLimit = characterCount > maxLength;
  const isEmpty = !content.trim();
  const canSubmit = !isEmpty && !isOverLimit && !isSubmitting;

  // Auto-resize textarea
  const adjustTextareaHeight = useCallback(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`;
    }
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;
    if (newContent.length <= maxLength + 50) {
      // Allow slight overflow for UX
      setContent(newContent);
      setError(null);
      adjustTextareaHeight();
    }
  };

  const handleSubmit = async () => {
    if (!canSubmit) return;

    setIsSubmitting(true);
    setError(null);

    try {
      await onSubmit(content.trim());
      setContent("");
      // Reset textarea height
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to post update");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Cmd/Ctrl + Enter to submit
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      handleSubmit();
    }
  };

  const clearContent = () => {
    setContent("");
    setError(null);
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  };

  // Calculate progress for visual indicator
  const progressPercent = Math.min(100, (characterCount / maxLength) * 100);
  const progressColor =
    progressPercent > 90 ? "text-error" : progressPercent > 75 ? "text-warning" : "text-text-muted";

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn(
        "bg-card border border-border-default rounded-xl p-4 transition-all",
        isFocused && "border-border-focus ring-2 ring-border-focus/20",
        className
      )}
    >
      <div className="flex gap-3">
        {/* Avatar */}
        <div className="shrink-0">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center text-white font-semibold overflow-hidden">
            {avatarUrl ? (
              <img src={avatarUrl} alt={userName} className="w-full h-full object-cover" />
            ) : (
              userName?.charAt(0).toUpperCase() || "?"
            )}
          </div>
        </div>

        {/* Input area */}
        <div className="flex-1 min-w-0">
          <Textarea
            ref={textareaRef}
            value={content}
            onChange={handleChange}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            rows={2}
            className={cn(
              "min-h-[80px] resize-none border-0 bg-transparent p-0 text-text-primary placeholder:text-text-muted focus-visible:ring-0 focus-visible:ring-offset-0",
              isOverLimit && "text-error"
            )}
            disabled={isSubmitting}
          />

          {/* Error message */}
          <AnimatePresence>
            {error && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                exit={{ opacity: 0, height: 0 }}
                className="flex items-center gap-2 text-error text-sm mt-2"
              >
                <AlertCircle className="w-4 h-4" />
                {error}
              </motion.div>
            )}
          </AnimatePresence>

          {/* Toolbar and actions */}
          <div className="flex items-center justify-between mt-3 pt-3 border-t border-border-subtle">
            {/* Left side - Toolbar */}
            <ComposerToolbar disabled={isSubmitting} />

            {/* Right side - Character count and submit */}
            <div className="flex items-center gap-3">
              {/* Character count */}
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "text-xs transition-colors",
                    progressColor,
                    characterCount === 0 && "text-text-muted"
                  )}
                >
                  {characterCount}/{maxLength}
                </span>
                {/* Circular progress indicator */}
                <svg className="w-5 h-5 -rotate-90" viewBox="0 0 20 20">
                  <circle
                    cx="10"
                    cy="10"
                    r="8"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className="text-border-subtle"
                  />
                  <circle
                    cx="10"
                    cy="10"
                    r="8"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeDasharray={`${progressPercent * 0.502} 50.2`}
                    className={cn("transition-all", progressColor)}
                  />
                </svg>
              </div>

              {/* Clear button (when has content) */}
              {content && !isSubmitting && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={clearContent}
                  className="h-8 px-2 text-text-muted hover:text-text-primary"
                >
                  <X className="w-4 h-4" />
                </Button>
              )}

              {/* Submit button */}
              <Button
                onClick={handleSubmit}
                disabled={!canSubmit}
                size="sm"
                className="gap-2"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Posting...
                  </>
                ) : (
                  <>
                    <Send className="w-4 h-4" />
                    Post
                  </>
                )}
              </Button>
            </div>
          </div>

          {/* Keyboard shortcut hint */}
          <p className="text-xs text-text-muted mt-2">
            Press{" "}
            <kbd className="px-1.5 py-0.5 bg-bg-tertiary rounded text-xs">Ctrl</kbd> +{" "}
            <kbd className="px-1.5 py-0.5 bg-bg-tertiary rounded text-xs">Enter</kbd> to post
          </p>
        </div>
      </div>
    </motion.div>
  );
}

// Mini composer for inline use
interface MiniActivityComposerProps {
  onSubmit: (content: string) => Promise<void>;
  className?: string;
}

export function MiniActivityComposer({ onSubmit, className }: MiniActivityComposerProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!content.trim() || isSubmitting) return;

    setIsSubmitting(true);
    try {
      await onSubmit(content.trim());
      setContent("");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={cn("flex gap-2", className)}>
      <input
        type="text"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
        placeholder="Write a quick update..."
        className="flex-1 bg-bg-tertiary border border-border-subtle rounded-lg px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-border-focus"
        disabled={isSubmitting}
      />
      <Button
        onClick={handleSubmit}
        disabled={!content.trim() || isSubmitting}
        size="sm"
        className="gap-1"
      >
        {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
      </Button>
    </div>
  );
}

export default ActivityComposer;

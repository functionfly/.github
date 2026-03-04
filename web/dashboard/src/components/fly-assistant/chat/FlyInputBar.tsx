/**
 * FlyInputBar.tsx
 *
 * Advanced input bar with slash commands, attachments, voice support,
 * context badges, and keyboard shortcuts.
 */

import React, { useRef, useState, useCallback, useEffect, KeyboardEvent } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Send,
  Paperclip,
  Mic,
  Command,
  X,
  FileText,
  Zap,
  Rocket,
  TestTube,
  Globe,
  Code,
  Sparkles,
  CornerDownLeft,
} from "lucide-react";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface SlashCommand {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  shortcut?: string;
}

export interface Attachment {
  id: string;
  name: string;
  type: string;
  size: number;
}

export interface FlyInputBarProps {
  /** Current input value */
  value: string;
  /** Callback when value changes */
  onChange: (value: string) => void;
  /** Callback when message is sent */
  onSend: () => void;
  /** Callback when command is selected */
  onCommand?: (commandId: string) => void;
  /** Current context (page/function the assistant is aware of) */
  context?: string;
  /** Whether input is disabled */
  disabled?: boolean;
  /** Whether in loading state */
  isLoading?: boolean;
  /** Custom placeholder text */
  placeholder?: string;
  /** Custom className */
  className?: string;
  /** Maximum character limit */
  maxLength?: number;
  /** Callback when attachment is added */
  onAttachmentAdd?: (files: FileList) => void;
  /** Callback when attachment is removed */
  onAttachmentRemove?: (id: string) => void;
  /** Current attachments */
  attachments?: Attachment[];
  /** Show voice button */
  showVoiceButton?: boolean;
}

// ============================================================================
// Slash Commands Configuration
// ============================================================================

const SLASH_COMMANDS: SlashCommand[] = [
  {
    id: "optimize",
    name: "optimize",
    description: "Optimize function performance and cost",
    icon: <Zap className="w-4 h-4" />,
    shortcut: "/optimize",
  },
  {
    id: "publish",
    name: "publish",
    description: "Publish function to marketplace",
    icon: <Rocket className="w-4 h-4" />,
    shortcut: "/publish",
  },
  {
    id: "test",
    name: "test",
    description: "Run tests on your function",
    icon: <TestTube className="w-4 h-4" />,
    shortcut: "/test",
  },
  {
    id: "deploy",
    name: "deploy",
    description: "Deploy function to production",
    icon: <Globe className="w-4 h-4" />,
    shortcut: "/deploy",
  },
  {
    id: "code",
    name: "code",
    description: "Generate or explain code",
    icon: <Code className="w-4 h-4" />,
    shortcut: "/code",
  },
];

// ============================================================================
// Component
// ============================================================================

export const FlyInputBar = React.forwardRef<HTMLTextAreaElement, FlyInputBarProps>(
  (
    {
      value,
      onChange,
      onSend,
      onCommand,
      context,
      disabled = false,
      isLoading = false,
      placeholder = "Ask me anything...",
      className,
      maxLength = 4000,
      onAttachmentAdd,
      onAttachmentRemove,
      attachments = [],
      showVoiceButton = true,
    },
    forwardedRef
  ) => {
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [showCommands, setShowCommands] = useState(false);
    const [selectedCommandIndex, setSelectedCommandIndex] = useState(0);
    const [isFocused, setIsFocused] = useState(false);

    // Merge refs
    const setRefs = useCallback(
      (element: HTMLTextAreaElement | null) => {
        (textareaRef as React.MutableRefObject<HTMLTextAreaElement | null>).current = element;
        if (typeof forwardedRef === "function") {
          forwardedRef(element);
        } else if (forwardedRef) {
          (forwardedRef as React.MutableRefObject<HTMLTextAreaElement | null>).current = element;
        }
      },
      [forwardedRef]
    );

    // Filter commands based on input
    const filteredCommands = SLASH_COMMANDS.filter((cmd) =>
      value.startsWith("/")
        ? cmd.name.toLowerCase().includes(value.slice(1).toLowerCase())
        : false
    );

    // Update command palette visibility
    useEffect(() => {
      if (value.startsWith("/")) {
        setShowCommands(true);
        setSelectedCommandIndex(0);
      } else {
        setShowCommands(false);
      }
    }, [value]);

    // Auto-resize textarea
    useEffect(() => {
      const textarea = textareaRef.current;
      if (textarea) {
        textarea.style.height = "auto";
        textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`;
      }
    }, [value]);

    // Handle keyboard shortcuts
    const handleKeyDown = useCallback(
      (e: KeyboardEvent<HTMLTextAreaElement>) => {
        // Command palette navigation
        if (showCommands && filteredCommands.length > 0) {
          switch (e.key) {
            case "ArrowDown":
              e.preventDefault();
              setSelectedCommandIndex((prev) =>
                prev < filteredCommands.length - 1 ? prev + 1 : prev
              );
              return;
            case "ArrowUp":
              e.preventDefault();
              setSelectedCommandIndex((prev) => (prev > 0 ? prev - 1 : 0));
              return;
            case "Enter":
              if (!e.shiftKey) {
                e.preventDefault();
                const cmd = filteredCommands[selectedCommandIndex];
                if (cmd) {
                  handleCommandSelect(cmd);
                }
              }
              return;
            case "Escape":
              setShowCommands(false);
              return;
          }
        }

        // Regular shortcuts
        switch (e.key) {
          case "Enter":
            if (!e.shiftKey && !showCommands) {
              e.preventDefault();
              if (value.trim() && !disabled && !isLoading) {
                onSend();
              }
            }
            break;
          case "k":
          case "K":
            if ((e.metaKey || e.ctrlKey) && !value.startsWith("/")) {
              e.preventDefault();
              onChange("/");
              textareaRef.current?.focus();
            }
            break;
        }
      },
      [
        showCommands,
        filteredCommands,
        selectedCommandIndex,
        value,
        disabled,
        isLoading,
        onSend,
        onChange,
      ]
    );

    // Handle command selection
    const handleCommandSelect = useCallback(
      (command: SlashCommand) => {
        onChange(`/${command.name} `);
        onCommand?.(command.id);
        setShowCommands(false);
        textareaRef.current?.focus();
      },
      [onChange, onCommand]
    );

    // Handle file attachment
    const handleFileSelect = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = e.target.files;
        if (files && files.length > 0) {
          onAttachmentAdd?.(files);
        }
        // Reset input
        if (fileInputRef.current) {
          fileInputRef.current.value = "";
        }
      },
      [onAttachmentAdd]
    );

    // Format file size
    const formatFileSize = (bytes: number): string => {
      if (bytes < 1024) return `${bytes} B`;
      if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
      return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    };

    const canSend = value.trim().length > 0 && !disabled && !isLoading;

    return (
      <div className={cn("relative", className)}>
        {/* Command Palette */}
        <AnimatePresence>
          {showCommands && filteredCommands.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 10 }}
              transition={{ duration: 0.15 }}
              className={cn(
                "absolute bottom-full left-0 right-0 mb-2",
                "bg-[var(--color-bg-secondary)]",
                "border border-[var(--color-border)]",
                "rounded-xl shadow-xl",
                "overflow-hidden z-50"
              )}
            >
              <div className="px-3 py-2 text-xs font-medium text-[var(--color-text-tertiary)] border-b border-[var(--color-border)]">
                Commands
              </div>
              <div className="max-h-48 overflow-y-auto">
                {filteredCommands.map((cmd, index) => (
                  <button
                    key={cmd.id}
                    onClick={() => handleCommandSelect(cmd)}
                    onMouseEnter={() => setSelectedCommandIndex(index)}
                    className={cn(
                      "w-full flex items-center gap-3 px-3 py-2.5",
                      "text-left transition-colors",
                      index === selectedCommandIndex
                        ? "bg-[var(--color-brand-500)]/10"
                        : "hover:bg-[var(--color-bg-tertiary)]"
                    )}
                  >
                    <span className="text-[var(--color-brand-500)]">{cmd.icon}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-[var(--color-text-primary)]">
                          /{cmd.name}
                        </span>
                        {cmd.shortcut && (
                          <kbd className="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 text-[10px] font-mono bg-[var(--color-bg-tertiary)] text-[var(--color-text-tertiary)] rounded">
                            {cmd.shortcut}
                          </kbd>
                        )}
                      </div>
                      <p className="text-xs text-[var(--color-text-secondary)] truncate">
                        {cmd.description}
                      </p>
                    </div>
                    {index === selectedCommandIndex && (
                      <CornerDownLeft className="w-4 h-4 text-[var(--color-brand-500)]" />
                    )}
                  </button>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Context Badge */}
        {context && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className={cn(
              "absolute -top-3 left-4",
              "flex items-center gap-1.5",
              "px-2 py-0.5",
              "text-xs font-medium",
              "bg-[var(--color-brand-500)]/10 text-[var(--color-brand-400)]",
              "rounded-full border border-[var(--color-brand-500)]/30",
              "z-10"
            )}
          >
            <Sparkles className="w-3 h-3" />
            <span className="max-w-[150px] truncate">{context}</span>
          </motion.div>
        )}

        {/* Main Input Container */}
        <div
          className={cn(
            "relative flex flex-col",
            "bg-[var(--color-bg-tertiary)]",
            "border border-[var(--color-border)]",
            "rounded-3xl",
            "transition-all duration-200",
            isFocused && "ring-2 ring-[var(--color-brand-500)] border-transparent",
            disabled && "opacity-60 cursor-not-allowed"
          )}
        >
          {/* Attachments */}
          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-2 px-4 pt-3 pb-2">
              {attachments.map((attachment) => (
                <motion.div
                  key={attachment.id}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.8 }}
                  className={cn(
                    "flex items-center gap-2",
                    "px-2.5 py-1.5",
                    "bg-[var(--color-bg-secondary)]",
                    "rounded-lg border border-[var(--color-border)]",
                    "text-sm"
                  )}
                >
                  <FileText className="w-4 h-4 text-[var(--color-brand-500)]" />
                  <span className="max-w-[120px] truncate text-[var(--color-text-primary)]">
                    {attachment.name}
                  </span>
                  <span className="text-xs text-[var(--color-text-tertiary)]">
                    {formatFileSize(attachment.size)}
                  </span>
                  <button
                    onClick={() => onAttachmentRemove?.(attachment.id)}
                    className={cn(
                      "p-0.5 rounded",
                      "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]",
                      "hover:bg-[var(--color-bg-tertiary)]",
                      "transition-colors"
                    )}
                    aria-label={`Remove ${attachment.name}`}
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </motion.div>
              ))}
            </div>
          )}

          {/* Input Row */}
          <div className="flex items-end gap-2 px-3 py-2">
            {/* Attachment Button */}
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={disabled || isLoading}
              className={cn(
                "p-2.5 rounded-xl",
                "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]",
                "hover:bg-[var(--color-bg-secondary)]",
                "transition-colors",
                "disabled:opacity-50 disabled:cursor-not-allowed"
              )}
              aria-label="Attach file"
            >
              <Paperclip className="w-5 h-5" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              onChange={handleFileSelect}
              className="hidden"
              multiple
              accept=".py,.js,.ts,.json,.md,.txt,.yaml,.yml"
            />

            {/* Textarea */}
            <textarea
              ref={setRefs}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              onKeyDown={handleKeyDown}
              onFocus={() => setIsFocused(true)}
              onBlur={() => setIsFocused(false)}
              disabled={disabled || isLoading}
              placeholder={placeholder}
              maxLength={maxLength}
              rows={1}
              className={cn(
                "flex-1 resize-none bg-transparent",
                "text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)]",
                "py-2.5 px-1",
                "focus:outline-none",
                "disabled:cursor-not-allowed",
                "min-h-[44px] max-h-[200px]"
              )}
              aria-label="Message input"
              aria-multiline="true"
            />

            {/* Voice Button */}
            {showVoiceButton && (
              <button
                disabled={disabled || isLoading}
                className={cn(
                  "p-2.5 rounded-xl",
                  "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]",
                  "hover:bg-[var(--color-bg-secondary)]",
                  "transition-colors",
                  "disabled:opacity-50 disabled:cursor-not-allowed"
                )}
                aria-label="Voice input"
              >
                <Mic className="w-5 h-5" />
              </button>
            )}

            {/* Send Button */}
            <motion.button
              whileHover={{ scale: canSend ? 1.05 : 1 }}
              whileTap={{ scale: canSend ? 0.95 : 1 }}
              onClick={onSend}
              disabled={!canSend}
              className={cn(
                "p-2.5 rounded-xl",
                "transition-all duration-200",
                canSend
                  ? "bg-[var(--color-brand-500)] text-white hover:bg-[var(--color-brand-600)]"
                  : "bg-[var(--color-bg-secondary)] text-[var(--color-text-tertiary)] cursor-not-allowed"
              )}
              aria-label="Send message"
            >
              {isLoading ? (
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                >
                  <Command className="w-5 h-5" />
                </motion.div>
              ) : (
                <Send className="w-5 h-5" />
              )}
            </motion.button>
          </div>

          {/* Character Count */}
          {value.length > maxLength * 0.8 && (
            <div
              className={cn(
                "px-4 pb-2 text-right text-xs",
                value.length >= maxLength
                  ? "text-red-400"
                  : "text-[var(--color-text-tertiary)]"
              )}
            >
              {value.length}/{maxLength}
            </div>
          )}
        </div>

        {/* Keyboard Shortcuts Hint */}
        <div className="flex items-center justify-between mt-2 px-2 text-xs text-[var(--color-text-tertiary)]">
          <div className="flex items-center gap-3">
            <span className="hidden sm:inline-flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 font-mono bg-[var(--color-bg-tertiary)] rounded">
                Enter
              </kbd>
              to send
            </span>
            <span className="hidden sm:inline-flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 font-mono bg-[var(--color-bg-tertiary)] rounded">
                Shift
              </kbd>
              +
              <kbd className="px-1.5 py-0.5 font-mono bg-[var(--color-bg-tertiary)] rounded">
                Enter
              </kbd>
              for new line
            </span>
          </div>
          <span className="hidden sm:inline-flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 font-mono bg-[var(--color-bg-tertiary)] rounded">
              {navigator.platform.includes("Mac") ? "⌘" : "Ctrl"}
            </kbd>
            +
            <kbd className="px-1.5 py-0.5 font-mono bg-[var(--color-bg-tertiary)] rounded">
              K
            </kbd>
            for commands
          </span>
        </div>
      </div>
    );
  }
);

FlyInputBar.displayName = "FlyInputBar";

export default FlyInputBar;

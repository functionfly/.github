/**
 * FlyTypingIndicator.tsx
 *
 * Animated AI typing indicator with bouncing dots and contextual hints.
 */

import React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Sparkles } from "lucide-react";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyTypingIndicatorProps {
  /** Whether the indicator is visible */
  isVisible: boolean;
  /** Optional context hint to display */
  context?: string;
  /** Custom className */
  className?: string;
  /** Delay before showing (in ms) */
  delay?: number;
}

// ============================================================================
// Animation Variants
// ============================================================================

const containerVariants = {
  hidden: {
    opacity: 0,
    y: 10,
  },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.3,
      ease: [0.25, 0.46, 0.45, 0.94] as const,
      staggerChildren: 0.1,
    },
  },
  exit: {
    opacity: 0,
    y: -10,
    transition: {
      duration: 0.2,
      ease: "easeIn" as const,
    },
  },
};

const dotVariants = {
  hidden: {
    opacity: 0,
    y: 0,
  },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.2,
    },
  },
};

const bounceVariants = {
  bounce: {
    y: [-4, 0, -4],
    transition: {
      duration: 0.6,
      repeat: Infinity,
      ease: "easeInOut",
    },
  },
};

const textVariants = {
  hidden: {
    opacity: 0,
  },
  visible: {
    opacity: 1,
    transition: {
      delay: 0.2,
      duration: 0.3,
    },
  },
};

// ============================================================================
// Component
// ============================================================================

export const FlyTypingIndicator: React.FC<FlyTypingIndicatorProps> = ({
  isVisible,
  context,
  className,
  delay = 0,
}) => {
  return (
    <AnimatePresence mode="wait">
      {isVisible && (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
          transition={{ delay: delay / 1000 }}
          className={cn(
            "flex items-center gap-3 px-4 py-3",
            "bg-[var(--color-bg-secondary)]",
            "border-l-[3px] border-[var(--color-brand-500)]",
            "rounded-[4px_18px_18px_18px]",
            "max-w-[85%]",
            className
          )}
          role="status"
          aria-live="polite"
          aria-label={context || "AI is typing"}
        >
          {/* AI Icon */}
          <div
            className={cn(
              "flex-shrink-0 w-8 h-8 rounded-full",
              "flex items-center justify-center",
              "bg-[var(--color-bg-tertiary)]",
              "border border-[var(--color-brand-500)]/30"
            )}
          >
            <motion.div
              animate={{
                scale: [1, 1.1, 1],
                opacity: [0.7, 1, 0.7],
              }}
              transition={{
                duration: 2,
                repeat: Infinity,
                ease: "easeInOut",
              }}
            >
              <Sparkles className="w-4 h-4 text-[var(--color-brand-500)]" />
            </motion.div>
          </div>

          {/* Dots Container */}
          <div className="flex items-center gap-1.5">
            {/* Bouncing Dots */}
            {[0, 1, 2].map((index) => (
              <motion.span
                key={index}
                variants={dotVariants}
                animate="bounce"
                custom={index}
                className={cn(
                  "w-2 h-2 rounded-full",
                  "bg-[var(--color-brand-500)]"
                )}
                style={{
                  animationDelay: `${index * 0.15}s`,
                }}
                transition={{
                  y: {
                    duration: 0.6,
                    repeat: Infinity,
                    repeatType: "reverse",
                    ease: "easeInOut",
                    delay: index * 0.15,
                  },
                }}
              />
            ))}
          </div>

          {/* Text Label */}
          <motion.span
            variants={textVariants}
            className="text-sm text-[var(--color-text-secondary)] ml-1"
          >
            {context || "AI is thinking"}
            <motion.span
              animate={{ opacity: [0, 1, 0] }}
              transition={{
                duration: 1.5,
                repeat: Infinity,
                ease: "easeInOut",
              }}
            >
              ...
            </motion.span>
          </motion.span>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

// ============================================================================
// Compact Variant
// ============================================================================

export interface FlyTypingIndicatorCompactProps {
  /** Whether the indicator is visible */
  isVisible: boolean;
  /** Custom className */
  className?: string;
}

export const FlyTypingIndicatorCompact: React.FC<FlyTypingIndicatorCompactProps> = ({
  isVisible,
  className,
}) => {
  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.8 }}
          transition={{ duration: 0.2 }}
          className={cn(
            "flex items-center gap-2 px-3 py-2",
            "bg-[var(--color-bg-tertiary)]",
            "rounded-full",
            className
          )}
          role="status"
          aria-label="AI is typing"
        >
          {[0, 1, 2].map((index) => (
            <motion.span
              key={index}
              className={cn(
                "w-1.5 h-1.5 rounded-full",
                "bg-[var(--color-brand-500)]"
              )}
              animate={{
                y: [0, -3, 0],
                opacity: [0.4, 1, 0.4],
              }}
              transition={{
                duration: 0.6,
                repeat: Infinity,
                ease: "easeInOut",
                delay: index * 0.15,
              }}
            />
          ))}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

// ============================================================================
// Skeleton Loading Variant
// ============================================================================

export interface FlyTypingSkeletonProps {
  /** Number of lines to show */
  lines?: number;
  /** Whether to show avatar placeholder */
  showAvatar?: boolean;
  /** Custom className */
  className?: string;
}

export const FlyTypingSkeleton: React.FC<FlyTypingSkeletonProps> = ({
  lines = 3,
  showAvatar = true,
  className,
}) => {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className={cn("flex gap-3 px-4 py-3", className)}
    >
      {/* Avatar Skeleton */}
      {showAvatar && (
        <motion.div
          className={cn(
            "flex-shrink-0 w-8 h-8 rounded-full",
            "bg-[var(--color-bg-tertiary)]"
          )}
          animate={{
            opacity: [0.5, 1, 0.5],
          }}
          transition={{
            duration: 1.5,
            repeat: Infinity,
            ease: "easeInOut",
          }}
        />
      )}

      {/* Lines Skeleton */}
      <div className="flex-1 space-y-2 max-w-[75%]">
        {Array.from({ length: lines }).map((_, index) => (
          <motion.div
            key={index}
            className={cn(
              "h-3 rounded",
              "bg-[var(--color-bg-tertiary)]"
            )}
            style={{
              width: index === lines - 1 ? "60%" : "100%",
            }}
            animate={{
              opacity: [0.3, 0.7, 0.3],
            }}
            transition={{
              duration: 1.5,
              repeat: Infinity,
              ease: "easeInOut",
              delay: index * 0.1,
            }}
          />
        ))}
      </div>
    </motion.div>
  );
};

export default FlyTypingIndicator;

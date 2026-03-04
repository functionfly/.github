/**
 * FlyAssistantPanel.tsx
 *
 * Main expandable container for the FlyAssistant interface.
 * Supports minimized, expanded, and fullscreen states with
 * smooth Framer Motion animations and resizable desktop support.
 */

import React, { useRef, useEffect, useCallback, useState } from "react";
import { motion, AnimatePresence, type Variants } from "framer-motion";
import { X, Minimize2, Maximize2, PanelLeft, GripVertical } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * Panel display mode
 */
export type PanelMode = "compact" | "default";

/**
 * Panel size state
 */
export type PanelSize = "minimized" | "expanded" | "fullscreen";

export interface FlyAssistantPanelProps {
  /** Whether the panel is currently open/visible */
  isOpen: boolean;
  /** Callback when panel should be closed */
  onClose: () => void;
  /** Display mode - compact or default */
  mode?: PanelMode;
  /** Child content (chat, insights, etc.) */
  children: React.ReactNode;
  /** Optional CSS class names */
  className?: string;
  /** Optional header title */
  title?: string;
  /** Optional test ID */
  testId?: string;
  /** Callback when panel is minimized */
  onMinimize?: () => void;
  /** Callback when panel is expanded */
  onExpand?: () => void;
  /** Current panel size state */
  size?: PanelSize;
}

// ============================================================================
// Animation Variants
// ============================================================================

const overlayVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { duration: 0.2 }
  },
  exit: {
    opacity: 0,
    transition: { duration: 0.15 }
  },
};

const panelVariants: Variants = {
  minimized: {
    width: 320,
    height: 56,
    x: 0,
    y: 0,
    transition: {
      type: "spring",
      stiffness: 300,
      damping: 30,
    },
  },
  expanded: {
    width: 420,
    height: 640,
    x: 0,
    y: 0,
    transition: {
      type: "spring",
      stiffness: 300,
      damping: 30,
    },
  },
  fullscreen: {
    width: "100vw",
    height: "100vh",
    x: 0,
    y: 0,
    transition: {
      type: "spring",
      stiffness: 300,
      damping: 30,
    },
  },
};

const contentVariants: Variants = {
  minimized: {
    opacity: 0,
    height: 0,
    transition: { duration: 0.15 },
  },
  expanded: {
    opacity: 1,
    height: "auto",
    transition: { duration: 0.2, delay: 0.1 },
  },
  fullscreen: {
    opacity: 1,
    height: "auto",
    transition: { duration: 0.2, delay: 0.1 },
  },
};

// ============================================================================
// Component
// ============================================================================

/**
 * FlyAssistantPanel - Main expandable container for FlyAssistant
 *
 * A floating panel that appears above the main application content.
 * Features include:
 * - Three size states: minimized (header only), expanded, fullscreen
 * - Compact mode option for smaller footprint
 * - Resizable on desktop (drag handle)
 * - Smooth open/close animations with Framer Motion
 * - Accessible header with minimize/close controls
 *
 * @example
 * ```tsx
 * <FlyAssistantPanel
 *   isOpen={isOpen}
 *   onClose={handleClose}
 *   size="expanded"
 *   mode="default"
 * >
 *   <ChatInterface />
 * </FlyAssistantPanel>
 * ```
 */
export function FlyAssistantPanel({
  isOpen,
  onClose,
  mode = "default",
  children,
  className,
  title = "FlyAssistant",
  testId = "fly-assistant-panel",
  onMinimize,
  onExpand,
  size = "expanded",
}: FlyAssistantPanelProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dimensions, setDimensions] = useState({ width: 420, height: 640 });

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen && size !== "minimized") {
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose, size]);

  // Handle resize (desktop only)
  const handleResizeStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (e: MouseEvent) => {
      if (panelRef.current) {
        const rect = panelRef.current.getBoundingClientRect();
        const newWidth = Math.max(320, Math.min(800, e.clientX - rect.left));
        const newHeight = Math.max(400, Math.min(800, e.clientY - rect.top));
        setDimensions({ width: newWidth, height: newHeight });
      }
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDragging]);

  // Determine if we should use custom dimensions
  const useCustomDimensions = size === "expanded" && !isDragging && dimensions.width !== 420;

  return (
    <AnimatePresence mode="wait">
      {isOpen && (
        <TooltipProvider>
          {/* Overlay (only in fullscreen) */}
          {size === "fullscreen" && (
            <motion.div
              className="fixed inset-0 bg-black/50 z-40"
              variants={overlayVariants}
              initial="hidden"
              animate="visible"
              exit="exit"
              onClick={onClose}
              aria-hidden="true"
            />
          )}

          {/* Main Panel */}
          <motion.div
            ref={panelRef}
            data-testid={testId}
            role="dialog"
            aria-modal={size !== "minimized"}
            aria-label={title}
            className={cn(
              // Positioning
              size === "fullscreen"
                ? "fixed inset-0 z-50"
                : "fixed bottom-24 right-6 z-50",
              // Styling
              "bg-bg-secondary rounded-2xl shadow-2xl",
              "border border-border-default",
              "flex flex-col overflow-hidden",
              "pointer-events-auto",
              // Compact mode
              mode === "compact" && "shadow-lg",
              className
            )}
            style={
              useCustomDimensions
                ? { width: dimensions.width, height: dimensions.height }
                : undefined
            }
            variants={panelVariants}
            initial="minimized"
            animate={size}
            exit="minimized"
          >
            {/* Header */}
            <header
              className={cn(
                "flex items-center justify-between",
                "px-4 py-3",
                "border-b border-border-default",
                "bg-bg-tertiary/50",
                size === "minimized" && "border-b-0"
              )}
            >
              {/* Left: Title and drag handle */}
              <div className="flex items-center gap-2">
                {size === "expanded" && (
                  <div
                    className="cursor-ns-resize p-1 rounded hover:bg-bg-hover"
                    onMouseDown={handleResizeStart}
                    role="button"
                    aria-label="Resize panel"
                    tabIndex={-1}
                  >
                    <GripVertical className="w-4 h-4 text-text-muted" />
                  </div>
                )}
                <div className="flex items-center gap-2">
                  <div className="w-6 h-6 rounded-lg bg-gradient-to-r from-brand-500 to-purple-500 flex items-center justify-center">
                    <span className="text-white text-xs font-bold">F</span>
                  </div>
                  <h2 className="text-sm font-semibold text-text-primary">
                    {title}
                  </h2>
                </div>
              </div>

              {/* Right: Controls */}
              <div className="flex items-center gap-1">
                {/* Minimize/Expand toggle */}
                {size !== "fullscreen" && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={size === "minimized" ? onExpand : onMinimize}
                        aria-label={size === "minimized" ? "Expand panel" : "Minimize panel"}
                      >
                        {size === "minimized" ? (
                          <PanelLeft className="w-4 h-4" />
                        ) : (
                          <Minimize2 className="w-4 h-4" />
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{size === "minimized" ? "Expand" : "Minimize"}</p>
                    </TooltipContent>
                  </Tooltip>
                )}

                {/* Fullscreen toggle (only in expanded) */}
                {size === "expanded" && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={onExpand}
                        aria-label="Enter fullscreen"
                      >
                        <Maximize2 className="w-4 h-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Fullscreen</p>
                    </TooltipContent>
                  </Tooltip>
                )}

                {/* Close button */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 hover:bg-error/10 hover:text-error"
                      onClick={onClose}
                      aria-label="Close panel"
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Close (Esc)</p>
                  </TooltipContent>
                </Tooltip>
              </div>
            </header>

            {/* Content Area */}
            <motion.div
              className="flex-1 overflow-hidden flex flex-col"
              variants={contentVariants}
              animate={size}
            >
              <div className={cn(
                "flex-1 overflow-y-auto",
                mode === "compact" ? "p-3" : "p-4"
              )}>
                {children}
              </div>
            </motion.div>

            {/* Resize handle (bottom-right corner) */}
            {size === "expanded" && (
              <div
                className="absolute bottom-0 right-0 w-4 h-4 cursor-nwse-resize"
                onMouseDown={handleResizeStart}
                role="button"
                aria-label="Resize panel"
                tabIndex={-1}
              >
                <div className="absolute bottom-1 right-1 w-2 h-2 border-r-2 border-b-2 border-text-muted/30 rounded-br" />
              </div>
            )}
          </motion.div>
        </TooltipProvider>
      )}
    </AnimatePresence>
  );
}

export default FlyAssistantPanel;

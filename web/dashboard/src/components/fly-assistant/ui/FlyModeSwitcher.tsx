/**
 * FlyModeSwitcher.tsx
 *
 * Switch between different assistant modes with smooth animated transitions.
 * Supports Chat, Insight, Marketplace, and Onboarding modes.
 *
 * @module fly-assistant/ui
 */

import React, { useCallback, useRef, useEffect, useState } from "react";
import { motion } from "framer-motion";
import {
  MessageSquare,
  Lightbulb,
  Store,
  GraduationCap,
} from "lucide-react";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export type AssistantMode = "chat" | "insight" | "marketplace" | "onboarding";

export interface FlyModeSwitcherProps {
  /** Currently active mode */
  currentMode: AssistantMode;
  /** Callback when mode changes */
  onModeChange: (mode: AssistantMode) => void;
  /** Available modes (defaults to all) */
  availableModes?: AssistantMode[];
  /** Disabled modes */
  disabledModes?: AssistantMode[];
  /** Custom className */
  className?: string;
}

interface ModeConfig {
  id: AssistantMode;
  label: string;
  icon: React.ElementType;
  description: string;
}

// ============================================================================
// Mode Configurations
// ============================================================================

const MODES: ModeConfig[] = [
  {
    id: "chat",
    label: "Chat",
    icon: MessageSquare,
    description: "Standard conversation mode",
  },
  {
    id: "insight",
    label: "Insights",
    icon: Lightbulb,
    description: "Proactive insights dashboard",
  },
  {
    id: "marketplace",
    label: "Marketplace",
    icon: Store,
    description: "Function discovery",
  },
  {
    id: "onboarding",
    label: "Help",
    icon: GraduationCap,
    description: "First-time user help",
  },
];

// ============================================================================
// Component
// ============================================================================

export const FlyModeSwitcher: React.FC<FlyModeSwitcherProps> = ({
  currentMode,
  onModeChange,
  availableModes = ["chat", "insight", "marketplace", "onboarding"],
  disabledModes = [],
  className,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [indicatorStyle, setIndicatorStyle] = useState({ left: 0, width: 0 });

  // Filter to only available modes
  const visibleModes = MODES.filter((mode) => availableModes.includes(mode.id));

  // Update indicator position when currentMode changes
  useEffect(() => {
    if (!containerRef.current) return;

    const activeButton = containerRef.current.querySelector(
      `[data-mode="${currentMode}"]`
    ) as HTMLElement;

    if (activeButton) {
      const containerRect = containerRef.current.getBoundingClientRect();
      const buttonRect = activeButton.getBoundingClientRect();

      setIndicatorStyle({
        left: buttonRect.left - containerRect.left,
        width: buttonRect.width,
      });
    }
  }, [currentMode, visibleModes.length]);

  // Handle mode change
  const handleModeChange = useCallback(
    (mode: AssistantMode) => {
      if (disabledModes.includes(mode)) return;
      onModeChange(mode);
    },
    [disabledModes, onModeChange]
  );

  // Handle keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, mode: AssistantMode) => {
      const currentIndex = visibleModes.findIndex((m) => m.id === mode);

      switch (e.key) {
        case "ArrowLeft":
          e.preventDefault();
          const prevIndex =
            currentIndex > 0 ? currentIndex - 1 : visibleModes.length - 1;
          const prevMode = visibleModes[prevIndex];
          if (!disabledModes.includes(prevMode.id)) {
            onModeChange(prevMode.id);
          }
          break;
        case "ArrowRight":
          e.preventDefault();
          const nextIndex =
            currentIndex < visibleModes.length - 1 ? currentIndex + 1 : 0;
          const nextMode = visibleModes[nextIndex];
          if (!disabledModes.includes(nextMode.id)) {
            onModeChange(nextMode.id);
          }
          break;
      }
    },
    [visibleModes, disabledModes, onModeChange]
  );

  if (visibleModes.length === 0) return null;

  return (
    <div
      ref={containerRef}
      className={cn(
        "relative inline-flex items-center p-1 rounded-lg",
        "bg-[var(--color-bg-tertiary)]",
        className
      )}
      role="tablist"
      aria-label="Assistant mode switcher"
    >
      {/* Animated indicator */}
      <motion.div
        className="absolute h-[calc(100%-8px)] top-1 rounded-md bg-[var(--color-brand-600)]"
        initial={false}
        animate={{
          left: indicatorStyle.left,
          width: indicatorStyle.width,
        }}
        transition={{
          type: "spring",
          stiffness: 500,
          damping: 35,
        }}
        aria-hidden="true"
      />

      {/* Mode tabs */}
      {visibleModes.map((mode) => {
        const isActive = currentMode === mode.id;
        const isDisabled = disabledModes.includes(mode.id);
        const Icon = mode.icon;
        const IconComponent = Icon as React.ComponentType<{ className?: string }>;

        return (
          <button
            key={mode.id}
            data-mode={mode.id}
            onClick={() => handleModeChange(mode.id)}
            onKeyDown={(e) => handleKeyDown(e, mode.id)}
            disabled={isDisabled}
            role="tab"
            aria-selected={isActive}
            aria-disabled={isDisabled}
            aria-label={`${mode.label}: ${mode.description}`}
            tabIndex={isActive ? 0 : -1}
            className={cn(
              "relative z-10 flex items-center gap-1.5 px-3 py-1.5",
              "text-[13px] font-medium rounded-md",
              "transition-colors duration-200",
              "focus:outline-none focus-visible:ring-2",
              "focus-visible:ring-[var(--color-brand-500)]/50",
              isActive
                ? "text-white"
                : "text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]",
              isDisabled &&
                "opacity-50 cursor-not-allowed hover:text-[var(--color-text-secondary)]"
            )}
          >
            <IconComponent
              className={cn(
                "w-3.5 h-3.5 transition-transform duration-200",
                isActive && "scale-110"
              )}
              aria-hidden="true"
            />
            <span>{mode.label}</span>

            {/* Disabled indicator */}
            {isDisabled && (
              <span className="sr-only">(disabled)</span>
            )}
          </button>
        );
      })}
    </div>
  );
};

export default FlyModeSwitcher;

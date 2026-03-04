/**
 * FlyCommandPalette.tsx
 *
 * Power-user command palette for advanced users.
 * Supports keyboard shortcuts, search, and grouped commands.
 *
 * @module fly-assistant/ui
 */

import React, { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import {
  Search,
  X,
  LayoutDashboard,
  FunctionSquare,
  Settings,
  Rocket,
  Play,
  Zap,
  ArrowRight,
  CornerDownLeft,
  Command,
  Clock,
  ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyCommandPaletteProps {
  /** Whether the palette is open */
  isOpen: boolean;
  /** Callback to close the palette */
  onClose: () => void;
  /** Callback when a function is selected */
  onSelectFunction: (funcId: string) => void;
  /** Callback when a page is selected */
  onSelectPage: (page: string) => void;
  /** Callback when an action is triggered */
  onTriggerAction: (action: string) => void;
  /** Available functions to search */
  availableFunctions?: Array<{ id: string; name: string }>;
  /** Recent commands history */
  recentCommands?: string[];
  /** Custom className */
  className?: string;
}

type CommandCategory = "pages" | "functions" | "actions" | "recent";

interface CommandItem {
  id: string;
  label: string;
  category: CommandCategory;
  icon: React.ReactNode;
  shortcut?: string;
  keywords?: string[];
  action: () => void;
}

// ============================================================================
// Page & Action Definitions
// ============================================================================

const PAGES = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, shortcut: "G D" },
  { id: "functions", label: "Functions", icon: FunctionSquare, shortcut: "G F" },
  { id: "settings", label: "Settings", icon: Settings, shortcut: "G S" },
];

const ACTIONS = [
  { id: "deploy", label: "Deploy", icon: Rocket, shortcut: "⌘ D" },
  { id: "test", label: "Test Function", icon: Play, shortcut: "⌘ T" },
  { id: "optimize", label: "Optimize", icon: Zap, shortcut: "⌘ O" },
];

// ============================================================================
// Component
// ============================================================================

export const FlyCommandPalette: React.FC<FlyCommandPaletteProps> = ({
  isOpen,
  onClose,
  onSelectFunction,
  onSelectPage,
  onTriggerAction,
  availableFunctions = [],
  recentCommands = [],
  className,
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  // Build command list
  const commands: CommandItem[] = useMemo(() => {
    const items: CommandItem[] = [];

    // Recent commands first
    if (recentCommands.length > 0 && !searchQuery) {
      recentCommands.forEach((cmdId) => {
        const action = ACTIONS.find((a) => a.id === cmdId);
        const page = PAGES.find((p) => p.id === cmdId);
        if (action) {
          items.push({
            id: `recent-${cmdId}`,
            label: action.label,
            category: "recent",
            icon: <action.icon className="w-4 h-4" />,
            action: () => onTriggerAction(cmdId),
          });
        } else if (page) {
          items.push({
            id: `recent-${cmdId}`,
            label: page.label,
            category: "recent",
            icon: <page.icon className="w-4 h-4" />,
            action: () => onSelectPage(cmdId),
          });
        }
      });
    }

    // Pages
    PAGES.forEach((page) => {
      items.push({
        id: `page-${page.id}`,
        label: page.label,
        category: "pages",
        icon: <page.icon className="w-4 h-4" />,
        shortcut: page.shortcut,
        keywords: [page.label.toLowerCase(), page.id],
        action: () => onSelectPage(page.id),
      });
    });

    // Functions
    availableFunctions.forEach((func) => {
      items.push({
        id: `func-${func.id}`,
        label: func.name,
        category: "functions",
        icon: <FunctionSquare className="w-4 h-4" />,
        keywords: [func.name.toLowerCase(), func.id.toLowerCase()],
        action: () => onSelectFunction(func.id),
      });
    });

    // Actions
    ACTIONS.forEach((action) => {
      items.push({
        id: `action-${action.id}`,
        label: action.label,
        category: "actions",
        icon: <action.icon className="w-4 h-4" />,
        shortcut: action.shortcut,
        keywords: [action.label.toLowerCase(), action.id],
        action: () => onTriggerAction(action.id),
      });
    });

    return items;
  }, [availableFunctions, recentCommands, searchQuery, onSelectFunction, onSelectPage, onTriggerAction]);

  // Filter commands by search
  const filteredCommands = useMemo(() => {
    if (!searchQuery.trim()) return commands;
    const query = searchQuery.toLowerCase();
    return commands.filter(
      (cmd) =>
        cmd.label.toLowerCase().includes(query) ||
        cmd.keywords?.some((k) => k.includes(query))
    );
  }, [commands, searchQuery]);

  // Reset selection when filtered changes
  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredCommands.length, searchQuery]);

  // Focus input on open
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
    } else {
      setSearchQuery("");
    }
  }, [isOpen]);

  // Keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setSelectedIndex((prev) =>
            prev < filteredCommands.length - 1 ? prev + 1 : prev
          );
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelectedIndex((prev) => (prev > 0 ? prev - 1 : prev));
          break;
        case "Enter":
          e.preventDefault();
          const selected = filteredCommands[selectedIndex];
          if (selected) {
            selected.action();
            onClose();
          }
          break;
        case "Escape":
          e.preventDefault();
          onClose();
          break;
      }
    },
    [filteredCommands, selectedIndex, onClose]
  );

  // Group commands by category
  const groupedCommands = useMemo(() => {
    const groups: Record<CommandCategory, CommandItem[]> = {
      recent: [],
      pages: [],
      functions: [],
      actions: [],
    };

    filteredCommands.forEach((cmd) => {
      groups[cmd.category].push(cmd);
    });

    return groups;
  }, [filteredCommands]);

  const categoryLabels: Record<CommandCategory, string> = {
    recent: "Recent",
    pages: "Pages",
    functions: "Functions",
    actions: "Actions",
  };

  return (
    <DialogPrimitive.Root open={isOpen} onOpenChange={onClose}>
      <DialogPrimitive.Portal>
        {/* Overlay */}
        <DialogPrimitive.Overlay asChild>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
            onClick={onClose}
          />
        </DialogPrimitive.Overlay>

        {/* Modal */}
        <DialogPrimitive.Content asChild>
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: -20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: -20 }}
            transition={{ type: "spring", stiffness: 400, damping: 30 }}
            className={cn(
              "fixed left-1/2 top-[20%] -translate-x-1/2 z-50",
              "w-full max-w-[640px] mx-auto",
              "bg-[var(--color-bg-primary)] border border-[var(--color-border)]",
              "rounded-xl shadow-2xl overflow-hidden",
              className
            )}
            onKeyDown={handleKeyDown}
          >
            {/* Search input */}
            <div className="flex items-center gap-3 px-4 py-3 border-b border-[var(--color-border)]">
              <Search
                className="w-5 h-5 text-[var(--color-text-muted)]"
                aria-hidden="true"
              />
              <input
                ref={inputRef}
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search functions, pages, or actions..."
                className={cn(
                  "flex-1 bg-transparent text-[var(--color-text-primary)]",
                  "placeholder:text-[var(--color-text-muted)]",
                  "focus:outline-none text-base"
                )}
                aria-label="Search commands"
              />
              <div className="flex items-center gap-1.5">
                <kbd className="hidden sm:flex items-center gap-0.5 px-1.5 py-0.5 text-xs bg-[var(--color-bg-tertiary)] text-[var(--color-text-muted)] rounded border border-[var(--color-border)]">
                  <Command className="w-3 h-3" />
                  <span>K</span>
                </kbd>
                <DialogPrimitive.Close asChild>
                  <button
                    className={cn(
                      "p-1 rounded",
                      "text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]",
                      "hover:bg-[var(--color-bg-secondary)] transition-colors"
                    )}
                    aria-label="Close"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </DialogPrimitive.Close>
              </div>
            </div>

            {/* Command list */}
            <div className="max-h-[400px] overflow-y-auto py-2">
              {filteredCommands.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                  <Search className="w-10 h-10 mb-3 opacity-30" />
                  <p className="text-sm">No commands found</p>
                  <p className="text-xs mt-1 opacity-70">
                    Try a different search term
                  </p>
                </div>
              ) : (
                (Object.keys(groupedCommands) as CommandCategory[]).map(
                  (category) => {
                    const items = groupedCommands[category];
                    if (items.length === 0) return null;

                    return (
                      <div key={category}>
                        {/* Category header */}
                        <div className="px-4 py-1.5 text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                          {categoryLabels[category]}
                        </div>

                        {/* Items */}
                        {items.map((cmd, index) => {
                          const globalIndex = filteredCommands.findIndex(
                            (c) => c.id === cmd.id
                          );
                          const isSelected = globalIndex === selectedIndex;

                          return (
                            <button
                              key={cmd.id}
                              onClick={() => {
                                cmd.action();
                                onClose();
                              }}
                              onMouseEnter={() => setSelectedIndex(globalIndex)}
                              className={cn(
                                "w-full flex items-center gap-3 px-4 py-2.5 text-left",
                                "transition-colors duration-150",
                                isSelected
                                  ? "bg-[var(--color-brand-500)]/20 border-l-3 border-[var(--color-brand-500)]"
                                  : "hover:bg-[var(--color-bg-secondary)] border-l-3 border-transparent"
                              )}
                              aria-selected={isSelected}
                            >
                              {/* Icon */}
                              <span
                                className={cn(
                                  "text-[var(--color-text-muted)]",
                                  isSelected && "text-[var(--color-brand-500)]"
                                )}
                              >
                                {cmd.icon}
                              </span>

                              {/* Label */}
                              <span
                                className={cn(
                                  "flex-1 text-sm",
                                  isSelected
                                    ? "text-[var(--color-text-primary)]"
                                    : "text-[var(--color-text-secondary)]"
                                )}
                              >
                                {cmd.label}
                              </span>

                              {/* Shortcut */}
                              {cmd.shortcut && (
                                <kbd className="hidden sm:flex items-center gap-0.5 px-1.5 py-0.5 text-xs bg-[var(--color-bg-tertiary)] text-[var(--color-text-muted)] rounded">
                                  {cmd.shortcut.split(" ").map((key, i) => (
                                    <React.Fragment key={i}>
                                      {i > 0 && <span className="opacity-50"> </span>}
                                      <span>{key}</span>
                                    </React.Fragment>
                                  ))}
                                </kbd>
                              )}

                              {/* Selection indicator */}
                              {isSelected && (
                                <CornerDownLeft className="w-4 h-4 text-[var(--color-brand-500)]" />
                              )}
                            </button>
                          );
                        })}
                      </div>
                    );
                  }
                )
              )}
            </div>

            {/* Footer */}
            <div className="flex items-center justify-between px-4 py-2 border-t border-[var(--color-border)] bg-[var(--color-bg-secondary)]/50">
              <div className="flex items-center gap-3 text-xs text-[var(--color-text-muted)]">
                <span className="flex items-center gap-1">
                  <CornerDownLeft className="w-3 h-3" /> to select
                </span>
                <span className="flex items-center gap-1">
                  <ArrowRight className="w-3 h-3 rotate-[-90deg]" /> {""}
                  <ArrowRight className="w-3 h-3 rotate-90" /> to navigate
                </span>
              </div>
              <span className="text-xs text-[var(--color-text-muted)]">
                {filteredCommands.length} commands
              </span>
            </div>
          </motion.div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
};

export default FlyCommandPalette;

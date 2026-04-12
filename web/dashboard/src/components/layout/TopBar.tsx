import { Menu, Bell, Command, Sparkles, Zap, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Breadcrumb } from "./Breadcrumb";
import { UserMenu } from "./UserMenu";
import { SearchButton } from "./SearchButton";
import { EnterpriseBadge } from "@/components/enterprise";
import { useNavigationStatus } from "@/hooks/useNavigationStatus";
import { useContextualActions } from "@/hooks/useContextualActions";
import { useThemeStore } from "@/stores/themeStore";
import { usePlan } from "@/hooks/usePlan";
import { useState, useCallback, useEffect } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface TopBarProps {
  onMenuClick: () => void;
  className?: string;
}

// Quick action shortcuts
const QUICK_ACTIONS = [
  { key: 'g', label: 'Go to...', icon: Command },
  { key: 'n', label: 'New Function', icon: Sparkles },
  { key: 'a', label: 'Agents', icon: Zap },
];

export function TopBar({ onMenuClick, className }: TopBarProps) {
  const status = useNavigationStatus();
  const contextualActions = useContextualActions();
  const { isEnterprise, plan } = usePlan();
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [notificationPulse, setNotificationPulse] = useState(false);

  // Calculate total notifications
  const totalNotifications = status.functions.pendingDeployments +
    (status.functions.hasIssues ? 1 : 0) +
    (status.providers.hasOffline ? 1 : 0) +
    (status.analytics.hasAlerts ? 1 : 0) +
    (status.settings.hasWarnings ? 1 : 0);

  // Pulse animation when notifications change
  useEffect(() => {
    if (totalNotifications > 0) {
      setNotificationPulse(true);
      const timer = setTimeout(() => setNotificationPulse(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [totalNotifications]);

  // Keyboard shortcuts
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Command palette: Cmd/Ctrl + K
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      setShowCommandPalette(true);
    }
  }, []);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  // Get status indicator color
  const getStatusColor = () => {
    if (status.providers.hasOffline || status.functions.hasIssues) return 'bg-red-500';
    if (status.analytics.hasAlerts) return 'bg-amber-500';
    if (status.functions.pendingDeployments > 0) return 'bg-blue-500';
    return null;
  };

  const statusColor = getStatusColor();

  return (
    <TooltipProvider delayDuration={0}>
      <>
        <header
          className={cn(
            "h-16 bg-aviation-bg-primary/95 backdrop-blur-xl",
            "border-b border-aviation-border-panel",
            "sticky top-0 z-30",
            className
          )}
        >
          <div className="h-full px-4 lg:px-6 flex items-center justify-between">
            {/* Left Section */}
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="icon"
                className="lg:hidden text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                onClick={onMenuClick}
                aria-label="Open navigation menu"
              >
                <Menu className="w-5 h-5" />
              </Button>

              {/* Dynamic Breadcrumb */}
              <Breadcrumb />
            </div>

            {/* Center: Contextual Actions with Aviation Styling */}
            {contextualActions.length > 0 && (
              <div className="hidden md:flex items-center gap-2">
                {contextualActions.map((action) => {
                  const Icon = action.icon;
                  return (
                    <Button
                      key={action.id}
                      variant={action.variant || "default"}
                      size="sm"
                      className={cn(
                        "aviation-button",
                        action.variant === "default" && "bg-aviation-amber text-aviation-bg-primary hover:bg-aviation-amber-glow"
                      )}
                      onClick={action.onClick}
                    >
                      <Icon className="w-4 h-4 mr-2" />
                      {action.label}
                    </Button>
                  );
                })}
              </div>
            )}

            {/* Right Section */}
            <div className="flex items-center gap-1 sm:gap-2">
              {/* Command Palette Trigger - Desktop */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setShowCommandPalette(true)}
                    className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-aviation-text-muted bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-all"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span className="hidden lg:inline">Search</span>
                    <kbd className="hidden xl:inline-flex items-center text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                      ⌘K
                    </kbd>
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>Command Palette</p>
                </TooltipContent>
              </Tooltip>

              {/* Global Search - Mobile/Tablet */}
              <div className="md:hidden">
                <SearchButton />
              </div>

              {/* Notifications */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={cn(
                      "relative text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument",
                      notificationPulse && "animate-pulse"
                    )}
                    aria-label={totalNotifications > 0 ? `Notifications (${totalNotifications} unread)` : 'Notifications'}
                  >
                    <Bell className="w-5 h-5" />
                    {totalNotifications > 0 && (
                      <>
                        {/* Status indicator dot */}
                        {statusColor && (
                          <span className={cn(
                            "absolute top-1 right-1 w-2 h-2 rounded-full",
                            statusColor
                          )} />
                        )}
                        {/* Count badge */}
                        <span className="absolute -top-1 -right-1 flex items-center justify-center min-w-[18px] h-[18px] bg-aviation-red text-white text-[10px] font-bold rounded-full px-1">
                          {totalNotifications > 9 ? '9+' : totalNotifications}
                        </span>
                      </>
                    )}
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <div className="space-y-1">
                    <p className="font-medium">Notifications</p>
                    {totalNotifications > 0 ? (
                      <div className="text-xs text-aviation-text-muted space-y-0.5">
                        {status.functions.pendingDeployments > 0 && (
                          <p>{status.functions.pendingDeployments} pending deployments</p>
                        )}
                        {status.functions.hasIssues && <p className="text-aviation-red">Function issues detected</p>}
                        {status.providers.hasOffline && <p className="text-aviation-red">Provider offline</p>}
                        {status.analytics.hasAlerts && <p className="text-aviation-amber">Analytics alert</p>}
                        {status.settings.hasWarnings && <p className="text-aviation-amber">Settings warning</p>}
                      </div>
                    ) : (
                      <p className="text-xs text-aviation-text-muted">No new notifications</p>
                    )}
                  </div>
                </TooltipContent>
              </Tooltip>

              {/* Plan Badge (non-enterprise) */}
              {!isEnterprise && plan && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className={cn(
                      "hidden sm:flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium cursor-pointer",
                      "bg-aviation-bg-instrument border border-aviation-border-instrument",
                      "text-aviation-text-secondary hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-colors"
                    )}>
                      <Shield className="w-3 h-3" />
                      <span className="capitalize">{plan}</span>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Current Plan: {plan}</p>
                    <p className="text-xs text-aviation-text-muted">Click to upgrade</p>
                  </TooltipContent>
                </Tooltip>
              )}

              {/* Enterprise Badge */}
              <EnterpriseBadge />

              {/* User Menu */}
              <UserMenu />
            </div>
          </div>
        </header>

        {/* Command Palette Overlay */}
        <AnimatePresence>
          {showCommandPalette && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-start justify-center pt-[20vh]"
              onClick={() => setShowCommandPalette(false)}
            >
              <motion.div
                initial={{ opacity: 0, y: -20, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -20, scale: 0.95 }}
                transition={{ type: "spring", stiffness: 300, damping: 30 }}
                className="w-full max-w-2xl mx-4 bg-aviation-bg-primary border border-aviation-border-panel rounded-xl shadow-2xl overflow-hidden"
                onClick={(e) => e.stopPropagation()}
              >
                {/* Search Input */}
                <div className="flex items-center gap-3 px-4 py-4 border-b border-aviation-border-panel">
                  <Command className="w-5 h-5 text-aviation-text-muted" />
                  <input
                    type="text"
                    placeholder="Search functions, agents, providers..."
                    className="flex-1 text-base text-aviation-text-primary placeholder:text-aviation-text-dim bg-transparent focus:outline-none"
                    autoFocus
                  />
                  <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-2 py-1 rounded">
                    ESC
                  </kbd>
                </div>

                {/* Quick Actions */}
                <div className="p-2">
                  <p className="px-3 py-2 text-xs font-semibold text-aviation-text-muted uppercase tracking-wider">
                    Quick Actions
                  </p>
                  <div className="space-y-1">
                    {QUICK_ACTIONS.map((action) => (
                      <button
                        key={action.key}
                        className="w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <action.icon className="w-4 h-4" />
                          <span>{action.label}</span>
                        </div>
                        <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                          ⌘{action.key.toUpperCase()}
                        </kbd>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Footer */}
                <div className="px-4 py-3 bg-aviation-bg-secondary border-t border-aviation-border-panel text-xs text-aviation-text-muted">
                  <p className="flex items-center gap-2">
                    <span>Use</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↑</kbd>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↓</kbd>
                    <span>to navigate,</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↵</kbd>
                    <span>to select</span>
                  </p>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </>
    </TooltipProvider>
  );
}

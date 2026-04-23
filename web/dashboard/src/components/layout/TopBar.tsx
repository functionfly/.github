import { Menu, Bell, Command, Sparkles, Zap, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Breadcrumb } from "./Breadcrumb";
import { UserMenu } from "./UserMenu";
import { GlobalCommandPalette } from "./GlobalCommandPalette";
import { EnterpriseBadge } from "@/components/enterprise";
import { LanguagePicker } from "@/components/common/LanguagePicker";
import { useNavigationStatus } from "@/hooks/useNavigationStatus";
import { useContextualActions } from "@/hooks/useContextualActions";
import { useThemeStore } from "@/stores/themeStore";
import { usePlan } from "@/hooks/usePlan";
import { useTranslation } from "react-i18next";
import { useState, useCallback, useEffect } from "react";
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

export function TopBar({ onMenuClick, className }: TopBarProps) {
  const { t } = useTranslation();
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
                    <span className="hidden lg:inline">{t('topbar.search')}</span>
                    <kbd className="hidden xl:inline-flex items-center text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                      ⌘K
                    </kbd>
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>{t('topbar.commandPalette')}</p>
                </TooltipContent>
              </Tooltip>

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
                    aria-label={totalNotifications > 0 ? `${t('topbar.notifications')} (${totalNotifications} unread)` : t('topbar.notifications')}
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
                    <p className="font-medium">{t('topbar.notifications')}</p>
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
                      <p className="text-xs text-aviation-text-muted">{t('common.noResults')}</p>
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
                    <p>{t('topbar.currentPlan')}: {plan}</p>
                    <p className="text-xs text-aviation-text-muted">{t('topbar.clickUpgrade')}</p>
                  </TooltipContent>
                </Tooltip>
              )}

              {/* Enterprise Badge */}
              <EnterpriseBadge />

              {/* Language Picker */}
              <LanguagePicker variant="icon" showLabel={false} />

              {/* User Menu */}
              <UserMenu />
            </div>
          </div>
        </header>

        {/* Command Palette */}
        <GlobalCommandPalette
          open={showCommandPalette}
          onOpenChange={setShowCommandPalette}
        />
      </>
    </TooltipProvider>
  );
}

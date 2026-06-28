import { Menu, Bell, Command, Shield } from "lucide-react";
import { cn } from "@/lib/utils";
import { Breadcrumb } from "./Breadcrumb";
import { UserMenu } from "./UserMenu";
import { GlobalCommandPalette } from "./GlobalCommandPalette";
import { EnterpriseBadge } from "@/components/enterprise";
import { LanguagePicker } from "@/components/common/LanguagePicker";
import { useNavigationStatus } from "@/hooks/useNavigationStatus";
import { useContextualActions } from "@/hooks/useContextualActions";
import { usePlan } from "@/hooks/usePlan";
import { useTranslation } from "react-i18next";
import { useState, useCallback, useEffect } from "react";
import "@/styles/sc-navbar.css";

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

  const totalNotifications = status.functions.pendingDeployments +
    (status.functions.hasIssues ? 1 : 0) +
    (status.providers.hasOffline ? 1 : 0) +
    (status.analytics.hasAlerts ? 1 : 0) +
    (status.settings.hasWarnings ? 1 : 0);

  useEffect(() => {
    if (totalNotifications > 0) {
      setNotificationPulse(true);
      const timer = setTimeout(() => setNotificationPulse(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [totalNotifications]);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      setShowCommandPalette(true);
    }
  }, []);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  const getStatusColor = () => {
    if (status.providers.hasOffline || status.functions.hasIssues) return 'sc-navbar__status-dot--error';
    if (status.analytics.hasAlerts) return 'sc-navbar__status-dot--warning';
    if (status.functions.pendingDeployments > 0) return 'sc-navbar__status-dot--info';
    return null;
  };

  const statusColor = getStatusColor();

  return (
    <>
      <header className={cn("sc-navbar", className)}>
        <div className="sc-navbar__inner">
          {/* Left */}
          <div className="sc-navbar__left">
            <button
              className="sc-navbar__menu-btn"
              onClick={onMenuClick}
              aria-label="Open navigation menu"
            >
              <Menu className="sc-navbar__icon" />
            </button>
            <Breadcrumb />
          </div>

          {/* Center: Contextual Actions */}
          {contextualActions.length > 0 && (
            <div className="sc-navbar__center">
              {contextualActions.map((action) => {
                const Icon = action.icon;
                return (
                  <button
                    key={action.id}
                    className={cn(
                      "sc-navbar__action-btn",
                      action.variant === "default" && "sc-navbar__action-btn--primary"
                    )}
                    onClick={action.onClick}
                  >
                    <Icon className="sc-navbar__action-icon" />
                    {action.label}
                  </button>
                );
              })}
            </div>
          )}

          {/* Right */}
          <div className="sc-navbar__right">
            {/* Command Palette Trigger */}
            <button
              onClick={() => setShowCommandPalette(true)}
              className="sc-navbar__cmd-btn"
            >
              <Command className="sc-navbar__cmd-icon" />
              <span className="sc-navbar__cmd-label">{t('topbar.search')}</span>
              <kbd className="sc-navbar__kbd">K</kbd>
            </button>

            {/* Notifications */}
            <button
              className={cn(
                "sc-navbar__icon-btn",
                notificationPulse && "sc-navbar__icon-btn--pulse"
              )}
              aria-label={totalNotifications > 0 ? `${t('topbar.notifications')} (${totalNotifications} unread)` : t('topbar.notifications')}
            >
              <Bell className="sc-navbar__icon" />
              {totalNotifications > 0 && (
                <>
                  {statusColor && <span className={cn("sc-navbar__status-dot", statusColor)} />}
                  <span className="sc-navbar__badge">
                    {totalNotifications > 9 ? '9+' : totalNotifications}
                  </span>
                </>
              )}
            </button>

            {/* Plan Badge */}
            {!isEnterprise && plan && (
              <div className="sc-navbar__plan-badge">
                <Shield className="sc-navbar__plan-icon" />
                <span className="capitalize">{plan}</span>
              </div>
            )}

            <EnterpriseBadge />
            <LanguagePicker variant="icon" showLabel={false} />
            <UserMenu />
          </div>
        </div>
      </header>

      <GlobalCommandPalette
        open={showCommandPalette}
        onOpenChange={setShowCommandPalette}
      />
    </>
  );
}

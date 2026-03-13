import { Menu, Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Breadcrumb } from "./Breadcrumb";
import { UserMenu } from "./UserMenu";
import { SearchButton } from "./SearchButton";
import { EnterpriseBadge } from "@/components/enterprise";
import { useNavigationStatus } from "@/hooks/useNavigationStatus";
import { useContextualActions } from "@/hooks/useContextualActions";
import { useThemeStore } from "@/stores/themeStore";

interface TopBarProps {
  onMenuClick: () => void;
  className?: string;
}

export function TopBar({ onMenuClick, className }: TopBarProps) {
  const status = useNavigationStatus();
  const contextualActions = useContextualActions();

  const totalNotifications = status.functions.pendingDeployments +
    (status.functions.hasIssues ? 1 : 0) +
    (status.providers.hasOffline ? 1 : 0) +
    (status.analytics.hasAlerts ? 1 : 0) +
    (status.settings.hasWarnings ? 1 : 0);

  return (
    <header
      className={cn(
        "h-16 bg-bg-primary/80 backdrop-blur-xl border-b border-border-subtle",
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
            className="lg:hidden text-text-secondary hover:text-text-primary"
            onClick={onMenuClick}
            aria-label="Open navigation menu"
          >
            <Menu className="w-5 h-5" />
          </Button>

          {/* Dynamic Breadcrumb */}
          <Breadcrumb />
        </div>

        {/* Contextual Actions */}
        {contextualActions.length > 0 && (
          <div className="flex items-center gap-2">
            {contextualActions.map((action) => {
              const Icon = action.icon;
              return (
                <Button
                  key={action.id}
                  variant={action.variant || "default"}
                  size="sm"
                  className="hidden md:flex"
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
        <div className="flex items-center gap-2">
          {/* Global Search */}
          <SearchButton />

          {/* Notifications */}
          <Button
            variant="ghost"
            size="icon"
            className="relative text-text-secondary hover:text-text-primary hover:bg-bg-hover"
            aria-label={totalNotifications > 0 ? `Notifications (${totalNotifications} unread)` : 'Notifications'}
          >
            <Bell className="w-5 h-5" />
            {totalNotifications > 0 && (
              <span className="absolute -top-1 -right-1 flex items-center justify-center min-w-[18px] h-[18px] bg-error text-white text-xs font-bold rounded-full px-1">
                {totalNotifications > 9 ? '9+' : totalNotifications}
              </span>
            )}
          </Button>

          {/* Enterprise Badge */}
          <EnterpriseBadge />

          {/* User Menu */}
          <UserMenu />
        </div>
      </div>
    </header>
  );
}

import { Switch } from '@/components/ui/switch';
import { PLANS } from '@/lib/constants';
import { isDesktopApp } from '@/lib/desktop-app';
import {
  openStudioExternalUrl,
  useStudioAccountPreferences,
} from '@/lib/studio-account-preferences';
import { cn } from '@/lib/utils';
import { getSettingsUrl, type SettingsTabValue } from '@/pages/SettingsPage/settings-utils';
import { useAuthStore } from '@/stores/authStore';
import { Badge, Button, GlassCard } from '@functionfly/ui-core';
import {
  ArrowUpRight,
  CreditCard,
  ExternalLink,
  Laptop,
  LogOut,
  Monitor,
  RotateCcw,
  Shield,
  User,
} from 'lucide-react';
import type { ReactNode } from 'react';
import { toast } from 'sonner';

function getInitials(name?: string | null, email?: string | null): string {
  const source = name || email || '';
  return (
    source
      .split(/[@.\s_-]+/)
      .filter(Boolean)
      .map((word) => word.charAt(0))
      .join('')
      .toUpperCase()
      .slice(0, 2) || '??'
  );
}

function openWebSettings(username: string, tab?: SettingsTabValue) {
  const path = getSettingsUrl(username, tab);
  const url = `${window.location.origin}${path}`;
  if (isDesktopApp()) {
    openStudioExternalUrl(url);
    return;
  }
  window.location.assign(path);
}

export function StudioAccountSettingsTab() {
  const { user, logout } = useAuthStore();
  const { preferences, updatePreference, resetPreferences, isLoading, isSaving, isResetting } =
    useStudioAccountPreferences();
  const isDesktop = isDesktopApp();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500" />
      </div>
    );
  }

  if (!user) {
    return (
      <GlassCard className="p-6">
        <p className="text-sm text-text-muted">Sign in to manage your Studio account.</p>
      </GlassCard>
    );
  }

  const planInfo = PLANS[user.plan.toUpperCase() as keyof typeof PLANS] || PLANS.FREE;
  const username = user.username || user.email.split('@')[0];

  const handleSignOut = async () => {
    try {
      await logout();
      toast.success('Signed out of Studio');
    } catch {
      toast.error('Failed to sign out');
    }
  };

  const webAccountLinks: Array<{
    tab?: SettingsTabValue;
    label: string;
    description: string;
    icon: ReactNode;
  }> = [
    {
      tab: 'account',
      label: 'Profile & password',
      description: 'Name, username, email, and password',
      icon: <User className="w-4 h-4" />,
    },
    {
      tab: 'security',
      label: 'Security & sessions',
      description: 'MFA, active sessions, and API access',
      icon: <Shield className="w-4 h-4" />,
    },
    {
      tab: 'billing',
      label: 'Billing & plan',
      description: 'Subscription, invoices, and payment methods',
      icon: <CreditCard className="w-4 h-4" />,
    },
  ];

  const desktopPreferences: Array<{
    key: keyof typeof preferences;
    label: string;
    description: string;
    icon: ReactNode;
    desktopOnly?: boolean;
  }> = [
    {
      key: 'launchAtLogin',
      label: 'Launch at login',
      description: 'Open Studio automatically when you sign in to your computer',
      icon: <Monitor className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />,
      desktopOnly: true,
    },
    {
      key: 'minimizeToTrayOnClose',
      label: 'Minimize to tray on close',
      description: 'Keep Studio running in the system tray when the window is closed',
      icon: <Laptop className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />,
      desktopOnly: true,
    },
    {
      key: 'restoreLastWorkspace',
      label: 'Restore last workspace',
      description: 'Reopen your previous Studio workspace when the app starts',
      icon: <RotateCcw className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />,
    },
    {
      key: 'openLinksExternally',
      label: 'Open links in browser',
      description: 'Open dashboard and docs links in your default browser instead of Studio',
      icon: <ExternalLink className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />,
      desktopOnly: true,
    },
  ];

  return (
    <div className="space-y-6 max-w-2xl">
      <GlassCard className="p-5 border border-border-subtle">
        <div className="flex items-start gap-4">
          <div
            className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl text-lg font-semibold text-white"
            style={{
              background:
                'linear-gradient(135deg, var(--button-primary), var(--button-primary-hover))',
            }}
          >
            {getInitials(user.name || user.username, user.email)}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-lg font-semibold text-text-primary truncate">
                {user.name || user.username || 'Studio user'}
              </h3>
              <Badge variant="outline" className="text-text-muted border-border-subtle capitalize">
                {planInfo.name} plan
              </Badge>
            </div>
            <p className="text-sm text-text-muted truncate">{user.email}</p>
            {user.username && <p className="text-xs text-text-muted mt-1">@{user.username}</p>}
          </div>
        </div>
      </GlassCard>

      <div>
        <h3 className="text-lg font-semibold text-text-primary mb-1">Studio preferences</h3>
        <p className="text-sm text-text-muted mb-4">
          Preferences synced to your FunctionFly account.
          {isDesktop && ' Desktop-only options apply when running the native Studio app.'}
        </p>
        <div className="space-y-3">
          {desktopPreferences.map((pref) => {
            if (pref.desktopOnly && !isDesktop) return null;
            return (
              <div
                key={pref.key}
                className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle"
              >
                <div className="flex items-center gap-3 min-w-0 pr-4">
                  {pref.icon}
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-text-primary">{pref.label}</p>
                    <p className="text-xs text-text-muted">{pref.description}</p>
                  </div>
                </div>
                <Switch
                  checked={preferences[pref.key]}
                  disabled={isSaving}
                  onCheckedChange={(value) => {
                    updatePreference(pref.key, value).catch(() => {
                      toast.error('Failed to save preference');
                    });
                  }}
                />
              </div>
            );
          })}
        </div>
        {!isDesktop && (
          <p className="text-xs text-text-muted mt-3">
            Desktop-only options appear when you run FunctionFly Studio as a native app.
          </p>
        )}
        <div className="mt-4">
          <Button
            variant="outline"
            size="sm"
            className="gap-2 border-border-subtle text-text-secondary hover:bg-bg-hover"
            disabled={isResetting || isSaving}
            onClick={() => {
              resetPreferences()
                .then(() => toast.success('Studio preferences reset'))
                .catch(() => toast.error('Failed to reset preferences'));
            }}
          >
            <RotateCcw className="w-4 h-4" />
            Reset Studio preferences
          </Button>
        </div>
      </div>

      <div>
        <h3 className="text-lg font-semibold text-text-primary mb-1">Web account</h3>
        <p className="text-sm text-text-muted mb-4">
          Profile, security, and billing are managed in the FunctionFly dashboard.
          {isDesktop && ' Links open in your browser so you stay in Studio.'}
        </p>
        <div className="space-y-2">
          {webAccountLinks.map((link) => (
            <button
              key={link.label}
              type="button"
              onClick={() => openWebSettings(username, link.tab)}
              className={cn(
                'w-full flex items-center justify-between gap-3 p-4 rounded-lg',
                'bg-bg-secondary border border-border-subtle text-left',
                'hover:bg-bg-hover transition-colors duration-200'
              )}
            >
              <div className="flex items-center gap-3 min-w-0">
                <span className="text-text-accent">{link.icon}</span>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-text-primary">{link.label}</p>
                  <p className="text-xs text-text-muted">{link.description}</p>
                </div>
              </div>
              <ArrowUpRight className="w-4 h-4 shrink-0 text-text-muted" />
            </button>
          ))}
          <button
            type="button"
            onClick={() => openWebSettings(username)}
            className={cn(
              'w-full flex items-center justify-between gap-3 p-4 rounded-lg',
              'bg-bg-secondary border border-border-subtle text-left',
              'hover:bg-bg-hover transition-colors duration-200'
            )}
          >
            <div className="flex items-center gap-3 min-w-0">
              <span className="text-text-accent">
                <ExternalLink className="w-4 h-4" />
              </span>
              <div className="min-w-0">
                <p className="text-sm font-medium text-text-primary">Open full dashboard</p>
                <p className="text-xs text-text-muted">
                  Billing, teams, marketplace, and all web settings
                </p>
              </div>
            </div>
            <ArrowUpRight className="w-4 h-4 shrink-0 text-text-muted" />
          </button>
        </div>
      </div>

      <GlassCard className="p-5 border border-border-subtle">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h3 className="text-sm font-semibold text-text-primary">Sign out of Studio</h3>
            <p className="text-xs text-text-muted mt-1">
              End your session on this device. Your Studio preferences remain saved to your account.
            </p>
          </div>
          <Button
            variant="outline"
            className="gap-2 border-red-500/30 text-red-400 hover:bg-red-500/10 hover:text-red-300 shrink-0"
            onClick={handleSignOut}
          >
            <LogOut className="w-4 h-4" />
            Sign out
          </Button>
        </div>
      </GlassCard>
    </div>
  );
}

import { usersApi } from '@/api/users';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

const NOTIFICATION_OPTIONS = [
  {
    key: 'deploymentSuccess' as const,
    label: 'Deployment Success',
    description: 'Get notified when a deployment succeeds',
  },
  {
    key: 'deploymentFailure' as const,
    label: 'Deployment Failure',
    description: 'Get notified when a deployment fails',
  },
  {
    key: 'failoverEvents' as const,
    label: 'Failover Events',
    description: 'Get notified when failover is triggered',
  },
  {
    key: 'providerIssues' as const,
    label: 'Provider Issues',
    description: 'Get notified when a provider has issues',
  },
] as const;

const DEFAULT_PREFS = {
  deploymentSuccess: true,
  deploymentFailure: true,
  failoverEvents: true,
  providerIssues: true,
};

type PrefsKey = keyof typeof DEFAULT_PREFS;

function prefsFromSettings(settings: Record<string, unknown> | undefined): typeof DEFAULT_PREFS {
  if (!settings) return DEFAULT_PREFS;
  return {
    deploymentSuccess: settings.deploymentSuccess !== false,
    deploymentFailure: settings.deploymentFailure !== false,
    failoverEvents: settings.failoverEvents !== false,
    providerIssues: settings.providerIssues !== false,
  };
}

export function NotificationsSettingsTab() {
  const [notifications, setNotifications] = useState<typeof DEFAULT_PREFS>(DEFAULT_PREFS);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    usersApi
      .getMySettings()
      .then((res) => {
        if (cancelled) return;
        const settings = (res as { settings?: Record<string, unknown> }).settings;
        setNotifications(prefsFromSettings(settings));
      })
      .catch(() => {
        if (!cancelled) toast.error('Failed to load notification preferences');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleToggle = async (key: PrefsKey, checked: boolean) => {
    const previous = { ...notifications };
    const updated = { ...notifications, [key]: checked };
    setNotifications(updated);
    try {
      await usersApi.updateMyNotificationSettings(updated);
      toast.success('Notification preference saved');
    } catch {
      setNotifications(previous);
      toast.error('Failed to save notification preference');
    }
  };

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">Notification Preferences</CardTitle>
          <CardDescription className="text-text-secondary">
            Choose what notifications you want to receive
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading && <p className="text-sm text-text-muted mb-4">Loading preferences…</p>}
          <div className="space-y-4">
            {NOTIFICATION_OPTIONS.map((item) => (
              <div key={item.key} className="flex items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <h4 className="font-medium text-text-primary">{item.label}</h4>
                  <p className="text-sm text-text-muted">{item.description}</p>
                </div>
                <div className="shrink-0">
                  <Switch
                    checked={notifications[item.key]}
                    onCheckedChange={(checked) => handleToggle(item.key, checked)}
                  />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

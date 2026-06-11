import { usersApi } from '@/api/users';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

const NOTIFICATION_KEYS = [
  'deploymentSuccess',
  'deploymentFailure',
  'failoverEvents',
  'providerIssues',
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
  const { t } = useTranslation();
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
        if (!cancelled) toast.error(t('notifSettings.toastFailedToLoad'));
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
      toast.success(t('notifSettings.toastSaved'));
    } catch {
      setNotifications(previous);
      toast.error(t('notifSettings.toastFailedToSave'));
    }
  };

  return (
    <div className="settings-page space-y-6">
      <Card className="settings-panel">
        <CardHeader>
          <CardTitle className="settings-section-title">{t('notifSettings.title')}</CardTitle>
          <CardDescription className="settings-section-description">
            {t('notifSettings.description')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading && <p className="text-sm text-text-muted mb-4">{t('notifSettings.loadingPreferences')}</p>}
          <div className="space-y-4">
            {NOTIFICATION_KEYS.map((key) => (
              <div key={key} className="flex items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <h4 className="font-medium text-text-primary">{t(`notifSettings.${key}Label`)}</h4>
                  <p className="text-sm text-text-muted">{t(`notifSettings.${key}Description`)}</p>
                </div>
                <div className="shrink-0">
                  <Switch
                    checked={notifications[key]}
                    onCheckedChange={(checked) => handleToggle(key, checked)}
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

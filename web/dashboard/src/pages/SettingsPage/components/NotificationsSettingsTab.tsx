import { usersApi } from '@/api/users';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { CollapsibleSection } from '@/components/ui/collapsible-section';
import { Switch } from '@/components/ui/switch';
import {
  Rocket,
  Shield,
  Server,
  CreditCard,
  Code,
  Users,
  DollarSign,
  Brain,
  Bell,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

const NOTIFICATION_PREFS = {
  deploymentSuccess: true,
  deploymentFailure: true,
  deploymentStarted: true,
  failoverEvents: true,
  providerIssues: true,
  paymentFailures: true,
  lowWalletBalance: true,
  spendCapWarnings: true,
  invoiceGenerated: true,
  subscriptionExpiring: true,
  newDeviceLogin: true,
  suspiciousActivity: true,
  functionErrors: true,
  functionPublished: true,
  functionUpdated: true,
  teamInvitations: true,
  roleChanges: true,
  directMessages: true,
  payoutCompleted: true,
  payoutFailed: true,
  payoutApprovalNeeded: true,
  consciousnessCritical: true,
  consciousnessAutoApplied: true,
} as const;

type PrefsKey = keyof typeof NOTIFICATION_PREFS;

interface NotificationToggle {
  key: PrefsKey;
  defaultOn?: boolean;
}

interface NotificationGroup {
  sectionKey: string;
  icon: React.ReactNode;
  defaultOpen?: boolean;
  toggles: NotificationToggle[];
}

const NOTIFICATION_GROUPS: NotificationGroup[] = [
  {
    sectionKey: 'deployment',
    icon: <Rocket className="w-4 h-4" />,
    defaultOpen: true,
    toggles: [
      { key: 'deploymentSuccess' },
      { key: 'deploymentFailure' },
      { key: 'deploymentStarted', defaultOn: false },
    ],
  },
  {
    sectionKey: 'infrastructure',
    icon: <Server className="w-4 h-4" />,
    defaultOpen: true,
    toggles: [
      { key: 'failoverEvents' },
      { key: 'providerIssues' },
    ],
  },
  {
    sectionKey: 'billing',
    icon: <CreditCard className="w-4 h-4" />,
    toggles: [
      { key: 'paymentFailures' },
      { key: 'lowWalletBalance' },
      { key: 'spendCapWarnings' },
      { key: 'invoiceGenerated', defaultOn: false },
      { key: 'subscriptionExpiring' },
    ],
  },
  {
    sectionKey: 'security',
    icon: <Shield className="w-4 h-4" />,
    toggles: [
      { key: 'newDeviceLogin' },
      { key: 'suspiciousActivity' },
    ],
  },
  {
    sectionKey: 'functions',
    icon: <Code className="w-4 h-4" />,
    toggles: [
      { key: 'functionErrors' },
      { key: 'functionPublished', defaultOn: false },
      { key: 'functionUpdated', defaultOn: false },
    ],
  },
  {
    sectionKey: 'team',
    icon: <Users className="w-4 h-4" />,
    toggles: [
      { key: 'teamInvitations' },
      { key: 'roleChanges' },
      { key: 'directMessages' },
    ],
  },
  {
    sectionKey: 'payouts',
    icon: <DollarSign className="w-4 h-4" />,
    toggles: [
      { key: 'payoutCompleted', defaultOn: false },
      { key: 'payoutFailed' },
      { key: 'payoutApprovalNeeded' },
    ],
  },
  {
    sectionKey: 'consciousness',
    icon: <Brain className="w-4 h-4" />,
    toggles: [
      { key: 'consciousnessCritical' },
      { key: 'consciousnessAutoApplied', defaultOn: false },
    ],
  },
];

function getDefaultPrefs(): Record<PrefsKey, boolean> {
  const defaults = {} as Record<PrefsKey, boolean>;
  for (const group of NOTIFICATION_GROUPS) {
    for (const toggle of group.toggles) {
      defaults[toggle.key] = toggle.defaultOn ?? true;
    }
  }
  return defaults;
}

function prefsFromSettings(settings: Record<string, unknown> | undefined): Record<PrefsKey, boolean> {
  const defaults = getDefaultPrefs();
  if (!settings) return defaults;
  const result = {} as Record<PrefsKey, boolean>;
  for (const key of Object.keys(defaults) as PrefsKey[]) {
    result[key] = settings[key] !== false;
  }
  return result;
}

export function NotificationsSettingsTab() {
  const { t } = useTranslation();
  const [notifications, setNotifications] = useState<Record<PrefsKey, boolean>>(getDefaultPrefs);
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
          <CardTitle className="flex items-center gap-2">
            <Bell className="w-5 h-5" />
            {t('notifSettings.title')}
          </CardTitle>
          <CardDescription className="settings-section-description">
            {t('notifSettings.description')}
            {' '}
            <span className="text-text-muted">
              {t('notifSettings.forSocialNotifications')}
            </span>
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading && <p className="text-sm text-text-muted mb-4">{t('notifSettings.loadingPreferences')}</p>}
          <div className="space-y-3">
            {NOTIFICATION_GROUPS.map((group) => (
              <CollapsibleSection
                key={group.sectionKey}
                title={t(`notifSettings.sections.${group.sectionKey}.title`)}
                icon={group.icon}
                defaultOpen={group.defaultOpen}
              >
                <div className="space-y-4">
                  {group.toggles.map((toggle) => (
                    <div key={toggle.key} className="flex items-center justify-between gap-4">
                      <div className="min-w-0 flex-1">
                        <h4 className="font-medium text-text-primary">
                          {t(`notifSettings.${toggle.key}Label`)}
                        </h4>
                        <p className="text-sm text-text-muted">
                          {t(`notifSettings.${toggle.key}Description`)}
                        </p>
                      </div>
                      <div className="shrink-0">
                        <Switch
                          checked={notifications[toggle.key]}
                          onCheckedChange={(checked) => handleToggle(toggle.key, checked)}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </CollapsibleSection>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

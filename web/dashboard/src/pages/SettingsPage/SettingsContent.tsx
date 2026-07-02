/**
 * Shared settings content: Account, Billing, Developer, Notifications, Security, Privacy.
 * Uses Sealed Containment design system with sidebar navigation.
 *
 * URL STRUCTURE (Hash-based routing - long term):
 *   /u/:username/settings#account       → Account tab (default)
 *   /u/:username/settings#billing       → Billing tab
 *   /u/:username/settings#developer     → Developer tab
 *   /u/:username/settings#notifications → Notifications tab
 *   /u/:username/settings#security      → Security tab
 *   /u/:username/settings#privacy       → Privacy tab
 *   /u/:username/settings#integrations  → Integrations tab
 */

import '@/styles/aviation-dashboard.css';
import './styles.css';

import { usersApi } from '@/api/users';
import { usePageTitle } from '@/hooks';
import { GitHubSettingsPage } from '@/pages/GitHubSettingsPage';
import { useApiReachableStore } from '@/stores/apiReachableStore';
import { useAuthStore } from '@/stores/authStore';
import type { UserProfile } from '@/types';
import { useQuery } from '@tanstack/react-query';
import {
  Bell,
  Brain,
  Code,
  CreditCard,
  Dna,
  Link2,
  Settings,
  Shield,
  ShieldCheck,
  Terminal,
  User,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import {
  AccountSettingsTab,
  AIKeysSettingsTab,
  AuthSettingsTab,
  BillingSettingsTab,
  DeveloperSettingsTab,
  IntegrationsSettingsTab,
  NotificationsSettingsTab,
  PlatformSettingsTab,
  SecuritySettingsTab,
  TrustAPISettingsTab,
} from './components';
import { PrivacySettingsTab } from './components/PrivacySettingsTab';
import { VALID_TABS, type SettingsTabValue } from './settings-utils';

export interface SettingsContentProps {
  showHeader?: boolean;
  profile?: UserProfile | null;
  initialTab?: string;
}

function getInitialTab(
  hash: string,
  initialTabProp: string | undefined,
  subtabFromUrl: string | null
): SettingsTabValue {
  const hashTab = hash.replace('#', '');
  if (hashTab && VALID_TABS.includes(hashTab as SettingsTabValue)) {
    return hashTab as SettingsTabValue;
  }
  if (initialTabProp && VALID_TABS.includes(initialTabProp as SettingsTabValue)) {
    return initialTabProp as SettingsTabValue;
  }
  if (subtabFromUrl && VALID_TABS.includes(subtabFromUrl as SettingsTabValue)) {
    return subtabFromUrl as SettingsTabValue;
  }
  return 'account';
}

type TabDef = { id: SettingsTabValue; label: string; icon: typeof User };

const TAB_GROUPS: { label: string; icon: typeof Settings; tabs: TabDef[] }[] = [
  {
    label: 'General',
    icon: Settings,
    tabs: [
      { id: 'account', label: 'Account', icon: User },
      { id: 'billing', label: 'Billing', icon: CreditCard },
      { id: 'notifications', label: 'Notifications', icon: Bell },
      { id: 'privacy', label: 'Privacy', icon: Shield },
    ],
  },
  {
    label: 'Security & Access',
    icon: ShieldCheck,
    tabs: [
      { id: 'security', label: 'Security', icon: ShieldCheck },
      { id: 'github', label: 'GitHub', icon: Code },
    ],
  },
  {
    label: 'Developer',
    icon: Terminal,
    tabs: [
      { id: 'developer', label: 'API & Webhooks', icon: Code },
      { id: 'ai-keys', label: 'AI Keys', icon: Brain },
      { id: 'trust-api', label: 'Trust API', icon: Shield },
    ],
  },
  {
    label: 'Platform',
    icon: Dna,
    tabs: [
      { id: 'platform', label: 'Platform', icon: Dna },
      { id: 'integrations', label: 'Integrations', icon: Link2 },
    ],
  },
];

function getTabLabel(tabId: SettingsTabValue, fallback: string, t: (k: string) => string): string {
  const map: Partial<Record<SettingsTabValue, string>> = {
    account: t('settings.account'),
    billing: t('settings.billing'),
    developer: t('settings.developer'),
    notifications: t('settings.notifications'),
    security: t('settings.security'),
    privacy: t('settings.privacy'),
  };
  return map[tabId] ?? fallback;
}

export function SettingsContent({
  showHeader = true,
  profile,
  initialTab: initialTabProp,
}: SettingsContentProps) {
  usePageTitle('Settings');
  const { t } = useTranslation();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const subtabFromUrl = searchParams.get('subtab');
  const [activeTab, setActiveTab] = useState<SettingsTabValue>(() =>
    getInitialTab(location.hash, initialTabProp, subtabFromUrl)
  );

  const apiReachable = useApiReachableStore((s) => s.apiReachable);
  const user = useAuthStore((s) => s.user);
  const setUserPlan = useAuthStore((s) => s.setUserPlan);
  const { data: meData } = useQuery({
    queryKey: ['users', 'me'],
    queryFn: async () => {
      try {
        return await usersApi.getMe();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return null;
        throw e;
      }
    },
    enabled: apiReachable === true,
    retry: false,
  });
  useEffect(() => {
    if (meData?.plan !== undefined && meData.plan !== useAuthStore.getState().user?.plan) {
      setUserPlan(meData.plan);
    }
  }, [meData?.plan, setUserPlan]);
  const displayPlan = meData?.plan ?? user?.plan ?? 'Starter';

  useEffect(() => {
    const newHash = `#${activeTab}`;
    if (location.hash !== newHash) {
      window.history.replaceState(
        {},
        document.title,
        `${location.pathname}${newHash}${location.search}`
      );
    }
  }, [activeTab, location.pathname, location.search, location.hash]);

  useEffect(() => {
    const handleHashChange = () => {
      const hashTab = location.hash.replace('#', '');
      if (hashTab && VALID_TABS.includes(hashTab as SettingsTabValue)) {
        setActiveTab(hashTab as SettingsTabValue);
      }
    };

    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, []);

  useEffect(() => {
    const next = initialTabProp || subtabFromUrl;
    if (next && VALID_TABS.includes(next as SettingsTabValue)) {
      setActiveTab(next as SettingsTabValue);
      window.history.replaceState(
        {},
        document.title,
        `${location.pathname}#${next}${location.search}`
      );
    }
  }, [initialTabProp, subtabFromUrl, location.pathname, location.search]);

  useEffect(() => {
    const success = searchParams.get('success');
    if (success === 'true') {
      toast.success('Subscription updated successfully!');
      const next = new URLSearchParams(location.search);
      next.delete('success');
      const q = next.toString();
      window.history.replaceState({}, document.title, location.pathname + (q ? `?${q}` : ''));
    }
  }, [searchParams, location.pathname, location.search]);

  const returnUrl = `${window.location.origin}${location.pathname}${location.search ? location.search : ''}`;

  return (
    <div className="settings-page">
      {showHeader && (
        <div className="settings-page-header">
          <h1 className="settings-page-title">{t('settings.title')}</h1>
          <p className="settings-page-subtitle">{t('settings.manageAccount')}</p>
        </div>
      )}

      <div className="settings-layout">
        {/* Sidebar Navigation */}
        <nav className="settings-sidebar" role="tablist" aria-orientation="vertical">
          {TAB_GROUPS.map((group, groupIdx) => (
            <div className="settings-sidebar-group" key={group.label}>
              {groupIdx > 0 && <div className="settings-sidebar-divider" />}
              <div className="settings-sidebar-group-label">
                <group.icon style={{ width: 12, height: 12 }} />
                {group.label}
              </div>
              {group.tabs.map((tab) => {
                const Icon = tab.icon;
                const isActive = activeTab === tab.id;
                return (
                  <button
                    key={tab.id}
                    role="tab"
                    aria-selected={isActive}
                    className="settings-sidebar-item"
                    data-state={isActive ? 'active' : 'inactive'}
                    onClick={() => setActiveTab(tab.id)}
                  >
                    <Icon className="settings-sidebar-item-icon" />
                    <span>{getTabLabel(tab.id, tab.label, t)}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </nav>

        {/* Tab Content */}
        <div className="settings-tab-content">
          {activeTab === 'account' && <AccountSettingsTab />}
          {activeTab === 'billing' && <BillingSettingsTab returnUrl={returnUrl} displayPlan={displayPlan} />}
          {activeTab === 'developer' && (
            <>
              <DeveloperSettingsTab />
              <AuthSettingsTab />
            </>
          )}
          {activeTab === 'notifications' && <NotificationsSettingsTab />}
          {activeTab === 'security' && <SecuritySettingsTab />}
          {activeTab === 'privacy' && <PrivacySettingsTab profile={profile ?? undefined} />}
          {activeTab === 'platform' && <PlatformSettingsTab />}
          {activeTab === 'ai-keys' && <AIKeysSettingsTab />}
          {activeTab === 'integrations' && <IntegrationsSettingsTab />}
          {activeTab === 'github' && <GitHubSettingsPage />}
          {activeTab === 'trust-api' && <TrustAPISettingsTab returnUrl={returnUrl} userPlan={displayPlan} />}
        </div>
      </div>
    </div>
  );
}

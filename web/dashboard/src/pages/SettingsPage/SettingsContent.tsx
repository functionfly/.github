/**
 * Shared settings content: Account, Billing, Developer, Notifications, Security, Privacy.
 * Uses Sealed Containment design system.
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
import { Bell, Code, CreditCard, Dna, Link2, Shield, ShieldCheck, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import {
  AccountSettingsTab,
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

const TABS: { id: SettingsTabValue; label: string; icon: typeof User }[] = [
  { id: 'account', label: 'Account', icon: User },
  { id: 'billing', label: 'Billing', icon: CreditCard },
  { id: 'developer', label: 'Developer', icon: Code },
  { id: 'notifications', label: 'Notifications', icon: Bell },
  { id: 'security', label: 'Security', icon: ShieldCheck },
  { id: 'privacy', label: 'Privacy', icon: Shield },
  { id: 'platform', label: 'Platform', icon: Dna },
  { id: 'integrations', label: 'Integrations', icon: Link2 },
  { id: 'github', label: 'GitHub', icon: Code },
  { id: 'trust-api', label: 'Trust API', icon: Shield },
];

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

      {/* Tab Navigation — containment-style tabs */}
      <div className="settings-page-tabs" role="tablist">
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          const tabLabel = tab.id === 'account' ? t('settings.account')
            : tab.id === 'billing' ? t('settings.billing')
            : tab.id === 'developer' ? t('settings.developer')
            : tab.id === 'notifications' ? t('settings.notifications')
            : tab.id === 'security' ? t('settings.security')
            : tab.id === 'privacy' ? t('settings.privacy')
            : tab.label;
          return (
            <button
              key={tab.id}
              role="tab"
              aria-selected={isActive}
              className="settings-page-tab"
              data-state={isActive ? 'active' : 'inactive'}
              onClick={() => setActiveTab(tab.id)}
            >
              <Icon style={{ width: 14, height: 14 }} />
              <span>{tabLabel}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      <div className="settings-tab-content" style={{ paddingTop: 'var(--space-6)' }}>
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
        {activeTab === 'integrations' && <IntegrationsSettingsTab />}
        {activeTab === 'github' && <GitHubSettingsPage />}
        {activeTab === 'trust-api' && <TrustAPISettingsTab returnUrl={returnUrl} />}
      </div>
    </div>
  );
}

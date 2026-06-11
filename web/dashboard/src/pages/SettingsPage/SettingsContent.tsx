/**
 * Shared settings content: Account, Billing, Developer, Notifications, Security, Privacy.
 * Used on the standalone /settings page and on /u/{username} (profile Settings tab).
 *
 * URL STRUCTURE (Hash-based routing - long term):
 *   /u/:username/settings#account       → Account tab (default)
 *   /u/:username/settings#billing       → Billing tab
 *   /u/:username/settings#developer     → Developer tab (API Keys, Deploy Keys, Webhooks)
 *   /u/:username/settings#notifications → Notifications tab
 *   /u/:username/settings#security      → Security tab
 *   /u/:username/settings#privacy       → Privacy tab
 *   /u/:username/settings#integrations  → Integrations tab (Brain connectors)
 *
 * BACKWARDS COMPATIBILITY:
 *   /u/:username/settings/billing      → Redirects to #billing (path-based, deprecated)
 *   ?subtab=billing                    → Redirects to #billing (query param, deprecated)
 *
 * Use getSettingsUrl(username, tab) from settings-utils.ts for generating URLs.
 */

import './styles.css';
import '@/styles/aviation-dashboard.css';

import { usersApi } from '@/api/users';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PrivacySettingsTab } from './components/PrivacySettingsTab';
import { GitHubSettingsPage } from '@/pages/GitHubSettingsPage';
import { useApiReachableStore } from '@/stores/apiReachableStore';
import { useAuthStore } from '@/stores/authStore';
import type { UserProfile } from '@/types';
import { useQuery } from '@tanstack/react-query';
import { Bell, CreditCard, Code, Dna, Key, Link2, Shield, ShieldCheck, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
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
import { VALID_TABS, type SettingsTabValue } from './settings-utils';

export interface SettingsContentProps {
  /** When false, omit the "Settings" page title (e.g. when embedded in profile). */
  showHeader?: boolean;
  /** Optional profile for Privacy tab (visibility, social links, etc.). */
  profile?: UserProfile | null;
  /** Initial tab when opened via path (e.g. "billing" for /u/username/settings/billing). */
  initialTab?: string;
}

/** Get tab from hash, path, or query param - priority: hash > path > query > default */
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

  // Update URL hash when tab changes (enables shareable links and browser history)
  useEffect(() => {
    const newHash = `#${activeTab}`;
    if (location.hash !== newHash) {
      // Use replaceState for initial load, pushState for user-initiated changes
      window.history.replaceState(
        {},
        document.title,
        `${location.pathname}${newHash}${location.search}`
      );
    }
  }, [activeTab, location.pathname, location.search, location.hash]);

  // Listen for hash changes (browser back/forward, external links)
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

  // Sync from path-based initialTab or query param on first load (backwards compat)
  useEffect(() => {
    const next = initialTabProp || subtabFromUrl;
    if (next && VALID_TABS.includes(next as SettingsTabValue)) {
      setActiveTab(next as SettingsTabValue);
      // Also update hash to match
      window.history.replaceState(
        {},
        document.title,
        `${location.pathname}#${next}${location.search}`
      );
    }
  }, [initialTabProp, subtabFromUrl, location.pathname, location.search]);

  // Billing portal return: show success toast and clean URL
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
    <div className="aviation-dashboard settings-page space-y-6">
      {showHeader && (
        <div className="settings-page-header">
          <h1 className="settings-page-title">
            {t('settings.title')}
          </h1>
          <p className="settings-page-subtitle">{t('settings.manageAccount')}</p>
        </div>
      )}

      <Tabs
        value={activeTab}
        onValueChange={(v) => setActiveTab(v as SettingsTabValue)}
        className="settings-tab-content space-y-6"
      >
        <TabsList className="settings-page-tabs inline-flex h-auto flex-wrap gap-1 rounded-xl border border-border-default bg-bg-secondary/80 p-1.5 text-text-secondary backdrop-blur-sm">
          <TabsTrigger
            value="account"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <User className="h-4 w-4 shrink-0" />
            {t('settings.account')}
          </TabsTrigger>
          <TabsTrigger
            value="billing"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <CreditCard className="h-4 w-4 shrink-0" />
            {t('settings.billing')}
          </TabsTrigger>
          <TabsTrigger
            value="developer"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Code className="h-4 w-4 shrink-0" />
            {t('settings.developer')}
          </TabsTrigger>
          <TabsTrigger
            value="notifications"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Bell className="h-4 w-4 shrink-0" />
            {t('settings.notifications')}
          </TabsTrigger>
          <TabsTrigger
            value="security"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <ShieldCheck className="h-4 w-4 shrink-0" />
            {t('settings.security')}
          </TabsTrigger>
          <TabsTrigger
            value="privacy"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Shield className="h-4 w-4 shrink-0" />
            {t('settings.privacy')}
          </TabsTrigger>
          <TabsTrigger
            value="platform"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Dna className="h-4 w-4 shrink-0" />
            Platform
          </TabsTrigger>
          <TabsTrigger
            value="integrations"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Link2 className="h-4 w-4 shrink-0" />
            Integrations
          </TabsTrigger>
          <TabsTrigger
            value="github"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <svg
              role="img"
              viewBox="0 0 24 24"
              className="h-4 w-4 shrink-0"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fill="currentColor"
                d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
              />
            </svg>
            GitHub
          </TabsTrigger>
          <TabsTrigger
            value="trust-api"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-brand-500 [&>svg]:data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm"
          >
            <Shield className="h-4 w-4 shrink-0" />
            Trust API
          </TabsTrigger>
        </TabsList>

        <TabsContent value="account" className="settings-tab-content space-y-6">
          <AccountSettingsTab />
        </TabsContent>

        <TabsContent value="billing" className="settings-tab-content space-y-6">
          <BillingSettingsTab returnUrl={returnUrl} displayPlan={displayPlan} />
        </TabsContent>

        <TabsContent value="developer" className="settings-tab-content space-y-6">
          <DeveloperSettingsTab />
          <AuthSettingsTab />
        </TabsContent>

        <TabsContent value="notifications" className="settings-tab-content space-y-6">
          <NotificationsSettingsTab />
        </TabsContent>

        <TabsContent value="security" className="settings-tab-content space-y-6">
          <SecuritySettingsTab />
        </TabsContent>

        <TabsContent value="privacy" className="settings-tab-content space-y-6">
          <PrivacySettingsTab profile={profile ?? undefined} />
        </TabsContent>

        <TabsContent value="platform" className="settings-tab-content space-y-6">
          <PlatformSettingsTab />
        </TabsContent>

        <TabsContent value="integrations" className="settings-tab-content space-y-6">
          <IntegrationsSettingsTab />
        </TabsContent>

        <TabsContent value="github" className="settings-tab-content space-y-6">
          <GitHubSettingsPage />
        </TabsContent>

        <TabsContent value="trust-api" className="settings-tab-content space-y-6">
          <TrustAPISettingsTab returnUrl={returnUrl} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

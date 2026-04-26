/**
 * Shared settings content: Account, Billing, Developer, Notifications, Security, Privacy.
 * Used on the standalone /settings page and on /u/{username} (profile Settings tab).
 *
 * URL STRUCTURE (Hash-based routing - long term):
 *   /u/:username/settings#account      → Account tab (default)
 *   /u/:username/settings#billing      → Billing tab
 *   /u/:username/settings#developer    → Developer tab (API Keys, Deploy Keys, Webhooks)
 *   /u/:username/settings#notifications → Notifications tab
 *   /u/:username/settings#security      → Security tab
 *   /u/:username/settings#privacy       → Privacy tab
 *
 * BACKWARDS COMPATIBILITY:
 *   /u/:username/settings/billing      → Redirects to #billing (path-based, deprecated)
 *   ?subtab=billing                    → Redirects to #billing (query param, deprecated)
 *
 * Use getSettingsUrl(username, tab) from settings-utils.ts for generating URLs.
 */

import { usersApi } from '@/api/users';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PrivacySettingsTab } from './components/PrivacySettingsTab';
import { useApiReachableStore } from '@/stores/apiReachableStore';
import { useAuthStore } from '@/stores/authStore';
import type { UserProfile } from '@/types';
import { useQuery } from '@tanstack/react-query';
import { Bell, CreditCard, Code, Key, Shield, ShieldCheck, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import {
  AccountSettingsTab,
  BillingSettingsTab,
  DeveloperSettingsTab,
  NotificationsSettingsTab,
  SecuritySettingsTab,
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
      window.history.replaceState({}, document.title, `${location.pathname}${newHash}${location.search}`);
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
    <div className="space-y-6">
      {showHeader && (
        <div>
          <h1 className="font-display text-2xl font-bold text-text-primary text-glow">
            {t('settings.title')}
          </h1>
          <p className="text-text-secondary">{t('settings.manageAccount')}</p>
        </div>
      )}

      <Tabs
        value={activeTab}
        onValueChange={(v) => setActiveTab(v as SettingsTabValue)}
        className="space-y-6"
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
        </TabsList>

        <TabsContent value="account" className="space-y-6">
          <AccountSettingsTab />
        </TabsContent>

        <TabsContent value="billing" className="space-y-6">
          <BillingSettingsTab returnUrl={returnUrl} displayPlan={displayPlan} />
        </TabsContent>

        <TabsContent value="developer" className="space-y-6">
          <DeveloperSettingsTab />
        </TabsContent>

        <TabsContent value="notifications" className="space-y-6">
          <NotificationsSettingsTab />
        </TabsContent>

        <TabsContent value="security" className="space-y-6">
          <SecuritySettingsTab />
        </TabsContent>

        <TabsContent value="privacy" className="space-y-6">
          <PrivacySettingsTab profile={profile ?? undefined} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

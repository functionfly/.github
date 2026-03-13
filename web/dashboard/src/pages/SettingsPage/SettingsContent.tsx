/**
 * Shared settings content: Account, Billing, API Keys, Notifications, Privacy.
 * Used on the standalone /settings page and on /u/{username} (profile Settings tab).
 */

import { useState, useEffect } from "react";
import { useLocation, useSearchParams } from "react-router-dom";
import { User, CreditCard, Key, Bell, Shield } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import { useQuery } from "@tanstack/react-query";
import { usersApi } from "@/api/users";
import { useAuthStore } from "@/stores/authStore";
import { useApiReachableStore } from "@/stores/apiReachableStore";
import { SettingsTab } from "@/pages/ProfilePage/components/tabs/SettingsTab";
import type { UserProfile } from "@/types";
import { VALID_TABS, type SettingsTabValue } from "./settings-utils";
import {
  AccountSettingsTab,
  BillingSettingsTab,
  ApiKeysSettingsTab,
  NotificationsSettingsTab,
} from "./components";

export interface SettingsContentProps {
  /** When false, omit the "Settings" page title (e.g. when embedded in profile). */
  showHeader?: boolean;
  /** Optional profile for Privacy tab (visibility, social links, etc.). */
  profile?: UserProfile | null;
  /** Initial tab when opened via path (e.g. "billing" for /u/username/settings/billing). */
  initialTab?: string;
}

export function SettingsContent({ showHeader = true, profile, initialTab: initialTabProp }: SettingsContentProps) {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const subtabFromUrl = searchParams.get("subtab");
  const [activeTab, setActiveTab] = useState<SettingsTabValue>(() => {
    if (initialTabProp && VALID_TABS.includes(initialTabProp as SettingsTabValue)) return initialTabProp as SettingsTabValue;
    if (subtabFromUrl && VALID_TABS.includes(subtabFromUrl as SettingsTabValue)) return subtabFromUrl as SettingsTabValue;
    return "account";
  });

  const apiReachable = useApiReachableStore((s) => s.apiReachable);
  const user = useAuthStore((s) => s.user);
  const setUserPlan = useAuthStore((s) => s.setUserPlan);
  const { data: meData } = useQuery({
    queryKey: ["users", "me"],
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
  const displayPlan = meData?.plan ?? user?.plan ?? "Starter";

  // Sync active tab from URL or initialTab prop (e.g. /u/username/settings/billing)
  useEffect(() => {
    const next = initialTabProp || subtabFromUrl;
    if (next && VALID_TABS.includes(next as SettingsTabValue)) {
      setActiveTab(next as SettingsTabValue);
    }
  }, [initialTabProp, subtabFromUrl]);

  // Billing portal return: show success toast and clean URL
  useEffect(() => {
    const success = searchParams.get("success");
    if (success === "true") {
      toast.success("Subscription updated successfully!");
      const next = new URLSearchParams(location.search);
      next.delete("success");
      const q = next.toString();
      window.history.replaceState({}, document.title, location.pathname + (q ? `?${q}` : ""));
    }
  }, [searchParams, location.pathname, location.search]);

  const returnUrl = `${window.location.origin}${location.pathname}${location.search ? location.search : ""}`;

  return (
    <div className="space-y-6">
      {showHeader && (
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Settings</h1>
          <p className="text-text-secondary">Manage your account and preferences</p>
        </div>
      )}

      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as SettingsTabValue)} className="space-y-6">
        <TabsList className="settings-page-tabs inline-flex h-auto flex-wrap gap-1 rounded-xl border border-border-default bg-bg-secondary/80 p-1.5 text-text-secondary backdrop-blur-sm">
          <TabsTrigger
            value="account"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200"
          >
            <User className="h-4 w-4 shrink-0" />
            Account
          </TabsTrigger>
          <TabsTrigger
            value="billing"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200"
          >
            <CreditCard className="h-4 w-4 shrink-0" />
            Billing
          </TabsTrigger>
          <TabsTrigger
            value="api"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200"
          >
            <Key className="h-4 w-4 shrink-0" />
            API Keys
          </TabsTrigger>
          <TabsTrigger
            value="notifications"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200"
          >
            <Bell className="h-4 w-4 shrink-0" />
            Notifications
          </TabsTrigger>
          <TabsTrigger
            value="privacy"
            className="settings-page-tab gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200"
          >
            <Shield className="h-4 w-4 shrink-0" />
            Privacy
          </TabsTrigger>
        </TabsList>

        <TabsContent value="account" className="space-y-6">
          <AccountSettingsTab />
        </TabsContent>

        <TabsContent value="billing" className="space-y-6">
          <BillingSettingsTab returnUrl={returnUrl} displayPlan={displayPlan} />
        </TabsContent>

        <TabsContent value="api" className="space-y-6">
          <ApiKeysSettingsTab />
        </TabsContent>

        <TabsContent value="notifications" className="space-y-6">
          <NotificationsSettingsTab />
        </TabsContent>

        <TabsContent value="privacy" className="space-y-6">
          <SettingsTab profile={profile ?? undefined} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

/**
 * VaultEnterprisePage
 *
 * DEPRECATED: This page has been replaced by VaultPage at /vault.
 * This page is kept for backward compatibility with a deprecation banner.
 */

import { useEffect } from "react";
import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCurrentPlan } from "@/hooks/useCurrentPlan";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";

import { VaultOverviewTab } from "@/components/VaultEnterprise/tabs/VaultOverviewTab";
import { SecurityTab } from "@/components/VaultEnterprise/tabs/SecurityTab";
import { DynamicCredsTab } from "@/components/VaultEnterprise/tabs/DynamicCredsTab";
import { AccessTab } from "@/components/VaultEnterprise/tabs/AccessTab";
import { EnterpriseTab } from "@/components/VaultEnterprise/tabs/EnterpriseTab";
import { ActivityTab } from "@/components/VaultEnterprise/tabs/ActivityTab";

const TABS: { value: string; label: string; plan: string[] }[] = [
  { value: "overview", label: "Overview", plan: ["free", "pro", "team", "enterprise"] },
  { value: "security", label: "Security", plan: ["pro", "team", "enterprise"] },
  { value: "dynamic", label: "Dynamic Creds", plan: ["pro", "team", "enterprise"] },
  { value: "access", label: "Access", plan: ["team", "enterprise"] },
  { value: "enterprise", label: "Enterprise", plan: ["team", "enterprise"] },
  { value: "activity", label: "Activity", plan: ["pro", "team", "enterprise"] },
];

interface VaultEnterprisePageProps {
  deprecated?: boolean;
}

export function VaultEnterprisePage({ deprecated }: VaultEnterprisePageProps) {
  const [tab, setTab] = useState<string>("overview");
  const { plan, isLoading } = useCurrentPlan();

  useEffect(() => {
    if (deprecated) {
      const dismissed = sessionStorage.getItem("vault-enterprise-deprecation-dismissed");
      if (!dismissed) {
        sessionStorage.setItem("vault-enterprise-deprecation-dismissed", "true");
        window.location.href = "/vault";
      }
    }
  }, [deprecated]);

  return (
    <div className="container mx-auto py-6 space-y-6">
      {deprecated && (
        <div className="mb-4 p-3 border border-amber-500/50 bg-amber-500/10 rounded-md flex items-center gap-3">
          <AlertTriangle className="h-5 w-5 text-amber-500 shrink-0" />
          <p className="text-sm flex-1">
            This page has moved to <a href="/vault" className="underline font-medium">/vault</a>. Please update your bookmarks.
          </p>
          <Button size="sm" variant="outline" onClick={() => window.location.href = "/vault"}>
            Go to Vault
          </Button>
        </div>
      )}
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Secrets Vault</h1>
          <p className="text-sm text-muted-foreground">
            Zero-knowledge encryption, dynamic credentials, and enterprise controls.
          </p>
        </div>
      </header>

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          {TABS.map((t) => (
            <TabsTrigger
              key={t.value}
              value={t.value}
              disabled={isLoading || !t.plan.includes(plan)}
            >
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <VaultOverviewTab plan={plan} />
        </TabsContent>
        <TabsContent value="security" className="space-y-6">
          <SecurityTab plan={plan} />
        </TabsContent>
        <TabsContent value="dynamic" className="space-y-6">
          <DynamicCredsTab plan={plan} />
        </TabsContent>
        <TabsContent value="access" className="space-y-6">
          <AccessTab plan={plan} />
        </TabsContent>
        <TabsContent value="enterprise" className="space-y-6">
          <EnterpriseTab plan={plan} />
        </TabsContent>
        <TabsContent value="activity" className="space-y-6">
          <ActivityTab plan={plan} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default VaultEnterprisePage;

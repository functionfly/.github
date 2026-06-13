/**
 * VaultEnterprisePage
 *
 * The unified UI surface for the FunctionFly vault's enterprise
 * features (Phases 1-5 + 6). Tabs map to the plan-gated feature
 * groups:
 *
 *   Overview      — plan + quota + cache + HA + secret browser
 *   Security      — MFA, IP allowlist, expiration, break-glass, escrow
 *   Dynamic Creds — DB targets + credential templates + leases
 *   Access        — RBAC roles + my assignments
 *   Enterprise    — namespaces, shares, SSO, SIEM webhooks
 *   Activity      — audit log + export + token monitor + dep graph
 *
 * Tabs that the tenant's plan doesn't unlock render as a
 * <LockedTabPlaceholder> so the user knows the capability exists.
 */

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCurrentPlan } from "@/hooks/useCurrentPlan";

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

export function VaultEnterprisePage() {
  const [tab, setTab] = useState<string>("overview");
  const { plan, isLoading } = useCurrentPlan();

  return (
    <div className="container mx-auto py-6 space-y-6">
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

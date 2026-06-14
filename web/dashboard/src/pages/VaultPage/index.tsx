/**
 * VaultPage
 *
 * Unified vault UI surface combining secrets management with enterprise features.
 * Tabs map to plan-gated feature groups:
 *
 *   Secrets        — Namespace tree + full SecretList (CRUD, search, rotation)
 *   Security       — MFA, IP allowlist, expiration, break-glass, escrow
 *   Dynamic Creds  — DB targets + credential templates + leases
 *   Access         — RBAC roles + my assignments
 *   Enterprise     — namespaces, shares, SSO, SIEM webhooks
 *   Activity       — audit log + export + token monitor + dep graph
 */

import "@/styles/professional-dashboard.css";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCurrentPlan } from "@/hooks/useCurrentPlan";
import { platformToVaultPlan } from "@/lib/vaultPlans";
import type { VaultPlan } from "@/types/vault-enterprise";

import { SecretsTab } from "@/components/VaultPage/tabs/SecretsTab";
import { SecurityTab } from "@/components/VaultPage/tabs/SecurityTab";
import { DynamicCredsTab } from "@/components/VaultPage/tabs/DynamicCredsTab";
import { AccessTab } from "@/components/VaultPage/tabs/AccessTab";
import { EnterpriseTab } from "@/components/VaultPage/tabs/EnterpriseTab";
import { ActivityTab } from "@/components/VaultPage/tabs/ActivityTab";

const TABS: { value: string; label: string; plan: VaultPlan[] }[] = [
  { value: "secrets", label: "Secrets", plan: ["free", "pro", "team", "enterprise"] },
  { value: "security", label: "Security", plan: ["pro", "team", "enterprise"] },
  { value: "dynamic", label: "Dynamic Creds", plan: ["pro", "team", "enterprise"] },
  { value: "access", label: "Access", plan: ["team", "enterprise"] },
  { value: "enterprise", label: "Enterprise", plan: ["team", "enterprise"] },
  { value: "activity", label: "Activity", plan: ["pro", "team", "enterprise"] },
];

interface VaultPageProps {
  deprecationBanner?: React.ReactNode;
}

export function VaultPage({ deprecationBanner }: VaultPageProps) {
  const [tab, setTab] = useState<string>("secrets");
  const { plan: platformPlan, isLoading } = useCurrentPlan();
  const plan = platformToVaultPlan(platformPlan) as VaultPlan;

  return (
    <div className="professional-dashboard container mx-auto py-6 space-y-6">
      {deprecationBanner}
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

        <TabsContent value="secrets" className="space-y-6">
          <SecretsTab plan={plan} />
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

export default VaultPage;

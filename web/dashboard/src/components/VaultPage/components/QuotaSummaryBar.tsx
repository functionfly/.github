/**
 * QuotaSummaryBar — Collapsible quota summary using <details>/<summary>
 */

import { useVaultSecrets } from "@/hooks/useVault";
import { getPlanLimits, PLAN_META } from "@/lib/vaultPlans";
import type { VaultPlan } from "@/types/vault-enterprise";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";

interface QuotaSummaryBarProps {
  plan: VaultPlan;
  platformPlan?: string;
}

function formatPlanName(plan: string): string {
  return plan
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export function QuotaSummaryBar({ plan, platformPlan }: QuotaSummaryBarProps) {
  const { data: secretsResp } = useVaultSecrets();
  const limits = getPlanLimits(plan);
  const displayName = platformPlan ? formatPlanName(platformPlan) : PLAN_META[plan].name;
  const currentSecretCount = secretsResp?.secrets?.length ?? 0;
  const secretPct = Math.round((currentSecretCount / limits.maxSecrets) * 100);

  return (
    <details className="border rounded-md">
      <summary className="flex items-center gap-2 p-3 cursor-pointer select-none text-sm font-medium">
        <Badge variant="outline">{displayName}</Badge>
        <span className="text-[var(--text-faint)]">
          {currentSecretCount.toLocaleString()}/{limits.maxSecrets.toLocaleString()} secrets
        </span>
        <span className="text-[var(--text-faint)]">·</span>
        <span className="text-[var(--text-faint)]">
          {limits.maxDynamicCreds.toLocaleString()} dyn creds
        </span>
        <Progress value={secretPct} className="h-1.5 w-24 ml-2" />
      </summary>
      <div className="px-3 pb-3 text-sm text-[var(--text-faint)]">
        Upgrade to {PLAN_META[plan === "enterprise" ? "enterprise" : plan === "team" ? "enterprise" : "pro"]?.name} for more secrets and dynamic credentials.
      </div>
    </details>
  );
}

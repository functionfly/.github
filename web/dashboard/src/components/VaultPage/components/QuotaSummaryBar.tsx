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
}

export function QuotaSummaryBar({ plan }: QuotaSummaryBarProps) {
  const { data: secretsResp } = useVaultSecrets();
  const limits = getPlanLimits(plan);
  const meta = PLAN_META[plan];
  const currentSecretCount = secretsResp?.secrets?.length ?? 0;
  const secretPct = Math.round((currentSecretCount / limits.maxSecrets) * 100);

  return (
    <details className="border rounded-md">
      <summary className="flex items-center gap-2 p-3 cursor-pointer select-none text-sm font-medium">
        <Badge variant="outline">{meta.name}</Badge>
        <span className="text-muted-foreground">
          {currentSecretCount.toLocaleString()}/{limits.maxSecrets.toLocaleString()} secrets
        </span>
        <span className="text-muted-foreground">·</span>
        <span className="text-muted-foreground">
          {limits.maxDynamicCreds.toLocaleString()} dyn creds
        </span>
        <Progress value={secretPct} className="h-1.5 w-24 ml-2" />
      </summary>
      <div className="px-3 pb-3 text-sm text-muted-foreground">
        Upgrade to {PLAN_META[plan === "enterprise" ? "enterprise" : plan === "team" ? "enterprise" : "pro"]?.name} for more secrets and dynamic credentials.
      </div>
    </details>
  );
}

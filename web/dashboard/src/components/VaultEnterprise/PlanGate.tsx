/**
 * PlanGate
 *
 * Wraps a feature panel and renders an upgrade prompt when the
 * tenant's plan doesn't include the feature. The component:
 *   - Reads the tenant's plan from the auth context (or `currentPlan` prop)
 *   - Looks up the feature in FEATURE_MIN_PLAN
 *   - Renders children when allowed, or an <UpgradePrompt> when not
 *
 * The locked variant of a feature is still implemented in the
 * parent — we don't block the import, only the rendering — so
 * route deep-links from docs work and the UI can be inspected.
 */

import { Lock, Sparkles } from "lucide-react";
import { Link } from "react-router-dom";
import type { ReactNode } from "react";
import { isFeatureAvailable, minPlanForFeature, PLAN_META } from "@/lib/vaultPlans";
import type { PlanLimits, VaultPlan } from "@/types/vault-enterprise";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface PlanGateProps {
  feature: keyof PlanLimits["features"];
  plan: VaultPlan;
  children: ReactNode;
  /** Optional title shown in the locked state. */
  title?: string;
  /** Optional description shown in the locked state. */
  description?: string;
  /** When true, only render children if locked (skip the upgrade card). */
  silent?: boolean;
}

export function PlanGate({
  feature,
  plan,
  children,
  title,
  description,
  silent = false,
}: PlanGateProps) {
  if (isFeatureAvailable(plan, feature)) {
    return <>{children}</>;
  }
  if (silent) {
    return null;
  }
  const minPlan = minPlanForFeature(feature);
  return (
    <UpgradePrompt
      title={title ?? "Upgrade to unlock"}
      description={
        description ??
        `This feature is available on ${PLAN_META[minPlan].name} and above.`
      }
      minPlan={minPlan}
      currentPlan={plan}
      feature={feature}
    />
  );
}

interface UpgradePromptProps {
  title: string;
  description: string;
  minPlan: VaultPlan;
  currentPlan: VaultPlan;
  feature: keyof PlanLimits["features"];
}

export function UpgradePrompt({
  title,
  description,
  minPlan,
  currentPlan,
  feature,
}: UpgradePromptProps) {
  return (
    <Card className="border-dashed bg-[var(--panel-raised)]/30">
      <CardHeader>
        <div className="flex items-center gap-2">
          <Lock className="h-5 w-5 text-[var(--text-faint)]" />
          <CardTitle className="text-base">{title}</CardTitle>
          <Badge variant="secondary" className="ml-auto">
            {PLAN_META[minPlan].name}+
          </Badge>
        </div>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <Button asChild>
            <Link to="/billing" state={{ fromFeature: feature, currentPlan }}>
              <Sparkles className="mr-2 h-4 w-4" />
              Upgrade to {PLAN_META[minPlan].name}
            </Link>
          </Button>
          <span className="text-sm text-[var(--text-faint)]">
            You're on {PLAN_META[currentPlan].name}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * WithFeature is a small HOC-style wrapper for inline use.
 *   <WithFeature feature="sso" plan={plan}><SsoConfigCard /></WithFeature>
 */
export function WithFeature({
  feature,
  plan,
  children,
  fallback,
}: PlanGateProps & { fallback?: ReactNode }) {
  if (isFeatureAvailable(plan, feature)) {
    return <>{children}</>;
  }
  return <>{fallback ?? null}</>;
}

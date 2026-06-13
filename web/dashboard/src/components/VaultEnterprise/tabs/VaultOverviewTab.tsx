import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Activity,
  Database,
  HardDrive,
  KeyRound,
  ServerCog,
  ShieldCheck,
  TrendingUp,
} from "lucide-react";

import { useCacheStats } from "@/api/vault";
import { useCurrentPlan } from "@/hooks/useCurrentPlan";
import { PLAN_META, getPlanLimits } from "@/lib/vaultPlans";
import { formatBytes, formatPlan, percent } from "@/components/VaultEnterprise/utils";
import { SecretBrowser } from "@/components/VaultEnterprise/SecretBrowser";
import type { VaultPlan } from "@/types/vault-enterprise";

interface VaultOverviewTabProps {
  plan: VaultPlan;
}

export function VaultOverviewTab({ plan }: VaultOverviewTabProps) {
  const limits = getPlanLimits(plan);
  const { data: cacheStats, isLoading: cacheLoading } = useCacheStats();
  const { plan: rehydratedPlan } = useCurrentPlan();
  // Use the passed plan in priority over the rehydrated one — the
  // parent has the freshest value.
  const currentPlan = plan ?? rehydratedPlan;

  return (
    <div className="space-y-6">
      {/* Plan summary card */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <PlanCard plan={currentPlan} />
        <QuotaCard label="Secrets" used={0} limit={limits.maxSecrets} unit="secrets" />
        <QuotaCard
          label="Dynamic creds / 30d"
          used={0}
          limit={limits.maxDynamicCreds}
          unit="credentials"
        />
        <QuotaCard
          label="Audit exports / day"
          used={0}
          limit={limits.auditExportsPerDay}
          unit="exports"
        />
      </div>

      {/* System health row */}
      <div className="grid gap-4 md:grid-cols-3">
        <CacheCard
          loading={cacheLoading}
          enabled={cacheStats?.enabled}
          meta={cacheStats?.meta_keys ?? 0}
          tokens={cacheStats?.token_keys ?? 0}
        />
        <HACard plan={currentPlan} />
        <KDFCard />
      </div>

      {/* The big secret browser */}
      <Card>
        <CardHeader>
          <CardTitle>Secrets</CardTitle>
          <CardDescription>
            {limits.maxSecrets > 0
              ? `Browse secrets by namespace, view expiration status, and open the dependency graph.`
              : `Your free plan includes up to ${limits.maxSecrets} secrets.`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SecretBrowser plan={currentPlan} />
        </CardContent>
      </Card>
    </div>
  );
}

function PlanCard({ plan }: { plan: VaultPlan }) {
  const meta = PLAN_META[plan];
  const limits = getPlanLimits(plan);
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>Plan</CardDescription>
        <CardTitle className="text-xl flex items-center gap-2">
          <ShieldCheck className="h-5 w-5 text-primary" />
          {meta.name}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{meta.tagline}</p>
        <p className="text-xs text-muted-foreground mt-2">
          {limits.maxSecrets.toLocaleString()} secrets · {limits.maxDynamicCreds.toLocaleString()} dyn creds
        </p>
      </CardContent>
    </Card>
  );
}

function QuotaCard({
  label,
  used,
  limit,
  unit,
}: {
  label: string;
  used: number;
  limit: number;
  unit: string;
}) {
  const pct = percent(used, limit);
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>{label}</CardDescription>
        <CardTitle className="text-xl">
          {used.toLocaleString()} <span className="text-muted-foreground text-sm">/ {limit.toLocaleString()}</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Progress value={pct} className="h-1.5" />
        <p className="text-xs text-muted-foreground mt-2">
          {pct}% of {unit} used
        </p>
      </CardContent>
    </Card>
  );
}

function CacheCard({
  loading,
  enabled,
  meta,
  tokens,
}: {
  loading: boolean;
  enabled?: boolean;
  meta: number;
  tokens: number;
}) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>Cache (Phase 5.1)</CardDescription>
        <CardTitle className="text-base flex items-center gap-2">
          <HardDrive className="h-4 w-4" />
          {loading ? <Skeleton className="h-5 w-24" /> : enabled ? "Connected" : "Disabled"}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-4 w-full" />
        ) : (
          <p className="text-sm text-muted-foreground">
            {meta.toLocaleString()} metadata keys · {tokens.toLocaleString()} token keys
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function HACard({ plan }: { plan: VaultPlan }) {
  // The HA leader status endpoint requires Enterprise; we just
  // show a placeholder for lower plans.
  const available = plan === "enterprise";
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>HA Leader (Phase 5.3)</CardDescription>
        <CardTitle className="text-base flex items-center gap-2">
          <ServerCog className="h-4 w-4" />
          {available ? "Active-Active" : "Single instance"}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          {available
            ? "Leader election via Redis SETNX; only the elected node runs sweepers."
            : "Upgrade to Enterprise to enable active-active leader election."}
        </p>
      </CardContent>
    </Card>
  );
}

function KDFCard() {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>KDF (Phase 1.5)</CardDescription>
        <CardTitle className="text-base flex items-center gap-2">
          <KeyRound className="h-4 w-4" />
          Argon2id
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          OWASP 2023 parameters: 64 MiB, 3 iterations, 4 parallel.
        </p>
      </CardContent>
    </Card>
  );
}

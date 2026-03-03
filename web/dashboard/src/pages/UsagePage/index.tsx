import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import {
  CreditCard,
  Loader2,
  TrendingUp,
  Zap,
  AlertCircle,
  ExternalLink,
  BarChart3,
  FunctionSquare,
  Cloud,
  Building2,
  Database,
  Bot,
  DollarSign,
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { dashboardApi } from "@/api/dashboard";
import { usePlan } from "@/hooks/usePlan";
import { useAuthStore } from "@/stores/authStore";
import { usersApi } from "@/api/users";
import { createBillingPortalSession, getBillingPortalErrorMessage } from "@/api/billing";
import { toast } from "sonner";
import { useState } from "react";
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { UsageGraph, ExecutionRateChart } from "@/components/dashboard";
import { ROUTES } from "@/lib/constants";
import { functionsApi } from "@/api/functions";
import { providersApi } from "@/api/providers";
import { appsApi } from "@/api/apps";
import { stateFabricApi } from "@/api/stateFabric";
import { agentApi } from "@/api/agent";

const USAGE_DAYS = 30;
const MAX_AGENTS_FOR_USAGE = 25;

export function UsagePage() {
  const user = useAuthStore((s) => s.user);
  const { plan, limits, displayName, isEnterprise } = usePlan();
  const [billingLoading, setBillingLoading] = useState(false);

  const { data: meData } = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      try {
        return await usersApi.getMe();
      } catch {
        return undefined;
      }
    },
    retry: false,
  });
  const displayPlan = meData?.plan ?? user?.plan ?? plan ?? "free";
  const username = user?.username ?? meData?.username;

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ["dashboard", "usage", USAGE_DAYS],
    queryFn: () => dashboardApi.getUsage(USAGE_DAYS),
  });

  const { data: executionRateDataRes, isLoading: executionRateLoading } = useQuery({
    queryKey: ["dashboard", "execution-rate", 168],
    queryFn: () => dashboardApi.getExecutionRate(168),
  });

  const { data: functionsData } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });
  const { data: providersData } = useQuery({
    queryKey: ["providers"],
    queryFn: () => providersApi.getConnectedProviders(),
  });
  const { data: appsData } = useQuery({
    queryKey: ["apps"],
    queryFn: async () => {
      const res = await appsApi.list();
      return res?.apps ?? [];
    },
  });

  const { data: stateFabricsList } = useQuery({
    queryKey: ["state-fabrics"],
    queryFn: () => stateFabricApi.list(),
  });
  const fabricIds = useMemo(
    () => (stateFabricsList ?? []).map((f) => f.id),
    [stateFabricsList]
  );
  const { data: stateFabricMetricsMap } = useQuery({
    queryKey: ["state-fabrics-metrics", fabricIds],
    queryFn: async () => {
      const map: Record<string, { totalOperations: number; storageUsed: number }> = {};
      await Promise.all(
        fabricIds.slice(0, 20).map(async (id) => {
          try {
            const m = await stateFabricApi.getMetrics(id);
            map[id] = {
              totalOperations: (m as { totalOperations?: number }).totalOperations ?? 0,
              storageUsed: (m as { storageUsed?: number }).storageUsed ?? 0,
            };
          } catch {
            map[id] = { totalOperations: 0, storageUsed: 0 };
          }
        })
      );
      return map;
    },
    enabled: fabricIds.length > 0,
  });

  const { data: agentsListRes } = useQuery({
    queryKey: ["agents-list"],
    queryFn: () => agentApi.listAgents({ limit: MAX_AGENTS_FOR_USAGE }),
  });
  const agentIds = useMemo(
    () => (agentsListRes?.agents ?? []).map((a) => a.agentId),
    [agentsListRes]
  );
  const { data: agentsUsageAndBalance } = useQuery({
    queryKey: ["agents-usage-balance", agentIds],
    queryFn: async () => {
      let totalCallsToday = 0;
      let totalSpendToday = 0;
      let totalBalanceUsd = 0;
      await Promise.all(
        agentIds.map(async (agentId) => {
          try {
            const [usageRes, balanceRes] = await Promise.all([
              agentApi.getUsage(agentId),
              agentApi.getCreditBalance(agentId),
            ]);
            const u = (usageRes as { usage?: { calls_today?: number; callsToday?: number; spend_today_usd?: number; spendToday?: number } })?.usage;
            if (u) {
              totalCallsToday += Number(u.calls_today ?? u.callsToday ?? 0);
              totalSpendToday += Number(u.spend_today_usd ?? u.spendToday ?? 0);
            }
            const b = (balanceRes as { balance?: { credit_balance_usd?: number; balanceUSD?: number } })?.balance;
            if (b) {
              totalBalanceUsd += Number(b.credit_balance_usd ?? b.balanceUSD ?? 0);
            }
          } catch {
            // skip failed agent
          }
        })
      );
      return { totalCallsToday, totalSpendToday, totalBalanceUsd };
    },
    enabled: agentIds.length > 0,
  });

  const functionsCount = (functionsData?.functions ?? []).length;
  const providersCount = (providersData ?? []).length;
  const appsCount = Array.isArray(appsData) ? appsData.length : 0;
  const functionsLimit = limits?.functions ?? 0;
  const providersLimit = limits?.providers ?? 0;
  const stateFabricsLimit = (limits as { stateFabrics?: number })?.stateFabrics ?? 0;
  const agentsLimit = (limits as { agents?: number })?.agents ?? 0;
  const formatLimit = (value: number) =>
    typeof value === "number" && value === Infinity ? "∞" : value?.toLocaleString?.() ?? "—";
  const stateFabricTotals = useMemo(() => {
    if (!stateFabricMetricsMap || !fabricIds.length) return { operations: 0, storage: 0 };
    let operations = 0;
    let storage = 0;
    fabricIds.forEach((id) => {
      const m = stateFabricMetricsMap[id];
      if (m) {
        operations += m.totalOperations;
        storage += m.storageUsed;
      }
    });
    return { operations, storage };
  }, [stateFabricMetricsMap, fabricIds]);

  const usageGraphData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + "Z").toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
      }),
      value: Number(d.value),
    }));
  }, [usageData]);

  const executionRateData = useMemo(() => {
    const raw = executionRateDataRes?.data ?? [];
    return raw.map((d) => ({
      time: d.time,
      rate: Number(d.rate),
    }));
  }, [executionRateDataRes]);

  const totalUsage = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.reduce((sum, d) => sum + Number(d.value), 0);
  }, [usageData]);

  const periodComparison = useMemo(() => {
    const raw = usageData?.data ?? [];
    const len = raw.length;
    const last7 = raw.slice(-7).reduce((s, d) => s + Number(d.value), 0);
    const prev7 = len >= 14 ? raw.slice(-14, -7).reduce((s, d) => s + Number(d.value), 0) : 0;
    const change = prev7 > 0 ? ((last7 - prev7) / prev7) * 100 : (last7 > 0 ? 100 : 0);
    return { last7, prev7, change };
  }, [usageData]);

  const usageDistributionData = useMemo(() => {
    const requests = totalUsage;
    const stateFabricOps = stateFabricTotals.operations;
    const agentCalls = agentsUsageAndBalance?.totalCallsToday ?? 0;
    const total = requests + stateFabricOps + agentCalls;
    if (total === 0) return [];
    return [
      { name: "Function requests", value: requests, color: "var(--color-brand-500)" },
      { name: "State Fabric ops", value: stateFabricOps, color: "#06b6d4" },
      { name: "Agent calls", value: agentCalls, color: "#8b5cf6" },
    ].filter((d) => d.value > 0);
  }, [totalUsage, stateFabricTotals.operations, agentsUsageAndBalance?.totalCallsToday]);

  const resourceUtilizationData = useMemo(() => {
    const toPct = (used: number, max: number) =>
      typeof max === "number" && max !== Infinity && max > 0
        ? Math.min(100, Math.round((used / max) * 100))
        : 0;
    return [
      { resource: "Functions", used: functionsCount, max: functionsLimit, pct: toPct(functionsCount, functionsLimit) },
      { resource: "Providers", used: providersCount, max: providersLimit, pct: toPct(providersCount, providersLimit) },
      { resource: "State Fabrics", used: (stateFabricsList ?? []).length, max: stateFabricsLimit, pct: toPct((stateFabricsList ?? []).length, stateFabricsLimit) },
      { resource: "Agents", used: agentIds.length, max: agentsLimit, pct: toPct(agentIds.length, agentsLimit) },
    ];
  }, [functionsCount, functionsLimit, providersCount, providersLimit, stateFabricsList, stateFabricsLimit, agentIds.length, agentsLimit]);

  const requestLimit = limits?.requests ?? 0;
  const isUnlimited = requestLimit === Infinity || isEnterprise;
  const remaining = isUnlimited ? null : Math.max(0, requestLimit - totalUsage);
  const usagePercent = isUnlimited
    ? 0
    : requestLimit > 0
      ? Math.min(100, (totalUsage / requestLimit) * 100)
      : 0;
  const isOverLimit = !isUnlimited && totalUsage > requestLimit;

  const openBillingPortal = async () => {
    setBillingLoading(true);
    try {
      const { url } = await createBillingPortalSession(
        `${window.location.origin}${ROUTES.USAGE}`
      );
      if (url) window.location.href = url;
    } catch (e) {
      toast.error(getBillingPortalErrorMessage(e));
    } finally {
      setBillingLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary tracking-tight">Usage</h1>
        <p className="text-text-secondary mt-1">
          Track platform usage across functions, State Fabric, agents, and billing.
        </p>
      </div>

      {/* Plan & remaining */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
              <Zap className="h-4 w-4" />
              Current plan
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xl font-semibold text-text-primary capitalize">
              {displayName || displayPlan || "Free"} Plan
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-3 w-full"
              onClick={openBillingPortal}
              disabled={billingLoading}
            >
              {billingLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <>
                  <CreditCard className="h-4 w-4 mr-2" />
                  Manage billing
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              Requests (last {USAGE_DAYS} days)
            </CardTitle>
          </CardHeader>
          <CardContent>
            {usageLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-xl font-semibold text-text-primary">
                {totalUsage.toLocaleString()}
              </p>
            )}
          </CardContent>
        </Card>

        {!isUnlimited && (
          <Card className="border-theme bg-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
                <TrendingUp className="h-4 w-4" />
                Remaining this period
              </CardTitle>
            </CardHeader>
            <CardContent>
              {usageLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
              ) : (
                <p
                  className={`text-xl font-semibold ${
                    isOverLimit ? "text-destructive" : "text-text-primary"
                  }`}
                >
                  {remaining !== null
                    ? remaining.toLocaleString()
                    : "—"}
                </p>
              )}
            </CardContent>
          </Card>
        )}

        {isOverLimit && (
          <Card className="border-theme bg-card border-destructive/50">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                Over limit
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-text-secondary">
                You’ve exceeded your plan’s request limit. Upgrade or wait for the next period.
              </p>
              <Button
                variant="default"
                size="sm"
                className="mt-3 w-full"
                onClick={openBillingPortal}
                disabled={billingLoading}
              >
                Upgrade plan
              </Button>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Usage vs limit progress */}
      {!isUnlimited && requestLimit > 0 && (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base">Usage vs plan limit</CardTitle>
            <CardDescription>
              Monthly request allowance. Overage may incur charges depending on your plan.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {usageLoading ? (
              <div className="h-4 w-full rounded-full bg-bg-hover animate-pulse" />
            ) : (
              <div className="space-y-2">
                <div className="flex justify-between text-sm text-text-secondary">
                  <span>
                    {totalUsage.toLocaleString()} / {requestLimit.toLocaleString()} requests
                  </span>
                  <span>{usagePercent.toFixed(0)}%</span>
                </div>
                <Progress
                  value={usagePercent}
                  className={isOverLimit ? "[&>div]:bg-destructive" : undefined}
                />
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Plan limits reference */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Your plan limits</CardTitle>
          <CardDescription>
            Maximum included in your {displayName || displayPlan} plan. Upgrade to increase limits.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 text-sm">
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Requests/mo</p>
              <p className="text-text-primary font-semibold">{formatLimit(requestLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Functions</p>
              <p className="text-text-primary font-semibold">{formatLimit(functionsLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Providers</p>
              <p className="text-text-primary font-semibold">{formatLimit(providersLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Custom domains</p>
              <p className="text-text-primary font-semibold">
                {formatLimit((limits as { customDomains?: number })?.customDomains ?? 0)}
              </p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">State Fabrics</p>
              <p className="text-text-primary font-semibold">{formatLimit(stateFabricsLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Agents</p>
              <p className="text-text-primary font-semibold">{formatLimit(agentsLimit)}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Resource counts: Functions, Providers, Apps */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Resources</CardTitle>
          <CardDescription>
            Current usage vs your plan limits above.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <FunctionSquare className="h-5 w-5 text-text-muted shrink-0" />
              <div>
                <p className="text-sm font-medium text-text-secondary">Functions</p>
                <p className="text-lg font-semibold text-text-primary">
                  {functionsCount}
                  <span className="text-text-muted font-normal text-sm">
                    {" "}/ {formatLimit(functionsLimit)}
                  </span>
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <Cloud className="h-5 w-5 text-text-muted shrink-0" />
              <div>
                <p className="text-sm font-medium text-text-secondary">Providers</p>
                <p className="text-lg font-semibold text-text-primary">
                  {providersCount}
                  <span className="text-text-muted font-normal text-sm">
                    {" "}/ {formatLimit(providersLimit)}
                  </span>
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <Building2 className="h-5 w-5 text-text-muted shrink-0" />
              <div>
                <p className="text-sm font-medium text-text-secondary">Apps</p>
                <p className="text-lg font-semibold text-text-primary">{appsCount}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* State Fabric usage */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Database className="h-4 w-4" />
            State Fabric
          </CardTitle>
          <CardDescription>
            Aggregate usage across your State Fabric instances (operations and storage).
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <div>
                <p className="text-sm font-medium text-text-secondary">Fabrics</p>
                <p className="text-lg font-semibold text-text-primary">
                  {(stateFabricsList ?? []).length}
                  <span className="text-text-muted font-normal text-sm">
                    {" "}/ {formatLimit(stateFabricsLimit)}
                  </span>
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <div>
                <p className="text-sm font-medium text-text-secondary">Total operations</p>
                <p className="text-lg font-semibold text-text-primary">
                  {stateFabricTotals.operations.toLocaleString()}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <div>
                <p className="text-sm font-medium text-text-secondary">Storage used</p>
                <p className="text-lg font-semibold text-text-primary">
                  {stateFabricTotals.storage >= 1024
                    ? `${(stateFabricTotals.storage / 1024).toFixed(1)} GB`
                    : `${stateFabricTotals.storage} MB`}
                </p>
              </div>
            </div>
          </div>
          {(stateFabricsList ?? []).length > 0 && (
            <Button variant="ghost" size="sm" className="mt-2" asChild>
              <Link to={ROUTES.STATE_FABRIC}>
                View State Fabric
                <ExternalLink className="h-4 w-4 ml-2 opacity-60" />
              </Link>
            </Button>
          )}
        </CardContent>
      </Card>

      {/* Agents usage & credits */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Bot className="h-4 w-4" />
            Agents
          </CardTitle>
          <CardDescription>
            Agent execution usage (calls today, spend) and total credit balance across your agents.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <div>
                <p className="text-sm font-medium text-text-secondary">Agents</p>
                <p className="text-lg font-semibold text-text-primary">
                  {agentIds.length}
                  <span className="text-text-muted font-normal text-sm">
                    {" "}/ {formatLimit(agentsLimit)}
                  </span>
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <div>
                <p className="text-sm font-medium text-text-secondary">Calls today</p>
                <p className="text-lg font-semibold text-text-primary">
                  {(agentsUsageAndBalance?.totalCallsToday ?? 0).toLocaleString()}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle bg-bg-secondary">
              <DollarSign className="h-5 w-5 text-text-muted shrink-0" />
              <div>
                <p className="text-sm font-medium text-text-secondary">Spend today / Balance</p>
                <p className="text-lg font-semibold text-text-primary">
                  ${(agentsUsageAndBalance?.totalSpendToday ?? 0).toFixed(2)} / $
                  {(agentsUsageAndBalance?.totalBalanceUsd ?? 0).toFixed(2)}
                </p>
              </div>
            </div>
          </div>
          {agentIds.length > 0 && (
            <Button variant="ghost" size="sm" className="mt-2" asChild>
              <Link to={ROUTES.AGENTS}>
                View Agents
                <ExternalLink className="h-4 w-4 ml-2 opacity-60" />
              </Link>
            </Button>
          )}
        </CardContent>
      </Card>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {usageLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <UsageGraph
            data={usageGraphData}
            title={`Requests (last ${USAGE_DAYS} days)`}
            valueLabel="Requests"
          />
        )}
        {executionRateLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <ExecutionRateChart
            data={executionRateData}
            title="Execution rate (last 7 days)"
            unit="exec/s"
          />
        )}
      </div>

      {/* Advanced visualizations */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-text-primary">Advanced visualizations</h2>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* Usage distribution (donut) */}
          <Card className="border-theme bg-card">
            <CardHeader>
              <CardTitle className="text-base">Usage by category</CardTitle>
              <CardDescription>
                Share of usage across function requests, State Fabric ops, and agent calls.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[240px] min-h-[240px] w-full min-w-0">
                {usageDistributionData.length === 0 ? (
                  <div className="flex h-full items-center justify-center text-text-muted text-sm">
                    No usage data yet
                  </div>
                ) : (
                  <ResponsiveContainer width="100%" height="100%" minHeight={240}>
                    <RechartsPieChart>
                      <Pie
                        data={usageDistributionData}
                        cx="50%"
                        cy="50%"
                        innerRadius={56}
                        outerRadius={80}
                        paddingAngle={2}
                        dataKey="value"
                        nameKey="name"
                        label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                        labelLine={{ stroke: "var(--color-border-subtle)" }}
                        isAnimationActive={false}
                      >
                        {usageDistributionData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: number) => [value.toLocaleString(), "Count"]}
                        contentStyle={{
                          backgroundColor: "var(--color-bg-tertiary)",
                          border: "1px solid var(--color-border-default)",
                          borderRadius: "8px",
                        }}
                        labelStyle={{ color: "var(--color-text-primary)" }}
                      />
                    </RechartsPieChart>
                  </ResponsiveContainer>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Resource utilization (horizontal bar) */}
          <Card className="border-theme bg-card">
            <CardHeader>
              <CardTitle className="text-base">Resource utilization</CardTitle>
              <CardDescription>
                Used vs plan limit (%) for each resource.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[240px] min-h-[240px] w-full min-w-0">
                <ResponsiveContainer width="100%" height="100%" minHeight={240}>
                  <RechartsBarChart
                    data={resourceUtilizationData}
                    layout="vertical"
                    margin={{ top: 4, right: 24, left: 70, bottom: 4 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border-subtle)" horizontal={false} />
                    <XAxis type="number" domain={[0, 100]} tick={{ fill: "var(--color-text-muted)", fontSize: 11 }} tickFormatter={(v) => `${v}%`} />
                    <YAxis type="category" dataKey="resource" width={70} tick={{ fill: "var(--color-text-muted)", fontSize: 11 }} />
                    <Tooltip
                      formatter={(value: number, _name: string, item: { payload?: { used: number; max: number } }) => {
                        const p = item?.payload;
                        const maxStr = p && typeof p.max === "number" && p.max !== Infinity ? p.max : "∞";
                        return [`${value}% (${p?.used ?? 0} / ${maxStr})`, "Utilization"];
                      }}
                      contentStyle={{
                        backgroundColor: "var(--color-bg-tertiary)",
                        border: "1px solid var(--color-border-default)",
                        borderRadius: "8px",
                      }}
                      labelStyle={{ color: "var(--color-text-primary)" }}
                    />
                    <Bar dataKey="pct" fill="var(--color-brand-500)" radius={[0, 4, 4, 0]} name="Utilization %" isAnimationActive={false} />
                  </RechartsBarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Period-over-period comparison */}
          <Card className="border-theme bg-card">
            <CardHeader>
              <CardTitle className="text-base">Request trend</CardTitle>
              <CardDescription>
                Last 7 days vs previous 7 days.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[240px] min-h-[240px] w-full min-w-0 flex flex-col justify-center">
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div className="rounded-lg border border-border-subtle bg-bg-secondary p-4">
                    <p className="text-xs font-medium text-text-muted uppercase tracking-wider">Last 7 days</p>
                    <p className="text-2xl font-bold text-text-primary mt-1">
                      {periodComparison.last7.toLocaleString()}
                    </p>
                    <p className="text-sm text-text-muted">requests</p>
                  </div>
                  <div className="rounded-lg border border-border-subtle bg-bg-secondary p-4">
                    <p className="text-xs font-medium text-text-muted uppercase tracking-wider">Previous 7 days</p>
                    <p className="text-2xl font-bold text-text-primary mt-1">
                      {periodComparison.prev7.toLocaleString()}
                    </p>
                    <p className="text-sm text-text-muted">requests</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {periodComparison.change > 0 && (
                    <TrendingUp className="h-5 w-5 text-emerald-500 shrink-0" />
                  )}
                  {periodComparison.change < 0 && (
                    <span className="text-red-500 text-lg">↓</span>
                  )}
                  <span
                    className={
                      periodComparison.change > 0
                        ? "text-emerald-600 dark:text-emerald-400 font-medium"
                        : periodComparison.change < 0
                          ? "text-red-600 dark:text-red-400 font-medium"
                          : "text-text-muted font-medium"
                    }
                  >
                    {periodComparison.change > 0 && "+"}
                    {periodComparison.change.toFixed(1)}%
                  </span>
                  <span className="text-text-secondary text-sm">vs previous 7 days</span>
                </div>
                <div className="mt-4 h-2 w-full rounded-full bg-bg-hover overflow-hidden flex">
                  <div
                    className="h-full bg-brand-500 rounded-l-full transition-all duration-500"
                    style={{
                      width: `${periodComparison.prev7 + periodComparison.last7 > 0 ? (periodComparison.last7 / (periodComparison.prev7 + periodComparison.last7)) * 100 : 50}%`,
                    }}
                  />
                </div>
                <p className="text-xs text-text-muted mt-2">Proportion: last 7 days (filled) vs previous 7 days</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Billing & docs */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Billing & limits</CardTitle>
          <CardDescription>
            Manage your subscription, payment methods, and view invoices. Usage includes
            function requests, State Fabric operations, and agent execution spend.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={openBillingPortal}
            disabled={billingLoading}
          >
            {billingLoading ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <CreditCard className="h-4 w-4 mr-2" />
            )}
            Open billing portal
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to={username ? `/u/${username}/settings/billing` : ROUTES.SETTINGS}>
              Settings
              <ExternalLink className="h-4 w-4 ml-2 opacity-60" />
            </Link>
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to={ROUTES.PRICING}>
              View plans
              <ExternalLink className="h-4 w-4 ml-2 opacity-60" />
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

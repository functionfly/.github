import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { dashboardApi } from '@/api/dashboard';
import { usersApi } from '@/api/users';
import { functionsApi } from '@/api/functions';
import { providersApi } from '@/api/providers';
import { appsApi } from '@/api/apps';
import { stateFabricApi } from '@/api/stateFabric';
import { agentApi } from '@/api/agent';
import {
  getCostSummary,
  getCostByFunction,
  getCostByPeriod,
  getCostByRegion,
  getUsageForecast,
  getSpendCap,
} from '@/api/usageAnalytics';
import { useAuthStore } from '@/stores/authStore';
import { usePlan } from '@/hooks/usePlan';
import { USAGE_DAYS, MAX_AGENTS_FOR_USAGE, type DateRangeValue } from '../constants';
import { getDateRange } from '../utils';

export function useUsagePageData(dateRange: DateRangeValue) {
  const user = useAuthStore((s) => s.user);
  const { plan, limits } = usePlan();

  const dateRangeParams = useMemo(() => getDateRange(dateRange), [dateRange]);

  // User data
  const { data: meData } = useQuery({
    queryKey: ['users', 'me'],
    queryFn: async () => {
      try {
        return await usersApi.getMe();
      } catch {
        return undefined;
      }
    },
    retry: false,
  });
  const displayPlan = meData?.plan ?? user?.plan ?? plan ?? 'free';
  const username = user?.username ?? meData?.username;

  // Basic usage data
  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ['dashboard', 'usage', USAGE_DAYS],
    queryFn: () => dashboardApi.getUsage(USAGE_DAYS),
  });

  const { data: executionRateDataRes, isLoading: executionRateLoading } = useQuery({
    queryKey: ['dashboard', 'execution-rate', 168],
    queryFn: () => dashboardApi.getExecutionRate(168),
  });

  // Resource counts
  const { data: functionsData } = useQuery({
    queryKey: ['functions'],
    queryFn: () => functionsApi.list(),
  });

  const { data: providersData } = useQuery({
    queryKey: ['providers'],
    queryFn: () => providersApi.getConnectedProviders(),
  });

  const { data: appsData } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appsApi.list();
      return res?.apps ?? [];
    },
  });

  // State Fabric data
  const { data: stateFabricsList } = useQuery({
    queryKey: ['state-fabrics'],
    queryFn: () => stateFabricApi.list(),
  });

  const fabricIds = useMemo(() => (stateFabricsList ?? []).map((f) => f.id), [stateFabricsList]);

  const { data: stateFabricMetricsMap } = useQuery({
    queryKey: ['state-fabrics-metrics', fabricIds],
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

  // Agents data
  const { data: agentsListRes } = useQuery({
    queryKey: ['agents-list'],
    queryFn: () => agentApi.listAgents({ limit: MAX_AGENTS_FOR_USAGE }),
  });

  const agentIds = useMemo(() => (agentsListRes?.agents ?? []).map((a) => a.agentId), [agentsListRes]);

  const { data: agentsUsageAndBalance } = useQuery({
    queryKey: ['agents-usage-balance', agentIds],
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

  // Cost analytics
  const { data: costSummary, isLoading: costSummaryLoading } = useQuery({
    queryKey: ['cost-analytics', 'summary', dateRange],
    queryFn: () => getCostSummary(dateRangeParams.start, dateRangeParams.end),
  });

  const { data: functionCosts, isLoading: functionCostsLoading } = useQuery({
    queryKey: ['cost-analytics', 'functions', dateRange],
    queryFn: () => getCostByFunction(dateRangeParams.start, dateRangeParams.end),
  });

  const { data: periodData, isLoading: periodLoading } = useQuery({
    queryKey: ['cost-analytics', 'period', dateRange],
    queryFn: () => getCostByPeriod(dateRangeParams.start, dateRangeParams.end),
  });

  const { data: regionData, isLoading: regionLoading } = useQuery({
    queryKey: ['cost-analytics', 'region', dateRange],
    queryFn: () => getCostByRegion(dateRangeParams.start, dateRangeParams.end),
  });

  const { data: forecast, isLoading: forecastLoading } = useQuery({
    queryKey: ['cost-analytics', 'forecast'],
    queryFn: () => getUsageForecast(),
  });

  const { data: spendCap, isLoading: spendCapLoading } = useQuery({
    queryKey: ['cost-analytics', 'spend-cap'],
    queryFn: () => getSpendCap(),
  });

  // Derived counts
  const functionsCount = (functionsData?.functions ?? []).length;
  const providersCount = (providersData ?? []).length;
  const appsCount = Array.isArray(appsData) ? appsData.length : 0;

  // Limits
  const functionsLimit = limits?.functions ?? 0;
  const providersLimit = limits?.providers ?? 0;
  const stateFabricsLimit = (limits as { stateFabrics?: number })?.stateFabrics ?? 0;
  const agentsLimit = (limits as { agents?: number })?.agents ?? 0;

  // State fabric totals
  const stateFabricTotals = useMemo(() => {
    return Object.values(stateFabricMetricsMap ?? {}).reduce(
      (acc, m) => ({
        operations: acc.operations + (m.totalOperations ?? 0),
        storage: acc.storage + (m.storageUsed ?? 0),
      }),
      { operations: 0, storage: 0 }
    );
  }, [stateFabricMetricsMap]);

  const isLoading = usageLoading || costSummaryLoading || functionCostsLoading || periodLoading || regionLoading || forecastLoading || spendCapLoading;

  return {
    // User
    user,
    meData,
    displayPlan,
    username,
    limits,

    // Data
    usageData,
    executionRateDataRes,
    functionsData,
    providersData,
    appsData,
    stateFabricsList,
    agentsListRes,
    agentsUsageAndBalance,
    costSummary,
    functionCosts,
    periodData,
    regionData,
    forecast,
    spendCap,

    // Counts
    functionsCount,
    providersCount,
    appsCount,
    fabricIds,
    agentIds,

    // Limits
    functionsLimit,
    providersLimit,
    stateFabricsLimit,
    agentsLimit,
    stateFabricTotals,

    // Loading
    isLoading,
    usageLoading,
    executionRateLoading,
    costSummaryLoading,
    functionCostsLoading,
    periodLoading,
    regionLoading,
    forecastLoading,
    spendCapLoading,
  };
}

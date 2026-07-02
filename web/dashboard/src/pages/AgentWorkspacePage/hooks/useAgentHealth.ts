import { agentApi } from '@/api/agent';
import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';

interface HealthData {
  status: 'healthy' | 'degraded' | 'critical';
  health_score: number;
  anomalies: Array<{ type: string; severity: string; description: string; timestamp: string }>;
  children: number | Array<{ agent_id: string; status: string; health_score: number }>;
}

interface ConcurrencyData {
  active_executions: number;
  max_concurrent: number;
  queued: number;
  avg_wait_ms: number;
}

export function useAgentHealth(agentId: string) {
  const healthQuery = useQuery({
    queryKey: ['agent-health', agentId],
    queryFn: () => agentApi.checkSwarmHealth(agentId) as Promise<HealthData>,
    refetchInterval: 10000,
    enabled: !!agentId,
  });

  const concurrencyQuery = useQuery({
    queryKey: ['agent-concurrency', agentId],
    queryFn: () => agentApi.getConcurrencyStats(),
    refetchInterval: 5000,
    enabled: !!agentId,
    select: (data: any) => data?.stats as ConcurrencyData,
  });

  return {
    health: healthQuery.data as HealthData | undefined,
    healthLoading: healthQuery.isLoading,
    healthError: healthQuery.error,
    concurrency: concurrencyQuery.data as ConcurrencyData | undefined,
    concurrencyLoading: concurrencyQuery.isLoading,
    refetch: () => {
      healthQuery.refetch();
      concurrencyQuery.refetch();
    },
  };
}

export function useCostRate(agentId: string) {
  const [rate, setRate] = useState<number>(0);
  const lastCostRef = useRef<number>(0);
  const lastTimeRef = useRef<number>(Date.now());

  const { data } = useQuery({
    queryKey: ['agent-billing', agentId],
    queryFn: () => agentApi.getBillingSummary(agentId),
    refetchInterval: 30000,
    enabled: !!agentId,
  });

  useEffect(() => {
    if (data?.summary) {
      const now = Date.now();
      const elapsed = (now - lastTimeRef.current) / 60000; // minutes
      const summary = data.summary as any;
      if (elapsed > 0 && lastCostRef.current > 0) {
        const delta = (summary.total_spend_usd ?? summary.spendThisPeriod ?? 0) - lastCostRef.current;
        setRate(delta / elapsed);
      }
      lastCostRef.current = summary.total_spend_usd ?? summary.spendThisPeriod ?? 0;
      lastTimeRef.current = now;
    }
  }, [data]);

  return rate;
}

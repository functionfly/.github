import { useQuery } from '@tanstack/react-query';
import { dashboardApi, type SystemHealthStatus } from '@/api/dashboard';

// Query keys
export const dashboardKeys = {
  all: ['dashboard'] as const,
  usage: (days?: number) => [...dashboardKeys.all, 'usage', days] as const,
  executionRate: (hours?: number) => [...dashboardKeys.all, 'execution-rate', hours] as const,
  activity: (limit?: number) => [...dashboardKeys.all, 'activity', limit] as const,
  memory: () => [...dashboardKeys.all, 'memory'] as const,
  metrics: () => [...dashboardKeys.all, 'metrics'] as const,
  health: () => [...dashboardKeys.all, 'health'] as const,
};

// Get usage data
export function useDashboardUsage(days = 14) {
  return useQuery({
    queryKey: dashboardKeys.usage(days),
    queryFn: () => dashboardApi.getUsage(days),
    staleTime: 1000 * 60,
  });
}

// Get execution rate
export function useDashboardExecutionRate(hours = 24) {
  return useQuery({
    queryKey: dashboardKeys.executionRate(hours),
    queryFn: () => dashboardApi.getExecutionRate(hours),
    staleTime: 1000 * 30,
    refetchInterval: 60000, // Refresh every minute
  });
}

// Get activity feed
export function useDashboardActivity(limit = 20) {
  return useQuery({
    queryKey: dashboardKeys.activity(limit),
    queryFn: () => dashboardApi.getActivity(limit),
    staleTime: 1000 * 30,
    refetchInterval: 30000, // Refresh every 30s
  });
}

// Get memory usage
export function useDashboardMemory() {
  return useQuery({
    queryKey: dashboardKeys.memory(),
    queryFn: () => dashboardApi.getMemoryUsage(),
    staleTime: 1000 * 30,
    refetchInterval: 30000,
  });
}

// Get dashboard metrics
export function useDashboardMetrics() {
  return useQuery({
    queryKey: dashboardKeys.metrics(),
    queryFn: () => dashboardApi.getMetrics(),
    staleTime: 1000 * 60,
    refetchInterval: 60000,
  });
}

// Get health status
export function useDashboardHealthStatus() {
  return useQuery<SystemHealthStatus>({
    queryKey: dashboardKeys.health(),
    queryFn: () => dashboardApi.getHealthStatus(),
    staleTime: 1000 * 30,
    refetchInterval: 30000,
  });
}

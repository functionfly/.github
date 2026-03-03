import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { useEffect, useCallback } from 'react';
import {
  statusApi,
  type PlatformStatus,
  type ComponentHealth,
  type ProviderStatus,
  type UptimeMetrics,
  type UptimePeriod,
  type LatencyMetrics,
  type ProviderType,
  type MaintenanceWindow,
  type Incident,
  type CreateIncidentRequest,
  type UpdateIncidentRequest,
  type GetIncidentsParams,
} from '@/api/status';

// Query keys for status-related queries
export const statusKeys = {
  all: ['status'] as const,
  platform: () => [...statusKeys.all, 'platform'] as const,
  components: () => [...statusKeys.all, 'components'] as const,
  providers: () => [...statusKeys.all, 'providers'] as const,
  uptime: (days: UptimePeriod) => [...statusKeys.all, 'uptime', days] as const,
  latency: (provider: ProviderType, region?: string) =>
    [...statusKeys.all, 'latency', provider, region] as const,
  maintenance: () => [...statusKeys.all, 'maintenance'] as const,
};

// Query keys for incidents
export const incidentKeys = {
  all: ['incidents'] as const,
  lists: (params: GetIncidentsParams) => [...incidentKeys.all, 'list', params] as const,
  detail: (id: string) => [...incidentKeys.all, 'detail', id] as const,
};

// Polling interval in milliseconds (30 seconds)
const POLLING_INTERVAL = 30000;

/**
 * Hook to fetch and manage platform status
 * Includes automatic polling fallback for real-time updates
 */
export function useStatus(options?: { enabled?: boolean }) {
  const queryClient = useQueryClient();

  const {
    data: platformStatus,
    isLoading: isPlatformLoading,
    error: platformError,
    refetch: refetchPlatform,
  } = useQuery<PlatformStatus>({
    queryKey: statusKeys.platform(),
    queryFn: statusApi.getPlatformStatus,
    staleTime: 10000, // Consider data stale after 10 seconds
    refetchInterval: POLLING_INTERVAL,
    retry: 2,
    enabled: options?.enabled !== false,
  });

  const {
    data: components,
    isLoading: isComponentsLoading,
    error: componentsError,
    refetch: refetchComponents,
  } = useQuery<ComponentHealth[]>({
    queryKey: statusKeys.components(),
    queryFn: statusApi.getComponents,
    staleTime: 10000,
    refetchInterval: POLLING_INTERVAL,
    retry: 2,
    enabled: options?.enabled !== false,
  });

  const {
    data: providers,
    isLoading: isProvidersLoading,
    error: providersError,
    refetch: refetchProviders,
  } = useQuery<ProviderStatus[]>({
    queryKey: statusKeys.providers(),
    queryFn: statusApi.getProviders,
    staleTime: 10000,
    refetchInterval: POLLING_INTERVAL,
    retry: 2,
    enabled: options?.enabled !== false,
  });

  // Combined loading state
  const isLoading = isPlatformLoading || isComponentsLoading || isProvidersLoading;

  // Combined error state
  const error = platformError || componentsError || providersError;

  // Refetch all status data
  const refetch = useCallback(async () => {
    await Promise.all([
      refetchPlatform(),
      refetchComponents(),
      refetchProviders(),
    ]);
  }, [refetchPlatform, refetchComponents, refetchProviders]);

  // Invalidate and refetch all status queries
  const invalidateStatus = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: statusKeys.all });
  }, [queryClient]);

  return {
    // Data
    platformStatus,
    components,
    providers,

    // Loading states
    isLoading,
    isPlatformLoading,
    isComponentsLoading,
    isProvidersLoading,

    // Error states
    error,
    platformError,
    componentsError,
    providersError,

    // Actions
    refetch,
    invalidateStatus,
  };
}

/**
 * Hook to fetch uptime metrics
 */
export function useUptimeMetrics(days: UptimePeriod = 30) {
  return useQuery<UptimeMetrics>({
    queryKey: statusKeys.uptime(days),
    queryFn: () => statusApi.getUptimeMetrics(days),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
}

/**
 * Hook to fetch latency metrics for a provider
 */
export function useLatencyMetrics(provider: ProviderType, region?: string) {
  return useQuery<LatencyMetrics>({
    queryKey: statusKeys.latency(provider, region),
    queryFn: () => statusApi.getLatencyMetrics(provider, region),
    staleTime: 60000, // 1 minute
    retry: 2,
    enabled: !!provider,
  });
}

/**
 * Hook to fetch scheduled maintenance windows
 */
export function useMaintenance() {
  return useQuery<MaintenanceWindow[]>({
    queryKey: statusKeys.maintenance(),
    queryFn: statusApi.getMaintenance,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
}

/**
 * Hook to get status color based on status type
 */
export function useStatusColors() {
  return {
    operational: {
      bg: 'bg-emerald-500',
      text: 'text-emerald-400',
      border: 'border-emerald-500/30',
      glow: 'shadow-emerald-500/20',
      gradient: 'from-emerald-500 to-teal-500',
    },
    degraded: {
      bg: 'bg-amber-500',
      text: 'text-amber-400',
      border: 'border-amber-500/30',
      glow: 'shadow-amber-500/20',
      gradient: 'from-amber-500 to-orange-500',
    },
    down: {
      bg: 'bg-red-500',
      text: 'text-red-400',
      border: 'border-red-500/30',
      glow: 'shadow-red-500/20',
      gradient: 'from-red-500 to-rose-500',
    },
    major_outage: {
      bg: 'bg-red-600',
      text: 'text-red-500',
      border: 'border-red-600/30',
      glow: 'shadow-red-600/30',
      gradient: 'from-red-600 to-rose-600',
    },
  };
}

/**
 * Hook to get severity colors for incidents
 */
export function useSeverityColors() {
  return {
    critical: {
      bg: 'bg-red-500',
      text: 'text-red-400',
      border: 'border-red-500/30',
      label: 'Critical',
    },
    high: {
      bg: 'bg-orange-500',
      text: 'text-orange-400',
      border: 'border-orange-500/30',
      label: 'High',
    },
    medium: {
      bg: 'bg-amber-500',
      text: 'text-amber-400',
      border: 'border-amber-500/30',
      label: 'Medium',
    },
    low: {
      bg: 'bg-blue-500',
      text: 'text-blue-400',
      border: 'border-blue-500/30',
      label: 'Low',
    },
  };
}

/**
 * Hook to get status label
 */
export function useStatusLabels() {
  return {
    operational: 'Operational',
    degraded: 'Degraded Performance',
    down: 'Service Disruption',
    major_outage: 'Major Outage',
  };
}

// ============================================================================
// Incident Hooks
// ============================================================================

/**
 * Hook to fetch incidents with filtering
 */
export function useIncidents(params: GetIncidentsParams = {}) {
  return useQuery<{
    incidents: Incident[];
    total: number;
    limit: number;
    offset: number;
  }>({
    queryKey: incidentKeys.lists(params),
    queryFn: () => statusApi.getIncidents(params),
    staleTime: 30000, // 30 seconds
    retry: 2,
  });
}

/**
 * Hook to fetch a single incident
 */
export function useIncident(id: string) {
  return useQuery<Incident>({
    queryKey: incidentKeys.detail(id),
    queryFn: () => statusApi.getIncident(id),
    staleTime: 30000,
    retry: 2,
    enabled: !!id,
  });
}

/**
 * Hook to create a new incident (admin only)
 */
export function useCreateIncident() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateIncidentRequest) => statusApi.createIncident(data),
    onSuccess: () => {
      // Invalidate incidents list
      queryClient.invalidateQueries({ queryKey: incidentKeys.all });
    },
  });
}

/**
 * Hook to update an incident (admin only)
 */
export function useUpdateIncident() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateIncidentRequest }) =>
      statusApi.updateIncident(id, data),
    onSuccess: (_, { id }) => {
      // Invalidate specific incident and list
      queryClient.invalidateQueries({ queryKey: incidentKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: incidentKeys.all });
    },
  });
}

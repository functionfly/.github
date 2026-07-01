import { useEffect, useMemo, useState } from 'react';
import { useNotificationStore } from '@/stores/notificationStore';
import { getAppsLimit } from '@/lib/plan-utils';
import { useAuthStore } from '@/stores/authStore';
import { apiClient } from '@/api/client';
import { API_URLS } from '@/lib/api-urls';
import { appsApi } from '@/api/apps';
import { agentApi } from '@/api/agent';
import { vaultApi } from '@/api/vault';
import { functionsApi } from '@/api/functions';
import { providersApi } from '@/api';
import { useQuery } from '@tanstack/react-query';

interface NavigationStatus {
  functions: {
    hasIssues: boolean;
    pendingDeployments: number;
    totalCount: number;
  };
  providers: {
    hasOffline: boolean;
    totalCount: number;
  };
  analytics: {
    hasAlerts: boolean;
    alertCount: number;
  };
  settings: {
    hasWarnings: boolean;
    warningCount: number;
  };
  wallet: {
    balance: number;
    currency: string;
    isLowBalance: boolean;
    lowBalanceThreshold: number;
  };
  agents: {
    totalCount: number;
    activeCount: number;
    hasOffline: boolean;
  };
  apps: {
    totalCount: number;
    deployedCount: number;
  };
  secrets: {
    totalCount: number;
  };
  teams: {
    totalCount: number;
    pendingInvites: number;
  };
}

const mockStatusData: NavigationStatus = {
  functions: {
    hasIssues: false,
    pendingDeployments: 0,
    totalCount: 0,
  },
  providers: {
    hasOffline: false,
    totalCount: 0,
  },
  analytics: {
    hasAlerts: false,
    alertCount: 0,
  },
  settings: {
    hasWarnings: false,
    warningCount: 0,
  },
  wallet: {
    balance: 0,
    currency: 'USD',
    isLowBalance: false,
    lowBalanceThreshold: 50.0,
  },
  agents: {
    totalCount: 0,
    activeCount: 0,
    hasOffline: false,
  },
  apps: {
    totalCount: 0,
    deployedCount: 0,
  },
  secrets: {
    totalCount: 0,
  },
  teams: {
    totalCount: 0,
    pendingInvites: 0,
  },
};

export function useNavigationStatus(): NavigationStatus {
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);
  const user = useAuthStore((state) => state.user);
  const plan = user?.plan;

  const appsLimit = getAppsLimit(plan);

  const { data: appsData } = useQuery({
    queryKey: ['apps', 'sidebar-count'],
    queryFn: () => appsApi.list(),
    staleTime: 1000 * 60,
    retry: false,
  });

  const { data: agentsData } = useQuery({
    queryKey: ['agents', 'sidebar-count'],
    queryFn: () => agentApi.listAgents(),
    staleTime: 1000 * 60,
    retry: false,
  });

  const { data: secretsData } = useQuery({
    queryKey: ['secrets', 'sidebar-count'],
    queryFn: () => vaultApi.listSecrets(),
    staleTime: 1000 * 60,
    retry: false,
  });

  const { data: functionsData } = useQuery({
    queryKey: ['functions', 'sidebar-count'],
    queryFn: () => functionsApi.list(),
    staleTime: 1000 * 60,
    retry: false,
  });

  const { data: providersData } = useQuery({
    queryKey: ['providers', 'sidebar-count'],
    queryFn: () => providersApi.getConnectedProviders(),
    staleTime: 1000 * 60,
    retry: false,
  });

  const appCount = appsData?.apps?.length ?? 0;
  const agentList = agentsData?.agents ?? [];
  const agentCount = agentList.length;
  const agentActiveCount = agentList.filter((a) => a.status === 'active').length;
  const agentHasOffline = agentList.some((a) => a.status === 'suspended' || a.status === 'deleted');
  const secretCount = secretsData?.total ?? 0;
  const functionList = functionsData?.functions ?? [];
  const functionCount = functionList.length;
  const pendingDeployments = functionList.filter((f) => f.status === 'deploying').length;
  const functionHasIssues = functionList.some((f) => f.status === 'failed');
  const providerList = providersData ?? [];
  const providerCount = providerList.length;
  const providerHasOffline = providerList.some((p) => p.status === 'offline' || p.status === 'error');

  return useMemo(() => {
    return {
      ...mockStatusData,
      functions: {
        hasIssues: functionHasIssues,
        pendingDeployments,
        totalCount: functionCount,
      },
      providers: {
        hasOffline: providerHasOffline,
        totalCount: providerCount,
      },
      agents: {
        totalCount: agentCount,
        activeCount: agentActiveCount,
        hasOffline: agentHasOffline,
      },
      apps: {
        totalCount: appsLimit,
        deployedCount: appCount,
      },
      secrets: {
        totalCount: secretCount,
      },
    };
  }, [unreadCount, appsLimit, appCount, agentCount, agentActiveCount, agentHasOffline, secretCount, functionCount, pendingDeployments, functionHasIssues, providerCount, providerHasOffline]);
}

// Helper hook for getting specific status badge content
export function useStatusBadge(path: string): { content: string | number | null; type: 'info' | 'warning' | 'error' | 'success' | null } {
  const status = useNavigationStatus();
  const [unvotedCount, setUnvotedCount] = useState(0);

  useEffect(() => {
    if (path !== '/founders') return;
    let cancelled = false;
    apiClient
      .get<{ votes: { has_voted: boolean }[] }>(API_URLS.founders.votes)
      .then((res) => {
        if (!cancelled && res?.votes) {
          setUnvotedCount(res.votes.filter((v) => !v.has_voted).length);
        }
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [path]);

  return useMemo(() => {
    switch (path) {
      case '/notifications':
        return status.analytics.hasAlerts
          ? { content: status.analytics.alertCount, type: 'warning' }
          : { content: null, type: null };
      case '/wallet':
        if (status.wallet.isLowBalance) {
          return { content: '$' + status.wallet.balance.toFixed(0), type: 'warning' };
        }
        return { content: '$' + status.wallet.balance.toFixed(0), type: 'info' };
      case '/agents':
        if (status.agents.hasOffline) {
          return { content: status.agents.activeCount + '/' + status.agents.totalCount, type: 'warning' };
        }
        return { content: status.agents.activeCount, type: 'success' };
      case '/apps':
        if (status.apps.totalCount < 0) {
          return { content: status.apps.deployedCount, type: 'info' };
        }
        return { content: status.apps.deployedCount + '/' + status.apps.totalCount, type: 'info' };
      case '/secrets':
        return { content: status.secrets.totalCount, type: 'info' };
      case '/teams':
        if (status.teams.pendingInvites > 0) {
          return { content: status.teams.pendingInvites, type: 'warning' };
        }
        return { content: status.teams.totalCount, type: 'info' };
      case '/functions/my':
        if (status.functions.pendingDeployments > 0) {
          return { content: status.functions.pendingDeployments, type: 'warning' };
        }
        return { content: status.functions.totalCount, type: 'info' };
      case '/settings':
        if (status.settings.hasWarnings) {
          return { content: status.settings.warningCount, type: 'warning' };
        }
        return { content: null, type: null };
      case '/providers':
        if (status.providers.hasOffline) {
          return { content: '!', type: 'error' };
        }
        return { content: status.providers.totalCount, type: 'success' };
      case '/founders':
        if (unvotedCount > 0) {
          return { content: unvotedCount, type: 'warning' };
        }
        return { content: null, type: null };
      default:
        return { content: null, type: null };
    }
  }, [status, path, unvotedCount]);
}

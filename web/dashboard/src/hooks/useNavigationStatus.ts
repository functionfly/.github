import { useMemo } from 'react';
import { useNotificationStore } from '@/stores/notificationStore';
import { getAppsLimit } from '@/lib/plan-utils';
import { useAuthStore } from '@/stores/authStore';

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

// Default values - apps.totalCount will be overridden in useNavigationStatus
// based on the user's plan
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

  // Apps limit from plan (unlimited = -1)
  const appsLimit = getAppsLimit(plan);

  // In a real app, this would fetch from API with polling/websockets
  // For now, return mock data enhanced with notifications
  return useMemo(() => {
    return {
      ...mockStatusData,
      apps: {
        totalCount: appsLimit,
        deployedCount: 0,
      },
      // Could incorporate unreadCount into relevant sections
    };
  }, [unreadCount, appsLimit]);
}

// Helper hook for getting specific status badge content
export function useStatusBadge(path: string): { content: string | number | null; type: 'info' | 'warning' | 'error' | 'success' | null } {
  const status = useNavigationStatus();

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
      default:
        return { content: null, type: null };
    }
  }, [status, path]);
}

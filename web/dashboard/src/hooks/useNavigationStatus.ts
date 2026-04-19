import { useMemo } from 'react';
import { useNotificationStore } from '@/stores/notificationStore';

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

// Enhanced mock data - in real app this would come from API/websockets
// This provides realistic data for badge displays
const mockStatusData: NavigationStatus = {
  functions: {
    hasIssues: true,
    pendingDeployments: 2,
    totalCount: 12,
  },
  providers: {
    hasOffline: false,
    totalCount: 3,
  },
  analytics: {
    hasAlerts: true,
    alertCount: 3,
  },
  settings: {
    hasWarnings: true,
    warningCount: 1,
  },
  wallet: {
    balance: 47.50,
    currency: 'USD',
    isLowBalance: true,
    lowBalanceThreshold: 50.00,
  },
  agents: {
    totalCount: 5,
    activeCount: 4,
    hasOffline: true,
  },
  apps: {
    totalCount: 3,
    deployedCount: 2,
  },
  secrets: {
    totalCount: 8,
  },
  teams: {
    totalCount: 4,
    pendingInvites: 2,
  },
};

export function useNavigationStatus(): NavigationStatus {
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);

  // In a real app, this would fetch from API with polling/websockets
  // For now, return mock data enhanced with notifications
  return useMemo(() => {
    return {
      ...mockStatusData,
      // Could incorporate unreadCount into relevant sections
    };
  }, [unreadCount]);
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

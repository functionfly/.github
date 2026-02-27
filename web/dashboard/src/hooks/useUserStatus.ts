import { useState } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';
import { UserStatusChangeEvent } from './types';

// Hook for tenant-wide user status changes
export function useTenantUserStatus() {
  const { user } = useAuthStore();
  const [statusChanges, setStatusChanges] = useState<UserStatusChangeEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<UserStatusChangeEvent>(
    `tenant_${user?.tenantId}_users`,
    'user_status_change',
    (event) => {
      setStatusChanges(prev => [event, ...prev]);
    }
  );

  return {
    statusChanges,
    isConnected,
  };
}
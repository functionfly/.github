import { useEffect, useState, useCallback, useRef } from 'react';
import { supabase, getUsersByIds, getUserLastActive } from '../lib/neon';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription';

// Generic hook for real-time subscriptions with callback pattern
export function useRealtime() {
  const subscriptionsRef = useRef<Map<string, Set<Function>>>(new Map());

  const subscribe = useCallback((channel: string, callback: Function) => {
    if (!subscriptionsRef.current.has(channel)) {
      subscriptionsRef.current.set(channel, new Set());
    }
    subscriptionsRef.current.get(channel)?.add(callback);
  }, []);

  const unsubscribe = useCallback((channel: string, callback: Function) => {
    const channelSubs = subscriptionsRef.current.get(channel);
    if (channelSubs) {
      channelSubs.delete(callback);
      if (channelSubs.size === 0) {
        subscriptionsRef.current.delete(channel);
      }
    }
  }, []);

  // Set up real-time subscription for all channels
  useRealtimeSubscription<RealtimeEvent>(
    'deployments',
    'deployment_update',
    (event) => {
      const channelSubs = subscriptionsRef.current.get('deployments');
      if (channelSubs) {
        channelSubs.forEach(callback => {
          try {
            callback(event);
          } catch (error) {
            console.error('Error in deployment subscription callback:', error);
          }
        });
      }
    }
  );

  return {
    subscribe,
    unsubscribe,
  };
}

// Types for real-time events
export interface RealtimeEvent {
  type: string;
  table: string;
  record_id?: string;
  tenant_id?: string;
  user_id?: string;
  data?: any;
  timestamp: string;
}

export interface UserStatusChangeEvent extends RealtimeEvent {
  type: 'user_status_change';
  old_status: 'verified' | 'unverified';
  new_status: 'verified' | 'unverified';
}

export interface ProfileUpdateEvent extends RealtimeEvent {
  type: 'profile_update';
  changes: {
    first_name?: boolean;
    last_name?: boolean;
    avatar_url?: boolean;
    bio?: boolean;
  };
}

export interface NewNotificationEvent extends RealtimeEvent {
  type: 'new_notification';
  notification_id: string;
  notification_type: 'info' | 'warning' | 'error' | 'success';
  title: string;
}

export interface PresenceEvent extends RealtimeEvent {
  type: 'presence_join' | 'presence_leave';
  key: string;
  current_presences: any[];
  new_presences?: any[];
  left_presences?: any[];
}

// Hook for user-specific notifications
export function useUserNotifications() {
  const { user } = useAuthStore();
  const [notifications, setNotifications] = useState<NewNotificationEvent[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  const { isConnected } = useRealtimeSubscription<NewNotificationEvent>(
    `user_${user?.id}_notifications`,
    'new_notification',
    (event) => {
      setNotifications(prev => [event, ...prev]);
      setUnreadCount(prev => prev + 1);
    }
  );

  const markAsRead = useCallback(async (notificationId: string) => {
    try {
      const { error } = await supabase
        .from('user_notifications')
        .update({ read_at: new Date().toISOString() })
        .eq('id', notificationId)
        .eq('user_id', user?.id);

      if (error) throw error;

      setNotifications(prev =>
        prev.map(n =>
          n.notification_id === notificationId
            ? { ...n, read_at: new Date().toISOString() }
            : n
        )
      );
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  }, [user?.id]);

  const markAllAsRead = useCallback(async () => {
    try {
      const { error } = await supabase
        .from('user_notifications')
        .update({ read_at: new Date().toISOString() })
        .eq('user_id', user?.id)
        .is('read_at', null);

      if (error) throw error;

      setNotifications(prev =>
        prev.map(n => ({ ...n, read_at: new Date().toISOString() }))
      );
      setUnreadCount(0);
    } catch (error) {
      console.error('Error marking all notifications as read:', error);
    }
  }, [user?.id]);

  return {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
  };
}

// Hook for profile updates
export function useProfileUpdates() {
  const { user } = useAuthStore();
  const [profileUpdates, setProfileUpdates] = useState<ProfileUpdateEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<ProfileUpdateEvent>(
    `user_${user?.id}_profile`,
    'profile_update',
    (event) => {
      setProfileUpdates(prev => [event, ...prev]);
    }
  );

  return {
    profileUpdates,
    isConnected,
  };
}

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

// Database change event interface
export interface DatabaseChangeEvent extends RealtimeEvent {
  schema: string;
  table: string;
  commit_timestamp: string;
  eventType: 'INSERT' | 'UPDATE' | 'DELETE';
  new: any | null;
  old: any | null;
  errors: string | null;
  ids: string[];
}

// RealtimePostgresChangesPayload compatible interface for Neon
export interface RealtimePostgresChangesPayload<T = any> {
  data: {
    schema: string;
    table: string;
    commit_timestamp: string;
    eventType: 'INSERT' | 'UPDATE' | 'DELETE';
    new: T | null;
    old: T | null;
    errors: string | null;
  };
  ids: string[];
}

// Hook for database changes using WebSocket-based notifications
export function useDatabaseChanges(table: string, filter?: string) {
  const [changes, setChanges] = useState<RealtimePostgresChangesPayload[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  // Use WebSocket-based real-time database change notifications
  const { isConnected: wsConnected } = useRealtimeSubscription<DatabaseChangeEvent>(
    `db_changes_${table}`,
    'db_change',
    (event) => {
      // Transform the database change event to match the expected payload format
      const payload: RealtimePostgresChangesPayload = {
        data: {
          schema: event.schema,
          table: event.table,
          commit_timestamp: event.commit_timestamp,
          eventType: event.eventType as 'INSERT' | 'UPDATE' | 'DELETE',
          new: event.new || null,
          old: event.old || null,
          errors: event.errors ? event.errors : null,
        },
        ids: event.ids || [],
      };

      setChanges(prev => [payload, ...prev.slice(0, 99)]); // Keep last 100 changes
    }
  );

  useEffect(() => {
    setIsConnected(wsConnected);
  }, [wsConnected]);

  return {
    changes,
    isConnected,
  };
}

// Custom hook for real-time activity feed
export function useActivityFeed() {
  const { user } = useAuthStore();
  const [activities, setActivities] = useState<RealtimeEvent[]>([]);

  // Combine multiple real-time sources
  useRealtimeSubscription(
    `user_${user?.id}_notifications`,
    'new_notification',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  useRealtimeSubscription(
    `user_${user?.id}_profile`,
    'profile_update',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  useRealtimeSubscription(
    `tenant_${user?.tenantId}_users`,
    'user_status_change',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  return activities.sort((a, b) =>
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );
}

// Utility hook for connection status - migrated to WebSocket
export function useRealtimeConnection() {
  // For now, return a simple connected status
  // In the future, this could track overall WebSocket connection health
  return {
    connectionStatus: 'connected' as const,
    reconnectAttempts: 0,
    isConnected: true,
  };
}

// Database monitoring types
export interface DatabaseHealth {
  status: 'healthy' | 'degraded' | 'unhealthy';
  connections: {
    active: number;
    idle: number;
    total: number;
    max: number;
  };
  performance: {
    avgQueryTime: number;
    slowQueries: number;
    throughput: number;
  };
  storage: {
    used: number;
    total: number;
    growthRate: number;
  };
  replication: {
    lag: number;
    status: 'healthy' | 'lagging' | 'error';
  };
  lastUpdated: string;
}

export interface DatabaseAlert {
  id: string;
  type: 'connection_pool_exhausted' | 'high_query_latency' | 'storage_warning' | 'replication_lag';
  severity: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  message: string;
  timestamp: string;
  resolved?: boolean;
}

export interface DatabaseMetric {
  timestamp: string;
  connections: number;
  queryCount: number;
  avgResponseTime: number;
  errorRate: number;
}

// Hook for real-time database health monitoring
export function useDatabaseHealth() {
  const [health, setHealth] = useState<DatabaseHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Subscribe to database health updates
  const { isConnected } = useRealtimeSubscription(
    'database_monitoring',
    'db_health_update',
    (event) => {
      if (event.type === 'db_health_update' && event.data) {
        setHealth(event.data);
        setError(null);
      }
    }
  );

  useEffect(() => {
    const fetchDatabaseHealth = async () => {
      try {
        setLoading(true);
        setError(null);

        // Call the database health API endpoint
        const response = await fetch('/v1/monitoring/database/health', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Transform API response to match our interface
        const transformedHealth: DatabaseHealth = {
          status: (data.status === 'healthy' || data.status === 'degraded' || data.status === 'unhealthy') ? data.status : 'healthy',
          connections: {
            active: data.connections?.active || 0,
            idle: data.connections?.idle || 0,
            total: data.connections?.total || 0,
            max: data.connections?.max || 100,
          },
          performance: {
            avgQueryTime: data.performance?.avgQueryTime || 0,
            slowQueries: data.performance?.slowQueries || 0,
            throughput: data.performance?.throughput || 0,
          },
          storage: {
            used: data.storage?.usedGB || 0,
            total: data.storage?.totalGB || 0,
            growthRate: data.storage?.growthRate || 0,
          },
          replication: {
            lag: data.replication?.lag || 0,
            status: (data.replication?.status === 'healthy' || data.replication?.status === 'lagging' || data.replication?.status === 'error') ? data.replication.status : 'healthy',
          },
          lastUpdated: data.lastUpdated || new Date().toISOString(),
        };

        setHealth(transformedHealth);
      } catch (err) {
        console.error('Error fetching database health:', err);
        setError(err instanceof Error ? err.message : 'Failed to load database health');

        // Fallback to mock data if API fails
        const mockHealth: DatabaseHealth = {
          status: 'healthy',
          connections: {
            active: 0,
            idle: 0,
            total: 0,
            max: 100
          },
          performance: {
            avgQueryTime: 0,
            slowQueries: 0,
            throughput: 0
          },
          storage: {
            used: 0,
            total: 0,
            growthRate: 0
          },
          replication: {
            lag: 0,
            status: 'healthy'
          },
          lastUpdated: new Date().toISOString()
        };
        setHealth(mockHealth);
      } finally {
        setLoading(false);
      }
    };

    fetchDatabaseHealth();

    // Set up periodic refresh (every 30 seconds)
    const interval = setInterval(fetchDatabaseHealth, 30000);

    return () => clearInterval(interval);
  }, []);

  return {
    health,
    loading,
    error,
    isRealtimeConnected: isConnected,
  };
}

// Hook for real-time database alerts
export function useDatabaseAlerts() {
  const [alerts, setAlerts] = useState<DatabaseAlert[]>([]);
  const [loading, setLoading] = useState(true);

  // Subscribe to new database alerts
  const { isConnected } = useRealtimeSubscription(
    'database_alerts',
    'db_alert_created',
    (event) => {
      if (event.type === 'db_alert_created' && event.data) {
        setAlerts(prev => [event.data, ...prev.slice(0, 9)]); // Keep last 10 alerts
      }
    }
  );

  // Subscribe to alert resolutions
  useRealtimeSubscription(
    'database_alerts',
    'db_alert_resolved',
    (event) => {
      if (event.type === 'db_alert_resolved' && event.data) {
        setAlerts(prev => prev.map(alert =>
          alert.id === event.data.id
            ? { ...alert, resolved: true }
            : alert
        ));
      }
    }
  );

  useEffect(() => {
    const fetchDatabaseAlerts = async () => {
      try {
        setLoading(true);

        // Call the database alerts API endpoint
        const response = await fetch('/v1/monitoring/database/alerts', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Transform API response to match our interface
        const transformedAlerts: DatabaseAlert[] = data.map((alert: any) => ({
          id: alert.id || 'unknown',
          type: alert.type || 'unknown',
          severity: alert.severity || 'low',
          title: alert.title || 'Unknown Alert',
          message: alert.message || '',
          timestamp: alert.timestamp || new Date().toISOString(),
          resolved: alert.resolved || false,
        }));

        setAlerts(transformedAlerts);
      } catch (err) {
        console.error('Error fetching database alerts:', err);

        // Fallback to empty array if API fails
        setAlerts([]);
      } finally {
        setLoading(false);
      }
    };

    fetchDatabaseAlerts();

    // Set up periodic refresh (every 60 seconds for alerts)
    const interval = setInterval(fetchDatabaseAlerts, 60000);

    return () => clearInterval(interval);
  }, []);

  return {
    alerts,
    loading,
    isRealtimeConnected: isConnected,
  };
}

// Hook for database performance metrics
export function useDatabaseMetrics(timeRange: '1h' | '6h' | '24h' | '7d' = '1h') {
  const [metrics, setMetrics] = useState<DatabaseMetric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDatabaseMetrics = async () => {
      try {
        setLoading(true);
        setError(null);

        // Call the database metrics API endpoint
        const response = await fetch(`/v1/monitoring/database/metrics?range=${timeRange}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Transform API response to match our interface
        const transformedMetrics: DatabaseMetric[] = data.map((metric: any) => ({
          timestamp: metric.timestamp || new Date().toISOString(),
          connections: metric.connections || 0,
          queryCount: metric.queryCount || 0,
          avgResponseTime: metric.avgResponseTime || 0,
          errorRate: metric.errorRate || 0,
        }));

        setMetrics(transformedMetrics);
      } catch (err) {
        console.error('Error fetching database metrics:', err);
        setError(err instanceof Error ? err.message : 'Failed to load database metrics');

        // Fallback to empty array if API fails
        setMetrics([]);
      } finally {
        setLoading(false);
      }
    };

    fetchDatabaseMetrics();

    // Set up periodic refresh based on time range
    const refreshInterval = timeRange === '1h' ? 30000 : timeRange === '6h' ? 60000 : 120000; // 30s, 1m, 2m
    const interval = setInterval(fetchDatabaseMetrics, refreshInterval);

    return () => clearInterval(interval);
  }, [timeRange]);

  return {
    metrics,
    loading,
    error,
  };
}

// Deployment event types
export interface DeploymentEvent extends RealtimeEvent {
  type: 'broadcast';
  event: 'deployment_update';
  deployment_id: string;
  status: string;
  details: any;
}

// Function execution event types
export interface FunctionExecutionEvent extends RealtimeEvent {
  type: 'broadcast';
  event: 'function_execution';
  function_id: string;
  execution_id: string;
  event_type: 'started' | 'completed' | 'failed' | 'log';
  details: any;
}

// Team event types
export interface TeamEvent extends RealtimeEvent {
  type: 'broadcast';
  event: 'team_update';
  event_type: 'member_added' | 'member_removed' | 'role_changed' | 'permissions_updated';
  details: any;
}

// Registry event types
export interface RegistryEvent extends RealtimeEvent {
  type: 'broadcast';
  event: 'registry_update';
  function_id: string;
  update_type: 'rating' | 'popularity' | 'download' | 'new_version';
  details: any;
}

// Hook for real-time deployment updates
export function useDeploymentUpdates(appId?: string) {
  const { user } = useAuthStore();
  const [deployments, setDeployments] = useState<DeploymentEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<DeploymentEvent>(
    appId ? `app_${appId}_deployments` : 'deployments',
    'deployment_update',
    (event) => {
      // Only show deployments for current user's tenant if no specific app
      if (!appId && user?.tenantId && event.details?.tenant_id !== user.tenantId) {
        return;
      }

      setDeployments(prev => [event, ...prev.slice(0, 49)]); // Keep last 50 events
    }
  );

  return {
    deployments,
    isConnected,
  };
}

// Hook for real-time function execution updates
export function useFunctionExecutionUpdates(functionId?: string) {
  const { user } = useAuthStore();
  const [executions, setExecutions] = useState<FunctionExecutionEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<FunctionExecutionEvent>(
    functionId ? `function_${functionId}_executions` : 'function_executions',
    'function_execution',
    (event) => {
      // Only show executions for current user's tenant if no specific function
      if (!functionId && user?.tenantId && event.details?.tenant_id !== user.tenantId) {
        return;
      }

      setExecutions(prev => [event, ...prev.slice(0, 99)]); // Keep last 100 events
    }
  );

  return {
    executions,
    isConnected,
  };
}

// Hook for real-time team updates
export function useTeamUpdates(teamId?: string) {
  const { user } = useAuthStore();
  const [teamEvents, setTeamEvents] = useState<TeamEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<TeamEvent>(
    teamId ? `team_${teamId}` : `tenant_${user?.tenantId}_team`,
    'team_update',
    (event) => {
      setTeamEvents(prev => [event, ...prev.slice(0, 49)]); // Keep last 50 events
    }
  );

  return {
    teamEvents,
    isConnected,
  };
}

// Hook for real-time registry updates
export function useRegistryUpdates() {
  const [registryEvents, setRegistryEvents] = useState<RegistryEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<RegistryEvent>(
    'registry_updates',
    'registry_update',
    (event) => {
      setRegistryEvents(prev => [event, ...prev.slice(0, 49)]); // Keep last 50 events
    }
  );

  return {
    registryEvents,
    isConnected,
  };
}

// Hook for combined activity feed including new real-time events
export function useEnhancedActivityFeed() {
  const { user } = useAuthStore();
  const [activities, setActivities] = useState<any[]>([]);

  // Existing subscriptions
  useRealtimeSubscription(
    `user_${user?.id}_notifications`,
    'new_notification',
    (event) => {
      setActivities(prev => [{ ...event, category: 'notification' }, ...prev]);
    }
  );

  useRealtimeSubscription(
    `user_${user?.id}_profile`,
    'profile_update',
    (event) => {
      setActivities(prev => [{ ...event, category: 'profile' }, ...prev]);
    }
  );

  useRealtimeSubscription(
    `tenant_${user?.tenantId}_users`,
    'user_status_change',
    (event) => {
      setActivities(prev => [{ ...event, category: 'user' }, ...prev]);
    }
  );

  // New real-time subscriptions
  useRealtimeSubscription(
    'deployments',
    'deployment_update',
    (event) => {
      setActivities(prev => [{ ...event, category: 'deployment' }, ...prev]);
    }
  );

  useRealtimeSubscription(
    'function_executions',
    'function_execution',
    (event) => {
      setActivities(prev => [{ ...event, category: 'function' }, ...prev]);
    }
  );

  useRealtimeSubscription(
    `tenant_${user?.tenantId}_team`,
    'team_update',
    (event) => {
      setActivities(prev => [{ ...event, category: 'team' }, ...prev]);
    }
  );

  useRealtimeSubscription(
    'registry_updates',
    'registry_update',
    (event) => {
      setActivities(prev => [{ ...event, category: 'registry' }, ...prev]);
    }
  );

  return activities
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 100); // Keep last 100 activities
}
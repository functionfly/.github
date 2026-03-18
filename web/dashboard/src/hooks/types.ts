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
  message?: string;
  read_at?: string;
}

export interface PresenceEvent extends RealtimeEvent {
  type: 'presence_join' | 'presence_leave';
  key: string;
  current_presences: any[];
  new_presences?: any[];
  left_presences?: any[];
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

// Admin API Client hook
export { useAdminApiClient } from './useAdminApiClient';

// Authentication hooks
export { useAdminAuth } from './useAdminAuth';
export { useSessionMonitor } from './useSessionMonitor';

// Real-time WebSocket hooks
export { useRealtimeSubscription } from './useRealtimeSubscription';
export { useStatusWebSocket } from './useStatusWebSocket';

export {
  useDeploymentUpdates,
  useFunctionMetricsUpdates,
  useRealtimeDeployments,
} from './useRealtimeDeployments';

export type { RealtimeEvent } from './useRealtimeSubscription';

export type {
  DeploymentUpdate,
  FunctionMetrics,
  UseRealtimeDeploymentsOptions,
  UseRealtimeDeploymentsResult,
} from './useRealtimeDeployments';

// Re-export useRealtime for convenience
export { useRealtime } from './useRealtime';

// Access control hook
export {
  useAccessControl,
  withAccessControl,
  type Permission,
  type AccessDeniedReason,
} from './useAccessControl';

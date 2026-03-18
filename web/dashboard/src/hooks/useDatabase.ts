import { useEffect, useState } from 'react';
import {
  validateDatabaseAlerts,
  validateDatabaseHealth,
  validateDatabaseMetrics,
  validateTableName,
  validateTimeRange,
} from '../lib/validation-utils';
import { DatabaseAlert, DatabaseHealth, DatabaseMetric } from './types';
import type { DatabaseChangeEvent, RealtimePostgresChangesPayload } from './useRealtime';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';

const MAX_CHANGES = 100;

// Hook for database changes via WebSocket realtime (backend can be backed by Neon or Postgres LISTEN/NOTIFY)
export function useDatabaseChanges(table: string, filter?: string) {
  const [changes, setChanges] = useState<RealtimePostgresChangesPayload<any>[]>([]);
  const [validationError, setValidationError] = useState<string | null>(null);

  useEffect(() => {
    const tableValidation = validateTableName(table);
    if (!tableValidation.success) {
      setValidationError(tableValidation.error ?? 'Invalid table name');
      return;
    }
    setValidationError(null);
  }, [table]);

  useRealtimeSubscription<DatabaseChangeEvent>(`db_changes_${table}`, 'db_change', (event) => {
    if (event.table !== table) return;
    const payload: RealtimePostgresChangesPayload = {
      data: {
        schema: event.schema ?? '',
        table: event.table ?? table,
        commit_timestamp: event.commit_timestamp ?? new Date().toISOString(),
        eventType: (event.eventType ?? 'INSERT') as 'INSERT' | 'UPDATE' | 'DELETE',
        new: event.new ?? null,
        old: event.old ?? null,
        errors: event.errors ?? null,
      },
      ids: event.ids ?? [],
    };
    if (filter) {
      const row = (event.new ?? event.old) as Record<string, unknown> | null;
      if (row && typeof row === 'object') {
        const [key, value] = filter.split('=').map((s) => s.trim());
        if (key && row[key] !== value) return;
      }
    }
    setChanges((prev) => [payload, ...prev].slice(0, MAX_CHANGES));
  });

  return { changes, validationError };
}

// Hook for real-time database health monitoring
export function useDatabaseHealth() {
  const [health, setHealth] = useState<DatabaseHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);

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

        // Validate and parse API response using Zod schema
        const validationResult = validateDatabaseHealth(data);
        if (validationResult.success && validationResult.data) {
          setHealth(validationResult.data as DatabaseHealth);
          setValidationError(null);
        } else {
          console.warn(
            'Database health validation failed, using fallback:',
            validationResult.error
          );
          setValidationError(validationResult.error || 'Validation failed');
          // Fallback to mock data if validation fails
          const mockHealth: DatabaseHealth = {
            status: 'healthy',
            connections: {
              active: 0,
              idle: 0,
              total: 0,
              max: 100,
            },
            performance: {
              avgQueryTime: 0,
              slowQueries: 0,
              throughput: 0,
            },
            storage: {
              used: 0,
              total: 0,
              growthRate: 0,
            },
            replication: {
              lag: 0,
              status: 'healthy',
            },
            lastUpdated: new Date().toISOString(),
          };
          setHealth(mockHealth);
        }
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
            max: 100,
          },
          performance: {
            avgQueryTime: 0,
            slowQueries: 0,
            throughput: 0,
          },
          storage: {
            used: 0,
            total: 0,
            growthRate: 0,
          },
          replication: {
            lag: 0,
            status: 'healthy',
          },
          lastUpdated: new Date().toISOString(),
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
    validationError,
    isRealtimeConnected: isConnected,
  };
}

// Hook for real-time database alerts
export function useDatabaseAlerts() {
  const [alerts, setAlerts] = useState<DatabaseAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [validationError, setValidationError] = useState<string | null>(null);

  // Subscribe to new database alerts
  const { isConnected } = useRealtimeSubscription(
    'database_alerts',
    'db_alert_created',
    (event) => {
      if (event.type === 'db_alert_created' && event.data) {
        setAlerts((prev) => [event.data, ...prev.slice(0, 9)]); // Keep last 10 alerts
      }
    }
  );

  // Subscribe to alert resolutions
  useRealtimeSubscription('database_alerts', 'db_alert_resolved', (event) => {
    if (event.type === 'db_alert_resolved' && event.data) {
      setAlerts((prev) =>
        prev.map((alert) => (alert.id === event.data.id ? { ...alert, resolved: true } : alert))
      );
    }
  });

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

        // Validate and parse API response using Zod schema
        const validationResult = validateDatabaseAlerts(data);
        setAlerts(validationResult as DatabaseAlert[]); // validateDatabaseAlerts returns DatabaseAlert[]
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
    validationError,
    isRealtimeConnected: isConnected,
  };
}

// Hook for database performance metrics
export function useDatabaseMetrics(timeRange: '1h' | '6h' | '24h' | '7d' = '1h') {
  const [metrics, setMetrics] = useState<DatabaseMetric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);

  useEffect(() => {
    // Validate timeRange parameter
    const timeRangeValidation = validateTimeRange(timeRange);
    if (!timeRangeValidation.success) {
      setValidationError(timeRangeValidation.error || 'Invalid time range');
      setLoading(false);
      return;
    }

    setValidationError(null);

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

        // Validate and parse API response using Zod schema
        const validationResult = validateDatabaseMetrics(data);
        setMetrics(validationResult as DatabaseMetric[]); // validateDatabaseMetrics returns DatabaseMetric[]
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
    validationError,
  };
}

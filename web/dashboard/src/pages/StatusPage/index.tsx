import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Activity, Radio, Wrench } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useStatus, useUptimeMetrics, useMaintenance, useIncidents } from '@/hooks/useStatus';
import { useRealtimeStatus } from '@/hooks/useStatusWebSocket';
import { useStatusStore } from '@/stores/statusStore';
import { HeroStatus } from './components/HeroStatus';
import { ComponentStatus } from './components/ComponentStatus';
import { ProviderGrid } from './components/ProviderCard';
import { IncidentTimeline } from './components/IncidentTimeline';
import { UptimeChart } from './components/UptimeChart';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

// Maintenance banner component
function MaintenanceBanner() {
  const { data: maintenance, isLoading } = useMaintenance();

  if (isLoading || !maintenance || maintenance.length === 0) return null;

  const upcomingMaintenance = maintenance
    .filter((m) => m.status === 'scheduled')
    .sort(
      (a, b) =>
        new Date(a.scheduled_start).getTime() -
        new Date(b.scheduled_start).getTime()
    )[0];

  if (!upcomingMaintenance) return null;

  const startDate = new Date(upcomingMaintenance.scheduled_start);
  const now = new Date();
  const hoursUntil = Math.ceil(
    (startDate.getTime() - now.getTime()) / (1000 * 60 * 60)
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      className="mb-6 rounded-lg border border-amber-500/30 bg-amber-500/10 p-4"
    >
      <div className="flex items-start gap-3">
        <Wrench className="h-5 w-5 text-amber-400 mt-0.5" />
        <div className="flex-1">
          <h3 className="font-medium text-amber-400">
            Scheduled Maintenance: {upcomingMaintenance.title}
          </h3>
          <p className="mt-1 text-sm text-text-secondary">
            {upcomingMaintenance.description}
          </p>
          <div className="mt-2 flex items-center gap-4 text-xs text-text-muted">
            <span>
              Starting: {startDate.toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })}
            </span>
            {hoursUntil <= 24 && (
              <Badge variant="warning" className="text-xs">
                {hoursUntil <= 1 ? 'Starting soon' : `In ${hoursUntil} hours`}
              </Badge>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
}

// Real-time connection indicator
function ConnectionStatus() {
  const { isRealtime, isConnecting, wsError, reconnectAttempt } = useRealtimeStatus();

  if (wsError) {
    return (
      <Badge variant="secondary" className="text-xs bg-red-500/10 text-red-400">
        <Activity className="mr-1 h-3 w-3" />
        Offline
      </Badge>
    );
  }

  if (isConnecting) {
    return (
      <Badge variant="secondary" className="text-xs">
        <Activity className="mr-1 h-3 w-3 animate-pulse" />
        Connecting{reconnectAttempt > 0 ? ` (${reconnectAttempt})` : ''}
      </Badge>
    );
  }

  if (isRealtime) {
    return (
      <Badge variant="success" className="text-xs">
        <Radio className="mr-1 h-3 w-3" />
        Live
      </Badge>
    );
  }

  return (
    <Badge variant="secondary" className="text-xs text-text-muted">
      <Activity className="mr-1 h-3 w-3" />
      Polling
    </Badge>
  );
}

// Subscribe to status updates from store
function useStatusSync() {
  const {
    platformStatus,
    components,
    providers,
    setPlatformStatus,
    setComponents,
    setProviders,
  } = useStatusStore();

  return {
    platformStatus,
    components,
    providers,
    setPlatformStatus,
    setComponents,
    setProviders,
  };
}

export default function StatusPage() {
  const { setPlatformStatus, setComponents, setProviders } = useStatusSync();

  // Fetch status data
  const {
    platformStatus,
    components,
    providers,
    isLoading: isStatusLoading,
    refetch: refetchStatus,
  } = useStatus();

  // Fetch incidents
  const { data: incidentsData, isLoading: isIncidentsLoading } = useIncidents({
    limit: 20,
  });

  // Fetch uptime metrics
  const { data: uptimeMetrics, isLoading: isUptimeLoading } = useUptimeMetrics(30);

  // Sync data to store when it changes
  useEffect(() => {
    if (platformStatus) {
      setPlatformStatus(platformStatus);
    }
  }, [platformStatus, setPlatformStatus]);

  useEffect(() => {
    if (components) {
      setComponents(components);
    }
  }, [components, setComponents]);

  useEffect(() => {
    if (providers) {
      setProviders(providers);
    }
  }, [providers, setProviders]);

  const isLoading = isStatusLoading || isIncidentsLoading;

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Header */}
      <header className="border-b border-border-subtle bg-bg-secondary/50 backdrop-blur-sm">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 items-center justify-between">
            <div>
              <h1 className="text-xl font-semibold text-text-primary">
                System Status
              </h1>
              <p className="text-sm text-text-secondary">
                Real-time platform health and incident tracking
              </p>
            </div>
            <ConnectionStatus />
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="space-y-8">
          {/* Maintenance banner */}
          <MaintenanceBanner />

          {/* Hero status section */}
          <HeroStatus
            status={platformStatus}
            lastUpdated={platformStatus?.timestamp || null}
            isLoading={isStatusLoading}
            onRefresh={refetchStatus}
          />

          {/* Component health grid (fallback to platform status components when dedicated list is empty) */}
          <ComponentStatus
            components={
              (components && components.length > 0)
                ? components
                : platformStatus?.components ?? []
            }
            isLoading={isStatusLoading}
          />

          {/* Provider status cards */}
          <ProviderGrid
            providers={providers || []}
            isLoading={isStatusLoading}
          />

          {/* Two-column layout for incidents and uptime */}
          <div className="grid gap-8 lg:grid-cols-2">
            {/* Recent incidents */}
            <IncidentTimeline
              incidents={incidentsData?.incidents || []}
              isLoading={isIncidentsLoading}
              showFilters={true}
              maxItems={5}
            />

            {/* Uptime chart */}
            <UptimeChart
              metrics={uptimeMetrics || null}
              isLoading={isUptimeLoading}
            />
          </div>

          {/* Footer note */}
          <div className="rounded-lg border border-border-subtle bg-bg-secondary p-4 text-center">
            <p className="text-sm text-text-muted">
              Status updates automatically every 30 seconds via WebSocket connection.
              Last refreshed: {new Date().toLocaleTimeString()}.
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}

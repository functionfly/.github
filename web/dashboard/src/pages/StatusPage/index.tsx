import { useEffect } from 'react';
import { Wrench } from 'lucide-react';
import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components';
import { useStatus, useUptimeMetrics, useMaintenance, useIncidents } from '@/hooks/useStatus';
import { useRealtimeStatus } from '@/hooks/useStatusWebSocket';
import { useStatusStore } from '@/stores/statusStore';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  StatusPill,
  AnnotationTag,
} from '@/components/containment';
import { HeroStatus } from './components/HeroStatus';
import { ComponentStatus } from './components/ComponentStatus';
import { ProviderGrid } from './components/ProviderCard';
import { IncidentTimeline } from './components/IncidentTimeline';
import { UptimeChart } from './components/UptimeChart';

import './status.css';

function MaintenanceBanner() {
  const { data: maintenance, isLoading } = useMaintenance();

  if (isLoading || !maintenance || maintenance.length === 0) return null;

  const upcomingMaintenance = maintenance
    .filter((m) => m.status === 'scheduled')
    .sort((a, b) => new Date(a.scheduled_start).getTime() - new Date(b.scheduled_start).getTime())[0];

  if (!upcomingMaintenance) return null;

  const startDate = new Date(upcomingMaintenance.scheduled_start);
  const now = new Date();
  const hoursUntil = Math.ceil((startDate.getTime() - now.getTime()) / (1000 * 60 * 60));

  return (
    <Chamber className="status-maintenance">
      <div className="status-maintenance__inner">
        <Wrench className="status-maintenance__icon" />
        <div className="status-maintenance__content">
          <h3 className="status-maintenance__title">
            Scheduled Maintenance: {upcomingMaintenance.title}
          </h3>
          <p className="status-maintenance__desc">{upcomingMaintenance.description}</p>
          <div className="status-maintenance__meta">
            <span>
              Starting: {startDate.toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })}
            </span>
            {hoursUntil <= 24 && (
              <StatusPill status="pending" label={hoursUntil <= 1 ? 'Starting soon' : `In ${hoursUntil} hours`} />
            )}
          </div>
        </div>
      </div>
    </Chamber>
  );
}

function ConnectionStatus() {
  const { isRealtime, isConnecting, wsError, reconnectAttempt } = useRealtimeStatus();

  if (wsError) {
    return <StatusPill status="revoked" label="Offline" />;
  }

  if (isConnecting) {
    return <StatusPill status="pending" label={`Connecting${reconnectAttempt > 0 ? ` (${reconnectAttempt})` : ''}`} />;
  }

  if (isRealtime) {
    return <StatusPill status="live" label="Live" />;
  }

  return <StatusPill status="pending" label="Polling" />;
}

function useStatusSync() {
  const { platformStatus, components, providers, setPlatformStatus, setComponents, setProviders } = useStatusStore();
  return { platformStatus, components, providers, setPlatformStatus, setComponents, setProviders };
}

export default function StatusPage() {
  const { setPlatformStatus, setComponents, setProviders } = useStatusSync();

  const {
    platformStatus,
    components,
    providers,
    isLoading: isStatusLoading,
    refetch: refetchStatus,
  } = useStatus();

  const { data: incidentsData, isLoading: isIncidentsLoading } = useIncidents({ limit: 20 });
  const { data: uptimeMetrics, isLoading: isUptimeLoading } = useUptimeMetrics(30);

  useEffect(() => {
    if (platformStatus) setPlatformStatus(platformStatus);
  }, [platformStatus, setPlatformStatus]);

  useEffect(() => {
    if (components) setComponents(components);
  }, [components, setComponents]);

  useEffect(() => {
    if (providers) setProviders(providers);
  }, [providers, setProviders]);

  const isLoading = isStatusLoading || isIncidentsLoading;

  return (
    <div className="status-page">
      <PageGrid />
      <Navbar variant="dashboard" />

      <main className="status-main">
        {/* Hero */}
        <Chamber className="status-hero" ribs>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="MODULE SYS-01" secondary="System Status" position="top-right" />

          <div className="status-hero__header">
            <div className="status-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="status-hero__title">System Status</h1>
              <ConnectionStatus />
            </div>
            <p className="status-hero__subtitle">
              Real-time platform health and incident tracking
            </p>
          </div>
        </Chamber>

        {/* Maintenance banner */}
        <MaintenanceBanner />

        {/* Hero status */}
        <HeroStatus
          status={platformStatus}
          lastUpdated={platformStatus?.timestamp || null}
          isLoading={isStatusLoading}
          onRefresh={refetchStatus}
        />

        {/* Component health */}
        <ComponentStatus
          components={
            (components && components.length > 0)
              ? components
              : platformStatus?.components ?? []
          }
          isLoading={isStatusLoading}
        />

        {/* Provider status */}
        <ProviderGrid providers={providers || []} isLoading={isStatusLoading} />

        {/* Incidents + Uptime */}
        <div className="status-two-col">
          <IncidentTimeline
            incidents={incidentsData?.incidents || []}
            isLoading={isIncidentsLoading}
            showFilters={true}
            maxItems={5}
          />
          <UptimeChart metrics={uptimeMetrics || null} isLoading={isUptimeLoading} />
        </div>

        {/* Footer note */}
        <Chamber className="status-footer">
          <p className="status-footer__text">
            Status updates automatically every 30 seconds via WebSocket connection.
            Last refreshed: {new Date().toLocaleTimeString()}.
          </p>
        </Chamber>
      </main>

      <Footer showScrollToTop={false} />
    </div>
  );
}

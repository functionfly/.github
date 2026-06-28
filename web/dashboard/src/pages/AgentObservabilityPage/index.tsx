'use client';

import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Bot, Play, Pause, SkipForward, BarChart3, GitBranch, Clock, DollarSign, AlertCircle, Settings, Download, Search, Info, Layers } from 'lucide-react';
import { Link } from 'react-router-dom';
import { ROUTES } from '@/lib/constants';
import { useAtlasRuns, useAtlasEvents, useAtlasStream, useAtlasConfig, RunsFilter, RunsPagination, EventsPagination } from '@/hooks/useAtlasObservability';
import { useAtlasTrace } from '@/hooks/useAtlasTraces';
import { ExecutionTraceViewer } from '@/components/atlas';
import { useKeyboardShortcuts, KeyboardShortcutsHelp } from '@/hooks/useKeyboardShortcuts';
import AgentRunTimeline from '@/components/observability/AgentRunTimeline';
import DecisionGraphViewer from '@/components/observability/DecisionGraphViewer';
import EventDetailPanel from '@/components/observability/EventDetailPanel';
import CostBreakdownPanel from '@/components/observability/CostBreakdownPanel';
import ReplayControls from '@/components/observability/ReplayControls';
import AtlasStatusBadge from '@/components/observability/AtlasStatusBadge';
import SpanNavigator from '@/components/observability/SpanNavigator';
import AtlasConfigPanel from '@/components/observability/AtlasConfigPanel';
import SpanDetailPanel from '@/components/observability/SpanDetailPanel';
import GraphNodeDetail from '@/components/observability/GraphNodeDetail';
import PaginationControls from '@/components/observability/PaginationControls';
import DataExport from '@/components/observability/DataExport';
import AutoRefreshToggle from '@/components/observability/AutoRefreshToggle';
import ReconnectButton from '@/components/observability/ReconnectButton';
import ConfirmDialog from '@/components/observability/ConfirmDialog';
import RunMetadataPanel from '@/components/observability/RunMetadataPanel';
import { RunsListSkeleton, EventsListSkeleton, StatsSkeleton } from '@/components/observability/Skeletons';
import { DateRangePicker } from '@/components/ui/date-picker';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
} from '@/components/containment';

import './observability.css';

export function AgentObservabilityPage() {
  const { t } = useTranslation();
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [agentIdFilter, setAgentIdFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [dateRange, setDateRange] = useState<{ from: Date | null; to: Date | null }>({ from: null, to: null });
  const [replaySpeed, setReplaySpeed] = useState<number>(1);
  const [isReplaying, setIsReplaying] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(10000);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSpanId, setSelectedSpanId] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [activeTab, setActiveTab] = useState<'events' | 'graph' | 'cost' | 'spans' | 'atlas-trace'>('events');

  const [runsPagination, setRunsPagination] = useState<RunsPagination>({ limit: 20, offset: 0, total: 0 });
  const [eventsPagination, setEventsPagination] = useState<EventsPagination>({ limit: 100, afterSequence: 0, total: 0 });

  const filter: RunsFilter = {
    agentId: agentIdFilter || undefined,
    status: statusFilter === 'all' ? undefined : statusFilter,
    startedAfter: dateRange.from?.toISOString(),
    startedBefore: dateRange.to?.toISOString(),
  };

  const { runs, loading: runsLoading, paginationInfo, refetch: refetchRuns } = useAtlasRuns(filter, { limit: runsPagination.limit, offset: runsPagination.offset });
  const { events, loading: eventsLoading, paginationInfo: eventsPaginationInfo } = useAtlasEvents(selectedRunId || undefined, { limit: eventsPagination.limit, afterSequence: eventsPagination.afterSequence });
  const { events: liveEvents, connected, reconnect } = useAtlasStream(selectedRunId || undefined, { autoRefresh, refreshInterval });
  const { config, loading: configLoading, updateConfig } = useAtlasConfig();

  const selectedRun = runs?.find(r => r.id === selectedRunId);
  const atlasRunId = selectedRun?.atlas_run_id || '';
  const { data: atlasTrace, isLoading: atlasTraceLoading } = useAtlasTrace(atlasRunId);

  const filteredEvents = searchQuery
    ? events.filter(e =>
        JSON.stringify(e.payload).toLowerCase().includes(searchQuery.toLowerCase()) ||
        e.kind.toLowerCase().includes(searchQuery.toLowerCase()) ||
        e.system_id.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : events;

  const handleRunsPageChange = useCallback((offset: number) => {
    setRunsPagination(prev => ({ ...prev, offset }));
  }, []);

  const handleEndRun = async () => {
    if (!selectedRunId) return;
    try {
      await fetch(`/v1/agent-observability/runs/${selectedRunId}/end`, { method: 'POST' });
      refetchRuns();
    } catch (error) {
      console.error('Failed to end run:', error);
    }
  };

  useKeyboardShortcuts({
    onNext: () => {
      const currentIndex = runs?.findIndex(r => r.id === selectedRunId) ?? -1;
      if (currentIndex < (runs?.length ?? 0) - 1) {
        setSelectedRunId(runs?.[currentIndex + 1]?.id ?? null);
      }
    },
    onPrevious: () => {
      const currentIndex = runs?.findIndex(r => r.id === selectedRunId) ?? -1;
      if (currentIndex > 0) {
        setSelectedRunId(runs?.[currentIndex - 1]?.id ?? null);
      }
    },
    onEscape: () => {
      setSelectedSpanId(null);
      setSelectedNodeId(null);
    },
    enabled: true,
  });

  const tabs = [
    { value: 'events' as const, label: 'Events' },
    { value: 'graph' as const, label: 'Decision Graph' },
    { value: 'cost' as const, label: 'Cost Breakdown' },
    { value: 'spans' as const, label: 'Spans' },
    { value: 'atlas-trace' as const, label: 'Atlas Trace', icon: Layers },
  ];

  const statusToPill = (status: string): 'live' | 'pending' | 'revoked' => {
    if (status === 'completed') return 'live';
    if (status === 'failed') return 'revoked';
    return 'pending';
  };

  return (
    <div className="obs-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="obs-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE OBS-01" secondary="Agent Observability" position="top-right" />

        <div className="obs-hero__header">
          <div className="obs-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="obs-hero__title">Agent Observability</h1>
            <AtlasStatusBadge connected={connected} />
          </div>
          <p className="obs-hero__subtitle">
            AI agent decision replay and debugging with Atlas
          </p>
          <div className="obs-hero__actions">
            <FrameButton size="sm" onClick={() => setShowShortcuts(!showShortcuts)} iconLeft={<Info className="obs-icon-sm" />}>
              Shortcuts
            </FrameButton>
            <FrameButton size="sm" onClick={() => setShowConfig(!showConfig)} iconLeft={<Settings className="obs-icon-sm" />}>
              Config
            </FrameButton>
          </div>
        </div>
      </Chamber>

      {showShortcuts && (
        <Chamber className="obs-shortcuts">
          <KeyboardShortcutsHelp />
        </Chamber>
      )}

      {showConfig && (
        <AtlasConfigPanel config={config} onUpdate={updateConfig} loading={configLoading} />
      )}

      {/* Main Grid */}
      <div className="obs-main-grid">
        {/* Left Panel: Runs */}
        <div className="obs-left-panel">
          <Chamber className="obs-runs-chamber">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <h2 className="obs-section-title">Runs</h2>
            <p className="obs-section-desc">AI agent execution history</p>

            <div className="obs-filters">
              <input
                className="obs-input"
                placeholder="Filter by agent ID..."
                value={agentIdFilter}
                onChange={(e) => setAgentIdFilter(e.target.value)}
              />
              <DateRangePicker
                value={dateRange}
                onChange={(range) => setDateRange({ from: range.from, to: range.to })}
                placeholder="Filter by date..."
              />
              <select className="obs-select" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                <option value="all">All Status</option>
                <option value="running">Running</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
              </select>

              <div className="obs-filters__row">
                <label className="obs-auto-refresh">
                  <input type="checkbox" checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} />
                  <span>Auto-refresh</span>
                </label>
                <DataExport data={runs || []} filename="runs" />
              </div>
            </div>

            <div className="obs-runs-list">
              {runsLoading ? (
                <RunsListSkeleton />
              ) : runs?.length === 0 ? (
                <div className="obs-empty">
                  <Bot className="obs-empty__icon" />
                  <h3 className="obs-empty__title">No runs found</h3>
                  <p className="obs-empty__desc">Try adjusting your filters or wait for new runs to appear</p>
                </div>
              ) : (
                runs?.map((run) => (
                  <button
                    key={run.id}
                    className={`obs-run-item ${selectedRunId === run.id ? 'obs-run-item--selected' : ''}`}
                    onClick={() => setSelectedRunId(run.id)}
                  >
                    <div className="obs-run-item__header">
                      <span className="obs-run-item__agent">{run.agent_id}</span>
                      <StatusPill status={statusToPill(run.status)} label={run.status} />
                    </div>
                    <div className="obs-run-item__meta">
                      <span><Clock className="obs-icon-xs" /> {new Date(run.started_at).toLocaleTimeString()}</span>
                      <span><DollarSign className="obs-icon-xs" /> ${run.total_cost_usd.toFixed(4)}</span>
                    </div>
                  </button>
                ))
              )}
            </div>

            <PaginationControls
              total={paginationInfo.total}
              limit={paginationInfo.limit}
              offset={paginationInfo.offset}
              onPageChange={handleRunsPageChange}
              loading={runsLoading}
            />
          </Chamber>

          {/* Statistics */}
          {selectedRun && (
            <Chamber className="obs-stats-chamber">
              <CornerBrace position="tr" />
              <CornerBrace position="bl" />
              <div className="obs-stats__header">
                <h2 className="obs-section-title">Statistics</h2>
                <RunMetadataPanel run={selectedRun}>
                  <FrameButton size="sm" iconLeft={<Info className="obs-icon-xs" />}>Details</FrameButton>
                </RunMetadataPanel>
              </div>
              {runsLoading ? (
                <StatsSkeleton />
              ) : (
                <div className="obs-stats-grid">
                  <div className="obs-stat-cell">
                    <p className="obs-stat-cell__value">{selectedRun.event_count}</p>
                    <p className="obs-stat-cell__label">Events</p>
                  </div>
                  <div className="obs-stat-cell">
                    <p className="obs-stat-cell__value">{selectedRun.error_count}</p>
                    <p className="obs-stat-cell__label">Errors</p>
                  </div>
                  <div className="obs-stat-cell">
                    <p className="obs-stat-cell__value">{selectedRun.total_input_tokens}</p>
                    <p className="obs-stat-cell__label">Input Tokens</p>
                  </div>
                  <div className="obs-stat-cell">
                    <p className="obs-stat-cell__value">{selectedRun.total_output_tokens}</p>
                    <p className="obs-stat-cell__label">Output Tokens</p>
                  </div>
                </div>
              )}
            </Chamber>
          )}
        </div>

        {/* Right Panel: Detail */}
        <div className="obs-right-panel">
          {selectedRunId ? (
            <>
              {/* Timeline */}
              <Chamber className="obs-timeline-chamber">
                <CornerBrace position="tl" />
                <CornerBrace position="br" />
                <div className="obs-timeline__header">
                  <div>
                    <h2 className="obs-section-title">Run Timeline</h2>
                    <p className="obs-section-desc">Event sequence visualization</p>
                  </div>
                  <div className="obs-timeline__controls">
                    {!connected && <ReconnectButton onReconnect={reconnect} />}
                    <ReplayControls speed={replaySpeed} onSpeedChange={setReplaySpeed} isPlaying={isReplaying} onPlayPause={() => setIsReplaying(!isReplaying)} />
                  </div>
                </div>
                <AgentRunTimeline events={liveEvents.length > 0 ? liveEvents : events} />
              </Chamber>

              {/* Tabs */}
              <div className="obs-tabs">
                {tabs.map((tab) => (
                  <button
                    key={tab.value}
                    className={`obs-tab ${activeTab === tab.value ? 'obs-tab--active' : ''}`}
                    onClick={() => setActiveTab(tab.value)}
                  >
                    {tab.icon && <tab.icon className="obs-icon-xs" />}
                    {tab.label}
                  </button>
                ))}
              </div>

              <div className="obs-tab-content">
                {activeTab === 'events' && (
                  <Chamber>
                    <div className="obs-search-bar">
                      <Search className="obs-icon-sm obs-search-icon" />
                      <input className="obs-input obs-search-input" placeholder="Search events..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} />
                      <DataExport data={filteredEvents} filename={`events-${selectedRunId}`} />
                    </div>
                    {eventsLoading ? <EventsListSkeleton /> : <EventDetailPanel events={filteredEvents} />}
                  </Chamber>
                )}

                {activeTab === 'graph' && (
                  <Chamber>
                    <DecisionGraphViewer runId={selectedRunId} onNodeClick={setSelectedNodeId} />
                    {selectedNodeId && (
                      <GraphNodeDetail nodeId={selectedNodeId} runId={selectedRunId} onClose={() => setSelectedNodeId(null)} />
                    )}
                  </Chamber>
                )}

                {activeTab === 'cost' && (
                  <Chamber>
                    <CostBreakdownPanel runId={selectedRunId} />
                  </Chamber>
                )}

                {activeTab === 'spans' && (
                  <Chamber>
                    <SpanNavigator runId={selectedRunId} onSpanClick={setSelectedSpanId} />
                    {selectedSpanId && (
                      <SpanDetailPanel spanId={selectedSpanId} runId={selectedRunId} onClose={() => setSelectedSpanId(null)} />
                    )}
                  </Chamber>
                )}

                {activeTab === 'atlas-trace' && (
                  <>
                    {atlasTrace ? (
                      <ExecutionTraceViewer run={atlasTrace.run} events={atlasTrace.events} graph={atlasTrace.graph} />
                    ) : atlasTraceLoading ? (
                      <Chamber className="obs-loading">
                        <span className="obs-loading__spinner" />
                        <span className="obs-loading__text">Loading Atlas trace...</span>
                      </Chamber>
                    ) : (
                      <Chamber className="obs-empty">
                        <Layers className="obs-empty__icon" />
                        <h3 className="obs-empty__title">No Atlas Trace</h3>
                        <p className="obs-empty__desc">
                          {atlasRunId
                            ? 'No trace data found for this run in Atlas Memory Engine.'
                            : 'This run does not have an associated Atlas trace. Configure ATLAS_URL to enable tracing.'}
                        </p>
                      </Chamber>
                    )}
                  </>
                )}
              </div>

              {selectedRun?.status === 'running' && (
                <div className="obs-end-run">
                  <ConfirmDialog
                    title="End Run"
                    description="Are you sure you want to end this run? This action cannot be undone."
                    confirmLabel="End Run"
                    variant="destructive"
                    onConfirm={handleEndRun}
                    trigger={
                      <SealedButton size="sm">End Run</SealedButton>
                    }
                  />
                </div>
              )}
            </>
          ) : (
            <Chamber className="obs-empty obs-empty--large">
              <Bot className="obs-empty__icon" />
              <h3 className="obs-empty__title">Select a run to view</h3>
              <p className="obs-empty__desc">Choose a run from the list to view its timeline, events, and decision graph</p>
            </Chamber>
          )}
        </div>
      </div>
    </div>
  );
}

export default AgentObservabilityPage;

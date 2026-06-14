'use client';

import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DateRangePicker } from '@/components/ui/date-picker';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Bot, Play, Pause, SkipForward, BarChart3, GitBranch, Clock, DollarSign, AlertCircle, Settings, Download, Search, Info } from 'lucide-react';
import { Link } from 'react-router-dom';
import { ROUTES } from '@/lib/constants';
import { useAtlasRuns, useAtlasEvents, useAtlasStream, useAtlasConfig, RunsFilter, RunsPagination, EventsPagination } from '@/hooks/useAtlasObservability';
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
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
            <Bot className="h-5 w-5 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">Agent Observability</h1>
            <p className="text-muted-foreground text-sm">AI agent decision replay and debugging with Atlas</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <AtlasStatusBadge connected={connected} />
          <Button variant="outline" size="sm" onClick={() => setShowShortcuts(!showShortcuts)} className="gap-2">
            <Info className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setShowConfig(!showConfig)} className="gap-2">
            <Settings className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {showShortcuts && (
        <Card className="p-4">
          <KeyboardShortcutsHelp />
        </Card>
      )}

      {showConfig && (
        <AtlasConfigPanel config={config} onUpdate={updateConfig} loading={configLoading} />
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Runs</CardTitle>
              <CardDescription>AI agent execution history</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Input
                  placeholder="Filter by agent ID..."
                  value={agentIdFilter}
                  onChange={(e) => setAgentIdFilter(e.target.value)}
                  className="flex-1"
                />
              </div>

              <div className="flex gap-2">
                <DateRangePicker
                  value={dateRange}
                  onChange={(range) => setDateRange({ from: range.from, to: range.to })}
                  placeholder="Filter by date..."
                />
              </div>

              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="running">Running</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="failed">Failed</SelectItem>
                </SelectContent>
              </Select>

              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Switch id="auto-refresh" checked={autoRefresh} onCheckedChange={setAutoRefresh} />
                  <Label htmlFor="auto-refresh" className="text-sm cursor-pointer">Auto-refresh</Label>
                </div>
                <DataExport data={runs || []} filename="runs" />
              </div>

              <div className="space-y-2 max-h-[400px] overflow-y-auto">
                {runsLoading ? (
                  <RunsListSkeleton />
                ) : runs?.length === 0 ? (
                  <EmptyState
                    icon={<Bot className="h-8 w-8" />}
                    title="No runs found"
                    description="Try adjusting your filters or wait for new runs to appear"
                  />
                ) : (
                  runs?.map((run) => (
                    <div
                      key={run.id}
                      className={`p-3 rounded-lg border cursor-pointer transition-colors ${
                        selectedRunId === run.id
                          ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20'
                          : 'border-border hover:border-purple-300'
                      }`}
                      onClick={() => setSelectedRunId(run.id)}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => e.key === 'Enter' && setSelectedRunId(run.id)}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-mono text-xs text-muted-foreground">{run.agent_id}</span>
                        <Badge
                          variant={run.status === 'completed' ? 'default' : run.status === 'failed' ? 'destructive' : 'secondary'}
                        >
                          {run.status}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {new Date(run.started_at).toLocaleTimeString()}
                        </span>
                        <span className="flex items-center gap-1">
                          <DollarSign className="h-3 w-3" />
                          ${run.total_cost_usd.toFixed(4)}
                        </span>
                      </div>
                    </div>
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
            </CardContent>
          </Card>

          {selectedRun && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="text-lg">Statistics</CardTitle>
                <RunMetadataPanel run={selectedRun}>
                  <Button variant="ghost" size="sm" className="gap-2">
                    <Info className="h-4 w-4" />
                    Details
                  </Button>
                </RunMetadataPanel>
              </CardHeader>
              <CardContent className="space-y-4">
                {runsLoading ? (
                  <StatsSkeleton />
                ) : (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="text-center p-3 rounded-lg bg-muted/50">
                      <p className="text-2xl font-bold">{selectedRun.event_count}</p>
                      <p className="text-xs text-muted-foreground">Events</p>
                    </div>
                    <div className="text-center p-3 rounded-lg bg-muted/50">
                      <p className="text-2xl font-bold">{selectedRun.error_count}</p>
                      <p className="text-xs text-muted-foreground">Errors</p>
                    </div>
                    <div className="text-center p-3 rounded-lg bg-muted/50">
                      <p className="text-2xl font-bold">{selectedRun.total_input_tokens}</p>
                      <p className="text-xs text-muted-foreground">Input Tokens</p>
                    </div>
                    <div className="text-center p-3 rounded-lg bg-muted/50">
                      <p className="text-2xl font-bold">{selectedRun.total_output_tokens}</p>
                      <p className="text-xs text-muted-foreground">Output Tokens</p>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        <div className="lg:col-span-2 space-y-4">
          {selectedRunId ? (
            <>
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle>Run Timeline</CardTitle>
                      <CardDescription>Event sequence visualization</CardDescription>
                    </div>
                    <div className="flex items-center gap-3">
                      {!connected && (
                        <ReconnectButton onReconnect={reconnect} />
                      )}
                      <ReplayControls
                        speed={replaySpeed}
                        onSpeedChange={setReplaySpeed}
                        isPlaying={isReplaying}
                        onPlayPause={() => setIsReplaying(!isReplaying)}
                      />
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <AgentRunTimeline events={liveEvents.length > 0 ? liveEvents : events} />
                </CardContent>
              </Card>

              <Tabs defaultValue="events" className="w-full">
                <TabsList>
                  <TabsTrigger value="events">Events</TabsTrigger>
                  <TabsTrigger value="graph">Decision Graph</TabsTrigger>
                  <TabsTrigger value="cost">Cost Breakdown</TabsTrigger>
                  <TabsTrigger value="spans">Spans</TabsTrigger>
                </TabsList>

                <TabsContent value="events" className="space-y-4">
                  <Card>
                    <CardContent className="pt-6">
                      <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-2 flex-1">
                          <Search className="h-4 w-4 text-muted-foreground" />
                          <Input
                            placeholder="Search events..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="flex-1"
                          />
                        </div>
                        <DataExport data={filteredEvents} filename={`events-${selectedRunId}`} />
                      </div>
                      {eventsLoading ? (
                        <EventsListSkeleton />
                      ) : (
                        <EventDetailPanel events={filteredEvents} />
                      )}
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="graph" className="space-y-4">
                  <Card>
                    <CardContent className="pt-6">
                      <DecisionGraphViewer runId={selectedRunId} onNodeClick={setSelectedNodeId} />
                      {selectedNodeId && (
                        <GraphNodeDetail
                          nodeId={selectedNodeId}
                          runId={selectedRunId}
                          onClose={() => setSelectedNodeId(null)}
                        />
                      )}
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="cost" className="space-y-4">
                  <Card>
                    <CardContent className="pt-6">
                      <CostBreakdownPanel runId={selectedRunId} />
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="spans" className="space-y-4">
                  <Card>
                    <CardContent className="pt-6">
                      <SpanNavigator runId={selectedRunId} onSpanClick={setSelectedSpanId} />
                      {selectedSpanId && (
                        <SpanDetailPanel
                          spanId={selectedSpanId}
                          runId={selectedRunId}
                          onClose={() => setSelectedSpanId(null)}
                        />
                      )}
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>

              {selectedRun?.status === 'running' && (
                <ConfirmDialog
                  title="End Run"
                  description="Are you sure you want to end this run? This action cannot be undone."
                  confirmLabel="End Run"
                  variant="destructive"
                  onConfirm={handleEndRun}
                  trigger={
                    <Button variant="destructive" size="sm" className="gap-2">
                      End Run
                    </Button>
                  }
                />
              )}
            </>
          ) : (
            <Card>
              <CardContent className="py-12 text-center">
                <EmptyState
                  icon={<Bot className="h-12 w-12" />}
                  title="Select a run to view"
                  description="Choose a run from the list to view its timeline, events, and decision graph"
                />
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

export default AgentObservabilityPage;

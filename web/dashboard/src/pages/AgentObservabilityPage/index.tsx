'use client';

import { useState } from 'react';
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
import { ArrowLeft, Bot, Play, Pause, SkipForward, BarChart3, GitBranch, Clock, DollarSign, AlertCircle } from 'lucide-react';
import { Link } from 'react-router-dom';
import { ROUTES } from '@/lib/constants';
import { useAtlasRuns, useAtlasEvents, useAtlasStream, useAtlasConfig } from '@/hooks/useAtlasObservability';
import AgentRunTimeline from '@/components/observability/AgentRunTimeline';
import DecisionGraphViewer from '@/components/observability/DecisionGraphViewer';
import EventDetailPanel from '@/components/observability/EventDetailPanel';
import CostBreakdownPanel from '@/components/observability/CostBreakdownPanel';
import ReplayControls from '@/components/observability/ReplayControls';
import AtlasStatusBadge from '@/components/observability/AtlasStatusBadge';
import SpanNavigator from '@/components/observability/SpanNavigator';

export function AgentObservabilityPage() {
  const { t } = useTranslation();
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [agentIdFilter, setAgentIdFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [replaySpeed, setReplaySpeed] = useState<number>(1);
  const [isReplaying, setIsReplaying] = useState(false);

  const { runs, loading: runsLoading } = useAtlasRuns(agentIdFilter || undefined, statusFilter === 'all' ? undefined : statusFilter);
  const { events, loading: eventsLoading } = useAtlasEvents(selectedRunId || undefined);
  const { events: liveEvents, connected } = useAtlasStream(selectedRunId || undefined);
  const { config, updateConfig } = useAtlasConfig();

  const selectedRun = runs?.find(r => r.id === selectedRunId);

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
        <AtlasStatusBadge connected={connected} />
      </div>

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

              <div className="space-y-2 max-h-[500px] overflow-y-auto">
                {runsLoading ? (
                  <p className="text-sm text-muted-foreground">Loading...</p>
                ) : runs?.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No runs found</p>
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
            </CardContent>
          </Card>

          {selectedRun && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Statistics</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
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
                    <ReplayControls
                      speed={replaySpeed}
                      onSpeedChange={setReplaySpeed}
                      isPlaying={isReplaying}
                      onPlayPause={() => setIsReplaying(!isReplaying)}
                    />
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
                      <EventDetailPanel events={events} />
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="graph" className="space-y-4">
                  <Card>
                    <CardContent className="pt-6">
                      <DecisionGraphViewer runId={selectedRunId} />
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
                      <SpanNavigator runId={selectedRunId} />
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </>
          ) : (
            <Card>
              <CardContent className="py-12 text-center">
                <Bot className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <p className="text-muted-foreground">Select a run to view its timeline and events</p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

export default AgentObservabilityPage;
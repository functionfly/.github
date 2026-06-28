import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  AlertCircle,
  ArrowRight,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock,
  GitBranch,
  ArrowDownToLine,
  Layers,
  Lightbulb,
  Play,
  Zap,
} from 'lucide-react';
import type { AtlasEvent, AtlasGraphResponse, AtlasRunRecord } from '@/api/atlas';

interface ExecutionTraceViewerProps {
  run: AtlasRunRecord;
  events: AtlasEvent[];
  graph?: AtlasGraphResponse;
  className?: string;
}

const kindConfig: Record<string, { color: string; bg: string; icon: typeof ArrowDownToLine; label: string }> = {
  input:    { color: 'text-blue-400',   bg: 'bg-blue-500/10 border-blue-500/20',   icon: ArrowDownToLine, label: 'Input' },
  decision: { color: 'text-amber-400',  bg: 'bg-amber-500/10 border-amber-500/20', icon: Lightbulb,    label: 'Decision' },
  action:   { color: 'text-violet-400', bg: 'bg-violet-500/10 border-violet-500/20', icon: Zap,          label: 'Action' },
  result:   { color: 'text-emerald-400',bg: 'bg-emerald-500/10 border-emerald-500/20', icon: CheckCircle2, label: 'Result' },
  error:    { color: 'text-red-400',    bg: 'bg-red-500/10 border-red-500/20',     icon: AlertCircle,  label: 'Error' },
};

function formatNs(ns: number): string {
  const ms = ns / 1_000_000;
  if (ms < 1000) return `${ms.toFixed(1)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

function formatTimestamp(ns: number): string {
  const date = new Date(ns / 1_000_000);
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3 });
}

function EventCard({ event, startNs, isExpanded, onToggle }: {
  event: AtlasEvent;
  startNs: number;
  isExpanded: boolean;
  onToggle: () => void;
}) {
  const config = kindConfig[event.kind] || kindConfig.input;
  const Icon = config.icon;
  const relativeMs = (event.timestamp_ns - startNs) / 1_000_000;
  const payload = typeof event.payload === 'string' ? JSON.parse(event.payload as string) : event.payload;

  return (
    <div className={`border rounded-lg ${config.bg} transition-all`}>
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-white/[0.02] transition-colors"
      >
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <Icon className={`h-4 w-4 shrink-0 ${config.color}`} />
          <Badge variant="outline" className={`text-xs ${config.color} border-current/20`}>
            {config.label}
          </Badge>
          <span className="text-xs text-muted-foreground font-mono">
            #{event.sequence}
          </span>
          <span className="text-xs text-muted-foreground truncate">
            {event.system_id}
          </span>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="text-xs text-muted-foreground font-mono flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  +{formatNs(event.timestamp_ns - startNs)}
                </span>
              </TooltipTrigger>
              <TooltipContent>
                <p className="text-xs">{formatTimestamp(event.timestamp_ns)}</p>
                <p className="text-xs">+{relativeMs.toFixed(3)}ms from start</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
          {isExpanded ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
      </button>

      {isExpanded && (
        <div className="px-4 pb-3">
          <Separator className="mb-3 bg-white/[0.06]" />
          <pre className="text-xs text-muted-foreground bg-black/20 rounded-md p-3 overflow-x-auto max-h-64 overflow-y-auto font-mono whitespace-pre-wrap break-all">
            {JSON.stringify(payload, null, 2)}
          </pre>
          {event.parent && (
            <p className="text-xs text-muted-foreground mt-2 flex items-center gap-1">
              <GitBranch className="h-3 w-3" />
              Parent: <span className="font-mono">{event.parent}</span>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function TraceTimeline({ events, startNs }: { events: AtlasEvent[]; startNs: number }) {
  const endNs = events.length > 0 ? events[events.length - 1].timestamp_ns : startNs;
  const totalNs = endNs - startNs || 1;

  return (
    <div className="relative h-8 bg-black/20 rounded-lg overflow-hidden">
      {events.map((event) => {
        const config = kindConfig[event.kind] || kindConfig.input;
        const left = ((event.timestamp_ns - startNs) / totalNs) * 100;
        return (
          <TooltipProvider key={event.event_id}>
            <Tooltip>
              <TooltipTrigger asChild>
                <div
                  className={`absolute top-1 bottom-1 w-2 rounded-full ${config.color.replace('text-', 'bg-')} cursor-pointer hover:scale-y-125 transition-transform`}
                  style={{ left: `${Math.max(0, Math.min(99, left))}%` }}
                />
              </TooltipTrigger>
              <TooltipContent>
                <p className="text-xs font-medium">{config.label} #{event.sequence}</p>
                <p className="text-xs text-muted-foreground">+{formatNs(event.timestamp_ns - startNs)}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        );
      })}
    </div>
  );
}

function DecisionGraphView({ graph }: { graph: AtlasGraphResponse }) {
  if (!graph.nodes.length) {
    return (
      <div className="text-center py-8 text-muted-foreground text-sm">
        No decision graph data available
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-xs text-muted-foreground mb-4">
        <GitBranch className="h-3.5 w-3.5" />
        <span>{graph.nodes.length} nodes, {graph.edges.length} edges</span>
      </div>
      <ScrollArea className="h-[400px]">
        <div className="space-y-1">
          {graph.nodes.map((node) => {
            const config = kindConfig[node.kind] || kindConfig.input;
            const Icon = config.icon;
            const parentEdge = graph.edges.find((e) => e.to === node.id);
            return (
              <div
                key={node.id}
                className={`flex items-center gap-3 px-3 py-2 rounded-md border ${config.bg}`}
              >
                {parentEdge && (
                  <ArrowRight className="h-3 w-3 text-muted-foreground shrink-0" />
                )}
                {!parentEdge && <div className="w-3 shrink-0" />}
                <Icon className={`h-4 w-4 shrink-0 ${config.color}`} />
                <div className="min-w-0 flex-1">
                  <span className="text-sm font-medium">{config.label}</span>
                  <span className="text-xs text-muted-foreground ml-2 font-mono">
                    {node.system_id}
                  </span>
                </div>
                <span className="text-xs text-muted-foreground font-mono">
                  #{node.sequence}
                </span>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}

function TraceStats({ events, run }: { events: AtlasEvent[]; run: AtlasRunRecord }) {
  const startNs = events.length > 0 ? events[0].timestamp_ns : run.created_at_ns;
  const endNs = events.length > 0 ? events[events.length - 1].timestamp_ns : run.created_at_ns;
  const durationNs = endNs - startNs;

  const kindCounts = events.reduce(
    (acc, e) => {
      acc[e.kind] = (acc[e.kind] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );

  const hasError = kindCounts.error && kindCounts.error > 0;

  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div className="bg-black/10 rounded-lg p-3">
        <p className="text-xs text-muted-foreground">Duration</p>
        <p className="text-lg font-semibold font-mono">{formatNs(durationNs)}</p>
      </div>
      <div className="bg-black/10 rounded-lg p-3">
        <p className="text-xs text-muted-foreground">Events</p>
        <p className="text-lg font-semibold font-mono">{events.length}</p>
      </div>
      <div className="bg-black/10 rounded-lg p-3">
        <p className="text-xs text-muted-foreground">Status</p>
        <p className={`text-lg font-semibold ${hasError ? 'text-red-400' : 'text-emerald-400'}`}>
          {hasError ? 'Error' : 'OK'}
        </p>
      </div>
      <div className="bg-black/10 rounded-lg p-3">
        <p className="text-xs text-muted-foreground">Run ID</p>
        <p className="text-xs font-mono truncate mt-1" title={run.run_id}>
          {run.run_id.slice(0, 12)}...
        </p>
      </div>
    </div>
  );
}

export function ExecutionTraceViewer({ run, events, graph, className }: ExecutionTraceViewerProps) {
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());
  const [view, setView] = useState<'timeline' | 'graph'>('timeline');

  const startNs = events.length > 0 ? events[0].timestamp_ns : run.created_at_ns;

  const toggleEvent = (eventId: string) => {
    setExpandedEvents((prev) => {
      const next = new Set(prev);
      if (next.has(eventId)) {
        next.delete(eventId);
      } else {
        next.add(eventId);
      }
      return next;
    });
  };

  const labels = run.labels || {};
  const functionName = (labels.function_name as string) || 'unknown';
  const author = (labels.author as string) || '';
  const runtime = (labels.runtime as string) || '';

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Play className="h-5 w-5 text-primary" />
            <div>
              <CardTitle className="text-base">
                Execution Trace
                {author && functionName !== 'unknown' && (
                  <span className="text-muted-foreground font-normal ml-2">
                    {author}/{functionName}
                  </span>
                )}
              </CardTitle>
              <div className="flex items-center gap-2 mt-1">
                {runtime && (
                  <Badge variant="outline" className="text-xs">
                    {runtime}
                  </Badge>
                )}
                <span className="text-xs text-muted-foreground font-mono">
                  {run.run_id}
                </span>
              </div>
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <TraceStats events={events} run={run} />

        <Tabs value={view} onValueChange={(v) => setView(v as 'timeline' | 'graph')}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="timeline" className="gap-1.5">
              <Layers className="h-3.5 w-3.5" />
              Timeline
            </TabsTrigger>
            <TabsTrigger value="graph" className="gap-1.5">
              <GitBranch className="h-3.5 w-3.5" />
              Graph
            </TabsTrigger>
          </TabsList>

          <TabsContent value="timeline" className="space-y-3 mt-3">
            {events.length > 1 && (
              <TraceTimeline events={events} startNs={startNs} />
            )}
            <ScrollArea className="h-[500px]">
              <div className="space-y-2 pr-4">
                {events.map((event) => (
                  <EventCard
                    key={event.event_id}
                    event={event}
                    startNs={startNs}
                    isExpanded={expandedEvents.has(event.event_id)}
                    onToggle={() => toggleEvent(event.event_id)}
                  />
                ))}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="graph" className="mt-3">
            {graph ? (
              <DecisionGraphView graph={graph} />
            ) : (
              <div className="text-center py-8 text-muted-foreground text-sm">
                Loading decision graph...
              </div>
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}

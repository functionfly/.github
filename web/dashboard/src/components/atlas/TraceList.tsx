import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Eye,
  Layers,
  RefreshCw,
} from 'lucide-react';
import { useAtlasTraces, useAtlasTrace } from '@/hooks/useAtlasTraces';
import { ExecutionTraceViewer } from './ExecutionTraceViewer';
import type { AtlasRunRecord } from '@/api/atlas';

function formatRelativeTime(ns: number): string {
  const now = Date.now() * 1_000_000;
  const diffMs = (now - ns) / 1_000_000;
  if (diffMs < 60_000) return 'just now';
  if (diffMs < 3_600_000) return `${Math.floor(diffMs / 60_000)}m ago`;
  if (diffMs < 86_400_000) return `${Math.floor(diffMs / 3_600_000)}h ago`;
  return `${Math.floor(diffMs / 86_400_000)}d ago`;
}

function TraceRow({ run, onSelect, isSelected }: {
  run: AtlasRunRecord;
  onSelect: () => void;
  isSelected: boolean;
}) {
  const labels = run.labels || {};
  const functionName = (labels.function_name as string) || 'unknown';
  const author = (labels.author as string) || '';
  const runtime = (labels.runtime as string) || '';
  const isError = (labels.type as string) === 'error';

  return (
    <button
      onClick={onSelect}
      className={`w-full text-left px-4 py-3 border-b border-white/[0.04] hover:bg-white/[0.02] transition-colors ${
        isSelected ? 'bg-primary/5 border-l-2 border-l-primary' : ''
      }`}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            {isError ? (
              <AlertCircle className="h-3.5 w-3.5 text-red-400 shrink-0" />
            ) : (
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
            )}
            <span className="text-sm font-medium truncate">
              {author ? `${author}/` : ''}{functionName}
            </span>
            {runtime && (
              <Badge variant="outline" className="text-[10px] px-1 py-0">
                {runtime}
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-3 mt-1">
            <span className="text-xs text-muted-foreground font-mono">
              {run.run_id.slice(0, 16)}...
            </span>
            <span className="text-xs text-muted-foreground flex items-center gap-1">
              <Layers className="h-3 w-3" />
              {run.event_count} events
            </span>
          </div>
        </div>
        <span className="text-xs text-muted-foreground flex items-center gap-1 shrink-0">
          <Clock className="h-3 w-3" />
          {formatRelativeTime(run.created_at_ns)}
        </span>
      </div>
    </button>
  );
}

export function TraceList({ functionFilter }: { functionFilter?: { author: string; name: string } }) {
  const { data, isLoading, refetch } = useAtlasTraces(100);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const { data: selectedTrace, isLoading: traceLoading } = useAtlasTrace(selectedRunId || '');

  const runs = data?.runs || [];

  const filteredRuns = functionFilter
    ? runs.filter((r) => {
        const labels = r.labels || {};
        return (
          (labels.author as string) === functionFilter.author &&
          (labels.function_name as string) === functionFilter.name
        );
      })
    : runs;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[380px_1fr] gap-4">
      <Card className="h-fit">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm flex items-center gap-2">
              <Layers className="h-4 w-4" />
              Traces
              <Badge variant="secondary" className="text-xs">
                {filteredRuns.length}
              </Badge>
            </CardTitle>
            <Button variant="ghost" size="sm" onClick={() => refetch()} className="h-7 px-2">
              <RefreshCw className="h-3.5 w-3.5" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {filteredRuns.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground text-sm px-4">
              <Layers className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No traces recorded yet</p>
              <p className="text-xs mt-1">Traces appear when ATLAS_URL is configured</p>
            </div>
          ) : (
            <div className="max-h-[600px] overflow-y-auto">
              {filteredRuns.map((run) => (
                <TraceRow
                  key={run.run_id}
                  run={run}
                  onSelect={() => setSelectedRunId(run.run_id)}
                  isSelected={selectedRunId === run.run_id}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div>
        {selectedRunId && traceLoading && (
          <Card>
            <CardContent className="flex items-center justify-center py-12">
              <LoadingSpinner />
            </CardContent>
          </Card>
        )}
        {selectedTrace && (
          <ExecutionTraceViewer
            run={selectedTrace.run}
            events={selectedTrace.events}
            graph={selectedTrace.graph}
          />
        )}
        {!selectedRunId && (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <Eye className="h-10 w-10 mb-3 opacity-40" />
              <p className="text-sm">Select a trace to view details</p>
              <p className="text-xs mt-1">Click any trace in the list to see its events and decision graph</p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

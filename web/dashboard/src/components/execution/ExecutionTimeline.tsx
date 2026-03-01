import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Clock,
  Code,
  Database,
  Download,
  Network,
  Server,
  Upload,
  Zap,
} from 'lucide-react';

export interface ExecutionPhase {
  id: string;
  name: string;
  durationMs: number;
  startTime: number;
  endTime: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  metadata?: Record<string, string | number>;
}

export interface ExecutionTimelineProps {
  phases: ExecutionPhase[];
  totalDurationMs: number;
  coldStart?: boolean;
  showDetails?: boolean;
  className?: string;
}

const getPhaseIcon = (phaseName: string) => {
  const name = phaseName.toLowerCase();
  if (name.includes('init') || name.includes('startup')) return Zap;
  if (name.includes('compile') || name.includes('build')) return Code;
  if (name.includes('download') || name.includes('fetch')) return Download;
  if (name.includes('upload') || name.includes('response')) return Upload;
  if (name.includes('query') || name.includes('database') || name.includes('db')) return Database;
  if (name.includes('network') || name.includes('http') || name.includes('api')) return Network;
  return Server;
};

const getStatusColor = (status: ExecutionPhase['status']) => {
  switch (status) {
    case 'pending':
      return 'bg-gray-400';
    case 'running':
      return 'bg-blue-500 animate-pulse';
    case 'completed':
      return 'bg-green-500';
    case 'failed':
      return 'bg-red-500';
    default:
      return 'bg-gray-400';
  }
};

export function ExecutionTimeline({
  phases,
  totalDurationMs,
  coldStart = false,
  showDetails = true,
  className = '',
}: ExecutionTimelineProps) {
  const calculateWidth = (durationMs: number) => {
    if (totalDurationMs === 0) return 0;
    return Math.max((durationMs / totalDurationMs) * 100, 1);
  };

  const calculateOffset = (startTime: number) => {
    if (totalDurationMs === 0) return 0;
    return (startTime / totalDurationMs) * 100;
  };

  const formatDuration = (ms: number) => {
    if (ms < 1) return '<1ms';
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Execution Timeline</CardTitle>
            <CardDescription>
              {coldStart && (
                <Badge variant="outline" className="mt-1 bg-blue-50 text-blue-700 border-blue-200">
                  <Zap className="mr-1 h-3 w-3" />
                  Cold Start
                </Badge>
              )}
            </CardDescription>
          </div>
          <div className="text-right">
            <p className="text-2xl font-bold">{formatDuration(totalDurationMs)}</p>
            <p className="text-xs text-muted-foreground">Total time</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Timeline Bar */}
        <div className="relative h-8 bg-secondary rounded-md overflow-hidden">
          {phases.map((phase, index) => (
            <TooltipProvider key={phase.id}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div
                    className={`absolute h-full ${getStatusColor(phase.status)} transition-all hover:opacity-80`}
                    style={{
                      left: `${calculateOffset(phase.startTime)}%`,
                      width: `${calculateWidth(phase.durationMs)}%`,
                    }}
                  >
                    {phase.durationMs > 10 && (
                      <div className="flex items-center justify-center h-full">
                        <span className="text-[10px] text-white font-medium truncate px-1">
                          {phase.name}
                        </span>
                      </div>
                    )}
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <div className="space-y-1">
                    <p className="font-medium">{phase.name}</p>
                    <p className="text-xs text-muted-foreground">
                      Duration: {formatDuration(phase.durationMs)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Status: {phase.status}
                    </p>
                    {phase.metadata && Object.keys(phase.metadata).length > 0 && (
                      <div className="pt-1 border-t mt-1">
                        {Object.entries(phase.metadata).map(([key, value]) => (
                          <p key={key} className="text-xs">
                            {key}: {String(value)}
                          </p>
                        ))}
                      </div>
                    )}
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          ))}
        </div>

        {/* Legend */}
        <div className="flex flex-wrap gap-3 text-xs">
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-gray-400" />
            <span>Pending</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
            <span>Running</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-green-500" />
            <span>Completed</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-red-500" />
            <span>Failed</span>
          </div>
        </div>

        {/* Phase Details */}
        {showDetails && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium">Phase Details</h4>
            <div className="space-y-2">
              {phases.map((phase) => {
                const Icon = getPhaseIcon(phase.name);
                return (
                  <div
                    key={phase.id}
                    className="flex items-center justify-between p-2 rounded-md hover:bg-secondary/50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-full ${getStatusColor(phase.status)}/10`}>
                        <Icon className={`h-4 w-4 ${getStatusColor(phase.status).replace('bg-', 'text-')}`} />
                      </div>
                      <div>
                        <p className="font-medium text-sm">{phase.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {phase.status === 'running'
                            ? 'In progress...'
                            : `${formatDuration(phase.durationMs)} (${calculateWidth(phase.durationMs).toFixed(1)}%)`}
                        </p>
                      </div>
                    </div>
                    <Badge
                      variant={
                        phase.status === 'completed'
                          ? 'default'
                          : phase.status === 'failed'
                          ? 'destructive'
                          : 'secondary'
                      }
                      className="text-xs"
                    >
                      {phase.status}
                    </Badge>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Demo component for testing
export function ExecutionTimelineDemo() {
  const demoPhases: ExecutionPhase[] = [
    {
      id: '1',
      name: 'Container Start',
      durationMs: 150,
      startTime: 0,
      endTime: 150,
      status: 'completed',
    },
    {
      id: '2',
      name: 'Runtime Init',
      durationMs: 80,
      startTime: 150,
      endTime: 230,
      status: 'completed',
    },
    {
      id: '3',
      name: 'Code Load',
      durationMs: 45,
      startTime: 230,
      endTime: 275,
      status: 'completed',
    },
    {
      id: '4',
      name: 'Function Execution',
      durationMs: 120,
      startTime: 275,
      endTime: 395,
      status: 'completed',
      metadata: {
        'Memory Used': '45MB',
        'CPU Time': '95ms',
      },
    },
    {
      id: '5',
      name: 'Response',
      durationMs: 15,
      startTime: 395,
      endTime: 410,
      status: 'completed',
    },
  ];

  return (
    <ExecutionTimeline
      phases={demoPhases}
      totalDurationMs={410}
      coldStart={true}
      showDetails={true}
    />
  );
}

export default ExecutionTimeline;

/**
 * CollapsibleExecutionPanel Component
 * Enhanced bottom panel for execution monitoring with drawer, tabs, and metrics
 */

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  ChevronUp,
  ChevronDown,
  Clock,
  DollarSign,
  Cpu,
  MemoryStick,
  Play,
  CheckCircle2,
  AlertCircle,
  Terminal,
  BarChart3,
  List,
  Bug,
  FileJson,
  XCircle,
  RotateCcw,
  Download,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

interface ExecutionEvent {
  id: string;
  timestamp: string;
  type: 'info' | 'success' | 'error' | 'warning' | 'node';
  message: string;
  nodeId?: string;
  duration?: number;
}

interface NodeResult {
  nodeId: string;
  nodeName: string;
  status: 'success' | 'error' | 'running' | 'pending';
  duration: number;
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: string;
}

interface ExecutionMetrics {
  totalDuration: number;
  costUSD: number;
  memoryMB: number;
  cpuPercent: number;
  nodeCount: number;
  completedNodes: number;
  errorCount: number;
}

interface CollapsibleExecutionPanelProps {
  isExpanded?: boolean;
  onToggle?: () => void;
  status?: 'idle' | 'running' | 'completed' | 'error';
  progress?: number;
  metrics?: ExecutionMetrics;
  events?: ExecutionEvent[];
  nodeResults?: NodeResult[];
  onRerun?: () => void;
  onExportLogs?: () => void;
  onRun?: () => void;
  onStop?: () => void;
  onStep?: () => void;
}

function StatusBadge({ status }: { status: string }) {
  const config = {
    idle: { color: 'bg-gray-500', label: 'Idle' },
    running: { color: 'bg-blue-500', label: 'Running' },
    completed: { color: 'bg-green-500', label: 'Completed' },
    error: { color: 'bg-red-500', label: 'Error' },
  };
  const { color, label } = config[status as keyof typeof config] || config.idle;

  return (
    <Badge className={cn("text-white", color)}>
      {status === 'running' && (
        <span className="w-1.5 h-1.5 bg-white rounded-full mr-1.5 animate-pulse" />
      )}
      {label}
    </Badge>
  );
}

function MetricCard({
  icon: Icon,
  label,
  value,
  subValue,
  color,
}: {
  icon: React.ComponentType<any>;
  label: string;
  value: string | number;
  subValue?: string;
  color: string;
}) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg bg-[var(--bg-tertiary)]">
      <div className={cn("w-10 h-10 rounded-lg flex items-center justify-center", color)}>
        <Icon className="w-5 h-5 text-white" />
      </div>
      <div>
        <p className="text-xs text-[var(--text-muted)]">{label}</p>
        <p className="text-lg font-semibold text-[var(--text-primary)]">{value}</p>
        {subValue && (
          <p className="text-xs text-[var(--text-secondary)]">{subValue}</p>
        )}
      </div>
    </div>
  );
}

function EventRow({ event }: { event: ExecutionEvent }) {
  const icons = {
    info: <Terminal className="w-4 h-4 text-blue-500" />,
    success: <CheckCircle2 className="w-4 h-4 text-green-500" />,
    error: <XCircle className="w-4 h-4 text-red-500" />,
    warning: <AlertCircle className="w-4 h-4 text-yellow-500" />,
    node: <Play className="w-4 h-4 text-purple-500" />,
  };

  return (
    <div className="flex items-start gap-3 py-2 px-3 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors">
      <div className="mt-0.5">{icons[event.type]}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-xs text-[var(--text-muted)] font-mono">
            {new Date(event.timestamp).toLocaleTimeString()}
          </span>
          {event.nodeId && (
            <Badge variant="secondary" className="text-[10px]">
              Node {event.nodeId.slice(0, 6)}
            </Badge>
          )}
          {event.duration && (
            <span className="text-xs text-[var(--text-secondary)]">
              {event.duration}ms
            </span>
          )}
        </div>
        <p className={cn(
          "text-sm mt-0.5",
          event.type === 'error' ? 'text-red-400' : 'text-[var(--text-primary)]'
        )}>
          {event.message}
        </p>
      </div>
    </div>
  );
}

// Mock data for demonstration
const mockEvents: ExecutionEvent[] = [
  {
    id: '1',
    timestamp: '2026-04-10T19:30:00Z',
    type: 'info',
    message: 'Execution started',
  },
  {
    id: '2',
    timestamp: '2026-04-10T19:30:01Z',
    type: 'node',
    message: 'Input validation completed',
    nodeId: 'node-input-1',
    duration: 45,
  },
  {
    id: '3',
    timestamp: '2026-04-10T19:30:02Z',
    type: 'node',
    message: 'Processing with GPT-4',
    nodeId: 'node-ai-1',
    duration: 1200,
  },
  {
    id: '4',
    timestamp: '2026-04-10T19:30:03Z',
    type: 'success',
    message: 'AI processing completed successfully',
  },
  {
    id: '5',
    timestamp: '2026-04-10T19:30:04Z',
    type: 'node',
    message: 'Output formatting',
    nodeId: 'node-output-1',
    duration: 89,
  },
];

const mockNodeResults: NodeResult[] = [
  {
    nodeId: 'node-input-1',
    nodeName: 'HTTP Input',
    status: 'success',
    duration: 45,
    input: { body: '{"prompt": "Hello AI"}' },
    output: { validated: true },
  },
  {
    nodeId: 'node-ai-1',
    nodeName: 'GPT-4 Processor',
    status: 'success',
    duration: 1200,
    input: { prompt: 'Hello AI' },
    output: { response: 'Hello! How can I assist you today?' },
  },
  {
    nodeId: 'node-transform-1',
    nodeName: 'Data Transformer',
    status: 'success',
    duration: 89,
    input: { data: 'raw' },
    output: { data: 'formatted' },
  },
  {
    nodeId: 'node-output-1',
    nodeName: 'Response Output',
    status: 'running',
    duration: 0,
  },
];

const mockMetrics: ExecutionMetrics = {
  totalDuration: 1334,
  costUSD: 0.0234,
  memoryMB: 256,
  cpuPercent: 12.5,
  nodeCount: 4,
  completedNodes: 3,
  errorCount: 0,
};

export function CollapsibleExecutionPanel({
  isExpanded,
  onToggle,
  status = 'running',
  progress = 75,
  metrics = mockMetrics,
  events = mockEvents,
  nodeResults = mockNodeResults,
  onRerun,
  onExportLogs,
}: CollapsibleExecutionPanelProps) {
  const [activeTab, setActiveTab] = useState('logs');

  return (
    <div className="absolute bottom-0 left-0 right-0 z-20">
      {/* Collapsed Bar */}
      {!isExpanded && (
        <motion.div
          initial={{ y: 100 }}
          animate={{ y: 0 }}
          className="bg-[var(--bg-secondary)] border-t border-[var(--border-subtle)] px-4 py-2 flex items-center justify-between"
        >
          <div className="flex items-center gap-4">
            <StatusBadge status={status} />
            <div className="flex items-center gap-2 w-48">
              <Progress value={progress} className="h-1.5" />
              <span className="text-xs text-[var(--text-muted)]">{progress}%</span>
            </div>
            <Separator orientation="vertical" className="h-4" />
            <div className="flex items-center gap-3 text-xs text-[var(--text-secondary)]">
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {metrics.totalDuration}ms
              </span>
              <span className="flex items-center gap-1">
                <DollarSign className="w-3 h-3" />
                ${metrics.costUSD.toFixed(4)}
              </span>
              <span className="flex items-center gap-1">
                <CheckCircle2 className="w-3 h-3" />
                {metrics.completedNodes}/{metrics.nodeCount} nodes
              </span>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onToggle}>
            <ChevronUp className="w-4 h-4" />
            Expand
          </Button>
        </motion.div>
      )}

      {/* Expanded Panel */}
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ height: 0 }}
            animate={{ height: 'auto' }}
            exit={{ height: 0 }}
            className="bg-[var(--bg-secondary)] border-t border-[var(--border-subtle)] overflow-hidden"
          >
            <div className="h-[400px] flex flex-col">
              {/* Header */}
              <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--border-subtle)]">
                <div className="flex items-center gap-4">
                  <StatusBadge status={status} />
                  <div className="flex items-center gap-2 w-48">
                    <Progress value={progress} className="h-1.5" />
                    <span className="text-xs text-[var(--text-muted)]">{progress}%</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onRerun}
                    disabled={status === 'running'}
                  >
                    <RotateCcw className="w-4 h-4 mr-1" />
                    Re-run
                  </Button>
                  <Button variant="ghost" size="sm" onClick={onExportLogs}>
                    <Download className="w-4 h-4 mr-1" />
                    Export
                  </Button>
                  <Button variant="ghost" size="sm" onClick={onToggle}>
                    <ChevronDown className="w-4 h-4" />
                    Collapse
                  </Button>
                </div>
              </div>

              {/* Content */}
              <div className="flex-1 overflow-hidden">
                <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
                  <TabsList className="px-4 pt-2 justify-start">
                    <TabsTrigger value="logs" className="gap-2">
                      <Terminal className="w-4 h-4" />
                      Logs
                      <Badge variant="secondary" className="ml-1 text-[10px]">
                        {events.length}
                      </Badge>
                    </TabsTrigger>
                    <TabsTrigger value="metrics" className="gap-2">
                      <BarChart3 className="w-4 h-4" />
                      Metrics
                    </TabsTrigger>
                    <TabsTrigger value="results" className="gap-2">
                      <List className="w-4 h-4" />
                      Node Results
                    </TabsTrigger>
                    <TabsTrigger value="errors" className="gap-2">
                      <Bug className="w-4 h-4" />
                      Errors
                      {metrics.errorCount > 0 && (
                        <Badge variant="destructive" className="ml-1 text-[10px]">
                          {metrics.errorCount}
                        </Badge>
                      )}
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value="logs" className="flex-1 m-0 p-4 pt-0 overflow-hidden">
                    <ScrollArea className="h-full">
                      <div className="space-y-1">
                        {events.map((event) => (
                          <EventRow key={event.id} event={event} />
                        ))}
                      </div>
                    </ScrollArea>
                  </TabsContent>

                  <TabsContent value="metrics" className="flex-1 m-0 p-4 pt-0">
                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                      <MetricCard
                        icon={Clock}
                        label="Execution Time"
                        value={`${metrics.totalDuration}ms`}
                        subValue="Total duration"
                        color="bg-blue-500"
                      />
                      <MetricCard
                        icon={DollarSign}
                        label="Cost"
                        value={`$${metrics.costUSD.toFixed(4)}`}
                        subValue="USD"
                        color="bg-green-500"
                      />
                      <MetricCard
                        icon={MemoryStick}
                        label="Memory"
                        value={`${metrics.memoryMB}MB`}
                        subValue="Peak usage"
                        color="bg-purple-500"
                      />
                      <MetricCard
                        icon={Cpu}
                        label="CPU"
                        value={`${metrics.cpuPercent}%`}
                        subValue="Average load"
                        color="bg-orange-500"
                      />
                    </div>

                    {/* Timeline visualization */}
                    <Card className="mt-4 border-[var(--border-subtle)]">
                      <CardHeader className="p-3 pb-2">
                        <CardTitle className="text-sm">Execution Timeline</CardTitle>
                      </CardHeader>
                      <CardContent className="p-3 pt-0">
                        <div className="relative h-16 bg-[var(--bg-tertiary)] rounded-lg overflow-hidden">
                          {nodeResults.map((result, index) => {
                            const startPercent = (index / nodeResults.length) * 100;
                            const widthPercent = (1 / nodeResults.length) * 100 - 1;
                            return (
                              <div
                                key={result.nodeId}
                                className={cn(
                                  "absolute top-2 bottom-2 rounded",
                                  result.status === 'success' && 'bg-green-500',
                                  result.status === 'error' && 'bg-red-500',
                                  result.status === 'running' && 'bg-blue-500 animate-pulse',
                                  result.status === 'pending' && 'bg-gray-500'
                                )}
                                style={{
                                  left: `${startPercent}%`,
                                  width: `${widthPercent}%`,
                                }}
                              >
                                <div className="absolute inset-0 flex items-center justify-center">
                                  <span className="text-[10px] text-white font-medium truncate px-1">
                                    {result.nodeName}
                                  </span>
                                </div>
                              </div>
                            );
                          })}
                        </div>
                        <div className="flex justify-between mt-2 text-xs text-[var(--text-muted)]">
                          <span>0ms</span>
                          <span>{Math.round(metrics.totalDuration / 2)}ms</span>
                          <span>{metrics.totalDuration}ms</span>
                        </div>
                      </CardContent>
                    </Card>
                  </TabsContent>

                  <TabsContent value="results" className="flex-1 m-0 p-4 pt-0 overflow-hidden">
                    <ScrollArea className="h-full">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Node</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Duration</TableHead>
                            <TableHead>Output Preview</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {nodeResults.map((result) => (
                            <TableRow key={result.nodeId}>
                              <TableCell className="font-medium">
                                <div className="flex items-center gap-2">
                                  <FileJson className="w-4 h-4 text-[var(--text-muted)]" />
                                  {result.nodeName}
                                </div>
                                <span className="text-xs text-[var(--text-muted)]">
                                  {result.nodeId.slice(0, 8)}
                                </span>
                              </TableCell>
                              <TableCell>
                                <Badge
                                  variant={
                                    result.status === 'success'
                                      ? 'default'
                                      : result.status === 'error'
                                      ? 'destructive'
                                      : 'secondary'
                                  }
                                  className="text-xs"
                                >
                                  {result.status}
                                </Badge>
                              </TableCell>
                              <TableCell>
                                {result.duration > 0 ? `${result.duration}ms` : '—'}
                              </TableCell>
                              <TableCell>
                                <div className="text-xs text-[var(--text-muted)] max-w-xs truncate">
                                  {result.output
                                    ? JSON.stringify(result.output).slice(0, 50)
                                    : result.error || 'No output yet'}
                                </div>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </ScrollArea>
                  </TabsContent>

                  <TabsContent value="errors" className="flex-1 m-0 p-4 pt-0 overflow-hidden">
                    <ScrollArea className="h-full">
                      {metrics.errorCount === 0 ? (
                        <div className="flex flex-col items-center justify-center h-32 text-[var(--text-muted)]">
                          <CheckCircle2 className="w-8 h-8 mb-2 text-green-500" />
                          <p>No errors found</p>
                        </div>
                      ) : (
                        <div className="space-y-2">
                          {nodeResults
                            .filter((r) => r.error)
                            .map((result) => (
                              <Card
                                key={result.nodeId}
                                className="border-red-500/30 bg-red-500/5"
                              >
                                <CardContent className="p-3">
                                  <div className="flex items-start gap-3">
                                    <AlertCircle className="w-5 h-5 text-red-500 mt-0.5" />
                                    <div className="flex-1">
                                      <p className="font-medium text-red-400">
                                        {result.nodeName} ({result.nodeId.slice(0, 8)})
                                      </p>
                                      <p className="text-sm text-red-300 mt-1">{result.error}</p>
                                    </div>
                                  </div>
                                </CardContent>
                              </Card>
                            ))}
                        </div>
                      )}
                    </ScrollArea>
                  </TabsContent>
                </Tabs>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default CollapsibleExecutionPanel;

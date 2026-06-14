'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Clock,
  DollarSign,
  Bot,
  Hash,
  Calendar,
  Activity,
  Zap,
  AlertCircle,
  Info,
  X,
} from 'lucide-react';

interface RunMetadata {
  id: string;
  atlas_run_id: string;
  atlas_tenant_id: string;
  agent_id: string;
  agent_type: string;
  status: string;
  total_cost_usd: number;
  total_input_tokens: number;
  total_output_tokens: number;
  event_count: number;
  error_count: number;
  tool_call_count: number;
  started_at: string;
  ended_at?: string;
  span_id?: string;
  parent_span_id?: string;
}

interface RunMetadataPanelProps {
  run: RunMetadata | null;
  children?: React.ReactNode;
}

export default function RunMetadataPanel({ run, children }: RunMetadataPanelProps) {
  const [open, setOpen] = useState(false);

  if (!run) return null;

  const duration = run.ended_at
    ? Math.floor((new Date(run.ended_at).getTime() - new Date(run.started_at).getTime()) / 1000)
    : null;

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  };

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>{children}</SheetTrigger>
      <SheetContent className="w-[400px] sm:max-w-[400px] overflow-y-auto">
        <SheetHeader className="pb-4">
          <div className="flex items-center justify-between">
            <SheetTitle className="flex items-center gap-2">
              <Bot className="h-5 w-5" />
              Run Metadata
            </SheetTitle>
            <Button variant="ghost" size="sm" onClick={() => setOpen(false)}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          <SheetDescription>Detailed information about this run</SheetDescription>
        </SheetHeader>

        <Tabs defaultValue="overview" className="w-full">
          <TabsList className="w-full">
            <TabsTrigger value="overview" className="flex-1">Overview</TabsTrigger>
            <TabsTrigger value="tokens" className="flex-1">Tokens</TabsTrigger>
            <TabsTrigger value="raw" className="flex-1">Raw</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-4 pt-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Activity className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Status</p>
                  <Badge
                    variant={run.status === 'completed' ? 'default' : run.status === 'failed' ? 'destructive' : 'secondary'}
                  >
                    {run.status}
                  </Badge>
                </div>
              </div>

              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <DollarSign className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Cost</p>
                  <p className="font-medium">${run.total_cost_usd.toFixed(4)}</p>
                </div>
              </div>

              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Hash className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Events</p>
                  <p className="font-medium">{run.event_count}</p>
                </div>
              </div>

              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <AlertCircle className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Errors</p>
                  <p className="font-medium">{run.error_count}</p>
                </div>
              </div>
            </div>

            {duration && (
              <div className="flex items-center gap-2 p-3 rounded-lg border">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Duration</p>
                  <p className="font-medium">{formatDuration(duration)}</p>
                </div>
              </div>
            )}

            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground flex items-center gap-2">
                  <Calendar className="h-3 w-3" /> Started
                </span>
                <span>{new Date(run.started_at).toLocaleString()}</span>
              </div>

              {run.ended_at && (
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Calendar className="h-3 w-3" /> Ended
                  </span>
                  <span>{new Date(run.ended_at).toLocaleString()}</span>
                </div>
              )}

              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground flex items-center gap-2">
                  <Bot className="h-3 w-3" /> Agent ID
                </span>
                <span className="font-mono text-xs">{run.agent_id}</span>
              </div>

              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground flex items-center gap-2">
                  <Zap className="h-3 w-3" /> Agent Type
                </span>
                <Badge variant="outline">{run.agent_type}</Badge>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="tokens" className="space-y-4 pt-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <p className="text-2xl font-bold">{run.total_input_tokens.toLocaleString()}</p>
                <p className="text-xs text-muted-foreground">Input Tokens</p>
              </div>
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <p className="text-2xl font-bold">{run.total_output_tokens.toLocaleString()}</p>
                <p className="text-xs text-muted-foreground">Output Tokens</p>
              </div>
            </div>

            <div className="p-4 rounded-lg border text-center">
              <p className="text-2xl font-bold">
                {(run.total_input_tokens + run.total_output_tokens).toLocaleString()}
              </p>
              <p className="text-xs text-muted-foreground">Total Tokens</p>
            </div>

            <div className="p-4 rounded-lg border">
              <p className="text-xs text-muted-foreground mb-2">Cost per 1K tokens</p>
              <p className="text-lg font-bold">
                $
                {(
                  (run.total_cost_usd / (run.total_input_tokens + run.total_output_tokens)) *
                  1000
                ).toFixed(4)}
              </p>
            </div>
          </TabsContent>

          <TabsContent value="raw" className="pt-4">
            <pre className="p-4 rounded-lg bg-muted/50 text-xs overflow-x-auto">
              {JSON.stringify(run, null, 2)}
            </pre>
          </TabsContent>
        </Tabs>
      </SheetContent>
    </Sheet>
  );
}

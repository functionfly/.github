import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Progress } from '@/components/ui/progress';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  useCreateReplay,
  useIsReplayStreamingEnabled,
  useReplayProgress,
  useStateFabricReplay,
  useStateFabricReplays,
} from '@/hooks/useStateFabric';
import type { ReplaySession, Snapshot } from '@/types';
import {
  CheckCircle,
  ChevronDown,
  ChevronUp,
  Clock,
  Loader2,
  Plus,
  RotateCcw,
  XCircle,
} from 'lucide-react';
import { useState } from 'react';
import { StateFabricAddonGate } from './StateFabricAddonGate';

interface ReplayManagerProps {
  fabricId: string;
  snapshots: Snapshot[];
}

const statusConfig: Record<
  ReplaySession['status'],
  { label: string; icon: React.ReactNode; color: string }
> = {
  pending: {
    label: 'Pending',
    icon: <Clock className="w-3 h-3" />,
    color: 'bg-yellow-500/10 text-yellow-400',
  },
  running: {
    label: 'Running',
    icon: <Loader2 className="w-3 h-3 animate-spin" />,
    color: 'bg-blue-500/10 text-blue-400',
  },
  completed: {
    label: 'Completed',
    icon: <CheckCircle className="w-3 h-3" />,
    color: 'bg-green-500/10 text-green-400',
  },
  failed: {
    label: 'Failed',
    icon: <XCircle className="w-3 h-3" />,
    color: 'bg-red-500/10 text-red-400',
  },
};

function ReplayDetailRow({ fabricId, replay }: { fabricId: string; replay: ReplaySession }) {
  const [expanded, setExpanded] = useState(false);
  const isActive = replay.status === 'pending' || replay.status === 'running';
  const isStreamingEnabled = useIsReplayStreamingEnabled();
  const { data: liveReplay } = useStateFabricReplay(fabricId, replay.id);
  const { data: streamedReplay } = useReplayProgress(fabricId, replay.id);
  const detail = (isActive && isStreamingEnabled && streamedReplay) ? streamedReplay : (liveReplay ?? replay);
  const cfg = statusConfig[detail.status];

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-lg bg-brand-500/10 flex items-center justify-center shrink-0">
              <RotateCcw className="w-5 h-5 text-brand-400" />
            </div>
            <div className="min-w-0">
              <CardTitle className="text-base font-mono truncate">
                {detail.id.slice(0, 8)}…
              </CardTitle>
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                <Badge className={`text-xs gap-1 ${cfg.color}`}>
                  {cfg.icon}
                  {cfg.label}
                </Badge>
                <span className="text-xs text-text-muted">
                  {new Date(detail.startedAt).toLocaleString()}
                </span>
              </div>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={() => setExpanded(!expanded)}>
            {expanded ? (
              <>
                <ChevronUp className="w-4 h-4 mr-1" />
                Hide
              </>
            ) : (
              <>
                <ChevronDown className="w-4 h-4 mr-1" />
                Details
              </>
            )}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span className="text-text-muted">Progress</span>
            <span className="text-text-primary">{detail.progress}%</span>
          </div>
          <Progress value={detail.progress} className="h-2" />
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
          <div>
            <p className="text-text-muted text-xs">Events replayed</p>
            <p className="font-medium">{detail.eventsReplayed.toLocaleString()}</p>
          </div>
          {detail.snapshotId && (
            <div>
              <p className="text-text-muted text-xs">Snapshot</p>
              <p className="font-mono text-xs truncate">{detail.snapshotId.slice(0, 8)}…</p>
            </div>
          )}
          {detail.startEventId && (
            <div>
              <p className="text-text-muted text-xs">Start event</p>
              <p className="font-mono text-xs truncate">{detail.startEventId.slice(0, 8)}…</p>
            </div>
          )}
          {detail.endEventId && (
            <div>
              <p className="text-text-muted text-xs">End event</p>
              <p className="font-mono text-xs truncate">{detail.endEventId.slice(0, 8)}…</p>
            </div>
          )}
        </div>
        {expanded && detail.error && (
          <p className="text-sm text-red-400 bg-red-500/10 p-3 rounded-lg">{detail.error}</p>
        )}
        {expanded && detail.completedAt && (
          <p className="text-xs text-text-muted">
            Completed {new Date(detail.completedAt).toLocaleString()}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export function ReplayManager({ fabricId, snapshots }: ReplayManagerProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [replayMode, setReplayMode] = useState<'snapshot' | 'events'>('snapshot');
  const [selectedSnapshotId, setSelectedSnapshotId] = useState('');
  const [startEventId, setStartEventId] = useState('');
  const [endEventId, setEndEventId] = useState('');

  const { data: replays, isLoading } = useStateFabricReplays(fabricId);
  const createReplay = useCreateReplay(fabricId);

  const handleCreate = async () => {
    const payload =
      replayMode === 'snapshot'
        ? { snapshotId: selectedSnapshotId }
        : {
            startEventId: startEventId || undefined,
            endEventId: endEventId || undefined,
          };

    await createReplay.mutateAsync(payload);
    setIsCreateOpen(false);
    setSelectedSnapshotId('');
    setStartEventId('');
    setEndEventId('');
  };

  const canSubmit =
    replayMode === 'snapshot' ? !!selectedSnapshotId : !!startEventId || !!endEventId;

  return (
    <StateFabricAddonGate addonId="hot_cache_booster">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-text-primary">Replay Sessions</h3>
            <p className="text-sm text-text-muted">
              Deterministic replay from snapshots or event ranges
            </p>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button size="sm">
                <Plus className="w-4 h-4 mr-2" />
                New Replay
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Start Replay Session</DialogTitle>
              </DialogHeader>
              <Tabs
                value={replayMode}
                onValueChange={(v) => setReplayMode(v as 'snapshot' | 'events')}
              >
                <TabsList className="grid w-full grid-cols-2">
                  <TabsTrigger value="snapshot">From Snapshot</TabsTrigger>
                  <TabsTrigger value="events">Event Range</TabsTrigger>
                </TabsList>
                <TabsContent value="snapshot" className="space-y-4 pt-4">
                  <div className="space-y-2">
                    <Label>Snapshot</Label>
                    <Select value={selectedSnapshotId} onValueChange={setSelectedSnapshotId}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a snapshot" />
                      </SelectTrigger>
                      <SelectContent>
                        {snapshots.map((s) => (
                          <SelectItem key={s.id} value={s.id}>
                            {s.name} ({new Date(s.createdAt).toLocaleDateString()})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </TabsContent>
                <TabsContent value="events" className="space-y-4 pt-4">
                  <div className="space-y-2">
                    <Label htmlFor="startEvent">Start Event ID (optional)</Label>
                    <Input
                      id="startEvent"
                      value={startEventId}
                      onChange={(e) => setStartEventId(e.target.value)}
                      placeholder="UUID of first event"
                      className="font-mono text-sm"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="endEvent">End Event ID (optional)</Label>
                    <Input
                      id="endEvent"
                      value={endEventId}
                      onChange={(e) => setEndEventId(e.target.value)}
                      placeholder="UUID of last event"
                      className="font-mono text-sm"
                    />
                  </div>
                </TabsContent>
              </Tabs>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleCreate} disabled={!canSubmit || createReplay.isPending}>
                  {createReplay.isPending ? (
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  ) : (
                    <RotateCcw className="w-4 h-4 mr-2" />
                  )}
                  Start Replay
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <LoadingSpinner />
          </div>
        ) : !replays?.length ? (
          <Card className="p-8 text-center">
            <RotateCcw className="w-12 h-12 mx-auto mb-4 text-text-muted" />
            <p className="text-text-muted mb-2">No replay sessions yet</p>
            <p className="text-sm text-text-muted">
              Start a replay from a snapshot or event range to debug state changes
            </p>
          </Card>
        ) : (
          <div className="grid gap-4">
            {replays.map((replay) => (
              <ReplayDetailRow key={replay.id} fabricId={fabricId} replay={replay} />
            ))}
          </div>
        )}
      </div>
    </StateFabricAddonGate>
  );
}

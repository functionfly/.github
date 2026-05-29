import { versionsApi, type PlatformFunctionVersion } from '@/api/versions';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { GitCompare, History, RotateCcw, Tag, Upload, Archive, AlertTriangle } from 'lucide-react';
import { useMemo, useState } from 'react';
import { toast } from 'sonner';

interface FunctionVersionsSectionProps {
  functionId: string;
  author: string;
  name: string;
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'published':
      return 'default';
    case 'deprecated':
      return 'destructive';
    case 'archived':
      return 'secondary';
    default:
      return 'outline';
  }
}

export function FunctionVersionsSection({
  functionId,
  author,
  name,
}: FunctionVersionsSectionProps) {
  const queryClient = useQueryClient();
  const [comparePair, setComparePair] = useState<[string, string] | null>(null);
  const [deprecateTarget, setDeprecateTarget] = useState<PlatformFunctionVersion | null>(null);
  const [deprecateReason, setDeprecateReason] = useState('');
  const [rollbackTarget, setRollbackTarget] = useState<string | null>(null);

  const versionsKey = ['platform-versions', functionId];

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: versionsKey,
    queryFn: () => versionsApi.list(functionId),
    enabled: !!functionId,
  });

  const { data: rollbackData } = useQuery({
    queryKey: ['platform-rollbacks', functionId],
    queryFn: () => versionsApi.rollbackHistory(functionId),
    enabled: !!functionId,
  });

  const { data: compareData, isLoading: compareLoading } = useQuery({
    queryKey: ['platform-version-compare', functionId, comparePair?.[0], comparePair?.[1]],
    queryFn: () => versionsApi.compare(functionId, comparePair![0], comparePair![1]),
    enabled: !!comparePair,
  });

  const versions = useMemo(
    () => [...(data?.versions ?? [])].sort((a, b) => b.version.localeCompare(a.version)),
    [data?.versions]
  );

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: versionsKey });
    queryClient.invalidateQueries({ queryKey: ['function-versions', author, name] });
    queryClient.invalidateQueries({ queryKey: ['function', author, name] });
    queryClient.invalidateQueries({ queryKey: ['platform-rollbacks', functionId] });
  };

  const publishMutation = useMutation({
    mutationFn: (version: string) => versionsApi.publish(functionId, version),
    onSuccess: () => {
      toast.success('Version published');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Publish failed'),
  });

  const archiveMutation = useMutation({
    mutationFn: (version: string) => versionsApi.archive(functionId, version),
    onSuccess: () => {
      toast.success('Version archived');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Archive failed'),
  });

  const deprecateMutation = useMutation({
    mutationFn: () =>
      versionsApi.deprecate(functionId, deprecateTarget!.version, {
        reason: deprecateReason || 'Deprecated by owner',
      }),
    onSuccess: () => {
      toast.success('Version deprecated');
      setDeprecateTarget(null);
      setDeprecateReason('');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Deprecate failed'),
  });

  const aliasMutation = useMutation({
    mutationFn: ({ version, alias }: { version: string; alias: 'latest' | 'stable' }) =>
      versionsApi.setAlias(functionId, version, alias),
    onSuccess: (_, vars) => {
      toast.success(`Set ${vars.alias} alias to v${vars.version}`);
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Failed to set alias'),
  });

  const rollbackMutation = useMutation({
    mutationFn: (version: string) => versionsApi.rollbackToVersion(functionId, version),
    onSuccess: (res) => {
      toast.success(`Rolled back to v${res.toVersion}`);
      setRollbackTarget(null);
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Rollback failed'),
  });

  if (isLoading) {
    return <p className="text-sm text-text-muted">Loading versions…</p>;
  }

  if (error) {
    return (
      <div className="text-sm text-destructive">
        Failed to load versions.{' '}
        <button type="button" className="underline" onClick={() => refetch()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm text-text-secondary">
          Manage publish state, aliases, deprecation, and rollbacks for{' '}
          <span className="font-mono text-text-primary">
            {author}/{name}
          </span>
        </p>
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            versionsApi
              .rollbackLatest(functionId)
              .then(() => {
                toast.success('Rolled back to previous version');
                invalidate();
              })
              .catch((e: Error) => toast.error(e.message || 'Rollback failed'));
          }}
          disabled={versions.length < 2 || rollbackMutation.isPending}
        >
          <RotateCcw className="h-4 w-4 mr-1" />
          Roll back one step
        </Button>
      </div>

      <div className="space-y-3">
        {versions.length === 0 ? (
          <p className="text-sm text-text-muted">No versions yet. Publish from Studio or the registry CLI.</p>
        ) : (
          versions.map((v) => (
            <Card
              key={v.id}
              className={cn(v.isLatest && 'border-brand-500/40 bg-brand-500/5')}
            >
              <CardContent className="p-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-mono font-semibold">v{v.version}</span>
                    <Badge variant={statusVariant(v.status)}>{v.status}</Badge>
                    {v.isLatest && <Badge variant="secondary">latest</Badge>}
                    {v.isStable && <Badge variant="outline">stable</Badge>}
                  </div>
                  {v.publishedAt && (
                    <p className="text-xs text-text-muted">
                      Published {new Date(v.publishedAt).toLocaleString()}
                    </p>
                  )}
                  {v.deprecation?.reason && (
                    <p className="text-xs text-amber-500">{v.deprecation.reason}</p>
                  )}
                </div>

                <div className="flex flex-wrap gap-2">
                  {v.status !== 'published' && (
                    <Button
                      size="sm"
                      onClick={() => publishMutation.mutate(v.version)}
                      disabled={publishMutation.isPending}
                    >
                      <Upload className="h-3 w-3 mr-1" />
                      Publish
                    </Button>
                  )}
                  {v.status === 'published' && !v.isLatest && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => aliasMutation.mutate({ version: v.version, alias: 'latest' })}
                      disabled={aliasMutation.isPending}
                    >
                      <Tag className="h-3 w-3 mr-1" />
                      Set latest
                    </Button>
                  )}
                  {v.status === 'published' && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setRollbackTarget(v.version)}
                    >
                      <RotateCcw className="h-3 w-3 mr-1" />
                      Rollback here
                    </Button>
                  )}
                  {v.status === 'published' && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setDeprecateTarget(v)}
                    >
                      <AlertTriangle className="h-3 w-3 mr-1" />
                      Deprecate
                    </Button>
                  )}
                  {v.status !== 'archived' && (
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => archiveMutation.mutate(v.version)}
                      disabled={archiveMutation.isPending}
                    >
                      <Archive className="h-3 w-3 mr-1" />
                      Archive
                    </Button>
                  )}
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      const other = versions.find((x) => x.version !== v.version);
                      if (other) setComparePair([v.version, other.version]);
                    }}
                    disabled={versions.length < 2}
                  >
                    <GitCompare className="h-3 w-3 mr-1" />
                    Compare
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {rollbackData?.rollbacks && rollbackData.rollbacks.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <History className="h-4 w-4" />
              Rollback history
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {rollbackData.rollbacks.map((r) => (
              <div
                key={r.id}
                className="text-xs text-text-secondary flex justify-between border-b border-border-subtle pb-2 last:border-0"
              >
                <span>
                  v{r.from_version} → v{r.to_version} ({r.status})
                </span>
                <span>{new Date(r.initiated_at).toLocaleString()}</span>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <Dialog open={!!comparePair} onOpenChange={(open) => !open && setComparePair(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              Compare v{comparePair?.[0]} → v{comparePair?.[1]}
            </DialogTitle>
            <DialogDescription>Metadata and changelog differences</DialogDescription>
          </DialogHeader>
          {compareLoading && <p className="text-sm text-text-muted">Loading diff…</p>}
          {compareData && (
            <div className="space-y-2 text-sm max-h-64 overflow-y-auto">
              {(compareData.changes ?? []).map((c, i) => (
                <div key={i} className="font-mono text-xs">
                  {c.field}: {c.fromValue} → {c.toValue}
                </div>
              ))}
              {compareData.isBreaking && (
                <p className="text-amber-500 text-xs">Contains breaking changes</p>
              )}
              {(compareData.changes ?? []).length === 0 && (
                <p className="text-text-muted">No metadata differences recorded.</p>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setComparePair(null)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deprecateTarget} onOpenChange={(open) => !open && setDeprecateTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Deprecate v{deprecateTarget?.version}</DialogTitle>
            <DialogDescription>
              Callers should migrate to a newer version. You can set a replacement in the API later.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="deprecate-reason">Reason</Label>
            <Textarea
              id="deprecate-reason"
              value={deprecateReason}
              onChange={(e) => setDeprecateReason(e.target.value)}
              placeholder="Why this version is being deprecated"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeprecateTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deprecateMutation.mutate()}
              disabled={deprecateMutation.isPending}
            >
              Deprecate
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!rollbackTarget} onOpenChange={(open) => !open && setRollbackTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rollback to v{rollbackTarget}?</AlertDialogTitle>
            <AlertDialogDescription>
              This sets the latest alias to the selected version immediately. Active traffic routing
              may take a moment to converge.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => rollbackTarget && rollbackMutation.mutate(rollbackTarget)}
            >
              Rollback
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

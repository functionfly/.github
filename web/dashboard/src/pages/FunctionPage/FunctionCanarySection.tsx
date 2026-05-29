import { registryCanaryApi, type CanaryConfig } from '@/api/versions';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FlaskConical, Play, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

interface FunctionCanarySectionProps {
  author: string;
  name: string;
  latestVersion?: string;
}

function isActiveCanary(
  data: CanaryConfig | { message?: string } | undefined
): data is CanaryConfig {
  return !!data && 'id' in data && !!data.id;
}

export function FunctionCanarySection({ author, name, latestVersion }: FunctionCanarySectionProps) {
  const queryClient = useQueryClient();
  const canaryKey = ['registry-canary', author, name];

  const { data, isLoading, refetch } = useQuery({
    queryKey: canaryKey,
    queryFn: () => registryCanaryApi.get(author, name),
    enabled: !!author && !!name,
  });

  const active = isActiveCanary(data);

  const [version, setVersion] = useState(latestVersion ?? '');
  const [trafficPercent, setTrafficPercent] = useState(10);
  const [autoPromote, setAutoPromote] = useState(false);

  useEffect(() => {
    if (latestVersion) setVersion(latestVersion);
  }, [latestVersion]);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: canaryKey });

  const createMutation = useMutation({
    mutationFn: () =>
      registryCanaryApi.create(author, name, {
        version,
        traffic_percent: trafficPercent,
        auto_promote: autoPromote,
        promote_threshold: 0.99,
        promote_window: 300,
      }),
    onSuccess: () => {
      toast.success('Canary deployment started');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Failed to start canary'),
  });

  const promoteMutation = useMutation({
    mutationFn: () => registryCanaryApi.promote(author, name),
    onSuccess: () => {
      toast.success('Canary promoted');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Promote failed'),
  });

  const rollbackMutation = useMutation({
    mutationFn: () => registryCanaryApi.rollback(author, name),
    onSuccess: () => {
      toast.success('Canary rolled back');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Canary rollback failed'),
  });

  const cancelMutation = useMutation({
    mutationFn: () => registryCanaryApi.cancel(author, name),
    onSuccess: () => {
      toast.success('Canary cancelled');
      invalidate();
    },
    onError: (e: Error) => toast.error(e.message || 'Cancel failed'),
  });

  if (isLoading) {
    return <p className="text-sm text-text-muted">Loading canary status…</p>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <FlaskConical className="h-4 w-4 text-brand-500" />
          Canary deployment
        </CardTitle>
        <CardDescription>
          Route a percentage of traffic to a candidate version before full promotion.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {active ? (
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge>v{data.version}</Badge>
              <Badge variant="secondary">{data.traffic_percent}% traffic</Badge>
              {data.auto_promote && <Badge variant="outline">auto-promote</Badge>}
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                size="sm"
                onClick={() => promoteMutation.mutate()}
                disabled={promoteMutation.isPending}
              >
                <Play className="h-3 w-3 mr-1" />
                Promote
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => rollbackMutation.mutate()}
                disabled={rollbackMutation.isPending}
              >
                Rollback canary
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => cancelMutation.mutate()}
                disabled={cancelMutation.isPending}
              >
                <XCircle className="h-3 w-3 mr-1" />
                Cancel
              </Button>
            </div>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="canary-version">Candidate version</Label>
              <Input
                id="canary-version"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder="1.0.0"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="canary-traffic">Traffic %</Label>
              <Input
                id="canary-traffic"
                type="number"
                min={1}
                max={100}
                value={trafficPercent}
                onChange={(e) => setTrafficPercent(Number(e.target.value))}
              />
            </div>
            <div className="flex items-center gap-2 sm:col-span-2">
              <Switch id="canary-auto" checked={autoPromote} onCheckedChange={setAutoPromote} />
              <Label htmlFor="canary-auto">Auto-promote when healthy</Label>
            </div>
            <div className="sm:col-span-2">
              <Button
                onClick={() => createMutation.mutate()}
                disabled={!version || createMutation.isPending}
              >
                Start canary
              </Button>
              <Button variant="ghost" size="sm" className="ml-2" onClick={() => refetch()}>
                Refresh
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

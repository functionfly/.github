import { useQuery } from '@tanstack/react-query';
import { Shield, Check, X, Info, RefreshCw, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { getSandboxStatus, type SandboxStatus, getTierColor, getTierLabel } from '@/api/sandbox';

export function SandboxStatusPanel() {
  const { data: status, isLoading, error, refetch } = useQuery<SandboxStatus>({
    queryKey: ['sandbox-status'],
    queryFn: getSandboxStatus,
    refetchInterval: 60000,
    retry: 1,
  });

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-32">
          <RefreshCw className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex items-center gap-2 h-32 text-destructive">
          <AlertTriangle className="w-5 h-5" />
          <span className="text-sm">Failed to load sandbox status</span>
          <Button variant="ghost" size="sm" onClick={() => refetch()}>Retry</Button>
        </CardContent>
      </Card>
    );
  }

  if (!status) return null;

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <Shield className="w-4 h-4 text-primary" />
            Sandbox Security
          </CardTitle>
          <Badge className={cn('text-xs', getTierColor(status.active_tier as any))}>
            {getTierLabel(status.active_tier as any)} Active
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="flex items-center gap-2 p-2 rounded-md bg-muted/30">
            {status.gvisor_available ? (
              <Check className="w-4 h-4 text-green-500" />
            ) : (
              <X className="w-4 h-4 text-red-500" />
            )}
            <div>
              <p className="text-xs font-medium">gVisor</p>
              <p className="text-[10px] text-muted-foreground">
                {status.gvisor_version || 'Not installed'}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 p-2 rounded-md bg-muted/30">
            {status.docker_available ? (
              <Check className="w-4 h-4 text-green-500" />
            ) : (
              <X className="w-4 h-4 text-red-500" />
            )}
            <div>
              <p className="text-xs font-medium">Docker</p>
              <p className="text-[10px] text-muted-foreground">
                {status.docker_version || 'Not available'}
              </p>
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <h4 className="text-xs font-medium text-muted-foreground">Available Tiers</h4>
          {status.supported_tiers.map((tier) => (
            <div
              key={tier.id}
              className="flex items-center justify-between p-2 rounded-md bg-muted/20"
            >
              <div className="flex items-center gap-2">
                <div className={cn('w-2 h-2 rounded-full', tier.available ? 'bg-green-500' : 'bg-gray-400')} />
                <span className="text-xs font-medium">{tier.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="text-[9px] px-1 py-0">
                  {tier.isolation_level.replace(/_/g, ' ')}
                </Badge>
                <span className={cn('text-[10px]', tier.available ? 'text-green-500' : 'text-muted-foreground')}>
                  {tier.available ? 'Ready' : tier.status}
                </span>
              </div>
            </div>
          ))}
        </div>

        <div className="flex items-center gap-2 p-2 rounded-md bg-muted/20 text-xs text-muted-foreground">
          <Info className="w-3.5 h-3.5 flex-shrink-0" />
          <span>
            Kernel: {status.system_info.kernel} | Cgroups: {status.system_info.cgroup_version} |
            Seccomp: {status.system_info.seccomp_enabled ? 'Yes' : 'No'}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

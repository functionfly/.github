import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import {
  useEnableEncryption,
  useEncryptionStats,
  useMigrateEncryption,
  useRotateEncryptionKey,
} from '@/hooks/useState';
import { Loader2, Lock, RefreshCw, Shield } from 'lucide-react';

interface StateEncryptionPanelProps {
  /** When set, shows per-state enable action */
  statePath?: string;
  compact?: boolean;
}

export function StateEncryptionPanel({ statePath, compact = false }: StateEncryptionPanelProps) {
  const { data: stats, isLoading, error, refetch } = useEncryptionStats();
  const migrate = useMigrateEncryption(statePath ?? '');
  const rotateKey = useRotateEncryptionKey();
  const enableEncryption = useEnableEncryption(statePath ?? '');

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <Card className="p-6 text-center text-text-muted">
        <p>Encryption stats unavailable</p>
        <p className="text-sm mt-1">{(error as Error).message}</p>
      </Card>
    );
  }

  if (!stats) return null;

  const stateCoverage =
    stats.total_states > 0 ? (stats.encrypted_states / stats.total_states) * 100 : 0;
  const valueCoverage =
    stats.total_values > 0 ? (stats.encrypted_values / stats.total_values) * 100 : 0;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text-primary flex items-center gap-2">
            <Shield className="w-5 h-5 text-brand-400" />
            Encryption at Rest
          </h3>
          {!compact && (
            <p className="text-sm text-text-muted mt-1">
              Server-side AES-256 encryption for state values
            </p>
          )}
        </div>
        <Badge variant={stats.encryption_enabled ? 'default' : 'secondary'}>
          {stats.encryption_enabled ? 'Enabled' : 'Not configured'}
        </Badge>
      </div>

      {!stats.encryption_enabled ? (
        <Card className="border-dashed">
          <CardContent className="py-6 text-center text-text-muted text-sm">
            Server-side encryption is not configured on this deployment.
          </CardContent>
        </Card>
      ) : (
        <>
          <div className={`grid gap-4 ${compact ? 'grid-cols-1' : 'grid-cols-1 md:grid-cols-2'}`}>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-text-muted">
                  States encrypted
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {stats.encrypted_states} / {stats.total_states}
                </p>
                <Progress value={stateCoverage} className="h-2 mt-2" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-text-muted">
                  Values encrypted
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {stats.encrypted_values} / {stats.total_values}
                </p>
                <Progress value={valueCoverage} className="h-2 mt-2" />
              </CardContent>
            </Card>
          </div>

          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => migrate.mutate({ dry_run: false })}
              disabled={migrate.isPending || stats.unencrypted_values === 0}
            >
              {migrate.isPending ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Lock className="w-4 h-4 mr-2" />
              )}
              Encrypt unencrypted values
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => rotateKey.mutate()}
              disabled={rotateKey.isPending}
            >
              {rotateKey.isPending ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4 mr-2" />
              )}
              Rotate key
            </Button>
            <Button variant="ghost" size="sm" onClick={() => refetch()}>
              Refresh stats
            </Button>
          </div>

          {statePath && (
            <Button
              size="sm"
              onClick={() => enableEncryption.mutate()}
              disabled={enableEncryption.isPending}
            >
              {enableEncryption.isPending ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Lock className="w-4 h-4 mr-2" />
              )}
              Enable encryption for this state
            </Button>
          )}
        </>
      )}
    </div>
  );
}

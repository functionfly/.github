/**
 * Alerts management dialog.
 * Lists, creates, and deletes usage alerts (spend cap, usage threshold, forecast overrun).
 */
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Loader2, Plus, Trash2, AlertCircle, Bell, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';
import {
  listUsageAlerts,
  createUsageAlert,
  deleteUsageAlert,
} from '@/api/usageAnalytics';
import { formatCostUsd } from '@/api/usageAnalytics';
import { cn } from '@/lib/utils';

const ALERT_TYPE_LABELS: Record<string, string> = {
  spend_cap: 'Spend cap',
  usage_threshold: 'Usage threshold',
  forecast_overrun: 'Forecast overrun',
};

export function AlertsDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [newType, setNewType] = useState<'spend_cap' | 'usage_threshold' | 'forecast_overrun'>(
    'spend_cap'
  );
  const [newThresholdCents, setNewThresholdCents] = useState('');
  const [newThresholdCount, setNewThresholdCount] = useState('');
  const [newAction, setNewAction] = useState<'notify' | 'throttle' | 'block'>('notify');

  const { data: alerts, isLoading, error } = useQuery({
    queryKey: ['usage-alerts'],
    queryFn: listUsageAlerts,
    enabled: open,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteUsageAlert(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['usage-alerts'] });
      toast.success('Alert deleted');
    },
    onError: () => toast.error('Failed to delete alert'),
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createUsageAlert({
        alert_type: newType,
        threshold_value:
          newType === 'spend_cap' ? Number(newThresholdCents) : Number(newThresholdCount),
        action: newAction,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['usage-alerts'] });
      toast.success('Alert created');
      setCreateOpen(false);
      setNewThresholdCents('');
      setNewThresholdCount('');
    },
    onError: (e: unknown) => {
      const msg = e instanceof Error ? e.message : 'Failed to create alert';
      toast.error(msg);
    },
  });

  const list = alerts?.alerts ?? [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col gap-0 p-0">
        <DialogHeader className="px-6 pt-5 pb-4 border-b border-border-subtle">
          <DialogTitle className="flex items-center gap-3">
            <span className="v-icon-brand w-9 h-9">
              <Bell className="h-4 w-4" />
            </span>
            <div>
              <span className="text-base font-semibold">Usage alerts</span>
              <DialogDescription className="mt-0.5">
                Get notified when spend or usage approaches limits. Trigger by webhook,
                email, or throttling.
              </DialogDescription>
            </div>
          </DialogTitle>
        </DialogHeader>

        <div className="flex justify-end px-6 pt-4">
          <Button size="sm" onClick={() => setCreateOpen((v) => !v)}>
            <Plus className="h-4 w-4 mr-1" />
            {createOpen ? 'Cancel' : 'New alert'}
          </Button>
        </div>

        {createOpen && (
          <div className="mx-6 mt-3 border border-border-subtle rounded-lg p-4 space-y-3 bg-bg-secondary/40">
            <div className="space-y-2">
              <Label htmlFor="alert-type" className="text-xs">
                Alert type
              </Label>
              <select
                id="alert-type"
                value={newType}
                onChange={(e) =>
                  setNewType(e.target.value as typeof newType)
                }
                className="w-full h-9 rounded-md border border-border-subtle bg-bg-primary px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ff-flame/30 focus:border-ff-flame/50"
              >
                <option value="spend_cap">Spend cap (cents)</option>
                <option value="usage_threshold">Usage threshold (request count)</option>
                <option value="forecast_overrun">Forecast overrun (cents)</option>
              </select>
            </div>
            {newType === 'spend_cap' || newType === 'forecast_overrun' ? (
              <div className="space-y-2">
                <Label htmlFor="threshold-cents" className="text-xs">
                  Threshold (cents)
                </Label>
                <Input
                  id="threshold-cents"
                  type="number"
                  min="0"
                  placeholder="e.g. 5000 (=$50.00)"
                  value={newThresholdCents}
                  onChange={(e) => setNewThresholdCents(e.target.value)}
                />
                {newThresholdCents && (
                  <p className="text-xs text-text-muted">
                    = {formatCostUsd(Number(newThresholdCents))}
                  </p>
                )}
              </div>
            ) : (
              <div className="space-y-2">
                <Label htmlFor="threshold-count" className="text-xs">
                  Request count
                </Label>
                <Input
                  id="threshold-count"
                  type="number"
                  min="0"
                  placeholder="e.g. 10000"
                  value={newThresholdCount}
                  onChange={(e) => setNewThresholdCount(e.target.value)}
                />
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="alert-action" className="text-xs">
                Action
              </Label>
              <select
                id="alert-action"
                value={newAction}
                onChange={(e) => setNewAction(e.target.value as typeof newAction)}
                className="w-full h-9 rounded-md border border-border-subtle bg-bg-primary px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ff-flame/30 focus:border-ff-flame/50"
              >
                <option value="notify">Notify only</option>
                <option value="throttle">Throttle traffic</option>
                <option value="block">Block traffic</option>
              </select>
            </div>
            <Button
              size="sm"
              className="w-full"
              disabled={
                createMutation.isPending ||
                (newType === 'spend_cap' && !newThresholdCents) ||
                (newType === 'usage_threshold' && !newThresholdCount) ||
                (newType === 'forecast_overrun' && !newThresholdCents)
              }
              onClick={() => createMutation.mutate()}
            >
              {createMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin mr-1" />
              ) : (
                <Plus className="h-4 w-4 mr-1" />
              )}
              Create alert
            </Button>
          </div>
        )}

        <div className="flex-1 overflow-auto space-y-2 px-6 py-3">
          {isLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))
          ) : error ? (
            <div className="text-center py-8 text-text-muted">
              <AlertCircle className="h-5 w-5 mx-auto mb-2 text-destructive" />
              Failed to load alerts
            </div>
          ) : list.length === 0 ? (
            <div className="text-center py-10">
              <div className="v-icon-muted w-12 h-12 mx-auto mb-3">
                <AlertTriangle className="h-5 w-5" />
              </div>
              <p className="text-sm font-medium text-text-secondary">No alerts configured</p>
              <p className="text-xs text-text-muted mt-1">
                Create one to be notified when spend or usage crosses a threshold.
              </p>
            </div>
          ) : (
            list.map((a) => (
              <div
                key={a.id}
                className="flex items-center justify-between p-3 rounded-lg border border-border-subtle bg-bg-secondary/40 hover:bg-bg-secondary transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div
                    className={cn(
                      'w-2 h-10 rounded-full',
                      a.status === 'active'
                        ? 'bg-emerald-500'
                        : a.status === 'triggered'
                          ? 'bg-red-500'
                          : 'bg-text-muted'
                    )}
                  />
                  <div className="min-w-0">
                    <p className="text-sm font-medium flex items-center gap-2">
                      {ALERT_TYPE_LABELS[a.alert_type] ?? a.alert_type}
                      <Badge
                        variant="secondary"
                        className={cn(
                          'text-[10px] h-4 px-1.5',
                          a.status === 'active'
                            ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/30'
                            : a.status === 'triggered'
                              ? 'bg-red-500/10 text-red-500 border-red-500/30'
                              : ''
                        )}
                      >
                        {a.status}
                      </Badge>
                    </p>
                    <p className="text-xs text-text-muted mt-0.5">
                      Threshold:{' '}
                      {a.alert_type === 'usage_threshold'
                        ? a.threshold_value.toLocaleString() + ' requests'
                        : formatCostUsd(a.threshold_value)}
                    </p>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => deleteMutation.mutate(a.id)}
                  disabled={deleteMutation.isPending}
                  title="Delete alert"
                  className="shrink-0"
                >
                  <Trash2 className="h-4 w-4 text-destructive" />
                </Button>
              </div>
            ))
          )}
        </div>

        <DialogFooter className="px-6 py-4 border-t border-border-subtle">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

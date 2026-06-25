import { CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { AlertCircle, Box, Cpu, HardDrive, Loader2, Server, Clock, CheckCircle2, XCircle } from 'lucide-react';
import type { MicroVMExecution, MicroVMQuota, MicroVMStats } from '@/api/microvm';

interface MicroVMMetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: React.ReactNode;
  tone?: 'neutral' | 'indigo' | 'amber' | 'emerald' | 'cyan' | 'violet' | 'rose';
  className?: string;
}

export function MicroVMMetricCard({
  title,
  value,
  subtitle,
  icon,
  tone = 'indigo',
  className,
}: MicroVMMetricCardProps) {
  const toneStyles = {
    neutral: { border: 'border-border', bg: 'bg-muted/50', icon: 'text-muted-foreground' },
    indigo: { border: 'border-blue-500/30', bg: 'bg-blue-500/10', icon: 'text-blue-500' },
    amber: { border: 'border-amber-500/30', bg: 'bg-amber-500/10', icon: 'text-amber-500' },
    emerald: { border: 'border-emerald-500/30', bg: 'bg-emerald-500/10', icon: 'text-emerald-500' },
    cyan: { border: 'border-cyan-500/30', bg: 'bg-cyan-500/10', icon: 'text-cyan-500' },
    violet: { border: 'border-violet-500/30', bg: 'bg-violet-500/10', icon: 'text-violet-500' },
    rose: { border: 'border-rose-500/30', bg: 'bg-rose-500/10', icon: 'text-rose-500' },
  };

  const style = toneStyles[tone];

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl border bg-card p-5 transition-colors',
        style.border,
        className
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <div className="mt-1.5 flex items-baseline gap-2">
            <span className="text-2xl font-semibold tabular-nums tracking-tight text-foreground">
              {value}
            </span>
          </div>
          {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
        </div>
        {icon && (
          <div
            className={cn(
              'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border',
              style.bg,
              style.border,
              style.icon
            )}
          >
            {icon}
          </div>
        )}
      </div>
    </div>
  );
}

interface MicroVMUsagePanelProps {
  stats: MicroVMStats | null;
  quota: MicroVMQuota | null;
  executions: MicroVMExecution[];
  maxConcurrentVMs: number;
  className?: string;
}

export function MicroVMUsagePanel({
  stats,
  quota,
  executions,
  maxConcurrentVMs,
  className,
}: MicroVMUsagePanelProps) {
  const activeVMs = stats?.running_vms ?? 0;
  const totalExecutions = stats?.total_executions ?? 0;
  const avgDuration = stats?.avg_duration_ms ?? 0;
  const successRate = stats?.success_rate ?? 0;

  const activePercent = maxConcurrentVMs > 0 ? (activeVMs / maxConcurrentVMs) * 100 : 0;
  const isWarning = activePercent >= 70 && activePercent < 90;
  const isCritical = activePercent >= 90;

  return (
    <div className={cn('space-y-4', className)}>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MicroVMMetricCard
          title="Active VMs"
          value={activeVMs}
          subtitle={`of ${maxConcurrentVMs} max`}
          icon={<Server className="h-5 w-5" />}
          tone={isCritical ? 'rose' : isWarning ? 'amber' : 'emerald'}
        />
        <MicroVMMetricCard
          title="Total Executions"
          value={totalExecutions.toLocaleString()}
          subtitle="this month"
          icon={<Box className="h-5 w-5" />}
          tone="indigo"
        />
        <MicroVMMetricCard
          title="Avg Duration"
          value={`${(avgDuration / 1000).toFixed(1)}s`}
          subtitle="per execution"
          icon={<Clock className="h-5 w-5" />}
          tone="cyan"
        />
        <MicroVMMetricCard
          title="Success Rate"
          value={`${successRate.toFixed(1)}%`}
          subtitle="completed successfully"
          icon={successRate >= 95 ? <CheckCircle2 className="h-5 w-5" /> : <XCircle className="h-5 w-5" />}
          tone={successRate >= 95 ? 'emerald' : 'rose'}
        />
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">VM Concurrency</span>
          </div>
          <div className="flex items-center gap-2">
            {isCritical && <AlertCircle className="h-4 w-4 text-rose-500" />}
            <span className="tabular-nums text-muted-foreground">
              {activeVMs} / {maxConcurrentVMs}
            </span>
          </div>
        </div>
        <div className="relative">
          <Progress value={activePercent} className="h-2 bg-muted" />
          <motion.div
            className={cn(
              'absolute top-0 left-0 h-2 rounded-full transition-all',
              isCritical ? 'bg-rose-500' : isWarning ? 'bg-amber-500' : 'bg-emerald-500'
            )}
            initial={{ width: 0 }}
            animate={{ width: `${activePercent}%` }}
            transition={{ duration: 0.5 }}
          />
        </div>
      </div>

      {executions.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium">Recent Executions</h4>
          <div className="max-h-48 space-y-1 overflow-auto rounded-lg border bg-muted/30 p-2">
            {executions.slice(0, 5).map((exec) => (
              <div
                key={exec.id}
                className="flex items-center justify-between rounded-md bg-card px-2 py-1.5 text-xs"
              >
                <div className="flex items-center gap-2">
                  {exec.status === 'running' ? (
                    <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
                  ) : exec.status === 'completed' ? (
                    <CheckCircle2 className="h-3 w-3 text-emerald-500" />
                  ) : (
                    <XCircle className="h-3 w-3 text-rose-500" />
                  )}
                  <span className="font-mono text-muted-foreground">{exec.function_id.slice(0, 8)}</span>
                </div>
                <div className="flex items-center gap-3 text-muted-foreground">
                  <span>{exec.duration_ms}ms</span>
                  <span>{exec.memory_mb}MB</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface MicroVMBillingPanelProps {
  currentBilling: {
    total_charge_cents: number;
    total_executions: number;
    total_compute_seconds: number;
    total_memory_seconds: number;
  } | null;
  limits: {
    max_memory_mb: number;
    max_vcpu: number;
    max_timeout: number;
  };
  className?: string;
}

export function MicroVMBillingPanel({
  currentBilling,
  limits,
  className,
}: MicroVMBillingPanelProps) {
  const formatCents = (cents: number) => `$${(cents / 100).toFixed(2)}`;

  return (
    <div className={cn('space-y-4 rounded-xl border bg-card p-5', className)}>
      <h3 className="text-lg font-semibold">MicroVM Billing</h3>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">Current Period</p>
          <p className="text-2xl font-semibold">
            {currentBilling ? formatCents(currentBilling.total_charge_cents) : '$0.00'}
          </p>
        </div>

        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">Executions</p>
          <p className="text-2xl font-semibold">
            {currentBilling?.total_executions ?? 0}
          </p>
        </div>
      </div>

      <div className="space-y-2 rounded-lg bg-muted/50 p-3">
        <h4 className="text-sm font-medium">Resource Limits</h4>
        <div className="grid gap-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Max Memory</span>
            <span>{limits.max_memory_mb}MB</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Max vCPUs</span>
            <span>{limits.max_vcpu}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Max Timeout</span>
            <span>{(limits.max_timeout / 1000).toFixed(0)}s</span>
          </div>
        </div>
      </div>
    </div>
  );
}

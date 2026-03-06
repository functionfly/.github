/**
 * RuntimeMetricsDisplay - Time, memory, cost display
 */

import { cn } from '@/lib/utils';
import {
  Clock,
  HardDrive,
  Cpu,
  DollarSign,
  Zap,
  Gauge,
} from 'lucide-react';
import type { ResourceUsage, PerformanceMetrics } from '../types';

interface RuntimeMetricsDisplayProps {
  resourceUsage?: ResourceUsage;
  performanceMetrics?: PerformanceMetrics;
  className?: string;
  compact?: boolean;
}

function formatDuration(ms: number): string {
  if (ms < 1) return `${(ms * 1000).toFixed(2)}µs`;
  if (ms < 1000) return `${ms.toFixed(2)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

function formatMemory(mb: number): string {
  if (mb < 1) return `${(mb * 1024).toFixed(1)}KB`;
  if (mb < 1024) return `${mb.toFixed(1)}MB`;
  return `${(mb / 1024).toFixed(2)}GB`;
}

function formatCost(cost: number): string {
  if (cost < 0.01) return `< $0.01`;
  return `$${cost.toFixed(2)}`;
}

export function RuntimeMetricsDisplay({
  resourceUsage,
  performanceMetrics,
  className,
  compact = false,
}: RuntimeMetricsDisplayProps) {
  if (!resourceUsage && !performanceMetrics) {
    return null;
  }

  if (compact) {
    return (
      <div className={cn('flex items-center gap-3 text-sm', className)}>
        {resourceUsage && (
          <>
            <span className="flex items-center gap-1 text-slate-400">
              <Clock className="h-3.5 w-3.5" />
              {formatDuration(resourceUsage.runtimeMs)}
            </span>
            <span className="flex items-center gap-1 text-slate-400">
              <HardDrive className="h-3.5 w-3.5" />
              {formatMemory(resourceUsage.memoryMb)}
            </span>
          </>
        )}
        {resourceUsage?.cost !== undefined && (
          <span className="flex items-center gap-1 text-slate-400">
            <DollarSign className="h-3.5 w-3.5" />
            {formatCost(resourceUsage.cost)}
          </span>
        )}
      </div>
    );
  }

  return (
    <div className={cn('rounded-lg border border-slate-800 bg-slate-900/50 p-4', className)}>
      <h4 className="mb-3 flex items-center gap-2 text-sm font-semibold text-white">
        <Gauge className="h-4 w-4 text-indigo-400" />
        Performance Metrics
      </h4>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {resourceUsage && (
          <>
            <MetricCard
              icon={Clock}
              label="Execution Time"
              value={formatDuration(resourceUsage.runtimeMs)}
              color="text-blue-400"
            />
            <MetricCard
              icon={HardDrive}
              label="Memory Usage"
              value={formatMemory(resourceUsage.memoryMb)}
              color="text-violet-400"
            />
            <MetricCard
              icon={Cpu}
              label="CPU Time"
              value={`${resourceUsage.cpuSeconds.toFixed(3)}s`}
              color="text-amber-400"
            />
            {resourceUsage.cost !== undefined && (
              <MetricCard
                icon={DollarSign}
                label="Execution Cost"
                value={formatCost(resourceUsage.cost)}
                color="text-emerald-400"
              />
            )}
          </>
        )}

        {performanceMetrics && (
          <>
            <MetricCard
              icon={Clock}
              label="Avg Execution Time"
              value={formatDuration(performanceMetrics.avgExecutionTimeMs)}
              color="text-blue-400"
            />
            <MetricCard
              icon={HardDrive}
              label="Avg Memory Usage"
              value={formatMemory(performanceMetrics.avgMemoryUsageMb)}
              color="text-violet-400"
            />
            {performanceMetrics.percentile95Ms && (
              <MetricCard
                icon={Zap}
                label="95th Percentile"
                value={formatDuration(performanceMetrics.percentile95Ms)}
                color="text-amber-400"
              />
            )}
          </>
        )}
      </div>

      {performanceMetrics?.deterministic && (
        <div className="mt-3 flex items-center gap-2 text-xs text-emerald-400">
          <div className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
          Deterministic execution verified
        </div>
      )}
    </div>
  );
}

function MetricCard({
  icon: Icon,
  label,
  value,
  color,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  color: string;
}) {
  return (
    <div className="flex items-center gap-3 rounded-md bg-slate-950/50 p-3">
      <div className={cn('rounded-md bg-slate-900 p-2', color)}>
        <Icon className="h-4 w-4" />
      </div>
      <div>
        <p className="text-xs text-slate-400">{label}</p>
        <p className={cn('font-mono text-sm font-medium', color)}>{value}</p>
      </div>
    </div>
  );
}

/**
 * Simple metrics row for inline display
 */
export function MetricsRow({
  resourceUsage,
  className,
}: {
  resourceUsage?: ResourceUsage;
  className?: string;
}) {
  if (!resourceUsage) return null;

  return (
    <div className={cn('flex flex-wrap items-center gap-4 text-xs text-slate-400', className)}>
      <span className="flex items-center gap-1">
        <Clock className="h-3 w-3" />
        {formatDuration(resourceUsage.runtimeMs)}
      </span>
      <span className="flex items-center gap-1">
        <HardDrive className="h-3 w-3" />
        {formatMemory(resourceUsage.memoryMb)}
      </span>
      {resourceUsage.cost !== undefined && (
        <span className="flex items-center gap-1">
          <DollarSign className="h-3 w-3" />
          {formatCost(resourceUsage.cost)}
        </span>
      )}
    </div>
  );
}

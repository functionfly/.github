import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle2, Circle, DollarSign, Sparkles } from 'lucide-react';
import { RUNTIME_META } from './constants';
import type { FunctionEditorModel } from './useFunctionEditor';
import { formatTimeout } from './utils';

type Props = { editor: FunctionEditorModel };

interface ChecklistItem {
  label: string;
  done: boolean;
}

function estimateCost(memoryMb: number, timeoutMs: number, warmInstances: number): string {
  // Rough estimate: $0.0000002 per GB-second
  const gbSeconds = (memoryMb / 1024) * (timeoutMs / 1000);
  const perInvocation = gbSeconds * 0.0000002;
  const warmCost = warmInstances * 0.000005; // ~$0.000005/hr per warm instance
  const monthly = perInvocation * 1_000_000 + warmCost * 720; // 1M invocations/month
  if (monthly < 0.01) return '< $0.01 / 1M invocations';
  return `~$${monthly.toFixed(2)} / 1M invocations`;
}

export function ConfigSummary({ editor }: Props) {
  const {
    functionName,
    slug,
    runtime,
    runtimeVersion,
    resources,
    visibility,
    httpTrigger,
    scheduleTrigger,
    envVars,
    tags,
    retryPolicy,
    warmInstances,
    code,
  } = editor;

  const checklist: ChecklistItem[] = [
    { label: 'Function name', done: !!functionName.trim() },
    { label: 'Slug / identifier', done: !!slug.trim() },
    { label: 'Code written', done: code.trim().length > 50 },
    { label: 'Runtime selected', done: true },
    { label: 'HTTP or schedule trigger', done: httpTrigger.enabled || scheduleTrigger.enabled },
    { label: 'Visibility set', done: true },
  ];

  const completedCount = checklist.filter((c) => c.done).length;
  const completionPct = Math.round((completedCount / checklist.length) * 100);

  const summaryRows = [
    { label: 'Name', value: functionName || '—' },
    { label: 'Slug', value: slug || '—', mono: true },
    { label: 'Runtime', value: `${RUNTIME_META[runtime].label} ${runtimeVersion}` },
    { label: 'Memory', value: `${resources.memoryMb} MB` },
    { label: 'Timeout', value: formatTimeout(resources.timeoutMs) },
    { label: 'Concurrency', value: `${resources.maxConcurrency}` },
    {
      label: 'Visibility',
      value: visibility === 'public' ? '🌐 Public' : '🔒 Private',
    },
    {
      label: 'HTTP Trigger',
      value: httpTrigger.enabled ? `${httpTrigger.method} ${httpTrigger.path}` : 'Disabled',
    },
    {
      label: 'Schedule',
      value: scheduleTrigger.enabled ? scheduleTrigger.cron : 'Disabled',
    },
    {
      label: 'Env Vars',
      value: `${envVars.length} variable${envVars.length !== 1 ? 's' : ''}`,
    },
    { label: 'Tags', value: tags.length > 0 ? tags.join(', ') : 'None' },
    {
      label: 'Retries',
      value: `${retryPolicy.maxRetries} (${retryPolicy.backoffStrategy})`,
    },
    {
      label: 'Warm Instances',
      value: warmInstances > 0 ? `${warmInstances}` : 'None',
    },
  ];

  return (
    <div className="space-y-4">
      {/* Configuration Summary */}
      <Card
        className="card border-border-subtle/50"
        style={{
          background: 'var(--bg-secondary, #12121a)',
        }}
      >
        <CardHeader className="pb-3 pt-4 px-5">
          <div className="flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-indigo-400" />
            <CardTitle className="text-sm font-semibold text-text-primary">
              Configuration Summary
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-5 space-y-2.5">
          {summaryRows.map(({ label, value, mono }) => (
            <div key={label} className="flex items-start justify-between gap-3 text-xs">
              <span className="text-text-muted shrink-0">{label}</span>
              <span className={`text-text-secondary text-right ${mono ? 'font-mono' : ''}`}>
                {value}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Estimated Cost */}
      <Card
        className="card border-border-subtle/50"
        style={{
          background: 'var(--bg-secondary, #12121a)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center gap-2">
            <DollarSign className="w-4 h-4 text-emerald-400" />
            <CardTitle className="text-sm font-semibold text-text-primary">
              Estimated Cost
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-4">
          <p className="text-sm font-mono text-emerald-400">
            {estimateCost(resources.memoryMb, resources.timeoutMs, warmInstances)}
          </p>
          <p className="text-xs text-text-muted mt-1.5 leading-relaxed">
            Based on current resource limits. Actual cost depends on invocation count.
          </p>
        </CardContent>
      </Card>

      {/* Deployment Checklist */}
      <Card
        className="card border-border-subtle/50"
        style={{
          background: 'var(--bg-secondary, #12121a)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-indigo-400" />
              <CardTitle className="text-sm font-semibold text-text-primary">
                Deployment Checklist
              </CardTitle>
            </div>
            <span className="text-xs font-mono text-text-muted">
              {completedCount}/{checklist.length}
            </span>
          </div>
          {/* Progress bar */}
          <div className="mt-2.5 h-1.5 rounded-full bg-bg-tertiary overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{
                width: `${completionPct}%`,
                background:
                  completionPct === 100
                    ? 'linear-gradient(90deg, #10b981, #34d399)'
                    : 'linear-gradient(90deg, #6366f1, #8b5cf6)',
              }}
            />
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-4 space-y-2">
          {checklist.map((item) => (
            <div key={item.label} className="flex items-center gap-2.5 text-xs">
              {item.done ? (
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
              ) : (
                <Circle className="w-3.5 h-3.5 text-text-muted shrink-0" />
              )}
              <span className={item.done ? 'text-text-secondary' : 'text-text-muted'}>
                {item.label}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

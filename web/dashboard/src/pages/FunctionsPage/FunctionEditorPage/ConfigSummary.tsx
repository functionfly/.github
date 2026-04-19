import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Slider } from '@/components/ui/slider';
import { useState } from 'react';
import { CheckCircle2, Circle, DollarSign, Info, Sparkles } from 'lucide-react';
import { RUNTIME_META } from './constants';
import type { FunctionEditorModel } from './useFunctionEditor';
import { formatTimeout } from './utils';

type Props = { editor: FunctionEditorModel };

interface ChecklistItem {
  label: string;
  done: boolean;
}

interface CostBreakdown {
  computeCost: number;
  warmInstanceCost: number;
  requestCost: number;
  total: number;
}

function calculateCost(
  memoryMb: number,
  timeoutMs: number,
  warmInstances: number,
  monthlyInvocations: number
): CostBreakdown {
  // Pricing constants (AWS Lambda-like pricing)
  const REQUEST_COST_PER_MILLION = 0.20; // $0.20 per million requests
  const COMPUTE_COST_PER_GB_SECOND = 0.0000166667; // $0.0000166667 per GB-second

  // Calculate compute cost
  const avgRuntimeSeconds = timeoutMs / 1000 / 2; // Assume average is half of max timeout
  const gbSeconds = (memoryMb / 1024) * avgRuntimeSeconds * monthlyInvocations;
  const computeCost = gbSeconds * COMPUTE_COST_PER_GB_SECOND;

  // Request cost
  const requestCost = (monthlyInvocations / 1_000_000) * REQUEST_COST_PER_MILLION;

  // Warm instance cost (rough estimate: $0.0075/hr per 128MB instance)
  const warmInstanceHourly = (memoryMb / 128) * 0.0075;
  const warmInstanceCost = warmInstances * warmInstanceHourly * 730; // ~730 hours per month

  return {
    computeCost,
    warmInstanceCost,
    requestCost,
    total: computeCost + warmInstanceCost + requestCost,
  };
}

function formatCurrency(amount: number): string {
  if (amount < 0.01) return '< $0.01';
  if (amount < 1) return `~$${amount.toFixed(3)}`;
  if (amount < 10) return `~$${amount.toFixed(2)}`;
  return `~$${amount.toFixed(1)}`;
}

const INVOCATION_PRESETS = [
  { value: 1000, label: '1K', description: 'Low traffic' },
  { value: 100_000, label: '100K', description: 'Medium traffic' },
  { value: 1_000_000, label: '1M', description: 'High traffic' },
  { value: 10_000_000, label: '10M', description: 'Very high traffic' },
  { value: 100_000_000, label: '100M', description: 'Enterprise' },
];

function scaleInvocations(value: number): number {
  // Map slider 0-100 to 1K - 100M with logarithmic scale
  if (value <= 20) return Math.round(1000 + (value / 20) * (100_000 - 1000));
  if (value <= 40) return Math.round(100_000 + ((value - 20) / 20) * (1_000_000 - 100_000));
  if (value <= 60) return Math.round(1_000_000 + ((value - 40) / 20) * (10_000_000 - 1_000_000));
  if (value <= 80) return Math.round(10_000_000 + ((value - 60) / 20) * (50_000_000 - 10_000_000));
  return Math.round(50_000_000 + ((value - 80) / 20) * (100_000_000 - 50_000_000));
}

function unscaleInvocations(invocations: number): number {
  // Reverse the scaling
  if (invocations <= 100_000) return ((invocations - 1000) / (100_000 - 1000)) * 20;
  if (invocations <= 1_000_000) return 20 + ((invocations - 100_000) / (1_000_000 - 100_000)) * 20;
  if (invocations <= 10_000_000) return 40 + ((invocations - 1_000_000) / (10_000_000 - 1_000_000)) * 20;
  if (invocations <= 50_000_000) return 60 + ((invocations - 10_000_000) / (50_000_000 - 10_000_000)) * 20;
  return 80 + ((invocations - 50_000_000) / (100_000_000 - 50_000_000)) * 20;
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

  const [monthlyInvocations, setMonthlyInvocations] = useState(100_000);
  const [showDetails, setShowDetails] = useState(false);

  const cost = calculateCost(
    resources.memoryMb,
    resources.timeoutMs,
    warmInstances,
    monthlyInvocations
  );

  const sliderValue = unscaleInvocations(monthlyInvocations);

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
          background: 'var(--bg-secondary)',
        }}
      >
        <CardHeader className="pb-3 pt-4 px-5">
          <div className="flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-[#FF6B35]" />
            <CardTitle className="text-sm font-semibold text-text-primary font-display">
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

      {/* Enhanced Cost Estimator */}
      <Card
        className="card border-border-subtle/50"
        style={{
          background: 'var(--bg-secondary)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <DollarSign className="w-4 h-4 text-emerald-400" />
              <CardTitle className="text-sm font-semibold text-text-primary font-display">
                Estimated Cost
              </CardTitle>
            </div>
            <span className="text-xs text-text-muted">
              {monthlyInvocations.toLocaleString()} invocations/month
            </span>
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-4">
          {/* Invocation slider */}
          <div className="mb-4">
            <Slider
              value={[sliderValue]}
              onValueChange={([v]) => setMonthlyInvocations(scaleInvocations(v))}
              min={0}
              max={100}
              step={1}
              className="w-full"
            />
            <div className="flex justify-between mt-2">
              {INVOCATION_PRESETS.map((preset) => (
                <button
                  key={preset.value}
                  onClick={() => setMonthlyInvocations(preset.value)}
                  className={`text-[10px] px-1.5 py-0.5 rounded transition-colors ${
                    monthlyInvocations === preset.value
                      ? 'bg-emerald-500/20 text-emerald-400'
                      : 'text-text-muted hover:text-text-secondary'
                  }`}
                  title={preset.description}
                >
                  {preset.label}
                </button>
              ))}
            </div>
          </div>

          {/* Total cost */}
          <div className="flex items-baseline gap-2">
            <p className="text-2xl font-semibold font-mono text-emerald-400">
              {formatCurrency(cost.total)}
            </p>
            <span className="text-xs text-text-muted">/month</span>
          </div>

          {/* Cost breakdown toggle */}
          <button
            onClick={() => setShowDetails(!showDetails)}
            className="flex items-center gap-1 text-xs text-text-muted hover:text-text-secondary transition-colors mt-2"
          >
            <Info className="w-3 h-3" />
            {showDetails ? 'Hide breakdown' : 'View breakdown'}
          </button>

          {/* Cost breakdown */}
          {showDetails && (
            <div className="mt-3 p-3 rounded-lg bg-bg-tertiary/50 space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-text-muted">Compute (GB-seconds)</span>
                <span className="text-text-secondary font-mono">
                  {formatCurrency(cost.computeCost)}
                </span>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-text-muted">Requests</span>
                <span className="text-text-secondary font-mono">
                  {formatCurrency(cost.requestCost)}
                </span>
              </div>
              {warmInstances > 0 && (
                <div className="flex justify-between text-xs">
                  <span className="text-text-muted">Warm instances ({warmInstances})</span>
                  <span className="text-text-secondary font-mono">
                    {formatCurrency(cost.warmInstanceCost)}
                  </span>
                </div>
              )}
              <div className="pt-2 border-t border-border-subtle/30 flex justify-between text-xs font-medium">
                <span className="text-text-primary">Total</span>
                <span className="text-emerald-400 font-mono">{formatCurrency(cost.total)}</span>
              </div>
            </div>
          )}

          {/* Free tier notice */}
          {cost.total < 5 && (
            <p className="text-xs text-emerald-500/80 mt-3 flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              Likely covered by free tier on most providers
            </p>
          )}

          {cost.total > 100 && (
            <p className="text-xs text-amber-400 mt-3 flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
              Consider optimizing: reduce memory, timeout, or warm instances
            </p>
          )}
        </CardContent>
      </Card>

      {/* Deployment Checklist */}
      <Card
        className="card border-border-subtle/50"
        style={{
          background: 'var(--bg-secondary)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-[#FF6B35]" />
              <CardTitle className="text-sm font-semibold text-text-primary font-display">
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
                      : 'linear-gradient(90deg, #FF6B35, #FF8C42)',
                }}
            />
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-4 space-y-2">
          {checklist.map((item) => (
            <div key={item.label} className="flex items-center gap-2.5 text-xs">
              {item.done ? (
                <CheckCircle2 className="w-3.5 h-3.5 text-[#FF6B35] shrink-0" />
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

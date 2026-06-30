import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Slider } from '@/components/ui/slider';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
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
  { value: 1000, label: '1K', descriptionKey: 'funcEditor.lowTraffic' as const },
  { value: 100_000, label: '100K', descriptionKey: 'funcEditor.mediumTraffic' as const },
  { value: 1_000_000, label: '1M', descriptionKey: 'funcEditor.highTraffic' as const },
  { value: 10_000_000, label: '10M', descriptionKey: 'funcEditor.veryHighTraffic' as const },
  { value: 100_000_000, label: '100M', descriptionKey: 'funcEditor.enterprise' as const },
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
  const { t } = useTranslation();
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
    { label: t('funcEditor.checklistFunctionName'), done: !!functionName.trim() },
    { label: t('funcEditor.checklistSlug'), done: !!slug.trim() },
    { label: t('funcEditor.checklistCodeWritten'), done: code.trim().length > 50 },
    { label: t('funcEditor.checklistRuntimeSelected'), done: true },
    { label: t('funcEditor.checklistTrigger'), done: httpTrigger.enabled || scheduleTrigger.enabled },
    { label: t('funcEditor.checklistVisibility'), done: true },
  ];

  const completedCount = checklist.filter((c) => c.done).length;
  const completionPct = Math.round((completedCount / checklist.length) * 100);

  const summaryRows = [
    { label: t('funcEditor.summaryName'), value: functionName || '—' },
    { label: t('funcEditor.summarySlug'), value: slug || '—', mono: true },
    { label: t('funcEditor.summaryRuntime'), value: `${RUNTIME_META[runtime].label} ${runtimeVersion}` },
    { label: t('funcEditor.summaryMemory'), value: `${resources.memoryMb} MB` },
    { label: t('funcEditor.summaryTimeout'), value: formatTimeout(resources.timeoutMs) },
    { label: t('funcEditor.summaryConcurrency'), value: `${resources.maxConcurrency}` },
    {
      label: t('funcEditor.summaryVisibility'),
      value: visibility === 'public' ? t('funcEditor.publicEmoji') : t('funcEditor.privateEmoji'),
    },
    {
      label: t('funcEditor.summaryHttpTrigger'),
      value: httpTrigger.enabled ? `${httpTrigger.method} ${httpTrigger.path}` : t('funcEditor.disabled'),
    },
    {
      label: t('funcEditor.summarySchedule'),
      value: scheduleTrigger.enabled ? scheduleTrigger.cron : t('funcEditor.disabled'),
    },
    {
      label: t('funcEditor.summaryEnvVars'),
      value: `${envVars.length} variable${envVars.length !== 1 ? 's' : ''}`,
    },
    { label: t('funcEditor.summaryTags'), value: tags.length > 0 ? tags.join(', ') : t('funcEditor.none') },
    {
      label: t('funcEditor.summaryRetries'),
      value: `${retryPolicy.maxRetries} (${retryPolicy.backoffStrategy})`,
    },
    {
      label: t('funcEditor.summaryWarmInstances'),
      value: warmInstances > 0 ? `${warmInstances}` : t('funcEditor.none'),
    },
  ];

  return (
    <div className="space-y-4">
      {/* Configuration Summary */}
      <Card
        className="overflow-hidden"
        style={{
          background: 'var(--panel)',
          backgroundImage: 'radial-gradient(140% 100% at 15% 0%, var(--glass-tint), transparent 55%)',
          borderColor: 'var(--panel-edge)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <CardHeader className="pb-3 pt-4 px-5">
          <div className="flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-[var(--status-ok)]" />
            <CardTitle className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>
              {t('funcEditor.configSummary')}
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-5 space-y-2.5">
          {summaryRows.map(({ label, value, mono }) => (
            <div key={label} className="flex items-start justify-between gap-3 text-xs">
              <span className="text-[var(--text-faint)] shrink-0">{label}</span>
              <span className={`text-[var(--text-dim)] text-right ${mono ? 'font-[var(--font-mono)]' : ''}`}>
                {value}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Enhanced Cost Estimator */}
      <Card
        className="overflow-hidden"
        style={{
          background: 'var(--panel)',
          borderColor: 'var(--panel-edge)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <DollarSign className="w-4 h-4 text-[var(--status-ok)]" />
              <CardTitle className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>
                {t('funcEditor.estimatedCost')}
              </CardTitle>
            </div>
            <span className="text-xs text-[var(--text-faint)]">
              {t('funcEditor.invocationsPerMonth', { count: monthlyInvocations.toLocaleString() })}
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
                  className={`text-[10px] px-1.5 py-0.5 rounded-[var(--radius-sm)] transition-colors ${
                    monthlyInvocations === preset.value
                      ? 'bg-[rgba(143,255,208,0.1)] text-[var(--status-ok)]'
                      : 'text-[var(--text-faint)] hover:text-[var(--text-dim)]'
                  }`}
                  title={t(preset.descriptionKey)}
                >
                  {preset.label}
                </button>
              ))}
            </div>
          </div>

          {/* Total cost */}
          <div className="flex items-baseline gap-2">
            <p className="text-2xl font-semibold font-[var(--font-mono)] text-[var(--status-ok)]">
              {formatCurrency(cost.total)}
            </p>
            <span className="text-xs text-[var(--text-faint)]">{t('funcEditor.perMonth')}</span>
          </div>

          {/* Cost breakdown toggle */}
          <button
            onClick={() => setShowDetails(!showDetails)}
            className="flex items-center gap-1 text-xs text-[var(--text-faint)] hover:text-[var(--text-dim)] transition-colors mt-2"
          >
            <Info className="w-3 h-3" />
            {showDetails ? t('funcEditor.hideBreakdown') : t('funcEditor.viewBreakdown')}
          </button>

          {/* Cost breakdown */}
          {showDetails && (
            <div className="mt-3 p-3 rounded-[var(--radius)] bg-[var(--panel-raised)] space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-[var(--text-faint)]">{t('funcEditor.computeGBSeconds')}</span>
                <span className="text-[var(--text-dim)] font-[var(--font-mono)]">
                  {formatCurrency(cost.computeCost)}
                </span>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-[var(--text-faint)]">{t('funcEditor.requests')}</span>
                <span className="text-[var(--text-dim)] font-[var(--font-mono)]">
                  {formatCurrency(cost.requestCost)}
                </span>
              </div>
              {warmInstances > 0 && (
                <div className="flex justify-between text-xs">
                  <span className="text-[var(--text-faint)]">{t('funcEditor.warmInstancesLabel', { count: warmInstances })}</span>
                  <span className="text-[var(--text-dim)] font-[var(--font-mono)]">
                    {formatCurrency(cost.warmInstanceCost)}
                  </span>
                </div>
              )}
              <div className="pt-2 border-t border-[var(--panel-edge)] flex justify-between text-xs font-medium">
                <span className="text-[var(--text)]">{t('funcEditor.total')}</span>
                <span className="text-[var(--status-ok)] font-[var(--font-mono)]">{formatCurrency(cost.total)}</span>
              </div>
            </div>
          )}

          {/* Free tier notice */}
          {cost.total < 5 && (
            <p className="text-xs text-[var(--status-ok)] mt-3 flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-[var(--status-ok)]" />
              {t('funcEditor.freeTierNotice')}
            </p>
          )}

          {cost.total > 100 && (
            <p className="text-xs text-[var(--status-pending)] mt-3 flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-[var(--status-pending)]" />
              {t('funcEditor.optimizeNotice')}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Deployment Checklist */}
      <Card
        className="overflow-hidden"
        style={{
          background: 'var(--panel)',
          borderColor: 'var(--panel-edge)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <CardHeader className="pb-2 pt-4 px-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-[var(--status-ok)]" />
              <CardTitle className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>
                {t('funcEditor.deploymentChecklist')}
              </CardTitle>
            </div>
            <span className="text-xs font-[var(--font-mono)] text-[var(--text-faint)]">
              {completedCount}/{checklist.length}
            </span>
          </div>
          {/* Progress bar */}
          <div className="mt-2.5 h-1.5 rounded-full bg-[var(--panel-raised)] overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
                style={{
                  width: `${completionPct}%`,
                  background: completionPct === 100
                    ? 'var(--status-ok)'
                    : 'var(--accent)',
                }}
            />
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-4 space-y-2">
          {checklist.map((item) => (
            <div key={item.label} className="flex items-center gap-2.5 text-xs">
              {item.done ? (
                <CheckCircle2 className="w-3.5 h-3.5 text-[var(--status-ok)] shrink-0" />
              ) : (
                <Circle className="w-3.5 h-3.5 text-[var(--text-faint)] shrink-0" />
              )}
              <span className={item.done ? 'text-[var(--text-dim)]' : 'text-[var(--text-faint)]'}>
                {item.label}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

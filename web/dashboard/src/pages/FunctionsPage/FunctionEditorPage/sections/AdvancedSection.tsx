import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { ChevronDown, ChevronUp, Settings2, Zap } from 'lucide-react';
import { useState } from 'react';
import { InfoTip, SectionCard } from '../components/editor-ui';
import type { BackoffStrategy } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

const BACKOFF_OPTIONS: { value: BackoffStrategy; label: string; description: string }[] = [
  { value: 'fixed', label: 'Fixed', description: 'Same delay between each retry' },
  { value: 'linear', label: 'Linear', description: 'Delay increases linearly' },
  { value: 'exponential', label: 'Exponential', description: 'Delay doubles each retry' },
];

export function AdvancedSection({ editor }: Props) {
  const { retryPolicy, setRetryPolicy, warmInstances, setWarmInstances, markDirty } = editor;
  const [isOpen, setIsOpen] = useState(false);

  return (
    <SectionCard
      icon={<Settings2 className="w-4 h-4" />}
      title="Advanced Settings"
      step={8}
      description="Retry policy, warm instances, and concurrency controls"
    >
      {/* Collapsible toggle */}
      <button
        type="button"
        onClick={() => setIsOpen((o) => !o)}
        className="flex items-center gap-2 text-xs text-text-secondary hover:text-text-primary transition-colors w-full text-left"
        aria-expanded={isOpen}
      >
        {isOpen ? (
          <ChevronUp className="w-3.5 h-3.5 shrink-0" />
        ) : (
          <ChevronDown className="w-3.5 h-3.5 shrink-0" />
        )}
        {isOpen ? 'Hide advanced settings' : 'Show advanced settings'}
      </button>

      {isOpen && (
        <div className="space-y-5 pt-1">
          {/* Retry Policy */}
          <div className="space-y-4">
            <p className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              Retry Policy
            </p>

            {/* Max Retries */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs text-text-secondary flex items-center">
                  Max Retries
                  <InfoTip content="Number of automatic retries on failure before giving up (0–5)." />
                </Label>
                <span className="text-xs font-mono font-semibold text-text-primary">
                  {retryPolicy.maxRetries}
                </span>
              </div>
              <Slider
                min={0}
                max={5}
                step={1}
                value={[retryPolicy.maxRetries]}
                onValueChange={([v]) => {
                  setRetryPolicy((r) => ({ ...r, maxRetries: v }));
                  markDirty();
                }}
                className="w-full"
                aria-label="Max retries"
              />
              <div className="flex justify-between mt-1">
                <span className="text-xs text-text-muted">0</span>
                <span className="text-xs text-text-muted">5</span>
              </div>
            </div>

            {/* Backoff Strategy */}
            <div>
              <Label className="text-xs text-text-secondary mb-1.5 block">
                Backoff Strategy
                <InfoTip content="How the delay between retries grows." />
              </Label>
              <div className="grid grid-cols-3 gap-2">
                {BACKOFF_OPTIONS.map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => {
                      setRetryPolicy((r) => ({ ...r, backoffStrategy: opt.value }));
                      markDirty();
                    }}
                    className={`flex flex-col gap-0.5 p-2.5 rounded-lg border text-left transition-all ${
                      retryPolicy.backoffStrategy === opt.value
                        ? 'border-indigo-500/50 bg-indigo-500/10'
                        : 'border-border-subtle bg-bg-tertiary hover:border-border-default'
                    }`}
                    aria-pressed={retryPolicy.backoffStrategy === opt.value}
                  >
                    <span className="text-xs font-semibold text-text-primary">{opt.label}</span>
                    <span className="text-[10px] text-text-muted leading-tight">
                      {opt.description}
                    </span>
                  </button>
                ))}
              </div>
            </div>

            {/* Retry Delay */}
            <div>
              <Label htmlFor="backoff-ms" className="text-xs text-text-secondary mb-1.5 block">
                Retry Delay (ms)
                <InfoTip content="Base delay in milliseconds between retry attempts." />
              </Label>
              <Input
                id="backoff-ms"
                type="number"
                min={100}
                max={60000}
                step={100}
                value={retryPolicy.backoffMs}
                onChange={(e) => {
                  setRetryPolicy((r) => ({ ...r, backoffMs: Number(e.target.value) }));
                  markDirty();
                }}
                className="input w-36"
              />
            </div>
          </div>

          {/* Warm Instances */}
          <div className="space-y-3 pt-2 border-t border-border-subtle">
            <p className="text-xs font-semibold text-text-secondary uppercase tracking-wider pt-2">
              Pre-warming
            </p>
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs text-text-secondary flex items-center">
                  <Zap className="w-3 h-3 mr-1 text-amber-400" />
                  Warm Instances
                  <InfoTip content="Number of pre-warmed instances to keep alive. Reduces cold start latency (0–5)." />
                </Label>
                <span className="text-xs font-mono font-semibold text-text-primary">
                  {warmInstances}
                </span>
              </div>
              <Slider
                min={0}
                max={5}
                step={1}
                value={[warmInstances]}
                onValueChange={([v]) => {
                  setWarmInstances(v);
                  markDirty();
                }}
                className="w-full"
                aria-label="Warm instances"
              />
              <div className="flex justify-between mt-1">
                <span className="text-xs text-text-muted">0 (cold)</span>
                <span className="text-xs text-text-muted">5</span>
              </div>
              {warmInstances > 0 && (
                <p className="text-xs text-amber-400 mt-1.5">
                  ⚡ {warmInstances} instance{warmInstances !== 1 ? 's' : ''} will be kept warm —
                  adds to base cost
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </SectionCard>
  );
}

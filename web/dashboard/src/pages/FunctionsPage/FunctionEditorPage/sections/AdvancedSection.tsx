import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { Check, ChevronDown, ChevronUp, Settings2, Zap } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { InfoTip, SectionCard } from '../components/editor-ui';
import type { BackoffStrategy } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

const BACKOFF_OPTIONS: { value: BackoffStrategy; labelKey: string; descKey: string }[] = [
  { value: 'fixed', labelKey: 'funcEditor.fixed' as const, descKey: 'funcEditor.fixedDescription' as const },
  { value: 'linear', labelKey: 'funcEditor.linear' as const, descKey: 'funcEditor.linearDescription' as const },
  { value: 'exponential', labelKey: 'funcEditor.exponential' as const, descKey: 'funcEditor.exponentialDescription' as const },
];

export function AdvancedSection({ editor }: Props) {
  const { t } = useTranslation();
  const { retryPolicy, setRetryPolicy, warmInstances, setWarmInstances, markDirty } = editor;
  const [isOpen, setIsOpen] = useState(false);

  return (
    <SectionCard
      icon={<Settings2 className="w-4 h-4" />}
      title={t('funcEditor.advancedSettings')}
      step={8}
      description={t('funcEditor.advancedDescription')}
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
        {isOpen ? t('funcEditor.hideAdvancedSettings') : t('funcEditor.showAdvancedSettings')}
      </button>

      {isOpen && (
        <div className="space-y-5 pt-1">
          {/* Retry Policy */}
          <div className="space-y-4">
            <p className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              {t('funcEditor.retryPolicy')}
            </p>

            {/* Max Retries */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs text-text-secondary flex items-center">
                  {t('funcEditor.maxRetries')}
                  <InfoTip content={t('funcEditor.maxRetriesInfoTip')} />
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
                {t('funcEditor.backoffStrategy')}
                <InfoTip content={t('funcEditor.backoffInfoTip')} />
              </Label>
              <div className="grid grid-cols-3 gap-2">
                {BACKOFF_OPTIONS.map((opt) => {
                  const selected = retryPolicy.backoffStrategy === opt.value;
                  return (
                    <button
                      key={opt.value}
                      type="button"
                      onClick={() => {
                        setRetryPolicy((r) => ({ ...r, backoffStrategy: opt.value }));
                        markDirty();
                      }}
                      className={`relative flex flex-col gap-0.5 rounded-lg border-2 p-2.5 text-left transition-all duration-200 ${
                        selected
                          ? 'border-[#FF6B35] bg-[#FFF1EB] shadow-sm dark:border-[#FF8C42]/90 dark:bg-[#FF6B35]/25 dark:shadow-[0_0_0_1px_rgba(255,140,66,0.35)]'
                          : 'border-transparent bg-bg-tertiary hover:border-border-default hover:bg-bg-hover'
                      }`}
                      aria-pressed={selected}
                    >
                      {selected ? (
                        <span
                          className="absolute right-1.5 top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-[#FF6B35] text-white shadow-sm dark:bg-[#FF8C42]"
                          aria-hidden
                        >
                          <Check className="h-3 w-3" strokeWidth={2.5} />
                        </span>
                      ) : null}
                      <span className="pr-6 text-xs font-semibold text-text-primary">
                        {t(opt.labelKey)}
                      </span>
                      <span className="pr-6 text-[10px] leading-tight text-text-muted">
                        {t(opt.descKey)}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Retry Delay */}
            <div>
              <Label htmlFor="backoff-ms" className="text-xs text-text-secondary mb-1.5 block">
                {t('funcEditor.retryDelayMs')}
                <InfoTip content={t('funcEditor.retryDelayInfoTip')} />
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
              {t('funcEditor.preWarming')}
            </p>
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-xs text-text-secondary flex items-center">
                  <Zap className="w-3 h-3 mr-1 text-amber-400" />
                  {t('funcEditor.warmInstances')}
                  <InfoTip content={t('funcEditor.warmInstancesInfoTip')} />
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
                <span className="text-xs text-text-muted">{t('funcEditor.coldLabel')}</span>
                <span className="text-xs text-text-muted">5</span>
              </div>
              {warmInstances > 0 && (
                <p className="text-xs text-amber-400 mt-1.5">
                  ⚡ {t('funcEditor.warmInstancesNotice', { count: warmInstances })}
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </SectionCard>
  );
}

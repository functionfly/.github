import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { Timer } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { InfoTip, SectionCard } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';
import { formatTimeout } from '../utils';

type Props = { editor: FunctionEditorModel };

const MEMORY_MIN = 128;
const MEMORY_MAX = 2048;
const TIMEOUT_MIN = 100;
const TIMEOUT_MAX = 30000;
const CONCURRENCY_MIN = 1;
const CONCURRENCY_MAX = 100;

export function ResourceLimitsSection({ editor }: Props) {
  const { t } = useTranslation();
  const { resources, setResources, markDirty } = editor;

  return (
    <SectionCard
      icon={<Timer className="w-4 h-4" />}
      title={t('funcEditor.resourceLimits')}
      step={5}
      description={t('funcEditor.resourceLimitsDescription')}
    >
      <div className="space-y-5">
        {/* Memory */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <Label className="text-xs text-text-secondary flex items-center">
              {t('funcEditor.memory')}
              <InfoTip content={t('funcEditor.memoryInfoTip')} />
            </Label>
            <span className="text-xs font-mono font-semibold text-text-primary">
              {resources.memoryMb} MB
            </span>
          </div>
          <Slider
            min={MEMORY_MIN}
            max={MEMORY_MAX}
            step={128}
            value={[resources.memoryMb]}
            onValueChange={([v]) => {
              setResources((r) => ({ ...r, memoryMb: v }));
              markDirty();
            }}
            className="w-full"
            aria-label="Memory limit"
          />
          <div className="flex justify-between mt-1">
            <span className="text-xs text-text-muted">{MEMORY_MIN} MB</span>
            <span className="text-xs text-text-muted">{MEMORY_MAX} MB</span>
          </div>
        </div>

        {/* Timeout */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <Label className="text-xs text-text-secondary flex items-center">
              {t('funcEditor.timeout')}
              <InfoTip content={t('funcEditor.timeoutInfoTip')} />
            </Label>
            <span className="text-xs font-mono font-semibold text-text-primary">
              {formatTimeout(resources.timeoutMs)}
            </span>
          </div>
          <Slider
            min={TIMEOUT_MIN}
            max={TIMEOUT_MAX}
            step={100}
            value={[resources.timeoutMs]}
            onValueChange={([v]) => {
              setResources((r) => ({ ...r, timeoutMs: v }));
              markDirty();
            }}
            className="w-full"
            aria-label="Timeout"
          />
          <div className="flex justify-between mt-1">
            <span className="text-xs text-text-muted">{formatTimeout(TIMEOUT_MIN)}</span>
            <span className="text-xs text-text-muted">{formatTimeout(TIMEOUT_MAX)}</span>
          </div>
        </div>

        {/* Max Concurrency */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <Label htmlFor="concurrency" className="text-xs text-text-secondary flex items-center">
              {t('funcEditor.maxConcurrency')}
              <InfoTip content={t('funcEditor.maxConcurrencyInfoTip')} />
            </Label>
            <span className="text-xs font-mono font-semibold text-text-primary">
              {resources.maxConcurrency}
            </span>
          </div>
          <div className="flex items-center gap-3">
            <Slider
              min={CONCURRENCY_MIN}
              max={CONCURRENCY_MAX}
              step={1}
              value={[resources.maxConcurrency]}
              onValueChange={([v]) => {
                setResources((r) => ({ ...r, maxConcurrency: v }));
                markDirty();
              }}
              className="flex-1"
              aria-label="Max concurrency"
            />
            <Input
              id="concurrency"
              type="number"
              min={CONCURRENCY_MIN}
              max={CONCURRENCY_MAX}
              value={resources.maxConcurrency}
              onChange={(e) => {
                const v = Math.min(
                  CONCURRENCY_MAX,
                  Math.max(CONCURRENCY_MIN, Number(e.target.value))
                );
                setResources((r) => ({ ...r, maxConcurrency: v }));
                markDirty();
              }}
              className="input w-20 text-center"
            />
          </div>
        </div>
      </div>
    </SectionCard>
  );
}

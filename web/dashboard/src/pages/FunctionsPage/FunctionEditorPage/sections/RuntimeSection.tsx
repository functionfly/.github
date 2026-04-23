import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Check, Code2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { SectionCard } from '../components/editor-ui';
import { RUNTIME_META, RUNTIME_VERSIONS } from '../constants';
import type { Runtime } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function RuntimeSection({ editor }: Props) {
  const { t } = useTranslation();
  const { runtime, runtimeVersion, setRuntimeVersion, handleRuntimeChange, markDirty } = editor;

  return (
    <SectionCard icon={<Code2 className="w-4 h-4" />} title={t('funcEditor.runtimeConfiguration')} step={2}>
      <div>
        <Label className="text-xs font-medium text-text-secondary mb-2 block">{t('funcEditor.runtime')}</Label>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {(Object.entries(RUNTIME_META) as [Runtime, (typeof RUNTIME_META)[Runtime]][]).map(
            ([key, meta]) => (
              <button
                key={key}
                type="button"
                onClick={() => handleRuntimeChange(key)}
                className={`relative flex flex-col gap-2 rounded-lg border-2 p-4 pr-10 text-left transition-all duration-200 ${
                  runtime === key
                    ? 'border-[#FF6B35] bg-[#FFF1EB] shadow-sm dark:border-[#FF8C42]/90 dark:bg-[#FF6B35]/25 dark:shadow-[0_0_0_1px_rgba(255,140,66,0.35)]'
                    : 'border-transparent bg-bg-tertiary hover:border-border-default hover:bg-bg-hover'
                }`}
                aria-pressed={runtime === key}
              >
                {runtime === key ? (
                  <span
                    className="absolute right-2 top-2 flex h-6 w-6 items-center justify-center rounded-full bg-[#FF6B35] text-white shadow-sm dark:bg-[#FF8C42]"
                    aria-hidden
                  >
                    <Check className="h-3.5 w-3.5" strokeWidth={2.5} />
                  </span>
                ) : null}
                <div className="flex items-center gap-2 pr-7">
                  <span
                    className="h-2 w-2 shrink-0 rounded-full"
                    style={{ background: meta.color }}
                  />
                  <span className="text-sm font-semibold text-text-primary">{meta.label}</span>
                </div>
                <span className="text-xs text-text-muted leading-relaxed">{meta.description}</span>
              </button>
            )
          )}
        </div>
      </div>
      <div className="w-40">
        <Label
          htmlFor="runtime-version"
          className="text-xs font-medium text-text-secondary mb-1.5 block"
        >
          {t('funcEditor.version')}
        </Label>
        <Select
          value={runtimeVersion}
          onValueChange={(v) => {
            setRuntimeVersion(v);
            markDirty();
          }}
        >
          <SelectTrigger id="runtime-version">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {RUNTIME_VERSIONS[runtime].map((v) => (
              <SelectItem key={v} value={v}>
                {v}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </SectionCard>
  );
}

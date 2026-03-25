import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Code2 } from 'lucide-react';
import { SectionCard } from '../components/editor-ui';
import { RUNTIME_META, RUNTIME_VERSIONS } from '../constants';
import type { Runtime } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function RuntimeSection({ editor }: Props) {
  const { runtime, runtimeVersion, setRuntimeVersion, handleRuntimeChange, markDirty } = editor;

  return (
    <SectionCard icon={<Code2 className="w-4 h-4" />} title="Runtime Configuration" step={2}>
      <div>
        <Label className="text-xs font-medium text-text-secondary mb-2 block">Runtime</Label>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          {(Object.entries(RUNTIME_META) as [Runtime, (typeof RUNTIME_META)[Runtime]][]).map(
            ([key, meta]) => (
              <button
                key={key}
                type="button"
                onClick={() => handleRuntimeChange(key)}
                className={`flex flex-col gap-1.5 p-3 rounded-lg border text-left transition-all ${
                  runtime === key
                    ? 'border-indigo-500/50 bg-indigo-500/10 shadow-[0_0_0_1px_rgba(99,102,241,0.3)]'
                    : 'border-border-subtle bg-bg-tertiary hover:border-border-default hover:bg-bg-tertiary/80'
                }`}
                aria-pressed={runtime === key}
              >
                <div className="flex items-center gap-2">
                  <span
                    className="w-2 h-2 rounded-full shrink-0"
                    style={{ background: meta.color }}
                  />
                  <span className="text-sm font-semibold text-text-primary">{meta.label}</span>
                </div>
                <span className="text-xs text-text-muted leading-tight">{meta.description}</span>
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
          Version
        </Label>
        <Select
          value={runtimeVersion}
          onValueChange={(v) => {
            setRuntimeVersion(v);
            markDirty();
          }}
        >
          <SelectTrigger id="runtime-version" className="select">
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

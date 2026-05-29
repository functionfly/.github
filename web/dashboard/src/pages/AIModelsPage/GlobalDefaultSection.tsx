import type { ModelSelection } from '@/api/aiModels';
import { ModelPicker } from '@/components/ai/ModelPicker';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

type Props = {
  useSameModelEverywhere: boolean;
  globalDefault?: ModelSelection;
  onToggleSameModel: (checked: boolean) => void;
  onGlobalDefaultChange: (next: ModelSelection | undefined) => void;
  disabled?: boolean;
};

export function GlobalDefaultSection({
  useSameModelEverywhere,
  globalDefault,
  onToggleSameModel,
  onGlobalDefaultChange,
  disabled = false,
}: Props) {
  return (
    <div className="space-y-4 rounded-xl border border-border-subtle bg-bg-secondary/30 p-4">
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-1">
          <Label className="text-base">Use same model everywhere</Label>
          <p className="text-sm text-text-muted">
            Apply one org default across Composer, FRG, and other AI features.
          </p>
        </div>
        <Switch
          checked={useSameModelEverywhere}
          onCheckedChange={onToggleSameModel}
          disabled={disabled}
        />
      </div>
      {useSameModelEverywhere && (
        <div className="space-y-2">
          <Label>Global default model</Label>
          <ModelPicker
            feature="composer"
            capability="code"
            value={globalDefault}
            onChange={onGlobalDefaultChange}
            disabled={disabled}
            showOrgDefaultOption={false}
          />
        </div>
      )}
    </div>
  );
}

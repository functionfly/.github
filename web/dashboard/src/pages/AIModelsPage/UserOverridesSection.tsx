import type { ModelSelection } from '@/api/aiModels';
import { ModelPicker } from '@/components/ai/ModelPicker';
import { Label } from '@/components/ui/label';
import { USER_OVERRIDE_FEATURES } from './utils';

type Props = {
  overrides: Record<string, ModelSelection | undefined>;
  onChange: (feature: string, next: ModelSelection | undefined) => void;
  disabled?: boolean;
};

export function UserOverridesSection({ overrides, onChange, disabled = false }: Props) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-text-muted">
        Override org defaults for your account. These apply only to you and only when your
        organization allows user overrides.
      </p>
      <div className="grid gap-4 sm:grid-cols-2">
        {USER_OVERRIDE_FEATURES.map((feature) => (
          <div key={feature.key} className="space-y-2 rounded-lg border border-border-subtle p-4">
            <Label>{feature.label}</Label>
            <ModelPicker
              feature={feature.key}
              capability={feature.capability}
              value={overrides[feature.key]}
              onChange={(next) => onChange(feature.key, next)}
              disabled={disabled}
              showDefaultBadge
            />
          </div>
        ))}
      </div>
    </div>
  );
}

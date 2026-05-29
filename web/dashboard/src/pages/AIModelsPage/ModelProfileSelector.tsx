import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { Gauge, Sparkles, Zap } from 'lucide-react';
import { PROFILE_OPTIONS } from './utils';

type ProfileId = 'fast' | 'balanced' | 'premium' | 'custom';

type Props = {
  value: ProfileId;
  onChange: (profile: ProfileId) => void;
  disabled?: boolean;
};

const ICONS = {
  fast: Zap,
  balanced: Gauge,
  premium: Sparkles,
} as const;

export function ModelProfileSelector({ value, onChange, disabled = false }: Props) {
  return (
    <div className="grid gap-3 sm:grid-cols-3">
      {PROFILE_OPTIONS.map((profile) => {
        const Icon = ICONS[profile.id];
        const selected = value === profile.id;
        return (
          <button
            key={profile.id}
            type="button"
            disabled={disabled}
            onClick={() => onChange(profile.id)}
            className={cn(
              'rounded-xl border p-4 text-left transition-colors',
              'hover:border-brand-500/50 hover:bg-bg-tertiary/40',
              selected
                ? 'border-brand-500 bg-brand-500/10 ring-1 ring-brand-500/30'
                : 'border-border-subtle bg-bg-secondary/30',
              disabled && 'cursor-not-allowed opacity-60'
            )}
          >
            <div className="mb-2 flex items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <Icon className="h-4 w-4 text-brand-500" />
                <span className="font-medium text-text-primary">{profile.label}</span>
              </div>
              {selected && (
                <Badge variant="outline" className="text-[10px] uppercase tracking-wide">
                  Active
                </Badge>
              )}
            </div>
            <p className="text-sm text-text-muted">{profile.description}</p>
          </button>
        );
      })}
    </div>
  );
}

import { Shield, Check, Lock, Cpu, Box, Server } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { type SandboxTier, type TierInfo, getTierColor, getIsolationDescription } from '@/api/sandbox';

interface SandboxTierSelectorProps {
  tiers: TierInfo[];
  value: SandboxTier;
  onChange: (tier: SandboxTier) => void;
  disabled?: boolean;
  showDescription?: boolean;
  recommendedTier?: SandboxTier;
}

const TIER_ICONS: Record<SandboxTier, typeof Shield> = {
  wasm: Box,
  gvisor: Shield,
  docker: Server,
  microvm: Cpu,
};

export function SandboxTierSelector({
  tiers,
  value,
  onChange,
  disabled = false,
  showDescription = true,
  recommendedTier,
}: SandboxTierSelectorProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-foreground flex items-center gap-2">
          <Shield className="w-4 h-4" />
          Sandbox Isolation Tier
        </label>
        {recommendedTier && (
          <Badge variant="outline" className="text-xs bg-amber-500/10 text-amber-400 border-amber-500/30">
            Recommended: {recommendedTier}
          </Badge>
        )}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {tiers.map((tier) => {
          const Icon = TIER_ICONS[tier.id as SandboxTier] || Shield;
          const isSelected = value === tier.id;
          const isRecommended = tier.id === recommendedTier;

          return (
            <button
              key={tier.id}
              onClick={() => !disabled && tier.available && onChange(tier.id as SandboxTier)}
              disabled={disabled || !tier.available}
              className={cn(
                'relative flex items-start gap-3 p-3 rounded-lg border text-left transition-all',
                isSelected
                  ? 'bg-primary/10 border-primary/40 ring-1 ring-primary/20'
                  : tier.available
                    ? 'bg-card border-border hover:border-primary/30 hover:bg-accent/50'
                    : 'bg-muted/30 border-border/50 opacity-60 cursor-not-allowed',
              )}
            >
              {isRecommended && (
                <div className="absolute -top-1.5 -right-1.5">
                  <Badge className="text-[9px] px-1 py-0 bg-amber-500 text-black">REC</Badge>
                </div>
              )}

              <div className={cn(
                'flex-shrink-0 w-8 h-8 rounded-md flex items-center justify-center',
                getTierColor(tier.id as SandboxTier)
              )}>
                <Icon className="w-4 h-4" />
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{tier.name}</span>
                  {isSelected && <Check className="w-3.5 h-3.5 text-primary" />}
                  {!tier.available && (
                    <Badge variant="outline" className="text-[9px] px-1 py-0 text-muted-foreground">
                      Unavailable
                    </Badge>
                  )}
                </div>
                {showDescription && (
                  <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                    {tier.description}
                  </p>
                )}
                <div className="flex items-center gap-2 mt-1">
                  <Badge variant="secondary" className="text-[9px] px-1 py-0">
                    {tier.isolation_level.replace(/_/g, ' ')}
                  </Badge>
                  <span className={cn(
                    'text-[9px]',
                    tier.available ? 'text-green-500' : 'text-muted-foreground'
                  )}>
                    {tier.available ? 'Ready' : tier.status}
                  </span>
                </div>
              </div>
            </button>
          );
        })}
      </div>

      {showDescription && value && (
        <div className="flex items-start gap-2 p-2 rounded-md bg-muted/30 text-xs text-muted-foreground">
          <Lock className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />
          <span>
            {getIsolationDescription(tiers.find(t => t.id === value)?.isolation_level || '')}
          </span>
        </div>
      )}
    </div>
  );
}

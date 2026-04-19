'use client';

import { PLANS } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { ArrowUpRight, Check, Sparkles } from 'lucide-react';
import { Button } from './button';
import { Badge } from '@/components/ui/badge';

interface PlanCardProps {
  planId: string;
  isCurrent: boolean;
  isUpgrade: boolean;
  isDowngrade: boolean;
  onSelect: () => void;
  loading?: boolean;
  recommended?: boolean;
}

export function PlanCard({
  planId,
  isCurrent,
  isUpgrade,
  isDowngrade,
  onSelect,
  loading = false,
  recommended = false,
}: PlanCardProps) {
  const plan = PLANS[planId.toUpperCase() as keyof typeof PLANS];
  if (!plan) return null;

  const { name, price, limits } = plan;

  return (
    <div
      className={cn(
        'relative p-4 rounded-lg border transition-all',
        isCurrent && 'border-brand-500/50 bg-brand-500/10',
        isUpgrade &&
          !isCurrent &&
          'border-border-default hover:border-brand-500/50 hover:bg-amber-500/5',
        isDowngrade && !isCurrent && 'border-border-default opacity-60 hover:opacity-80',
        !isCurrent && !isUpgrade && !isDowngrade && 'border-border-default',
        isCurrent && 'shadow-[0_0_20px_rgba(255,107,53,0.15)]'
      )}
    >
      {recommended && !isCurrent && (
        <div className="absolute -top-2 left-1/2 -translate-x-1/2">
          <span className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-brand-700 text-white rounded-full">
            <Sparkles className="w-3 h-3" />
            Recommended
          </span>
        </div>
      )}

      {isCurrent && (
        <div className="absolute -top-2 left-1/2 -translate-x-1/2">
          <Badge variant="success" className="gap-1 border border-green-500/30">
            <Check className="w-3 h-3" />
            Current Plan
          </Badge>
        </div>
      )}

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <h4 className="font-medium text-text-primary">{name}</h4>
            <div className="flex items-baseline gap-1 mt-0.5">
              <span className="text-lg font-semibold text-text-primary">
                {typeof price === 'number' ? `$${price}` : price}
              </span>
              {typeof price === 'number' && <span className="text-xs text-text-muted">/mo</span>}
            </div>
          </div>
          {isUpgrade && !isCurrent && <ArrowUpRight className="w-5 h-5 text-brand-500" />}
        </div>

        <div className="space-y-1">
          <p className="text-xs text-text-muted">
            {limits.requests === Infinity
              ? 'Unlimited requests'
              : `${(limits.requests / 1_000_000).toFixed(0)}M requests/mo`}
          </p>
          <p className="text-xs text-text-muted">
            {limits.functions === Infinity
              ? 'Unlimited functions'
              : `${limits.functions} functions`}
          </p>
          <p className="text-xs text-text-muted">
            {limits.agents === Infinity ? 'Unlimited agents' : `${limits.agents} agents`}
          </p>
        </div>

        {isUpgrade && !isCurrent && (
          <Button
            onClick={onSelect}
            disabled={loading}
            size="sm"
            className="w-full bg-brand-500 hover:bg-brand-500/90"
          >
            <span className="flex items-center gap-2">
              <ArrowUpRight className="w-4 h-4" />
              Upgrade to {name}
            </span>
          </Button>
        )}

        {isCurrent && (
          <Button variant="ghost" size="sm" className="w-full text-text-muted" disabled>
            Current Plan
          </Button>
        )}

        {isDowngrade && !isCurrent && (
          <Button
            onClick={onSelect}
            disabled={loading}
            variant="outline"
            size="sm"
            className="w-full"
          >
            Downgrade
          </Button>
        )}

        {!isUpgrade && !isDowngrade && !isCurrent && (
          <Button
            onClick={onSelect}
            disabled={loading}
            variant="outline"
            size="sm"
            className="w-full"
          >
            Select Plan
          </Button>
        )}
      </div>
    </div>
  );
}

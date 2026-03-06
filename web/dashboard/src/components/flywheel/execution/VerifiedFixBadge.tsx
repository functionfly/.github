/**
 * VerifiedFixBadge - Verification status badge
 */

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  CheckCircle2,
  Shield,
  Award,
  Sparkles,
  Clock,
  AlertCircle,
} from 'lucide-react';

interface VerifiedFixBadgeProps {
  isVerified: boolean;
  score?: number;
  verifiedAt?: string;
  verifiedBy?: 'system' | 'community' | 'expert';
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function VerifiedFixBadge({
  isVerified,
  score,
  verifiedAt,
  verifiedBy = 'system',
  className,
  size = 'md',
}: VerifiedFixBadgeProps) {
  if (!isVerified) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge
              variant="outline"
              className={cn(
                'gap-1.5 border-slate-700 bg-slate-800/50 text-slate-400',
                size === 'sm' && 'text-xs',
                size === 'lg' && 'text-base px-3 py-1',
                className
              )}
            >
              <Clock className={cn(
                size === 'sm' ? 'h-3 w-3' : size === 'lg' ? 'h-5 w-5' : 'h-4 w-4'
              )} />
              Pending Verification
            </Badge>
          </TooltipTrigger>
          <TooltipContent className="bg-slate-900 border-slate-800">
            <p className="text-sm text-slate-300">
              This solution hasn't been verified yet.
              <br />
              Run the code to verify it works correctly.
            </p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  const verifierConfig = {
    system: { icon: Sparkles, label: 'AI Verified', color: 'text-indigo-400', bgColor: 'bg-indigo-500/10', borderColor: 'border-indigo-500/30' },
    community: { icon: Shield, label: 'Community Verified', color: 'text-emerald-400', bgColor: 'bg-emerald-500/10', borderColor: 'border-emerald-500/30' },
    expert: { icon: Award, label: 'Expert Verified', color: 'text-amber-400', bgColor: 'bg-amber-500/10', borderColor: 'border-amber-500/30' },
  };

  const config = verifierConfig[verifiedBy];
  const Icon = config.icon;

  const badgeContent = (
    <Badge
      variant="outline"
      className={cn(
        'gap-1.5 font-medium',
        config.bgColor,
        config.color,
        config.borderColor,
        size === 'sm' && 'text-xs',
        size === 'lg' && 'text-base px-3 py-1',
        className
      )}
    >
      <Icon className={cn(
        size === 'sm' ? 'h-3 w-3' : size === 'lg' ? 'h-5 w-5' : 'h-4 w-4'
      )} />
      {config.label}
      {score !== undefined && (
        <span className="ml-1 opacity-80">({score}/100)</span>
      )}
    </Badge>
  );

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {badgeContent}
        </TooltipTrigger>
        <TooltipContent className="bg-slate-900 border-slate-800 p-3">
          <div className="space-y-1">
            <p className="font-medium text-white">{config.label}</p>
            {score !== undefined && (
              <p className="text-sm text-slate-400">
                Score: <span className={cn('font-medium', config.color)}>{score}/100</span>
              </p>
            )}
            {verifiedAt && (
              <p className="text-xs text-slate-400">
                Verified {new Date(verifiedAt).toLocaleDateString()}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Compact verification icon
 */
export function VerificationIcon({
  isVerified,
  className,
}: {
  isVerified: boolean;
  className?: string;
}) {
  if (!isVerified) {
    return (
      <AlertCircle className={cn('h-4 w-4 text-slate-400', className)} />
    );
  }

  return (
    <CheckCircle2 className={cn('h-4 w-4 text-indigo-400', className)} />
  );
}

/**
 * Verification requirement checklist
 */
interface VerificationChecklistProps {
  hasTests: boolean;
  testPassRate?: number;
  hasBenchmark?: boolean;
  isDeterministic?: boolean;
  className?: string;
}

export function VerificationChecklist({
  hasTests,
  testPassRate,
  hasBenchmark,
  isDeterministic,
  className,
}: VerificationChecklistProps) {
  const items = [
    { label: 'Test cases pass', pass: testPassRate === 100 || (hasTests && testPassRate && testPassRate >= 80) },
    { label: 'Performance benchmark', pass: hasBenchmark },
    { label: 'Deterministic output', pass: isDeterministic },
  ];

  const passedCount = items.filter(i => i.pass).length;
  const totalCount = items.length;

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between text-sm">
        <span className="text-slate-400">Verification Requirements</span>
        <span className={cn(
          'font-medium',
          passedCount === totalCount ? 'text-emerald-400' : 'text-amber-400'
        )}>
          {passedCount}/{totalCount}
        </span>
      </div>
      <div className="space-y-1">
        {items.map((item) => (
          <div key={item.label} className="flex items-center gap-2 text-xs">
            {item.pass ? (
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
            ) : (
              <div className="h-3.5 w-3.5 rounded-full border border-slate-600" />
            )}
            <span className={item.pass ? 'text-slate-300' : 'text-slate-400'}>
              {item.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

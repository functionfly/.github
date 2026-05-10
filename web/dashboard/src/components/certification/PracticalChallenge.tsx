import { Wrench, ExternalLink } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import type { CertPracticalChallengePublic } from '@/api/certification';

interface PracticalChallengeProps {
  challenge: CertPracticalChallengePublic;
  examId: string;
}

const difficultyColors: Record<string, string> = {
  easy: 'bg-emerald-500/10 text-emerald-500',
  medium: 'bg-amber-500/10 text-amber-500',
  hard: 'bg-red-500/10 text-red-500',
};

export function PracticalChallenge({ challenge, examId }: PracticalChallengeProps) {
  return (
    <div className="rounded-xl border border-theme bg-card p-6">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-500/10">
            <Wrench className="h-5 w-5 text-purple-500" />
          </div>
          <div>
            <h4 className="font-medium text-text-primary">{challenge.name}</h4>
            <div className="flex items-center gap-2 mt-1">
              <Badge variant="outline" className="text-xs">{challenge.category}</Badge>
              <span className={cn(
                'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                difficultyColors[challenge.difficulty] || difficultyColors.medium
              )}>
                {challenge.difficulty}
              </span>
            </div>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm font-bold text-text-primary">{challenge.points} pts</p>
          <p className="text-xs text-text-muted">{challenge.time_limit_minutes} min</p>
        </div>
      </div>

      <p className="text-sm text-text-secondary mb-4 whitespace-pre-wrap">
        {challenge.description}
      </p>

      <Button variant="outline" size="sm">
        <ExternalLink className="h-4 w-4" />
        Open Challenge Environment
      </Button>
    </div>
  );
}

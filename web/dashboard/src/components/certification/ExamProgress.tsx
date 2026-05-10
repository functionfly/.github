import { cn } from '@/lib/utils';
import { CheckCircle2, Circle, ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ExamProgressProps {
  currentIndex: number;
  totalQuestions: number;
  answeredCount: number;
  answers: Record<string, string>;
  questionIds: string[];
  onNavigate: (index: number) => void;
}

export function ExamProgress({
  currentIndex,
  totalQuestions,
  answeredCount,
  answers,
  questionIds,
  onNavigate,
}: ExamProgressProps) {
  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex items-center justify-between gap-2 text-sm flex-wrap">
        <span className="text-text-muted">
          {answeredCount} of {totalQuestions} answered
        </span>
        <span className="text-text-muted">
          {totalQuestions - answeredCount} remaining
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-2 w-full rounded-full bg-bg-tertiary overflow-hidden">
        <div
          className="h-full rounded-full bg-brand-500 transition-all duration-300"
          style={{ width: `${totalQuestions > 0 ? (answeredCount / totalQuestions) * 100 : 0}%` }}
        />
      </div>

      {/* Question grid — scrollable on small screens */}
      <div className="overflow-x-auto pb-1">
        <div className="flex flex-wrap gap-2">
          {questionIds.map((qId, index) => {
            const isAnswered = !!answers[qId];
            const isCurrent = index === currentIndex;

            return (
              <button
                key={qId}
                onClick={() => onNavigate(index)}
                className={cn(
                  'flex-none w-10 h-10 items-center justify-center rounded-md text-xs font-medium transition-all',
                  isCurrent
                    ? 'bg-brand-500 text-white ring-2 ring-brand-500/30'
                    : isAnswered
                      ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20'
                      : 'bg-card border border-theme text-text-muted hover:border-brand-500/50'
                )}
              >
                {isAnswered ? (
                  <CheckCircle2 className="h-4 w-4" />
                ) : (
                  <span>{index + 1}</span>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between gap-2 pt-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onNavigate(currentIndex - 1)}
          disabled={currentIndex <= 0}
        >
          <ChevronLeft className="h-4 w-4" />
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onNavigate(currentIndex + 1)}
          disabled={currentIndex >= totalQuestions - 1}
        >
          Next
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

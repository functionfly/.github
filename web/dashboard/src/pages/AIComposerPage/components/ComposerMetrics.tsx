import { AlertTriangle, BarChart3, CheckCircle2, Clock, Coins, XCircle } from 'lucide-react';
import { Progress } from '@/components/ui/progress';
import { Separator } from '@/components/ui/separator';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { COMPLEXITY_COLORS } from '../constants';
import { calculateCost } from '../utils';

export function ComplexityProgress({
  complexity,
  isGenerating,
}: {
  complexity?: string;
  isGenerating: boolean;
}) {
  const complexityLevel = complexity || 'simple';
  const colors =
    COMPLEXITY_COLORS[complexityLevel as keyof typeof COMPLEXITY_COLORS] ||
    COMPLEXITY_COLORS.simple;

  const progressValue =
    {
      simple: 33,
      moderate: 66,
      complex: 100,
    }[complexityLevel] || 33;

  if (isGenerating) {
    return (
      <div className="space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">Analyzing complexity...</span>
          <span className="animate-pulse">Processing</span>
        </div>
        <Progress value={undefined} className="h-1.5 animate-pulse" />
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">Complexity</span>
        <span className={`font-medium ${colors.text}`}>{colors.label}</span>
      </div>
      <div className="relative h-1.5 bg-muted rounded-full overflow-hidden">
        <div
          className={`absolute h-full ${colors.bg} transition-all duration-500 ease-out rounded-full`}
          style={{ width: `${progressValue}%` }}
        />
        <div className="absolute top-0 left-[33%] w-0.5 h-full bg-white/50" />
        <div className="absolute top-0 left-[66%] w-0.5 h-full bg-white/50" />
      </div>
      <div className="flex justify-between text-[10px] text-muted-foreground">
        <span>Simple</span>
        <span>Moderate</span>
        <span>Complex</span>
      </div>
    </div>
  );
}

export function ConfidenceDisplay({ score }: { score?: number }) {
  if (score === undefined || score === null) return null;

  const percentage = Math.round(score * 100);
  let colorClass = 'text-green-600 dark:text-green-400';
  let bgClass = 'bg-green-500/10';
  let Icon = CheckCircle2;

  if (percentage < 70) {
    colorClass = 'text-yellow-600 dark:text-yellow-400';
    bgClass = 'bg-yellow-500/10';
    Icon = AlertTriangle;
  }
  if (percentage < 50) {
    colorClass = 'text-red-600 dark:text-red-400';
    bgClass = 'bg-red-500/10';
    Icon = XCircle;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={`flex items-center gap-1.5 px-2 py-1 rounded-md ${bgClass} ${colorClass} text-xs font-medium`}
          >
            <Icon className="w-3.5 h-3.5" />
            <span>{percentage}% confidence</span>
          </div>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p className="text-xs max-w-[200px]">
            AI confidence score based on code complexity, clarity of requirements, and pattern
            recognition.
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export function TokenUsageDisplay({
  tokens_used,
  latency_ms,
}: {
  tokens_used?: { prompt: number; completion: number; total: number };
  latency_ms?: number;
}) {
  if (!tokens_used || tokens_used.total === 0) return null;

  const cost = calculateCost(tokens_used);

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="flex items-center gap-3 px-3 py-2 rounded-md bg-muted/50 text-xs">
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <BarChart3 className="w-3.5 h-3.5" />
              <span>{tokens_used.total.toLocaleString()} tokens</span>
            </div>
            <Separator orientation="vertical" className="h-3" />
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <Coins className="w-3.5 h-3.5" />
              <span>~${cost.toFixed(4)}</span>
            </div>
            {latency_ms && latency_ms > 0 && (
              <>
                <Separator orientation="vertical" className="h-3" />
                <div className="flex items-center gap-1.5 text-muted-foreground">
                  <Clock className="w-3.5 h-3.5" />
                  <span>{(latency_ms / 1000).toFixed(2)}s</span>
                </div>
              </>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <div className="text-xs space-y-1">
            <p>
              <strong>Prompt:</strong> {tokens_used.prompt.toLocaleString()} tokens
            </p>
            <p>
              <strong>Completion:</strong> {tokens_used.completion.toLocaleString()} tokens
            </p>
            <p>
              <strong>Total:</strong> {tokens_used.total.toLocaleString()} tokens
            </p>
            <Separator className="my-1" />
            <p>
              <strong>Est. Cost:</strong> ${cost.toFixed(6)} USD
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

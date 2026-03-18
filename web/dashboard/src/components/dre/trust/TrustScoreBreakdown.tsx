import { Shield, RefreshCw, HardDrive, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface TrustScoreBreakdownProps {
  /** Determinism score (0-100) */
  determinismScore: number;
  /** Replay consistency score (0-100) */
  replayConsistency: number;
  /** Resource stability score (0-100); omit when not available from API */
  resourceStability?: number;
  /** Number of drift incidents */
  driftIncidents: number;
  /** Overall trust score (0-100) */
  overallScore: number;
  /** Custom className */
  className?: string;
}

interface ScoreBarProps {
  label: string;
  value: number;
  icon: React.ElementType;
  color: string;
}

function ScoreBar({ label, value, icon: Icon, color }: ScoreBarProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2">
          <Icon className={cn("h-4 w-4", color)} />
          <span>{label}</span>
        </div>
        <span className="font-medium">{value}%</span>
      </div>
      <div className="h-2 bg-bg-secondary rounded-full overflow-hidden">
        <div
          className={cn("h-full rounded-full transition-all", color.replace("text-", "bg-"))}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  );
}

export function TrustScoreBreakdown({
  determinismScore,
  replayConsistency,
  resourceStability,
  driftIncidents,
  overallScore,
  className,
}: TrustScoreBreakdownProps) {
  const getScoreColor = (score: number) => {
    if (score >= 80) return "text-green-500";
    if (score >= 50) return "text-yellow-500";
    return "text-red-500";
  };

  const getScoreBgColor = (score: number) => {
    if (score >= 80) return "bg-green-500";
    if (score >= 50) return "bg-yellow-500";
    return "bg-red-500";
  };

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          <Shield className="h-4 w-4" />
          Trust Score Breakdown
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Overall Score */}
        <div className="text-center py-4">
          <div className="relative inline-block">
            <svg className="w-24 h-24" viewBox="0 0 100 100">
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="currentColor"
                strokeWidth="8"
                className="text-bg-secondary"
              />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="currentColor"
                strokeWidth="8"
                strokeDasharray={`${overallScore * 2.83} 283`}
                strokeLinecap="round"
                className={cn("transition-all", getScoreBgColor(overallScore))}
                transform="rotate(-90 50 50)"
              />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className={cn("text-2xl font-bold", getScoreColor(overallScore))}>
                {overallScore}
              </span>
            </div>
          </div>
          <p className="text-sm text-muted-foreground mt-2">Overall Trust Score</p>
        </div>

        {/* Score Bars */}
        <div className="space-y-4">
          <ScoreBar
            label="Determinism"
            value={determinismScore}
            icon={Shield}
            color="text-blue-500"
          />
          <ScoreBar
            label="Replay Consistency"
            value={replayConsistency}
            icon={RefreshCw}
            color="text-purple-500"
          />
          {resourceStability !== undefined ? (
            <ScoreBar
              label="Resource Stability"
              value={resourceStability}
              icon={HardDrive}
              color="text-cyan-500"
            />
          ) : (
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-muted-foreground" />
                <span>Resource Stability</span>
              </div>
              <span className="text-muted-foreground">—</span>
            </div>
          )}
        </div>

        {/* Drift Incidents */}
        <div className="flex items-center justify-between p-3 bg-bg-secondary rounded-lg">
          <div className="flex items-center gap-2">
            <AlertTriangle className={cn("h-4 w-4", driftIncidents > 0 ? "text-yellow-500" : "text-green-500")} />
            <span className="text-sm">Drift Incidents</span>
          </div>
          <span className={cn("font-medium", driftIncidents > 0 ? "text-yellow-500" : "text-green-500")}>
            {driftIncidents}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

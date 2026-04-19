import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { SeverityBadge, SeverityBar, SeverityCounts } from './SeverityBadge';
import { type ScanSummary, type SeverityLevel, calculateRiskScore, formatDuration, groupVulnerabilitiesBySeverity } from './types';
import { 
  Shield, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  Activity,
  Target,
  Percent,
  ScanLine
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface ScanSummaryCardProps {
  summary: ScanSummary;
  scanType: string;
  target: string;
  startedAt: string;
  completedAt?: string;
  duration?: number;
  status: 'running' | 'completed' | 'failed';
  className?: string;
}

export function ScanSummaryCard({
  summary,
  scanType,
  target,
  startedAt,
  completedAt,
  duration,
  status,
  className,
}: ScanSummaryCardProps) {
  const riskScore = calculateRiskScore(summary);
  const counts = {
    critical: summary.critical_count,
    high: summary.high_count,
    medium: summary.medium_count,
    low: summary.low_count,
    info: summary.info_count,
  };

  const getRiskColor = (score: number) => {
    if (score >= 80) return 'text-red-500';
    if (score >= 60) return 'text-orange-500';
    if (score >= 40) return 'text-yellow-500';
    if (score >= 20) return 'text-blue-500';
    return 'text-green-500';
  };

  const getRiskLabel = (score: number) => {
    if (score >= 80) return 'Critical Risk';
    if (score >= 60) return 'High Risk';
    if (score >= 40) return 'Medium Risk';
    if (score >= 20) return 'Low Risk';
    return 'Minimal Risk';
  };

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              'p-2 rounded-lg',
              status === 'completed' ? 'bg-green-500/10' :
              status === 'running' ? 'bg-blue-500/10' :
              'bg-red-500/10'
            )}>
              {status === 'completed' ? (
                <CheckCircle className="w-5 h-5 text-green-500" />
              ) : status === 'running' ? (
                <ScanLine className="w-5 h-5 text-blue-500 animate-pulse" />
              ) : (
                <AlertTriangle className="w-5 h-5 text-red-500" />
              )}
            </div>
            <div>
              <CardTitle className="text-lg">{scanType}</CardTitle>
              <p className="text-sm text-text-secondary">{target}</p>
            </div>
          </div>
          <div className="text-right">
            <div className={cn('text-2xl font-bold', getRiskColor(riskScore))}>
              {riskScore}%
            </div>
            <p className="text-xs text-text-muted">{getRiskLabel(riskScore)}</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Risk Score Progress */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-text-secondary">Risk Score</span>
            <span className={cn('font-medium', getRiskColor(riskScore))}>
              {getRiskLabel(riskScore)}
            </span>
          </div>
          <Progress 
            value={riskScore} 
            className={cn(
              'h-2',
              riskScore >= 80 ? '[&>div]:bg-red-500' :
              riskScore >= 60 ? '[&>div]:bg-orange-500' :
              riskScore >= 40 ? '[&>div]:bg-yellow-500' :
              riskScore >= 20 ? '[&>div]:bg-blue-500' :
              '[&>div]:bg-green-500'
            )}
          />
        </div>

        {/* Severity Distribution Bar */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-text-secondary">Vulnerability Distribution</span>
            <span className="font-medium text-white">{summary.total_vulnerabilities} total</span>
          </div>
          <SeverityBar 
            counts={counts} 
            total={summary.total_vulnerabilities}
            className="h-3"
          />
        </div>

        {/* Severity Counts */}
        <SeverityCounts counts={counts} />

        {/* Stats Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
          <StatCard
            icon={Target}
            label="Target"
            value={target.split('/').pop() || target}
            className="col-span-2 sm:col-span-1"
          />
          <StatCard
            icon={Shield}
            label="Coverage"
            value={`${Math.round(summary.coverage_percentage)}%`}
          />
          <StatCard
            icon={Activity}
            label="Compliance"
            value={summary.compliance_score ? `${Math.round(summary.compliance_score)}%` : 'N/A'}
          />
          <StatCard
            icon={Clock}
            label="Duration"
            value={duration ? formatDuration(duration) : 'In Progress'}
          />
        </div>
      </CardContent>
    </Card>
  );
}

interface StatCardProps {
  icon: typeof Shield;
  label: string;
  value: string;
  className?: string;
}

function StatCard({ icon: Icon, label, value, className }: StatCardProps) {
  return (
    <div className={cn(
      'flex items-center gap-3 p-3 rounded-lg bg-muted/50',
      className
    )}>
      <Icon className="w-4 h-4 text-text-muted shrink-0" />
      <div className="min-w-0">
        <p className="text-xs text-text-muted">{label}</p>
        <p className="text-sm font-medium text-white truncate" title={value}>
          {value}
        </p>
      </div>
    </div>
  );
}

interface EmptyScanStateProps {
  message?: string;
  className?: string;
}

export function EmptyScanState({ 
  message = 'No vulnerabilities found. Your containers are secure!',
  className 
}: EmptyScanStateProps) {
  return (
    <Card className={className}>
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <div className="w-16 h-16 rounded-full bg-green-500/10 flex items-center justify-center mb-4">
          <Shield className="w-8 h-8 text-green-500" />
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">All Clear!</h3>
        <p className="text-text-secondary max-w-sm">{message}</p>
      </CardContent>
    </Card>
  );
}

interface ScanErrorStateProps {
  error: string;
  onRetry?: () => void;
  className?: string;
}

export function ScanErrorState({ error, onRetry, className }: ScanErrorStateProps) {
  return (
    <Card className={cn('border-red-500/30', className)}>
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <div className="w-16 h-16 rounded-full bg-red-500/10 flex items-center justify-center mb-4">
          <AlertTriangle className="w-8 h-8 text-red-500" />
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">Scan Failed</h3>
        <p className="text-text-secondary max-w-sm mb-4">{error}</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            Retry Scan
          </button>
        )}
      </CardContent>
    </Card>
  );
}

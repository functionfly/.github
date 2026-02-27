import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Shield, CheckCircle, TrendingUp, TrendingDown, AlertTriangle } from 'lucide-react';
import { EXCELLENT_SECURITY_THRESHOLD } from '../constants';

interface SecurityScoreOverviewProps {
  securityScore: number;
  lastUpdated: Date;
}

export function SecurityScoreOverview({ securityScore, lastUpdated }: SecurityScoreOverviewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-green-500" />
          Security Score
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-2xl font-bold text-green-500">{securityScore}%</p>
            <p className="text-sm text-muted-foreground">Overall security rating</p>
          </div>
          <div className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-500" />
            <span className="text-sm font-medium">Excellent Security Posture</span>
          </div>
        </div>
        <Progress value={securityScore} className="h-2" />
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Last updated: {new Intl.DateTimeFormat('en-US', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit'
            }).format(lastUpdated)}
          </p>
          <div className="flex items-center gap-2">
            {securityScore >= EXCELLENT_SECURITY_THRESHOLD ? (
              <TrendingUp className="h-4 w-4 text-green-500" />
            ) : securityScore >= 95 ? (
              <TrendingDown className="h-4 w-4 text-yellow-500" />
            ) : (
              <AlertTriangle className="h-4 w-4 text-red-500" />
            )}
            <span className="text-xs text-muted-foreground">Live</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
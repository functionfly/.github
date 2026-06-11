import { Card, CardContent } from '@/components/ui/card';
import { TrendingUp, ArrowDownRight, AlertCircle, Activity, Zap } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Insight } from '../hooks/useInsights';

interface InsightsSectionProps {
  insights: Insight[];
}

const iconComponents = {
  TrendingUp,
  ArrowDownRight,
  AlertCircle,
  Activity,
  Zap,
};

export function InsightsSection({ insights }: InsightsSectionProps) {
  if (insights.length === 0) return null;

  return (
    <div className="usage-insights-grid">
      {insights.slice(0, 4).map((insight, idx) => {
        const IconComponent = iconComponents[insight.iconName];
        return (
          <Card
            key={idx}
            className={cn(
              "usage-insight-card",
              insight.type === 'success' && "usage-insight-card-success",
              insight.type === 'warning' && "usage-insight-card-warning",
              insight.type === 'error' && "usage-insight-card-error"
            )}
          >
            <CardContent className="p-4">
              <div className="usage-insight-header">
                <div
                  className={cn(
                    "usage-insight-icon",
                    insight.type === 'success' && "usage-insight-icon-success",
                    insight.type === 'warning' && "usage-insight-icon-warning",
                    insight.type === 'error' && "usage-insight-icon-error",
                    insight.type === 'info' && "usage-insight-icon-info"
                  )}
                >
                  <IconComponent className="h-4 w-4" />
                </div>
                <div className="usage-insight-content">
                  <p className="usage-insight-title">{insight.title}</p>
                  <p className="usage-insight-message">{insight.message}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

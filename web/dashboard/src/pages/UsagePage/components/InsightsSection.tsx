import { Card, CardContent } from '@/components/ui/card';
import { TrendingUp, ArrowDownRight, AlertCircle, Activity, Zap } from 'lucide-react';
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
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {insights.slice(0, 4).map((insight, idx) => {
        const IconComponent = iconComponents[insight.iconName];
        return (
          <Card
            key={idx}
            className={`border-l-4 ${
              insight.type === 'success'
                ? 'border-l-emerald-500'
                : insight.type === 'warning'
                  ? 'border-l-amber-500'
                  : insight.type === 'error'
                    ? 'border-l-red-500'
                    : 'border-l-blue-500'
            } bg-card`}
          >
            <CardContent className="p-4">
              <div className="flex items-start gap-3">
                <div
                  className={`mt-0.5 ${
                    insight.type === 'success'
                      ? 'text-emerald-500'
                      : insight.type === 'warning'
                        ? 'text-amber-500'
                        : insight.type === 'error'
                          ? 'text-red-500'
                          : 'text-blue-500'
                  }`}
                >
                  <IconComponent className="h-4 w-4" />
                </div>
                <div>
                  <p className="font-semibold text-sm">{insight.title}</p>
                  <p className="text-xs text-text-secondary mt-1">{insight.message}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

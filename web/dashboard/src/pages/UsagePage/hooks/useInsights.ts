import { useMemo } from 'react';
import type { CostSummary, FunctionCostSummary } from '@/api/usageAnalytics';
import { formatCostUsd } from '@/api/usageAnalytics';

interface FunctionCostsResponse {
  tenant_id: string;
  period_start: string;
  period_end: string;
  function_count: number;
  functions: FunctionCostSummary[];
}

export type IconName = 'TrendingUp' | 'ArrowDownRight' | 'AlertCircle' | 'Activity' | 'Zap';

interface InsightsParams {
  costSummary: CostSummary | null | undefined;
  forecast: { predicted_monthly_cost_usd?: number } | null | undefined;
  spendCap: { spend_cap_usd?: number } | null | undefined;
  functionCosts: FunctionCostsResponse | null | undefined;
  periodComparison: { change: number };
}

export type InsightType = 'success' | 'warning' | 'error' | 'info';

export interface Insight {
  type: InsightType;
  title: string;
  message: string;
  iconName: IconName;
}

export function useInsights({
  costSummary,
  forecast,
  spendCap,
  functionCosts,
}: InsightsParams): Insight[] {
  return useMemo(() => {
    const items: Insight[] = [];

    // Cost trend alerts
    if (costSummary) {
      const { cost_trend_percent } = costSummary;
      if (cost_trend_percent > 20) {
        items.push({
          type: 'warning',
          title: 'Cost Trend Alert',
          message: `Costs increased by ${cost_trend_percent.toFixed(1)}% vs previous period`,
          iconName: 'TrendingUp',
        });
      } else if (cost_trend_percent < -10) {
        items.push({
          type: 'success',
          title: 'Cost Optimization',
          message: `Costs decreased by ${Math.abs(cost_trend_percent).toFixed(1)}% vs previous period`,
          iconName: 'ArrowDownRight',
        });
      }
    }

    // Forecast alerts
    if (forecast?.predicted_monthly_cost_usd && spendCap?.spend_cap_usd) {
      if (forecast.predicted_monthly_cost_usd > spendCap.spend_cap_usd) {
        items.push({
          type: 'error',
          title: 'Forecast Alert',
          message: `Predicted monthly cost (${formatCostUsd(forecast.predicted_monthly_cost_usd)}) exceeds your spend cap`,
          iconName: 'AlertCircle',
        });
      }
    }

    // Error rate alerts
    if (functionCosts?.functions && functionCosts.functions.length > 0) {
      const totalSuccesses = functionCosts.functions.reduce(
        (acc, fn) => acc + fn.success_executions,
        0
      );
      const totalExecutions = functionCosts.functions.reduce(
        (acc, fn) => acc + fn.total_executions,
        0
      );
      const overallSuccessRate = totalExecutions > 0 ? totalSuccesses / totalExecutions : 1;
      if (overallSuccessRate < 0.95) {
        items.push({
          type: 'warning',
          title: 'Error Rate Alert',
          message: `Overall success rate is ${(overallSuccessRate * 100).toFixed(1)}% - consider reviewing error logs`,
          iconName: 'Activity',
        });
      }
    }

    // Data transfer optimization tip
    if (costSummary?.cost_breakdown) {
      const { execution, data_transfer } = costSummary.cost_breakdown;
      const total = execution + data_transfer;
      if (total > 0 && data_transfer / total > 0.3) {
        items.push({
          type: 'info',
          title: 'Data Transfer Optimization',
          message: 'Data transfer costs are significant - consider caching strategies',
          iconName: 'Zap',
        });
      }
    }

    return items;
  }, [costSummary, forecast, spendCap, functionCosts]);
}

import { useMemo } from 'react';
import { formatDate } from '../utils';
import { COLORS } from '../constants';
import type { CostSummary, FunctionCostSummary, DailyCostBreakdown, RegionCostBreakdown } from '@/api/usageAnalytics';

interface UsageDataPoint {
  time: string;
  value: number;
}

interface ExecutionRatePoint {
  time: string;
  rate: number;
}

interface ChartDataParams {
  usageData: { data?: Array<{ time: string; value: number }> } | undefined;
  executionRateDataRes: { data?: Array<{ time: string; rate: number }> } | undefined;
  periodData: { daily_breakdown?: DailyCostBreakdown[] } | null | undefined;
  costSummary: CostSummary | null | undefined;
  regionData: { regions?: Record<string, RegionCostBreakdown> } | null | undefined;
  functionCosts: FunctionCostSummary | null | undefined;
}

export function useChartData({
  usageData,
  executionRateDataRes,
  periodData,
  costSummary,
  regionData,
  functionCosts,
}: ChartDataParams) {
  // Usage graph data transformation
  const usageGraphData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + 'Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      value: Number(d.value),
    }));
  }, [usageData]);

  // Execution rate data transformation
  const executionRateData = useMemo(() => {
    const raw = executionRateDataRes?.data ?? [];
    return raw.map((d) => ({ time: d.time, rate: Number(d.rate) }));
  }, [executionRateDataRes]);

  // Total usage calculation
  const totalUsage = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.reduce((sum, d) => sum + Number(d.value), 0);
  }, [usageData]);

  // Period comparison for insights
  const periodComparison = useMemo(() => {
    const raw = usageData?.data ?? [];
    const len = raw.length;
    const last7 = raw.slice(-7).reduce((s, d) => s + Number(d.value), 0);
    const prev7 = len >= 14 ? raw.slice(-14, -7).reduce((s, d) => s + Number(d.value), 0) : 0;
    const change = prev7 > 0 ? ((last7 - prev7) / prev7) * 100 : last7 > 0 ? 100 : 0;
    return { last7, prev7, change };
  }, [usageData]);

  // Daily chart data for area/bar charts
  const dailyChartData = useMemo(() => {
    if (!periodData?.daily_breakdown) return [];
    return periodData.daily_breakdown.map((day) => ({
      date: formatDate(day.date),
      fullDate: day.date,
      executions: day.executions,
      cost: day.cost_usd,
    }));
  }, [periodData]);

  // Cost breakdown data for pie chart
  const costBreakdownData = useMemo(() => {
    if (!costSummary?.cost_breakdown) return [];
    const { execution, compute, platform_fee, data_transfer } = costSummary.cost_breakdown;
    return [
      { name: 'Execution', value: execution, color: COLORS.execution },
      { name: 'Compute', value: compute, color: COLORS.compute },
      { name: 'Platform Fee', value: platform_fee, color: COLORS.platform_fee },
      { name: 'Data Transfer', value: data_transfer, color: COLORS.data_transfer },
    ].filter((d) => d.value > 0);
  }, [costSummary]);

  // Region chart data for pie chart
  const regionChartData = useMemo(() => {
    if (!regionData?.regions) return [];
    return Object.entries(regionData.regions)
      .map(([region, data]) => ({
        name: region,
        executions: data.total_executions,
        cost: data.total_cost_usd,
      }))
      .sort((a, b) => b.cost - a.cost);
  }, [regionData]);

  // Function chart data for list/bar chart
  const functionChartData = useMemo(() => {
    if (!functionCosts?.functions) return [];
    return functionCosts.functions
      .slice(0, 10)
      .map((fn) => ({
        name: fn.function_name,
        executions: fn.total_executions,
        cost: fn.total_cost_usd,
        successRate: fn.total_executions > 0 ? (fn.success_executions / fn.total_executions) * 100 : 0,
      }))
      .sort((a, b) => b.cost - a.cost);
  }, [functionCosts]);

  return {
    usageGraphData,
    executionRateData,
    totalUsage,
    periodComparison,
    dailyChartData,
    costBreakdownData,
    regionChartData,
    functionChartData,
  };
}

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatCard } from "@/components/common/StatCard";
import { Globe, Clock, AlertTriangle, TrendingUp, Loader2, Rocket } from "lucide-react";
import { EmptyState } from "@/components/ui";
import { functionsApi } from "@/api/functions";
import { dashboardApi } from "@/api/dashboard";
import { UsageGraph } from "@/components/dashboard";
import { LineChart } from "@/components/common/LineChart";

export function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState<"24h" | "7d" | "30d">("7d");

  // Map time range to days for API calls
  const days = timeRange === "24h" ? 1 : timeRange === "7d" ? 7 : 30;
  const hours = timeRange === "24h" ? 24 : timeRange === "7d" ? 168 : 720;

  // Fetch functions
  const { data: functionsData, isLoading: functionsLoading } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });

  const functions = functionsData?.functions ?? [];
  const activeFunctions = functions.filter((f) => f.status === "deployed").length;

  // Fetch usage data
  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ["analytics", "usage", days],
    queryFn: () => dashboardApi.getUsage(days),
    enabled: functions.length > 0,
  });

  // Fetch execution rate data
  const { data: executionRateData, isLoading: executionRateLoading } = useQuery({
    queryKey: ["analytics", "execution-rate", hours],
    queryFn: () => dashboardApi.getExecutionRate(hours),
    enabled: functions.length > 0,
  });

  // Fetch dashboard metrics (latency, success/error rate)
  const { data: metricsData, isLoading: metricsLoading } = useQuery({
    queryKey: ["analytics", "dashboard-metrics"],
    queryFn: () => dashboardApi.getMetrics(),
    enabled: functions.length > 0,
  });

  // Process usage data for chart
  const usageChartData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + "Z").toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      value: Number(d.value),
    }));
  }, [usageData]);

  // Process execution rate data for chart
  const executionChartData = useMemo(() => {
    const raw = executionRateData?.data ?? [];
    return raw.map((d) => ({
      time: d.time,
      value: Number(d.rate),
    }));
  }, [executionRateData]);

  // Calculate aggregate metrics from data
  const totalRequests = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.reduce((sum, d) => sum + Number(d.value), 0);
  }, [usageData]);

  const avgRequestsPerDay = useMemo(() => {
    const raw = usageData?.data ?? [];
    if (raw.length === 0) return 0;
    return Math.round(totalRequests / raw.length);
  }, [usageData, totalRequests]);

  // Latency and success/error rate from dashboard metrics API
  const avgLatency = metricsData?.avg_latency_ms != null ? Math.round(metricsData.avg_latency_ms) : undefined;
  const successRate = metricsData?.uptime_pct != null ? Math.round(metricsData.uptime_pct * 10) / 10 : undefined;
  const errorRate = successRate != null ? Math.round((100 - successRate) * 10) / 10 : undefined;

  const stats = [
    {
      title: "Total Requests",
      value: functionsLoading ? "—" : totalRequests > 0 ? totalRequests.toLocaleString() : "0",
      change: { value: 0, label: totalRequests > 0 ? `last ${timeRange}` : "no data yet" },
      icon: <Globe className="w-5 h-5 text-brand-500" />,
      trend: totalRequests > 0 ? "up" as const : "neutral" as const,
    },
    {
      title: "Avg Latency",
      value: functionsLoading || metricsLoading ? "—" : avgLatency != null && avgLatency >= 0 ? `${avgLatency}ms` : "—",
      change: { value: 0, label: avgLatency != null ? "last 7d" : "no data yet" },
      icon: <Clock className="w-5 h-5 text-brand-500" />,
      trend: avgLatency != null && avgLatency < 100 ? "down" as const : "neutral" as const,
    },
    {
      title: "Error Rate",
      value: functionsLoading || metricsLoading ? "—" : errorRate != null ? `${errorRate}%` : "—",
      change: { value: 0, label: errorRate != null ? "last 7d" : "no data yet" },
      icon: <AlertTriangle className="w-5 h-5 text-error" />,
      trend: errorRate != null && errorRate < 1 ? "down" as const : "neutral" as const,
    },
    {
      title: "Success Rate",
      value: functionsLoading || metricsLoading ? "—" : successRate != null ? `${successRate}%` : "—",
      change: { value: 0, label: successRate != null ? "last 7d" : "no data yet" },
      icon: <TrendingUp className="w-5 h-5 text-success" />,
      trend: successRate != null && successRate > 99 ? "up" as const : "neutral" as const,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Analytics</h1>
          <p className="text-text-secondary">Monitor your application's performance</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Time Range Controls */}
          {(["24h", "7d", "30d"] as const).map((range) => (
            <Button
              key={range}
              variant={timeRange === range ? "default" : "outline"}
              size="sm"
              onClick={() => setTimeRange(range)}
            >
              {range === "24h" ? "24 Hours" : range === "7d" ? "7 Days" : "30 Days"}
            </Button>
          ))}
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Loading or empty state */}
      {(functionsLoading || usageLoading || executionRateLoading || metricsLoading) && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
        </div>
      )}

      {!functionsLoading && !usageLoading && !executionRateLoading && !metricsLoading && functions.length === 0 && (
        <EmptyState
          icon={<Rocket className="h-8 w-8" />}
          title="No functions deployed yet"
          description="Deploy your first function to start seeing analytics data and performance metrics."
          action={
            <Button asChild>
              <a href="/functions">Deploy a Function</a>
            </Button>
          }
        />
      )}

      {/* Charts Grid - shown when functions exist */}
      {!functionsLoading && !usageLoading && !executionRateLoading && !metricsLoading && functions.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Requests Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              {usageChartData.length > 0 ? (
                <div className="h-[300px]">
                  <UsageGraph
                    data={usageChartData}
                    title=""
                    valueLabel="Requests"
                  />
                </div>
              ) : (
                <div className="h-[300px] flex items-center justify-center">
                  <p className="text-text-muted text-sm">No request data available for this time period.</p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Execution Rate</CardTitle>
            </CardHeader>
            <CardContent>
              {executionChartData.length > 0 ? (
                <div className="h-[300px]">
                  <LineChart
                    data={executionChartData}
                    series={[
                      {
                        key: "value",
                        name: "Executions",
                        color: "#6366f1",
                      },
                    ]}
                    xAxisKey="time"
                    showLegend={false}
                  />
                </div>
              ) : (
                <div className="h-[300px] flex items-center justify-center">
                  <p className="text-text-muted text-sm">No execution data available for this time period.</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

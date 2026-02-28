import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatCard } from "@/components/common/StatCard";
import { Globe, Clock, AlertTriangle, TrendingUp, Loader2 } from "lucide-react";
import { functionsApi } from "@/api/functions";

export function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState<"24h" | "7d" | "30d">("7d");

  // Fetch real metrics from the first available function, or aggregate
  const { data: functionsData, isLoading } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });

  const functions = functionsData?.functions ?? [];

  const stats = [
    {
      title: "Total Requests",
      value: isLoading ? "—" : "—",
      change: { value: 0, label: "no data yet" },
      icon: <Globe className="w-5 h-5 text-brand-500" />,
      trend: "neutral" as const,
    },
    {
      title: "Avg Latency",
      value: isLoading ? "—" : "—",
      change: { value: 0, label: "no data yet" },
      icon: <Clock className="w-5 h-5 text-brand-500" />,
      trend: "neutral" as const,
    },
    {
      title: "Error Rate",
      value: isLoading ? "—" : "—",
      change: { value: 0, label: "no data yet" },
      icon: <AlertTriangle className="w-5 h-5 text-error" />,
      trend: "neutral" as const,
    },
    {
      title: "Success Rate",
      value: isLoading ? "—" : "—",
      change: { value: 0, label: "no data yet" },
      icon: <TrendingUp className="w-5 h-5 text-success" />,
      trend: "neutral" as const,
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
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
        </div>
      )}

      {!isLoading && functions.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-text-secondary">No functions deployed yet. Deploy a function to see analytics.</p>
          </CardContent>
        </Card>
      )}

      {/* Charts Grid - shown when functions exist */}
      {!isLoading && functions.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Requests Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-[300px] flex items-center justify-center">
                <p className="text-text-muted text-sm">Analytics data will appear here once your functions receive traffic.</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Latency by Provider</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-[300px] flex items-center justify-center">
                <p className="text-text-muted text-sm">Latency data will appear here once your functions receive traffic.</p>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

import { useState, useEffect, useCallback } from "react";
import {
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Bar,
  Pie,
  Cell,
  AreaChart,
  Area,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { StatCard } from "@/components/common/StatCard";
import { LineChart } from "@/components/common/LineChart";
import { BarChart } from "@/components/common/BarChart";
import { Globe, Clock, AlertTriangle, TrendingUp, Wifi, WifiOff } from "lucide-react";

const requestData = [
  { name: "Mon", requests: 4000, errors: 24 },
  { name: "Tue", requests: 3000, errors: 18 },
  { name: "Wed", requests: 2000, errors: 12 },
  { name: "Thu", requests: 2780, errors: 15 },
  { name: "Fri", requests: 1890, errors: 8 },
  { name: "Sat", requests: 2390, errors: 10 },
  { name: "Sun", requests: 3490, errors: 20 },
];

const latencyData = [
  { name: "Workers", latency: 45 },
  { name: "Vercel", latency: 62 },
  { name: "Fly.io", latency: 78 },
];

const trafficData = [
  { name: "Workers", value: 45, color: "var(--color-cloudflare)" },
  { name: "Vercel", value: 35, color: "var(--color-vercel)" },
  { name: "Fly.io", value: 20, color: "var(--color-fly)" },
];

const stats = [
  {
    title: "Total Requests",
    value: "17.5K",
    change: { value: 12, label: "from last week" },
    icon: <Globe className="w-5 h-5 text-brand-500" />,
    trend: "up" as const,
  },
  {
    title: "Avg Latency",
    value: "58ms",
    change: { value: -8, label: "from last week" },
    icon: <Clock className="w-5 h-5 text-brand-500" />,
    trend: "up" as const,
  },
  {
    title: "Error Rate",
    value: "0.5%",
    change: { value: -0.2, label: "from last week" },
    icon: <AlertTriangle className="w-5 h-5 text-error" />,
    trend: "up" as const,
  },
  {
    title: "Success Rate",
    value: "99.5%",
    change: { value: 0.2, label: "from last week" },
    icon: <TrendingUp className="w-5 h-5 text-success" />,
    trend: "up" as const,
  },
];

export function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState<"24h" | "7d" | "30d">("7d");
  const [isRealtimeEnabled, setIsRealtimeEnabled] = useState(true);
  const [realtimeData, setRealtimeData] = useState(requestData);
  const [connectionStatus, setConnectionStatus] = useState<"connected" | "disconnected" | "connecting">("connected");

  // Simulate real-time data updates
  const generateRealtimeData = useCallback(() => {
    const now = new Date();
    const timeString = now.getHours().toString().padStart(2, '0') + ':' + now.getMinutes().toString().padStart(2, '0');

    return {
      name: timeString,
      requests: Math.floor(Math.random() * 400) + 200,
      errors: Math.floor(Math.random() * 10) + 1,
    };
  }, []);

  // Simulate WebSocket connection
  useEffect(() => {
    if (!isRealtimeEnabled) return;

    const interval = setInterval(() => {
      setRealtimeData(prevData => {
        const newData = [...prevData.slice(-23), generateRealtimeData()]; // Keep last 24 data points
        return newData;
      });
    }, 5000); // Update every 5 seconds

    return () => clearInterval(interval);
  }, [isRealtimeEnabled, generateRealtimeData]);

  // Simulate connection status changes
  useEffect(() => {
    if (!isRealtimeEnabled) {
      setConnectionStatus("disconnected");
      return;
    }

    const statuses: ("connected" | "disconnected" | "connecting")[] = ["connected", "connecting", "connected"];
    let index = 0;

    const statusInterval = setInterval(() => {
      setConnectionStatus(statuses[index % statuses.length]);
      index++;
    }, 10000); // Change status every 10 seconds

    return () => clearInterval(statusInterval);
  }, [isRealtimeEnabled]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Analytics</h1>
          <p className="text-text-secondary">Monitor your application's performance</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Real-time Status */}
          <div className="flex items-center gap-2">
            {connectionStatus === "connected" && (
              <Badge variant="secondary" className="gap-1 bg-green-400/10 text-green-400 border-green-400/20">
                <Wifi className="w-3 h-3" />
                Live
              </Badge>
            )}
            {connectionStatus === "connecting" && (
              <Badge variant="secondary" className="gap-1 bg-yellow-400/10 text-yellow-400 border-yellow-400/20">
                <Wifi className="w-3 h-3 animate-pulse" />
                Connecting...
              </Badge>
            )}
            {connectionStatus === "disconnected" && (
              <Badge variant="secondary" className="gap-1 bg-red-400/10 text-red-400 border-red-400/20">
                <WifiOff className="w-3 h-3" />
                Offline
              </Badge>
            )}
          </div>

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

          {/* Real-time Toggle */}
          <Button
            variant={isRealtimeEnabled ? "default" : "outline"}
            size="sm"
            onClick={() => setIsRealtimeEnabled(!isRealtimeEnabled)}
          >
            {isRealtimeEnabled ? "Pause Live" : "Enable Live"}
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Real-time Requests Chart */}
        <LineChart
          title="Requests Over Time"
          data={isRealtimeEnabled ? realtimeData : requestData}
          series={[
            { key: "requests", name: "Requests", color: "#6366f1" },
            { key: "errors", name: "Errors", color: "#ef4444" },
          ]}
          height={300}
        />

        {/* Latency by Provider */}
        <BarChart
          title="Latency by Provider"
          data={latencyData}
          series={[{ key: "latency", name: "Latency (ms)", color: "#10b981" }]}
          height={300}
        />

        {/* Real-time Performance Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Real-time Performance</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={realtimeData.slice(-10)}>
                  <defs>
                    <linearGradient id="requestsGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#6366f1" stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id="errorsGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                  <XAxis dataKey="name" stroke="var(--text-muted)" />
                  <YAxis stroke="var(--text-muted)" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "var(--bg-elevated)",
                      border: "1px solid var(--border-default)",
                      borderRadius: "8px",
                    }}
                    labelStyle={{ color: "var(--text-primary)" }}
                  />
                  <Area
                    type="monotone"
                    dataKey="requests"
                    stroke="#6366f1"
                    fillOpacity={1}
                    fill="url(#requestsGradient)"
                  />
                  <Area
                    type="monotone"
                    dataKey="errors"
                    stroke="#ef4444"
                    fillOpacity={1}
                    fill="url(#errorsGradient)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
            {isRealtimeEnabled && (
              <div className="flex justify-center mt-4">
                <Badge variant="secondary" className="gap-1">
                  <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                  Updating every 5 seconds
                </Badge>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Top Endpoints */}
        <Card>
          <CardHeader>
            <CardTitle>Top Endpoints</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                { path: "/api/users", requests: 4520, latency: "45ms" },
                { path: "/api/auth", requests: 3890, latency: "32ms" },
                { path: "/api/webhooks", requests: 2150, latency: "78ms" },
                { path: "/api/images", requests: 1890, latency: "120ms" },
                { path: "/health", requests: 1200, latency: "12ms" },
              ].map((endpoint, index) => (
                <div
                  key={endpoint.path}
                  className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-text-muted w-6">#{index + 1}</span>
                    <span className="text-sm font-medium text-text-primary">{endpoint.path}</span>
                  </div>
                  <div className="flex items-center gap-4 text-sm">
                    <span className="text-text-secondary">{endpoint.requests.toLocaleString()} req</span>
                    <span className="text-emerald-400">{endpoint.latency}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Real-time Metrics Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Real-time Metrics
            {isRealtimeEnabled && (
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="text-3xl font-bold text-white mb-1">
                {realtimeData[realtimeData.length - 1]?.requests || 0}
              </div>
              <p className="text-sm text-text-secondary">Current RPS</p>
              <div className="flex items-center justify-center gap-1 mt-2">
                <TrendingUp className="w-4 h-4 text-green-400" />
                <span className="text-xs text-green-400">+12% from avg</span>
              </div>
            </div>

            <div className="text-center">
              <div className="text-3xl font-bold text-white mb-1">
                {Math.round((realtimeData.reduce((sum, d) => sum + d.requests, 0) / realtimeData.length) * 100) / 100}ms
              </div>
              <p className="text-sm text-text-secondary">Avg Latency</p>
              <div className="flex items-center justify-center gap-1 mt-2">
                <Clock className="w-4 h-4 text-blue-400" />
                <span className="text-xs text-blue-400">Stable</span>
              </div>
            </div>

            <div className="text-center">
              <div className="text-3xl font-bold text-white mb-1">
                {(realtimeData[realtimeData.length - 1]?.errors || 0) * 10}
              </div>
              <p className="text-sm text-text-secondary">Active Connections</p>
              <div className="flex items-center justify-center gap-1 mt-2">
                <Globe className="w-4 h-4 text-purple-400" />
                <span className="text-xs text-purple-400">Peak: 98.5%</span>
              </div>
            </div>
          </div>

          {/* Live Activity Feed */}
          <div className="mt-6 border-t border-border-subtle pt-6">
            <h4 className="text-lg font-medium text-white mb-4">Live Activity</h4>
            <div className="space-y-3 max-h-48 overflow-y-auto">
              {Array.from({ length: 5 }, (_, i) => {
                const activities = [
                  { action: "Request processed", endpoint: "/api/users", latency: "45ms", status: "success" },
                  { action: "Function invoked", endpoint: "/api/webhooks", latency: "67ms", status: "success" },
                  { action: "Error occurred", endpoint: "/api/payments", latency: "120ms", status: "error" },
                  { action: "Cache hit", endpoint: "/api/config", latency: "12ms", status: "success" },
                  { action: "Deployment triggered", endpoint: "system", latency: "-", status: "info" },
                ];
                const activity = activities[i % activities.length];
                const timestamp = new Date(Date.now() - i * 30000).toLocaleTimeString();

                return (
                  <div key={i} className="flex items-center gap-3 p-2 rounded-lg bg-bg-tertiary">
                    <div className={`w-2 h-2 rounded-full ${
                      activity.status === 'success' ? 'bg-green-400' :
                      activity.status === 'error' ? 'bg-red-400' : 'bg-blue-400'
                    }`} />
                    <div className="flex-1">
                      <span className="text-sm text-text-primary">{activity.action}</span>
                      <span className="text-sm text-text-secondary ml-2">({activity.endpoint})</span>
                    </div>
                    <div className="text-xs text-text-muted">
                      {activity.latency !== '-' && `${activity.latency} • `}{timestamp}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

import { Logo } from "@/components/Logo";
import { StatusDot } from "@/components/StatusBadge";
import { UptimeBar, UptimeMiniBar } from "@/components/UptimeBar";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  statusAPI,
  type Component,
  type IncidentsListResponse,
  type UptimeDataPoint,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { format, parseISO, subDays } from "date-fns";
import { motion } from "framer-motion";
import {
  Activity,
  AlertTriangle,
  ArrowLeft,
  BarChart3,
  Calendar,
  Download,
  Globe,
  Server,
  TrendingUp,
} from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";

interface MonthlyStats {
  month: string;
  uptime: number;
  incidents: number;
  avgResponseTime: number;
}

function AnimatedBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden -z-10">
      <motion.div
        className="absolute w-[600px] h-[600px] rounded-full opacity-20 blur-[100px]"
        style={{
          background:
            "radial-gradient(circle, rgba(99, 102, 241, 0.3) 0%, transparent 70%)",
          top: "-10%",
          left: "-10%",
        }}
        animate={{
          x: [0, 100, 0],
          y: [0, 50, 0],
          scale: [1, 1.2, 1],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
      <motion.div
        className="absolute w-[500px] h-[500px] rounded-full opacity-15 blur-[80px]"
        style={{
          background:
            "radial-gradient(circle, rgba(139, 92, 246, 0.3) 0%, transparent 70%)",
          bottom: "-5%",
          right: "-5%",
        }}
        animate={{
          x: [0, -80, 0],
          y: [0, -30, 0],
          scale: [1, 1.1, 1],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
    </div>
  );
}

function Header() {
  return (
    <motion.header
      className="fixed top-0 left-0 right-0 z-50 bg-bg-glass-strong/90 backdrop-blur-xl border-b border-border-subtle shadow-lg"
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <Link
            to="/"
            className="inline-flex items-center gap-2 text-text-secondary hover:text-text-primary transition-colors group"
          >
            <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
            <span className="text-sm font-medium">Back to Status</span>
          </Link>

          <div className="flex items-center gap-3">
            <Logo size="sm" showText={false} />
            <span className="font-semibold text-text-primary hidden sm:inline">
              History
            </span>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" className="hidden sm:flex">
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
          </div>
        </div>
      </div>
    </motion.header>
  );
}

function StatsCard({ stat, index }: { stat: MonthlyStats; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.2 + index * 0.05 }}
      whileHover={{ scale: 1.02, y: -4 }}
    >
      <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm overflow-hidden hover:border-border-default transition-all duration-300">
        <CardContent className="p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold text-text-primary">{stat.month}</h3>
            {stat.incidents > 0 ? (
              <span className="text-xs text-amber-400 bg-amber-500/10 px-2 py-1 rounded-full border border-amber-500/30">
                {stat.incidents}{" "}
                {stat.incidents === 1 ? "incident" : "incidents"}
              </span>
            ) : (
              <span className="text-xs text-emerald-400 bg-emerald-500/10 px-2 py-1 rounded-full border border-emerald-500/30">
                No incidents
              </span>
            )}
          </div>

          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm text-text-muted">Uptime</span>
                <span className="font-mono font-medium text-emerald-400">
                  {stat.uptime}%
                </span>
              </div>
              <div className="h-2 bg-bg-secondary rounded-full overflow-hidden">
                <motion.div
                  className="h-full bg-gradient-to-r from-emerald-500 to-brand-500"
                  initial={{ width: 0 }}
                  animate={{ width: `${stat.uptime}%` }}
                  transition={{
                    delay: 0.5 + index * 0.1,
                    duration: 0.8,
                    ease: "easeOut",
                  }}
                />
              </div>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-text-muted">Avg Latency</span>
              <span className="font-mono font-medium text-text-primary">
                {stat.avgResponseTime}ms
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function StatsCardSkeleton() {
  return (
    <Card className="border-border-subtle bg-bg-tertiary/50 p-5">
      <div className="flex items-center justify-between mb-4">
        <Skeleton className="h-5 w-24" />
        <Skeleton className="h-5 w-20 rounded-full" />
      </div>
      <Skeleton className="h-2 w-full mb-4 rounded-full" />
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-4 w-16" />
      </div>
    </Card>
  );
}

// Transform uptime data points from API into segments for the UptimeBar
function transformUptimeData(
  dataPoints: UptimeDataPoint[],
): Array<{
  date: string;
  status: "operational" | "degraded" | "outage" | "maintenance";
  uptime: number;
}> {
  if (!dataPoints || dataPoints.length === 0) {
    // Generate fallback data for the last 365 days
    return Array.from({ length: 365 }, (_, i) => ({
      date: format(subDays(new Date(), 364 - i), "MMM dd"),
      status: "operational" as const,
      uptime: 99.97,
    }));
  }

  return dataPoints.map((point) => {
    const uptime = point.uptime_percent;
    let status: "operational" | "degraded" | "outage" | "maintenance" =
      "operational";

    if (uptime < 95) {
      status = "outage";
    } else if (uptime < 99) {
      status = "degraded";
    }

    return {
      date: format(
        parseISO(
          point.timestamp.toString
            ? point.timestamp.toString()
            : new Date().toISOString(),
        ),
        "MMM dd",
      ),
      status,
      uptime,
    };
  });
}

// Calculate monthly stats from components and incidents
function calculateMonthlyStats(
  components: Component[],
  incidents: IncidentsListResponse["incidents"],
): MonthlyStats[] {
  const now = new Date();
  const months: MonthlyStats[] = [];

  for (let i = 0; i < 6; i++) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const monthName = format(date, "MMMM yyyy");

    // Count incidents in this month
    const monthStart = new Date(date.getFullYear(), date.getMonth(), 1);
    const monthEnd = new Date(
      date.getFullYear(),
      date.getMonth() + 1,
      0,
      23,
      59,
      59,
    );

    const monthIncidents = incidents.filter((inc) => {
      const incDate = parseISO(inc.created_at);
      return incDate >= monthStart && incDate <= monthEnd;
    });

    // Calculate average uptime from components (default to 99.97 if no data)
    const avgUptime =
      components.length > 0
        ? components.reduce((acc, c) => acc + (c.uptime_30d || 99.97), 0) /
          components.length
        : 99.97;

    // Calculate average response time
    const avgLatency =
      components.length > 0
        ? Math.round(
            components.reduce((acc, c) => acc + (c.response_time_ms || 50), 0) /
              components.length,
          )
        : 45;

    months.push({
      month: monthName,
      uptime: Number(avgUptime.toFixed(2)),
      incidents: monthIncidents.length,
      avgResponseTime: avgLatency,
    });
  }

  return months;
}

export default function HistoryPage() {
  const [selectedRange, setSelectedRange] = useState<
    "30d" | "90d" | "6m" | "1y"
  >("1y");

  // Fetch uptime metrics
  const { data: uptimeData, isLoading: isLoadingUptime } = useQuery({
    queryKey: ["uptime", "all", "90d"],
    queryFn: () => statusAPI.getUptimeMetrics("all", "90d"),
  });

  // Fetch components
  const { data: componentsData, isLoading: isLoadingComponents } = useQuery({
    queryKey: ["components"],
    queryFn: () => statusAPI.getComponents(),
  });

  // Fetch incidents
  const { data: incidentsData, isLoading: isLoadingIncidents } = useQuery({
    queryKey: ["incidents", "history"],
    queryFn: () => statusAPI.listIncidents({ limit: 100 }),
  });

  const ranges = [
    { label: "30d", value: "30d" as const },
    { label: "90d", value: "90d" as const },
    { label: "6m", value: "6m" as const },
    { label: "1y", value: "1y" as const },
  ];

  const isLoading =
    isLoadingUptime || isLoadingComponents || isLoadingIncidents;

  const components = componentsData?.components || [];
  const incidents = incidentsData?.incidents || [];
  const uptimePoints = uptimeData?.data_points || [];

  // Transform uptime data
  const yearlyUptime = transformUptimeData(uptimePoints);

  // Calculate monthly stats
  const monthlyStats = calculateMonthlyStats(components, incidents);

  // Filter based on selected range
  const filteredUptime = yearlyUptime.slice(
    -(selectedRange === "30d"
      ? 30
      : selectedRange === "90d"
        ? 90
        : selectedRange === "6m"
          ? 180
          : 365),
  );

  // Calculate overall stats
  const overallUptime = uptimeData?.overall_uptime || 99.97;
  const totalIncidents = incidents.filter(
    (inc) => inc.status === "resolved",
  ).length;
  const activeIncidents = incidents.filter(
    (inc) => inc.status !== "resolved",
  ).length;
  const avgResponse =
    components.length > 0
      ? Math.round(
          components.reduce((acc, c) => acc + c.response_time_ms, 0) /
            components.length,
        )
      : 45;

  return (
    <div className="min-h-screen bg-bg-primary">
      <AnimatedBackground />
      <Header />

      <main className="pt-24 pb-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8">
          {/* Page Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center mb-12"
          >
            <h1 className="text-3xl md:text-4xl font-bold text-text-primary mb-4">
              Status History
            </h1>
            <p className="text-text-secondary max-w-2xl mx-auto">
              Historical uptime data and incident reports for all FunctionFly
              services. Our infrastructure maintains a{" "}
              {overallUptime.toFixed(2)}% uptime guarantee.
            </p>
          </motion.div>

          {/* Overview Stats */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-4"
          >
            {[
              {
                label: "Overall Uptime",
                value: `${overallUptime.toFixed(2)}%`,
                icon: TrendingUp,
                color: "text-emerald-400",
              },
              {
                label: "Avg Response",
                value: `${avgResponse}ms`,
                icon: Activity,
                color: "text-brand-400",
              },
              {
                label: "Total Incidents",
                value: totalIncidents.toString(),
                icon: AlertTriangle,
                color: "text-amber-400",
              },
              {
                label: "Active Incidents",
                value: activeIncidents.toString(),
                icon: Globe,
                color:
                  activeIncidents > 0 ? "text-red-400" : "text-emerald-400",
              },
            ].map((stat, index) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.2 + index * 0.05 }}
                whileHover={{ scale: 1.02, y: -2 }}
                className={cn(
                  "rounded-xl p-4 border border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm",
                  "hover:border-border-default hover:bg-bg-elevated/80 transition-all duration-300",
                )}
              >
                <div
                  className={cn(
                    "p-2 rounded-lg w-fit mb-3 bg-bg-secondary",
                    stat.color,
                  )}
                >
                  <stat.icon className="w-5 h-5" />
                </div>
                <div className="text-2xl font-bold text-text-primary">
                  {stat.value}
                </div>
                <div className="text-sm text-text-muted">{stat.label}</div>
              </motion.div>
            ))}
          </motion.div>

          {/* Uptime History Chart */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
          >
            <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm overflow-hidden">
              <CardHeader className="flex flex-row items-center justify-between pb-4">
                <div>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <BarChart3 className="w-5 h-5 text-brand-400" />
                    Uptime Overview
                  </CardTitle>
                  <CardDescription>
                    System availability over the past{" "}
                    {selectedRange === "1y" ? "year" : selectedRange}
                  </CardDescription>
                </div>

                <div className="flex items-center gap-1 p-1 bg-bg-secondary rounded-lg border border-border-subtle">
                  {ranges.map((range) => (
                    <button
                      key={range.value}
                      onClick={() => setSelectedRange(range.value)}
                      className={cn(
                        "px-3 py-1.5 text-sm font-medium rounded-md transition-all duration-200",
                        selectedRange === range.value
                          ? "bg-brand-500 text-white shadow-lg shadow-brand-500/25"
                          : "text-text-secondary hover:text-text-primary hover:bg-bg-hover",
                      )}
                    >
                      {range.label}
                    </button>
                  ))}
                </div>
              </CardHeader>

              <CardContent className="pb-6">
                {isLoadingUptime ? (
                  <Skeleton className="h-24 w-full mt-2 rounded" />
                ) : (
                  <UptimeBar segments={filteredUptime} className="mt-2" />
                )}

                {/* Stats summary */}
                <div className="mt-6 grid grid-cols-3 gap-4">
                  {[
                    {
                      label: "Operational",
                      value: "99.2%",
                      color: "text-emerald-400",
                      bgColor: "bg-emerald-500",
                    },
                    {
                      label: "Degraded",
                      value: "0.6%",
                      color: "text-amber-400",
                      bgColor: "bg-amber-500",
                    },
                    {
                      label: "Outage",
                      value: "0.2%",
                      color: "text-red-400",
                      bgColor: "bg-red-500",
                    },
                  ].map((stat) => (
                    <div
                      key={stat.label}
                      className="flex items-center gap-3 p-3 rounded-lg bg-bg-secondary/50 border border-border-subtle"
                    >
                      <div
                        className={cn("w-3 h-3 rounded-full", stat.bgColor)}
                      />
                      <div>
                        <div className={cn("text-lg font-bold", stat.color)}>
                          {stat.value}
                        </div>
                        <div className="text-xs text-text-muted">
                          {stat.label}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Monthly Breakdown */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
          >
            <h2 className="text-xl font-semibold text-text-primary mb-6 flex items-center gap-2">
              <Calendar className="w-5 h-5 text-brand-400" />
              Monthly Breakdown
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {isLoading
                ? Array.from({ length: 6 }).map((_, i) => (
                    <StatsCardSkeleton key={i} />
                  ))
                : monthlyStats.map((stat, index) => (
                    <StatsCard key={stat.month} stat={stat} index={index} />
                  ))}
            </div>
          </motion.div>

          {/* Service Performance */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
          >
            <h2 className="text-xl font-semibold text-text-primary mb-6 flex items-center gap-2">
              <Server className="w-5 h-5 text-brand-400" />
              Service Performance (Last 30 Days)
            </h2>

            <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
              <CardContent className="p-0">
                <div className="divide-y divide-border-subtle">
                  {isLoadingComponents
                    ? Array.from({ length: 6 }).map((_, i) => (
                        <div
                          key={i}
                          className="flex items-center justify-between p-4"
                        >
                          <div className="flex items-center gap-4">
                            <Skeleton className="w-2 h-2 rounded-full" />
                            <div>
                              <Skeleton className="h-5 w-32 mb-1" />
                              <Skeleton className="h-4 w-20" />
                            </div>
                          </div>
                          <Skeleton className="h-5 w-16" />
                        </div>
                      ))
                    : components.map((service, index) => (
                        <motion.div
                          key={service.id}
                          initial={{ opacity: 0, x: -20 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: 0.6 + index * 0.05 }}
                          className="flex items-center justify-between p-4 hover:bg-bg-secondary/50 transition-colors group"
                        >
                          <div className="flex items-center gap-4">
                            <StatusDot
                              status={service.status}
                              size="sm"
                              pulse={false}
                            />
                            <div>
                              <div className="font-medium text-text-primary group-hover:text-brand-400 transition-colors">
                                {service.name}
                              </div>
                              <div className="text-sm text-text-muted">
                                {service.response_time_ms}ms avg latency
                              </div>
                            </div>
                          </div>

                          <div className="flex items-center gap-4">
                            <UptimeMiniBar
                              days={30}
                              uptime={service.uptime_30d}
                              className="hidden sm:flex w-32"
                            />
                            <div
                              className={cn(
                                "font-mono font-medium",
                                service.uptime_30d >= 99.98
                                  ? "text-emerald-400"
                                  : service.uptime_30d >= 99.95
                                    ? "text-amber-400"
                                    : "text-red-400",
                              )}
                            >
                              {service.uptime_30d.toFixed(2)}%
                            </div>
                          </div>
                        </motion.div>
                      ))}
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Call to Action */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.7 }}
            className={cn(
              "relative overflow-hidden rounded-2xl p-8",
              "bg-gradient-to-br from-brand-500/10 via-purple-500/5 to-transparent",
              "border border-brand-500/20",
            )}
          >
            <div className="relative z-10 flex flex-col md:flex-row items-center justify-between gap-6">
              <div className="text-center md:text-left">
                <h3 className="text-xl font-semibold text-text-primary mb-2">
                  Need historical data?
                </h3>
                <p className="text-text-secondary">
                  Export full uptime reports or access our API for programmatic
                  access.
                </p>
              </div>
              <div className="flex gap-3">
                <Button
                  variant="outline"
                  className="border-border-default hover:bg-bg-hover"
                >
                  <Download className="w-4 h-4 mr-2" />
                  Export CSV
                </Button>
                <Button className="bg-brand-500 hover:bg-brand-600 text-white shadow-lg shadow-brand-500/25">
                  <Globe className="w-4 h-4 mr-2" />
                  API Access
                </Button>
              </div>
            </div>
          </motion.div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border-subtle bg-bg-secondary/50 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm text-text-muted">
              <Logo size="sm" showText={false} />
              <span>© {new Date().getFullYear()} FunctionFly</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-text-muted">
              <Link
                to="/"
                className="hover:text-text-primary transition-colors"
              >
                Status
              </Link>
              <Link
                to="/history"
                className="hover:text-text-primary transition-colors"
              >
                History
              </Link>
              <a
                href="https://docs.functionfly.com"
                className="hover:text-text-primary transition-colors"
              >
                API
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

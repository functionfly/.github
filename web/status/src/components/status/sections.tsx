import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { UptimeBar } from "@/components/UptimeBar";
import { getStatusColor, statusAPI } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { format, formatDistanceToNow, parseISO } from "date-fns";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  CheckCircle,
  ChevronDown,
  Clock,
  Gauge,
  Globe,
  Layers,
  TrendingDown,
  TrendingUp,
  Wifi,
  Wrench,
  Zap,
} from "lucide-react";
import { useMemo, useState } from "react";
import { ProviderCard } from "./cards";
import {
  IncidentSkeleton,
  MetricsSectionSkeleton,
  ServiceCardSkeleton,
} from "./skeletons";
import type {
  IncidentTimelineProps,
  IncidentUpdate,
  MaintenanceSectionProps,
  Metric,
  MetricsSectionProps,
  ProviderSectionProps,
  UptimeDataPoint,
  UptimeHistorySectionProps,
} from "./types";

export function MetricsSection({
  components,
  isLoading,
  probeLatencyMs,
  probeLatencyLoading,
}: MetricsSectionProps) {
  if (isLoading) {
    return <MetricsSectionSkeleton />;
  }

  const operationalCount = components.filter(
    (c) => c.status === "operational",
  ).length;
  const avgResponse =
    components.length > 0
      ? Math.round(
          components.reduce((acc, c) => acc + c.response_time_ms, 0) /
            components.length,
        )
      : 0;
  const avgUptime30 =
    components.length > 0
      ? components.reduce((acc, c) => acc + c.uptime_30d, 0) / components.length
      : 0;
  const avgUptime7 =
    components.length > 0
      ? components.reduce((acc, c) => acc + c.uptime_7d, 0) / components.length
      : 0;
  const uptimeWeekVsMonth = avgUptime7 - avgUptime30;
  const formatSignedPercent = (x: number) => {
    if (!Number.isFinite(x) || Math.abs(x) < 0.005) return "—";
    const sign = x > 0 ? "+" : "";
    return `${sign}${x.toFixed(2)}%`;
  };
  const healthPct =
    components.length > 0 ? (100 * operationalCount) / components.length : 0;
  const notOperational = components.length - operationalCount;

  const metrics: Metric[] = [
    {
      label: "Uptime (30d)",
      value: `${avgUptime30.toFixed(2)}%`,
      change: `7d vs 30d avg ${formatSignedPercent(uptimeWeekVsMonth)}`,
      trend: uptimeWeekVsMonth >= 0 ? "up" : "down",
      icon: <Wifi className="w-5 h-5" />,
    },
    {
      label: "Avg Response",
      value: `${avgResponse}ms`,
      ...(avgResponse > 0
        ? {
            change: "Latest system_health_checks",
            trend: "neutral" as const,
          }
        : {
            change: "No health-check samples yet",
            trend: "neutral" as const,
          }),
      icon: <Zap className="w-5 h-5" />,
    },
    {
      label: "Probe latency",
      value: probeLatencyLoading
        ? "…"
        : probeLatencyMs != null && probeLatencyMs > 0
          ? `${Math.round(probeLatencyMs)}ms`
          : "—",
      change: probeLatencyLoading
        ? "Loading Prometheus…"
        : probeLatencyMs != null && probeLatencyMs > 0
          ? "p95 · all providers · 24h"
          : "No probe series yet (functionfly_probe_latency_ms_bucket)",
      trend: "neutral" as const,
      icon: <Gauge className="w-5 h-5" />,
    },
    {
      label: "Operational",
      value: `${healthPct.toFixed(1)}%`,
      change:
        notOperational > 0
          ? `${notOperational} component${notOperational === 1 ? "" : "s"} not operational`
          : "All monitored components operational",
      trend: notOperational === 0 ? "up" : "down",
      icon: <TrendingUp className="w-5 h-5" />,
    },
    {
      label: "Components",
      value: `${operationalCount}/${components.length}`,
      change: "Tracked in status catalog",
      trend: "neutral",
      icon: <Globe className="w-5 h-5" />,
    },
  ];

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 0.3 }}
      className="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-5 gap-4"
    >
      {metrics.map((metric, index) => (
        <motion.div
          key={metric.label}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 + index * 0.1 }}
          whileHover={{ scale: 1.02, y: -2 }}
          className={cn(
            "relative overflow-hidden rounded-xl p-4",
            "bg-gradient-to-br from-bg-tertiary/80 to-bg-elevated/50",
            "border border-border-subtle backdrop-blur-sm",
            "hover:border-border-default transition-all duration-300",
            "group cursor-pointer",
          )}
        >
          <div className="absolute inset-0 bg-gradient-to-br from-brand-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

          <div className="relative z-10 flex items-start justify-between mb-3">
            <div
              className={cn(
                "p-2 rounded-lg",
                "bg-bg-secondary border border-border-subtle",
                "text-text-secondary group-hover:text-brand-400 transition-colors",
              )}
            >
              {metric.icon}
            </div>
            {metric.change ? (
              <div
                className={cn(
                  "flex items-center gap-1 text-xs font-medium max-w-[140px] text-right leading-tight",
                  metric.trend === "up" && "text-emerald-400",
                  metric.trend === "down" && "text-amber-400",
                  metric.trend === "neutral" && "text-text-muted",
                )}
              >
                {metric.trend === "down" ? (
                  <TrendingDown className="w-3 h-3 shrink-0" />
                ) : metric.trend === "up" ? (
                  <TrendingUp className="w-3 h-3 shrink-0" />
                ) : null}
                <span className="line-clamp-2">{metric.change}</span>
              </div>
            ) : null}
          </div>

          <div className="relative z-10">
            <div className="text-2xl font-bold text-text-primary mb-1">
              {metric.value}
            </div>
            <div className="text-xs text-text-muted">{metric.label}</div>
          </div>
        </motion.div>
      ))}
    </motion.div>
  );
}

export function IncidentTimeline({
  incidents,
  isLoading,
}: IncidentTimelineProps) {
  const [expandedIncident, setExpandedIncident] = useState<string | null>(null);

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2].map((i) => (
          <IncidentSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (incidents.length === 0) {
    return (
      <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
        <CardContent className="p-8 text-center">
          <CheckCircle className="w-12 h-12 text-emerald-500 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-semibold text-text-primary mb-2">
            No Recent Incidents
          </h3>
          <p className="text-text-secondary">
            All systems have been operating normally. Great job team!
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {incidents.map((incident, index) => (
        <motion.div
          key={incident.id}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: index * 0.1 }}
        >
          <Card
            className={cn(
              "border border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm",
              "hover:border-border-default transition-all duration-300",
              "overflow-hidden",
            )}
          >
            <div
              className="p-4 cursor-pointer"
              onClick={() =>
                setExpandedIncident(
                  expandedIncident === incident.id ? null : incident.id,
                )
              }
            >
              <div className="flex items-start gap-4">
                <div
                  className={cn(
                    "mt-1 w-3 h-3 rounded-full shrink-0",
                    incident.severity === "critical"
                      ? "bg-red-500 shadow-lg shadow-red-500/50"
                      : incident.severity === "high"
                        ? "bg-amber-500 shadow-lg shadow-amber-500/50"
                        : "bg-emerald-500 shadow-lg shadow-emerald-500/50",
                  )}
                />

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <h3 className="font-semibold text-text-primary">
                      {incident.title}
                    </h3>
                    <span
                      className={cn(
                        "text-xs px-2 py-0.5 rounded-full border",
                        getStatusColor(incident.status),
                      )}
                    >
                      {incident.status}
                    </span>
                  </div>

                  <p className="text-sm text-text-secondary mb-2">
                    {incident.description}
                  </p>

                  <div className="flex items-center gap-4 text-xs text-text-muted">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {formatDistanceToNow(parseISO(incident.created_at), {
                        addSuffix: true,
                      })}
                    </span>
                    <span className="flex items-center gap-1">
                      <Layers className="w-3 h-3" />
                      {incident.affected_components?.length || 0} services
                      affected
                    </span>
                  </div>
                </div>

                <ChevronDown
                  className={cn(
                    "w-5 h-5 text-text-muted transition-transform duration-300",
                    expandedIncident === incident.id && "rotate-180",
                  )}
                />
              </div>
            </div>

            <AnimatePresence>
              {expandedIncident === incident.id && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                  className="border-t border-border-subtle bg-bg-secondary/50"
                >
                  <div className="p-4 space-y-4">
                    <h4 className="text-sm font-medium text-text-primary">
                      Incident Updates
                    </h4>
                    <div className="space-y-3">
                      {incident.updates?.map(
                        (update: IncidentUpdate, i: number) => (
                          <div key={update.id} className="flex gap-3">
                            <div className="flex flex-col items-center">
                              <div
                                className={cn(
                                  "w-2 h-2 rounded-full",
                                  update.status === "resolved"
                                    ? "bg-emerald-500"
                                    : update.status === "identified"
                                      ? "bg-blue-500"
                                      : "bg-amber-500",
                                )}
                              />
                              {i < (incident.updates?.length || 0) - 1 && (
                                <div className="w-px h-full bg-border-subtle my-1" />
                              )}
                            </div>
                            <div className="flex-1 pb-4">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="text-xs font-medium text-text-primary">
                                  {update.status.charAt(0).toUpperCase() +
                                    update.status.slice(1)}
                                </span>
                                <span className="text-xs text-text-muted">
                                  {formatDistanceToNow(
                                    parseISO(update.created_at),
                                    { addSuffix: true },
                                  )}
                                </span>
                              </div>
                              <p className="text-sm text-text-secondary">
                                {update.message}
                              </p>
                            </div>
                          </div>
                        ),
                      )}
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </Card>
        </motion.div>
      ))}
    </div>
  );
}

export function MaintenanceSection({
  maintenance,
  isLoading,
}: MaintenanceSectionProps) {
  if (isLoading) {
    return (
      <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-20 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (maintenance.length === 0) {
    return (
      <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Wrench className="w-5 h-5 text-purple-400" />
            Scheduled Maintenance
          </CardTitle>
          <CardDescription>
            Upcoming maintenance windows and scheduled downtime
          </CardDescription>
        </CardHeader>
        <CardContent className="p-8 text-center">
          <CheckCircle className="w-12 h-12 text-emerald-500 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-semibold text-text-primary mb-2">
            No Scheduled Maintenance
          </h3>
          <p className="text-text-secondary">
            There are no upcoming maintenance windows scheduled at this time.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
      <CardHeader>
        <CardTitle className="text-lg flex items-center gap-2">
          <Wrench className="w-5 h-5 text-purple-400" />
          Scheduled Maintenance
        </CardTitle>
        <CardDescription>
          Upcoming maintenance windows and scheduled downtime
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {maintenance.map((item, index) => (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className={cn(
                "flex items-start gap-4 p-4 rounded-lg border",
                "bg-bg-secondary/50 border-border-subtle",
                "hover:border-purple-500/30 transition-colors",
              )}
            >
              <div className="p-2 rounded-lg bg-purple-500/10 border border-purple-500/20">
                <Wrench className="w-4 h-4 text-purple-400" />
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-medium text-text-primary">{item.title}</h4>
                <div className="flex items-center gap-3 text-sm text-text-muted mt-1">
                  <span>
                    {formatDistanceToNow(parseISO(item.scheduled_start), {
                      addSuffix: true,
                    })}
                  </span>
                  <span>•</span>
                  <span>
                    {format(parseISO(item.scheduled_start), "MMM d, HH:mm")}
                  </span>
                </div>
              </div>
              <span
                className={cn(
                  "px-2 py-1 rounded-full text-xs font-medium border",
                  item.status === "scheduled"
                    ? "text-purple-400 bg-purple-500/10 border-purple-500/30"
                    : item.status === "in_progress"
                      ? "text-amber-400 bg-amber-500/10 border-amber-500/30"
                      : "text-emerald-400 bg-emerald-500/10 border-emerald-500/30",
                )}
              >
                {item.status}
              </span>
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function uptimePointsToBarSegments(dataPoints: UptimeDataPoint[]): Array<{
  date: string;
  status: "operational" | "degraded" | "outage" | "maintenance";
  uptime: number;
}> {
  if (!dataPoints.length) return [];
  return dataPoints.map((point) => {
    const uptime = point.uptime_percent;
    let status: "operational" | "degraded" | "outage" | "maintenance" =
      "operational";
    if (uptime < 95) status = "outage";
    else if (uptime < 99) status = "degraded";
    const raw = point.timestamp as unknown;
    const ts =
      typeof raw === "string"
        ? raw
        : raw != null &&
            typeof (raw as { toString?: () => string }).toString === "function"
          ? (raw as { toString: () => string }).toString()
          : new Date().toISOString();
    return {
      date: format(parseISO(ts), "MMM d"),
      status,
      uptime,
    };
  });
}

export function UptimeHistorySection({ isLoading }: UptimeHistorySectionProps) {
  const [selectedRange, setSelectedRange] = useState(30);

  const ranges = [
    { label: "7d", value: 7 },
    { label: "30d", value: 30 },
    { label: "90d", value: 90 },
  ];

  const period =
    selectedRange === 7 ? "7d" : selectedRange === 30 ? "30d" : "90d";

  const { data: uptimeMetrics, isLoading: isLoadingUptime } = useQuery({
    queryKey: ["uptime", "all", period, "day"],
    queryFn: () => statusAPI.getUptimeMetrics("all", period, "day"),
    staleTime: 60_000,
  });

  const segments = useMemo(
    () => uptimePointsToBarSegments(uptimeMetrics?.data_points ?? []),
    [uptimeMetrics?.data_points],
  );

  const legendStats = useMemo(() => {
    const total = segments.length;
    const share = (n: number) =>
      total > 0 ? `${((100 * n) / total).toFixed(1)}%` : "—";
    const op = segments.filter((s) => s.status === "operational").length;
    const deg = segments.filter((s) => s.status === "degraded").length;
    const out = segments.filter((s) => s.status === "outage").length;
    const maint = segments.filter((s) => s.status === "maintenance").length;
    return [
      {
        label: "Operational",
        value: share(op),
        color: "text-emerald-400",
        bgColor: "bg-emerald-500",
      },
      {
        label: "Degraded",
        value: share(deg),
        color: "text-amber-400",
        bgColor: "bg-amber-500",
      },
      {
        label: "Outage",
        value: share(out),
        color: "text-red-400",
        bgColor: "bg-red-500",
      },
      ...(maint > 0
        ? [
            {
              label: "Maintenance",
              value: share(maint),
              color: "text-purple-400",
              bgColor: "bg-purple-500",
            },
          ]
        : []),
    ];
  }, [segments]);

  const showSkeleton = isLoading || isLoadingUptime;

  return (
    <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle className="text-lg flex items-center gap-2">
            <Activity className="w-5 h-5 text-brand-400" />
            Uptime History
          </CardTitle>
          <CardDescription>
            System availability over the past {selectedRange} days
          </CardDescription>
        </div>

        <div className="flex items-center gap-1 p-1 bg-bg-secondary rounded-lg border border-border-subtle">
          {ranges.map((range) => (
            <button
              key={range.value}
              type="button"
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
        {showSkeleton ? (
          <Skeleton className="h-24 w-full mt-2" />
        ) : segments.length === 0 ? (
          <p className="text-sm text-text-muted mt-2 py-6">
            No uptime samples for this range yet. The chart fills when
            Prometheus records{" "}
            <span className="font-mono text-xs text-text-secondary">
              functionfly_uptime_ratio
            </span>{" "}
            (see ops / monitoring docs).
          </p>
        ) : (
          <UptimeBar segments={segments} className="mt-2" />
        )}

        <div
          className={cn(
            "mt-6 gap-4",
            legendStats.length === 4
              ? "grid grid-cols-2 sm:grid-cols-4"
              : "grid grid-cols-3",
          )}
        >
          {legendStats.map((stat) => (
            <div
              key={stat.label}
              className="flex items-center gap-3 p-3 rounded-lg bg-bg-secondary/50 border border-border-subtle"
            >
              <div
                className={cn("w-3 h-3 rounded-full shrink-0", stat.bgColor)}
              />
              <div className="min-w-0">
                <div className={cn("text-lg font-bold", stat.color)}>
                  {stat.value}
                </div>
                <div className="text-xs text-text-muted">{stat.label}</div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

// Import Skeleton here to avoid circular dependency
import { Skeleton } from "@/components/ui/skeleton";

export function ProviderSection({
  providers,
  isLoading,
}: ProviderSectionProps) {
  if (isLoading) {
    return (
      <section>
        <div className="flex items-center justify-between mb-6">
          <Skeleton className="h-7 w-40" />
          <Skeleton className="h-5 w-32" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <ServiceCardSkeleton key={i} />
          ))}
        </div>
      </section>
    );
  }

  return (
    <section>
      <motion.div
        className="flex items-center justify-between mb-6"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.4 }}
      >
        <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
          <Layers className="w-5 h-5 text-brand-400" />
          Infrastructure Providers
        </h2>
        <span className="text-sm text-text-muted">
          {providers.filter((p) => p.status === "operational").length} /{" "}
          {providers.length} operational
        </span>
      </motion.div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {providers.map((provider, index) => (
          <ProviderCard key={provider.id} provider={provider} index={index} />
        ))}
      </div>
    </section>
  );
}

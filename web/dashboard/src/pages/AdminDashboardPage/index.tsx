import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Users,
  Building2,
  CreditCard,
  Shield,
  Settings,
  Mail,
  Calendar,
  FileText,
  MessageSquare,
  ArrowRight,
  Code,
  Database,
  Layers,
  TrendingUp,
  TrendingDown,
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  RefreshCw,
  BarChart3,
  Zap,
  Globe,
  ClipboardList,
} from "lucide-react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import { adminUsersApi, billingApi, auditApi, adminFunctionsApi, healthApi, adminDashboardApi } from "@/api/admin";
import { cn } from "@/lib/utils";

const DAY_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

const adminSections = [
  {
    title: "Tenants",
    description: "Manage multi-tenant organizations",
    path: "/admin/tenants",
    icon: Building2,
    iconColor: "text-blue-500",
    bgColor: "bg-blue-500/10",
    borderColor: "border-blue-500/20 hover:border-blue-500/40",
  },
  {
    title: "Users",
    description: "User management and permissions",
    path: "/admin/users",
    icon: Users,
    iconColor: "text-emerald-500",
    bgColor: "bg-emerald-500/10",
    borderColor: "border-emerald-500/20 hover:border-emerald-500/40",
  },
  {
    title: "Billing",
    description: "Subscriptions and revenue",
    path: "/admin/billing",
    icon: CreditCard,
    iconColor: "text-purple-500",
    bgColor: "bg-purple-500/10",
    borderColor: "border-purple-500/20 hover:border-purple-500/40",
  },
  {
    title: "Features",
    description: "Tier-specific features and permissions",
    path: "/admin/features",
    icon: Shield,
    iconColor: "text-amber-500",
    bgColor: "bg-amber-500/10",
    borderColor: "border-amber-500/20 hover:border-amber-500/40",
  },
  {
    title: "Audit Log",
    description: "Security and system events",
    path: "/admin/audit",
    icon: Shield,
    iconColor: "text-red-500",
    bgColor: "bg-red-500/10",
    borderColor: "border-red-500/20 hover:border-red-500/40",
  },
  {
    title: "System",
    description: "Configuration and maintenance",
    path: "/admin/system",
    icon: Settings,
    iconColor: "text-slate-500",
    bgColor: "bg-slate-500/10",
    borderColor: "border-slate-500/20 hover:border-slate-500/40",
  },
  {
    title: "Newsletter",
    description: "Subscribers and campaigns",
    path: "/admin/newsletter",
    icon: Mail,
    iconColor: "text-indigo-500",
    bgColor: "bg-indigo-500/10",
    borderColor: "border-indigo-500/20 hover:border-indigo-500/40",
  },
  {
    title: "Content Calendar",
    description: "Publication schedule",
    path: "/admin/content-calendar",
    icon: Calendar,
    iconColor: "text-orange-500",
    bgColor: "bg-orange-500/10",
    borderColor: "border-orange-500/20 hover:border-orange-500/40",
  },
  {
    title: "Content",
    description: "Blog and content management",
    path: "/admin/content",
    icon: FileText,
    iconColor: "text-teal-500",
    bgColor: "bg-teal-500/10",
    borderColor: "border-teal-500/20 hover:border-teal-500/40",
  },
  {
    title: "Feedback",
    description: "User feedback and tickets",
    path: "/admin/feedback",
    icon: MessageSquare,
    iconColor: "text-pink-500",
    bgColor: "bg-pink-500/10",
    borderColor: "border-pink-500/20 hover:border-pink-500/40",
  },
  {
    title: "Functions",
    description: "All functions across tenants",
    path: "/admin/functions",
    icon: Code,
    iconColor: "text-violet-500",
    bgColor: "bg-violet-500/10",
    borderColor: "border-violet-500/20 hover:border-violet-500/40",
  },
  {
    title: "Registry",
    description: "Function registry moderation",
    path: "/admin/registry",
    icon: Database,
    iconColor: "text-cyan-500",
    bgColor: "bg-cyan-500/10",
    borderColor: "border-cyan-500/20 hover:border-cyan-500/40",
  },
  {
    title: "State Fabric",
    description: "State fabrics across tenants",
    path: "/admin/state-fabric",
    icon: Layers,
    iconColor: "text-amber-500",
    bgColor: "bg-amber-500/10",
    borderColor: "border-amber-500/20 hover:border-amber-500/40",
  },
  {
    title: "Trust Dashboard",
    description: "Monitor trust distribution and detect suspicious trust activity",
    path: "/admin/trust-dashboard",
    icon: Shield,
    iconColor: "text-emerald-500",
    bgColor: "bg-emerald-500/10",
    borderColor: "border-emerald-500/20 hover:border-emerald-500/40",
  },
  {
    title: "Execution Audit",
    description: "Searchable audit trail for all function executions",
    path: "/admin/execution-audit",
    icon: ClipboardList,
    iconColor: "text-blue-500",
    bgColor: "bg-blue-500/10",
    borderColor: "border-blue-500/20 hover:border-blue-500/40",
  },
  {
    title: "Fraud Detection",
    description: "AI-powered fraud and bot pattern detection",
    path: "/admin/fraud-detection",
    icon: AlertTriangle,
    iconColor: "text-rose-500",
    bgColor: "bg-rose-500/10",
    borderColor: "border-rose-500/20 hover:border-rose-500/40",
  },
  {
    title: "Economic Leaderboard",
    description: "Revenue leaders and economic manipulation detection",
    path: "/admin/economic-leaderboard",
    icon: TrendingUp,
    iconColor: "text-violet-500",
    bgColor: "bg-violet-500/10",
    borderColor: "border-violet-500/20 hover:border-violet-500/40",
  },
];

interface SystemHealthItem {
  name: string;
  status: "healthy" | "degraded" | "down";
  latency?: string;
  uptime?: string;
}

function HealthStatusDot({ status }: { status: SystemHealthItem["status"] }) {
  return (
    <span
      className={cn(
        "inline-block w-2 h-2 rounded-full",
        status === "healthy" && "bg-emerald-500",
        status === "degraded" && "bg-amber-500",
        status === "down" && "bg-red-500"
      )}
    />
  );
}

export function AdminDashboardPage() {
  const navigate = useNavigate();

  // Fetch real data
  const { data: userStats, isLoading: usersLoading } = useQuery({
    queryKey: ["admin-user-stats"],
    queryFn: () => adminUsersApi.getUserStats(),
  });

  const { data: subscriptionsData, isLoading: subsLoading } = useQuery({
    queryKey: ["subscriptions"],
    queryFn: () => billingApi.listSubscriptions(),
  });

  const { data: auditData, isLoading: auditLoading } = useQuery({
    queryKey: ["audit-events-recent"],
    queryFn: () => auditApi.listAuditEvents({ limit: 5 }),
  });

  const { data: functionsData, isLoading: functionsLoading } = useQuery({
    queryKey: ["admin-functions-stats"],
    queryFn: () => adminFunctionsApi.listFunctions({ limit: 100 }),
  });

  const { data: dashboardActivity, isLoading: activityLoading } = useQuery({
    queryKey: ["admin-dashboard-activity", 7],
    queryFn: () => adminDashboardApi.getActivity({ days: 7 }),
  });

  const { data: dashboardRevenue, isLoading: revenueLoading } = useQuery({
    queryKey: ["admin-dashboard-revenue", 7],
    queryFn: () => adminDashboardApi.getRevenue({ months: 7 }),
  });

  const { data: systemHealthData, isLoading: healthLoading } = useQuery({
    queryKey: ["admin-system-health"],
    queryFn: () => healthApi.getSystemHealth(),
  });

  const { data: quickStats } = useQuery({
    queryKey: ["admin-quick-stats"],
    queryFn: () => adminDashboardApi.getQuickStats(),
  });

  const activityData = useMemo(() => {
    const series = dashboardActivity?.series ?? [];
    return series.map((p) => ({
      day: p.day_label,
      users: p.new_users,
      functions: p.function_calls,
    }));
  }, [dashboardActivity]);

  const revenueData = useMemo(() => {
    const series = dashboardRevenue?.series ?? [];
    return series.map((p) => ({
      month: p.month,
      revenue: Math.round(p.revenue_cents / 100),
    }));
  }, [dashboardRevenue]);

  const systemHealth: SystemHealthItem[] = useMemo(() => {
    const services = systemHealthData?.services ?? [];
    return services.map((s) => ({
      name: s.name,
      status: (s.status === "unhealthy" ? "down" : s.status === "degraded" ? "degraded" : "healthy") as SystemHealthItem["status"],
      latency: s.latency_ms != null ? `${s.latency_ms}ms` : undefined,
      uptime: s.uptime_percent != null ? `${s.uptime_percent.toFixed(2)}%` : undefined,
    }));
  }, [systemHealthData]);

  const totalUsers = userStats?.total_users ?? 0;
  const activeUsers = userStats?.active_users ?? 0;
  const adminUsers = userStats?.admin_users ?? 0;
  const activeSubscriptions = subscriptionsData?.subscriptions?.filter(s => s.status === "active").length ?? 0;
  const totalFunctions = functionsData?.total ?? 0;
  const recentAuditEvents = auditData?.events ?? [];

  const isLoading = usersLoading || subsLoading || auditLoading || functionsLoading;
  const mrrCents = dashboardRevenue?.mrr_cents ?? 0;
  const mrrDisplay = `$${(mrrCents / 100).toLocaleString(undefined, { maximumFractionDigits: 0 })}`;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-text-primary">
            Admin Dashboard
          </h1>
          <p className="text-text-secondary mt-1">
            Platform overview and management hub
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span className="text-xs font-medium text-emerald-600 dark:text-emerald-400">All Systems Operational</span>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="border-border-default hover:bg-bg-hover text-text-secondary"
            onClick={() => window.location.reload()}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          {
            title: "Total Users",
            value: usersLoading ? null : totalUsers,
            icon: Users,
            iconColor: "text-blue-500",
            bgColor: "bg-blue-500/10",
            trend: "up",
            change: "+12%",
            sub: `${activeUsers} active`,
          },
          {
            title: "Active Subscriptions",
            value: subsLoading ? null : activeSubscriptions,
            icon: CreditCard,
            iconColor: "text-purple-500",
            bgColor: "bg-purple-500/10",
            trend: "up",
            change: "+8%",
            sub: "paying tenants",
          },
          {
            title: "Total Functions",
            value: functionsLoading ? null : totalFunctions,
            icon: Code,
            iconColor: "text-violet-500",
            bgColor: "bg-violet-500/10",
            trend: "up",
            change: "+23%",
            sub: "deployed & published",
          },
          {
            title: "Admin Users",
            value: usersLoading ? null : adminUsers,
            icon: Shield,
            iconColor: "text-red-500",
            bgColor: "bg-red-500/10",
            trend: "neutral",
            change: "0%",
            sub: "with elevated access",
          },
        ].map((metric) => (
          <Card key={metric.title} className="glass-card hover-lift overflow-hidden">
            <CardContent className="p-5">
              <div className="flex items-start justify-between mb-3">
                <div className={cn("p-2.5 rounded-xl", metric.bgColor)}>
                  <metric.icon className={cn("w-5 h-5", metric.iconColor)} />
                </div>
                <div className={cn(
                  "flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full",
                  metric.trend === "up" && "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
                  metric.trend === "down" && "bg-red-500/10 text-red-600 dark:text-red-400",
                  metric.trend === "neutral" && "bg-slate-500/10 text-slate-600 dark:text-slate-400",
                )}>
                  {metric.trend === "up" && <TrendingUp className="w-3 h-3" />}
                  {metric.trend === "down" && <TrendingDown className="w-3 h-3" />}
                  {metric.change}
                </div>
              </div>
              {metric.value === null ? (
                <Skeleton className="h-8 w-16 mb-1" />
              ) : (
                <div className="text-2xl font-bold text-text-primary">{metric.value.toLocaleString()}</div>
              )}
              <p className="text-xs text-text-muted mt-1">{metric.sub}</p>
              <p className="text-sm font-medium text-text-secondary mt-0.5">{metric.title}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Activity Chart */}
        <Card className="lg:col-span-2 glass-card">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-text-primary flex items-center gap-2">
                <Activity className="w-5 h-5 text-brand-500" />
                Platform Activity
              </CardTitle>
              <Badge variant="outline" className="text-xs border-border-default text-text-muted">
                Last 7 days
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            {activityLoading ? (
              <div className="h-[200px] flex items-center justify-center">
                <Skeleton className="h-full w-full" />
              </div>
            ) : (
              <>
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={activityData} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorFunctions" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                <XAxis dataKey="day" tick={{ fill: "var(--text-muted)", fontSize: 12 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: "var(--text-muted)", fontSize: 12 }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{
                    background: "var(--bg-tertiary)",
                    border: "1px solid var(--border-default)",
                    borderRadius: "8px",
                    color: "var(--text-primary)",
                    fontSize: "12px",
                  }}
                />
                <Area type="monotone" dataKey="users" stroke="#6366f1" strokeWidth={2} fill="url(#colorUsers)" name="New Users" isAnimationActive={false} />
                <Area type="monotone" dataKey="functions" stroke="#10b981" strokeWidth={2} fill="url(#colorFunctions)" name="Function Calls" isAnimationActive={false} />
              </AreaChart>
            </ResponsiveContainer>
            <div className="flex items-center gap-4 mt-2">
              <div className="flex items-center gap-1.5">
                <span className="w-3 h-0.5 bg-brand-500 rounded" />
                <span className="text-xs text-text-muted">New Users</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="w-3 h-0.5 bg-emerald-500 rounded" />
                <span className="text-xs text-text-muted">Function Calls</span>
              </div>
            </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Revenue Chart */}
        <Card className="glass-card">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-text-primary flex items-center gap-2">
                <BarChart3 className="w-5 h-5 text-purple-500" />
                Revenue
              </CardTitle>
              <Badge variant="outline" className="text-xs border-border-default text-text-muted">
                7 months
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            {revenueLoading ? (
              <div className="h-[200px] flex items-center justify-center">
                <Skeleton className="h-full w-full" />
              </div>
            ) : (
              <>
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={revenueData} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                    <XAxis dataKey="month" tick={{ fill: "var(--text-muted)", fontSize: 11 }} axisLine={false} tickLine={false} />
                    <YAxis tick={{ fill: "var(--text-muted)", fontSize: 11 }} axisLine={false} tickLine={false} tickFormatter={(v) => `$${v}`} />
                    <Tooltip
                      contentStyle={{
                        background: "var(--bg-tertiary)",
                        border: "1px solid var(--border-default)",
                        borderRadius: "8px",
                        color: "var(--text-primary)",
                        fontSize: "12px",
                      }}
                      formatter={(value) => [`$${value}`, "Revenue"]}
                    />
                    <Bar dataKey="revenue" fill="#8b5cf6" radius={[4, 4, 0, 0]} isAnimationActive={false} />
                  </BarChart>
                </ResponsiveContainer>
                <div className="mt-2 flex items-center justify-between">
                  <span className="text-xs text-text-muted">MRR</span>
                  <span className="text-sm font-semibold text-text-primary">{mrrDisplay}</span>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* System Health + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* System Health */}
        <Card className="glass-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-text-primary flex items-center gap-2">
              <Zap className="w-5 h-5 text-amber-500" />
              System Health
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {healthLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))
            ) : (
              <>
                {systemHealth.length === 0 ? (
                  <div className="text-center py-6 text-text-muted text-sm">No health data</div>
                ) : (
                  systemHealth.map((service) => (
                    <div
                      key={service.name}
                      className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-subtle hover:border-border-default transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <HealthStatusDot status={service.status} />
                        <span className="text-sm font-medium text-text-primary">{service.name}</span>
                      </div>
                      <div className="flex items-center gap-4 text-xs text-text-muted">
                        {service.latency && (
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {service.latency}
                          </span>
                        )}
                        {service.uptime && (
                          <span className="text-emerald-600 dark:text-emerald-400 font-medium">{service.uptime}</span>
                        )}
                      </div>
                    </div>
                  ))
                )}
                <div className="pt-2 flex items-center justify-between text-xs text-text-muted">
                  <span className="flex items-center gap-1">
                    <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />
                    {systemHealth.every((s) => s.status === "healthy") ? "All services operational" : "Some issues detected"}
                  </span>
                  <span>Updated just now</span>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Recent Audit Events */}
        <Card className="glass-card">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-text-primary flex items-center gap-2">
                <Shield className="w-5 h-5 text-red-500" />
                Recent Audit Events
              </CardTitle>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate("/admin/audit")}
                className="text-text-muted hover:text-text-primary text-xs"
              >
                View all
                <ArrowRight className="w-3 h-3 ml-1" />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-2">
            {auditLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))
            ) : recentAuditEvents.length === 0 ? (
              <div className="text-center py-6 text-text-muted text-sm">
                No recent audit events
              </div>
            ) : (
              recentAuditEvents.slice(0, 5).map((event) => (
                <div
                  key={event.id}
                  className="flex items-start gap-3 p-2.5 rounded-lg hover:bg-bg-hover transition-colors"
                >
                  <div className={cn(
                    "mt-0.5 w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0",
                    event.success ? "bg-emerald-500/10" : "bg-red-500/10"
                  )}>
                    {event.success ? (
                      <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />
                    ) : (
                      <AlertTriangle className="w-3.5 h-3.5 text-red-500" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-text-primary truncate">
                      <span className="font-medium">{event.actor_email || "System"}</span>
                      {" · "}
                      <span className="text-text-secondary">{event.action}</span>
                    </p>
                    <p className="text-xs text-text-muted">
                      {new Date(event.timestamp).toLocaleString()}
                    </p>
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>

      {/* Admin Sections Grid */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-text-primary">Management Sections</h2>
          <span className="text-sm text-text-muted">{adminSections.length} sections</span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {adminSections.map((section) => {
            const Icon = section.icon;
            return (
              <button
                key={section.path}
                onClick={() => navigate(section.path)}
                className={cn(
                  "group text-left p-4 rounded-xl border transition-all duration-200",
                  "bg-bg-secondary hover:bg-bg-hover",
                  "hover:shadow-md hover:-translate-y-0.5",
                  section.borderColor
                )}
              >
                <div className="flex items-start gap-3">
                  <div className={cn("p-2 rounded-lg flex-shrink-0", section.bgColor)}>
                    <Icon className={cn("w-5 h-5", section.iconColor)} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <p className="font-semibold text-text-primary text-sm">{section.title}</p>
                      <ArrowRight className="w-3.5 h-3.5 text-text-muted opacity-0 group-hover:opacity-100 transition-opacity -translate-x-1 group-hover:translate-x-0 duration-200" />
                    </div>
                    <p className="text-xs text-text-muted mt-0.5 leading-relaxed">{section.description}</p>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Quick Stats Footer */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-5 rounded-xl bg-bg-secondary border border-border-subtle">
        {[
          { label: "Platform Uptime", value: quickStats ? `${quickStats.platform_uptime_percent}%` : "—", icon: Globe, color: "text-emerald-500" },
          { label: "Avg Response Time", value: quickStats ? `${quickStats.avg_response_time_ms}ms` : "—", icon: Zap, color: "text-amber-500" },
          { label: "Functions Executed", value: quickStats?.functions_executed ?? "—", icon: Code, color: "text-violet-500" },
          { label: "Data Processed", value: quickStats?.data_processed ?? "—", icon: Database, color: "text-cyan-500" },
        ].map((stat) => (
          <div key={stat.label} className="flex items-center gap-3">
            <div className={cn("p-2 rounded-lg bg-bg-tertiary")}>
              <stat.icon className={cn("w-4 h-4", stat.color)} />
            </div>
            <div>
              <p className="text-sm font-bold text-text-primary">{stat.value}</p>
              <p className="text-xs text-text-muted">{stat.label}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

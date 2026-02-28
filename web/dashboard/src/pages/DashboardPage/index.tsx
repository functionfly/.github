import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { FunctionSquare, Activity, Globe, Zap, Play, X, Loader2 } from "lucide-react";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { functionsApi } from "@/api/functions";
import { providersApi } from "@/api/providers";
import { dashboardApi } from "@/api/dashboard";
import {
  MetricCard,
  UsageGraph,
  ExecutionRateChart,
  MemoryUsageGauge,
  TrustScoreBadge,
  AgentActivityFeed,
  SystemHealthIndicator,
  QuickCreateAgentCard,
} from "@/components/dashboard";
import type { AgentActivityItem } from "@/components/dashboard";

export function DashboardPage() {
  const { canResume, completedSteps } = useOnboardingStore();
  const navigate = useNavigate();

  const { data: functionsData, isLoading: functionsLoading } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });

  const { data: providers, isLoading: providersLoading } = useQuery({
    queryKey: ["providers"],
    queryFn: () => providersApi.getConnectedProviders(),
  });

  const functions = functionsData?.functions ?? [];
  const activeFunctions = functions.filter((f) => f.status === "deployed").length;

  const handleResumeOnboarding = () => {
    navigate("/onboarding");
  };

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ["dashboard", "usage"],
    queryFn: () => dashboardApi.getUsage(14),
  });

  const { data: executionRateDataRes, isLoading: executionRateLoading } = useQuery({
    queryKey: ["dashboard", "execution-rate"],
    queryFn: () => dashboardApi.getExecutionRate(24),
  });

  const { data: activityData, isLoading: activityLoading } = useQuery({
    queryKey: ["dashboard", "activity"],
    queryFn: () => dashboardApi.getActivity(20),
  });

  const usageGraphData = useMemo(() => {
    const raw = usageData?.data ?? [];
    return raw.map((d) => ({
      time: new Date(d.time + "Z").toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      value: Number(d.value),
    }));
  }, [usageData]);

  const executionRateData = useMemo(() => {
    const raw = executionRateDataRes?.data ?? [];
    return raw.map((d) => ({
      time: d.time,
      rate: Number(d.rate),
    }));
  }, [executionRateDataRes]);

  const agentActivities: AgentActivityItem[] = useMemo(() => {
    const raw = activityData?.activities ?? [];
    return raw.map((a) => ({
      id: a.id,
      type: (a.type as AgentActivityItem["type"]) || "info",
      title: a.title,
      description: a.description,
      timestamp: new Date(a.timestamp),
      agentId: a.function_id,
      agentName: a.function_name,
    }));
  }, [activityData]);

  const sparklineUp = useMemo(() => [10, 14, 12, 18, 22, 20, 24], []);
  const sparklineFlat = useMemo(() => [20, 22, 19, 21, 20, 23, 22], []);

  return (
    <div className="relative space-y-6">
      {/* Resume Onboarding Banner */}
      {canResume() && (
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass-card glow hover-lift p-4"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-[#6366f1]/20 rounded-full flex items-center justify-center">
                <Play className="w-5 h-5 text-[#6366f1]" />
              </div>
              <div>
                <h3 className="font-semibold text-text-primary">
                  Complete Your Setup
                </h3>
                <p className="text-sm text-text-secondary">
                  You've completed {completedSteps.length} of 4 onboarding steps.
                  Continue where you left off to unlock all features.
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={handleResumeOnboarding}
                className="btn-primary"
                size="sm"
              >
                <Play className="w-4 h-4 mr-2" />
                Resume Setup
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  localStorage.setItem('onboarding-banner-dismissed', 'true');
                  window.location.reload();
                }}
              >
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </motion.div>
      )}

      {/* Header + System Health */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <div className="text-center lg:text-left">
          <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
            <span className="text-text-primary text-glow">Dashboard</span>
          </h1>
          <p className="text-text-secondary text-lg">
            Welcome back! Here&apos;s what&apos;s happening with your functions.
          </p>
        </div>
        <div className="flex justify-center sm:justify-end">
          <SystemHealthIndicator status="healthy" showLabel size="md" />
        </div>
      </motion.div>

      {/* Metric Cards */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.1 }}
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
      >
        <MetricCard
          title="Active Functions"
          value={functionsLoading ? "—" : activeFunctions}
          changeLabel="total deployed"
          changePercent={functions.length > 0 ? 12 : undefined}
          sparklineData={sparklineUp}
          icon={<FunctionSquare className="h-5 w-5" />}
        />
        <MetricCard
          title="Avg Latency"
          value="—"
          changeLabel="no data yet"
          icon={<Zap className="h-5 w-5" />}
        />
        <MetricCard
          title="Uptime"
          value="99.9%"
          changePercent={0.1}
          changeLabel="vs last 7d"
          sparklineData={sparklineFlat}
          icon={<Activity className="h-5 w-5" />}
        />
        <MetricCard
          title="Requests This Month"
          value="12.4k"
          changePercent={8.2}
          changeLabel="vs last month"
          sparklineData={sparklineUp}
          icon={<Globe className="h-5 w-5" />}
        />
      </motion.div>

      {/* Quick Create */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.15 }}
      >
        <QuickCreateAgentCard
          title="Deploy a function"
          description="Create and deploy a new function in minutes."
          actionLabel="New function"
          onCreateClick={() => navigate("/functions/new")}
        />
      </motion.div>

      {/* Usage & Execution Charts */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="grid grid-cols-1 lg:grid-cols-2 gap-4"
      >
        {usageLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <UsageGraph
            data={usageGraphData}
            title="Usage (last 14 days)"
            valueLabel="Requests"
          />
        )}
        {executionRateLoading ? (
          <Card className="border-theme bg-card h-[280px] flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <ExecutionRateChart
            data={executionRateData}
            title="Execution rate (last 24h)"
            unit="exec/s"
          />
        )}
      </motion.div>

      {/* Memory & Trust */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.25 }}
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
      >
        <MemoryUsageGauge percent={62} label="Memory" size="md" />
        <Card className="border-theme bg-card flex flex-col justify-center p-6">
          <CardHeader className="p-0 pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Trust score
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <TrustScoreBadge trustScore={85} showScore size="lg" />
          </CardContent>
        </Card>
      </motion.div>

      {/* Main Content Grid */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.3 }}
        className="grid grid-cols-1 lg:grid-cols-3 gap-6"
      >
        {/* Provider Status */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          className="lg:col-span-2"
        >
          <Card className="glass-card glow hover-lift">
            <CardHeader>
              <CardTitle className="text-text-primary text-glow">Provider Status</CardTitle>
            </CardHeader>
            <CardContent>
              {providersLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
                </div>
              ) : !providers || providers.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-text-secondary text-sm">No providers connected yet.</p>
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-3"
                    onClick={() => navigate("/providers")}
                  >
                    Connect a Provider
                  </Button>
                </div>
              ) : (
                <div className="space-y-4">
                  {providers.map((provider, index) => (
                    <motion.div
                      key={provider.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                      className="glass-light hover-lift p-4 rounded-lg border border-white/8 hover:border-brand-500/30 transition-all duration-300"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div className="w-10 h-10 rounded-lg bg-bg-tertiary flex items-center justify-center">
                            <ProviderIcon provider={provider.id} size="lg" />
                          </div>
                          <div>
                            <p className="font-medium text-white">{provider.name}</p>
                            <p className="text-sm text-text-muted">Global</p>
                          </div>
                        </div>
                        <StatusBadge status={provider.status} />
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* Recent Functions */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
        >
          <Card className="glass-card glow hover-lift">
            <CardHeader>
              <CardTitle className="text-text-primary text-glow">Recent Functions</CardTitle>
            </CardHeader>
            <CardContent>
              {functionsLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
                </div>
              ) : functions.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-text-secondary text-sm">No functions deployed yet.</p>
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-3"
                    onClick={() => navigate("/functions/new")}
                  >
                    Deploy a Function
                  </Button>
                </div>
              ) : (
                <div className="space-y-4">
                  {functions.slice(0, 5).map((fn, index) => (
                    <motion.div
                      key={fn.id}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.5, delay: 0.5 + index * 0.1 }}
                      className="flex gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors duration-200 cursor-pointer"
                      onClick={() => navigate(`/functions/${fn.id}`)}
                    >
                      <div className="w-2 h-2 mt-2 rounded-full bg-linear-to-r from-[#6366f1] to-[#8b5cf6]" />
                      <div>
                        <p className="text-sm text-text-primary font-medium">{fn.name}</p>
                        <p className="text-xs text-text-muted capitalize">{fn.status || "unknown"}</p>
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      </motion.div>

      {/* Agent Activity Feed */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.35 }}
      >
        {activityLoading ? (
          <Card className="border-theme bg-card flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
          </Card>
        ) : (
          <AgentActivityFeed
            activities={agentActivities}
            title="Recent activity"
            maxItems={5}
          />
        )}
      </motion.div>
    </div>
  );
}

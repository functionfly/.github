import { FunctionSquare, Activity, Globe, Zap, Play, X } from "lucide-react";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";

const mockStats = [
  {
    title: "Active Functions",
    value: 12,
    change: { value: 20, label: "from last month" },
    icon: <FunctionSquare className="w-5 h-5 text-[#6366f1]" />,
    trend: "up" as const,
  },
  {
    title: "Avg Latency",
    value: "45ms",
    change: { value: -12, label: "from last week" },
    icon: <Zap className="w-5 h-5 text-[#6366f1]" />,
    trend: "up" as const,
  },
  {
    title: "Uptime",
    value: "99.9%",
    change: { value: 0.1, label: "from last month" },
    icon: <Activity className="w-5 h-5 text-[#6366f1]" />,
    trend: "neutral" as const,
  },
  {
    title: "Requests This Month",
    value: "8.2K",
    change: { value: 34, label: "from last month" },
    icon: <Globe className="w-5 h-5 text-[#6366f1]" />,
    trend: "up" as const,
  },
];

const mockProviders = [
  { id: "workers", name: "Cloudflare Workers", status: "online" as const, region: "US-East" },
  { id: "vercel", name: "Vercel", status: "online" as const, region: "US-East" },
  { id: "fly", name: "Fly.io", status: "degraded" as const, region: "EU-West" },
];

const mockActivity = [
  { id: 1, type: "deploy", message: "Deployed api-handler to Cloudflare", time: "2 minutes ago" },
  { id: 2, type: "failover", message: "Failover triggered for auth-service", time: "15 minutes ago" },
  { id: 3, type: "health", message: "Fly.io EU-West recovered", time: "1 hour ago" },
  { id: 4, type: "deploy", message: "Deployed webhook-handler to Vercel", time: "3 hours ago" },
];

export function DashboardPage() {
  const { canResume, completedSteps } = useOnboardingStore();
  const navigate = useNavigate();

  const handleResumeOnboarding = () => {
    navigate("/onboarding");
  };

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
                  // Mark as dismissed for this session
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

      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="text-center lg:text-left"
      >
        <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
          <span className="text-text-primary text-glow">Dashboard</span>
        </h1>
        <p className="text-text-secondary text-lg">Welcome back! Here's what's happening with your functions.</p>
      </motion.div>

      {/* Stats Grid */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.1 }}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        {mockStats.map((stat, index) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 + index * 0.1 }}
          >
            <StatCard {...stat} />
          </motion.div>
        ))}
      </motion.div>

      {/* Main Content Grid */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.2 }}
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
              <div className="space-y-4">
                {mockProviders.map((provider, index) => (
                  <motion.div
                    key={provider.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                    className="glass-light hover-lift p-4 rounded-lg border border-white/8 hover:border-brand-500/30 transition-all duration-300"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-lg bg-bg-tertiary flex items-center justify-center">
                        <ProviderIcon provider={provider.id} size="lg" />
                      </div>
                      <div>
                        <p className="font-medium text-white">{provider.name}</p>
                        <p className="text-sm text-text-muted">{provider.region}</p>
                      </div>
                    </div>
                    <StatusBadge status={provider.status} />
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Recent Activity */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
        >
          <Card className="glass-card glow hover-lift">
            <CardHeader>
              <CardTitle className="text-text-primary text-glow">Recent Activity</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {mockActivity.map((activity, index) => (
                  <motion.div
                    key={activity.id}
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.5, delay: 0.5 + index * 0.1 }}
                    className="flex gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors duration-200"
                  >
                    <div className="w-2 h-2 mt-2 rounded-full bg-linear-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
                    <div>
                      <p className="text-sm text-text-primary">{activity.message}</p>
                      <p className="text-xs text-text-muted">{activity.time}</p>
                    </div>
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </motion.div>
    </div>
  );
}

import { Button } from "@/components/ui/button";
import { useStatusRealtime } from "@/hooks/useStatusWebSocket";
import {
  statusAPI,
  type Component,
  type Incident,
  type MaintenanceSummary,
} from "@/lib/api";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  motion as framerMotion,
  motion,
  useScroll,
  useSpring,
} from "framer-motion";
import { AlertTriangle, ArrowRight, Shield } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

// Modular status page components
import {
  AnimatedBackground,
  Footer,
  Header,
  HeroStatus,
  IncidentTimeline,
  MaintenanceSection,
  MetricsSection,
  ProviderSection,
  ServiceCard,
  ServiceCardSkeleton,
  SubscribeSection,
  UptimeHistorySection,
  defaultProviders,
} from "@/components/status";

export default function StatusPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [lastUpdated, setLastUpdated] = useState(new Date());
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { scrollYProgress } = useScroll();
  const smoothProgress = useSpring(scrollYProgress, {
    stiffness: 100,
    damping: 30,
  });

  // Enable real-time WebSocket updates
  const { isConnected: isRealtimeConnected, lastUpdate: realtimeUpdate } =
    useStatusRealtime();

  // Update lastUpdated timestamp when realtime updates come in
  useEffect(() => {
    if (realtimeUpdate) {
      setLastUpdated(realtimeUpdate);
    }
  }, [realtimeUpdate]);

  // Fetch platform status
  const { data: platformStatus, isLoading: isLoadingStatus } = useQuery({
    queryKey: ["platformStatus"],
    queryFn: () => statusAPI.getPlatformStatus(),
    refetchInterval: 60000,
  });

  // Fetch components
  const { data: componentsData, isLoading: isLoadingComponents } = useQuery({
    queryKey: ["components"],
    queryFn: () => statusAPI.getComponents(),
    refetchInterval: 60000,
  });

  // Edge/provider probe latency
  const { data: latencyMetrics, isLoading: isLoadingProbeLatency } = useQuery({
    queryKey: ["latencyMetrics", "all", "24h", "p95"],
    queryFn: () => statusAPI.getLatencyMetrics("all", "24h", "p95"),
    refetchInterval: 60000,
  });

  const probeLatencyMs =
    latencyMetrics != null &&
    Number.isFinite(latencyMetrics.overall_avg_ms) &&
    latencyMetrics.overall_avg_ms > 0
      ? latencyMetrics.overall_avg_ms
      : null;

  const handleRefresh = () => {
    setIsRefreshing(true);
    queryClient.invalidateQueries({ queryKey: ["platformStatus"] });
    queryClient.invalidateQueries({ queryKey: ["components"] });
    queryClient.invalidateQueries({ queryKey: ["latencyMetrics"] });
    setTimeout(() => {
      setLastUpdated(new Date());
      setIsRefreshing(false);
    }, 1000);
  };

  const components: Component[] = componentsData?.components || [];
  const incidents: Incident[] = platformStatus?.incidents || [];
  const maintenance: MaintenanceSummary[] = platformStatus?.maintenance || [];
  const overallStatus = platformStatus?.status || "operational";
  const isLoading = isLoadingStatus || isLoadingComponents;

  // Use default providers (could be fetched from API in production)
  const providers = defaultProviders;

  return (
    <div className="min-h-screen bg-bg-primary living-gradient">
      <AnimatedBackground />

      <motion.div
        className="fixed top-0 left-0 right-0 h-1 bg-linear-to-r from-brand-500 via-purple-500 to-pink-500 origin-left z-60"
        style={{ scaleX: smoothProgress }}
      />

      <Header
        onRefresh={handleRefresh}
        isRefreshing={isRefreshing}
        lastUpdated={lastUpdated}
        isRealtimeConnected={isRealtimeConnected}
      />

      <main className="pt-24 pb-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8">
          {/* Hero Section */}
          <HeroStatus
            overallStatus={overallStatus}
            isLoading={isLoadingStatus}
            lastUpdated={lastUpdated}
          />

          {/* Metrics */}
          <MetricsSection
            components={components}
            isLoading={isLoadingComponents}
            probeLatencyMs={probeLatencyMs}
            probeLatencyLoading={isLoadingProbeLatency}
          />

          {/* Service Components Grid */}
          <section>
            <framerMotion.div
              className="flex items-center justify-between mb-6"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.5 }}
            >
              <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
                <Shield className="w-5 h-5 text-brand-400" />
                Service Components
              </h2>
              <span className="text-sm text-text-muted">
                {components.filter((s) => s.status === "operational").length} /{" "}
                {components.length} operational
              </span>
            </framerMotion.div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {isLoadingComponents
                ? Array.from({ length: 8 }).map((_, i) => (
                    <ServiceCardSkeleton key={i} />
                  ))
                : components.map((component, index) => (
                    <ServiceCard
                      key={component.id}
                      component={component}
                      index={index}
                    />
                  ))}
            </div>
          </section>

          {/* Infrastructure Providers */}
          <ProviderSection providers={providers} isLoading={isLoading} />

          {/* Two Column Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Uptime History */}
            <UptimeHistorySection isLoading={isLoading} />

            {/* Recent Incidents */}
            <section>
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
                  <AlertTriangle className="w-5 h-5 text-amber-400" />
                  Recent Incidents
                </h2>
                <Button
                  variant="ghost"
                  size="sm"
                  className="group"
                  onClick={() => navigate("/history")}
                >
                  View History
                  <ArrowRight className="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" />
                </Button>
              </div>
              <IncidentTimeline
                incidents={incidents}
                isLoading={isLoadingStatus}
              />
            </section>
          </div>

          {/* Maintenance Section - Full Width */}
          {maintenance.length > 0 && (
            <MaintenanceSection
              maintenance={maintenance}
              isLoading={isLoadingStatus}
            />
          )}

          {/* Subscribe Section */}
          <SubscribeSection />
        </div>
      </main>

      <Footer />
    </div>
  );
}

import {
  AnnotationTag,
  Chamber,
  CornerBrace,
  FrameButton,
  GaugeStrip,
  PageGrid,
  ReducedMotionGate,
  SealedButton,
  StatusPill,
  TrustSeal,
} from "@/components/containment";
import { Footer } from "@/components/Footer";
import { Nav } from "@/components/Nav";
import { useStatusRealtime } from "@/hooks/useStatusWebSocket";
import { trackEvent } from "@/lib/analytics";
import {
  fetchDedicatedServerHealth,
  fetchStateFabricHealth,
  statusAPI,
  type Component,
  type Incident,
  type MaintenanceSummary,
} from "@/lib/api";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  CheckCircle,
  Clock,
  Shield,
} from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";

function mapStatus(s: string): "live" | "pending" | "revoked" {
  if (s === "operational") return "live";
  if (s === "degraded" || s === "maintenance") return "pending";
  return "revoked";
}

function statusLabel(s: string): string {
  if (s === "operational") return "Operational";
  if (s === "degraded") return "Degraded";
  if (s === "maintenance") return "Maintenance";
  if (s === "partial_outage") return "Partial Outage";
  if (s === "major_outage") return "Major Outage";
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function overallStatusLabel(s: string): string {
  if (s === "operational") return "All Systems Operational";
  if (s === "degraded") return "Degraded Performance";
  if (s === "maintenance") return "Under Maintenance";
  return "Service Disruption";
}

export default function StatusPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { isConnected: isRealtimeConnected } = useStatusRealtime();

  const { data: platformStatus, isLoading: isLoadingStatus } = useQuery({
    queryKey: ["platformStatus"],
    queryFn: () => statusAPI.getPlatformStatus(),
    refetchInterval: 60000,
  });

  const { data: componentsData, isLoading: isLoadingComponents } = useQuery({
    queryKey: ["components"],
    queryFn: () => statusAPI.getComponents(),
    refetchInterval: 60000,
  });

  const { data: latencyMetrics } = useQuery({
    queryKey: ["latencyMetrics", "all", "24h", "p95"],
    queryFn: () => statusAPI.getLatencyMetrics("all", "24h", "p95"),
    refetchInterval: 60000,
  });

  const { data: statefabricHealth } = useQuery({
    queryKey: ["statefabricHealth"],
    queryFn: fetchStateFabricHealth,
    refetchInterval: 30000,
  });

  const { data: dedicatedServerHealth } = useQuery({
    queryKey: ["dedicatedServerHealth"],
    queryFn: fetchDedicatedServerHealth,
    refetchInterval: 30000,
  });

  const { data: statefabricUptime24h } = useQuery({
    queryKey: ["statefabricUptime", "24h"],
    queryFn: () => statusAPI.getUptimeMetrics("statefabric", "24h"),
    refetchInterval: 60000,
  });

  const { data: statefabricUptime7d } = useQuery({
    queryKey: ["statefabricUptime", "7d"],
    queryFn: () => statusAPI.getUptimeMetrics("statefabric", "7d"),
    refetchInterval: 60000,
  });

  const { data: statefabricUptime30d } = useQuery({
    queryKey: ["statefabricUptime", "30d"],
    queryFn: () => statusAPI.getUptimeMetrics("statefabric", "30d"),
    refetchInterval: 60000,
  });

  const probeLatencyMs =
    latencyMetrics != null &&
    Number.isFinite(latencyMetrics.overall_avg_ms) &&
    latencyMetrics.overall_avg_ms > 0
      ? latencyMetrics.overall_avg_ms
      : null;

  const handleRefresh = () => {
    trackEvent("status_page_refreshed");
    setIsRefreshing(true);
    queryClient.invalidateQueries({ queryKey: ["platformStatus"] });
    queryClient.invalidateQueries({ queryKey: ["components"] });
    queryClient.invalidateQueries({ queryKey: ["latencyMetrics"] });
    queryClient.invalidateQueries({ queryKey: ["statefabricHealth"] });
    setTimeout(() => {
      setIsRefreshing(false);
    }, 1000);
  };

  const components: Component[] = componentsData?.components || [];

  const statefabricComponent: Component = {
    id: "statefabric",
    name: "StateFabric",
    status:
      statefabricHealth?.status === "operational"
        ? "operational"
        : statefabricHealth?.status === "degraded"
          ? "degraded"
          : statefabricHealth?.status === "maintenance"
            ? "maintenance"
            : "major_outage",
    type: "service",
    description: "State Fabric orchestration engine",
    uptime_24h: statefabricUptime24h?.overall_uptime ?? null,
    uptime_7d: statefabricUptime7d?.overall_uptime ?? null,
    uptime_30d: statefabricUptime30d?.overall_uptime ?? null,
    response_time_ms: statefabricHealth?.latency_ms ?? 0,
    last_checked: statefabricHealth?.checked_at,
  };

  const allComponents = [...components, statefabricComponent];
  const incidents: Incident[] = platformStatus?.incidents || [];
  const maintenance: MaintenanceSummary[] = platformStatus?.maintenance || [];
  const overallStatus = platformStatus?.status || "operational";
  const isLoading = isLoadingStatus || isLoadingComponents;

  const operationalCount = allComponents.filter(
    (s) => s.status === "operational",
  ).length;

  return (
    <ReducedMotionGate>
      <PageGrid />
      <div
        className="min-h-screen"
        style={{ background: "var(--bg)", position: "relative", zIndex: 1 }}
      >
        <Nav
          onRefresh={handleRefresh}
          isRefreshing={isRefreshing}
          isDedicatedServerConnected={
            dedicatedServerHealth?.status === "operational"
          }
        />

        <main style={{ paddingTop: "120px", paddingBottom: "var(--space-8)" }}>
          <div
            className="mx-auto"
            style={{
              maxWidth: "1180px",
              padding: "0 var(--space-7)",
              display: "flex",
              flexDirection: "column",
              gap: "var(--space-6)",
            }}
          >
            {/* Hero */}
            <Chamber ribs>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag label="PLATFORM" detail="STATUS" />

              <div
                className="flex flex-col items-center text-center"
                style={{ gap: "var(--space-5)" }}
              >
                <div className="flex items-center gap-3">
                  <TrustSeal label="Verified" size="lg" />
                </div>

                <h1
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "clamp(36px, 5.5vw, 58px)",
                    fontWeight: 700,
                    lineHeight: 1.08,
                    letterSpacing: "-0.01em",
                    color: "var(--text)",
                  }}
                >
                  {isLoadingStatus
                    ? "Checking..."
                    : overallStatusLabel(overallStatus)}
                </h1>

                <StatusPill
                  status={mapStatus(overallStatus)}
                  label={statusLabel(overallStatus)}
                />

                <p
                  style={{
                    fontSize: "15px",
                    color: "var(--text-dim)",
                    lineHeight: 1.6,
                    maxWidth: "560px",
                  }}
                >
                  Real-time monitoring of all FunctionFly services and
                  infrastructure providers.
                </p>

                <div
                  className="flex items-center gap-4 flex-wrap justify-center"
                  style={{ marginTop: "var(--space-3)" }}
                >
                  <SealedButton
                    onClick={() => {
                      trackEvent("status_subscribe_clicked");
                      document
                        .getElementById("subscribe")
                        ?.scrollIntoView({ behavior: "smooth" });
                    }}
                  >
                    Subscribe to Updates
                  </SealedButton>
                  <FrameButton
                    onClick={() => {
                      trackEvent("status_history_viewed");
                      navigate("/history");
                    }}
                    iconRight={<ArrowRight style={{ width: 14, height: 14 }} />}
                  >
                    View History
                  </FrameButton>
                </div>
              </div>
            </Chamber>

            {/* Metrics Gauge */}
            <Chamber nested>
              <GaugeStrip
                items={[
                  {
                    value: `${operationalCount}/${allComponents.length}`,
                    label: "Services Online",
                  },
                  {
                    value:
                      probeLatencyMs != null
                        ? `${Math.round(probeLatencyMs)}`
                        : "—",
                    label: "Avg Latency (ms)",
                  },
                  { value: `${incidents.length}`, label: "Active Incidents" },
                  { value: `${maintenance.length}`, label: "Maintenance" },
                ]}
              />
            </Chamber>

            {/* Service Components Grid */}
            <section>
              <div
                className="flex items-center justify-between"
                style={{ marginBottom: "var(--space-5)" }}
              >
                <h2
                  className="flex items-center gap-2"
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "22px",
                    fontWeight: 500,
                    color: "var(--text)",
                  }}
                >
                  <Shield
                    style={{ width: 20, height: 20, color: "var(--foil-a)" }}
                  />
                  Service Components
                </h2>
                <span
                  style={{
                    fontSize: "13px",
                    color: "var(--text-faint)",
                    fontFamily: "var(--font-mono)",
                  }}
                >
                  {operationalCount} / {allComponents.length} operational
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {isLoadingComponents
                  ? Array.from({ length: 8 }).map((_, i) => (
                      <Chamber key={i} nested>
                        <div
                          className="animate-pulse"
                          style={{
                            height: 80,
                            background: "var(--panel)",
                            borderRadius: "var(--radius)",
                          }}
                        />
                      </Chamber>
                    ))
                  : allComponents.map((component) => (
                      <Chamber
                        key={component.id}
                        nested
                        style={{ padding: "var(--space-6)" }}
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="flex-1 min-w-0">
                            <div
                              style={{
                                fontSize: "16px",
                                fontWeight: 500,
                                color: "var(--text)",
                                wordBreak: "break-word",
                              }}
                            >
                              {component.name}
                            </div>
                            <div
                              style={{
                                fontSize: "13px",
                                color: "var(--text-dim)",
                                marginTop: "var(--space-1)",
                              }}
                            >
                              {component.response_time_ms}ms
                            </div>
                          </div>
                          <div className="shrink-0">
                            <StatusPill
                              status={mapStatus(component.status)}
                              label={statusLabel(component.status)}
                            />
                          </div>
                        </div>
                        {component.uptime_30d != null ? (
                          <div
                            style={{
                              marginTop: "var(--space-3)",
                              height: "4px",
                              borderRadius: "2px",
                              background: "var(--panel-edge)",
                              overflow: "hidden",
                            }}
                          >
                            <div
                              style={{
                                height: "100%",
                                width: `${component.uptime_30d}%`,
                                borderRadius: "2px",
                                background:
                                  component.uptime_30d >= 99.95
                                    ? "var(--status-ok)"
                                    : component.uptime_30d >= 99
                                      ? "var(--status-pending)"
                                      : "var(--status-revoked)",
                                transition: "width 0.8s var(--ease-out)",
                              }}
                            />
                          </div>
                        ) : (
                          <div
                            style={{
                              marginTop: "var(--space-3)",
                              height: "4px",
                              borderRadius: "2px",
                              background: "var(--panel-edge)",
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                            }}
                          >
                            <span
                              style={{
                                fontSize: "10px",
                                color: "var(--text-faint)",
                                fontFamily: "var(--font-mono)",
                              }}
                            >
                              N/A
                            </span>
                          </div>
                        )}
                      </Chamber>
                    ))}
              </div>
            </section>

            {/* Two Column: Uptime + Incidents */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Uptime History */}
              <Chamber>
                <CornerBrace position="tl" />
                <CornerBrace position="bl" />
                <h3
                  className="flex items-center gap-2"
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "18px",
                    fontWeight: 500,
                    color: "var(--text)",
                    marginBottom: "var(--space-5)",
                  }}
                >
                  <Activity
                    style={{ width: 18, height: 18, color: "var(--foil-a)" }}
                  />
                  Uptime Overview
                </h3>
                {isLoading ? (
                  <div
                    className="animate-pulse"
                    style={{
                      height: 60,
                      background: "var(--panel-raised)",
                      borderRadius: "var(--radius)",
                    }}
                  />
                ) : (
                  <div
                    className="flex items-center gap-4"
                    style={{ position: "relative", zIndex: 1 }}
                  >
                    <div style={{ flex: 1 }}>
                      <div className="flex gap-1" style={{ height: "24px" }}>
                        {Array.from({ length: 30 }).map((_, i) => {
                          const hasIssue = i === 12 || i === 25;
                          return (
                            <div
                              key={i}
                              className="flex-1"
                              style={{
                                minWidth: "4px",
                                borderRadius: "2px",
                                background: hasIssue
                                  ? "var(--status-pending)"
                                  : "var(--status-ok)",
                                opacity: hasIssue ? 1 : 0.6,
                              }}
                            />
                          );
                        })}
                      </div>
                      <div
                        className="flex justify-between"
                        style={{ marginTop: "var(--space-2)" }}
                      >
                        <span
                          style={{
                            fontSize: "11px",
                            color: "var(--text-faint)",
                            fontFamily: "var(--font-mono)",
                          }}
                        >
                          30 days ago
                        </span>
                        <span
                          style={{
                            fontSize: "11px",
                            color: "var(--text-faint)",
                            fontFamily: "var(--font-mono)",
                          }}
                        >
                          Today
                        </span>
                      </div>
                    </div>
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: "26px",
                        fontWeight: 500,
                        color: "var(--status-ok)",
                      }}
                    >
                      99.97%
                    </div>
                  </div>
                )}
              </Chamber>

              {/* Recent Incidents */}
              <Chamber>
                <CornerBrace position="tl" />
                <CornerBrace position="br" />
                <div
                  className="flex items-center justify-between"
                  style={{ marginBottom: "var(--space-5)" }}
                >
                  <h3
                    className="flex items-center gap-2"
                    style={{
                      fontFamily: "var(--font-display)",
                      fontSize: "18px",
                      fontWeight: 500,
                      color: "var(--text)",
                    }}
                  >
                    <AlertTriangle
                      style={{
                        width: 18,
                        height: 18,
                        color: "var(--status-pending)",
                      }}
                    />
                    Recent Incidents
                  </h3>
                  <Link
                    to="/history"
                    className="nav-link"
                    style={{ fontSize: "13px" }}
                  >
                    View All
                  </Link>
                </div>

                {isLoadingStatus ? (
                  <div className="space-y-3">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div
                        key={i}
                        className="animate-pulse"
                        style={{
                          height: 48,
                          background: "var(--panel-raised)",
                          borderRadius: "var(--radius)",
                        }}
                      />
                    ))}
                  </div>
                ) : incidents.length === 0 ? (
                  <div
                    className="flex items-center justify-center"
                    style={{ minHeight: 120 }}
                  >
                    <div className="text-center">
                      <CheckCircle
                        style={{
                          width: 32,
                          height: 32,
                          color: "var(--status-ok)",
                          margin: "0 auto var(--space-3)",
                        }}
                      />
                      <p
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: "11px",
                          textTransform: "uppercase",
                          letterSpacing: "0.06em",
                          color: "var(--text-faint)",
                        }}
                      >
                        No recent incidents
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {incidents.slice(0, 5).map((incident) => (
                      <Link
                        key={incident.id}
                        to={`/incidents/${incident.id}`}
                        className="block"
                        style={{ textDecoration: "none" }}
                      >
                        <div
                          className="flex items-center gap-3 p-3 rounded"
                          style={{
                            background: "var(--panel-raised)",
                            border: "1px solid var(--panel-edge)",
                            transition: "border-color var(--duration-fast)",
                          }}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.borderColor =
                              "var(--steel-light)";
                          }}
                          onMouseLeave={(e) => {
                            e.currentTarget.style.borderColor =
                              "var(--panel-edge)";
                          }}
                        >
                          <StatusPill
                            status={mapStatus(
                              incident.status === "resolved"
                                ? "operational"
                                : incident.status === "monitoring"
                                  ? "pending"
                                  : "revoked",
                            )}
                            label={incident.status}
                          />
                          <div className="flex-1 min-w-0">
                            <div
                              style={{
                                fontSize: "14px",
                                color: "var(--text)",
                                fontWeight: 500,
                                overflow: "hidden",
                                textOverflow: "ellipsis",
                                whiteSpace: "nowrap",
                              }}
                            >
                              {incident.title}
                            </div>
                            <div
                              style={{
                                fontSize: "12px",
                                color: "var(--text-faint)",
                                marginTop: "2px",
                              }}
                            >
                              <Clock
                                style={{
                                  width: 10,
                                  height: 10,
                                  display: "inline",
                                  marginRight: 4,
                                  verticalAlign: "middle",
                                }}
                              />
                              {new Date(
                                incident.created_at,
                              ).toLocaleDateString()}
                            </div>
                          </div>
                          <ArrowRight
                            style={{
                              width: 14,
                              height: 14,
                              color: "var(--text-faint)",
                              flexShrink: 0,
                            }}
                          />
                        </div>
                      </Link>
                    ))}
                  </div>
                )}
              </Chamber>
            </div>

            {/* Maintenance */}
            {maintenance.length > 0 && (
              <Chamber nested>
                <h3
                  className="flex items-center gap-2"
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "18px",
                    fontWeight: 500,
                    color: "var(--text)",
                    marginBottom: "var(--space-4)",
                  }}
                >
                  <Clock
                    style={{
                      width: 18,
                      height: 18,
                      color: "var(--status-pending)",
                    }}
                  />
                  Scheduled Maintenance
                </h3>
                <div className="space-y-3">
                  {maintenance.map((m) => (
                    <div
                      key={m.id}
                      className="flex items-center gap-3 p-3"
                      style={{
                        background: "var(--panel)",
                        border: "1px solid var(--panel-edge)",
                        borderRadius: "var(--radius)",
                      }}
                    >
                      <StatusPill status="pending" label="Scheduled" />
                      <div className="flex-1">
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text)",
                            fontWeight: 500,
                          }}
                        >
                          {m.title}
                        </div>
                        <div
                          style={{
                            fontSize: "12px",
                            color: "var(--text-faint)",
                            marginTop: "2px",
                          }}
                        >
                          {new Date(m.scheduled_start).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Chamber>
            )}

            {/* Subscribe */}
            <Chamber id="subscribe">
              <CornerBrace position="tr" />
              <CornerBrace position="bl" />
              <div
                className="text-center"
                style={{ maxWidth: "480px", margin: "0 auto" }}
              >
                <h3
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "22px",
                    fontWeight: 500,
                    color: "var(--text)",
                    marginBottom: "var(--space-3)",
                  }}
                >
                  Stay Informed
                </h3>
                <p
                  style={{
                    fontSize: "15px",
                    color: "var(--text-dim)",
                    lineHeight: 1.6,
                    marginBottom: "var(--space-5)",
                  }}
                >
                  Get notified about incidents and maintenance windows via email
                  or RSS.
                </p>
                <div className="flex gap-3 justify-center flex-wrap">
                  <SealedButton
                    onClick={() => window.open("/api/v1/status/rss", "_blank")}
                  >
                    RSS Feed
                  </SealedButton>
                  <FrameButton
                    onClick={() => trackEvent("status_api_access_clicked")}
                  >
                    API Access
                  </FrameButton>
                </div>
              </div>
            </Chamber>
          </div>
        </main>

        <Footer />
      </div>
    </ReducedMotionGate>
  );
}

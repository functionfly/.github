import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  Plus,
  Database,
  Zap,
  Network,
  Activity,
  Settings,
  MoreVertical,
  Search,
  RefreshCw,
  AlertTriangle,
} from "lucide-react";
import {
  Chamber,
  CornerBrace,
  PageGrid,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  Modal,
} from "@/components/containment";
import {
  useStateFabrics,
  useDeleteStateFabric,
} from "@/hooks/useStateFabric";
import { usePlan } from "@/hooks/usePlan";
import {
  canCreateStateFabric,
  getStateFabricsLimit,
  hasFeature,
} from "@/lib/plan-utils";
import { ROUTES } from "@/lib/constants";
import type { StateFabric } from "@/types";
import "@/styles/sc-tokens.css";

const getTypeIcon = (type: string) => {
  switch (type) {
    case "session": return "👤";
    case "catalog": return "📦";
    case "cache": return "⚡";
    case "workflow": return "🔄";
    default: return "🧵";
  }
};

const getTypeLabel = (type: string) => {
  switch (type) {
    case "session": return "Session Store";
    case "catalog": return "Data Catalog";
    case "cache": return "Cache Layer";
    case "workflow": return "Workflow Engine";
    default: return "Custom Fabric";
  }
};

const statusToPill = (status: string): "live" | "pending" | "revoked" => {
  if (status === "online") return "live";
  if (status === "degraded") return "pending";
  return "revoked";
};

export function StateFabricPage() {
  const navigate = useNavigate();
  const { plan } = usePlan();
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState("all");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [fabricToDelete, setFabricToDelete] = useState<StateFabric | null>(null);
  const [menuOpenId, setMenuOpenId] = useState<string | null>(null);

  const { data: fabrics, isLoading, error, refetch } = useStateFabrics();
  const deleteFabric = useDeleteStateFabric();

  const fabricCount = fabrics?.length ?? 0;
  const canCreate = canCreateStateFabric(plan, fabricCount);
  const stateFabricUnlocked = hasFeature(plan, "STATE_FABRIC");
  const fabricLimit = getStateFabricsLimit(plan);

  const filteredFabrics = fabrics?.filter((fabric) => {
    const matchesSearch =
      fabric.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      fabric.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter === "all" || fabric.status === statusFilter;
    const matchesType = typeFilter === "all" || fabric.type === typeFilter;
    return matchesSearch && matchesStatus && matchesType;
  });

  const stats = {
    total: fabrics?.length || 0,
    active: fabrics?.filter((f) => f.status === "online").length || 0,
    stores: fabrics?.reduce((acc, f) => acc + (f.stores?.length || 0), 0) || 0,
    pipelines: fabrics?.reduce((acc, f) => acc + (f.pipelines?.length || 0), 0) || 0,
  };

  const handleConfirmDelete = async () => {
    if (fabricToDelete) {
      await deleteFabric.mutateAsync(fabricToDelete.id);
      setDeleteDialogOpen(false);
      setFabricToDelete(null);
    }
  };

  if (isLoading) {
    return (
      <div style={{ display: "flex", alignItems: "center", justifyContent: "center", height: 384 }}>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em", color: "var(--text-faint)" }}>
          Loading state fabrics...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ maxWidth: 1180, margin: "0 auto", padding: "var(--space-7)", display: "flex", flexDirection: "column", gap: "var(--space-6)" }}>
        <PageGrid />
        <div>
          <h1 style={{ fontFamily: "var(--font-display)", fontSize: 36, fontWeight: 700, color: "var(--text)" }}>State Fabric</h1>
          <p style={{ color: "var(--text-dim)", marginTop: "var(--space-2)" }}>Manage state and data orchestration across your applications</p>
        </div>
        <Chamber>
          <div style={{ textAlign: "center", padding: "var(--space-8) 0" }}>
            <p style={{ color: "var(--status-revoked)", marginBottom: "var(--space-4)" }}>Failed to load state fabrics</p>
            <FrameButton onClick={() => refetch()} iconLeft={<RefreshCw style={{ width: 14, height: 14 }} />}>
              Retry
            </FrameButton>
          </div>
        </Chamber>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 1180, margin: "0 auto", padding: "var(--space-7)", display: "flex", flexDirection: "column", gap: "var(--space-6)" }}>
      <PageGrid />

      {/* Header */}
      <div style={{ display: "flex", flexWrap: "wrap", alignItems: "center", justifyContent: "space-between", gap: "var(--space-4)" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-display)", fontSize: 36, fontWeight: 700, letterSpacing: "-0.005em", color: "var(--text)" }}>State Fabric</h1>
          <p style={{ color: "var(--text-dim)", marginTop: "var(--space-2)" }}>Manage state and data orchestration across your applications</p>
        </div>
        <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: "var(--space-1)" }}>
          <SealedButton
            onClick={() => navigate("/state-fabric/new")}
            disabled={!canCreate}
            iconLeft={<Plus style={{ width: 14, height: 14 }} />}
            title={
              !stateFabricUnlocked
                ? "State Fabric is available on Starter and higher plans"
                : !canCreate
                  ? `Plan limit reached (${fabricCount} of ${fabricLimit >= 10000 ? "∞" : fabricLimit})`
                  : undefined
            }
          >
            Create State Fabric
          </SealedButton>
          {!canCreate && (
            <p style={{ fontSize: 11, color: "var(--text-faint)", maxWidth: 280, textAlign: "right" }}>
              {!stateFabricUnlocked ? (
                <>Upgrade to use State Fabric. <Link to={ROUTES.PRICING} style={{ color: "var(--status-ok)" }}>View plans</Link></>
              ) : (
                <>Limit reached for your plan. <Link to={ROUTES.PRICING} style={{ color: "var(--status-ok)" }}>Upgrade</Link></>
              )}
            </p>
          )}
        </div>
      </div>

      {/* Stats */}
      <Chamber nested>
        <GaugeStrip>
          <Gauge data={{ value: stats.total, label: "Total Fabrics" }} isFirst />
          <Gauge data={{ value: stats.active, label: "Active" }} />
          <Gauge data={{ value: stats.stores, label: "Stores" }} />
          <Gauge data={{ value: stats.pipelines, label: "Pipelines" }} />
        </GaugeStrip>
      </Chamber>

      {/* Toolbar */}
      <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-3)" }}>
        <div style={{ position: "relative", flex: 1, minWidth: 200 }}>
          <Search style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)", width: 14, height: 14, color: "var(--text-faint)" }} />
          <input
            type="text"
            placeholder="Search state fabrics..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input"
            style={{ paddingLeft: 36 }}
          />
        </div>
        <div style={{ display: "flex", gap: "var(--space-2)" }}>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="input"
            style={{ width: 140, cursor: "pointer", appearance: "none" }}
          >
            <option value="all">All Status</option>
            <option value="online">Online</option>
            <option value="degraded">Degraded</option>
            <option value="offline">Offline</option>
          </select>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="input"
            style={{ width: 140, cursor: "pointer", appearance: "none" }}
          >
            <option value="all">All Types</option>
            <option value="session">Session</option>
            <option value="catalog">Catalog</option>
            <option value="cache">Cache</option>
            <option value="workflow">Workflow</option>
            <option value="custom">Custom</option>
          </select>
          <FrameButton
            size="sm"
            onClick={() => refetch()}
            iconLeft={<RefreshCw style={{ width: 14, height: 14 }} />}
          />
        </div>
      </div>

      {/* Fabrics Grid */}
      {filteredFabrics && filteredFabrics.length > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(480px, 1fr))", gap: "var(--space-5)" }}>
          {filteredFabrics.map((fabric) => (
            <Chamber nested key={fabric.id} style={{ cursor: "pointer" }} onClick={() => navigate(`/state-fabric/${fabric.id}`)}>
              {/* Header row */}
              <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", marginBottom: "var(--space-4)" }}>
                <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                  <div style={{
                    width: 40, height: 40, borderRadius: "var(--radius)",
                    display: "flex", alignItems: "center", justifyContent: "center",
                    fontSize: 18, background: "var(--panel)", border: "1px solid var(--panel-edge)",
                  }}>
                    {getTypeIcon(fabric.type)}
                  </div>
                  <div>
                    <h3 style={{ fontFamily: "var(--font-display)", fontSize: 18, fontWeight: 600, color: "var(--text)" }}>{fabric.name}</h3>
                    <p style={{ fontSize: 13, color: "var(--text-dim)" }}>{fabric.description}</p>
                    <span style={{
                      display: "inline-block", marginTop: "var(--space-1)",
                      fontFamily: "var(--font-mono)", fontSize: 10, fontWeight: 500,
                      textTransform: "uppercase", letterSpacing: "0.06em",
                      color: "var(--text-faint)", padding: "2px 8px",
                      borderRadius: "var(--radius-sm)", border: "1px solid var(--panel-edge)",
                    }}>
                      {getTypeLabel(fabric.type)}
                    </span>
                  </div>
                </div>
                <StatusPill status={statusToPill(fabric.status)} label={fabric.status} />
              </div>

              {/* Metrics grid */}
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "var(--space-4)", marginBottom: "var(--space-4)" }}>
                {[
                  { label: "Throughput", value: fabric.metrics?.operationsPerSecond ? `${fabric.metrics.operationsPerSecond.toFixed(1)} ops/sec` : "N/A" },
                  { label: "Latency", value: fabric.metrics?.averageLatency ? `${fabric.metrics.averageLatency.toFixed(0)}ms` : "N/A" },
                  { label: "Stores", value: String(fabric.stores?.length || 0) },
                  { label: "Pipelines", value: String(fabric.pipelines?.length || 0) },
                ].map((m) => (
                  <div key={m.label}>
                    <p style={{ fontFamily: "var(--font-mono)", fontSize: 10, textTransform: "uppercase", letterSpacing: "0.06em", color: "var(--text-faint)", marginBottom: "var(--space-1)" }}>{m.label}</p>
                    <p style={{ fontFamily: "var(--font-mono)", fontSize: 16, fontWeight: 500, color: "var(--text)" }}>{m.value}</p>
                  </div>
                ))}
              </div>

              {/* Footer */}
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", paddingTop: "var(--space-4)", borderTop: "1px solid var(--panel-edge)" }}>
                <p style={{ fontSize: 11, color: "var(--text-faint)" }}>
                  Updated {new Date(fabric.updatedAt).toLocaleDateString()}
                </p>
                <div style={{ display: "flex", gap: "var(--space-2)" }}>
                  <FrameButton
                    size="sm"
                    onClick={(e) => { e.stopPropagation(); navigate(`/state-fabric/${fabric.id}/edit`); }}
                    iconLeft={<Settings style={{ width: 12, height: 12 }} />}
                  >
                    Configure
                  </FrameButton>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setFabricToDelete(fabric);
                      setDeleteDialogOpen(true);
                    }}
                    style={{
                      display: "flex", alignItems: "center", justifyContent: "center",
                      width: 28, height: 28, borderRadius: "var(--radius)",
                      background: "transparent", border: "1px solid var(--steel)",
                      cursor: "pointer", color: "var(--status-revoked)",
                      transition: "border-color var(--duration-fast) var(--ease-out)",
                    }}
                    title="Delete"
                  >
                    <Database style={{ width: 12, height: 12 }} />
                  </button>
                </div>
              </div>
            </Chamber>
          ))}
        </div>
      ) : (
        <Chamber>
          <div style={{ textAlign: "center", padding: "var(--space-8) 0" }}>
            <div style={{ width: 64, height: 64, margin: "0 auto var(--space-4)", borderRadius: "50%", background: "var(--panel-raised)", display: "flex", alignItems: "center", justifyContent: "center" }}>
              <Database style={{ width: 32, height: 32, color: "var(--text-faint)" }} />
            </div>
            <h3 style={{ fontFamily: "var(--font-display)", fontSize: 18, fontWeight: 500, color: "var(--text)", marginBottom: "var(--space-2)" }}>
              No state fabrics yet
            </h3>
            <p style={{ color: "var(--text-dim)", marginBottom: "var(--space-6)" }}>
              {searchQuery
                ? "No fabrics match your search."
                : "Create your first state fabric to get started with data orchestration."}
            </p>
            <SealedButton onClick={() => navigate("/state-fabric/new")} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>
              Create State Fabric
            </SealedButton>
          </div>
        </Chamber>
      )}

      {/* Delete Confirmation Modal */}
      <Modal open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)} title="Delete State Fabric">
        <div style={{ display: "flex", alignItems: "flex-start", gap: "var(--space-3)", marginBottom: "var(--space-4)" }}>
          <AlertTriangle style={{ width: 18, height: 18, color: "var(--status-revoked)", flexShrink: 0, marginTop: 2 }} />
          <p style={{ fontSize: 14, color: "var(--text-dim)", lineHeight: 1.6 }}>
            Are you sure you want to delete "{fabricToDelete?.name}"? This action cannot be undone.
            All associated stores and pipelines will be permanently removed.
          </p>
        </div>
        <div style={{ display: "flex", justifyContent: "flex-end", gap: "var(--space-3)" }}>
          <FrameButton onClick={() => setDeleteDialogOpen(false)}>Cancel</FrameButton>
          <SealedButton
            onClick={handleConfirmDelete}
            disabled={deleteFabric.isPending}
            style={{ background: "var(--status-revoked)", color: "#fff" }}
          >
            {deleteFabric.isPending ? "Deleting..." : "Delete"}
          </SealedButton>
        </div>
      </Modal>
    </div>
  );
}

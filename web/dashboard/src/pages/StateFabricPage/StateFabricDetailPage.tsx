import { useState } from "react";
import { Link, useParams, useNavigate, useLocation } from "react-router-dom";
import {
  ArrowLeft,
  Database,
  Network,
  Activity,
  Clock,
  Settings,
  Play,
  Pause,
  RefreshCw,
  Trash2,
  Plus,
  History,
  Camera,
  RotateCcw,
  User,
  Package,
  Zap,
  Workflow,
  Sparkles,
  Info,
  Check,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { TriggerConfiguration } from "./components/TriggerConfiguration";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  AnnotationTag,
  StatusPill,
} from "@/components/containment";
import {
  useStateFabrics,
  useStateFabric,
  useStateFabricMetrics,
  useStateFabricStores,
  useStateFabricPipelines,
  useStateFabricEventLogs,
  useStateFabricSnapshots,
  useDeleteStateFabric,
  useCreateStateFabric,
  useUpdateStateFabric,
} from "@/hooks/useStateFabric";
import { usePlan } from "@/hooks/usePlan";
import { usePageTitle } from "@/hooks/usePageTitle";
import {
  canCreateStateFabric,
  getStateFabricsLimit,
  hasFeature,
} from "@/lib/plan-utils";
import { ROUTES } from "@/lib/constants";
import { EventLogViewer } from "./components/EventLogViewer";
import { PipelineVisualization } from "./components/PipelineVisualization";
import { StoreConfiguration } from "./components/StoreConfiguration";
import { StateFabricMetrics as MetricsDashboard } from "./components/StateFabricMetrics";
import { SnapshotManager } from "./components/SnapshotManager";
import { StateFabricAddonGate } from "./components/StateFabricAddonGate";
import type {
  CreateStateFabricRequest,
  UpdateStateFabricRequest,
  StateFabric,
  StateFabricSettings,
} from "@/types";

const defaultSettings = (): StateFabricSettings => ({
  autoSnapshot: false,
  snapshotIntervalMinutes: 60,
  retentionDays: 30,
  enableReplication: false,
  regions: [],
  conflictResolution: "last-write-wins",
});

export function StateFabricDetailPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { plan } = usePlan();
  const { data: fabricsList } = useStateFabrics();
  const [activeTab, setActiveTab] = useState("overview");
  // /state-fabric/new has no :id param (route is literal "new"), so id can be undefined
  const isNew = !id || id === "new";
  const isEditPage = location.pathname.endsWith("/edit");

  const fabricCount = fabricsList?.length ?? 0;
  const allowCreateFabric = canCreateStateFabric(plan, fabricCount);
  const stateFabricUnlocked = hasFeature(plan, "STATE_FABRIC");
  const fabricLimit = getStateFabricsLimit(plan);

  const { data: fabric, isLoading: fabricLoading, error } = useStateFabric(isNew ? "" : id!);
  const fabricName = (fabric as StateFabric)?.name;
  usePageTitle(isNew ? 'State Fabric / New' : fabricName ? `State Fabric / ${fabricName}` : 'State Fabric');
  const { data: metrics } = useStateFabricMetrics(id || "");
  const { data: stores } = useStateFabricStores(id || "");
  const { data: pipelines } = useStateFabricPipelines(id || "");
  const { data: eventLogs } = useStateFabricEventLogs(id || "", { limit: 100 });
  const { data: snapshots } = useStateFabricSnapshots(id || "");
  const deleteFabric = useDeleteStateFabric();
  const createFabric = useCreateStateFabric();
  const updateFabric = useUpdateStateFabric(id || "");

  // Create new state fabric form (id === "new" — no API fetches run)
  if (isNew && !allowCreateFabric) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => navigate("/state-fabric")}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to State Fabrics
        </Button>
        <Card className="border-dashed border-white/20 rgba(14,19,24,0.5)/50">
          <CardHeader>
            <CardTitle className="text-[var(--text)]">
              {!stateFabricUnlocked
                ? "State Fabric is not on your plan"
                : "State fabric limit reached"}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-[var(--text-dim)] text-sm">
            {!stateFabricUnlocked ? (
              <p>
                State Fabric is available on Starter, Professional, and Enterprise plans.
              </p>
            ) : (
              <p>
                Your plan allows {fabricLimit >= 10000 ? "unlimited" : fabricLimit} state fabric
                {fabricLimit === 1 ? "" : "s"} (you have {fabricCount}).
              </p>
            )}
            <Button asChild variant="default">
              <Link to={ROUTES.PRICING}>View plans</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isNew) {
    return (
      <StateFabricCreateForm
        onCancel={() => navigate("/state-fabric")}
        onCreate={async (data) => {
          const created = await createFabric.mutateAsync(data);
          if (created?.id) {
            navigate(`/state-fabric/${created.id}`);
          } else {
            navigate("/state-fabric");
          }
        }}
        isSubmitting={createFabric.isPending}
      />
    );
  }

  const handleDelete = async () => {
    if (!id) return;
    if (confirm("Are you sure you want to delete this state fabric? This action cannot be undone.")) {
      await deleteFabric.mutateAsync(id);
      navigate("/state-fabric");
    }
  };

  if (fabricLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (error || !fabric) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => navigate("/state-fabric")}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to State Fabrics
        </Button>
        <Card className="p-12 text-center">
          <div className="text-[var(--status-revoked)] mb-4">Failed to load state fabric</div>
          <Button onClick={() => window.location.reload()} variant="outline">
            Retry
          </Button>
        </Card>
      </div>
    );
  }

  // Dynamic edit page: /state-fabric/:id/edit
  if (isEditPage) {
    return (
      <StateFabricEditForm
        fabric={fabric}
        onSave={async (data) => {
          await updateFabric.mutateAsync(data);
          navigate(`/state-fabric/${id}`);
        }}
        onCancel={() => navigate(`/state-fabric/${id}`)}
        isSubmitting={updateFabric.isPending}
      />
    );
  }

  const detailTabs = [
    { value: 'overview', label: 'Overview', icon: Activity },
    { value: 'stores', label: 'Stores', icon: Database },
    { value: 'pipelines', label: 'Pipelines', icon: Network },
    { value: 'events', label: 'Events', icon: History },
    { value: 'snapshots', label: 'Snapshots', icon: Camera },
    { value: 'triggers', label: 'Triggers', icon: Zap },
    { value: 'metrics', label: 'Metrics', icon: Activity },
  ];

  return (
    <div className="space-y-6">
      <PageGrid />

      {/* Hero Header */}
      <Chamber ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE SF-02" secondary="State Fabric" position="top-right" />

        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4" style={{ paddingTop: 'var(--space-3)', paddingRight: '140px' }}>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-3">
              <TrustSeal size="lg" />
              <div>
                <h1 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)', color: 'var(--text)' }}>
                  {fabric.name}
                </h1>
                <p style={{ color: 'var(--text-dim)' }}>{fabric.description}</p>
              </div>
            </div>
            <StatusPill status={fabric.status === 'online' ? 'live' : fabric.status === 'degraded' ? 'pending' : 'revoked'} />
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => navigate(`/state-fabric/${id}/edit`)}>
              <Settings className="w-4 h-4 mr-2" />
              Configure
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              <Trash2 className="w-4 h-4 mr-2" />
              Delete
            </Button>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-6">
          {[
            { label: 'Throughput', value: metrics?.operationsPerSecond ? `${metrics.operationsPerSecond.toFixed(1)}` : 'N/A', sub: 'ops/sec', icon: Activity },
            { label: 'Latency', value: metrics?.averageLatency ? `${metrics.averageLatency.toFixed(0)}` : 'N/A', sub: 'ms avg', icon: Clock },
            { label: 'Stores', value: String(stores?.length || 0), sub: 'active stores', icon: Database },
            { label: 'Pipelines', value: String(pipelines?.length || 0), sub: 'configured', icon: Network },
          ].map((stat) => (
            <div key={stat.label} className="p-4 rounded-[var(--radius)]" style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}>
              <div className="flex items-center gap-2 mb-2" style={{ color: 'var(--text-faint)' }}>
                <stat.icon className="w-4 h-4" />
                <span className="text-sm">{stat.label}</span>
              </div>
              <p className="text-2xl font-bold font-mono tabular-nums" style={{ color: 'var(--text)' }}>{stat.value}</p>
              <p className="text-xs" style={{ color: 'var(--text-faint)' }}>{stat.sub}</p>
            </div>
          ))}
        </div>
      </Chamber>

      {/* Tabs */}
      <div>
        <div
          className="inline-flex gap-1 p-1 rounded-[var(--radius)] overflow-x-auto"
          style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}
          role="tablist"
        >
          {detailTabs.map(({ value, label, icon: Icon }) => {
            const isActive = activeTab === value;
            return (
              <button
                key={value}
                role="tab"
                aria-selected={isActive}
                onClick={() => setActiveTab(value)}
                className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-[var(--radius-sm)] transition-all whitespace-nowrap"
                style={{
                  background: isActive ? 'var(--panel-raised)' : 'transparent',
                  color: isActive ? 'var(--text)' : 'var(--text-faint)',
                  boxShadow: isActive ? '0 1px 3px rgba(0,0,0,0.2)' : 'none',
                  fontFamily: 'var(--font-body)',
                }}
              >
                <Icon className="w-3.5 h-3.5" />
                {label}
              </button>
            );
          })}
        </div>

        <div className="mt-6 space-y-6">
          {activeTab === 'overview' && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Chamber nested>
                <CardHeader>
                  <CardTitle className="text-lg" style={{ fontFamily: 'var(--font-display)' }}>Configuration</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-4">
                    {[
                      ['Type', fabric.type],
                      ['Auto Snapshot', fabric.settings?.autoSnapshot ? 'Enabled' : 'Disabled'],
                      ['Snapshot Interval', `${fabric.settings?.snapshotIntervalMinutes || 60} minutes`],
                      ['Retention', `${fabric.settings?.retentionDays || 30} days`],
                      ['Replication', fabric.settings?.enableReplication ? 'Enabled' : 'Disabled'],
                      ['Conflict Resolution', (fabric.settings?.conflictResolution || 'last-write-wins').replace(/-/g, ' ')],
                    ].map(([label, value]) => (
                      <div key={label}>
                        <p className="text-sm" style={{ color: 'var(--text-faint)' }}>{label}</p>
                        <p className="font-medium capitalize" style={{ color: 'var(--text)' }}>{value}</p>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Chamber>

              <Chamber nested>
                <CardHeader>
                  <CardTitle className="text-lg" style={{ fontFamily: 'var(--font-display)' }}>Regions</CardTitle>
                </CardHeader>
                <CardContent>
                  {fabric.settings?.regions && fabric.settings.regions.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {fabric.settings.regions.map((region) => (
                        <span key={region} className="text-xs px-2.5 py-0.5 rounded-[var(--radius-sm)]" style={{ background: 'var(--panel)', color: 'var(--text-dim)', border: '1px solid var(--panel-edge)' }}>
                          {region}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <p style={{ color: 'var(--text-faint)' }}>No regions configured</p>
                  )}
                </CardContent>
              </Chamber>

              <Chamber nested className="lg:col-span-2">
                <CardHeader>
                  <CardTitle className="text-lg" style={{ fontFamily: 'var(--font-display)' }}>Recent Activity</CardTitle>
                </CardHeader>
                <CardContent>
                  {eventLogs?.events && eventLogs.events.length > 0 ? (
                    <div className="space-y-2">
                      {eventLogs.events.slice(0, 5).map((event) => (
                        <div
                          key={event.id}
                          className="flex items-center justify-between py-2"
                          style={{ borderBottom: '1px solid var(--panel-edge)' }}
                        >
                          <div className="flex items-center gap-3">
                            <span className="text-xs px-2 py-0.5 rounded-[var(--radius-sm)] capitalize" style={{ background: 'var(--panel)', color: 'var(--text-dim)', border: '1px solid var(--panel-edge)' }}>
                              {event.eventType}
                            </span>
                            <span className="text-sm" style={{ color: 'var(--text-dim)' }}>
                              {event.correlationId?.slice(0, 8) || 'System'}
                            </span>
                          </div>
                          <span className="text-xs" style={{ color: 'var(--text-faint)' }}>
                            {new Date(event.timestamp).toLocaleString()}
                          </span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p style={{ color: 'var(--text-faint)' }}>No recent activity</p>
                  )}
                </CardContent>
              </Chamber>
            </div>
          )}

          {activeTab === 'stores' && <StoreConfiguration fabricId={id || ''} stores={stores || []} />}
          {activeTab === 'pipelines' && <PipelineVisualization fabricId={id || ''} pipelines={pipelines || []} />}
          {activeTab === 'events' && (
            <StateFabricAddonGate addonId="advanced_security_pack">
              <EventLogViewer fabricId={id || ''} events={eventLogs?.events || []} total={eventLogs?.total || 0} />
            </StateFabricAddonGate>
          )}
          {activeTab === 'snapshots' && <SnapshotManager fabricId={id || ''} snapshots={snapshots || []} />}
          {activeTab === 'triggers' && <TriggerConfiguration fabricId={id || ''} />}
          {activeTab === 'metrics' && (
            <StateFabricAddonGate addonId="advanced_insights">
              <MetricsDashboard fabricId={id || ''} metrics={metrics} />
            </StateFabricAddonGate>
          )}
        </div>
      </div>
    </div>
  );
}

interface FabricTypeConfig {
  value: CreateStateFabricRequest["type"];
  label: string;
  description: string;
  icon: React.ReactNode;
  features: string[];
  iconBg: string;
  defaultSettings: StateFabricSettings;
}

const FABRIC_TYPES: FabricTypeConfig[] = [
  {
    value: "session",
    label: "Session Store",
    description: "Manage user sessions with automatic expiry and distributed replication",
    icon: <User className="w-6 h-6" />,
    features: ["Auto-expiry", "Distributed", "High availability"],
    iconBg: "rgba(159,216,255,0.1)",
    defaultSettings: {
      autoSnapshot: false,
      snapshotIntervalMinutes: 30,
      retentionDays: 7,
      enableReplication: true,
      regions: [],
      conflictResolution: "last-write-wins",
    },
  },
  {
    value: "catalog",
    label: "Data Catalog",
    description: "Centralized metadata management for your data assets",
    icon: <Package className="w-6 h-6" />,
    features: ["Schema registry", "Data lineage", "Search & discovery"],
    iconBg: "rgba(143,255,208,0.1)",
    defaultSettings: {
      autoSnapshot: true,
      snapshotIntervalMinutes: 120,
      retentionDays: 90,
      enableReplication: false,
      regions: [],
      conflictResolution: "last-write-wins",
    },
  },
  {
    value: "cache",
    label: "Cache Layer",
    description: "High-performance distributed caching with TTL support",
    icon: <Zap className="w-6 h-6" />,
    features: ["Sub-millisecond latency", "TTL support", "LRU eviction"],
    iconBg: "rgba(232,196,104,0.1)",
    defaultSettings: {
      autoSnapshot: false,
      snapshotIntervalMinutes: 60,
      retentionDays: 3,
      enableReplication: true,
      regions: [],
      conflictResolution: "last-write-wins",
    },
  },
  {
    value: "workflow",
    label: "Workflow Engine",
    description: "Orchestrate complex business processes with state management",
    icon: <Workflow className="w-6 h-6" />,
    features: ["State machines", "Event-driven", "Retry logic"],
    iconBg: "rgba(217,196,255,0.1)",
    defaultSettings: {
      autoSnapshot: true,
      snapshotIntervalMinutes: 60,
      retentionDays: 90,
      enableReplication: false,
      regions: [],
      conflictResolution: "manual",
    },
  },
  {
    value: "custom",
    label: "Custom Fabric",
    description: "Build your own state fabric with full customization",
    icon: <Sparkles className="w-6 h-6" />,
    features: ["Fully customizable", "Multi-store", "Custom pipelines"],
    iconBg: "rgba(74,86,95,0.15)",
    defaultSettings: {
      autoSnapshot: false,
      snapshotIntervalMinutes: 60,
      retentionDays: 30,
      enableReplication: false,
      regions: [],
      conflictResolution: "last-write-wins",
    },
  },
];

function StateFabricCreateForm({
  onCancel,
  onCreate,
  isSubmitting,
}: {
  onCancel: () => void;
  onCreate: (data: CreateStateFabricRequest) => Promise<void>;
  isSubmitting: boolean;
}) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<CreateStateFabricRequest["type"]>("session");
  const [errors, setErrors] = useState<{ name?: string }>({});
  const [settings, setSettings] = useState<StateFabricSettings>(
    FABRIC_TYPES[0].defaultSettings
  );
  const [showSettings, setShowSettings] = useState(false);

  const selectedType = FABRIC_TYPES.find((t) => t.value === type)!;

  const handleTypeChange = (newType: CreateStateFabricRequest["type"]) => {
    setType(newType);
    const typeConfig = FABRIC_TYPES.find((t) => t.value === newType);
    if (typeConfig) {
      setSettings(typeConfig.defaultSettings);
    }
  };

  const validateForm = (): boolean => {
    const newErrors: { name?: string } = {};
    if (!name.trim()) {
      newErrors.name = "Name is required";
    } else if (name.length < 3) {
      newErrors.name = "Name must be at least 3 characters";
    } else if (name.length > 50) {
      newErrors.name = "Name must be less than 50 characters";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    try {
      await onCreate({
        name: name.trim(),
        description: description.trim(),
        type,
        settings: showSettings ? settings : undefined,
      });
    } catch (err: any) {
      const message = err?.message || "Failed to create state fabric";
      toast.error(message);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <PageGrid />

      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={onCancel}
          className="shrink-0"
          disabled={isSubmitting}
        >
          <ArrowLeft className="w-5 h-5" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)', color: 'var(--text)' }}>Create State Fabric</h1>
          <p style={{ color: 'var(--text-dim)' }}>
            Configure a new state fabric for your application
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Type Selection */}
        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="MODULE SF-01" secondary="Fabric Type" position="top-right" />

          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
              <Database className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
              Choose Fabric Type
            </CardTitle>
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              Select the type that best fits your use case
            </p>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {FABRIC_TYPES.map((fabricType) => {
                const isSelected = type === fabricType.value;
                return (
                  <button
                    key={fabricType.value}
                    type="button"
                    onClick={() => handleTypeChange(fabricType.value)}
                    disabled={isSubmitting}
                    className="relative p-4 rounded-[var(--radius)] text-left transition-all duration-200"
                    style={{
                      background: isSelected ? 'rgba(143,255,208,0.03)' : 'var(--panel-raised)',
                      border: isSelected ? '2px solid var(--status-ok)' : '2px solid var(--panel-edge)',
                      opacity: isSubmitting ? 0.5 : 1,
                      cursor: isSubmitting ? 'not-allowed' : 'pointer',
                    }}
                  >
                    {isSelected && (
                      <div className="absolute top-3 right-3 w-5 h-5 rounded-full flex items-center justify-center" style={{ background: 'var(--status-ok)' }}>
                        <Check className="w-3 h-3" style={{ color: 'var(--bg)' }} />
                      </div>
                    )}

                    <div
                      className="w-12 h-12 rounded-[var(--radius)] flex items-center justify-center mb-3"
                      style={{ background: fabricType.iconBg, color: isSelected ? 'var(--status-ok)' : 'var(--text-dim)' }}
                    >
                      {fabricType.icon}
                    </div>

                    <h3 className="font-semibold mb-1" style={{ color: 'var(--text)' }}>{fabricType.label}</h3>

                    <p className="text-xs mb-3 line-clamp-2" style={{ color: 'var(--text-dim)' }}>
                      {fabricType.description}
                    </p>

                    <div className="flex flex-wrap gap-1">
                      {fabricType.features.map((feature) => (
                        <span
                          key={feature}
                          className="text-[10px] px-2 py-0.5 rounded-[var(--radius-sm)]"
                          style={{ background: 'var(--panel)', color: 'var(--text-faint)', border: '1px solid var(--panel-edge)' }}
                        >
                          {feature}
                        </span>
                      ))}
                    </div>
                  </button>
                );
              })}
            </div>
          </CardContent>
        </Chamber>

        {/* Configuration */}
        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />

          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
              <Settings className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
              Configuration
            </CardTitle>
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              Basic settings for your state fabric
            </p>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name" style={{ color: 'var(--text)' }}>
                Name <span style={{ color: 'var(--status-revoked)' }}>*</span>
              </Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (errors.name) setErrors({});
                }}
                placeholder={`e.g., My ${selectedType.label}`}
                required
                disabled={isSubmitting}
                style={errors.name ? { borderColor: 'var(--status-revoked)' } : undefined}
              />
              {errors.name ? (
                <p className="text-sm" style={{ color: 'var(--status-revoked)' }}>{errors.name}</p>
              ) : (
                <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                  A unique name to identify this state fabric (3-50 characters)
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description" style={{ color: 'var(--text)' }}>Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe the purpose of this state fabric..."
                disabled={isSubmitting}
                rows={3}
                className="resize-none"
              />
              <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                Optional description to help team members understand its purpose
              </p>
            </div>
          </CardContent>
        </Chamber>

        {/* Advanced Settings (collapsible) */}
        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />

          <CardHeader>
            <button
              type="button"
              onClick={() => setShowSettings(!showSettings)}
              className="flex items-center justify-between w-full text-left"
            >
              <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                <Settings className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
                Advanced Settings
              </CardTitle>
              <span className="text-xs" style={{ color: 'var(--text-faint)' }}>
                {showSettings ? "Hide" : "Show"} &mdash; defaults for {selectedType.label}
              </span>
            </button>
          </CardHeader>
          {showSettings && (
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-lg border border-[var(--panel-edge)] p-4">
                <div>
                  <Label style={{ color: 'var(--text)' }}>Auto snapshot</Label>
                  <p className="text-xs" style={{ color: 'var(--text-faint)' }}>Create snapshots automatically</p>
                </div>
                <Switch
                  checked={settings.autoSnapshot}
                  onCheckedChange={(v) => setSettings((s) => ({ ...s, autoSnapshot: v }))}
                  disabled={isSubmitting}
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="snapshotInterval" style={{ color: 'var(--text)' }}>Snapshot interval (min)</Label>
                  <Input
                    id="snapshotInterval"
                    type="number"
                    min={1}
                    max={10080}
                    value={settings.snapshotIntervalMinutes}
                    onChange={(e) =>
                      setSettings((s) => ({
                        ...s,
                        snapshotIntervalMinutes: Math.max(1, parseInt(e.target.value, 10) || 1),
                      }))
                    }
                    disabled={isSubmitting}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="retentionDays" style={{ color: 'var(--text)' }}>Retention (days)</Label>
                  <Input
                    id="retentionDays"
                    type="number"
                    min={1}
                    max={3650}
                    value={settings.retentionDays}
                    onChange={(e) =>
                      setSettings((s) => ({
                        ...s,
                        retentionDays: Math.max(1, parseInt(e.target.value, 10) || 1),
                      }))
                    }
                    disabled={isSubmitting}
                  />
                </div>
              </div>
              <div className="flex items-center justify-between rounded-lg border border-[var(--panel-edge)] p-4">
                <div>
                  <Label style={{ color: 'var(--text)' }}>Enable replication</Label>
                  <p className="text-xs" style={{ color: 'var(--text-faint)' }}>Replicate across regions</p>
                </div>
                <Switch
                  checked={settings.enableReplication}
                  onCheckedChange={(v) => setSettings((s) => ({ ...s, enableReplication: v }))}
                  disabled={isSubmitting}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="conflictResolution" style={{ color: 'var(--text)' }}>Conflict resolution</Label>
                <Select
                  value={settings.conflictResolution}
                  onValueChange={(v) =>
                    setSettings((s) => ({
                      ...s,
                      conflictResolution: v as StateFabricSettings["conflictResolution"],
                    }))
                  }
                  disabled={isSubmitting}
                >
                  <SelectTrigger id="conflictResolution">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="last-write-wins">Last write wins</SelectItem>
                    <SelectItem value="first-write-wins">First write wins</SelectItem>
                    <SelectItem value="manual">Manual</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          )}
        </Chamber>

        {/* Actions */}
        <div className="flex items-center justify-between pt-4">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={!name.trim() || isSubmitting}
            className="gap-2"
          >
            {isSubmitting ? (
              <>
                <LoadingSpinner size="sm" />
                Creating...
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Create State Fabric
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}

function StateFabricEditForm({
  fabric,
  onSave,
  onCancel,
  isSubmitting,
}: {
  fabric: StateFabric;
  onSave: (data: UpdateStateFabricRequest) => Promise<void>;
  onCancel: () => void;
  isSubmitting: boolean;
}) {
  const [name, setName] = useState(fabric.name);
  const [description, setDescription] = useState(fabric.description ?? "");
  const [settings, setSettings] = useState<StateFabricSettings>(() => ({
    ...defaultSettings(),
    ...fabric.settings,
  }));
  const [errors, setErrors] = useState<{ name?: string }>({});

  const validateForm = (): boolean => {
    const newErrors: { name?: string } = {};
    if (!name.trim()) {
      newErrors.name = "Name is required";
    } else if (name.length < 3) {
      newErrors.name = "Name must be at least 3 characters";
    } else if (name.length > 50) {
      newErrors.name = "Name must be less than 50 characters";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    onSave({
      name: name.trim(),
      description: description.trim(),
      settings: {
        autoSnapshot: settings.autoSnapshot,
        snapshotIntervalMinutes: settings.snapshotIntervalMinutes,
        retentionDays: settings.retentionDays,
        enableReplication: settings.enableReplication,
        regions: settings.regions,
        conflictResolution: settings.conflictResolution,
      },
    });
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={onCancel}
          className="shrink-0"
          disabled={isSubmitting}
        >
          <ArrowLeft className="w-5 h-5" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold text-[var(--text)]">Edit State Fabric</h1>
          <p className="text-[var(--text-dim)]">
            Update name, description, and settings for this fabric
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Settings className="w-5 h-5 text-[var(--status-ok)]" />
              Basic info
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">
                Name <span className="text-[var(--status-revoked)]">*</span>
              </Label>
              <Input
                id="edit-name"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (errors.name) setErrors({});
                }}
                placeholder="Fabric name"
                disabled={isSubmitting}
                className={errors.name ? "border-[var(--status-revoked)] focus-visible:shadow-[0_0_0_1px_var(--status-revoked)_inset]" : ""}
              />
              {errors.name && (
                <p className="text-sm text-[var(--status-revoked)]">{errors.name}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-description">Description</Label>
              <Textarea
                id="edit-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description"
                disabled={isSubmitting}
                rows={3}
                className="resize-none"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Database className="w-5 h-5 text-[var(--status-ok)]" />
              Settings
            </CardTitle>
            <p className="text-sm text-[var(--text-dim)]">
              Snapshot, retention, and replication options
            </p>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-lg border border-[var(--panel-edge)] p-4">
              <div>
                <Label htmlFor="edit-autoSnapshot">Auto snapshot</Label>
                <p className="text-xs text-[var(--text-faint)]">Create snapshots automatically</p>
              </div>
              <Switch
                id="edit-autoSnapshot"
                checked={settings.autoSnapshot}
                onCheckedChange={(v) =>
                  setSettings((s) => ({ ...s, autoSnapshot: v }))
                }
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-snapshotInterval">Snapshot interval (minutes)</Label>
              <Input
                id="edit-snapshotInterval"
                type="number"
                min={1}
                max={10080}
                value={settings.snapshotIntervalMinutes}
                onChange={(e) =>
                  setSettings((s) => ({
                    ...s,
                    snapshotIntervalMinutes: Math.max(1, parseInt(e.target.value, 10) || 1),
                  }))
                }
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-retentionDays">Retention (days)</Label>
              <Input
                id="edit-retentionDays"
                type="number"
                min={1}
                max={3650}
                value={settings.retentionDays}
                onChange={(e) =>
                  setSettings((s) => ({
                    ...s,
                    retentionDays: Math.max(1, parseInt(e.target.value, 10) || 1),
                  }))
                }
                disabled={isSubmitting}
              />
            </div>
            <div className="flex items-center justify-between rounded-lg border border-[var(--panel-edge)] p-4">
              <div>
                <Label htmlFor="edit-enableReplication">Enable replication</Label>
                <p className="text-xs text-[var(--text-faint)]">Replicate across regions</p>
              </div>
              <Switch
                id="edit-enableReplication"
                checked={settings.enableReplication}
                onCheckedChange={(v) =>
                  setSettings((s) => ({ ...s, enableReplication: v }))
                }
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-conflictResolution">Conflict resolution</Label>
              <Select
                value={settings.conflictResolution}
                onValueChange={(v) =>
                  setSettings((s) => ({
                    ...s,
                    conflictResolution: v as StateFabricSettings["conflictResolution"],
                  }))
                }
                disabled={isSubmitting}
              >
                <SelectTrigger id="edit-conflictResolution">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="last-write-wins">Last write wins</SelectItem>
                  <SelectItem value="first-write-wins">First write wins</SelectItem>
                  <SelectItem value="manual">Manual</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <div className="flex items-center justify-between pt-4">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={!name.trim() || isSubmitting} className="gap-2">
            {isSubmitting ? (
              <>
                <LoadingSpinner size="sm" />
                Saving...
              </>
            ) : (
              <>
                <Check className="w-4 h-4" />
                Save changes
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}

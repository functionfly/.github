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
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { StatusBadge } from "@/components/common/StatusBadge";
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
        <Card className="border-dashed border-white/20 bg-bg-secondary/50">
          <CardHeader>
            <CardTitle className="text-text-primary">
              {!stateFabricUnlocked
                ? "State Fabric is not on your plan"
                : "State fabric limit reached"}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-text-secondary text-sm">
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
          navigate(`/state-fabric/${created.id}`);
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
          <div className="text-red-400 mb-4">Failed to load state fabric</div>
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/state-fabric")}
            className="shrink-0"
            aria-label="Back to State Fabric"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-text-primary">
                {fabric.name}
              </h1>
              <StatusBadge status={fabric.status} />
            </div>
            <p className="text-text-secondary">{fabric.description}</p>
          </div>
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
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-text-muted mb-2">
              <Activity className="w-4 h-4" />
              <span className="text-sm">Throughput</span>
            </div>
            <p className="text-2xl font-bold text-text-primary">
              {metrics?.operationsPerSecond
                ? `${metrics.operationsPerSecond.toFixed(1)}`
                : "N/A"}
            </p>
            <p className="text-xs text-text-muted">ops/sec</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-text-muted mb-2">
              <Clock className="w-4 h-4" />
              <span className="text-sm">Latency</span>
            </div>
            <p className="text-2xl font-bold text-text-primary">
              {metrics?.averageLatency
                ? `${metrics.averageLatency.toFixed(0)}`
                : "N/A"}
            </p>
            <p className="text-xs text-text-muted">ms avg</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-text-muted mb-2">
              <Database className="w-4 h-4" />
              <span className="text-sm">Stores</span>
            </div>
            <p className="text-2xl font-bold text-text-primary">
              {stores?.length || 0}
            </p>
            <p className="text-xs text-text-muted">active stores</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-text-muted mb-2">
              <Network className="w-4 h-4" />
              <span className="text-sm">Pipelines</span>
            </div>
            <p className="text-2xl font-bold text-text-primary">
              {pipelines?.length || 0}
            </p>
            <p className="text-xs text-text-muted">configured</p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid grid-cols-3 md:grid-cols-6 lg:w-fit">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="stores">Stores</TabsTrigger>
          <TabsTrigger value="pipelines">Pipelines</TabsTrigger>
          <TabsTrigger value="events">Events</TabsTrigger>
          <TabsTrigger value="snapshots">Snapshots</TabsTrigger>
          <TabsTrigger value="triggers">Triggers</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Configuration */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Configuration</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-text-muted">Type</p>
                    <p className="font-medium text-text-primary capitalize">
                      {fabric.type}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-muted">Auto Snapshot</p>
                    <p className="font-medium text-text-primary">
                      {fabric.settings?.autoSnapshot ? "Enabled" : "Disabled"}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-muted">Snapshot Interval</p>
                    <p className="font-medium text-text-primary">
                      {fabric.settings?.snapshotIntervalMinutes || 60} minutes
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-muted">Retention</p>
                    <p className="font-medium text-text-primary">
                      {fabric.settings?.retentionDays || 30} days
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-muted">Replication</p>
                    <p className="font-medium text-text-primary">
                      {fabric.settings?.enableReplication ? "Enabled" : "Disabled"}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-muted">Conflict Resolution</p>
                    <p className="font-medium text-text-primary capitalize">
                      {(fabric.settings?.conflictResolution || "last-write-wins").replace(/-/g, " ")}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Regions */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Regions</CardTitle>
              </CardHeader>
              <CardContent>
                {fabric.settings?.regions && fabric.settings.regions.length > 0 ? (
                  <div className="flex flex-wrap gap-2">
                    {fabric.settings.regions.map((region) => (
                      <Badge key={region} variant="secondary">
                        {region}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <p className="text-text-muted">No regions configured</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Recent Activity */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Recent Activity</CardTitle>
            </CardHeader>
            <CardContent>
              {eventLogs?.events && eventLogs.events.length > 0 ? (
                <div className="space-y-2">
                  {eventLogs.events.slice(0, 5).map((event) => (
                    <div
                      key={event.id}
                      className="flex items-center justify-between py-2 border-b border-border-subtle last:border-0"
                    >
                      <div className="flex items-center gap-3">
                        <Badge variant="outline" className="capitalize">
                          {event.eventType}
                        </Badge>
                        <span className="text-sm text-text-secondary">
                          {event.correlationId?.slice(0, 8) || "System"}
                        </span>
                      </div>
                      <span className="text-xs text-text-muted">
                        {new Date(event.timestamp).toLocaleString()}
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-text-muted">No recent activity</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="stores">
          <StoreConfiguration fabricId={id || ""} stores={stores || []} />
        </TabsContent>

        <TabsContent value="pipelines">
          <PipelineVisualization fabricId={id || ""} pipelines={pipelines || []} />
        </TabsContent>

        <TabsContent value="events">
          <StateFabricAddonGate addonId="advanced_security_pack">
            <EventLogViewer fabricId={id || ""} events={eventLogs?.events || []} total={eventLogs?.total || 0} />
          </StateFabricAddonGate>
        </TabsContent>

        <TabsContent value="snapshots">
          <SnapshotManager fabricId={id || ""} snapshots={snapshots || []} />
        </TabsContent>

        <TabsContent value="triggers">
          <TriggerConfiguration fabricId={id || ""} />
        </TabsContent>

        <TabsContent value="metrics">
          <StateFabricAddonGate addonId="advanced_insights">
            <MetricsDashboard fabricId={id || ""} metrics={metrics} />
          </StateFabricAddonGate>
        </TabsContent>
      </Tabs>
    </div>
  );
}

interface FabricTypeConfig {
  value: CreateStateFabricRequest["type"];
  label: string;
  description: string;
  icon: React.ReactNode;
  features: string[];
  gradient: string;
}

const FABRIC_TYPES: FabricTypeConfig[] = [
  {
    value: "session",
    label: "Session Store",
    description: "Manage user sessions with automatic expiry and distributed replication",
    icon: <User className="w-6 h-6" />,
    features: ["Auto-expiry", "Distributed", "High availability"],
    gradient: "from-blue-500/20 via-blue-500/10 to-transparent",
  },
  {
    value: "catalog",
    label: "Data Catalog",
    description: "Centralized metadata management for your data assets",
    icon: <Package className="w-6 h-6" />,
    features: ["Schema registry", "Data lineage", "Search & discovery"],
    gradient: "from-green-500/20 via-green-500/10 to-transparent",
  },
  {
    value: "cache",
    label: "Cache Layer",
    description: "High-performance distributed caching with TTL support",
    icon: <Zap className="w-6 h-6" />,
    features: ["Sub-millisecond latency", "TTL support", "LRU eviction"],
    gradient: "from-yellow-500/20 via-yellow-500/10 to-transparent",
  },
  {
    value: "workflow",
    label: "Workflow Engine",
    description: "Orchestrate complex business processes with state management",
    icon: <Workflow className="w-6 h-6" />,
    features: ["State machines", "Event-driven", "Retry logic"],
    gradient: "from-purple-500/20 via-purple-500/10 to-transparent",
  },
  {
    value: "custom",
    label: "Custom Fabric",
    description: "Build your own state fabric with full customization",
    icon: <Sparkles className="w-6 h-6" />,
    features: ["Fully customizable", "Multi-store", "Custom pipelines"],
    gradient: "from-gray-500/20 via-gray-500/10 to-transparent",
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

  const selectedType = FABRIC_TYPES.find((t) => t.value === type)!;

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
    onCreate({ name: name.trim(), description: description.trim(), type });
  };

  return (
    <div className="space-y-6 max-w-4xl">
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
          <h1 className="text-2xl font-bold text-text-primary">Create State Fabric</h1>
          <p className="text-text-secondary">
            Configure a new state fabric for your application
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Type Selection */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Database className="w-5 h-5 text-brand-500" />
              Choose Fabric Type
            </CardTitle>
            <p className="text-sm text-text-secondary">
              Select the type that best fits your use case
            </p>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {FABRIC_TYPES.map((fabricType) => (
                <button
                  key={fabricType.value}
                  type="button"
                  onClick={() => setType(fabricType.value)}
                  disabled={isSubmitting}
                  className={`relative p-4 rounded-xl border-2 text-left transition-all duration-200 ${
                    type === fabricType.value
                      ? "border-brand-500 bg-brand-500/5"
                      : "border-border-subtle hover:border-border-default hover:bg-bg-hover"
                  } disabled:opacity-50 disabled:cursor-not-allowed`}
                >
                  {/* Selection indicator */}
                  {type === fabricType.value && (
                    <div className="absolute top-3 right-3 w-5 h-5 rounded-full bg-brand-500 flex items-center justify-center">
                      <Check className="w-3 h-3 text-white" />
                    </div>
                  )}

                  {/* Icon with gradient background */}
                  <div
                    className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 bg-gradient-to-br ${fabricType.gradient} ${
                      type === fabricType.value ? "text-brand-400" : "text-text-secondary"
                    }`}
                  >
                    {fabricType.icon}
                  </div>

                  {/* Label */}
                  <h3 className="font-semibold text-text-primary mb-1">{fabricType.label}</h3>

                  {/* Description */}
                  <p className="text-xs text-text-secondary mb-3 line-clamp-2">
                    {fabricType.description}
                  </p>

                  {/* Features */}
                  <div className="flex flex-wrap gap-1">
                    {fabricType.features.map((feature) => (
                      <span
                        key={feature}
                        className="text-[10px] px-2 py-0.5 rounded-full bg-bg-tertiary text-text-muted"
                      >
                        {feature}
                      </span>
                    ))}
                  </div>
                </button>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Configuration */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Settings className="w-5 h-5 text-brand-500" />
              Configuration
            </CardTitle>
            <p className="text-sm text-text-secondary">
              Basic settings for your state fabric
            </p>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Name Field */}
            <div className="space-y-2">
              <Label htmlFor="name" className="flex items-center gap-2">
                Name
                <span className="text-red-500">*</span>
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
                className={errors.name ? "border-red-500 focus-visible:ring-red-500" : ""}
              />
              {errors.name ? (
                <p className="text-sm text-red-500">{errors.name}</p>
              ) : (
                <p className="text-xs text-text-muted">
                  A unique name to identify this state fabric (3-50 characters)
                </p>
              )}
            </div>

            {/* Description Field */}
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe the purpose of this state fabric..."
                disabled={isSubmitting}
                rows={3}
                className="resize-none"
              />
              <p className="text-xs text-text-muted">
                Optional description to help team members understand its purpose
              </p>
            </div>

            {/* Selected Type Info */}
            <div className="p-4 rounded-lg bg-bg-tertiary border border-border-subtle">
              <div className="flex items-start gap-3">
                <Info className="w-5 h-5 text-brand-500 mt-0.5 shrink-0" />
                <div>
                  <h4 className="font-medium text-text-primary text-sm">
                    About {selectedType.label}
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    {selectedType.description}
                  </p>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {selectedType.features.map((feature) => (
                      <Badge key={feature} variant="secondary" className="text-xs">
                        {feature}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

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
          <h1 className="text-2xl font-bold text-text-primary">Edit State Fabric</h1>
          <p className="text-text-secondary">
            Update name, description, and settings for this fabric
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Settings className="w-5 h-5 text-brand-500" />
              Basic info
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">
                Name <span className="text-red-500">*</span>
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
                className={errors.name ? "border-red-500 focus-visible:ring-red-500" : ""}
              />
              {errors.name && (
                <p className="text-sm text-red-500">{errors.name}</p>
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
              <Database className="w-5 h-5 text-brand-500" />
              Settings
            </CardTitle>
            <p className="text-sm text-text-secondary">
              Snapshot, retention, and replication options
            </p>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-lg border border-border-subtle p-4">
              <div>
                <Label htmlFor="edit-autoSnapshot">Auto snapshot</Label>
                <p className="text-xs text-text-muted">Create snapshots automatically</p>
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
            <div className="flex items-center justify-between rounded-lg border border-border-subtle p-4">
              <div>
                <Label htmlFor="edit-enableReplication">Enable replication</Label>
                <p className="text-xs text-text-muted">Replicate across regions</p>
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

import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import {
  ArrowLeft,
  Activity,
  Globe,
  Clock,
  AlertTriangle,
  Play,
  Pause,
  Trash2,
  Eye,
  Edit,
} from "lucide-react";
import "@/styles/components.css";

interface FunctionData {
  id: string;
  name: string;
  author: string;
  authorId: string;
  title?: string;
  description: string;
  category?: string;
  tags?: any;
  status: "online" | "offline" | "degraded";
  isPublic: boolean;
  isFeatured: boolean;
  runtime: string;
  version: string;
  providers: string[];
  region: string;
  createdAt: string;
  updatedAt: string;
  lastDeployedAt?: string;
  pricePerCall?: number;
  popularityScore?: number;
  reliabilityScore?: number;
  deterministicScore?: number;
  capabilities?: any;
  embedConfig?: any;
  tenantId?: string;
  ownerUserId?: string;
  totalRatings?: number;
  overallScore?: number;
  isFlagged?: boolean;
  flagReason?: string | null;
  versions?: any[];
  stats: {
    totalInvocations: number;
    avgLatency: number;
    errorRate: number;
    uptime: number;
  };
  owner: {
    id: string;
    name: string;
    email: string;
  };
}

export function AdminFunctionDetailPage() {
  const { functionId } = useParams<{ functionId: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");
  const [isToggling, setIsToggling] = useState(false);

  const [funcData, setFuncData] = useState<FunctionData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch function data
  useEffect(() => {
    const fetchFunction = async () => {
      if (!functionId) return;

      try {
        setLoading(true);
        setError(null);

        const response = await apiClient.get<any>(`/v1/admin/registry/functions/${functionId}`);
        // Transform registry API response to expected format
        const registryData = response.function;
        const transformedData: FunctionData = {
          id: registryData.id,
          name: registryData.name,
          author: registryData.author,
          authorId: registryData.author, // Using author as authorId for now
          title: registryData.title,
          description: registryData.description || '',
          category: registryData.category,
          tags: registryData.tags,
          status: registryData.visibility === 'public' ? 'online' : 'offline', // Map visibility to status
          isPublic: registryData.visibility === 'public',
          isFeatured: false, // Not available in registry data
          runtime: response.versions?.[0]?.runtime || '',
          version: registryData.latest_version || '',
          providers: [], // TODO: Extract from deployment/backend info or manifest
          region: '', // TODO: Extract from backend info
          createdAt: registryData.created_at,
          updatedAt: registryData.updated_at,
          pricePerCall: registryData.price_per_call,
          popularityScore: registryData.popularity_score,
          reliabilityScore: registryData.reliability_score,
          deterministicScore: registryData.deterministic_score,
          capabilities: registryData.capabilities,
          embedConfig: registryData.embed_config,
          tenantId: registryData.tenant_id,
          ownerUserId: registryData.owner_user_id,
          totalRatings: registryData.total_ratings,
          overallScore: registryData.overall_score,
          isFlagged: registryData.is_flagged,
          flagReason: registryData.flag_reason,
          versions: response.versions,
          stats: {
            totalInvocations: 0, // Not available in registry data
            avgLatency: 0,
            errorRate: 0,
            uptime: 100,
          },
          owner: {
            id: registryData.author,
            name: registryData.author,
            email: '', // Not available
          },
        };
        setFuncData(transformedData);
      } catch (err) {
        console.error("Failed to load function:", err);
        setError("Failed to load function");
        toast.error("Failed to load function");
      } finally {
        setLoading(false);
      }
    };

    fetchFunction();
  }, [functionId]);

  const handleToggleStatus = async () => {
    if (!functionId || !funcData) return;

    setIsToggling(true);
    try {
      const newVisibility = funcData.status === "online" ? "private" : "public";
      await apiClient.patch(`/v1/admin/registry/functions/${functionId}/visibility`, { visibility: newVisibility });
      const newStatus = newVisibility === "public" ? "online" : "offline";
      setFuncData({ ...funcData, status: newStatus, isPublic: newVisibility === "public" });
      toast.success(`Function ${newStatus === "online" ? "enabled" : "disabled"} successfully`);
    } catch (error) {
      console.error("Failed to toggle function status:", error);
      toast.error("Failed to toggle function status");
    } finally {
      setIsToggling(false);
    }
  };

  const handleTogglePublic = async () => {
    if (!functionId || !funcData) return;

    try {
      const newVisibility = !funcData.isPublic ? "public" : "private";
      await apiClient.patch(`/v1/admin/registry/functions/${functionId}/visibility`, {
        visibility: newVisibility
      });
      const newStatus = newVisibility === "public" ? "online" : "offline";
      setFuncData({ ...funcData, isPublic: !funcData.isPublic, status: newStatus });
      toast.success(`Function is now ${!funcData.isPublic ? "public" : "private"}`);
    } catch (error) {
      console.error("Failed to toggle function visibility:", error);
      toast.error("Failed to toggle function visibility");
    }
  };

  const handleToggleFeatured = async () => {
    if (!functionId || !funcData) return;

    try {
      await apiClient.patch(`/v1/admin/registry/functions/${functionId}`, {
        is_featured: !funcData.isFeatured
      });
      setFuncData({ ...funcData, isFeatured: !funcData.isFeatured });
      toast.success(`Function ${!funcData.isFeatured ? "featured" : "unfeatured"} successfully`);
    } catch (error) {
      console.error("Failed to toggle featured status:", error);
      toast.error("Failed to toggle featured status");
    }
  };

  if (loading || !funcData) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 bg-muted rounded animate-pulse" />
          <div className="space-y-2">
            <div className="w-48 h-6 bg-muted rounded animate-pulse" />
            <div className="w-32 h-4 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="p-6 border rounded-lg">
          <div className="w-full h-64 bg-muted rounded animate-pulse" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/admin/registry/functions")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">
                {funcData.title || `${funcData.author}/${funcData.name}`}
              </h1>
              <StatusBadge status={funcData.status} />
              {funcData.isFeatured && (
                <Badge variant="default">Featured</Badge>
              )}
              {funcData.category && (
                <Badge variant="secondary">{funcData.category}</Badge>
              )}
            </div>
            <p className="text-text-secondary">
              {funcData.description || "No description"}
            </p>
            {funcData.title && (
              <p className="text-text-muted text-sm">
                {funcData.author}/{funcData.name}
              </p>
            )}
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" onClick={handleToggleStatus} disabled={isToggling}>
            {funcData.status === "online" ? (
              <>
                <Pause className="w-4 h-4 mr-2" />
                Disable
              </>
            ) : (
              <>
                <Play className="w-4 h-4 mr-2" />
                Enable
              </>
            )}
          </Button>
          <Button variant="outline" onClick={() => navigate(`/fx/${funcData.author}/${funcData.name}`)}>
            <Eye className="w-4 h-4 mr-2" />
            View Public Page
          </Button>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          title="Total Ratings"
          value={funcData.totalRatings?.toLocaleString() || '0'}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Overall Score"
          value={funcData.overallScore?.toFixed(1) || 'N/A'}
          icon={<Clock className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Price per Call"
          value={funcData.pricePerCall ? `$${funcData.pricePerCall.toFixed(4)}` : 'Free'}
          icon={<AlertTriangle className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Popularity Score"
          value={funcData.popularityScore?.toFixed(2) || 'N/A'}
          icon={<Globe className="w-5 h-5" />}
          trend="neutral"
        />
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="owner">Owner</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Function Details</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Runtime</Label>
                  <p className="text-text-primary mt-1">{funcData.runtime}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Version</Label>
                  <p className="text-text-primary mt-1">{funcData.version}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Category</Label>
                  <p className="text-text-primary mt-1">{funcData.category || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Region</Label>
                  <p className="text-text-primary mt-1">{funcData.region || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Providers</Label>
                  <p className="text-text-primary mt-1">{funcData.providers.join(", ") || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Reliability Score</Label>
                  <p className="text-text-primary mt-1">{funcData.reliabilityScore?.toFixed(2) || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Deterministic Score</Label>
                  <p className="text-text-primary mt-1">{funcData.deterministicScore?.toFixed(2) || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Flagged</Label>
                  <p className="text-text-primary mt-1">{funcData.isFlagged ? 'Yes' : 'No'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Tags</Label>
                  <p className="text-text-primary mt-1">{funcData.tags ? JSON.stringify(funcData.tags) : 'None'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Tenant ID</Label>
                  <p className="text-text-primary mt-1 font-mono text-sm">{funcData.tenantId || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Owner User ID</Label>
                  <p className="text-text-primary mt-1 font-mono text-sm">{funcData.ownerUserId || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Created</Label>
                  <p className="text-text-primary mt-1">{funcData.createdAt}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Last Deployed</Label>
                  <p className="text-text-primary mt-1">{funcData.lastDeployedAt || "Never"}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Visibility</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-medium text-text-primary text-base">Public Function</h2>
                  <p className="text-sm text-text-muted">
                    Make this function visible in the public registry
                  </p>
                </div>
                <Switch
                  checked={funcData.isPublic}
                  onCheckedChange={handleTogglePublic}
                />
              </div>

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-medium text-text-primary text-base">Featured</h2>
                  <p className="text-sm text-text-muted">
                    Feature this function on the homepage
                  </p>
                </div>
                <Switch
                  checked={funcData.isFeatured}
                  onCheckedChange={handleToggleFeatured}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Owner Tab */}
        <TabsContent value="owner" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Owner Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Name</Label>
                  <p className="text-text-primary mt-1">{funcData.owner.name}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Email</Label>
                  <p className="text-text-primary mt-1">{funcData.owner.email}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Owner ID</Label>
                  <p className="text-text-primary mt-1 font-mono text-sm">{funcData.owner.id}</p>
                </div>
              </div>
              <div className="flex gap-2 mt-4">
                <Button
                  variant="outline"
                  onClick={() => navigate(`/admin/users/${funcData.owner.id}`)}
                >
                  View Owner Profile
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function Label({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <label className={`text-sm font-medium ${className || ""}`}>
      {children}
    </label>
  );
}

export default AdminFunctionDetailPage;

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
  description: string;
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

        const response = await apiClient.get<FunctionData>(`/v1/admin/functions/${functionId}`);
        setFuncData(response);
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
      const newStatus = funcData.status === "online" ? "offline" : "online";
      await apiClient.patch(`/v1/admin/functions/${functionId}`, { status: newStatus });
      setFuncData({ ...funcData, status: newStatus });
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
      await apiClient.patch(`/v1/admin/functions/${functionId}`, { 
        isPublic: !funcData.isPublic 
      });
      setFuncData({ ...funcData, isPublic: !funcData.isPublic });
      toast.success(`Function is now ${!funcData.isPublic ? "public" : "private"}`);
    } catch (error) {
      console.error("Failed to toggle function visibility:", error);
      toast.error("Failed to toggle function visibility");
    }
  };

  const handleToggleFeatured = async () => {
    if (!functionId || !funcData) return;

    try {
      await apiClient.patch(`/v1/admin/functions/${functionId}`, { 
        isFeatured: !funcData.isFeatured 
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
            onClick={() => navigate("/admin/functions")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">
                {funcData.author}/{funcData.name}
              </h1>
              <StatusBadge status={funcData.status} />
              {funcData.isFeatured && (
                <Badge variant="default">Featured</Badge>
              )}
            </div>
            <p className="text-text-secondary">
              {funcData.description || "No description"}
            </p>
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
          <Button variant="outline" onClick={() => navigate(`/functions/${funcData.id}`)}>
            <Eye className="w-4 h-4 mr-2" />
            View
          </Button>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          title="Total Invocations"
          value={funcData.stats.totalInvocations.toLocaleString()}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Avg Latency"
          value={`${funcData.stats.avgLatency}ms`}
          icon={<Clock className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Error Rate"
          value={`${funcData.stats.errorRate}%`}
          icon={<AlertTriangle className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Uptime"
          value={`${funcData.stats.uptime}%`}
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
                  <Label className="text-text-muted">Providers</Label>
                  <p className="text-text-primary mt-1">{funcData.providers.join(", ")}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Region</Label>
                  <p className="text-text-primary mt-1">{funcData.region.toUpperCase()}</p>
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
                  <h4 className="font-medium text-text-primary">Public Function</h4>
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
                  <h4 className="font-medium text-text-primary">Featured</h4>
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
                  onClick={() => navigate(`/admin/users/${funcData.ownerId}`)}
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

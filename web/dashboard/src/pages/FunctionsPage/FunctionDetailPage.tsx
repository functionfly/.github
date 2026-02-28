import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import { LineChart } from "@/components/common/LineChart";
import { BarChart } from "@/components/common/BarChart";
import { PieChart } from "@/components/common/PieChart";
import {
  ArrowLeft,
  Edit,
  Rocket,
  Trash2,
  MoreVertical,
  Calendar,
  Clock,
  Globe,
  Activity,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Play,
  Pause,
  RotateCcw
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import "@/styles/components.css";

interface FunctionData {
  id: string;
  name: string;
  status: "online" | "offline" | "degraded";
  providers: string[];
  region: string;
  lastDeployed: string;
  createdAt: string;
  version: string;
  runtime: string;
  requests: number;
  avgLatency: number;
  errorRate: number;
  uptime: number;
}

interface Deployment {
  id: string;
  version: string;
  status: "success" | "failed" | "pending";
  timestamp: string;
  duration: number;
  triggeredBy: string;
  commit?: string;
}

interface LogEntry {
  id: string;
  timestamp: string;
  level: "info" | "warn" | "error";
  message: string;
  source: string;
}


const requestData = [
  { name: "00:00", requests: 120, errors: 1 },
  { name: "04:00", requests: 98, errors: 0 },
  { name: "08:00", requests: 245, errors: 2 },
  { name: "12:00", requests: 189, errors: 1 },
  { name: "16:00", requests: 312, errors: 3 },
  { name: "20:00", requests: 156, errors: 1 },
];

const latencyData = [
  { name: "Workers", latency: 45, color: "#f48120" },
  { name: "Vercel", latency: 62, color: "#000000" },
];

const errorData = [
  { name: "4xx", value: 65, color: "#f59e0b" },
  { name: "5xx", value: 25, color: "#ef4444" },
  { name: "Timeout", value: 10, color: "#8b5cf6" },
];

export function FunctionDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");
  const [isRedeploying, setIsRedeploying] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  // State for API data
  const [functionData, setFunctionData] = useState<FunctionData | null>(null);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch function data
  useEffect(() => {
    const fetchFunctionData = async () => {
      if (!id) return;

      try {
        setLoading(true);
        setError(null);

        // Fetch function details
        const functionResponse = await apiClient.get<FunctionData>(`/v1/functions/${id}`);
        setFunctionData(functionResponse);

        // Fetch deployments
        const deploymentsResponse = await apiClient.get<{ deployments: Deployment[] }>(`/v1/functions/${id}/deployments`);
        setDeployments(deploymentsResponse.deployments || []);

        // Fetch logs
        const logsResponse = await apiClient.get<{ logs: LogEntry[] }>(`/v1/functions/${id}/logs`);
        setLogs(logsResponse.logs || []);

      } catch (err) {
        console.error("Failed to load function data:", err);
        setError("Failed to load function data");
        toast.error("Failed to load function data");
      } finally {
        setLoading(false);
      }
    };

    fetchFunctionData();
  }, [id]);

  const handleRedeploy = async () => {
    if (!id) return;
    setIsRedeploying(true);
    try {
      await apiClient.post(`/v1/functions/${id}/redeploy`);
      toast.success("Function redeployed successfully");
    } catch (error) {
      toast.error("Failed to redeploy function. Please try again.");
    } finally {
      setIsRedeploying(false);
    }
  };

  const handleDelete = () => {
    setShowDeleteDialog(true);
  };

  const confirmDelete = async () => {
    if (!id || !functionData) return;

    try {
      setIsDeleting(true);
      await apiClient.delete(`/v1/functions/${id}`);
      toast.success(`Function "${functionData.name}" has been deleted successfully`);
      navigate("/functions");
    } catch (error) {
      console.error("Failed to delete function:", error);
      toast.error("Failed to delete function. Please try again.");
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  };

  if (loading || !functionData) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 bg-muted rounded animate-pulse" />
          <div className="space-y-2">
            <div className="w-48 h-6 bg-muted rounded animate-pulse" />
            <div className="w-32 h-4 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="p-6 border rounded-lg">
              <div className="w-16 h-16 bg-muted rounded animate-pulse mb-4" />
              <div className="w-20 h-4 bg-muted rounded animate-pulse mb-2" />
              <div className="w-16 h-6 bg-muted rounded animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  const stats = [
    {
      title: "Total Requests",
      value: functionData.requests.toLocaleString(),
      change: { value: 12, label: "from last week" },
      icon: <Globe className="w-5 h-5" />,
      trend: "up" as const,
    },
    {
      title: "Avg Latency",
      value: `${functionData.avgLatency}ms`,
      change: { value: -8, label: "from last week" },
      icon: <Clock className="w-5 h-5" />,
      trend: "up" as const,
    },
    {
      title: "Error Rate",
      value: `${functionData.errorRate}%`,
      change: { value: -0.2, label: "from last week" },
      icon: <AlertTriangle className="w-5 h-5" />,
      trend: "up" as const,
    },
    {
      title: "Uptime",
      value: `${functionData.uptime}%`,
      change: { value: 0.1, label: "from last week" },
      icon: <Activity className="w-5 h-5" />,
      trend: "up" as const,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/functions")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">{functionData.name}</h1>
              <StatusBadge status={functionData.status} />
            </div>
            <p className="text-text-secondary">Function management and monitoring</p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => navigate(`/functions/${id}/edit`)}
          >
            <Edit className="w-4 h-4 mr-2" />
            Edit
          </Button>
          <Button
            onClick={handleRedeploy}
            disabled={isRedeploying}
          >
            <Rocket className="w-4 h-4 mr-2" />
            {isRedeploying ? "Redeploying..." : "Redeploy"}
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-bg-tertiary border-white/8">
              <DropdownMenuItem>
                <Play className="w-4 h-4 mr-2" />
                Test Function
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Pause className="w-4 h-4 mr-2" />
                Pause Deployments
              </DropdownMenuItem>
              <DropdownMenuItem className="text-red-400">
                <Trash2 className="w-4 h-4 mr-2" />
                Delete Function
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Function Info Card */}
      <Card className="card">
        <CardContent className="card-content p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Providers</h3>
              <div className="flex items-center gap-2">
                {functionData.providers.map((provider) => (
                  <div key={provider} className="flex items-center gap-2">
                    <ProviderIcon provider={provider} size="sm" />
                    <span className="text-sm text-text-primary capitalize">{provider}</span>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Region</h3>
              <p className="text-text-primary">{functionData.region.toUpperCase()}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Runtime</h3>
              <p className="text-text-primary">{functionData.runtime}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Version</h3>
              <Badge variant="secondary">{functionData.version}</Badge>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Last Deployed</h3>
              <p className="text-text-primary">{functionData.lastDeployed}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">Created</h3>
              <p className="text-text-primary">{functionData.createdAt}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Row */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <LineChart
              title="Requests Over Time"
              data={requestData}
              series={[
                { key: "requests", name: "Requests", color: "#6366f1" },
                { key: "errors", name: "Errors", color: "#ef4444" },
              ]}
              height={300}
            />

            <BarChart
              title="Latency by Provider"
              data={latencyData}
              series={[{ key: "latency", name: "Latency (ms)", color: "#10b981" }]}
              height={300}
            />
          </div>
        </TabsContent>

        <TabsContent value="deployments" className="space-y-4">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Deployment History</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="space-y-4">
                {deployments.map((deployment) => (
                  <div key={deployment.id} className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                    <div className="flex items-center gap-4">
                      {deployment.status === "success" && <CheckCircle2 className="w-5 h-5 text-green-400" />}
                      {deployment.status === "failed" && <XCircle className="w-5 h-5 text-red-400" />}
                      {deployment.status === "pending" && <RotateCcw className="w-5 h-5 text-yellow-400 animate-spin" />}

                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-text-primary">v{deployment.version}</span>
                          <Badge
                            variant={deployment.status === "success" ? "default" : "destructive"}
                            className="text-xs"
                          >
                            {deployment.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-text-secondary">
                          {deployment.timestamp} • {deployment.duration}s • by {deployment.triggeredBy}
                        </p>
                        {deployment.commit && (
                          <p className="text-xs text-text-muted font-mono">{deployment.commit}</p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs" className="space-y-4">
          <Card className="card h-[600px]">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Recent Logs</CardTitle>
            </CardHeader>
            <CardContent className="card-content p-0">
              <ScrollArea className="h-[520px] p-6">
                <div className="space-y-3">
                  {logs.map((log) => (
                    <div key={log.id} className="flex items-start gap-3 p-3 rounded-lg bg-bg-tertiary">
                      <div className="text-text-muted font-mono text-xs w-24">
                        {log.timestamp.split(' ')[1]}
                      </div>
                      <div className="flex items-center gap-2 flex-1">
                        {log.level === "error" && <XCircle className="w-4 h-4 text-red-400" />}
                        {log.level === "warn" && <AlertTriangle className="w-4 h-4 text-yellow-400" />}
                        {log.level === "info" && <Activity className="w-4 h-4 text-blue-400" />}
                        <div className="flex-1">
                          <span className="text-sm text-text-primary">{log.message}</span>
                          <span className="text-xs text-text-muted ml-2">({log.source})</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="analytics" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <LineChart
              title="Error Rate Over Time"
              data={requestData}
              series={[{ key: "errors", name: "Errors", color: "#ef4444" }]}
              height={300}
            />

            <Card className="card">
              <CardHeader className="card-header">
                <CardTitle className="card-title">Error Distribution</CardTitle>
              </CardHeader>
              <CardContent className="card-content">
                <div className="h-[300px]">
                  <PieChart
                    data={errorData}
                    height={300}
                  />
                </div>
                <div className="flex justify-center gap-6 mt-4">
                  {errorData.map((item) => (
                    <div key={item.name} className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: item.color }}
                      />
                      <span className="text-sm text-text-secondary">
                        {item.name} ({item.value}%)
                      </span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Function</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the function "{functionData?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete Function"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
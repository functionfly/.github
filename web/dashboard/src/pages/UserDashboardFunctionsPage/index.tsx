import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import {
  ArrowLeft,
  Search,
  Plus,
  MoreVertical,
  Activity,
  Clock,
  Globe,
  Eye,
  Edit,
  Trash2,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import "@/styles/components.css";

interface FunctionItem {
  id: string;
  name: string;
  author: string;
  description: string;
  status: "online" | "offline" | "degraded";
  isPublic: boolean;
  runtime: string;
  lastDeployedAt?: string;
  invocations: number;
  avgLatency: number;
}

interface UserData {
  id: string;
  username: string;
  name: string;
  avatar?: string;
  functionCount: number;
}

export function UserDashboardFunctionsPage() {
  const { username } = useParams<{ username: string }>();
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [functions, setFunctions] = useState<FunctionItem[]>([]);
  const [user, setUser] = useState<UserData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch user and functions data
  useEffect(() => {
    const fetchData = async () => {
      if (!username) return;

      try {
        setLoading(true);
        setError(null);

        // Fetch user data
        const userResponse = await apiClient.get<UserData>(`/v1/users/${username}`);
        setUser(userResponse);

        // Fetch user's functions
        const functionsResponse = await apiClient.get<{ functions: FunctionItem[] }>(
          `/v1/users/${username}/functions`
        );
        setFunctions(functionsResponse.functions || []);
      } catch (err) {
        console.error("Failed to load data:", err);
        setError("Failed to load user functions");
        toast.error("Failed to load user functions");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [username]);

  const filteredFunctions = functions.filter((fn) =>
    fn.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const totalInvocations = functions.reduce((sum, fn) => sum + fn.invocations, 0);
  const onlineFunctions = functions.filter((fn) => fn.status === "online").length;

  if (loading) {
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
            onClick={() => navigate(`/u/${username}`)}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-white">
              {user?.name}'s Functions
            </h1>
            <p className="text-text-secondary">
              @{username} • {functions.length} functions
            </p>
          </div>
        </div>

        <Button onClick={() => navigate("/functions/new")}>
          <Plus className="w-4 h-4 mr-2" />
          Create Function
        </Button>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard
          title="Total Functions"
          value={functions.length.toString()}
          icon={<Globe className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Online Functions"
          value={onlineFunctions.toString()}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Total Invocations"
          value={totalInvocations.toLocaleString()}
          icon={<Clock className="w-5 h-5" />}
          trend="neutral"
        />
      </div>

      {/* Search */}
      <Card className="card">
        <CardContent className="card-content p-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search functions..."
              className="pl-9"
            />
          </div>
        </CardContent>
      </Card>

      {/* Functions List */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle className="card-title">Functions</CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          {filteredFunctions.length === 0 ? (
            <div className="text-center py-8 text-text-muted">
              <Globe className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p>No functions found</p>
              <Button
                variant="link"
                onClick={() => navigate("/functions/new")}
                className="mt-2"
              >
                Create your first function
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredFunctions.map((fn) => (
                <div
                  key={fn.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary hover:bg-bg-hover transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h3 className="font-medium text-text-primary">{fn.name}</h3>
                        <StatusBadge status={fn.status} />
                        {fn.isPublic && (
                          <Badge variant="outline" className="text-xs">
                            Public
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-text-muted">
                        {fn.description || "No description"}
                      </p>
                      <div className="flex items-center gap-4 mt-1 text-xs text-text-muted">
                        <span>{fn.runtime}</span>
                        <span>{fn.invocations} invocations</span>
                        <span>{fn.avgLatency}ms avg latency</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => navigate(`/fx/${fn.author}/${fn.name}`)}
                    >
                      <Eye className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => navigate(`/functions/${fn.id}/edit`)}
                    >
                      <Edit className="w-4 h-4" />
                    </Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => navigate(`/functions/${fn.id}/logs`)}
                        >
                          View Logs
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => navigate(`/functions/${fn.id}/settings`)}
                        >
                          Settings
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-red-400">
                          <Trash2 className="w-4 h-4 mr-2" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default UserDashboardFunctionsPage;

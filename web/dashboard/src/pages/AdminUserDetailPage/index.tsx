import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatCard } from "@/components/common/StatCard";
import {
  ArrowLeft,
  Edit,
  Save,
  User,
  Mail,
  Calendar,
  Shield,
  Activity,
  Key,
  AlertTriangle,
} from "lucide-react";
import "@/styles/components.css";

interface UserData {
  id: string;
  email: string;
  name: string;
  username: string;
  avatar?: string;
  role: string;
  status: "active" | "suspended" | "pending";
  createdAt: string;
  lastLoginAt?: string;
  tenantId?: string;
  tenantName?: string;
  metadata: {
    functions: number;
    totalInvocations: number;
    apiKeys: number;
  };
}

export function AdminUserDetailPage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [user, setUser] = useState<UserData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Editable fields
  const [name, setName] = useState("");
  const [role, setRole] = useState("");

  // Fetch user data
  useEffect(() => {
    const fetchUser = async () => {
      if (!userId) return;

      try {
        setLoading(true);
        setError(null);

        const response = await apiClient.get<UserData>(`/v1/admin/users/${userId}`);
        setUser(response);

        // Populate form fields
        setName(response.name);
        setRole(response.role);
      } catch (err) {
        console.error("Failed to load user:", err);
        setError("Failed to load user");
        toast.error("Failed to load user");
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [userId]);

  const handleSave = async () => {
    if (!userId) return;

    setIsSaving(true);
    try {
      const updates = {
        name,
        role,
      };

      await apiClient.patch(`/v1/admin/users/${userId}`, updates);
      toast.success("User updated successfully");
      setIsEditing(false);
    } catch (error) {
      console.error("Failed to update user:", error);
      toast.error("Failed to update user");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSuspend = async () => {
    if (!userId || !user) return;

    try {
      const newStatus = user.status === "active" ? "suspended" : "active";
      await apiClient.patch(`/v1/admin/users/${userId}`, { status: newStatus });
      setUser({ ...user, status: newStatus });
      toast.success(`User ${newStatus === "active" ? "activated" : "suspended"} successfully`);
    } catch (error) {
      console.error("Failed to update user status:", error);
      toast.error("Failed to update user status");
    }
  };

  const handleResetPassword = async () => {
    if (!userId) return;

    try {
      await apiClient.post(`/v1/admin/users/${userId}/reset-password`);
      toast.success("Password reset email sent");
    } catch (error) {
      console.error("Failed to reset password:", error);
      toast.error("Failed to reset password");
    }
  };

  if (loading || !user) {
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
            onClick={() => navigate("/admin/users")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div className="flex items-center gap-3">
            {user.avatar ? (
              <img
                src={user.avatar}
                alt={user.name}
                className="w-10 h-10 rounded-full"
              />
            ) : (
              <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                <User className="w-5 h-5 text-primary" />
              </div>
            )}
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold text-white">{user.name}</h1>
                <Badge
                  variant={user.status === "active" ? "default" : "destructive"}
                >
                  {user.status}
                </Badge>
                <Badge variant="outline">{user.role}</Badge>
              </div>
              <p className="text-text-secondary">@{user.username}</p>
            </div>
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" onClick={handleResetPassword}>
            <Key className="w-4 h-4 mr-2" />
            Reset Password
          </Button>
          <Button
            variant="outline"
            onClick={handleSuspend}
          >
            <Shield className="w-4 h-4 mr-2" />
            {user.status === "active" ? "Suspend" : "Activate"}
          </Button>
          <Button
            variant={isEditing ? "default" : "outline"}
            onClick={isEditing ? handleSave : () => setIsEditing(true)}
            disabled={isSaving}
          >
            {isEditing ? (
              <>
                <Save className="w-4 h-4 mr-2" />
                {isSaving ? "Saving..." : "Save"}
              </>
            ) : (
              <>
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          title="Functions"
          value={user.metadata.functions.toString()}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Total Invocations"
          value={user.metadata.totalInvocations.toLocaleString()}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="API Keys"
          value={user.metadata.apiKeys.toString()}
          icon={<Key className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Last Login"
          value={user.lastLoginAt || "Never"}
          icon={<Calendar className="w-5 h-5" />}
          trend="neutral"
        />
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">User Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Full Name</Label>
                  {isEditing ? (
                    <Input
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">{user.name}</p>
                  )}
                </div>
                <div>
                  <Label className="text-text-muted">Username</Label>
                  <p className="text-text-primary mt-1">@{user.username}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Email</Label>
                  <p className="text-text-primary mt-1 flex items-center gap-2">
                    <Mail className="w-4 h-4" />
                    {user.email}
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Role</Label>
                  {isEditing ? (
                    <Input
                      value={role}
                      onChange={(e) => setRole(e.target.value)}
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">{user.role}</p>
                  )}
                </div>
                <div>
                  <Label className="text-text-muted">Created</Label>
                  <p className="text-text-primary mt-1 flex items-center gap-2">
                    <Calendar className="w-4 h-4" />
                    {user.createdAt}
                  </p>
                </div>
                {user.tenantName && (
                  <div>
                    <Label className="text-text-muted">Tenant</Label>
                    <p className="text-text-primary mt-1">{user.tenantName}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Security</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div>
                  <h4 className="font-medium text-text-primary">Two-Factor Authentication</h4>
                  <p className="text-sm text-text-muted">Add an extra layer of security</p>
                </div>
                <Badge variant="secondary">Not Enabled</Badge>
              </div>

              <Separator />

              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div>
                  <h4 className="font-medium text-text-primary">API Keys</h4>
                  <p className="text-sm text-text-muted">Manage API keys for this user</p>
                </div>
                <Badge variant="outline">{user.metadata.apiKeys} keys</Badge>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Activity Tab */}
        <TabsContent value="activity" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Recent Activity</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="text-center py-8 text-text-muted">
                <Activity className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p>No recent activity</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AdminUserDetailPage;

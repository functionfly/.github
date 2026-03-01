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
  Users,
  Activity,
  DollarSign,
  Globe,
  Mail,
  Calendar,
  Shield,
  AlertTriangle,
} from "lucide-react";
import "@/styles/components.css";

interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: "active" | "suspended" | "pending";
  plan: string;
  createdAt: string;
  updatedAt: string;
  ownerEmail: string;
  customDomain?: string;
  settings: {
    allowPublicFunctions: boolean;
    allowCustomDomains: boolean;
    maxFunctions: number;
    maxUsers: number;
  };
  usage: {
    functions: number;
    users: number;
    apiCalls: number;
    storage: number;
  };
  billing: {
    monthlySpend: number;
    lastPaymentDate?: string;
    paymentMethod?: string;
  };
}

export function AdminTenantDetailPage() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Editable fields
  const [name, setName] = useState("");
  const [customDomain, setCustomDomain] = useState("");
  const [maxFunctions, setMaxFunctions] = useState(0);
  const [maxUsers, setMaxUsers] = useState(0);
  const [allowPublicFunctions, setAllowPublicFunctions] = useState(false);
  const [allowCustomDomains, setAllowCustomDomains] = useState(false);

  // Fetch tenant data
  useEffect(() => {
    const fetchTenant = async () => {
      if (!tenantId) return;

      try {
        setLoading(true);
        setError(null);

        const response = await apiClient.get<Tenant>(`/v1/admin/tenants/${tenantId}`);
        setTenant(response);

        // Populate form fields
        setName(response.name);
        setCustomDomain(response.customDomain || "");
        setMaxFunctions(response.settings.maxFunctions);
        setMaxUsers(response.settings.maxUsers);
        setAllowPublicFunctions(response.settings.allowPublicFunctions);
        setAllowCustomDomains(response.settings.allowCustomDomains);
      } catch (err) {
        console.error("Failed to load tenant:", err);
        setError("Failed to load tenant");
        toast.error("Failed to load tenant");
      } finally {
        setLoading(false);
      }
    };

    fetchTenant();
  }, [tenantId]);

  const handleSave = async () => {
    if (!tenantId) return;

    setIsSaving(true);
    try {
      const updates = {
        name,
        customDomain,
        settings: {
          maxFunctions,
          maxUsers,
          allowPublicFunctions,
          allowCustomDomains,
        },
      };

      await apiClient.patch(`/v1/admin/tenants/${tenantId}`, updates);
      toast.success("Tenant updated successfully");
      setIsEditing(false);
    } catch (error) {
      console.error("Failed to update tenant:", error);
      toast.error("Failed to update tenant");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSuspend = async () => {
    if (!tenantId || !tenant) return;

    try {
      const newStatus = tenant.status === "active" ? "suspended" : "active";
      await apiClient.patch(`/v1/admin/tenants/${tenantId}`, { status: newStatus });
      setTenant({ ...tenant, status: newStatus });
      toast.success(`Tenant ${newStatus === "active" ? "activated" : "suspended"} successfully`);
    } catch (error) {
      console.error("Failed to update tenant status:", error);
      toast.error("Failed to update tenant status");
    }
  };

  if (loading || !tenant) {
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
            onClick={() => navigate("/admin/tenants")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">{tenant.name}</h1>
              <Badge
                variant={tenant.status === "active" ? "default" : "destructive"}
              >
                {tenant.status}
              </Badge>
              <Badge variant="outline">{tenant.plan}</Badge>
            </div>
            <p className="text-text-secondary">Tenant ID: {tenant.id}</p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleSuspend}
          >
            <Shield className="w-4 h-4 mr-2" />
            {tenant.status === "active" ? "Suspend" : "Activate"}
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
          value={`${tenant.usage.functions} / ${tenant.settings.maxFunctions}`}
          icon={<Activity className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Users"
          value={`${tenant.usage.users} / ${tenant.settings.maxUsers}`}
          icon={<Users className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="API Calls"
          value={tenant.usage.apiCalls.toLocaleString()}
          icon={<Globe className="w-5 h-5" />}
          trend="neutral"
        />
        <StatCard
          title="Monthly Spend"
          value={`$${tenant.billing.monthlySpend.toFixed(2)}`}
          icon={<DollarSign className="w-5 h-5" />}
          trend="neutral"
        />
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="billing">Billing</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Tenant Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Tenant Name</Label>
                  {isEditing ? (
                    <Input
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">{tenant.name}</p>
                  )}
                </div>
                <div>
                  <Label className="text-text-muted">Slug</Label>
                  <p className="text-text-primary mt-1">{tenant.slug}</p>
                </div>
                <div>
                  <Label className="text-text-muted">Owner Email</Label>
                  <p className="text-text-primary mt-1 flex items-center gap-2">
                    <Mail className="w-4 h-4" />
                    {tenant.ownerEmail}
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Created</Label>
                  <p className="text-text-primary mt-1 flex items-center gap-2">
                    <Calendar className="w-4 h-4" />
                    {tenant.createdAt}
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Custom Domain</Label>
                  {isEditing ? (
                    <Input
                      value={customDomain}
                      onChange={(e) => setCustomDomain(e.target.value)}
                      placeholder="example.com"
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">
                      {tenant.customDomain || "None"}
                    </p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Resource Limits</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Max Functions</Label>
                  {isEditing ? (
                    <Input
                      type="number"
                      value={maxFunctions}
                      onChange={(e) => setMaxFunctions(parseInt(e.target.value))}
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">{tenant.settings.maxFunctions}</p>
                  )}
                </div>
                <div>
                  <Label className="text-text-muted">Max Users</Label>
                  {isEditing ? (
                    <Input
                      type="number"
                      value={maxUsers}
                      onChange={(e) => setMaxUsers(parseInt(e.target.value))}
                      className="mt-1"
                    />
                  ) : (
                    <p className="text-text-primary mt-1">{tenant.settings.maxUsers}</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Feature Flags</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Allow Public Functions</Label>
                  <p className="text-sm text-text-muted">
                    Allow tenant to publish public functions
                  </p>
                </div>
                {isEditing ? (
                  <Switch
                    checked={allowPublicFunctions}
                    onCheckedChange={setAllowPublicFunctions}
                  />
                ) : (
                  <Badge variant={tenant.settings.allowPublicFunctions ? "default" : "secondary"}>
                    {tenant.settings.allowPublicFunctions ? "Enabled" : "Disabled"}
                  </Badge>
                )}
              </div>

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <Label>Allow Custom Domains</Label>
                  <p className="text-sm text-text-muted">
                    Allow tenant to use custom domains
                  </p>
                </div>
                {isEditing ? (
                  <Switch
                    checked={allowCustomDomains}
                    onCheckedChange={setAllowCustomDomains}
                  />
                ) : (
                  <Badge variant={tenant.settings.allowCustomDomains ? "default" : "secondary"}>
                    {tenant.settings.allowCustomDomains ? "Enabled" : "Disabled"}
                  </Badge>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Billing Tab */}
        <TabsContent value="billing" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Current Plan</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div>
                  <h3 className="font-semibold text-text-primary">{tenant.plan}</h3>
                  <p className="text-sm text-text-muted">Current subscription plan</p>
                </div>
                <Button variant="outline">Change Plan</Button>
              </div>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Payment Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Monthly Spend</Label>
                  <p className="text-text-primary mt-1">
                    ${tenant.billing.monthlySpend.toFixed(2)}
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Last Payment</Label>
                  <p className="text-text-primary mt-1">
                    {tenant.billing.lastPaymentDate || "N/A"}
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Payment Method</Label>
                  <p className="text-text-primary mt-1">
                    {tenant.billing.paymentMethod || "Not set"}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AdminTenantDetailPage;

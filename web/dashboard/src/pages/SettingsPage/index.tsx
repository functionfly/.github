import { useState } from "react";
import { User, CreditCard, Key, Bell } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthStore } from "@/stores/authStore";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { useQuery } from "@tanstack/react-query";

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState("account");
  const user = useAuthStore((state) => state.user);

  // Split name into first/last
  const nameParts = (user?.name || "").split(" ");
  const [firstName, setFirstName] = useState(nameParts[0] || "");
  const [lastName, setLastName] = useState(nameParts.slice(1).join(" ") || "");
  const [email, setEmail] = useState(user?.email || "");
  const [isSavingProfile, setIsSavingProfile] = useState(false);

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isUpdatingPassword, setIsUpdatingPassword] = useState(false);

  const [notifications, setNotifications] = useState({
    deploymentSuccess: true,
    deploymentFailure: true,
    failoverEvents: true,
    providerIssues: true,
  });

  // Fetch API keys
  const { data: apiKeysData, isLoading: apiKeysLoading, refetch: refetchApiKeys } = useQuery({
    queryKey: ["api-keys"],
    queryFn: () => apiClient.get("/v1/api-keys"),
    retry: false,
  });

  const apiKeys = (apiKeysData as any)?.keys ?? [];

  const handleSaveProfile = async () => {
    setIsSavingProfile(true);
    try {
      await apiClient.patch("/v1/users/me", {
        name: `${firstName} ${lastName}`.trim(),
        email,
      });
      toast.success("Profile updated successfully");
    } catch {
      toast.error("Failed to update profile");
    } finally {
      setIsSavingProfile(false);
    }
  };

  const handleUpdatePassword = async () => {
    if (newPassword !== confirmPassword) {
      toast.error("New passwords do not match");
      return;
    }
    if (newPassword.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }
    setIsUpdatingPassword(true);
    try {
      await apiClient.post("/v1/users/me/change-password", {
        currentPassword,
        newPassword,
      });
      toast.success("Password updated successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch {
      toast.error("Failed to update password. Check your current password.");
    } finally {
      setIsUpdatingPassword(false);
    }
  };

  const handleGenerateApiKey = async () => {
    try {
      await apiClient.post("/v1/api-keys", { name: "New Key" });
      refetchApiKeys();
      toast.success("API key generated");
    } catch {
      toast.error("Failed to generate API key");
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-text-secondary">Manage your account and preferences</p>
      </div>

      {/* Settings Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="bg-bg-secondary border border-white/8">
          <TabsTrigger value="account" className="gap-2">
            <User className="w-4 h-4" />
            Account
          </TabsTrigger>
          <TabsTrigger value="billing" className="gap-2">
            <CreditCard className="w-4 h-4" />
            Billing
          </TabsTrigger>
          <TabsTrigger value="api" className="gap-2">
            <Key className="w-4 h-4" />
            API Keys
          </TabsTrigger>
          <TabsTrigger value="notifications" className="gap-2">
            <Bell className="w-4 h-4" />
            Notifications
          </TabsTrigger>
        </TabsList>

        {/* Account Settings */}
        <TabsContent value="account" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription className="text-text-secondary">
                Update your account details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="firstName">First Name</Label>
                  <Input
                    id="firstName"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="lastName">Last Name</Label>
                  <Input
                    id="lastName"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              <Button onClick={handleSaveProfile} disabled={isSavingProfile}>
                {isSavingProfile ? "Saving..." : "Save Changes"}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Password</CardTitle>
              <CardDescription className="text-text-secondary">
                Update your password
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="currentPassword">Current Password</Label>
                <Input
                  id="currentPassword"
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="newPassword">New Password</Label>
                <Input
                  id="newPassword"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm New Password</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                />
              </div>
              <Button onClick={handleUpdatePassword} disabled={isUpdatingPassword}>
                {isUpdatingPassword ? "Updating..." : "Update Password"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Billing Settings */}
        <TabsContent value="billing" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Current Plan</CardTitle>
              <CardDescription className="text-text-secondary">
                Manage your subscription
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-[#6366f1]/20">
                <div>
                  <h3 className="font-semibold text-white capitalize">{user?.plan || "Starter"} Plan</h3>
                  <p className="text-sm text-text-secondary">
                    {user?.plan === "free" ? "Free forever" : "Active subscription"}
                  </p>
                </div>
                <Badge>Current</Badge>
              </div>
              <div className="mt-6">
                <Button variant="outline" onClick={() => window.open("/pricing", "_blank")}>
                  Upgrade Plan
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* API Keys */}
        <TabsContent value="api" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>API Keys</CardTitle>
              <CardDescription className="text-text-secondary">
                Manage your API keys for programmatic access
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {apiKeysLoading ? (
                  <p className="text-text-muted text-sm">Loading API keys...</p>
                ) : apiKeys.length === 0 ? (
                  <p className="text-text-muted text-sm">No API keys yet. Generate one below.</p>
                ) : (
                  apiKeys.map((key: any) => (
                    <div key={key.id} className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8">
                      <div>
                        <h4 className="font-medium text-white">{key.name || "API Key"}</h4>
                        <p className="text-sm text-text-muted">
                          Created {key.created_at ? new Date(key.created_at).toLocaleDateString() : "recently"}
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <code className="px-3 py-1 rounded bg-bg-tertiary text-sm text-text-secondary">
                          {key.prefix ? `${key.prefix}••••••••••••` : "ff_live_••••••••••••"}
                        </code>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            navigator.clipboard.writeText(key.key || "");
                            toast.success("Copied to clipboard");
                          }}
                        >
                          Copy
                        </Button>
                      </div>
                    </div>
                  ))
                )}
                <Button className="gap-2" onClick={handleGenerateApiKey}>
                  <Key className="w-4 h-4" />
                  Generate New Key
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Notifications */}
        <TabsContent value="notifications" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Notification Preferences</CardTitle>
              <CardDescription className="text-text-secondary">
                Choose what notifications you want to receive
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {[
                  { key: "deploymentSuccess" as const, label: "Deployment Success", description: "Get notified when a deployment succeeds" },
                  { key: "deploymentFailure" as const, label: "Deployment Failure", description: "Get notified when a deployment fails" },
                  { key: "failoverEvents" as const, label: "Failover Events", description: "Get notified when failover is triggered" },
                  { key: "providerIssues" as const, label: "Provider Issues", description: "Get notified when a provider has issues" },
                ].map((item) => (
                  <div key={item.key} className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-white">{item.label}</h4>
                      <p className="text-sm text-text-muted">{item.description}</p>
                    </div>
                    <Switch
                      checked={notifications[item.key]}
                      onCheckedChange={(checked) =>
                        setNotifications((prev) => ({ ...prev, [item.key]: checked }))
                      }
                    />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

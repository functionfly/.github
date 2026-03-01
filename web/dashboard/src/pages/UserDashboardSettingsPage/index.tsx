import { useState, useEffect } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { PLANS } from "@/lib/constants";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { usersApi } from "@/api/users";
import { createBillingPortalSession, getBillingPortalErrorMessage } from "@/api/billing";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  ArrowLeft,
  Save,
  User,
  Mail,
  Lock,
  Bell,
  Globe,
  Shield,
  Key,
  Trash2,
  CreditCard,
} from "lucide-react";
import "@/styles/components.css";

interface UserProfile {
  id: string;
  username: string;
  name: string;
  email: string;
  avatar?: string;
  bio?: string;
  website?: string;
  twitter?: string;
  github?: string;
  settings: {
    emailNotifications: boolean;
    marketingEmails: boolean;
    publicProfile: boolean;
    allowMessaging: boolean;
  };
}

interface UserDashboardSettingsPageProps {
  initialTab?: string;
}

export function UserDashboardSettingsPage({ initialTab }: UserDashboardSettingsPageProps) {
  const { username } = useParams<{ username: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const user = useAuthStore((s) => s.user);
  const [activeTab, setActiveTab] = useState(initialTab || "profile");
  const [isSaving, setIsSaving] = useState(false);
  const [billingPortalLoading, setBillingPortalLoading] = useState(false);

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Plan from GET /users/me (authoritative) for billing section
  const { data: meData } = useQuery({
    queryKey: ["users", "me"],
    queryFn: () => usersApi.getMe(),
    retry: false,
  });
  const setUserPlan = useAuthStore((s) => s.setUserPlan);
  useEffect(() => {
    if (meData?.plan !== undefined && meData.plan !== useAuthStore.getState().user?.plan) {
      setUserPlan(meData.plan);
    }
  }, [meData?.plan, setUserPlan]);
  const displayPlan = meData?.plan ?? user?.plan ?? "Free";

  // Form fields
  const [name, setName] = useState("");
  const [bio, setBio] = useState("");
  const [website, setWebsite] = useState("");
  const [twitter, setTwitter] = useState("");
  const [github, setGithub] = useState("");
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [marketingEmails, setMarketingEmails] = useState(false);
  const [publicProfile, setPublicProfile] = useState(true);
  const [allowMessaging, setAllowMessaging] = useState(false);

  // Fetch user profile
  useEffect(() => {
    const fetchProfile = async () => {
      if (!username) return;

      try {
        setLoading(true);
        setError(null);

        const response = await apiClient.get<UserProfile>(`/v1/users/${username}/settings`);
        setProfile(response);

        // Populate form fields
        setName(response.name);
        setBio(response.bio || "");
        setWebsite(response.website || "");
        setTwitter(response.twitter || "");
        setGithub(response.github || "");
        setEmailNotifications(response.settings.emailNotifications);
        setMarketingEmails(response.settings.marketingEmails);
        setPublicProfile(response.settings.publicProfile);
        setAllowMessaging(response.settings.allowMessaging);
      } catch (err) {
        console.error("Failed to load profile:", err);
        setError("Failed to load profile settings");
        toast.error("Failed to load profile settings");
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [username]);

  const handleSaveProfile = async () => {
    if (!username) return;

    setIsSaving(true);
    try {
      const updates = {
        name,
        bio,
        website,
        twitter,
        github,
      };

      await apiClient.patch(`/v1/users/${username}/settings/profile`, updates);
      toast.success("Profile updated successfully");
    } catch (error) {
      console.error("Failed to update profile:", error);
      toast.error("Failed to update profile");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveNotifications = async () => {
    if (!username) return;

    setIsSaving(true);
    try {
      const updates = {
        emailNotifications,
        marketingEmails,
      };

      await apiClient.patch(`/v1/users/${username}/settings/notifications`, updates);
      toast.success("Notification settings updated successfully");
    } catch (error) {
      console.error("Failed to update notifications:", error);
      toast.error("Failed to update notification settings");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSavePrivacy = async () => {
    if (!username) return;

    setIsSaving(true);
    try {
      const updates = {
        publicProfile,
        allowMessaging,
      };

      await apiClient.patch(`/v1/users/${username}/settings/privacy`, updates);
      toast.success("Privacy settings updated successfully");
    } catch (error) {
      console.error("Failed to update privacy settings:", error);
      toast.error("Failed to update privacy settings");
    } finally {
      setIsSaving(false);
    }
  };

  if (loading || !profile) {
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
            <h1 className="text-2xl font-bold text-text-primary">Settings</h1>
            <p className="text-text-secondary">
              Manage your profile and preferences
            </p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="billing">Billing</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
          <TabsTrigger value="privacy">Privacy</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
        </TabsList>

        {/* Profile Tab */}
        <TabsContent value="profile" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Profile Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-muted">Username</Label>
                  <Input value={profile.username} disabled className="mt-1" />
                  <p className="text-xs text-text-muted mt-1">
                    Username cannot be changed
                  </p>
                </div>
                <div>
                  <Label className="text-text-muted">Email</Label>
                  <Input value={profile.email} disabled className="mt-1" />
                </div>
                <div className="col-span-2">
                  <Label className="text-text-muted">Display Name</Label>
                  <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="mt-1"
                  />
                </div>
                <div className="col-span-2">
                  <Label className="text-text-muted">Bio</Label>
                  <Input
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                    placeholder="Tell us about yourself"
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label className="text-text-muted">Website</Label>
                  <Input
                    value={website}
                    onChange={(e) => setWebsite(e.target.value)}
                    placeholder="https://example.com"
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label className="text-text-muted">Twitter</Label>
                  <Input
                    value={twitter}
                    onChange={(e) => setTwitter(e.target.value)}
                    placeholder="@username"
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label className="text-text-muted">GitHub</Label>
                  <Input
                    value={github}
                    onChange={(e) => setGithub(e.target.value)}
                    placeholder="username"
                    className="mt-1"
                  />
                </div>
              </div>

              <div className="flex justify-end pt-4">
                <Button onClick={handleSaveProfile} disabled={isSaving}>
                  <Save className="w-4 h-4 mr-2" />
                  {isSaving ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Billing Tab */}
        <TabsContent value="billing" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Current Plan</CardTitle>
              <p className="text-sm text-text-muted mt-1">Manage your subscription</p>
            </CardHeader>
            <CardContent className="card-content">
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-[#6366f1]/20">
                <div>
                  <h3 className="font-semibold text-text-primary capitalize">
                    {(displayPlan && PLANS[displayPlan.toUpperCase() as keyof typeof PLANS]?.name) || displayPlan} Plan
                  </h3>
                  <p className="text-sm text-text-muted">
                    {displayPlan === "free" ? "Free forever" : "Active subscription"}
                  </p>
                </div>
                <Badge>Current</Badge>
              </div>
              <div className="mt-6 flex flex-wrap gap-3">
                <Button
                  variant="default"
                  onClick={async () => {
                    setBillingPortalLoading(true);
                    try {
                      const returnUrl = `${window.location.origin}${location.pathname}`;
                      const { url } = await createBillingPortalSession(returnUrl);
                      window.location.href = url;
                    } catch (e: unknown) {
                      setBillingPortalLoading(false);
                      toast.error(getBillingPortalErrorMessage(e));
                    }
                  }}
                  disabled={billingPortalLoading}
                >
                  {billingPortalLoading ? "Opening…" : "Manage billing"}
                </Button>
                <Button
                  variant="outline"
                  className="settings-upgrade-btn"
                  disabled={billingPortalLoading}
                  onClick={async () => {
                    setBillingPortalLoading(true);
                    try {
                      const returnUrl = `${window.location.origin}/pricing`;
                      const { url } = await createBillingPortalSession(returnUrl);
                      window.location.href = url;
                    } catch (e: unknown) {
                      setBillingPortalLoading(false);
                      toast.error(getBillingPortalErrorMessage(e));
                    }
                  }}
                >
                  {billingPortalLoading ? "Opening…" : "Upgrade Plan"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Notifications Tab */}
        <TabsContent value="notifications" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Email Preferences</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Email Notifications</Label>
                  <p className="text-sm text-text-muted">
                    Receive emails about your function activity
                  </p>
                </div>
                <Switch
                  checked={emailNotifications}
                  onCheckedChange={setEmailNotifications}
                />
              </div>

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <Label>Marketing Emails</Label>
                  <p className="text-sm text-text-muted">
                    Receive updates about new features and announcements
                  </p>
                </div>
                <Switch
                  checked={marketingEmails}
                  onCheckedChange={setMarketingEmails}
                />
              </div>

              <div className="flex justify-end pt-4">
                <Button onClick={handleSaveNotifications} disabled={isSaving}>
                  <Save className="w-4 h-4 mr-2" />
                  {isSaving ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Privacy Tab */}
        <TabsContent value="privacy" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Profile Visibility</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Public Profile</Label>
                  <p className="text-sm text-text-muted">
                    Allow anyone to view your profile and functions
                  </p>
                </div>
                <Switch
                  checked={publicProfile}
                  onCheckedChange={setPublicProfile}
                />
              </div>

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <Label>Allow Messaging</Label>
                  <p className="text-sm text-text-muted">
                    Allow others to send you messages
                  </p>
                </div>
                <Switch
                  checked={allowMessaging}
                  onCheckedChange={setAllowMessaging}
                />
              </div>

              <div className="flex justify-end pt-4">
                <Button onClick={handleSavePrivacy} disabled={isSaving}>
                  <Save className="w-4 h-4 mr-2" />
                  {isSaving ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Security Tab */}
        <TabsContent value="security" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Password</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-primary/10 rounded-lg">
                    <Lock className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-medium text-text-primary">Change Password</h4>
                    <p className="text-sm text-text-muted">
                      Update your password regularly for security
                    </p>
                  </div>
                </div>
                <Button variant="outline">Change Password</Button>
              </div>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">API Keys</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-primary/10 rounded-lg">
                    <Key className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-medium text-text-primary">Manage API Keys</h4>
                    <p className="text-sm text-text-muted">
                      Create and manage your API keys
                    </p>
                  </div>
                </div>
                <Button variant="outline">Manage Keys</Button>
              </div>
            </CardContent>
          </Card>

          <Card className="card border-red-500/20">
            <CardHeader className="card-header">
              <CardTitle className="card-title text-red-400">Danger Zone</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="flex items-center justify-between p-4 rounded-lg bg-red-500/5 border border-red-500/20">
                <div>
                  <h4 className="font-medium text-text-primary">Delete Account</h4>
                  <p className="text-sm text-text-muted">
                    Permanently delete your account and all data
                  </p>
                </div>
                <Button variant="destructive">
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete Account
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default UserDashboardSettingsPage;

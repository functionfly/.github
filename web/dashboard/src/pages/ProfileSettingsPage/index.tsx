/**
 * ProfileSettingsPage Component
 *
 * A comprehensive settings page for user profile management.
 * Features tabbed interface for Profile, Account, Notifications, and Privacy settings.
 *
 * @example
 * <Route path="/settings/profile" element={<ProfileSettingsPage />} />
 */

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import {
  User,
  Lock,
  Bell,
  Shield,
  Globe,
  Camera,
  Save,
  Loader2,
  AlertCircle,
  CheckCircle2,
  ChevronRight,
  Mail,
  Smartphone,
  Eye,
  EyeOff,
  Trash2,
  ArrowLeft,
  AtSign,
  Building2,
  Briefcase,
  MapPin,
  Link as LinkIcon,
  Github,
  Twitter,
  Linkedin,
  ImageIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useAuthStore } from "@/stores/authStore";
import { usersApi, type UpdateProfileRequest } from "@/api/users";

type SettingsTab = "profile" | "account" | "notifications" | "privacy";

interface FormErrors {
  name?: string;
  username?: string;
  website?: string;
  twitterUrl?: string;
  githubUrl?: string;
  linkedinUrl?: string;
  currentPassword?: string;
  newPassword?: string;
  confirmPassword?: string;
}

const URL_REGEX = /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/;
const USERNAME_REGEX = /^[a-z0-9_-]+$/;
const MAX_BIO_LENGTH = 500;

export function ProfileSettingsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const [activeTab, setActiveTab] = useState<SettingsTab>("profile");

  // Fetch latest profile from server
  const { data: serverProfile, isLoading: isLoadingProfile } = useQuery({
    queryKey: ["my-profile"],
    queryFn: () => usersApi.getMe(),
    staleTime: 60 * 1000,
  });

  const profile = serverProfile ?? {
    id: user?.id ?? "",
    name: user?.name ?? "",
    username: user?.username ?? "",
    companyName: user?.companyName ?? "",
    bio: (user as any)?.bio ?? "",
    email: user?.email ?? "",
    avatar: user?.avatar,
    updatedAt: user?.updatedAt ?? "",
  };

  // Update profile mutation
  const updateMutation = useMutation({
    mutationFn: (data: UpdateProfileRequest) => usersApi.updateMe(data),
    onSuccess: () => {
      toast.success("Profile updated successfully");
      queryClient.invalidateQueries({ queryKey: ["my-profile"] });
      useAuthStore.getState().initialize();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to update profile");
    },
  });

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Header */}
      <div className="border-b border-border-subtle bg-bg-secondary">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate(-1)}
              className="shrink-0"
            >
              <ArrowLeft className="w-5 h-5" />
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-text-primary">Profile Settings</h1>
              <p className="text-text-secondary">
                Manage your profile, account, and preferences
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as SettingsTab)}>
          <div className="flex flex-col lg:flex-row gap-8">
            {/* Sidebar */}
            <div className="lg:w-64 shrink-0">
              <TabsList className="flex flex-col h-auto bg-transparent gap-1 w-full">
                <TabsTrigger
                  value="profile"
                  className="w-full justify-start gap-3 px-4 py-3 data-[state=active]:bg-bg-tertiary"
                >
                  <User className="w-4 h-4" />
                  Profile
                </TabsTrigger>
                <TabsTrigger
                  value="account"
                  className="w-full justify-start gap-3 px-4 py-3 data-[state=active]:bg-bg-tertiary"
                >
                  <Lock className="w-4 h-4" />
                  Account
                </TabsTrigger>
                <TabsTrigger
                  value="notifications"
                  className="w-full justify-start gap-3 px-4 py-3 data-[state=active]:bg-bg-tertiary"
                >
                  <Bell className="w-4 h-4" />
                  Notifications
                </TabsTrigger>
                <TabsTrigger
                  value="privacy"
                  className="w-full justify-start gap-3 px-4 py-3 data-[state=active]:bg-bg-tertiary"
                >
                  <Shield className="w-4 h-4" />
                  Privacy
                </TabsTrigger>
              </TabsList>
            </div>

            {/* Main Content */}
            <div className="flex-1 min-w-0">
              <AnimatePresence mode="wait">
                <motion.div
                  key={activeTab}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  transition={{ duration: 0.2 }}
                >
                  <TabsContent value="profile" className="mt-0">
                    <ProfileTab
                      profile={profile}
                      onUpdate={updateMutation.mutate}
                      isLoading={updateMutation.isPending}
                    />
                  </TabsContent>

                  <TabsContent value="account" className="mt-0">
                    <AccountTab />
                  </TabsContent>

                  <TabsContent value="notifications" className="mt-0">
                    <NotificationsTab />
                  </TabsContent>

                  <TabsContent value="privacy" className="mt-0">
                    <PrivacyTab />
                  </TabsContent>
                </motion.div>
              </AnimatePresence>
            </div>
          </div>
        </Tabs>
      </div>
    </div>
  );
}

// ============================================================================
// Profile Tab
// ============================================================================

interface ProfileTabProps {
  profile: {
    id: string;
    name: string;
    username?: string;
    companyName?: string;
    bio?: string;
    email: string;
    avatar?: string;
    updatedAt: string;
  };
  onUpdate: (data: UpdateProfileRequest) => void;
  isLoading: boolean;
}

function ProfileTab({ profile, onUpdate, isLoading }: ProfileTabProps) {
  const [formData, setFormData] = useState({
    name: profile.name || "",
    username: profile.username || "",
    bio: profile.bio || "",
    location: "",
    company: profile.companyName || "",
    jobTitle: "",
    website: "",
    twitterUrl: "",
    githubUrl: "",
    linkedinUrl: "",
    coverImageUrl: "",
  });

  const [errors, setErrors] = useState<FormErrors>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Sync with profile data
  useEffect(() => {
    setFormData((prev) => ({
      ...prev,
      name: profile.name || "",
      username: profile.username || "",
      bio: profile.bio || "",
      company: profile.companyName || "",
    }));
  }, [profile]);

  const validateField = (field: string, value: string): string | undefined => {
    switch (field) {
      case "name":
        if (!value.trim()) return "Name is required";
        if (value.trim().length < 2) return "Name must be at least 2 characters";
        if (value.trim().length > 50) return "Name must be less than 50 characters";
        return;
      case "username":
        if (!value) return "Username is required";
        if (value.length < 3) return "Username must be at least 3 characters";
        if (value.length > 30) return "Username must be less than 30 characters";
        if (!USERNAME_REGEX.test(value)) {
          return "Username can only contain lowercase letters, numbers, hyphens, and underscores";
        }
        return;
      case "website":
      case "twitterUrl":
      case "githubUrl":
      case "linkedinUrl":
        if (value && !URL_REGEX.test(value) && !value.startsWith("http")) {
          return "Please enter a valid URL";
        }
        return;
      default:
        return;
    }
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
    const error = validateField(field, formData[field as keyof typeof formData]);
    setErrors((prev) => ({ ...prev, [field]: error }));
  };

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (touched[field]) {
      const error = validateField(field, value);
      setErrors((prev) => ({ ...prev, [field]: error }));
    }
  };

  const handleSave = () => {
    // Validate all fields
    const newErrors: FormErrors = {};
    Object.keys(formData).forEach((key) => {
      const error = validateField(key, formData[key as keyof typeof formData]);
      if (error) newErrors[key as keyof FormErrors] = error;
    });

    setErrors(newErrors);
    setTouched(Object.keys(formData).reduce((acc, key) => ({ ...acc, [key]: true }), {}));

    if (Object.values(newErrors).some(Boolean)) return;

    onUpdate({
      name: formData.name.trim() || undefined,
      username: formData.username.trim() || undefined,
      bio: formData.bio.trim() || undefined,
      location: formData.location.trim() || undefined,
      companyName: formData.company.trim() || undefined,
      jobTitle: formData.jobTitle.trim() || undefined,
      website: formData.website.trim() || undefined,
      twitterUrl: formData.twitterUrl.trim() || undefined,
      githubUrl: formData.githubUrl.trim() || undefined,
      linkedinUrl: formData.linkedinUrl.trim() || undefined,
    });
  };

  const isDirty =
    formData.name !== (profile.name || "") ||
    formData.username !== (profile.username || "") ||
    formData.bio !== (profile.bio || "") ||
    formData.company !== (profile.companyName || "");

  return (
    <div className="space-y-6">
      {/* Avatar Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Camera className="w-5 h-5 text-brand-500" />
            Profile Picture
          </CardTitle>
          <CardDescription>Your avatar is synced from your social login provider</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-6">
            <div className="w-24 h-24 rounded-full bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center text-white text-3xl font-bold overflow-hidden">
              {profile.avatar ? (
                <img src={profile.avatar} alt={profile.name} className="w-full h-full object-cover" />
              ) : (
                profile.name.charAt(0).toUpperCase()
              )}
            </div>
            <div>
              <p className="text-text-secondary text-sm">
                Avatar is automatically synced from your social login provider (Google, GitHub, etc.)
              </p>
              <p className="text-text-muted text-xs mt-1">
                Last updated: {profile.updatedAt ? new Date(profile.updatedAt).toLocaleDateString() : "Never"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Basic Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="w-5 h-5 text-brand-500" />
            Basic Information
          </CardTitle>
          <CardDescription>Update your public profile information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Email (read-only) */}
          <div className="space-y-2">
            <Label htmlFor="email" className="flex items-center gap-2">
              <Mail className="w-4 h-4" />
              Email Address
            </Label>
            <Input id="email" value={profile.email} disabled className="bg-bg-tertiary opacity-60" />
            <p className="text-xs text-text-muted">Email cannot be changed here. Contact support to update.</p>
          </div>

          <Separator />

          {/* Display Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Display Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleChange("name", e.target.value)}
              onBlur={() => handleBlur("name")}
              placeholder="Your full name"
              className={cn(touched.name && errors.name && "border-error")}
            />
            {touched.name && errors.name && (
              <p className="text-xs text-error flex items-center gap-1">
                <AlertCircle className="w-3 h-3" />
                {errors.name}
              </p>
            )}
          </div>

          {/* Username */}
          <div className="space-y-2">
            <Label htmlFor="username" className="flex items-center gap-2">
              <AtSign className="w-4 h-4" />
              Username *
            </Label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">@</span>
              <Input
                id="username"
                value={formData.username}
                onChange={(e) =>
                  handleChange("username", e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""))
                }
                onBlur={() => handleBlur("username")}
                placeholder="username"
                className={cn("pl-7", touched.username && errors.username && "border-error")}
              />
            </div>
            {touched.username && errors.username ? (
              <p className="text-xs text-error flex items-center gap-1">
                <AlertCircle className="w-3 h-3" />
                {errors.username}
              </p>
            ) : (
              <p className="text-xs text-text-muted">
                Your public profile URL: /u/{formData.username || "username"}
              </p>
            )}
          </div>

          {/* Bio */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="bio">Bio</Label>
              <span
                className={cn(
                  "text-xs",
                  formData.bio.length > MAX_BIO_LENGTH ? "text-error" : "text-text-muted"
                )}
              >
                {formData.bio.length}/{MAX_BIO_LENGTH}
              </span>
            </div>
            <Textarea
              id="bio"
              value={formData.bio}
              onChange={(e) => handleChange("bio", e.target.value.slice(0, MAX_BIO_LENGTH))}
              placeholder="Tell others about yourself..."
              rows={4}
            />
          </div>
        </CardContent>
      </Card>

      {/* Work & Location */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Briefcase className="w-5 h-5 text-brand-500" />
            Work & Location
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Company */}
            <div className="space-y-2">
              <Label htmlFor="company" className="flex items-center gap-2">
                <Building2 className="w-4 h-4" />
                Company
              </Label>
              <Input
                id="company"
                value={formData.company}
                onChange={(e) => handleChange("company", e.target.value)}
                placeholder="Acme Inc"
              />
            </div>

            {/* Job Title */}
            <div className="space-y-2">
              <Label htmlFor="jobTitle">Job Title</Label>
              <Input
                id="jobTitle"
                value={formData.jobTitle}
                onChange={(e) => handleChange("jobTitle", e.target.value)}
                placeholder="Software Engineer"
              />
            </div>
          </div>

          {/* Location */}
          <div className="space-y-2">
            <Label htmlFor="location" className="flex items-center gap-2">
              <MapPin className="w-4 h-4" />
              Location
            </Label>
            <Input
              id="location"
              value={formData.location}
              onChange={(e) => handleChange("location", e.target.value)}
              placeholder="San Francisco, CA"
            />
          </div>
        </CardContent>
      </Card>

      {/* Social Links */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="w-5 h-5 text-brand-500" />
            Online Presence
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Website */}
          <div className="space-y-2">
            <Label htmlFor="website" className="flex items-center gap-2">
              <LinkIcon className="w-4 h-4" />
              Personal Website
            </Label>
            <Input
              id="website"
              value={formData.website}
              onChange={(e) => handleChange("website", e.target.value)}
              onBlur={() => handleBlur("website")}
              placeholder="https://yourwebsite.com"
              className={cn(touched.website && errors.website && "border-error")}
            />
            {touched.website && errors.website && (
              <p className="text-xs text-error flex items-center gap-1">
                <AlertCircle className="w-3 h-3" />
                {errors.website}
              </p>
            )}
          </div>

          <Separator />

          {/* GitHub */}
          <div className="space-y-2">
            <Label htmlFor="githubUrl" className="flex items-center gap-2">
              <Github className="w-4 h-4" />
              GitHub
            </Label>
            <Input
              id="githubUrl"
              value={formData.githubUrl}
              onChange={(e) => handleChange("githubUrl", e.target.value)}
              onBlur={() => handleBlur("githubUrl")}
              placeholder="https://github.com/username"
              className={cn(touched.githubUrl && errors.githubUrl && "border-error")}
            />
          </div>

          {/* Twitter */}
          <div className="space-y-2">
            <Label htmlFor="twitterUrl" className="flex items-center gap-2">
              <Twitter className="w-4 h-4" />
              Twitter / X
            </Label>
            <Input
              id="twitterUrl"
              value={formData.twitterUrl}
              onChange={(e) => handleChange("twitterUrl", e.target.value)}
              onBlur={() => handleBlur("twitterUrl")}
              placeholder="https://twitter.com/username"
              className={cn(touched.twitterUrl && errors.twitterUrl && "border-error")}
            />
          </div>

          {/* LinkedIn */}
          <div className="space-y-2">
            <Label htmlFor="linkedinUrl" className="flex items-center gap-2">
              <Linkedin className="w-4 h-4" />
              LinkedIn
            </Label>
            <Input
              id="linkedinUrl"
              value={formData.linkedinUrl}
              onChange={(e) => handleChange("linkedinUrl", e.target.value)}
              onBlur={() => handleBlur("linkedinUrl")}
              placeholder="https://linkedin.com/in/username"
              className={cn(touched.linkedinUrl && errors.linkedinUrl && "border-error")}
            />
          </div>
        </CardContent>
      </Card>

      {/* Cover Image */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ImageIcon className="w-5 h-5 text-brand-500" />
            Cover Image
          </CardTitle>
          <CardDescription>Customize your profile header</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="h-32 rounded-lg bg-gradient-to-br from-brand-500/20 to-purple-500/20 border border-dashed border-border-default flex items-center justify-center">
            <div className="text-center">
              <ImageIcon className="w-8 h-8 text-text-muted mx-auto mb-2" />
              <p className="text-sm text-text-muted">Cover image upload coming soon</p>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="coverImageUrl">Cover Image URL</Label>
            <Input
              id="coverImageUrl"
              value={formData.coverImageUrl}
              onChange={(e) => handleChange("coverImageUrl", e.target.value)}
              placeholder="https://example.com/cover.jpg"
            />
            <p className="text-xs text-text-muted">Recommended: 1500x500 pixels</p>
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end gap-4">
        <Button
          variant="outline"
          onClick={() =>
            setFormData({
              name: profile.name || "",
              username: profile.username || "",
              bio: profile.bio || "",
              location: "",
              company: profile.companyName || "",
              jobTitle: "",
              website: "",
              twitterUrl: "",
              githubUrl: "",
              linkedinUrl: "",
              coverImageUrl: "",
            })
          }
          disabled={!isDirty || isLoading}
        >
          Reset
        </Button>
        <Button onClick={handleSave} disabled={!isDirty || isLoading} className="gap-2">
          {isLoading ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="w-4 h-4" />
              Save Changes
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

// ============================================================================
// Account Tab
// ============================================================================

function AccountTab() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isChangingPassword, setIsChangingPassword] = useState(false);

  const handleChangePassword = async () => {
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }
    if (newPassword.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }

    setIsChangingPassword(true);
    try {
      await usersApi.changePassword({
        currentPassword,
        newPassword,
      });
      toast.success("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      toast.error(err.message || "Failed to change password");
    } finally {
      setIsChangingPassword(false);
    }
  };

  const canChangePassword =
    currentPassword && newPassword && confirmPassword && newPassword === confirmPassword;

  return (
    <div className="space-y-6">
      {/* Password Change */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lock className="w-5 h-5 text-brand-500" />
            Change Password
          </CardTitle>
          <CardDescription>Update your account password</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">Current Password</Label>
            <div className="relative">
              <Input
                id="currentPassword"
                type={showPassword ? "text" : "password"}
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="newPassword">New Password</Label>
            <Input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="••••••••"
            />
            <p className="text-xs text-text-muted">Must be at least 8 characters</p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">Confirm New Password</Label>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="••••••••"
            />
            {confirmPassword && newPassword !== confirmPassword && (
              <p className="text-xs text-error">Passwords do not match</p>
            )}
          </div>

          <Button
            onClick={handleChangePassword}
            disabled={!canChangePassword || isChangingPassword}
            className="gap-2"
          >
            {isChangingPassword ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Changing...
              </>
            ) : (
              <>
                <Lock className="w-4 h-4" />
                Change Password
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {/* Sessions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Smartphone className="w-5 h-5 text-brand-500" />
            Active Sessions
          </CardTitle>
          <CardDescription>Manage your active login sessions</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-text-muted">
            Session management coming soon. You can currently sign out of all devices by changing your password.
          </p>
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card className="border-error/30">
        <CardHeader>
          <CardTitle className="text-error flex items-center gap-2">
            <Trash2 className="w-5 h-5" />
            Danger Zone
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="font-medium text-text-primary">Delete Account</h4>
              <p className="text-sm text-text-muted">
                Permanently delete your account and all associated data
              </p>
            </div>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive">Delete Account</Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete your account and remove
                    all your data from our servers.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction className="bg-error hover:bg-error/90">
                    Delete Account
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// ============================================================================
// Notifications Tab
// ============================================================================

function NotificationsTab() {
  const [settings, setSettings] = useState({
    emailNotifications: true,
    pushNotifications: false,
    marketingEmails: false,
    newFollower: true,
    functionPublished: true,
    mentions: true,
    weeklyDigest: true,
    securityAlerts: true,
  });

  const toggleSetting = (key: keyof typeof settings) => {
    setSettings((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="w-5 h-5 text-brand-500" />
            Notification Preferences
          </CardTitle>
          <CardDescription>Choose how you want to be notified</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Email Notifications */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="emailNotifications" className="text-base">
                Email Notifications
              </Label>
              <p className="text-sm text-text-muted">Receive notifications via email</p>
            </div>
            <Switch
              id="emailNotifications"
              checked={settings.emailNotifications}
              onCheckedChange={() => toggleSetting("emailNotifications")}
            />
          </div>

          <Separator />

          {/* Push Notifications */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="pushNotifications" className="text-base">
                Push Notifications
              </Label>
              <p className="text-sm text-text-muted">Receive browser push notifications</p>
            </div>
            <Switch
              id="pushNotifications"
              checked={settings.pushNotifications}
              onCheckedChange={() => toggleSetting("pushNotifications")}
            />
          </div>

          <Separator />

          {/* Activity Notifications */}
          <div className="space-y-4">
            <h4 className="font-medium text-text-primary">Activity Notifications</h4>

            {[
              { key: "newFollower", label: "New followers", description: "When someone follows you" },
              { key: "functionPublished", label: "Function activity", description: "When your functions get interactions" },
              { key: "mentions", label: "Mentions", description: "When someone mentions you" },
            ].map(({ key, label, description }) => (
              <div key={key} className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor={key} className="text-sm">
                    {label}
                  </Label>
                  <p className="text-xs text-text-muted">{description}</p>
                </div>
                <Switch
                  id={key}
                  checked={settings[key as keyof typeof settings]}
                  onCheckedChange={() => toggleSetting(key as keyof typeof settings)}
                />
              </div>
            ))}
          </div>

          <Separator />

          {/* Email Digests */}
          <div className="space-y-4">
            <h4 className="font-medium text-text-primary">Email Digests</h4>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="weeklyDigest" className="text-sm">
                  Weekly Digest
                </Label>
                <p className="text-xs text-text-muted">Weekly summary of your activity</p>
              </div>
              <Switch
                id="weeklyDigest"
                checked={settings.weeklyDigest}
                onCheckedChange={() => toggleSetting("weeklyDigest")}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="marketingEmails" className="text-sm">
                  Marketing Emails
                </Label>
                <p className="text-xs text-text-muted">Product updates and promotions</p>
              </div>
              <Switch
                id="marketingEmails"
                checked={settings.marketingEmails}
                onCheckedChange={() => toggleSetting("marketingEmails")}
              />
            </div>
          </div>

          <Separator />

          {/* Security */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="securityAlerts" className="text-base">
                Security Alerts
              </Label>
              <p className="text-sm text-text-muted">Important security notifications</p>
            </div>
            <Switch
              id="securityAlerts"
              checked={settings.securityAlerts}
              onCheckedChange={() => toggleSetting("securityAlerts")}
              disabled
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// ============================================================================
// Privacy Tab
// ============================================================================

function PrivacyTab() {
  const [settings, setSettings] = useState({
    profileVisibility: "public",
    showEmail: false,
    showActivity: true,
    allowMentions: true,
    allowFollowing: true,
    dataSharing: false,
  });

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-brand-500" />
            Privacy Settings
          </CardTitle>
          <CardDescription>Control your privacy and visibility</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Profile Visibility */}
          <div className="space-y-3">
            <Label className="text-base">Profile Visibility</Label>
            <div className="grid grid-cols-3 gap-3">
              {[
                { value: "public", label: "Public", description: "Everyone can see" },
                { value: "followers", label: "Followers", description: "Only followers" },
                { value: "private", label: "Private", description: "Only you" },
              ].map((option) => (
                <button
                  key={option.value}
                  onClick={() => setSettings((prev) => ({ ...prev, profileVisibility: option.value }))}
                  className={cn(
                    "p-3 rounded-lg border text-left transition-all",
                    settings.profileVisibility === option.value
                      ? "border-brand-500 bg-brand-500/10"
                      : "border-border-subtle hover:border-border-default"
                  )}
                >
                  <div className="font-medium text-text-primary">{option.label}</div>
                  <div className="text-xs text-text-muted">{option.description}</div>
                </button>
              ))}
            </div>
          </div>

          <Separator />

          {/* Profile Details */}
          <div className="space-y-4">
            <h4 className="font-medium text-text-primary">Profile Details</h4>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="showEmail" className="text-sm">
                  Show Email Address
                </Label>
                <p className="text-xs text-text-muted">Display email on your public profile</p>
              </div>
              <Switch
                id="showEmail"
                checked={settings.showEmail}
                onCheckedChange={() => setSettings((prev) => ({ ...prev, showEmail: !prev.showEmail }))}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="showActivity" className="text-sm">
                  Show Activity
                </Label>
                <p className="text-xs text-text-muted">Display your recent activity feed</p>
              </div>
              <Switch
                id="showActivity"
                checked={settings.showActivity}
                onCheckedChange={() => setSettings((prev) => ({ ...prev, showActivity: !prev.showActivity }))}
              />
            </div>
          </div>

          <Separator />

          {/* Interactions */}
          <div className="space-y-4">
            <h4 className="font-medium text-text-primary">Interactions</h4>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="allowMentions" className="text-sm">
                  Allow Mentions
                </Label>
                <p className="text-xs text-text-muted">Let others mention you in posts</p>
              </div>
              <Switch
                id="allowMentions"
                checked={settings.allowMentions}
                onCheckedChange={() => setSettings((prev) => ({ ...prev, allowMentions: !prev.allowMentions }))}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="allowFollowing" className="text-sm">
                  Allow Following
                </Label>
                <p className="text-xs text-text-muted">Let others follow your profile</p>
              </div>
              <Switch
                id="allowFollowing"
                checked={settings.allowFollowing}
                onCheckedChange={() => setSettings((prev) => ({ ...prev, allowFollowing: !prev.allowFollowing }))}
              />
            </div>
          </div>

          <Separator />

          {/* Data */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="dataSharing" className="text-base">
                Data Sharing
              </Label>
              <p className="text-sm text-text-muted">Share anonymous usage data to help improve our services</p>
            </div>
            <Switch
              id="dataSharing"
              checked={settings.dataSharing}
              onCheckedChange={() => setSettings((prev) => ({ ...prev, dataSharing: !prev.dataSharing }))}
            />
          </div>
        </CardContent>
      </Card>

      {/* Data Export */}
      <Card>
        <CardHeader>
          <CardTitle>Your Data</CardTitle>
          <CardDescription>Download or manage your personal data</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="font-medium text-text-primary">Export Your Data</h4>
              <p className="text-sm text-text-muted">Download a copy of your data</p>
            </div>
            <Button variant="outline">Request Export</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default ProfileSettingsPage;

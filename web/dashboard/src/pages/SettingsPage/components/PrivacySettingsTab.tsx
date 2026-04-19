/**
 * Settings Tab Component
 *
 * Displays profile settings for the user's own profile.
 * Includes visibility preferences, notification settings, and privacy controls.
 */

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import {
  Settings,
  Eye,
  Bell,
  Shield,
  User,
  Link as LinkIcon,
  Save,
  Loader2,
  Globe,
  Lock,
  Users,
  Check,
  AlertCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { tabContentVariants } from "@/pages/ProfilePage/animations";
import { usersApi, type UpdateProfileRequest } from "@/api/users";
import { useAuthStore } from "@/stores/authStore";
import { toast } from "sonner";
import type { UserProfile } from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

interface PrivacySettingsTabProps {
  profile?: UserProfile;
}

interface ProfileSettings {
  // Visibility settings
  profileVisibility: "public" | "followers" | "private";
  showEmail: boolean;
  showLocation: boolean;
  showCompany: boolean;
  showActivity: boolean;
  showAnalytics: boolean;

  // Notification settings
  emailNotifications: boolean;
  pushNotifications: boolean;
  notifyOnFollow: boolean;
  notifyOnMention: boolean;
  notifyOnFunctionUsage: boolean;
  notifyOnReviews: boolean;
  weeklyDigest: boolean;

  // Privacy settings
  allowTagging: boolean;
  allowIndexing: boolean;
  showLastActive: boolean;
}

const defaultSettings: ProfileSettings = {
  profileVisibility: "public",
  showEmail: false,
  showLocation: true,
  showCompany: true,
  showActivity: true,
  showAnalytics: true,
  emailNotifications: true,
  pushNotifications: false,
  notifyOnFollow: true,
  notifyOnMention: true,
  notifyOnFunctionUsage: true,
  notifyOnReviews: true,
  weeklyDigest: true,
  allowTagging: true,
  allowIndexing: true,
  showLastActive: true,
};

export function PrivacySettingsTab({ profile }: PrivacySettingsTabProps = {}) {
  const currentUser = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const username = profile?.username || currentUser?.username || "";

  // Fetch settings from backend
  const { data: settingsData, isLoading: isLoadingSettings } = useQuery({
    queryKey: ["my-settings"],
    queryFn: async () => {
      const response = await usersApi.getMySettings();
      return response.settings as unknown as ProfileSettings;
    },
  });

  // Local state for settings
  const [settings, setSettings] = useState<ProfileSettings>(defaultSettings);
  const [socialLinks, setSocialLinks] = useState({
    website: profile?.website || "",
    github: profile?.socialLinks?.github || "",
    twitter: profile?.socialLinks?.twitter || "",
    linkedin: profile?.socialLinks?.linkedin || "",
  });

  // Update local state when data is fetched
  useEffect(() => {
    if (settingsData) {
      setSettings((prev) => ({ ...prev, ...settingsData }));
    }
  }, [settingsData]);

  // Update social links when profile changes
  useEffect(() => {
    if (profile) {
      setSocialLinks({
        website: profile.website || "",
        github: profile.socialLinks?.github || "",
        twitter: profile.socialLinks?.twitter || "",
        linkedin: profile.socialLinks?.linkedin || "",
      });
    }
  }, [profile]);

  // Mutations
  const updateVisibilityMutation = useMutation({
    mutationFn: (data: Partial<ProfileSettings>) => usersApi.updateMyVisibilitySettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["my-settings"] });
      toast.success("Visibility settings saved");
    },
    onError: () => toast.error("Failed to save visibility settings"),
  });

  const updateNotificationsMutation = useMutation({
    mutationFn: (data: Record<string, boolean>) => usersApi.updateMyNotificationSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["my-settings"] });
      toast.success("Notification settings saved");
    },
    onError: () => toast.error("Failed to save notification settings"),
  });

  const updatePrivacyMutation = useMutation({
    mutationFn: (data: Record<string, boolean>) => usersApi.updateMyPrivacySettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["my-settings"] });
      toast.success("Privacy settings saved");
    },
    onError: () => toast.error("Failed to save privacy settings"),
  });

  const updateSocialLinksMutation = useMutation({
    mutationFn: async () => {
      const socialUpdate: UpdateProfileRequest = {};
      if (socialLinks.website !== profile?.website) {
        socialUpdate.website = socialLinks.website;
      }
      if (socialLinks.github !== profile?.socialLinks?.github) {
        socialUpdate.githubUrl = socialLinks.github;
      }
      if (socialLinks.twitter !== profile?.socialLinks?.twitter) {
        socialUpdate.twitterUrl = socialLinks.twitter;
      }
      if (socialLinks.linkedin !== profile?.socialLinks?.linkedin) {
        socialUpdate.linkedinUrl = socialLinks.linkedin;
      }

      if (Object.keys(socialUpdate).length > 0) {
        await usersApi.updateMe(socialUpdate);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enhanced-profile", username] });
      toast.success("Social links saved");
    },
    onError: () => toast.error("Failed to save social links"),
  });

  const updateSetting = <K extends keyof ProfileSettings>(
    key: K,
    value: ProfileSettings[K]
  ) => {
    setSettings((prev) => ({ ...prev, [key]: value }));
  };

  const updateSocialLink = (key: keyof typeof socialLinks, value: string) => {
    setSocialLinks((prev) => ({ ...prev, [key]: value }));
  };

  const handleSaveVisibility = () => {
    updateVisibilityMutation.mutate({
      profileVisibility: settings.profileVisibility,
      showEmail: settings.showEmail,
      showLocation: settings.showLocation,
      showCompany: settings.showCompany,
      showActivity: settings.showActivity,
      showAnalytics: settings.showAnalytics,
    });
  };

  const handleSaveNotifications = () => {
    updateNotificationsMutation.mutate({
      emailNotifications: settings.emailNotifications,
      pushNotifications: settings.pushNotifications,
      notifyOnFollow: settings.notifyOnFollow,
      notifyOnMention: settings.notifyOnMention,
      notifyOnFunctionUsage: settings.notifyOnFunctionUsage,
      notifyOnReviews: settings.notifyOnReviews,
      weeklyDigest: settings.weeklyDigest,
    });
  };

  const handleSavePrivacy = () => {
    updatePrivacyMutation.mutate({
      allowTagging: settings.allowTagging,
      allowIndexing: settings.allowIndexing,
      showLastActive: settings.showLastActive,
    });
  };

  const handleSaveSocialLinks = () => {
    updateSocialLinksMutation.mutate();
  };

  const handleSaveAll = async () => {
    await Promise.all([
      updateVisibilityMutation.mutateAsync({
        profileVisibility: settings.profileVisibility,
        showEmail: settings.showEmail,
        showLocation: settings.showLocation,
        showCompany: settings.showCompany,
        showActivity: settings.showActivity,
        showAnalytics: settings.showAnalytics,
      }),
      updateNotificationsMutation.mutateAsync({
        emailNotifications: settings.emailNotifications,
        pushNotifications: settings.pushNotifications,
        notifyOnFollow: settings.notifyOnFollow,
        notifyOnMention: settings.notifyOnMention,
        notifyOnFunctionUsage: settings.notifyOnFunctionUsage,
        notifyOnReviews: settings.notifyOnReviews,
        weeklyDigest: settings.weeklyDigest,
      }),
      updatePrivacyMutation.mutateAsync({
        allowTagging: settings.allowTagging,
        allowIndexing: settings.allowIndexing,
        showLastActive: settings.showLastActive,
      }),
      updateSocialLinksMutation.mutateAsync(),
    ]);
    toast.success("All settings saved successfully");
  };

  const getVisibilityIcon = (visibility: string) => {
    switch (visibility) {
      case "public":
        return <Globe className="w-4 h-4" />;
      case "followers":
        return <Users className="w-4 h-4" />;
      case "private":
        return <Lock className="w-4 h-4" />;
      default:
        return <Globe className="w-4 h-4" />;
    }
  };

  const isSaving = updateVisibilityMutation.isPending ||
                   updateNotificationsMutation.isPending ||
                   updatePrivacyMutation.isPending ||
                   updateSocialLinksMutation.isPending;

  if (isLoadingSettings) {
    return (
      <motion.div
        variants={tabContentVariants}
        initial="hidden"
        animate="visible"
        exit="exit"
        className="space-y-6 px-4 md:px-8 pb-8"
      >
        <div className="animate-pulse space-y-4">
          <div className="h-32 bg-bg-secondary rounded-lg" />
          <div className="h-48 bg-bg-secondary rounded-lg" />
          <div className="h-48 bg-bg-secondary rounded-lg" />
        </div>
      </motion.div>
    );
  }

  return (
    <TooltipProvider>
      <motion.div
        variants={tabContentVariants}
        initial="hidden"
        animate="visible"
        exit="exit"
        className="space-y-6 px-4 md:px-8 pb-8"
      >
        {/* Header with save button */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-display text-xl font-semibold bg-gradient-to-r from-brand-500 via-ff-afterburner to-brand-400 bg-clip-text text-transparent">
              Profile Settings
            </h2>
            <p className="text-sm text-text-muted">
              Manage your profile visibility, notifications, and privacy preferences
            </p>
          </div>
          <Button
            onClick={handleSaveAll}
            disabled={isSaving}
            className="ff-btn-velocity min-w-[120px]"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            {isSaving ? "Saving..." : "Save All"}
          </Button>
        </div>

        {/* Profile Visibility Section */}
        <Card className="ff-card-velocity border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Eye className="w-5 h-5 text-brand-500" />
              Profile Visibility
            </CardTitle>
            <CardDescription>
              Control who can see your profile and what information is displayed
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Visibility Level */}
            <div className="space-y-3">
              <Label className="text-sm font-medium">Profile Visibility Level</Label>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {(["public", "followers", "private"] as const).map((level) => (
                  <button
                    key={level}
                    onClick={() => updateSetting("profileVisibility", level)}
                    className={`flex items-center gap-3 p-3 rounded-lg border transition-all ${
                      settings.profileVisibility === level
                        ? "border-brand-500 bg-brand-500/10"
                        : "border-border-subtle hover:border-border-default bg-bg-secondary"
                    }`}
                  >
                    <div className={`p-2 rounded-md ${
                      settings.profileVisibility === level
                        ? "bg-brand-500 text-white"
                        : "bg-bg-tertiary text-text-secondary"
                    }`}>
                      {getVisibilityIcon(level)}
                    </div>
                    <div className="text-left">
                      <p className={`font-medium capitalize ${
                        settings.profileVisibility === level
                          ? "text-brand-400"
                          : "text-text-primary"
                      }`}>
                        {level}
                      </p>
                      <p className="text-xs text-text-muted">
                        {level === "public" && "Everyone can see"}
                        {level === "followers" && "Followers only"}
                        {level === "private" && "Only you"}
                      </p>
                    </div>
                    {settings.profileVisibility === level && (
                      <Check className="w-4 h-4 text-brand-500 ml-auto" />
                    )}
                  </button>
                ))}
              </div>
            </div>

            <div className="ff-divider-flame" />

            {/* Field Visibility Toggles */}
            <div className="space-y-4">
              <Label className="text-sm font-medium">Field Visibility</Label>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <User className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">Email Address</p>
                    <p className="text-sm text-text-muted">Show email on your profile</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showEmail}
                  onCheckedChange={(checked) => updateSetting("showEmail", checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Globe className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">Location</p>
                    <p className="text-sm text-text-muted">Show your location</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showLocation}
                  onCheckedChange={(checked) => updateSetting("showLocation", checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Users className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">Company</p>
                    <p className="text-sm text-text-muted">Show your company/organization</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showCompany}
                  onCheckedChange={(checked) => updateSetting("showCompany", checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Eye className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">Activity Status</p>
                    <p className="text-sm text-text-muted">Show your recent activity</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showActivity}
                  onCheckedChange={(checked) => updateSetting("showActivity", checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Shield className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">Analytics</p>
                    <p className="text-sm text-text-muted">Show function usage analytics</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showAnalytics}
                  onCheckedChange={(checked) => updateSetting("showAnalytics", checked)}
                />
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveVisibility}
                disabled={updateVisibilityMutation.isPending}
                className="ff-btn-velocity"
              >
                {updateVisibilityMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Save Visibility
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Social Links Section */}
        <Card className="ff-card-velocity border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <LinkIcon className="w-5 h-5 text-brand-500" />
              Social Links
            </CardTitle>
            <CardDescription>
              Manage your external links and social media profiles
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="website" className="text-sm font-medium">
                Personal Website
              </Label>
              <Input
                id="website"
                type="url"
                placeholder="https://yourwebsite.com"
                value={socialLinks.website}
                onChange={(e) => updateSocialLink("website", e.target.value)}
                className="bg-bg-secondary"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="github" className="text-sm font-medium">
                GitHub Profile
              </Label>
              <Input
                id="github"
                type="url"
                placeholder="https://github.com/username"
                value={socialLinks.github}
                onChange={(e) => updateSocialLink("github", e.target.value)}
                className="bg-bg-secondary"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="twitter" className="text-sm font-medium">
                Twitter/X Profile
              </Label>
              <Input
                id="twitter"
                type="url"
                placeholder="https://twitter.com/username"
                value={socialLinks.twitter}
                onChange={(e) => updateSocialLink("twitter", e.target.value)}
                className="bg-bg-secondary"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="linkedin" className="text-sm font-medium">
                LinkedIn Profile
              </Label>
              <Input
                id="linkedin"
                type="url"
                placeholder="https://linkedin.com/in/username"
                value={socialLinks.linkedin}
                onChange={(e) => updateSocialLink("linkedin", e.target.value)}
                className="bg-bg-secondary"
              />
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveSocialLinks}
                disabled={updateSocialLinksMutation.isPending}
                className="ff-btn-velocity"
              >
                {updateSocialLinksMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Save Links
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card className="ff-card-velocity border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Bell className="w-5 h-5 text-brand-500" />
              Notification Preferences
            </CardTitle>
            <CardDescription>
              Choose what notifications you want to receive
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Master Toggles */}
            <div className="flex flex-wrap gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  checked={settings.emailNotifications}
                  onCheckedChange={(checked) => updateSetting("emailNotifications", checked)}
                />
                <Label className="text-sm font-medium cursor-pointer">
                  Email Notifications
                </Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch
                  checked={settings.pushNotifications}
                  onCheckedChange={(checked) => updateSetting("pushNotifications", checked)}
                />
                <Label className="text-sm font-medium cursor-pointer">
                  Push Notifications
                </Label>
              </div>
            </div>

            <div className="ff-divider-flame" />

            {/* Notification Types */}
            <div className="space-y-4">
              <Label className="text-sm font-medium">Notification Types</Label>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">New Followers</p>
                  <p className="text-sm text-text-muted">When someone follows you</p>
                </div>
                <Switch
                  checked={settings.notifyOnFollow}
                  onCheckedChange={(checked) => updateSetting("notifyOnFollow", checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">Mentions</p>
                  <p className="text-sm text-text-muted">When you're mentioned in comments</p>
                </div>
                <Switch
                  checked={settings.notifyOnMention}
                  onCheckedChange={(checked) => updateSetting("notifyOnMention", checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">Function Usage</p>
                  <p className="text-sm text-text-muted">When someone uses your functions</p>
                </div>
                <Switch
                  checked={settings.notifyOnFunctionUsage}
                  onCheckedChange={(checked) => updateSetting("notifyOnFunctionUsage", checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">Reviews</p>
                  <p className="text-sm text-text-muted">When someone reviews your functions</p>
                </div>
                <Switch
                  checked={settings.notifyOnReviews}
                  onCheckedChange={(checked) => updateSetting("notifyOnReviews", checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">Weekly Digest</p>
                  <p className="text-sm text-text-muted">Weekly summary of your activity</p>
                </div>
                <Switch
                  checked={settings.weeklyDigest}
                  onCheckedChange={(checked) => updateSetting("weeklyDigest", checked)}
                  disabled={!settings.emailNotifications}
                />
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveNotifications}
                disabled={updateNotificationsMutation.isPending}
                className="ff-btn-velocity"
              >
                {updateNotificationsMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Save Notifications
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Privacy Settings */}
        <Card className="ff-card-velocity border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Shield className="w-5 h-5 text-brand-500" />
              Privacy & Security
            </CardTitle>
            <CardDescription>
              Control your privacy and security preferences
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Users className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">Allow Tagging</p>
                  <p className="text-sm text-text-muted">Let others tag you in posts and comments</p>
                </div>
              </div>
              <Switch
                checked={settings.allowTagging}
                onCheckedChange={(checked) => updateSetting("allowTagging", checked)}
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Globe className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">Search Engine Indexing</p>
                  <p className="text-sm text-text-muted">Allow search engines to index your profile</p>
                </div>
              </div>
              <Switch
                checked={settings.allowIndexing}
                onCheckedChange={(checked) => updateSetting("allowIndexing", checked)}
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Eye className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">Last Active Status</p>
                  <p className="text-sm text-text-muted">Show when you were last active</p>
                </div>
              </div>
              <Switch
                checked={settings.showLastActive}
                onCheckedChange={(checked) => updateSetting("showLastActive", checked)}
              />
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSavePrivacy}
                disabled={updatePrivacyMutation.isPending}
                className="ff-btn-velocity"
              >
                {updatePrivacyMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Save Privacy
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Save Button at bottom */}
        <div className="flex items-center justify-end gap-4">
          <Button
            onClick={handleSaveAll}
            disabled={isSaving}
            size="lg"
            className="ff-btn-velocity min-w-[160px]"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            {isSaving ? "Saving..." : "Save All Changes"}
          </Button>
        </div>
      </motion.div>
    </TooltipProvider>
  );
}

export default PrivacySettingsTab;

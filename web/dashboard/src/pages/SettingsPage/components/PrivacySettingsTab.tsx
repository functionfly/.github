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
  CircleDot,
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
import { useCustomStatus, CUSTOM_STATUS_OPTIONS, type CustomStatusValue } from "@/hooks/useCustomStatus";

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
  const { status: customStatus, isLoading: isLoadingCustomStatus, setStatus } = useCustomStatus();

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
        className="settings-page space-y-6 px-4 md:px-8 pb-8"
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
        className="settings-page space-y-6 px-4 md:px-8 pb-8"
      >
        {/* Header with save button */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-display text-xl font-semibold text-white">
              Profile Settings
            </h2>
            <p className="text-sm text-gray-400">
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
        <Card className="settings-panel border-border-subtle">
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
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <LinkIcon className="w-5 h-5 text-brand-500" />
              Social Links
            </CardTitle>
            <CardDescription>
              Connect your social profiles to help others find you
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            {/* Personal Website */}
            <div className="group">
              <Label htmlFor="website" className="text-sm font-medium mb-2 block">
                Personal Website
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-brand-500/10 group-hover:text-brand-500 transition-colors">
                  <Globe className="w-4 h-4" />
                </div>
                <Input
                  id="website"
                  type="url"
                  placeholder="https://yourwebsite.com"
                  value={socialLinks.website}
                  onChange={(e) => updateSocialLink("website", e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* GitHub Profile */}
            <div className="group">
              <Label htmlFor="github" className="text-sm font-medium mb-2 block">
                GitHub Profile
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-[#333] group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                  </svg>
                </div>
                <Input
                  id="github"
                  type="url"
                  placeholder="https://github.com/username"
                  value={socialLinks.github}
                  onChange={(e) => updateSocialLink("github", e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* Twitter/X Profile */}
            <div className="group">
              <Label htmlFor="twitter" className="text-sm font-medium mb-2 block">
                Twitter/X Profile
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-black group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                  </svg>
                </div>
                <Input
                  id="twitter"
                  type="url"
                  placeholder="https://twitter.com/username"
                  value={socialLinks.twitter}
                  onChange={(e) => updateSocialLink("twitter", e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* LinkedIn Profile */}
            <div className="group">
              <Label htmlFor="linkedin" className="text-sm font-medium mb-2 block">
                LinkedIn Profile
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-[#0077b5] group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
                  </svg>
                </div>
                <Input
                  id="linkedin"
                  type="url"
                  placeholder="https://linkedin.com/in/username"
                  value={socialLinks.linkedin}
                  onChange={(e) => updateSocialLink("linkedin", e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <Button
                onClick={handleSaveSocialLinks}
                disabled={updateSocialLinksMutation.isPending}
                className="ff-btn-velocity"
              >
                {updateSocialLinksMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                {updateSocialLinksMutation.isPending ? "Saving..." : "Save Links"}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Custom Status Section */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <CircleDot className="w-5 h-5 text-brand-500" />
              Presence Status
            </CardTitle>
            <CardDescription>
              Set how you appear to others. Choose "Auto" to let your activity determine your status.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoadingCustomStatus ? (
              <div className="flex items-center gap-2 text-text-muted">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="text-sm">Loading status...</span>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-2">
                  {CUSTOM_STATUS_OPTIONS.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => setStatus(option.value, option.emoji)}
                      disabled={option.value === customStatus.customStatus}
                      className={`flex flex-col items-center gap-1.5 p-3 rounded-lg border transition-all ${
                        option.value === customStatus.customStatus
                          ? "border-brand-500 bg-brand-500/10"
                          : "border-border-subtle hover:border-border-default bg-bg-secondary disabled:opacity-50"
                      }`}
                    >
                      <span className="text-xl">{option.emoji}</span>
                      <span className={`text-xs font-medium ${
                        option.value === customStatus.customStatus
                          ? "text-brand-400"
                          : "text-text-primary"
                      }`}>
                        {option.label}
                      </span>
                    </button>
                  ))}
                </div>
                {customStatus.customStatusEmoji && (
                  <div className="flex items-center gap-2 text-sm text-text-muted">
                    <span>Current status:</span>
                    <span className="text-lg">{customStatus.customStatusEmoji}</span>
                    <span className="font-medium capitalize">{customStatus.customStatus}</span>
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card className="settings-panel border-border-subtle">
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
        <Card className="settings-panel border-border-subtle">
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

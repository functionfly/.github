/**
 * ProfilePage Component
 *
 * A comprehensive, feature-rich profile page for FunctionFly users.
 * Includes tabs for Overview, Functions, Activity, Analytics, About, and Settings.
 *
 * @example
 * <ProfilePage username="johndoe" isOwnProfile={false} />
 *
 * @example
 * <ProfilePage username="currentuser" isOwnProfile={true} />
 */

import { useState, useEffect } from "react";
import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { User, Package, Activity, BarChart3, BookOpen, Settings } from "lucide-react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/common/Navbar";
import { EditProfileModal } from "@/components/profile/EditProfileModal";
import { AvatarPicker } from "@/components/profile/AvatarPicker";
import type {
  UserProfile,
  ProfileTab,
  UserActivity,
  Achievement as AchievementType,
  Skill,
  ProfileAnalytics,
  ActivityType,
} from "@/types";
import {
  usersApi,
  type UserAnalyticsResponse,
  type UserAchievementsResponse,
  type UserActivityResponse,
  type UserSkillsResponse,
} from "@/api/users";
import { useAuthStore } from "@/stores/authStore";
import { toast } from "sonner";
import { Footer } from "@/pages/LandingPage/components";
import { AlertCircle } from "lucide-react";
import { format } from "date-fns";

import { transformToUserProfile } from "./transformers";
import {
  ProfileHeaderSkeleton,
  StatsOverviewSkeleton,
  TabContentSkeleton,
} from "./components/Skeletons";
import { ProfileHeader } from "./components/ProfileHeader";
import { StatsOverview } from "./components/StatsOverview";
import { OverviewTab } from "./components/tabs/OverviewTab";
import { FunctionsTab } from "./components/tabs/FunctionsTab";
import { ActivityTab } from "./components/tabs/ActivityTab";
import { AnalyticsTab } from "./components/tabs/AnalyticsTab";
import { AboutTab } from "./components/tabs/AboutTab";
import { SettingsTab } from "./components/tabs/SettingsTab";
import { registryApi } from "@/api/registry";

// ============================================================================
// Main Profile Page Component
// ============================================================================

export interface ProfilePageProps {
  username?: string;
  isOwnProfile?: boolean;
}

export function ProfilePage({
  username: propUsername,
  isOwnProfile: propIsOwnProfile,
}: ProfilePageProps = {}) {
  const { username: paramUsername } = useParams<{ username: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const urlTab = searchParams.get("tab") as ProfileTab | null;
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  const username = propUsername || paramUsername;

  const isOwnProfile =
    propIsOwnProfile ??
    (!!currentUser &&
      !!username &&
      (currentUser.username === username || username === "me"));

  const [activeTab, setActiveTab] = useState<ProfileTab>(urlTab || "overview");
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isAvatarPickerOpen, setIsAvatarPickerOpen] = useState(false);

  useEffect(() => {
    if (urlTab && urlTab !== activeTab) {
      setActiveTab(urlTab);
    }
  }, [urlTab]);

  const handleTabChange = (value: string) => {
    const newTab = value as ProfileTab;
    setActiveTab(newTab);
    setSearchParams({ tab: newTab });
  };

  const {
    data: profile,
    isLoading,
    isError,
    error,
  } = useQuery<UserProfile>({
    queryKey: ["enhanced-profile", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");

      // Use authenticated endpoint for own profile, public endpoint for others
      const profileApiCall = isOwnProfile
        ? usersApi.getMe()
        : usersApi.getPublicProfile(username);

      const [profileResponse, functionsResponse] = await Promise.all([
        profileApiCall,
        registryApi.getFunctions({ author: username, limit: 100 }),
      ]);

      return transformToUserProfile(
        profileResponse,
        functionsResponse.functions || []
      );
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const { data: analyticsResponse } = useQuery<UserAnalyticsResponse>({
    queryKey: ["profile-analytics", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      try {
        return await usersApi.getUserAnalytics(username);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response
          ?.status;
        if (status === 404) {
          return {
            executionStats: {
              totalExecutions: 0,
              totalUniqueUsers: 0,
              functionCount: 0,
              executionHistory: [],
            },
            popularFunctions: [],
            geographicStats: { regions: [] },
            deviceStats: { devices: [] },
          };
        }
        throw err;
      }
    },
    enabled:
      !!username &&
      (activeTab === "analytics" || activeTab === "overview"),
    staleTime: 5 * 60 * 1000,
  });

  const analytics: ProfileAnalytics | undefined = analyticsResponse
    ? {
        executionHistory:
          analyticsResponse.executionStats?.executionHistory?.map((h) => ({
            date: h.date,
            executions: Number(h.executions) || 0,
            uniqueUsers: Number(h.uniqueUsers) || 0,
          })) || [],
        popularFunctions:
          analyticsResponse.popularFunctions?.map((f) => ({
            functionId: f.id,
            name: f.name,
            executions: Number(f.executionCount) || 0,
            percentage: 0,
          })) || [],
        geographicDistribution:
          analyticsResponse.geographicStats?.regions?.map((r) => ({
            country: r.region,
            executions: Number(r.executions) || 0,
            percentage: 0,
          })) || [],
        deviceStats:
          analyticsResponse.deviceStats?.devices?.map((d) => ({
            device: d.device,
            percentage: 0,
          })) || [],
        browserStats: [],
      }
    : undefined;

  const { data: achievementsResponse } = useQuery<UserAchievementsResponse>({
    queryKey: ["profile-achievements", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      try {
        return await usersApi.getUserAchievements(username);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response
          ?.status;
        if (status === 404) {
          return { achievements: [], totalPoints: 0, available: 0 };
        }
        throw err;
      }
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  const { data: activityResponse } = useQuery<UserActivityResponse>({
    queryKey: ["profile-activity", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      try {
        return await usersApi.getUserActivity(username, { limit: 20 });
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response
          ?.status;
        if (status === 404) {
          return { activities: [], limit: 20, offset: 0, total: 0 };
        }
        throw err;
      }
    },
    enabled:
      !!username &&
      (activeTab === "activity" || activeTab === "overview"),
    staleTime: 5 * 60 * 1000,
  });

  const { data: skillsResponse } = useQuery<UserSkillsResponse>({
    queryKey: ["profile-skills", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      try {
        return await usersApi.getUserSkills(username);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response
          ?.status;
        if (status === 404) {
          return { skills: [] };
        }
        throw err;
      }
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  const achievementsData: AchievementType[] =
    achievementsResponse?.achievements?.map((a) => ({
      id: a.id,
      name: a.name,
      description: a.description,
      icon: a.icon || "Award",
      color: a.color || "blue",
      unlockedAt: a.earnedAt,
      tier: a.isCompleted
        ? a.points >= 500
          ? "platinum"
          : a.points >= 200
            ? "gold"
            : a.points >= 100
              ? "silver"
              : "bronze"
        : "bronze",
      progress: {
        current: a.progress,
        target: 100,
      },
    })) || [];

  const typeMap: Record<string, ActivityType> = {
    function_published: "function_published",
    function_updated: "function_updated",
    badge_earned: "achievement_earned",
    profile_updated: "deployment",
    review_submitted: "review_received",
    comment_posted: "contribution",
  };

  const rawActivity: UserActivity[] =
    activityResponse?.activities?.map((a) => ({
      id: a.id,
      type: typeMap[a.type] || "contribution",
      title: a.title,
      description: a.description || "",
      timestamp: a.createdAt,
      metadata: a.metadata,
    })) || [];

  // Prepend "Joined FunctionFly" activity with join date (synthetic, from profile)
  const joinedActivity: UserActivity | null =
    profile
      ? {
          id: `joined-${profile.id}`,
          type: "joined",
          title: "Joined FunctionFly",
          description: profile.createdAt
            ? format(new Date(profile.createdAt), "MMMM d, yyyy")
            : "",
          timestamp: profile.createdAt ?? new Date().toISOString(),
        }
      : null;
  const activityData: UserActivity[] = joinedActivity
    ? [joinedActivity, ...rawActivity]
    : rawActivity;

  const skillsData: Skill[] =
    skillsResponse?.skills?.map((s) => ({
      id: s.id,
      name: s.name,
      level: s.level,
      category: (s.category as Skill["category"]) || "concept",
    })) || [];

  const mergedProfile: UserProfile | undefined = profile
    ? {
        ...profile,
        achievements:
          achievementsData.length > 0 ? achievementsData : profile.achievements,
        recentActivity:
          activityData.length > 0 ? activityData : profile.recentActivity,
        skills: skillsData.length > 0 ? skillsData : profile.skills,
      }
    : undefined;

  const updateProfileMutation = useMutation({
    mutationFn: async (data: import("@/api/users").UpdateProfileRequest) => {
      await usersApi.updateMe(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enhanced-profile", username] });
      queryClient.invalidateQueries({ queryKey: ["my-profile"] });
      queryClient.invalidateQueries({ queryKey: ["my-settings"] });
      useAuthStore.getState().initialize();
    },
  });

  const tabs: { value: ProfileTab; label: string; icon: React.ReactNode }[] = [
    { value: "overview", label: "Overview", icon: <User className="w-4 h-4" /> },
    {
      value: "functions",
      label: "Functions",
      icon: <Package className="w-4 h-4" />,
    },
    {
      value: "activity",
      label: "Activity",
      icon: <Activity className="w-4 h-4" />,
    },
    {
      value: "analytics",
      label: "Analytics",
      icon: <BarChart3 className="w-4 h-4" />,
    },
    {
      value: "about",
      label: "About",
      icon: <BookOpen className="w-4 h-4" />,
    },
  ];

  if (isOwnProfile) {
    tabs.push({
      value: "settings",
      label: "Settings",
      icon: <Settings className="w-4 h-4" />,
    });
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar variant="dashboard" />

      <main className="pt-16 pb-16">
        <div className="max-w-7xl mx-auto">
          {isLoading && (
            <div className="bg-card rounded-xl border border-border-subtle">
              <ProfileHeaderSkeleton />
              <div className="px-4 md:px-8 py-4">
                <StatsOverviewSkeleton />
              </div>
              <div className="px-4 md:px-8 pb-8">
                <TabContentSkeleton />
              </div>
            </div>
          )}

          {isError && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex flex-col items-center justify-center py-24 text-center px-4"
            >
              <AlertCircle className="w-12 h-12 text-text-muted mb-4" />
              <h1 className="text-2xl font-bold text-text-primary mb-2">
                User not found
              </h1>
              <p className="text-text-secondary mb-6">
                {(error as Error)?.message?.includes("404")
                  ? `No user with username "@${username}" exists.`
                  : "Failed to load this profile. Please try again."}
              </p>
              <Link to="/registry">
                <Button variant="outline">Browse Functions</Button>
              </Link>
            </motion.div>
          )}

          {profile && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="bg-card rounded-xl border border-border-subtle overflow-hidden"
            >
              <ProfileHeader
                profile={profile}
                isOwnProfile={isOwnProfile}
                onEditProfile={() => setIsEditModalOpen(true)}
                onAvatarClick={
                  isOwnProfile ? () => setIsAvatarPickerOpen(true) : undefined
                }
              />

              <StatsOverview stats={profile.stats} />

              <div className="border-t border-border-subtle mt-4">
                <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
                  <div className="px-4 md:px-8 overflow-x-auto">
                    <TabsList className="bg-transparent border-b border-border-subtle rounded-none w-full justify-start h-auto p-0">
                      {tabs.map((tab) => (
                        <TabsTrigger
                          key={tab.value}
                          value={tab.value}
                          className="rounded-none border-b-2 border-transparent data-[state=active]:border-brand-500 data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-3 gap-2"
                        >
                          {tab.icon}
                          {tab.label}
                        </TabsTrigger>
                      ))}
                    </TabsList>
                  </div>

                  <TabsContent value="overview" className="m-0">
                    <OverviewTab profile={mergedProfile!} />
                  </TabsContent>
                  <TabsContent value="functions" className="m-0">
                    <FunctionsTab profile={mergedProfile!} />
                  </TabsContent>
                  <TabsContent value="activity" className="m-0">
                    <ActivityTab profile={mergedProfile!} />
                  </TabsContent>
                  <TabsContent value="analytics" className="m-0">
                    {analytics ? (
                      <AnalyticsTab analytics={analytics} />
                    ) : (
                      <div className="px-4 md:px-8 py-16 text-center">
                        <BarChart3 className="w-12 h-12 mx-auto text-text-muted mb-4" />
                        <p className="text-text-muted">Loading analytics...</p>
                      </div>
                    )}
                  </TabsContent>
                  <TabsContent value="about" className="m-0">
                    <AboutTab
                      profile={mergedProfile!}
                      isOwnProfile={isOwnProfile}
                      userSkills={skillsResponse?.skills}
                      onAddSkill={async (skill) => {
                        await usersApi.addSkill(
                          skill as import("@/api/users").AddSkillRequest
                        );
                        queryClient.invalidateQueries({
                          queryKey: ["profile-skills", username],
                        });
                      }}
                      onRemoveSkill={async (skillId) => {
                        await usersApi.removeSkill(skillId);
                        queryClient.invalidateQueries({
                          queryKey: ["profile-skills", username],
                        });
                      }}
                      isSkillsLoading={skillsResponse === undefined}
                    />
                  </TabsContent>
                  <TabsContent value="settings" className="m-0">
                    <SettingsTab />
                  </TabsContent>
                </Tabs>
              </div>
            </motion.div>
          )}
        </div>
      </main>

      {isOwnProfile && profile && (
        <EditProfileModal
          isOpen={isEditModalOpen}
          onClose={() => setIsEditModalOpen(false)}
          profile={profile}
          onSave={updateProfileMutation.mutateAsync}
          isLoading={updateProfileMutation.isPending}
        />
      )}

      {isOwnProfile && profile && (
        <AvatarPicker
          open={isAvatarPickerOpen}
          onOpenChange={setIsAvatarPickerOpen}
          currentAvatar={profile.avatar}
          onSelect={async (avatarUrl) => {
            try {
              await updateProfileMutation.mutateAsync({ avatar: avatarUrl });
              setIsAvatarPickerOpen(false);
              toast.success("Profile picture updated");
            } catch {
              toast.error("Failed to update profile picture");
            }
          }}
          isLoading={updateProfileMutation.isPending}
        />
      )}

      <Footer />
    </div>
  );
}

export default ProfilePage;

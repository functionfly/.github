import {
  usersApi,
  type MeResponse,
  type UserAchievementsResponse,
  type UserActivityResponse,
  type UserAnalyticsResponse,
  type UserContributionsResponse,
  type UserSkillsResponse,
} from '@/api/users';
import { Navbar } from '@/components/common/Navbar';
import { AvatarPicker } from '@/components/profile/AvatarPicker';
import { EditProfileModal } from '@/components/profile/EditProfileModal';
import { UserNotFoundView } from '@/components/ui/UserNotFoundView';
import { Footer } from '@/pages/LandingPage/components';
import { useAuthStore } from '@/stores/authStore';
import type {
  Achievement as AchievementType,
  ActivityType,
  ProfileAnalytics,
  ProfileTab,
  Skill,
  UserActivity,
  UserProfile,
} from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { format } from 'date-fns';
import { Activity, BarChart3, BookOpen, Package, Settings, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { usePageTitle } from '@/hooks';

import { registryApi } from '@/api/registry';
import { usePlan } from '@/hooks/usePlan';
import { usePublicBadges } from '@/hooks/useCertification';
import { SettingsContent } from '@/pages/SettingsPage/SettingsContent';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  AnnotationTag,
  StatusPill,
} from '@/components/containment';
import { ProfileHeader } from './components/ProfileHeader';
import {
  ProfileHeaderSkeleton,
  StatsOverviewSkeleton,
  TabContentSkeleton,
} from './components/Skeletons';
import { StatsOverview } from './components/StatsOverview';
import { AboutTab } from './components/tabs/AboutTab';
import { ActivityTab } from './components/tabs/ActivityTab';
import { AnalyticsTab } from './components/tabs/AnalyticsTab';
import { FunctionsTab } from './components/tabs/FunctionsTab';
import { OverviewTab } from './components/tabs/OverviewTab';
import { transformToUserProfile } from './transformers';

import './profile.css';

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
  const urlTab = searchParams.get('tab') as ProfileTab | null;
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);
  const authChecked = useAuthStore((state) => state.authChecked);
  const { isEnterprise: isCurrentUserEnterprise } = usePlan();

  const username = propUsername || paramUsername;

  const isOwnProfile =
    propIsOwnProfile ??
    (!!currentUser &&
      !!username &&
      (username === 'me' ||
        currentUser.username.toLowerCase() === String(username).toLowerCase()));

  const [activeTab, setActiveTab] = useState<ProfileTab>(urlTab || 'overview');
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
    queryKey: ['enhanced-profile', username, isOwnProfile],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      const profileApiCall = isOwnProfile ? usersApi.getMe() : usersApi.getPublicProfile(username);
      const [profileResponse, functionsResponse] = await Promise.all([
        profileApiCall,
        registryApi.getFunctions({ author: username, limit: 100 }),
      ]);

      let profileData: any;
      if (isOwnProfile) {
        const meResponse = profileResponse as MeResponse;
        profileData = {
          id: meResponse.id,
          username: meResponse.username,
          name: meResponse.name,
          avatar: meResponse.avatar,
          bio: meResponse.bio,
          location: meResponse.location,
          website: meResponse.website,
          jobTitle: meResponse.jobTitle,
          companyName: meResponse.companyName,
          twitterUrl: meResponse.twitterUrl,
          githubUrl: meResponse.githubUrl,
          linkedinUrl: meResponse.linkedinUrl,
          socialLinks: meResponse.socialLinks,
          createdAt: meResponse.createdAt ?? meResponse.updatedAt,
          stats: meResponse.stats,
          publishedFunctions: [],
          profileNumber: meResponse.profileNumber,
          isOnline: meResponse.isOnline ?? true,
          lastActive: meResponse.lastActive,
          role: meResponse.role,
          founderNumber: meResponse.founderNumber,
        };
      } else {
        profileData = profileResponse;
      }

      return transformToUserProfile(profileData, functionsResponse.functions || []);
    },
    enabled: !!username && !!authChecked,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const { data: analyticsResponse } = useQuery<UserAnalyticsResponse>({
    queryKey: ['profile-analytics', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try {
        return await usersApi.getUserAnalytics(username);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) {
          return {
            executionStats: { totalExecutions: 0, totalUniqueUsers: 0, functionCount: 0, executionHistory: [] },
            popularFunctions: [],
            geographicStats: { regions: [] },
            deviceStats: { devices: [] },
          };
        }
        throw err;
      }
    },
    enabled: !!username && (activeTab === 'analytics' || activeTab === 'overview'),
    staleTime: 5 * 60 * 1000,
  });

  const analytics: ProfileAnalytics | undefined = analyticsResponse
    ? {
        executionHistory: analyticsResponse.executionStats?.executionHistory?.map((h) => ({
          date: h.date,
          executions: Number(h.executions) || 0,
          uniqueUsers: Number(h.uniqueUsers) || 0,
        })) || [],
        popularFunctions: analyticsResponse.popularFunctions?.map((f) => ({
          functionId: f.id, name: f.name, executions: Number(f.executionCount) || 0, percentage: 0,
        })) || [],
        geographicDistribution: analyticsResponse.geographicStats?.regions?.map((r) => ({
          country: r.region, executions: Number(r.executions) || 0, percentage: 0,
        })) || [],
        deviceStats: analyticsResponse.deviceStats?.devices?.map((d) => ({ device: d.device, percentage: 0 })) || [],
        browserStats: [],
      }
    : undefined;

  const { data: achievementsResponse } = useQuery<UserAchievementsResponse>({
    queryKey: ['profile-achievements', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try { return await usersApi.getUserAchievements(username); }
      catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { achievements: [], totalPoints: 0, available: 0 };
        throw err;
      }
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  const { data: certificatesResponse } = useQuery({
    queryKey: ['profile-certifications', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try {
        const { certificationApi } = await import('@/api/certification');
        return certificationApi.getPublicBadges(username);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { badges: [], count: 0, username: '' };
        throw err;
      }
    },
    enabled: !!username,
    staleTime: 1000 * 60 * 2,
  });

  const certificationsData = (certificatesResponse as { badges?: import('@/api/certification').PublicBadge[] } | undefined)?.badges || [];

  const { data: activityResponse } = useQuery<UserActivityResponse>({
    queryKey: ['profile-activity', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try { return await usersApi.getUserActivity(username, { limit: 20 }); }
      catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { activities: [], limit: 20, offset: 0, total: 0 };
        throw err;
      }
    },
    enabled: !!username && (activeTab === 'activity' || activeTab === 'overview'),
    staleTime: 5 * 60 * 1000,
  });

  const { data: contributionResponse } = useQuery<UserContributionsResponse>({
    queryKey: ['profile-contributions', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try { return await usersApi.getUserContributions(username); }
      catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) return undefined;
        throw err;
      }
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  const { data: skillsResponse } = useQuery<UserSkillsResponse>({
    queryKey: ['profile-skills', username],
    queryFn: async () => {
      if (!username) throw new Error('Username is required');
      try { return await usersApi.getUserSkills(username); }
      catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { skills: [] };
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
      icon: a.icon || 'Award',
      color: a.color || 'blue',
      unlockedAt: a.earnedAt,
      tier: a.isCompleted ? a.points >= 500 ? 'platinum' : a.points >= 200 ? 'gold' : a.points >= 100 ? 'silver' : 'bronze' : 'bronze',
      progress: { current: a.progress, target: 100 },
    })) || [];

  const typeMap: Record<string, ActivityType> = {
    function_published: 'function_published',
    function_updated: 'function_updated',
    badge_earned: 'achievement_earned',
    profile_updated: 'deployment',
    review_submitted: 'review_received',
    comment_posted: 'contribution',
    membership_upgraded: 'membership_upgraded',
  };

  const rawActivity: UserActivity[] =
    activityResponse?.activities?.map((a) => ({
      id: a.id,
      type: typeMap[a.type] || 'contribution',
      title: a.title,
      description: a.description || '',
      timestamp: a.createdAt,
      metadata: a.metadata,
    })) || [];

  const joinedActivity: UserActivity | null = profile
    ? {
        id: `joined-${profile.id}`,
        type: 'joined',
        title: 'Joined FunctionFly',
        description: profile.createdAt ? format(new Date(profile.createdAt), 'MMMM d, yyyy') : '',
        timestamp: profile.createdAt ?? new Date().toISOString(),
      }
    : null;
  const activityData: UserActivity[] = joinedActivity ? [joinedActivity, ...rawActivity] : rawActivity;

  const skillsData: Skill[] =
    skillsResponse?.skills?.map((s) => ({
      id: s.id,
      name: s.name,
      level: s.level,
      category: (s.category as Skill['category']) || 'concept',
    })) || [];

  const pageTitle = profile ? profile.username : 'Profile';
  usePageTitle(pageTitle);

  const mergedProfile: UserProfile | undefined = profile
    ? {
        ...profile,
        certifications: certificationsData,
        achievements: achievementsData.length > 0 ? achievementsData : profile.achievements,
        recentActivity: activityData.length > 0 ? activityData : profile.recentActivity,
        skills: skillsData.length > 0 ? skillsData : profile.skills,
        stats: {
          ...profile.stats,
          ...(contributionResponse?.days?.length
            ? {
                contributionGraph: contributionResponse.days.map((d) => ({
                  date: d.date,
                  count: Number(d.count) || 0,
                  level: (Math.min(4, Math.max(0, Number(d.level))) || 0) as 0 | 1 | 2 | 3 | 4,
                })),
                contributionStreak: {
                  current: contributionResponse.currentStreak ?? 0,
                  longest: contributionResponse.longestStreak ?? 0,
                  lastContribution: contributionResponse.lastContributionDate ?? profile.stats.contributionStreak.lastContribution,
                },
              }
            : {}),
        },
      }
    : undefined;

  const updateProfileMutation = useMutation({
    mutationFn: async (data: import('@/api/users').UpdateProfileRequest) => {
      return usersApi.updateMe(data);
    },
    onSuccess: (response) => {
      if (response?.user?.avatar !== undefined) {
        queryClient.setQueryData<UserProfile>(
          ['enhanced-profile', username, isOwnProfile],
          (old) => old ? { ...old, avatar: response.user.avatar } : old,
        );
      }
      queryClient.invalidateQueries({ queryKey: ['enhanced-profile', username] });
      queryClient.invalidateQueries({ queryKey: ['my-profile'] });
      queryClient.invalidateQueries({ queryKey: ['my-settings'] });
      useAuthStore.getState().initialize();
    },
  });

  const tabs: { value: ProfileTab; label: string; icon: React.ReactNode }[] = [
    { value: 'overview', label: 'Overview', icon: <User className="profile-tab-icon" /> },
    { value: 'functions', label: 'Functions', icon: <Package className="profile-tab-icon" /> },
    { value: 'activity', label: 'Activity', icon: <Activity className="profile-tab-icon" /> },
    { value: 'analytics', label: 'Analytics', icon: <BarChart3 className="profile-tab-icon" /> },
    { value: 'about', label: 'About', icon: <BookOpen className="profile-tab-icon" /> },
  ];

  if (isOwnProfile) {
    tabs.push({ value: 'settings', label: 'Settings', icon: <Settings className="profile-tab-icon" /> });
  }

  return (
    <div className="profile-page">
      <PageGrid />
      <Navbar variant="dashboard" />

      <main className="profile-main">
        <div className="profile-container">
          {isLoading && (
            <Chamber className="profile-loading-chamber">
              <ProfileHeaderSkeleton />
              <div className="profile-loading-stats">
                <StatsOverviewSkeleton />
              </div>
              <div className="profile-loading-tabs">
                <TabContentSkeleton />
              </div>
            </Chamber>
          )}

          {isError && (
            <div className="profile-error">
              <UserNotFoundView
                username={username}
                is404={(error as Error)?.message?.includes('404')}
                compact
              />
            </div>
          )}

          {profile && (
            <Chamber className="profile-chamber">
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag primary={`USER ${profile.profileNumber ? `#${profile.profileNumber}` : ''}`} secondary={profile.username} position="top-right" />

              <ProfileHeader
                profile={profile}
                isOwnProfile={isOwnProfile}
                isViewerSignedIn={!!currentUser}
                onEditProfile={() => setIsEditModalOpen(true)}
                onAvatarClick={isOwnProfile ? () => setIsAvatarPickerOpen(true) : undefined}
                isEnterprise={
                  isOwnProfile
                    ? isCurrentUserEnterprise
                    : profile.username.toLowerCase().includes('enterprise')
                }
              />

              <StatsOverview stats={mergedProfile!.stats} />

              {/* Tabs */}
              <div className="profile-tabs-wrapper">
                <div className="profile-tabs">
                  {tabs.map((tab) => (
                    <button
                      key={tab.value}
                      className={`profile-tab ${activeTab === tab.value ? 'profile-tab--active' : ''}`}
                      onClick={() => handleTabChange(tab.value)}
                    >
                      {tab.icon}
                      {tab.label}
                    </button>
                  ))}
                </div>

                <div className="profile-tab-content">
                  {activeTab === 'overview' && <OverviewTab profile={mergedProfile!} />}
                  {activeTab === 'functions' && <FunctionsTab profile={mergedProfile!} />}
                  {activeTab === 'activity' && <ActivityTab profile={mergedProfile!} />}
                  {activeTab === 'analytics' && (
                    analytics ? <AnalyticsTab analytics={analytics} /> : (
                      <div className="profile-tab-empty">
                        <BarChart3 className="profile-tab-empty__icon" />
                        <p>Loading analytics...</p>
                      </div>
                    )
                  )}
                  {activeTab === 'about' && (
                    <AboutTab
                      profile={mergedProfile!}
                      isOwnProfile={isOwnProfile}
                      userSkills={skillsResponse?.skills}
                      onAddSkill={async (skill) => {
                        await usersApi.addSkill(skill as import('@/api/users').AddSkillRequest);
                        queryClient.invalidateQueries({ queryKey: ['profile-skills', username] });
                      }}
                      onRemoveSkill={async (skillId) => {
                        await usersApi.removeSkill(skillId);
                        queryClient.invalidateQueries({ queryKey: ['profile-skills', username] });
                      }}
                      isSkillsLoading={skillsResponse === undefined}
                    />
                  )}
                  {activeTab === 'settings' && (
                    <div className="profile-settings">
                      <SettingsContent showHeader={false} profile={mergedProfile ?? undefined} />
                    </div>
                  )}
                </div>
              </div>
            </Chamber>
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
              toast.success('Profile picture updated');
            } catch {
              toast.error('Failed to update profile picture');
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

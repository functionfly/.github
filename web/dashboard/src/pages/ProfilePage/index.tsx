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

import { useState, useEffect, useMemo } from "react";
import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import {
  User,
  Calendar,
  Package,
  ExternalLink,
  AlertCircle,
  MapPin,
  Building2,
  Link as LinkIcon,
  Github,
  Twitter,
  Linkedin,
  Globe,
  MessageCircle,
  Share2,
  MoreHorizontal,
  Users,
  TrendingUp,
  Activity,
  Award,
  Code2,
  DollarSign,
  Flame,
  Filter,
  Search,
  ChevronDown,
  Heart,
  GitBranch,
  Star,
  Clock,
  BarChart3,
  PieChart,
  Map,
  Monitor,
  Mail,
  Briefcase,
  GraduationCap,
  BookOpen,
  Settings,
  Edit3,
  Eye,
  Zap,
  CheckCircle2,
  Shield,
  Target,
} from "lucide-react";
import { format, formatDistanceToNow, subDays } from "date-fns";
import { cn, formatNumber } from "@/lib/utils";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Navbar } from "@/components/common/Navbar";
import { FunctionCard } from "@/components/functions/FunctionCard";
import { TrustScoreBadge } from "@/components/common/TrustScoreBadge";
import { LineChart } from "@/components/common/LineChart";
import { BarChart } from "@/components/common/BarChart";
import { Sparkline } from "@/components/common/Sparkline";
import type {
  UserProfile,
  ProfileTab,
  UserActivity,
  Achievement as AchievementType,
  Skill,
  ProfileAnalytics,
  FunctionFilters,
  ActivityType,
  PublicUserProfile,
  FunctionCardData,
} from "@/types";
import {
  usersApi,
  type UserAnalyticsResponse,
  type UserAchievementsResponse,
  type UserActivityResponse,
  type UserSkillsResponse,
  type Achievement,
  type UserActivityItem,
  type UserSkill,
} from "@/api/users";
import { registryApi, RegistryFunction } from "@/api/registry";

// ============================================================================
// API Data Transformers
// ============================================================================

/**
 * Transform registry function to FunctionCardData format
 */
function transformRegistryFunction(fn: RegistryFunction, author: { id: string; username: string; name: string; avatar?: string }): FunctionCardData {
  return {
    id: fn.id,
    name: fn.name,
    description: fn.description || "",
    author,
    trustScore: Math.round(fn.overall_score * 100),
    metrics: {
      executionCount: 0, // Will be populated from stats if needed
      executionTrend: [],
      averageLatency: 0,
      errorRate: 0,
    },
    pricing: {
      model: fn.price_per_call > 0 ? "per_call" : "free",
      pricePerCall: fn.price_per_call,
      currency: "USD",
    },
    isVerified: fn.overall_score >= 0.8,
    isDeterministic: fn.deterministic_score >= 0.9,
    rating: {
      average: fn.total_ratings > 0 ? fn.overall_score * 5 : 0,
      count: fn.total_ratings,
      distribution: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 },
    },
    tags: fn.tags || [],
    category: fn.category || "other",
    language: fn.latest_version ? "typescript" : "unknown", // Default to typescript
    lastUpdated: fn.created_at,
    version: fn.latest_version || "1.0.0",
    isFavorite: false,
    isFeatured: false,
  };
}

/**
 * Generate empty contribution graph (365 days)
 */
function generateEmptyContributionGraph(): UserProfile["stats"]["contributionGraph"] {
  const data: UserProfile["stats"]["contributionGraph"] = [];
  for (let i = 364; i >= 0; i--) {
    const date = subDays(new Date(), i);
    data.push({
      date: format(date, "yyyy-MM-dd"),
      count: 0,
      level: 0,
    });
  }
  return data;
}

/**
 * Transform API response to UserProfile format
 */
function transformToUserProfile(
  apiProfile: PublicUserProfile,
  registryFunctions: RegistryFunction[]
): UserProfile {
  const authorInfo = {
    id: apiProfile.id,
    username: apiProfile.username,
    name: apiProfile.name,
    avatar: apiProfile.avatar,
  };

  const publishedFunctions = registryFunctions.map(fn =>
    transformRegistryFunction(fn, authorInfo)
  );

  // Calculate stats from functions
  const totalExecutions = publishedFunctions.reduce((sum, f) => sum + (f.metrics?.executionCount || 0), 0);
  const totalViews = totalExecutions * 2; // Estimate

  return {
    id: apiProfile.id,
    username: apiProfile.username,
    name: apiProfile.name,
    avatar: apiProfile.avatar,
    coverImage: undefined,
    bio: apiProfile.bio,
    location: apiProfile.location,
    company: apiProfile.companyName,
    jobTitle: apiProfile.jobTitle,
    website: apiProfile.website,
    socialLinks: {
      github: apiProfile.githubUrl,
      twitter: apiProfile.twitterUrl,
      linkedin: apiProfile.linkedinUrl,
      website: apiProfile.website,
      discord: apiProfile.socialLinks?.discord,
    },
    skills: [], // Will be populated from separate API call
    createdAt: apiProfile.createdAt,
    updatedAt: undefined,
    isOnline: false,
    lastActive: undefined,
    experience: [],
    education: [],
    openSourceContributions: [],
    languages: [],
    stats: {
      functionsPublished: publishedFunctions.length,
      functionsTrend: 0,
      totalExecutions,
      executionsTrend: 0,
      totalViews,
      viewsTrend: 0,
      trustScore: 75, // Default trust score for users with functions
      reputationRank: "Contributor",
      followersCount: 0,
      followingCount: 0,
      followersTrend: 0,
      contributionStreak: {
        current: 0,
        longest: 0,
        lastContribution: new Date().toISOString(),
      },
      contributionGraph: generateEmptyContributionGraph(),
    },
    achievements: [], // Will be populated from separate API call
    recentActivity: [], // Will be populated from separate API call
    publishedFunctions,
  };
}

// ============================================================================
// Animation Variants
// ============================================================================

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.1 },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] as const },
  },
};

const tabContentVariants = {
  hidden: { opacity: 0, x: -20 },
  visible: { opacity: 1, x: 0, transition: { duration: 0.3 } },
  exit: { opacity: 0, x: 20, transition: { duration: 0.2 } },
};

// ============================================================================
// Skeleton Components
// ============================================================================

function ProfileHeaderSkeleton() {
  return (
    <div className="animate-pulse">
      {/* Cover */}
      <div className="h-48 md:h-64 bg-gradient-to-r from-gray-700 to-gray-800 rounded-t-xl" />

      <div className="px-4 md:px-8 pb-6">
        <div className="flex flex-col md:flex-row md:items-end -mt-16 md:-mt-20 gap-4 md:gap-6">
          {/* Avatar */}
          <Skeleton className="w-32 h-32 md:w-40 md:h-40 rounded-full border-4 border-background shrink-0" />

          {/* Info */}
          <div className="flex-1 min-w-0 space-y-3">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-full max-w-md" />
            <div className="flex gap-4">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2">
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-24" />
          </div>
        </div>
      </div>
    </div>
  );
}

function StatsOverviewSkeleton() {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="bg-card">
          <CardContent className="p-4">
            <Skeleton className="h-4 w-20 mb-2" />
            <Skeleton className="h-8 w-16 mb-1" />
            <Skeleton className="h-3 w-12" />
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function TabContentSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-40 rounded-lg" />
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// Profile Header Section
// ============================================================================

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  onEditProfile?: () => void;
}

function ProfileHeader({ profile, isOwnProfile, onEditProfile }: ProfileHeaderProps) {
  const [isFollowing, setIsFollowing] = useState(false);
  const [followersCount, setFollowersCount] = useState(profile.stats.followersCount);

  const handleFollow = () => {
    setIsFollowing(!isFollowing);
    setFollowersCount(prev => isFollowing ? prev - 1 : prev + 1);
  };

  const handleShare = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      // Could show toast here
    } catch (err) {
      console.error("Failed to copy URL", err);
    }
  };

  const joinedDate = format(new Date(profile.createdAt), "MMMM yyyy");
  const lastActiveText = profile.lastActive
    ? formatDistanceToNow(new Date(profile.lastActive), { addSuffix: true })
    : null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="relative"
    >
      {/* Cover Image */}
      <div className="h-48 md:h-64 rounded-t-xl overflow-hidden relative">
        {profile.coverImage ? (
          <img
            src={profile.coverImage}
            alt="Profile cover"
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-brand-500 via-brand-600 to-indigo-700">
            <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4wNSI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
          </div>
        )}

        {/* Gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent" />
      </div>

      {/* Profile Info */}
      <div className="px-4 md:px-8 pb-6">
        <div className="flex flex-col md:flex-row md:items-end -mt-16 md:-mt-20 gap-4 md:gap-6">
          {/* Avatar */}
          <div className="relative shrink-0">
            <div className="w-32 h-32 md:w-40 md:h-40 rounded-full border-4 border-background bg-gradient-to-br from-brand-500 to-brand-600 flex items-center justify-center overflow-hidden shadow-xl">
              {profile.avatar ? (
                <img
                  src={profile.avatar}
                  alt={profile.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <span className="text-4xl md:text-5xl font-bold text-white">
                  {profile.name.charAt(0).toUpperCase()}
                </span>
              )}
            </div>

            {/* Online status indicator */}
            <div
              className={cn(
                "absolute bottom-2 right-2 w-6 h-6 rounded-full border-4 border-background",
                profile.isOnline ? "bg-green-500" : "bg-gray-400"
              )}
              title={profile.isOnline ? "Online" : `Last active ${lastActiveText}`}
            />
          </div>

          {/* User Info */}
          <div className="flex-1 min-w-0 space-y-2">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl md:text-3xl font-bold text-text-primary">
                {profile.name}
              </h1>
              {profile.stats.trustScore >= 80 && (
                <Badge variant="default" className="bg-brand-500/20 text-brand-400 border-brand-500/30">
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Verified
                </Badge>
              )}
            </div>

            <p className="text-lg text-brand-400 font-medium">@{profile.username}</p>

            {profile.bio && (
              <p className="text-text-secondary max-w-2xl">{profile.bio}</p>
            )}

            {/* Meta info */}
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-text-muted">
              {profile.location && (
                <span className="flex items-center gap-1">
                  <MapPin className="w-4 h-4" />
                  {profile.location}
                </span>
              )}
              {profile.company && (
                <span className="flex items-center gap-1">
                  <Building2 className="w-4 h-4" />
                  {profile.company}
                  {profile.jobTitle && ` · ${profile.jobTitle}`}
                </span>
              )}
              <span className="flex items-center gap-1">
                <Calendar className="w-4 h-4" />
                Joined {joinedDate}
              </span>
              {lastActiveText && !profile.isOnline && (
                <span className="flex items-center gap-1">
                  <Clock className="w-4 h-4" />
                  Active {lastActiveText}
                </span>
              )}
            </div>

            {/* Social Links */}
            {Object.values(profile.socialLinks).some(Boolean) && (
              <div className="flex items-center gap-3 pt-1">
                {profile.socialLinks.github && (
                  <a
                    href={profile.socialLinks.github}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-muted hover:text-text-primary transition-colors"
                    title="GitHub"
                  >
                    <Github className="w-5 h-5" />
                  </a>
                )}
                {profile.socialLinks.twitter && (
                  <a
                    href={profile.socialLinks.twitter}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-muted hover:text-text-primary transition-colors"
                    title="Twitter"
                  >
                    <Twitter className="w-5 h-5" />
                  </a>
                )}
                {profile.socialLinks.linkedin && (
                  <a
                    href={profile.socialLinks.linkedin}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-muted hover:text-text-primary transition-colors"
                    title="LinkedIn"
                  >
                    <Linkedin className="w-5 h-5" />
                  </a>
                )}
                {profile.socialLinks.website && (
                  <a
                    href={profile.socialLinks.website}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-muted hover:text-text-primary transition-colors"
                    title="Website"
                  >
                    <Globe className="w-5 h-5" />
                  </a>
                )}
                {profile.socialLinks.discord && (
                  <a
                    href={profile.socialLinks.discord}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-muted hover:text-text-primary transition-colors"
                    title="Discord"
                  >
                    <MessageCircle className="w-5 h-5" />
                  </a>
                )}
                {profile.website && (
                  <a
                    href={profile.website}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1 text-sm text-text-muted hover:text-brand-400 transition-colors"
                  >
                    <LinkIcon className="w-4 h-4" />
                    {new URL(profile.website).hostname}
                  </a>
                )}
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 shrink-0">
            {!isOwnProfile ? (
              <>
                <Button
                  variant={isFollowing ? "outline" : "default"}
                  onClick={handleFollow}
                  className={cn(
                    "gap-2",
                    isFollowing && "border-brand-500 text-brand-400"
                  )}
                >
                  <Users className="w-4 h-4" />
                  {isFollowing ? "Following" : "Follow"}
                </Button>
                <Button variant="outline" className="gap-2">
                  <MessageCircle className="w-4 h-4" />
                  Message
                </Button>
              </>
            ) : (
              <Button variant="outline" onClick={onEditProfile} className="gap-2">
                <Edit3 className="w-4 h-4" />
                Edit Profile
              </Button>
            )}

            <Button variant="outline" size="icon" onClick={handleShare}>
              <Share2 className="w-4 h-4" />
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="icon">
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <ExternalLink className="w-4 h-4 mr-2" />
                  Copy Profile URL
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Eye className="w-4 h-4 mr-2" />
                  View in Registry
                </DropdownMenuItem>
                {!isOwnProfile && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-500">
                      Report User
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

// ============================================================================
// Stats Overview Section
// ============================================================================

interface StatsOverviewProps {
  stats: UserProfile["stats"];
}

function StatsOverview({ stats }: StatsOverviewProps) {
  const statItems = [
    {
      label: "Functions",
      value: stats.functionsPublished,
      trend: stats.functionsTrend,
      icon: Package,
      color: "text-blue-500",
      bgColor: "bg-blue-500/10",
    },
    {
      label: "Executions",
      value: stats.totalExecutions,
      trend: stats.executionsTrend,
      icon: Activity,
      color: "text-green-500",
      bgColor: "bg-green-500/10",
    },
    {
      label: "Trust Score",
      value: stats.trustScore,
      suffix: "%",
      trend: null,
      icon: Shield,
      color: stats.trustScore >= 80 ? "text-emerald-500" : stats.trustScore >= 60 ? "text-yellow-500" : "text-orange-500",
      bgColor: stats.trustScore >= 80 ? "bg-emerald-500/10" : stats.trustScore >= 60 ? "bg-yellow-500/10" : "bg-orange-500/10",
    },
    {
      label: "Followers",
      value: stats.followersCount,
      trend: stats.followersTrend,
      icon: Users,
      color: "text-purple-500",
      bgColor: "bg-purple-500/10",
    },
    {
      label: "Following",
      value: stats.followingCount,
      trend: null,
      icon: Heart,
      color: "text-pink-500",
      bgColor: "bg-pink-500/10",
    },
    {
      label: "Streak",
      value: stats.contributionStreak.current,
      suffix: " days",
      trend: null,
      icon: Flame,
      color: "text-orange-500",
      bgColor: "bg-orange-500/10",
      highlight: stats.contributionStreak.current >= 30,
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 px-4 md:px-8 py-4"
    >
      {statItems.map((item, index) => (
        <motion.div key={item.label} variants={itemVariants}>
          <Card
            className={cn(
              "group hover:shadow-lg transition-all duration-300 border-border-subtle hover:border-brand-500/30",
              item.highlight && "ring-2 ring-orange-500/30"
            )}
          >
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className={cn("p-1.5 rounded-md", item.bgColor, item.color)}>
                  <item.icon className="w-4 h-4" />
                </div>
                <span className="text-xs text-text-muted">{item.label}</span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold text-text-primary">
                  {formatNumber(item.value)}{item.suffix || ""}
                </span>
                {item.trend !== null && item.trend !== undefined && (
                  <span
                    className={cn(
                      "text-xs font-medium",
                      item.trend >= 0 ? "text-green-500" : "text-red-500"
                    )}
                  >
                    {item.trend >= 0 ? "+" : ""}{item.trend}%
                  </span>
                )}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </motion.div>
  );
}

// ============================================================================
// Contribution Graph (GitHub-style Heatmap)
// ============================================================================

interface ContributionGraphProps {
  data: UserProfile["stats"]["contributionGraph"];
}

function ContributionGraph({ data }: ContributionGraphProps) {
  const levelColors = [
    "bg-border-subtle",
    "bg-brand-500/20",
    "bg-brand-500/40",
    "bg-brand-500/60",
    "bg-brand-500",
  ];

  // Group by weeks
  const weeks = useMemo(() => {
    const result: typeof data[] = [];
    for (let i = 0; i < data.length; i += 7) {
      result.push(data.slice(i, i + 7));
    }
    return result;
  }, [data]);

  const totalContributions = data.reduce((sum, day) => sum + day.count, 0);

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <GitBranch className="w-5 h-5 text-brand-500" />
            Contribution Activity
          </CardTitle>
          <span className="text-sm text-text-muted">
            {formatNumber(totalContributions)} contributions in the last year
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <div className="flex gap-1 min-w-max">
            {weeks.map((week, weekIndex) => (
              <div key={weekIndex} className="flex flex-col gap-1">
                {week.map((day, dayIndex) => (
                  <div
                    key={dayIndex}
                    className={cn(
                      "w-3 h-3 rounded-sm transition-all duration-200 hover:ring-2 hover:ring-brand-500/50 cursor-pointer",
                      levelColors[day.level]
                    )}
                    title={`${day.count} contributions on ${format(new Date(day.date), "MMM d, yyyy")}`}
                  />
                ))}
              </div>
            ))}
          </div>
        </div>

        {/* Legend */}
        <div className="flex items-center gap-2 mt-4 text-xs text-text-muted">
          <span>Less</span>
          {levelColors.map((color, i) => (
            <div key={i} className={cn("w-3 h-3 rounded-sm", color)} />
          ))}
          <span>More</span>
        </div>
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Achievements Section
// ============================================================================

interface AchievementsSectionProps {
  achievements: Achievement[];
}

function AchievementsSection({ achievements }: AchievementsSectionProps) {
  const tierIcons = {
    bronze: "🥉",
    silver: "🥈",
    gold: "🥇",
    platinum: "💎",
  };

  const tierColors = {
    bronze: "from-amber-700 to-amber-600",
    silver: "from-gray-400 to-gray-300",
    gold: "from-yellow-500 to-yellow-400",
    platinum: "from-cyan-500 to-blue-500",
  };

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          <Award className="w-5 h-5 text-brand-500" />
          Achievements
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {achievements.slice(0, 4).map((achievement) => (
            <div
              key={achievement.id}
              className="flex items-start gap-3 p-3 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors group cursor-pointer"
            >
              <div
                className={cn(
                  "w-12 h-12 rounded-lg bg-gradient-to-br flex items-center justify-center text-2xl shrink-0",
                  tierColors[achievement.tier]
                )}
              >
                {tierIcons[achievement.tier]}
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-medium text-text-primary text-sm">{achievement.name}</h4>
                <p className="text-xs text-text-muted line-clamp-2">{achievement.description}</p>
                {achievement.progress && (
                  <div className="mt-2">
                    <Progress
                      value={(achievement.progress.current / achievement.progress.target) * 100}
                      className="h-1"
                    />
                    <span className="text-xs text-text-muted mt-1">
                      {achievement.progress.current} / {achievement.progress.target}
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        {achievements.length > 4 && (
          <Button variant="ghost" className="w-full mt-3 text-sm text-brand-400">
            View all {achievements.length} achievements
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Skills Section
// ============================================================================

interface SkillsSectionProps {
  skills: Skill[];
}

function SkillsSection({ skills }: SkillsSectionProps) {
  const levelColors = {
    beginner: "bg-gray-500/20 text-gray-400",
    intermediate: "bg-blue-500/20 text-blue-400",
    advanced: "bg-purple-500/20 text-purple-400",
    expert: "bg-brand-500/20 text-brand-400",
  };

  const categories = {
    language: "Languages",
    framework: "Frameworks",
    tool: "Tools",
    platform: "Platforms",
    concept: "Concepts",
  };

  const groupedSkills = skills.reduce((acc, skill) => {
    if (!acc[skill.category]) acc[skill.category] = [];
    acc[skill.category].push(skill);
    return acc;
  }, {} as Record<string, Skill[]>);

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          <Code2 className="w-5 h-5 text-brand-500" />
          Skills & Technologies
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {Object.entries(groupedSkills).map(([category, categorySkills]) => (
          <div key={category}>
            <h4 className="text-sm font-medium text-text-muted mb-2">
              {categories[category as keyof typeof categories]}
            </h4>
            <div className="flex flex-wrap gap-2">
              {categorySkills.map((skill) => (
                <Badge
                  key={skill.name}
                  variant="secondary"
                  className={cn(
                    "px-2 py-1 text-xs cursor-default transition-all hover:scale-105",
                    levelColors[skill.level]
                  )}
                  title={`${skill.level}${skill.endorsements ? ` · ${skill.endorsements} endorsements` : ""}`}
                >
                  {skill.name}
                </Badge>
              ))}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Trust Metrics Section
// ============================================================================

interface TrustMetricsSectionProps {
  trustScore: number;
}

function TrustMetricsSection({ trustScore }: TrustMetricsSectionProps) {
  const metrics = [
    { name: "Reliability", score: Math.min(100, trustScore + Math.random() * 10 - 5), icon: Shield },
    { name: "Performance", score: Math.min(100, trustScore + Math.random() * 10 - 5), icon: Zap },
    { name: "Community", score: Math.min(100, trustScore + Math.random() * 15 - 7), icon: Users },
    { name: "Documentation", score: Math.min(100, trustScore + Math.random() * 20 - 10), icon: BookOpen },
  ];

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          <Target className="w-5 h-5 text-brand-500" />
          Trust Metrics
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4 mb-6">
          <div className="relative w-24 h-24">
            <svg className="w-full h-full -rotate-90" viewBox="0 0 40 40">
              <circle
                cx="20"
                cy="20"
                r="16"
                fill="none"
                stroke="currentColor"
                strokeWidth="3"
                className="text-border-subtle"
              />
              <circle
                cx="20"
                cy="20"
                r="16"
                fill="none"
                stroke="currentColor"
                strokeWidth="3"
                strokeLinecap="round"
                className={cn(
                  trustScore >= 80 ? "text-emerald-500" :
                  trustScore >= 60 ? "text-yellow-500" :
                  "text-orange-500"
                )}
                style={{
                  strokeDasharray: 2 * Math.PI * 16,
                  strokeDashoffset: 2 * Math.PI * 16 * (1 - trustScore / 100),
                }}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-2xl font-bold text-text-primary">{trustScore}</span>
              <span className="text-xs text-text-muted">Score</span>
            </div>
          </div>
          <div>
            <h4 className="font-medium text-text-primary">
              {trustScore >= 80 ? "Excellent" : trustScore >= 60 ? "Good" : "Fair"} Reputation
            </h4>
            <p className="text-sm text-text-muted">
              Based on function quality, community engagement, and execution reliability
            </p>
          </div>
        </div>

        <div className="space-y-3">
          {metrics.map((metric) => (
            <div key={metric.name} className="flex items-center gap-3">
              <metric.icon className="w-4 h-4 text-text-muted" />
              <span className="text-sm text-text-secondary w-28">{metric.name}</span>
              <div className="flex-1">
                <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                  <div
                    className={cn(
                      "h-full rounded-full transition-all duration-1000",
                      metric.score >= 80 ? "bg-emerald-500" :
                      metric.score >= 60 ? "bg-yellow-500" :
                      "bg-orange-500"
                    )}
                    style={{ width: `${metric.score}%` }}
                  />
                </div>
              </div>
              <span className="text-sm font-medium text-text-primary w-10 text-right">
                {Math.round(metric.score)}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Activity Timeline
// ============================================================================

interface ActivityTimelineProps {
  activities: UserActivity[];
  filter?: ActivityType | "all";
}

function ActivityTimeline({ activities, filter = "all" }: ActivityTimelineProps) {
  const filteredActivities = filter === "all"
    ? activities
    : activities.filter(a => a.type === filter);

  const typeIcons: Record<ActivityType, React.ReactNode> = {
    function_published: <Package className="w-4 h-4" />,
    function_updated: <Edit3 className="w-4 h-4" />,
    function_deleted: <AlertCircle className="w-4 h-4" />,
    achievement_earned: <Award className="w-4 h-4" />,
    review_received: <Star className="w-4 h-4" />,
    milestone_reached: <Target className="w-4 h-4" />,
    followed: <Users className="w-4 h-4" />,
    follower_gained: <Heart className="w-4 h-4" />,
    contribution: <GitBranch className="w-4 h-4" />,
    deployment: <Zap className="w-4 h-4" />,
  };

  const typeColors: Record<ActivityType, string> = {
    function_published: "bg-blue-500/20 text-blue-400",
    function_updated: "bg-yellow-500/20 text-yellow-400",
    function_deleted: "bg-red-500/20 text-red-400",
    achievement_earned: "bg-amber-500/20 text-amber-400",
    review_received: "bg-purple-500/20 text-purple-400",
    milestone_reached: "bg-emerald-500/20 text-emerald-400",
    followed: "bg-pink-500/20 text-pink-400",
    follower_gained: "bg-pink-500/20 text-pink-400",
    contribution: "bg-brand-500/20 text-brand-400",
    deployment: "bg-green-500/20 text-green-400",
  };

  return (
    <div className="space-y-4">
      {filteredActivities.map((activity, index) => (
        <motion.div
          key={activity.id}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: index * 0.05 }}
          className="flex gap-4"
        >
          <div className="flex flex-col items-center">
            <div className={cn("w-10 h-10 rounded-full flex items-center justify-center", typeColors[activity.type])}>
              {typeIcons[activity.type]}
            </div>
            {index < filteredActivities.length - 1 && (
              <div className="w-px flex-1 bg-border-subtle my-2" />
            )}
          </div>
          <div className="flex-1 pb-6">
            <div className="flex items-start justify-between gap-2">
              <div>
                <h4 className="font-medium text-text-primary">{activity.title}</h4>
                {activity.description && (
                  <p className="text-sm text-text-muted mt-0.5">{activity.description}</p>
                )}
                {activity.relatedFunction && (
                  <Link
                    to={`/fx/${activity.relatedFunction.author}/${activity.relatedFunction.name}`}
                    className="inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300 mt-1"
                  >
                    <Code2 className="w-3.5 h-3.5" />
                    {activity.relatedFunction.name}
                  </Link>
                )}
              </div>
              <span className="text-xs text-text-muted shrink-0">
                {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
              </span>
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  );
}

// ============================================================================
// Tab Content Components
// ============================================================================

function OverviewTab({ profile }: { profile: UserProfile }) {
  const featuredFunctions = profile.publishedFunctions
    .filter(f => f.isFeatured || f.metrics.executionCount > 1000)
    .slice(0, 4);

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Featured Functions */}
      {featuredFunctions.length > 0 && (
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-text-primary flex items-center gap-2">
              <Star className="w-5 h-5 text-brand-500" />
              Featured Functions
            </h2>
            <Link to="?tab=functions" className="text-sm text-brand-400 hover:text-brand-300">
              View all
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {featuredFunctions.map((fn) => (
              <FunctionCard
                key={fn.id}
                data={fn}
                variant="compact"
                onView={(id) => console.log("View", id)}
              />
            ))}
          </div>
        </section>
      )}

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Activity & Contribution */}
        <div className="lg:col-span-2 space-y-6">
          <ContributionGraph data={profile.stats.contributionGraph} />

          <Card className="border-border-subtle">
            <CardHeader className="pb-3">
              <CardTitle className="text-lg flex items-center gap-2">
                <Activity className="w-5 h-5 text-brand-500" />
                Recent Activity
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ActivityTimeline activities={profile.recentActivity.slice(0, 6)} />
            </CardContent>
          </Card>
        </div>

        {/* Right Column - Achievements, Skills, Trust */}
        <div className="space-y-6">
          <AchievementsSection achievements={profile.achievements} />
          <TrustMetricsSection trustScore={profile.stats.trustScore} />
          <SkillsSection skills={profile.skills} />
        </div>
      </div>
    </motion.div>
  );
}

function FunctionsTab({ profile }: { profile: UserProfile }) {
  const [filters, setFilters] = useState<FunctionFilters>({
    search: "",
    sortBy: "popular",
  });

  const filteredFunctions = useMemo(() => {
    let result = [...profile.publishedFunctions];

    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      result = result.filter(
        f =>
          f.name.toLowerCase().includes(searchLower) ||
          f.description.toLowerCase().includes(searchLower) ||
          f.tags?.some(t => t.toLowerCase().includes(searchLower))
      );
    }

    result.sort((a, b) => {
      switch (filters.sortBy) {
        case "popular":
          return (b.metrics.executionCount || 0) - (a.metrics.executionCount || 0);
        case "recent":
          return new Date(b.lastUpdated || 0).getTime() - new Date(a.lastUpdated || 0).getTime();
        case "name":
          return a.name.localeCompare(b.name);
        case "rating":
          return (b.rating?.average || 0) - (a.rating?.average || 0);
        default:
          return 0;
      }
    });

    return result;
  }, [profile.publishedFunctions, filters]);

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Filter Bar */}
      <Card className="border-border-subtle">
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Search functions..."
                value={filters.search}
                onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                className="pl-9"
              />
            </div>
            <Select
              value={filters.sortBy}
              onValueChange={(value) => setFilters(prev => ({ ...prev, sortBy: value as FunctionFilters["sortBy"] }))}
            >
              <SelectTrigger className="w-full sm:w-40">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="popular">Most Popular</SelectItem>
                <SelectItem value="recent">Recently Updated</SelectItem>
                <SelectItem value="name">Name</SelectItem>
                <SelectItem value="rating">Highest Rated</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Statistics Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold text-text-primary">{profile.publishedFunctions.length}</p>
            <p className="text-sm text-text-muted">Total Functions</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold text-text-primary">
              {formatNumber(profile.stats.totalExecutions)}
            </p>
            <p className="text-sm text-text-muted">Total Executions</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold text-text-primary">
              {formatNumber(profile.stats.totalViews)}
            </p>
            <p className="text-sm text-text-muted">Total Views</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold text-text-primary">
              {(profile.publishedFunctions.reduce((sum, f) => sum + (f.rating?.average || 0), 0) / profile.publishedFunctions.length || 0).toFixed(1)}
            </p>
            <p className="text-sm text-text-muted">Avg Rating</p>
          </CardContent>
        </Card>
      </div>

      {/* Functions Grid */}
      {filteredFunctions.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredFunctions.map((fn) => (
            <FunctionCard
              key={fn.id}
              data={fn}
              variant="compact"
              onView={(id) => console.log("View", id)}
              onExecute={(id) => console.log("Execute", id)}
              onShare={(id) => console.log("Share", id)}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-16">
          <Package className="w-16 h-16 mx-auto text-text-muted mb-4" />
          <h3 className="text-lg font-medium text-text-primary mb-2">No functions found</h3>
          <p className="text-text-muted">Try adjusting your search or filters</p>
        </div>
      )}
    </motion.div>
  );
}

function ActivityTab({ profile }: { profile: UserProfile }) {
  const [activityFilter, setActivityFilter] = useState<ActivityType | "all">("all");

  const filterOptions: { value: ActivityType | "all"; label: string }[] = [
    { value: "all", label: "All Activity" },
    { value: "function_published", label: "Functions" },
    { value: "achievement_earned", label: "Achievements" },
    { value: "review_received", label: "Reviews" },
    { value: "milestone_reached", label: "Milestones" },
  ];

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <Card className="border-border-subtle">
        <CardHeader className="pb-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <CardTitle className="text-lg flex items-center gap-2">
              <Activity className="w-5 h-5 text-brand-500" />
              Activity Timeline
            </CardTitle>
            <Select value={activityFilter} onValueChange={(v) => setActivityFilter(v as ActivityType | "all")}>
              <SelectTrigger className="w-full sm:w-40">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {filterOptions.map(opt => (
                  <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          <ActivityTimeline activities={profile.recentActivity} filter={activityFilter} />
        </CardContent>
      </Card>
    </motion.div>
  );
}

function AnalyticsTab({ analytics }: { analytics: ProfileAnalytics }) {
  const executionChartData = analytics.executionHistory.map(h => ({
    date: format(new Date(h.date), "MMM d"),
    executions: h.executions,
    users: h.uniqueUsers,
  }));

  const popularFunctionsData = analytics.popularFunctions.map(f => ({
    name: f.name.length > 20 ? f.name.slice(0, 20) + "..." : f.name,
    executions: f.executions,
  }));

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Execution Chart */}
      <LineChart
        data={executionChartData}
        series={[
          { key: "executions", name: "Executions", color: "#6366f1" },
          { key: "users", name: "Unique Users", color: "#10b981" },
        ]}
        title="Execution History"
        xAxisKey="date"
        height={300}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Popular Functions */}
        <BarChart
          data={popularFunctionsData}
          series={[{ key: "executions", name: "Executions", color: "#6366f1" }]}
          title="Popular Functions"
          xAxisKey="name"
          height={300}
          layout="horizontal"
        />

        {/* Geographic Distribution */}
        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Map className="w-5 h-5 text-brand-500" />
              Geographic Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.geographicDistribution.slice(0, 6).map((country) => (
                <div key={country.country} className="flex items-center gap-3">
                  <span className="text-sm text-text-secondary w-24">{country.country}</span>
                  <div className="flex-1">
                    <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                      <div
                        className="h-full bg-brand-500 rounded-full"
                        style={{ width: `${country.percentage}%` }}
                      />
                    </div>
                  </div>
                  <span className="text-sm font-medium text-text-primary w-16 text-right">
                    {country.percentage}%
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Device & Browser Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Monitor className="w-5 h-5 text-brand-500" />
              Device Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.deviceStats.map((device) => (
                <div key={device.device} className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">{device.device}</span>
                  <span className="text-sm font-medium text-text-primary">{device.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-brand-500" />
              Browser Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.browserStats.map((browser) => (
                <div key={browser.browser} className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">{browser.browser}</span>
                  <span className="text-sm font-medium text-text-primary">{browser.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  );
}

function AboutTab({ profile }: { profile: UserProfile }) {
  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Bio */}
          <Card className="border-border-subtle">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <User className="w-5 h-5 text-brand-500" />
                About
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-text-secondary whitespace-pre-wrap">
                {profile.bio || "No bio provided yet."}
              </p>
            </CardContent>
          </Card>

          {/* Experience */}
          {profile.experience && profile.experience.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Briefcase className="w-5 h-5 text-brand-500" />
                  Experience
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.experience.map((exp, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/10 flex items-center justify-center shrink-0">
                      <Building2 className="w-5 h-5 text-brand-500" />
                    </div>
                    <div>
                      <h4 className="font-medium text-text-primary">{exp.title}</h4>
                      <p className="text-sm text-text-secondary">{exp.company}</p>
                      <p className="text-xs text-text-muted mt-1">
                        {format(new Date(exp.startDate), "MMM yyyy")} -{" "}
                        {exp.current ? "Present" : exp.endDate ? format(new Date(exp.endDate), "MMM yyyy") : ""}
                      </p>
                      {exp.description && (
                        <p className="text-sm text-text-muted mt-2">{exp.description}</p>
                      )}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Education */}
          {profile.education && profile.education.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <GraduationCap className="w-5 h-5 text-brand-500" />
                  Education
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.education.map((edu, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-lg bg-purple-500/10 flex items-center justify-center shrink-0">
                      <GraduationCap className="w-5 h-5 text-purple-500" />
                    </div>
                    <div>
                      <h4 className="font-medium text-text-primary">{edu.institution}</h4>
                      <p className="text-sm text-text-secondary">{edu.degree} in {edu.field}</p>
                      <p className="text-xs text-text-muted mt-1">
                        {format(new Date(edu.startDate), "yyyy")} -{" "}
                        {edu.endDate ? format(new Date(edu.endDate), "yyyy") : "Present"}
                      </p>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Open Source */}
          {profile.openSourceContributions && profile.openSourceContributions.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Github className="w-5 h-5 text-brand-500" />
                  Open Source Contributions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {profile.openSourceContributions.map((contrib, index) => (
                    <a
                      key={index}
                      href={contrib.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center justify-between p-3 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors group"
                    >
                      <span className="font-medium text-text-primary group-hover:text-brand-400 transition-colors">
                        {contrib.project}
                      </span>
                      <span className="text-sm text-text-muted">
                        {contrib.contributions} contributions
                      </span>
                    </a>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Contact Info */}
          <Card className="border-border-subtle">
            <CardHeader>
              <CardTitle className="text-lg">Contact</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {profile.socialLinks.github && (
                <a
                  href={profile.socialLinks.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Github className="w-5 h-5" />
                  <span className="text-sm">GitHub</span>
                </a>
              )}
              {profile.socialLinks.twitter && (
                <a
                  href={profile.socialLinks.twitter}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Twitter className="w-5 h-5" />
                  <span className="text-sm">Twitter</span>
                </a>
              )}
              {profile.socialLinks.linkedin && (
                <a
                  href={profile.socialLinks.linkedin}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Linkedin className="w-5 h-5" />
                  <span className="text-sm">LinkedIn</span>
                </a>
              )}
              {profile.website && (
                <a
                  href={profile.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Globe className="w-5 h-5" />
                  <span className="text-sm">Website</span>
                </a>
              )}
            </CardContent>
          </Card>

          {/* Languages */}
          {profile.languages && profile.languages.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg">Languages</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {profile.languages.map((lang) => (
                    <Badge key={lang} variant="secondary">
                      {lang}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Skills Preview */}
          <SkillsSection skills={profile.skills.slice(0, 8)} />
        </div>
      </div>
    </motion.div>
  );
}

// ============================================================================
// Settings Tab (for own profile)
// ============================================================================

function SettingsTab() {
  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <Card className="border-border-subtle">
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Settings className="w-5 h-5 text-brand-500" />
            Profile Settings
          </CardTitle>
          <CardDescription>
            Manage your profile visibility and preferences
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-muted">
            Profile settings would be managed here. This is a placeholder for future implementation.
          </p>
        </CardContent>
      </Card>
    </motion.div>
  );
}

// ============================================================================
// Main Profile Page Component
// ============================================================================

interface ProfilePageProps {
  username?: string;
  isOwnProfile?: boolean;
}

export function ProfilePage({ username: propUsername, isOwnProfile: propIsOwnProfile }: ProfilePageProps = {}) {
  const { username: paramUsername } = useParams<{ username: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const urlTab = searchParams.get("tab") as ProfileTab | null;

  const username = propUsername || paramUsername;
  const isOwnProfile = propIsOwnProfile ?? false;

  // Tab state with URL sync
  const [activeTab, setActiveTab] = useState<ProfileTab>(urlTab || "overview");

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

  // Fetch profile data from API
  const {
    data: profile,
    isLoading,
    isError,
    error,
  } = useQuery<UserProfile>({
    queryKey: ["enhanced-profile", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");

      // Fetch user profile and registry functions in parallel
      const [profileResponse, functionsResponse] = await Promise.all([
        usersApi.getPublicProfile(username),
        registryApi.getFunctions({ author: username, limit: 100 }),
      ]);

      // Transform API response to UserProfile format
      return transformToUserProfile(
        profileResponse,
        functionsResponse.functions || []
      );
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  // Fetch analytics data from real API
  const { data: analyticsResponse } = useQuery<UserAnalyticsResponse>({
    queryKey: ["profile-analytics", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      return usersApi.getUserAnalytics(username);
    },
    enabled: !!username && (activeTab === "analytics" || activeTab === "overview"),
    staleTime: 5 * 60 * 1000,
  });

  // Transform API analytics to ProfileAnalytics format
  const analytics: ProfileAnalytics | undefined = analyticsResponse ? {
    executionHistory: analyticsResponse.executionStats?.executionHistory?.map(h => ({
      date: h.date,
      executions: Number(h.executions) || 0,
      uniqueUsers: Number(h.uniqueUsers) || 0,
    })) || [],
    popularFunctions: analyticsResponse.popularFunctions?.map(f => ({
      functionId: f.id,
      name: f.name,
      executions: Number(f.executionCount) || 0,
      percentage: 0, // Calculate if needed
    })) || [],
    geographicDistribution: analyticsResponse.geographicStats?.regions?.map(r => ({
      country: r.region,
      executions: Number(r.executions) || 0,
      percentage: 0, // Calculate if needed
    })) || [],
    deviceStats: analyticsResponse.deviceStats?.devices?.map(d => ({
      device: d.device,
      percentage: 0, // Calculate if needed
    })) || [],
    browserStats: [], // Not yet implemented in API
  } : undefined;

  // Fetch achievements data from real API
  const { data: achievementsResponse } = useQuery<UserAchievementsResponse>({
    queryKey: ["profile-achievements", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      return usersApi.getUserAchievements(username);
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  // Fetch activity data from real API
  const { data: activityResponse } = useQuery<UserActivityResponse>({
    queryKey: ["profile-activity", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      return usersApi.getUserActivity(username, { limit: 20 });
    },
    enabled: !!username && (activeTab === "activity" || activeTab === "overview"),
    staleTime: 5 * 60 * 1000,
  });

  // Fetch skills data from real API
  const { data: skillsResponse } = useQuery<UserSkillsResponse>({
    queryKey: ["profile-skills", username],
    queryFn: async () => {
      if (!username) throw new Error("Username is required");
      return usersApi.getUserSkills(username);
    },
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
  });

  // Transform API achievements to component format
  const achievementsData: AchievementType[] = achievementsResponse?.achievements?.map(a => ({
    id: a.id,
    name: a.name,
    description: a.description,
    icon: a.icon || "Award",
    color: a.color || "blue",
    unlockedAt: a.earnedAt,
    tier: a.isCompleted ? (a.points >= 500 ? "platinum" : a.points >= 200 ? "gold" : a.points >= 100 ? "silver" : "bronze") : "bronze",
    progress: {
      current: a.progress,
      target: 100,
    },
  })) || [];

  // Transform API activity to component format
  const activityData: UserActivity[] = activityResponse?.activities?.map(a => {
    const typeMap: Record<string, ActivityType> = {
      function_published: "function_published",
      function_updated: "function_updated",
      badge_earned: "achievement_earned",
      profile_updated: "deployment",
      review_submitted: "review_received",
      comment_posted: "contribution",
    };
    return {
      id: a.id,
      type: typeMap[a.type] || "contribution",
      title: a.title,
      description: a.description || "",
      timestamp: a.createdAt,
      metadata: a.metadata,
    };
  }) || [];

  // Transform API skills to component format
  const skillsData: Skill[] = skillsResponse?.skills?.map(s => ({
    id: s.id,
    name: s.name,
    level: s.level,
    category: (s.category as Skill["category"]) || "concept",
  })) || [];

  // Create merged profile with real API data
  const mergedProfile: UserProfile | undefined = profile ? {
    ...profile,
    achievements: achievementsData.length > 0 ? achievementsData : profile.achievements,
    recentActivity: activityData.length > 0 ? activityData : profile.recentActivity,
    skills: skillsData.length > 0 ? skillsData : profile.skills,
  } : undefined;

  const tabs: { value: ProfileTab; label: string; icon: React.ReactNode }[] = [
    { value: "overview", label: "Overview", icon: <User className="w-4 h-4" /> },
    { value: "functions", label: "Functions", icon: <Package className="w-4 h-4" /> },
    { value: "activity", label: "Activity", icon: <Activity className="w-4 h-4" /> },
    { value: "analytics", label: "Analytics", icon: <BarChart3 className="w-4 h-4" /> },
    { value: "about", label: "About", icon: <BookOpen className="w-4 h-4" /> },
  ];

  if (isOwnProfile) {
    tabs.push({ value: "settings", label: "Settings", icon: <Settings className="w-4 h-4" /> });
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar variant="landing" />

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
              {/* Profile Header */}
              <ProfileHeader
                profile={profile}
                isOwnProfile={isOwnProfile}
                onEditProfile={() => setActiveTab("settings")}
              />

              {/* Stats Overview */}
              <StatsOverview stats={profile.stats} />

              {/* Navigation Tabs */}
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

                  <AnimatePresence mode="wait">
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
                      <AboutTab profile={mergedProfile!} />
                    </TabsContent>
                    <TabsContent value="settings" className="m-0">
                      <SettingsTab />
                    </TabsContent>
                  </AnimatePresence>
                </Tabs>
              </div>
            </motion.div>
          )}
        </div>
      </main>
    </div>
  );
}

export default ProfilePage;

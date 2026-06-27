/**
 * Data Transformers for ProfilePage
 *
 * Contains functions to transform API responses into component-ready data formats.
 */

import type { RegistryFunction } from '@/api/registry';
import type { FunctionCardData, UserProfile } from '@/types';
import { format, subDays } from 'date-fns';

/**
 * Transform registry function to FunctionCardData format
 */
export function transformRegistryFunction(
  fn: RegistryFunction,
  author: { id: string; username: string; name: string; avatar?: string }
): FunctionCardData {
  return {
    id: fn.id,
    name: fn.name,
    description: fn.description || '',
    author,
    trustScore: Math.round(fn.overall_score * 100),
    metrics: {
      executionCount: 0, // Will be populated from stats if needed
      executionTrend: [],
      averageLatency: 0,
      errorRate: 0,
    },
    pricing: {
      model: fn.price_per_call > 0 ? 'per_call' : 'free',
      pricePerCall: fn.price_per_call,
      currency: 'USD',
    },
    isVerified: fn.overall_score >= 0.8,
    isDeterministic: fn.deterministic_score >= 0.9,
    rating: {
      average: fn.total_ratings > 0 ? fn.overall_score * 5 : 0,
      count: fn.total_ratings,
      distribution: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 },
    },
    tags: fn.tags || [],
    category: fn.category || 'other',
    language: fn.latest_version ? 'typescript' : 'unknown', // Default to typescript
    lastUpdated: fn.created_at,
    version: fn.latest_version || '1.0.0',
    isFavorite: false,
    isFeatured: false,
  };
}

/**
 * Generate empty contribution graph (365 days)
 */
export function generateEmptyContributionGraph(): UserProfile['stats']['contributionGraph'] {
  const data: UserProfile['stats']['contributionGraph'] = [];
  for (let i = 364; i >= 0; i--) {
    const date = subDays(new Date(), i);
    data.push({
      date: format(date, 'yyyy-MM-dd'),
      count: 0,
      level: 0,
    });
  }
  return data;
}

/**
 * Transform API response to UserProfile format
 */
export function transformToUserProfile(
  apiProfile: import('@/types').PublicUserProfile,
  registryFunctions: RegistryFunction[]
): UserProfile {
  const authorInfo = {
    id: apiProfile.id,
    username: apiProfile.username,
    name: apiProfile.name,
    avatar: apiProfile.avatar,
  };

  const publishedFunctions = registryFunctions.map((fn) =>
    transformRegistryFunction(fn, authorInfo)
  );

  const s = apiProfile.stats;
  const totalExecutionsFromList = publishedFunctions.reduce(
    (sum, f) => sum + (f.metrics?.executionCount || 0),
    0
  );
  const totalExecutions = s?.totalExecutions ?? totalExecutionsFromList;
  const totalViews = totalExecutions * 2; // Estimate when no separate view metric

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
    isOnline: apiProfile.isOnline ?? false,
    lastActive: apiProfile.lastActive,
    profileNumber: apiProfile.profileNumber,
    founderNumber: apiProfile.founderNumber,
    role: apiProfile.role, // Platform admin role for badge display (public or own profile)
    experience: [],
    education: [],
    openSourceContributions: [],
    languages: [],
    stats: {
      functionsPublished: s?.functionsCount ?? publishedFunctions.length,
      functionsTrend: 0,
      totalExecutions,
      executionsTrend: 0,
      totalViews,
      viewsTrend: 0,
      trustScore:
        s?.trustScore ??
        (publishedFunctions.length > 0
          ? Math.round(
              publishedFunctions.reduce((sum, f) => sum + f.trustScore, 0) /
                publishedFunctions.length
            )
          : 0),
      reputationRank: 'Contributor',
      followersCount: s?.followersCount ?? 0,
      followingCount: s?.followingCount ?? 0,
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
    certifications: [], // Will be populated from separate API call
    publishedFunctions,
  };
}

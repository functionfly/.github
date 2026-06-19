/**
 * Mock Data for ProfilePage Component
 *
 * Comprehensive mock data for demonstration and development purposes.
 */

import type {
  UserProfile,
  ProfileAnalytics,
  Achievement,
  UserActivity,
  Skill,
  FunctionCardData,
} from "@/types";
import type { PublicBadge } from "@/api/certification";
import { subDays, format } from "date-fns";

// ============================================================================
// Helper Functions
// ============================================================================

function generateContributionGraph() {
  const data: { date: string; count: number; level: 0 | 1 | 2 | 3 | 4 }[] = [];
  for (let i = 364; i >= 0; i--) {
    const date = subDays(new Date(), i);
    const count = Math.random() > 0.6 ? Math.floor(Math.random() * 20) : 0;
    let level: 0 | 1 | 2 | 3 | 4 = 0;
    if (count > 0) level = 1;
    if (count > 5) level = 2;
    if (count > 10) level = 3;
    if (count > 15) level = 4;
    data.push({
      date: format(date, "yyyy-MM-dd"),
      count,
      level,
    });
  }
  return data;
}

function generateExecutionHistory(days: number) {
  return Array.from({ length: days }, (_, i) => ({
    date: format(subDays(new Date(), days - i - 1), "yyyy-MM-dd"),
    executions: Math.floor(Math.random() * 5000) + 1000,
    uniqueUsers: Math.floor(Math.random() * 500) + 100,
  }));
}

// ============================================================================
// Mock Functions
// ============================================================================

const mockFunctions: FunctionCardData[] = [
  {
    id: "fn-001",
    name: "json-formatter",
    description: "A powerful JSON formatter and validator with syntax highlighting and error detection. Supports large files up to 10MB.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 94,
    metrics: {
      executionCount: 125420,
      executionTrend: [1200, 1350, 1100, 1450, 1600, 1800, 2100],
      averageLatency: 45,
      errorRate: 0.02,
    },
    pricing: {
      model: "free",
      pricePerCall: 0,
      currency: "USD",
    },
    isVerified: true,
    isDeterministic: true,
    rating: {
      average: 4.8,
      count: 342,
      distribution: { 1: 2, 2: 5, 3: 15, 4: 78, 5: 242 },
    },
    tags: ["json", "formatter", "validator", "utility"],
    category: "utilities",
    language: "typescript",
    lastUpdated: "2024-03-01T10:30:00Z",
    version: "2.3.1",
    isFavorite: false,
    isFeatured: true,
  },
  {
    id: "fn-002",
    name: "image-resizer",
    description: "High-performance image resizing and optimization service. Supports WebP, JPEG, PNG conversion with quality control.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 89,
    metrics: {
      executionCount: 89300,
      executionTrend: [800, 950, 1100, 1050, 1200, 1350, 1400],
      averageLatency: 120,
      errorRate: 0.05,
    },
    pricing: {
      model: "per_call",
      pricePerCall: 0.001,
      currency: "USD",
    },
    isVerified: true,
    isDeterministic: false,
    rating: {
      average: 4.6,
      count: 189,
      distribution: { 1: 3, 2: 8, 3: 22, 4: 56, 5: 100 },
    },
    tags: ["image", "resize", "optimization", "media"],
    category: "media",
    language: "typescript",
    lastUpdated: "2024-02-28T14:20:00Z",
    version: "1.5.2",
    isFavorite: true,
    isFeatured: true,
  },
  {
    id: "fn-003",
    name: "csv-to-json",
    description: "Convert CSV files to JSON with schema detection, type inference, and streaming support for large datasets.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 91,
    metrics: {
      executionCount: 67800,
      executionTrend: [600, 700, 800, 750, 900, 950, 1000],
      averageLatency: 85,
      errorRate: 0.03,
    },
    pricing: {
      model: "free",
      pricePerCall: 0,
      currency: "USD",
    },
    isVerified: true,
    isDeterministic: true,
    rating: {
      average: 4.7,
      count: 156,
      distribution: { 1: 1, 2: 4, 3: 12, 4: 45, 5: 94 },
    },
    tags: ["csv", "json", "converter", "data"],
    category: "data",
    language: "typescript",
    lastUpdated: "2024-02-25T09:15:00Z",
    version: "1.8.0",
    isFavorite: false,
    isFeatured: false,
  },
  {
    id: "fn-004",
    name: "jwt-validator",
    description: "Secure JWT token validation with signature verification, expiration checking, and custom claim validation.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 96,
    metrics: {
      executionCount: 234500,
      executionTrend: [2000, 2200, 2100, 2500, 2800, 3000, 3200],
      averageLatency: 15,
      errorRate: 0.01,
    },
    pricing: {
      model: "free",
      pricePerCall: 0,
      currency: "USD",
    },
    isVerified: true,
    isDeterministic: true,
    rating: {
      average: 4.9,
      count: 523,
      distribution: { 1: 1, 2: 2, 3: 8, 4: 62, 5: 450 },
    },
    tags: ["jwt", "auth", "security", "validator"],
    category: "security",
    language: "typescript",
    lastUpdated: "2024-03-02T16:45:00Z",
    version: "3.1.0",
    isFavorite: true,
    isFeatured: true,
  },
  {
    id: "fn-005",
    name: "password-generator",
    description: "Generate secure, random passwords with customizable length, character sets, and entropy requirements.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 88,
    metrics: {
      executionCount: 45600,
      executionTrend: [400, 450, 500, 480, 550, 600, 650],
      averageLatency: 12,
      errorRate: 0.02,
    },
    pricing: {
      model: "free",
      pricePerCall: 0,
      currency: "USD",
    },
    isVerified: true,
    isDeterministic: false,
    rating: {
      average: 4.5,
      count: 98,
      distribution: { 1: 2, 2: 3, 3: 10, 4: 35, 5: 48 },
    },
    tags: ["password", "security", "generator", "utility"],
    category: "security",
    language: "typescript",
    lastUpdated: "2024-02-20T11:30:00Z",
    version: "1.2.0",
    isFavorite: false,
    isFeatured: false,
  },
  {
    id: "fn-006",
    name: "url-shortener",
    description: "Fast URL shortening service with custom aliases, click analytics, and QR code generation.",
    author: {
      id: "user-001",
      username: "sarahchen",
      name: "Sarah Chen",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
    },
    trustScore: 85,
    metrics: {
      executionCount: 32100,
      executionTrend: [300, 320, 350, 380, 400, 420, 450],
      averageLatency: 25,
      errorRate: 0.04,
    },
    pricing: {
      model: "per_call",
      pricePerCall: 0.0005,
      currency: "USD",
    },
    isVerified: false,
    isDeterministic: false,
    rating: {
      average: 4.3,
      count: 76,
      distribution: { 1: 2, 2: 5, 3: 15, 4: 28, 5: 26 },
    },
    tags: ["url", "shortener", "utility", "analytics"],
    category: "utilities",
    language: "typescript",
    lastUpdated: "2024-02-15T13:00:00Z",
    version: "0.9.5",
    isFavorite: false,
    isFeatured: false,
  },
];

// ============================================================================
// Mock Achievements
// ============================================================================

const mockAchievements: Achievement[] = [
  {
    id: "ach-001",
    name: "First Function",
    description: "Published your first function to the registry",
    icon: "rocket",
    color: "#6366f1",
    unlockedAt: "2023-06-15T10:00:00Z",
    tier: "bronze",
  },
  {
    id: "ach-002",
    name: "Popular Maker",
    description: "Reached 100,000 total function executions",
    icon: "trending",
    color: "#10b981",
    unlockedAt: "2023-09-20T14:30:00Z",
    tier: "gold",
  },
  {
    id: "ach-003",
    name: "Trusted Developer",
    description: "Achieved a trust score of 90 or higher",
    icon: "shield",
    color: "#f59e0b",
    unlockedAt: "2023-11-05T09:15:00Z",
    tier: "silver",
  },
  {
    id: "ach-004",
    name: "Streak Master",
    description: "Maintained a 30-day contribution streak",
    icon: "flame",
    color: "#ef4444",
    unlockedAt: "2024-01-10T16:45:00Z",
    tier: "gold",
    progress: {
      current: 45,
      target: 100,
    },
  },
  {
    id: "ach-005",
    name: "Community Champion",
    description: "Received 500+ positive ratings on your functions",
    icon: "star",
    color: "#8b5cf6",
    unlockedAt: "2024-02-14T11:20:00Z",
    tier: "platinum",
  },
  {
    id: "ach-006",
    name: "Verified Creator",
    description: "Completed identity verification",
    icon: "check",
    color: "#06b6d4",
    unlockedAt: "2023-08-01T08:00:00Z",
    tier: "silver",
  },
  {
    id: "ach-007",
    name: "Enterprise Pioneer",
    description: "Upgraded to Enterprise tier - elite member with unlimited access, dedicated support, and premium features. Less than 1% of users achieve this status.",
    icon: "Crown",
    color: "#8b5cf6",
    unlockedAt: "2024-03-20T10:30:00Z",
    tier: "platinum",
  },
];

// ============================================================================
// Mock Activity
// ============================================================================

const mockActivities: UserActivity[] = [
  {
    id: "act-001",
    type: "function_published",
    title: "Published jwt-validator v3.1.0",
    description: "Added support for RS256 algorithm and improved error messages",
    timestamp: "2024-03-02T16:45:00Z",
    relatedFunction: {
      id: "fn-004",
      name: "jwt-validator",
      author: "sarahchen",
    },
  },
  {
    id: "act-002",
    type: "milestone_reached",
    title: "Reached 500,000 total executions",
    description: "Your functions have been executed over half a million times!",
    timestamp: "2024-03-01T10:00:00Z",
  },
  {
    id: "act-003",
    type: "achievement_earned",
    title: "Earned Community Champion",
    description: "Received 500+ positive ratings on your functions",
    timestamp: "2024-02-14T11:20:00Z",
  },
  {
    id: "act-003b",
    type: "membership_upgraded",
    title: "Upgraded to Enterprise",
    description: "Unlimited functions, dedicated support, and premium features",
    timestamp: "2024-03-20T10:30:00Z",
    metadata: {
      plan: "enterprise",
      previousPlan: "professional",
    },
  },
  {
    id: "act-004",
    type: "function_updated",
    title: "Updated json-formatter v2.3.1",
    description: "Fixed memory leak with large JSON files",
    timestamp: "2024-03-01T10:30:00Z",
    relatedFunction: {
      id: "fn-001",
      name: "json-formatter",
      author: "sarahchen",
    },
  },
  {
    id: "act-005",
    type: "review_received",
    title: "Received a 5-star review",
    description: '"This JWT validator saved me hours of work. Thank you!"',
    timestamp: "2024-02-28T09:15:00Z",
    relatedFunction: {
      id: "fn-004",
      name: "jwt-validator",
      author: "sarahchen",
    },
  },
  {
    id: "act-006",
    type: "follower_gained",
    title: "Gained 100 new followers",
    description: "Your profile is gaining traction in the community",
    timestamp: "2024-02-25T14:00:00Z",
  },
  {
    id: "act-007",
    type: "contribution",
    title: "30-day contribution streak",
    description: "You've been active for 30 consecutive days",
    timestamp: "2024-02-20T08:00:00Z",
  },
  {
    id: "act-008",
    type: "deployment",
    title: "Deployed to production",
    description: "Successfully deployed 3 functions to production",
    timestamp: "2024-02-18T16:30:00Z",
  },
];

// ============================================================================
// Mock Skills
// ============================================================================

const mockSkills: Skill[] = [
  { name: "TypeScript", level: "expert", category: "language", endorsements: 45 },
  { name: "JavaScript", level: "expert", category: "language", endorsements: 52 },
  { name: "Python", level: "advanced", category: "language", endorsements: 28 },
  { name: "Go", level: "intermediate", category: "language", endorsements: 15 },
  { name: "React", level: "expert", category: "framework", endorsements: 38 },
  { name: "Node.js", level: "expert", category: "framework", endorsements: 42 },
  { name: "Next.js", level: "advanced", category: "framework", endorsements: 25 },
  { name: "Docker", level: "advanced", category: "tool", endorsements: 30 },
  { name: "Kubernetes", level: "intermediate", category: "tool", endorsements: 18 },
  { name: "AWS", level: "advanced", category: "platform", endorsements: 35 },
  { name: "Serverless", level: "expert", category: "concept", endorsements: 40 },
  { name: "GraphQL", level: "advanced", category: "concept", endorsements: 22 },
  { name: "REST APIs", level: "expert", category: "concept", endorsements: 48 },
  { name: "PostgreSQL", level: "advanced", category: "tool", endorsements: 27 },
];

// ============================================================================
// Mock Profile
// ============================================================================

export const mockUserProfile: UserProfile = {
  id: "user-001",
  username: "sarahchen",
  name: "Sarah Chen",
  avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=sarah",
  coverImage: undefined,
  bio: "Full-stack developer passionate about serverless architecture and developer tools. Building useful functions to make developers' lives easier. Former tech lead at Stripe, now exploring the future of edge computing.",
  location: "San Francisco, CA",
  company: "FunctionFly",
  jobTitle: "Senior Developer Advocate",
  website: "https://sarahchen.dev",
  socialLinks: {
    github: "https://github.com/sarahchen",
    twitter: "https://twitter.com/sarahchen",
    linkedin: "https://linkedin.com/in/sarahchen",
    website: "https://sarahchen.dev",
    discord: "https://discord.gg/sarahchen",
  },
  skills: mockSkills,
  createdAt: "2023-01-15T08:00:00Z",
  updatedAt: "2024-03-02T10:00:00Z",
  isOnline: true,
  lastActive: "2024-03-02T10:00:00Z",

  experience: [
    {
      company: "FunctionFly",
      title: "Senior Developer Advocate",
      startDate: "2023-06-01T00:00:00Z",
      current: true,
      description: "Building the future of serverless computing and helping developers succeed on the platform.",
    },
    {
      company: "Stripe",
      title: "Tech Lead",
      startDate: "2020-03-01T00:00:00Z",
      endDate: "2023-05-31T00:00:00Z",
      current: false,
      description: "Led the developer experience team, focusing on APIs and SDKs.",
    },
    {
      company: "GitHub",
      title: "Senior Software Engineer",
      startDate: "2018-01-15T00:00:00Z",
      endDate: "2020-02-28T00:00:00Z",
      current: false,
      description: "Worked on the Actions platform and developer tools.",
    },
  ],

  education: [
    {
      institution: "Stanford University",
      degree: "M.S.",
      field: "Computer Science",
      startDate: "2016-09-01T00:00:00Z",
      endDate: "2018-06-15T00:00:00Z",
    },
    {
      institution: "UC Berkeley",
      degree: "B.S.",
      field: "Electrical Engineering & Computer Science",
      startDate: "2012-09-01T00:00:00Z",
      endDate: "2016-05-15T00:00:00Z",
    },
  ],

  openSourceContributions: [
    {
      project: "serverless-framework",
      url: "https://github.com/serverless/serverless",
      contributions: 47,
    },
    {
      project: "next.js",
      url: "https://github.com/vercel/next.js",
      contributions: 23,
    },
    {
      project: "react",
      url: "https://github.com/facebook/react",
      contributions: 12,
    },
  ],

  languages: ["English", "Mandarin", "Spanish"],

  stats: {
    functionsPublished: 6,
    functionsTrend: 20,
    totalExecutions: 594620,
    executionsTrend: 15,
    totalViews: 1250000,
    viewsTrend: 25,
    trustScore: 91,
    reputationRank: "Top 1%",
    followersCount: 2847,
    followingCount: 342,
    followersTrend: 12,
    contributionStreak: {
      current: 45,
      longest: 67,
      lastContribution: "2024-03-02T10:00:00Z",
    },
    contributionGraph: generateContributionGraph(),
  },

  achievements: mockAchievements,
  recentActivity: mockActivities,
  certifications: [
    {
      tier_slug: "verified-developer",
      tier_name: "Verified Developer",
      tier_color: "#22c55e",
      tier_icon: "badge-check",
      credential_number: "FF-2024-00142",
      issued_at: "2024-01-15T00:00:00Z",
      expires_at: "2025-01-15T00:00:00Z",
    },
    {
      tier_slug: "top-publisher",
      tier_name: "Top Publisher",
      tier_color: "#8b5cf6",
      tier_icon: "star",
      credential_number: "FF-2024-00089",
      issued_at: "2024-02-20T00:00:00Z",
      expires_at: "2025-02-20T00:00:00Z",
    },
  ] as PublicBadge[],
  publishedFunctions: mockFunctions,
};

// ============================================================================
// Mock Analytics
// ============================================================================

export const mockAnalytics: ProfileAnalytics = {
  executionHistory: generateExecutionHistory(30),
  popularFunctions: [
    { functionId: "fn-004", name: "jwt-validator", executions: 234500, percentage: 39 },
    { functionId: "fn-001", name: "json-formatter", executions: 125420, percentage: 21 },
    { functionId: "fn-002", name: "image-resizer", executions: 89300, percentage: 15 },
    { functionId: "fn-003", name: "csv-to-json", executions: 67800, percentage: 11 },
    { functionId: "fn-005", name: "password-generator", executions: 45600, percentage: 8 },
    { functionId: "fn-006", name: "url-shortener", executions: 32100, percentage: 5 },
  ],
  revenueHistory: [
    { date: "2024-02-01", revenue: 125.5, calls: 12500 },
    { date: "2024-02-02", revenue: 142.3, calls: 14230 },
    { date: "2024-02-03", revenue: 98.7, calls: 9870 },
    { date: "2024-02-04", revenue: 156.2, calls: 15620 },
    { date: "2024-02-05", revenue: 178.9, calls: 17890 },
    { date: "2024-02-06", revenue: 134.5, calls: 13450 },
    { date: "2024-02-07", revenue: 189.4, calls: 18940 },
  ],
  geographicDistribution: [
    { country: "United States", executions: 245000, percentage: 41 },
    { country: "Germany", executions: 89000, percentage: 15 },
    { country: "United Kingdom", executions: 72000, percentage: 12 },
    { country: "India", executions: 65000, percentage: 11 },
    { country: "Canada", executions: 48000, percentage: 8 },
    { country: "Japan", executions: 35000, percentage: 6 },
    { country: "France", executions: 28000, percentage: 5 },
    { country: "Other", executions: 26620, percentage: 2 },
  ],
  deviceStats: [
    { device: "Desktop", percentage: 65 },
    { device: "Mobile", percentage: 28 },
    { device: "Tablet", percentage: 7 },
  ],
  browserStats: [
    { browser: "Chrome", percentage: 58 },
    { browser: "Safari", percentage: 18 },
    { browser: "Firefox", percentage: 12 },
    { browser: "Edge", percentage: 8 },
    { browser: "Other", percentage: 4 },
  ],
};

// ============================================================================
// Empty State Data
// ============================================================================

export const emptyUserProfile: UserProfile = {
  ...mockUserProfile,
  stats: {
    ...mockUserProfile.stats,
    functionsPublished: 0,
    totalExecutions: 0,
    totalViews: 0,
    trustScore: 50,
    followersCount: 0,
    followingCount: 0,
    contributionStreak: {
      current: 0,
      longest: 0,
      lastContribution: "",
    },
    contributionGraph: generateContributionGraph().map(d => ({ ...d, count: 0, level: 0 })),
  },
  achievements: [],
  recentActivity: [],
  certifications: [],
  publishedFunctions: [],
};

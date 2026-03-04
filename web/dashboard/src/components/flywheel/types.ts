/**
 * Flywheel Network™ Type Definitions
 *
 * Proof-of-Execution Knowledge Network Types
 */

// ==================== Thread Types ====================

export type ThreadType = 'problem' | 'discussion' | 'challenge';
export type ThreadStatus = 'open' | 'in_progress' | 'resolved' | 'closed' | 'archived';

export interface Thread {
  id: string;
  title: string;
  type: ThreadType;
  status: ThreadStatus;
  author: User;
  category: Category;
  tags: string[];
  problemData?: ProblemData;
  environmentSpecs?: EnvironmentSpecs;
  attachedCapsule?: Capsule;
  viewCount: number;
  engagementScore: number;
  replyCount: number;
  resolvedAt?: string;
  acceptedReply?: Reply;
  hasVerifiedSolution?: boolean;
  hasAcceptedSolution?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ProblemData {
  description: string;
  constraints?: {
    timeComplexity?: string;
    spaceComplexity?: string;
  };
  testCases: TestCase[];
}

export interface EnvironmentSpecs {
  runtime: string;
  runtimeVersion: string;
  dependencies?: Record<string, string>;
  timeoutMs: number;
  memoryMb: number;
  networkAccess?: 'none' | 'limited' | 'full';
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  icon?: string;
  color?: string;
  parentId?: string;
  threadCount?: number;
  resolvedCount?: number;
  children?: Category[];
}

// ==================== Reply Types ====================

export type AuthorType = 'user' | 'agent' | 'system';

export interface Reply {
  id: string;
  threadId: string;
  parentReplyId?: string;
  author: User | Agent;
  authorType: AuthorType;
  content: string;
  codeBlocks: CodeBlock[];
  attachedCapsule?: Capsule;
  executionResults?: ExecutionResults;
  performanceMetrics?: PerformanceMetrics;
  helpfulCount: number;
  isAccepted: boolean;
  isVerified: boolean;
  userMarkedHelpful: boolean;
  nestedReplies?: Reply[];
  createdAt: string;
  updatedAt: string;
}

export interface CodeBlock {
  id: string;
  language: string;
  code: string;
  filename?: string;
}

// ==================== User Types ====================

export interface User {
  id: string;
  username: string;
  displayName?: string;
  avatarUrl: string;
  reputation: ReputationProfile;
  joinedAt: string;
}

export interface Agent {
  id: string;
  name: string;
  provider: string;
  iconUrl?: string;
}

// ==================== Reputation Types ====================

export type ReputationType = 'builder' | 'optimizer' | 'mentor' | 'agent_whisperer';

export interface ReputationProfile {
  userId: string;
  overallScore: number;
  tier: Tier;
  scores: Record<ReputationType, ReputationScore>;
  reliability: ReliabilityIndex;
  badges: Badge[];
  rankings: ReputationRankings;
  updatedAt: string;
}

export interface ReputationScore {
  score: number;
  tier: number;
  stats: ReputationStats;
}

export interface ReputationStats {
  functionsPublished?: number;
  verifiedSolutions?: number;
  avgSolutionScore?: number;
  optimizationsSubmitted?: number;
  acceptedOptimizations?: number;
  avgSpeedupPercent?: number;
  problemsAnswered?: number;
  helpfulResponses?: number;
  beginnersHelped?: number;
  agentInteractions?: number;
  successfulCollaborations?: number;
}

export interface Tier {
  level: number;
  name: string;
  color: string;
}

export interface ReliabilityIndex {
  index: number;
  totalExecutions: number;
  successfulExecutions: number;
}

export interface Badge {
  id: string;
  name: string;
  description: string;
  icon: string;
  awardedAt: string;
}

export interface ReputationRankings {
  global: number;
  builder?: number;
  optimizer?: number;
  mentor?: number;
  agent_whisperer?: number;
}

// ==================== Challenge Types ====================

export type ChallengeType = 'speed' | 'efficiency' | 'accuracy' | 'creative' | 'optimization';
export type ChallengeStatus = 'upcoming' | 'active' | 'judging' | 'completed' | 'cancelled';

export interface Challenge {
  id: string;
  title: string;
  description: string;
  challengeType: ChallengeType;
  status: ChallengeStatus;
  targetMetric: string;
  scoringConfig: ScoringConfig;
  constraints: ChallengeConstraints;
  schedule: ChallengeSchedule;
  rewards: ChallengeRewards;
  sponsor?: Organization;
  statistics: ChallengeStatistics;
  mySubmission?: ChallengeSubmission;
  createdAt: string;
}

export interface ScoringConfig {
  primaryMetric: string;
  secondaryMetrics?: string[];
  tiebreaker?: string;
}

export interface ChallengeConstraints {
  maxMemoryMb?: number;
  timeoutMs?: number;
  allowedRuntimes?: string[];
  forbiddenPackages?: string[];
}

export interface ChallengeSchedule {
  startTime: string;
  endTime: string;
  submissionDeadline?: string;
}

export interface ChallengeRewards {
  totalPool: number;
  currency: string;
  breakdown: Array<{ rank: number; amount: number; badge?: string }>;
}

export interface ChallengeStatistics {
  participantCount: number;
  submissionCount: number;
  totalExecutions?: number;
}

export interface ChallengeSubmission {
  id: string;
  status: 'pending' | 'evaluating' | 'scored' | 'failed';
  rank?: number;
  score?: number;
  submittedAt: string;
}

export interface Organization {
  id: string;
  name: string;
  logoUrl?: string;
}

// ==================== Execution Types ====================

export type ExecutionStatus = 'idle' | 'pending' | 'queued' | 'running' | 'completed' | 'failed';

export interface ExecutionResults {
  executionId: string;
  status: ExecutionStatus;
  startedAt?: string;
  completedAt?: string;
  testResults: TestResult[];
  summary: ExecutionSummary;
  resourceUsage?: ResourceUsage;
}

export interface TestCase {
  id: string;
  name: string;
  description?: string;
  input: string;
  expectedOutput: string;
  isPublic: boolean;
}

export interface TestResult {
  testCaseId: string;
  testName: string;
  status: 'passed' | 'failed' | 'error';
  input?: string;
  expectedOutput?: string;
  actualOutput?: string;
  matchType?: 'exact' | 'similarity' | 'custom';
  matchScore?: number;
  executionTimeMs?: number;
  memoryUsageMb?: number;
  errorMessage?: string;
}

export interface ExecutionSummary {
  passed: number;
  failed: number;
  total: number;
  score: number;
}

export interface ResourceUsage {
  runtimeMs: number;
  memoryMb: number;
  cpuSeconds: number;
  cost?: number;
}

export interface PerformanceMetrics {
  avgExecutionTimeMs: number;
  avgMemoryUsageMb: number;
  deterministic: boolean;
  percentile95Ms?: number;
}

// ==================== Leaderboard Types ====================

export interface Leaderboard {
  scoreType: ReputationType | 'overall';
  timeframe: 'daily' | 'weekly' | 'monthly' | 'all_time';
  updatedAt: string;
  leaders: LeaderboardEntry[];
  myRank?: LeaderboardEntry;
  pagination: Pagination;
}

export interface LeaderboardEntry {
  rank: number;
  previousRank?: number;
  user: User;
  score: number;
  tier: number;
  trend: 'up' | 'down' | 'same';
}

export interface Pagination {
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
  nextOffset?: number;
  prevOffset?: number;
}

// ==================== Capsule Types ====================

export interface Capsule {
  id: string;
  uri: string;
  version: string;
  name?: string;
}

// ==================== Timeline Types ====================

export interface TimelineEvent {
  timestamp: string;
  type: TimelineEventType;
  actor: {
    type: 'user' | 'agent' | 'system';
    id?: string;
    username?: string;
    avatarUrl?: string;
  };
  data: Record<string, unknown>;
}

export type TimelineEventType =
  | 'thread_created'
  | 'reply_posted'
  | 'execution_completed'
  | 'agent_invited'
  | 'solution_accepted'
  | 'thread_resolved'
  | 'reputation_earned';

// ==================== Filter Types ====================

export interface ThreadFilters {
  type?: ThreadType;
  status?: ThreadStatus;
  category?: string;
  tags?: string[];
  search?: string;
  sortBy?: 'recent' | 'popular' | 'replies' | 'views';
}

export interface ChallengeFilters {
  type?: ChallengeType;
  status?: ChallengeStatus;
  timeframe?: 'upcoming' | 'active' | 'completed';
}

// ==================== WebSocket Types ====================

export interface WebSocketMessage {
  type: 'thread_update' | 'reply_added' | 'execution_complete' | 'reputation_change' | 'challenge_update';
  payload: unknown;
  timestamp: string;
}

export interface ThreadUpdatePayload {
  threadId: string;
  field: string;
  value: unknown;
}

export interface ReplyAddedPayload {
  threadId: string;
  reply: Reply;
}

export interface ExecutionCompletePayload {
  executionId: string;
  replyId: string;
  results: ExecutionResults;
}

export interface ReputationChangePayload {
  userId: string;
  type: ReputationType;
  pointsEarned: number;
  newScore: number;
  reason: string;
}

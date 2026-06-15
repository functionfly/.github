export interface User {
  id: string;
  email: string;
  username?: string;
  companyName?: string;
  name: string;
  bio?: string;
  avatar?: string; // Profile picture URL from social providers
  tenantId: string;
  plan: string; // Tenant plan from API: free, starter, pro, enterprise, etc.
  role?: string; // Admin role for admin users
  createdAt: string;
  updatedAt?: string;
  profileNumber?: number; // Sequential number assigned based on registration order (e.g., Member #123)
  language?: string; // User's preferred language code (e.g., 'en', 'es')
  isOnline?: boolean; // Online status from presence system
}

export interface Session {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  token_type: string;
  user: {
    id: string;
    email: string;
    user_metadata: {
      name?: string;
      avatar_url?: string;
    };
    created_at: string;
    updated_at?: string;
  };
}

/** Aggregates from the API (registry + follows), not limited by registry list pagination. */
export interface PublicProfileStats {
  functionsCount: number;
  totalExecutions: number;
  trustScore: number;
  followersCount: number;
  followingCount: number;
}

export interface PublicUserProfile {
  id: string;
  username: string;
  name: string;
  avatar?: string;
  bio?: string;
  location?: string;
  website?: string;
  jobTitle?: string;
  companyName?: string;
  twitterUrl?: string;
  githubUrl?: string;
  linkedinUrl?: string;
  socialLinks?: Record<string, string>;
  createdAt: string;
  /** When present, use for stat cards; functions list may still be paginated separately. */
  stats?: PublicProfileStats;
  publishedFunctions: PublicRegistryFunction[];
  // Online status fields from API
  isOnline?: boolean;
  lastActive?: string;
  // Profile number for early adopter tracking (e.g., Member #123)
  profileNumber?: number;
  // Platform admin role (super_admin, admin, support) for badge display
  role?: string;
}

export interface PublicRegistryFunction {
  name: string;
  author: string;
  description: string;
  version: string;
  tags?: string[];
  executionCount?: number;
  rating?: number;
  createdAt: string;
}

export interface App {
  id: string;
  name: string;
  slug: string;
  tenantId: string;
  createdAt: string;
}

export interface Backend {
  id: string;
  /** Present on API responses (Go json: app_id) */
  app_id?: string;
  provider: string;
  region: string;
  url: string;
  sharedSecret: string;
  priority?: number;
  createdAt: string;
}

export interface CircuitState {
  state: 'closed' | 'open' | 'half-open';
  sinceTs: string;
  failCount: number;
  successCount: number;
  lastFailureTs?: string;
}

export interface HealthCheck {
  timestamp: string;
  ok: boolean;
  statusCode: number;
  latencyMs: number;
  errorMessage?: string;
}

export interface BackendStatus {
  backend: Backend;
  circuitState?: CircuitState;
  latestHealthCheck?: HealthCheck;
}

export interface AppStatus {
  app: App;
  backends: BackendStatus[];
}

export interface Deployment {
  id: string;
  appId: string;
  provider: string;
  region: string;
  status: 'pending' | 'building' | 'deploying' | 'success' | 'failed' | 'rolled_back';
  artifactUrl?: string;
  deployedUrl?: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
}

export interface RoutingDecision {
  selectedBackend?: Backend;
  failoverBackends: Backend[];
  reason: string;
  requestId: string;
}

export interface LoginRequest {
  /** Can be either an email address or a username */
  email: string;
  password: string;
  recaptchaToken?: string;
}

export interface SignupRequest {
  name?: string;
  /** ISO date YYYY-MM-DD */
  dateOfBirth: string;
  email: string;
  password: string;
  confirmPassword: string;
  termsAccepted: boolean;
  username: string;
  companyName?: string;
  inviteCode?: string;
  recaptchaToken?: string;
}

export interface SignupResponse {
  message: string;
  emailSent: boolean;
  requiresVerification: boolean;
}

export interface LoginResponse {
  token: string;
  expiresIn: number;
  user: User;
}

export interface CreateAppRequest {
  name: string;
  slug: string;
}

export interface CreateBackendRequest {
  provider: string;
  region: string;
  url: string;
  sharedSecret?: string;
  priority?: number;
}

export interface BackendTemplate {
  provider: string;
  name: string;
  icon: string;
  description: string;
  features: string[];
  supportedRegions: string[];
  maxBackends?: number;
  capabilities: string[];
  documentationUrl?: string;
}

export const BACKEND_PROVIDERS: BackendTemplate[] = [
  {
    provider: 'functionfly-edge',
    name: 'FunctionFly Edge',
    icon: '⬡',
    description:
      'FunctionFly managed edge network with global distribution, automatic scaling, and built-in DDoS protection.',
    features: ['Global CDN', 'Auto-scaling', 'DDoS Protection', 'Edge Caching', '99.9% SLA'],
    supportedRegions: ['us-east-1', 'eu-central-1'],
    capabilities: ['http', 'websocket', 'grpc'],
    documentationUrl: '/docs/backends/functionfly-edge',
  },
  {
    provider: 'workers',
    name: 'Cloudflare Workers',
    icon: '☁',
    description:
      "V8-based serverless functions running on Cloudflare's global network with sub-millisecond cold starts.",
    features: ['Global Network', 'Cold Start < 1ms', 'KV Storage', 'Durable Objects', 'Cache API'],
    supportedRegions: [
      'us-east-1',
      'us-west-1',
      'eu-west-1',
      'ap-southeast-1',
      'ap-northeast-1',
      'ap-southeast-2',
      'sa-east-1',
      'af-south-1',
    ],
    capabilities: ['http', 'websocket', 'queue'],
    documentationUrl: '/docs/backends/cloudflare-workers',
  },
  {
    provider: 'vercel',
    name: 'Vercel',
    icon: '▲',
    description:
      'Zero-config deployment platform with instant scaling, edge functions, and built-in CI/CD.',
    features: [
      'Instant Scaling',
      'Edge Functions',
      'Preview Deployments',
      'Git Integration',
      'Analytics',
    ],
    supportedRegions: [
      'arn1',
      'bom1',
      'cdg1',
      'cle1',
      'cpt1',
      'dub1',
      'fra1',
      'gru1',
      'hkg1',
      'hnd1',
      'iad1',
      'icn1',
      'jnb1',
      'lax1',
      'lhr1',
      'pdx1',
      'sfo1',
      'sin1',
      'syd1',
    ],
    capabilities: ['http', 'serverless'],
    documentationUrl: '/docs/backends/vercel',
  },
  {
    provider: 'fly',
    name: 'Fly.io',
    icon: '✈',
    description:
      "Run applications close to users worldwide with Fly's lightweight VMs and anycast routing.",
    features: [
      'Anycast Routing',
      'Fly Machines',
      'Private Networking',
      'Auto Health Checks',
      'Edge Locations',
    ],
    supportedRegions: [
      'ams',
      'arn',
      'atl',
      'bog',
      'bos',
      'bru',
      'cdg',
      'den',
      'dfw',
      'ewr',
      'eze',
      'fra',
      'gig',
      'gru',
      'hkg',
      'iad',
      'jnb',
      'lax',
      'lhr',
      'mad',
      'mia',
      'nrt',
      'ord',
      'phx',
      'pma',
      'sea',
      'sfo',
      'sin',
      'syd',
      'tsn',
      'waw',
      'yyz',
    ],
    capabilities: ['http', 'websocket', 'tcp'],
    documentationUrl: '/docs/backends/fly',
  },
  {
    provider: 'deno-deploy',
    name: 'Deno Deploy',
    icon: '  ',
    description:
      'Edge runtime powered by Deno with TypeScript native support, built-in KV, and zero config deployment.',
    features: ['TypeScript Native', 'Built-in KV', 'Queues', 'Cron Triggers', 'Edge Caching'],
    supportedRegions: ['us-east4', 'europe-west4', 'asia-southeast1', 'us-west2'],
    capabilities: ['http', 'websocket', 'kv'],
    documentationUrl: '/docs/backends/deno-deploy',
  },
];

export interface DeployRequest {
  provider: string;
  region: string;
  artifact: string;
  routes?: string[];
  envVars?: Record<string, string>;
  secrets?: Record<string, string>;
  providerConfig?: Record<string, unknown>;
}

export interface DeployResult {
  deploymentId: string;
  status: string;
  url?: string;
}

export interface AnalyticsMetrics {
  totalRequests: number;
  averageLatency: number;
  errorRate: number;
  uptimePercentage: number;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
  label?: string;
}

export interface ProviderMetrics {
  provider: string;
  requests: number;
  latency: number;
  errors: number;
}

export interface ConnectedProvider {
  /** Database row id (DELETE /v1/providers/:id uses this). */
  id: string;
  /** Provider slug from the API (e.g. workers, vercel, functionfly-edge) — not the display name. */
  name: string;
  status: 'online' | 'offline' | 'degraded' | 'pending';
  connectedAt: string;
  apiKey?: string;
  /** Whether the provider connection is stale (not used in 30+ days) */
  isStale?: boolean;
  /** ISO timestamp of when the provider was last used */
  lastUsedAt?: string;
}

export interface ProviderMaintenanceStatus {
  disabled: boolean;
  reason?: string;
  disabledAt?: string;
  disabledBy?: string;
}

export interface ConnectProviderRequest {
  providerId: string;
  apiKey: string;
}

export interface ConnectProviderResponse {
  provider: ConnectedProvider;
}

// Admin types
export interface Tenant {
  id: string;
  name: string;
  plan?: string;
  status: 'active' | 'suspended';
  createdAt: string;
  updatedAt: string;
}

export interface AuditEvent {
  id: string;
  actorUserId?: string;
  actorEmail?: string;
  tenantId?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  requestId?: string;
  beforeState?: any;
  afterState?: any;
  ipAddress?: string;
  userAgent?: string;
  timestamp: string;
  success: boolean;
}

export interface PricingTier {
  id: string;
  name: string;
  description: string;
  priceCents: number;
  currency: string;
  features: any;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Subscription {
  id: string;
  tenantId: string;
  pricingTierId: string;
  status: 'active' | 'canceled' | 'past_due' | 'trialing';
  currentPeriodStart: string;
  currentPeriodEnd: string;
  trialEnd?: string;
  cancelAtPeriodEnd: boolean;
  canceledAt?: string;
  createdAt: string;
  updatedAt: string;
  pricingTier?: PricingTier;
}

export interface Invoice {
  id: string;
  tenantId: string;
  subscriptionId?: string;
  status: 'draft' | 'open' | 'paid' | 'void' | 'uncollectible';
  amountDueCents: number;
  amountPaidCents: number;
  currency: string;
  invoicePdfUrl?: string;
  hostedInvoiceUrl?: string;
  periodStart?: string;
  periodEnd?: string;
  dueDate?: string;
  paidAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface UsageEvent {
  id: string;
  tenantId: string;
  eventType: string;
  quantity: number;
  unitPriceCents?: number;
  metadata?: any;
  timestamp: string;
}

export interface UsageRollup {
  id: string;
  tenantId: string;
  eventType: string;
  periodDate: string;
  totalQuantity: number;
  createdAt: string;
  updatedAt: string;
}

export interface Coupon {
  id: string;
  code: string;
  name: string;
  description: string;
  discountType: 'percent' | 'amount';
  discountValue: number;
  currency?: string;
  maxRedemptions?: number;
  timesRedeemed: number;
  validFrom?: string;
  validUntil?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface SystemHealth {
  status: 'healthy' | 'unhealthy';
  version: string;
  timestamp: string;
  checks: Record<
    string,
    {
      status: string;
      healthy: boolean;
      responseTimeMs?: number;
      goroutines?: number;
    }
  >;
}

// Analytics types
export interface AnalyticsService {
  name: string;
  enabled: boolean;
  status: 'loading' | 'loaded' | 'error' | 'disabled';
  config: Record<string, any>;
  lastUsed?: string;
}

export interface GoogleAnalyticsConfig {
  measurementId: string;
  enabled: boolean;
}

export interface HotjarConfig {
  siteId: string;
  enabled: boolean;
}

export interface AnalyticsSettings {
  googleAnalytics?: GoogleAnalyticsConfig;
  hotjar?: HotjarConfig;
  services: AnalyticsService[];
}

export interface UpdateAnalyticsRequest {
  googleAnalytics?: GoogleAnalyticsConfig;
  hotjar?: HotjarConfig;
}

export interface UpdateAnalyticsResponse {
  message: string;
  settings: UpdateAnalyticsRequest;
  note?: string;
}

// Function types
export interface EnvironmentVariable {
  key: string;
  value: string;
  isSecret: boolean;
}

export interface FunctionConfig {
  id: string;
  name: string;
  providers: string[];
  region: string;
  code: string;
  envVars: EnvironmentVariable[];
  tenantId: string;
  /** When set, deploy must use a backend from this app (API enforced) */
  appId?: string;
  createdAt: string;
  updatedAt: string;
  version: string;
  status: 'draft' | 'deploying' | 'deployed' | 'failed' | 'active' | 'suspended';
  /** Trust score 0-100, available when function has execution history */
  trustScore?: number;
  /** Number of times this function has been executed */
  executionCount?: number;
  /** Average execution duration in milliseconds */
  avgDurationMs?: number;
}

export interface CreateFunctionRequest {
  name: string;
  providers: string[];
  region: string;
  code: string;
  envVars?: EnvironmentVariable[];
}

export interface UpdateFunctionRequest {
  name?: string;
  providers?: string[];
  region?: string;
  code?: string;
  envVars?: EnvironmentVariable[];
}

export interface FunctionDeployment {
  id: string;
  functionId: string;
  version: string;
  status: 'pending' | 'deploying' | 'success' | 'failed';
  provider: string;
  region: string;
  deployedUrl?: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
}

export interface FunctionLog {
  id: string;
  functionId: string;
  deploymentId?: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  source: string;
  metadata?: Record<string, any>;
}

export interface DeployFunctionRequest {
  functionId: string;
  backendId: string;
  version?: string;
  environment?: 'dev' | 'staging' | 'prod';
}

export interface DeployFunctionResponse {
  functionId: string;
  deploymentId: string;
  url: string;
  region: string;
  providers: string[];
  status: string;
  deployments: FunctionDeployment[];
}

export interface TestFunctionRequest {
  functionId?: string;
  code?: string;
  envVars?: EnvironmentVariable[];
  testInput?: any;
}

export interface TestFunctionResponse {
  success: boolean;
  output?: any;
  error?: string;
  executionTimeMs: number;
  logs: FunctionLog[];
}

// Function Card Types

export type FunctionCardVariant = 'compact' | 'expanded' | 'analytics';

export type PricingModel = 'free' | 'per_call' | 'subscription' | 'revenue_share';

export interface FunctionAuthor {
  id: string;
  username: string;
  name: string;
  avatar?: string;
  profileUrl?: string;
}

export interface FunctionMetrics {
  executionCount: number;
  executionTrend?: number[]; // Last 7 days for sparkline
  averageLatency?: number;
  errorRate?: number;
}

export interface FunctionRating {
  average: number; // 0-5
  count: number;
  distribution?: Record<number, number>; // rating -> count
}

export interface FunctionCardData {
  id: string;
  name: string;
  description: string;
  author: FunctionAuthor;
  trustScore: number; // 0-100
  metrics: FunctionMetrics;
  pricing: {
    model: PricingModel;
    pricePerCall?: number; // in cents or USD
    currency?: string;
  };
  isVerified: boolean;
  isDeterministic: boolean;
  rating: FunctionRating;
  tags?: string[];
  category?: string;
  language?: string;
  lastUpdated?: string;
  version?: string;
  isFavorite?: boolean;
  isFeatured?: boolean;
}

export interface FunctionCardProps {
  data: FunctionCardData;
  variant?: FunctionCardVariant;
  className?: string;
  // Action handlers
  onView?: (id: string) => void;
  onExecute?: (id: string) => void;
  onFavorite?: (id: string, isFavorite: boolean) => void;
  onShare?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAdminAction?: (id: string, action: string) => void;
  onRate?: (id: string, stars: number) => void;
}

// Function Header Types

export type TrustTier = 'critical' | 'high' | 'medium' | 'low' | 'untrusted';

export interface FunctionHeaderData {
  /** Function name */
  name: string;
  /** Function ID */
  id: string;
  /** Hash identifier for the function execution */
  executionRootHash: string;
  /** Trust level indicator */
  trustTier: TrustTier;
  /** Economic/scoring metric (0-100) */
  economicScore: number;
  /** Runtime environment (e.g., workers, vercel, fly, deno) */
  runtime: string;
  /** Resource identifier/signature */
  resourceSignature: string;
  /** Certificate verification status */
  fxcert: {
    verified: boolean;
    issuedAt?: string;
    expiresAt?: string;
    issuer?: string;
  };
  /** Optional description */
  description?: string;
  /** Optional status for the status badge */
  status?: 'online' | 'offline' | 'degraded' | 'pending';
  /** Optional version */
  version?: string;
}

export interface FunctionHeaderProps {
  data: FunctionHeaderData;
  className?: string;
  /** Optional back button handler */
  onBack?: () => void;
  /** Optional action handlers */
  onEdit?: () => void;
  onDeploy?: () => void;
  onTest?: () => void;
  onShare?: () => void;
  /** Handler for reporting a function issue to the author */
  onReportIssue?: () => void;
}

// Trust Score Badge Types

/** Trust score band classification */
export type TrustScoreBand = 'excellent' | 'good' | 'fair' | 'poor';

/** Fraud risk level assessment */
export type FraudRiskLevel = 'low' | 'medium' | 'high';

/**
 * Comprehensive trust metrics for a function
 * All scores are 0-100 except where noted
 */
export interface TrustMetrics {
  /** Overall trust score (0-100%) */
  overallScore: number;
  /** Reliability score - uptime/execution success rate */
  reliability: number;
  /** Latency score - response time metric (0-100, higher is better) */
  latency: number;
  /** Determinism score - consistency of outputs */
  determinism: number;
  /** Community reputation score - user ratings/votes */
  communityReputation: number;
  /** Fraud risk assessment */
  fraudRisk: FraudRiskLevel;
  /** Optional: detailed breakdown data */
  details?: {
    /** Total number of executions analyzed */
    totalExecutions?: number;
    /** Number of failed executions */
    failedExecutions?: number;
    /** Average response time in ms */
    averageResponseTimeMs?: number;
    /** Number of community votes/ratings */
    voteCount?: number;
    /** Last updated timestamp */
    lastUpdated?: string;
  };
}

/**
 * Props for the TrustScoreBadge component
 */
export interface TrustScoreBadgeProps {
  /** Trust metrics data object */
  metrics: TrustMetrics;
  /** Display variant */
  variant?: 'compact' | 'expanded' | 'mini';
  /** Whether to show detailed tooltip on hover */
  showDetails?: boolean;
  /** Optional additional CSS classes */
  className?: string;
  /** Optional callback when badge is clicked */
  onClick?: () => void;
}

// State Fabric Types

export interface StateFabric {
  id: string;
  name: string;
  description: string;
  status: 'online' | 'offline' | 'degraded' | 'pending';
  type: 'session' | 'catalog' | 'cache' | 'workflow' | 'custom';
  tenantId: string;
  stores: StateFabricStore[];
  pipelines: Pipeline[];
  throughput: number;
  latency: number;
  lastUpdated: string;
  createdAt: string;
  updatedAt: string;
  settings: StateFabricSettings;
  metrics: StateFabricMetrics;
}

export interface StateFabricStore {
  id: string;
  name: string;
  type: 'memory' | 'persistent' | 'cache' | 'queue' | 'vector' | 'embedding' | 'ai-memory';
  status: 'active' | 'inactive' | 'error';
  size: number;
  maxSize: number;
  region: string;
  provider: string;
  throughput: number;
  latency: number;
  createdAt: string;
  updatedAt: string;
}

export interface Pipeline {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'paused' | 'error' | 'draft';
  steps: PipelineStep[];
  inputSchema?: Record<string, any>;
  outputSchema?: Record<string, any>;
  throughput: number;
  errorRate: number;
  lastExecutedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PipelineStep {
  id: string;
  name: string;
  type: 'transform' | 'filter' | 'aggregate' | 'enrich' | 'custom';
  config: Record<string, any>;
  order: number;
  enabled: boolean;
  timeout: number;
  retry_count: number;
  description?: string;
  condition?: string;
}

export interface EventLog {
  id: string;
  fabricId: string;
  storeId?: string;
  eventType: 'create' | 'update' | 'delete' | 'snapshot' | 'sync';
  payload: Record<string, any>;
  timestamp: string;
  sequenceNumber: number;
  correlationId?: string;
}

export interface Snapshot {
  id: string;
  fabricId: string;
  storeId?: string;
  name: string;
  description?: string;
  state: Record<string, any>;
  eventCount: number;
  sizeBytes: number;
  createdAt: string;
  createdBy?: string;
  expiresAt?: string;
}

export interface ReplaySession {
  id: string;
  fabricId: string;
  snapshotId?: string;
  startEventId?: string;
  endEventId?: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  eventsReplayed: number;
  startedAt: string;
  completedAt?: string;
  error?: string;
}

export interface StateFabricSettings {
  autoSnapshot: boolean;
  snapshotIntervalMinutes: number;
  retentionDays: number;
  enableReplication: boolean;
  regions: string[];
  conflictResolution: 'last-write-wins' | 'first-write-wins' | 'manual';
}

/** One data point for metrics time-series (from API when available). */
export interface StateFabricMetricsHistoryPoint {
  time: string;
  operations: number;
  latency: number;
  errors: number;
}

export interface StateFabricMetrics {
  totalOperations: number;
  operationsPerSecond: number;
  averageLatency: number;
  errorRate: number;
  cacheHitRate?: number;
  storageUsed: number;
  lastCalculatedAt: string;
  /** Time-series for charts; when omitted, charts show empty state. */
  history?: StateFabricMetricsHistoryPoint[];
}

export interface CreateStateFabricRequest {
  name: string;
  description: string;
  type: StateFabric['type'];
  settings?: Partial<StateFabricSettings>;
}

export interface UpdateStateFabricRequest {
  name?: string;
  description?: string;
  settings?: Partial<StateFabricSettings>;
}

export interface CreatePipelineRequest {
  name: string;
  description: string;
  steps: Omit<PipelineStep, 'id'>[];
}

export interface UpdatePipelineRequest {
  name?: string;
  description?: string;
  steps?: PipelineStep[];
  status?: Pipeline['status'];
}

export interface CreateStoreRequest {
  name: string;
  type: StateFabricStore['type'];
  maxSize: number;
  region: string;
}

// State Fabric Triggers
export interface StateFabricTrigger {
  id: string;
  tenantId: string;
  sourceStateId: string;
  triggerType: string;
  keyPattern?: string;
  condition?: Record<string, any>;
  targetFunctionId?: string;
  targetFunction?: string;
  includePrevious: boolean;
  includeNew: boolean;
  maxInvocationsPerMinute: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTriggerRequest {
  triggerType: string;
  keyPattern?: string;
  condition?: Record<string, any>;
  targetFunctionId?: string;
  targetFunction?: string;
  includePrevious: boolean;
  includeNew: boolean;
  maxInvocationsPerMinute: number;
  isActive: boolean;
}

// ============================================================================
// Enhanced Profile Page Types
// ============================================================================

/** Tab types for profile page navigation */
export type ProfileTab = 'overview' | 'functions' | 'activity' | 'analytics' | 'about' | 'settings';

/** Social links for user profile */
export interface SocialLinks {
  github?: string;
  twitter?: string;
  linkedin?: string;
  website?: string;
  discord?: string;
  devto?: string;
  medium?: string;
}

/** Achievement/Badge data */
export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  unlockedAt: string;
  tier: 'bronze' | 'silver' | 'gold' | 'platinum';
  progress?: {
    current: number;
    target: number;
  };
}

/** Activity feed item types */
export type ActivityType =
  | 'joined'
  | 'function_published'
  | 'function_updated'
  | 'function_deleted'
  | 'achievement_earned'
  | 'review_received'
  | 'milestone_reached'
  | 'followed'
  | 'follower_gained'
  | 'contribution'
  | 'deployment'
  | 'membership_upgraded';

/** Activity feed item */
export interface UserActivity {
  id: string;
  type: ActivityType;
  title: string;
  description?: string;
  timestamp: string;
  metadata?: Record<string, any>;
  relatedFunction?: {
    id: string;
    name: string;
    author: string;
  };
  relatedUser?: {
    id: string;
    username: string;
    avatar?: string;
  };
}

/** User statistics for profile page */
export interface UserStats {
  // Function stats
  functionsPublished: number;
  functionsTrend: number; // Percentage change
  totalExecutions: number;
  executionsTrend: number;
  totalViews: number;
  viewsTrend: number;

  // Reputation
  trustScore: number;
  reputationRank: string;

  // User Reputation Profile scores (from reputation_profiles table)
  builderScore?: number;
  optimizerScore?: number;
  mentorScore?: number;
  agentWhispererScore?: number;
  reputationTier?: string;
  overallReputationScore?: number;

  // Social
  followersCount: number;
  followingCount: number;
  followersTrend: number;

  // Financial (if applicable)
  totalEarnings?: number;
  earningsTrend?: number;
  currency?: string;

  // Contribution streak
  contributionStreak: {
    current: number;
    longest: number;
    lastContribution: string;
  };

  // Contribution graph data (GitHub-style heatmap)
  contributionGraph: {
    date: string;
    count: number;
    level: 0 | 1 | 2 | 3 | 4;
  }[];
}

/** Skill/Technology expertise */
export interface Skill {
  name: string;
  level: 'beginner' | 'intermediate' | 'advanced' | 'expert';
  category: 'language' | 'framework' | 'tool' | 'platform' | 'concept';
  endorsements?: number;
}

/** Extended user profile data */
export interface UserProfile {
  id: string;
  username: string;
  name: string;
  avatar?: string;
  coverImage?: string;
  bio?: string;
  location?: string;
  company?: string;
  jobTitle?: string;
  website?: string;
  socialLinks: SocialLinks;
  skills: Skill[];
  createdAt: string;
  updatedAt?: string;
  isOnline: boolean;
  lastActive?: string;
  // Profile number for early adopter tracking (e.g., Member #123)
  profileNumber?: number;
  // Platform admin role (super_admin, admin, support, etc.)
  role?: string;

  // Extended info for About tab
  experience?: {
    company: string;
    title: string;
    startDate: string;
    endDate?: string;
    current: boolean;
    description?: string;
  }[];
  education?: {
    institution: string;
    degree: string;
    field: string;
    startDate: string;
    endDate?: string;
  }[];
  openSourceContributions?: {
    project: string;
    url: string;
    contributions: number;
  }[];
  languages?: string[];

  // Stats and content
  stats: UserStats;
  achievements: Achievement[];
  recentActivity: UserActivity[];
  certifications: import('@/api/certification').PublicBadge[];
  publishedFunctions: FunctionCardData[];
}

/** Analytics data for profile */
export interface ProfileAnalytics {
  executionHistory: {
    date: string;
    executions: number;
    uniqueUsers: number;
  }[];
  popularFunctions: {
    functionId: string;
    name: string;
    executions: number;
    percentage: number;
  }[];
  revenueHistory?: {
    date: string;
    revenue: number;
    calls: number;
  }[];
  geographicDistribution: {
    country: string;
    executions: number;
    percentage: number;
  }[];
  deviceStats: {
    device: string;
    percentage: number;
  }[];
  browserStats: {
    browser: string;
    percentage: number;
  }[];
}

/** Filter and sort options for functions tab */
export interface FunctionFilters {
  search: string;
  sortBy: 'popular' | 'recent' | 'name' | 'rating';
  category?: string;
  language?: string;
  visibility?: 'all' | 'public' | 'private';
}

/** Trust metrics visualization data */
export interface TrustMetricsVisualization {
  overallScore: number;
  breakdown: {
    reliability: number;
    performance: number;
    security: number;
    community: number;
    documentation: number;
  };
  history: {
    date: string;
    score: number;
  }[];
}

// ============================================================================
// Agent Memory Types
// ============================================================================

/** Agent memory types */
export type AgentMemoryType = 'working' | 'longterm' | 'context' | 'episodic';

/** Agent memory model */
export interface AgentMemory {
  id: string;
  tenant_id: string;
  agent_id: string;
  memory_type: AgentMemoryType;
  content?: string;
  structured_data?: Record<string, unknown>;
  embedding?: number[];
  importance_score: number;
  access_count: number;
  last_accessed_at?: string;
  ttl_days: number;
  expires_at?: string;
  source_event_id?: string;
  created_at: string;
  updated_at: string;
}

/** Request to create an agent memory */
export interface CreateAgentMemoryRequest {
  agent_id: string;
  memory_type: AgentMemoryType;
  content: string;
  structured_data?: Record<string, unknown>;
  embedding?: number[];
  importance_score?: number;
  ttl_days?: number;
}

/** Request to update an agent memory */
export interface UpdateAgentMemoryRequest {
  content?: string;
  structured_data?: Record<string, unknown>;
  embedding?: number[];
  importance_score?: number;
}

/** Request to search agent memories */
export interface AgentMemorySearchRequest {
  agent_id?: string;
  query: string;
  embedding?: number[];
  memory_type?: AgentMemoryType;
  limit?: number;
  threshold?: number;
}

/** Response from searching agent memories */
export interface AgentMemorySearchResponse {
  memories: AgentMemory[];
  count: number;
}

/** Request to rebuild the search index */
export interface RebuildIndexRequest {
  agent_id?: string;
  memory_type?: AgentMemoryType;
}

/** Response from rebuilding the index */
export interface RebuildIndexResponse {
  success: boolean;
  index_id?: string;
  memory_count: number;
  message: string;
}

/** Response from listing agent memories */
export interface ListAgentMemoriesResponse {
  memories: AgentMemory[];
  total: number;
  limit: number;
  offset: number;
}

// ============================================================================
// Vault Types - Re-export from vault module
// ============================================================================

export * from './vault';

// ============================================================================
// API Key Types - Re-export from api-key module
// ============================================================================

export * from './api-key';

// ============================================================================
// GitHub Integration Types - Re-export from github module
// ============================================================================

export * from './github';

// ============================================================================
// Simple State Types
// ============================================================================

/** Simple State - basic key-value state management */
export interface SimpleState {
  path: string;
  key: string;
  value: StateValue;
  createdAt: string;
  updatedAt: string;
  version: number;
  ttl?: number;
  metadata?: Record<string, string>;
}

/** Primitive JSON-serializable state value */
export type StatePrimitive = string | number | boolean | null;

/** State value - JSON-serializable, non-recursive in type definition to satisfy TypeScript */
export type StateValue =
  | StatePrimitive
  | StatePrimitive[]
  | Record<string, StatePrimitive | StatePrimitive[]>;

/** History entry for state changes */
export interface StateHistoryEntry {
  id: string;
  path: string;
  key: string;
  operation: 'create' | 'update' | 'delete' | 'patch';
  previousValue?: StateValue;
  newValue?: StateValue;
  timestamp: string;
  actor?: string;
}

/** Snapshot of state at a point in time */
export interface StateSnapshot {
  id: string;
  path: string;
  name: string;
  description?: string;
  state: Record<string, StateValue>;
  version: number;
  createdAt: string;
  expiresAt?: string;
}

/** Permission for state access */
export interface StatePermission {
  id: string;
  path: string;
  principal: string;
  principalType: 'user' | 'team' | 'service';
  permissions: ('read' | 'write' | 'admin')[];
  grantedAt: string;
  grantedBy?: string;
}

/** Request to create a state */
export interface CreateStateRequest {
  path: string;
  key: string;
  value: StateValue;
  ttl?: number;
  metadata?: Record<string, string>;
}

/** Request to update a state */
export interface UpdateStateRequest {
  value: StateValue;
  ttl?: number;
  metadata?: Record<string, string>;
}

/** Request to patch a state value */
export interface PatchStateRequest {
  path: string;
  key?: string;
  operations: JsonPatchOperation[];
}

/** JSON Patch operation */
export interface JsonPatchOperation {
  op: 'add' | 'remove' | 'replace' | 'move' | 'copy';
  path: string;
  value?: StateValue;
  from?: string;
}

/** Request to create a snapshot */
export interface CreateSnapshotRequest {
  path: string;
  name: string;
  description?: string;
}

/** Request to grant permission */
export interface GrantPermissionRequest {
  path: string;
  principal: string;
  principalType: 'user' | 'team' | 'service';
  permissions: ('read' | 'write' | 'admin')[];
}

/** Time travel request */
export interface TimeTravelRequest {
  path: string;
  timestamp?: string;
  version?: number;
}

/** Time travel response */
export interface TimeTravelResponse {
  path: string;
  value: StateValue;
  version: number;
  timestamp: string;
}

// ============================================================================
// Tenant Auth Types for Backend-in-a-Box Bundles
// ============================================================================

/** Per-tenant authentication settings */
export interface TenantAuthSettings {
  id: string;
  tenant_id: string;
  mfa_required: boolean;
  mfa_mode: 'optional' | 'required' | 'enforced';
  password_policy: PasswordPolicy;
  session_timeout_minutes: number;
  ip_allowlist_enabled: boolean;
  ip_allowlist: string[];
  allowed_domains: string[];
  sso_provider: 'none' | 'saml' | 'oidc';
  saml_metadata_url?: string;
  saml_entity_id?: string;
  use_custom_branding: boolean;
  email_from_name: string;
  email_from_address: string;
  require_email_verification: boolean;
  allow_password_login: boolean;
  allow_magic_link: boolean;
  max_login_attempts: number;
  lockout_duration_minutes: number;
  created_at: string;
  updated_at: string;
}

/** Password policy requirements */
export interface PasswordPolicy {
  min_length: number;
  require_uppercase: boolean;
  require_lowercase: boolean;
  require_digit: boolean;
  require_special: boolean;
  max_age_days?: number;
  prevent_reuse_count?: number;
}

/** OAuth provider configuration for a tenant */
export interface TenantOAuthProvider {
  id: string;
  tenant_id: string;
  provider: 'github' | 'google' | 'microsoft' | 'apple' | 'gitlab' | 'bitbucket';
  client_id: string;
  enabled: boolean;
  callback_url?: string;
  scopes: string[];
  created_at: string;
  updated_at: string;
}

/** Team membership record */
export interface TenantMembership {
  id: string;
  tenant_id: string;
  user_id: string;
  role: 'team_owner' | 'team_admin' | 'team_member' | 'team_viewer';
  invited_by?: string;
  invited_at?: string;
  joined_at: string;
  last_active_at?: string;
  status: 'active' | 'suspended' | 'invited';
  user?: {
    id: string;
    email: string;
    name?: string;
    avatar?: string;
  };
}

/** Invitation for a user to join a tenant */
export interface TenantInviteCode {
  id: string;
  tenant_id: string;
  code: string;
  email: string;
  role: 'team_owner' | 'team_admin' | 'team_member' | 'team_viewer';
  invited_by: string;
  expires_at: string;
  accepted_at?: string;
  accepted_by?: string;
  max_uses: number;
  uses: number;
  created_at: string;
}

/** Auth audit log entry */
export interface TenantAuthAuditLog {
  id: string;
  tenant_id: string;
  user_id?: string;
  action: string;
  resource_type?: string;
  resource_id?: string;
  ip_address?: string;
  user_agent?: string;
  metadata?: Record<string, unknown>;
  success: boolean;
  error_message?: string;
  created_at: string;
}

// Function DNA types
export * from './dna';

/** User certification badge */
export interface UserCertification extends Omit<import('@/api/certification').PublicBadge, 'tier'> {
  id: string;
  tier_slug: string;
  tier_name: string;
  tier_color: string;
  tier_icon: string;
  credential_number: string;
  issued_at: string;
  expires_at: string;
  status: string;
  verification_hash: string;
  verification_url?: string;
}

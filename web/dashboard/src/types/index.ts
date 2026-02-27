export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string; // Profile picture URL from social providers
  tenantId: string;
  plan: "starter" | "pro";
  role?: string; // Admin role for admin users
  createdAt: string;
  updatedAt?: string;
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
  provider: string;
  region: string;
  url: string;
  sharedSecret: string;
  priority?: number;
  createdAt: string;
}

export interface CircuitState {
  state: "closed" | "open" | "half-open";
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
  status: "pending" | "building" | "deploying" | "success" | "failed" | "rolled_back";
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
  email: string;
  password: string;
  recaptchaToken?: string;
}

export interface SignupRequest {
  name?: string;
  email: string;
  password: string;
  confirmPassword: string;
  termsAccepted: boolean;
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
  id: string;
  name: string;
  status: "online" | "offline" | "degraded" | "pending";
  connectedAt: string;
  apiKey?: string;
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
  status: "active" | "suspended";
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
  status: "active" | "canceled" | "past_due" | "trialing";
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
  status: "draft" | "open" | "paid" | "void" | "uncollectible";
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
  discountType: "percent" | "amount";
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
  status: "healthy" | "unhealthy";
  version: string;
  timestamp: string;
  checks: Record<string, {
    status: string;
    healthy: boolean;
    responseTimeMs?: number;
    goroutines?: number;
  }>;
}

// Analytics types
export interface AnalyticsService {
  name: string;
  enabled: boolean;
  status: "loading" | "loaded" | "error" | "disabled";
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
  createdAt: string;
  updatedAt: string;
  version: string;
  status: "draft" | "deploying" | "deployed" | "failed";
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
  status: "pending" | "deploying" | "success" | "failed";
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
  level: "info" | "warn" | "error" | "debug";
  message: string;
  timestamp: string;
  source: string;
  metadata?: Record<string, any>;
}

export interface DeployFunctionRequest {
  functionId: string;
  providers?: string[];
  region?: string;
}

export interface DeployFunctionResponse {
  deploymentId: string;
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

import { z } from 'zod';

// Base schemas for common patterns
const timestampSchema = z.string().refine((val) => !isNaN(Date.parse(val)), {
  message: 'Invalid timestamp format'
});

const idSchema = z.string().uuid('Invalid UUID format');

const emailSchema = z.string().email('Invalid email format');

const urlSchema = z.string().url('Invalid URL format');

// Database monitoring schemas
export const databaseHealthSchema = z.object({
  status: z.enum(['healthy', 'degraded', 'unhealthy']),
  connections: z.object({
    active: z.number().int().min(0),
    idle: z.number().int().min(0),
    total: z.number().int().min(0),
    max: z.number().int().min(1)
  }),
  performance: z.object({
    avgQueryTime: z.number().min(0),
    slowQueries: z.number().int().min(0),
    throughput: z.number().min(0)
  }),
  storage: z.object({
    used: z.number().min(0),
    total: z.number().min(0),
    growthRate: z.number()
  }),
  replication: z.object({
    lag: z.number().min(0),
    status: z.enum(['healthy', 'lagging', 'error'])
  }),
  lastUpdated: timestampSchema
});

export const databaseAlertSchema = z.object({
  id: z.string(),
  type: z.enum(['connection_pool_exhausted', 'high_query_latency', 'storage_warning', 'replication_lag']),
  severity: z.enum(['low', 'medium', 'high', 'critical']),
  title: z.string().min(1),
  message: z.string(),
  timestamp: timestampSchema,
  resolved: z.boolean().optional()
});

export const databaseMetricSchema = z.object({
  timestamp: timestampSchema,
  connections: z.number().int().min(0),
  queryCount: z.number().int().min(0),
  avgResponseTime: z.number().min(0),
  errorRate: z.number().min(0)
});

// User and authentication schemas
export const userSchema = z.object({
  id: idSchema,
  email: emailSchema,
  name: z.string().min(1),
  avatar: z.string().url().optional(),
  tenantId: idSchema,
  plan: z.enum(['starter', 'pro']),
  role: z.string().optional(),
  createdAt: timestampSchema,
  updatedAt: timestampSchema.optional()
});

export const loginRequestSchema = z.object({
  email: emailSchema,
  password: z.string().min(1),
  recaptchaToken: z.string().optional()
});

export const signupRequestSchema = z.object({
  name: z.string().optional(),
  email: emailSchema,
  password: z.string().min(1),
  confirmPassword: z.string().min(1),
  termsAccepted: z.boolean(),
  recaptchaToken: z.string().optional()
});

export const loginResponseSchema = z.object({
  token: z.string().min(1),
  expiresIn: z.number().int().positive(),
  user: userSchema
});

export const signupResponseSchema = z.object({
  message: z.string(),
  emailSent: z.boolean(),
  requiresVerification: z.boolean()
});

// App and backend schemas
export const appSchema = z.object({
  id: idSchema,
  name: z.string().min(1),
  slug: z.string().regex(/^[a-z0-9-]+$/, 'Slug must contain only lowercase letters, numbers, and hyphens'),
  tenantId: idSchema,
  createdAt: timestampSchema
});

export const backendSchema = z.object({
  id: idSchema,
  provider: z.string().min(1),
  region: z.string().min(1),
  url: urlSchema,
  sharedSecret: z.string().min(1),
  priority: z.number().int().min(0).optional(),
  createdAt: timestampSchema
});

export const circuitStateSchema = z.object({
  state: z.enum(['closed', 'open', 'half-open']),
  sinceTs: timestampSchema,
  failCount: z.number().int().min(0),
  successCount: z.number().int().min(0),
  lastFailureTs: timestampSchema.optional()
});

export const healthCheckSchema = z.object({
  timestamp: timestampSchema,
  ok: z.boolean(),
  statusCode: z.number().int().min(100).max(599),
  latencyMs: z.number().min(0),
  errorMessage: z.string().optional()
});

export const backendStatusSchema = z.object({
  backend: backendSchema,
  circuitState: circuitStateSchema.optional(),
  latestHealthCheck: healthCheckSchema.optional()
});

export const appStatusSchema = z.object({
  app: appSchema,
  backends: z.array(backendStatusSchema)
});

// Deployment schemas
export const deploymentSchema = z.object({
  id: idSchema,
  appId: idSchema,
  provider: z.string().min(1),
  region: z.string().min(1),
  status: z.enum(['pending', 'building', 'deploying', 'success', 'failed', 'rolled_back']),
  artifactUrl: urlSchema.optional(),
  deployedUrl: urlSchema.optional(),
  errorMessage: z.string().optional(),
  createdAt: timestampSchema,
  updatedAt: timestampSchema
});

export const routingDecisionSchema = z.object({
  selectedBackend: backendSchema.optional(),
  failoverBackends: z.array(backendSchema),
  reason: z.string().min(1),
  requestId: z.string().min(1)
});

// Function schemas
export const environmentVariableSchema = z.object({
  key: z.string().regex(/^[A-Z_][A-Z0-9_]*$/, 'Environment variable key must be uppercase with underscores'),
  value: z.string(),
  isSecret: z.boolean()
});

export const functionConfigSchema = z.object({
  id: idSchema,
  name: z.string().min(1).max(100),
  providers: z.array(z.string().min(1)),
  region: z.string().min(1),
  code: z.string().min(1),
  envVars: z.array(environmentVariableSchema),
  tenantId: idSchema,
  createdAt: timestampSchema,
  updatedAt: timestampSchema,
  version: z.string().min(1),
  status: z.enum(['draft', 'deploying', 'deployed', 'failed'])
});

export const createFunctionRequestSchema = z.object({
  name: z.string().min(1).max(100),
  providers: z.array(z.string().min(1)).min(1),
  region: z.string().min(1),
  code: z.string().min(1),
  envVars: z.array(environmentVariableSchema).optional()
});

export const updateFunctionRequestSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  providers: z.array(z.string().min(1)).min(1).optional(),
  region: z.string().min(1).optional(),
  code: z.string().min(1).optional(),
  envVars: z.array(environmentVariableSchema).optional()
});

export const functionDeploymentSchema = z.object({
  id: idSchema,
  functionId: idSchema,
  version: z.string().min(1),
  status: z.enum(['pending', 'deploying', 'success', 'failed']),
  provider: z.string().min(1),
  region: z.string().min(1),
  deployedUrl: urlSchema.optional(),
  errorMessage: z.string().optional(),
  createdAt: timestampSchema,
  updatedAt: timestampSchema
});

export const functionLogSchema = z.object({
  id: idSchema,
  functionId: idSchema,
  deploymentId: idSchema.optional(),
  level: z.enum(['info', 'warn', 'error', 'debug']),
  message: z.string(),
  timestamp: timestampSchema,
  source: z.string().min(1),
  metadata: z.record(z.any()).optional()
});

export const deployFunctionRequestSchema = z.object({
  functionId: idSchema,
  providers: z.array(z.string().min(1)).optional(),
  region: z.string().min(1).optional()
});

export const deployFunctionResponseSchema = z.object({
  deploymentId: idSchema,
  status: z.string().min(1),
  deployments: z.array(functionDeploymentSchema)
});

export const testFunctionRequestSchema = z.object({
  functionId: idSchema.optional(),
  code: z.string().optional(),
  envVars: z.array(environmentVariableSchema).optional(),
  testInput: z.any().optional()
});

export const testFunctionResponseSchema = z.object({
  success: z.boolean(),
  output: z.any().optional(),
  error: z.string().optional(),
  executionTimeMs: z.number().min(0),
  logs: z.array(functionLogSchema)
});

// Admin schemas
export const tenantSchema = z.object({
  id: idSchema,
  name: z.string().min(1),
  plan: z.string().optional(),
  status: z.enum(['active', 'suspended']),
  createdAt: timestampSchema,
  updatedAt: timestampSchema
});

export const auditEventSchema = z.object({
  id: idSchema,
  actorUserId: idSchema.optional(),
  actorEmail: emailSchema.optional(),
  tenantId: idSchema.optional(),
  action: z.string().min(1),
  resourceType: z.string().min(1),
  resourceId: idSchema.optional(),
  requestId: z.string().optional(),
  beforeState: z.any().optional(),
  afterState: z.any().optional(),
  ipAddress: z.string().optional(),
  userAgent: z.string().optional(),
  timestamp: timestampSchema,
  success: z.boolean()
});

// Analytics schemas
export const analyticsServiceSchema = z.object({
  name: z.string().min(1),
  enabled: z.boolean(),
  status: z.enum(['loading', 'loaded', 'error', 'disabled']),
  config: z.record(z.any()),
  lastUsed: timestampSchema.optional()
});

export const googleAnalyticsConfigSchema = z.object({
  measurementId: z.string().optional(),
  enabled: z.boolean()
});

export const hotjarConfigSchema = z.object({
  siteId: z.string().optional(),
  enabled: z.boolean()
});

export const analyticsSettingsSchema = z.object({
  googleAnalytics: googleAnalyticsConfigSchema.optional(),
  hotjar: hotjarConfigSchema.optional(),
  services: z.array(analyticsServiceSchema)
});

export const updateAnalyticsRequestSchema = z.object({
  googleAnalytics: googleAnalyticsConfigSchema.optional(),
  hotjar: hotjarConfigSchema.optional()
});

export const updateAnalyticsResponseSchema = z.object({
  message: z.string().optional(),
  settings: updateAnalyticsRequestSchema.optional(),
  note: z.string().optional()
});

// Real-time event schemas
export const realtimeEventSchema = z.object({
  type: z.string().min(1),
  table: z.string().min(1),
  record_id: idSchema.optional(),
  tenant_id: idSchema.optional(),
  user_id: idSchema.optional(),
  data: z.any().optional(),
  timestamp: timestampSchema
});

export const userStatusChangeEventSchema = realtimeEventSchema.extend({
  type: z.literal('user_status_change'),
  old_status: z.enum(['verified', 'unverified']),
  new_status: z.enum(['verified', 'unverified'])
});

export const profileUpdateEventSchema = realtimeEventSchema.extend({
  type: z.literal('profile_update'),
  changes: z.object({
    first_name: z.boolean().optional(),
    last_name: z.boolean().optional(),
    avatar_url: z.boolean().optional(),
    bio: z.boolean().optional()
  })
});

export const newNotificationEventSchema = realtimeEventSchema.extend({
  type: z.literal('new_notification'),
  notification_id: idSchema,
  notification_type: z.enum(['info', 'warning', 'error', 'success']),
  title: z.string().min(1)
});

export const presenceEventSchema = realtimeEventSchema.extend({
  type: z.enum(['presence_join', 'presence_leave']),
  key: z.string().min(1),
  current_presences: z.array(z.any()),
  new_presences: z.array(z.any()).optional(),
  left_presences: z.array(z.any()).optional()
});

// Hook input validation schemas
export const timeRangeSchema = z.enum(['1h', '6h', '24h', '7d']);

export const tableNameSchema = z.string().regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, 'Invalid table name format');

// Type exports for TypeScript inference
export type DatabaseHealth = z.infer<typeof databaseHealthSchema>;
export type DatabaseAlert = z.infer<typeof databaseAlertSchema>;
export type DatabaseMetric = z.infer<typeof databaseMetricSchema>;
export type User = z.infer<typeof userSchema>;
export type LoginRequest = z.infer<typeof loginRequestSchema>;
export type SignupRequest = z.infer<typeof signupRequestSchema>;
export type LoginResponse = z.infer<typeof loginResponseSchema>;
export type SignupResponse = z.infer<typeof signupResponseSchema>;
export type App = z.infer<typeof appSchema>;
export type Backend = z.infer<typeof backendSchema>;
export type CircuitState = z.infer<typeof circuitStateSchema>;
export type HealthCheck = z.infer<typeof healthCheckSchema>;
export type BackendStatus = z.infer<typeof backendStatusSchema>;
export type AppStatus = z.infer<typeof appStatusSchema>;
export type Deployment = z.infer<typeof deploymentSchema>;
export type RoutingDecision = z.infer<typeof routingDecisionSchema>;
export type EnvironmentVariable = z.infer<typeof environmentVariableSchema>;
export type FunctionConfig = z.infer<typeof functionConfigSchema>;
export type CreateFunctionRequest = z.infer<typeof createFunctionRequestSchema>;
export type UpdateFunctionRequest = z.infer<typeof updateFunctionRequestSchema>;
export type FunctionDeployment = z.infer<typeof functionDeploymentSchema>;
export type FunctionLog = z.infer<typeof functionLogSchema>;
export type DeployFunctionRequest = z.infer<typeof deployFunctionRequestSchema>;
export type DeployFunctionResponse = z.infer<typeof deployFunctionResponseSchema>;
export type TestFunctionRequest = z.infer<typeof testFunctionRequestSchema>;
export type TestFunctionResponse = z.infer<typeof testFunctionResponseSchema>;
export type Tenant = z.infer<typeof tenantSchema>;
export type AuditEvent = z.infer<typeof auditEventSchema>;
export type AnalyticsService = z.infer<typeof analyticsServiceSchema>;
export type GoogleAnalyticsConfig = z.infer<typeof googleAnalyticsConfigSchema>;
export type HotjarConfig = z.infer<typeof hotjarConfigSchema>;
export type AnalyticsSettings = z.infer<typeof analyticsSettingsSchema>;
export type UpdateAnalyticsRequest = z.infer<typeof updateAnalyticsRequestSchema>;
export type UpdateAnalyticsResponse = z.infer<typeof updateAnalyticsResponseSchema>;
export type RealtimeEvent = z.infer<typeof realtimeEventSchema>;
export type UserStatusChangeEvent = z.infer<typeof userStatusChangeEventSchema>;
export type ProfileUpdateEvent = z.infer<typeof profileUpdateEventSchema>;
export type NewNotificationEvent = z.infer<typeof newNotificationEventSchema>;
export type PresenceEvent = z.infer<typeof presenceEventSchema>;
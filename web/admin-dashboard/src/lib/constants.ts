/**
 * Constants for Admin Dashboard
 */

export const ROUTES = {
  // Auth
  AUTH_LOGIN: '/auth/login',
  AUTH_MFA: '/auth/mfa',
  AUTH_LOGOUT: '/auth/logout',

  // Admin Dashboard
  ADMIN_DASHBOARD: '/',
  ADMIN_TENANTS: '/tenants',
  ADMIN_USERS: '/users',
  ADMIN_BILLING: '/billing',
  ADMIN_AUDIT: '/audit',
  ADMIN_SYSTEM: '/system',
  ADMIN_BACKENDS: '/backends',
  ADMIN_PROVIDERS: '/providers',
  ADMIN_CONTENT: '/content',
  ADMIN_BLOG: '/blog',
  ADMIN_FUNCTIONS: '/functions',
  ADMIN_STATE_FABRIC: '/state-fabric',
  ADMIN_FEEDBACK: '/feedback',
  ADMIN_FEATURES: '/features',
  ADMIN_STATUS: '/status',
  ADMIN_STATUS_INCIDENTS: '/status/incidents',
  ADMIN_TRUST_DASHBOARD: '/trust-dashboard',
  ADMIN_EXECUTION_AUDIT: '/execution-audit',
  ADMIN_FRAUD_DETECTION: '/fraud-detection',
  ADMIN_ECONOMIC_LEADERBOARD: '/economic-leaderboard',
  ADMIN_REDIRECTS: '/redirects',
  ADMIN_EMAIL: '/email',
  ADMIN_NEWSLETTER: '/newsletter',
  ADMIN_CONTENT_CALENDAR: '/content-calendar',
  ADMIN_FACTORY: '/factory',
  ADMIN_FACTORY_OPPORTUNITIES: '/factory/opportunities',
  ADMIN_FACTORY_REVIEWS: '/factory/reviews',
  ADMIN_MAINTENANCE: '/maintenance',
  ADMIN_CACHE: '/cache',
  ADMIN_MONITORING: '/monitoring',
  ADMIN_CLOUDFLARE_ANALYTICS: '/cloudflare',
  ADMIN_SUPPORT: '/support',
  ADMIN_IP_ALLOWLIST: '/ip-allowlist',
  ADMIN_SIEM: '/siem',
  ADMIN_AUTH_AUDIT: '/auth-audit',
  ADMIN_SIGNUP_INVITES: '/signup-invites',
  ADMIN_WAITLIST: '/waitlist',
};

export const API_ROUTES = {
  ADMIN_SESSION: '/auth/session',
  ADMIN_HEALTH: '/health',
  ADMIN_DASHBOARD_ACTIVITY: '/dashboard/activity',
  ADMIN_DASHBOARD_REVENUE: '/dashboard/revenue',
  ADMIN_DASHBOARD_QUICK_STATS: '/dashboard/quick-stats',
  ADMIN_TENANTS: '/tenants',
  ADMIN_USERS: '/users',
  ADMIN_AUDIT_EVENTS: '/audit-events',
  // Maintenance
  ADMIN_MAINTENANCE: '/maintenance',
  // Cache
  ADMIN_CACHE_STATS: '/cache/stats',
  // Monitoring
  ADMIN_MONITORING_ALERTS: '/monitoring/alerts',
  ADMIN_MONITORING_METRICS: '/monitoring/metrics',
  ADMIN_MONITORING_HEALTH: '/monitoring/health',
  // Cloudflare Analytics
  ADMIN_CLOUDFLARE_ANALYTICS: '/cloudflare/analytics',
  // Factory
  FACTORY_STATUS: '/factory/status',
  FACTORY_CONFIG: '/factory/config',
  FACTORY_PIPELINE_RUN: '/factory/pipeline/run',
  FACTORY_REVIEWS_PENDING: '/factory/reviews/pending',
  FACTORY_OPPORTUNITIES: '/factory/opportunities',
  FACTORY_FUNCTIONS: '/factory/functions',
  FACTORY_VERSION_CODE: '/factory/versions',
} as const;

export type APIRoute = (typeof API_ROUTES)[keyof typeof API_ROUTES];

export function factoryRoute(opportunityId?: string, action?: 'approve' | 'reject'): string {
  if (opportunityId && action) {
    return `/factory/opportunities/${opportunityId}/${action}`;
  }
  if (opportunityId) {
    return `/factory/opportunities/${opportunityId}`;
  }
  return API_ROUTES.FACTORY_OPPORTUNITIES;
}

/**
 * Canonical auth audit event types returned by GET /v1/admin/auth-audit.
 * The list is what the backend actually emits — the filter dropdown and
 * the "event type" column use this exact set.
 */
export const AUDIT_EVENT_TYPES = [
  { value: 'login', label: 'Login' },
  { value: 'logout', label: 'Logout' },
  { value: 'mfa_enable', label: 'MFA enabled' },
  { value: 'mfa_disable', label: 'MFA disabled' },
  { value: 'mfa_challenge', label: 'MFA challenge' },
  { value: 'mfa_failed', label: 'MFA failed' },
  { value: 'sso_login', label: 'SSO login' },
  { value: 'password_change', label: 'Password change' },
  { value: 'password_reset', label: 'Password reset' },
  { value: 'session_revoke', label: 'Session revoked' },
  { value: 'api_key_create', label: 'API key created' },
  { value: 'api_key_revoke', label: 'API key revoked' },
  { value: 'ip_allowlist_create', label: 'IP allowlist entry added' },
  { value: 'ip_allowlist_update', label: 'IP allowlist entry updated' },
  { value: 'ip_allowlist_delete', label: 'IP allowlist entry removed' },
  { value: 'siem_config_create', label: 'SIEM config created' },
  { value: 'siem_config_update', label: 'SIEM config updated' },
  { value: 'passkey_register', label: 'Passkey registered' },
  { value: 'passkey_delete', label: 'Passkey deleted' },
  { value: 'tenant_suspend', label: 'Tenant suspended' },
  { value: 'user_create', label: 'User created' },
  { value: 'user_update', label: 'User updated' },
  { value: 'user_delete', label: 'User deleted' },
] as const;

export type AuditEventType = (typeof AUDIT_EVENT_TYPES)[number]['value'];

/**
 * MFA policy options exposed in the tenant settings UI. The string values
 * are what the backend stores on tenant.mfa_policy; the labels are
 * display-only.
 */
export const MFA_POLICY_OPTIONS = [
  { value: 'disabled', label: 'Disabled' },
  { value: 'optional', label: 'Optional' },
  { value: 'required_admins', label: 'Required for admins' },
  { value: 'required_all', label: 'Required for everyone' },
] as const;

export type MFAPolicyValue = (typeof MFA_POLICY_OPTIONS)[number]['value'];

/**
 * Default session policy applied to new tenants. Matches what the
 * backend's defaults look like — keep these in lockstep with
 * internal/storage/sql/sessions.go's createSession defaults.
 */
export const SESSION_POLICY_DEFAULTS = {
  idle_timeout_minutes: 30,
  absolute_timeout_minutes: 24 * 60,
  mfa_reverify_interval_minutes: 4 * 60,
  max_concurrent_sessions: 5,
  ip_binding: 'none',
} as const;

/**
 * SIEM destination types the admin SIEM integration supports. Kept in
 * lockstep with internal/adapters/siem/sink.go's destination schemas.
 */
export const SIEM_DESTINATION_TYPES = [
  { value: 'webhook', label: 'Generic webhook' },
  { value: 'splunk_hec', label: 'Splunk HEC' },
  { value: 'datadog_logs', label: 'Datadog Logs' },
  { value: 's3', label: 'Amazon S3' },
  { value: 'gcs', label: 'Google Cloud Storage' },
] as const;

export type SIEMDestinationType = (typeof SIEM_DESTINATION_TYPES)[number]['value'];

/**
 * SIEM event payload formats supported by the admin SIEM integration.
 * Values are what the backend's SIEM adapter recognises.
 */
export const SIEM_FORMATS = [
  { value: 'json', label: 'JSON (one event per line)' },
  { value: 'cef', label: 'CEF (Common Event Format)' },
  { value: 'leef', label: 'LEEF (Log Event Extended Format)' },
  { value: 'syslog', label: 'Syslog RFC 5424' },
] as const;

export type SIEMFormatValue = (typeof SIEM_FORMATS)[number]['value'];

export const SESSION = {
  TIMEOUT: parseInt(import.meta.env.VITE_SESSION_TIMEOUT || '1800000', 10),
  IDLE_TIMEOUT: parseInt(import.meta.env.VITE_IDLE_TIMEOUT || '900000', 10),
  MFA_REVERIFY_INTERVAL: parseInt(import.meta.env.VITE_MFA_REVERIFY_INTERVAL || '14400000', 10),
  CHECK_INTERVAL: 60000, // 1 minute
};

export const SECURITY = {
  ENABLE_IP_WHITELIST: import.meta.env.VITE_ENABLE_IP_WHITELIST === 'true',
  ENABLE_DEVICE_FINGERPRINT: import.meta.env.VITE_ENABLE_DEVICE_FINGERPRINT === 'true',
  ENABLE_AUDIT_LOGGING: import.meta.env.VITE_ENABLE_AUDIT_LOGGING === 'true',
  ENABLE_SESSION_RECORDING: import.meta.env.VITE_ENABLE_SESSION_RECORDING === 'true',
};

export const CACHE_KEYS = {
  ADMIN_USER: 'admin_user',
  ADMIN_SESSION: 'admin_session',
  ADMIN_ACCESS_TOKEN: 'admin_access_token',
  ADMIN_PERMISSIONS: 'admin_permissions',
  BLOG_SETTINGS: 'admin_blog_settings',
  PLATFORM_SETTINGS: 'admin_platform_settings',
};

export const ERROR_MESSAGES = {
  UNAUTHORIZED: 'Unauthorized access',
  SESSION_EXPIRED: 'Your session has expired',
  IP_NOT_WHITELISTED: 'Your IP address is not whitelisted',
  MFA_REQUIRED: 'MFA verification required',
  PERMISSION_DENIED: 'You do not have permission to perform this action',
  INTERNAL_ERROR: 'An internal error occurred',
};

export function getApiBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL || 'https://api.functionfly.com';
}

export function getAdminApiBaseUrl(): string {
  return import.meta.env.VITE_ADMIN_API_BASE_URL || `${getApiBaseUrl()}/v1/admin`;
}

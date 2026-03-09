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
  ADMIN_FUNCTIONS: '/functions',
  ADMIN_REGISTRY: '/registry',
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
  ADMIN_NEWSLETTER: '/newsletter',
  ADMIN_CONTENT_CALENDAR: '/content-calendar',
  ADMIN_FACTORY: '/factory',
};

export const API_ROUTES = {
  ADMIN_SESSION: '/admin/auth/session',
  ADMIN_HEALTH: '/admin/health',
  ADMIN_DASHBOARD_ACTIVITY: '/admin/dashboard/activity',
  ADMIN_DASHBOARD_REVENUE: '/admin/dashboard/revenue',
  ADMIN_DASHBOARD_QUICK_STATS: '/admin/dashboard/quick-stats',
  ADMIN_TENANTS: '/admin/tenants',
  ADMIN_USERS: '/admin/users',
  ADMIN_AUDIT_EVENTS: '/admin/audit-events',
};

export const SESSION = {
  TIMEOUT: parseInt(import.meta.env.VITE_SESSION_TIMEOUT || '1800000', 10),
  IDLE_TIMEOUT: parseInt(
    import.meta.env.VITE_IDLE_TIMEOUT || '900000',
    10
  ),
  MFA_REVERIFY_INTERVAL: parseInt(
    import.meta.env.VITE_MFA_REVERIFY_INTERVAL || '14400000',
    10
  ),
  CHECK_INTERVAL: 60000, // 1 minute
};

export const SECURITY = {
  ENABLE_IP_WHITELIST:
    import.meta.env.VITE_ENABLE_IP_WHITELIST === 'true',
  ENABLE_DEVICE_FINGERPRINT:
    import.meta.env.VITE_ENABLE_DEVICE_FINGERPRINT === 'true',
  ENABLE_AUDIT_LOGGING:
    import.meta.env.VITE_ENABLE_AUDIT_LOGGING === 'true',
  ENABLE_SESSION_RECORDING:
    import.meta.env.VITE_ENABLE_SESSION_RECORDING === 'true',
};

export const CACHE_KEYS = {
  ADMIN_USER: 'admin_user',
  ADMIN_SESSION: 'admin_session',
  ADMIN_ACCESS_TOKEN: 'admin_access_token',
  ADMIN_PERMISSIONS: 'admin_permissions',
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
  return import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
}

export function getAdminApiBaseUrl(): string {
  return import.meta.env.VITE_ADMIN_API_BASE_URL || `${getApiBaseUrl()}/v1/admin`;
}

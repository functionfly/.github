import type { Permission } from '@/hooks/useAccessControl';

export interface AdminRouteDef {
  path: string;
  permission?: Permission;
  featureName?: string;
}

/** Route-level permissions aligned with backend JWT permission strings (mapped in useAccessControl). */
export const ADMIN_ROUTE_PERMISSIONS: Record<string, AdminRouteDef> = {
  '/': { path: '/', permission: 'system:read', featureName: 'Dashboard' },
  '/tenants': { path: '/tenants', permission: 'tenants:read', featureName: 'Tenants' },
  '/tenants/:tenantId': { path: '/tenants/:tenantId', permission: 'tenants:read', featureName: 'Tenant details' },
  '/tenants/:tenantId/settings': {
    path: '/tenants/:tenantId/settings',
    permission: 'tenants:write',
    featureName: 'Tenant settings',
  },
  '/users': { path: '/users', permission: 'users:read', featureName: 'Users' },
  '/users/:userId': { path: '/users/:userId', permission: 'users:read', featureName: 'User details' },
  '/billing': { path: '/billing', permission: 'billing:read', featureName: 'Billing' },
  '/audit': { path: '/audit', permission: 'audit:read', featureName: 'Audit log' },
  '/auth-audit': { path: '/auth-audit', permission: 'audit:read', featureName: 'Auth audit' },
  '/system': { path: '/system', permission: 'system:read', featureName: 'System' },
  '/backends': { path: '/backends', permission: 'system:read', featureName: 'Backends' },
  '/providers': { path: '/providers', permission: 'system:read', featureName: 'Providers' },
  '/functions': { path: '/functions', permission: 'functions:read', featureName: 'Functions' },
  '/functions/:functionId': {
    path: '/functions/:functionId',
    permission: 'functions:read',
    featureName: 'Function details',
  },
  '/content': { path: '/content', permission: 'system:read', featureName: 'Content' },
  '/blog': { path: '/blog', permission: 'system:read', featureName: 'Blog' },
  '/state-fabric': { path: '/state-fabric', permission: 'tenants:read', featureName: 'State Fabric' },
  '/feedback': { path: '/feedback', permission: 'system:read', featureName: 'Feedback' },
  '/features': { path: '/features', permission: 'features:read', featureName: 'Feature flags' },
  '/status': { path: '/status', permission: 'system:read', featureName: 'Status page' },
  '/status/incidents': { path: '/status/incidents', permission: 'system:read', featureName: 'Incidents' },
  '/redirects': { path: '/redirects', permission: 'system:read', featureName: 'Redirects' },
  '/email': { path: '/email', permission: 'system:read', featureName: 'Email' },
  '/newsletter': { path: '/newsletter', permission: 'system:read', featureName: 'Newsletter' },
  '/content-calendar': { path: '/content-calendar', permission: 'system:read', featureName: 'Content calendar' },
  '/trust-dashboard': { path: '/trust-dashboard', permission: 'system:read', featureName: 'Trust dashboard' },
  '/execution-audit': { path: '/execution-audit', permission: 'audit:read', featureName: 'Execution audit' },
  '/fraud-detection': { path: '/fraud-detection', permission: 'system:read', featureName: 'Fraud detection' },
  '/economic-leaderboard': {
    path: '/economic-leaderboard',
    permission: 'system:read',
    featureName: 'Economic leaderboard',
  },
  '/factory': { path: '/factory', permission: 'system:read', featureName: 'Function factory' },
  '/maintenance': { path: '/maintenance', permission: 'system:write', featureName: 'Maintenance' },
  '/cache': { path: '/cache', permission: 'system:write', featureName: 'Cache' },
  '/monitoring': { path: '/monitoring', permission: 'system:read', featureName: 'Monitoring' },
  '/cloudflare': { path: '/cloudflare', permission: 'system:read', featureName: 'Cloudflare analytics' },
  '/support': { path: '/support', permission: 'system:read', featureName: 'Support' },
  '/signup-invites': { path: '/signup-invites', permission: 'users:write', featureName: 'Signup invites' },
  '/waitlist': { path: '/waitlist', permission: 'users:read', featureName: 'Waitlist' },
  '/ip-allowlist': { path: '/ip-allowlist', permission: 'system:write', featureName: 'IP allowlist' },
  '/siem': { path: '/siem', permission: 'system:write', featureName: 'SIEM' },
};

export function getRoutePermission(path: string): AdminRouteDef | undefined {
  if (ADMIN_ROUTE_PERMISSIONS[path]) {
    return ADMIN_ROUTE_PERMISSIONS[path];
  }
  const dynamic = Object.values(ADMIN_ROUTE_PERMISSIONS).find((def) => {
    const pattern = def.path.replace(/:[^/]+/g, '[^/]+');
    return new RegExp(`^${pattern}$`).test(path);
  });
  return dynamic;
}

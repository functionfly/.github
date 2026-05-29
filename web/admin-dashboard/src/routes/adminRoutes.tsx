/**
 * Admin dashboard route definitions with lazy-loaded pages and permissions.
 */

import { AdminPage } from '@/components/auth/AdminPage';
import type { Permission } from '@/hooks/useAccessControl';
import { lazy, type ReactNode } from 'react';
import { Route } from 'react-router-dom';

const AdminDashboardPage = lazy(() =>
  import('@/pages/AdminDashboardPage').then((m) => ({ default: m.AdminDashboardPage }))
);
const AdminTenantsPage = lazy(() =>
  import('@/pages/AdminTenantsPage').then((m) => ({ default: m.AdminTenantsPage }))
);
const AdminTenantDetailPage = lazy(() =>
  import('@/pages/AdminTenantDetailPage').then((m) => ({ default: m.AdminTenantDetailPage }))
);
const AdminTenantSettingsPage = lazy(() =>
  import('@/pages/AdminTenantSettingsPage').then((m) => ({ default: m.AdminTenantSettingsPage }))
);
const AdminUsersPage = lazy(() =>
  import('@/pages/AdminUsersPage').then((m) => ({ default: m.AdminUsersPage }))
);
const AdminUserDetailPage = lazy(() =>
  import('@/pages/AdminUserDetailPage').then((m) => ({ default: m.AdminUserDetailPage }))
);
const AdminBillingPage = lazy(() =>
  import('@/pages/AdminBillingPage').then((m) => ({ default: m.AdminBillingPage }))
);
const AdminAuditPage = lazy(() =>
  import('@/pages/AdminAuditPage').then((m) => ({ default: m.AdminAuditPage }))
);
const AdminAuthAuditPage = lazy(() =>
  import('@/pages/AdminAuthAuditPage').then((m) => ({ default: m.AdminAuthAuditPage }))
);
const AdminSystemPage = lazy(() =>
  import('@/pages/AdminSystemPage').then((m) => ({ default: m.AdminSystemPage }))
);
const AdminBackendsPage = lazy(() =>
  import('@/pages/AdminBackendsPage').then((m) => ({ default: m.AdminBackendsPage }))
);
const AdminProvidersPage = lazy(() =>
  import('@/pages/AdminProvidersPage').then((m) => ({ default: m.AdminProvidersPage }))
);
const AdminFunctionsPage = lazy(() =>
  import('@/pages/AdminFunctionsPage').then((m) => ({ default: m.AdminFunctionsPage }))
);
const AdminFunctionDetailPage = lazy(() =>
  import('@/pages/AdminFunctionDetailPage').then((m) => ({ default: m.AdminFunctionDetailPage }))
);
const AdminContentPage = lazy(() =>
  import('@/pages/AdminContentPage').then((m) => ({ default: m.AdminContentPage }))
);
const AdminBlogPage = lazy(() =>
  import('@/pages/AdminBlogPage').then((m) => ({ default: m.AdminBlogPage }))
);
const AdminStateFabricPage = lazy(() =>
  import('@/pages/AdminStateFabricPage').then((m) => ({ default: m.AdminStateFabricPage }))
);
const AdminFeedbackPage = lazy(() =>
  import('@/pages/AdminFeedbackPage').then((m) => ({ default: m.AdminFeedbackPage }))
);
const AdminFeaturesPage = lazy(() =>
  import('@/pages/AdminFeaturesPage').then((m) => ({ default: m.AdminFeaturesPage }))
);
const AdminStatusPage = lazy(() =>
  import('@/pages/AdminStatusPage').then((m) => ({ default: m.AdminStatusPage }))
);
const AdminIncidentsPage = lazy(() =>
  import('@/pages/AdminIncidentsPage').then((m) => ({ default: m.AdminIncidentsPage }))
);
const AdminRedirectsPage = lazy(() =>
  import('@/pages/AdminRedirectsPage').then((m) => ({ default: m.AdminRedirectsPage }))
);
const AdminEmailPage = lazy(() =>
  import('@/pages/AdminEmailPage').then((m) => ({ default: m.AdminEmailPage }))
);
const AdminNewsletterPage = lazy(() =>
  import('@/pages/AdminNewsletterPage').then((m) => ({ default: m.AdminNewsletterPage }))
);
const AdminContentCalendarPage = lazy(() =>
  import('@/pages/AdminContentCalendarPage').then((m) => ({ default: m.AdminContentCalendarPage }))
);
const AdminTrustDashboardPage = lazy(() =>
  import('@/pages/AdminTrustDashboardPage').then((m) => ({ default: m.AdminTrustDashboardPage }))
);
const AdminExecutionAuditPage = lazy(() =>
  import('@/pages/AdminExecutionAuditPage').then((m) => ({ default: m.AdminExecutionAuditPage }))
);
const AdminFraudDetectionPage = lazy(() =>
  import('@/pages/AdminFraudDetectionPage').then((m) => ({ default: m.AdminFraudDetectionPage }))
);
const AdminEconomicLeaderboardPage = lazy(() =>
  import('@/pages/AdminEconomicLeaderboardPage').then((m) => ({
    default: m.AdminEconomicLeaderboardPage,
  }))
);
const AdminFactoryPage = lazy(() =>
  import('@/pages/AdminFactoryPage').then((m) => ({ default: m.AdminFactoryPage }))
);
const AdminMaintenancePage = lazy(() =>
  import('@/pages/AdminMaintenancePage').then((m) => ({ default: m.AdminMaintenancePage }))
);
const AdminCachePage = lazy(() =>
  import('@/pages/AdminCachePage').then((m) => ({ default: m.AdminCachePage }))
);
const AdminMonitoringPage = lazy(() =>
  import('@/pages/AdminMonitoringPage').then((m) => ({ default: m.AdminMonitoringPage }))
);
const AdminCloudflareAnalyticsPage = lazy(() =>
  import('@/pages/AdminCloudflareAnalyticsPage').then((m) => ({
    default: m.AdminCloudflareAnalyticsPage,
  }))
);
const AdminSupportPage = lazy(() =>
  import('@/pages/AdminSupportPage').then((m) => ({ default: m.AdminSupportPage }))
);
const AdminSignupInvitesPage = lazy(() =>
  import('@/pages/AdminSignupInvitesPage').then((m) => ({ default: m.AdminSignupInvitesPage }))
);
const AdminWaitlistPage = lazy(() => import('@/pages/AdminWaitlistPage'));
const AdminIPAllowlistPage = lazy(() =>
  import('@/pages/AdminIPAllowlistPage').then((m) => ({ default: m.AdminIPAllowlistPage }))
);
const AdminSIEMPage = lazy(() =>
  import('@/pages/AdminSIEMPage').then((m) => ({ default: m.AdminSIEMPage }))
);

interface AdminRouteConfig {
  path: string;
  index?: boolean;
  component: React.ComponentType;
  permission?: Permission;
  featureName?: string;
}

const ADMIN_ROUTES: AdminRouteConfig[] = [
  { path: '', index: true, component: AdminDashboardPage, permission: 'system:read', featureName: 'Dashboard' },
  { path: 'tenants', component: AdminTenantsPage, permission: 'tenants:read', featureName: 'Tenants' },
  { path: 'tenants/:tenantId', component: AdminTenantDetailPage, permission: 'tenants:read', featureName: 'Tenant details' },
  { path: 'tenants/:tenantId/settings', component: AdminTenantSettingsPage, permission: 'tenants:write', featureName: 'Tenant settings' },
  { path: 'users', component: AdminUsersPage, permission: 'users:read', featureName: 'Users' },
  { path: 'users/:userId', component: AdminUserDetailPage, permission: 'users:read', featureName: 'User details' },
  { path: 'billing', component: AdminBillingPage, permission: 'billing:read', featureName: 'Billing' },
  { path: 'audit', component: AdminAuditPage, permission: 'audit:read', featureName: 'Audit log' },
  { path: 'auth-audit', component: AdminAuthAuditPage, permission: 'audit:read', featureName: 'Auth audit' },
  { path: 'system', component: AdminSystemPage, permission: 'system:read', featureName: 'System' },
  { path: 'backends', component: AdminBackendsPage, permission: 'system:read', featureName: 'Backends' },
  { path: 'providers', component: AdminProvidersPage, permission: 'system:read', featureName: 'Providers' },
  { path: 'functions', component: AdminFunctionsPage, permission: 'functions:read', featureName: 'Functions' },
  { path: 'functions/:functionId', component: AdminFunctionDetailPage, permission: 'functions:read', featureName: 'Function details' },
  { path: 'content', component: AdminContentPage, permission: 'system:read', featureName: 'Content' },
  { path: 'blog', component: AdminBlogPage, permission: 'system:read', featureName: 'Blog' },
  { path: 'state-fabric', component: AdminStateFabricPage, permission: 'tenants:read', featureName: 'State Fabric' },
  { path: 'feedback', component: AdminFeedbackPage, permission: 'system:read', featureName: 'Feedback' },
  { path: 'features', component: AdminFeaturesPage, permission: 'features:read', featureName: 'Feature flags' },
  { path: 'status', component: AdminStatusPage, permission: 'system:read', featureName: 'Status page' },
  { path: 'status/incidents', component: AdminIncidentsPage, permission: 'system:read', featureName: 'Incidents' },
  { path: 'redirects', component: AdminRedirectsPage, permission: 'system:read', featureName: 'Redirects' },
  { path: 'email', component: AdminEmailPage, permission: 'system:read', featureName: 'Email' },
  { path: 'newsletter', component: AdminNewsletterPage, permission: 'system:read', featureName: 'Newsletter' },
  { path: 'content-calendar', component: AdminContentCalendarPage, permission: 'system:read', featureName: 'Content calendar' },
  { path: 'trust-dashboard', component: AdminTrustDashboardPage, permission: 'system:read', featureName: 'Trust dashboard' },
  { path: 'execution-audit', component: AdminExecutionAuditPage, permission: 'audit:read', featureName: 'Execution audit' },
  { path: 'fraud-detection', component: AdminFraudDetectionPage, permission: 'system:read', featureName: 'Fraud detection' },
  { path: 'economic-leaderboard', component: AdminEconomicLeaderboardPage, permission: 'system:read', featureName: 'Economic leaderboard' },
  { path: 'factory', component: AdminFactoryPage, permission: 'system:read', featureName: 'Function factory' },
  { path: 'maintenance', component: AdminMaintenancePage, permission: 'system:write', featureName: 'Maintenance' },
  { path: 'cache', component: AdminCachePage, permission: 'system:write', featureName: 'Cache' },
  { path: 'monitoring', component: AdminMonitoringPage, permission: 'system:read', featureName: 'Monitoring' },
  { path: 'cloudflare', component: AdminCloudflareAnalyticsPage, permission: 'system:read', featureName: 'Cloudflare analytics' },
  { path: 'support', component: AdminSupportPage, permission: 'system:read', featureName: 'Support' },
  { path: 'signup-invites', component: AdminSignupInvitesPage, permission: 'users:write', featureName: 'Signup invites' },
  { path: 'waitlist', component: AdminWaitlistPage, permission: 'users:read', featureName: 'Waitlist' },
  { path: 'ip-allowlist', component: AdminIPAllowlistPage, permission: 'system:write', featureName: 'IP allowlist' },
  { path: 'siem', component: AdminSIEMPage, permission: 'system:write', featureName: 'SIEM' },
];

export function renderAdminRoutes(): ReactNode[] {
  return ADMIN_ROUTES.map(({ path, index, component, permission, featureName }) => (
    <Route
      key={path || 'index'}
      index={index}
      path={index ? undefined : path}
      element={
        <AdminPage component={component} permission={permission} featureName={featureName} />
      }
    />
  ));
}

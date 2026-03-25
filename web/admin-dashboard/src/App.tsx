/**
 * Main App Component
 * Sets up routing for admin dashboard
 */

import { AdminAuthRestore } from '@/components/auth/AdminAuthRestore';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { AdminLayout } from '@/components/layout/AdminLayout';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

// Pages
import { AdminAuditPage } from '@/pages/AdminAuditPage';
import { AdminBillingPage } from '@/pages/AdminBillingPage';
import { AdminDashboardPage } from '@/pages/AdminDashboardPage';
import { AdminLoginPage } from '@/pages/AdminLoginPage';
import { AdminTenantsPage } from '@/pages/AdminTenantsPage';
import { AdminUsersPage } from '@/pages/AdminUsersPage';

// Lazy loaded pages (to be migrated)
import { Suspense, lazy } from 'react';

// Priority 2 pages (lazy loaded)
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
const AdminTenantDetailPage = lazy(() =>
  import('@/pages/AdminTenantDetailPage').then((m) => ({ default: m.AdminTenantDetailPage }))
);
const AdminUserDetailPage = lazy(() =>
  import('@/pages/AdminUserDetailPage').then((m) => ({ default: m.AdminUserDetailPage }))
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
const AdminRegistryPage = lazy(() =>
  import('@/pages/AdminRegistryPage').then((m) => ({ default: m.AdminRegistryPage }))
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
const AdminSupportPage = lazy(() =>
  import('@/pages/AdminSupportPage').then((m) => ({ default: m.AdminSupportPage }))
);
const AdminRedirectsPage = lazy(() =>
  import('@/pages/AdminRedirectsPage').then((m) => ({ default: m.AdminRedirectsPage }))
);
const AdminEmailPage = lazy(() =>
  import('@/pages/AdminEmailPage').then((m) => ({ default: m.AdminEmailPage }))
);
const AdminContentCalendarPage = lazy(() =>
  import('@/pages/AdminContentCalendarPage').then((m) => ({ default: m.AdminContentCalendarPage }))
);
const AdminIncidentsPage = lazy(() =>
  import('@/pages/AdminIncidentsPage').then((m) => ({ default: m.AdminIncidentsPage }))
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

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 30, // 30 minutes
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AdminAuthRestore>
          <Routes>
            {/* All admin routes use dashboard layout (sidebar + header) */}
            <Route element={<AdminLayout />}>
              {/* Public: login page inside same layout */}
              <Route path="/auth/login" element={<AdminLoginPage />} />
              <Route path="/auth/*" element={<Navigate to="/auth/login" replace />} />

              {/* Protected admin routes */}
              <Route element={<ProtectedRoute />}>
                <Route
                  index
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminDashboardPage />
                    </Suspense>
                  }
                />

                {/* Priority 1 Pages - Migrated */}
                <Route
                  path="tenants"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminTenantsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="tenants/:tenantId"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminTenantDetailPage />
                    </Suspense>
                  }
                />
                <Route
                  path="users"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminUsersPage />
                    </Suspense>
                  }
                />
                <Route
                  path="users/:userId"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminUserDetailPage />
                    </Suspense>
                  }
                />
                <Route
                  path="billing"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminBillingPage />
                    </Suspense>
                  }
                />
                <Route
                  path="audit"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminAuditPage />
                    </Suspense>
                  }
                />

                {/* Priority 2 Pages - In Progress */}
                <Route
                  path="system"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminSystemPage />
                    </Suspense>
                  }
                />
                <Route
                  path="backends"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminBackendsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="providers"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminProvidersPage />
                    </Suspense>
                  }
                />
                <Route
                  path="functions"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFunctionsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="functions/:functionId"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFunctionDetailPage />
                    </Suspense>
                  }
                />

                <Route
                  path="content"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminContentPage />
                    </Suspense>
                  }
                />
                <Route
                  path="blog"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminBlogPage />
                    </Suspense>
                  }
                />
                <Route
                  path="registry"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminRegistryPage />
                    </Suspense>
                  }
                />
                <Route
                  path="state-fabric"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminStateFabricPage />
                    </Suspense>
                  }
                />
                <Route
                  path="feedback"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFeedbackPage />
                    </Suspense>
                  }
                />
                <Route
                  path="features"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFeaturesPage />
                    </Suspense>
                  }
                />
                <Route
                  path="status"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminStatusPage />
                    </Suspense>
                  }
                />
                <Route
                  path="status/incidents"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminIncidentsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="redirects"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminRedirectsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="email"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminEmailPage />
                    </Suspense>
                  }
                />
                <Route path="newsletter" element={<Navigate to="/email" replace />} />
                <Route
                  path="content-calendar"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminContentCalendarPage />
                    </Suspense>
                  }
                />
                <Route
                  path="trust-dashboard"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminTrustDashboardPage />
                    </Suspense>
                  }
                />
                <Route
                  path="execution-audit"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminExecutionAuditPage />
                    </Suspense>
                  }
                />
                <Route
                  path="fraud-detection"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFraudDetectionPage />
                    </Suspense>
                  }
                />
                <Route
                  path="economic-leaderboard"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminEconomicLeaderboardPage />
                    </Suspense>
                  }
                />
                <Route
                  path="factory"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminFactoryPage />
                    </Suspense>
                  }
                />
                <Route
                  path="maintenance"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminMaintenancePage />
                    </Suspense>
                  }
                />
                <Route
                  path="cache"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminCachePage />
                    </Suspense>
                  }
                />
                <Route
                  path="monitoring"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminMonitoringPage />
                    </Suspense>
                  }
                />
                <Route
                  path="cloudflare"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminCloudflareAnalyticsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="support"
                  element={
                    <Suspense fallback={<LoadingScreen />}>
                      <AdminSupportPage />
                    </Suspense>
                  }
                />
              </Route>
            </Route>

            {/* Catch-all */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AdminAuthRestore>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;

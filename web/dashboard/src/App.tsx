import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet, useParams, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster, toast } from 'sonner';
import { Bell, Shield, DollarSign, AlertTriangle, MessageSquare } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useNotificationRealtime } from '@/hooks/useNotificationRealtime';
import { CookieConsentProvider } from '@/components/cookie-consent';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { DevA11y } from '@/components/dev/DevA11y';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import type { Notification, NotificationCategory } from '@/types/notifications';
import { LandingPage } from '@/pages/LandingPage';
import { TeamPage } from '@/pages/TeamPage';
import { PricingPage } from '@/pages/PricingPage';
import { FeaturesPage } from '@/pages/FeaturesPage';
import { IntegrationsPage } from '@/pages/IntegrationsPage';
import { AuthPage } from '@/pages/AuthPage';
import { VerifyEmailPage } from '@/pages/AuthPage/VerifyEmailPage';
import { OAuthCallback } from '@/pages/AuthPage/OAuthCallback';
import { OnboardingPage } from '@/pages/OnboardingPage';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { DashboardPage } from '@/pages/DashboardPage';
import { FunctionsPage } from '@/pages/FunctionsPage';
import { FunctionEditorPage } from '@/pages/FunctionsPage/FunctionEditorPage';
import { FunctionDetailPage } from '@/pages/FunctionsPage/FunctionDetailPage';
import { ProvidersPage } from '@/pages/ProvidersPage';
import { AnalyticsPage } from '@/pages/AnalyticsPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { AdminTenantsPage } from '@/pages/AdminTenantsPage';
import { AdminUsersPage } from '@/pages/AdminUsersPage';
import { AdminBillingPage } from '@/pages/AdminBillingPage';
import { AdminAuditPage } from '@/pages/AdminAuditPage';
import { AdminFeaturesPage } from '@/pages/AdminFeaturesPage';
import { AdminSystemPage } from '@/pages/AdminSystemPage';
import { AdminRedirectsPage } from '@/pages/AdminRedirectsPage';
import { AdminNewsletterPage } from '@/pages/AdminNewsletterPage';
import { AdminContentCalendarPage } from '@/pages/AdminContentCalendarPage';
import { AdminFeedbackPage } from '@/pages/AdminFeedbackPage';
import { AdminDashboardPage } from '@/pages/AdminDashboardPage';
import { AdminFunctionsPage } from '@/pages/AdminFunctionsPage';
import { AdminRegistryPage } from '@/pages/AdminRegistryPage';
import { AdminBackendsPage } from '@/pages/AdminBackendsPage';
import { AdminProvidersPage } from '@/pages/AdminProvidersPage';
import { PrivacyPage } from '@/pages/PrivacyPage';
import { SecurityPage } from '@/pages/SecurityPage';
import { TermsPage } from '@/pages/TermsPage';
import ChangelogPage from '@/pages/ChangelogPage';
import { FeedbackPage } from '@/pages/FeedbackPage';
import { FAQPage } from '@/pages/FAQPage';
import BlogPage from '@/pages/BlogPage';
import BlogPostPage from '@/pages/BlogPostPage';
import AdminContentPage from '@/pages/AdminContentPage';
import { ContactPage } from '@/pages/ContactPage';
import { PasswordResetPage } from '@/pages/AuthPage/PasswordResetPage';
import { NotFoundPage } from '@/pages/NotFoundPage';
import { GlobalKeyboardShortcuts } from '@/components/common/GlobalKeyboardShortcuts';
import { Analytics } from '@/components/common/Analytics';
import { ThemeProvider } from '@/components/common/ThemeProvider';
import FunctionPage from '@/pages/FunctionPage';
import ExecutionExplorerPage from '@/pages/ExecutionExplorerPage';
import { PlaygroundPage } from '@/pages/PlaygroundPage';
import { StandalonePlaygroundPage } from '@/pages/StandalonePlaygroundPage';
import { ReplayPage } from '@/pages/ReplayPage';
import { StateFabricPage } from '@/pages/StateFabricPage';
import { StateFabricDetailPage } from '@/pages/StateFabricPage/StateFabricDetailPage';
import { AdminStateFabricPage } from '@/pages/AdminStateFabricPage';
import { AdminTrustDashboardPage } from '@/pages/AdminTrustDashboardPage';
import { AdminExecutionAuditPage } from '@/pages/AdminExecutionAuditPage';
import { AdminFraudDetectionPage } from '@/pages/AdminFraudDetectionPage';
import { AdminEconomicLeaderboardPage } from '@/pages/AdminEconomicLeaderboardPage';
import { StateFabricMarketingPage } from '@/pages/StateFabricMarketingPage';
import { BrowseFunctionsPage } from '@/pages/BrowseFunctionsPage';
import RegistryDeployPage from '@/pages/RegistryDeployPage';
import { DOCS_SITE_URL } from '@/lib/constants';
import { DocsPage } from '@/pages/DocsPage';
import { MyProfilePage } from '@/pages/MyProfilePage';
import { ProfilePage } from '@/pages/ProfilePage/ProfilePage';
import { ProfileSettingsPage } from '@/pages/ProfileSettingsPage';
import { AppsPage } from '@/pages/AppsPage';
import { AppDetailPage } from '@/pages/AppDetailPage';
import { FunctionSettingsPage } from '@/pages/FunctionsPage/FunctionSettingsPage';
import { FunctionLogsPage } from '@/pages/FunctionsPage/FunctionLogsPage';
import { AdminTenantDetailPage } from '@/pages/AdminTenantDetailPage';
import { AdminUserDetailPage } from '@/pages/AdminUserDetailPage';
import { AdminFunctionDetailPage } from '@/pages/AdminFunctionDetailPage';
import { UserDashboardFunctionsPage } from '@/pages/UserDashboardFunctionsPage';
import { UserDashboardSettingsPage } from '@/pages/UserDashboardSettingsPage';
import { TeamsPage } from '@/pages/TeamsPage';
import { ForbiddenPage } from '@/pages/ForbiddenPage';
import { ServerErrorPage } from '@/pages/ServerErrorPage';
import AgentsPage from '@/pages/AgentsPage';
import AgentMarketplacePage from '@/pages/AgentMarketplacePage';
import FunctionMarketplacePage from '@/pages/FunctionMarketplacePage';
import EvolutionPage from '@/pages/EvolutionPage';
import EnterpriseSLAPage from '@/pages/EnterpriseSLAPage';
import { EnterpriseAuditPage } from '@/pages/EnterpriseAuditPage';
import { EnterpriseSupportPage } from '@/pages/EnterpriseSupportPage';
import { UsagePage } from '@/pages/UsagePage';
import { NotificationsPage } from '@/pages/NotificationsPage';
import StatusPage from '@/pages/StatusPage';
import AdminIncidentsPage from '@/pages/AdminStatusPage/AdminIncidentsPage';

function RegistryFunctionRedirect() {
  const { author, name } = useParams<{ author: string; name: string }>();
  if (!author || !name) return <Navigate to="/registry" replace />;
  return <Navigate to={`/fx/${author}/${name}`} replace />;
}

/** Docs landing: link to main docs site (no redirect to avoid loops). */
function DocsRedirectPage() {
  const { slug } = useParams<{ slug?: string }>();
  const target = slug ? `${DOCS_SITE_URL}/${slug}` : DOCS_SITE_URL;
  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-6 bg-bg-primary p-6 text-center">
      <h1 className="text-xl font-semibold text-text-primary">Documentation</h1>
      <p className="text-text-secondary max-w-sm">
        Our full documentation, API reference, and guides live on our docs site.
      </p>
      <a
        href={target}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 rounded-lg bg-brand-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-brand-600 transition-colors"
      >
        Open docs site
        <span aria-hidden>→</span>
      </a>
      <a href="/" className="text-sm text-text-muted hover:text-text-primary transition-colors">
        ← Back to home
      </a>
    </div>
  );
}

// Alias for any remaining references (e.g. hot reload cache)
const RedirectToDocs = DocsRedirectPage;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
    },
  },
});

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check if user is an admin (skip onboarding for admins)
  const isAdmin =
    user?.role &&
    ['super_admin', 'support', 'billing_admin', 'developer_admin'].includes(user.role);
  if (isAdmin) {
    return <>{children}</>;
  }

  // Allow access to dashboard even with incomplete onboarding - banner will show resume option
  // Only redirect to onboarding for completely new users (no steps completed)
  const { completedSteps } = useOnboardingStore.getState();
  if (!isOnboardingComplete && !hasSkippedOnboarding && completedSteps.length === 0) {
    return <Navigate to="/onboarding" replace />;
  }

  return <>{children}</>;
}

function OnboardingRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check if user is an admin (admins skip onboarding entirely)
  const isAdmin =
    user?.role &&
    ['super_admin', 'support', 'billing_admin', 'developer_admin'].includes(user.role);
  if (isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  // If onboarding is complete or skipped, redirect to dashboard
  if (isOnboardingComplete || hasSkippedOnboarding) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const user = useAuthStore((state) => state.user);

  // Show loading state while auth is being initialized
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" text="Loading..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  const isAdmin =
    user?.role &&
    ['super_admin', 'support', 'billing_admin', 'developer_admin'].includes(user.role);
  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return !isAuthenticated ? <>{children}</> : <Navigate to="/dashboard" replace />;
}

/**
 * Icon mapping for notification categories
 */
const CATEGORY_ICONS: Record<NotificationCategory, React.ReactNode> = {
  all: <Bell className="h-4 w-4" />,
  trust: <Shield className="h-4 w-4" />,
  revenue: <DollarSign className="h-4 w-4" />,
  issues: <AlertTriangle className="h-4 w-4" />,
  messages: <MessageSquare className="h-4 w-4" />,
  security: <Shield className="h-4 w-4" />,
};

/**
 * Color mapping for notification priorities
 */
const PRIORITY_COLORS: Record<string, string> = {
  critical: 'var(--error)',
  high: 'var(--warning)',
  medium: 'var(--info)',
  low: 'var(--success)',
};

/**
 * NotificationsProvider - Handles real-time notifications and toast integration
 */
function NotificationsProvider({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const { updateUnreadCount, updateUnreadCounts } = useNotificationStore();

  useNotificationRealtime({
    enabled: isAuthenticated && !!user?.id,
    onNewNotification: (notification: Notification) => {
      // Update unread counts
      const currentState = useNotificationStore.getState();
      updateUnreadCount('all', currentState.unreadCounts.all + 1);
      if (notification.category !== 'all') {
        updateUnreadCount(
          notification.category,
          currentState.unreadCounts[notification.category] + 1
        );
      }

      // Show toast notification
      const icon = CATEGORY_ICONS[notification.category] || CATEGORY_ICONS.all;
      const borderColor = PRIORITY_COLORS[notification.priority] || 'var(--brand-500)';

      toast.info(
        <div className="flex items-start gap-3">
          <div className="mt-0.5" style={{ color: borderColor }}>
            {icon}
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-medium text-text-primary">{notification.title}</p>
            <p className="text-sm text-text-secondary truncate">{notification.message}</p>
          </div>
        </div>,
        {
          duration: 5000,
          action: {
            label: 'View',
            onClick: () => {
              if (notification.actionUrl) {
                navigate(notification.actionUrl);
              } else {
                navigate('/notifications');
              }
            },
          },
          style: {
            borderLeft: `4px solid ${borderColor}`,
            cursor: 'pointer',
          },
          onDismiss: () => {
            // Optional: track dismissed notifications
          },
          onAutoClose: () => {
            // Optional: track auto-closed notifications
          },
        }
      );
    },
    onTrustAlert: (alert) => {
      // Show critical trust alerts as error toasts
      if (alert.severity === 'critical' || alert.severity === 'emergency') {
        toast.error(
          <div className="flex items-start gap-3">
            <Shield className="h-4 w-4 mt-0.5 text-error" />
            <div className="flex-1 min-w-0">
              <p className="font-medium text-text-primary">Trust Alert: {alert.title}</p>
              <p className="text-sm text-text-secondary truncate">{alert.description}</p>
            </div>
          </div>,
          {
            duration: 10000,
            action: {
              label: 'View Details',
              onClick: () => {
                if (alert.actionUrl) {
                  navigate(alert.actionUrl);
                } else {
                  navigate('/notifications');
                }
              },
            },
            style: {
              borderLeft: '4px solid var(--error)',
              cursor: 'pointer',
            },
          }
        );
      }
    },
  });

  return <>{children}</>;
}

function AppContent() {
  const initialize = useAuthStore((state) => state.initialize);

  useEffect(() => {
    initialize();
  }, [initialize]);

  return (
    <NotificationsProvider>
      <Routes>
      {/* Public Routes */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/status" element={<StatusPage />} />
      <Route path="/pricing" element={<PricingPage />} />
      <Route path="/features" element={<FeaturesPage />} />
      <Route path="/integrations" element={<IntegrationsPage />} />
      <Route path="/team" element={<TeamPage />} />
      <Route path="/privacy" element={<PrivacyPage />} />
      <Route path="/security" element={<SecurityPage />} />
      <Route path="/terms" element={<TermsPage />} />
      <Route path="/changelog" element={<ChangelogPage />} />
      <Route path="/feedback" element={<FeedbackPage />} />
      <Route path="/faq" element={<FAQPage />} />
      <Route path="/contact" element={<ContactPage />} />
      <Route path="/docs" element={<DocsPage />} />
      <Route path="/docs/:slug" element={<DocsPage />} />
      <Route path="/blog" element={<BlogPage />} />
      <Route path="/blog/:slug" element={<BlogPostPage />} />
      <Route path="/products/state-fabric" element={<StateFabricMarketingPage />} />
      <Route path="/registry" element={<BrowseFunctionsPage />} />
      <Route path="/registry/:author/:name" element={<RegistryFunctionRedirect />} />

      {/* Public user profile pages */}
      <Route path="/u/:username" element={<ProfilePage />} />
      <Route path="/profile/:username" element={<ProfilePage />} />

      {/* Standalone Playground (Public) */}
      <Route path="/playground" element={<StandalonePlaygroundPage />} />

      {/* Registry Playground Routes (Public) */}
      <Route path="/fx/:author/:name" element={<FunctionPage />} />
      <Route path="/run/:author/:name" element={<PlaygroundPage />} />
      <Route path="/registry/:author/:name/executions" element={<ExecutionExplorerPage />} />
      <Route path="/run/:appSlug/:functionName" element={<PlaygroundPage />} />
      <Route path="/replay/:execId" element={<ReplayPage />} />
      <Route
        path="/login"
        element={
          <PublicRoute>
            <AuthPage />
          </PublicRoute>
        }
      />
      <Route
        path="/signup"
        element={
          <PublicRoute>
            <AuthPage />
          </PublicRoute>
        }
      />
      <Route
        path="/auth/verify-email"
        element={
          <PublicRoute>
            <VerifyEmailPage />
          </PublicRoute>
        }
      />
      <Route
        path="/auth/oauth/callback"
        element={
          <PublicRoute>
            <OAuthCallback />
          </PublicRoute>
        }
      />
      <Route path="/auth/reset-password" element={<PasswordResetPage />} />

      {/* Onboarding Route */}
      <Route
        path="/onboarding"
        element={
          <OnboardingRoute>
            <OnboardingPage />
          </OnboardingRoute>
        }
      />

      {/* Protected Dashboard Routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <DashboardLayout />
          </ProtectedRoute>
        }
      >
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="notifications" element={<NotificationsPage />} />
        <Route path="apps" element={<AppsPage />} />
        <Route path="apps/:appId" element={<AppDetailPage />} />
        <Route path="functions" element={<FunctionsPage />} />
        <Route path="functions/new" element={<FunctionEditorPage />} />
        <Route path="functions/deploy" element={<RegistryDeployPage />} />
        <Route path="functions/:id" element={<FunctionDetailPage />} />
        <Route path="functions/:id/edit" element={<FunctionEditorPage />} />
        <Route path="functions/:author/:name/settings" element={<FunctionSettingsPage />} />
        <Route path="functions/:author/:name/logs" element={<FunctionLogsPage />} />
        <Route path="providers" element={<ProvidersPage />} />
        <Route path="analytics" element={<AnalyticsPage />} />
        <Route path="usage" element={<UsagePage />} />
        <Route path="state-fabric" element={<StateFabricPage />} />
        <Route path="state-fabric/new" element={<StateFabricDetailPage />} />
        <Route path="state-fabric/:id" element={<StateFabricDetailPage />} />
        <Route path="state-fabric/:id/edit" element={<StateFabricDetailPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="teams" element={<TeamsPage />} />
        {/* Enterprise Routes */}
        <Route path="enterprise/sla" element={<EnterpriseSLAPage />} />
        <Route path="enterprise/audit" element={<EnterpriseAuditPage />} />
        <Route path="enterprise/support" element={<EnterpriseSupportPage />} />
        <Route path="profile" element={<MyProfilePage />} />
        <Route path="profile/settings" element={<ProfileSettingsPage />} />
        <Route path="dashboard/:username/functions" element={<UserDashboardFunctionsPage />} />
        <Route
          path="u/:username/settings/billing"
          element={<UserDashboardSettingsPage initialTab="billing" />}
        />
        <Route path="u/:username/settings" element={<UserDashboardSettingsPage />} />

        {/* Agent Routes */}
        <Route path="agents" element={<AgentsPage />} />
        <Route path="agents/:agentId" element={<AgentsPage />} />
        <Route path="marketplace/agents" element={<AgentMarketplacePage />} />
        <Route path="marketplace/functions" element={<FunctionMarketplacePage />} />
        <Route path="evolution" element={<EvolutionPage />} />
        <Route path="evolution/:agentId" element={<EvolutionPage />} />

        {/* Admin Routes - use Outlet so only one DashboardLayout (parent) is used */}
        <Route
          path="admin"
          element={
            <AdminRoute>
              <Outlet />
            </AdminRoute>
          }
        >
          <Route index element={<AdminDashboardPage />} />
          <Route path="tenants" element={<AdminTenantsPage />} />
          <Route path="tenants/:tenantId" element={<AdminTenantDetailPage />} />
          <Route path="users" element={<AdminUsersPage />} />
          <Route path="users/:userId" element={<AdminUserDetailPage />} />
          <Route path="billing" element={<AdminBillingPage />} />
          <Route path="features" element={<AdminFeaturesPage />} />
          <Route path="audit" element={<AdminAuditPage />} />
          <Route path="system" element={<AdminSystemPage />} />
          <Route path="backends" element={<AdminBackendsPage />} />
          <Route path="providers" element={<AdminProvidersPage />} />
          <Route path="redirects" element={<AdminRedirectsPage />} />
          <Route path="newsletter" element={<AdminNewsletterPage />} />
          <Route path="content-calendar" element={<AdminContentCalendarPage />} />
          <Route path="content" element={<AdminContentPage />} />
          <Route path="feedback" element={<AdminFeedbackPage />} />
          <Route path="functions" element={<AdminFunctionsPage />} />
          <Route path="functions/:functionId" element={<AdminFunctionDetailPage />} />
          <Route path="registry" element={<AdminRegistryPage />} />
          <Route path="registry/functions" element={<AdminRegistryPage />} />
          <Route path="registry/functions/:functionId" element={<AdminFunctionDetailPage />} />
          <Route path="state-fabric" element={<AdminStateFabricPage />} />
          <Route path="trust-dashboard" element={<AdminTrustDashboardPage />} />
          <Route path="execution-audit" element={<AdminExecutionAuditPage />} />
          <Route path="fraud-detection" element={<AdminFraudDetectionPage />} />
          <Route path="economic-leaderboard" element={<AdminEconomicLeaderboardPage />} />
          <Route path="status/incidents" element={<AdminIncidentsPage />} />
        </Route>
      </Route>

      {/* 404 - Not Found */}
      <Route path="*" element={<NotFoundPage />} />

      {/* 403 - Forbidden */}
      <Route path="/forbidden" element={<ForbiddenPage />} />

      {/* 500 - Internal Server Error */}
      <Route path="/error" element={<ServerErrorPage />} />
      </Routes>
    </NotificationsProvider>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <CookieConsentProvider>
            <Analytics />
            <DevA11y />
            <BrowserRouter>
              <GlobalKeyboardShortcuts />
              <AppContent />
            </BrowserRouter>
            <Toaster
              position="bottom-right"
              toastOptions={{
                style: {
                  background: 'var(--bg-secondary)',
                  border: '1px solid var(--border-subtle)',
                  color: 'var(--text-primary)',
                },
              }}
            />
          </CookieConsentProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;

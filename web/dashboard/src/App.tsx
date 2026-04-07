import { Analytics } from '@/components/common/Analytics';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { GlobalKeyboardShortcuts } from '@/components/common/GlobalKeyboardShortcuts';
import { ThemeProvider } from '@/components/common/ThemeProvider';
import { CookieConsentProvider } from '@/components/cookie-consent';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { useNotificationRealtime } from '@/hooks/useNotificationRealtime';
import { useNotificationUnreadPolling } from '@/hooks/useNotificationUnreadPolling';
import { redirectToAuthSite } from '@/lib/auth-integration';
import {
  COMING_SOON_ONLY,
  getMarketingPageUrl,
  getMarketingRedirectOrigin,
  getPublicDocsSiteOrigin,
} from '@/lib/constants';
import { unreadStoreKeyFromEventCategory } from '@/lib/notification-unread-sync';
import { isPlatformAdminRole } from '@/lib/platform-admin';
import AgentMarketplacePage from '@/pages/AgentMarketplacePage';
import { AgentMemoryPage } from '@/pages/AgentMemoryPage';
import { AgentMemoryDetailPage } from '@/pages/AgentMemoryPage/AgentMemoryDetailPage';
import AgentSDKIntegrationsPage from '@/pages/AgentSDKIntegrationsPage';
import AgentsPage from '@/pages/AgentsPage';
import { AnalyticsPage } from '@/pages/AnalyticsPage';
import { APIKeyDetailPage, APIKeysPage } from '@/pages/api-keys';
import { AppDetailPage } from '@/pages/AppDetailPage';
import { AppsPage, CreateAppPage } from '@/pages/AppsPage';
import { AuthCallbackPage } from '@/pages/AuthCallbackPage';
import { BrowseFunctionsPage } from '@/pages/BrowseFunctionsPage';
import ChangelogPage from '@/pages/ChangelogPage';
import { ContactPage } from '@/pages/ContactPage';
import ConversationsPage from '@/pages/ConversationsPage';
import { DashboardPage } from '@/pages/DashboardPage';
import EnterpriseSLAPage from '@/pages/EnterpriseSLAPage';
import EvolutionPage from '@/pages/EvolutionPage';
import ExecutionExplorerPage from '@/pages/ExecutionExplorerPage';
import { FAQPage } from '@/pages/FAQPage';
import { FeaturesPage } from '@/pages/FeaturesPage';
import { FeedbackPage } from '@/pages/FeedbackPage';
import { ForbiddenPage } from '@/pages/ForbiddenPage';
import FunctionMarketplacePage from '@/pages/FunctionMarketplacePage';
import FunctionPage from '@/pages/FunctionPage';
import { FunctionsPage } from '@/pages/FunctionsPage';
import { FunctionDetailPage } from '@/pages/FunctionsPage/FunctionDetailPage';
import { FunctionEditorPage } from '@/pages/FunctionsPage/FunctionEditorPage';
import { FunctionLogsPage } from '@/pages/FunctionsPage/FunctionLogsPage';
import { FunctionSettingsPage } from '@/pages/FunctionsPage/FunctionSettingsPage';
import { IntegrationsPage } from '@/pages/IntegrationsPage';
import { LaunchPage } from '@/pages/LaunchPage';
import { MyProfilePage } from '@/pages/MyProfilePage';
import { NotFoundPage } from '@/pages/NotFoundPage';
import { OnboardingPage } from '@/pages/OnboardingPage';
import { PlaygroundPage } from '@/pages/PlaygroundPage';
import { ProfilePage } from '@/pages/ProfilePage/ProfilePage';
import { ProfileSettingsPage } from '@/pages/ProfileSettingsPage';
import { ProvidersPage } from '@/pages/ProvidersPage';
import RegistryDeployPage from '@/pages/RegistryDeployPage';
import { ReplayPage } from '@/pages/ReplayPage';
import { SecretsPage } from '@/pages/SecretsPage';
import { SecurityPage } from '@/pages/SecurityPage';
import { ServerErrorPage } from '@/pages/ServerErrorPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { StandalonePlaygroundPage } from '@/pages/StandalonePlaygroundPage';
import { StateFabricMarketingPage } from '@/pages/StateFabricMarketingPage';
import { StateFabricPage } from '@/pages/StateFabricPage';
import { StateFabricDetailPage } from '@/pages/StateFabricPage/StateFabricDetailPage';
import { StatePage } from '@/pages/StatePage';
import { StateDetailPage } from '@/pages/StatePage/StateDetailPage';
import { TeamPage } from '@/pages/TeamPage';
import { TeamsPage } from '@/pages/TeamsPage';
import { UserDashboardFunctionsPage } from '@/pages/UserDashboardFunctionsPage';
import { UserDashboardSettingsPage } from '@/pages/UserDashboardSettingsPage';
import WalletPage from '@/pages/WalletPage';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import type { Notification, NotificationCategory } from '@/types/notifications';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AlertTriangle, Bell, DollarSign, Loader2, MessageSquare, Shield } from 'lucide-react';
import { useEffect, useState } from 'react';
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from 'react-router-dom';
import { Toaster, toast } from 'sonner';

// Flywheel Network imports
import { FlywheelLayout } from '@/components/flywheel/layout/FlywheelLayout';
import { EnterpriseAuditPage } from '@/pages/EnterpriseAuditPage';
import { EnterpriseSupportPage } from '@/pages/EnterpriseSupportPage';
import ChallengePage from '@/pages/flywheel/ChallengePage';
import FlywheelHubPage from '@/pages/flywheel/FlywheelHubPage';
import LeaderboardPage from '@/pages/flywheel/LeaderboardPage';
import ThreadListPage from '@/pages/flywheel/ThreadListPage';
import { NotificationsPage } from '@/pages/NotificationsPage';
import StatusPage from '@/pages/StatusPage';
import { UsagePage } from '@/pages/UsagePage';

function RegistryFunctionRedirect() {
  const { author, name } = useParams<{ author: string; name: string }>();
  if (!author || !name) return <Navigate to="/registry" replace />;
  return <Navigate to={`/fx/${author}/${name}`} replace />;
}

/** Shown while `initialize()` has not finished — avoids protected pages firing API calls without a token. */
function AuthSessionLoading() {
  return (
    <div
      className="flex min-h-[40vh] items-center justify-center"
      aria-busy="true"
      aria-label="Checking session"
    >
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );
}

/** Handles "/": logged-out users go to the Astro marketing site (web/site); authenticated → dashboard. */
function HomeRedirect() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const authChecked = useAuthStore((s) => s.authChecked);
  const marketingOrigin = getMarketingRedirectOrigin();

  useEffect(() => {
    if (!authChecked) return;
    if (isAuthenticated) return;
    window.location.replace(marketingOrigin);
  }, [authChecked, isAuthenticated, marketingOrigin]);

  if (!authChecked) return <AuthSessionLoading />;
  if (isAuthenticated) return <Navigate to="/overview" replace />;
  return null;
}

/** Sends /docs and /docs/:slug to the standalone Astro docs site (web/docs). */
function DocsOutboundRedirect() {
  const { slug } = useParams<{ slug?: string }>();
  const origin = getPublicDocsSiteOrigin();

  useEffect(() => {
    const path = slug ? `/docs/${slug}` : '/';
    window.location.replace(`${origin}${path}`);
  }, [slug, origin]);

  return null;
}

/** Privacy / terms live on the Astro marketing site (web/site). */
function MarketingLegalRedirect({ page }: { page: 'privacy' | 'terms' }) {
  const url = getMarketingPageUrl(`/${page}`);
  useEffect(() => {
    window.location.replace(url);
  }, [url]);
  return null;
}

/** Public blog lives on the Astro marketing site (web/site); redirect app /blog bookmarks. */
function MarketingBlogRedirect() {
  const { slug } = useParams<{ slug?: string }>();
  const path = slug ? `/blog/${slug}` : '/blog';
  const url = getMarketingPageUrl(path);
  useEffect(() => {
    window.location.replace(url);
  }, [url]);
  return null;
}

/** Pricing lives on the Astro marketing site (web/site). */
function MarketingPricingRedirect() {
  const url = getMarketingPageUrl('/pricing');
  useEffect(() => {
    window.location.replace(url);
  }, [url]);
  return null;
}

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
  const authChecked = useAuthStore((state) => state.authChecked);
  const user = useAuthStore((state) => state.user);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);
  const location = useLocation();

  if (!authChecked) {
    return <AuthSessionLoading />;
  }

  if (!isAuthenticated) {
    // Non-signed-in users on app.functionfly.com redirect to auth.functionfly.com
    redirectToAuthSite(location.pathname + location.search);
    return <AuthSessionLoading />;
  }

  // Platform admins skip onboarding (matches backend IsAdminRole)
  if (isPlatformAdminRole(user?.role)) {
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
  const authChecked = useAuthStore((state) => state.authChecked);
  const user = useAuthStore((state) => state.user);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);
  const location = useLocation();

  if (!authChecked) {
    return <AuthSessionLoading />;
  }

  if (!isAuthenticated) {
    // Non-signed-in users on app.functionfly.com redirect to auth.functionfly.com
    redirectToAuthSite(location.pathname + location.search);
    return <AuthSessionLoading />;
  }

  if (isPlatformAdminRole(user?.role)) {
    return <Navigate to="/overview" replace />;
  }

  // If onboarding is complete or skipped, redirect to dashboard
  if (isOnboardingComplete || hasSkippedOnboarding) {
    return <Navigate to="/overview" replace />;
  }

  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return !isAuthenticated ? <>{children}</> : <Navigate to="/overview" replace />;
}

/** Redirects unauthenticated users directly to the standalone auth site (auth.functionfly.com) */
function RedirectToAuth() {
  const location = useLocation();
  const authChecked = useAuthStore((s) => s.authChecked);

  useEffect(() => {
    if (!authChecked) return;
    redirectToAuthSite(location.pathname + location.search);
  }, [authChecked, location]);

  return <AuthSessionLoading />;
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
  const { updateUnreadCount } = useNotificationStore();

  useNotificationUnreadPolling();

  useNotificationRealtime({
    enabled: isAuthenticated && !!user?.id,
    onNewNotification: (notification: Notification) => {
      // Update unread counts
      const currentState = useNotificationStore.getState();
      updateUnreadCount('all', currentState.unreadCounts.all + 1);
      const slice = unreadStoreKeyFromEventCategory(notification.category);
      if (slice) {
        const prev = currentState.unreadCounts[slice] ?? 0;
        updateUnreadCount(slice, prev + 1);
      }

      // Show toast notification (API may use `team`; UI buckets use `messages`)
      const rawCat = notification.category as string;
      const toastCategory: NotificationCategory =
        rawCat === 'team' ? 'messages' : notification.category;
      const icon = CATEGORY_ICONS[toastCategory] || CATEGORY_ICONS.all;
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

/** Auth-related paths that must work even when COMING_SOON_ONLY is active. */
const AUTH_BYPASS_PATHS = [
  '/auth/',
  '/login',
  '/signup',
  '/auth/verify-email',
  '/auth/reset-password',
  '/auth/callback',
  '/auth/oauth/callback',
];

function AppContent() {
  const initialize = useAuthStore((state) => state.initialize);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const authChecked = useAuthStore((s) => s.authChecked);
  const location = useLocation();
  const [hasInitialized, setHasInitialized] = useState(false);

  useEffect(() => {
    if (!hasInitialized) {
      setHasInitialized(true);
      initialize();
    }
  }, [initialize, hasInitialized]);

  // functionfly.com pre-launch: redirect unauthenticated users to auth site.
  // Signed-in users bypass it and see the full dashboard.
  // Auth-related paths always bypass so login/callback flows work.
  if (COMING_SOON_ONLY) {
    const isAuthPath = AUTH_BYPASS_PATHS.some((p) => location.pathname.startsWith(p));
    if (!isAuthPath) {
      if (!authChecked) return <AuthSessionLoading />;
      if (!isAuthenticated) {
        // Redirect non-signed-in users to auth.functionfly.com instead of showing launch page
        redirectToAuthSite(location.pathname + location.search);
        return <AuthSessionLoading />;
      }
      // authenticated → fall through to full routes
    }
  }

  return (
    <NotificationsProvider>
      <Routes>
        {/* Public Routes */}
        <Route path="/" element={<HomeRedirect />} />
        <Route path="/launch" element={<LaunchPage />} />
        <Route path="/coming-soon" element={<LaunchPage />} />
        <Route path="/status" element={<StatusPage />} />
        <Route path="/pricing" element={<MarketingPricingRedirect />} />
        <Route path="/features" element={<FeaturesPage />} />
        <Route path="/integrations" element={<IntegrationsPage />} />
        <Route path="/team" element={<TeamPage />} />
        <Route path="/privacy" element={<MarketingLegalRedirect page="privacy" />} />
        <Route path="/security" element={<SecurityPage />} />
        <Route path="/terms" element={<MarketingLegalRedirect page="terms" />} />
        <Route path="/changelog" element={<ChangelogPage />} />
        <Route path="/feedback" element={<FeedbackPage />} />
        <Route path="/faq" element={<FAQPage />} />
        <Route path="/contact" element={<ContactPage />} />
        <Route path="/docs" element={<DocsOutboundRedirect />} />
        <Route path="/docs/:slug" element={<DocsOutboundRedirect />} />
        <Route path="/blog" element={<MarketingBlogRedirect />} />
        <Route path="/blog/:slug" element={<MarketingBlogRedirect />} />
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
        <Route path="/auth" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<RedirectToAuth />} />
        <Route path="/auth/login" element={<RedirectToAuth />} />
        <Route path="/signup" element={<RedirectToAuth />} />
        <Route path="/auth/verify-email" element={<RedirectToAuth />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route path="/auth/oauth/callback" element={<AuthCallbackPage />} />
        <Route path="/auth/reset-password" element={<RedirectToAuth />} />

        {/* Onboarding Route */}
        <Route
          path="/onboarding"
          element={
            <OnboardingRoute>
              <OnboardingPage />
            </OnboardingRoute>
          }
        />

        {/* Flywheel Network – must be before path="/" so /flywheel doesn't match dashboard layout */}
        <Route
          path="/flywheel"
          element={
            <ProtectedRoute>
              <FlywheelLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<FlywheelHubPage />} />
          <Route path="threads" element={<ThreadListPage />} />
          <Route path="threads/new" element={<div>New Thread Page (Coming Soon)</div>} />
          <Route path="threads/:id" element={<div>Thread Detail Page (Coming Soon)</div>} />
          <Route path="challenges" element={<ChallengePage />} />
          <Route path="challenges/:id" element={<div>Challenge Detail Page (Coming Soon)</div>} />
          <Route path="leaderboards" element={<LeaderboardPage />} />
          <Route path="reputation/:userId" element={<div>Reputation Page (Coming Soon)</div>} />
        </Route>

        {/* Protected Dashboard Routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route path="dashboard" element={<FunctionMarketplacePage />} />
          <Route path="overview" element={<DashboardPage />} />
          <Route path="notifications" element={<NotificationsPage />} />
          <Route path="apps" element={<AppsPage />} />
          <Route path="apps/new" element={<CreateAppPage />} />
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
          <Route path="state" element={<StatePage />} />
          <Route path="state/new" element={<StateDetailPage />} />
          <Route path="state/:path" element={<StateDetailPage />} />
          <Route path="agent-memories" element={<AgentMemoryPage />} />
          <Route path="agent-memories/new" element={<AgentMemoryDetailPage />} />
          <Route path="agent-memories/:id" element={<AgentMemoryDetailPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="api-keys" element={<APIKeysPage />} />
          <Route path="api-keys/:keyId" element={<APIKeyDetailPage />} />
          {/* Same pages at /dashboard/api-keys for consistent nav links */}
          <Route path="dashboard/api-keys" element={<APIKeysPage />} />
          <Route path="dashboard/api-keys/:keyId" element={<APIKeyDetailPage />} />
          <Route path="secrets" element={<SecretsPage />} />
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
          <Route path="sdk-integrations" element={<AgentSDKIntegrationsPage />} />
          <Route path="marketplace/agents" element={<AgentMarketplacePage />} />
          <Route path="marketplace/functions" element={<Navigate to="/dashboard" replace />} />
          <Route path="evolution" element={<EvolutionPage />} />
          <Route path="evolution/:agentId" element={<EvolutionPage />} />
          <Route path="wallet" element={<WalletPage />} />
          <Route path="wallet/:agentId" element={<WalletPage />} />

          <Route path="conversations" element={<ConversationsPage />} />
          <Route path="conversations/:id" element={<ConversationsPage />} />
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
            <BrowserRouter>
              <Analytics />
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

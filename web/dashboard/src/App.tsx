import { Analytics } from '@/components/common/Analytics';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { GlobalKeyboardShortcuts } from '@/components/common/GlobalKeyboardShortcuts';
import { PageViewTracker } from '@/components/common/PageViewTracker';
import { ThemeAwareToaster } from '@/components/common/ThemeAwareToaster';
import { ThemeProvider } from '@/components/common/ThemeProvider';
import { CookieConsentProvider } from '@/components/cookie-consent';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { UserProfileLayout } from '@/components/layout/UserProfileLayout';
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
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import type { Notification, NotificationCategory } from '@/types/notifications';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AlertTriangle, Bell, DollarSign, Loader2, MessageSquare, Shield } from 'lucide-react';
import { NuqsAdapter } from 'nuqs/adapters/react-router/v7';
import { lazy, Suspense, useEffect, useState } from 'react';
import { HelmetProvider } from 'react-helmet-async';
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from 'react-router-dom';
import { toast } from 'sonner';

// Helper to create lazy loaded page components with named exports
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const lazyPage = (importFn: () => Promise<any>, exportName: string) =>
  lazy(() => importFn().then((m) => ({ default: m[exportName] ?? m.default })));

// Route components - all lazy-loaded for code splitting
const AgentAnalyticsPage = lazyPage(() => import('@/pages/AgentAnalyticsPage'), 'AgentAnalyticsPage');
const AgentObservabilityPage = lazyPage(() => import('@/pages/AgentObservabilityPage'), 'AgentObservabilityPage');
const AgentCreatePage = lazyPage(() => import('@/pages/AgentCreatePage'), 'AgentCreatePage');
const AgentDetailPage = lazyPage(() => import('@/pages/AgentDetailPage'), 'AgentDetailPage');
const AgentEditPage = lazyPage(() => import('@/pages/AgentEditPage'), 'AgentEditPage');
const AgentMarketplaceDetailPage = lazyPage(() => import('@/pages/AgentMarketplaceDetailPage'), 'AgentMarketplaceDetailPage');
const AgentMemoryPage = lazyPage(() => import('@/pages/AgentMemoryPage'), 'AgentMemoryPage');
const AgentMemoryDetailPage = lazyPage(() => import('@/pages/AgentMemoryPage/AgentMemoryDetailPage'), 'AgentMemoryDetailPage');
const AgentSDKIntegrationsPage = lazyPage(() => import('@/pages/AgentSDKIntegrationsPage'), 'AgentSDKIntegrationsPage');
const MarketplacePage = lazyPage(() => import('@/pages/MarketplacePage'), 'MarketplacePage');
const ExtensionDetailPage = lazyPage(() => import('@/pages/ExtensionDetailPage'), 'ExtensionDetailPage');
const AgentsPage = lazyPage(() => import('@/pages/AgentsPage'), 'AgentsPage');
const AgentWalletPage = lazyPage(() => import('@/pages/AgentWalletPage'), 'AgentWalletPage');
const AgentWorkspacePage = lazyPage(() => import('@/pages/AgentWorkspacePage'), 'AgentWorkspacePage');
const AIComposerPage = lazyPage(() => import('@/pages/AIComposerPage'), 'AIComposerPage');
const AnalyticsPage = lazyPage(() => import('@/pages/AnalyticsPage'), 'AnalyticsPage');
const APIKeyDetailPage = lazyPage(() => import('@/pages/api-keys'), 'APIKeyDetailPage');
const APIKeysPage = lazyPage(() => import('@/pages/api-keys'), 'APIKeysPage');
const AppDetailPage = lazyPage(() => import('@/pages/AppDetailPage'), 'AppDetailPage');
const AppsPage = lazyPage(() => import('@/pages/AppsPage'), 'AppsPage');
const CreateAppPage = lazyPage(() => import('@/pages/AppsPage'), 'CreateAppPage');
const AuthCallbackPage = lazyPage(() => import('@/pages/AuthCallbackPage'), 'AuthCallbackPage');
const MagicLinkVerifyPage = lazyPage(() => import('@/pages/AuthPage/MagicLinkVerifyPage'), 'MagicLinkVerifyPage');
const BrowseFunctionsPage = lazyPage(() => import('@/pages/BrowseFunctionsPage'), 'BrowseFunctionsPage');
const BundlePricingPage = lazyPage(() => import('@/pages/BundlePricingPage'), 'BundlePricingPage');
const BundleProvisioningPage = lazyPage(() => import('@/pages/BundleProvisioningPage'), 'BundleProvisioningPage');
const BundleOverviewPage = lazyPage(() => import('@/pages/BundleOverviewPage'), 'BundleOverviewPage');
const BundleFunctionsPage = lazyPage(() => import('@/pages/BundleFunctionsPage'), 'BundleFunctionsPage');
const BundleIntegrationsPage = lazyPage(() => import('@/pages/BundleIntegrationsPage'), 'BundleIntegrationsPage');
const BundleIntegrationDetailPage = lazyPage(() => import('@/pages/BundleIntegrationDetailPage'), 'BundleIntegrationDetailPage');
const BundleConfigPage = lazyPage(() => import('@/pages/BundleConfigPage'), 'BundleConfigPage');
const MyBundlesPage = lazyPage(() => import('@/pages/MyBundlesPage'), 'MyBundlesPage');
const BrainPage = lazyPage(() => import('@/pages/BrainPage'), 'BrainPage');
const CertificationPage = lazyPage(() => import('@/pages/CertificationPage'), 'CertificationPage');
const ChangelogPage = lazyPage(() => import('@/pages/ChangelogPage'), 'ChangelogPage');
const CommunityPage = lazyPage(() => import('@/pages/CommunityPage'), 'CommunityPage');
const CommunityThreadPage = lazyPage(() => import('@/pages/CommunityPage/CommunityThreadPage'), 'CommunityThreadPage');
const CommunityUserProfilePage = lazyPage(() => import('@/pages/CommunityPage/CommunityUserProfilePage'), 'CommunityUserProfilePage');
const CommunityBookmarksPage = lazyPage(() => import('@/pages/CommunityPage/CommunityBookmarksPage'), 'CommunityBookmarksPage');
const CommunityNotificationsPage = lazyPage(() => import('@/pages/CommunityPage/CommunityNotificationsPage'), 'CommunityNotificationsPage');
const ConnectorsCallbackPage = lazyPage(() => import('@/pages/ConnectorsCallbackPage'), 'ConnectorsCallbackPage');
const ContactPage = lazyPage(() => import('@/pages/ContactPage'), 'ContactPage');
const ConversationsPage = lazyPage(() => import('@/pages/ConversationsPage'), 'ConversationsPage');
const CredentialsPage = lazyPage(() => import('@/pages/CredentialsPage'), 'CredentialsPage');
const DashboardPage = lazyPage(() => import('@/pages/DashboardPage'), 'DashboardPage');
const DecisionsPage = lazyPage(() => import('@/pages/DecisionsPage'), 'DecisionsPage');
const DNAOverviewPage = lazyPage(() => import('@/pages/DNAOverviewPage'), 'DNAOverviewPage');
const EnterpriseAuditPage = lazyPage(() => import('@/pages/EnterpriseAuditPage'), 'EnterpriseAuditPage');
const EnterpriseSLAPage = lazyPage(() => import('@/pages/EnterpriseSLAPage'), 'EnterpriseSLAPage');
const EnterpriseSupportPage = lazyPage(() => import('@/pages/EnterpriseSupportPage'), 'EnterpriseSupportPage');
const EvolutionPage = lazyPage(() => import('@/pages/EvolutionPage'), 'EvolutionPage');
const ExamPage = lazyPage(() => import('@/pages/ExamPage'), 'ExamPage');
const ExamResultsPage = lazyPage(() => import('@/pages/ExamResultsPage'), 'ExamResultsPage');
const ExecutionExplorerPage = lazyPage(() => import('@/pages/ExecutionExplorerPage'), 'ExecutionExplorerPage');
const FAQPage = lazyPage(() => import('@/pages/FAQPage'), 'FAQPage');
const FavoritesPage = lazyPage(() => import('@/pages/FavoritesPage'), 'FavoritesPage');
const FeaturesPage = lazyPage(() => import('@/pages/FeaturesPage'), 'FeaturesPage');
const FeedbackPage = lazyPage(() => import('@/pages/FeedbackPage'), 'FeedbackPage');
const ForbiddenPage = lazyPage(() => import('@/pages/ForbiddenPage'), 'ForbiddenPage');
const FRGEditorPage = lazyPage(() => import('@/pages/FRGEditorPage'), 'FRGEditorPage');
const FRGGraphsPage = lazyPage(() => import('@/pages/FRGGraphsPage'), 'FRGGraphsPage');
const FRGShowcasePage = lazyPage(() => import('@/pages/FRGShowcasePage'), 'FRGShowcasePage');
const FoundersPage = lazyPage(() => import('@/pages/FoundersPage'), 'FoundersPage');
const ProposalDetailPage = lazyPage(() => import('@/pages/FoundersPage/ProposalDetailPage'), 'ProposalDetailPage');
const FunctionDNAPage = lazyPage(() => import('@/pages/FunctionDNAPage'), 'FunctionDNAPage');
const FunctionMarketplacePage = lazyPage(() => import('@/pages/FunctionMarketplacePage'), 'FunctionMarketplacePage');
const FunctionPage = lazyPage(() => import('@/pages/FunctionPage'), 'FunctionPage');
const FunctionsDiscoveryPage = lazyPage(() => import('@/pages/FunctionsDiscoveryPage'), 'FunctionsDiscoveryPage');
const FunctionsPage = lazyPage(() => import('@/pages/FunctionsPage'), 'FunctionsPage');
const FunctionDetailPage = lazyPage(() => import('@/pages/FunctionsPage/FunctionDetailPage'), 'FunctionDetailPage');
const FunctionEditorPage = lazyPage(() => import('@/pages/FunctionsPage/FunctionEditorPage'), 'FunctionEditorPage');
const FunctionLogsPage = lazyPage(() => import('@/pages/FunctionsPage/FunctionLogsPage'), 'FunctionLogsPage');
const FunctionSettingsPage = lazyPage(() => import('@/pages/FunctionsPage/FunctionSettingsPage'), 'FunctionSettingsPage');
const GalleryPage = lazyPage(() => import('@/pages/GalleryPage'), 'GalleryPage');
const GitHubPage = lazyPage(() => import('@/pages/GitHubPage'), 'GitHubPage');
const GitHubRepoImportPage = lazyPage(() => import('@/pages/GitHubRepoImportPage'), 'GitHubRepoImportPage');
const HelpCenterPage = lazyPage(() => import('@/pages/HelpCenterPage'), 'HelpCenterPage');
const IntegrationsPage = lazyPage(() => import('@/pages/IntegrationsPage'), 'IntegrationsPage');
const LaunchPage = lazyPage(() => import('@/pages/LaunchPage'), 'LaunchPage');
const MyProfilePage = lazyPage(() => import('@/pages/MyProfilePage'), 'MyProfilePage');
const MyTeamPage = lazyPage(() => import('@/pages/MyTeamPage'), 'MyTeamPage');
const NotFoundPage = lazyPage(() => import('@/pages/NotFoundPage'), 'NotFoundPage');
const OnboardingPage = lazyPage(() => import('@/pages/OnboardingPage'), 'OnboardingPage');
const PlaygroundPage = lazyPage(() => import('@/pages/PlaygroundPage'), 'PlaygroundPage');
const ProfilePage = lazyPage(() => import('@/pages/ProfilePage/ProfilePage'), 'ProfilePage');
const ProfileSettingsPage = lazyPage(() => import('@/pages/ProfileSettingsPage'), 'ProfileSettingsPage');
const ProvidersPage = lazyPage(() => import('@/pages/ProvidersPage'), 'ProvidersPage');
const PublishRegistryFunctionPage = lazyPage(() => import('@/pages/PublishRegistryFunctionPage'), 'PublishRegistryFunctionPage');
const RegistryDeployPage = lazyPage(() => import('@/pages/RegistryDeployPage'), 'RegistryDeployPage');
const ReplayPage = lazyPage(() => import('@/pages/ReplayPage'), 'ReplayPage');
const SecretsPage = lazyPage(() => import('@/pages/SecretsPage'), 'SecretsPage');
const SecurityPage = lazyPage(() => import('@/pages/SecurityPage'), 'SecurityPage');
const ServerErrorPage = lazyPage(() => import('@/pages/ServerErrorPage'), 'ServerErrorPage');
const SettingsPage = lazyPage(() => import('@/pages/SettingsPage'), 'SettingsPage');
const StandalonePlaygroundPage = lazyPage(() => import('@/pages/StandalonePlaygroundPage'), 'StandalonePlaygroundPage');
const StateFabricMarketingPage = lazyPage(() => import('@/pages/StateFabricMarketingPage'), 'StateFabricMarketingPage');
const StateFabricPage = lazyPage(() => import('@/pages/StateFabricPage'), 'StateFabricPage');
const StateFabricDetailPage = lazyPage(() => import('@/pages/StateFabricPage/StateFabricDetailPage'), 'StateFabricDetailPage');
const StatePage = lazyPage(() => import('@/pages/StatePage'), 'StatePage');
const StateDetailPage = lazyPage(() => import('@/pages/StatePage/StateDetailPage'), 'StateDetailPage');
const TeamDecisionsPage = lazyPage(() => import('@/pages/TeamDecisionsPage'), 'TeamDecisionsPage');
const TeamMemoryPage = lazyPage(() => import('@/pages/TeamMemoryPage'), 'TeamMemoryPage');
const TeamsPage = lazyPage(() => import('@/pages/TeamsPage'), 'TeamsPage');
const ConsciousnessPage = lazyPage(() => import('@/pages/ConsciousnessPage'), 'ConsciousnessPage');
const TimeMachinePage = lazyPage(() => import('@/pages/TimeMachinePage'), 'TimeMachinePage');
const NewReplayPage = lazyPage(() => import('@/pages/TimeMachinePage/NewReplayPage'), 'NewReplayPage');
const ReplayDetailPage = lazyPage(() => import('@/pages/TimeMachinePage/ReplayDetailPage'), 'ReplayDetailPage');
const UserDashboardFunctionsPage = lazyPage(() => import('@/pages/UserDashboardFunctionsPage'), 'UserDashboardFunctionsPage');
const UserDashboardSettingsPage = lazyPage(() => import('@/pages/UserDashboardSettingsPage'), 'UserDashboardSettingsPage') as React.ComponentType<{ initialTab?: string }>;
const VaultPage = lazyPage(() => import('@/pages/VaultPage'), 'VaultPage');
const VerifyPage = lazyPage(() => import('@/pages/VerifyPage'), 'VerifyPage');
const WalletPage = lazyPage(() => import('@/pages/WalletPage'), 'WalletPage');
const AdaptiveUXPage = lazyPage(() => import('@/pages/AdaptiveUXPage'), 'AdaptiveUXPage');
const CodeIntelligencePage = lazyPage(() => import('@/pages/CodeIntelligencePage'), 'CodeIntelligencePage');
const CollaborationPage = lazyPage(() => import('@/pages/CollaborationPage'), 'CollaborationPage');
const DataVisualizationPage = lazyPage(() => import('@/pages/DataVisualizationPage'), 'DataVisualizationPage');
const DevOpsPage = lazyPage(() => import('@/pages/DevOpsPage'), 'DevOpsPage');
const FuturisticPage = lazyPage(() => import('@/pages/FuturisticPage'), 'FuturisticPage');
const MarketplaceEconomyPage = lazyPage(() => import('@/pages/MarketplaceEconomyPage'), 'MarketplaceEconomyPage');
const MCPCenterPage = lazyPage(() => import('@/pages/MCPCenterPage'), 'MCPCenterPage');
const MemoryPage = lazyPage(() => import('@/pages/MemoryPage'), 'MemoryPage');
const NotificationsPage = lazyPage(() => import('@/pages/NotificationsPage'), 'NotificationsPage');
const PasteCodePage = lazyPage(() => import('@/pages/PasteCodePage'), 'PasteCodePage');
const RoboticsPage = lazyPage(() => import('@/pages/RoboticsPage'), 'RoboticsPage');
const SimulationPage = lazyPage(() => import('@/pages/SimulationPage'), 'SimulationPage');
const StatusPage = lazyPage(() => import('@/pages/StatusPage'), 'StatusPage');
const StudioPage = lazyPage(() => import('@/pages/StudioPage'), 'StudioPage');
const UniversalRuntimePage = lazyPage(() => import('@/pages/UniversalRuntimePage'), 'UniversalRuntimePage');
const UsagePage = lazyPage(() => import('@/pages/UsagePage'), 'UsagePage');
const BillingHubPage = lazyPage(() => import('@/pages/BillingHubPage'), 'BillingHubPage');
const TrustAPIRegisterPage = lazyPage(() => import('@/pages/TrustAPIRegisterPage'), 'TrustAPIRegisterPage');

// Loading fallback for Suspense boundaries
function PageLoader() {
  return (
    <div
      className="flex min-h-[40vh] items-center justify-center"
      aria-busy="true"
      aria-label="Loading page"
    >
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );
}

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
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Public Routes */}
          <Route path="/" element={<HomeRedirect />} />
          <Route path="/launch" element={<LaunchPage />} />
          <Route path="/coming-soon" element={<LaunchPage />} />
          <Route path="/status" element={<StatusPage />} />
          <Route path="/pricing" element={<MarketingPricingRedirect />} />
          <Route path="/features" element={<FeaturesPage />} />
          <Route path="/integrations" element={<IntegrationsPage />} />
          {/* /teams is the dashboard-protected route inside DashboardLayout */}
          <Route path="/privacy" element={<MarketingLegalRedirect page="privacy" />} />
          {/* SecurityPage is routed inside DashboardLayout */}
          <Route path="/terms" element={<MarketingLegalRedirect page="terms" />} />
          <Route path="/changelog" element={<ChangelogPage />} />
          <Route path="/feedback" element={<FeedbackPage />} />
          <Route path="/faq" element={<FAQPage />} />
          <Route path="/help" element={<HelpCenterPage />} />
          <Route path="/contact" element={<ContactPage />} />
          <Route path="/docs" element={<DocsOutboundRedirect />} />
          <Route path="/docs/:slug" element={<DocsOutboundRedirect />} />
          <Route path="/blog" element={<MarketingBlogRedirect />} />
          <Route path="/blog/:slug" element={<MarketingBlogRedirect />} />
          <Route path="/products/state-fabric" element={<StateFabricMarketingPage />} />
          <Route path="/frg-showcase" element={<FRGShowcasePage />} />
          {/* Gallery is now inside DashboardLayout - see below */}
          <Route path="/registry" element={<BrowseFunctionsPage />} />
          <Route path="/registry/:author/:name" element={<RegistryFunctionRedirect />} />

          {/* Public user profile pages */}
          <Route path="/u/:username" element={<ProfilePage />} />
          <Route path="/profile/:username" element={<ProfilePage />} />

          {/* Public credential verification */}
          <Route path="/verify/:username" element={<VerifyPage />} />

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
          <Route path="/auth/magic-link" element={<RedirectToAuth />} />
          <Route path="/auth/magic-link/verify" element={<MagicLinkVerifyPage />} />

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
            <Route index element={<Navigate to="/overview" replace />} />
            <Route path="dashboard" element={<FunctionMarketplacePage />} />
            <Route path="overview" element={<DashboardPage />} />
            <Route path="notifications" element={<NotificationsPage />} />
            <Route path="apps" element={<AppsPage />} />
            <Route path="apps/new" element={<CreateAppPage />} />
            <Route path="apps/:slug" element={<AppDetailPage />} />
            <Route path="apps/:slug/bundle" element={<BundleConfigPage />} />
            <Route path="functions" element={<FunctionsPage />} />
            <Route path="functions/hot" element={<FunctionsDiscoveryPage />} />
            <Route path="functions/trending" element={<FunctionsDiscoveryPage />} />
            <Route path="functions/explore/new" element={<FunctionsDiscoveryPage />} />
            <Route path="functions/popular" element={<FunctionsDiscoveryPage />} />
            <Route path="functions/favorites" element={<FavoritesPage />} />
            <Route path="functions/my" element={<FunctionsDiscoveryPage />} />
            <Route path="functions/discovery" element={<Navigate to="/marketplace?type=functions" replace />} />
            <Route path="functions/discovery/:filter" element={<FunctionsDiscoveryPage />} />
            <Route path="gallery" element={<GalleryPage />} />
            <Route path="functions/new" element={<FunctionEditorPage />} />
            <Route path="functions/paste" element={<PasteCodePage />} />
            <Route path="functions/publish" element={<PublishRegistryFunctionPage />} />
            <Route path="functions/deploy" element={<RegistryDeployPage />} />
            <Route path="functions/:author/:name" element={<FunctionPage />} />
            <Route path="functions/:author/:name/settings" element={<FunctionSettingsPage />} />
            <Route path="functions/:author/:name/logs" element={<FunctionLogsPage />} />
            <Route path="functions/:id" element={<FunctionDetailPage />} />
            <Route path="functions/:id/edit" element={<FunctionEditorPage />} />
            <Route path="functions/:id/dna" element={<FunctionDNAPage />} />
            <Route path="dna/overview" element={<DNAOverviewPage />} />
            {/* AI Composer Routes - Multiple aliases for flexibility */}
            <Route path="ai-composer" element={<AIComposerPage />} />
            <Route path="composer" element={<AIComposerPage />} />
            <Route path="generate" element={<AIComposerPage />} />
            <Route path="functions/generate" element={<AIComposerPage />} />
            {/* AI Namespace - Future expansion routes */}
            <Route path="ai/composer" element={<AIComposerPage />} />
            <Route path="ai/chat" element={<AIComposerPage />} />
            <Route path="ai/suggest" element={<AIComposerPage />} />
            {/* Code Intelligence Routes */}
            <Route path="code-intelligence" element={<CodeIntelligencePage />} />
            <Route path="code-intelligence/:panel" element={<CodeIntelligencePage />} />
            <Route path="studio/code" element={<CodeIntelligencePage />} />
            {/* FRG (Function Runtime Graph) Routes */}
            <Route path="frg" element={<FRGGraphsPage />} />
            <Route path="frg/new" element={<FRGEditorPage />} />
            <Route path="frg/:author/:name" element={<FRGEditorPage />} />

            {/* GitHub Integration Routes */}
            <Route path="github" element={<GitHubPage />} />
            <Route path="github/import/:repoId" element={<GitHubRepoImportPage />} />

            <Route path="providers" element={<ProvidersPage />} />
            <Route path="analytics" element={<AnalyticsPage />} />
            <Route path="mcp" element={<MCPCenterPage />} />
            <Route path="usage" element={<UsagePage />} />
            <Route path="billing" element={<BillingHubPage />} />
            <Route path="state-fabric" element={<StateFabricPage />} />
            <Route path="state-fabric/new" element={<StateFabricDetailPage />} />
            <Route path="state-fabric/:id" element={<StateFabricDetailPage />} />
            <Route path="state-fabric/:id/edit" element={<StateFabricDetailPage />} />
            <Route path="bundles" element={<BundlePricingPage />} />
            <Route path="bundles/provisioning" element={<BundleProvisioningPage />} />
            <Route path="bundles/overview" element={<BundleOverviewPage />} />
            <Route path="bundles/functions" element={<BundleFunctionsPage />} />
            <Route path="bundles/integrations" element={<BundleIntegrationsPage />} />
            <Route path="bundles/integrations/:type" element={<BundleIntegrationDetailPage />} />
            <Route path="bundles/mine" element={<MyBundlesPage />} />
            <Route path="founders" element={<FoundersPage />} />
            <Route path="founders/votes/:id" element={<ProposalDetailPage />} />

            {/* Time Machine Routes */}
            <Route path="time-machine" element={<TimeMachinePage />} />
            <Route path="time-machine/new" element={<NewReplayPage />} />
            <Route path="time-machine/:id" element={<ReplayDetailPage />} />

            {/* Function Consciousness Routes */}
            <Route path="consciousness" element={<ConsciousnessPage />} />

            {/* Certification Routes */}
            <Route path="certification" element={<CertificationPage />} />
            <Route path="certification/exam/:examId" element={<ExamPage />} />
            <Route path="certification/exam/:examId/results" element={<ExamResultsPage />} />
            <Route path="credentials" element={<CredentialsPage />} />

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
            <Route path="vault" element={<VaultPage />} />
            <Route path="connectors/callback" element={<ConnectorsCallbackPage />} />
            <Route path="connectors" element={<Navigate to="/settings#integrations" replace />} />
            <Route path="brain" element={<BrainPage />} />
            <Route path="teams" element={<TeamsPage />} />
            <Route path="my-team" element={<MyTeamPage />} />
            <Route path="teams/:teamId/memory" element={<TeamMemoryPage />} />
            <Route path="decisions" element={<DecisionsPage />} />
            <Route path="teams/:teamId/decisions" element={<TeamDecisionsPage />} />
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
            <Route path="agents/new" element={<AgentCreatePage />} />
            <Route path="agents/:id" element={<AgentDetailPage />} />
            <Route path="agents/:id/edit" element={<AgentEditPage />} />
            <Route path="agents/:id/wallet" element={<AgentWalletPage />} />
            <Route path="agents/:id/analytics" element={<AgentAnalyticsPage />} />
            <Route path="agent-observability" element={<AgentObservabilityPage />} />
            <Route path="sdk-integrations" element={<AgentSDKIntegrationsPage />} />
            <Route path="marketplace" element={<MarketplacePage />} />
            <Route path="marketplace/agents/:id" element={<AgentMarketplaceDetailPage />} />
            <Route path="marketplace/extensions/:id" element={<ExtensionDetailPage />} />
            <Route path="marketplace/agents" element={<Navigate to="/marketplace?type=agents" replace />} />
            <Route path="wallet" element={<WalletPage />} />
            <Route path="wallet/agents/:id" element={<WalletPage />} />
            <Route path="wallet/:slug" element={<WalletPage />} />
            <Route path="evolution" element={<EvolutionPage />} />
            <Route path="evolution/:slug" element={<EvolutionPage />} />
            <Route path="conversations" element={<ConversationsPage />} />
            <Route path="conversations/:id" element={<ConversationsPage />} />
            <Route path="community" element={<CommunityPage />} />
            <Route path="community/bookmarks" element={<CommunityBookmarksPage />} />
            <Route path="community/notifications" element={<CommunityNotificationsPage />} />
            <Route path="community/user/:userId" element={<CommunityUserProfilePage />} />
            <Route path="community/:postId" element={<CommunityThreadPage />} />
          </Route>

          {/* Agent Workspace - Outside DashboardLayout for fullscreen */}
          <Route
            path="agents/:id/workspace"
            element={
              <ProtectedRoute>
                <AgentWorkspacePage />
              </ProtectedRoute>
            }
          />

          {/* Studio Route - Outside DashboardLayout for fullscreen */}
          <Route
            path="studio"
            element={
              <ProtectedRoute>
                <StudioPage />
              </ProtectedRoute>
            }
          />

          {/* Trust API Registration - Standalone page outside DashboardLayout */}
          <Route
            path="trust-api/register"
            element={
              <ProtectedRoute>
                <TrustAPIRegisterPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="studio/:environment"
            element={
              <ProtectedRoute>
                <StudioPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="devops"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DevOpsPage />} />
            <Route path=":panel" element={<DevOpsPage />} />
          </Route>

          <Route
            path="security"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<SecurityPage />} />
            <Route path=":panel" element={<SecurityPage />} />
          </Route>

          <Route
            path="collaboration"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<CollaborationPage />} />
            <Route path=":panel" element={<CollaborationPage />} />
          </Route>
          <Route
            path="memory"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<MemoryPage />} />
            <Route path=":panel" element={<MemoryPage />} />
          </Route>

          <Route
            path="simulation"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<SimulationPage />} />
            <Route path=":panel" element={<SimulationPage />} />
          </Route>

          <Route
            path="robotics"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<RoboticsPage />} />
            <Route path=":panel" element={<RoboticsPage />} />
          </Route>

          <Route
            path="marketplace-economy"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<MarketplaceEconomyPage />} />
            <Route path=":panel" element={<MarketplaceEconomyPage />} />
          </Route>

          <Route
            path="adaptive-ux"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<AdaptiveUXPage />} />
            <Route path=":panel" element={<AdaptiveUXPage />} />
          </Route>

          <Route
            path="universal-runtime"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<UniversalRuntimePage />} />
            <Route path=":panel" element={<UniversalRuntimePage />} />
          </Route>

          <Route
            path="data-visualization"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DataVisualizationPage />} />
            <Route path=":panel" element={<DataVisualizationPage />} />
          </Route>


          <Route
            path="futuristic"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<FuturisticPage />} />
            <Route path=":panel" element={<FuturisticPage />} />
          </Route>

          <Route
            path="u/:username"
            element={
              <ProtectedRoute>
                <UserProfileLayout />
              </ProtectedRoute>
            }
          >
            <Route path="agents" element={<AgentsPage />} />
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
      </Suspense>
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
              <NuqsAdapter>
                <HelmetProvider>
                  <Analytics />
                  <PageViewTracker />
                  <GlobalKeyboardShortcuts />
                  <AppContent />
                </HelmetProvider>
              </NuqsAdapter>
            </BrowserRouter>
            <ThemeAwareToaster />
          </CookieConsentProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;

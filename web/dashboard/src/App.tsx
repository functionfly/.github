import { useEffect } from "react";
import { BrowserRouter, Routes, Route, Navigate, Outlet, useParams } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { useAuthStore } from "@/stores/authStore";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { CookieConsentProvider } from "@/components/cookie-consent";
import { LandingPage } from "@/pages/LandingPage";
import { TeamPage } from "@/pages/TeamPage";
import { PricingPage } from "@/pages/PricingPage";
import { FeaturesPage } from "@/pages/FeaturesPage";
import { IntegrationsPage } from "@/pages/IntegrationsPage";
import { AuthPage } from "@/pages/AuthPage";
import { VerifyEmailPage } from "@/pages/AuthPage/VerifyEmailPage";
import { OAuthCallback } from "@/pages/AuthPage/OAuthCallback";
import { OnboardingPage } from "@/pages/OnboardingPage";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { DashboardPage } from "@/pages/DashboardPage";
import { FunctionsPage } from "@/pages/FunctionsPage";
import { FunctionEditorPage } from "@/pages/FunctionsPage/FunctionEditorPage";
import { FunctionDetailPage } from "@/pages/FunctionsPage/FunctionDetailPage";
import { ProvidersPage } from "@/pages/ProvidersPage";
import { AnalyticsPage } from "@/pages/AnalyticsPage";
import { SettingsPage } from "@/pages/SettingsPage";
import { AdminTenantsPage } from "@/pages/AdminTenantsPage";
import { AdminUsersPage } from "@/pages/AdminUsersPage";
import { AdminBillingPage } from "@/pages/AdminBillingPage";
import { AdminAuditPage } from "@/pages/AdminAuditPage";
import { AdminSystemPage } from "@/pages/AdminSystemPage";
import { AdminRedirectsPage } from "@/pages/AdminRedirectsPage";
import { AdminNewsletterPage } from "@/pages/AdminNewsletterPage";
import { AdminContentCalendarPage } from "@/pages/AdminContentCalendarPage";
import { AdminFeedbackPage } from "@/pages/AdminFeedbackPage";
import { AdminDashboardPage } from "@/pages/AdminDashboardPage";
import { AdminFunctionsPage } from "@/pages/AdminFunctionsPage";
import { AdminRegistryPage } from "@/pages/AdminRegistryPage";
import { PrivacyPage } from "@/pages/PrivacyPage";
import { SecurityPage } from "@/pages/SecurityPage";
import ChangelogPage from "@/pages/ChangelogPage";
import { FeedbackPage } from "@/pages/FeedbackPage";
import { FAQPage } from "@/pages/FAQPage";
import BlogPage from "@/pages/BlogPage";
import BlogPostPage from "@/pages/BlogPostPage";
import AdminContentPage from "@/pages/AdminContentPage";
import { ContactPage } from "@/pages/ContactPage";
import { PasswordResetPage } from "@/pages/AuthPage/PasswordResetPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { GlobalKeyboardShortcuts } from "@/components/common/GlobalKeyboardShortcuts";
import { Analytics } from "@/components/common/Analytics";
import { ThemeProvider } from "@/components/common/ThemeProvider";
import FunctionPage from "@/pages/FunctionPage";
import { PlaygroundPage } from "@/pages/PlaygroundPage";
import { ReplayPage } from "@/pages/ReplayPage";
import { StateFabricPage } from "@/pages/StateFabricPage";
import { StateFabricDetailPage } from "@/pages/StateFabricPage/StateFabricDetailPage";
import { AdminStateFabricPage } from "@/pages/AdminStateFabricPage";
import { StateFabricMarketingPage } from "@/pages/StateFabricMarketingPage";
import { BrowseFunctionsPage } from "@/pages/BrowseFunctionsPage";
import RegistryDeployPage from "@/pages/RegistryDeployPage";
import { DocsPage } from "@/pages/DocsPage";

function RegistryFunctionRedirect() {
  const { author, name } = useParams<{ author: string; name: string }>();
  if (!author || !name) return <Navigate to="/registry" replace />;
  return <Navigate to={`/fx/${author}/${name}`} replace />;
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
  const user = useAuthStore((state) => state.user);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check if user is an admin (skip onboarding for admins)
  const isAdmin = user?.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);
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
  const isAdmin = user?.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);
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
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  const isAdmin = user?.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);
  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return !isAuthenticated ? <>{children}</> : <Navigate to="/dashboard" replace />;
}

function AppContent() {
  const initialize = useAuthStore((state) => state.initialize);

  useEffect(() => {
    initialize();
  }, [initialize]);

  return (
    <Routes>
      {/* Public Routes */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/pricing" element={<PricingPage />} />
      <Route path="/features" element={<FeaturesPage />} />
      <Route path="/integrations" element={<IntegrationsPage />} />
      <Route path="/team" element={<TeamPage />} />
      <Route path="/privacy" element={<PrivacyPage />} />
      <Route path="/security" element={<SecurityPage />} />
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
      <Route
        path="/registry/:author/:name"
        element={<RegistryFunctionRedirect />}
      />

      {/* Registry Playground Routes (Public) */}
      <Route path="/fx/:author/:name" element={<FunctionPage />} />
      <Route path="/run/:author/:name" element={<PlaygroundPage />} />
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
        <Route path="functions" element={<FunctionsPage />} />
        <Route path="functions/new" element={<FunctionEditorPage />} />
        <Route path="functions/deploy" element={<RegistryDeployPage />} />
        <Route path="functions/:id" element={<FunctionDetailPage />} />
        <Route path="functions/:id/edit" element={<FunctionEditorPage />} />
        <Route path="providers" element={<ProvidersPage />} />
        <Route path="analytics" element={<AnalyticsPage />} />
        <Route path="state-fabric" element={<StateFabricPage />} />
        <Route path="state-fabric/new" element={<StateFabricDetailPage />} />
        <Route path="state-fabric/:id" element={<StateFabricDetailPage />} />
        <Route path="state-fabric/:id/edit" element={<StateFabricDetailPage />} />
        <Route path="settings" element={<SettingsPage />} />

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
          <Route path="users" element={<AdminUsersPage />} />
          <Route path="billing" element={<AdminBillingPage />} />
          <Route path="audit" element={<AdminAuditPage />} />
          <Route path="system" element={<AdminSystemPage />} />
          <Route path="redirects" element={<AdminRedirectsPage />} />
          <Route path="newsletter" element={<AdminNewsletterPage />} />
          <Route path="content-calendar" element={<AdminContentCalendarPage />} />
          <Route path="content" element={<AdminContentPage />} />
          <Route path="feedback" element={<AdminFeedbackPage />} />
          <Route path="functions" element={<AdminFunctionsPage />} />
          <Route path="registry" element={<AdminRegistryPage />} />
          <Route path="state-fabric" element={<AdminStateFabricPage />} />
        </Route>
      </Route>

      {/* 404 - Not Found */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <CookieConsentProvider>
          <Analytics />
          <BrowserRouter>
            <GlobalKeyboardShortcuts />
            <AppContent />
          </BrowserRouter>
          <Toaster
            position="bottom-right"
            toastOptions={{
              style: {
                background: "var(--bg-secondary)",
                border: "1px solid var(--border-subtle)",
                color: "var(--text-primary)",
              },
            }}
          />
        </CookieConsentProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;

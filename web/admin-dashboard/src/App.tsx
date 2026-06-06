/**
 * Main App Component
 * Sets up routing for admin dashboard
 *
 * Route definitions live in @/routes/adminRoutes.tsx — that file is the
 * single source of truth for the admin SPA. App.tsx only wires up
 * providers, the auth-restore gate, the layout, and the public/auth
 * routes. To add a new admin page, edit adminRoutes.tsx.
 */

import { AdminAuthRestore } from '@/components/auth/AdminAuthRestore';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/AdminLayout';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import { ToastProvider } from '@/components/ui/Toast';
import { AdminAccessDeniedPage } from '@/pages/AdminAccessDeniedPage';
import { AdminLoginPage } from '@/pages/AdminLoginPage';
import { renderAdminRoutes } from '@/routes/adminRoutes';

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
      <ToastProvider>
        <BrowserRouter>
          <AdminAuthRestore>
            <Routes>
              {/* All admin routes use the dashboard layout (sidebar + header) */}
              <Route element={<AdminLayout />}>
                {/* Public: login page inside same layout */}
                <Route path="/auth/login" element={<AdminLoginPage />} />
                <Route path="/auth/*" element={<Navigate to="/auth/login" replace />} />

                {/* Protected admin routes — auth gate */}
                <Route element={<ProtectedRoute />}>
                  {/*
                    Source of truth for the admin SPA's protected routes is
                    @/routes/adminRoutes.tsx. Each entry there wraps its
                    component in <AdminPage> which handles per-permission
                    checks and redirects to /access-denied.
                   */}
                  {renderAdminRoutes()}

                  {/* Access denied lives inside the layout so the user can
                      still navigate back, and inside the protected gate so
                      unauthenticated users can't reach it. */}
                  <Route path="access-denied" element={<AdminAccessDeniedPage />} />
                </Route>
              </Route>

              {/* Catch-all */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </AdminAuthRestore>
        </BrowserRouter>
      </ToastProvider>
    </QueryClientProvider>
  );
}

export default App;

import { useAuthStore } from '@/stores/authStore';
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import type { ReactNode } from 'react';
import { LoginPage } from '@/pages/LoginPage';

interface AuthGuardProps {
  children: ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { isAuthenticated, isLoading, initialize } = useAuthStore();
  const location = useLocation();

  useEffect(() => {
    initialize();
  }, [initialize]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-950">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
      </div>
    );
  }

  if (!isAuthenticated && location.pathname !== '/login') {
    return <LoginPage />;
  }

  return <>{children}</>;
}

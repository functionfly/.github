/**
 * Admin Layout Component
 * Main layout for admin dashboard with navigation
 */

import { useState, useEffect } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAdminAuth } from '@/hooks/useAdminAuth';
import { AdminSidebar } from './AdminSidebar';
import { AdminHeader } from './AdminHeader';
import { SessionTimeoutWarning } from '@/components/security/SessionTimeout';
import { MFAReVerificationChecker } from '@/components/security/MFAReVerification';
import { useSessionMonitor } from '@/hooks/useSessionMonitor';

export function AdminLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { user, logout } = useAdminAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // Monitor session
  useSessionMonitor();

  useEffect(() => {
    // Close sidebar on route change
    setSidebarOpen(false);
  }, [location.pathname]);

  const handleLogout = () => {
    logout();
    navigate('/auth/login');
  };

  const isLoggedIn = !!user;

  return (
    <MFAReVerificationChecker>
      <div className="flex h-screen bg-gray-50">
        {/* Sidebar: only when logged in */}
        {isLoggedIn && (
          <AdminSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        )}

        {/* Main content */}
        <div className="flex-1 flex flex-col overflow-hidden min-w-0">
          {/* Header */}
          <AdminHeader
            user={user}
            onMenuClick={() => setSidebarOpen(!sidebarOpen)}
            onLogout={handleLogout}
            showMenuButton={isLoggedIn}
          />

          {/* Content: centered when logged out (login page), normal padding when logged in */}
          <main
            className={`flex-1 overflow-y-auto ${!isLoggedIn ? 'flex items-center justify-center p-4 sm:p-6' : ''}`}
          >
            <div className={isLoggedIn ? 'max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-8' : 'w-full max-w-md'}>
              <Outlet />
            </div>
          </main>
        </div>

        {/* Session timeout warning */}
        <SessionTimeoutWarning />
      </div>
    </MFAReVerificationChecker>
  );
}

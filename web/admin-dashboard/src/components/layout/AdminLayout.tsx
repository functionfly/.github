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

  return (
    <MFAReVerificationChecker>
      <div className="flex h-screen bg-gray-50">
        {/* Sidebar */}
        <AdminSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />

        {/* Main content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Header */}
          <AdminHeader
            user={user}
            onMenuClick={() => setSidebarOpen(!sidebarOpen)}
            onLogout={handleLogout}
          />

          {/* Content */}
          <main className="flex-1 overflow-y-auto">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-8">
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

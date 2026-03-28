import { Navbar } from '@/components/common/Navbar';
import { SupportBubble, SupportChatProvider, UnifiedChatWindow } from '@/components/support';
import { Footer } from '@/pages/LandingPage/components';
import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';

export function DashboardLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <SupportChatProvider>
      <div
        className="min-h-screen bg-bg-primary flex flex-row mesh-gradient-bg dashboard-enhanced"
        style={{ backgroundColor: 'var(--bg-primary)' }}
      >
        {/* Background Effects */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-[128px] animate-float" />
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[128px] animate-float-rotate" />
          <div className="spotlight-container">
            <div className="spotlight-bg animate-spotlight" />
          </div>
        </div>

        {/* Sidebar - min-h-screen ensures it extends full height */}
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />

        {/* Main Content */}
        <div className="flex-1 flex flex-col min-w-0 relative">
          <Navbar variant="dashboard" onMenuClick={() => setSidebarOpen(true)} />

          <main className="flex-1 pt-20 lg:pt-24 p-4 lg:p-6" aria-label="Main content">
            <div className="max-w-7xl mx-auto">
              <Outlet />
            </div>
          </main>

          <Footer showScrollToTop={false} />

          {/* Unified AI + support chat - bottom right */}
          <SupportBubble />
          <UnifiedChatWindow />
        </div>
      </div>
    </SupportChatProvider>
  );
}

import { ErrorBoundary, SectionErrorBoundary } from '@/components/common/ErrorBoundary';
import { Navbar } from '@/components/common/Navbar';
import { Sidebar } from '@/components/layout/Sidebar';
import { useSidebarStore } from '@/stores/sidebarStore';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { useCallback, useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';

export function UserProfileLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { isCollapsed, setCollapsed } = useSidebarStore();

  // Swipe to open sidebar on desktop (left edge swipe)
  const { gestureHandlers: openGestureHandlers } = useSwipeGesture({
    onSwipeRight: () => {
      if (window.innerWidth >= 1024) {
        setSidebarOpen(true);
      }
    },
  });

  // Handle keyboard shortcut to toggle sidebar
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault();
      setCollapsed(!isCollapsed);
    }
  }, [isCollapsed, setCollapsed]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div
      className="min-h-screen flex flex-col relative mesh-gradient-bg"
      {...openGestureHandlers}
    >
      {/* Background Effects */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-[128px] animate-float" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[128px] animate-float-rotate" />
      </div>

      {/* Navbar spans full width above sidebar and main content */}
      <Navbar variant="dashboard" className="sticky top-0 z-50 shrink-0" onMenuClick={() => setSidebarOpen(true)} />

      {/* Content area: sidebar + main */}
      <div className="flex flex-1 min-h-0 relative">
        {/* Sidebar sits below the navbar on desktop */}
        <Sidebar
          isOpen={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
        />

        {/* Main Content */}
        <div className="flex-1 flex flex-col min-w-0 relative dashboard-main-bg transition-all duration-300 ease-in-out">

          <main className="flex-1 p-4 lg:p-6" aria-label="Main content">
            <div className="max-w-7xl mx-auto">
              {/* ErrorBoundary wraps each profile page to prevent one failing component from crashing the entire section */}
              <ErrorBoundary
                fallback={
                  <div className="min-h-[40vh] flex items-center justify-center">
                    <div className="text-center space-y-4">
                      <p className="text-lg font-medium">This page encountered an error</p>
                      <p className="text-sm text-muted-foreground">Try refreshing the page</p>
                    </div>
                  </div>
                }
              >
                <SectionErrorBoundary sectionName="Profile page">
                  <Outlet />
                </SectionErrorBoundary>
              </ErrorBoundary>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}

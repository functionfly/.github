import { ErrorBoundary, SectionErrorBoundary } from '@/components/common/ErrorBoundary';
import { Navbar } from '@/components/common/Navbar';
import { SupportBubble, SupportChatProvider, UnifiedChatWindow } from '@/components/support';
import { Footer } from '@/pages/LandingPage/components';
import { useSidebarStore } from '@/stores/sidebarStore';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { useEffect, useState, useCallback } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';

export function DashboardLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [showSwipeIndicator, setShowSwipeIndicator] = useState(false);
  const { isCollapsed, setCollapsed } = useSidebarStore();

  // Swipe to open sidebar on desktop (left edge swipe)
  const { gestureHandlers: openGestureHandlers } = useSwipeGesture({
    onSwipeRight: () => {
      // Only trigger if near left edge (first 20% of screen)
      if (window.innerWidth >= 1024) {
        setSidebarOpen(true);
      }
    },
  });

  // Handle keyboard shortcut to toggle sidebar
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Cmd/Ctrl + B to toggle sidebar
    if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault();
      setCollapsed(!isCollapsed);
    }
    // Cmd/Ctrl + Shift + S to open sidebar on mobile
    if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'S') {
      e.preventDefault();
      setSidebarOpen(true);
    }
  }, [isCollapsed, setCollapsed]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  // Show swipe indicator briefly on mobile
  useEffect(() => {
    if (window.innerWidth < 1024) {
      const timer = setTimeout(() => {
        setShowSwipeIndicator(true);
        const hideTimer = setTimeout(() => setShowSwipeIndicator(false), 3000);
        return () => clearTimeout(hideTimer);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, []);

  return (
    <SupportChatProvider>
      {/* Swipe Indicator for Mobile */}
      <div
        className={`aviation-swipe-indicator ${showSwipeIndicator ? 'visible' : ''} lg:hidden`}
        onClick={() => setSidebarOpen(true)}
      />

      <div
        className="min-h-screen flex flex-row relative mesh-gradient-bg"
        {...openGestureHandlers}
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
        <Sidebar
          isOpen={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
        />

        {/* Main Content - No margin needed on desktop since sidebar is in flex flow */}
        <div className="flex-1 flex flex-col min-w-0 relative dashboard-main-bg transition-all duration-300 ease-in-out">
          <Navbar variant="dashboard" onMenuClick={() => setSidebarOpen(true)} />

          <main className="flex-1 pt-20 lg:pt-24 p-4 lg:p-6" aria-label="Main content">
            <div className="max-w-7xl mx-auto">
              {/* ErrorBoundary wraps each dashboard page to prevent one failing component from crashing the entire dashboard */}
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
                <SectionErrorBoundary sectionName="Dashboard page">
                  <Outlet />
                </SectionErrorBoundary>
              </ErrorBoundary>
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

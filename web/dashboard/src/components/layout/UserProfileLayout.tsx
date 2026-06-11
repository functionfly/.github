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
      className="min-h-screen flex flex-row relative mesh-gradient-bg"
      {...openGestureHandlers}
    >
      {/* Background Effects */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-[128px] animate-float" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[128px] animate-float-rotate" />
      </div>

      {/* Sidebar */}
      <Sidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0 relative dashboard-main-bg transition-all duration-300 ease-in-out">
        <Navbar variant="dashboard" onMenuClick={() => setSidebarOpen(true)} />

        <main className="flex-1 pt-20 lg:pt-24 p-4 lg:p-6" aria-label="Main content">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}

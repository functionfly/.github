import { useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { NavLink, useLocation } from "react-router-dom";
import {
  X,
  Menu,
  Home,
  LayoutDashboard,
  FunctionSquare,
  Cloud,
  BarChart3,
  Settings,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { ADMIN_DASHBOARD_URL, ROUTES } from "@/lib/constants";
import { useSwipeGesture } from "@/hooks/useSwipeGesture";
import { useKeyboardNavigation } from "@/hooks/useKeyboardNavigation";

interface NavItem {
  path: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface MobileNavProps {
  className?: string;
}

const mainNavItems: NavItem[] = [
  { path: ROUTES.DASHBOARD, label: "Marketplace", icon: Home },
  { path: ROUTES.OVERVIEW, label: "Overview", icon: LayoutDashboard },
  { path: ROUTES.FUNCTIONS, label: "Functions", icon: FunctionSquare },
  { path: ROUTES.PROVIDERS, label: "Providers", icon: Cloud },
  { path: ROUTES.ANALYTICS, label: "Analytics", icon: BarChart3 },
  { path: ROUTES.SETTINGS, label: "Settings", icon: Settings },
];

export function MobileNav({ className }: MobileNavProps) {
  const [isOpen, setIsOpen] = useState(false);
  const location = useLocation();
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === 'admin';

  const adminNavItems: NavItem[] =
    isAdmin && ADMIN_DASHBOARD_URL
      ? [{ path: ADMIN_DASHBOARD_URL, label: "Admin Panel", icon: Shield }]
      : [];

  const allNavItems = [...mainNavItems, ...adminNavItems];

  // Swipe gesture for closing
  const { gestureHandlers } = useSwipeGesture({
    onSwipeLeft: () => setIsOpen(false),
  });

  // Keyboard navigation
  const [focusedIndex, setFocusedIndex] = useState(-1);

  const handleArrowUp = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex(prev => prev <= 0 ? allNavItems.length - 1 : prev - 1);
  }, [isOpen, allNavItems.length]);

  const handleArrowDown = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex(prev => prev >= allNavItems.length - 1 ? 0 : prev + 1);
  }, [isOpen, allNavItems.length]);

  const handleEnter = useCallback(() => {
    if (!isOpen || focusedIndex < 0 || focusedIndex >= allNavItems.length) return;
    const item = allNavItems[focusedIndex];
    if (item) {
      setIsOpen(false);
      window.location.href = item.path;
    }
  }, [isOpen, focusedIndex, allNavItems]);

  const handleEscape = useCallback(() => {
    if (isOpen) {
      setIsOpen(false);
      setFocusedIndex(-1);
    }
  }, [isOpen]);

  useKeyboardNavigation({
    enabled: isOpen,
    onArrowUp: handleArrowUp,
    onArrowDown: handleArrowDown,
    onEnter: handleEnter,
    onEscape: handleEscape,
  });

  const toggleNav = () => {
    setIsOpen(!isOpen);
    if (!isOpen) {
      setFocusedIndex(-1);
    }
  };

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={toggleNav}
        className={cn(
          "lg:hidden p-2 rounded-lg text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-colors",
          className
        )}
        aria-label="Toggle navigation menu"
      >
        <Menu className="w-6 h-6" />
      </button>

      {/* Mobile Navigation Drawer */}
      <AnimatePresence>
        {isOpen && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="fixed inset-0 bg-black/50 z-50 lg:hidden"
              onClick={() => setIsOpen(false)}
            />

            {/* Drawer */}
            <motion.nav
              {...gestureHandlers}
              initial={{ x: "-100%" }}
              animate={{ x: 0 }}
              exit={{ x: "-100%" }}
              transition={{
                type: "spring",
                stiffness: 300,
                damping: 30,
                mass: 0.8
              }}
              className="fixed left-0 top-0 z-50 h-full w-80 bg-bg-secondary border-r border-border-subtle lg:hidden"
            >
              {/* Header */}
              <div className="flex items-center justify-between p-4 border-b border-border-subtle">
                <h2 className="text-lg font-semibold text-text-primary">Navigation</h2>
                <button
                  onClick={() => setIsOpen(false)}
                  className="p-2 rounded-lg hover:bg-bg-hover text-text-secondary hover:text-text-primary transition-colors"
                  aria-label="Close navigation menu"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Navigation Items */}
              <div className="py-4">
                {allNavItems.map((item, index) => {
                  const isExternal = item.path.startsWith('http');
                  const isActive =
                    !isExternal &&
                    (location.pathname === item.path ||
                      (item.path !== ROUTES.DASHBOARD &&
                        item.path !== ROUTES.OVERVIEW &&
                        location.pathname.startsWith(item.path)));
                  const isFocused = focusedIndex === index;
                  const Icon = item.icon;

                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      onClick={() => setIsOpen(false)}
                      className={cn(
                        "flex items-center gap-3 px-4 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200",
                        "hover:bg-bg-hover",
                        isActive
                          ? "text-brand-500 bg-brand-500/10 border-l-3 border-brand-500"
                          : isFocused
                          ? "text-text-primary bg-bg-hover ring-2 ring-border-focus"
                          : "text-text-secondary hover:text-text-primary"
                      )}
                    >
                      <Icon className={cn(
                        "w-5 h-5",
                        isActive ? "text-brand-500" : "text-text-muted"
                      )} />
                      <span>{item.label}</span>
                    </NavLink>
                  );
                })}
              </div>
            </motion.nav>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
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
  Store,
  Bot,
  CreditCard,
  ShoppingBag,
  Settings,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { ADMIN_DASHBOARD_URL, DOCS_SITE_URL, ROUTES } from "@/lib/constants";
import { isPlatformAdminRole } from "@/lib/platform-admin";
import { useSwipeGesture } from "@/hooks/useSwipeGesture";
import { useKeyboardNavigation } from "@/hooks/useKeyboardNavigation";

interface NavItem {
  path: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  section?: string;
  external?: boolean;
}

interface MobileNavProps {
  className?: string;
}

// Main nav items for authenticated users
const mainNavItems: NavItem[] = [
  { path: ROUTES.DASHBOARD, label: "Dashboard", icon: LayoutDashboard, section: "main" },
  { path: "/functions/my", label: "Functions", icon: FunctionSquare, section: "main" },
  { path: "/functions/discovery", label: "Browse Functions", icon: Store, section: "marketplace" },
  { path: "/marketplace/agents", label: "Browse Agents", icon: Bot, section: "marketplace" },
  { path: "/marketplace/purchases", label: "My Purchases", icon: ShoppingBag, section: "marketplace" },
  { path: ROUTES.PROVIDERS, label: "Connected Providers", icon: Cloud, section: "providers" },
  { path: "/providers/billing", label: "Usage & Billing", icon: CreditCard, section: "providers" },
  { path: ROUTES.SETTINGS, label: "Settings", icon: Settings, section: "settings" },
];

// Unauthenticated nav items
const unauthNavItems: NavItem[] = [
  { path: ROUTES.HOME, label: "Home", icon: Home, external: true },
  { path: "/functions/discovery", label: "Functions", icon: FunctionSquare },
  { path: "/pricing", label: "Pricing", icon: LayoutDashboard },
  { path: DOCS_SITE_URL, label: "Docs", icon: FunctionSquare, external: true },
];

export function MobileNav({ className }: MobileNavProps) {
  const [isOpen, setIsOpen] = useState(false);
  const location = useLocation();
  const user = useAuthStore((state) => state.user);
  const isAdmin = isPlatformAdminRole(user?.role);

  const adminNavItems: NavItem[] =
    isAdmin && ADMIN_DASHBOARD_URL
      ? [{ path: ADMIN_DASHBOARD_URL, label: "Admin Panel", icon: Shield }]
      : [];

  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const allNavItems = isAuthenticated 
    ? [...mainNavItems, ...adminNavItems] 
    : unauthNavItems;

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
      if (item.external) {
        window.open(item.path, '_blank');
      } else {
        window.location.href = item.path;
      }
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

  // Group items by section
  const mainItems = allNavItems.filter(item => !item.section || item.section === "main");
  const marketplaceItems = allNavItems.filter(item => item.section === "marketplace");
  const providerItems = allNavItems.filter(item => item.section === "providers");
  const settingsItems = allNavItems.filter(item => item.section === "settings");

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
              <div className="py-4 overflow-y-auto">
                {/* Main Section */}
                {mainItems.length > 0 && (
                  <div className="px-3 mb-4">
                    <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mb-2">
                      Menu
                    </p>
                    {mainItems.map((item) => {
                      const isActive = location.pathname === item.path ||
                        (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path));
                      const absoluteIndex = allNavItems.indexOf(item);
                      const isFocused = focusedIndex === absoluteIndex;
                      const Icon = item.icon;

                      return (
                        <NavLink
                          key={item.path}
                          to={item.external ? '#' : item.path}
                          onClick={(e) => {
                            if (item.external) {
                              e.preventDefault();
                              window.open(item.path, '_blank');
                            }
                            setIsOpen(false);
                          }}
                          className={cn(
                            "flex items-center gap-3 px-3 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200",
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
                )}

                {/* Marketplace Section */}
                {marketplaceItems.length > 0 && (
                  <div className="px-3 mb-4">
                    <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mb-2">
                      Marketplace
                    </p>
                    {marketplaceItems.map((item) => {
                      const isActive = location.pathname === item.path;
                      const absoluteIndex = allNavItems.indexOf(item);
                      const isFocused = focusedIndex === absoluteIndex;
                      const Icon = item.icon;

                      return (
                        <NavLink
                          key={item.path}
                          to={item.path}
                          onClick={() => setIsOpen(false)}
                          className={cn(
                            "flex items-center gap-3 px-3 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200",
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
                )}

                {/* Providers Section */}
                {providerItems.length > 0 && (
                  <div className="px-3 mb-4">
                    <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mb-2">
                      Providers
                    </p>
                    {providerItems.map((item) => {
                      const isActive = location.pathname === item.path;
                      const absoluteIndex = allNavItems.indexOf(item);
                      const isFocused = focusedIndex === absoluteIndex;
                      const Icon = item.icon;

                      return (
                        <NavLink
                          key={item.path}
                          to={item.path}
                          onClick={() => setIsOpen(false)}
                          className={cn(
                            "flex items-center gap-3 px-3 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200",
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
                )}

                {/* Settings Section */}
                {settingsItems.length > 0 && (
                  <div className="px-3">
                    <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mb-2">
                      Account
                    </p>
                    {settingsItems.map((item) => {
                      const isActive = location.pathname === item.path;
                      const Icon = item.icon;

                      return (
                        <NavLink
                          key={item.path}
                          to={item.path}
                          onClick={() => setIsOpen(false)}
                          className={cn(
                            "flex items-center gap-3 px-3 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200",
                            "hover:bg-bg-hover",
                            isActive
                              ? "text-brand-500 bg-brand-500/10 border-l-3 border-brand-500"
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
                )}
              </div>
            </motion.nav>
          </>
        )}
      </AnimatePresence>
    </>
  );
}

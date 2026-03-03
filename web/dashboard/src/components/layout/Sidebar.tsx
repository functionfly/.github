import { useState, useCallback, useEffect } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  LayoutDashboard,
  Package,
  Cloud,
  BarChart3,
  Settings,
  LogOut,
  X,
  ChevronDown,
  Search,
  Database,
  FunctionSquare,
  Shield,
  Users,
  Building2,
  CreditCard,
  FileText,
  Mail,
  Calendar,
  MessageSquare,
  Code,
  Layers,
  RotateCcw,
  Wrench,
  Server,
  Bot,
  PieChart,
  Bell,
  Activity,
  AlertTriangle,
} from "lucide-react";
import { Logo } from "@/components/common/Logo";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/authStore";
import { useRecentNavStore } from "@/stores/recentNavStore";
import { useNotificationStore } from "@/stores/notificationStore";
import { cn } from "@/lib/utils";
import { ROUTES } from "@/lib/constants";
import { useNavigationStatus } from "@/hooks/useNavigationStatus";
import { useSwipeGesture } from "@/hooks/useSwipeGesture";
import { useKeyboardNavigation } from "@/hooks/useKeyboardNavigation";
import { Input } from "@/components/ui/input";

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

interface NavSection {
  title: string;
  items: Array<{
    path: string;
    label: string;
    icon: React.ComponentType<{ className?: string }>;
  }>;
}

const navigationSections: NavSection[] = [
  {
    title: "Overview",
    items: [
      { path: ROUTES.DASHBOARD, label: "Dashboard", icon: LayoutDashboard },
      { path: "/status", label: "Status", icon: Activity },
      { path: "/notifications", label: "Notifications", icon: Bell }
    ]
  },
  {
    title: "Management",
    items: [
      { path: ROUTES.FUNCTIONS, label: "Functions", icon: FunctionSquare },
      { path: ROUTES.APPS, label: "Apps", icon: Building2 },
      { path: ROUTES.REGISTRY, label: "Registry", icon: Package },
      { path: ROUTES.PROVIDERS, label: "Providers", icon: Cloud },
      { path: ROUTES.TEAMS, label: "Teams", icon: Users },
      { path: ROUTES.STATE_FABRIC, label: "State Fabric", icon: Database },
      { path: ROUTES.AGENTS, label: "Agents", icon: Bot }
    ]
  },
  {
    title: "Insights",
    items: [
      { path: ROUTES.ANALYTICS, label: "Analytics", icon: BarChart3 },
      { path: ROUTES.USAGE, label: "Usage", icon: PieChart }
    ]
  },
  {
    title: "Account",
    items: [
      { path: ROUTES.SETTINGS, label: "Settings", icon: Settings }
    ]
  }
];

const LG_BREAKPOINT = 1024;

function Sidebar({ isOpen, onClose }: SidebarProps) {
  const location = useLocation();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const status = useNavigationStatus();
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);

  const isAdmin = user?.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);

  const [mobileSearchQuery, setMobileSearchQuery] = useState("");
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [isLg, setIsLg] = useState(() => typeof window !== "undefined" && window.innerWidth >= LG_BREAKPOINT);

  const recordRecent = useRecentNavStore((s) => s.record);
  const recentPaths = useRecentNavStore((s) => s.recentPaths);

  useEffect(() => {
    const mq = window.matchMedia(`(min-width: ${LG_BREAKPOINT}px)`);
    const handler = () => setIsLg(mq.matches);
    handler();
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  // Record current route for recent-tab tracking (only when inside dashboard layout)
  useEffect(() => {
    recordRecent(location.pathname);
  }, [location.pathname, recordRecent]);

  // Swipe gesture for mobile
  const { gestureHandlers } = useSwipeGesture({
    onSwipeLeft: () => onClose(), // Close sidebar on swipe left
  });

  const adminSection: NavSection | null = isAdmin ? {
    title: "Admin",
    items: [
      { path: ROUTES.ADMIN_TENANTS, label: "Tenants", icon: Building2 },
      { path: ROUTES.ADMIN_USERS, label: "Users", icon: Users },
      { path: ROUTES.ADMIN_BILLING, label: "Billing", icon: CreditCard },
      { path: ROUTES.ADMIN_AUDIT, label: "Audit Log", icon: Shield },
      { path: ROUTES.ADMIN_SYSTEM, label: "System", icon: Wrench },
      { path: ROUTES.ADMIN_BACKENDS, label: "Platform Backends", icon: Server },
      { path: ROUTES.ADMIN_PROVIDERS, label: "Providers", icon: Cloud },
      { path: ROUTES.ADMIN_CONTENT, label: "Content", icon: FileText },
      { path: ROUTES.ADMIN_REDIRECTS, label: "Redirects", icon: RotateCcw },
      { path: ROUTES.ADMIN_NEWSLETTER, label: "Newsletter", icon: Mail },
      { path: ROUTES.ADMIN_CONTENT_CALENDAR, label: "Content Calendar", icon: Calendar },
      { path: ROUTES.ADMIN_FEEDBACK, label: "Feedback", icon: MessageSquare },
      { path: ROUTES.ADMIN_FUNCTIONS, label: "Functions", icon: Code },
      { path: ROUTES.ADMIN_REGISTRY, label: "Registry", icon: Database },
      { path: ROUTES.ADMIN_STATE_FABRIC, label: "State Fabric", icon: Layers },
      { path: "/admin/status/incidents", label: "Status Incidents", icon: AlertTriangle },
    ]
  } : null;

  const allSections = adminSection ? [...navigationSections, adminSection] : navigationSections;

  // Build path -> nav item map and derive recent items from stored paths (only show items that exist in current nav)
  const pathToItem = new Map(
    allSections.flatMap((s) => s.items.map((item) => [item.path, item] as const))
  );
  const recentItems = recentPaths
    .map((path) => pathToItem.get(path))
    .filter((item): item is NonNullable<typeof item> => item != null);

  // Initialize expanded state - all sections expanded by default except admin (collapsed by default)
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => {
    const initial = new Set(allSections.map(section => section.title));
    if (adminSection) {
      initial.delete("Admin"); // Collapse admin section by default
    }
    return initial;
  });

  const toggleSection = (sectionTitle: string) => {
    setExpandedSections(prev => {
      const newSet = new Set(prev);
      if (newSet.has(sectionTitle)) {
        newSet.delete(sectionTitle);
      } else {
        newSet.add(sectionTitle);
      }
      return newSet;
    });
  };

  // Filter sections and items based on mobile search
  const filteredSections = mobileSearchQuery.trim()
    ? allSections.map(section => ({
        ...section,
        items: section.items.filter(item =>
          item.label.toLowerCase().includes(mobileSearchQuery.toLowerCase())
        )
      })).filter(section => section.items.length > 0)
    : allSections;

  // Get all navigable items (flattened for keyboard navigation)
  const allNavigableItems = filteredSections.flatMap(section =>
    expandedSections.has(section.title) ? section.items : []
  );

  // Keyboard navigation
  const handleArrowUp = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex(prev => {
      const newIndex = prev <= 0 ? allNavigableItems.length - 1 : prev - 1;
      return newIndex;
    });
  }, [isOpen, allNavigableItems.length]);

  const handleArrowDown = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex(prev => {
      const newIndex = prev >= allNavigableItems.length - 1 ? 0 : prev + 1;
      return newIndex;
    });
  }, [isOpen, allNavigableItems.length]);

  const handleEnter = useCallback(() => {
    if (!isOpen || focusedIndex < 0 || focusedIndex >= allNavigableItems.length) return;
    const item = allNavigableItems[focusedIndex];
    if (item) {
      onClose();
      window.location.href = item.path;
    }
  }, [isOpen, focusedIndex, allNavigableItems, onClose]);

  const handleEscape = useCallback(() => {
    if (isOpen) {
      onClose();
    }
  }, [isOpen, onClose]);

  useKeyboardNavigation({
    enabled: isOpen,
    onArrowUp: handleArrowUp,
    onArrowDown: handleArrowDown,
    onEnter: handleEnter,
    onEscape: handleEscape,
  });

  const handleLogout = () => {
    logout();
  };

  return (
    <>
      {/* Mobile Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <motion.aside
        {...gestureHandlers}
        initial={false}
        animate={{
          x: isOpen || isLg ? 0 : -280,
        }}
        transition={{ type: "spring", stiffness: 300, damping: 30 }}
        className={cn(
          "dashboard-sidebar fixed left-0 top-0 z-50 h-screen w-[260px] min-w-[260px] bg-bg-primary border-r border-border-subtle",
          "flex flex-col lg:translate-x-0 lg:static lg:shrink-0"
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border-subtle">
          <Logo size="sm" />
          <button
            onClick={onClose}
            className="lg:hidden p-2 rounded-lg hover:bg-bg-hover text-text-secondary hover:text-text-primary"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Mobile Search */}
        <div className="px-4 pb-4 lg:hidden border-b border-border-subtle">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
            <Input
              placeholder="Search navigation..."
              value={mobileSearchQuery}
              onChange={(e) => setMobileSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-4" aria-label="Primary navigation">
          {/* Recent Items */}
          {recentItems.length > 0 && !mobileSearchQuery && (
            <div className="space-y-2">
              <h3 className="px-3 text-xs font-semibold text-text-muted uppercase tracking-wider">
                Recent
              </h3>
              <div className="space-y-1">
                {recentItems.map((item) => {
                  const isActive = location.pathname === item.path ||
                    (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path));
                  const Icon = item.icon;
                  const itemIndex = allNavigableItems.findIndex(navItem => navItem.path === item.path);
                  const isFocused = focusedIndex === itemIndex;

                  return (
                    <NavLink
                      key={`recent-${item.path}`}
                      to={item.path}
                      onClick={() => onClose()}
                      className={cn(
                        "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                        "relative overflow-hidden",
                        isActive
                          ? "text-text-primary bg-bg-hover"
                          : isFocused
                          ? "text-text-primary bg-bg-hover ring-2 ring-border-focus"
                          : "text-text-secondary hover:text-text-primary hover:bg-bg-hover"
                      )}
                    >
                      {isActive && (
                        <motion.div
                          layoutId="activeNav"
                          className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-6 rounded-full bg-linear-to-b from-brand-500 to-brand-600"
                          transition={{ type: "spring", stiffness: 300, damping: 30 }}
                        />
                      )}
                      <Icon className={cn("w-5 h-5", isActive && "text-brand-500")} />
                      <span>{item.label}</span>
                    </NavLink>
                  );
                })}
              </div>
            </div>
          )}

          {filteredSections.map((section) => {
            const isExpanded = expandedSections.has(section.title);
            const hasActiveItem = section.items.some(item =>
              location.pathname === item.path ||
              (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path))
            );

            return (
              <div key={section.title} className="space-y-2">
                <button
                  onClick={() => toggleSection(section.title)}
                  className={cn(
                    "flex items-center justify-between w-full px-3 py-2 text-xs font-semibold uppercase tracking-wider rounded-lg transition-colors",
                    "hover:bg-bg-hover",
                    hasActiveItem ? "text-text-primary" : "text-text-muted"
                  )}
                >
                  <span>{section.title}</span>
                  <ChevronDown className={cn(
                    "w-4 h-4 transition-transform duration-200",
                    isExpanded ? "rotate-0" : "-rotate-90"
                  )} />
                </button>

                <AnimatePresence>
                  {isExpanded && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="space-y-1 overflow-hidden"
                    >
                      {section.items.map((item) => {
                  const isActive = location.pathname === item.path ||
                    (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path));
                  const Icon = item.icon;

                  // Determine if this item has status indicators
                  const hasStatusIndicator = (() => {
                    switch (item.path) {
                      case ROUTES.REGISTRY:
                        return false;
                      case ROUTES.PROVIDERS:
                        return status.providers.hasOffline;
                      case ROUTES.ANALYTICS:
                        return status.analytics.hasAlerts;
                      case ROUTES.SETTINGS:
                        return status.settings.hasWarnings;
                      case "/notifications":
                        return unreadCount > 0;
                      default:
                        return false;
                    }
                  })();

                  const getStatusBadge = () => {
                    switch (item.path) {
                      case ROUTES.REGISTRY:
                        return null;
                      case ROUTES.PROVIDERS:
                        if (status.providers.hasOffline) {
                          return {
                            count: "!",
                            color: "bg-warning"
                          };
                        }
                        break;
                      case ROUTES.ANALYTICS:
                        if (status.analytics.hasAlerts) {
                          return {
                            count: "!",
                            color: "bg-warning"
                          };
                        }
                        break;
                      case ROUTES.SETTINGS:
                        if (status.settings.hasWarnings) {
                          return {
                            count: "!",
                            color: "bg-warning"
                          };
                        }
                        break;
                      case "/notifications":
                        if (unreadCount > 0) {
                          return {
                            count: unreadCount > 99 ? "99+" : unreadCount.toString(),
                            color: "bg-error"
                          };
                        }
                        break;
                    }
                    return null;
                  };

                  const statusBadge = getStatusBadge();
                  const itemIndex = allNavigableItems.findIndex(navItem => navItem.path === item.path);
                  const isFocused = focusedIndex === itemIndex;

                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      onClick={() => onClose()}
                      className={cn(
                        "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                        "relative overflow-hidden",
                        isActive
                          ? "text-text-primary bg-bg-hover"
                          : isFocused
                          ? "text-text-primary bg-bg-hover ring-2 ring-border-focus"
                          : "text-text-secondary hover:text-text-primary hover:bg-bg-hover"
                      )}
                    >
                      {isActive && (
                        <motion.div
                          layoutId="activeNav"
                          className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-6 rounded-full bg-linear-to-b from-brand-500 to-brand-600"
                          transition={{ type: "spring", stiffness: 300, damping: 30 }}
                        />
                      )}
                      <Icon className={cn("w-5 h-5", isActive && "text-brand-500")} />
                        <span className="flex-1">{item.label}</span>
                        {statusBadge && (
                          <span className={cn(
                            "flex items-center justify-center min-w-[18px] h-[18px] text-xs font-bold text-white rounded-full text-[10px]",
                            statusBadge.color
                          )}>
                            {statusBadge.count}
                          </span>
                        )}
                      </NavLink>
                    );
                  })}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            );
          })}
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-border-subtle">
          <Button
            variant="ghost"
            className="w-full justify-start gap-3 text-text-secondary hover:text-error hover:bg-error/10"
            onClick={handleLogout}
          >
            <LogOut className="w-5 h-5" />
            <span>Logout</span>
          </Button>
        </div>
      </motion.aside>
    </>
  );
}

export { Sidebar };

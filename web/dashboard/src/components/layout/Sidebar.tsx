import { Logo } from '@/components/common/Logo';
import { UpgradeBanner } from '@/components/enterprise';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { useKeyboardNavigation } from '@/hooks/useKeyboardNavigation';
import { useNavigationStatus } from '@/hooks/useNavigationStatus';
import { usePlan } from '@/hooks/usePlan';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { ROUTES } from '@/lib/constants';
import { hasFeature as planHasFeature } from '@/lib/plan-utils';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useRecentNavStore } from '@/stores/recentNavStore';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Activity,
  BarChart3,
  Bell,
  Bot,
  Briefcase,
  Building2,
  ChevronDown,
  Cloud,
  Code,
  Command,
  Database,
  FileSearch,
  Flame,
  FunctionSquare,
  Key,
  KeyRound,
  LayoutDashboard,
  LayoutGrid,
  LogOut,
  MessageSquare,
  PieChart,
  Puzzle,
  Rocket,
  Scale,
  Search,
  Settings,
  Shield,
  ShieldCheck,
  Sparkles,
  Star,
  LifeBuoy,
  TrendingUp,
  Users,
  Wallet,
  X,
  Network,
  Workflow,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { NavLink, useLocation, useNavigate } from 'react-router-dom';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

interface NavItem {
  path: string;
  label: string;
  icon: LucideIcon;
  badge?: 'new' | 'beta' | number;
  shortcut?: string;
  description?: string;
}

interface NavSection {
  id: string;
  title: string;
  icon: LucideIcon;
  items: NavItem[];
  collapsible?: boolean;
}

const navigationSections: NavSection[] = [
  {
    id: 'overview',
    title: 'Overview',
    icon: LayoutDashboard,
    collapsible: true,
    items: [
      {
        path: ROUTES.DASHBOARD,
        label: 'Discover',
        icon: Code,
        shortcut: 'G',
        description: 'Browse functions and marketplace',
      },
      {
        path: ROUTES.OVERVIEW,
        label: 'Overview',
        icon: LayoutDashboard,
        description: 'Your dashboard summary',
      },
      {
        path: '/notifications',
        label: 'Notifications',
        icon: Bell,
        description: 'View your notifications',
      },
      {
        path: ROUTES.CONVERSATIONS,
        label: 'Conversations',
        icon: MessageSquare,
        shortcut: 'C',
        description: 'Chat and message history',
      },
    ],
  },
  {
    id: 'functions',
    title: 'Functions',
    icon: FunctionSquare,
    collapsible: true,
    items: [
      {
        path: '/functions/hot',
        label: 'Hot',
        icon: Flame,
        shortcut: 'H',
        description: 'Trending hot functions right now',
      },
      {
        path: '/functions/trending',
        label: 'Trending',
        icon: TrendingUp,
        shortcut: 'T',
        description: 'Functions gaining popularity',
      },
      {
        path: '/functions/explore/new',
        label: 'New',
        icon: Sparkles,
        shortcut: 'N',
        description: 'Recently added functions',
      },
      {
        path: '/functions/popular',
        label: 'Popular',
        icon: Zap,
        description: 'Most used functions of all time',
      },
      {
        path: '/functions/favorites',
        label: 'Favorites',
        icon: Star,
        description: 'Your starred functions',
      },
      {
        path: '/functions/my',
        label: 'My Functions',
        icon: Code,
        description: 'Functions you created',
      },
    ],
  },
  {
    id: 'swarm',
    title: 'Agent Swarm',
    icon: Bot,
    collapsible: true,
    items: [
      {
        path: ROUTES.AGENTS,
        label: 'Agents',
        icon: Bot,
        shortcut: 'A',
        badge: 'new',
        description: 'Manage AI agents',
      },
      {
        path: ROUTES.EVOLUTION,
        label: 'Evolution',
        icon: Sparkles,
        badge: 'beta',
        description: 'Agent evolution tracking',
      },
      {
        path: ROUTES.MARKETPLACE_AGENTS,
        label: 'Marketplace',
        icon: Shield,
        description: 'Browse agent marketplace',
      },
      {
        path: ROUTES.AGENT_MEMORIES,
        label: 'Memory',
        icon: Database,
        description: 'Agent memory and context storage',
      },
    ],
  },
  {
    id: 'teams',
    title: 'Teams',
    icon: Users,
    collapsible: true,
    items: [
      {
        path: ROUTES.TEAMS,
        label: 'All Teams',
        icon: Users,
        shortcut: 'M',
        description: 'Manage your teams',
      },
      {
        path: '/my-team',
        label: 'My Team',
        icon: Shield,
        description: 'Your primary team and memory',
      },
    ],
  },
  {
    id: 'management',
    title: 'Management',
    icon: Building2,
    collapsible: true,
    items: [
      {
        path: '/ai-composer',
        label: 'AI Composer',
        icon: Sparkles,
        badge: 'new',
        description: 'AI-powered function generation',
      },
      {
        path: '/frg',
        label: 'Graph Editor',
        icon: Network,
        badge: 'beta',
        shortcut: 'R',
        description: 'Visual function graph editor',
      },
      {
        path: '/gallery',
        label: 'Gallery',
        icon: LayoutGrid,
        badge: 'new',
        description: 'Browse and remix public functions',
      },
      { path: ROUTES.APPS, label: 'Apps', icon: Building2, description: 'Your applications' },
      {
        path: ROUTES.SDK_INTEGRATIONS,
        label: 'SDK',
        icon: Puzzle,
        description: 'SDK integrations',
      },
      {
        path: ROUTES.PROVIDERS,
        label: 'Providers',
        icon: Cloud,
        shortcut: 'P',
        description: 'Cloud providers',
      },
      {
        path: ROUTES.STATE_FABRIC,
        label: 'State Fabric',
        icon: Network,
        badge: 'beta',
        description: 'Distributed state management',
      },
      {
        path: ROUTES.STATE,
        label: 'State',
        icon: Database,
        description: 'Function state management',
      },
      { path: ROUTES.SECRETS, label: 'Secrets', icon: Key, description: 'Secure secret storage' },
      {
        path: ROUTES.API_KEYS,
        label: 'API Keys',
        icon: KeyRound,
        description: 'API key management',
      },
      {
        path: '/wallet',
        label: 'Wallet',
        icon: Wallet,
        description: 'Platform wallet & credits',
      },
      {
        path: '/pricing/bundles',
        label: 'Bundles',
        icon: Rocket,
        badge: 'new',
        description: 'Backend-in-a-Box pricing bundles',
      },
    ],
  },
  {
    id: 'insights',
    title: 'Insights',
    icon: BarChart3,
    collapsible: true,
    items: [
      {
        path: ROUTES.ANALYTICS,
        label: 'Analytics',
        icon: BarChart3,
        shortcut: 'Y',
        description: 'Performance analytics',
      },
      { path: ROUTES.USAGE, label: 'Usage', icon: PieChart, description: 'Resource usage & cost analytics' },
      { path: '/status', label: 'Status', icon: Activity, description: 'System status' },
    ],
  },
  {
    id: 'enterprise',
    title: 'Enterprise',
    icon: Briefcase,
    collapsible: true,
    items: [
      {
        path: ROUTES.ENTERPRISE_SLA,
        label: 'SLA',
        icon: Scale,
        description: 'Service Level Agreements',
      },
      {
        path: ROUTES.ENTERPRISE_AUDIT,
        label: 'Audit Log',
        icon: FileSearch,
        description: 'Security audit and compliance logs',
      },
      {
        path: ROUTES.ENTERPRISE_SUPPORT,
        label: 'Support',
        icon: LifeBuoy,
        description: 'Enterprise support center',
      },
    ],
  },
  {
    id: 'account',
    title: 'Account',
    icon: Settings,
    collapsible: false,
    items: [
      {
        path: ROUTES.SETTINGS,
        label: 'Settings',
        icon: Settings,
        shortcut: 'S',
        description: 'Account settings',
      },
    ],
  },
];

const LG_BREAKPOINT = 1024;

// Animation variants
const sectionVariants = {
  collapsed: { height: 0, opacity: 0 },
  expanded: { height: 'auto', opacity: 1 },
};

function Sidebar({ isOpen, onClose }: SidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const status = useNavigationStatus();
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);

  const [mobileSearchQuery, setMobileSearchQuery] = useState('');
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [isLg, setIsLg] = useState(
    () => typeof window !== 'undefined' && window.innerWidth >= LG_BREAKPOINT
  );
  const searchInputRef = useRef<HTMLInputElement>(null);

  const recordRecent = useRecentNavStore((s) => s.record);
  const recentPaths = useRecentNavStore((s) => s.recentPaths);

  // Record current route for recent-tab tracking
  useEffect(() => {
    recordRecent(location.pathname);
  }, [location.pathname, recordRecent]);

  // Handle resize
  useEffect(() => {
    const mq = window.matchMedia(`(min-width: ${LG_BREAKPOINT}px)`);
    const handler = () => setIsLg(mq.matches);
    handler();
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  // Focus search on mobile when sidebar opens
  useEffect(() => {
    if (isOpen && !isLg && searchInputRef.current) {
      setTimeout(() => searchInputRef.current?.focus(), 100);
    }
  }, [isOpen, isLg]);

  /** Prefer /wallet/agents/:agentId when we know it */
  const walletNavPath = useMemo(() => {
    const m = location.pathname.match(/^\/wallet\/agents\/([^/]+)/);
    if (m?.[1]) return `/wallet/agents/${m[1]}`;
    // Platform wallet is always available at /wallet
    return '/wallet';
  }, [location.pathname]);

  const resolveNavPath = useCallback(
    (item: NavItem) => (item.label === 'Wallet' ? walletNavPath : item.path),
    [walletNavPath]
  );

  // Swipe gesture for mobile
  const { gestureHandlers } = useSwipeGesture({
    onSwipeLeft: () => onClose(),
  });

  const { plan } = usePlan();

  // Filter sections based on plan features
  const allSections = useMemo(() => {
    return navigationSections
      .map((section) => ({
        ...section,
        items: section.items.filter((item) => {
          if (item.path === ROUTES.STATE_FABRIC) {
            return planHasFeature(plan, 'STATE_FABRIC');
          }
          if (item.path === ROUTES.AGENTS) {
            return planHasFeature(plan, 'AGENTS');
          }
          return true;
        }),
      }))
      .filter((section) => {
        if (section.items.length === 0) return false;
        // Hide enterprise section for free/starter plans
        if (section.id === 'enterprise') {
          return planHasFeature(plan, 'ENTERPRISE_SECTION');
        }
        return true;
      });
  }, [plan]);

  // Build searchable items
  const searchableItems = useMemo(() => {
    return allSections.flatMap((section) =>
      section.items.map((item) => ({
        ...item,
        section: section.title,
      }))
    );
  }, [allSections]);

  // Filtered search results
  const searchResults = useMemo(() => {
    if (!mobileSearchQuery.trim()) return [];
    const query = mobileSearchQuery.toLowerCase();
    return searchableItems.filter(
      (item) =>
        item.label.toLowerCase().includes(query) ||
        item.section.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
    );
  }, [mobileSearchQuery, searchableItems]);

  // Determine expanded sections based on active route
  const getInitialExpanded = useCallback(() => {
    const expanded = new Set<string>();
    allSections.forEach((section) => {
      const hasActive = section.items.some(
        (item) =>
          location.pathname === item.path ||
          (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path))
      );
      if (hasActive || !section.collapsible) {
        expanded.add(section.id);
      }
    });
    return expanded;
  }, [allSections, location.pathname]);

  const [expandedSections, setExpandedSections] = useState<Set<string>>(getInitialExpanded);

  // Update expanded sections when route changes
  useEffect(() => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      allSections.forEach((section) => {
        const hasActive = section.items.some(
          (item) =>
            location.pathname === item.path ||
            (item.path !== ROUTES.DASHBOARD && location.pathname.startsWith(item.path))
        );
        if (hasActive) {
          next.add(section.id);
        }
      });
      return next;
    });
  }, [location.pathname, allSections]);

  const toggleSection = (sectionId: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      if (next.has(sectionId)) {
        next.delete(sectionId);
      } else {
        next.add(sectionId);
      }
      return next;
    });
  };

  // Get all visible navigable items for keyboard navigation
  const visibleNavigableItems = useMemo(() => {
    if (mobileSearchQuery && searchResults.length > 0) {
      return searchResults;
    }
    return allSections.flatMap((section) =>
      expandedSections.has(section.id) ? section.items : []
    );
  }, [mobileSearchQuery, searchResults, allSections, expandedSections]);

  // Keyboard navigation
  const handleArrowUp = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex((prev) => {
      if (prev <= 0) return visibleNavigableItems.length - 1;
      return prev - 1;
    });
  }, [isOpen, visibleNavigableItems.length]);

  const handleArrowDown = useCallback(() => {
    if (!isOpen) return;
    setFocusedIndex((prev) => {
      if (prev >= visibleNavigableItems.length - 1) return 0;
      return prev + 1;
    });
  }, [isOpen, visibleNavigableItems.length]);

  const handleEnter = useCallback(() => {
    if (!isOpen || focusedIndex < 0 || focusedIndex >= visibleNavigableItems.length) return;
    const item = visibleNavigableItems[focusedIndex];
    if (item) {
      onClose();
      navigate(resolveNavPath(item));
      setMobileSearchQuery('');
    }
  }, [isOpen, focusedIndex, visibleNavigableItems, onClose, navigate, resolveNavPath]);

  const handleEscape = useCallback(() => {
    if (isOpen) {
      if (mobileSearchQuery) {
        setMobileSearchQuery('');
      } else {
        onClose();
      }
    }
  }, [isOpen, mobileSearchQuery, onClose]);

  // Keyboard shortcuts for quick navigation
  const handleShortcut = useCallback(
    (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey) {
        const shortcut = e.key.toUpperCase();
        const item = searchableItems.find((i) => i.shortcut === shortcut);
        if (item) {
          e.preventDefault();
          navigate(resolveNavPath(item));
        }
      }
    },
    [searchableItems, navigate, resolveNavPath]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleShortcut);
    return () => document.removeEventListener('keydown', handleShortcut);
  }, [handleShortcut]);

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

  // Check if item is active
  const isItemActive = (path: string) => {
    return (
      location.pathname === path ||
      (path !== ROUTES.DASHBOARD && path !== ROUTES.OVERVIEW && location.pathname.startsWith(path))
    );
  };

  // Get status badge for item
  const getStatusBadge = (path: string): { count: string; color: string } | null => {
    switch (path) {
      case ROUTES.PROVIDERS:
        return status.providers.hasOffline ? { count: '!', color: 'bg-warning' } : null;
      case ROUTES.ANALYTICS:
        return status.analytics.hasAlerts ? { count: '!', color: 'bg-warning' } : null;
      case ROUTES.SETTINGS:
        return status.settings.hasWarnings ? { count: '!', color: 'bg-warning' } : null;
      case '/notifications':
        return unreadCount > 0
          ? { count: unreadCount > 99 ? '99+' : unreadCount.toString(), color: 'bg-error' }
          : null;
      default:
        return null;
    }
  };

  // Build path -> item map for recent items
  const pathToItem = useMemo(() => {
    const map = new Map<string, NavItem>();
    allSections.forEach((section) => {
      section.items.forEach((item) => map.set(item.path, item));
    });
    return map;
  }, [allSections]);

  const recentItems = useMemo(() => {
    return recentPaths
      .map((path) => pathToItem.get(path))
      .filter((item): item is NonNullable<typeof item> => item != null)
      .slice(0, 4);
  }, [recentPaths, pathToItem]);

  return (
    <TooltipProvider delayDuration={0}>
      <>
        {/* Mobile Overlay */}
        <AnimatePresence>
          {isOpen && !isLg && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
              onClick={onClose}
            />
          )}
        </AnimatePresence>

        {/* Sidebar */}
        <motion.aside
          {...gestureHandlers}
          initial={false}
          animate={{
            x: isOpen || isLg ? 0 : -300,
          }}
          transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          className={cn(
            'fixed left-0 top-0 z-50 h-screen w-[280px] min-w-[280px]',
            'aviation-sidebar',
            'flex flex-col',
            'lg:static lg:translate-x-0 lg:self-stretch'
          )}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
            <Logo size="sm" />
            <div className="flex items-center gap-1">
              {/* Command Palette Trigger (Desktop) */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => {
                      document.dispatchEvent(
                        new KeyboardEvent('keydown', { key: 'k', metaKey: true })
                      );
                    }}
                    className="hidden lg:flex items-center gap-1.5 px-2 py-1.5 rounded-md bg-aviation-bg-instrument/50 border border-aviation-border-instrument text-aviation-text-muted hover:text-aviation-text-secondary hover:border-aviation-amber/30 transition-colors"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span className="text-[10px] font-medium">K</span>
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>Command Palette</p>
                </TooltipContent>
              </Tooltip>

              <button
                onClick={onClose}
                aria-label="Close navigation"
                className="lg:hidden p-2 rounded-lg hover:bg-aviation-bg-instrument text-aviation-text-secondary hover:text-aviation-text-primary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Mobile Search */}
          <div className="px-3 py-3 lg:hidden border-b border-aviation-border-panel">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
              <Input
                ref={searchInputRef}
                placeholder="Search navigation..."
                value={mobileSearchQuery}
                onChange={(e) => setMobileSearchQuery(e.target.value)}
                className="pl-9 bg-aviation-bg-instrument border-aviation-border-instrument text-aviation-text-primary placeholder:text-aviation-text-dim focus:border-aviation-amber focus:ring-aviation-amber/20"
              />
            </div>
          </div>

          {/* Navigation */}
          <nav
            className="flex-1 min-h-0 overflow-y-auto aviation-scroll py-3"
            aria-label="Primary navigation"
          >
            {/* Search Results */}
            <AnimatePresence mode="wait">
              {mobileSearchQuery && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="px-3 pb-3"
                >
                  <p className="px-3 text-xs font-medium text-aviation-text-muted mb-2">
                    Search Results
                  </p>
                  {searchResults.length > 0 ? (
                    <div className="space-y-1">
                      {searchResults.map((item, index) => {
                        const isActive = isItemActive(item.path);
                        const Icon = item.icon;
                        const isFocused = focusedIndex === index;

                        return (
                          <NavLink
                            key={`search-${item.path}`}
                            to={resolveNavPath(item)}
                            onClick={() => {
                              onClose();
                              setMobileSearchQuery('');
                            }}
                            className={cn(
                              'aviation-sidebar-item',
                              isActive && 'aviation-sidebar-item-active',
                              isFocused && 'ring-2 ring-aviation-amber/50'
                            )}
                          >
                            <Icon className="aviation-sidebar-icon" />
                            <div className="flex-1 min-w-0">
                              <span className="font-medium block truncate">{item.label}</span>
                              <span className="text-xs text-aviation-text-muted block truncate">
                                {item.section}
                              </span>
                            </div>
                          </NavLink>
                        );
                      })}
                    </div>
                  ) : (
                    <div className="px-3 py-8 text-center">
                      <Search className="w-8 h-8 text-aviation-text-dim mx-auto mb-2" />
                      <p className="text-sm text-aviation-text-muted">No results found</p>
                    </div>
                  )}
                </motion.div>
              )}
            </AnimatePresence>

            {/* Recent Items */}
            {!mobileSearchQuery && recentItems.length > 0 && (
              <div className="px-3 mb-4">
                <p className="px-3 text-[10px] font-semibold text-aviation-text-muted uppercase tracking-wider mb-2">
                  Recent
                </p>
                <div className="space-y-0.5">
                  {recentItems.map((item) => {
                    const isActive = isItemActive(item.path);
                    const Icon = item.icon;

                    return (
                      <Tooltip key={`recent-${item.path}`}>
                        <TooltipTrigger asChild>
                          <NavLink
                            to={resolveNavPath(item)}
                            onClick={() => onClose()}
                            className={cn(
                              'aviation-sidebar-item',
                              isActive && 'aviation-sidebar-item-active'
                            )}
                          >
                            <Icon className="aviation-sidebar-icon" />
                            <span className="flex-1 font-medium truncate">{item.label}</span>
                            {item.shortcut && (
                              <kbd className="aviation-sidebar-kbd">
                                ⌘{item.shortcut}
                              </kbd>
                            )}
                          </NavLink>
                        </TooltipTrigger>
                        <TooltipContent side="right">
                          <p>{item.description || item.label}</p>
                        </TooltipContent>
                      </Tooltip>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Navigation Sections */}
            {!mobileSearchQuery && (
              <div className="space-y-1 px-3">
                {allSections.map((section) => {
                  const isExpanded = expandedSections.has(section.id);
                  const SectionIcon = section.icon;
                  const hasActiveItem = section.items.some((item) => isItemActive(item.path));

                  return (
                    <div key={section.id} className="mb-1">
                      {/* Section Header */}
                      {section.collapsible ? (
                        <button
                          onClick={() => toggleSection(section.id)}
                          className={cn(
                            'flex items-center justify-between w-full px-3 py-2 rounded-lg transition-all duration-200',
                            'aviation-sidebar-section',
                            hasActiveItem && 'aviation-sidebar-section-active'
                          )}
                        >
                          <div className="flex items-center gap-2">
                            <SectionIcon className="w-4 h-4" />
                            <span>{section.title}</span>
                          </div>
                          <motion.div
                            animate={{ rotate: isExpanded ? 0 : -90 }}
                            transition={{ duration: 0.2 }}
                            className="aviation-sidebar-toggle"
                          >
                            <ChevronDown className="w-3.5 h-3.5 aviation-sidebar-toggle-icon" />
                          </motion.div>
                        </button>
                      ) : (
                        <div
                          className={cn(
                            'px-3 py-2 aviation-sidebar-section',
                            hasActiveItem && 'aviation-sidebar-section-active'
                          )}
                        >
                          {section.title}
                        </div>
                      )}

                      {/* Section Items */}
                      <AnimatePresence initial={false}>
                        {isExpanded && (
                          <motion.div
                            initial="collapsed"
                            animate="expanded"
                            exit="collapsed"
                            variants={sectionVariants}
                            transition={{ duration: 0.2, ease: 'easeInOut' }}
                            className="overflow-hidden"
                          >
                            <div className="space-y-0.5 pt-1">
                              {section.items.map((item, itemIndex) => {
                                const isActive = isItemActive(item.path);
                                const Icon = item.icon;
                                const statusBadge = getStatusBadge(item.path);
                                const globalIndex =
                                  allSections
                                    .slice(0, allSections.indexOf(section))
                                    .reduce(
                                      (acc, s) =>
                                        acc + (expandedSections.has(s.id) ? s.items.length : 0),
                                      0
                                    ) + itemIndex;
                                const isFocused = focusedIndex === globalIndex;

                                return (
                                  <Tooltip key={item.path}>
                                    <TooltipTrigger asChild>
                                      <NavLink
                                        to={resolveNavPath(item)}
                                        onClick={() => onClose()}
                                        className={cn(
                                          'aviation-sidebar-item',
                                          isActive && 'aviation-sidebar-item-active',
                                          isFocused && 'ring-2 ring-aviation-amber/50'
                                        )}
                                      >
                                        {/* Active indicator */}
                                        {isActive && (
                                          <motion.div
                                            layoutId="activeNavIndicator"
                                            className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-full bg-linear-to-b from-aviation-amber to-aviation-amber-glow"
                                            transition={{
                                              type: 'spring',
                                              stiffness: 400,
                                              damping: 30,
                                            }}
                                          />
                                        )}

                                        <Icon className="aviation-sidebar-icon" />

                                        <span className="flex-1 font-medium truncate">
                                          {item.label}
                                        </span>

                                        {/* Badge */}
                                        {item.badge && (
                                          <span
                                            className={cn(
                                              'aviation-sidebar-badge',
                                              item.badge === 'new'
                                                ? 'aviation-sidebar-badge-new'
                                                : 'aviation-sidebar-badge-beta'
                                            )}
                                          >
                                            {item.badge}
                                          </span>
                                        )}

                                        {/* Status badge */}
                                        {statusBadge && (
                                          <span
                                            className={cn(
                                              'flex items-center justify-center min-w-[18px] h-[18px] text-[10px] font-bold text-white rounded-full aviation-sidebar-status',
                                              statusBadge.color
                                            )}
                                          >
                                            {statusBadge.count}
                                          </span>
                                        )}

                                        {/* Shortcut hint */}
                                        {item.shortcut && (
                                          <kbd className="aviation-sidebar-kbd">
                                            ⌘{item.shortcut}
                                          </kbd>
                                        )}
                                      </NavLink>
                                    </TooltipTrigger>
                                    <TooltipContent side="right">
                                      <p>{item.description || item.label}</p>
                                      {item.shortcut && (
                                        <p className="text-xs text-aviation-text-muted mt-1">
                                          Shortcut: ⌘{item.shortcut}
                                        </p>
                                      )}
                                    </TooltipContent>
                                  </Tooltip>
                                );
                              })}
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}
              </div>
            )}
          </nav>

          {/* Upgrade Banner - only for free users */}
          <div className="px-3 pb-3">
            <UpgradeBanner placement="sidebar" />
          </div>

          {/* Footer */}
          <div className="p-3 border-t border-aviation-border-panel">
            <div className="aviation-profile mb-3">
              <div className="aviation-profile-avatar">
                {user?.avatar ? (
                  <img
                    src={user.avatar}
                    alt={user.name || 'User'}
                  />
                ) : (
                  <div className="aviation-profile-initials">
                    {(user?.name?.[0] || user?.email?.[0] || 'U').toUpperCase()}
                  </div>
                )}
                <span className="aviation-profile-status" />
              </div>
              <div className="aviation-profile-info">
                <p className="aviation-profile-name">
                  {user?.username || user?.name || user?.email?.split('@')[0] || 'User'}
                </p>
                <p className="aviation-profile-plan">{plan || 'Free'} Plan</p>
              </div>
            </div>

            <button
              className="aviation-signout"
              onClick={handleLogout}
            >
              <LogOut className="aviation-signout-icon" />
              <span className="aviation-signout-text">Sign Out</span>
            </button>
          </div>
        </motion.aside>
      </>
    </TooltipProvider>
  );
}

export { Sidebar };

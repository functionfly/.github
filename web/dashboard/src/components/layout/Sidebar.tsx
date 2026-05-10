import { Logo } from '@/components/common/Logo';
import { UpgradeBanner } from '@/components/enterprise';
import { Input } from '@/components/ui/input';
import { useTranslation } from 'react-i18next';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { useActiveEnvironment, type Environment } from '@/hooks/useActiveEnvironment';
import { useKeyboardNavigation } from '@/hooks/useKeyboardNavigation';
import { useNavigationStatus, useStatusBadge } from '@/hooks/useNavigationStatus';
import { usePlan } from '@/hooks/usePlan';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { ROUTES } from '@/lib/constants';
import { hasFeature as planHasFeature } from '@/lib/plan-utils';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useRecentNavStore } from '@/stores/recentNavStore';
import { useSidebarStore } from '@/stores/sidebarStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Activity,
  ArrowRight,
  Award,
  BarChart3,
  BadgeCheck,
  Bell,
  Bot,
  Boxes,
  Building2,
  CheckCircle,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Cloud,
  Code,
  Command,
  Database,
  Dna,
  Flame,
  GitFork,
  History,
  GripVertical,
  Key,
  KeyRound,
  LayoutGrid,
  LifeBuoy,
  LogOut,
  MessageSquare,
  Network,
  PieChart,
  Pin,
  Puzzle,
  Rocket,
  Search,
  Server,
  Settings,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Users,
  Wallet,
  Workflow,
  X,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { NavLink, useLocation, useNavigate } from 'react-router-dom';

const GitHubIcon = () => (
  <svg role="img" viewBox="0 0 24 24" className="w-5 h-5" xmlns="http://www.w3.org/2000/svg">
    <path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
  </svg>
);

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

interface NavItem {
  path: string;
  label: string;
  icon: LucideIcon | React.ComponentType;
  badge?: 'new' | 'beta' | number;
  shortcut?: string;
  description?: string;
  onboardingHint?: string;
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
    id: 'discover',
    title: 'Discover',
    icon: Search,
    collapsible: true,
    items: [
      {
        path: ROUTES.DISCOVER,
        label: 'Discover',
        icon: Code,
        shortcut: 'G',
        description: 'Browse functions and marketplace',
        onboardingHint: 'Start here to explore functions',
      },
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
        path: '/functions/favorites',
        label: 'Favorites',
        icon: Star,
        description: 'Your starred functions',
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
    id: 'build',
    title: 'Build',
    icon: Workflow,
    collapsible: true,
    items: [
      {
        path: '/functions/my',
        label: 'My Functions',
        icon: Code,
        description: 'Functions you created',
        onboardingHint: 'Create your first function here',
      },
      {
        path: '/ai/composer',
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
        path: ROUTES.STATE,
        label: 'State',
        icon: Database,
        description: 'Function state management',
      },
      {
        path: '/github',
        label: 'GitHub Import',
        icon: GitHubIcon,
        badge: 'new',
        description: 'Import functions from GitHub repositories',
      },
      {
        path: '/functions/paste',
        label: 'Paste Code',
        icon: Code,
        description: 'Paste and import code snippets',
      },
    ],
  },
  {
    id: 'deploy',
    title: 'Deploy',
    icon: Rocket,
    collapsible: true,
    items: [
      {
        path: ROUTES.APPS,
        label: 'Apps',
        icon: Building2,
        description: 'Your applications',
        onboardingHint: 'Deploy your first app',
      },
      {
        path: '/agents', // Resolved dynamically in resolveNavPath based on username
        label: 'Agents',
        icon: Bot,
        shortcut: 'A',
        badge: 'new',
        description: 'Manage AI agents',
      },
      {
        path: ROUTES.PROVIDERS,
        label: 'Providers',
        icon: Cloud,
        shortcut: 'P',
        description: 'Cloud providers',
      },
      {
        path: ROUTES.SDK_INTEGRATIONS,
        label: 'SDK',
        icon: Puzzle,
        description: 'SDK integrations',
      },
      {
        path: ROUTES.SECRETS,
        label: 'Secrets',
        icon: Key,
        description: 'Secure secret storage',
      },
      {
        path: ROUTES.API_KEYS,
        label: 'API Keys',
        icon: KeyRound,
        description: 'API key management',
      },
      {
        path: '/backends',
        label: 'One-Click Backends',
        icon: Boxes,
        badge: 'new',
        shortcut: 'B',
        description: 'Deploy backends with one click',
      },
    ],
  },
  {
    id: 'operate',
    title: 'Operate',
    icon: Activity,
    collapsible: true,
    items: [
      {
        path: ROUTES.ANALYTICS,
        label: 'Analytics',
        icon: BarChart3,
        shortcut: 'Y',
        description: 'Performance analytics',
      },
      {
        path: ROUTES.USAGE,
        label: 'Usage',
        icon: PieChart,
        description: 'Resource usage & cost analytics',
      },
      {
        path: '/wallet',
        label: 'Wallet',
        icon: Wallet,
        description: 'Platform wallet & credits',
      },
      {
        path: '/bundles',
        label: 'Bundles',
        icon: LayoutGrid,
        badge: 'new',
        description: 'Backend-in-a-Box pricing bundles',
      },
      {
        path: '/status',
        label: 'Status',
        icon: Activity,
        description: 'System status',
      },
      {
        path: '/dna/overview',
        label: 'Function DNA',
        icon: Dna,
        badge: 'new',
        description: 'Living code that evolves itself',
      },
      {
        path: '/time-machine',
        label: 'Time Machine',
        icon: History,
        badge: 'new',
        description: 'Rewind and fix production bugs',
      },
      {
        path: '/certification',
        label: 'Certification',
        icon: Award,
        badge: 'new',
        description: 'Earn verifiable developer credentials',
      },
      {
        path: '/credentials',
        label: 'My Credentials',
        icon: BadgeCheck,
        description: 'View earned certifications',
      },
    ],
  },
  {
    id: 'advanced',
    title: 'Advanced',
    icon: Shield,
    collapsible: true,
    items: [
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
      {
        path: ROUTES.TEAMS,
        label: 'Teams',
        icon: Users,
        shortcut: 'M',
        description: 'Manage your teams',
      },
      {
        path: ROUTES.DECISIONS,
        label: 'Decisions',
        icon: CheckCircle,
        badge: 'new',
        description: 'Team decision recorder',
      },
      {
        path: ROUTES.STATE_FABRIC,
        label: 'State Fabric',
        icon: Network,
        badge: 'beta',
        description: 'Distributed state management',
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
      {
        path: ROUTES.ENTERPRISE_SUPPORT,
        label: 'Support',
        icon: LifeBuoy,
        description: 'Help and support center',
      },
    ],
  },
];

const LG_BREAKPOINT = 1024;

// Map sidebar labels to translation keys
const NAV_LABEL_KEYS: Record<string, string> = {
  'Discover': 'nav.discover',
  'Hot': 'nav.hot',
  'Trending': 'nav.trending',
  'New': 'nav.new',
  'Notifications': 'nav.notifications',
  'Conversations': 'nav.conversations',
  'My Functions': 'nav.myFunctions',
  'AI Composer': 'nav.aiComposer',
  'Graph Editor': 'nav.graphEditor',
  'State': 'nav.state',
  'Apps': 'nav.apps',
  'Agents': 'nav.agents',
  'Providers': 'nav.providers',
  'SDK': 'nav.sdk',
  'Secrets': 'nav.secrets',
  'API Keys': 'nav.apiKeys',
  'One-Click Backends': 'nav.oneClickBackends',
  'Analytics': 'nav.analytics',
  'Usage': 'nav.usage',
  'Wallet': 'nav.wallet',
  'Bundles': 'nav.bundles',
  'Status': 'nav.status',
  'Time Machine': 'nav.timeMachine',
  'Evolution': 'nav.evolution',
  'Marketplace': 'nav.marketplace',
  'Memory': 'nav.memory',
  'Teams': 'nav.teams',
  'Decisions': 'nav.decisions',
  'State Fabric': 'nav.stateFabric',
  'GitHub Import': 'nav.githubImport',
  'Settings': 'nav.settings',
  'Support': 'nav.support',
  'Recent': 'nav.recent',
  'Getting Started': 'nav.gettingStarted',
  'Sign Out': 'nav.signOut',
  'Search Results': 'nav.search',
};

// Animation variants
const sectionVariants = {
  collapsed: { height: 0, opacity: 0 },
  expanded: { height: 'auto', opacity: 1 },
};

function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const status = useNavigationStatus();
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);
  const { plan } = usePlan();

  const translateLabel = (label: string) => NAV_LABEL_KEYS[label] ? t(NAV_LABEL_KEYS[label]) : label;

  // Sidebar store for persistent state
  const {
    isCollapsed,
    toggleCollapsed,
    expandedSections,
    toggleSection,
    favorites,
    toggleFavorite,
    isFavorite,
    showOnboardingHints,
    completedOnboardingSteps,
  } = useSidebarStore();

  // Active environment with backend sync
  const { environment: currentEnvironment, setEnvironment, isLoading: isEnvironmentLoading } = useActiveEnvironment();

  const { isOnboardingComplete } = useOnboardingStore();

  // Local state
  const [mobileSearchQuery, setMobileSearchQuery] = useState('');
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [isLg, setIsLg] = useState(
    () => typeof window !== 'undefined' && window.innerWidth >= LG_BREAKPOINT
  );
  const [draggingSection, setDraggingSection] = useState<string | null>(null);
  const [dragOverSection, setDragOverSection] = useState<string | null>(null);
  const [showShortcutsHint, setShowShortcutsHint] = useState(false);
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

  // Show keyboard shortcuts hint on first load
  useEffect(() => {
    const hasShownHint = localStorage.getItem('sidebar-shortcuts-hint-shown');
    if (!hasShownHint) {
      setTimeout(() => {
        setShowShortcutsHint(true);
        setTimeout(() => {
          setShowShortcutsHint(false);
          localStorage.setItem('sidebar-shortcuts-hint-shown', 'true');
        }, 5000);
      }, 2000);
    }
  }, []);

  /** Prefer /wallet/agents/:agentId when we know it */
  const walletNavPath = useMemo(() => {
    const m = location.pathname.match(/^\/wallet\/agents\/([^/]+)/);
    if (m?.[1]) return `/wallet/agents/${m[1]}`;
    return '/wallet';
  }, [location.pathname]);

  const resolveNavPath = useCallback(
    (item: NavItem) => {
      if (item.label === 'Wallet') return walletNavPath;
      if (item.label === 'Agents' && user?.username) {
        return ROUTES.AGENTS(user.username);
      }
      return item.path;
    },
    [walletNavPath, user]
  );

  // Swipe gesture for mobile
  const { gestureHandlers } = useSwipeGesture({
    onSwipeLeft: () => onClose(),
  });

  // Filter sections based on plan features
  const allSections = useMemo(() => {
    return navigationSections
      .map((section) => ({
        ...section,
        items: section.items.filter((item) => {
          if (item.path === ROUTES.STATE_FABRIC) {
            return planHasFeature(plan, 'STATE_FABRIC');
          }
          if (item.path === '/agents') {
            return planHasFeature(plan, 'AGENTS');
          }
          if (item.path === ROUTES.ENTERPRISE_SUPPORT) {
            return planHasFeature(plan, 'DEDICATED_SUPPORT');
          }
          return true;
        }),
      }))
      .filter((section) => {
        if (section.items.length === 0) return false;
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

  // Get favorites items
  const favoriteItems = useMemo(() => {
    return favorites
      .map((path) => searchableItems.find((item) => item.path === path))
      .filter((item): item is (NavItem & { section: string }) => item != null);
  }, [favorites, searchableItems]);

  // Drag and drop handlers
  const handleDragStart = (sectionId: string) => {
    setDraggingSection(sectionId);
  };

  const handleDragOver = (e: React.DragEvent, sectionId: string) => {
    e.preventDefault();
    if (draggingSection && draggingSection !== sectionId) {
      setDragOverSection(sectionId);
    }
  };

  const handleDragLeave = () => {
    setDragOverSection(null);
  };

  const handleDrop = (e: React.DragEvent, targetSectionId: string) => {
    e.preventDefault();
    setDraggingSection(null);
    setDragOverSection(null);
  };

  // Keyboard navigation
  const visibleNavigableItems = useMemo(() => {
    if (mobileSearchQuery && searchResults.length > 0) {
      return searchResults;
    }
    return allSections.flatMap((section) =>
      expandedSections.has(section.id) ? section.items : []
    );
  }, [mobileSearchQuery, searchResults, allSections, expandedSections]);

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
    if (!isOpen || focusedIndex < 0 || focusedIndex >= visibleNavigableItems.length)
      return;
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

  // Onboarding progress calculation
  const onboardingProgress = useMemo(() => {
    const totalSteps = 5;
    const completed = completedOnboardingSteps.length;
    return Math.round((completed / totalSteps) * 100);
  }, [completedOnboardingSteps]);

  // Collapse button position
  const CollapseButton = () => (
    <button
      onClick={toggleCollapsed}
      className="aviation-collapse-btn hidden lg:flex"
      title={isCollapsed ? 'Expand sidebar (⌘B)' : 'Collapse sidebar (⌘B)'}
      aria-label={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
    >
      {isCollapsed ? (
        <ChevronRight className="w-4 h-4" />
      ) : (
        <ChevronLeft className="w-4 h-4" />
      )}
    </button>
  );

  // Environment switcher component - memoized to prevent unnecessary re-renders
  const EnvironmentSwitcher = useCallback(() => (
    <div className={cn('aviation-environment-tabs', isEnvironmentLoading && 'opacity-50 pointer-events-none')}>
      {(['production', 'staging', 'development'] as const).map((env) => (
        <button
          key={env}
          onClick={() => setEnvironment(env)}
          data-env={env}
          disabled={isEnvironmentLoading}
          className={cn(
            'aviation-environment-tab',
            currentEnvironment === env && 'active'
          )}
          title={`Switch to ${env} environment`}
        >
          {env.charAt(0).toUpperCase()}
        </button>
      ))}
    </div>
  ), [currentEnvironment, setEnvironment, isEnvironmentLoading]);

  // Onboarding progress component
  const OnboardingProgress = () => {
    return (
      <div className="aviation-onboarding-progress">
        <div className="aviation-onboarding-progress-header">
          <span>{translateLabel('Getting Started')}</span>
          <span>{onboardingProgress}%</span>
        </div>
        <div className="aviation-onboarding-progress-bar">
          <div
            className="aviation-onboarding-progress-fill"
            style={{ width: `${onboardingProgress}%` }}
          />
        </div>
      </div>
    );
  };

  // Enhanced NavItem component with favorites and badges
  const NavItemComponent = ({
    item,
    isActive,
    isFocused,
    sectionId,
  }: {
    item: NavItem & { section?: string };
    isActive: boolean;
    isFocused: boolean;
    sectionId: string;
  }) => {
    const Icon = item.icon;
    const favorite = isFavorite(item.path);

    return (
      <Tooltip key={item.path}>
        <TooltipTrigger asChild>
          <NavLink
            to={resolveNavPath(item)}
            onClick={() => {
              onClose();
            }}
            className={cn(
              'aviation-sidebar-item group',
              isActive && 'aviation-sidebar-item-active',
              isFocused && 'ring-2 ring-aviation-amber/50'
            )}
            title={item.description}
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

            {Icon && <Icon className="aviation-sidebar-icon flex-shrink-0" />}

            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
              {translateLabel(item.label)}
            </span>

            {/* Favorite button */}
            <button
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                toggleFavorite(item.path);
              }}
              className={cn(
                'aviation-favorite-btn p-1 rounded hover:bg-aviation-bg-instrument',
                favorite && 'is-favorite'
              )}
              title={favorite ? 'Remove from favorites' : 'Add to favorites'}
            >
              <Pin className="w-3 h-3" fill={favorite ? 'currentColor' : 'none'} />
            </button>

            {/* Enhanced status badge */}
            <StatusBadge path={item.path} />

            {/* Original badge */}
            {item.badge && (
              <span
                className={cn(
                  'aviation-sidebar-badge',
                  item.badge === 'new' ? 'aviation-sidebar-badge-new' : 'aviation-sidebar-badge-beta'
                )}
              >
                {item.badge}
              </span>
            )}

            {/* Shortcut hint */}
            {item.shortcut && !isCollapsed && (
              <kbd className="aviation-sidebar-kbd">⌘{item.shortcut}</kbd>
            )}
          </NavLink>
        </TooltipTrigger>
        <TooltipContent side={isCollapsed ? 'right' : 'right'}>
          <p>{item.description || item.label}</p>
          {item.shortcut && (
            <p className="text-xs text-aviation-text-muted mt-1">Shortcut: ⌘{item.shortcut}</p>
          )}
        </TooltipContent>
      </Tooltip>
    );
  };

  // Status badge component using the new hook
  const StatusBadge = ({ path }: { path: string }) => {
    const badge = useStatusBadge(path);
    if (!badge.content) return null;

    return (
      <span
        className="aviation-status-badge-count"
        data-type={badge.type}
        title={badge.type === 'warning' ? 'Requires attention' : undefined}
      >
        {badge.content}
      </span>
    );
  };

  // Section header component with drag handle and expand/collapse
  const SectionHeader = ({
    section,
    isExpanded,
    hasActiveItem,
  }: {
    section: NavSection;
    isExpanded: boolean;
    hasActiveItem: boolean;
  }) => {
    const SectionIcon = section.icon;

    // Collapsed state: show icon with tooltip menu for items
    if (isCollapsed) {
      const activeItem = section.items.find((item) => isItemActive(item.path));

      return (
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <button
              onClick={() => section.collapsible && toggleSection(section.id)}
              className={cn(
                'w-full flex items-center justify-center py-3 rounded-lg transition-all duration-200',
                'aviation-sidebar-section aviation-sidebar-section-collapsed',
                (hasActiveItem || activeItem) && 'aviation-sidebar-section-active'
              )}
            >
              <SectionIcon className="w-5 h-5 flex-shrink-0 text-aviation-cyan" />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right" className="p-0 bg-aviation-bg-panel border-aviation-border-panel">
            <div className="py-2">
              <p className="px-3 py-1 text-xs font-semibold text-aviation-cyan uppercase tracking-wider">
                {section.title}
              </p>
              <div className="mt-1 space-y-0.5">
                {section.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = isItemActive(item.path);
                  return (
                    <NavLink
                      key={item.path}
                      to={resolveNavPath(item)}
                      onClick={() => onClose()}
                      className={cn(
                        'flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-aviation-bg-instrument transition-colors',
                        isActive ? 'text-aviation-amber' : 'text-aviation-text-secondary'
                      )}
                    >
                      <Icon className="w-4 h-4" />
                      <span>{translateLabel(item.label)}</span>
                    </NavLink>
                  );
                })}
              </div>
            </div>
          </TooltipContent>
        </Tooltip>
      );
    }

    return (
      <div
        className={cn(
          'flex items-center gap-2',
          section.collapsible && 'cursor-pointer'
        )}
        onClick={() => section.collapsible && toggleSection(section.id)}
      >
        {/* Drag handle (desktop only, not on mobile) */}
        {isLg && section.collapsible && (
          <button
            className="aviation-sidebar-drag-handle p-1 rounded hover:bg-aviation-bg-instrument"
            draggable
            onDragStart={() => handleDragStart(section.id)}
            onDragOver={(e) => handleDragOver(e, section.id)}
            onDragLeave={handleDragLeave}
            onDrop={(e) => handleDrop(e, section.id)}
            title="Drag to reorder"
            aria-label="Drag to reorder section"
          >
            <GripVertical className="w-3 h-3" />
          </button>
        )}

        <div
          className={cn(
            'flex-1 px-3 py-2 rounded-lg transition-all duration-200 flex items-center gap-2',
            'aviation-sidebar-section aviation-sidebar-section-title',
            hasActiveItem && 'aviation-sidebar-section-active'
          )}
          title={section.title}
        >
          <SectionIcon className="w-4 h-4 flex-shrink-0 text-aviation-cyan" />
          <span className="flex-1 text-aviation-cyan font-semibold">{translateLabel(section.title)}</span>
        </div>

        {section.collapsible && (
          <motion.div
            animate={{ rotate: isExpanded ? 0 : -90 }}
            transition={{ duration: 0.2 }}
            className="aviation-sidebar-toggle"
            aria-expanded={isExpanded}
            aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
          >
            <ChevronDown className="w-3.5 h-3.5 aviation-sidebar-toggle-icon" />
          </motion.div>
        )}
      </div>
    );
  };

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
            x: isOpen || isLg ? 0 : isCollapsed ? -64 : -300,
            width: isCollapsed ? 64 : 280,
          }}
          transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          className={cn(
            'fixed left-0 top-0 z-50 h-screen flex flex-col',
            'aviation-sidebar',
            isCollapsed && 'aviation-sidebar-collapsed',
            !isLg && 'aviation-sidebar-mobile-sheet',
            'lg:relative lg:translate-x-0 lg:self-stretch lg:z-auto'
          )}
        >
          {/* Collapse Button (Desktop) */}
          <CollapseButton />

          {/* Mobile Handle */}
          {!isLg && <div className="aviation-sidebar-mobile-handle lg:hidden" />}

          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
            <Logo size={isCollapsed ? 'xs' : 'sm'} />
            <div className="flex items-center gap-1">
              {/* Command Palette Trigger (Desktop, not collapsed) */}
              {!isCollapsed && (
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
              )}

              <button
                onClick={onClose}
                aria-label="Close navigation"
                className="lg:hidden p-2 rounded-lg hover:bg-aviation-bg-instrument text-aviation-text-secondary hover:text-aviation-text-primary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Environment Switcher (Desktop only, not collapsed) */}
          {!isCollapsed && isLg && (
            <div className="aviation-workspace-switcher !pt-8">
              <EnvironmentSwitcher />
            </div>
          )}

          {/* Onboarding Progress (if applicable) */}
          {!isCollapsed && <OnboardingProgress />}

          {/* Mobile Search */}
          {!isCollapsed && (
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
          )}

          {/* Navigation */}
          <nav
            className="flex-1 min-h-0 overflow-y-auto aviation-scroll py-3"
            aria-label="Primary navigation"
          >
            {/* Search Results */}
            <AnimatePresence mode="wait">
              {mobileSearchQuery && !isCollapsed && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="px-3 pb-3"
                >
                  <p className="px-3 text-xs font-medium text-aviation-text-muted mb-2">
                    {translateLabel('Search Results')}
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
                              <span className="font-medium block truncate">{translateLabel(item.label)}</span>
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

            {/* Favorites Section */}
            {!mobileSearchQuery && favoriteItems.length > 0 && !isCollapsed && (
              <div className="aviation-sidebar-favorites px-3 mb-4">
                <p className="aviation-sidebar-favorites-title">{translateLabel('Favorites')}</p>
                <div className="space-y-0.5">
                  {favoriteItems.map((item) => {
                    const isActive = isItemActive(item.path);
                    const Icon = item.icon;

                    return (
                      <Tooltip key={`fav-${item.path}`}>
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
                            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
                              {translateLabel(item.label)}
                            </span>
                            <button
                              onClick={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                                toggleFavorite(item.path);
                              }}
                              className="aviation-favorite-btn is-favorite p-1 rounded hover:bg-aviation-bg-instrument"
                            >
                              <Pin className="w-3 h-3" fill="currentColor" />
                            </button>
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

            {/* Recent Items */}
            {!mobileSearchQuery && recentItems.length > 0 && !isCollapsed && (
              <div className="aviation-sidebar-recent px-3 mb-4">
                <p className="aviation-sidebar-recent-title">{translateLabel('Recent')}</p>
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
                            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
                              {translateLabel(item.label)}
                            </span>
                            {item.shortcut && (
                              <kbd className="aviation-sidebar-kbd">⌘{item.shortcut}</kbd>
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

            {/* Divider before main navigation */}
            {(favoriteItems.length > 0 || recentItems.length > 0) && !isCollapsed && (
              <div className="aviation-sidebar-divider" />
            )}

            {/* Navigation Sections */}
            {!mobileSearchQuery && (
              <div className={cn('space-y-1', isCollapsed ? 'px-1' : 'px-3')}>
                {allSections.map((section, sectionIndex) => {
                  const isExpanded = expandedSections.has(section.id);
                  const hasActiveItem = section.items.some((item) => isItemActive(item.path));
                  const isDragging = draggingSection === section.id;
                  const isDragOver = dragOverSection === section.id;

                  return (
                    <div
                      key={section.id}
                      className={cn(
                        'mb-1',
                        isDragging && 'aviation-sidebar-section-dragging',
                        isDragOver && 'aviation-sidebar-section-drag-over'
                      )}
                      draggable={section.collapsible && !isCollapsed}
                      onDragStart={() => handleDragStart(section.id)}
                      onDragOver={(e) => handleDragOver(e, section.id)}
                      onDragLeave={handleDragLeave}
                      onDrop={(e) => handleDrop(e, section.id)}
                    >
                      {/* Section Header */}
                      {section.collapsible ? (
                        <div
                          role="button"
                          tabIndex={0}
                          onClick={() => toggleSection(section.id)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ') {
                              e.preventDefault();
                              toggleSection(section.id);
                            }
                          }}
                          className={cn(
                            'flex items-center justify-between w-full rounded-lg transition-all duration-200',
                            'aviation-sidebar-section cursor-pointer',
                            hasActiveItem && 'aviation-sidebar-section-active'
                          )}
                          aria-expanded={isExpanded}
                          aria-controls={`section-${section.id}`}
                          title={`${section.title} ${isExpanded ? '(click to collapse)' : '(click to expand)'}`}
                        >
                          <SectionHeader
                            section={section}
                            isExpanded={isExpanded}
                            hasActiveItem={hasActiveItem}
                          />
                        </div>
                      ) : (
                        <div
                          className={cn(
                            isCollapsed ? 'py-1' : 'px-3 py-2',
                            'aviation-sidebar-section',
                            hasActiveItem && 'aviation-sidebar-section-active'
                          )}
                        >
                          <SectionHeader
                            section={section}
                            isExpanded={true}
                            hasActiveItem={hasActiveItem}
                          />
                        </div>
                      )}

                      {/* Section Items - hidden when collapsed */}
                      {!isCollapsed && (
                        <AnimatePresence initial={false}>
                          {isExpanded && (
                            <motion.div
                              id={`section-${section.id}`}
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
                                  const globalIndex =
                                    allSections
                                      .slice(0, sectionIndex)
                                      .reduce(
                                        (acc, s) =>
                                          acc + (expandedSections.has(s.id) ? s.items.length : 0),
                                        0
                                      ) + itemIndex;
                                  const isFocused = focusedIndex === globalIndex;

                                  return (
                                    <NavItemComponent
                                      key={item.path}
                                      item={item}
                                      isActive={isActive}
                                      isFocused={isFocused}
                                      sectionId={section.id}
                                    />
                                  );
                                })}
                              </div>
                            </motion.div>
                          )}
                        </AnimatePresence>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </nav>

          {/* Divider before footer */}
          {!isCollapsed && <div className="aviation-sidebar-divider mt-auto" />}

          {/* Footer */}
          <div className="p-3 border-t border-aviation-border-panel">
            {/* Quick Links - Changelog & Feedback (not collapsed) */}
            {!isCollapsed && (
              <div className="flex items-center justify-center gap-4 mb-3 px-2">
                <NavLink
                  to="/changelog"
                  className={({ isActive }) =>
                    cn(
                      'text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors',
                      isActive && 'text-aviation-text-primary font-medium'
                    )
                  }
                >
                  {translateLabel('Changelog')}
                </NavLink>
                <span className="text-aviation-border-subtle">·</span>
                <NavLink
                  to="/feedback"
                  className={({ isActive }) =>
                    cn(
                      'text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors',
                      isActive && 'text-aviation-text-primary font-medium'
                    )
                  }
                >
                  {translateLabel('Feedback')}
                </NavLink>
              </div>
            )}

            {/* Upgrade Banner - only for free users (not collapsed) */}
            {!isCollapsed && (
              <div className="mb-3">
                <UpgradeBanner placement="sidebar" />
              </div>
            )}

            {/* User Profile */}
            <div className={cn('aviation-profile mb-3', isCollapsed && 'justify-center')}>
              <div className="aviation-profile-avatar">
                {user?.avatar ? (
                  <img src={user.avatar} alt={user.name || 'User'} />
                ) : (
                  <div className="aviation-profile-initials">
                    {(user?.name?.[0] || user?.email?.[0] || 'U').toUpperCase()}
                  </div>
                )}
                <span className="aviation-profile-status" />
              </div>
              {!isCollapsed && (
                <div className="aviation-profile-info">
                  <p className="aviation-profile-name">
                    {user?.username || user?.name || user?.email?.split('@')[0] || 'User'}
                  </p>
                  <p className="aviation-profile-plan">{plan || 'Free'} Plan</p>
                </div>
              )}
            </div>

            {/* Sign Out */}
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  className={cn(
                    'aviation-signout',
                    isCollapsed && 'justify-center px-0'
                  )}
                  onClick={handleLogout}
                >
                  <LogOut className="aviation-signout-icon" />
                  {!isCollapsed && <span className="aviation-signout-text">{translateLabel('Sign Out')}</span>}
                </button>
              </TooltipTrigger>
              {isCollapsed && (
                <TooltipContent side="right">
                  <p>{translateLabel('Sign Out')}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </div>
        </motion.aside>

        {/* Keyboard Shortcuts Hint */}
        <div className={cn('aviation-shortcut-hint', showShortcutsHint && 'visible')}>
          <span className="font-medium">Tip:</span> Press{' '}
          <kbd className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-aviation-text-primary border border-aviation-border-subtle">
            ⌘B
          </kbd>{' '}
          to toggle sidebar
        </div>
      </>
    </TooltipProvider>
  );
}

export { Sidebar };

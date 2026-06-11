/**
 * Main Sidebar component.
 *
 * Production-readiness fixes:
 * - Rate-limited keyboard shortcuts (prevents rapid-fire abuse)
 * - Debounced mobile search (prevents excessive re-renders)
 * - Logout confirmation dialog (prevents accidental sign-out)
 * - Fixed isItemActive path matching (/settings won't match /settings-other)
 * - ErrorBoundary wrapper (sidebar failure doesn't crash the whole app)
 * - HTML-escaped user data in profile (defense-in-depth XSS guard)
 * - Avatar alt text safe fallback (no user-controlled content leak)
 * - Onboarding progress gated on mid-onboarding users only
 */

import { Logo } from '@/components/common/Logo';
import { UpgradeBanner } from '@/components/enterprise';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Input } from '@/components/ui/input';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { useActiveEnvironment } from '@/hooks/useActiveEnvironment';
import { useKeyboardNavigation } from '@/hooks/useKeyboardNavigation';
import { useStatusBadge } from '@/hooks/useNavigationStatus';
import { usePlan } from '@/hooks/usePlan';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { hasFeature as planHasFeature } from '@/lib/plan-utils';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { useRecentNavStore } from '@/stores/recentNavStore';
import { useSidebarStore } from '@/stores/sidebarStore';
import { AnimatePresence, motion } from 'framer-motion';
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Command,
  GripVertical,
  LogOut,
  Pin,
  Search,
  X,
} from 'lucide-react';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { NavLink, useLocation, useNavigate } from 'react-router-dom';
import {
  createRateLimitedHandler,
  escapeHtml,
  isItemActive,
  translateLabel,
  useDebounce,
} from './helpers';
import {
  LG_BREAKPOINT,
  navigationSections,
  SECTION_VARIANTS,
  type NavItem,
  type NavSection,
  type SidebarProps,
} from './navigation';

// ============================================================================
// Error Boundary — prevents sidebar crash from taking down the whole app
// ============================================================================

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class SidebarErrorBoundary extends React.Component<
  React.PropsWithChildren<Record<string, unknown>>,
  ErrorBoundaryState
> {
  constructor(props: React.PropsWithChildren<Record<string, unknown>>) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('[Sidebar] Render error:', error, errorInfo);
  }

  componentDidUpdate(prevProps: React.PropsWithChildren<Record<string, unknown>>) {
    // If the children changed (e.g. route change triggered a re-render),
    // attempt to recover from a stale error state.
    if (this.state.hasError && this.props.children !== prevProps.children) {
      this.setState({ hasError: false, error: null });
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center h-screen bg-aviation-bg-primary text-aviation-text-secondary p-6">
          <div className="text-center">
            <p className="text-sm font-medium text-aviation-red mb-1">
              Sidebar failed to load
            </p>
            <p className="text-xs text-aviation-text-muted">
              Refresh the page or sign out to continue.
            </p>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}

// ============================================================================
// Sidebar
// ============================================================================

function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const { plan } = usePlan();

  // ─── Wallet path resolution ────────────────────────────────────────────────
  const walletNavPath = useMemo(() => {
    const m = location.pathname.match(/^\/wallet\/agents\/([^/]+)/);
    if (m?.[1]) return `/wallet/agents/${m[1]}`;
    return '/wallet';
  }, [location.pathname]);

  // ─── Resolve dynamic nav paths ─────────────────────────────────────────────
  const resolveNavPath = useCallback(
    (item: NavItem) => {
      if (item.label === 'Wallet') return walletNavPath;
      // Agents and Conversations use dashboard routes (inside DashboardLayout)
      // NOT /u/:username routes which are outside DashboardLayout
      return item.path;
    },
    [walletNavPath]
  );

  // ─── Filter sections based on plan features ───────────────────────────────
  const allSections = useMemo(() => {
    return navigationSections
      .map((section) => ({
        ...section,
        items: section.items.filter((item) => {
          if (item.path === '/state-fabric') return planHasFeature(plan, 'STATE_FABRIC');
          if (item.path === '/agents') return planHasFeature(plan, 'AGENTS');
          if (item.path === '/enterprise/support') return planHasFeature(plan, 'DEDICATED_SUPPORT');
          if (item.path === '/dna/overview') return planHasFeature(plan, 'FUNCTION_DNA');
          return true;
        }),
      }))
      .filter((section) => section.items.length > 0);
  }, [plan]);

  // ─── Build searchable items ────────────────────────────────────────────────
  const searchableItems = useMemo(() => {
    return allSections.flatMap((section) =>
      section.items.map((item) => ({
        ...item,
        section: section.title,
      }))
    );
  }, [allSections]);

  const {
    isCollapsed,
    toggleCollapsed,
    expandedSections,
    toggleSection,
    favorites,
    toggleFavorite,
    isFavorite,
    completedOnboardingSteps,
    moveSection,
  } = useSidebarStore();

  const {
    environment: currentEnvironment,
    setEnvironment,
    isLoading: isEnvironmentLoading,
  } = useActiveEnvironment();

  const { isOnboardingComplete } = useOnboardingStore();

  // Local state
  const [mobileSearchQuery, setMobileSearchQuery] = useState('');

  // ─── Filtered search results ───────────────────────────────────────────────
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

  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [isLg, setIsLg] = useState(
    () => typeof window !== 'undefined' && window.innerWidth >= LG_BREAKPOINT
  );
  const [draggingSection, setDraggingSection] = useState<string | null>(null);
  const [dragOverSection, setDragOverSection] = useState<string | null>(null);
  // Throttle drag-over updates to one per animation frame (~60fps) to avoid excess re-renders
  const dragOverThrottleRef = useRef<number>(0);
  const [showShortcutsHint, setShowShortcutsHint] = useState(false);
  const [showLogoutDialog, setShowLogoutDialog] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);

  const recordRecent = useRecentNavStore((s) => s.record);
  const recentPaths = useRecentNavStore((s) => s.recentPaths);

  // ─── Debounced search update ─────────────────────────────────────────────
  const debouncedSetSearch = useDebounce((query: string) => {
    setMobileSearchQuery(query);
  }, 150);

  // ─── Keyboard shortcut handler (rate-limited) ──────────────────────────────
  const handleShortcut = useCallback(
    createRateLimitedHandler((e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey) {
        const shortcut = e.key.toUpperCase();
        const item = searchableItems.find((i) => i.shortcut === shortcut);
        if (item) {
          e.preventDefault();
          navigate(resolveNavPath(item));
        }
      }
    }),
    [searchableItems, navigate, resolveNavPath]
  );

  // ─── Record current route ─────────────────────────────────────────────────
  useEffect(() => {
    recordRecent(location.pathname);
  }, [location.pathname, recordRecent]);

  // ─── Handle resize ────────────────────────────────────────────────────────
  useEffect(() => {
    const mq = window.matchMedia(`(min-width: ${LG_BREAKPOINT}px)`);
    const handler = () => setIsLg(mq.matches);
    handler();
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  // ─── Focus search on mobile when sidebar opens ────────────────────────────
  useEffect(() => {
    if (isOpen && !isLg && searchInputRef.current) {
      setTimeout(() => searchInputRef.current?.focus(), 100);
    }
  }, [isOpen, isLg]);

  // ─── Show keyboard shortcuts hint once ─────────────────────────────────────
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

  // ─── Keyboard shortcut listener ───────────────────────────────────────────
  useEffect(() => {
    document.addEventListener('keydown', handleShortcut);
    return () => document.removeEventListener('keydown', handleShortcut);
  }, [handleShortcut]);

  // ─── Keyboard navigation ──────────────────────────────────────────────────
  const visibleNavigableItems = useMemo(() => {
    if (mobileSearchQuery && searchResults.length > 0) return searchResults;
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
      if (mobileSearchQuery) setMobileSearchQuery('');
      else onClose();
    }
  }, [isOpen, mobileSearchQuery, onClose]);

  useKeyboardNavigation({
    enabled: isOpen,
    onArrowUp: handleArrowUp,
    onArrowDown: handleArrowDown,
    onEnter: handleEnter,
    onEscape: handleEscape,
  });

  // ─── Swipe gesture ─────────────────────────────────────────────────────────
  const { gestureHandlers } = useSwipeGesture({
    onSwipeLeft: () => onClose(),
  });

  // ─── Favorites ─────────────────────────────────────────────────────────────
  const favoriteItems = useMemo(() => {
    return favorites
      .map((path) => searchableItems.find((item) => item.path === path))
      .filter((item): item is (NavItem & { section: string }) => item != null);
  }, [favorites, searchableItems]);

  // ─── Drag and drop ─────────────────────────────────────────────────────────
  const handleDragStart = (sectionId: string) => setDraggingSection(sectionId);
  const handleDragOver = (e: React.DragEvent, sectionId: string) => {
    e.preventDefault();
    if (draggingSection && draggingSection !== sectionId) {
      const now = Date.now();
      if (now - dragOverThrottleRef.current >= 16) {
        dragOverThrottleRef.current = now;
        setDragOverSection(sectionId);
      }
    }
  };
  const handleDragLeave = () => setDragOverSection(null);
  const handleDrop = (e: React.DragEvent, targetSectionId: string) => {
    e.preventDefault();
    if (draggingSection && draggingSection !== targetSectionId) {
      const order = useSidebarStore.getState().sectionOrder;
      const fromIndex = order.indexOf(draggingSection);
      const toIndex = order.indexOf(targetSectionId);
      if (fromIndex !== -1 && toIndex !== -1) {
        moveSection(fromIndex, toIndex);
      }
    }
    setDraggingSection(null);
    setDragOverSection(null);
  };

  // ─── Path → item map for recent items ─────────────────────────────────────
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
      .slice(0, 5);
  }, [recentPaths, pathToItem]);

  // ─── Onboarding progress ────────────────────────────────────────────────────
  const onboardingProgress = useMemo(() => {
    const totalSteps = 5;
    const completed = completedOnboardingSteps.length;
    return Math.round((completed / totalSteps) * 100);
  }, [completedOnboardingSteps]);

  // ─── User display name — safe, no XSS possible ─────────────────────────────
  const userDisplayName = useMemo(() => {
    if (user?.username) return escapeHtml(user.username);
    if (user?.name) return escapeHtml(user.name);
    if (user?.email) return escapeHtml(user.email.split('@')[0]);
    return 'User';
  }, [user]);

  // =======================================================================
  // Sub-components
  // =======================================================================

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

  const EnvironmentSwitcher = useCallback(
    () => (
      <div
        className={cn(
          'aviation-environment-tabs',
          isEnvironmentLoading && 'opacity-50 pointer-events-none'
        )}
      >
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
    ),
    [currentEnvironment, setEnvironment, isEnvironmentLoading]
  );

  // Only show onboarding progress for users mid-onboarding (not complete, not 0)
  const OnboardingProgress = () => {
    if (isOnboardingComplete || onboardingProgress === 0) return null;
    return (
      <div className="aviation-onboarding-progress">
        <div className="aviation-onboarding-progress-header">
          <span>{translateLabel(t, 'Getting Started')}</span>
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

  const NavItemComponent = ({
    item,
    isActive,
    isFocused,
  }: {
    item: NavItem & { section?: string };
    isActive: boolean;
    isFocused: boolean;
  }) => {
    const Icon = item.icon;
    const favorite = isFavorite(item.path);

    return (
      <Tooltip key={item.path}>
        <TooltipTrigger asChild>
          <NavLink
            to={resolveNavPath(item)}
            onClick={onClose}
            className={cn(
              'aviation-sidebar-item group',
              isActive && 'aviation-sidebar-item-active',
              isFocused && 'ring-2 ring-aviation-amber/50'
            )}
            title={item.description}
          >
            {isActive && (
              <motion.div
                layoutId="activeNavIndicator"
                className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-full bg-linear-to-b from-aviation-amber to-aviation-amber-glow"
                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              />
            )}
            {Icon && <Icon className="aviation-sidebar-icon flex-shrink-0" />}
            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
              {translateLabel(t, item.label)}
            </span>
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
            <StatusBadge path={item.path} />
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
            {item.shortcut && !isCollapsed && (
              <kbd className="aviation-sidebar-kbd">⌘{item.shortcut}</kbd>
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
  };

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

    if (isCollapsed) {
      const activeItem = section.items.find((item) => isItemActive(item.path, location.pathname));
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
          <TooltipContent
            side="right"
            className="p-0 bg-aviation-bg-panel border-aviation-border-panel"
          >
            <div className="py-2">
              <p className="px-3 py-1 text-xs font-semibold text-aviation-cyan uppercase tracking-wider">
                {translateLabel(t, section.title)}
              </p>
              <div className="mt-1 space-y-0.5">
                {section.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = isItemActive(item.path, location.pathname);
                  return (
                    <NavLink
                      key={item.path}
                      to={resolveNavPath(item)}
                      onClick={onClose}
                      className={cn(
                        'flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-aviation-bg-instrument transition-colors',
                        isActive ? 'text-aviation-amber' : 'text-aviation-text-secondary'
                      )}
                    >
                      <Icon className="w-4 h-4" />
                      <span>{translateLabel(t, item.label)}</span>
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
          <span className="flex-1 text-aviation-cyan font-semibold">
            {translateLabel(t, section.title)}
          </span>
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

  // =======================================================================
  // Render
  // =======================================================================

  return (
    <SidebarErrorBoundary>
      <TooltipProvider delayDuration={0}>
        {/* ── Logout confirmation dialog ────────────────────────────────────── */}
        <AlertDialog open={showLogoutDialog} onOpenChange={setShowLogoutDialog}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Sign out?</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to sign out of your account?
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={() => {
                  setShowLogoutDialog(false);
                  logout();
                }}
                className="bg-aviation-red hover:bg-aviation-red/90"
              >
                Sign out
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        {/* ── Mobile Overlay ─────────────────────────────────────────────────── */}
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

        {/* ── Sidebar ────────────────────────────────────────────────────────── */}
        <motion.aside
          {...gestureHandlers}
          initial={false}
          animate={{
            x: isOpen || isLg ? 0 : isCollapsed ? -64 : -300,
            width: isCollapsed ? 64 : 280,
          }}
          transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          className={cn(
            'fixed left-0 top-0 z-50 min-h-screen flex flex-col',
            'aviation-sidebar',
            isCollapsed && 'aviation-sidebar-collapsed',
            !isLg && 'aviation-sidebar-mobile-sheet',
            'lg:relative lg:translate-x-0 lg:self-stretch lg:z-auto'
          )}
        >
          <CollapseButton />
          {!isLg && <div className="aviation-sidebar-mobile-handle lg:hidden" />}

          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
            <Logo size={isCollapsed ? 'xs' : 'sm'} />
            <div className="flex items-center gap-1">
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

          {/* Environment Switcher */}
          {!isCollapsed && isLg && (
            <div className="aviation-workspace-switcher !pt-8">
              <EnvironmentSwitcher />
            </div>
          )}

          {/* Onboarding Progress — only for mid-onboarding users */}
          {!isCollapsed && <OnboardingProgress />}

          {/* Mobile Search — debounced */}
          {!isCollapsed && (
            <div className="px-3 py-3 lg:hidden border-b border-aviation-border-panel">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
                <Input
                  ref={searchInputRef}
                  placeholder="Search navigation..."
                  value={mobileSearchQuery}
                  onChange={(e) => debouncedSetSearch(e.target.value)}
                  className="pl-9 bg-aviation-bg-instrument border-aviation-border-instrument text-aviation-text-primary placeholder:text-aviation-text-dim focus:border-aviation-amber focus:ring-aviation-amber/20"
                />
              </div>
            </div>
          )}

          {/* Navigation */}
          <nav className="flex-1 min-h-0 overflow-y-auto aviation-scroll py-3" aria-label="Primary navigation">
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
                    {translateLabel(t, 'Search Results')}
                  </p>
                  {searchResults.length > 0 ? (
                    <div className="space-y-1">
                      {searchResults.map((item, index) => {
                        const isActive = isItemActive(item.path, location.pathname);
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
                              <span className="font-medium block truncate">
                                {translateLabel(t, item.label)}
                              </span>
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

            {/* Favorites */}
            {!mobileSearchQuery && favoriteItems.length > 0 && !isCollapsed && (
              <div className="aviation-sidebar-favorites px-3 mb-4">
                <p className="aviation-sidebar-favorites-title">
                  {translateLabel(t, 'Favorites')}
                </p>
                <div className="space-y-0.5">
                  {favoriteItems.map((item) => {
                    const isActive = isItemActive(item.path, location.pathname);
                    const Icon = item.icon;
                    return (
                      <Tooltip key={`fav-${item.path}`}>
                        <TooltipTrigger asChild>
                          <NavLink
                            to={resolveNavPath(item)}
                            onClick={onClose}
                            className={cn(
                              'aviation-sidebar-item',
                              isActive && 'aviation-sidebar-item-active'
                            )}
                          >
                            <Icon className="aviation-sidebar-icon" />
                            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
                              {translateLabel(t, item.label)}
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

            {/* Recent */}
            {!mobileSearchQuery && recentItems.length > 0 && !isCollapsed && (
              <div className="aviation-sidebar-recent px-3 mb-4">
                <p className="aviation-sidebar-recent-title">
                  {translateLabel(t, 'Recent')}
                </p>
                <div className="space-y-0.5">
                  {recentItems.map((item) => {
                    const isActive = isItemActive(item.path, location.pathname);
                    const Icon = item.icon;
                    return (
                      <Tooltip key={`recent-${item.path}`}>
                        <TooltipTrigger asChild>
                          <NavLink
                            to={resolveNavPath(item)}
                            onClick={onClose}
                            className={cn(
                              'aviation-sidebar-item',
                              isActive && 'aviation-sidebar-item-active'
                            )}
                          >
                            <Icon className="aviation-sidebar-icon" />
                            <span className="aviation-sidebar-item-label flex-1 font-medium truncate">
                              {translateLabel(t, item.label)}
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

            {/* Divider */}
            {(favoriteItems.length > 0 || recentItems.length > 0) && !isCollapsed && (
              <div className="aviation-sidebar-divider" />
            )}

            {/* Navigation Sections */}
            {!mobileSearchQuery && (
              <div className={cn('space-y-1', isCollapsed ? 'px-1' : 'px-3')}>
                {allSections.map((section, sectionIndex) => {
                  const isExpanded = expandedSections.has(section.id);
                  const hasActiveItem = section.items.some((item) =>
                    isItemActive(item.path, location.pathname)
                  );
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
                      {/* Section header */}
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

                      {/* Section items — hidden when collapsed */}
                      {!isCollapsed && (
                        <AnimatePresence initial={false}>
                          {isExpanded && (
                            <motion.div
                              id={`section-${section.id}`}
                              initial="collapsed"
                              animate="expanded"
                              exit="collapsed"
                              variants={SECTION_VARIANTS}
                              transition={{ duration: 0.2, ease: 'easeInOut' }}
                              className="overflow-hidden grid"
                            >
                              <div className="space-y-0.5 pt-1 min-h-0">
                                {section.items.map((item, itemIndex) => {
                                  const isActive = isItemActive(item.path, location.pathname);
                                  const globalIndex =
                                    allSections
                                      .slice(0, sectionIndex)
                                      .reduce(
                                        (acc, s) => acc + s.items.length,
                                        0
                                      ) + itemIndex;
                                  const isFocused = focusedIndex === globalIndex;
                                  return (
                                    <NavItemComponent
                                      key={item.path}
                                      item={item}
                                      isActive={isActive}
                                      isFocused={isFocused}
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
            {/* Quick links */}
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
                  {translateLabel(t, 'Changelog')}
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
                  {translateLabel(t, 'Feedback')}
                </NavLink>
              </div>
            )}

            {/* Upgrade Banner */}
            {!isCollapsed && (
              <div className="mb-3">
                <UpgradeBanner placement="sidebar" />
              </div>
            )}

            {/* User Profile — avatar alt text is always safe fallback */}
            <div className={cn('aviation-profile mb-3', isCollapsed && 'justify-center')}>
              <div className="aviation-profile-avatar">
                {user?.avatar ? (
                  <img
                    src={user.avatar}
                    alt={userDisplayName}
                    onError={(e) => {
                      // Fallback to initials if avatar fails to load
                      (e.target as HTMLImageElement).style.display = 'none';
                      (e.target as HTMLImageElement).nextElementSibling?.classList.remove('hidden');
                    }}
                  />
                ) : null}
                <div className="aviation-profile-initials">
                  {userDisplayName.charAt(0).toUpperCase()}
                </div>
                <span className="aviation-profile-status" />
              </div>
              {!isCollapsed && (
                <div className="aviation-profile-info">
                  <p className="aviation-profile-name">{userDisplayName}</p>
                  <p className="aviation-profile-plan">{plan || 'Free'} Plan</p>
                </div>
              )}
            </div>

            {/* Sign Out — opens confirmation dialog */}
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  className={cn(
                    'aviation-signout',
                    isCollapsed && 'justify-center px-0'
                  )}
                  onClick={() => setShowLogoutDialog(true)}
                >
                  <LogOut className="aviation-signout-icon" />
                  {!isCollapsed && (
                    <span className="aviation-signout-text">
                      {translateLabel(t, 'Sign Out')}
                    </span>
                  )}
                </button>
              </TooltipTrigger>
              {isCollapsed && (
                <TooltipContent side="right">
                  <p>{translateLabel(t, 'Sign Out')}</p>
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
      </TooltipProvider>
    </SidebarErrorBoundary>
  );
}

export type { SidebarProps } from './navigation';
export { Sidebar };


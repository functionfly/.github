/**
 * Admin Sidebar Component - Production Ready
 * Organized sections, search, keyboard shortcuts, collapsible groups
 */

import { ROUTES } from '@/lib/constants';
import { cn } from '@/lib/utils';
import {
    Activity,
    AlertTriangle,
    BarChart3,
    BookOpen,
    Boxes,
    Building2,
    Calendar,
    ChevronDown,
    CircleDot,
    Cloud,
    Command,
    CreditCard,
    Factory,
    FileText,
    Gavel,
    Home,
    KeyRound,
    Landmark,
    LayoutDashboard,
    Mail,
    MessageSquare,
    PanelTop,
    RotateCcw,
    Search,
    Settings,
    Shield,
    TrendingUp,
    UserPlus,
    Users,
    Wrench,
    X,
    Zap,
    type LucideIcon,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';

// CSS-based transitions instead of framer-motion

interface AdminSidebarProps {
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

// Organized navigation with logical grouping
const navigationSections: NavSection[] = [
  {
    id: 'dashboard',
    title: 'Dashboard',
    icon: LayoutDashboard,
    collapsible: false,
    items: [
      {
        path: ROUTES.ADMIN_DASHBOARD,
        label: 'Overview',
        icon: Home,
        shortcut: 'H',
        description: 'Admin dashboard overview',
      },
      {
        path: ROUTES.ADMIN_AUDIT,
        label: 'Audit Log',
        icon: FileText,
        description: 'System audit logs',
      },
    ],
  },
  {
    id: 'platform',
    title: 'Platform',
    icon: Building2,
    collapsible: true,
    items: [
      {
        path: ROUTES.ADMIN_TENANTS,
        label: 'Tenants',
        icon: Building2,
        shortcut: 'T',
        description: 'Manage tenants',
      },
      {
        path: ROUTES.ADMIN_USERS,
        label: 'Users',
        icon: Users,
        shortcut: 'U',
        description: 'User management',
      },
      {
        path: ROUTES.ADMIN_SIGNUP_INVITES,
        label: 'Invites',
        icon: Shield,
        description: 'Signup invitations',
      },
      {
        path: ROUTES.ADMIN_WAITLIST,
        label: 'Waitlist',
        icon: UserPlus,
        description: 'Waitlist applicants',
      },
      {
        path: ROUTES.ADMIN_BILLING,
        label: 'Billing',
        icon: CreditCard,
        description: 'Billing management',
      },
    ],
  },
  {
    id: 'infrastructure',
    title: 'Infrastructure',
    icon: Boxes,
    collapsible: true,
    items: [
      {
        path: ROUTES.ADMIN_BACKENDS,
        label: 'Backends',
        icon: Boxes,
        description: 'Backend services',
      },
      {
        path: ROUTES.ADMIN_PROVIDERS,
        label: 'Providers',
        icon: Zap,
        description: 'Cloud providers',
      },
      {
        path: ROUTES.ADMIN_SYSTEM,
        label: 'System',
        icon: Settings,
        description: 'System settings',
      },
      {
        path: ROUTES.ADMIN_CACHE,
        label: 'Cache',
        icon: CircleDot,
        description: 'Cache management',
      },
      {
        path: ROUTES.ADMIN_MONITORING,
        label: 'Monitoring',
        icon: Activity,
        description: 'System monitoring',
      },
      {
        path: ROUTES.ADMIN_MAINTENANCE,
        label: 'Maintenance',
        icon: Wrench,
        description: 'Maintenance mode',
      },
    ],
  },
  {
    id: 'content',
    title: 'Content',
    icon: PanelTop,
    collapsible: true,
    items: [
      { path: ROUTES.ADMIN_CONTENT, label: 'Pages', icon: PanelTop, description: 'Content pages' },
      {
        path: ROUTES.ADMIN_BLOG,
        label: 'Blog',
        icon: BookOpen,
        shortcut: 'B',
        description: 'Blog management',
      },
      {
        path: ROUTES.ADMIN_CHANGELOG,
        label: 'Changelog',
        icon: FileText,
        description: 'Changelog management',
      },
      {
        path: ROUTES.ADMIN_CONTENT_CALENDAR,
        label: 'Calendar',
        icon: Calendar,
        description: 'Content calendar',
      },
      {
        path: ROUTES.ADMIN_REDIRECTS,
        label: 'Redirects',
        icon: RotateCcw,
        description: 'URL redirects',
      },
      {
        path: ROUTES.ADMIN_COMMUNITY_RULES,
        label: 'Community Rules',
        icon: Gavel,
        description: 'Community guidelines',
      },
    ],
  },
  {
    id: 'functions',
    title: 'Functions',
    icon: BarChart3,
    collapsible: true,
    items: [
      {
        path: ROUTES.ADMIN_FUNCTIONS,
        label: 'Functions & Registry',
        icon: FileText,
        description: 'Function management and registry',
      },
      {
        path: ROUTES.ADMIN_STATE_FABRIC,
        label: 'State Fabric',
        icon: CircleDot,
        badge: 'beta',
        description: 'Distributed state',
      },
      {
        path: ROUTES.ADMIN_FACTORY,
        label: 'Factory',
        icon: Factory,
        badge: 'new',
        description: 'Function factory',
      },
    ],
  },
  {
    id: 'trust',
    title: 'Trust & Safety',
    icon: Shield,
    collapsible: true,
    items: [
      {
        path: ROUTES.ADMIN_TRUST_DASHBOARD,
        label: 'Trust Dashboard',
        icon: Shield,
        description: 'Trust metrics',
      },
      {
        path: ROUTES.ADMIN_AUTH_AUDIT,
        label: 'Auth Audit',
        icon: KeyRound,
        description: 'Authentication events',
      },
      {
        path: ROUTES.ADMIN_EXECUTION_AUDIT,
        label: 'Execution Audit',
        icon: BarChart3,
        description: 'Execution logs',
      },
      {
        path: ROUTES.ADMIN_FRAUD_DETECTION,
        label: 'Fraud Detection',
        icon: AlertTriangle,
        badge: 'beta',
        description: 'Fraud monitoring',
      },
      {
        path: ROUTES.ADMIN_IP_ALLOWLIST,
        label: 'IP Allowlist',
        icon: Shield,
        description: 'Restrict access by IP',
      },
      {
        path: ROUTES.ADMIN_SIEM,
        label: 'SIEM',
        icon: AlertTriangle,
        description: 'Security information & event management',
      },
      {
        path: ROUTES.ADMIN_ECONOMIC_LEADERBOARD,
        label: 'Economic',
        icon: TrendingUp,
        description: 'Economic metrics',
      },
    ],
  },
  {
    id: 'communications',
    title: 'Communications',
    icon: MessageSquare,
    collapsible: true,
    items: [
      { path: ROUTES.ADMIN_EMAIL, label: 'Email', icon: Mail, description: 'Email templates' },
      { path: ROUTES.ADMIN_NEWSLETTER, label: 'Newsletter', icon: Mail, description: 'Newsletter campaigns', badge: 'new' },
      {
        path: ROUTES.ADMIN_SUPPORT,
        label: 'Support',
        icon: MessageSquare,
        description: 'Support tickets',
      },
      {
        path: ROUTES.ADMIN_FEEDBACK,
        label: 'Feedback',
        icon: MessageSquare,
        description: 'User feedback',
      },
    ],
  },
  {
    id: 'status',
    title: 'Status',
    icon: Landmark,
    collapsible: true,
    items: [
      {
        path: ROUTES.ADMIN_STATUS,
        label: 'Status Page',
        icon: Landmark,
        description: 'Public status page',
      },
      {
        path: ROUTES.ADMIN_STATUS_INCIDENTS,
        label: 'Incidents',
        icon: AlertTriangle,
        description: 'Status incidents',
      },
      {
        path: ROUTES.ADMIN_CLOUDFLARE_ANALYTICS,
        label: 'Cloudflare',
        icon: Cloud,
        description: 'CDN analytics',
      },
    ],
  },
];

const MD_BREAKPOINT = 768;

export function AdminSidebar({ isOpen, onClose }: AdminSidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const searchInputRef = useRef<HTMLInputElement>(null);

  const [searchQuery, setSearchQuery] = useState('');
  const [isMd, setIsMd] = useState(() => window.innerWidth >= MD_BREAKPOINT);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => {
    // Start with all sections expanded
    return new Set(navigationSections.map((s) => s.id));
  });
  const [focusedIndex, setFocusedIndex] = useState(-1);

  // Handle resize
  useEffect(() => {
    const mq = window.matchMedia(`(min-width: ${MD_BREAKPOINT}px)`);
    const handler = () => setIsMd(mq.matches);
    handler();
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  // Focus search on mobile when sidebar opens
  useEffect(() => {
    if (isOpen && !isMd && searchInputRef.current) {
      setTimeout(() => searchInputRef.current?.focus(), 100);
    }
  }, [isOpen, isMd]);

  // Auto-expand sections with active items
  useEffect(() => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      navigationSections.forEach((section) => {
        const hasActive = section.items.some(
          (item) => location.pathname === item.path || location.pathname.startsWith(`${item.path}/`)
        );
        if (hasActive) {
          next.add(section.id);
        }
      });
      return next;
    });
  }, [location.pathname]);

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

  // Search functionality
  const searchableItems = useMemo(() => {
    return navigationSections.flatMap((section) =>
      section.items.map((item) => ({ ...item, section: section.title }))
    );
  }, []);

  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return [];
    const query = searchQuery.toLowerCase();
    return searchableItems.filter(
      (item) =>
        item.label.toLowerCase().includes(query) ||
        item.section.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
    );
  }, [searchQuery, searchableItems]);

  // Check if item is active
  const isItemActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(`${path}/`);
  };

  // Keyboard shortcuts
  const handleShortcut = useCallback(
    (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey) {
        const shortcut = e.key.toUpperCase();
        const item = searchableItems.find((i) => i.shortcut === shortcut);
        if (item) {
          e.preventDefault();
          navigate(item.path);
        }
      }
    },
    [searchableItems, navigate]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleShortcut);
    return () => document.removeEventListener('keydown', handleShortcut);
  }, [handleShortcut]);

  // Keyboard navigation
  const visibleItems = useMemo(() => {
    if (searchQuery && searchResults.length > 0) return searchResults;
    return navigationSections.flatMap((s) => (expandedSections.has(s.id) ? s.items : []));
  }, [searchQuery, searchResults, expandedSections]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (!isOpen) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setFocusedIndex((prev) => (prev >= visibleItems.length - 1 ? 0 : prev + 1));
          break;
        case 'ArrowUp':
          e.preventDefault();
          setFocusedIndex((prev) => (prev <= 0 ? visibleItems.length - 1 : prev - 1));
          break;
        case 'Enter':
          if (focusedIndex >= 0 && focusedIndex < visibleItems.length) {
            const item = visibleItems[focusedIndex];
            if (item) {
              navigate(item.path);
              setSearchQuery('');
              onClose();
            }
          }
          break;
        case 'Escape':
          if (searchQuery) {
            setSearchQuery('');
          } else {
            onClose();
          }
          break;
      }
    },
    [isOpen, visibleItems, focusedIndex, navigate, searchQuery, onClose]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <>
      {/* Mobile Overlay */}
      {isOpen && !isMd && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 md:hidden transition-opacity duration-200"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed left-0 top-0 z-50 h-screen w-[280px] min-w-[280px]',
          'bg-gray-900 text-white',
          'border-r border-gray-800',
          'flex flex-col',
          'transition-transform duration-300 ease-out',
          isOpen || isMd ? 'translate-x-0' : '-translate-x-full',
          'md:static md:translate-x-0 md:self-stretch'
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800">
          <div className="flex items-center gap-3">
            <img
              src="/favicon.svg"
              alt=""
              width="28"
              height="28"
              className="shrink-0"
            />
            <div>
              <span className="text-sm font-bold text-white">FunctionFly</span>
              <span className="text-[10px] text-gray-400 block">Admin</span>
            </div>
          </div>

          <div className="flex items-center gap-1">
            {/* Command Palette Trigger */}
            <button
              onClick={() => {
                document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }));
              }}
              className="hidden md:flex items-center gap-1 px-2 py-1 rounded bg-gray-800 border border-gray-700 text-gray-400 hover:text-white hover:border-indigo-500/50 transition-colors"
            >
              <Command className="w-3 h-3" />
              <span className="text-[10px]">K</span>
            </button>

            <button
              onClick={onClose}
              className="md:hidden p-2 hover:bg-gray-800 rounded-lg transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Search */}
        <div className="px-3 py-3 border-b border-gray-800">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              ref={searchInputRef}
              type="text"
              placeholder="Search..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-sm text-white placeholder:text-gray-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-all"
            />
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 min-h-0 overflow-y-auto py-3">
          {/* Search Results */}
          {searchQuery && (
            <div className="px-3 pb-3 animate-in fade-in slide-in-from-top-2 duration-200">
              <p className="px-3 text-xs font-medium text-gray-500 mb-2">Search Results</p>
              {searchResults.length > 0 ? (
                <div className="space-y-1">
                  {searchResults.map((item, index) => {
                    const isActive = isItemActive(item.path);
                    const Icon = item.icon;
                    const isFocused = focusedIndex === index;

                    return (
                      <Link
                        key={`search-${item.path}`}
                        to={item.path}
                        onClick={() => {
                          onClose();
                          setSearchQuery('');
                        }}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all',
                          isActive
                            ? 'bg-indigo-500/20 text-indigo-400'
                            : isFocused
                              ? 'bg-gray-800 text-white ring-2 ring-indigo-500/50'
                              : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                        )}
                      >
                        <Icon className={cn('w-4 h-4', isActive && 'text-indigo-400')} />
                        <div className="flex-1 min-w-0">
                          <span className="block truncate">{item.label}</span>
                          <span className="text-xs text-gray-500 block truncate">
                            {item.section}
                          </span>
                        </div>
                      </Link>
                    );
                  })}
                </div>
              ) : (
                <div className="px-3 py-8 text-center">
                  <Search className="w-8 h-8 text-gray-600 mx-auto mb-2" />
                  <p className="text-sm text-gray-500">No results found</p>
                </div>
              )}
            </div>
          )}

          {/* Navigation Sections */}
          {!searchQuery && (
            <div className="space-y-1 px-3">
              {navigationSections.map((section) => {
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
                          'flex items-center justify-between w-full px-3 py-2 rounded-lg transition-all',
                          'hover:bg-gray-800',
                          hasActiveItem ? 'text-white' : 'text-gray-400'
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <SectionIcon className="w-4 h-4" />
                          <span className="text-xs font-semibold uppercase tracking-wider">
                            {section.title}
                          </span>
                        </div>
                        <div
                          className={cn(
                            'transition-transform duration-200',
                            isExpanded ? 'rotate-0' : '-rotate-90'
                          )}
                        >
                          <ChevronDown className="w-3.5 h-3.5" />
                        </div>
                      </button>
                    ) : (
                      <div className="px-3 py-2 text-xs font-semibold uppercase tracking-wider text-gray-400">
                        {section.title}
                      </div>
                    )}

                    {/* Section Items */}
                    <div
                      className={cn(
                        'overflow-hidden transition-all duration-200 ease-in-out',
                        isExpanded ? 'max-h-[500px] opacity-100' : 'max-h-0 opacity-0'
                      )}
                    >
                      <div className="space-y-0.5 pt-1">
                        {section.items.map((item, itemIndex) => {
                          const isActive = isItemActive(item.path);
                          const Icon = item.icon;
                          const globalIndex =
                            navigationSections
                              .slice(0, navigationSections.indexOf(section))
                              .reduce(
                                (acc, s) => acc + (expandedSections.has(s.id) ? s.items.length : 0),
                                0
                              ) + itemIndex;
                          const isFocused = focusedIndex === globalIndex;

                          return (
                            <Link
                              key={item.path}
                              to={item.path}
                              onClick={() => onClose()}
                              className={cn(
                                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all relative group',
                                isActive
                                  ? 'bg-indigo-500/20 text-indigo-400'
                                  : isFocused
                                    ? 'bg-gray-800 text-white ring-2 ring-indigo-500/50'
                                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                              )}
                            >
                              {/* Active indicator */}
                              {isActive && (
                                <div className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-full bg-indigo-500" />
                              )}

                              <Icon className={cn('w-4 h-4', isActive && 'text-indigo-400')} />
                              <span className="flex-1 font-medium truncate">{item.label}</span>

                              {/* Badge */}
                              {item.badge && (
                                <span
                                  className={cn(
                                    'text-[9px] font-bold px-1.5 py-0.5 rounded-full uppercase',
                                    item.badge === 'new'
                                      ? 'bg-green-500/20 text-green-400'
                                      : 'bg-blue-500/20 text-blue-400'
                                  )}
                                >
                                  {item.badge}
                                </span>
                              )}

                              {/* Shortcut hint */}
                              {item.shortcut && (
                                <kbd className="text-[10px] font-mono text-gray-500 bg-gray-800 px-1.5 py-0.5 rounded hidden group-hover:inline-block">
                                  ⌘{item.shortcut}
                                </kbd>
                              )}
                            </Link>
                          );
                        })}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </nav>

        {/* Footer */}
        <div className="p-3 border-t border-gray-800">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <span>FunctionFly Admin</span>
            <span>v2.0</span>
          </div>
        </div>
      </aside>
    </>
  );
}

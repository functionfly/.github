/**
 * Admin Header Component - Production Ready
 * Breadcrumbs, quick actions, dropdown menus, notifications
 */

import { DarkModeToggle } from '@/components/common/DarkModeToggle';
import { ROUTES } from '@/lib/constants';
import { cn } from '@/lib/utils';
import type { AdminUser } from '@/types';
import {
  Bell,
  ChevronDown,
  Command,
  HelpCircle,
  Keyboard,
  LifeBuoy,
  LogOut,
  Menu,
  Settings,
  Shield,
  User,
  type LucideIcon,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';

interface AdminHeaderProps {
  user: AdminUser | null;
  onMenuClick: () => void;
  onLogout: () => void;
  /** Show sidebar toggle (only when logged in and sidebar is present) */
  showMenuButton?: boolean;
}

interface BreadcrumbItem {
  label: string;
  path?: string;
}

interface QuickAction {
  label: string;
  icon: LucideIcon;
  href: string;
  shortcut?: string;
}

const quickActions: QuickAction[] = [
  { label: 'Dashboard', icon: Command, href: ROUTES.ADMIN_DASHBOARD, shortcut: 'H' },
  { label: 'Tenants', icon: Command, href: ROUTES.ADMIN_TENANTS, shortcut: 'T' },
  { label: 'Users', icon: Command, href: ROUTES.ADMIN_USERS, shortcut: 'U' },
  { label: 'Blog', icon: Command, href: ROUTES.ADMIN_BLOG, shortcut: 'B' },
];

// Route to breadcrumb mapping
const breadcrumbMap: Record<string, BreadcrumbItem[]> = {
  [ROUTES.ADMIN_DASHBOARD]: [{ label: 'Dashboard' }],
  [ROUTES.ADMIN_TENANTS]: [
    { label: 'Dashboard', path: ROUTES.ADMIN_DASHBOARD },
    { label: 'Tenants' },
  ],
  [ROUTES.ADMIN_USERS]: [{ label: 'Dashboard', path: ROUTES.ADMIN_DASHBOARD }, { label: 'Users' }],
  [ROUTES.ADMIN_BILLING]: [
    { label: 'Dashboard', path: ROUTES.ADMIN_DASHBOARD },
    { label: 'Billing' },
  ],
  [ROUTES.ADMIN_AUDIT]: [
    { label: 'Dashboard', path: ROUTES.ADMIN_DASHBOARD },
    { label: 'Audit Log' },
  ],
  [ROUTES.ADMIN_SYSTEM]: [{ label: 'Infrastructure' }, { label: 'System' }],
  [ROUTES.ADMIN_BACKENDS]: [{ label: 'Infrastructure' }, { label: 'Backends' }],
  [ROUTES.ADMIN_PROVIDERS]: [{ label: 'Infrastructure' }, { label: 'Providers' }],
  [ROUTES.ADMIN_CACHE]: [{ label: 'Infrastructure' }, { label: 'Cache' }],
  [ROUTES.ADMIN_MONITORING]: [{ label: 'Infrastructure' }, { label: 'Monitoring' }],
  [ROUTES.ADMIN_MAINTENANCE]: [{ label: 'Infrastructure' }, { label: 'Maintenance' }],
  [ROUTES.ADMIN_CONTENT]: [{ label: 'Content' }, { label: 'Pages' }],
  [ROUTES.ADMIN_BLOG]: [{ label: 'Content' }, { label: 'Blog' }],
  [ROUTES.ADMIN_CONTENT_CALENDAR]: [{ label: 'Content' }, { label: 'Calendar' }],
  [ROUTES.ADMIN_REDIRECTS]: [{ label: 'Content' }, { label: 'Redirects' }],
  [ROUTES.ADMIN_FUNCTIONS]: [{ label: 'Functions' }, { label: 'Management' }],
  [ROUTES.ADMIN_STATE_FABRIC]: [{ label: 'Functions' }, { label: 'State Fabric' }],
  [ROUTES.ADMIN_FACTORY]: [{ label: 'Functions' }, { label: 'Factory' }],
  [ROUTES.ADMIN_TRUST_DASHBOARD]: [{ label: 'Trust & Safety' }, { label: 'Dashboard' }],
  [ROUTES.ADMIN_EXECUTION_AUDIT]: [{ label: 'Trust & Safety' }, { label: 'Execution Audit' }],
  [ROUTES.ADMIN_FRAUD_DETECTION]: [{ label: 'Trust & Safety' }, { label: 'Fraud Detection' }],
  [ROUTES.ADMIN_ECONOMIC_LEADERBOARD]: [{ label: 'Trust & Safety' }, { label: 'Economic' }],
  [ROUTES.ADMIN_STATUS]: [{ label: 'Status' }, { label: 'Status Page' }],
  [ROUTES.ADMIN_STATUS_INCIDENTS]: [{ label: 'Status' }, { label: 'Incidents' }],
  [ROUTES.ADMIN_CLOUDFLARE_ANALYTICS]: [{ label: 'Status' }, { label: 'Cloudflare' }],
  [ROUTES.ADMIN_EMAIL]: [{ label: 'Communications' }, { label: 'Email' }],
  [ROUTES.ADMIN_SUPPORT]: [{ label: 'Communications' }, { label: 'Support' }],
  [ROUTES.ADMIN_FEEDBACK]: [{ label: 'Communications' }, { label: 'Feedback' }],
};

function displayName(user: AdminUser): string {
  if (user.name?.trim()) return user.name;
  if (user.email) return user.email.split('@')[0].replace(/[._]/g, ' ');
  return 'Admin';
}

function displayRole(role: string): string {
  return role.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export function AdminHeader({
  user,
  onMenuClick,
  onLogout,
  showMenuButton = true,
}: AdminHeaderProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [notifications, setNotifications] = useState([
    {
      id: 1,
      title: 'New user signup',
      message: 'user@example.com just signed up',
      time: '2 min ago',
      unread: true,
    },
    {
      id: 2,
      title: 'System alert',
      message: 'High CPU usage on backend-3',
      time: '15 min ago',
      unread: true,
    },
    {
      id: 3,
      title: 'Deployment complete',
      message: 'v2.4.1 deployed successfully',
      time: '1 hour ago',
      unread: false,
    },
  ]);

  // Dropdown states
  const [helpOpen, setHelpOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const [userOpen, setUserOpen] = useState(false);

  const helpRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);
  const userRef = useRef<HTMLDivElement>(null);

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (helpRef.current && !helpRef.current.contains(e.target as Node)) {
        setHelpOpen(false);
      }
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setNotifOpen(false);
      }
      if (userRef.current && !userRef.current.contains(e.target as Node)) {
        setUserOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const unreadCount = notifications.filter((n) => n.unread).length;

  // Generate breadcrumbs based on current route
  const breadcrumbs = useMemo(() => {
    // Try exact match first
    if (breadcrumbMap[location.pathname]) {
      return breadcrumbMap[location.pathname];
    }

    // Check for dynamic routes (e.g., /users/123)
    for (const [route, crumbs] of Object.entries(breadcrumbMap)) {
      if (location.pathname.startsWith(route + '/')) {
        // Add the dynamic segment as the last breadcrumb
        const lastSegment = location.pathname.split('/').pop();
        return [...crumbs, { label: lastSegment || 'Detail' }];
      }
    }

    // Fallback: generate from path
    const segments = location.pathname.split('/').filter(Boolean);
    return segments.map((segment, index) => ({
      label: segment.charAt(0).toUpperCase() + segment.slice(1),
      path: index < segments.length - 1 ? '/' + segments.slice(0, index + 1).join('/') : undefined,
    }));
  }, [location.pathname]);

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Command palette: Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setShowCommandPalette((prev) => !prev);
      }

      // Quick navigation
      if (e.metaKey || e.ctrlKey) {
        const action = quickActions.find((a) => a.shortcut === e.key.toUpperCase());
        if (action) {
          e.preventDefault();
          navigate(action.href);
        }
      }

      // Close command palette on escape
      if (e.key === 'Escape' && showCommandPalette) {
        setShowCommandPalette(false);
      }
    },
    [navigate, showCommandPalette]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  const markAllRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, unread: false })));
  };

  return (
    <>
      <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 shadow-sm shrink-0 sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-14">
            {/* Left: Menu button + Breadcrumbs */}
            <div className="flex items-center gap-4">
              {showMenuButton && (
                <button
                  onClick={onMenuClick}
                  className="md:hidden p-2 -ml-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                  aria-label="Toggle menu"
                >
                  <Menu className="w-5 h-5 text-gray-600 dark:text-gray-300" />
                </button>
              )}

              {/* Breadcrumbs */}
              <nav className="hidden sm:flex items-center text-sm">
                <ol className="flex items-center gap-2">
                  {breadcrumbs.map((crumb, index) => (
                    <li key={index} className="flex items-center gap-2">
                      {index > 0 && <ChevronDown className="w-3 h-3 text-gray-400 -rotate-90" />}
                      {crumb.path ? (
                        <Link
                          to={crumb.path}
                          className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
                        >
                          {crumb.label}
                        </Link>
                      ) : (
                        <span
                          className={cn(
                            'font-medium',
                            index === breadcrumbs.length - 1 ? 'text-gray-900 dark:text-white' : 'text-gray-500 dark:text-gray-400'
                          )}
                        >
                          {crumb.label}
                        </span>
                      )}
                    </li>
                  ))}
                </ol>
              </nav>

              {/* Mobile page title */}
              <h1 className="sm:hidden text-lg font-semibold text-gray-900 dark:text-white">
                {breadcrumbs[breadcrumbs.length - 1]?.label || 'Admin'}
              </h1>
            </div>

            {/* Right: Actions + User */}
            <div className="flex items-center gap-2">
              {user && (
                <>
                  {/* Command Palette Trigger */}
                  <button
                    onClick={() => setShowCommandPalette(true)}
                    className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 border border-gray-200 dark:border-gray-600 rounded-lg transition-colors"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span>Search</span>
                    <kbd className="text-[10px] font-mono text-gray-400 dark:text-gray-500 bg-gray-200 dark:bg-gray-600 px-1.5 py-0.5 rounded">
                      ⌘K
                    </kbd>
                  </button>

                  {/* Help Dropdown */}
                  <div className="relative" ref={helpRef}>
                    <button
                      onClick={() => setHelpOpen(!helpOpen)}
                      className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                    >
                      <HelpCircle className="w-5 h-5 text-gray-600 dark:text-gray-300" />
                    </button>
                    {helpOpen && (
                      <div className="absolute right-0 mt-2 w-56 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50">
                        <div className="px-3 py-2 text-sm font-semibold text-gray-700 dark:text-gray-200">
                          Help & Resources
                        </div>
                        <hr className="my-1 border-gray-100 dark:border-gray-700" />
                        <button
                          onClick={() => {
                            setHelpOpen(false);
                            setShowCommandPalette(true);
                          }}
                          className="w-full px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                        >
                          <Command className="w-4 h-4" />
                          Command Palette
                          <kbd className="ml-auto text-[10px] font-mono text-gray-400">⌘K</kbd>
                        </button>
                        <button className="w-full px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2">
                          <Keyboard className="w-4 h-4" />
                          Keyboard Shortcuts
                        </button>
                        <button className="w-full px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2">
                          <LifeBuoy className="w-4 h-4" />
                          Support Center
                        </button>
                        <hr className="my-1 border-gray-100 dark:border-gray-700" />
                        <div className="px-3 py-2">
                          <p className="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Quick Navigation</p>
                          <div className="space-y-1">
                            {quickActions.map((action) => (
                              <div
                                key={action.label}
                                className="flex justify-between text-xs text-gray-600 dark:text-gray-400"
                              >
                                <span>{action.label}</span>
                                <kbd className="font-mono text-gray-400">⌘{action.shortcut}</kbd>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Notifications */}
                  <div className="relative" ref={notifRef}>
                    <button
                      onClick={() => setNotifOpen(!notifOpen)}
                      className="relative p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                    >
                      <Bell className="w-5 h-5 text-gray-600 dark:text-gray-300" />
                      {unreadCount > 0 && (
                        <span className="absolute top-1 right-1 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
                          {unreadCount > 9 ? '9+' : unreadCount}
                        </span>
                      )}
                    </button>
                    {notifOpen && (
                      <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50">
                        <div className="flex items-center justify-between px-3 py-2 border-b border-gray-100 dark:border-gray-700">
                          <span className="font-medium text-sm text-gray-900 dark:text-white">Notifications</span>
                          {unreadCount > 0 && (
                            <button
                              onClick={markAllRead}
                              className="text-xs text-indigo-600 hover:text-indigo-700 dark:text-indigo-400"
                            >
                              Mark all read
                            </button>
                          )}
                        </div>
                        <div className="max-h-64 overflow-y-auto">
                          {notifications.length > 0 ? (
                            notifications.map((notification) => (
                              <div
                                key={notification.id}
                                onClick={() => {
                                  setNotifications((prev) =>
                                    prev.map((n) =>
                                      n.id === notification.id ? { ...n, unread: false } : n
                                    )
                                  );
                                }}
                                className={cn(
                                  'px-3 py-3 border-b border-gray-50 dark:border-gray-700 last:border-0 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer',
                                  notification.unread && 'bg-indigo-50/50 dark:bg-indigo-900/20'
                                )}
                              >
                                <div className="flex items-start gap-3">
                                  <div
                                    className={cn(
                                      'w-2 h-2 rounded-full mt-1.5 shrink-0',
                                      notification.unread ? 'bg-indigo-500' : 'bg-gray-300 dark:bg-gray-600'
                                    )}
                                  />
                                  <div className="flex-1 min-w-0">
                                    <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                                      {notification.title}
                                    </p>
                                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                                      {notification.message}
                                    </p>
                                    <p className="text-[10px] text-gray-400 dark:text-gray-500 mt-1">
                                      {notification.time}
                                    </p>
                                  </div>
                                </div>
                              </div>
                            ))
                          ) : (
                            <div className="px-3 py-8 text-center text-gray-500 dark:text-gray-400">
                              <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                              <p className="text-sm">No notifications</p>
                            </div>
                          )}
                        </div>
                        <div className="px-3 py-2 border-t border-gray-100 dark:border-gray-700">
                          <Link
                            to="#"
                            className="block text-center text-xs text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 font-medium"
                          >
                            View all notifications
                          </Link>
                        </div>
                      </div>
                    )}
                  </div>
                </>
              )}

              {/* Dark Mode Toggle */}
              <div className="hidden sm:block">
                <DarkModeToggle />
              </div>

              {/* User Dropdown */}
              {user ? (
                <div className="relative" ref={userRef}>
                  <button
                    onClick={() => setUserOpen(!userOpen)}
                    className="flex items-center gap-2 p-1 pr-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors ml-2"
                  >
                    <div className="w-8 h-8 rounded-full bg-linear-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-sm font-semibold">
                      {user.name?.trim()
                        ? user.name.charAt(0).toUpperCase()
                        : (user.email?.charAt(0) || 'A').toUpperCase()}
                    </div>
                    <div className="hidden md:block text-left">
                      <p className="text-sm font-medium text-gray-900 dark:text-white leading-tight">
                        {displayName(user)}
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-400 leading-tight">
                        {displayRole(user.role)}
                      </p>
                    </div>
                    <ChevronDown
                      className={cn(
                        'w-4 h-4 text-gray-400 hidden md:block transition-transform',
                        userOpen && 'rotate-180'
                      )}
                    />
                  </button>
                  {userOpen && (
                    <div className="absolute right-0 mt-2 w-56 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50">
                      <div className="px-3 py-2">
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{displayName(user)}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">{user.email}</p>
                      </div>
                      <hr className="my-1 border-gray-100 dark:border-gray-700" />
                      <Link
                        to={`${ROUTES.ADMIN_USERS}/${user.id}`}
                        className="flex items-center gap-2 px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                        onClick={() => setUserOpen(false)}
                      >
                        <User className="w-4 h-4" />
                        Profile
                      </Link>
                      <Link
                        to={ROUTES.ADMIN_SYSTEM}
                        className="flex items-center gap-2 px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                        onClick={() => setUserOpen(false)}
                      >
                        <Settings className="w-4 h-4" />
                        Settings
                      </Link>
                      <button
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                        onClick={() => setUserOpen(false)}
                      >
                        <Shield className="w-4 h-4" />
                        Security
                      </button>
                      <hr className="my-1 border-gray-100 dark:border-gray-700" />
                      <button
                        onClick={() => {
                          setUserOpen(false);
                          onLogout();
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
                      >
                        <LogOut className="w-4 h-4" />
                        Sign out
                      </button>
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400">
                  <User className="w-5 h-5" />
                  <p className="text-sm font-medium">Sign in</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Command Palette Overlay */}
      {showCommandPalette && (
        <div
          className="fixed inset-0 bg-black/50 dark:bg-black/70 z-50 flex items-start justify-center pt-[20vh]"
          onClick={() => setShowCommandPalette(false)}
        >
          <div
            className="w-full max-w-lg bg-white dark:bg-gray-800 rounded-xl shadow-2xl overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-100 dark:border-gray-700">
              <Command className="w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search pages, actions, or users..."
                className="flex-1 text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none"
                autoFocus
              />
              <kbd className="text-[10px] font-mono text-gray-400 dark:text-gray-500 bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">
                ESC
              </kbd>
            </div>
            <div className="py-2">
              <div className="px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                Quick Navigation
              </div>
              {quickActions.map((action) => (
                <button
                  key={action.label}
                  onClick={() => {
                    navigate(action.href);
                    setShowCommandPalette(false);
                  }}
                  className="w-full flex items-center justify-between px-4 py-2.5 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <action.icon className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                    <span className="text-sm text-gray-900 dark:text-gray-100">{action.label}</span>
                  </div>
                  <kbd className="text-[10px] font-mono text-gray-400 dark:text-gray-500 bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">
                    ⌘{action.shortcut}
                  </kbd>
                </button>
              ))}
            </div>
            <div className="px-4 py-2 bg-gray-50 dark:bg-gray-900 text-xs text-gray-500 dark:text-gray-400">
              <p>
                Use <kbd className="font-mono bg-gray-200 dark:bg-gray-700 px-1 rounded">↑</kbd>{' '}
                <kbd className="font-mono bg-gray-200 dark:bg-gray-700 px-1 rounded">↓</kbd> to navigate,{' '}
                <kbd className="font-mono bg-gray-200 dark:bg-gray-700 px-1 rounded">↵</kbd> to select
              </p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

import { Logo } from '@/components/common/Logo';
import { MarketplaceDropdown } from '@/components/common/MarketplaceDropdown';
import { ProvidersDropdown } from '@/components/common/ProvidersDropdown';
import { ThemeToggle } from '@/components/common/ThemeToggle';
import { FrameButton, SealedButton } from '@/components/containment';
import { UserMenu } from '@/components/layout/UserMenu';
import { NotificationBell } from '@/components/notifications';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { DOCS_SITE_URL, getMarketingRedirectOrigin, PROVIDERS } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useProvidersStore } from '@/stores/providersStore';
import { useThemeStore } from '@/stores/themeStore';
import '@/styles/sc-navbar.css';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Bot,
  Cloud,
  Command,
  CreditCard,
  FunctionSquare,
  Home,
  Menu,
  MessageCircle,
  ShoppingBag,
  Sparkles,
  X,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';

interface NavbarProps {
  variant?: 'landing' | 'dashboard';
  className?: string;
  onMenuClick?: () => void;
}

// Quick action shortcuts with actions
interface QuickAction {
  key: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  action: (
    navigate: ReturnType<typeof useNavigate>,
    setShowCommandPalette: (show: boolean) => void
  ) => void;
}

const QUICK_ACTIONS: QuickAction[] = [
  {
    key: 'g',
    label: 'Go to...',
    icon: Command,
    action: (_, setShow) => setShow(true),
  },
  {
    key: 'n',
    label: 'New Function',
    icon: Sparkles,
    action: (navigate, setShow) => {
      setShow(false);
      navigate('/functions/new');
    },
  },
  {
    key: 'a',
    label: 'Agents',
    icon: Zap,
    action: (navigate, setShow) => {
      setShow(false);
      navigate('/marketplace?type=agents');
    },
  },
];

export function Navbar({ variant = 'landing', className, onMenuClick }: NavbarProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const paletteRef = useRef<HTMLDivElement>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const theme = useThemeStore((state) => state.theme);
  const messagesUnread = useNotificationStore((state) => state.unreadCounts.messages);
  const marketingHomeUrl = getMarketingRedirectOrigin();
  const connectedProviders = useProvidersStore((state) => state.providers);
  const fetchProviders = useProvidersStore((state) => state.fetchProviders);
  const totalProviders = Object.keys(PROVIDERS).length;
  const connectedCount = connectedProviders.length;

  useEffect(() => {
    if (isAuthenticated && variant === 'dashboard') {
      fetchProviders();
    }
  }, [isAuthenticated, variant, fetchProviders]);

  // Scroll-aware background
  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', handler);
    return () => window.removeEventListener('scroll', handler);
  }, []);

  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Skip if focus is in an interactive element (search, code editor, etc.)
      const target = e.target as HTMLElement;
      const isInteractive =
        target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;

      // Command palette: Cmd/Ctrl + K (skip if in an input)
      if ((e.metaKey || e.ctrlKey) && e.key === 'k' && !isInteractive) {
        e.preventDefault();
        setShowCommandPalette(true);
      }

      // Close on escape
      if (e.key === 'Escape' && showCommandPalette) {
        setShowCommandPalette(false);
      }
    },
    [showCommandPalette]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  // Focus trap within command palette
  useEffect(() => {
    if (!showCommandPalette) return;

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const focusable = paletteRef.current?.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      if (!focusable || focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', handleTab);
    return () => document.removeEventListener('keydown', handleTab);
  }, [showCommandPalette]);

  return (
    <TooltipProvider delayDuration={0}>
      <>
        <nav
          className={cn(
            'top-0 left-0 right-0 z-50 transition-all duration-300 sc-navbar',
            className
          )}
        >
          <div className="max-w-7xl mx-auto px-4 lg:px-6 h-16 flex items-center justify-between gap-4">
            {/* Left: menu button + logo */}
            <div className="flex items-center shrink-0 min-w-0">
              {/* Mobile Menu Button (Dashboard only) */}
              {variant === 'dashboard' && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="lg:hidden text-text-secondary hover:text-text-primary hover:bg-bg-secondary mr-2"
                      onClick={onMenuClick}
                      aria-label="Open navigation menu"
                    >
                      <Menu className="w-5 h-5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent
                    side="bottom"
                    className="bg-bg-secondary border border-border-subtle shadow-lg"
                  >
                    <p className="text-text-primary">Open sidebar</p>
                  </TooltipContent>
                </Tooltip>
              )}

              {/* Logo — hide on desktop in dashboard layout because the sidebar already shows it */}
              {isAuthenticated ? (
                <Link
                  to="/dashboard"
                  className={cn('shrink-0 mr-4 md:mr-6', variant === 'dashboard' && 'lg:hidden')}
                  aria-label="FunctionFly home"
                >
                  <Logo />
                </Link>
              ) : (
                <a
                  href={marketingHomeUrl}
                  className={cn('shrink-0 mr-4 md:mr-6', variant === 'dashboard' && 'lg:hidden')}
                  aria-label="FunctionFly home"
                >
                  <Logo />
                </a>
              )}

              {/* Breadcrumbs - shown on nested pages */}
              {variant === 'dashboard' &&
                location.pathname.split('/').filter(Boolean).length > 1 && (
                  <nav className="hidden lg:flex items-center gap-2 text-sm text-text-muted ml-2">
                    <Link to="/dashboard" className="hover:text-text-primary transition-colors">
                      Home
                    </Link>
                    {location.pathname
                      .split('/')
                      .filter(Boolean)
                      .slice(0, -1)
                      .map((segment, i) => {
                        const path =
                          '/' +
                          location.pathname
                            .split('/')
                            .filter(Boolean)
                            .slice(0, i + 1)
                            .join('/');
                        return (
                          <span key={i} className="flex items-center gap-2">
                            <span className="text-text-muted">/</span>
                            <Link
                              to={path}
                              className="hover:text-text-primary transition-colors capitalize"
                            >
                              {segment}
                            </Link>
                          </span>
                        );
                      })}
                  </nav>
                )}
            </div>

            {/* Desktop Navigation - Authenticated */}
            <div className="hidden md:flex items-center gap-5">
              {isAuthenticated ? (
                <>
                  {/* Dashboard - direct link */}
                  <Link
                    to="/dashboard"
                    className={cn(
                      'sc-navbar-link-indicator',
                      location.pathname === '/dashboard' && 'active'
                    )}
                  >
                    Dashboard
                    {location.pathname === '/dashboard' && <span className="active-bar" />}
                  </Link>

                  {/* Functions - direct link */}
                  <Link
                    to="/functions/my"
                    className={cn(
                      'sc-navbar-link-indicator',
                      (location.pathname === '/functions' ||
                        location.pathname.startsWith('/functions/')) &&
                        'active'
                    )}
                  >
                    Functions
                    {(location.pathname === '/functions' ||
                      location.pathname.startsWith('/functions/')) && (
                      <span className="active-bar" />
                    )}
                  </Link>

                  {/* Marketplace - dropdown */}
                  <MarketplaceDropdown />

                  {/* Providers - dropdown */}
                  <ProvidersDropdown />
                </>
              ) : (
                <>
                  {/* Unauthenticated nav */}
                  <Link
                    to="/"
                    className={cn(
                      'sc-navbar-link-indicator',
                      location.pathname === '/' && 'active'
                    )}
                  >
                    Home
                    {location.pathname === '/' && <span className="active-bar" />}
                  </Link>

                  {/* Functions - direct link */}
                  <Link
                    to="/marketplace?type=functions"
                    className={cn(
                      'sc-navbar-link-indicator',
                      location.pathname === '/marketplace' && location.search.includes('type=functions') && 'active'
                    )}
                  >
                    Functions
                    {location.pathname === '/marketplace' && location.search.includes('type=functions') && (
                      <span className="active-bar" />
                    )}
                  </Link>

                  {/* Marketplace - dropdown */}
                  <MarketplaceDropdown />

                  <Link
                    to="/pricing"
                    className={cn(
                      'sc-navbar-link-indicator',
                      location.pathname === '/pricing' && 'active'
                    )}
                  >
                    Pricing
                    {location.pathname === '/pricing' && <span className="active-bar" />}
                  </Link>
                  <a
                    href={DOCS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-navbar-link"
                  >
                    Docs
                  </a>
                </>
              )}
            </div>

            {/* Right Section */}
            <div className="flex items-center gap-2 shrink-0">
              {/* Provider Health Status - Authenticated Dashboard only */}
              {isAuthenticated && variant === 'dashboard' && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Link
                      to="/providers"
                      className="hidden lg:flex items-center gap-1.5 px-2 py-1 rounded-full bg-bg-secondary/30 border border-border-subtle hover:bg-bg-secondary/50 transition-colors cursor-pointer"
                    >
                      <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                      <span className="text-xs text-text-secondary">
                        {connectedCount}/{totalProviders} Providers
                      </span>
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent
                    side="bottom"
                    className="bg-bg-secondary border border-border-subtle shadow-lg"
                  >
                    <p className="text-text-primary font-medium">
                      {connectedCount === totalProviders
                        ? 'All providers connected'
                        : `${connectedCount} of ${totalProviders} providers connected`}
                    </p>
                    <p className="text-xs text-text-muted mt-1">
                      {Object.values(PROVIDERS)
                        .map((p) => p.name)
                        .join(' • ')}
                    </p>
                  </TooltipContent>
                </Tooltip>
              )}

              {/* Command Palette Trigger - Desktop */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setShowCommandPalette(true)}
                    className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-text-muted bg-bg-secondary/50 border border-border-subtle rounded-lg hover:text-text-primary hover:border-warning/30 transition-all"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span className="hidden lg:inline">Search</span>
                    <kbd className="hidden xl:inline text-[10px] font-mono text-text-muted bg-bg-secondary px-1.5 py-0.5 rounded">
                      ⌘K
                    </kbd>
                  </button>
                </TooltipTrigger>
                <TooltipContent
                  side="bottom"
                  className="bg-bg-secondary border border-border-subtle shadow-lg"
                >
                  <p className="text-text-primary">Command Palette</p>
                </TooltipContent>
              </Tooltip>

              {isAuthenticated ? (
                <>
                  {/* Theme Toggle */}
                  <ThemeToggle />

                  {/* Messages (Dashboard only) */}
                  {variant === 'dashboard' && user?.username && (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          to={`/u/${user.username}/conversations`}
                          aria-label={
                            messagesUnread > 0 ? `Messages (${messagesUnread} unread)` : 'Messages'
                          }
                          className="relative flex items-center justify-center rounded-lg p-2 text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
                        >
                          <MessageCircle className="w-5 h-5" />
                          {messagesUnread > 0 && (
                            <span className="absolute -top-0.5 -right-0.5 flex min-h-[18px] min-w-[18px] items-center justify-center rounded-full bg-error px-1 text-[10px] font-bold leading-none text-white">
                              {messagesUnread > 99 ? '99+' : messagesUnread}
                            </span>
                          )}
                        </Link>
                      </TooltipTrigger>
                      <TooltipContent
                        side="bottom"
                        className="bg-bg-secondary border border-border-subtle shadow-lg"
                      >
                        <p className="text-text-primary">Messages</p>
                        {messagesUnread > 0 && (
                          <p className="text-xs text-text-muted">{messagesUnread} unread</p>
                        )}
                      </TooltipContent>
                    </Tooltip>
                  )}

                  {/* Notifications (Dashboard only) */}
                  {variant === 'dashboard' && (
                    <NotificationBell
                      variant="ghost"
                      size="md"
                      className="relative text-text-secondary hover:text-text-primary hover:bg-bg-secondary"
                    />
                  )}

                  {/* User Menu - includes Settings */}
                  <UserMenu />
                </>
              ) : (
                <>
                  {/* Theme Toggle */}
                  <ThemeToggle />

                  {/* Auth Buttons */}
                  <Link to="/login">
                    <FrameButton size="sm" className="hidden sm:inline-flex">
                      Sign In
                    </FrameButton>
                  </Link>
                  <Link to="/signup">
                    <SealedButton size="sm">Get Started</SealedButton>
                  </Link>
                </>
              )}

              {/* Mobile Menu Toggle */}
              <Button
                variant="ghost"
                size="icon"
                className="md:hidden text-text-secondary hover:text-text-primary hover:bg-bg-secondary"
                onClick={toggleMobileMenu}
                aria-label={isMobileMenuOpen ? 'Close navigation menu' : 'Open navigation menu'}
                aria-expanded={isMobileMenuOpen}
              >
                {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
              </Button>
            </div>
          </div>
        </nav>

        {/* Mobile Menu */}
        <AnimatePresence>
          {isMobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40 md:hidden"
            >
              {/* Overlay */}
              <div
                className="fixed inset-0 bg-black/60 backdrop-blur-sm"
                onClick={() => setIsMobileMenuOpen(false)}
              />

              {/* Menu */}
              <motion.div
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="fixed top-16 left-0 right-0 bg-bg-primary border-b border-border-default shadow-xl"
              >
                <div className="px-4 py-4 space-y-4 max-h-[70vh] overflow-y-auto">
                  {/* Search for mobile */}
                  <button
                    type="button"
                    onClick={() => {
                      setIsMobileMenuOpen(false);
                      setShowCommandPalette(true);
                    }}
                    className="w-full flex items-center gap-3 px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-muted hover:border-warning/30 transition-colors text-left"
                  >
                    <Command className="w-4 h-4 shrink-0" />
                    <span className="text-sm">Search...</span>
                    <kbd className="ml-auto text-[10px] font-mono bg-bg-primary px-1.5 py-0.5 rounded border border-border-subtle">
                      ⌘K
                    </kbd>
                  </button>

                  {isAuthenticated ? (
                    <>
                      {/* Dashboard Section */}
                      <Link
                        to="/dashboard"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'flex items-center gap-2 py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/dashboard' && 'text-text-primary bg-bg-secondary'
                        )}
                      >
                        <Home className="w-4 h-4" /> Dashboard
                      </Link>

                      {/* Functions */}
                      <Link
                        to="/functions/my"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'flex items-center gap-2 py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname.startsWith('/functions') &&
                            'text-text-primary bg-bg-secondary'
                        )}
                      >
                        <FunctionSquare className="w-4 h-4" /> Functions
                      </Link>

                      {/* Marketplace Section */}
                      <div className="space-y-2">
                        <div className="text-sm font-semibold text-text-primary px-2">
                          Marketplace
                        </div>
                        <Link
                          to="/marketplace?type=functions"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/marketplace' && location.search.includes('type=functions') &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <FunctionSquare className="w-4 h-4" /> Browse Functions
                        </Link>
                        <Link
                          to="/marketplace?type=agents"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/marketplace' && location.search.includes('type=agents') &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <Bot className="w-4 h-4" /> Browse Agents
                        </Link>
                        <Link
                          to="/marketplace/purchases"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/marketplace/purchases' &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <ShoppingBag className="w-4 h-4" /> My Purchases
                        </Link>
                      </div>

                      {/* Providers Section */}
                      <div className="space-y-2">
                        <div className="text-sm font-semibold text-text-primary px-2">
                          Providers
                        </div>
                        <Link
                          to="/providers"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/providers' &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <Cloud className="w-4 h-4" /> Connected Providers
                        </Link>
                        <Link
                          to="/providers/billing"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/providers/billing' &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <CreditCard className="w-4 h-4" /> Usage & Billing
                        </Link>
                      </div>
                    </>
                  ) : (
                    <>
                      <a
                        href={marketingHomeUrl}
                        onClick={() => setIsMobileMenuOpen(false)}
                        className="block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50"
                      >
                        Home
                      </a>
                      <Link
                        to="/marketplace"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'flex items-center gap-2 py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/marketplace' &&
                            'text-text-primary bg-bg-secondary'
                        )}
                      >
                        <FunctionSquare className="w-4 h-4" /> Functions
                      </Link>
                      <Link
                        to="/pricing"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/pricing' && 'text-text-primary bg-bg-secondary'
                        )}
                      >
                        Pricing
                      </Link>
                      <a
                        href={DOCS_SITE_URL}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className="block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50"
                      >
                        Docs
                      </a>
                    </>
                  )}

                  {/* Theme Toggle */}
                  <div className="py-2 border-t border-border-default">
                    <ThemeToggle />
                  </div>

                  {!isAuthenticated && (
                    <div className="pt-4 border-t border-border-default space-y-2">
                      <Link to="/login" onClick={() => setIsMobileMenuOpen(false)}>
                        <FrameButton size="sm" className="w-full justify-start">
                          Sign In
                        </FrameButton>
                      </Link>
                      <Link to="/signup" onClick={() => setIsMobileMenuOpen(false)}>
                        <SealedButton size="sm" className="w-full">
                          Get Started
                        </SealedButton>
                      </Link>
                    </div>
                  )}
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Command Palette Overlay */}
        <AnimatePresence>
          {showCommandPalette && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-start justify-center pt-[20vh]"
              onClick={() => setShowCommandPalette(false)}
            >
              <motion.div
                ref={paletteRef}
                initial={{ opacity: 0, y: -20, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -20, scale: 0.95 }}
                transition={{ type: 'spring', stiffness: 300, damping: 30 }}
                className="w-full max-w-2xl mx-4 bg-bg-primary border border-border-default rounded-xl shadow-2xl overflow-hidden"
                onClick={(e) => e.stopPropagation()}
              >
                {/* Search Input */}
                <div className="flex items-center gap-3 px-4 py-4 bg-bg-secondary border-b border-border-default">
                  <Command className="w-5 h-5 text-text-muted shrink-0" />
                  <input
                    type="text"
                    placeholder="Search functions, agents, providers..."
                    className="flex-1 text-base text-text-primary placeholder:text-text-muted bg-bg-secondary focus:outline-none min-w-0"
                    autoFocus
                  />
                  <kbd className="hidden sm:block text-[10px] font-mono text-text-secondary bg-bg-primary px-2 py-1 rounded border border-border-subtle shrink-0">
                    ESC
                  </kbd>
                </div>

                {/* Quick Actions */}
                <div className="p-2">
                  <p className="px-3 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider">
                    Quick Actions
                  </p>
                  <div className="space-y-1">
                    {QUICK_ACTIONS.map((action) => (
                      <button
                        key={action.key}
                        onClick={() => action.action(navigate, setShowCommandPalette)}
                        className="w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <action.icon className="w-4 h-4" />
                          <span>{action.label}</span>
                        </div>
                        <kbd className="text-[10px] font-mono text-text-muted bg-bg-secondary px-1.5 py-0.5 rounded">
                          ⌘{action.key.toUpperCase()}
                        </kbd>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Footer */}
                <div className="px-4 py-3 bg-bg-secondary border-t border-border-default text-xs text-text-muted">
                  <p className="flex items-center gap-2">
                    <span>Use</span>
                    <kbd className="font-mono bg-bg-secondary px-1 py-0.5 rounded">↑</kbd>
                    <kbd className="font-mono bg-bg-secondary px-1 py-0.5 rounded">↓</kbd>
                    <span>to navigate,</span>
                    <kbd className="font-mono bg-bg-secondary px-1 py-0.5 rounded">↵</kbd>
                    <span>to select</span>
                  </p>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </>
    </TooltipProvider>
  );
}

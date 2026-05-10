import { Logo } from '@/components/common/Logo';
import { MarketplaceDropdown } from '@/components/common/MarketplaceDropdown';
import { ProductsDropdown } from '@/components/common/ProductsDropdown';
import { ThemeToggle } from '@/components/common/ThemeToggle';
import { UserMenu } from '@/components/layout/UserMenu';
import { NotificationBell } from '@/components/notifications';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { DOCS_SITE_URL, getMarketingRedirectOrigin } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useThemeStore } from '@/stores/themeStore';
import { AnimatePresence, motion } from 'framer-motion';
import { BarChart3, Bot, Cloud, Command, FunctionSquare, Home, Menu, MessageCircle, Sparkles, X, Zap } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
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
  action: (navigate: ReturnType<typeof useNavigate>, setShowCommandPalette: (show: boolean) => void) => void;
}

const QUICK_ACTIONS: QuickAction[] = [
  { 
    key: 'g', 
    label: 'Go to...', 
    icon: Command,
    action: (_, setShow) => setShow(true)
  },
  { 
    key: 'n', 
    label: 'New Function', 
    icon: Sparkles,
    action: (navigate, setShow) => {
      setShow(false);
      navigate('/functions/new');
    }
  },
  { 
    key: 'a', 
    label: 'Agents', 
    icon: Zap,
    action: (navigate, setShow) => {
      setShow(false);
      navigate('/marketplace/agents');
    }
  },
];

export function Navbar({ variant = 'landing', className, onMenuClick }: NavbarProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const theme = useThemeStore((state) => state.theme);
  const messagesUnread = useNotificationStore((state) => state.unreadCounts.messages);
  const marketingHomeUrl = getMarketingRedirectOrigin();

  // Scroll-aware background
  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', handler);
    return () => window.removeEventListener('scroll', handler);
  }, []);

  const settingsPath = user?.username ? `/u/${user.username}/settings` : '/settings';
  const navigationItems = isAuthenticated
    ? [
        { path: '/dashboard', label: 'Marketplace' },
        { path: '/overview', label: 'Overview' },
        { path: '/functions', label: 'Functions' },
        { path: '/providers', label: 'Providers' },
        { path: '/analytics', label: 'Analytics' },
        { path: settingsPath, label: 'Settings' },
      ]
    : [
        {
          path: marketingHomeUrl,
          label: 'Home',
          external: true,
        },
        { path: '/registry', label: 'Functions' },
        { path: '/pricing', label: 'Pricing' },
        { path: DOCS_SITE_URL, label: 'Docs', external: true },
      ];

  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Command palette: Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
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

  return (
    <TooltipProvider delayDuration={0}>
      <>
        <nav
          className={cn(
            'fixed top-0 left-0 right-0 z-50 transition-all duration-300',
            variant === 'dashboard'
              ? 'bg-bg-primary/95 backdrop-blur-xl border-b border-border-default'
              : cn(
                  'bg-glass backdrop-blur-md border-b border-subtle',
                  scrolled && 'bg-bg-primary/80 backdrop-blur-md border-b border-border-subtle'
                ),
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
                  <TooltipContent side="bottom" className="bg-bg-secondary border border-border-subtle shadow-lg">
                    <p className="text-text-primary">Open sidebar</p>
                  </TooltipContent>
                </Tooltip>
              )}

              {/* Logo */}
              {isAuthenticated ? (
                <Link
                  to="/dashboard"
                  className="shrink-0 mr-4 md:mr-6"
                  aria-label="FunctionFly home"
                >
                  <Logo />
                </Link>
              ) : (
                <a
                  href={marketingHomeUrl}
                  className="shrink-0 mr-4 md:mr-6"
                  aria-label="FunctionFly home"
                >
                  <Logo />
                </a>
              )}

              {/* Breadcrumbs - shown on nested pages */}
              {variant === 'dashboard' && location.pathname.split('/').filter(Boolean).length > 1 && (
                <nav className="hidden lg:flex items-center gap-2 text-sm text-text-muted ml-2">
                  <Link to="/" className="hover:text-text-primary transition-colors">Home</Link>
                  {location.pathname.split('/').filter(Boolean).slice(0, -1).map((segment, i) => {
                    const path = '/' + location.pathname.split('/').filter(Boolean).slice(0, i + 1).join('/');
                    return (
                      <span key={i} className="flex items-center gap-2">
                        <span className="text-text-muted">/</span>
                        <Link to={path} className="hover:text-text-primary transition-colors capitalize">
                          {segment}
                        </Link>
                      </span>
                    );
                  })}
                </nav>
              )}
            </div>

            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center gap-6">
              {isAuthenticated ? (
                <>
                  {variant !== 'dashboard' && (
                    <>
                      <ProductsDropdown />
                      <MarketplaceDropdown />
                    </>
                  )}
                  {variant === 'dashboard' && (
                    <>
                      <Link
                        to="/dashboard"
                        className={cn(
                          'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                          location.pathname === '/dashboard' && 'text-text-primary'
                        )}
                      >
                        Marketplace
                        {location.pathname === '/dashboard' && (
                          <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                        )}
                      </Link>
                      <Link
                        to="/overview"
                        className={cn(
                          'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                          location.pathname === '/overview' && 'text-text-primary'
                        )}
                      >
                        Overview
                        {location.pathname === '/overview' && (
                          <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                        )}
                      </Link>
                      <Link
                        to="/functions/my"
                        className={cn(
                          'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                          (location.pathname === '/functions' || location.pathname.startsWith('/functions/')) && 'text-text-primary'
                        )}
                      >
                        Functions
                        {(location.pathname === '/functions' || location.pathname.startsWith('/functions/')) && (
                          <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                        )}
                      </Link>
                      <Link
                        to="/providers"
                        className={cn(
                          'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                          location.pathname === '/providers' && 'text-text-primary'
                        )}
                      >
                        Providers
                        {location.pathname === '/providers' && (
                          <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                        )}
                      </Link>
                      <Link
                        to="/analytics"
                        className={cn(
                          'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                          location.pathname === '/analytics' && 'text-text-primary'
                        )}
                      >
                        Analytics
                        {location.pathname === '/analytics' && (
                          <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                        )}
                      </Link>
                    </>
                  )}
                  <Link
                    to={settingsPath}
                    className={cn(
                      'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                      (location.pathname === '/settings' ||
                        location.pathname.match(/^\/u\/[^/]+\/settings$/)) &&
                        'text-text-primary'
                    )}
                  >
                    Settings
                    {(location.pathname === '/settings' || location.pathname.match(/^\/u\/[^/]+\/settings$/)) && (
                      <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                    )}
                  </Link>
                </>
              ) : (
                <>
                  <Link
                    to="/"
                    className={cn(
                      'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                      location.pathname === '/' && 'text-text-primary'
                    )}
                  >
                    Home
                    {location.pathname === '/' && (
                      <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                    )}
                  </Link>
                  <ProductsDropdown />
                  <Link
                    to="/registry"
                    className={cn(
                      'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                      location.pathname === '/registry' && 'text-text-primary'
                    )}
                  >
                    Functions
                    {location.pathname === '/registry' && (
                      <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                    )}
                  </Link>
                  <MarketplaceDropdown />
                  <Link
                    to="/pricing"
                    className={cn(
                      'relative text-text-secondary hover:text-text-primary transition-colors font-medium py-1',
                      location.pathname === '/pricing' && 'text-text-primary'
                    )}
                  >
                    Pricing
                    {location.pathname === '/pricing' && (
                      <span className="absolute left-0 -bottom-0.5 w-full h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 rounded-full" />
                    )}
                  </Link>
                  <a
                    href={DOCS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-secondary hover:text-text-primary transition-colors font-medium"
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
                    <div className="hidden lg:flex items-center gap-1.5 px-2 py-1 rounded-full bg-bg-secondary/30 border border-border-subtle cursor-pointer hover:bg-bg-secondary/50 transition-colors">
                      <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                      <span className="text-xs text-text-secondary">3/3 Providers</span>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="bg-bg-secondary border border-border-subtle shadow-lg">
                    <p className="text-text-primary">All providers operational</p>
                    <p className="text-xs text-text-muted">Cloudflare, Vercel, Fly.io</p>
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
                  <TooltipContent side="bottom" className="bg-bg-secondary border border-border-subtle shadow-lg">
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
                      <TooltipContent side="bottom" className="bg-bg-secondary border border-border-subtle shadow-lg">
                        <p className="text-text-primary">Messages</p>
                        {messagesUnread > 0 && (
                          <p className="text-xs text-text-muted">
                            {messagesUnread} unread
                          </p>
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

                  {/* User Menu */}
                  <UserMenu />
                </>
              ) : (
                <>
                  {/* Theme Toggle */}
                  <ThemeToggle />

                  {/* Auth Buttons */}
                  <Link to="/login">
                    <Button
                      variant="ghost"
                      className="text-text-secondary hover:text-text-primary hidden sm:inline-flex"
                    >
                      Sign In
                    </Button>
                  </Link>
                  <Link to="/signup">
                    <Button className="button-primary">Get Started</Button>
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
                  <div className="relative">
                    <Command className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                    <input
                      type="text"
                      placeholder="Search..."
                      className="w-full pl-9 pr-4 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-muted focus:outline-none focus:border-warning"
                      onClick={() => {
                        setIsMobileMenuOpen(false);
                        setShowCommandPalette(true);
                      }}
                      readOnly
                    />
                  </div>

                  {isAuthenticated ? (
                    <>
                      {variant !== 'dashboard' && (
                        <>
                          {/* Products Section */}
                          <div className="space-y-2">
                            <div className="text-sm font-semibold text-text-primary px-2">
                              Products
                            </div>
<Link
                        to="/functions/my"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/functions' &&
                            'text-text-primary bg-bg-secondary'
                        )}
                      >
                              <FunctionSquare className="w-4 h-4" /> Functions
                            </Link>
                            <Link
                              to="/providers"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                                location.pathname === '/providers' &&
                                  'text-text-primary bg-bg-secondary'
                              )}
                            >
                              <Cloud className="w-4 h-4" /> Providers
                            </Link>
                            <Link
                              to="/analytics"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                                location.pathname === '/analytics' &&
                                  'text-text-primary bg-bg-secondary'
                              )}
                            >
                              <BarChart3 className="w-4 h-4" /> Analytics
                            </Link>
                          </div>

                          {/* Marketplace Section */}
                          <div className="space-y-2">
                            <div className="text-sm font-semibold text-text-primary px-2">
                              Marketplace
                            </div>
                            <Link
                              to="/dashboard"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                                (location.pathname === '/dashboard' ||
                                  location.pathname.startsWith('/marketplace')) &&
                                  'text-text-primary bg-bg-secondary'
                              )}
                            >
                              <Home className="w-4 h-4" /> Function Marketplace
                            </Link>
                            <Link
                              to="/marketplace/agents"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                                location.pathname === '/marketplace/agents' &&
                                  'text-text-primary bg-bg-secondary'
                              )}
                            >
                              <Bot className="w-4 h-4" /> Agent Marketplace
                            </Link>
                          </div>
                        </>
                      )}

                      <Link
                        to={settingsPath}
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          (location.pathname === '/settings' ||
                            location.pathname.match(/^\/u\/[^/]+\/settings$/)) &&
                            'text-text-primary bg-bg-secondary'
                        )}
                      >
                        Settings
                      </Link>
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
                        to="/registry"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'flex items-center gap-2 py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/registry' &&
                            'text-text-primary bg-bg-secondary'
                        )}
                      >
                        <FunctionSquare className="w-4 h-4" /> Browse Functions
                      </Link>
                      {/* Marketplace Section */}
                      <div className="space-y-2">
                        <div className="text-sm font-semibold text-text-primary px-2">
                          Marketplace
                        </div>
                        <Link
                          to="/dashboard"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            (location.pathname === '/dashboard' ||
                              location.pathname.startsWith('/marketplace')) &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <Home className="w-4 h-4" /> Function Marketplace
                        </Link>
                        <Link
                          to="/marketplace/agents"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-2 py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                            location.pathname === '/marketplace/agents' &&
                              'text-text-primary bg-bg-secondary'
                          )}
                        >
                          <Bot className="w-4 h-4" /> Agent Marketplace
                        </Link>
                      </div>

                      <Link
                        to="/pricing"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium rounded-lg hover:bg-bg-secondary/50',
                          location.pathname === '/pricing' &&
                            'text-text-primary bg-bg-secondary'
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
                        <Button variant="ghost" className="w-full justify-start">
                          Sign In
                        </Button>
                      </Link>
                      <Link to="/signup" onClick={() => setIsMobileMenuOpen(false)}>
                        <Button className="w-full button-primary">Get Started</Button>
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

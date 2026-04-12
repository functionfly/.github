import { Logo } from '@/components/common/Logo';
import { MarketplaceDropdown } from '@/components/common/MarketplaceDropdown';
import { ProductsDropdown } from '@/components/common/ProductsDropdown';
import { ThemeToggle } from '@/components/common/ThemeToggle';
import { SearchButton } from '@/components/layout/SearchButton';
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
import { Command, Menu, MessageCircle, Sparkles, X, Zap } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';

interface NavbarProps {
  variant?: 'landing' | 'dashboard';
  className?: string;
  onMenuClick?: () => void;
}

// Quick action shortcuts
const QUICK_ACTIONS = [
  { key: 'g', label: 'Go to...', icon: Command },
  { key: 'n', label: 'New Function', icon: Sparkles },
  { key: 'a', label: 'Agents', icon: Zap },
];

export function Navbar({ variant = 'landing', className, onMenuClick }: NavbarProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const location = useLocation();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const theme = useThemeStore((state) => state.theme);
  const messagesUnread = useNotificationStore((state) => state.unreadCounts.messages);
  const marketingHomeUrl = getMarketingRedirectOrigin();

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
            'fixed top-0 left-0 right-0 z-50',
            variant === 'dashboard'
              ? 'bg-aviation-bg-primary/95 backdrop-blur-xl border-b border-aviation-border-panel'
              : 'glass-navbar',
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
                      className="lg:hidden text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument mr-2"
                      onClick={onMenuClick}
                      aria-label="Open navigation menu"
                    >
                      <Menu className="w-5 h-5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Open sidebar</p>
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
                  <Link
                    to={settingsPath}
                    className={cn(
                      'text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium',
                      (location.pathname === '/settings' ||
                        location.pathname.match(/^\/u\/[^/]+\/settings$/)) &&
                        'text-aviation-text-primary'
                    )}
                  >
                    Settings
                  </Link>
                </>
              ) : (
                <>
                  <a
                    href={marketingHomeUrl}
                    className="text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium"
                  >
                    Home
                  </a>
                  <ProductsDropdown />
                  <Link
                    to="/registry"
                    className={cn(
                      'text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium',
                      location.pathname === '/registry' && 'text-aviation-text-primary'
                    )}
                  >
                    Functions
                  </Link>
                  <MarketplaceDropdown />
                  <Link
                    to="/pricing"
                    className={cn(
                      'text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium',
                      location.pathname === '/pricing' && 'text-aviation-text-primary'
                    )}
                  >
                    Pricing
                  </Link>
                  <a
                    href={DOCS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium"
                  >
                    Docs
                  </a>
                </>
              )}
            </div>

            {/* Right Section */}
            <div className="flex items-center gap-2 shrink-0">
              {/* Command Palette Trigger - Desktop */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setShowCommandPalette(true)}
                    className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-aviation-text-muted bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-all"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span className="hidden lg:inline">Search</span>
                    <kbd className="hidden xl:inline text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                      ⌘K
                    </kbd>
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>Command Palette</p>
                </TooltipContent>
              </Tooltip>

              {isAuthenticated ? (
                <>
                  {/* Search (Dashboard only) */}
                  {variant === 'dashboard' && (
                    <div className="hidden sm:block">
                      <SearchButton />
                    </div>
                  )}

                  {/* Theme Toggle */}
                  <ThemeToggle />

                  {/* Messages (Dashboard only) */}
                  {variant === 'dashboard' && (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          to="/conversations"
                          aria-label={
                            messagesUnread > 0 ? `Messages (${messagesUnread} unread)` : 'Messages'
                          }
                          className="relative flex items-center justify-center rounded-lg p-2 text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument transition-colors"
                        >
                          <MessageCircle className="w-5 h-5" />
                          {messagesUnread > 0 && (
                            <span className="absolute -top-0.5 -right-0.5 flex min-h-[18px] min-w-[18px] items-center justify-center rounded-full bg-aviation-red px-1 text-[10px] font-bold leading-none text-white">
                              {messagesUnread > 99 ? '99+' : messagesUnread}
                            </span>
                          )}
                        </Link>
                      </TooltipTrigger>
                      <TooltipContent side="bottom">
                        <p>Messages</p>
                        {messagesUnread > 0 && (
                          <p className="text-xs text-aviation-text-muted">
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
                      className="relative text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
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
                      className="text-aviation-text-secondary hover:text-aviation-text-primary hidden sm:inline-flex"
                    >
                      Sign In
                    </Button>
                  </Link>
                  <Link to="/signup">
                    <Button className="aviation-button-primary">Get Started</Button>
                  </Link>
                </>
              )}

              {/* Mobile Menu Toggle */}
              <Button
                variant="ghost"
                size="icon"
                className="md:hidden text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
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
                className="fixed top-16 left-0 right-0 bg-aviation-bg-primary border-b border-aviation-border-panel shadow-xl"
              >
                <div className="px-4 py-4 space-y-4 max-h-[70vh] overflow-y-auto">
                  {/* Search for mobile */}
                  <div className="relative">
                    <Command className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
                    <input
                      type="text"
                      placeholder="Search..."
                      className="w-full pl-9 pr-4 py-2 bg-aviation-bg-instrument border border-aviation-border-instrument rounded-lg text-aviation-text-primary placeholder:text-aviation-text-dim focus:outline-none focus:border-aviation-amber"
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
                            <div className="text-sm font-semibold text-aviation-text-primary px-2">
                              Products
                            </div>
                            <Link
                              to="/functions"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                                location.pathname === '/functions' &&
                                  'text-aviation-text-primary bg-aviation-bg-instrument'
                              )}
                            >
                              <span className="mr-2">⚡</span> Functions
                            </Link>
                            <Link
                              to="/providers"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                                location.pathname === '/providers' &&
                                  'text-aviation-text-primary bg-aviation-bg-instrument'
                              )}
                            >
                              <span className="mr-2">☁️</span> Providers
                            </Link>
                            <Link
                              to="/analytics"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                                location.pathname === '/analytics' &&
                                  'text-aviation-text-primary bg-aviation-bg-instrument'
                              )}
                            >
                              <span className="mr-2">📊</span> Analytics
                            </Link>
                          </div>

                          {/* Marketplace Section */}
                          <div className="space-y-2">
                            <div className="text-sm font-semibold text-aviation-text-primary px-2">
                              Marketplace
                            </div>
                            <Link
                              to="/dashboard"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                                (location.pathname === '/dashboard' ||
                                  location.pathname.startsWith('/marketplace')) &&
                                  'text-aviation-text-primary bg-aviation-bg-instrument'
                              )}
                            >
                              <span className="mr-2">⚡</span> Function Marketplace
                            </Link>
                            <Link
                              to="/marketplace/agents"
                              onClick={() => setIsMobileMenuOpen(false)}
                              className={cn(
                                'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                                location.pathname === '/marketplace/agents' &&
                                  'text-aviation-text-primary bg-aviation-bg-instrument'
                              )}
                            >
                              <span className="mr-2">🤖</span> Agent Marketplace
                            </Link>
                          </div>
                        </>
                      )}

                      <Link
                        to={settingsPath}
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                          (location.pathname === '/settings' ||
                            location.pathname.match(/^\/u\/[^/]+\/settings$/)) &&
                            'text-aviation-text-primary bg-aviation-bg-instrument'
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
                        className="block py-2 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50"
                      >
                        Home
                      </a>
                      <Link
                        to="/registry"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                          location.pathname === '/registry' &&
                            'text-aviation-text-primary bg-aviation-bg-instrument'
                        )}
                      >
                        <span className="mr-2">⚡</span> Browse Functions
                      </Link>
                      {/* Marketplace Section */}
                      <div className="space-y-2">
                        <div className="text-sm font-semibold text-aviation-text-primary px-2">
                          Marketplace
                        </div>
                        <Link
                          to="/dashboard"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                            (location.pathname === '/dashboard' ||
                              location.pathname.startsWith('/marketplace')) &&
                              'text-aviation-text-primary bg-aviation-bg-instrument'
                          )}
                        >
                          <span className="mr-2">⚡</span> Function Marketplace
                        </Link>
                        <Link
                          to="/marketplace/agents"
                          onClick={() => setIsMobileMenuOpen(false)}
                          className={cn(
                            'block py-2 pl-4 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                            location.pathname === '/marketplace/agents' &&
                              'text-aviation-text-primary bg-aviation-bg-instrument'
                          )}
                        >
                          <span className="mr-2">🤖</span> Agent Marketplace
                        </Link>
                      </div>

                      <Link
                        to="/pricing"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className={cn(
                          'block py-2 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50',
                          location.pathname === '/pricing' &&
                            'text-aviation-text-primary bg-aviation-bg-instrument'
                        )}
                      >
                        Pricing
                      </Link>
                      <a
                        href={DOCS_SITE_URL}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={() => setIsMobileMenuOpen(false)}
                        className="block py-2 text-aviation-text-secondary hover:text-aviation-text-primary transition-colors font-medium rounded-lg hover:bg-aviation-bg-instrument/50"
                      >
                        Docs
                      </a>
                    </>
                  )}

                  {/* Theme Toggle */}
                  <div className="py-2 border-t border-aviation-border-panel">
                    <ThemeToggle />
                  </div>

                  {!isAuthenticated && (
                    <div className="pt-4 border-t border-aviation-border-panel space-y-2">
                      <Link to="/login" onClick={() => setIsMobileMenuOpen(false)}>
                        <Button variant="ghost" className="w-full justify-start">
                          Sign In
                        </Button>
                      </Link>
                      <Link to="/signup" onClick={() => setIsMobileMenuOpen(false)}>
                        <Button className="w-full aviation-button-primary">Get Started</Button>
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
                className="w-full max-w-2xl mx-4 bg-aviation-bg-primary border border-aviation-border-panel rounded-xl shadow-2xl overflow-hidden"
                onClick={(e) => e.stopPropagation()}
              >
                {/* Search Input */}
                <div className="flex items-center gap-3 px-4 py-4 border-b border-aviation-border-panel">
                  <Command className="w-5 h-5 text-aviation-text-muted" />
                  <input
                    type="text"
                    placeholder="Search functions, agents, providers..."
                    className="flex-1 text-base text-aviation-text-primary placeholder:text-aviation-text-dim bg-transparent focus:outline-none"
                    autoFocus
                  />
                  <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-2 py-1 rounded">
                    ESC
                  </kbd>
                </div>

                {/* Quick Actions */}
                <div className="p-2">
                  <p className="px-3 py-2 text-xs font-semibold text-aviation-text-muted uppercase tracking-wider">
                    Quick Actions
                  </p>
                  <div className="space-y-1">
                    {QUICK_ACTIONS.map((action) => (
                      <button
                        key={action.key}
                        className="w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <action.icon className="w-4 h-4" />
                          <span>{action.label}</span>
                        </div>
                        <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                          ⌘{action.key.toUpperCase()}
                        </kbd>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Footer */}
                <div className="px-4 py-3 bg-aviation-bg-secondary border-t border-aviation-border-panel text-xs text-aviation-text-muted">
                  <p className="flex items-center gap-2">
                    <span>Use</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↑</kbd>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↓</kbd>
                    <span>to navigate,</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↵</kbd>
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

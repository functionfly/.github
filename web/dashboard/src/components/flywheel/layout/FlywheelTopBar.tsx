/**
 * FlywheelTopBar - Production Ready Header
 * Enhanced with aviation theme, command palette, and better UX
 */

import { useMyReputation } from '@/api/flywheel';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Bell,
  Command,
  HelpCircle,
  LogOut,
  Menu,
  Search,
  Settings,
  Sparkles,
  Trophy,
  User,
  X,
} from 'lucide-react';
import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ReputationBadge } from '../reputation/ReputationBadge';

interface FlywheelTopBarProps {
  className?: string;
  onMenuClick?: () => void;
  isMobileMenuOpen?: boolean;
}

// Quick action shortcuts
const QUICK_ACTIONS = [
  { key: 's', label: 'Search Threads', icon: Search },
  { key: 'n', label: 'New Thread', icon: Sparkles },
  { key: 'r', label: 'Reputation', icon: Trophy },
];

export function FlywheelTopBar({ className, onMenuClick, isMobileMenuOpen }: FlywheelTopBarProps) {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [notifications, setNotifications] = useState([
    {
      id: 1,
      title: 'New solution accepted',
      message: 'Your solution was marked as accepted',
      time: '2 min ago',
      unread: true,
    },
    {
      id: 2,
      title: 'Reputation increased',
      message: '+50 points for helpful contribution',
      time: '1 hour ago',
      unread: true,
    },
  ]);
  const { data: reputationData } = useMyReputation();

  const unreadCount = notifications.filter((n) => n.unread).length;

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/flywheel/threads?search=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Command palette: Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setShowCommandPalette(true);
      }

      // Quick navigation
      if (e.metaKey || e.ctrlKey) {
        const action = QUICK_ACTIONS.find((a) => a.key === e.key.toLowerCase());
        if (action) {
          e.preventDefault();
          // Navigate based on action
          if (action.key === 's') navigate('/flywheel/search');
          if (action.key === 'n') navigate('/flywheel/threads/new');
          if (action.key === 'r') navigate('/flywheel/reputation/me');
        }
      }
    },
    [navigate]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  const markAllRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, unread: false })));
  };

  return (
    <TooltipProvider delayDuration={0}>
      <>
        <header
          className={cn(
            'fixed left-0 right-0 top-0 z-50 h-16',
            'bg-aviation-bg-primary/95 backdrop-blur-xl',
            'border-b border-aviation-border-panel',
            className
          )}
        >
          <div className="flex h-full items-center justify-between px-4 lg:px-6">
            {/* Left Section */}
            <div className="flex items-center gap-4">
              {/* Mobile Menu Button */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="lg:hidden text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                    onClick={onMenuClick}
                    aria-label={isMobileMenuOpen ? 'Close menu' : 'Open menu'}
                  >
                    {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>{isMobileMenuOpen ? 'Close menu' : 'Open menu'}</p>
                </TooltipContent>
              </Tooltip>

              {/* Logo */}
              <a href="/flywheel" className="flex items-center gap-2 group">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-linear-to-br from-aviation-amber to-aviation-amber-glow group-hover:shadow-lg group-hover:shadow-aviation-amber/20 transition-shadow">
                  <Trophy className="h-4 w-4 text-aviation-bg-primary" />
                </div>
                <span className="hidden text-lg font-bold text-aviation-text-primary sm:inline">
                  Flywheel
                </span>
              </a>
            </div>

            {/* Center - Search */}
            <form onSubmit={handleSearch} className="mx-4 hidden max-w-md flex-1 md:block">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-aviation-text-muted" />
                <Input
                  type="search"
                  placeholder="Search threads, users, solutions..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="h-10 border-aviation-border-instrument bg-aviation-bg-instrument pl-10 text-sm text-aviation-text-primary placeholder:text-aviation-text-dim focus-visible:border-aviation-amber focus-visible:ring-aviation-amber/20"
                />
                <kbd className="absolute right-3 top-1/2 -translate-y-1/2 hidden lg:inline-flex text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-secondary px-1.5 py-0.5 rounded">
                  ⌘K
                </kbd>
              </div>
            </form>

            {/* Right Section */}
            <div className="flex items-center gap-2">
              {/* Command Palette Trigger - Desktop */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setShowCommandPalette(true)}
                    className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-aviation-text-muted bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-all"
                  >
                    <Command className="w-3.5 h-3.5" />
                    <span className="hidden lg:inline">Quick Actions</span>
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>Command Palette (⌘K)</p>
                </TooltipContent>
              </Tooltip>

              {/* Mobile Search Button */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="md:hidden text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                    onClick={() => navigate('/flywheel/search')}
                    aria-label="Search"
                  >
                    <Search className="w-5 h-5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>Search</p>
                </TooltipContent>
              </Tooltip>

              {/* Notifications */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="relative text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                    aria-label="Notifications"
                  >
                    <Bell className="w-5 h-5" />
                    {unreadCount > 0 && (
                      <>
                        <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-aviation-amber ring-2 ring-aviation-bg-primary" />
                        <span className="absolute -top-1 -right-1 flex items-center justify-center min-w-[18px] h-[18px] bg-aviation-red text-white text-[10px] font-bold rounded-full px-1">
                          {unreadCount > 9 ? '9+' : unreadCount}
                        </span>
                      </>
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="w-80 bg-aviation-bg-primary border-aviation-border-panel"
                >
                  <div className="flex items-center justify-between px-3 py-2 border-b border-aviation-border-panel">
                    <span className="text-sm font-medium text-aviation-text-primary">
                      Notifications
                    </span>
                    {unreadCount > 0 && (
                      <button
                        onClick={markAllRead}
                        className="text-xs text-aviation-amber hover:text-aviation-amber-glow"
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
                            'px-3 py-3 border-b border-aviation-border-panel/50 last:border-0 hover:bg-aviation-bg-instrument/50 transition-colors cursor-pointer',
                            notification.unread && 'bg-aviation-amber/5'
                          )}
                        >
                          <div className="flex items-start gap-3">
                            <div
                              className={cn(
                                'w-2 h-2 rounded-full mt-1.5 shrink-0',
                                notification.unread ? 'bg-aviation-amber' : 'bg-aviation-text-dim'
                              )}
                            />
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium text-aviation-text-primary">
                                {notification.title}
                              </p>
                              <p className="text-xs text-aviation-text-secondary mt-0.5">
                                {notification.message}
                              </p>
                              <p className="text-[10px] text-aviation-text-muted mt-1">
                                {notification.time}
                              </p>
                            </div>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="px-3 py-8 text-center text-aviation-text-muted">
                        <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                        <p className="text-sm">No new notifications</p>
                      </div>
                    )}
                  </div>
                </DropdownMenuContent>
              </DropdownMenu>

              {/* User Menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="flex items-center gap-2 px-2 hover:bg-aviation-bg-instrument"
                  >
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-linear-to-br from-aviation-bg-instrument to-aviation-bg-secondary ring-2 ring-aviation-border-instrument">
                      <User className="h-4 w-4 text-aviation-text-secondary" />
                    </div>
                    <span className="hidden text-sm font-medium text-aviation-text-secondary lg:inline">
                      Profile
                    </span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="w-56 bg-aviation-bg-primary border-aviation-border-panel"
                >
                  <div className="px-3 py-2">
                    <p className="text-sm font-medium text-aviation-text-primary">Your Account</p>
                    {reputationData?.profile && (
                      <div className="mt-2 flex items-center gap-2">
                        <ReputationBadge
                          score={reputationData.profile.overallScore}
                          type="overall"
                          tier={reputationData.profile.tier?.level}
                          size="sm"
                        />
                      </div>
                    )}
                  </div>
                  <DropdownMenuSeparator className="bg-aviation-border-panel" />
                  <DropdownMenuItem
                    onClick={() => navigate('/flywheel/reputation/me')}
                    className="text-aviation-text-secondary focus:bg-aviation-bg-instrument focus:text-aviation-text-primary"
                  >
                    <Trophy className="mr-2 h-4 w-4" />
                    Reputation
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate('/settings')}
                    className="text-aviation-text-secondary focus:bg-aviation-bg-instrument focus:text-aviation-text-primary"
                  >
                    <Settings className="mr-2 h-4 w-4" />
                    Settings
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate('/help')}
                    className="text-aviation-text-secondary focus:bg-aviation-bg-instrument focus:text-aviation-text-primary"
                  >
                    <HelpCircle className="mr-2 h-4 w-4" />
                    Help
                  </DropdownMenuItem>
                  <DropdownMenuSeparator className="bg-aviation-border-panel" />
                  <DropdownMenuItem className="text-aviation-red focus:bg-aviation-red/10 focus:text-aviation-red">
                    <LogOut className="mr-2 h-4 w-4" />
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </header>

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
                    placeholder="Search threads, users, solutions..."
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
                        onClick={() => {
                          setShowCommandPalette(false);
                          if (action.key === 's') navigate('/flywheel/search');
                          if (action.key === 'n') navigate('/flywheel/threads/new');
                          if (action.key === 'r') navigate('/flywheel/reputation/me');
                        }}
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

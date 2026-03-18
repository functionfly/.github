/**
 * FlywheelTopBar - Header with search and user actions
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Search,
  Bell,
  Settings,
  LogOut,
  User,
  Trophy,
  HelpCircle,
  Menu,
  X,
} from 'lucide-react';
import { useMyReputation } from '@/api/flywheel';
import { ReputationBadge } from '../reputation/ReputationBadge';

interface FlywheelTopBarProps {
  className?: string;
  onMenuClick?: () => void;
  isMobileMenuOpen?: boolean;
}

export function FlywheelTopBar({
  className,
  onMenuClick,
  isMobileMenuOpen,
}: FlywheelTopBarProps) {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const { data: reputationData } = useMyReputation();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/flywheel/threads?search=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  return (
    <header
      className={cn(
        'flywheel-topbar fixed left-0 right-0 top-0 z-50 h-16 border-b border-border-default bg-bg-secondary/80 backdrop-blur-md',
        className
      )}
    >
      <div className="flex h-full items-center justify-between px-4 lg:px-6">
        {/* Left Section */}
        <div className="flex items-center gap-4">
          {/* Mobile Menu Button */}
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={onMenuClick}
            aria-label={isMobileMenuOpen ? 'Close menu' : 'Open menu'}
          >
            {isMobileMenuOpen ? (
              <X className="flywheel-nav-icon h-5 w-5 text-text-muted" />
            ) : (
              <Menu className="flywheel-nav-icon h-5 w-5 text-text-muted" />
            )}
          </Button>

          {/* Logo */}
          <a href="/flywheel" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
              <Trophy className="h-4 w-4 text-white" />
            </div>
            <span className="flywheel-logo-text hidden text-lg font-bold text-text-primary sm:inline">
              Flywheel
            </span>
          </a>
        </div>

        {/* Center - Search */}
        <form onSubmit={handleSearch} className="mx-4 hidden max-w-md flex-1 md:block">
          <div className="relative">
            <Search className="flywheel-search-icon absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
            <Input
              type="search"
              placeholder="Search threads, users, solutions..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="flywheel-search-input h-10 border-border-default bg-bg-tertiary pl-10 text-sm text-text-primary placeholder:text-text-muted focus-visible:ring-indigo-500"
            />
          </div>
        </form>

        {/* Right Section */}
        <div className="flex items-center gap-2">
          {/* Mobile Search Button */}
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={() => navigate('/flywheel/search')}
            aria-label="Search"
          >
            <Search className="flywheel-nav-icon h-5 w-5 text-text-muted" />
          </Button>

          {/* Notifications */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="relative"
                aria-label="Notifications"
              >
                <Bell className="flywheel-nav-icon h-5 w-5 text-text-muted" />
                <span className="flywheel-notification-dot absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-indigo-500 ring-2 ring-bg-primary" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80 border-border-default bg-bg-elevated">
              <div className="flex items-center justify-between px-3 py-2">
                <span className="text-sm font-medium text-text-primary">Notifications</span>
                <Button variant="ghost" size="sm" className="h-auto text-xs text-indigo-400 hover:text-indigo-300">
                  Mark all read
                </Button>
              </div>
              <DropdownMenuSeparator className="bg-border-subtle" />
              <div className="max-h-64 overflow-y-auto">
                <div className="px-3 py-4 text-center text-sm text-text-muted">
                  No new notifications
                </div>
              </div>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* User Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="flywheel-profile-text flex items-center gap-2 px-2 hover:bg-bg-hover"
              >
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-slate-700 to-slate-600">
                  <User className="h-4 w-4 text-text-secondary" />
                </div>
                <span className="hidden text-sm font-medium text-text-secondary lg:inline">
                  Profile
                </span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56 border-border-default bg-bg-elevated">
              <div className="px-3 py-2">
                <p className="text-sm font-medium text-text-primary">Your Account</p>
                {reputationData?.profile && (
                  <div className="mt-2 flex items-center gap-2">
                    <ReputationBadge
                      score={reputationData.profile.overallScore}
                      type="overall"
                      tier={reputationData.profile.tier.level}
                      size="sm"
                    />
                  </div>
                )}
              </div>
              <DropdownMenuSeparator className="bg-border-subtle" />
              <DropdownMenuItem
                onClick={() => navigate('/flywheel/reputation/me')}
                className="text-text-secondary focus:bg-bg-hover focus:text-text-primary"
              >
                <Trophy className="mr-2 h-4 w-4" />
                Reputation
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => navigate('/settings')}
                className="text-text-secondary focus:bg-bg-hover focus:text-text-primary"
              >
                <Settings className="mr-2 h-4 w-4" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => navigate('/help')}
                className="text-text-secondary focus:bg-bg-hover focus:text-text-primary"
              >
                <HelpCircle className="mr-2 h-4 w-4" />
                Help
              </DropdownMenuItem>
              <DropdownMenuSeparator className="bg-border-subtle" />
              <DropdownMenuItem
                className="text-red-400 focus:bg-bg-hover focus:text-red-300"
              >
                <LogOut className="mr-2 h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
}

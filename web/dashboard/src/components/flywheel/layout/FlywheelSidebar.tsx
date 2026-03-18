/**
 * FlywheelSidebar - Navigation sidebar for Flywheel Network
 */

import { NavLink, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import {
  MessageSquare,
  Trophy,
  BarChart3,
  Flame,
  Plus,
  Hash,
  Users,
  Sparkles,
} from 'lucide-react';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
}

const mainNavItems: NavItem[] = [
  { label: 'Hub', path: '/flywheel', icon: Flame },
  { label: 'Threads', path: '/flywheel/threads', icon: MessageSquare },
  { label: 'Challenges', path: '/flywheel/challenges', icon: Trophy },
  { label: 'Leaderboards', path: '/flywheel/leaderboards', icon: BarChart3 },
];

const discoveryNavItems: NavItem[] = [
  { label: 'Categories', path: '/flywheel/categories', icon: Hash },
  { label: 'Users', path: '/flywheel/users', icon: Users },
  { label: 'Solutions', path: '/flywheel/solutions', icon: Sparkles },
];

interface FlywheelSidebarProps {
  className?: string;
}

export function FlywheelSidebar({ className }: FlywheelSidebarProps) {
  const location = useLocation();

  return (
    <aside
      className={cn(
        'flywheel-sidebar fixed left-0 top-0 z-40 h-full w-64 border-r border-border-default bg-bg-secondary pt-16',
        className
      )}
    >
      <div className="flex h-full flex-col">
        {/* New Thread Button */}
        <div className="px-4 py-3">
            <NavLink
            to="/flywheel/threads/new"
            className={cn(
              'flywheel-btn-new-thread flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white transition-all hover:bg-indigo-500',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-2 focus-visible:ring-offset-bg-primary'
            )}
          >
            <Plus className="h-4 w-4" />
            New Thread
          </NavLink>
        </div>

        {/* Main Navigation */}
        <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-4">
          {/* Primary Nav */}
          <div className="space-y-1">
            <p className="flywheel-nav-section-label px-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
              Main
            </p>
            {mainNavItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) => {
                  const active = isActive || (item.path !== '/flywheel' && location.pathname.startsWith(item.path));
                  return cn(
                    'flywheel-nav-link flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                    active ? 'flywheel-nav-link-active bg-bg-hover text-indigo-400' : 'text-text-secondary hover:bg-bg-hover hover:text-text-primary'
                  );
                }}
              >
                <item.icon className="h-4 w-4" />
                {item.label}
                {item.badge && (
                  <span className="ml-auto rounded-full bg-indigo-500/10 px-2 py-0.5 text-xs font-medium text-indigo-400">
                    {item.badge}
                  </span>
                )}
              </NavLink>
            ))}
          </div>

          {/* Discovery Nav */}
          <div className="space-y-1">
            <p className="flywheel-nav-section-label px-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
              Discovery
            </p>
            {discoveryNavItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  cn(
                    'flywheel-nav-link flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                    isActive ? 'flywheel-nav-link-active bg-bg-hover text-indigo-400' : 'text-text-secondary hover:bg-bg-hover hover:text-text-primary'
                  )
                }
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            ))}
          </div>

          {/* Popular Tags */}
          <div className="space-y-2">
            <p className="flywheel-nav-section-label px-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
              Popular Tags
            </p>
            <div className="flex flex-wrap gap-1.5 px-3">
              {['algorithms', 'optimization', 'python', 'javascript', 'rust'].map((tag) => (
                <NavLink
                  key={tag}
                  to={`/flywheel/threads?tags=${tag}`}
                  className="flywheel-tag rounded-md bg-bg-tertiary px-2 py-1 text-xs text-text-secondary transition-colors hover:bg-bg-hover hover:text-text-primary"
                >
                  #{tag}
                </NavLink>
              ))}
            </div>
          </div>
        </nav>

        {/* Footer */}
        <div className="flywheel-sidebar-footer border-t border-border-default p-4">
          <p className="text-xs text-text-muted">
            Flywheel Network™
            <br />
            Proof-of-Execution Knowledge Network
          </p>
        </div>
      </div>
    </aside>
  );
}

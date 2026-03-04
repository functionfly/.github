/**
 * FlywheelMobileNav - Mobile navigation drawer
 */

import { NavLink } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import {
  MessageSquare,
  Trophy,
  BarChart3,
  Flame,
  Plus,
  Hash,
  Users,
  Sparkles,
  X,
} from 'lucide-react';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
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

interface FlywheelMobileNavProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function FlywheelMobileNav({ isOpen, onOpenChange }: FlywheelMobileNavProps) {
  return (
    <Sheet open={isOpen} onOpenChange={onOpenChange}>
      <SheetContent
        side="left"
        className="w-72 border-slate-800 bg-slate-950 p-0"
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-slate-800 p-4">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
                <Trophy className="h-4 w-4 text-white" />
              </div>
              <span className="text-lg font-bold text-white">Flywheel</span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onOpenChange(false)}
              className="text-slate-400"
            >
              <X className="h-5 w-5" />
            </Button>
          </div>

          {/* New Thread Button */}
          <div className="p-4">
            <NavLink
              to="/flywheel/threads/new"
              onClick={() => onOpenChange(false)}
              className={cn(
                'flex items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-3 text-sm font-medium text-white transition-all hover:bg-indigo-500'
              )}
            >
              <Plus className="h-4 w-4" />
              New Thread
            </NavLink>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-2">
            {/* Main Nav */}
            <div className="space-y-1">
              <p className="px-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
                Main
              </p>
              {mainNavItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  onClick={() => onOpenChange(false)}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-slate-900 text-indigo-400'
                        : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
                    )
                  }
                >
                  <item.icon className="h-5 w-5" />
                  {item.label}
                </NavLink>
              ))}
            </div>

            {/* Discovery Nav */}
            <div className="space-y-1">
              <p className="px-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
                Discovery
              </p>
              {discoveryNavItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  onClick={() => onOpenChange(false)}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-slate-900 text-indigo-400'
                        : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
                    )
                  }
                >
                  <item.icon className="h-5 w-5" />
                  {item.label}
                </NavLink>
              ))}
            </div>

            {/* Popular Tags */}
            <div className="space-y-2">
              <p className="px-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
                Popular Tags
              </p>
              <div className="flex flex-wrap gap-2 px-3">
                {['algorithms', 'optimization', 'python', 'javascript', 'rust'].map((tag) => (
                  <NavLink
                    key={tag}
                    to={`/flywheel/threads?tags=${tag}`}
                    onClick={() => onOpenChange(false)}
                    className="rounded-md bg-slate-900 px-3 py-1.5 text-sm text-slate-400 transition-colors hover:bg-slate-800 hover:text-slate-200"
                  >
                    #{tag}
                  </NavLink>
                ))}
              </div>
            </div>
          </nav>

          {/* Footer */}
          <div className="border-t border-slate-800 p-4">
            <p className="text-xs text-slate-500">
              Flywheel Network™
              <br />
              Proof-of-Execution Knowledge Network
            </p>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

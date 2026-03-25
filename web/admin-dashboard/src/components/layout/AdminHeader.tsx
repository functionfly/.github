/**
 * Admin Header Component
 */

import { DarkModeToggle } from '@/components/common/DarkModeToggle';
import { ROUTES } from '@/lib/constants';
import type { AdminUser } from '@/types';
import { Bell, LogOut, User } from 'lucide-react';
import { Link } from 'react-router-dom';

interface AdminHeaderProps {
  user: AdminUser | null;
  onMenuClick: () => void;
  onLogout: () => void;
  /** Show sidebar toggle (only when logged in and sidebar is present) */
  showMenuButton?: boolean;
}

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
  return (
    <header className="bg-white border-b border-gray-200 shadow-sm shrink-0">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          {showMenuButton && (
            <button
              onClick={onMenuClick}
              className="md:hidden p-2 hover:bg-gray-100 rounded-lg transition-colors"
              aria-label="Toggle menu"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
          )}

          <div>
            <h1 className="text-xl sm:text-2xl font-bold text-gray-900 tracking-tight">
              {user ? 'Admin Dashboard' : 'FlyAdmin'}
            </h1>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {user ? (
            <>
              {/* Notifications */}
              <button
                className="p-2.5 rounded-xl border border-gray-200 bg-gray-50/80 hover:bg-gray-100 hover:border-gray-300 transition-colors relative"
                aria-label="Notifications"
              >
                <Bell className="w-5 h-5 text-gray-600" />
                <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full ring-2 ring-white" />
              </button>

              {/* Dark Mode Toggle */}
              <DarkModeToggle />

              {/* User block: pill-style container */}
              <div className="flex items-center gap-3 pl-3 pr-2 py-1.5 rounded-xl border border-gray-200 bg-linear-to-r from-gray-50 to-slate-50/80 shadow-sm min-w-0">
                <Link
                  to={`${ROUTES.ADMIN_USERS}/${user.id}`}
                  className="flex items-center gap-3 min-w-0 rounded-lg -m-1 p-1 hover:bg-gray-200/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-50 transition-colors"
                  title="View profile"
                >
                  <div className="w-10 h-10 rounded-full bg-linear-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-semibold text-sm shadow-sm shrink-0 ring-2 ring-white ring-offset-2 ring-offset-gray-50">
                    {user.name?.trim()
                      ? user.name.charAt(0).toUpperCase()
                      : (user.email?.charAt(0) || 'A').toUpperCase()}
                  </div>
                  <div className="hidden sm:block text-left min-w-0">
                    <span className="text-sm font-semibold text-blue-700 hover:text-blue-800 hover:underline truncate block">
                      {displayName(user)}
                    </span>
                    <span className="text-xs text-gray-500 font-medium block">
                      {displayRole(user.role)}
                    </span>
                  </div>
                </Link>
                <button
                  onClick={onLogout}
                  className="p-2 rounded-lg hover:bg-gray-200/80 text-gray-600 hover:text-gray-900 transition-colors shrink-0"
                  title="Logout"
                >
                  <LogOut className="w-5 h-5" />
                </button>
              </div>
            </>
          ) : (
            <div className="flex items-center gap-2 text-gray-500">
              <User className="w-5 h-5" />
              <p className="text-sm font-medium">Sign in to continue</p>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

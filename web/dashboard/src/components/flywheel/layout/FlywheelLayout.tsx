/**
 * FlywheelLayout - Main layout with sidebar for Flywheel Network
 */

import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { FlywheelSidebar } from './FlywheelSidebar';
import { FlywheelTopBar } from './FlywheelTopBar';
import { FlywheelMobileNav } from './FlywheelMobileNav';
import { FlywheelChatAssistant } from '../ai/FlywheelChatAssistant';

interface FlywheelLayoutProps {
  className?: string;
}

export function FlywheelLayout({ className }: FlywheelLayoutProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  return (
    <div className={cn('min-h-screen bg-slate-950', className)}>
      {/* Top Bar */}
      <FlywheelTopBar
        onMenuClick={() => setIsMobileMenuOpen(true)}
        isMobileMenuOpen={isMobileMenuOpen}
      />

      {/* Desktop Sidebar */}
      <div className="hidden lg:block">
        <FlywheelSidebar />
      </div>

      {/* Mobile Navigation */}
      <div className="lg:hidden">
        <FlywheelMobileNav
          isOpen={isMobileMenuOpen}
          onOpenChange={setIsMobileMenuOpen}
        />
      </div>

      {/* Main Content */}
      <main className="pt-16 lg:pl-64">
        <div className="min-h-[calc(100vh-4rem)]">
          <Outlet />
        </div>
      </main>

      {/* AI Chat Assistant */}
      <FlywheelChatAssistant />
    </div>
  );
}

/**
 * FlywheelPageLayout - Standard page layout with consistent padding and max-width
 */
interface FlywheelPageLayoutProps {
  children: React.ReactNode;
  className?: string;
  fullWidth?: boolean;
}

export function FlywheelPageLayout({
  children,
  className,
  fullWidth = false,
}: FlywheelPageLayoutProps) {
  return (
    <div
      className={cn(
        'mx-auto px-4 py-6 sm:px-6 lg:px-8',
        !fullWidth && 'max-w-7xl',
        className
      )}
    >
      {children}
    </div>
  );
}

/**
 * FlywheelSection - Section container with consistent spacing
 */
interface FlywheelSectionProps {
  children: React.ReactNode;
  className?: string;
  title?: string;
  description?: string;
  action?: React.ReactNode;
}

export function FlywheelSection({
  children,
  className,
  title,
  description,
  action,
}: FlywheelSectionProps) {
  return (
    <section className={cn('space-y-4', className)}>
      {(title || action) && (
        <div className="flex items-center justify-between gap-4">
          <div>
            {title && (
              <h2 className="text-xl font-semibold text-white">{title}</h2>
            )}
            {description && (
              <p className="mt-1 text-sm text-slate-400">{description}</p>
            )}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      {children}
    </section>
  );
}

/**
 * FlywheelCard - Consistent card styling for Flywheel components
 */
interface FlywheelCardProps {
  children: React.ReactNode;
  className?: string;
  hoverable?: boolean;
  onClick?: () => void;
}

export function FlywheelCard({
  children,
  className,
  hoverable = false,
  onClick,
}: FlywheelCardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-xl border border-slate-800 bg-slate-900 p-5',
        hoverable && 'cursor-pointer transition-all duration-200 hover:-translate-y-0.5 hover:border-slate-700 hover:shadow-lg',
        onClick && 'cursor-pointer',
        className
      )}
    >
      {children}
    </div>
  );
}

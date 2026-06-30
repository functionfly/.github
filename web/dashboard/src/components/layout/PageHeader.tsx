import { Button } from '@/components/ui/button';
import { generateBreadcrumbs } from '@/lib/breadcrumbs';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { AlertCircle, ChevronRight, Shield, Sparkles } from 'lucide-react';
import { useLocation } from 'react-router-dom';
import React, { useMemo } from 'react';

interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ComponentType<{ className?: string }>;
  isActive?: boolean;
}

interface ActionButton {
  label: string;
  onClick: () => void;
  variant?: 'default' | 'outline' | 'ghost' | 'secondary';
  size?: 'sm' | 'md' | 'lg';
  icon?: React.ComponentType<{ className?: string }>;
  disabled?: boolean;
  className?: string;
}

// Enhanced badge types
interface PageBadge {
  label: string;
  variant?:
    | 'default'
    | 'secondary'
    | 'destructive'
    | 'outline'
    | 'new'
    | 'beta'
    | 'enterprise'
    | 'warning';
  icon?: React.ComponentType<{ className?: string }>;
  className?: string;
}

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  breadcrumbs?: BreadcrumbItem[];
  actions?: ActionButton[];
  badges?: PageBadge[];
  className?: string;
  animate?: boolean;
}

export function PageHeader({
  title,
  subtitle,
  breadcrumbs,
  actions,
  badges,
  className,
  animate = true,
}: PageHeaderProps) {
  const location = useLocation();

  // Auto-generate breadcrumbs if not provided
  const autoBreadcrumbs = useMemo(() => {
    const crumbs = generateBreadcrumbs(location.pathname);
    return crumbs.map((c) => ({
      label: c.label,
      path: c.path,
      icon: c.icon,
      isActive: c.isActive,
    }));
  }, [location.pathname]);

  const displayBreadcrumbs = breadcrumbs || autoBreadcrumbs;

  const Header = animate ? motion.div : 'div';
  const headerProps = animate
    ? {
        initial: { opacity: 0, y: -10 },
        animate: { opacity: 1, y: 0 },
        transition: { duration: 0.3, ease: 'easeOut' as const },
      }
    : {};

  // Get badge styles based on variant
  const getBadgeStyles = (variant?: string) => {
    switch (variant) {
      case 'new':
        return 'bg-aviation-green/20 text-aviation-green border-aviation-green/30';
      case 'beta':
        return 'bg-aviation-cyan/20 text-aviation-cyan border-aviation-cyan/30';
      case 'enterprise':
        return 'bg-aviation-amber/20 text-aviation-amber border-aviation-amber/30';
      case 'warning':
        return 'bg-aviation-amber/20 text-aviation-amber border-aviation-amber/30';
      case 'destructive':
        return 'bg-aviation-red/20 text-aviation-red border-aviation-red/30';
      case 'secondary':
        return 'bg-aviation-bg-instrument text-aviation-text-secondary border-aviation-border-instrument';
      case 'outline':
        return 'bg-transparent text-aviation-text-muted border-aviation-border-instrument';
      default:
        return 'bg-aviation-amber/20 text-aviation-amber border-aviation-amber/30';
    }
  };

  // Get badge icon based on variant
  const getBadgeIcon = (variant?: string) => {
    switch (variant) {
      case 'new':
        return Sparkles;
      case 'beta':
        return Shield;
      case 'enterprise':
        return Shield;
      case 'warning':
        return AlertCircle;
      default:
        return null;
    }
  };

  return (
    <Header {...headerProps} className={cn('space-y-4', className)}>
      {/* Breadcrumbs — uses sc-breadcrumb CSS classes from sc-navbar.css */}
      {displayBreadcrumbs && displayBreadcrumbs.length > 1 && (
        <nav className="sc-breadcrumb__nav">
          {displayBreadcrumbs.map((crumb, index) => {
            const Icon = crumb.icon;

            return (
              <div key={crumb.path ?? crumb.label} className="sc-breadcrumb__item">
                {index > 0 && <ChevronRight className="sc-breadcrumb__sep" />}
                {crumb.path && !crumb.isActive ? (
                  <Link to={crumb.path} className="sc-breadcrumb__link">
                    {Icon && <Icon className="sc-breadcrumb__icon" />}
                    <span>{crumb.label}</span>
                  </Link>
                ) : (
                  <span className={cn('sc-breadcrumb__text', crumb.isActive && 'sc-breadcrumb__text--active')}>
                    {Icon && <Icon className="sc-breadcrumb__icon" />}
                    <span>{crumb.label}</span>
                  </span>
                )}
              </div>
            );
          })}
        </nav>
      )}

      {/* Header Content */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-2xl font-bold text-aviation-text-primary">{title}</h1>
            {badges && badges.length > 0 && (
              <div className="flex gap-2">
                {badges.map((badge, index) => {
                  const Icon = badge.icon || getBadgeIcon(badge.variant);
                  return (
                    <span
                      key={index}
                      className={cn(
                        'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border',
                        getBadgeStyles(badge.variant)
                      )}
                    >
                      {Icon && <Icon className="w-3 h-3" />}
                      {badge.label}
                    </span>
                  );
                })}
              </div>
            )}
          </div>
          {subtitle && <p className="text-aviation-text-secondary">{subtitle}</p>}
        </div>

        {/* Actions */}
        {actions && actions.length > 0 && (
          <div className="flex items-center gap-2">
            {actions.map((action, index) => {
              const Icon = action.icon;
              return (
                <Button
                  key={index}
                  variant={action.variant || 'default'}
                  size={action.size === 'md' ? 'default' : action.size || 'sm'}
                  onClick={action.onClick}
                  disabled={action.disabled}
                  className={cn(action.variant === 'default' && 'aviation-button-primary')}
                >
                  {Icon && <Icon className="w-4 h-4 mr-2" />}
                  {action.label}
                </Button>
              );
            })}
          </div>
        )}
      </div>
    </Header>
  );
}

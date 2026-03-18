import { Button } from '@/components/ui/button';
import { ROUTES } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { ChevronRight, Plus } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

interface BreadcrumbItem {
  label: string;
  path?: string;
  isActive?: boolean;
}

interface ContextualAction {
  label: string;
  onClick: () => void;
}

export function Breadcrumb() {
  const location = useLocation();

  const getContextualActions = (): ContextualAction[] => {
    if (location.pathname.startsWith('/functions')) {
      return [
        {
          label: 'New Function',
          onClick: () => (window.location.href = '/functions/new'),
        },
      ];
    }
    if (location.pathname.startsWith('/providers')) {
      return [
        {
          label: 'Connect Provider',
          onClick: () => console.log('Open connect provider modal'),
        },
      ];
    }
    return [];
  };

  const generateBreadcrumbs = (): BreadcrumbItem[] => {
    const pathSegments = location.pathname.split('/').filter(Boolean);
    const breadcrumbs: BreadcrumbItem[] = [];

    // Always start with Dashboard
    breadcrumbs.push({
      label: 'Dashboard',
      path: ROUTES.DASHBOARD,
      isActive: location.pathname === ROUTES.DASHBOARD,
    });

    // Handle different routes
    if (location.pathname.startsWith('/functions')) {
      breadcrumbs.push({
        label: 'Functions',
        path: ROUTES.FUNCTIONS,
        isActive: location.pathname === ROUTES.FUNCTIONS,
      });

      // If we're on a specific function page
      if (pathSegments.length > 1 && pathSegments[0] === 'functions') {
        const functionId = pathSegments[1];
        if (functionId && functionId !== 'new') {
          breadcrumbs.push({
            label: `Function ${functionId}`,
            isActive: true,
          });
        } else if (functionId === 'new') {
          breadcrumbs.push({
            label: 'New Function',
            isActive: true,
          });
        }
      }
    } else if (location.pathname.startsWith('/providers')) {
      breadcrumbs.push({
        label: 'Providers',
        path: ROUTES.PROVIDERS,
        isActive: location.pathname === ROUTES.PROVIDERS,
      });

      // If we're on a specific provider page
      if (pathSegments.length > 1 && pathSegments[0] === 'providers') {
        const providerId = pathSegments[1];
        if (providerId) {
          breadcrumbs.push({
            label: `Provider ${providerId}`,
            isActive: true,
          });
        }
      }
    } else if (location.pathname.startsWith('/analytics')) {
      breadcrumbs.push({
        label: 'Analytics',
        path: ROUTES.ANALYTICS,
        isActive: true,
      });
    } else if (location.pathname.startsWith('/settings')) {
      breadcrumbs.push({
        label: 'Settings',
        path: ROUTES.SETTINGS,
        isActive: true,
      });
    }

    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();
  const contextualActions = getContextualActions();

  if (breadcrumbs.length <= 1) {
    return null; // Don't show breadcrumbs if we're just on Dashboard
  }

  return (
    <div className="hidden md:flex items-center gap-4">
      <nav className="flex items-center gap-2 text-sm">
        {breadcrumbs.map((crumb, index) => (
          <div key={index} className="flex items-center gap-2">
            {index > 0 && <ChevronRight className="w-4 h-4 text-text-muted" />}
            {crumb.path && !crumb.isActive ? (
              <Link
                to={crumb.path}
                className="text-text-secondary hover:text-white transition-colors"
              >
                {crumb.label}
              </Link>
            ) : (
              <span
                className={cn(crumb.isActive ? 'text-white font-medium' : 'text-text-secondary')}
              >
                {crumb.label}
              </span>
            )}
          </div>
        ))}
      </nav>

      {/* Contextual Actions */}
      {contextualActions.length > 0 && (
        <div className="flex items-center gap-2">
          {contextualActions.map((action, index) => (
            <Button
              key={index}
              variant="outline"
              size="sm"
              onClick={action.onClick}
              className="h-7 px-2 text-xs border-white/20 hover:bg-white/5"
            >
              <Plus className="w-3 h-3 mr-1" />
              {action.label}
            </Button>
          ))}
        </div>
      )}
    </div>
  );
}

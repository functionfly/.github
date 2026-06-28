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
      return [{ label: 'New Function', onClick: () => (window.location.href = '/functions/new') }];
    }
    if (location.pathname.startsWith('/providers')) {
      return [{ label: 'Connect Provider', onClick: () => console.log('Open connect provider modal') }];
    }
    return [];
  };

  const generateBreadcrumbs = (): BreadcrumbItem[] => {
    const pathSegments = location.pathname.split('/').filter(Boolean);
    const breadcrumbs: BreadcrumbItem[] = [];

    breadcrumbs.push({ label: 'Home', path: ROUTES.DASHBOARD, isActive: location.pathname === ROUTES.DASHBOARD });

    if (location.pathname === ROUTES.OVERVIEW || location.pathname.startsWith(`${ROUTES.OVERVIEW}/`)) {
      breadcrumbs[0].isActive = false;
      breadcrumbs.push({ label: 'Overview', path: ROUTES.OVERVIEW, isActive: true });
    } else if (location.pathname.startsWith('/functions')) {
      breadcrumbs.push({ label: 'Functions', path: ROUTES.FUNCTIONS, isActive: location.pathname === ROUTES.FUNCTIONS });
      if (pathSegments.length > 1 && pathSegments[0] === 'functions') {
        const functionId = pathSegments[1];
        if (functionId && functionId !== 'new') breadcrumbs.push({ label: `Function ${functionId}`, isActive: true });
        else if (functionId === 'new') breadcrumbs.push({ label: 'New Function', isActive: true });
      }
    } else if (location.pathname.startsWith('/providers')) {
      breadcrumbs.push({ label: 'Providers', path: ROUTES.PROVIDERS, isActive: location.pathname === ROUTES.PROVIDERS });
      if (pathSegments.length > 1 && pathSegments[0] === 'providers') {
        const providerId = pathSegments[1];
        if (providerId) breadcrumbs.push({ label: `Provider ${providerId}`, isActive: true });
      }
    } else if (location.pathname.startsWith('/analytics')) {
      breadcrumbs.push({ label: 'Analytics', path: ROUTES.ANALYTICS, isActive: true });
    } else if (location.pathname.startsWith('/settings')) {
      breadcrumbs.push({ label: 'Settings', path: ROUTES.SETTINGS, isActive: true });
    }

    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();
  const contextualActions = getContextualActions();

  if (breadcrumbs.length <= 1) return null;

  return (
    <div className="sc-breadcrumb">
      <nav className="sc-breadcrumb__nav">
        {breadcrumbs.map((crumb, index) => (
          <div key={index} className="sc-breadcrumb__item">
            {index > 0 && <ChevronRight className="sc-breadcrumb__sep" />}
            {crumb.path && !crumb.isActive ? (
              <Link to={crumb.path} className="sc-breadcrumb__link">{crumb.label}</Link>
            ) : (
              <span className={cn('sc-breadcrumb__text', crumb.isActive && 'sc-breadcrumb__text--active')}>
                {crumb.label}
              </span>
            )}
          </div>
        ))}
      </nav>

      {contextualActions.length > 0 && (
        <div className="sc-breadcrumb__actions">
          {contextualActions.map((action, index) => (
            <button key={index} className="sc-breadcrumb__action" onClick={action.onClick}>
              <Plus className="sc-breadcrumb__action-icon" />
              {action.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

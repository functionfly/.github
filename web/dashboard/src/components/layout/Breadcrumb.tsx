/**
 * Breadcrumb component — used in TopBar.
 *
 * Uses the centralized breadcrumbs registry so it stays in sync
 * with PageHeader's auto-generation.
 *
 * Production-readiness:
 * - Keyed by path (not index) for stable identity
 * - Contextual actions derived from route prefix
 * - Responsive: hidden below 768px (same threshold as TopBar)
 */

import { ROUTES } from '@/lib/constants';
import { generateBreadcrumbs } from '@/lib/breadcrumbs';
import { cn } from '@/lib/utils';
import { ChevronRight, Plus } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

interface ContextualAction {
  label: string;
  onClick: () => void;
}

export function Breadcrumb() {
  const location = useLocation();

  const getContextualActions = (): ContextualAction[] => {
    const path = location.pathname;
    if (path.startsWith('/functions')) {
      return [{ label: 'New Function', onClick: () => { window.location.href = '/functions/new'; } }];
    }
    if (path.startsWith('/apps')) {
      return [{ label: 'New App', onClick: () => { window.location.href = '/apps/new'; } }];
    }
    if (path.startsWith('/agents')) {
      return [{ label: 'New Agent', onClick: () => { window.location.href = '/agents/new'; } }];
    }
    if (path.startsWith('/providers')) {
      return [{ label: 'Connect Provider', onClick: () => console.log('Open connect provider modal') }];
    }
    return [];
  };

  const crumbs = generateBreadcrumbs(location.pathname);
  const contextualActions = getContextualActions();

  // Hide when at home/overview only
  if (crumbs.length <= 1) return null;

  return (
    <div className="sc-breadcrumb">
      <nav className="sc-breadcrumb__nav">
        {crumbs.map((crumb, index) => (
          <div key={crumb.path ?? crumb.label} className="sc-breadcrumb__item">
            {index > 0 && <ChevronRight className="sc-breadcrumb__sep" />}
            {crumb.path && !crumb.isActive ? (
              <Link to={crumb.path} className="sc-breadcrumb__link">
                {crumb.label}
              </Link>
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
          {contextualActions.map((action) => (
            <button key={action.label} className="sc-breadcrumb__action" onClick={action.onClick}>
              <Plus className="sc-breadcrumb__action-icon" />
              {action.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

import { Building2, Globe, Layers, Plus, Zap } from 'lucide-react';
import {
  Chamber,
  SealedButton,
} from '@/components/containment';
import { CreateAppModal } from './CreateAppModal';

interface AppsEmptyStateProps {
  isFiltered?: boolean;
  searchQuery?: string;
  onCreateSuccess?: () => void;
}

export function AppsEmptyState({ isFiltered, searchQuery, onCreateSuccess }: AppsEmptyStateProps) {
  if (isFiltered) {
    return (
      <Chamber className="apps-empty">
        <Building2 className="apps-empty__icon" />
        <h3 className="apps-empty__title">No apps found</h3>
        <p className="apps-empty__desc">
          No apps match{' '}
          {searchQuery ? <>your search for <strong>"{searchQuery}"</strong></> : 'your current filters'}.
          Try adjusting your search.
        </p>
      </Chamber>
    );
  }

  return (
    <Chamber className="apps-empty">
      <div className="apps-empty__illustration">
        <div className="apps-empty__glow" />
        <div className="apps-empty__main-icon">
          <Building2 className="apps-empty__main-icon-svg" />
        </div>
        <div className="apps-empty__float apps-empty__float--violet">
          <Layers className="apps-empty__float-icon" />
        </div>
        <div className="apps-empty__float apps-empty__float--emerald">
          <Zap className="apps-empty__float-icon" />
        </div>
        <div className="apps-empty__float apps-empty__float--blue">
          <Globe className="apps-empty__float-icon apps-empty__float-icon--sm" />
        </div>
      </div>

      <h3 className="apps-empty__title">Create your first app</h3>
      <p className="apps-empty__desc">
        Apps are the foundation of FunctionFly. Organize your functions, manage multi-cloud
        deployments, and monitor performance — all in one place.
      </p>

      <div className="apps-empty__features">
        {[
          { icon: Layers, label: 'Organize functions' },
          { icon: Globe, label: 'Multi-cloud deploy' },
          { icon: Zap, label: 'Edge performance' },
        ].map(({ icon: Icon, label }) => (
          <div key={label} className="apps-empty__feature">
            <Icon className="apps-empty__feature-icon" />
            <span className="apps-empty__feature-label">{label}</span>
          </div>
        ))}
      </div>

      <CreateAppModal
        onSuccess={onCreateSuccess}
        trigger={
          <SealedButton iconLeft={<Plus className="apps-icon-sm" />}>
            Create Your First App
          </SealedButton>
        }
      />
    </Chamber>
  );
}

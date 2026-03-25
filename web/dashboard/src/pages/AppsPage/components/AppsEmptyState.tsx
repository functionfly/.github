import { Button } from '@/components/ui/button';
import { Building2, Globe, Layers, Plus, Zap } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface AppsEmptyStateProps {
  isFiltered?: boolean;
  searchQuery?: string;
}

export function AppsEmptyState({ isFiltered, searchQuery }: AppsEmptyStateProps) {
  const navigate = useNavigate();

  if (isFiltered) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="w-16 h-16 rounded-2xl bg-muted/50 border border-border/50 flex items-center justify-center mb-5">
          <Building2 className="w-7 h-7 text-muted-foreground/60" />
        </div>
        <h3 className="text-lg font-semibold text-foreground mb-2">No apps found</h3>
        <p className="text-sm text-muted-foreground max-w-sm">
          No apps match{' '}
          {searchQuery ? (
            <>
              your search for <span className="font-medium text-foreground">"{searchQuery}"</span>
            </>
          ) : (
            'your current filters'
          )}
          . Try adjusting your search.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      {/* Illustration */}
      <div className="relative mb-8">
        {/* Background glow */}
        <div className="absolute inset-0 bg-brand-500/10 rounded-full blur-3xl scale-150" />

        {/* Main icon */}
        <div className="relative w-20 h-20 rounded-2xl bg-gradient-to-br from-brand-500/20 to-brand-600/10 border border-brand-500/20 flex items-center justify-center">
          <Building2 className="w-9 h-9 text-brand-500" />
        </div>

        {/* Floating feature icons */}
        <div className="absolute -top-2 -right-4 w-9 h-9 rounded-xl bg-violet-500/15 border border-violet-500/20 flex items-center justify-center">
          <Layers className="w-4 h-4 text-violet-500" />
        </div>
        <div className="absolute -bottom-2 -left-4 w-9 h-9 rounded-xl bg-emerald-500/15 border border-emerald-500/20 flex items-center justify-center">
          <Zap className="w-4 h-4 text-emerald-500" />
        </div>
        <div className="absolute top-1/2 -translate-y-1/2 -right-8 w-8 h-8 rounded-xl bg-blue-500/15 border border-blue-500/20 flex items-center justify-center">
          <Globe className="w-3.5 h-3.5 text-blue-500" />
        </div>
      </div>

      <h3 className="text-xl font-semibold text-foreground mb-3">Create your first app</h3>
      <p className="text-sm text-muted-foreground mb-8 max-w-md leading-relaxed">
        Apps are the foundation of FunctionFly. Organize your functions, manage multi-cloud
        deployments, and monitor performance — all in one place.
      </p>

      {/* Feature highlights */}
      <div className="grid grid-cols-3 gap-4 mb-8 max-w-sm w-full">
        {[
          { icon: Layers, label: 'Organize functions', color: 'text-violet-500' },
          { icon: Globe, label: 'Multi-cloud deploy', color: 'text-blue-500' },
          { icon: Zap, label: 'Edge performance', color: 'text-emerald-500' },
        ].map(({ icon: Icon, label, color }) => (
          <div
            key={label}
            className="flex flex-col items-center gap-2 p-3 rounded-xl bg-muted/30 border border-border/50"
          >
            <Icon className={`w-5 h-5 ${color}`} />
            <span className="text-xs text-muted-foreground text-center leading-tight">{label}</span>
          </div>
        ))}
      </div>

      <Button onClick={() => navigate('/apps/new')} size="lg" className="gap-2">
        <Plus className="w-4 h-4" />
        Create Your First App
      </Button>
    </div>
  );
}

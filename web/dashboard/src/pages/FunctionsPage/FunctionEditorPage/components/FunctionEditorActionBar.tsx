import { Button } from '@/components/ui/button';
import type { FunctionEditorModel } from '../useFunctionEditor';
import { ArrowLeft, CheckCircle2, ChevronRight, Loader2, Play, Rocket, Save } from 'lucide-react';
import { Link } from 'react-router-dom';

type Props = { editor: FunctionEditorModel };

export function FunctionEditorActionBar({ editor }: Props) {
  const {
    navigate,
    isEditing,
    functionName,
    isLoading,
    isDirty,
    lastSaved,
    handleTest,
    handleSaveDraft,
    handleDeploy,
    isSaving,
    isDeploying,
    isTesting,
  } = editor;

  return (
    <div
      className="sticky top-0 z-40 border-b"
      style={{
        background: 'rgba(10,10,15,0.9)',
        backdropFilter: 'blur(12px)',
        borderColor: 'rgba(255,255,255,0.08)',
      }}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2 min-w-0">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate('/functions')}
            className="shrink-0 text-text-secondary hover:text-text-primary"
            aria-label="Back to functions"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <nav className="flex items-center gap-1 text-sm min-w-0" aria-label="Breadcrumb">
            <Link
              to="/functions"
              className="text-text-muted hover:text-text-primary transition-colors truncate"
            >
              Functions
            </Link>
            <ChevronRight className="w-3.5 h-3.5 text-text-muted shrink-0" />
            <span className="text-text-primary font-medium truncate">
              {isEditing ? functionName || 'Edit Function' : functionName || 'New Function'}
            </span>
          </nav>
        </div>

        <div className="hidden sm:flex items-center gap-2 text-xs text-text-muted">
          {isLoading ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-400" />
          ) : isDirty ? (
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
              Unsaved changes
            </span>
          ) : lastSaved ? (
            <span className="flex items-center gap-1.5">
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
              Draft saved
            </span>
          ) : null}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/functions')}
            className="text-text-secondary hover:text-text-primary hidden sm:flex"
          >
            Cancel
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleTest}
            disabled={isLoading}
            className="gap-1.5 border-border-default text-text-secondary hover:text-text-primary hidden md:flex"
          >
            {isTesting ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Play className="w-3.5 h-3.5" />
            )}
            Test
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleSaveDraft}
            disabled={isLoading}
            className="gap-1.5 border-border-default text-text-secondary hover:text-text-primary"
          >
            {isSaving ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Save className="w-3.5 h-3.5" />
            )}
            <span className="hidden sm:inline">Save Draft</span>
          </Button>
          <Button
            size="sm"
            onClick={handleDeploy}
            disabled={isLoading}
            className="gap-1.5 font-semibold"
            style={{
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              color: '#fff',
              border: 'none',
            }}
          >
            {isDeploying ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Rocket className="w-3.5 h-3.5" />
            )}
            {isDeploying ? 'Deploying…' : 'Deploy'}
          </Button>
        </div>
      </div>
    </div>
  );
}

import { Button } from '@/components/ui/button';
import {
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  Clock,
  Loader2,
  Play,
  Rocket,
  Save,
  X,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { FunctionEditorModel } from './useFunctionEditor';

type Props = { editor: FunctionEditorModel };

function formatRelativeTime(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 5) return 'just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  return `${Math.floor(minutes / 60)}h ago`;
}

export function ActionBar({ editor }: Props) {
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
    showDraftRestorePrompt,
    handleRestoreDraft,
    handleDiscardDraft,
  } = editor;

  const [relativeTime, setRelativeTime] = useState('');

  useEffect(() => {
    if (!lastSaved) return;
    setRelativeTime(formatRelativeTime(lastSaved));
    const interval = setInterval(() => {
      setRelativeTime(formatRelativeTime(lastSaved));
    }, 10000);
    return () => clearInterval(interval);
  }, [lastSaved]);

  return (
    <>
      {/* Draft restore prompt */}
      {showDraftRestorePrompt && (
        <div
          className="sticky top-0 z-50 border-b"
          style={{
            background: 'rgba(99,102,241,0.12)',
            backdropFilter: 'blur(12px)',
            borderColor: 'rgba(99,102,241,0.3)',
          }}
        >
          <div className="max-w-7xl mx-auto px-4 sm:px-6 h-11 flex items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-indigo-400 shrink-0" />
              <span className="text-text-secondary">
                You have an unsaved draft. Would you like to restore it?
              </span>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Button
                size="sm"
                variant="ghost"
                onClick={handleDiscardDraft}
                className="text-text-muted hover:text-text-primary h-7 gap-1"
              >
                <X className="w-3.5 h-3.5" />
                Discard
              </Button>
              <Button
                size="sm"
                onClick={handleRestoreDraft}
                className="h-7 gap-1"
                style={{
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  color: '#fff',
                  border: 'none',
                }}
              >
                Restore Draft
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Main action bar — theme-aware (light: surface + semantic text; dark: frosted chrome) */}
      <div
        className="sticky top-0 z-40 border-b border-border-subtle bg-bg-secondary/95 backdrop-blur-xl dark:border-white/10 dark:bg-[rgba(10,10,15,0.9)]"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
          {/* Left: breadcrumb */}
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

          {/* Center: save status */}
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
                Draft saved {relativeTime}
              </span>
            ) : null}
          </div>

          {/* Right: actions */}
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
    </>
  );
}

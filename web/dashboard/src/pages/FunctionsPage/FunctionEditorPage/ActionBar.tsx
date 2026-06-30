import { Button } from '@/components/ui/button';
import {
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  Clock,
  Keyboard,
  Loader2,
  Play,
  Rocket,
  Save,
  X,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { KeyboardShortcutsDialog } from './components/KeyboardShortcutsDialog';
import type { FunctionEditorModel } from './useFunctionEditor';

type Props = { editor: FunctionEditorModel };

function formatRelativeTime(date: Date, t: (key: string, opts?: Record<string, unknown>) => string): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 5) return t('funcEditor.justNow');
  if (seconds < 60) return t('funcEditor.secondsAgo', { count: seconds });
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return t('funcEditor.minutesAgo', { count: minutes });
  return t('funcEditor.hoursAgo', { count: Math.floor(minutes / 60) });
}

export function ActionBar({ editor }: Props) {
  const { t } = useTranslation();
  const {
    navigate,
    isEditing,
    functionName,
    isLoading,
    isDirty,
    lastSaved,
    draftTimestamp,
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
    setRelativeTime(formatRelativeTime(lastSaved, t));
    const interval = setInterval(() => {
      setRelativeTime(formatRelativeTime(lastSaved, t));
    }, 10000);
    return () => clearInterval(interval);
  }, [lastSaved]);

  return (
    <>
      {/* Draft restore prompt */}
      {showDraftRestorePrompt && (
        <div
          className="sticky top-0 z-30 border-b"
          style={{
            background: 'rgba(255, 122, 61, 0.08)',
            backdropFilter: 'blur(12px)',
            borderColor: 'var(--accent-dim)',
          }}
        >
          <div className="max-w-7xl mx-auto px-4 sm:px-6 h-11 flex items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-[var(--accent)] shrink-0" />
              <span className="text-[var(--text-dim)]">
                {t('funcEditor.unsavedDraft')}{draftTimestamp && (
                  <span className="text-[var(--text-faint)] ml-1">
                    {t('funcEditor.unsavedDraftFrom', { time: formatRelativeTime(draftTimestamp, t) })}
                  </span>
                )}. {t('funcEditor.restoreItQuestion')}
              </span>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Button
                size="sm"
                variant="ghost"
                onClick={handleDiscardDraft}
                className="h-7 gap-1"
              >
                <X className="w-3.5 h-3.5" />
                {t('funcEditor.discard')}
              </Button>
              <Button
                size="sm"
                onClick={handleRestoreDraft}
                className="h-7 gap-1"
              >
                {t('funcEditor.restoreDraft')}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Main action bar */}
      <div
        className="sticky top-0 z-20 border-b"
        style={{
          background: 'rgba(14, 19, 24, 0.85)',
          backdropFilter: 'blur(12px)',
          borderColor: 'var(--panel-edge)',
        }}
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
          {/* Left: breadcrumb */}
          <div className="flex items-center gap-2 min-w-0">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => navigate('/functions')}
                className="shrink-0"
                aria-label={t('funcEditor.backToFunctions')}
              >
                <ArrowLeft className="w-4 h-4" />
              </Button>
            <nav className="flex items-center gap-1 text-sm min-w-0" aria-label="Breadcrumb">
              <Link
                to="/functions/my"
                className="text-[var(--text-faint)] hover:text-[var(--text)] transition-colors truncate"
              >
                {t('funcEditor.breadcrumbFunctions')}
              </Link>
              <ChevronRight className="w-3.5 h-3.5 text-[var(--text-faint)] shrink-0" />
              <span className="text-[var(--accent)] font-medium truncate" style={{ fontFamily: 'var(--font-display)' }}>
                {isEditing ? functionName || t('funcEditor.breadcrumbEditFunction') : functionName || t('funcEditor.breadcrumbNewFunction')}
              </span>
            </nav>
          </div>

          {/* Center: save status */}
          <div className="hidden sm:flex items-center gap-2 text-xs text-[var(--text-faint)]">
            {isLoading ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin text-[var(--accent)]" />
            ) : isDirty ? (
              <span className="flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-[var(--status-pending)] animate-pulse" />
                {t('funcEditor.unsavedChanges')}
              </span>
            ) : lastSaved ? (
              <span className="flex items-center gap-1.5">
                <CheckCircle2 className="w-3.5 h-3.5 text-[var(--status-ok)]" />
                {t('funcEditor.draftSavedTime', { time: relativeTime })}
              </span>
            ) : null}
          </div>

          {/* Right: actions */}
          <div className="flex items-center gap-2 shrink-0">
            <KeyboardShortcutsDialog>
              <button
                className="hidden md:flex items-center gap-1.5 text-xs text-[var(--text-faint)] hover:text-[var(--text)] transition-colors px-2 py-1 rounded-[var(--radius-sm)] hover:bg-[rgba(255,255,255,0.04)]"
                aria-label={t('funcEditor.keyboardShortcuts')}
              >
                <Keyboard className="w-3.5 h-3.5" />
                <kbd className="hidden lg:inline-flex items-center px-1.5 py-0.5 rounded-[var(--radius-sm)] bg-[var(--panel-raised)] border border-[var(--panel-edge)] text-[10px] font-[var(--font-mono)]">
                  ?
                </kbd>
              </button>
            </KeyboardShortcutsDialog>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/functions')}
              className="hidden sm:flex"
            >
              {t('funcEditor.cancelAction')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleTest}
              disabled={isLoading}
              className="gap-1.5 hidden md:flex"
            >
              {isTesting ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Play className="w-3.5 h-3.5" />
              )}
              {t('funcEditor.test')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleSaveDraft}
              disabled={isLoading}
              className="gap-1.5"
            >
              {isSaving ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Save className="w-3.5 h-3.5" />
              )}
              <span className="hidden sm:inline">{t('funcEditor.saveDraft')}</span>
            </Button>
            <Button
              size="sm"
              onClick={handleDeploy}
              disabled={isLoading}
              className="gap-1.5"
            >
              {isDeploying ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Rocket className="w-3.5 h-3.5" />
              )}
              {isDeploying ? t('funcEditor.deploying') : t('funcEditor.deployAction')}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
}

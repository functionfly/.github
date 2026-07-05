import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { LazyMonacoEditor } from '@/components/LazyMonacoEditor';
import { Code2, Maximize2, Minimize2 } from 'lucide-react';
import { useTheme } from '@/components/common/ThemeProvider';
import { RUNTIME_META } from '../constants';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function CodeEditorSectionMobile({ editor }: Props) {
  const { t } = useTranslation();
  const { runtime, code, markDirty } = editor;
  const { theme } = useTheme();
  const monacoTheme = theme === 'light' ? 'vs' : 'vs-dark';
  const [open, setOpen] = useState(false);

  if (open) {
    return (
      <div className="lg:hidden">
        <CodeEditorSectionMobileInner editor={editor} onBack={() => setOpen(false)} monacoTheme={monacoTheme} />
      </div>
    );
  }

  return (
    <div className="lg:hidden rounded-lg border border-[var(--panel-edge)] bg-[var(--panel)] p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Code2 className="w-4 h-4 text-[var(--status-ok)]" />
          <span className="text-sm font-semibold text-[var(--text)]">{t('funcEditor.codeEditor')}</span>
          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-[var(--panel-edge)] text-[var(--text-faint)]">
            {RUNTIME_META[runtime].label}
          </span>
        </div>
        <Button size="sm" variant="outline" className="gap-1.5" onClick={() => setOpen(true)}>
          {t('funcEditor.editCode') || 'Edit Code'}
        </Button>
      </div>
      <p className="text-xs text-[var(--text-faint)] mb-2">
        {code.split('\n').length}L · {code.length}C
      </p>
      <pre className="text-xs font-mono text-[var(--text-dim)] whitespace-pre-wrap break-words max-h-48 overflow-auto bg-[var(--panel-raised)] rounded p-3">
        {code}
      </pre>
    </div>
  );
}

function CodeEditorSectionMobileInner({ editor, onBack, monacoTheme }: { editor: FunctionEditorModel; onBack: () => void; monacoTheme: string }) {
  const { t } = useTranslation();
  const { runtime, code, setCode, markDirty, errors } = editor;
  const [isFullscreen, setIsFullscreen] = useState(false);
  const fullscreenCardRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isFullscreen) return;
    const previouslyActive = document.activeElement as HTMLElement | null;
    const card = fullscreenCardRef.current;
    if (!card) return;
    const focusable = card.querySelectorAll<HTMLElement>('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    first?.focus();
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { setIsFullscreen(false); return; }
      if (e.key === 'Tab' && focusable.length) {
        if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus(); }
        else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus(); }
      }
    };
    document.addEventListener('keydown', handleKey);
    return () => { document.removeEventListener('keydown', handleKey); previouslyActive?.focus(); };
  }, [isFullscreen]);

  const editorHeight = isFullscreen ? 'calc(100vh - 200px)' : '400px';

  return (
    <Card
      ref={fullscreenCardRef}
      className={`overflow-hidden transition-all duration-300 ${isFullscreen ? 'fixed inset-4 z-50' : ''}`}
      style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)', display: 'flex', flexDirection: 'column' }}
      role={isFullscreen ? 'dialog' : undefined}
      aria-modal={isFullscreen ? true : undefined}
      aria-label={t('funcEditor.codeEditor')}
    >
      <CardHeader className="pb-0 pt-4 px-5 shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Code2 className="w-4 h-4 text-[var(--status-ok)]" />
            <span className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>{t('funcEditor.codeEditor')}</span>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-[var(--panel-edge)] text-[var(--text-faint)]">
              {RUNTIME_META[runtime].label}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={onBack}>Back</Button>
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => setIsFullscreen((f) => !f)} aria-label={isFullscreen ? t('funcEditor.exitFullscreen') : t('funcEditor.enterFullscreen')}>
              {isFullscreen ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col min-h-0 mt-3">
        <div className="w-full" style={{ height: editorHeight }}>
          <LazyMonacoEditor
            language={RUNTIME_META[runtime].monacoLang}
            value={code}
            onChange={(v) => { setCode(v || ''); markDirty(); }}
            theme={monacoTheme}
            options={{ minimap: { enabled: false }, fontSize: 13, lineNumbers: 'on', roundedSelection: false, scrollBeyondLastLine: false, automaticLayout: true, tabSize: 2, wordWrap: 'on', padding: { top: 12, bottom: 12 } }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

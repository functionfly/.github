import { useTheme } from '@/components/common/ThemeProvider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { LazyMonacoEditor } from '@/components/LazyMonacoEditor';
import {
  AlertCircle,
  CheckCircle2,
  ChevronRight,
  Clock,
  Code2,
  Copy,
  FileJson,
  Loader2,
  Maximize2,
  Minimize2,
  Play,
  RefreshCw,
  Terminal,
  Zap,
} from 'lucide-react';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { FieldError } from '../components/editor-ui';
import { CODE_TEMPLATES, RUNTIME_META } from '../constants';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function CodeEditorSection({ editor }: Props) {
  const { t } = useTranslation();
  const {
    runtime,
    code,
    setCode,
    activeTab,
    setActiveTab,
    logs,
    errors,
    markDirty,
    testInput,
    setTestInput,
    testResult,
    testTab,
    setTestTab,
    isTesting,
    handleTest,
  } = editor;
  const [isFullscreen, setIsFullscreen] = useState(false);
  const { theme } = useTheme();
  const monacoTheme = theme === 'light' ? 'vs' : 'vs-dark';

  const lineCount = code.split('\n').length;
  const charCount = code.length;

  const handleCopy = useCallback(() => {
    void navigator.clipboard.writeText(code);
    toast.success(t('funcEditor.codeCopiedToClipboard'));
  }, [code]);

  const handleCopyOutput = useCallback(() => {
    if (testResult?.output) {
      void navigator.clipboard.writeText(JSON.stringify(testResult.output, null, 2));
      toast.success(t('funcEditor.outputCopiedToClipboard'));
    }
  }, [testResult]);

  const handleReset = useCallback(() => {
    setCode(CODE_TEMPLATES[runtime]);
    markDirty();
    toast.info(t('funcEditor.codeResetToTemplate'));
  }, [runtime, setCode, markDirty]);

  const editorHeight = isFullscreen ? 'calc(100vh - 200px)' : '400px';
  const testHeight = isFullscreen ? 'calc(100vh - 240px)' : '360px';

  return (
    <Card
      className={`overflow-hidden transition-all duration-300 ${
        isFullscreen ? 'fixed inset-4 z-50' : ''
      }`}
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--panel-edge)',
        borderRadius: 'var(--radius-lg)',
        boxShadow: isFullscreen ? 'var(--shadow-chamber)' : 'var(--shadow-chamber)',
        ...(isFullscreen ? {} : { maxHeight: '520px' }),
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <CardHeader className="pb-0 pt-4 px-5 shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Code2 className="w-4 h-4 text-[var(--status-ok)]" />
            <CardTitle className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>{t('funcEditor.codeEditor')}</CardTitle>
            <Badge
              variant="outline"
              className="text-xs font-mono"
              style={{
                borderColor: RUNTIME_META[runtime].color,
                color: RUNTIME_META[runtime].color,
              }}
            >
              {RUNTIME_META[runtime].label}
            </Badge>
          </div>
          <div className="flex items-center gap-1">
            {activeTab === 'editor' && (
              <span className="text-xs text-[var(--text-faint)] font-mono mr-2 hidden sm:inline">
                {lineCount}L · {charCount}C
              </span>
            )}
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-[var(--text-faint)] hover:text-[var(--text)]"
              onClick={handleCopy}
              aria-label={t('funcEditor.copyCode')}
            >
              <Copy className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-[var(--text-faint)] hover:text-[var(--text)]"
              onClick={handleReset}
              aria-label={t('funcEditor.resetToTemplate')}
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-[var(--text-faint)] hover:text-[var(--text)]"
              onClick={() => setIsFullscreen((f) => !f)}
              aria-label={isFullscreen ? t('funcEditor.exitFullscreen') : t('funcEditor.enterFullscreen')}
            >
              {isFullscreen ? (
                <Minimize2 className="w-3.5 h-3.5" />
              ) : (
                <Maximize2 className="w-3.5 h-3.5" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col min-h-0 mt-3">
        <Tabs
          value={activeTab}
          onValueChange={setActiveTab}
          className="flex-1 flex flex-col min-h-0"
        >
          <TabsList className="grid h-9 w-full grid-cols-4 shrink-0 rounded-none border-b border-[var(--panel-edge)]">
            <TabsTrigger value="editor" className="rounded-none text-xs">
              <Code2 className="w-3 h-3 mr-1.5 hidden sm:inline" />
              {t('funcEditor.editor')}
            </TabsTrigger>
            <TabsTrigger value="test-input" className="rounded-none text-xs">
              <FileJson className="w-3 h-3 mr-1.5 hidden sm:inline" />
              {t('funcEditor.testInput')}
            </TabsTrigger>
            <TabsTrigger value="test-output" className="rounded-none text-xs">
              <Play className="w-3 h-3 mr-1.5 hidden sm:inline" />
              {t('funcEditor.output')}
              {testResult && (
                <span
                  className={`ml-1.5 w-1.5 h-1.5 rounded-full inline-block ${
                    testResult.success ? 'bg-[var(--status-ok)]' : 'bg-[var(--status-revoked)]'
                  }`}
                />
              )}
            </TabsTrigger>
            <TabsTrigger value="logs" className="rounded-none text-xs">
              <Terminal className="w-3 h-3 mr-1.5 hidden sm:inline" />
              {t('funcEditor.logs')}
              {logs.some((l) => l.level === 'error') && (
                <span className="ml-1.5 w-1.5 h-1.5 rounded-full bg-[var(--status-revoked)] inline-block" />
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
            {/* Explicit height: Monaco with height="100%" collapses here because flex parents use auto height. */}
            <div className="w-full" style={{ height: editorHeight }}>
              {activeTab === 'editor' && (
                <LazyMonacoEditor
                  height={editorHeight}
                  language={RUNTIME_META[runtime].monacoLang}
                  value={code}
                  onChange={(v) => {
                    setCode(v || '');
                    markDirty();
                  }}
                  theme={monacoTheme}
                  loading={
                    <div
                        className="flex h-full w-full items-center justify-center text-[var(--text-dim)]"
                      style={{
                        backgroundColor: 'var(--panel-raised)',
                      }}
                    >
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  }
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: 'on',
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    tabSize: 2,
                    wordWrap: 'on',
                    padding: { top: 12, bottom: 12 },
                  }}
                />
              )}
            </div>
            <FieldError message={errors.code} />
          </TabsContent>

          <TabsContent value="test-input" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
            <div className="flex flex-col h-full">
              <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--panel-edge)] bg-[var(--panel-raised)]/50">
                <span className="text-xs text-[var(--text-faint)]">{t('funcEditor.jsonPayloadSent')}</span>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 text-xs gap-1.5"
                    onClick={() =>
                      setTestInput(
                        JSON.stringify(
                          {
                            method: 'GET',
                            path: '/',
                            headers: { 'Content-Type': 'application/json' },
                            body: {},
                          },
                          null,
                          2
                        )
                      )
                    }
                  >
                    {t('funcEditor.reset')}
                  </Button>
                  <Button
                    size="sm"
                    className="h-7 text-xs gap-1.5"
                    onClick={handleTest}
                    disabled={isTesting}
                  >
                    {isTesting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3" />}
                    {t('funcEditor.runTest')}
                  </Button>
                </div>
              </div>
              <div className="flex-1 min-h-0">
                <LazyMonacoEditor
                  height={testHeight}
                  language="json"
                  value={testInput}
                  onChange={(v) => setTestInput(v || '{}')}
                  theme={monacoTheme}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 12,
                    lineNumbers: 'off',
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    tabSize: 2,
                    wordWrap: 'on',
                    padding: { top: 12, bottom: 12 },
                    folding: true,
                    renderLineHighlight: 'none',
                  }}
                />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="test-output" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
            <div className="flex flex-col h-full">
              <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--panel-edge)] bg-[var(--panel-raised)]/50">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[var(--text-faint)]">{t('funcEditor.testResults')}</span>
                  {testResult?.executionTimeMs && (
                    <Badge variant="outline" className="text-xs h-5">
                      <Clock className="w-3 h-3 mr-1" />
                      {testResult.executionTimeMs}ms
                    </Badge>
                  )}
                  {testResult?.coldStartMs && (
                    <Badge variant="outline" className="text-xs h-5">
                      <Zap className="w-3 h-3 mr-1" />
                      Cold: {testResult.coldStartMs}ms
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {testResult?.success ? (
                    <span className="text-xs text-[var(--status-ok)] flex items-center gap-1">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      {t('funcEditor.success')}
                    </span>
                  ) : testResult ? (
                    <span className="text-xs text-[var(--status-revoked)] flex items-center gap-1">
                      <AlertCircle className="w-3.5 h-3.5" />
                      {t('funcEditor.failed')}
                    </span>
                  ) : null}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={handleCopyOutput}
                    disabled={!testResult?.output}
                  >
                    <Copy className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
              <div className="flex-1 min-h-0 p-4">
                {!testResult ? (
                  <div className="flex flex-col items-center justify-center h-full text-[var(--text-faint)]">
                    <Play className="w-8 h-8 mb-3 opacity-30" />
                    <p className="text-sm">{t('funcEditor.runTestToSeeResults')}</p>
                    <p className="text-xs mt-1">{t('funcEditor.clickRunTest')}</p>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {testResult.error && (
                      <div className="p-3 rounded-[var(--radius)]" style={{ background: 'rgba(255, 107, 107, 0.06)', border: '1px solid rgba(255, 107, 107, 0.2)' }}>
                        <p className="text-sm text-[var(--status-revoked)] flex items-center gap-2">
                          <AlertCircle className="w-4 h-4" />
                          {t('funcEditor.error')}
                        </p>
                        <pre className="mt-2 text-xs text-[var(--status-revoked)] font-mono whitespace-pre-wrap">
                          {testResult.error}
                        </pre>
                      </div>
                    )}
                    {testResult.output && (
                      <div>
                        <p className="text-xs text-[var(--text-faint)] mb-2 flex items-center gap-2">
                          <ChevronRight className="w-3 h-3" />
                          Output
                        </p>
                        <pre className="p-3 rounded-lg bg-[var(--panel-raised)] text-xs font-mono text-[var(--text-dim)] overflow-auto">
                          {JSON.stringify(testResult.output, null, 2)}
                        </pre>
                      </div>
                    )}
                    {testResult.logs && testResult.logs.length > 0 && (
                      <div>
                        <p className="text-xs text-[var(--text-faint)] mb-2">{t('funcEditor.executionLogs')}</p>
                        <div className="space-y-1">
                          {testResult.logs.map((log, idx) => (
                            <div key={idx} className="text-xs font-mono flex items-start gap-2">
                              <span
                                className={`w-2 h-2 rounded-full mt-1 shrink-0 ${
                                  log.level === 'error'
                                    ? 'bg-[var(--status-revoked)]'
                                    : log.level === 'warn'
                                      ? 'bg-[var(--status-pending)]'
                                      : log.level === 'success'
                                        ? 'bg-[var(--status-ok)]'
                                        : 'bg-blue-400'
                                }`}
                              />
                              <span className="text-[var(--text-dim)]">{log.message}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="logs" className="mt-0 flex-1 min-h-0 overflow-auto">
            <ScrollArea className="h-full p-4">
              <div className="space-y-1.5">
                {logs.map((log) => (
                  <div key={log.id} className="flex items-start gap-2.5 text-xs font-mono">
                    <span className="text-[var(--text-faint)] w-[72px] shrink-0 pt-0.5">
                      {log.timestamp.split(' ')[1] ?? log.timestamp.slice(-8)}
                    </span>
                    <span className="shrink-0 pt-0.5">
                      {log.level === 'success' && (
                        <CheckCircle2 className="w-3.5 h-3.5 text-[var(--status-ok)]" />
                      )}
                      {log.level === 'error' && (
                        <AlertCircle className="w-3.5 h-3.5 text-[var(--status-revoked)]" />
                      )}
                      {log.level === 'warn' && (
                        <AlertCircle className="w-3.5 h-3.5 text-[var(--status-pending)]" />
                      )}
                      {log.level === 'info' && <Terminal className="w-3.5 h-3.5 text-[var(--foil-a)]" />}
                    </span>
                    <span
                      className={`flex-1 ${
                        log.level === 'error'
                          ? 'text-[var(--status-revoked)]'
                          : log.level === 'success'
                            ? 'text-[var(--status-ok)]'
                            : log.level === 'warn'
                              ? 'text-[var(--status-pending)]'
                              : 'text-[var(--text-dim)]'
                      }`}
                    >
                      {log.message}
                    </span>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </CardContent>

      {/* Fullscreen backdrop */}
      {isFullscreen && (
        <div
          className="fixed inset-0 bg-black/60 z-40"
          onClick={() => setIsFullscreen(false)}
          aria-hidden
        />
      )}
    </Card>
  );
}

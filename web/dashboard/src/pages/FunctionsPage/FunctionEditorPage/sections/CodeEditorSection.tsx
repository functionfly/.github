import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Editor from '@monaco-editor/react';
import {
  AlertCircle,
  CheckCircle2,
  Code2,
  Copy,
  Loader2,
  Maximize2,
  Minimize2,
  RefreshCw,
  Terminal,
} from 'lucide-react';
import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { FieldError } from '../components/editor-ui';
import { CODE_TEMPLATES, RUNTIME_META } from '../constants';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function CodeEditorSection({ editor }: Props) {
  const { runtime, code, setCode, activeTab, setActiveTab, logs, errors, markDirty } = editor;
  const [isFullscreen, setIsFullscreen] = useState(false);

  const lineCount = code.split('\n').length;
  const charCount = code.length;

  const handleCopy = useCallback(() => {
    void navigator.clipboard.writeText(code);
    toast.success('Code copied to clipboard');
  }, [code]);

  const handleReset = useCallback(() => {
    setCode(CODE_TEMPLATES[runtime]);
    markDirty();
    toast.info('Code reset to template');
  }, [runtime, setCode, markDirty]);

  const editorHeight = isFullscreen ? 'calc(100vh - 200px)' : '420px';

  return (
    <Card
      className={`card overflow-hidden transition-all duration-300 ${
        isFullscreen ? 'fixed inset-4 z-50 shadow-2xl' : ''
      }`}
      style={{
        background: 'var(--bg-secondary, #12121a)',
        border: '1px solid rgba(255,255,255,0.08)',
        ...(isFullscreen ? {} : { height: '560px' }),
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <CardHeader className="pb-0 pt-4 px-5 shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Code2 className="w-4 h-4 text-indigo-400" />
            <CardTitle className="text-sm font-semibold text-text-primary">Code Editor</CardTitle>
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
              <span className="text-xs text-text-muted font-mono mr-2 hidden sm:inline">
                {lineCount}L · {charCount}C
              </span>
            )}
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
              onClick={handleCopy}
              aria-label="Copy code"
            >
              <Copy className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
              onClick={handleReset}
              aria-label="Reset to template"
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
              onClick={() => setIsFullscreen((f) => !f)}
              aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
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
          <TabsList className="grid w-full grid-cols-2 rounded-none border-b border-border-subtle shrink-0 bg-transparent h-9">
            <TabsTrigger value="editor" className="rounded-none text-xs">
              Editor
            </TabsTrigger>
            <TabsTrigger value="logs" className="rounded-none text-xs">
              Logs
              {logs.some((l) => l.level === 'error') && (
                <span className="ml-1.5 w-1.5 h-1.5 rounded-full bg-red-400 inline-block" />
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
            <div className="w-full h-full" style={{ minHeight: editorHeight }}>
              {activeTab === 'editor' && (
                <Editor
                  height={isFullscreen ? editorHeight : '100%'}
                  language={RUNTIME_META[runtime].monacoLang}
                  value={code}
                  onChange={(v) => {
                    setCode(v || '');
                    markDirty();
                  }}
                  theme="vs-dark"
                  loading={
                    <div className="flex items-center justify-center w-full h-full bg-[#1e1e1e] text-text-secondary">
                      <Loader2 className="w-6 h-6 animate-spin" />
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

          <TabsContent value="logs" className="mt-0 flex-1 min-h-0 overflow-auto">
            <ScrollArea className="h-full p-4">
              <div className="space-y-1.5">
                {logs.map((log) => (
                  <div key={log.id} className="flex items-start gap-2.5 text-xs font-mono">
                    <span className="text-text-muted w-[72px] shrink-0 pt-0.5">
                      {log.timestamp.split(' ')[1] ?? log.timestamp.slice(-8)}
                    </span>
                    <span className="shrink-0 pt-0.5">
                      {log.level === 'success' && (
                        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                      )}
                      {log.level === 'error' && (
                        <AlertCircle className="w-3.5 h-3.5 text-red-400" />
                      )}
                      {log.level === 'warn' && (
                        <AlertCircle className="w-3.5 h-3.5 text-amber-400" />
                      )}
                      {log.level === 'info' && <Terminal className="w-3.5 h-3.5 text-blue-400" />}
                    </span>
                    <span
                      className={`flex-1 ${
                        log.level === 'error'
                          ? 'text-red-300'
                          : log.level === 'success'
                            ? 'text-emerald-300'
                            : log.level === 'warn'
                              ? 'text-amber-300'
                              : 'text-text-secondary'
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

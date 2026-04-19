import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Editor from '@monaco-editor/react';
import { useTheme } from '@/components/common/ThemeProvider';
import { CODE_TEMPLATES, RUNTIME_META } from '../constants';
import type { FunctionEditorModel } from '../useFunctionEditor';
import { formatTimeout } from '../utils';
import { FieldError } from './editor-ui';
import {
  AlertCircle,
  CheckCircle2,
  Code2,
  Copy,
  Loader2,
  RefreshCw,
  Sparkles,
  Terminal,
} from 'lucide-react';
import { toast } from 'sonner';

type Props = { editor: FunctionEditorModel };

export function FunctionEditorRightColumn({ editor }: Props) {
  const { theme } = useTheme();
  const monacoTheme = theme === 'light' ? 'vs' : 'vs-dark';
  const {
    runtime,
    code,
    setCode,
    activeTab,
    setActiveTab,
    logs,
    errors,
    markDirty,
    functionName,
    slug,
    runtimeVersion,
    resources,
    visibility,
    httpTrigger,
    scheduleTrigger,
    envVars,
    tags,
    retryPolicy,
  } = editor;

  return (
    <div className="space-y-5">
      <Card
        className="card overflow-hidden"
        style={{
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border-subtle)',
          height: '560px',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <CardHeader className="pb-0 pt-4 px-5 shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Code2 className="w-4 h-4 text-[#FF6B35]" />
              <CardTitle className="text-sm font-semibold text-text-primary font-display">Code Editor</CardTitle>
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
            <div className="flex gap-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                onClick={() => {
                  void navigator.clipboard.writeText(code);
                  toast.success('Code copied');
                }}
                aria-label="Copy code"
              >
                <Copy className="w-3.5 h-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                onClick={() => {
                  setCode(CODE_TEMPLATES[runtime]);
                  markDirty();
                  toast.info('Code reset to template');
                }}
                aria-label="Reset to template"
              >
                <RefreshCw className="w-3.5 h-3.5" />
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

            <TabsContent
              value="editor"
              className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden"
            >
              <div className="w-full h-full" style={{ minHeight: '420px' }}>
                {activeTab === 'editor' && (
                  <Editor
                    height="100%"
                    language={RUNTIME_META[runtime].monacoLang}
                    value={code}
                    onChange={(v) => {
                      setCode(v || '');
                      markDirty();
                    }}
                    theme={monacoTheme}
                    loading={
                      <div
                        className="flex h-full w-full items-center justify-center text-text-secondary"
                        style={{
                          backgroundColor: 'var(--bg-tertiary)',
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
                        {log.level === 'info' && (
                          <Terminal className="w-3.5 h-3.5 text-blue-400" />
                        )}
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
      </Card>

      <Card
        className="card"
        style={{
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border-subtle)',
        }}
      >
        <CardHeader className="pb-3 pt-4 px-5">
          <div className="flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-[#FF6B35]" />
            <CardTitle className="text-sm font-semibold text-text-primary font-display">
              Configuration Summary
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="px-5 pb-5 space-y-3">
          {[
            { label: 'Name', value: functionName || '—' },
            { label: 'Slug', value: slug || '—', mono: true },
            { label: 'Runtime', value: `${RUNTIME_META[runtime].label} ${runtimeVersion}` },
            { label: 'Memory', value: `${resources.memoryMb} MB` },
            { label: 'Timeout', value: formatTimeout(resources.timeoutMs) },
            { label: 'Concurrency', value: `${resources.maxConcurrency}` },
            {
              label: 'Visibility',
              value: visibility === 'public' ? '🌐 Public' : '🔒 Private',
            },
            {
              label: 'HTTP Trigger',
              value: httpTrigger.enabled
                ? `${httpTrigger.method} ${httpTrigger.path}`
                : 'Disabled',
            },
            {
              label: 'Schedule',
              value: scheduleTrigger.enabled ? scheduleTrigger.cron : 'Disabled',
            },
            {
              label: 'Env Vars',
              value: `${envVars.length} variable${envVars.length !== 1 ? 's' : ''}`,
            },
            { label: 'Tags', value: tags.length > 0 ? tags.join(', ') : 'None' },
            {
              label: 'Retries',
              value: `${retryPolicy.maxRetries} (${retryPolicy.backoffMs}ms backoff)`,
            },
          ].map(({ label, value, mono }) => (
            <div key={label} className="flex items-start justify-between gap-3 text-xs">
              <span className="text-text-muted shrink-0">{label}</span>
              <span className={`text-text-secondary text-right ${mono ? 'font-mono' : ''}`}>
                {value}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

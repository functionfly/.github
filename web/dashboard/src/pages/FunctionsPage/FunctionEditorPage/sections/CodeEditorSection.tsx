import { useTheme } from '@/components/common/ThemeProvider';
import { Badge } from '@/components/ui/badge';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { LazyMonacoEditor, LazyMonacoDiffEditor } from '@/components/LazyMonacoEditor';
import type { OnMount, BeforeMount } from '@monaco-editor/react';
import {
  AlertCircle,
  CheckCircle2,
  ChevronRight,
  Clock,
  Code2,
  Copy,
  Diff,
  FileJson,
  Loader2,
  Maximize2,
  Minimize2,
  Play,
  RefreshCw,
  Terminal,
  X,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import * as monaco from 'monaco-editor';
import { FieldError } from '../components/editor-ui';
import { CODE_TEMPLATES, RUNTIME_META } from '../constants';
import { validateCodeAsync, type ValidationIssue, getMonacoLanguage } from '../utils/codeValidation';
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
    cancelTest,
    isDirty,
    handleRuntimeChange,
    validationIssues,
    setValidationIssues,
  } = editor;
  const [isFullscreen, setIsFullscreen] = useState(() => {
    if (typeof window === 'undefined') return false;
    try {
      const saved = sessionStorage.getItem('functionfly:editor-fullscreen');
      return saved === 'true';
    } catch {
      return false;
    }
  });
  const [isDiffMode, setIsDiffMode] = useState(false);
  const [originalCode, setOriginalCode] = useState<string | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  } | null>(null);
  const [cursorPosition, setCursorPosition] = useState({ line: 1, column: 1 });
  const [lsStatus, setLsStatus] = useState<'loading' | 'ready' | 'error'>('ready');
  const [tabSize, setTabSize] = useState(2);
  const { theme } = useTheme();
  const monacoTheme = theme === 'light' ? 'vs' : 'vs-dark';
  const fullscreenCardRef = useRef<HTMLDivElement>(null);
  const monacoEditorRef = useRef<import('monaco-editor').editor.IStandaloneCodeEditor | null>(null);
  const monacoInstanceRef = useRef<typeof import('monaco-editor') | null>(null);

  const lineCount = code.split('\n').length;
  const charCount = code.length;

  useEffect(() => {
    if (typeof window === 'undefined') return;
    try {
      sessionStorage.setItem('functionfly:editor-fullscreen', String(isFullscreen));
    } catch {
      /* ignore */
    }
  }, [isFullscreen]);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (!isDirty) return;
    const timeoutId = setTimeout(() => {
      try {
        localStorage.setItem(
          'functionfly:editor-draft',
          JSON.stringify({
            code,
            runtime,
            savedAt: new Date().toISOString(),
          })
        );
      } catch {
        /* ignore */
      }
    }, 1000);
    return () => clearTimeout(timeoutId);
  }, [code, runtime, isDirty]);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    try {
      const saved = localStorage.getItem('functionfly:editor-draft');
      if (saved) {
        const draft = JSON.parse(saved);
        if (draft.runtime === runtime && draft.code !== code) {
          setConfirmDialog({
            open: true,
            title: 'Recover Draft?',
            message: `A draft was saved ${new Date(draft.savedAt).toLocaleString()}. Would you like to recover it?`,
            onConfirm: () => {
              setCode(draft.code);
              markDirty();
              setConfirmDialog(null);
            },
          });
        }
      }
    } catch {
      /* ignore */
    }
  }, [runtime]);

  useEffect(() => {
    if (!isFullscreen) return;
    const previouslyActive = document.activeElement as HTMLElement | null;
    const card = fullscreenCardRef.current;
    if (!card) return;
    const focusable = card.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    first?.focus();

    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsFullscreen(false);
        return;
      }
      if (e.key === 'Tab' && focusable.length) {
        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last?.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first?.focus();
        }
      }
    };
    document.addEventListener('keydown', handleKey);
    return () => {
      document.removeEventListener('keydown', handleKey);
      previouslyActive?.focus();
    };
  }, [isFullscreen]);

  const isTestInputInvalid = useMemo(() => {
    if (activeTab !== 'test-input') return false;
    try {
      JSON.parse(testInput);
      return false;
    } catch {
      return true;
    }
  }, [testInput, activeTab]);

  const fallbackCopy = useCallback((text: string) => {
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    ta.style.pointerEvents = 'none';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    if (ok) toast.success(t('funcEditor.codeCopiedToClipboard'));
    else toast.error('Clipboard unavailable');
  }, [t]);

  const handleCopy = useCallback(() => {
    if (navigator?.clipboard?.writeText) {
      navigator.clipboard.writeText(code).catch(() => fallbackCopy(code));
    } else {
      fallbackCopy(code);
    }
  }, [code, fallbackCopy]);

  const handleCopyOutput = useCallback(() => {
    if (!testResult?.output) return;
    const text = JSON.stringify(testResult.output, null, 2);
    if (navigator?.clipboard?.writeText) {
      navigator.clipboard.writeText(text).catch(() => fallbackCopy(text));
    } else {
      fallbackCopy(text);
    }
  }, [testResult, fallbackCopy]);

  useEffect(() => {
    const editor = monacoEditorRef.current;
    if (!editor) return;

    const handleScroll = () => {
      if (typeof window === 'undefined') return;
      try {
        const state = editor.saveViewState();
        if (state) {
          sessionStorage.setItem('functionfly:editor-scroll', JSON.stringify(state));
        }
      } catch {
        /* ignore */
      }
    };

    const disposable = editor.onDidScrollChange(handleScroll);
    return () => {
      disposable.dispose();
    };
  }, []);

  useEffect(() => {
    const editor = monacoEditorRef.current;
    if (!editor) return;
    if (typeof window === 'undefined') return;
    try {
      const saved = sessionStorage.getItem('functionfly:editor-scroll');
      if (saved) {
        const state = JSON.parse(saved);
        editor.restoreViewState(state);
        editor.focus();
      }
    } catch {
      /* ignore */
    }
  }, [code, runtime]);

  const handleFullscreenToggle = useCallback(() => {
    setIsFullscreen((f) => !f);
  }, []);

  const editorHeight = isFullscreen ? 'calc(100vh - 200px)' : '400px';
  const testHeight = isFullscreen ? 'calc(100vh - 240px)' : '360px';

  const monacoSeverity = (type: ValidationIssue['type']) => {
    switch (type) {
      case 'error':
        return monaco.MarkerSeverity.Error;
      case 'warning':
        return monaco.MarkerSeverity.Warning;
      default:
        return monaco.MarkerSeverity.Info;
    }
  };

  useEffect(() => {
    const editor = monacoEditorRef.current;
    if (!editor) return;
    const model = editor.getModel();
    if (!model) return;

    let cancelled = false;
    let timeoutId: ReturnType<typeof setTimeout>;

    const runValidation = () => {
      validateCodeAsync(code, runtime).then((issues) => {
        if (cancelled) return;
        setValidationIssues(issues.issues);

        const markers = issues.issues.map((issue) => ({
          severity: monacoSeverity(issue.type),
          message: issue.message,
          startLineNumber: issue.line ?? 1,
          startColumn: issue.column ?? 1,
          endLineNumber: issue.endLine ?? issue.line ?? 1,
          endColumn: issue.endColumn ?? (issue.column ?? 1) + (issue.message.length || 10),
        }));

        monaco.editor.setModelMarkers(model, 'functionfly', markers);
      });
    };

    timeoutId = setTimeout(runValidation, 300);

    return () => {
      cancelled = true;
      clearTimeout(timeoutId);
    };
  }, [code, runtime, setValidationIssues]);

  const handleBeforeMount: BeforeMount = useCallback((monacoInstance) => {
    monacoInstance.languages.typescript.typescriptDefaults.setMaximumWorkerIdleTime(5000);

    const functionflyTypes = `
      declare const Request: {
        new (input: RequestInfo | URL, init?: RequestInit): Request;
        prototype: Request;
      };

      declare const Response: {
        new (body?: BodyInit | null, init?: ResponseInit): Response;
        prototype: Response;
        json(data: unknown, init?: ResponseInit): Response;
        redirect(url: string, status?: number): Response;
        error(): Response;
      };

      declare const Headers: {
        prototype: Headers;
        new (init?: HeadersInit): Headers;
      };

      declare const URL: {
        prototype: URL;
        new (url: string | URL, base?: string | URL): URL;
      };

      declare const URLSearchParams: {
        prototype: URLSearchParams;
        new (init?: string | URLSearchParams | Record<string, string>): URLSearchParams;
      };

      interface ExecutionContext {
        waitUntil(promise: Promise<unknown>): void;
        passThroughOnException(): void;
      }

      interface Env {
        [key: string]: string | undefined;
      }

      declare const env: Env;
      declare const ctx: ExecutionContext;
    `;

    monacoInstance.languages.typescript.typescriptDefaults.setCompilerOptions({
      target: monacoInstance.languages.typescript.ScriptTarget.ES2020,
      lib: ['ES2022', 'ES2023', 'DOM'],
      allowJs: true,
      allowSyntheticDefaultImports: true,
      esModuleInterop: true,
      strict: true,
      noImplicitAny: true,
      strictNullChecks: true,
      noUnusedLocals: false,
      noUnusedParameters: false,
    });

    monacoInstance.languages.typescript.typescriptDefaults.setDiagnosticsOptions({
      noSemanticValidation: false,
      noSyntaxValidation: false,
      noSuggestionDiagnostics: false,
      diagnosticCodesToIgnore: [
        1108,
        2322,
        2339,
        2345,
        2377,
        2384,
        2393,
        2403,
        2420,
        2445,
        2451,
        2456,
        2471,
        2481,
        2497,
        2503,
        2504,
        2515,
        2538,
        2540,
        2542,
        2545,
        2547,
        2549,
        2551,
        2552,
        2554,
        2555,
        2557,
        2565,
        2571,
        2586,
        2591,
        2600,
        2601,
        2614,
        2622,
        2628,
        2636,
        2641,
        2649,
        2651,
        2653,
        2663,
        2664,
        2666,
        2676,
        2686,
        2693,
        2700,
        2705,
        2724,
        2726,
        2727,
        2732,
        2737,
        2738,
        2740,
        2741,
        2743,
        2744,
        2745,
        2747,
        2748,
        2749,
        2753,
        2754,
        2755,
        2756,
        2757,
        2762,
        2763,
        2766,
        2769,
        2770,
        2771,
        2774,
        2775,
        2779,
        2784,
        2785,
        2786,
        2787,
        2792,
        2794,
        2796,
        2799,
        2800,
        2801,
        2802,
        2803,
        2805,
        2806,
        2810,
        2812,
        2813,
        2814,
        2815,
        2816,
        2817,
        2818,
        2819,
        2820,
        2821,
        2822,
        2823,
        2824,
        2825,
        2826,
        2827,
        2828,
        2829,
        2830,
        2831,
        2832,
        2833,
        2834,
        2835,
        2836,
        2837,
        2838,
        2839,
        2840,
        2841,
        2842,
        2843,
        2844,
        2845,
        2846,
        2847,
        2850,
        2851,
        2852,
        2853,
        2854,
        2855,
        2856,
        2857,
        2858,
        2859,
        2860,
        2861,
        2862,
        2863,
        2864,
        2865,
        2866,
        2867,
        2868,
        2869,
        2870,
        2871,
        2872,
        2873,
        2874,
        2875,
        2876,
        2877,
        2878,
        2879,
        2881,
        2883,
        2884,
        2885,
        2886,
        2887,
        2888,
        2889,
        2890,
        2891,
        2892,
        2893,
        2894,
        2895,
        2896,
        2897,
        2898,
        2899,
        2900,
        2901,
        2902,
        2903,
        2904,
        2905,
        2906,
        2907,
        2908,
        2909,
        2910,
        2911,
        2912,
        2913,
        2914,
        2915,
        2916,
        2917,
        2918,
        2919,
        2920,
        2921,
        2922,
        2923,
        2924,
        2925,
        2926,
        2927,
        2928,
        2929,
        2930,
        2931,
        2932,
        2933,
        2934,
        2935,
        2936,
        2937,
        2938,
        2939,
        2940,
        2941,
        2942,
        2943,
        2944,
        2945,
        2946,
        2947,
        2948,
        2949,
        2950,
        2951,
        2952,
        2953,
        2954,
        2955,
        2956,
        2957,
        2958,
        2959,
        2960,
        2961,
        2962,
        2963,
        2964,
        2965,
        2966,
        2967,
        2968,
        2969,
        2970,
        2971,
        2972,
        2973,
        2974,
        2975,
        2976,
        2977,
        2978,
        2979,
        2980,
        2981,
        2982,
        2983,
        2984,
        2985,
        2986,
        2987,
        2988,
        2989,
        2990,
        2991,
        2992,
        2993,
        2994,
        2995,
        2996,
        2997,
        2998,
        2999,
        4028,
      ],
    });

    monacoInstance.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
      noSemanticValidation: false,
      noSyntaxValidation: false,
      noSuggestionDiagnostics: false,
    });

    monacoInstance.languages.typescript.javascriptDefaults.setCompilerOptions({
      target: monacoInstance.languages.typescript.ScriptTarget.ES2020,
      lib: ['ES2022', 'ES2023', 'DOM'],
      allowJs: true,
      allowSyntheticDefaultImports: true,
      esModuleInterop: true,
    });

    monacoInstance.languages.typescript.typescriptDefaults.addExtraLib(
      functionflyTypes,
      'file:///node_modules/@functionfly/runtime-types/index.d.ts'
    );

    monacoInstance.languages.typescript.javascriptDefaults.addExtraLib(
      functionflyTypes.replace(/:\s*Promise</g, '<').replace(/Promise</g, '<'),
      'file:///node_modules/@functionfly/runtime-types/javascript.d.ts'
    );
  }, []);

  const handleEditorMount: OnMount = useCallback((editor, _monaco) => {
    monacoEditorRef.current = editor as unknown as import('monaco-editor').editor.IStandaloneCodeEditor;
    monacoInstanceRef.current = monaco;

    editor.onDidChangeCursorPosition((e) => {
      setCursorPosition({ line: e.position.lineNumber, column: e.position.column });
    });

    editor.onDidChangeCursorSelection((e) => {
      const model = editor.getModel();
      if (model) {
        setTabSize(model.getOptions().tabSize);
      }
    });

    setLsStatus('ready');

    editor.addAction({
      id: 'functionfly.save',
      label: 'Save',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS],
      run: () => {
        if (isDirty) {
          editor.trigger('keyboard', 'editor.action.formatDocument', null);
        }
      },
    });

    editor.addAction({
      id: 'functionfly.run-test',
      label: 'Run Test',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
      run: () => {
        if (!isTesting && !isTestInputInvalid) {
          handleTest();
        }
      },
    });

    editor.addAction({
      id: 'functionfly.toggle-fullscreen',
      label: 'Toggle Fullscreen',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF],
      run: () => {
        handleFullscreenToggle();
      },
    });

    editor.addAction({
      id: 'functionfly.reset-code',
      label: 'Reset to Template',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyR],
      run: () => {
        if (isDirty) {
          setConfirmDialog({
            open: true,
            title: 'Reset to Template?',
            message: 'Resetting the code will discard all unsaved changes. Continue?',
            onConfirm: () => {
              setCode(CODE_TEMPLATES[runtime]);
              markDirty();
              toast.info(t('funcEditor.codeResetToTemplate'));
              setConfirmDialog(null);
            },
          });
        } else {
          setCode(CODE_TEMPLATES[runtime]);
          markDirty();
          toast.info(t('funcEditor.codeResetToTemplate'));
        }
      },
    });

    editor.addAction({
      id: 'functionfly.find',
      label: 'Find',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyF],
      run: () => {
        editor.getAction('actions.find')?.run();
      },
    });

    editor.addAction({
      id: 'functionfly.find-and-replace',
      label: 'Find and Replace',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyH],
      run: () => {
        editor.getAction('editor.action.startFindReplaceAction')?.run();
      },
    });

    editor.addAction({
      id: 'functionfly.toggle-folding',
      label: 'Toggle Code Folding',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.BracketLeft],
      run: () => {
        editor.getAction('editor.toggleFold')?.run();
      },
    });

    editor.addAction({
      id: 'functionfly.fold-all',
      label: 'Fold All',
      run: () => {
        editor.getAction('editor.foldAll')?.run();
      },
    });

    editor.addAction({
      id: 'functionfly.unfold-all',
      label: 'Unfold All',
      run: () => {
        editor.getAction('editor.unfoldAll')?.run();
      },
    });

    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyP, () => {
      editor.trigger('keyboard', 'editor.action.quickCommand', null);
    });
  }, [handleTest, handleFullscreenToggle, isDirty, isTesting, isTestInputInvalid, runtime, setCode, markDirty, t]);

  return (
    <Card
      ref={fullscreenCardRef}
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
      role={isFullscreen ? 'dialog' : undefined}
      aria-modal={isFullscreen ? true : undefined}
      aria-label={t('funcEditor.codeEditor')}
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
              className={`h-7 w-7 p-0 hover:text-[var(--text)] ${
                isDiffMode ? 'text-[var(--status-ok)]' : 'text-[var(--text-faint)]'
              }`}
              onClick={() => {
                if (!isDiffMode) {
                  setOriginalCode(code);
                  setIsDiffMode(true);
                  toast.info('Comparing with original. Edit the right side to see changes.');
                } else {
                  setIsDiffMode(false);
                  setOriginalCode(null);
                }
              }}
              aria-label={isDiffMode ? t('funcEditor.exitDiffMode') : t('funcEditor.compareVersions')}
            >
              <Diff className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 px-2 gap-1.5 text-[var(--text-faint)] hover:text-[var(--text)]"
              onClick={() => {
                const editor = monacoEditorRef.current;
                if (editor) {
                  editor.getAction('editor.action.formatDocument')?.run();
                }
              }}
              aria-label={t('funcEditor.formatCode')}
            >
              <span className="text-xs font-mono">fmt</span>
            </Button>
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
              onClick={() => {
                if (isDirty) {
                  setConfirmDialog({
                    open: true,
                    title: 'Reset to Template?',
                    message: 'Resetting the code will discard all unsaved changes. Continue?',
                    onConfirm: () => {
                      setCode(CODE_TEMPLATES[runtime]);
                      markDirty();
                      toast.info(t('funcEditor.codeResetToTemplate'));
                      setConfirmDialog(null);
                    },
                  });
                } else {
                  setCode(CODE_TEMPLATES[runtime]);
                  markDirty();
                  toast.info(t('funcEditor.codeResetToTemplate'));
                }
              }}
              aria-label={t('funcEditor.resetToTemplate')}
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-[var(--text-faint)] hover:text-[var(--text)]"
              onClick={handleFullscreenToggle}
              aria-label={isFullscreen ? t('funcEditor.exitFullscreen') : t('funcEditor.enterFullscreen')}
            >
              {isFullscreen ? (
                <Minimize2 className="w-3.5 h-3.5" />
              ) : (
                <Maximize2 className="w-3.5 h-3.5" />
              )}
            </Button>
            {isFullscreen && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-[var(--text-faint)] hover:text-[var(--text)]"
                onClick={handleFullscreenToggle}
                aria-label="Close fullscreen (Esc)"
              >
                <X className="w-3.5 h-3.5" />
              </Button>
            )}
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
            <div className="w-full" style={{ height: editorHeight }}>
              {activeTab === 'editor' && isDiffMode && originalCode !== null ? (
                <LazyMonacoDiffEditor
                  language={RUNTIME_META[runtime].monacoLang}
                  original={originalCode}
                  modified={code}
                  theme={monacoTheme}
                  options={{
                    readOnly: true,
                    renderSideBySide: true,
                    diffWordWrap: 'on',
                    automaticLayout: true,
                    folding: true,
                    matchBrackets: 'always',
                    fontSize: 13,
                    minimap: { enabled: false },
                  }}
                />
              ) : (
                <LazyMonacoEditor
                  language={RUNTIME_META[runtime].monacoLang}
                  value={code}
                  onChange={(v) => {
                    setCode(v || '');
                    markDirty();
                  }}
                  theme={monacoTheme}
                  onMount={handleEditorMount}
                  beforeMount={handleBeforeMount}
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
                    folding: true,
                    foldingStrategy: 'auto',
                    showFoldingControls: 'mouseover',
                    matchBrackets: 'always',
                    autoClosingBrackets: 'languageDefined',
                    autoClosingQuotes: 'languageDefined',
                    autoIndent: 'full',
                    formatOnPaste: true,
                    formatOnType: true,
                    suggestOnTriggerCharacters: true,
                    acceptSuggestionOnEnter: 'smart',
                    tabCompletion: 'on',
                    parameterHints: { enabled: true },
                    quickSuggestions: { other: true, comments: false, strings: false },
                  }}
                />
              )}
            </div>
            <div
              className="mt-1.5 flex items-center justify-between gap-4 px-2 py-1 rounded bg-[var(--panel-raised)]/50"
              role="status"
              aria-live="polite"
              aria-label="Editor status"
            >
              <div className="flex items-center gap-4">
                <FieldError message={errors.code} />
                {isDiffMode && (
                  <span className="text-[10px] text-[var(--status-ok)]">Diff mode: comparing with saved version</span>
                )}
                {code.length > 100000 && (
                  <span className="text-[10px] text-[var(--status-pending)] flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    Large file ({Math.round(code.length / 1024)}KB)
                  </span>
                )}
              </div>
              <div className="flex items-center gap-3 text-[10px] text-[var(--text-faint)] font-mono">
                <span aria-label={`Cursor at line ${cursorPosition.line}, column ${cursorPosition.column}`}>
                  Ln {cursorPosition.line}, Col {cursorPosition.column}
                </span>
                <Separator orientation="vertical" className="h-3" />
                <span className="flex items-center gap-1" aria-label={`Language server status: ${lsStatus}`}>
                  {lsStatus === 'loading' && <Loader2 className="w-3 h-3 animate-spin" />}
                  {lsStatus === 'error' && <AlertCircle className="w-3 h-3 text-[var(--status-revoked)]" />}
                  {lsStatus === 'ready' && <CheckCircle2 className="w-3 h-3 text-[var(--status-ok)]" />}
                  {RUNTIME_META[runtime].label}
                </span>
                <Separator orientation="vertical" className="h-3" />
                <span aria-label="UTF-8 encoding">UTF-8</span>
                <Separator orientation="vertical" className="h-3" />
                <span aria-label={`Indentation: ${tabSize} spaces`}>Spaces: {tabSize}</span>
              </div>
            </div>
            <div className="sr-only" role="alert" aria-live="assertive">
              {validationIssues.filter(i => i.type === 'error').map((issue, idx) => (
                <p key={idx}>Error on line {issue.line}: {issue.message}</p>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="test-input" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
            <div className="flex flex-col h-full">
              <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--panel-edge)] bg-[var(--panel-raised)]/50">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[var(--text-faint)]">{t('funcEditor.jsonPayloadSent')}</span>
                  {isTestInputInvalid && (
                    <span className="text-[10px] text-[var(--status-revoked)] font-medium">Invalid JSON</span>
                  )}
                </div>
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
                    disabled={isTesting || isTestInputInvalid}
                  >
                    {isTesting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3" />}
                    {t('funcEditor.runTest')}
                  </Button>
                  {isTesting && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs gap-1.5"
                      onClick={cancelTest}
                    >
                      <X className="w-3 h-3" />
                      Cancel
                    </Button>
                  )}
                </div>
              </div>
              <div className={`flex-1 min-h-0 ${isTestInputInvalid ? 'ring-1 ring-inset ring-[var(--status-revoked)]' : ''}`}>
                <LazyMonacoEditor
                  language="json"
                  value={testInput}
                  onChange={(v) => setTestInput(v || '{}')}
                  theme={monacoTheme}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 12,
                    lineNumbers: 'on',
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    tabSize: 2,
                    wordWrap: 'on',
                    padding: { top: 12, bottom: 12 },
                    folding: true,
                    renderLineHighlight: 'none',
                    matchBrackets: 'always',
                    autoClosingBrackets: 'always',
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
                    aria-label={t('funcEditor.copyOutput')}
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
                      {log.timestamp.includes('T')
                        ? log.timestamp.slice(11, 19)
                        : log.timestamp.split(' ')[1] ?? log.timestamp.slice(-8)}
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
          onClick={handleFullscreenToggle}
          aria-hidden
        />
      )}

      {/* Confirm Dialog */}
      <AlertDialog open={confirmDialog?.open ?? false} onOpenChange={(open) => {
        if (!open) setConfirmDialog(null);
      }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmDialog?.title}</AlertDialogTitle>
            <AlertDialogDescription>{confirmDialog?.message}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setConfirmDialog(null)}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDialog?.onConfirm}>Continue</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}

import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { useCodeAnalyzer } from '../../hooks/useCodeAnalyzer';
import { useFunctionImport } from '../../hooks/useFunctionImport';
import { useCodePasteStore } from '../../stores/codePasteStore';
import type { ParsedFunction } from '../../types/codePaste';
import { CodeEditor } from './CodeEditor';
import { CodeSizeIndicator } from './CodeSizeIndicator';
import './CodePasteWorkspace.css';
import { FunctionList } from './FunctionList';
import { FunctionPreviewPanel } from './FunctionPreviewPanel';
import { ImportConfigPanel } from './ImportConfigPanel';
import { ImportErrorsList } from './ImportErrorsList';
import { LanguageDetectorBadge } from './LanguageDetectorBadge';
import { CodePasteEmptyState } from './CodePasteEmptyState';

export interface CodePasteWorkspaceProps {
  initialCode?: string;
  onCancel?: () => void;
  onImportComplete?: (functions: Array<{ id: string; name: string }>) => void;
}

const EXAMPLE_CODE = {
  python: `def hello(name: str) -> str:
    """Greet someone by name."""
    return f"Hello, {name}!"

def calculate_total(items: list, tax_rate: float = 0.1) -> float:
    """Calculate total with tax."""
    subtotal = sum(items)
    return subtotal * (1 + tax_rate)

def process_data(data: dict, include_metadata: bool = False) -> dict:
    """Process data and optionally include metadata."""
    result = {"processed": True, "data": data}
    if include_metadata:
        result["metadata"] = {"timestamp": "2024-01-01"}
    return result`,

  javascript: `function hello(name) {
    return \`Hello, \${name}!\`;
}

const calculateTotal = (items, taxRate = 0.1) => {
    const subtotal = items.reduce((a, b) => a + b, 0);
    return subtotal * (1 + taxRate);
};

async function fetchData(url) {
    const response = await fetch(url);
    return response.json();
}`,

  go: `package main

func hello(name string) string {
    return "Hello, " + name + "!"
}

func calculateTotal(items []float64, taxRate float64) float64 {
    var subtotal float64
    for _, item := range items {
        subtotal += item
    }
    return subtotal * (1 + taxRate)
}`,
};

export function CodePasteWorkspace({
  initialCode = '',
  onCancel,
  onImportComplete,
}: CodePasteWorkspaceProps) {
  const [previewFunction, setPreviewFunction] = useState<ParsedFunction | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [showDiscardDialog, setShowDiscardDialog] = useState(false);
  const [importErrors, setImportErrors] = useState<Array<{ name: string; error: string }>>([]);

  const {
    code,
    language,
    confidence,
    functions,
    status,
    error,
    setCode,
    setLanguage,
    parseCode,
    clearAnalysis,
  } = useCodeAnalyzer();

  const { status: importStatus, importFunctions, createdFunctions, reset: resetImport } = useFunctionImport();

  const {
    selectedIds,
    importConfig,
    toggleSelection,
    selectAll,
    clearSelection,
    setImportConfig,
    setFunctions,
    reset: resetStore,
  } = useCodePasteStore();

  const isDirty = useMemo(
    () => code.trim().length > 0 || functions.length > 0 || status === 'parsed',
    [code, functions.length, status]
  );

  useEffect(() => {
    if (initialCode.trim()) {
      setCode(initialCode);
    }
  }, [initialCode, setCode]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault();
        if (status !== 'parsing' && code.trim() && !error) {
          void parseCode();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [code, error, parseCode, status]);

  const resetWorkspace = useCallback(() => {
    clearAnalysis();
    resetImport();
    resetStore();
    setImportErrors([]);
  }, [clearAnalysis, resetImport, resetStore]);

  const handleCancel = useCallback(() => {
    if (isDirty) {
      setShowDiscardDialog(true);
      return;
    }
    resetWorkspace();
    onCancel?.();
  }, [isDirty, onCancel, resetWorkspace]);

  const handleConfirmDiscard = useCallback(() => {
    setShowDiscardDialog(false);
    resetWorkspace();
    onCancel?.();
  }, [onCancel, resetWorkspace]);

  const handleParse = useCallback(async () => {
    setImportErrors([]);
    await parseCode();
  }, [parseCode]);

  const handleImport = useCallback(async () => {
    const success = await importFunctions(functions, selectedIds, importConfig);

    if (success && createdFunctions.length > 0) {
      toast.success(
        `Successfully imported ${createdFunctions.length} function${createdFunctions.length !== 1 ? 's' : ''}`
      );
      onImportComplete?.(createdFunctions);
      resetWorkspace();
      return;
    }

    if (importErrors.length > 0) {
      setImportErrors(importErrors);
      toast.error(
        importErrors.length === 1
          ? `Import failed: ${importErrors[0].error}`
          : `${importErrors.length} functions failed to import`
      );
    }
  }, [functions, selectedIds, importConfig, importFunctions, createdFunctions, importErrors, onImportComplete, resetWorkspace]);

  const handleFunctionPreview = useCallback((func: ParsedFunction) => {
    setPreviewFunction(func);
    setShowPreview(true);
  }, []);

  const handleFunctionNameChange = useCallback(
    (id: string, name: string) => {
      setFunctions(functions.map((f) => (f.id === id ? { ...f, name } : f)));
    },
    [functions, setFunctions]
  );

  const handleExampleClick = useCallback(
    (lang: keyof typeof EXAMPLE_CODE) => {
      setCode(EXAMPLE_CODE[lang]);
    },
    [setCode]
  );

  return (
    <>
      <section className="code-paste-workspace">
        <div className="code-paste-workspace__body">
          <div className="code-paste-workspace__input-panel">
            <div className="code-paste-workspace__editor-header">
              <LanguageDetectorBadge
                language={language}
                confidence={confidence}
                isParsing={status === 'parsing'}
                hasCode={code.trim().length > 0}
                onLanguageChange={setLanguage}
              />
              <div className="code-paste-workspace__examples">
                <span>Try:</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => handleExampleClick('python')}
                >
                  Python
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => handleExampleClick('javascript')}
                >
                  JavaScript
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => handleExampleClick('go')}
                >
                  Go
                </Button>
              </div>
            </div>

            <div className="code-paste-workspace__editor">
              <CodeEditor
                value={code}
                onChange={setCode}
                language={language}
                height="350px"
              />
            </div>

            <CodeSizeIndicator code={code} />

            <div className="code-paste-workspace__actions">
              <Button
                type="button"
                variant="outline"
                onClick={clearAnalysis}
                disabled={status === 'parsing' || !code}
              >
                Clear
              </Button>
              <Button
                type="button"
                onClick={handleParse}
                disabled={status === 'parsing' || !code || !!error}
                isLoading={status === 'parsing'}
              >
                {status === 'parsing' ? 'Analyzing...' : 'Parse Code'}
              </Button>
            </div>

            <p className="code-paste-workspace__shortcut-hint">
              Tip: press <kbd>Ctrl</kbd>+<kbd>Enter</kbd> to parse
            </p>

            {error && (
              <div className="code-paste-workspace__error" role="alert">
                {error}
              </div>
            )}
          </div>

          <div className="code-paste-workspace__results-panel">
            {status === 'parsed' && functions.length > 0 ? (
              <>
                <FunctionList
                  functions={functions}
                  selectedIds={selectedIds}
                  onSelectionChange={toggleSelection}
                  onSelectAll={selectAll}
                  onClearSelection={clearSelection}
                  onFunctionPreview={handleFunctionPreview}
                  onFunctionNameChange={handleFunctionNameChange}
                />

                <ImportConfigPanel
                  selectedCount={selectedIds.size}
                  importConfig={importConfig}
                  onConfigChange={setImportConfig}
                  onImport={handleImport}
                  onCancel={handleCancel}
                  isImporting={importStatus === 'importing'}
                  importErrors={importErrors}
                />
              </>
            ) : (
              <CodePasteEmptyState />
            )}
          </div>
        </div>
      </section>

      <AlertDialog open={showDiscardDialog} onOpenChange={setShowDiscardDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard pasted code?</AlertDialogTitle>
            <AlertDialogDescription>
              You have unsaved pasted code or parsed functions. Leaving now will lose your progress.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep editing</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmDiscard}>Discard</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <FunctionPreviewPanel
        func={previewFunction}
        isOpen={showPreview}
        onClose={() => setShowPreview(false)}
      />
    </>
  );
}

/** @deprecated Use CodePasteWorkspace */
export const CodePasteModal = CodePasteWorkspace;

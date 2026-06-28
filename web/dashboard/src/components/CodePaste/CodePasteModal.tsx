import { useState, useCallback } from 'react';
import { Toaster, toast } from 'react-hot-toast';
import { CodeEditor } from './CodeEditor';
import { LanguageDetectorBadge } from './LanguageDetectorBadge';
import { FunctionList } from './FunctionList';
import { ImportConfigPanel } from './ImportConfigPanel';
import { FunctionPreviewPanel } from './FunctionPreviewPanel';
import { useCodeAnalyzer } from '../../hooks/useCodeAnalyzer';
import { useFunctionImport } from '../../hooks/useFunctionImport';
import { useCodePasteStore } from '../../stores/codePasteStore';
import type { ParsedFunction } from '../../types/codePaste';
import { X } from 'lucide-react';
import './CodePasteModal.css';

interface CodePasteModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImportComplete?: (functions: Array<{ id: string; name: string }>) => void;
  initialCode?: string;
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

export function CodePasteModal({
  isOpen,
  onClose,
  onImportComplete,
  initialCode = '',
}: CodePasteModalProps) {
  const [previewFunction, setPreviewFunction] = useState<ParsedFunction | null>(null);
  const [showPreview, setShowPreview] = useState(false);

  const { code, language, confidence, functions, status, error, setCode, setLanguage, parseCode, clearAnalysis } = useCodeAnalyzer();
  const { status: importStatus, createdFunctions, errors: importErrors, importFunctions, reset: resetImport } = useFunctionImport();
  const { selectedIds, importConfig, toggleSelection, selectAll, clearSelection, setImportConfig, setFunctions, reset: resetStore } = useCodePasteStore();

  const handleParse = useCallback(async () => { await parseCode(); }, [parseCode]);

  const handleClose = useCallback(() => {
    clearAnalysis();
    resetImport();
    resetStore();
    onClose();
  }, [clearAnalysis, resetImport, resetStore, onClose]);

  const handleImport = useCallback(async () => {
    const success = await importFunctions(functions, selectedIds, importConfig);
    if (success && createdFunctions.length > 0) {
      toast.success(`Successfully imported ${createdFunctions.length} function${createdFunctions.length !== 1 ? 's' : ''}!`);
      if (onImportComplete) onImportComplete(createdFunctions);
      setTimeout(handleClose, 1500);
    } else if (importErrors.length > 0) {
      toast.error(`Import failed: ${importErrors[0].error}`);
    }
  }, [functions, selectedIds, importConfig, importFunctions, createdFunctions, importErrors, onImportComplete, handleClose]);

  const handleFunctionPreview = useCallback((func: ParsedFunction) => { setPreviewFunction(func); setShowPreview(true); }, []);
  const handleFunctionNameChange = useCallback((id: string, name: string) => { setFunctions(functions.map((f) => f.id === id ? { ...f, name } : f)); }, [functions, setFunctions]);
  const handleExampleClick = useCallback((lang: keyof typeof EXAMPLE_CODE) => { setCode(EXAMPLE_CODE[lang]); }, [setCode]);

  if (!isOpen) return null;

  return (
    <div className="cpm">
      <Toaster position="top-right" toastOptions={{ duration: 4000, style: { background: 'var(--panel)', color: 'var(--text)', border: '1px solid var(--panel-edge)' } }} />

      <div className="cpm__content">
        {/* Header */}
        <div className="cpm__header">
          <div>
            <h1 className="cpm__title">Paste Your Code</h1>
            <p className="cpm__subtitle">Transform code snippets into deployable functions</p>
          </div>
          <button className="cpm__close" onClick={handleClose} aria-label="Close modal">
            <X className="cpm__close-icon" />
          </button>
        </div>

        {/* Body */}
        <div className="cpm__body">
          {/* Input Panel */}
          <div className="cpm__input-panel">
            <div className="cpm__editor-header">
              <LanguageDetectorBadge language={language} confidence={confidence} onLanguageChange={setLanguage} />
              <div className="cpm__examples">
                <span className="cpm__examples-label">Try:</span>
                {(['python', 'javascript', 'go'] as const).map((lang) => (
                  <button key={lang} className="cpm__example-btn" onClick={() => handleExampleClick(lang)}>
                    {lang === 'python' ? 'Python' : lang === 'javascript' ? 'JavaScript' : 'Go'}
                  </button>
                ))}
              </div>
            </div>

            <div className="cpm__editor">
              <CodeEditor value={code} onChange={setCode} language={language} height="350px" />
            </div>

            <div className="cpm__actions">
              <button className="cpm__clear-btn" onClick={clearAnalysis} disabled={status === 'parsing' || !code}>Clear</button>
              <button className="cpm__parse-btn" onClick={handleParse} disabled={status === 'parsing' || !code}>
                {status === 'parsing' ? <><span className="cpm__spinner" /> Analyzing...</> : <>Parse Code →</>}
              </button>
            </div>

            {error && <div className="cpm__error">{error}</div>}
          </div>

          {/* Results Panel */}
          <div className="cpm__results-panel">
            {status === 'parsed' && functions.length > 0 ? (
              <>
                <FunctionList functions={functions} selectedIds={selectedIds} onSelectionChange={toggleSelection}
                  onSelectAll={selectAll} onClearSelection={clearSelection} onFunctionPreview={handleFunctionPreview}
                  onFunctionNameChange={handleFunctionNameChange} />
                <ImportConfigPanel selectedCount={selectedIds.size} importConfig={importConfig}
                  onConfigChange={setImportConfig} onImport={handleImport} onCancel={handleClose}
                  isImporting={importStatus === 'importing'} />
              </>
            ) : (
              <div className="cpm__empty">
                <div className="cpm__empty-icon">🔍</div>
                <h3 className="cpm__empty-title">Paste code and click Parse</h3>
                <p className="cpm__empty-desc">We'll detect functions and help you import them as deployable functions.</p>
                <div className="cpm__features">
                  {[
                    { icon: '🔮', label: 'Auto-detect language' },
                    { icon: '📜', label: 'Extract function signatures' },
                    { icon: '✅', label: 'Choose what to import' },
                  ].map(({ icon, label }) => (
                    <div key={label} className="cpm__feature">
                      <span className="cpm__feature-icon">{icon}</span>
                      <span>{label}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      <FunctionPreviewPanel func={previewFunction} isOpen={showPreview} onClose={() => setShowPreview(false)} />
    </div>
  );
}

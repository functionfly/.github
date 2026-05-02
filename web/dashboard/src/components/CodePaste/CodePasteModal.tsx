import { useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
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

  const {
    status: importStatus,
    createdFunctions,
    errors: importErrors,
    importFunctions,
    reset: resetImport,
  } = useFunctionImport();

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

  const handleParse = useCallback(async () => {
    await parseCode();
  }, [parseCode]);

  const handleClose = useCallback(() => {
    clearAnalysis();
    resetImport();
    resetStore();
    onClose();
  }, [clearAnalysis, resetImport, resetStore, onClose]);

  const handleImport = useCallback(async () => {
    const success = await importFunctions(functions, selectedIds, importConfig);
    if (success && createdFunctions.length > 0) {
      toast.success(
        `Successfully imported ${createdFunctions.length} function${createdFunctions.length !== 1 ? 's' : ''}!`
      );
      if (onImportComplete) {
        onImportComplete(createdFunctions);
      }
      setTimeout(handleClose, 1500);
    } else if (importErrors.length > 0) {
      toast.error(`Import failed: ${importErrors[0].error}`);
    }
  }, [functions, selectedIds, importConfig, importFunctions, createdFunctions, importErrors, onImportComplete, handleClose]);

  const handleFunctionPreview = useCallback((func: ParsedFunction) => {
    setPreviewFunction(func);
    setShowPreview(true);
  }, []);

  const handleFunctionNameChange = useCallback((id: string, name: string) => {
    setFunctions(
      functions.map((f) =>
        f.id === id ? { ...f, name } : f
      )
    );
  }, [functions, setFunctions]);

  const handleExampleClick = useCallback((lang: keyof typeof EXAMPLE_CODE) => {
    setCode(EXAMPLE_CODE[lang]);
  }, [setCode]);

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          className="code-paste-modal"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
        >
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: '#333',
                color: '#fff',
              },
            }}
          />

          <motion.div
            className="code-paste-modal__content"
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
          >
            <div className="code-paste-modal__header">
              <div>
                <h1>Paste Your Code</h1>
                <p>Transform code snippets into deployable functions</p>
              </div>
              <button
                className="code-paste-modal__close"
                onClick={handleClose}
                aria-label="Close modal"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <div className="code-paste-modal__body">
              <div className="code-paste-modal__input-panel">
                <div className="code-paste-modal__editor-header">
                  <LanguageDetectorBadge
                    language={language}
                    confidence={confidence}
                    onLanguageChange={setLanguage}
                  />
                  <div className="code-paste-modal__examples">
                    <span>Try:</span>
                    <button onClick={() => handleExampleClick('python')}>Python</button>
                    <button onClick={() => handleExampleClick('javascript')}>JavaScript</button>
                    <button onClick={() => handleExampleClick('go')}>Go</button>
                  </div>
                </div>

                <div className="code-paste-modal__editor">
                  <CodeEditor
                    value={code}
                    onChange={setCode}
                    language={language}
                    height="350px"
                  />
                </div>

                <div className="code-paste-modal__actions">
                  <button
                    className="code-paste-modal__clear-btn"
                    onClick={clearAnalysis}
                    disabled={status === 'parsing' || !code}
                  >
                    Clear
                  </button>
                  <button
                    className="code-paste-modal__parse-btn"
                    onClick={handleParse}
                    disabled={status === 'parsing' || !code}
                  >
                    {status === 'parsing' ? (
                      <>
                        <span className="code-paste-modal__spinner" />
                        Analyzing...
                      </>
                    ) : (
                      <>Parse Code →</>
                    )}
                  </button>
                </div>

                {error && (
                  <div className="code-paste-modal__error">
                    {error}
                  </div>
                )}
              </div>

              <div className="code-paste-modal__results-panel">
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
                      onCancel={handleClose}
                      isImporting={importStatus === 'importing'}
                    />
                  </>
                ) : (
                  <div className="code-paste-modal__empty-state">
                    <div className="code-paste-modal__empty-icon">🔍</div>
                    <h3>Paste code and click Parse</h3>
                    <p>We'll detect functions and help you import them as deployable functions.</p>
                    <div className="code-paste-modal__features">
                      <div className="code-paste-modal__feature">
                        <span>🔮</span>
                        <span>Auto-detect language</span>
                      </div>
                      <div className="code-paste-modal__feature">
                        <span>📜</span>
                        <span>Extract function signatures</span>
                      </div>
                      <div className="code-paste-modal__feature">
                        <span>✅</span>
                        <span>Choose what to import</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </motion.div>

          <FunctionPreviewPanel
            func={previewFunction}
            isOpen={showPreview}
            onClose={() => setShowPreview(false)}
          />
        </motion.div>
      )}
    </AnimatePresence>
  );
}
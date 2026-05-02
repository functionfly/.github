import { useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { ParsedFunction } from '../../types/codePaste';
import './FunctionPreviewPanel.css';

interface FunctionPreviewPanelProps {
  func: ParsedFunction | null;
  isOpen: boolean;
  onClose: () => void;
}

const languageToMonaco: Record<string, string> = {
  python: 'python',
  javascript: 'javascript',
  typescript: 'typescript',
  go: 'go',
  rust: 'rust',
  ruby: 'ruby',
  java: 'java',
  kotlin: 'kotlin',
  swift: 'swift',
  cpp: 'cpp',
  c: 'c',
};

export function FunctionPreviewPanel({ func, isOpen, onClose }: FunctionPreviewPanelProps) {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  if (!func) return null;

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            className="function-preview-panel__backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.div
            className="function-preview-panel"
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 30, stiffness: 300 }}
          >
            <div className="function-preview-panel__header">
              <div>
                <h2>{func.name}</h2>
                <span className="function-preview-panel__language">
                  {func.language}
                </span>
              </div>
              <button
                className="function-preview-panel__close"
                onClick={onClose}
                aria-label="Close preview"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <div className="function-preview-panel__content">
              <div className="function-preview-panel__section">
                <h3>Signature</h3>
                <pre className="function-preview-panel__signature">
                  <code>{func.signature}</code>
                </pre>
              </div>

              {func.parameters.length > 0 && (
                <div className="function-preview-panel__section">
                  <h3>Parameters</h3>
                  <ul className="function-preview-panel__params">
                    {func.parameters.map((param, idx) => (
                      <li key={idx}>
                        <span className="param-name">{param.name}</span>
                        {param.type && (
                          <span className="param-type">: {param.type}</span>
                        )}
                        {param.has_default && (
                          <span className="param-default">
                            {' '}
                            = {param.default_value}
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {func.return_type && (
                <div className="function-preview-panel__section">
                  <h3>Return Type</h3>
                  <span className="function-preview-panel__return-type">
                    {func.return_type}
                  </span>
                </div>
              )}

              {func.docstring && (
                <div className="function-preview-panel__section">
                  <h3>Docstring</h3>
                  <p className="function-preview-panel__docstring">
                    {func.docstring}
                  </p>
                </div>
              )}

              <div className="function-preview-panel__section">
                <div className="function-preview-panel__code-header">
                  <h3>Code</h3>
                  <button
                    className="function-preview-panel__copy-btn"
                    onClick={() => navigator.clipboard.writeText(func.code)}
                  >
                    Copy Code
                  </button>
                </div>
                <pre className="function-preview-panel__code">
                  <code>{func.code}</code>
                </pre>
              </div>

              <div className="function-preview-panel__meta">
                <span>Lines {func.start_line}-{func.end_line}</span>
                <span>|</span>
                <span>{func.language}</span>
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
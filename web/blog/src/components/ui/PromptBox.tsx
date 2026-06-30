import { useState, useCallback, useRef, useEffect } from 'react';

interface PromptBoxProps {
  /** The prompt text content */
  children: string;
  /** Optional label displayed in the header (default: "PROMPT") */
  label?: string;
  /** Optional model name displayed in the header */
  model?: string;
  /** Optional context/system info displayed below the header */
  context?: string;
  /** Whether to show line numbers */
  lineNumbers?: boolean;
  /** Custom className */
  className?: string;
}

export default function PromptBox({
  children,
  label = 'PROMPT',
  model,
  context,
  lineNumbers = false,
  className = '',
}: PromptBoxProps) {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(children);
      setCopied(true);
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    } catch {
      const textarea = document.createElement('textarea');
      textarea.value = children;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopied(true);
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    }
  }, [children]);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  const lines = children.split('\n');
  const trimmedContent = children.replace(/\n+$/, '');

  return (
    <div className={`prompt-box ${className}`}>
      <div className="prompt-box__header">
        <div className="prompt-box__header-left">
          <span className="prompt-box__label">{label}</span>
          {model && <span className="prompt-box__model">{model}</span>}
        </div>
        <button
          className={`prompt-box__copy ${copied ? 'prompt-box__copy--copied' : ''}`}
          onClick={handleCopy}
          aria-label={copied ? 'Copied' : 'Copy prompt'}
          type="button"
        >
          {copied ? (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          ) : (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
            </svg>
          )}
          <span>{copied ? 'Copied' : 'Copy'}</span>
        </button>
      </div>

      {context && (
        <div className="prompt-box__context">
          <span className="prompt-box__context-label">SYSTEM</span>
          <span className="prompt-box__context-text">{context}</span>
        </div>
      )}

      <div className="prompt-box__body">
        {lineNumbers && (
          <div className="prompt-box__lines" aria-hidden="true">
            {lines.map((_, i) => (
              <span key={i} className="prompt-box__line-number">{i + 1}</span>
            ))}
          </div>
        )}
        <pre className="prompt-box__pre">
          <code className="prompt-box__code">{trimmedContent}</code>
        </pre>
      </div>

      <style>{`
        .prompt-box {
          position: relative;
          background: var(--panel);
          border: 1px solid var(--panel-edge);
          border-radius: var(--radius);
          overflow: hidden;
          margin: var(--space-5) 0;
          box-shadow: var(--shadow-chamber);
        }

        .prompt-box__header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: var(--space-2) var(--space-4);
          background: var(--panel-raised);
          border-bottom: 1px solid var(--panel-edge);
        }

        .prompt-box__header-left {
          display: flex;
          align-items: center;
          gap: var(--space-3);
        }

        .prompt-box__label {
          font-family: var(--font-mono);
          font-size: 11px;
          font-weight: 500;
          letter-spacing: 0.06em;
          text-transform: uppercase;
          color: var(--accent);
        }

        .prompt-box__model {
          font-family: var(--font-mono);
          font-size: 11px;
          font-weight: 500;
          letter-spacing: 0.06em;
          text-transform: uppercase;
          color: var(--text-faint);
          padding: var(--space-1) var(--space-2);
          background: var(--panel);
          border: 1px solid var(--panel-edge);
          border-radius: var(--radius-sm);
        }

        .prompt-box__copy {
          display: inline-flex;
          align-items: center;
          gap: var(--space-1);
          padding: var(--space-1) var(--space-2);
          background: transparent;
          border: 1px solid var(--steel);
          border-radius: var(--radius);
          color: var(--text-dim);
          font-family: var(--font-mono);
          font-size: 11px;
          font-weight: 500;
          letter-spacing: 0.04em;
          cursor: pointer;
          transition:
            border-color var(--duration-fast) var(--ease-out),
            color var(--duration-fast) var(--ease-out),
            background var(--duration-fast) var(--ease-out);
        }

        .prompt-box__copy:hover {
          border-color: var(--steel-light);
          color: var(--text);
          background: rgba(255, 255, 255, 0.04);
        }

        .prompt-box__copy:focus-visible {
          outline: none;
          box-shadow: var(--shadow-focus);
        }

        .prompt-box__copy--copied {
          color: var(--status-ok);
          border-color: rgba(143, 255, 208, 0.3);
        }

        .prompt-box__context {
          display: flex;
          align-items: baseline;
          gap: var(--space-3);
          padding: var(--space-2) var(--space-4);
          background: rgba(143, 255, 208, 0.03);
          border-bottom: 1px solid var(--panel-edge);
        }

        .prompt-box__context-label {
          font-family: var(--font-mono);
          font-size: 10px;
          font-weight: 500;
          letter-spacing: 0.08em;
          text-transform: uppercase;
          color: var(--status-ok);
          flex-shrink: 0;
        }

        .prompt-box__context-text {
          font-family: var(--font-body);
          font-size: 13px;
          color: var(--text-dim);
          line-height: 1.5;
        }

        .prompt-box__body {
          display: flex;
          overflow-x: auto;
        }

        .prompt-box__lines {
          flex-shrink: 0;
          padding: var(--space-4) var(--space-3);
          background: var(--panel-raised);
          border-right: 1px solid var(--panel-edge);
          text-align: right;
          user-select: none;
        }

        .prompt-box__line-number {
          display: block;
          font-family: var(--font-mono);
          font-size: 12px;
          line-height: 1.7;
          color: var(--text-faint);
        }

        .prompt-box__pre {
          margin: 0;
          padding: var(--space-4);
          background: transparent;
          overflow-x: auto;
          flex: 1;
        }

        .prompt-box__code {
          font-family: var(--font-mono);
          font-size: 13px;
          line-height: 1.7;
          color: var(--text);
          background: transparent;
          padding: 0;
          display: block;
          white-space: pre-wrap;
          word-break: break-word;
        }

        [data-theme="light"] .prompt-box {
          background: #ffffff;
          border-color: #e2e8f0;
          box-shadow: 0 0 0 1px #e2e8f0;
        }

        [data-theme="light"] .prompt-box__header {
          background: #f8fafc;
          border-color: #e2e8f0;
        }

        [data-theme="light"] .prompt-box__label {
          color: #e85a2a;
        }

        [data-theme="light"] .prompt-box__model {
          background: #ffffff;
          border-color: #e2e8f0;
          color: #64748b;
        }

        [data-theme="light"] .prompt-box__copy {
          border-color: #e2e8f0;
          color: #64748b;
        }

        [data-theme="light"] .prompt-box__copy:hover {
          border-color: #94a3b8;
          color: #0f172a;
          background: rgba(0, 0, 0, 0.02);
        }

        [data-theme="light"] .prompt-box__copy--copied {
          color: #059669;
          border-color: rgba(5, 150, 105, 0.3);
        }

        [data-theme="light"] .prompt-box__context {
          background: rgba(5, 150, 105, 0.03);
          border-color: #e2e8f0;
        }

        [data-theme="light"] .prompt-box__context-label {
          color: #059669;
        }

        [data-theme="light"] .prompt-box__context-text {
          color: #475569;
        }

        [data-theme="light"] .prompt-box__lines {
          background: #f8fafc;
          border-color: #e2e8f0;
        }

        [data-theme="light"] .prompt-box__line-number {
          color: #94a3b8;
        }

        [data-theme="light"] .prompt-box__code {
          color: #0f172a;
        }

        @media (max-width: 640px) {
          .prompt-box__header {
            padding: var(--space-2) var(--space-3);
          }

          .prompt-box__pre {
            padding: var(--space-3);
          }

          .prompt-box__lines {
            padding: var(--space-3) var(--space-2);
          }
        }
      `}</style>
    </div>
  );
}

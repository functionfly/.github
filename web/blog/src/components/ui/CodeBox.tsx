import { useState, useCallback, useRef, useEffect } from 'react';

interface CodeBoxProps {
  /** The code content */
  children: string;
  /** Programming language */
  language?: string;
  /** Optional filename displayed in the header */
  filename?: string;
  /** Whether to show line numbers (default: true) */
  lineNumbers?: boolean;
  /** Whether the code is an inline snippet (single line, no header) */
  inline?: boolean;
  /** Starting line number (default: 1) */
  startLine?: number;
  /** Highlight specific lines (1-indexed) */
  highlightLines?: number[];
  /** Custom className */
  className?: string;
}

const LANGUAGE_COLORS: Record<string, string> = {
  typescript: '#3178c6',
  ts: '#3178c6',
  javascript: '#f7df1e',
  js: '#f7df1e',
  tsx: '#3178c6',
  jsx: '#f7df1e',
  go: '#00add8',
  golang: '#00add8',
  python: '#3776ab',
  py: '#3776ab',
  rust: '#dea584',
  bash: '#4eaa25',
  sh: '#4eaa25',
  shell: '#4eaa25',
  json: '#f7df1e',
  yaml: '#cb171e',
  yml: '#cb171e',
  sql: '#f29111',
  html: '#e34c26',
  css: '#264de4',
  ruby: '#cc342d',
  rb: '#cc342d',
  java: '#b07219',
  kotlin: '#A97BFF',
  swift: '#F05138',
  php: '#4F5D95',
  c: '#555555',
  cpp: '#f34b7d',
  'c++': '#f34b7d',
  csharp: '#178600',
  'c#': '#178600',
  toml: '#9c4221',
  markdown: '#083fa1',
  md: '#083fa1',
  dockerfile: '#384d54',
  terraform: '#5C4EE5',
  hcl: '#5C4EE5',
};

function normalizeLanguage(lang?: string): string {
  if (!lang) return 'text';
  const lower = lang.toLowerCase();
  const aliases: Record<string, string> = {
    ts: 'typescript',
    js: 'javascript',
    tsx: 'typescript',
    jsx: 'javascript',
    py: 'python',
    rb: 'ruby',
    sh: 'bash',
    shell: 'bash',
    yml: 'yaml',
    golang: 'go',
    'c++': 'cpp',
    'c#': 'csharp',
    md: 'markdown',
  };
  return aliases[lower] || lower;
}

function getLanguageColor(lang: string): string | undefined {
  return LANGUAGE_COLORS[lang];
}

export default function CodeBox({
  children,
  language,
  filename,
  lineNumbers = true,
  inline = false,
  startLine = 1,
  highlightLines,
  className = '',
}: CodeBoxProps) {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const normalizedLang = normalizeLanguage(language);
  const langColor = getLanguageColor(normalizedLang);
  const code = children.replace(/\n+$/, '');

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    } catch {
      const textarea = document.createElement('textarea');
      textarea.value = code;
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
  }, [code]);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  if (inline) {
    return (
      <code className={`code-box-inline ${className}`}>
        {children}
        <style>{`
          .code-box-inline {
            font-family: var(--font-mono);
            font-size: 0.875em;
            padding: 0.15em 0.4em;
            background: var(--panel);
            border: 1px solid var(--panel-edge);
            border-radius: var(--radius-sm);
            color: var(--status-ok);
          }

          [data-theme="light"] .code-box-inline {
            background: #f1f5f9;
            border-color: #e2e8f0;
            color: #059669;
          }
        `}</style>
      </code>
    );
  }

  const lines = code.split('\n');
  const highlightSet = new Set(highlightLines?.map(l => l - 1) ?? []);

  return (
    <div className={`code-box ${className}`}>
      <div className="code-box__header">
        <div className="code-box__header-left">
          {filename && (
            <span className="code-box__filename">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
              </svg>
              {filename}
            </span>
          )}
          {normalizedLang !== 'text' && (
            <span
              className="code-box__lang"
              style={langColor ? { color: langColor } : undefined}
            >
              {normalizedLang}
            </span>
          )}
        </div>
        <button
          className={`code-box__copy ${copied ? 'code-box__copy--copied' : ''}`}
          onClick={handleCopy}
          aria-label={copied ? 'Copied' : 'Copy code'}
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

      <div className="code-box__body">
        {lineNumbers && (
          <div className="code-box__lines" aria-hidden="true">
            {lines.map((_, i) => (
              <span
                key={i}
                className={`code-box__line-number ${highlightSet.has(i) ? 'code-box__line-number--highlight' : ''}`}
              >
                {startLine + i}
              </span>
            ))}
          </div>
        )}
        <pre className="code-box__pre">
          <code className={`code-box__code language-${normalizedLang}`}>
            {lines.map((line, i) => (
              <span
                key={i}
                className={`code-box__line ${highlightSet.has(i) ? 'code-box__line--highlight' : ''}`}
              >
                {line}
                {i < lines.length - 1 ? '\n' : ''}
              </span>
            ))}
          </code>
        </pre>
      </div>

      <style>{`
        .code-box {
          position: relative;
          background: var(--panel);
          border: 1px solid var(--panel-edge);
          border-radius: var(--radius);
          overflow: hidden;
          margin: var(--space-5) 0;
          box-shadow: var(--shadow-chamber);
        }

        .code-box__header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: var(--space-2) var(--space-4);
          background: var(--panel-raised);
          border-bottom: 1px solid var(--panel-edge);
        }

        .code-box__header-left {
          display: flex;
          align-items: center;
          gap: var(--space-3);
        }

        .code-box__filename {
          display: inline-flex;
          align-items: center;
          gap: var(--space-1);
          font-family: var(--font-mono);
          font-size: 12px;
          font-weight: 500;
          color: var(--text-dim);
        }

        .code-box__lang {
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

        .code-box__copy {
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
            background var(--duration-fast) var(--ease-out),
            opacity var(--duration-fast) var(--ease-out);
          opacity: 0;
        }

        .code-box:hover .code-box__copy {
          opacity: 1;
        }

        .code-box__copy:hover {
          border-color: var(--steel-light);
          color: var(--text);
          background: rgba(255, 255, 255, 0.04);
        }

        .code-box__copy:focus-visible {
          outline: none;
          box-shadow: var(--shadow-focus);
          opacity: 1;
        }

        .code-box__copy--copied {
          color: var(--status-ok);
          border-color: rgba(143, 255, 208, 0.3);
          opacity: 1;
        }

        .code-box__body {
          display: flex;
          overflow-x: auto;
        }

        .code-box__lines {
          flex-shrink: 0;
          padding: var(--space-4) var(--space-3);
          background: var(--panel-raised);
          border-right: 1px solid var(--panel-edge);
          text-align: right;
          user-select: none;
        }

        .code-box__line-number {
          display: block;
          font-family: var(--font-mono);
          font-size: 12px;
          line-height: 1.7;
          color: var(--text-faint);
          transition: color var(--duration-fast) var(--ease-out);
        }

        .code-box__line-number--highlight {
          color: var(--status-ok);
        }

        .code-box__pre {
          margin: 0;
          padding: var(--space-4);
          background: transparent;
          overflow-x: auto;
          flex: 1;
        }

        .code-box__code {
          font-family: var(--font-mono);
          font-size: 13px;
          line-height: 1.7;
          color: var(--text-dim);
          background: transparent;
          padding: 0;
          display: block;
          white-space: pre;
          tab-size: 2;
        }

        .code-box__line {
          display: block;
          padding: 0 var(--space-1);
          margin: 0 calc(var(--space-1) * -1);
          border-radius: var(--radius-sm);
        }

        .code-box__line--highlight {
          background: rgba(143, 255, 208, 0.06);
          border-left: 2px solid var(--status-ok);
          padding-left: var(--space-2);
          margin-left: calc(var(--space-2) * -1 - 2px);
        }

        /* Syntax color tokens per spec 8.5 */
        .code-box__code .tok-keyword,
        .code-box__code .keyword { color: var(--foil-a); }
        .code-box__code .tok-string,
        .code-box__code .string { color: var(--status-ok); }
        .code-box__code .tok-comment,
        .code-box__code .comment { color: var(--text-faint); font-style: italic; }
        .code-box__code .tok-fn,
        .code-box__code .function { color: var(--foil-b); }
        .code-box__code .tok-num,
        .code-box__code .number { color: var(--foil-c); }

        /* Light mode */
        [data-theme="light"] .code-box {
          background: #ffffff;
          border-color: #e2e8f0;
          box-shadow: 0 0 0 1px #e2e8f0;
        }

        [data-theme="light"] .code-box__header {
          background: #f8fafc;
          border-color: #e2e8f0;
        }

        [data-theme="light"] .code-box__filename {
          color: #475569;
        }

        [data-theme="light"] .code-box__lang {
          color: #64748b;
          background: #ffffff;
          border-color: #e2e8f0;
        }

        [data-theme="light"] .code-box__copy {
          border-color: #e2e8f0;
          color: #64748b;
        }

        [data-theme="light"] .code-box__copy:hover {
          border-color: #94a3b8;
          color: #0f172a;
          background: rgba(0, 0, 0, 0.02);
        }

        [data-theme="light"] .code-box__copy--copied {
          color: #059669;
          border-color: rgba(5, 150, 105, 0.3);
        }

        [data-theme="light"] .code-box__lines {
          background: #f8fafc;
          border-color: #e2e8f0;
        }

        [data-theme="light"] .code-box__line-number {
          color: #94a3b8;
        }

        [data-theme="light"] .code-box__line-number--highlight {
          color: #059669;
        }

        [data-theme="light"] .code-box__code {
          color: #475569;
        }

        [data-theme="light"] .code-box__line--highlight {
          background: rgba(5, 150, 105, 0.06);
          border-left-color: #059669;
        }

        @media (max-width: 640px) {
          .code-box__header {
            padding: var(--space-2) var(--space-3);
          }

          .code-box__pre {
            padding: var(--space-3);
          }

          .code-box__lines {
            padding: var(--space-3) var(--space-2);
          }

          .code-box__copy {
            opacity: 1;
          }
        }
      `}</style>
    </div>
  );
}

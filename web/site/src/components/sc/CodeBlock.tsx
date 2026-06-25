import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface CodeBlockProps {
  code: string;
  language?: string;
  showLineNumbers?: boolean;
  className?: string;
}

const SYNTAX_COLORS = {
  keyword: 'var(--foil-a)',
  string: 'var(--status-ok)',
  comment: 'var(--text-faint)',
  function: 'var(--foil-b)',
  default: 'var(--text-dim)',
};

export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language = 'typescript',
  showLineNumbers = false,
  className = '',
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const [isCopied, setIsCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setIsCopied(true);
      setTimeout(() => setIsCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy code:', err);
    }
  };

  const lines = code.split('\n');

  return (
    <div
      className={`relative group ${className}`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <div
        style={{
          background: 'var(--panel)',
          border: '1px solid var(--panel-edge)',
          borderRadius: 'var(--radius)',
          overflow: 'hidden',
        }}
      >
        <div
          className="ff-font-mono flex items-center gap-[var(--space-2)] px-[var(--space-4)] py-[var(--space-3)]"
          style={{
            borderBottom: '1px solid var(--panel-edge)',
          }}
        >
          <span
            className="w-[12px] h-[12px] rounded-full"
            style={{ background: '#ef4444' }}
          />
          <span
            className="w-[12px] h-[12px] rounded-full"
            style={{ background: '#f59e0b' }}
          />
          <span
            className="w-[12px] h-[12px] rounded-full"
            style={{ background: '#10b981' }}
          />
          <span
            className="ml-[var(--space-2)] text-[11px] uppercase tracking-wider text-[var(--text-faint)]"
          >
            {language}
          </span>
        </div>

        <div className="relative" style={{ minWidth: 0 }}>
          <pre
            className="ff-font-mono overflow-x-auto"
            style={{
              fontSize: '12px',
              lineHeight: 1.7,
              color: SYNTAX_COLORS.default,
              margin: 0,
              padding: '20px 38px 20px var(--space-4)',
              whiteSpace: 'pre',
              wordBreak: 'normal',
              overflowWrap: 'normal',
            }}
          >
            {showLineNumbers && (
              <span
                className="select-none pr-[var(--space-4)] text-[var(--text-faint)]"
                style={{ userSelect: 'none' }}
              >
                {lines.map((_, i) => (
                  <span key={i} className="block">
                    {String(i + 1).padStart(2, ' ')}
                  </span>
                ))}
              </span>
            )}
            <code
              style={{
                whiteSpace: 'pre',
                wordBreak: 'normal',
                overflowWrap: 'normal',
              }}
            >
              {code}
            </code>
          </pre>

          <AnimatePresence>
            {isHovered && (
              <motion.button
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.12 }}
                type="button"
                onClick={handleCopy}
                className="absolute top-[10px] right-[10px] rounded-[var(--radius-sm)]"
                style={{
                  padding: '6px',
                  background: 'var(--panel-raised)',
                  border: '1px solid var(--steel)',
                  color: isCopied ? 'var(--status-ok)' : 'var(--text-dim)',
                  boxSizing: 'border-box',
                  zIndex: 2,
                }}
                aria-label={isCopied ? 'Copied!' : 'Copy code'}
              >
                {isCopied ? (
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : (
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                  </svg>
                )}
              </motion.button>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
};

interface InlineCodeProps {
  children: React.ReactNode;
  className?: string;
}

export const InlineCode: React.FC<InlineCodeProps> = ({
  children,
  className = '',
}) => {
  return (
    <code
      className={`ff-font-mono inline-block px-[6px] py-[2px] rounded-[var(--radius-sm)] text-[13px] ${className}`}
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--panel-edge)',
        color: 'var(--text-dim)',
      }}
    >
      {children}
    </code>
  );
};

import type { ReactNode } from 'react';
import { useState } from 'react';

interface CodeBlockProps {
  children: ReactNode;
  language?: string;
}

export function CodeBlock({ children, language = 'text' }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = typeof children === 'string' ? children : '';
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="code-block" data-language={language}>
      <pre className="code-block__pre">
        <code>{children}</code>
      </pre>
      <button
        type="button"
        className="code-block__copy"
        onClick={handleCopy}
        aria-label="Copy code"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
    </div>
  );
}

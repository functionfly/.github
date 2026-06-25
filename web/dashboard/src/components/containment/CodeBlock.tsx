import { useState, type ReactNode } from 'react';
import { Copy, Check } from 'lucide-react';

export interface CodeBlockProps {
  children: ReactNode;
  language?: string;
  className?: string;
}

export function CodeBlock({ children, language, className = '' }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = extractText(children);
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={`code-block ${className}`}>
      <div className="code-block__header">
        {language && <span className="code-block__language">{language}</span>}
        <button
          type="button"
          className="code-block__copy"
          onClick={handleCopy}
          aria-label={copied ? 'Copied' : 'Copy code'}
        >
          {copied ? <Check size={14} /> : <Copy size={14} />}
        </button>
      </div>
      <pre className="code-block__pre">
        <code className="code-block__code">{children}</code>
      </pre>
    </div>
  );
}

function extractText(children: ReactNode): string {
  if (typeof children === 'string') return children;
  if (Array.isArray(children)) return children.map(extractText).join('');
  if (children && typeof children === 'object' && 'props' in children) {
    return extractText((children as { props: { children?: ReactNode } }).props.children);
  }
  return '';
}

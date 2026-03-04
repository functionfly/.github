'use client';

import React, { useMemo, useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Check, Copy, Info, AlertTriangle, AlertCircle, Lightbulb } from 'lucide-react';
import Prism from 'prismjs';
import 'prismjs/components/prism-bash';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-typescript';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-jsx';
import 'prismjs/components/prism-tsx';
import 'prismjs/components/prism-go';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-sql';
import 'prismjs/components/prism-yaml';
import 'prismjs/components/prism-markdown';
import 'prismjs/components/prism-css';
import 'prismjs/components/prism-scss';
import 'prismjs/components/prism-rust';
import 'prismjs/components/prism-java';
import 'prismjs/components/prism-c';
import 'prismjs/components/prism-cpp';
import 'prismjs/components/prism-csharp';
import 'prismjs/components/prism-ruby';
import 'prismjs/components/prism-php';
import 'prismjs/components/prism-swift';
import 'prismjs/components/prism-kotlin';

// ============================================================================
// Types
// ============================================================================

export interface TextNode {
  text: string;
  bold?: boolean;
  italic?: boolean;
  code?: boolean;
  underline?: boolean;
  strikethrough?: boolean;
  link?: string;
  superscript?: boolean;
  subscript?: boolean;
  highlight?: boolean;
  color?: string;
}

export type BlockType =
  | 'paragraph'
  | 'heading'
  | 'heading-one'
  | 'heading-two'
  | 'heading-three'
  | 'heading-four'
  | 'heading-five'
  | 'heading-six'
  | 'block-quote'
  | 'bulleted-list'
  | 'numbered-list'
  | 'list-item'
  | 'code-block'
  | 'image'
  | 'divider'
  | 'table'
  | 'table-row'
  | 'table-cell'
  | 'callout'
  | 'details'
  | 'expandable'
  | 'video'
  | 'audio'
  | 'embed'
  | 'math';

export type CalloutType = 'info' | 'warning' | 'error' | 'success' | 'tip';

export interface ContentNode {
  type: BlockType;
  children: Array<TextNode | ContentNode>;
  level?: number; // for headings (1-6)
  language?: string; // for code blocks
  url?: string; // for images, videos, audio
  alt?: string; // for images
  align?: 'left' | 'center' | 'right' | 'justify';
  calloutType?: CalloutType; // for callouts
  title?: string; // for expandable sections
  caption?: string; // for media
  width?: number; // for media
  height?: number; // for media
  src?: string; // for embeds
  provider?: string; // for embeds (youtube, vimeo, etc.)
  expression?: string; // for math
}

export type RichTextContent = ContentNode[];

interface RichTextRendererProps {
  content: string | RichTextContent;
  className?: string;
}

interface CodeBlockProps {
  code: string;
  language?: string;
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Check if content looks like Slate.js JSON format
 */
function isSlateContent(content: string): boolean {
  const trimmed = content.trim();
  if (!trimmed.startsWith('[')) return false;

  try {
    const parsed = JSON.parse(trimmed);
    return Array.isArray(parsed) && parsed.length > 0 &&
           (parsed[0].type !== undefined || parsed[0].children !== undefined);
  } catch {
    return false;
  }
}

/**
 * Parse content string to RichTextContent
 */
function parseContent(content: string | RichTextContent): RichTextContent | null {
  if (typeof content !== 'string') {
    return Array.isArray(content) ? content : null;
  }

  if (!isSlateContent(content)) {
    return null;
  }

  try {
    const parsed = JSON.parse(content);
    return Array.isArray(parsed) ? parsed : [parsed];
  } catch {
    return null;
  }
}

/**
 * Check if a node is a text node
 */
function isTextNode(node: TextNode | ContentNode): node is TextNode {
  return 'text' in node && typeof node.text === 'string';
}

/**
 * Normalize language name for Prism
 */
function normalizeLanguage(lang?: string): string | undefined {
  if (!lang) return undefined;
  const normalized = lang.toLowerCase().trim();
  // Map common aliases
  const aliases: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'py': 'python',
    'sh': 'bash',
    'shell': 'bash',
    'yml': 'yaml',
    'md': 'markdown',
    'jsx': 'jsx',
    'tsx': 'tsx',
  };
  return aliases[normalized] || normalized;
}

// ============================================================================
// Code Block with Copy Button & Syntax Highlighting
// ============================================================================

function CodeBlock({ code, language }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const normalizedLang = normalizeLanguage(language);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Ignore copy errors
    }
  }, [code]);

  // Highlight the code
  const highlightedCode = useMemo(() => {
    if (!normalizedLang || !Prism.languages[normalizedLang]) {
      return code;
    }
    try {
      return Prism.highlight(code, Prism.languages[normalizedLang], normalizedLang);
    } catch {
      return code;
    }
  }, [code, normalizedLang]);

  return (
    <div className="relative group my-6 rounded-xl overflow-hidden bg-muted/80 border border-border/50">
      {/* Header with language and copy button */}
      <div className="flex items-center justify-between px-4 py-2 bg-muted border-b border-border/50">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {normalizedLang || 'text'}
        </span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-2 py-1 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted-foreground/10"
          aria-label={copied ? 'Copied!' : 'Copy code'}
        >
          {copied ? (
            <>
              <Check className="h-3.5 w-3.5 text-green-500" />
              <span className="text-green-500">Copied!</span>
            </>
          ) : (
            <>
              <Copy className="h-3.5 w-3.5" />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>

      {/* Code content */}
      <pre className="p-4 overflow-x-auto">
        <code
          className={cn('font-mono text-sm text-foreground/90', normalizedLang && `language-${normalizedLang}`)}
          dangerouslySetInnerHTML={{ __html: highlightedCode }}
        />
      </pre>
    </div>
  );
}

// ============================================================================
// Callout Component
// ============================================================================

const calloutStyles: Record<CalloutType, { icon: React.ReactNode; classes: string }> = {
  info: {
    icon: <Info className="h-5 w-5" />,
    classes: 'bg-blue-500/10 border-blue-500/30 text-blue-900 dark:text-blue-100',
  },
  warning: {
    icon: <AlertTriangle className="h-5 w-5" />,
    classes: 'bg-amber-500/10 border-amber-500/30 text-amber-900 dark:text-amber-100',
  },
  error: {
    icon: <AlertCircle className="h-5 w-5" />,
    classes: 'bg-red-500/10 border-red-500/30 text-red-900 dark:text-red-100',
  },
  success: {
    icon: <Check className="h-5 w-5" />,
    classes: 'bg-green-500/10 border-green-500/30 text-green-900 dark:text-green-100',
  },
  tip: {
    icon: <Lightbulb className="h-5 w-5" />,
    classes: 'bg-purple-500/10 border-purple-500/30 text-purple-900 dark:text-purple-100',
  },
};

function Callout({ type = 'info', children }: { type?: CalloutType; children: React.ReactNode }) {
  const { icon, classes } = calloutStyles[type];

  return (
    <div className={cn('my-6 p-4 rounded-xl border-l-4', classes)}>
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-0.5 opacity-70">{icon}</div>
        <div className="flex-1">{children}</div>
      </div>
    </div>
  );
}

// ============================================================================
// Expandable/Details Component
// ============================================================================

function Expandable({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <details className="my-6 rounded-xl border border-border/50 bg-muted/30 overflow-hidden group">
      {title && (
        <summary className="px-4 py-3 font-medium text-foreground cursor-pointer hover:bg-muted/50 transition-colors list-none flex items-center justify-between">
          <span>{title}</span>
          <span className="text-muted-foreground transform transition-transform group-open:rotate-180">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </span>
        </summary>
      )}
      <div className="px-4 py-3 border-t border-border/50">{children}</div>
    </details>
  );
}

// ============================================================================
// Leaf Renderer (inline text formatting)
// ============================================================================

interface LeafProps {
  leaf: TextNode;
  children: React.ReactNode;
}

function Leaf({ leaf, children }: LeafProps) {
  let content: React.ReactNode = children;

  // Apply inline formatting (order matters for nesting)
  if (leaf.color) {
    content = <span style={{ color: leaf.color }}>{content}</span>;
  }

  if (leaf.highlight) {
    content = (
      <mark className="bg-yellow-200 dark:bg-yellow-900/40 px-1 rounded text-foreground">
        {content}
      </mark>
    );
  }

  if (leaf.code) {
    content = (
      <code className="px-1.5 py-0.5 rounded bg-muted font-mono text-sm text-foreground/90 border border-border/50">
        {content}
      </code>
    );
  }

  if (leaf.superscript) {
    content = <sup className="text-xs">{content}</sup>;
  }

  if (leaf.subscript) {
    content = <sub className="text-xs">{content}</sub>;
  }

  if (leaf.bold) {
    content = <strong className="font-semibold text-foreground">{content}</strong>;
  }

  if (leaf.italic) {
    content = <em className="italic">{content}</em>;
  }

  if (leaf.underline) {
    content = <u className="underline underline-offset-2">{content}</u>;
  }

  if (leaf.strikethrough) {
    content = <s className="line-through opacity-70">{content}</s>;
  }

  if (leaf.link) {
    content = (
      <a
        href={leaf.link}
        target="_blank"
        rel="noopener noreferrer"
        className="text-brand-600 dark:text-brand-400 hover:underline underline-offset-2 font-medium"
      >
        {content}
      </a>
    );
  }

  return <>{content}</>;
}

// ============================================================================
// Text Renderer (handles nested text with marks)
// ============================================================================

interface TextRendererProps {
  nodes: Array<TextNode | ContentNode>;
}

function TextRenderer({ nodes }: TextRendererProps) {
  return (
    <>
      {nodes.map((node, index) => {
        if (isTextNode(node)) {
          // Handle empty text nodes
          if (!node.text && index === nodes.length - 1) {
            return null;
          }
          return (
            <Leaf key={index} leaf={node}>
              {node.text || ''}
            </Leaf>
          );
        }
        // Nested content node
        return <Element key={index} node={node} />;
      })}
    </>
  );
}

// ============================================================================
// Element Renderer (block-level elements)
// ============================================================================

interface ElementProps {
  node: ContentNode;
}

function Element({ node }: ElementProps) {
  const { type, children, align } = node;

  // Alignment classes
  const alignClass = align ? {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
    justify: 'text-justify',
  }[align] : '';

  switch (type) {
    case 'paragraph':
      return (
        <p className={cn('mb-4 leading-[1.75] text-foreground/90 last:mb-0', alignClass)}>
          <TextRenderer nodes={children} />
        </p>
      );

    case 'heading':
    case 'heading-one':
      return (
        <h1 className={cn('text-3xl sm:text-4xl font-bold tracking-tight mt-8 mb-4 text-foreground first:mt-0', alignClass)}>
          <TextRenderer nodes={children} />
        </h1>
      );

    case 'heading-two':
      return (
        <h2 className={cn('text-2xl sm:text-3xl font-semibold tracking-tight mt-8 mb-4 text-foreground', alignClass)}>
          <TextRenderer nodes={children} />
        </h2>
      );

    case 'heading-three':
      return (
        <h3 className={cn('text-xl sm:text-2xl font-semibold tracking-tight mt-6 mb-3 text-foreground', alignClass)}>
          <TextRenderer nodes={children} />
        </h3>
      );

    case 'heading-four':
      return (
        <h4 className={cn('text-lg sm:text-xl font-semibold tracking-tight mt-6 mb-3 text-foreground', alignClass)}>
          <TextRenderer nodes={children} />
        </h4>
      );

    case 'heading-five':
      return (
        <h5 className={cn('text-base sm:text-lg font-semibold mt-4 mb-2 text-foreground', alignClass)}>
          <TextRenderer nodes={children} />
        </h5>
      );

    case 'heading-six':
      return (
        <h6 className={cn('text-sm sm:text-base font-semibold mt-4 mb-2 text-foreground', alignClass)}>
          <TextRenderer nodes={children} />
        </h6>
      );

    case 'block-quote':
      return (
        <blockquote className="my-6 pl-4 border-l-4 border-brand-500/50 bg-muted/30 py-3 pr-4 rounded-r-lg">
          <div className="text-foreground/80 italic leading-relaxed">
            <TextRenderer nodes={children} />
          </div>
        </blockquote>
      );

    case 'bulleted-list':
      return (
        <ul className="my-4 ml-6 list-disc space-y-1">
          <TextRenderer nodes={children} />
        </ul>
      );

    case 'numbered-list':
      return (
        <ol className="my-4 ml-6 list-decimal space-y-1">
          <TextRenderer nodes={children} />
        </ol>
      );

    case 'list-item':
      return (
        <li className="leading-relaxed text-foreground/90">
          <TextRenderer nodes={children} />
        </li>
      );

    case 'code-block': {
      // Extract plain text from code block children
      const codeText = children
        .map(child => (isTextNode(child) ? child.text : ''))
        .join('');
      return <CodeBlock code={codeText} language={node.language} />;
    }

    case 'image':
      if (!node.url) return null;
      return (
        <figure className={cn('my-6', alignClass)}>
          <img
            src={node.url}
            alt={node.alt || ''}
            loading="lazy"
            className="rounded-xl max-w-full h-auto shadow-lg shadow-black/5 dark:shadow-none"
          />
          {(node.caption || node.alt) && (
            <figcaption className="mt-2 text-sm text-center text-muted-foreground italic">
              {node.caption || node.alt}
            </figcaption>
          )}
        </figure>
      );

    case 'video':
      if (!node.url) return null;
      return (
        <figure className={cn('my-6', alignClass)}>
          <video
            src={node.url}
            controls
            className="rounded-xl max-w-full shadow-lg shadow-black/5 dark:shadow-none"
            style={{ width: node.width, height: node.height }}
          />
          {node.caption && (
            <figcaption className="mt-2 text-sm text-center text-muted-foreground italic">
              {node.caption}
            </figcaption>
          )}
        </figure>
      );

    case 'audio':
      if (!node.url) return null;
      return (
        <div className="my-6">
          <audio
            src={node.url}
            controls
            className="w-full rounded-lg"
          />
        </div>
      );

    case 'embed':
      if (!node.src) return null;
      return (
        <div className={cn('my-6', alignClass)}>
          <div className="relative w-full aspect-video rounded-xl overflow-hidden bg-muted">
            <iframe
              src={node.src}
              title={node.title || 'Embedded content'}
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
              className="absolute inset-0 w-full h-full"
            />
          </div>
          {node.caption && (
            <p className="mt-2 text-sm text-center text-muted-foreground italic">
              {node.caption}
            </p>
          )}
        </div>
      );

    case 'divider':
      return <hr className="my-8 border-border/50" />;

    case 'table':
      return (
        <div className="my-6 overflow-x-auto">
          <table className="w-full border-collapse text-sm">
            <TextRenderer nodes={children} />
          </table>
        </div>
      );

    case 'table-row':
      return (
        <tr className="border-b border-border/50 last:border-b-0">
          <TextRenderer nodes={children} />
        </tr>
      );

    case 'table-cell':
      return (
        <td className="px-4 py-3 text-foreground/90">
          <TextRenderer nodes={children} />
        </td>
      );

    case 'callout':
      return (
        <Callout type={node.calloutType || 'info'}>
          <TextRenderer nodes={children} />
        </Callout>
      );

    case 'details':
    case 'expandable':
      return (
        <Expandable title={node.title}>
          <TextRenderer nodes={children} />
        </Expandable>
      );

    case 'math':
      return (
        <div className="my-6 overflow-x-auto">
          <code className="block p-4 rounded-xl bg-muted font-mono text-center">
            {node.expression || <TextRenderer nodes={children} />}
          </code>
        </div>
      );

    default:
      // Fallback for unknown types - render as paragraph
      return (
        <p className="mb-4 leading-[1.75] text-foreground/90 last:mb-0">
          <TextRenderer nodes={children} />
        </p>
      );
  }
}

// ============================================================================
// Main Component
// ============================================================================

export function RichTextRenderer({ content, className }: RichTextRendererProps) {
  const parsedContent = useMemo(() => parseContent(content), [content]);

  // Handle empty or invalid content
  if (!parsedContent || parsedContent.length === 0) {
    return (
      <div className={cn('text-muted-foreground italic', className)}>
        No content available.
      </div>
    );
  }

  return (
    <div className={className}>
      {parsedContent.map((node, index) => (
        <Element key={index} node={node} />
      ))}
    </div>
  );
}

// ============================================================================
// Export utility for checking content type
// ============================================================================

export { isSlateContent, parseContent };
export default RichTextRenderer;

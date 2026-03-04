'use client';

import { useEffect, useRef, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize from 'rehype-sanitize';
import DOMPurify from 'dompurify';
import Prism from 'prismjs';
import type { Components } from 'react-markdown';
import { RichTextRenderer, isSlateContent } from '@/components/blog';
import 'prismjs/components/prism-bash';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-typescript';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-go';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-sql';
import 'prismjs/components/prism-yaml';

const ALLOWED_HTML_TAGS = [
  'p', 'br', 'strong', 'em', 'u', 's', 'a', 'ul', 'ol', 'li', 'blockquote',
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'pre', 'code', 'img', 'hr', 'div', 'span',
  'table', 'thead', 'tbody', 'tr', 'th', 'td',
];

function looksLikeHtml(content: string): boolean {
  const trimmed = content.trim();
  return trimmed.startsWith('<') && (trimmed.includes('</') || trimmed.includes('/>'));
}

interface BlogPostBodyProps {
  html: string;
  className?: string;
}

export function BlogPostBody({ html, className = '' }: BlogPostBodyProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  // Determine content type
  const contentType = useMemo(() => {
    if (isSlateContent(html)) return 'slate';
    if (looksLikeHtml(html)) return 'html';
    return 'markdown';
  }, [html]);

  const isMarkdown = contentType === 'markdown';
  const isSlate = contentType === 'slate';

  const sanitizedHtml = useMemo(() => {
    if (contentType === 'html') {
      return DOMPurify.sanitize(html, {
        ALLOWED_TAGS: ALLOWED_HTML_TAGS,
        ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'class', 'id'],
        ADD_ATTR: ['target'],
      });
    }
    return '';
  }, [html, contentType]);

  const markdownContent = isMarkdown ? html : '';

  useEffect(() => {
    if (!containerRef.current || isSlate) return;
    Prism.highlightAllUnder(containerRef.current);
  }, [isMarkdown, isSlate, sanitizedHtml, markdownContent]);

  const markdownComponents: Components = useMemo(
    () => ({
      code({ node, className: codeClassName, children, ...props }) {
        const isBlock =
          (codeClassName?.startsWith('language-') ?? false) ||
          (typeof children === 'string' && children.includes('\n'));
        if (isBlock) {
          return (
            <pre className={codeClassName ?? ''}>
              <code className={codeClassName ?? ''} {...props}>
                {children}
              </code>
            </pre>
          );
        }
        return (
          <code className={codeClassName ?? ''} {...props}>
            {children}
          </code>
        );
      },
    }),
    []
  );

  // Handle Slate.js rich text format
  if (isSlate) {
    return <RichTextRenderer content={html} className={className} />;
  }

  if (isMarkdown) {
    return (
      <div ref={containerRef} className={className}>
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeRaw, rehypeSanitize]}
          components={markdownComponents}
        >
          {markdownContent}
        </ReactMarkdown>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={className}
      dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
    />
  );
}

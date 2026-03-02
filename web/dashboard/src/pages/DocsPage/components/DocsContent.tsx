import { useEffect, useRef } from "react";
import { Check, Copy, ExternalLink, AlertTriangle, Info, CheckCircle, XCircle } from "lucide-react";
import { DOCS_SITE_URL } from "@/lib/constants";
import { type DocPage } from "../data/docs";

interface DocsContentProps {
  page: DocPage;
}

// Simple markdown parser for documentation content
function parseMarkdown(content: string): string {
  let html = content;

  // Escape HTML entities
  html = html
    .replace(/&/g, "&")
    .replace(/</g, "<")
    .replace(/>/g, ">");

  // Headers - add IDs for anchor links
  html = html.replace(/^### (.+)$/gm, (_, title) => {
    const id = title
      .toLowerCase()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-");
    return `<h3 id="${id}" class="docs-h3">${title}</h3>`;
  });

  html = html.replace(/^## (.+)$/gm, (_, title) => {
    const id = title
      .toLowerCase()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-");
    return `<h2 id="${id}" class="docs-h2">${title}</h2>`;
  });

  html = html.replace(/^# (.+)$/gm, (_, title) => {
    const id = title
      .toLowerCase()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-");
    return `<h1 id="${id}" class="docs-h1">${title}</h1>`;
  });

  // Bold and italic
  html = html.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Code blocks
  html = html.replace(
    /```(\w+)?\n([\s\S]*?)```/g,
    (_, lang, code) => `
      <div class="docs-code-block">
        <div class="docs-code-header">
          <span class="docs-code-lang">${lang || 'text'}</span>
          <button class="docs-copy-btn" data-code="${encodeURIComponent(code)}">
            <svg class="copy-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
            <svg class="check-icon hidden" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
            <span>Copy</span>
          </button>
        </div>
        <pre class="docs-pre"><code class="docs-code ${lang ? `language-${lang}` : ''}">${code.trim()}</code></pre>
      </div>
    `
  );

  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code class="docs-inline-code">$1</code>');

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="docs-link">$1</a>');

  // Tables
  html = html.replace(
    /\|(.+)\|\n\|[-:\| ]+\|\n((?:\|.+\|\n?)+)/g,
    (_, header, rows) => {
      const headers = header.split('|').map((h: string) => h.trim()).filter(Boolean);
      const rowLines = rows.trim().split('\n');

      let tableHtml = '<div class="docs-table-wrapper"><table class="docs-table"><thead><tr>';
      headers.forEach((h: string) => {
        tableHtml += `<th>${h}</th>`;
      });
      tableHtml += '</tr></thead><tbody>';

      rowLines.forEach((line: string) => {
        const cells = line.split('|').map((c: string) => c.trim()).filter(Boolean);
        tableHtml += '<tr>';
        cells.forEach((cell: string) => {
          tableHtml += `<td>${cell}</td>`;
        });
        tableHtml += '</tr>';
      });

      tableHtml += '</tbody></table></div>';
      return tableHtml;
    }
  );

  // Blockquotes
  html = html.replace(
    /^> (.+)$/gm,
    '<blockquote class="docs-blockquote">$1</blockquote>'
  );

  // Lists
  html = html.replace(/(^|\n)((?:- .+\n?)+)/g, (_, prefix, list) => {
    const items = list.trim().split('\n').map((line: string) => {
      const content = line.replace(/^- /, '');
      return `<li>${content}</li>`;
    }).join('');
    return `${prefix}<ul class="docs-ul">${items}</ul>`;
  });

  // Numbered lists
  html = html.replace(/(^|\n)((?:\d+\. .+\n?)+)/g, (_, prefix, list) => {
    const items = list.trim().split('\n').map((line: string) => {
      const content = line.replace(/^\d+\. /, '');
      return `<li>${content}</li>`;
    }).join('');
    return `${prefix}<ol class="docs-ol">${items}</ol>`;
  });

  // Paragraphs (must be last)
  html = html.replace(/\n\n([^<\n].*?)\n\n/g, '\n\n<p class="docs-p">$1</p>\n\n');
  html = html.replace(/\n\n([^<\n].*?)$/g, '\n\n<p class="docs-p">$1</p>');

  // Horizontal rules
  html = html.replace(/\n---\n/g, '\n<hr class="docs-hr" />\n');

  return html;
}

export function DocsContent({ page }: DocsContentProps) {
  const contentRef = useRef<HTMLDivElement>(null);

  // Handle copy button clicks
  useEffect(() => {
    const content = contentRef.current;
    if (!content) return;

    const handleClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      const copyBtn = target.closest('.docs-copy-btn') as HTMLButtonElement;

      if (copyBtn) {
        const code = decodeURIComponent(copyBtn.dataset.code || '');
        navigator.clipboard.writeText(code).then(() => {
          const copyIcon = copyBtn.querySelector('.copy-icon');
          const checkIcon = copyBtn.querySelector('.check-icon');
          const span = copyBtn.querySelector('span');

          copyIcon?.classList.add('hidden');
          checkIcon?.classList.remove('hidden');
          if (span) span.textContent = 'Copied!';

          setTimeout(() => {
            copyIcon?.classList.remove('hidden');
            checkIcon?.classList.add('hidden');
            if (span) span.textContent = 'Copy';
          }, 2000);
        });
      }
    };

    content.addEventListener('click', handleClick);
    return () => content.removeEventListener('click', handleClick);
  }, [page.content]);

  const htmlContent = parseMarkdown(page.content);

  return (
    <article className="docs-article">
      {/* Page Header */}
      <header className="mb-8 pb-8 border-b border-border-subtle">
        <h1 className="docs-title">{page.title}</h1>
        <p className="docs-description">{page.description}</p>
        {page.lastUpdated && (
          <div className="mt-4 flex items-center gap-2 text-sm text-text-muted">
            <span>Last updated:</span>
            <time dateTime={page.lastUpdated}>
              {new Date(page.lastUpdated).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}
            </time>
          </div>
        )}
      </header>

      {/* Content */}
      <div
        ref={contentRef}
        className="docs-content"
        dangerouslySetInnerHTML={{ __html: htmlContent }}
      />

      {/* Page Navigation */}
      <footer className="mt-12 pt-8 border-t border-border-subtle">
        <div className="flex items-center justify-between">
          <a
            href={DOCS_SITE_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-text-secondary hover:text-text-primary transition-colors"
          >
            <ExternalLink className="w-4 h-4" />
            <span>View all docs</span>
          </a>
          <a
            href="/contact"
            className="flex items-center gap-2 text-brand-500 hover:text-brand-400 transition-colors"
          >
            <span>Need help?</span>
            <ExternalLink className="w-4 h-4" />
          </a>
        </div>
      </footer>
    </article>
  );
}

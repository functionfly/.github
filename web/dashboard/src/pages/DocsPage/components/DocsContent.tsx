import { useEffect, useRef } from "react";
import { ExternalLink } from "lucide-react";
import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import { DOCS_SITE_URL } from "@/lib/constants";
import { type DocPage } from "../data/docs";

interface DocsContentProps {
  page: DocPage;
}

export function DocsContent({ page }: DocsContentProps) {
  const contentRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const content = contentRef.current;
    if (!content) return;

    const handleClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      const copyBtn = target.closest('.docs-copy-btn') as HTMLButtonElement;

      if (copyBtn) {
        const codeEl = copyBtn.closest('.docs-code-block')?.querySelector('code');
        const code = codeEl?.textContent || '';
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

  return (
    <article className="docs-article">
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

      <div ref={contentRef} className="docs-content">
        <ReactMarkdown
          rehypePlugins={[rehypeSanitize]}
          remarkPlugins={[remarkGfm]}
          components={{
            pre: ({ children }) => (
              <div className="docs-code-block">
                <div className="docs-code-header">
                  <span className="docs-code-lang">code</span>
                  <button className="docs-copy-btn" type="button">
                    <svg className="copy-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                    <svg className="check-icon hidden" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    <span>Copy</span>
                  </button>
                </div>
                <pre className="docs-pre"><code className="docs-code">{children}</code></pre>
              </div>
            ),
            a: ({ href, children }) => (
              <a href={href} className="docs-link" target="_blank" rel="noopener noreferrer">{children}</a>
            ),
          }}
        >
          {page.content}
        </ReactMarkdown>
      </div>

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

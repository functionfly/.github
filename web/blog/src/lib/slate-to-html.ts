/**
 * Enhanced Slate/TipTap/Markdown to HTML converter with syntax highlighting support
 * Handles paragraphs, headings, inline marks, code blocks with language detection,
 * and Markdown text from Sanity CMS.
 */

import { marked } from "marked";

// Configure marked with custom renderer for CodeBox-style code blocks
const renderer = new marked.Renderer();
renderer.code = function({ text, lang }: { text: string; lang?: string }) {
  const normalizedLang = languageAliases[lang || ''] || lang || 'text';
  const lines = text.split('\n');
  const langColor = getLanguageColor(normalizedLang);
  const lineNumbers = lines
    .map((_: string, i: number) => `<span class="code-box__line-number">${i + 1}</span>`)
    .join('\n');
  const escapedCode = escapeHtml(text);

  return `
    <div class="code-box" data-language="${normalizedLang}">
      <div class="code-box__header">
        <div class="code-box__header-left">
          <span class="code-box__lang"${langColor ? ` style="color:${langColor}"` : ''}>${normalizedLang}</span>
        </div>
        <button class="code-box__copy" data-code="${escapeHtml(text)}" type="button">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
          </svg>
          <span>Copy</span>
        </button>
      </div>
      <div class="code-box__body">
        <div class="code-box__lines" aria-hidden="true">
          ${lineNumbers}
        </div>
        <pre class="code-box__pre"><code class="code-box__code language-${normalizedLang}">${escapedCode}</code></pre>
      </div>
    </div>
  `;
};

marked.use({ renderer });

/**
 * Post-process HTML to fix CLI command references.
 * The ff-cli binary is `ff`, not `fly`.
 * Matches: fly init, fly deploy, fly auth, etc. in visible content AND data-code attributes.
 */
function fixCliCommands(html: string): string {
  // Replace standalone `fly` CLI commands (preceded by whitespace/start/>/$, followed by known subcommand)
  const flyPattern = /(^|[\s>$])fly(\s+(?:init|deploy|auth|logs|status|run|publish|test|config|whoami|version|--help|--version|secret|env|domains|scale|restart|ssh|proxy|dashboard|secrets|tokens|org|apps|regions|machines|volumes|certs|wireguard|turboku|launch|doctor|image|builder|pg|redis|consul|ip|check|releases|rollback|suspend|resume|move|destroy|open))(?=[\s<"'\n])/gm;

  // Also fix inside data-code attributes (HTML-encoded quotes)
  return html
    .replace(flyPattern, '$1ff$2')
    .replace(/data-code="([^"]*?)fly(\s+(?:init|deploy|auth|logs|status|run|publish|test|config|whoami|version|--help|--version|secret|env|domains|scale|restart|ssh|proxy|dashboard|secrets|tokens|org|apps|regions|machines|volumes|certs|wireguard|turboku|launch|doctor|image|builder|pg|redis|consul|ip|check|releases|rollback|suspend|resume|move|destroy|open))/g, (match, prefix, cmd) => {
      return `data-code="${prefix}ff${cmd}`;
    });
}

// Language aliases for syntax highlighting
const languageAliases: Record<string, string> = {
  'ts': 'typescript',
  'js': 'javascript',
  'tsx': 'typescript',
  'jsx': 'javascript',
  'py': 'python',
  'rb': 'ruby',
  'sh': 'bash',
  'shell': 'bash',
  'yml': 'yaml',
  'golang': 'go',
};

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

interface TextLeaf {
  text?: string;
  bold?: boolean;
  italic?: boolean;
  code?: boolean;
  underline?: boolean;
  strikethrough?: boolean;
  link?: string;
  // TipTap marks format (optional - for TipTap JSON)
  marks?: Array<{ type: string; attrs?: Record<string, string> }>;
}

function renderLeaf(leaf: TextLeaf): string {
  let t = escapeHtml(leaf.text ?? '');
  if (leaf.code) t = `<code class="inline-code">${t}</code>`;
  if (leaf.bold) t = `<strong>${t}</strong>`;
  if (leaf.italic) t = `<em>${t}</em>`;
  if (leaf.underline) t = `<u>${t}</u>`;
  if (leaf.strikethrough) t = `<s>${t}</s>`;
  if (leaf.link) {
    const href = escapeHtml(leaf.link);
    t = `<a href="${href}" class="content-link" target="_blank" rel="noopener noreferrer">${t}</a>`;
  }
  return t;
}

function paragraphInlineHtml(children: unknown): string {
  if (!Array.isArray(children)) return '';
  return children
    .map((child) => {
      if (child && typeof child === 'object' && 'text' in child) {
        return renderLeaf(child as TextLeaf);
      }
      return '';
    })
    .join('');
}

function rawParagraphText(children: unknown): string {
  if (!Array.isArray(children)) return '';
  return children
    .map((child) => {
      if (child && typeof child === 'object' && 'text' in child) {
        return String((child as { text?: string }).text ?? '');
      }
      return '';
    })
    .join('');
}

/**
 * Detect language from code block text (supports ```language syntax)
 */
function detectLanguage(text: string): { lang: string; cleanCode: string } {
  const firstLine = text.split('\n')[0]?.trim() || '';
  const match = firstLine.match(/^```(\w+)$/);
  
  if (match) {
    const rawLang = match[1].toLowerCase();
    const lang = languageAliases[rawLang] || rawLang;
    const cleanCode = text.split('\n').slice(1).join('\n').replace(/```$/, '').trim();
    return { lang, cleanCode };
  }
  
  // Try to detect from content
  if (text.includes('package main') && text.includes('import')) {
    return { lang: 'go', cleanCode: text };
  }
  if (text.includes('import React') || text.includes('export default')) {
    return { lang: 'typescript', cleanCode: text };
  }
  if (text.includes('def ') || text.includes('import ') && !text.includes(';')) {
    return { lang: 'python', cleanCode: text };
  }
  
  return { lang: 'text', cleanCode: text };
}

/**
 * Create enhanced code block HTML with copy button and line numbers
 * Uses the CodeBox component class structure
 */
function createCodeBlockHtml(code: string, lang: string): string {
  const normalizedLang = languageAliases[lang] || lang || 'text';
  const lines = code.split('\n').filter((_, i, arr) => i < arr.length - 1 || code.endsWith('\n') || code.split('\n').pop());
  
  const lineNumbers = lines
    .map((_, i) => `<span class="code-box__line-number">${i + 1}</span>`)
    .join('\n');
  
  const escapedCode = escapeHtml(code);
  const langColor = getLanguageColor(normalizedLang);
  
  return `
    <div class="code-box" data-language="${normalizedLang}">
      <div class="code-box__header">
        <div class="code-box__header-left">
          <span class="code-box__lang"${langColor ? ` style="color:${langColor}"` : ''}>${normalizedLang}</span>
        </div>
        <button class="code-box__copy" data-code="${escapeHtml(code)}" type="button">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
          </svg>
          <span>Copy</span>
        </button>
      </div>
      <div class="code-box__body">
        <div class="code-box__lines" aria-hidden="true">
          ${lineNumbers}
        </div>
        <pre class="code-box__pre"><code class="code-box__code language-${normalizedLang}">${escapedCode}</code></pre>
      </div>
    </div>
  `;
}

function getLanguageColor(lang: string): string | undefined {
  const colors: Record<string, string> = {
    typescript: '#3178c6', ts: '#3178c6',
    javascript: '#f7df1e', js: '#f7df1e',
    tsx: '#3178c6', jsx: '#f7df1e',
    go: '#00add8',
    python: '#3776ab', py: '#3776ab',
    rust: '#dea584',
    bash: '#4eaa25', sh: '#4eaa25',
    json: '#f7df1e',
    yaml: '#cb171e', yml: '#cb171e',
    sql: '#f29111',
    html: '#e34c26',
    css: '#264de4',
    ruby: '#cc342d',
    java: '#b07219',
    kotlin: '#A97BFF',
    swift: '#F05138',
    php: '#4F5D95',
    c: '#555555',
    cpp: '#f34b7d',
    csharp: '#178600',
    toml: '#9c4221',
    markdown: '#083fa1', md: '#083fa1',
    dockerfile: '#384d54',
    terraform: '#5C4EE5', hcl: '#5C4EE5',
  };
  return colors[lang];
}

/**
 * Convert TipTap/Slate-like JSON to HTML with enhanced code blocks
 */
export function slateBodyToHtml(body: unknown): string {
  if (body == null) return '';

  let processedBody = body;

  if (typeof body === 'string') {
    try {
      const parsed = JSON.parse(body);
      if (parsed && typeof parsed === 'object') {
        processedBody = parsed;
      }
    } catch {
    }
  }

  if (typeof processedBody === 'string') {
    const str = processedBody;
    try {
      const decoded = atob(str);
      const parsed = JSON.parse(decoded);
      processedBody = parsed;
    } catch {
      return fixCliCommands(marked.parse(str, { async: false }) as string);
    }
  }

  if (Array.isArray(processedBody)) {
    return tipTapArrayToHtml(processedBody);
  }

  if (!Array.isArray(processedBody) && typeof processedBody === 'object' && processedBody !== null) {
    const obj = processedBody as Record<string, unknown>;
    if (obj.type === 'doc' && obj.content) {
      return tipTapToHtml(processedBody as TipTapDoc);
    }
    return tipTapToHtml(processedBody as TipTapDoc);
  }

  return '';
}

function tipTapArrayToHtml(nodes: unknown[]): string {
  if (!nodes || nodes.length === 0) return '';
  return nodes.map((node) => renderTipTapNodeFromArray(node)).join('\n');
}

function renderTipTapNodeFromArray(node: unknown): string {
  if (!node || typeof node !== 'object') return '';
  const n = node as Record<string, unknown>;
  if (!n.type || typeof n.type !== 'string') return '';

  const type = n.type as string;
  const content = n.content as unknown[] | undefined;
  const children = n.children as unknown[] | undefined;
  const attrs = n.attrs as Record<string, unknown> | undefined;

  switch (type) {
    case 'heading': {
      const level = (attrs?.level as number) || 2;
      const inner = renderContentArray(children || content);
      return `<h${level} class="content-heading-${level}">${inner}</h${level}>`;
    }
    case 'paragraph': {
      const inner = renderContentArray(children || content);
      if (!inner.trim()) return '<p class="content-paragraph">&nbsp;</p>';
      return `<p class="content-paragraph">${inner}</p>`;
    }
    case 'blockquote': {
      const inner = (children || content || []).map(renderTipTapNodeFromArray).join('');
      return `<blockquote class="content-blockquote">${inner}</blockquote>`;
    }
    case 'bulletList': {
      const items = (children || content || []).map(item => {
        if (typeof item === 'object' && item !== null && (item as Record<string, unknown>).type === 'listItem') {
          const listItem = item as Record<string, unknown>;
          const itemContent = listItem.content as unknown[] | undefined;
          const inner = renderContentArray(itemContent);
          return `<li>${inner}</li>`;
        }
        return '';
      }).join('');
      return `<ul class="content-list content-list--bullet">${items}</ul>`;
    }
    case 'orderedList': {
      const items = (children || content || []).map(item => {
        if (typeof item === 'object' && item !== null && (item as Record<string, unknown>).type === 'listItem') {
          const listItem = item as Record<string, unknown>;
          const itemContent = listItem.content as unknown[] | undefined;
          const inner = renderContentArray(itemContent);
          return `<li>${inner}</li>`;
        }
        return '';
      }).join('');
      return `<ol class="content-list content-list--numbered">${items}</ol>`;
    }
    case 'listItem': {
      const inner = renderContentArray(children || content);
      return inner;
    }
    case 'codeBlock': {
      const lang = (attrs?.language as string) || 'text';
      const rawCode = renderContentArray(children || content);
      return createCodeBlockHtml(rawCode, lang);
    }
    case 'horizontalRule':
      return '<hr class="content-divider">';
    default:
      return '';
  }
}

function renderContentArray(nodes: unknown[] | undefined): string {
  if (!nodes || !Array.isArray(nodes)) return '';
  return nodes.map(node => {
    if (!node || typeof node !== 'object') return '';
    const n = node as Record<string, unknown>;
    if (n.text !== undefined && typeof n.text === 'string') {
      let t = escapeHtml(n.text);
      if (n.marks && Array.isArray(n.marks)) {
        for (const mark of n.marks as Array<{type: string; attrs?: Record<string, unknown>}>) {
          switch (mark.type) {
            case 'code':
              t = `<code class="inline-code">${t}</code>`;
              break;
            case 'bold':
              t = `<strong>${t}</strong>`;
              break;
            case 'italic':
              t = `<em>${t}</em>`;
              break;
            case 'link':
              const href = escapeHtml((mark.attrs?.href as string) || '#');
              t = `<a href="${href}" class="content-link" target="_blank" rel="noopener noreferrer">${t}</a>`;
              break;
            case 'strike':
              t = `<s>${t}</s>`;
              break;
            case 'underline':
              t = `<u>${t}</u>`;
              break;
          }
        }
      }
      return t;
    }
    if (n.type && typeof n.type === 'string') {
      return renderTipTapNodeFromArray(node);
    }
    return '';
  }).join('');
}

// TipTap types
interface TipTapDoc {
  type: 'doc';
  content?: TipTapNode[];
}

interface TipTapNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: TipTapNode[] | TextLeaf[];
  text?: string;
  marks?: Array<{ type: string; attrs?: Record<string, string> }>;
}

// Type guard to check if a node is a TipTapNode (has type property) vs TextLeaf
function isTipTapNode(node: unknown): node is TipTapNode {
  return typeof node === 'object' && node !== null && 'type' in node;
}

/**
 * Convert TipTap JSON format to HTML
 */
function tipTapToHtml(doc: TipTapDoc): string {
  if (!doc.content) return '';
  return doc.content.map(renderTipTapNode).join('\n');
}

function renderTipTapNode(node: TipTapNode | TextLeaf): string {
  // Handle TextLeaf (leaf nodes with text content)
  if ('text' in node && node.text !== undefined) {
    return renderTipTapText(node);
  }
  
  // Must be a TipTapNode at this point
  if (!isTipTapNode(node)) {
    return '';
  }
  
  const { type, attrs, content } = node;

  switch (type) {
    case 'heading': {
      const level = (attrs?.level as number) || 2;
      const inner = content?.map(renderTipTapText).join('') || '';
      return `<h${level} class="content-heading-${level}">${inner}</h${level}>`;
    }
    
    case 'paragraph': {
      const inner = content?.map(renderTipTapText).join('') || '';
      if (!inner.trim()) return '<p class="content-paragraph">&nbsp;</p>';
      return `<p class="content-paragraph">${inner}</p>`;
    }
    
    case 'blockquote': {
      const inner = content?.filter(isTipTapNode).map(renderTipTapNode).join('') || '';
      return `<blockquote class="content-blockquote">${inner}</blockquote>`;
    }
    
    case 'bulletList': {
      const items = content?.map((item) => {
        if (isTipTapNode(item) && item.type === 'listItem') {
          const inner = item.content?.map(renderTipTapNode).join('') || '';
          return `<li>${inner}</li>`;
        }
        return '';
      }).join('') || '';
      return `<ul class="content-list content-list--bullet">${items}</ul>`;
    }
    
    case 'orderedList': {
      const items = content?.map((item) => {
        if (isTipTapNode(item) && item.type === 'listItem') {
          const inner = item.content?.map(renderTipTapNode).join('') || '';
          return `<li>${inner}</li>`;
        }
        return '';
      }).join('') || '';
      return `<ol class="content-list content-list--numbered">${items}</ol>`;
    }
    
    case 'codeBlock': {
      const lang = (attrs?.language as string) || 'text';
      const rawCode = content?.map(renderTipTapText).join('') || '';
      return createCodeBlockHtml(rawCode, lang);
    }
    
    case 'horizontalRule':
      return '<hr class="content-divider">';
    
    default:
      return '';
  }
}

function renderTipTapText(node: TipTapNode | TextLeaf): string {
  if ('text' in node && node.text !== undefined) {
    let t = escapeHtml(node.text);
    
    if (node.marks) {
      for (const mark of node.marks) {
        switch (mark.type) {
          case 'code':
            t = `<code class="inline-code">${t}</code>`;
            break;
          case 'bold':
            t = `<strong>${t}</strong>`;
            break;
          case 'italic':
            t = `<em>${t}</em>`;
            break;
          case 'link':
            const href = escapeHtml(mark.attrs?.href || '#');
            t = `<a href="${href}" class="content-link" target="_blank" rel="noopener noreferrer">${t}</a>`;
            break;
          case 'strike':
            t = `<s>${t}</s>`;
            break;
          case 'underline':
            t = `<u>${t}</u>`;
            break;
        }
      }
    }
    return t;
  }
  return '';
}

/**
 * Legacy Slate format converter
 */
function legacySlateToHtml(blocks: Array<Record<string, unknown>>): string {
  const out: string[] = [];
  let i = 0;

  while (i < blocks.length) {
    const block = blocks[i];
    if (!block || typeof block !== 'object' || typeof block.type !== 'string') {
      i++;
      continue;
    }

    const type = block.type as string;

    if (type === 'heading') {
      const level = Math.min(6, Math.max(1, Number(block.level) || 2));
      const inner = paragraphInlineHtml(block.children);
      out.push(`<h${level} class="content-heading-${level}">${inner}</h${level}>`);
      i++;
      continue;
    }

    if (type === 'paragraph') {
      const raw = rawParagraphText(block.children).trim();
      const fence = raw.match(/^```(\w*)/);
      if (fence) {
        const lang = (fence[1] || 'text').toLowerCase();
        i++;
        const codeLines: string[] = [];
        while (i < blocks.length) {
          const b = blocks[i];
          if (!b || b.type !== 'paragraph') break;
          const line = rawParagraphText(b.children);
          if (line.trim() === '```') {
            i++;
            break;
          }
          codeLines.push(line);
          i++;
        }
        const code = codeLines.join('\n');
        out.push(createCodeBlockHtml(code, lang));
        continue;
      }
      const inner = paragraphInlineHtml(block.children);
      if (inner.trim()) out.push(`<p class="content-paragraph">${inner}</p>`);
      i++;
      continue;
    }

    if (type === 'codeBlock') {
      const lang = String(block.language || 'text').toLowerCase();
      const code = rawParagraphText(block.children);
      out.push(createCodeBlockHtml(code, lang));
      i++;
      continue;
    }

    if (type === 'blockquote') {
      const inner = paragraphInlineHtml(block.children);
      out.push(`<blockquote class="content-blockquote">${inner}</blockquote>`);
      i++;
      continue;
    }

    if (type === 'bulleted-list' || type === 'numbered-list') {
      const listTag = type === 'numbered-list' ? 'ol' : 'ul';
      const listClass = type === 'numbered-list' ? 'content-list--numbered' : 'content-list--bullet';
      const items: string[] = [];
      i++;
      while (i < blocks.length && blocks[i]?.type === 'list-item') {
        const liBlock = blocks[i] as { children?: unknown };
        items.push(`<li>${paragraphInlineHtml(liBlock.children)}</li>`);
        i++;
      }
      if (items.length) out.push(`<${listTag} class="content-list ${listClass}">${items.join('')}</${listTag}>`);
      continue;
    }

    if (type === 'list-item') {
      const inner = paragraphInlineHtml(block.children);
      out.push(`<ul class="content-list"><li>${inner}</li></ul>`);
      i++;
      continue;
    }

    i++;
  }

  return out.join('\n');
}

/**
 * Calculate reading time in minutes from content
 */
export function calculateReadingTime(body: unknown): number {
  if (!body) return 1;

  let text = '';

  if (typeof body === 'string') {
    try {
      const parsed = JSON.parse(body);
      if (parsed.content) {
        // TipTap format
        text = extractTipTapText(parsed);
      } else {
        text = body;
      }
    } catch {
      text = body;
    }
  } else if (typeof body === 'object' && body !== null) {
    if ('content' in body) {
      text = extractTipTapText(body as TipTapDoc);
    } else if (Array.isArray(body)) {
      // Legacy Slate format
      text = body
        .map((block: Record<string, unknown>) => {
          if (block.children && Array.isArray(block.children)) {
            return block.children
              .map((child: unknown) => {
                if (child && typeof child === 'object' && 'text' in child) {
                  return String((child as { text?: string }).text ?? '');
                }
                return '';
              })
              .join(' ');
          }
          return '';
        })
        .join(' ');
    }
  }

  const words = text.trim().split(/\s+/).filter((w) => w.length > 0).length;
  return Math.max(1, Math.ceil(words / 200));
}

function extractTipTapText(doc: TipTapDoc): string {
  if (!doc.content) return '';
  
  const extractNode = (node: TipTapNode | TextLeaf): string => {
    if ('text' in node && node.text) {
      return node.text;
    }
    if ('content' in node && node.content && Array.isArray(node.content)) {
      return node.content.map(extractNode).join(' ');
    }
    return '';
  };

  return doc.content.map(extractNode).join(' ');
}

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

  let html = '';

  if (typeof processedBody === 'string') {
    const str = processedBody;
    try {
      const decoded = atob(str);
      const parsed = JSON.parse(decoded);
      processedBody = parsed;
    } catch {
      html = fixCliCommands(marked.parse(str, { async: false }) as string);
      return enhanceHtml(html);
    }
  }

  if (Array.isArray(processedBody)) {
    html = tipTapArrayToHtml(processedBody);
  } else if (!Array.isArray(processedBody) && typeof processedBody === 'object' && processedBody !== null) {
    const obj = processedBody as Record<string, unknown>;
    if (obj.type === 'doc' && obj.content) {
      html = tipTapToHtml(processedBody as TipTapDoc);
    } else {
      html = tipTapToHtml(processedBody as TipTapDoc);
    }
  } else {
    html = '';
  }

  return enhanceHtml(html);
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

/**
 * Enhance HTML with shortcode syntax for enhanced components
 * Supports:
 * - :::callout[type] content ::: - Callout boxes (tip, warning, info, important)
 * - :::comparison ... ::: - Side-by-side comparison cards
 * - :::api-table ... [/api-table] - API comparison table
 * - :::workflow ... ::: - Numbered workflow steps
 * - :::lifecycle stage1 > stage2 > stage3 ::: - Lifecycle flow badges
 * - :::decision ... ::: - Decision grid cards
 */
function enhanceHtml(html: string): string {
  // Callout boxes: :::callout[type] content :::
  html = html.replace(
    /:::callout\[(\w+)\]([\s\S]*?):::/g,
    (match, type, content) => {
      const icons: Record<string, string> = {
        tip: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>',
        warning: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
        info: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>',
        important: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"/><path d="M12 8v4"/><path d="M12 16h.01"/></svg>',
      };
      const icon = icons[type] || icons.tip;
      return `<div class="callout callout-${type}"><div class="callout-icon">${icon}</div><div class="callout-content">${content.trim()}</div></div>`;
    }
  );

  // Lifecycle flow: :::lifecycle draft > published > deprecated > archived :::
  html = html.replace(
    /:::lifecycle([\s\S]*?):::/g,
    (match, stages) => {
      const stagesList = stages.split('>').map((s: string) => s.trim());
      const badges = stagesList.map((stage: string, i: number) => {
        const arrows = stagesList.slice(0, i).join('');
        return `<span class="lifecycle-badge ${stage}">${stage}</span>`;
      }).join('<svg class="lifecycle-arrow" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>');
      return `<div class="lifecycle-flow">${badges}</div>`;
    }
  );

  // API Table: :::api-table\n| label | platform | registry | ...\n[/api-table]
  html = html.replace(
    /:::api-table\n([\s\S]*?)\[\/api-table\]/g,
    (match, tableContent) => {
      const lines = tableContent.trim().split('\n').filter((l: string) => l.startsWith('|'));
      if (lines.length < 2) return match;

      const headerMatch = lines[0].match(/\|(.*?)\|/g);
      if (!headerMatch) return match;

      let headerHtml = '<div class="api-comparison-header"><div class="api-header-label"></div>';
      const hasPlatform = lines[0].includes('Platform');
      const hasRegistry = lines[0].includes('Registry');

      if (hasPlatform) {
        headerHtml += '<div class="api-header-col"><span class="platform-badge"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>Platform (Create)</span></div>';
      }
      if (hasRegistry) {
        headerHtml += '<div class="api-header-col"><span class="registry-badge"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>Registry (Publish)</span></div>';
      }
      headerHtml += '</div>';

      let rowsHtml = '';
      for (let i = 1; i < lines.length; i++) {
        const cells = lines[i].split('|').filter((c: string, idx: number) => idx > 0 && idx < lines[i].split('|').length - 1);
        if (cells.length < 2) continue;

        const label = cells[0].trim();
        const platform = cells[1]?.trim() || '';
        const registry = cells[2]?.trim() || '';

        rowsHtml += `<div class="api-row"><div class="api-label"><strong>${label}</strong></div><div class="api-cell platform-cell">${platform}</div><div class="api-cell registry-cell">${registry}</div></div>`;
      }

      return `<div class="api-comparison-card">${headerHtml}${rowsHtml}</div>`;
    }
  );

  // Workflow: :::workflow\n1. step code — description\n2. ...
  html = html.replace(
    /:::workflow\n([\s\S]*?)\[\/workflow\]/g,
    (match, stepsContent) => {
      const steps = stepsContent.trim().split('\n').filter((l: string) => l.trim());
      let stepsHtml = '<div class="workflow-card"><ol>';

      steps.forEach((step: string, i: number) => {
        const match = step.match(/^\d+\.\s*(.+?)(?:\s*[-–—]\s*(.+))?$/);
        if (match) {
          const code = match[1].trim();
          const desc = match[2] ? `<span class="step-desc">${match[2].trim()}</span>` : '';
          stepsHtml += `<li><code>${code}</code>${desc}</li>`;
        }
      });

      stepsHtml += '</ol></div>';
      return stepsHtml;
    }
  );

  // Decision Grid: :::decision\n[platform]\nTitle\n- item1\n- item2\n[/platform]
  html = html.replace(
    /:::decision\n([\s\S]*?)\[\/decision\]/g,
    (match, content) => {
      const platformMatch = content.match(/\[platform\]\s*\n*(.+?)\n([\s\S]*?)\[\/platform\]/);
      const registryMatch = content.match(/\[registry\]\s*\n*(.+?)\n([\s\S]*?)\[\/registry\]/);

      let cards = '';

      if (platformMatch) {
        const title = platformMatch[1].trim();
        const items = platformMatch[2].split('\n').filter((l: string) => l.startsWith('-')).map((l: string) => `<li>${l.replace(/^-\s*/, '')}</li>`).join('');
        cards += `<div class="decision-card platform"><h3>${title}</h3><ul>${items}</ul></div>`;
      }

      if (registryMatch) {
        const title = registryMatch[1].trim();
        const items = registryMatch[2].split('\n').filter((l: string) => l.startsWith('-')).map((l: string) => `<li>${l.replace(/^-\s*/, '')}</li>`).join('');
        cards += `<div class="decision-card registry"><h3>${title}</h3><ul>${items}</ul></div>`;
      }

      return `<div class="decision-grid">${cards}</div>`;
    }
  );

  // Comparison: :::comparison\n[create]\nTitle\nDescription\n[/create]
  html = html.replace(
    /:::comparison\n([\s\S]*?)\[\/comparison\]/g,
    (match, content) => {
      const createMatch = content.match(/\[create\]\s*\n*(.+?)\n\n([\s\S]*?)\[\/create\]/);
      const publishMatch = content.match(/\[publish\]\s*\n*(.+?)\n\n([\s\S]*?)\[\/publish\]/);

      let cards = '';

      if (createMatch) {
        cards += `<div class="comparison-card create"><h3>${createMatch[1].trim()}</h3><p>${createMatch[2].trim()}</p></div>`;
      }

      if (publishMatch) {
        cards += `<div class="comparison-card publish"><h3>${publishMatch[1].trim()}</h3><p>${publishMatch[2].trim()}</p></div>`;
      }

      return `<div class="comparison-highlight">${cards}</div>`;
    }
  );

  return html;
}

/**
 * Serialize NestJS blog post `body` (Slate-like JSON) to safe HTML for the marketing site.
 * Handles paragraphs, headings, inline marks, and ``` fenced code blocks split across paragraphs.
 */

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
}

function renderLeaf(leaf: TextLeaf): string {
  let t = escapeHtml(leaf.text ?? '');
  if (leaf.code) t = `<code>${t}</code>`;
  if (leaf.bold) t = `<strong>${t}</strong>`;
  if (leaf.italic) t = `<em>${t}</em>`;
  if (leaf.underline) t = `<u>${t}</u>`;
  if (leaf.strikethrough) t = `<s>${t}</s>`;
  if (leaf.link) {
    const href = escapeHtml(leaf.link);
    t = `<a href="${href}" target="_blank" rel="noopener noreferrer">${t}</a>`;
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
 * Convert Slate-like block array to HTML string for use with set:html in Astro.
 */
export function slateBodyToHtml(body: unknown): string {
  if (body == null) return '';
  if (typeof body === 'string') {
    return `<p>${escapeHtml(body)}</p>`;
  }
  if (!Array.isArray(body)) return '';

  const blocks = body as Array<Record<string, unknown>>;
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
      out.push(`<h${level}>${inner}</h${level}>`);
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
        const code = escapeHtml(codeLines.join('\n'));
        out.push(`<pre class="blog-code-block"><code class="language-${escapeHtml(lang)}">${code}</code></pre>`);
        continue;
      }
      const inner = paragraphInlineHtml(block.children);
      if (inner.trim()) out.push(`<p>${inner}</p>`);
      i++;
      continue;
    }

    if (type === 'bulleted-list' || type === 'numbered-list') {
      const listTag = type === 'numbered-list' ? 'ol' : 'ul';
      const items: string[] = [];
      i++;
      while (i < blocks.length && blocks[i]?.type === 'list-item') {
        const liBlock = blocks[i] as { children?: unknown };
        items.push(`<li>${paragraphInlineHtml(liBlock.children)}</li>`);
        i++;
      }
      if (items.length) out.push(`<${listTag} class="blog-list">${items.join('')}</${listTag}>`);
      continue;
    }

    if (type === 'list-item') {
      const inner = paragraphInlineHtml(block.children);
      out.push(`<ul class="blog-list"><li>${inner}</li></ul>`);
      i++;
      continue;
    }

    i++;
  }

  return out.join('\n');
}

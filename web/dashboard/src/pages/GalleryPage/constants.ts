export type GalleryCategory =
  | 'all'
  | 'data-processing'
  | 'api'
  | 'ml'
  | 'web-scraping'
  | 'automation'
  | 'utility'
  | 'finance';

export type ViewMode = 'grid' | 'runway' | 'radar';

export const RUNTIME_COLORS: Record<string, { primary: string; glow: string; accent: string }> = {
  python: { primary: '#3b82f6', glow: '#60a5fa', accent: '#1d4ed8' },
  'python-light': { primary: '#60a5fa', glow: '#93c5fd', accent: '#2563eb' },
  nodejs: { primary: '#10b981', glow: '#34d399', accent: '#059669' },
  typescript: { primary: '#10b981', glow: '#34d399', accent: '#059669' },
  go: { primary: '#06b6d4', glow: '#22d3ee', accent: '#0891b2' },
  rust: { primary: '#f97316', glow: '#fb923c', accent: '#ea580c' },
  deno: { primary: '#8b5cf6', glow: '#a78bfa', accent: '#7c3aed' },
  bun: { primary: '#f59e0b', glow: '#fbbf24', accent: '#d97706' },
  java: { primary: '#ef4444', glow: '#f87171', accent: '#dc2626' },
  csharp: { primary: '#ec4899', glow: '#f472b6', accent: '#db2777' },
  ruby: { primary: '#dc2626', glow: '#ef4444', accent: '#b91c1c' },
  php: { primary: '#6366f1', glow: '#818cf8', accent: '#4f46e5' },
};

export const RUNTIME_ICONS: Record<string, string> = {
  python: '🐍',
  'python-light': '🐍',
  nodejs: '🟢',
  typescript: '📘',
  go: '🐹',
  rust: '🦀',
  deno: '🦕',
  bun: '🥯',
  java: '☕',
  csharp: '#️⃣',
  ruby: '💎',
  php: '🐘',
};

export const CATEGORY_META: Record<
  string,
  { label: string; color: string; center: [number, number, number] }
> = {
  'data-processing': { label: 'Data', color: '#3b82f6', center: [-15, 5, -10] },
  api: { label: 'API', color: '#8b5cf6', center: [0, 8, -5] },
  ml: { label: 'ML/AI', color: '#ec4899', center: [15, 5, -10] },
  'web-scraping': { label: 'Web', color: '#10b981', center: [-12, -5, 8] },
  automation: { label: 'Auto', color: '#f59e0b', center: [0, -8, 5] },
  utility: { label: 'Utils', color: '#6366f1', center: [12, -5, 8] },
  finance: { label: 'Finance', color: '#14b8a6', center: [0, 0, 15] },
  default: { label: 'General', color: '#64748b', center: [0, 0, 0] },
};

/** Map backend runtime strings (e.g. python3.12) to gallery palette keys */
export function normalizeRuntime(runtime: string): string {
  const r = runtime.toLowerCase();
  if (r.startsWith('python')) return 'python';
  if (r.startsWith('node') || r === 'javascript') return 'nodejs';
  if (r.startsWith('typescript') || r === 'deno' || r === 'bun') return 'typescript';
  if (r.startsWith('go')) return 'go';
  if (r.startsWith('rust')) return 'rust';
  if (r.startsWith('java')) return 'java';
  if (r.startsWith('csharp') || r.startsWith('c#')) return 'csharp';
  if (r.startsWith('ruby')) return 'ruby';
  if (r.startsWith('php')) return 'php';
  return r.split(/[^a-z]/)[0] || 'python';
}

/** Normalize API function records for gallery UI */
export function mapGalleryFunction<T extends { runtime?: string; trust_score?: number }>(fn: T): T & { runtime: string; trust_score: number } {
  const trust = fn.trust_score ?? 0;
  return {
    ...fn,
    runtime: normalizeRuntime(fn.runtime || 'python'),
    trust_score: trust > 0 && trust <= 1 ? Math.round(trust * 100) : trust,
  };
}

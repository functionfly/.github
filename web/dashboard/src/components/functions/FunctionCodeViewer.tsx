import { useState, useCallback, useMemo, useEffect, type ReactNode } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import {
  Copy,
  Check,
  Download,
  Maximize2,
  Minimize2,
  FileCode2,
  Terminal,
  Hash,
  FileJson,
  FileText,
  Settings,
  Code2,
  Braces,
  FlaskConical,
  Shield,
  ChevronDown,
  ChevronUp,
  Lock,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

// =============================================================================
// Types
// =============================================================================

interface FileNode {
  name: string;
  icon: ReactNode;
  isActive?: boolean;
  isConfig?: boolean;
}

interface FunctionCodeViewerProps {
  code: string;
  runtime: string;
  functionName: string;
  version?: string;
  lastModified?: string;
  onEdit?: () => void;
  className?: string;
}

// =============================================================================
// Runtime Configuration
// =============================================================================

const RUNTIME_CONFIG: Record<
  string,
  {
    label: string;
    color: string;
    glowColor: string;
    syntaxLang: string;
    files: FileNode[];
  }
> = {
  typescript: {
    label: 'TypeScript',
    color: '#3178c6',
    glowColor: 'rgba(49, 120, 198, 0.15)',
    syntaxLang: 'typescript',
    files: [
      { name: 'handler.ts', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'package.json', icon: <FileJson className="w-3.5 h-3.5" />, isConfig: true },
      { name: 'tsconfig.json', icon: <Settings className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  javascript: {
    label: 'JavaScript',
    color: '#f7df1e',
    glowColor: 'rgba(247, 223, 30, 0.15)',
    syntaxLang: 'javascript',
    files: [
      { name: 'handler.js', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'package.json', icon: <FileJson className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  python: {
    label: 'Python',
    color: '#3572A5',
    glowColor: 'rgba(53, 114, 165, 0.15)',
    syntaxLang: 'python',
    files: [
      { name: 'handler.py', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'requirements.txt', icon: <FileText className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  'python-wasm': {
    label: 'MicroPython',
    color: '#2b5b84',
    glowColor: 'rgba(43, 91, 132, 0.15)',
    syntaxLang: 'python',
    files: [
      { name: 'handler.py', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'manifest.json', icon: <FileJson className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  'rust-wasm': {
    label: 'Rust / WASM',
    color: '#dea584',
    glowColor: 'rgba(222, 165, 132, 0.15)',
    syntaxLang: 'rust',
    files: [
      { name: 'lib.rs', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'Cargo.toml', icon: <Settings className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  go: {
    label: 'Go',
    color: '#00add8',
    glowColor: 'rgba(0, 173, 216, 0.15)',
    syntaxLang: 'go',
    files: [
      { name: 'handler.go', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'go.mod', icon: <Settings className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  deno: {
    label: 'Deno',
    color: '#ffffff',
    glowColor: 'rgba(255, 255, 255, 0.10)',
    syntaxLang: 'typescript',
    files: [
      { name: 'handler.ts', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'deno.json', icon: <FileJson className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  bun: {
    label: 'Bun',
    color: '#f9f1e5',
    glowColor: 'rgba(249, 241, 229, 0.10)',
    syntaxLang: 'typescript',
    files: [
      { name: 'handler.ts', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'package.json', icon: <FileJson className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
  'browser-wasm': {
    label: 'Browser WASM',
    color: '#6b46c1',
    glowColor: 'rgba(107, 70, 193, 0.15)',
    syntaxLang: 'rust',
    files: [
      { name: 'lib.rs', icon: <Code2 className="w-3.5 h-3.5" />, isActive: true },
      { name: 'Cargo.toml', icon: <Settings className="w-3.5 h-3.5" />, isConfig: true },
    ],
  },
};

const FALLBACK_RUNTIME_CONFIG = {
  label: 'Unknown',
  color: '#6366f1',
  glowColor: 'rgba(99, 102, 241, 0.15)',
  syntaxLang: 'text',
  files: [{ name: 'handler.txt', icon: <FileCode2 className="w-3.5 h-3.5" />, isActive: true }],
};

function getRuntimeConfig(runtime: string) {
  return RUNTIME_CONFIG[runtime] ?? { ...FALLBACK_RUNTIME_CONFIG, label: runtime };
}

// =============================================================================
// Monaco-style theme override (module-level — never recreated)
// =============================================================================

const syntaxHighlighterCustomStyle = {
  ...vscDarkPlus,
  'pre[class*="language-"]': {
    ...vscDarkPlus['pre[class*="language-"]'],
    background: 'transparent',
    margin: 0,
    padding: '1rem 1.25rem',
    fontSize: '13px',
    lineHeight: '1.6',
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
  },
  'code[class*="language-"]': {
    ...vscDarkPlus['code[class*="language-"]'],
    background: 'transparent',
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
    fontSize: '13px',
    lineHeight: '1.6',
  },
};

// =============================================================================
// Helpers
// =============================================================================

const textEncoder = typeof TextEncoder !== 'undefined' ? new TextEncoder() : null;

function generateSourceFingerprint(code: string): string {
  let hash = 0;
  for (let i = 0; i < code.length; i++) {
    const char = code.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash |= 0;
  }
  const hex = Math.abs(hash).toString(16).padStart(8, '0');
  return `0x${hex}${hex}`.slice(0, 18);
}

function formatBytes(str: string): string {
  const bytes = textEncoder ? textEncoder.encode(str).length : new Blob([str]).size;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}

function analyzeLines(code: string): Array<{ depth: number; type: 'empty' | 'comment' | 'string' | 'code' }> {
  return code.split('\n').map((line) => {
    const trimmed = line.trim();
    if (trimmed.length === 0) return { depth: 0, type: 'empty' as const };
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('#') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      return { depth: line.search(/\S/) || 0, type: 'comment' as const };
    }
    if (trimmed.startsWith('"') || trimmed.startsWith("'") || trimmed.startsWith('`')) {
      return { depth: line.search(/\S/) || 0, type: 'string' as const };
    }
    return { depth: line.search(/\S/) || 0, type: 'code' as const };
  });
}

/**
 * Clipboard write with fallback for non-HTTPS / legacy contexts.
 * Returns true on success.
 */
async function writeToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // fall through to legacy approach
  }
  // Legacy fallback
  try {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(textarea);
    return ok;
  } catch {
    return false;
  }
}

/**
 * Trigger a file download from a string.
 */
function downloadString(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  anchor.style.display = 'none';
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

// =============================================================================
// Sub-components
// =============================================================================

const LINE_COLORS = {
  comment: (opacity: number) => `rgba(100, 116, 139, ${opacity * 0.6})`,
  string: (opacity: number) => `rgba(16, 185, 129, ${opacity * 0.7})`,
  empty: () => 'transparent',
} as const;

function hexToRgb(hex: string): string {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? `rgb(${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)})`
    : 'rgb(99, 102, 241)';
}

function CodeMinimap({ code, runtimeColor }: { code: string; runtimeColor: string }) {
  const lines = useMemo(() => analyzeLines(code), [code]);
  const rgbColor = useMemo(() => hexToRgb(runtimeColor), [runtimeColor]);

  const getLineColor = useCallback(
    (type: string, depth: number) => {
      const opacity = Math.min(0.3 + depth * 0.08, 0.9);
      if (type === 'comment') return LINE_COLORS.comment(opacity);
      if (type === 'string') return LINE_COLORS.string(opacity);
      if (type === 'empty') return LINE_COLORS.empty();
      return `${rgbColor.replace(')', `, ${opacity})`).replace('rgb', 'rgba')}`;
    },
    [rgbColor]
  );

  return (
    <div className="w-12 shrink-0 border-l border-border-subtle bg-bg-secondary/50 flex flex-col items-center py-2 select-none">
      <div className="text-[9px] text-text-muted font-mono mb-1 tracking-wider uppercase">Map</div>
      <ScrollArea className="flex-1 w-full">
        <div className="px-1 space-y-[1px]">
          {lines.map((line, i) => (
            <div
              key={i}
              className="w-full rounded-[1px] transition-opacity hover:opacity-100"
              style={{
                height: line.type === 'empty' ? '2px' : '3px',
                backgroundColor: line.type === 'empty' ? 'transparent' : getLineColor(line.type, line.depth),
                opacity: line.type === 'empty' ? 0 : 0.7,
                marginLeft: `${Math.min(line.depth * 1.5, 8)}px`,
                width: line.type === 'empty' ? '100%' : `${Math.max(30, 100 - line.depth * 6)}%`,
              }}
              title={`Line ${i + 1}`}
            />
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}

function FileTree({ files, runtimeColor }: { files: FileNode[]; runtimeColor: string }) {
  return (
    <div className="w-44 shrink-0 border-r border-border-subtle bg-bg-secondary/30 flex flex-col">
      <div className="px-3 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-1.5 text-xs text-text-muted font-medium uppercase tracking-wider">
          <Braces className="w-3 h-3" />
          Source Files
        </div>
      </div>
      <div className="p-2 space-y-0.5">
        {files.map((file) => (
          <div
            key={file.name}
            className={cn(
              'flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs transition-colors',
              file.isActive
                ? 'bg-bg-tertiary text-text-primary cursor-default'
                : 'text-text-muted cursor-default'
            )}
            title={file.isActive ? file.name : `${file.name} — not available for single-file source`}
          >
            <span style={{ color: file.isActive ? runtimeColor : undefined }}>{file.icon}</span>
            <span className="truncate">{file.name}</span>
            {file.isConfig && (
              <Badge variant="outline" className="text-[9px] h-4 px-1 ml-auto border-border-subtle text-text-muted">
                config
              </Badge>
            )}
          </div>
        ))}
      </div>
      <div className="mt-auto p-3 border-t border-border-subtle">
        <div className="flex items-center gap-1.5 text-[10px] text-text-muted">
          <Shield className="w-3 h-3 text-emerald-400" />
          <span>Zero-knowledge vault</span>
        </div>
      </div>
    </div>
  );
}

// =============================================================================
// Collapsed Header
// =============================================================================

function CollapsedHeader({
  config,
  functionName,
  lineCount,
  sizeStr,
  fingerprint,
  runtimeColor,
}: {
  config: ReturnType<typeof getRuntimeConfig>;
  functionName: string;
  lineCount: number;
  sizeStr: string;
  fingerprint: string;
  runtimeColor: string;
}) {
  return (
    <div className="flex items-center justify-between gap-4 px-5 py-4">
      <div className="flex items-center gap-3 min-w-0">
        <div
          className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: `${runtimeColor}15`, border: `1px solid ${runtimeColor}30` }}
        >
          <Code2 className="w-5 h-5" style={{ color: runtimeColor }} />
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-text-primary truncate">{functionName}</span>
            <Badge
              variant="outline"
              className="text-[10px] h-5 px-1.5 font-mono border-border-subtle shrink-0"
              style={{ color: config.color, borderColor: `${config.color}40` }}
            >
              {config.label}
            </Badge>
          </div>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-xs text-text-muted font-mono">{lineCount} lines</span>
            <span className="text-border-default text-xs">·</span>
            <span className="text-xs text-text-muted font-mono">{sizeStr}</span>
            <span className="text-border-default text-xs hidden sm:inline">·</span>
            <span className="text-xs text-text-muted font-mono hidden sm:inline">{fingerprint}</span>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        <div className="flex items-center gap-1 text-xs text-text-muted">
          <Lock className="w-3 h-3 text-emerald-400" />
          <span className="hidden sm:inline">Verified source</span>
        </div>
        <div className="flex items-center gap-1 text-xs text-brand-400 font-medium">
          <span>View source</span>
          <ChevronDown className="w-4 h-4" />
        </div>
      </div>
    </div>
  );
}

// =============================================================================
// Main Component
// =============================================================================

export function FunctionCodeViewer({
  code,
  runtime,
  functionName,
  version,
  lastModified,
  onEdit,
  className,
}: FunctionCodeViewerProps) {
  const [copied, setCopied] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showMinimap, setShowMinimap] = useState(true);
  const [isExpanded, setIsExpanded] = useState(false);
  const [isRevealed, setIsRevealed] = useState(false);

  const config = useMemo(() => getRuntimeConfig(runtime), [runtime]);
  const fingerprint = useMemo(() => generateSourceFingerprint(code), [code]);
  const lineCount = useMemo(() => code.split('\n').length, [code]);
  const charCount = useMemo(() => code.length, [code]);
  const sizeStr = useMemo(() => formatBytes(code), [code]);
  const activeFileName = useMemo(
    () => config.files.find((f) => f.isActive)?.name ?? 'handler',
    [config]
  );

  useEffect(() => {
    if (isExpanded) {
      const timer = setTimeout(() => setIsRevealed(true), 100);
      return () => clearTimeout(timer);
    }
    setIsRevealed(false);
  }, [isExpanded]);

  // Lock body scroll when fullscreen
  useEffect(() => {
    if (isFullscreen) {
      const prev = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
      return () => {
        document.body.style.overflow = prev;
      };
    }
  }, [isFullscreen]);

  const handleCopy = useCallback(async () => {
    const ok = await writeToClipboard(code);
    if (ok) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [code]);

  const handleDownload = useCallback(() => {
    const ext = config.files[0]?.name.split('.').pop() ?? 'txt';
    downloadString(code, `${functionName}.${ext}`);
  }, [code, functionName, config]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: 'easeOut' }}
      className={cn(
        'flex flex-col overflow-hidden border border-border-subtle bg-bg-secondary rounded-xl transition-shadow duration-300',
        isFullscreen ? 'fixed inset-4 z-50 shadow-2xl' : 'relative',
        isExpanded && 'cursor-default',
        !isExpanded && 'cursor-pointer hover:border-border-default',
        className
      )}
      style={{
        boxShadow: isFullscreen
          ? `0 0 60px ${config.glowColor}, 0 25px 50px -12px rgba(0,0,0,0.5)`
          : isExpanded
            ? `0 0 30px ${config.glowColor}, 0 4px 6px -1px rgba(0,0,0,0.1)`
            : `0 0 15px ${config.glowColor}, 0 1px 3px rgba(0,0,0,0.08)`,
      }}
      onClick={() => {
        if (!isExpanded) setIsExpanded(true);
      }}
      role={!isExpanded ? 'button' : undefined}
      tabIndex={!isExpanded ? 0 : undefined}
      onKeyDown={(e) => {
        if (!isExpanded && (e.key === 'Enter' || e.key === ' ')) {
          e.preventDefault();
          setIsExpanded(true);
        }
      }}
      aria-label={!isExpanded ? `View ${functionName} source code` : undefined}
    >
      {/* Collapsed Header */}
      {!isExpanded && (
        <CollapsedHeader
          config={config}
          functionName={functionName}
          lineCount={lineCount}
          sizeStr={sizeStr}
          fingerprint={fingerprint}
          runtimeColor={config.color}
        />
      )}

      {/* Expanded Content */}
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.3, ease: 'easeInOut' }}
            className="flex flex-col overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Top Bar */}
            <div className="flex items-center justify-between px-4 py-2.5 border-b border-border-subtle bg-bg-tertiary/60 backdrop-blur-sm shrink-0">
              <div className="flex items-center gap-3 min-w-0">
                <div className="flex items-center gap-1.5 text-xs text-text-muted">
                  <Terminal className="w-3.5 h-3.5" />
                  <span className="hidden sm:inline">src</span>
                  <span className="text-border-default">/</span>
                  <span className="text-text-secondary truncate">{activeFileName}</span>
                </div>
                <Badge
                  variant="outline"
                  className="text-[10px] h-5 px-1.5 font-mono border-border-subtle shrink-0"
                  style={{ color: config.color, borderColor: `${config.color}40` }}
                >
                  <FlaskConical className="w-3 h-3 mr-1" />
                  {config.label}
                </Badge>
                {version && (
                  <Badge variant="secondary" className="text-[10px] h-5 px-1.5 shrink-0">
                    v{version}
                  </Badge>
                )}
              </div>

              <div className="flex items-center gap-1 shrink-0">
                <AnimatePresence mode="wait">
                  {copied ? (
                    <motion.span
                      key="copied"
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.8 }}
                      className="text-xs text-emerald-400 flex items-center gap-1 mr-2"
                    >
                      <Check className="w-3.5 h-3.5" />
                      <span className="hidden sm:inline">Copied</span>
                    </motion.span>
                  ) : null}
                </AnimatePresence>

                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                  onClick={handleCopy}
                  aria-label="Copy code to clipboard"
                >
                  <Copy className="w-3.5 h-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                  onClick={handleDownload}
                  aria-label="Download source file"
                >
                  <Download className="w-3.5 h-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary hidden sm:flex"
                  onClick={() => setShowMinimap((v) => !v)}
                  aria-label={showMinimap ? 'Hide minimap' : 'Show minimap'}
                  aria-pressed={showMinimap}
                >
                  <Hash className="w-3.5 h-3.5" />
                </Button>
                {onEdit && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-xs text-text-muted hover:text-text-primary hidden sm:inline-flex"
                    onClick={onEdit}
                  >
                    <FileCode2 className="w-3.5 h-3.5 mr-1" />
                    Edit
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                  onClick={() => setIsFullscreen((f) => !f)}
                  aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
                >
                  {isFullscreen ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                  onClick={() => setIsExpanded(false)}
                  aria-label="Collapse source viewer"
                >
                  <ChevronUp className="w-3.5 h-3.5" />
                </Button>
              </div>
            </div>

            {/* Main Body */}
            <div
              className="flex flex-1 min-h-0 overflow-hidden"
              style={{ maxHeight: isFullscreen ? 'calc(100vh - 100px)' : '600px' }}
            >
              <FileTree files={config.files} runtimeColor={config.color} />

              {/* Code Area */}
              <div className="flex-1 min-w-0 overflow-auto bg-[#0d0d12]">
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: isRevealed ? 1 : 0 }}
                  transition={{ duration: 0.6, delay: 0.15 }}
                >
                  <SyntaxHighlighter
                    language={config.syntaxLang}
                    style={syntaxHighlighterCustomStyle}
                    showLineNumbers
                    lineNumberStyle={{
                      minWidth: '3em',
                      paddingRight: '1.25em',
                      color: '#3d3d4d',
                      fontSize: '12px',
                      fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
                      userSelect: 'none',
                    }}
                    wrapLines
                    wrapLongLines={false}
                  >
                    {code.trim() || '// No source code available'}
                  </SyntaxHighlighter>
                </motion.div>
              </div>

              {/* Code Minimap */}
              <AnimatePresence>
                {showMinimap && (
                  <motion.div
                    initial={{ width: 0, opacity: 0 }}
                    animate={{ width: 48, opacity: 1 }}
                    exit={{ width: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="shrink-0 overflow-hidden"
                  >
                    <CodeMinimap code={code} runtimeColor={config.color} />
                  </motion.div>
                )}
              </AnimatePresence>
            </div>

            {/* Bottom Bar */}
            <div className="flex items-center justify-between px-4 py-1.5 border-t border-border-subtle bg-bg-tertiary/40 backdrop-blur-sm shrink-0">
              <div className="flex items-center gap-3 text-[11px] text-text-muted font-mono">
                <span className="flex items-center gap-1">
                  <Code2 className="w-3 h-3" />
                  {lineCount} lines
                </span>
                <span className="text-border-default">|</span>
                <span>{charCount.toLocaleString()} chars</span>
                <span className="text-border-default">|</span>
                <span>{sizeStr}</span>
                {lastModified && (
                  <>
                    <span className="text-border-default hidden sm:inline">|</span>
                    <span className="hidden sm:inline">Modified {lastModified}</span>
                  </>
                )}
              </div>

              <div className="flex items-center gap-2">
                <div className="flex items-center gap-1 text-[10px] text-text-muted font-mono">
                  <Hash className="w-3 h-3" />
                  <span className="hidden sm:inline">{fingerprint}</span>
                  <span className="sm:hidden">{fingerprint.slice(0, 10)}…</span>
                </div>
                <div
                  className="w-1.5 h-1.5 rounded-full animate-pulse"
                  style={{ backgroundColor: config.color }}
                  title={`${config.label} runtime`}
                />
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Fullscreen backdrop */}
      {isFullscreen && (
        <div
          className="fixed inset-0 bg-black/70 backdrop-blur-sm -z-10"
          onClick={() => setIsFullscreen(false)}
          aria-hidden="true"
        />
      )}
    </motion.div>
  );
}

FunctionCodeViewer.displayName = 'FunctionCodeViewer';

export default FunctionCodeViewer;

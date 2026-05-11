import { useState, useEffect, useCallback } from 'react';
import { CodeBlock } from '@/components/common/CodeBlock';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BookOpen,
  ExternalLink,
  Play,
  ChevronDown,
  ChevronRight,
  AlertTriangle,
  Copy,
  Check,
  Terminal,
  Shield,
  Key,
  Zap,
  Link2,
  Loader2,
  X,
} from 'lucide-react';
import type { FunctionInfo, ManifestProperty } from './types';

interface ApiReferenceSectionProps {
  functionInfo: FunctionInfo;
}

interface TryItPanelProps {
  functionInfo: FunctionInfo;
  onClose: () => void;
}

interface PropertyRowProps {
  keyName: string;
  property: ManifestProperty;
  requiredFields: string[];
  nestLevel?: number;
  onToggleExpand?: (key: string) => void;
  expandedKeys?: Set<string>;
}

function PropertyRow({ keyName, property, requiredFields, nestLevel = 0, onToggleExpand, expandedKeys }: PropertyRowProps) {
  const isRequired = requiredFields.includes(keyName);
  const hasNested = property.type === 'object' && property.properties;
  const isExpanded = expandedKeys?.has(`${nestLevel}-${keyName}`);
  const hasDefault = property.default !== undefined;

  return (
    <>
      <tr className="border-b border-border-subtle/50 last:border-0">
        <td className="px-4 py-2.5">
          <div className="flex items-center gap-2">
            {hasNested && (
              <button
                onClick={() => onToggleExpand?.(`${nestLevel}-${keyName}`)}
                className="flex-shrink-0 p-0.5 hover:bg-bg-tertiary rounded"
              >
                {isExpanded ? (
                  <ChevronDown className="h-3 w-3 text-text-muted" />
                ) : (
                  <ChevronRight className="h-3 w-3 text-text-muted" />
                )}
              </button>
            )}
            <div className="flex items-center gap-2">
              {nestLevel > 0 && (
                <span className="text-text-muted text-xs">
                  {'  '.repeat(nestLevel)}
                </span>
              )}
              <span className="font-mono text-xs text-foreground">{keyName}</span>
              {isRequired && (
                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/10 text-red-500">
                  required
                </span>
              )}
            </div>
          </div>
        </td>
        <td className="px-4 py-2.5">
          <span className="text-muted-foreground text-sm">
            {property.type === 'object' && property.properties ? (
              <span className="text-amber-500">object</span>
            ) : property.type === 'array' ? (
              <span className="text-purple-500">array</span>
            ) : (
              property.type ?? '—'
            )}
          </span>
        </td>
        <td className="px-4 py-2.5 text-muted-foreground text-sm">
          {property.description ?? '—'}
        </td>
        <td className="px-4 py-2.5">
          {hasDefault ? (
            <code className="text-xs bg-bg-tertiary px-1.5 py-0.5 rounded font-mono">
              {typeof property.default === 'object'
                ? JSON.stringify(property.default)
                : String(property.default)}
            </code>
          ) : (
            <span className="text-text-muted text-xs">—</span>
          )}
        </td>
      </tr>
      {isExpanded && hasNested && property.properties && (
        <tr className="bg-bg-tertiary/30">
          <td colSpan={4} className="px-4 py-2">
            <div className="rounded border border-border-subtle overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-bg-tertiary/70">
                    <th className="px-3 py-1.5 font-medium text-foreground text-xs text-left">Property</th>
                    <th className="px-3 py-1.5 font-medium text-foreground text-xs text-left">Type</th>
                    <th className="px-3 py-1.5 font-medium text-foreground text-xs text-left">Description</th>
                    <th className="px-3 py-1.5 font-medium text-foreground text-xs text-left">Default</th>
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(property.properties).map(([nestedKey, nestedProp]: [string, ManifestProperty]) => (
                    <PropertyRow
                      key={nestedKey}
                      keyName={nestedKey}
                      property={nestedProp}
                      requiredFields={property.required ?? []}
                      nestLevel={nestLevel + 1}
                      onToggleExpand={onToggleExpand}
                      expandedKeys={expandedKeys}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

function TryItPanel({ functionInfo, onClose }: TryItPanelProps) {
  const [inputValue, setInputValue] = useState('');
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<{ success: boolean; output?: string; error?: string } | null>(null);

  const initialValue = functionInfo.input_example
    ? typeof functionInfo.input_example === 'string'
      ? functionInfo.input_example
      : JSON.stringify(functionInfo.input_example, null, 2)
    : functionInfo.manifest?.input?.example
    ? typeof functionInfo.manifest.input.example === 'string'
      ? functionInfo.manifest.input.example
      : JSON.stringify(functionInfo.manifest.input.example, null, 2)
    : functionInfo.manifest?.input?.properties
    ? JSON.stringify(
        Object.fromEntries(
          Object.entries(functionInfo.manifest.input.properties).map(([k, v]) => [
            k,
            v.default ?? v.example ?? '',
          ])
        ),
        null,
        2
      )
    : '';

  useState(() => {
    setInputValue(initialValue);
  });

  const handleRun = useCallback(async () => {
    setIsRunning(true);
    setResult(null);

    try {
      const response = await fetch(`/v1/run/${functionInfo.author}/${functionInfo.name}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: inputValue,
      });

      if (response.ok) {
        const data = await response.json();
        setResult({
          success: true,
          output: typeof data === 'object' ? JSON.stringify(data, null, 2) : String(data),
        });
      } else {
        const errorText = await response.text();
        setResult({
          success: false,
          error: `${response.status} ${response.statusText}: ${errorText}`,
        });
      }
    } catch (err) {
      setResult({
        success: false,
        error: err instanceof Error ? err.message : 'Failed to execute function',
      });
    } finally {
      setIsRunning(false);
    }
  }, [functionInfo.author, functionInfo.name, inputValue]);

  return (
    <motion.div
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      exit={{ opacity: 0, height: 0 }}
      transition={{ duration: 0.2 }}
      className="api-try-it-panel"
    >
      <div className="api-try-it-header">
        <div className="flex items-center gap-2">
          <Play className="h-4 w-4" />
          <span className="font-medium text-sm">Try It</span>
        </div>
        <button onClick={onClose} className="p-1 hover:bg-bg-tertiary rounded">
          <X className="h-4 w-4" />
        </button>
      </div>
      <div className="api-try-it-content">
        <Textarea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder="Enter JSON input..."
          className="font-mono text-sm"
          rows={8}
        />
        <div className="flex items-center gap-3 mt-3">
          <Button onClick={handleRun} disabled={isRunning} size="sm" className="gap-2">
            {isRunning ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Running...
              </>
            ) : (
              <>
                <Play className="h-4 w-4" />
                Run Function
              </>
            )}
          </Button>
          <span className="text-xs text-text-muted">
            POST /v1/run/{functionInfo.author}/{functionInfo.name}
          </span>
        </div>
        {result && (
          <div className="mt-4">
            {result.success ? (
              <div>
                <p className="text-xs font-medium text-green-500 mb-2">Output</p>
                <CodeBlock code={result.output ?? ''} language="json" showLineNumbers maxHeight="200px" />
              </div>
            ) : (
              <Alert className="border-red-500/20 bg-red-500/10">
                <AlertDescription className="text-red-600 dark:text-red-400 text-sm">
                  {result.error}
                </AlertDescription>
              </Alert>
            )}
          </div>
        )}
      </div>
    </motion.div>
  );
}

function generateCurlExample(functionInfo: FunctionInfo): string {
  const endpoint = `/v1/run/${functionInfo.author}/${functionInfo.name}`;
  const authHeader = functionInfo.manifest?.auth ? 'YOUR_API_KEY' : 'YOUR_KEY';
  const inputExample = functionInfo.input_example
    ? typeof functionInfo.input_example === 'string'
      ? functionInfo.input_example
      : JSON.stringify(functionInfo.input_example, null, 2)
    : functionInfo.manifest?.input?.example
    ? JSON.stringify(functionInfo.manifest.input.example, null, 2)
    : '{}';

  return `curl -X POST ${endpoint} \\
  -H "Authorization: Bearer ${authHeader}" \\
  -H "Content-Type: application/json" \\
  -d '${inputExample}'`;
}

export function ApiReferenceSection({ functionInfo }: ApiReferenceSectionProps) {
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());
  const [activeSection, setActiveSection] = useState<string>('doc-overview');
  const [showTryIt, setShowTryIt] = useState(false);
  const [curlCopied, setCurlCopied] = useState(false);

  const toggleExpanded = useCallback((key: string) => {
    setExpandedKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }, []);

  useEffect(() => {
    const handleScroll = () => {
      const sections = ['doc-overview', 'doc-input', 'doc-output', 'doc-examples'];
      for (const id of sections) {
        const el = document.getElementById(id);
        if (el) {
          const rect = el.getBoundingClientRect();
          if (rect.top <= 150 && rect.bottom >= 150) {
            setActiveSection(id);
            return;
          }
        }
      }
    };

    const contentEl = document.querySelector('.function-page-api-content');
    contentEl?.addEventListener('scroll', handleScroll);
    return () => contentEl?.removeEventListener('scroll', handleScroll);
  }, []);

  const handleCopyCurl = async () => {
    await navigator.clipboard.writeText(generateCurlExample(functionInfo));
    setCurlCopied(true);
    setTimeout(() => setCurlCopied(false), 2000);
  };

  const hasInput = functionInfo.manifest?.input || functionInfo.input_example != null;
  const hasOutput = functionInfo.manifest?.output || functionInfo.output_example != null;
  const isDeprecated = functionInfo.manifest?.deprecated === true;
  const inputRequired = functionInfo.manifest?.input?.required ?? [];
  const outputRequired = functionInfo.manifest?.output?.required ?? [];

  return (
    <motion.div
      id="function-api-reference"
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
      className="function-page-api-reference"
    >
      {isDeprecated && (
        <div className="api-deprecation-banner">
          <AlertTriangle className="h-4 w-4" />
          <span>This function is deprecated.</span>
          {functionInfo.manifest?.successor && (
            <span className="text-text-secondary">
              Successor: <a href={`/registry/${functionInfo.manifest.successor}`} className="underline">{functionInfo.manifest.successor}</a>
            </span>
          )}
        </div>
      )}

      <div className="function-page-api-header">
        <div className="flex items-center gap-3">
          <div className="function-page-api-header-icon">
            <BookOpen className="h-5 w-5" />
          </div>
          <div>
            <h2 className="function-page-api-header-title">API Reference</h2>
            <p className="function-page-api-header-subtitle">
              Auto-generated from the function manifest
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="gap-2"
            onClick={() => setShowTryIt(!showTryIt)}
          >
            <Play className="h-3.5 w-3.5" />
            Try It
          </Button>
          <Button variant="outline" size="sm" className="gap-2" asChild>
            <a
              href={`/v1/docs/${functionInfo.author}/${functionInfo.name}`}
              target="_blank"
              rel="noopener noreferrer"
            >
              <ExternalLink className="h-3.5 w-3.5" />
              Full docs
            </a>
          </Button>
        </div>
      </div>

      <AnimatePresence>{showTryIt && <TryItPanel functionInfo={functionInfo} onClose={() => setShowTryIt(false)} />}</AnimatePresence>

      <div className="function-page-api-layout">
        <div className="function-page-api-nav">
          <nav className="function-page-api-nav-list" aria-label="Documentation sections">
            <a
              href="#doc-overview"
              className={`function-page-api-nav-link ${activeSection === 'doc-overview' ? 'function-page-api-nav-link--active' : ''}`}
            >
              Overview
            </a>
            {hasInput && (
              <a
                href="#doc-input"
                className={`function-page-api-nav-link ${activeSection === 'doc-input' ? 'function-page-api-nav-link--active' : ''}`}
              >
                Input
              </a>
            )}
            {hasOutput && (
              <a
                href="#doc-output"
                className={`function-page-api-nav-link ${activeSection === 'doc-output' ? 'function-page-api-nav-link--active' : ''}`}
              >
                Output
              </a>
            )}
            <a
              href="#doc-examples"
              className={`function-page-api-nav-link ${activeSection === 'doc-examples' ? 'function-page-api-nav-link--active' : ''}`}
            >
              Examples
            </a>
          </nav>
        </div>

        <ScrollArea className="function-page-api-content">
          <div className="function-page-api-section">
            <section id="doc-overview" className="scroll-mt-6">
              <div className="function-page-api-card">
                <h3 className="function-page-api-section-title">
                  <span className="function-page-api-section-number">1</span>
                  Overview
                </h3>
                <p className="function-page-api-description">
                  {functionInfo.description || 'No description provided.'}
                </p>
                <div className="function-page-api-meta-grid">
                  <div className="function-page-api-meta-item">
                    <p className="function-page-api-meta-label">Runtime</p>
                    <p className="function-page-api-meta-value">{functionInfo.runtime}</p>
                  </div>
                  {functionInfo.manifest?.deterministic != null && (
                    <div className="function-page-api-meta-item">
                      <p className="function-page-api-meta-label">Determinism</p>
                      <p className="function-page-api-meta-value">
                        {functionInfo.manifest.deterministic ? 'Deterministic' : 'Non-deterministic'}
                      </p>
                    </div>
                  )}
                  {functionInfo.cache_ttl != null && functionInfo.cache_ttl > 0 && (
                    <div className="function-page-api-meta-item">
                      <p className="function-page-api-meta-label">Cache TTL</p>
                      <p className="function-page-api-meta-value">{functionInfo.cache_ttl}s</p>
                    </div>
                  )}
                </div>

                <div className="function-page-api-endpoint">
                  <div className="flex items-center gap-2 mb-3">
                    <Link2 className="h-4 w-4 text-text-muted" />
                    <span className="text-sm font-medium">Endpoint</span>
                  </div>
                  <div className="api-endpoint-url">
                    <Badge variant="outline" className="font-mono text-xs">POST</Badge>
                    <code className="text-sm font-mono text-foreground">
                      /v1/run/{functionInfo.author}/{functionInfo.name}
                    </code>
                  </div>
                  <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-text-muted">
                    {functionInfo.manifest?.auth && (
                      <div className="flex items-center gap-1">
                        <Key className="h-3.5 w-3.5" />
                        <span>{functionInfo.manifest.auth}</span>
                      </div>
                    )}
                    {functionInfo.manifest?.rate_limit && (
                      <div className="flex items-center gap-1">
                        <Zap className="h-3.5 w-3.5" />
                        <span>{functionInfo.manifest.rate_limit} req/min</span>
                      </div>
                    )}
                    {functionInfo.trust_level && (
                      <div className="flex items-center gap-1">
                        <Shield className="h-3.5 w-3.5" />
                        <span className="capitalize">{functionInfo.trust_level}</span>
                      </div>
                    )}
                  </div>
                </div>

                {functionInfo.verified && (
                  <div className="api-verified-badge">
                    <Shield className="h-4 w-4 text-green-500" />
                    <span>FXCert Verified</span>
                    {functionInfo.trust_score && (
                      <span className="text-text-secondary">
                        Trust Score: {functionInfo.trust_score}%
                      </span>
                    )}
                  </div>
                )}
              </div>
            </section>

            {hasInput && (
              <section id="doc-input" className="scroll-mt-6 pt-6">
                <div className="function-page-api-card">
                  <h3 className="function-page-api-section-title">
                    <span className="function-page-api-section-number function-page-api-section-number--input">2</span>
                    Input
                  </h3>
                  {functionInfo.manifest?.input?.properties &&
                    typeof functionInfo.manifest.input.properties === 'object' && (
                      <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                        <table className="function-page-api-table">
                          <thead>
                            <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                              <th className="px-4 py-2.5 font-medium text-foreground">Property</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Type</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Description</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Default</th>
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(functionInfo.manifest.input.properties).map(
                              ([key, val]: [string, ManifestProperty]) => (
                                <PropertyRow
                                  key={key}
                                  keyName={key}
                                  property={val}
                                  requiredFields={inputRequired}
                                  onToggleExpand={toggleExpanded}
                                  expandedKeys={expandedKeys}
                                />
                              )
                            )}
                          </tbody>
                        </table>
                      </div>
                    )}
                  <Tabs defaultValue="schema" className="w-full">
                    <TabsList className="mb-2 h-9">
                      <TabsTrigger value="schema" className="text-xs">Schema</TabsTrigger>
                      {functionInfo.input_example != null && (
                        <TabsTrigger value="example" className="text-xs">Example</TabsTrigger>
                      )}
                      <TabsTrigger value="curl" className="text-xs">cURL</TabsTrigger>
                    </TabsList>
                    <TabsContent value="schema" className="mt-0">
                      {functionInfo.manifest?.input ? (
                        <CodeBlock code={JSON.stringify(functionInfo.manifest.input, null, 2)} language="json" />
                      ) : (
                        <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">
                          No schema defined.
                        </p>
                      )}
                    </TabsContent>
                    {functionInfo.input_example != null && (
                      <TabsContent value="example" className="mt-0">
                        <CodeBlock
                          code={
                            typeof functionInfo.input_example === 'string'
                              ? functionInfo.input_example
                              : JSON.stringify(functionInfo.input_example, null, 2)
                          }
                          language="json"
                        />
                      </TabsContent>
                    )}
                    <TabsContent value="curl" className="mt-0">
                      <div className="rounded-lg border border-border-subtle overflow-hidden">
                        <div className="flex items-center justify-between px-4 py-2 bg-bg-tertiary border-b border-border-subtle">
                          <div className="flex items-center gap-2">
                            <Terminal className="h-4 w-4 text-text-muted" />
                            <span className="text-xs font-medium text-text-secondary">cURL</span>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={handleCopyCurl}
                            className="h-7 gap-1.5 text-xs text-text-muted hover:text-text-primary"
                          >
                            {curlCopied ? (
                              <>
                                <Check className="h-3.5 w-3.5 text-green-500" />
                                <span className="text-green-500">Copied!</span>
                              </>
                            ) : (
                              <>
                                <Copy className="h-3.5 w-3.5" />
                                Copy
                              </>
                            )}
                          </Button>
                        </div>
                        <div className="bg-[#1e1e1e] p-4 overflow-x-auto">
                          <pre className="text-sm font-mono text-[#d4d4d4] whitespace-pre-wrap">
                            {generateCurlExample(functionInfo)}
                          </pre>
                        </div>
                      </div>
                    </TabsContent>
                  </Tabs>
                </div>
              </section>
            )}

            {hasOutput && (
              <section id="doc-output" className="scroll-mt-6 pt-6">
                <div className="function-page-api-card">
                  <h3 className="function-page-api-section-title">
                    <span className="function-page-api-section-number function-page-api-section-number--output">3</span>
                    Output
                  </h3>
                  {functionInfo.manifest?.output?.properties &&
                    typeof functionInfo.manifest.output.properties === 'object' && (
                      <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                        <table className="function-page-api-table">
                          <thead>
                            <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                              <th className="px-4 py-2.5 font-medium text-foreground">Property</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Type</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Description</th>
                              <th className="px-4 py-2.5 font-medium text-foreground">Default</th>
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(functionInfo.manifest.output.properties).map(
                              ([key, val]: [string, ManifestProperty]) => (
                                <PropertyRow
                                  key={key}
                                  keyName={key}
                                  property={val}
                                  requiredFields={outputRequired}
                                  onToggleExpand={toggleExpanded}
                                  expandedKeys={expandedKeys}
                                />
                              )
                            )}
                          </tbody>
                        </table>
                      </div>
                    )}
                  <Tabs defaultValue="schema" className="w-full">
                    <TabsList className="mb-2 h-9">
                      <TabsTrigger value="schema" className="text-xs">Schema</TabsTrigger>
                      {functionInfo.output_example != null && (
                        <TabsTrigger value="example" className="text-xs">Example</TabsTrigger>
                      )}
                    </TabsList>
                    <TabsContent value="schema" className="mt-0">
                      {functionInfo.manifest?.output ? (
                        <CodeBlock code={JSON.stringify(functionInfo.manifest.output, null, 2)} language="json" />
                      ) : (
                        <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">
                          No schema defined.
                        </p>
                      )}
                    </TabsContent>
                    {functionInfo.output_example != null && (
                      <TabsContent value="example" className="mt-0">
                        <CodeBlock
                          code={
                            typeof functionInfo.output_example === 'string'
                              ? functionInfo.output_example
                              : JSON.stringify(functionInfo.output_example, null, 2)
                          }
                          language="json"
                        />
                      </TabsContent>
                    )}
                  </Tabs>
                </div>
              </section>
            )}

            <section id="doc-examples" className="scroll-mt-6 pt-6">
              <div className="function-page-api-card">
                <h3 className="function-page-api-section-title">
                  <span className="function-page-api-section-number" style={{ background: 'rgba(var(--ff-cyan-rgb), 0.15)', color: 'var(--ff-cyan)' }}>4</span>
                  Code Examples
                </h3>
                <div className="space-y-6">
                  <div>
                    <h4 className="text-sm font-medium mb-2">JavaScript</h4>
                    <CodeBlock
                      code={`const response = await fetch('/v1/run/${functionInfo.author}/${functionInfo.name}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_API_KEY'
  },
  body: JSON.stringify(${JSON.stringify(functionInfo.input_example ?? functionInfo.manifest?.input?.example ?? {}, null, 2)})
});

const result = await response.json();
console.log(result);`}
                      language="javascript"
                      showLineNumbers
                    />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">Python</h4>
                    <CodeBlock
                      code={`import requests

response = requests.post(
    '/v1/run/${functionInfo.author}/${functionInfo.name}',
    json=${JSON.stringify(functionInfo.input_example ?? functionInfo.manifest?.input?.example ?? {}, null, 2)},
    headers={'Authorization': 'Bearer YOUR_API_KEY'}
)

result = response.json()
print(result)`}
                      language="python"
                      showLineNumbers
                    />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">cURL</h4>
                    <CodeBlock
                      code={generateCurlExample(functionInfo)}
                      language="bash"
                      showLineNumbers
                    />
                  </div>
                </div>
              </div>
            </section>
          </div>
        </ScrollArea>
      </div>
    </motion.div>
  );
}
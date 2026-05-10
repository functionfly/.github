import { CodeBlock } from '@/components/common/CodeBlock';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { motion } from 'framer-motion';
import { BookOpen, ExternalLink } from 'lucide-react';
import type { FunctionInfo } from './types';

interface ApiReferenceSectionProps {
  functionInfo: FunctionInfo;
}

export function ApiReferenceSection({ functionInfo }: ApiReferenceSectionProps) {
  return (
    <motion.div
      id="function-api-reference"
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
      className="function-page-api-reference"
    >
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

      <div className="function-page-api-layout">
        <div className="function-page-api-nav">
          <nav className="function-page-api-nav-list" aria-label="Documentation sections">
            <a href="#doc-overview" className="function-page-api-nav-link">
              Overview
            </a>
            {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
              <a href="#doc-input" className="function-page-api-nav-link">
                Input
              </a>
            )}
            {(functionInfo.manifest?.output || functionInfo.output_example != null) && (
              <a href="#doc-output" className="function-page-api-nav-link">
                Output
              </a>
            )}
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
              </div>
            </section>

            {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
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
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(functionInfo.manifest.input.properties).map(
                              ([key, val]: [string, unknown]) => {
                                const v = val as { type?: string; description?: string; default?: unknown };
                                return (
                                  <tr key={key} className="border-b border-border-subtle/50 last:border-0">
                                    <td className="px-4 py-2 font-mono text-xs text-foreground">{key}</td>
                                    <td className="px-4 py-2 text-muted-foreground">{v?.type ?? '—'}</td>
                                    <td className="px-4 py-2 text-muted-foreground">{v?.description ?? '—'}</td>
                                  </tr>
                                );
                              }
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
                  </Tabs>
                </div>
              </section>
            )}

            {(functionInfo.manifest?.output || functionInfo.output_example != null) && (
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
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(functionInfo.manifest.output.properties).map(
                              ([key, val]: [string, unknown]) => {
                                const v = val as { type?: string; description?: string };
                                return (
                                  <tr key={key} className="border-b border-border-subtle/50 last:border-0">
                                    <td className="px-4 py-2 font-mono text-xs text-foreground">{key}</td>
                                    <td className="px-4 py-2 text-muted-foreground">{v?.type ?? '—'}</td>
                                    <td className="px-4 py-2 text-muted-foreground">{v?.description ?? '—'}</td>
                                  </tr>
                                );
                              }
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
          </div>
        </ScrollArea>
      </div>
    </motion.div>
  );
}

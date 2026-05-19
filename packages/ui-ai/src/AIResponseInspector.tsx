/**
 * @functionfly/ui-ai
 * AI Response Inspector - Inspect and analyze AI responses
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Eye, ChevronDown, ChevronRight, Copy, Download, CheckCircle2, AlertTriangle } from "lucide-react";

export interface ResponseMetadata {
  model?: string;
  tokensUsed?: number;
  promptTokens?: number;
  completionTokens?: number;
  finishReason?: string;
  latency?: number;
  cost?: number;
}

export interface AIResponseInspectorProps {
  content: string;
  metadata?: ResponseMetadata;
  reasoning?: string;
  onContentCopy?: () => void;
  onContentExport?: () => void;
  className?: string;
}

export function AIResponseInspector({
  content,
  metadata,
  reasoning,
  onContentCopy,
  onContentExport,
  className,
}: AIResponseInspectorProps) {
  const [expandedSections, setExpandedSections] = React.useState<Set<string>>(new Set(["content"]));

  const toggleSection = (section: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev);
      if (next.has(section)) next.delete(section);
      else next.add(section);
      return next;
    });
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Eye className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">Response Inspector</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={onContentCopy}
            className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
          >
            <Copy className="size-4" />
          </button>
          <button
            onClick={onContentExport}
            className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
          >
            <Download className="size-4" />
          </button>
        </div>
      </div>

      {/* Metadata Summary */}
      {metadata && (
        <div className="flex items-center gap-3 px-4 py-2 border-b border-border-subtle bg-bg-tertiary/30">
          {metadata.model && (
            <Badge variant="outline" size="sm">{metadata.model}</Badge>
          )}
          {metadata.tokensUsed && (
            <span className="text-xs text-text-muted">{metadata.tokensUsed} tokens</span>
          )}
          {metadata.latency && (
            <span className="text-xs text-text-muted">{metadata.latency}ms</span>
          )}
          {metadata.cost !== undefined && (
            <span className="text-xs text-text-muted">${metadata.cost.toFixed(4)}</span>
          )}
          {metadata.finishReason && (
            <Badge variant="success" size="sm">
              <CheckCircle2 className="size-3 mr-1" />
              {metadata.finishReason}
            </Badge>
          )}
        </div>
      )}

      {/* Content Sections */}
      <div className="flex-1 overflow-y-auto">
        {/* Content Section */}
        <div className="border-b border-border-subtle">
          <button
            onClick={() => toggleSection("content")}
            className="w-full flex items-center gap-2 px-4 py-2 hover:bg-bg-hover transition-colors"
          >
            {expandedSections.has("content") ? (
              <ChevronDown className="size-4 text-text-muted" />
            ) : (
              <ChevronRight className="size-4 text-text-muted" />
            )}
            <span className="text-sm font-medium text-text-primary">Response</span>
            <Badge variant="ghost" size="sm" className="ml-auto">{content.length} chars</Badge>
          </button>
          
          {expandedSections.has("content") && (
            <div className="px-4 pb-4">
              <pre className="p-4 bg-bg-tertiary/50 rounded-lg text-sm text-text-secondary whitespace-pre-wrap font-mono border border-border-subtle">
                {content}
              </pre>
            </div>
          )}
        </div>

        {/* Reasoning Section */}
        {reasoning && (
          <div className="border-b border-border-subtle">
            <button
              onClick={() => toggleSection("reasoning")}
              className="w-full flex items-center gap-2 px-4 py-2 hover:bg-bg-hover transition-colors"
            >
              {expandedSections.has("reasoning") ? (
                <ChevronDown className="size-4 text-text-muted" />
              ) : (
                <ChevronRight className="size-4 text-text-muted" />
              )}
              <span className="text-sm font-medium text-text-primary">Reasoning</span>
              <Badge variant="brand" size="sm" className="ml-auto">
                <AlertTriangle className="size-3 mr-1" />
                Chain of Thought
              </Badge>
            </button>
            
            {expandedSections.has("reasoning") && (
              <div className="px-4 pb-4">
                <pre className="p-4 bg-bg-tertiary/50 rounded-lg text-sm text-text-secondary whitespace-pre-wrap border border-border-subtle">
                  {reasoning}
                </pre>
              </div>
            )}
          </div>
        )}

        {/* Token Breakdown */}
        {metadata && (
          <div className="border-b border-border-subtle">
            <button
              onClick={() => toggleSection("tokens")}
              className="w-full flex items-center gap-2 px-4 py-2 hover:bg-bg-hover transition-colors"
            >
              {expandedSections.has("tokens") ? (
                <ChevronDown className="size-4 text-text-muted" />
              ) : (
                <ChevronRight className="size-4 text-text-muted" />
              )}
              <span className="text-sm font-medium text-text-primary">Token Breakdown</span>
            </button>
            
            {expandedSections.has("tokens") && (
              <div className="px-4 pb-4 space-y-2">
                <div className="flex items-center justify-between p-2 bg-bg-tertiary/50 rounded">
                  <span className="text-xs text-text-muted">Prompt Tokens</span>
                  <span className="text-sm font-mono text-text-primary">{metadata.promptTokens || 0}</span>
                </div>
                <div className="flex items-center justify-between p-2 bg-bg-tertiary/50 rounded">
                  <span className="text-xs text-text-muted">Completion Tokens</span>
                  <span className="text-sm font-mono text-text-primary">{metadata.completionTokens || 0}</span>
                </div>
                <div className="flex items-center justify-between p-2 bg-brand-500/5 rounded border border-brand-500/20">
                  <span className="text-xs text-text-primary font-medium">Total</span>
                  <span className="text-sm font-mono text-brand-500">{metadata.tokensUsed || 0}</span>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

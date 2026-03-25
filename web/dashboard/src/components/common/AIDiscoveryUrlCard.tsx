import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { WELL_KNOWN_DISCOVERY_URL } from '@/lib/api-urls';
import { getPublicDocsSiteOrigin } from '@/lib/constants';
import { Bot, Check, Copy, ExternalLink } from 'lucide-react';
import { useState } from 'react';

/**
 * Card that shows the AI/LLM discovery endpoint URL with copy button.
 * Use on marketplace or docs-related pages so developers can find the well-known URL.
 */
export function AIDiscoveryUrlCard() {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(WELL_KNOWN_DISCOVERY_URL);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore
    }
  };

  return (
    <Card className="border-theme bg-card/80 ai-discovery-card">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <Bot className="h-5 w-5 text-brand-500" aria-hidden />
          <CardTitle className="text-base">AI & LLM discovery</CardTitle>
        </div>
        <CardDescription>
          Public manifest for LLMs and agents. Fetch this URL to get OpenAI-compatible tool schemas
          for all public functions.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="ai-discovery-url-row flex items-center gap-2 rounded-lg bg-bg-secondary px-3 py-2 font-mono text-sm text-text-secondary break-all">
          <code className="flex-1 min-w-0">{WELL_KNOWN_DISCOVERY_URL}</code>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="shrink-0 h-8 w-8"
            onClick={handleCopy}
            aria-label={copied ? 'Copied' : 'Copy URL'}
          >
            {copied ? (
              <Check className="h-4 w-4 text-green-500" aria-hidden />
            ) : (
              <Copy className="h-4 w-4" aria-hidden />
            )}
          </Button>
        </div>
        <a
          href={`${getPublicDocsSiteOrigin()}/docs/agent-ai-llm-integration`}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 text-sm text-brand-500 hover:text-brand-400 transition-colors"
        >
          <span>Docs: AI & LLM integration</span>
          <ExternalLink className="h-3.5 w-3.5" aria-hidden />
        </a>
      </CardContent>
    </Card>
  );
}

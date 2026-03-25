import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { BookOpen, Check, Copy, ExternalLink, Info, Shield, ShieldCheck, Zap } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

interface TrustPolicyConfig {
  minTrustScore: number;
  requireVerified: boolean;
  requiredTrustLevels: string[];
  capabilitiesDeny: string[];
}

const DEFAULT_POLICY: TrustPolicyConfig = {
  minTrustScore: 80,
  requireVerified: true,
  requiredTrustLevels: ['high'],
  capabilitiesDeny: ['secrets_read'],
};

function policyHash(policy: TrustPolicyConfig): string {
  const normalized = {
    capabilities_deny: [...policy.capabilitiesDeny].sort(),
    min_trust_score: policy.minTrustScore,
    require_verified: policy.requireVerified,
    required_trust_levels: [...policy.requiredTrustLevels].sort(),
  };
  const str = JSON.stringify(normalized);
  // Simple deterministic representation (not a crypto hash in the browser, just for display)
  let h = 0;
  for (let i = 0; i < str.length; i++) {
    h = (Math.imul(31, h) + str.charCodeAt(i)) | 0;
  }
  return Math.abs(h).toString(16).padStart(8, '0');
}

function buildLangChainSnippet(apiBase: string, apiKey: string, policy: TrustPolicyConfig): string {
  const deny = policy.capabilitiesDeny.length
    ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},`
    : '';
  const levels = policy.requiredTrustLevels.length
    ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},`
    : '';
  return `# pip install "flypy[agents-langchain]"
from flypy import AgentClient, LangChainAdapter, TrustPolicy

client = AgentClient(
    api_base="${apiBase}",
    api_key="${apiKey || 'YOUR_API_KEY'}",
)
adapter = LangChainAdapter(client)

policy = TrustPolicy(
    min_trust_score=${policy.minTrustScore},
    require_verified=${policy.requireVerified ? 'True' : 'False'},${levels}${deny}
)

tools = adapter.build_tools(
    policy=policy,
    query="text transform",
    limit=10,
)

# Use with LangChain agent
# agent = initialize_agent(tools, llm, agent=AgentType.STRUCTURED_CHAT_ZERO_SHOT_REACT_DESCRIPTION)`;
}

function buildAutoGenSnippet(apiBase: string, apiKey: string, policy: TrustPolicyConfig): string {
  const deny = policy.capabilitiesDeny.length
    ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},`
    : '';
  const levels = policy.requiredTrustLevels.length
    ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},`
    : '';
  return `# pip install "flypy[agents-autogen]"
from flypy import AgentClient, AutoGenAdapter, TrustPolicy

client = AgentClient(
    api_base="${apiBase}",
    api_key="${apiKey || 'YOUR_API_KEY'}",
)
adapter = AutoGenAdapter(client)

policy = TrustPolicy(
    min_trust_score=${policy.minTrustScore},
    require_verified=${policy.requireVerified ? 'True' : 'False'},${levels}${deny}
)

tools = adapter.build_tools(policy=policy, query="json", limit=5)

# Each tool has: name, description, parameters, function, metadata
# Register with AutoGen:
# for tool in tools:
#     user_proxy.register_function({tool["name"]: tool["function"]})`;
}

function buildCrewAISnippet(apiBase: string, apiKey: string, policy: TrustPolicyConfig): string {
  const deny = policy.capabilitiesDeny.length
    ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},`
    : '';
  const levels = policy.requiredTrustLevels.length
    ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},`
    : '';
  return `# pip install "flypy[agents-crewai]"
from flypy import AgentClient, CrewAIAdapter, TrustPolicy

client = AgentClient(
    api_base="${apiBase}",
    api_key="${apiKey || 'YOUR_API_KEY'}",
)
adapter = CrewAIAdapter(client)

policy = TrustPolicy(
    min_trust_score=${policy.minTrustScore},
    require_verified=${policy.requireVerified ? 'True' : 'False'},${levels}${deny}
)

tools = adapter.build_tools(policy=policy, query="text", limit=5)

# Each tool has: name, description, args_schema, func, metadata
# Use with CrewAI:
# from crewai import Agent, Task, Crew
# agent = Agent(role="analyst", tools=[t["func"] for t in tools], ...)`;
}

function CodeBlock({ code, label }: { code: string; label: string }) {
  const [copied, setCopied] = useState(false);
  const lines = code.split('\n');
  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    toast.success('Copied to clipboard');
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="relative sdk-code-block">
      <div className="sdk-code-header flex items-center justify-between px-3 py-2 bg-muted/50 border border-b-0 rounded-t-md">
        <span className="text-xs font-mono text-muted-foreground">{label}</span>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 px-2 text-xs gap-1 sdk-copy-btn"
          onClick={handleCopy}
        >
          {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
          {copied ? 'Copied' : 'Copy'}
        </Button>
      </div>
      <pre className="sdk-code-pre bg-slate-950 border rounded-b-md p-4 overflow-x-auto text-sm font-mono text-slate-200 leading-relaxed">
        <code className="sdk-code-lines">
          {lines.map((line, index) => (
            <div key={`${label}-${index}`} className="sdk-code-line">
              <span className="sdk-code-line-number">{index + 1}</span>
              <span className="sdk-code-line-content">{line || ' '}</span>
            </div>
          ))}
        </code>
      </pre>
    </div>
  );
}

export function AgentSDKIntegrationsPage() {
  const [policy, setPolicy] = useState<TrustPolicyConfig>(DEFAULT_POLICY);
  const [apiKeyOverride, setApiKeyOverride] = useState('');
  const [toolSlug, setToolSlug] = useState('functionfly/text-transform');
  const [toolVersion, setToolVersion] = useState('latest');
  const [toolInput, setToolInput] = useState('{"text":"hello world"}');
  const apiBase = 'https://api.functionfly.com';
  const displayKey = apiKeyOverride || 'YOUR_API_KEY';
  const trustLevelLabel =
    policy.minTrustScore >= 90
      ? 'Strict'
      : policy.minTrustScore >= 70
        ? 'Balanced'
        : policy.minTrustScore >= 40
          ? 'Permissive'
          : 'Open';

  const hash = policyHash(policy);
  let parsedInput: unknown = {};
  let inputError = '';
  try {
    parsedInput = toolInput.trim() ? JSON.parse(toolInput) : {};
  } catch {
    inputError = 'Input must be valid JSON.';
  }
  const requestPreview = {
    input: parsedInput,
  };
  const curlPreview = `curl -X POST "${apiBase}/v1/fx/${toolSlug}${
    toolVersion.trim() ? `?version=${encodeURIComponent(toolVersion.trim())}` : ''
  }" \\
  -H "Authorization: Bearer ${displayKey}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(requestPreview)}'`;

  const toggleLevel = (level: string) => {
    setPolicy((p) => ({
      ...p,
      requiredTrustLevels: p.requiredTrustLevels.includes(level)
        ? p.requiredTrustLevels.filter((l) => l !== level)
        : [...p.requiredTrustLevels, level],
    }));
  };

  const toggleDenyCap = (cap: string) => {
    setPolicy((p) => ({
      ...p,
      capabilitiesDeny: p.capabilitiesDeny.includes(cap)
        ? p.capabilitiesDeny.filter((c) => c !== cap)
        : [...p.capabilitiesDeny, cap],
    }));
  };

  return (
    <div className="space-y-6 max-w-4xl agent-sdk-integrations-page">
      {/* Header */}
      <div className="sdk-header">
        <div className="flex items-center gap-2 mb-1.5">
          <ShieldCheck className="h-7 w-7 text-violet-500" />
          <h1 className="text-2xl font-bold">SDK Integrations</h1>
          <Badge variant="secondary" className="text-xs">
            Python
          </Badge>
        </div>
        <p className="text-muted-foreground max-w-3xl">
          Connect FunctionFly trusted tools to LangChain, AutoGen, and CrewAI. All adapters are
          fail-closed by default — only verified, high-trust tools are exposed to your agents.
        </p>
      </div>

      <div className="sdk-section-nav sticky top-16 z-20 flex flex-wrap items-center gap-2 rounded-lg border border-border/60 bg-background/90 px-2 py-2 backdrop-blur">
        <a href="#policy" className="sdk-section-link">
          Policy
        </a>
        <a href="#snippets" className="sdk-section-link">
          Snippets
        </a>
        <a href="#payload" className="sdk-section-link">
          Payload
        </a>
        <a href="#security" className="sdk-section-link">
          Security
        </a>
      </div>

      {/* Install banner */}
      <Card className="border-violet-500/30 bg-violet-500/5 sdk-install-banner">
        <CardContent className="pt-4 pb-4">
          <div className="flex flex-col sm:flex-row sm:items-center gap-3">
            <div className="flex-1">
              <p className="text-sm font-medium mb-1">Install the SDK</p>
              <code className="text-xs font-mono text-muted-foreground">
                pip install "flypy[agents]"
              </code>
              <span className="text-xs text-muted-foreground ml-3">
                or per-framework: <code className="font-mono">flypy[agents-langchain]</code>
              </span>
            </div>
            <Button
              variant="default"
              size="sm"
              className="gap-1.5 shrink-0 sdk-quickstart-btn"
              asChild
            >
              <a
                href="https://docs.functionfly.com/sdk/agent-integrations"
                target="_blank"
                rel="noopener noreferrer"
              >
                <BookOpen className="h-3.5 w-3.5" />
                Quickstart Guide
                <ExternalLink className="h-3 w-3 text-muted-foreground" />
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Trust Policy Builder */}
        <Card id="policy" className="lg:col-span-1 sdk-policy-card scroll-mt-28">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Shield className="h-4 w-4 text-violet-500" />
              Trust Policy
            </CardTitle>
            <CardDescription className="text-xs">
              Configure which tools agents can use. Code snippets update live.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            {/* API Key override */}
            <div className="space-y-1.5 sdk-control-group">
              <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold">
                Auth
              </p>
              <Label className="text-xs">API Key (optional preview)</Label>
              <Input
                className="font-mono text-xs h-8"
                placeholder="ff_sk_…"
                value={apiKeyOverride}
                onChange={(e) => setApiKeyOverride(e.target.value)}
              />
              <p className="text-[11px] text-muted-foreground">
                Not stored. Used only to prefill code snippets.
              </p>
            </div>

            {/* Min trust score */}
            <div className="space-y-1.5 sdk-control-group">
              <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold">
                Scoring
              </p>
              <Label className="text-xs">
                Min Trust Score: <span className="font-mono font-bold">{policy.minTrustScore}</span>
                <Badge variant="secondary" className="ml-2 text-[10px] align-middle">
                  {trustLevelLabel}
                </Badge>
              </Label>
              <input
                type="range"
                min={0}
                max={100}
                step={5}
                value={policy.minTrustScore}
                onChange={(e) =>
                  setPolicy((p) => ({ ...p, minTrustScore: Number(e.target.value) }))
                }
                className="w-full accent-violet-500"
              />
              <div className="sdk-slider-ticks">
                {[0, 50, 80, 100].map((tick) => (
                  <span
                    key={tick}
                    className={`sdk-slider-tick ${tick === 80 ? 'sdk-slider-tick-default' : ''}`}
                  >
                    {tick}
                  </span>
                ))}
              </div>
              <div className="flex justify-between text-[10px] text-muted-foreground">
                <span>0 (open)</span>
                <span>80 (default)</span>
                <span>100</span>
              </div>
            </div>

            {/* Require verified */}
            <div className="flex items-center justify-between sdk-control-group">
              <div>
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-1">
                  Verification
                </p>
                <Label className="text-xs">Require Verified</Label>
                <p className="text-[11px] text-muted-foreground">Only FXCert-verified tools</p>
              </div>
              <Switch
                checked={policy.requireVerified}
                onCheckedChange={(v) => setPolicy((p) => ({ ...p, requireVerified: v }))}
              />
            </div>

            {/* Trust levels */}
            <div className="space-y-1.5 sdk-control-group">
              <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold">
                Restrictions
              </p>
              <Label className="text-xs">Required Trust Levels</Label>
              <div className="flex flex-wrap gap-1.5">
                {['high', 'medium', 'low'].map((level) => (
                  <button
                    key={level}
                    onClick={() => toggleLevel(level)}
                    className={`px-2.5 py-1 rounded-full text-xs font-medium border transition-colors ${
                      policy.requiredTrustLevels.includes(level)
                        ? 'bg-violet-500 border-violet-500 text-white'
                        : 'border-border text-muted-foreground hover:border-violet-400'
                    }`}
                  >
                    {level}
                  </button>
                ))}
              </div>
            </div>

            {/* Deny capabilities */}
            <div className="space-y-1.5 sdk-control-group">
              <Label className="text-xs">Deny Capabilities</Label>
              <div className="flex flex-wrap gap-1.5">
                {['secrets_read', 'filesystem', 'network', 'exec'].map((cap) => (
                  <button
                    key={cap}
                    onClick={() => toggleDenyCap(cap)}
                    className={`px-2 py-0.5 rounded text-[11px] font-mono border transition-colors ${
                      policy.capabilitiesDeny.includes(cap)
                        ? 'bg-red-500/20 border-red-500 text-red-400'
                        : 'border-border text-muted-foreground hover:border-red-400'
                    }`}
                  >
                    {cap}
                  </button>
                ))}
              </div>
            </div>

            {/* Policy hash */}
            <div className="rounded-md bg-muted/50 border px-3 py-2">
              <div className="flex items-center gap-1.5 mb-0.5">
                <Info className="h-3 w-3 text-muted-foreground" />
                <span className="text-[11px] text-muted-foreground font-medium">Policy Hash</span>
              </div>
              <code className="text-xs font-mono text-violet-400">{hash}</code>
              <p className="text-[10px] text-muted-foreground mt-0.5">
                Included in every execution envelope for audit.
              </p>
            </div>

            {!policy.requireVerified && (
              <div className="flex items-start gap-2 rounded-md bg-amber-500/10 border border-amber-500/30 px-3 py-2 text-xs text-amber-600 dark:text-amber-400">
                <Zap className="h-3.5 w-3.5 shrink-0 mt-0.5" />
                <span>
                  Verified enforcement is disabled. Unverified tools may be exposed to your agent.
                </span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Code snippets */}
        <div id="snippets" className="lg:col-span-2 space-y-4 scroll-mt-28">
          <Tabs defaultValue="langchain">
            <TabsList className="grid w-full grid-cols-3 sdk-tabs-list">
              <TabsTrigger value="langchain">LangChain</TabsTrigger>
              <TabsTrigger value="autogen">AutoGen</TabsTrigger>
              <TabsTrigger value="crewai">CrewAI</TabsTrigger>
            </TabsList>

            <TabsContent value="langchain" className="space-y-3 pt-2">
              <div className="flex items-center gap-2">
                <Badge className="bg-green-700/30 text-green-400 border-green-700/50 text-xs sdk-available-badge">
                  available
                </Badge>
                <span className="text-xs text-muted-foreground">
                  flypy[agents-langchain] · langchain-core≥0.2
                </span>
              </div>
              <CodeBlock
                label="langchain_trusted_tools.py"
                code={buildLangChainSnippet(apiBase, displayKey, policy)}
              />
              <p className="text-xs text-muted-foreground">
                Returns <code className="font-mono">StructuredTool</code> instances when{' '}
                <code className="font-mono">langchain-core</code> is installed, or plain dict
                descriptors as fallback. Execution metadata includes{' '}
                <code className="font-mono">policy_hash</code> for auditability.
              </p>
            </TabsContent>

            <TabsContent value="autogen" className="space-y-3 pt-2">
              <div className="flex items-center gap-2">
                <Badge className="bg-green-700/30 text-green-400 border-green-700/50 text-xs sdk-available-badge">
                  available
                </Badge>
                <span className="text-xs text-muted-foreground">
                  flypy[agents-autogen] · pyautogen≥0.2
                </span>
              </div>
              <CodeBlock
                label="autogen_trusted_tools.py"
                code={buildAutoGenSnippet(apiBase, displayKey, policy)}
              />
              <p className="text-xs text-muted-foreground">
                Returns dicts with <code className="font-mono">name</code>,{' '}
                <code className="font-mono">description</code>,{' '}
                <code className="font-mono">parameters</code>, and a{' '}
                <code className="font-mono">function</code> callable. Register with{' '}
                <code className="font-mono">user_proxy.register_function()</code>.
              </p>
            </TabsContent>

            <TabsContent value="crewai" className="space-y-3 pt-2">
              <div className="flex items-center gap-2">
                <Badge className="bg-green-700/30 text-green-400 border-green-700/50 text-xs sdk-available-badge">
                  available
                </Badge>
                <span className="text-xs text-muted-foreground">
                  flypy[agents-crewai] · crewai≥0.70
                </span>
              </div>
              <CodeBlock
                label="crewai_trusted_tools.py"
                code={buildCrewAISnippet(apiBase, displayKey, policy)}
              />
              <p className="text-xs text-muted-foreground">
                Returns dicts with <code className="font-mono">name</code>,{' '}
                <code className="font-mono">args_schema</code>, and a{' '}
                <code className="font-mono">func</code> callable. Trust enforcement happens at
                discovery, not execution.
              </p>
            </TabsContent>
          </Tabs>

          {/* Execution payload preview */}
          <Card id="payload" className="border-muted sdk-preview-card scroll-mt-28">
            <CardHeader className="pb-2 pt-4">
              <CardTitle className="text-sm flex items-center gap-2">
                <Info className="h-4 w-4 text-blue-500" />
                Execution Payload Preview
              </CardTitle>
              <CardDescription className="text-xs">
                Validate JSON input and preview the exact execute request shape.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 pb-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label className="text-xs">Function Slug</Label>
                  <Input
                    value={toolSlug}
                    onChange={(e) => setToolSlug(e.target.value)}
                    className="font-mono text-xs h-8"
                    placeholder="author/name"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs">Version</Label>
                  <Input
                    value={toolVersion}
                    onChange={(e) => setToolVersion(e.target.value)}
                    className="font-mono text-xs h-8"
                    placeholder="latest"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <Label className="text-xs">Tool Input (JSON)</Label>
                <Textarea
                  value={toolInput}
                  onChange={(e) => setToolInput(e.target.value)}
                  className="font-mono text-xs min-h-[110px]"
                  placeholder='{"text":"hello world"}'
                />
                {inputError ? (
                  <p className="text-[11px] text-red-500">{inputError}</p>
                ) : (
                  <p className="text-[11px] text-muted-foreground">
                    JSON parsed successfully. Policy hash for this session:{' '}
                    <code className="font-mono">{hash}</code>
                  </p>
                )}
              </div>

              {!inputError && (
                <>
                  <CodeBlock
                    label="request_body.json"
                    code={JSON.stringify(requestPreview, null, 2)}
                  />
                  <CodeBlock label="curl_execute.sh" code={curlPreview} />
                </>
              )}
            </CardContent>
          </Card>

          {/* Security notes */}
          <Card id="security" className="border-muted sdk-security-card scroll-mt-28">
            <CardHeader className="pb-2 pt-4">
              <CardTitle className="text-sm flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 text-emerald-500" />
                Production Security Notes
              </CardTitle>
            </CardHeader>
            <CardContent className="pb-4">
              <ul className="space-y-1.5 text-xs text-muted-foreground">
                <li className="flex items-start gap-2">
                  <Check className="h-3.5 w-3.5 text-emerald-500 mt-0.5 shrink-0" />
                  API keys must be stored in environment variables, never in source code.
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-3.5 w-3.5 text-emerald-500 mt-0.5 shrink-0" />
                  <code className="font-mono">AgentClient</code> enforces HTTPS for all
                  non-localhost hosts.
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-3.5 w-3.5 text-emerald-500 mt-0.5 shrink-0" />
                  Trust policy is evaluated at <em>discovery time</em>. Disallowed tools are never
                  returned to the agent.
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-3.5 w-3.5 text-emerald-500 mt-0.5 shrink-0" />
                  Every execution envelope includes <code className="font-mono">
                    policy_hash
                  </code>, <code className="font-mono">tool_id</code>, and{' '}
                  <code className="font-mono">version</code> for end-to-end auditability.
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-3.5 w-3.5 text-emerald-500 mt-0.5 shrink-0" />
                  Set <code className="font-mono">timeout_seconds</code> and{' '}
                  <code className="font-mono">max_retries</code> explicitly for your production
                  environment.
                </li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export default AgentSDKIntegrationsPage;

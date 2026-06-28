import { BookOpen, Check, Copy, ExternalLink, Info, Shield, ShieldCheck, Zap } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  AnnotationTag,
} from '@/components/containment';
import './styles.css';

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
  let h = 0;
  for (let i = 0; i < str.length; i++) { h = (Math.imul(31, h) + str.charCodeAt(i)) | 0; }
  return Math.abs(h).toString(16).padStart(8, '0');
}

function buildLangChainSnippet(apiBase: string, apiKey: string, policy: TrustPolicyConfig): string {
  const deny = policy.capabilitiesDeny.length ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},` : '';
  const levels = policy.requiredTrustLevels.length ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},` : '';
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
  const deny = policy.capabilitiesDeny.length ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},` : '';
  const levels = policy.requiredTrustLevels.length ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},` : '';
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
  const deny = policy.capabilitiesDeny.length ? `\n    capabilities_deny=${JSON.stringify(policy.capabilitiesDeny)},` : '';
  const levels = policy.requiredTrustLevels.length ? `\n    required_trust_levels=${JSON.stringify(policy.requiredTrustLevels)},` : '';
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
    <div className="sdk-code-block">
      <div className="sdk-code-header">
        <span className="sdk-code-label">{label}</span>
        <button className="sdk-copy-btn" onClick={handleCopy}>
          {copied ? <Check className="sdk-icon-xs sdk-icon-ok" /> : <Copy className="sdk-icon-xs" />}
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      <pre className="sdk-code-pre">
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

function ToggleButton({ active, onClick, label, variant = 'default' }: { active: boolean; onClick: () => void; label: string; variant?: 'default' | 'danger' }) {
  const cls = variant === 'danger'
    ? `sdk-toggle-btn sdk-toggle-btn--danger ${active ? 'sdk-toggle-btn--active-danger' : ''}`
    : `sdk-toggle-btn ${active ? 'sdk-toggle-btn--active' : ''}`;
  return <button className={cls} onClick={onClick}>{label}</button>;
}

export function AgentSDKIntegrationsPage() {
  const [policy, setPolicy] = useState<TrustPolicyConfig>(DEFAULT_POLICY);
  const [apiKeyOverride, setApiKeyOverride] = useState('');
  const [toolSlug, setToolSlug] = useState('functionfly/text-transform');
  const [toolVersion, setToolVersion] = useState('latest');
  const [toolInput, setToolInput] = useState('{"text":"hello world"}');
  const apiBase = 'https://api.functionfly.com';
  const displayKey = apiKeyOverride || 'YOUR_API_KEY';
  const trustLevelLabel = policy.minTrustScore >= 90 ? 'Strict' : policy.minTrustScore >= 70 ? 'Balanced' : policy.minTrustScore >= 40 ? 'Permissive' : 'Open';

  const hash = policyHash(policy);
  let parsedInput: unknown = {};
  let inputError = '';
  try { parsedInput = toolInput.trim() ? JSON.parse(toolInput) : {}; } catch { inputError = 'Input must be valid JSON.'; }
  const requestPreview = { input: parsedInput };
  const curlPreview = `curl -X POST "${apiBase}/v1/fx/${toolSlug}${toolVersion.trim() ? `?version=${encodeURIComponent(toolVersion.trim())}` : ''}" \\
  -H "Authorization: Bearer ${displayKey}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(requestPreview)}'`;

  const toggleLevel = (level: string) => {
    setPolicy((p) => ({ ...p, requiredTrustLevels: p.requiredTrustLevels.includes(level) ? p.requiredTrustLevels.filter((l) => l !== level) : [...p.requiredTrustLevels, level] }));
  };

  const toggleDenyCap = (cap: string) => {
    setPolicy((p) => ({ ...p, capabilitiesDeny: p.capabilitiesDeny.includes(cap) ? p.capabilitiesDeny.filter((c) => c !== cap) : [...p.capabilitiesDeny, cap] }));
  };

  const [activeTab, setActiveTab] = useState<'langchain' | 'autogen' | 'crewai'>('langchain');

  return (
    <div className="sdk-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="sdk-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE SDK-01" secondary="SDK Integrations" position="top-right" />

        <div className="sdk-hero__header">
          <div className="sdk-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="sdk-hero__title">SDK Integrations</h1>
            <StatusPill status="live" label="Python" />
          </div>
          <p className="sdk-hero__subtitle">
            Connect FunctionFly trusted tools to LangChain, AutoGen, and CrewAI. All adapters are fail-closed by default — only verified, high-trust tools are exposed to your agents.
          </p>
        </div>
      </Chamber>

      {/* Section Nav */}
      <nav className="sdk-section-nav">
        <a href="#policy" className="sdk-section-link">Policy</a>
        <a href="#snippets" className="sdk-section-link">Snippets</a>
        <a href="#payload" className="sdk-section-link">Payload</a>
        <a href="#security" className="sdk-section-link">Security</a>
      </nav>

      {/* Install Banner */}
      <Chamber className="sdk-install-banner">
        <div className="sdk-install__content">
          <div className="sdk-install__info">
            <p className="sdk-install__title">Install the SDK</p>
            <code className="sdk-install__code">pip install "flypy[agents]"</code>
            <span className="sdk-install__hint">or per-framework: <code>flypy[agents-langchain]</code></span>
          </div>
          <a href="https://docs.functionfly.com/sdk/agent-integrations" target="_blank" rel="noopener noreferrer">
            <SealedButton size="sm" iconLeft={<BookOpen className="sdk-icon-sm" />} iconRight={<ExternalLink className="sdk-icon-xs" />}>
              Quickstart Guide
            </SealedButton>
          </a>
        </div>
      </Chamber>

      {/* Main Grid */}
      <div className="sdk-main-grid">
        {/* Policy Builder */}
        <Chamber id="policy" className="sdk-policy-card">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="sdk-policy__header">
            <Shield className="sdk-icon-sm sdk-icon-accent" />
            <h2 className="sdk-policy__title">Trust Policy</h2>
          </div>
          <p className="sdk-policy__desc">Configure which tools agents can use. Code snippets update live.</p>

          <div className="sdk-controls">
            {/* API Key */}
            <div className="sdk-control-group">
              <span className="sdk-control-label">Auth</span>
              <label className="sdk-field-label">API Key (optional preview)</label>
              <input className="sdk-input" placeholder="ff_sk_…" value={apiKeyOverride} onChange={(e) => setApiKeyOverride(e.target.value)} />
              <p className="sdk-field-hint">Not stored. Used only to prefill code snippets.</p>
            </div>

            {/* Min trust score */}
            <div className="sdk-control-group">
              <span className="sdk-control-label">Scoring</span>
              <label className="sdk-field-label">Min Trust Score: <span className="sdk-field-value">{policy.minTrustScore}</span> <span className="sdk-trust-badge">{trustLevelLabel}</span></label>
              <input type="range" min={0} max={100} step={5} value={policy.minTrustScore} onChange={(e) => setPolicy((p) => ({ ...p, minTrustScore: Number(e.target.value) }))} className="sdk-range" />
              <div className="sdk-slider-ticks">
                {[0, 50, 80, 100].map((tick) => (
                  <span key={tick} className={`sdk-slider-tick ${tick === 80 ? 'sdk-slider-tick--default' : ''}`}>{tick}</span>
                ))}
              </div>
              <div className="sdk-range-labels">
                <span>0 (open)</span>
                <span>80 (default)</span>
                <span>100</span>
              </div>
            </div>

            {/* Require verified */}
            <div className="sdk-control-group sdk-control-group--row">
              <div>
                <span className="sdk-control-label">Verification</span>
                <label className="sdk-field-label">Require Verified</label>
                <p className="sdk-field-hint">Only FXCert-verified tools</p>
              </div>
              <button
                className={`sdk-switch ${policy.requireVerified ? 'sdk-switch--on' : ''}`}
                onClick={() => setPolicy((p) => ({ ...p, requireVerified: !p.requireVerified }))}
              >
                <span className="sdk-switch__thumb" />
              </button>
            </div>

            {/* Trust levels */}
            <div className="sdk-control-group">
              <span className="sdk-control-label">Restrictions</span>
              <label className="sdk-field-label">Required Trust Levels</label>
              <div className="sdk-toggle-group">
                {['high', 'medium', 'low'].map((level) => (
                  <ToggleButton key={level} active={policy.requiredTrustLevels.includes(level)} onClick={() => toggleLevel(level)} label={level} />
                ))}
              </div>
            </div>

            {/* Deny capabilities */}
            <div className="sdk-control-group">
              <label className="sdk-field-label">Deny Capabilities</label>
              <div className="sdk-toggle-group">
                {['secrets_read', 'filesystem', 'network', 'exec'].map((cap) => (
                  <ToggleButton key={cap} active={policy.capabilitiesDeny.includes(cap)} onClick={() => toggleDenyCap(cap)} label={cap} variant="danger" />
                ))}
              </div>
            </div>

            {/* Policy hash */}
            <div className="sdk-hash-box">
              <div className="sdk-hash-box__header">
                <Info className="sdk-icon-xs" />
                <span className="sdk-hash-box__label">Policy Hash</span>
              </div>
              <code className="sdk-hash-box__value">{hash}</code>
              <p className="sdk-hash-box__hint">Included in every execution envelope for audit.</p>
            </div>

            {!policy.requireVerified && (
              <div className="sdk-alert sdk-alert--warning">
                <Zap className="sdk-icon-xs" />
                <span>Verified enforcement is disabled. Unverified tools may be exposed to your agent.</span>
              </div>
            )}
          </div>
        </Chamber>

        {/* Code Snippets */}
        <div id="snippets" className="sdk-snippets-col">
          <div className="sdk-tabs">
            {(['langchain', 'autogen', 'crewai'] as const).map((tab) => (
              <button key={tab} className={`sdk-tab ${activeTab === tab ? 'sdk-tab--active' : ''}`} onClick={() => setActiveTab(tab)}>
                {tab === 'langchain' ? 'LangChain' : tab === 'autogen' ? 'AutoGen' : 'CrewAI'}
              </button>
            ))}
          </div>

          <div className="sdk-tab-panel">
            <div className="sdk-availability">
              <StatusPill status="live" label="available" />
              <span className="sdk-availability__text">
                {activeTab === 'langchain' ? 'flypy[agents-langchain] · langchain-core≥0.2' : activeTab === 'autogen' ? 'flypy[agents-autogen] · pyautogen≥0.2' : 'flypy[agents-crewai] · crewai≥0.70'}
              </span>
            </div>
            <CodeBlock
              label={`${activeTab}_trusted_tools.py`}
              code={activeTab === 'langchain' ? buildLangChainSnippet(apiBase, displayKey, policy) : activeTab === 'autogen' ? buildAutoGenSnippet(apiBase, displayKey, policy) : buildCrewAISnippet(apiBase, displayKey, policy)}
            />
            <p className="sdk-snippet-hint">
              {activeTab === 'langchain'
                ? 'Returns StructuredTool instances when langchain-core is installed, or plain dict descriptors as fallback. Execution metadata includes policy_hash for auditability.'
                : activeTab === 'autogen'
                  ? 'Returns dicts with name, description, parameters, and a function callable. Register with user_proxy.register_function().'
                  : 'Returns dicts with name, args_schema, and a func callable. Trust enforcement happens at discovery, not execution.'}
            </p>
          </div>

          {/* Execution Payload Preview */}
          <Chamber id="payload" className="sdk-preview-card">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="sdk-preview__header">
              <Info className="sdk-icon-sm sdk-icon-info" />
              <h2 className="sdk-preview__title">Execution Payload Preview</h2>
            </div>
            <p className="sdk-preview__desc">Validate JSON input and preview the exact execute request shape.</p>

            <div className="sdk-preview__fields">
              <div className="sdk-field">
                <label className="sdk-field-label">Function Slug</label>
                <input className="sdk-input" value={toolSlug} onChange={(e) => setToolSlug(e.target.value)} placeholder="author/name" />
              </div>
              <div className="sdk-field">
                <label className="sdk-field-label">Version</label>
                <input className="sdk-input" value={toolVersion} onChange={(e) => setToolVersion(e.target.value)} placeholder="latest" />
              </div>
            </div>

            <div className="sdk-field">
              <label className="sdk-field-label">Tool Input (JSON)</label>
              <textarea className="sdk-textarea" value={toolInput} onChange={(e) => setToolInput(e.target.value)} placeholder='{"text":"hello world"}' />
              {inputError ? (
                <p className="sdk-field-error">{inputError}</p>
              ) : (
                <p className="sdk-field-hint">JSON parsed successfully. Policy hash for this session: <code>{hash}</code></p>
              )}
            </div>

            {!inputError && (
              <>
                <CodeBlock label="request_body.json" code={JSON.stringify(requestPreview, null, 2)} />
                <CodeBlock label="curl_execute.sh" code={curlPreview} />
              </>
            )}
          </Chamber>

          {/* Security Notes */}
          <Chamber id="security" className="sdk-security-card">
            <CornerBrace position="tr" />
            <CornerBrace position="bl" />
            <div className="sdk-security__header">
              <ShieldCheck className="sdk-icon-sm sdk-icon-ok" />
              <h2 className="sdk-security__title">Production Security Notes</h2>
            </div>
            <ul className="sdk-security__list">
              <li><Check className="sdk-icon-xs sdk-icon-ok" /> API keys must be stored in environment variables, never in source code.</li>
              <li><Check className="sdk-icon-xs sdk-icon-ok" /> <code>AgentClient</code> enforces HTTPS for all non-localhost hosts.</li>
              <li><Check className="sdk-icon-xs sdk-icon-ok" /> Trust policy is evaluated at <em>discovery time</em>. Disallowed tools are never returned to the agent.</li>
              <li><Check className="sdk-icon-xs sdk-icon-ok" /> Every execution envelope includes <code>policy_hash</code>, <code>tool_id</code>, and <code>version</code> for end-to-end auditability.</li>
              <li><Check className="sdk-icon-xs sdk-icon-ok" /> Set <code>timeout_seconds</code> and <code>max_retries</code> explicitly for your production environment.</li>
            </ul>
          </Chamber>
        </div>
      </div>
    </div>
  );
}

export default AgentSDKIntegrationsPage;

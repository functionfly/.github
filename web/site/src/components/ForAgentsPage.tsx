import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN, DOCS_ORIGIN } from '../config'
import './homepage.css'

const trustApiUrl = `${DOCS_ORIGIN.replace(/\/$/, '')}/docs/trust-api/`
const trustProtocolUrl = `${DOCS_ORIGIN.replace(/\/$/, '')}/docs/trust-protocol-spec/`

const ForAgentsPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-agents-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Agents & orchestrators</span>
          </div>
          <h1 className="ff-hero-headline">
            Built for agents that need<br />real trust signals
          </h1>
          <p className="ff-hero-sub">
            FunctionFly exposes functions with clear manifests, verification metadata, and execution-backed trust scores—so policies can distinguish "callable" from "appropriate to call."
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Start free</a>
            <a className="ff-btn ff-btn-secondary" href="/trust">Trust layer</a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">Pricing</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pillars">
            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M11 19a7 7 0 1 0 0-14 7 7 0 0 0 0 14Zm8-2 3 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M8 11h6M11 8v6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Discovery & manifests</h2>
              <p>
                Tools ship with structured configuration and I/O schemas so agents and orchestrators can match intent to capability without guessing.
              </p>
              <ul className="ff-pillar-list">
                <li>Stable tool identity and versioning</li>
                <li>Schema-first inputs and outputs</li>
                <li>Policy hooks for capability and constraints</li>
              </ul>
            </div>

            <div className="ff-pillar-card ff-pillar-card--accent">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Trust API</h2>
              <p>
                Query attestations, scores, and verification-friendly metadata to enforce budgets and trust tiers at selection time and runtime.
              </p>
              <ul className="ff-pillar-list">
                <li>Batch-friendly score lookups</li>
                <li>Verification and reporting flows</li>
                <li>Usage-based metering for partner workloads</li>
              </ul>
              <p className="ff-pillar-links">
                <a className="ff-link" href={trustApiUrl}>Trust API docs →</a>
                <a className="ff-link" href={trustProtocolUrl}>Trust protocol spec →</a>
              </p>
            </div>

            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                  <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Human + platform verification</h2>
              <p>
                Verification levels stack from format checks through platform-signed tooling. See the full model and what agents actually consume at runtime.
              </p>
              <ul className="ff-pillar-list">
                <li>Multi-level verification ladder</li>
                <li>Signing, attestations, and revocation posture</li>
                <li>Execution-backed score components</li>
              </ul>
              <p className="ff-pillar-links">
                <a className="ff-link" href="/trust">How trust works →</a>
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">From discovery to a safe tool call</h2>
          <p className="ff-section-lead">
            Orchestrators can treat FunctionFly as the source of truth for "what this tool is" and "whether it is allowed right now."
          </p>
          <div className="ff-flow-grid">
            <div className="ff-flow-step">
              <span className="ff-flow-k">1</span>
              <div>
                <div className="ff-flow-t">Discover</div>
                <div className="ff-flow-d">List candidate tools from the registry with manifests and declared capabilities.</div>
              </div>
            </div>
            <div className="ff-flow-step">
              <span className="ff-flow-k">2</span>
              <div>
                <div className="ff-flow-t">Score</div>
                <div className="ff-flow-d">Call the Trust API for trust tier, verification level, and components that match your policy.</div>
              </div>
            </div>
            <div className="ff-flow-step">
              <span className="ff-flow-k">3</span>
              <div>
                <div className="ff-flow-t">Decide</div>
                <div className="ff-flow-d">Apply budgets, tiers, and org rules before the model is allowed to bind a tool.</div>
              </div>
            </div>
            <div className="ff-flow-step">
              <span className="ff-flow-k">4</span>
              <div>
                <div className="ff-flow-t">Execute</div>
                <div className="ff-flow-d">Run the function with observability; scores update from real execution history.</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-split-grid">
            <div className="ff-split-card">
              <h2>Agent execution</h2>
              <p className="ff-muted">
                Dedicated agent plans cover tool-call volume, concurrency, and guardrails when your agents outgrow ad-hoc usage. Align spend with how often tools actually run.
              </p>
              <ul className="ff-checklist">
                <li><span className="ff-tick" aria-hidden="true">✓</span> Tool-call quotas and burst limits</li>
                <li><span className="ff-tick" aria-hidden="true">✓</span> Per-agent cost visibility</li>
                <li><span className="ff-tick" aria-hidden="true">✓</span> Enterprise options for custom caps</li>
              </ul>
              <a className="ff-btn ff-btn-outline" href="/pricing">See agent pricing →</a>
            </div>
            <div className="ff-split-card">
              <h2 id="state-fabric-agents">Stateful tools with State Fabric</h2>
              <p className="ff-muted">
                Long-running or multi-step agents often need durable state, not just stateless HTTP calls. State Fabric adds structured state, snapshots, and replay-oriented workflows on top of your functions.
              </p>
              <ul className="ff-checklist">
                <li><span className="ff-tick" aria-hidden="true">✓</span> Optional add-ons for cache, security, and AI memory</li>
                <li><span className="ff-tick" aria-hidden="true">✓</span> Tiers from sandbox through enterprise</li>
                <li><span className="ff-tick" aria-hidden="true">✓</span> Billed alongside your workspace plan</li>
              </ul>
              <a className="ff-btn ff-btn-outline" href="/pricing#state-fabric">State Fabric pricing →</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Trust API at a glance</h2>
          <p className="ff-section-lead">
            Authenticate with a partner API key, then read scores or drive verification. Full request and response shapes live in the docs.
          </p>
          <div className="ff-endpoint-grid">
            <div className="ff-endpoint-card">
              <div className="ff-endpoint-method">GET</div>
              <code className="ff-endpoint-path">/v1/trust/score/<span className="ff-param">function_id</span></code>
              <p className="ff-endpoint-desc">Current trust score, tier, verification level, and component breakdown.</p>
            </div>
            <div className="ff-endpoint-card">
              <div className="ff-endpoint-method">POST</div>
              <code className="ff-endpoint-path">/v1/trust/batch</code>
              <p className="ff-endpoint-desc">Batch score lookup for tool shortlists and ranking.</p>
            </div>
            <div className="ff-endpoint-card">
              <div className="ff-endpoint-method">GET</div>
              <code className="ff-endpoint-path">/v1/trust/history/<span className="ff-param">function_id</span></code>
              <p className="ff-endpoint-desc">Historical windows for drift detection and policy audits.</p>
            </div>
            <div className="ff-endpoint-card">
              <div className="ff-endpoint-method">POST</div>
              <code className="ff-endpoint-path">/v1/trust/verify</code>
              <p className="ff-endpoint-desc">Submit a verification request (scoped key required).</p>
            </div>
          </div>
          <p className="ff-docs-cta">
            <a className="ff-link" href={trustApiUrl}>Read the Trust API guide →</a>
          </p>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container ff-container--narrow">
          <h2 className="ff-section-title">FAQ</h2>
          <dl className="ff-faq">
            <div className="ff-faq-item">
              <dt>Do agents need FunctionFly accounts?</dt>
              <dd>
                Your product integrates with FunctionFly using API keys and workspace configuration. End-user agents typically talk to your backend, which enforces policy and calls FunctionFly as needed.
              </dd>
            </div>
            <div className="ff-faq-item">
              <dt>How does this relate to the zero-knowledge vault?</dt>
              <dd>
                Tools can use secrets without the platform ever seeing plaintext: encryption stays client-side. That pairs well with agent toolchains that must not leak credentials into prompts or logs.
              </dd>
            </div>
            <div className="ff-faq-item">
              <dt>Where do I start?</dt>
              <dd>
                Create your workspace to publish or connect functions, then wire the Trust API into your tool-selection path. Use <a className="ff-inline-link" href="/trust">Trust layer</a> for the verification model and <a className="ff-inline-link" href="/pricing">Pricing</a> for execution and State Fabric.
              </dd>
            </div>
          </dl>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Ship agents that choose tools with evidence</h2>
          <p>Connect manifests, trust scores, and execution-backed signals in one place—then scale with agent plans and State Fabric when workloads grow.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Start free</a>
            <a className="ff-btn ff-btn-secondary" href="/contact">Talk to us</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default ForAgentsPage

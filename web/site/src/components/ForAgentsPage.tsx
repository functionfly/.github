import React from 'react'
import { AUTH_ORIGIN, DOCS_ORIGIN } from '../config'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { TrustSeal } from './containment/TrustSeal'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'
import { StatusPill } from './containment/StatusPill'

const trustApiUrl = `${DOCS_ORIGIN.replace(/\/$/, '')}/docs/trust-api/`
const trustProtocolUrl = `${DOCS_ORIGIN.replace(/\/$/, '')}/docs/trust-protocol-spec/`

const FFContainer: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({ children, narrow }) => (
  <div style={{ maxWidth: narrow ? 720 : 1100, margin: '0 auto', padding: '0 var(--space-4)' }}>
    {children}
  </div>
)

const Section: React.FC<{ children: React.ReactNode; id?: string }> = ({ children, id }) => (
  <section id={id} style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>
    {children}
  </section>
)

const SectionTitle: React.FC<{ children: React.ReactNode; lead?: React.ReactNode }> = ({ children, lead }) => (
  <div style={{ textAlign: 'center', marginBottom: 'var(--space-7)' }}>
    <h2 style={{
      fontFamily: 'var(--font-display)',
      fontSize: '36px',
      fontWeight: 700,
      letterSpacing: '-0.005em',
      lineHeight: 1.15,
      color: 'var(--text)',
      marginBottom: lead ? 'var(--space-4)' : 0,
    }}>{children}</h2>
    {lead && <p style={{ color: 'var(--text-dim)', maxWidth: 640, margin: '0 auto', lineHeight: 1.6 }}>{lead}</p>}
  </div>
)

const ForAgentsPage: React.FC = () => {
  return (
    <>
      <PageGrid />

      <Section>
        <FFContainer>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{
              fontFamily: 'var(--font-mono)',
              fontSize: 11,
              letterSpacing: '0.08em',
              textTransform: 'uppercase',
              color: 'var(--status-ok)',
              marginBottom: 'var(--space-5)',
              display: 'inline-flex',
              alignItems: 'center',
              gap: 'var(--space-2)',
            }}>
              <span style={{
                width: 8, height: 8, borderRadius: '50%',
                background: 'var(--status-ok)',
                boxShadow: '0 0 12px var(--status-ok), 0 0 24px rgba(143, 255, 208, 0.4)',
              }} />
              Agents & orchestrators
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)',
              fontSize: '58px',
              fontWeight: 700,
              letterSpacing: '-0.01em',
              lineHeight: 1.08,
              color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Built for agents that need<br />real trust signals
            </h1>
            <p style={{
              fontFamily: 'var(--font-body)',
              fontSize: 17,
              lineHeight: 1.6,
              color: 'var(--text-dim)',
              maxWidth: 720,
              marginBottom: 'var(--space-6)',
            }}>
              FunctionFly exposes functions with clear manifests, verification metadata, and execution-backed trust scores—so policies can distinguish "callable" from "appropriate to call."
            </p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>
                Start free
              </SealedButton>
              <FrameButton onClick={() => { window.location.href = '/trust' }}>
                Trust layer
              </FrameButton>
              <FrameButton onClick={() => { window.location.href = '/pricing' }}>
                Pricing
              </FrameButton>
            </div>
            <div style={{ marginTop: 'var(--space-6)' }}>
              <TrustSeal label="Verified manifest" size="default" />
            </div>
          </Chamber>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer>
          <SectionTitle lead="Tools ship with structured configuration, schemas, and attestations—so agents can match intent to capability without guessing.">
            Discovery, trust, and execution
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <div style={{
                width: 48, height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center',
                border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)',
                color: 'var(--status-ok)',
              }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="11" cy="11" r="7" />
                  <path d="m21 21-4.3-4.3" />
                </svg>
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Discovery & manifests
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Tools ship with structured configuration and I/O schemas so agents and orchestrators can match intent to capability without guessing.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0, color: 'var(--text-dim)', fontSize: 14 }}>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Stable tool identity and versioning
                </li>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Schema-first inputs and outputs
                </li>
                <li style={{ paddingLeft: 20, position: 'relative' }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Policy hooks for capability and constraints
                </li>
              </ul>
            </Card>

            <Card>
              <div style={{
                width: 48, height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center',
                border: '1px solid var(--accent)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)',
                color: 'var(--accent)',
              }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" />
                  <path d="m9 12 2 2 4-4" />
                </svg>
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Trust API
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Query attestations, scores, and verification-friendly metadata to enforce budgets and trust tiers at selection time and runtime.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0, color: 'var(--text-dim)', fontSize: 14 }}>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Batch-friendly score lookups
                </li>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Verification and reporting flows
                </li>
                <li style={{ paddingLeft: 20, position: 'relative' }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Usage-based metering for partner workloads
                </li>
              </ul>
              <div style={{ marginTop: 'var(--space-4)', display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
                <a href={trustApiUrl} style={{ color: 'var(--status-ok)', fontSize: 14 }}>Trust API docs →</a>
                <a href={trustProtocolUrl} style={{ color: 'var(--status-ok)', fontSize: 14 }}>Trust protocol spec →</a>
              </div>
            </Card>

            <Card>
              <div style={{
                width: 48, height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center',
                border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)',
                color: 'var(--status-ok)',
              }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
                </svg>
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Human + platform verification
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Verification levels stack from format checks through platform-signed tooling. See the full model and what agents actually consume at runtime.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0, color: 'var(--text-dim)', fontSize: 14 }}>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Multi-level verification ladder
                </li>
                <li style={{ paddingLeft: 20, position: 'relative', marginBottom: 8 }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Signing, attestations, and revocation posture
                </li>
                <li style={{ paddingLeft: 20, position: 'relative' }}>
                  <span style={{ position: 'absolute', left: 0, color: 'var(--status-ok)' }}>→</span>
                  Execution-backed score components
                </li>
              </ul>
              <div style={{ marginTop: 'var(--space-4)' }}>
                <a href="/trust" style={{ color: 'var(--status-ok)', fontSize: 14 }}>How trust works →</a>
              </div>
            </Card>
          </div>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer>
          <SectionTitle lead='Orchestrators can treat FunctionFly as the source of truth for "what this tool is" and "whether it is allowed right now."'>
            From discovery to a safe tool call
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { k: 1, t: 'Discover', d: 'List candidate tools from the registry with manifests and declared capabilities.' },
              { k: 2, t: 'Score', d: 'Call the Trust API for trust tier, verification level, and components that match your policy.' },
              { k: 3, t: 'Decide', d: 'Apply budgets, tiers, and org rules before the model is allowed to bind a tool.' },
              { k: 4, t: 'Execute', d: 'Run the function with observability; scores update from real execution history.' },
            ].map((s) => (
              <Card key={s.k}>
                <div style={{ display: 'flex', gap: 'var(--space-4)', alignItems: 'flex-start' }}>
                  <div style={{
                    flexShrink: 0, width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    border: '1px solid var(--accent)', borderRadius: '50%',
                    fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--accent)',
                  }}>{s.k}</div>
                  <div>
                    <div style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 4 }}>{s.t}</div>
                    <div style={{ fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.5 }}>{s.d}</div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Agent execution
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Dedicated agent plans cover tool-call volume, concurrency, and guardrails when your agents outgrow ad-hoc usage. Align spend with how often tools actually run.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-4) 0' }}>
                {['Tool-call quotas and burst limits', 'Per-agent cost visibility', 'Enterprise options for custom caps'].map((item) => (
                  <li key={item} style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', color: 'var(--text-dim)', fontSize: 14, marginBottom: 8 }}>
                    <span style={{ color: 'var(--status-ok)' }}>✓</span> {item}
                  </li>
                ))}
              </ul>
              <FrameButton onClick={() => { window.location.href = '/pricing' }}>See agent pricing →</FrameButton>
            </Card>
            <Card>
              <h3 id="state-fabric-agents" style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Stateful tools with State Fabric
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Long-running or multi-step agents often need durable state, not just stateless HTTP calls. State Fabric adds structured state, snapshots, and replay-oriented workflows on top of your functions.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-4) 0' }}>
                {['Optional add-ons for cache, security, and AI memory', 'Tiers from sandbox through enterprise', 'Billed alongside your workspace plan'].map((item) => (
                  <li key={item} style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', color: 'var(--text-dim)', fontSize: 14, marginBottom: 8 }}>
                    <span style={{ color: 'var(--status-ok)' }}>✓</span> {item}
                  </li>
                ))}
              </ul>
              <FrameButton onClick={() => { window.location.href = '/pricing#state-fabric' }}>State Fabric pricing →</FrameButton>
            </Card>
          </div>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer>
          <SectionTitle lead="Authenticate with a partner API key, then read scores or drive verification. Full request and response shapes live in the docs.">
            Trust API at a glance
          </SectionTitle>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-4)' }}>
              {[
                { method: 'GET', path: '/v1/trust/score/:function_id', desc: 'Current trust score, tier, verification level, and component breakdown.' },
                { method: 'POST', path: '/v1/trust/batch', desc: 'Batch score lookup for tool shortlists and ranking.' },
                { method: 'GET', path: '/v1/trust/history/:function_id', desc: 'Historical windows for drift detection and policy audits.' },
                { method: 'POST', path: '/v1/trust/verify', desc: 'Submit a verification request (scoped key required).' },
              ].map((ep) => (
                <div key={ep.path} style={{
                  padding: 'var(--space-4)',
                  background: 'var(--bg)',
                  border: '1px solid var(--panel-edge)',
                  borderRadius: 'var(--radius)',
                }}>
                  <div style={{
                    display: 'inline-block',
                    fontFamily: 'var(--font-mono)',
                    fontSize: 11, fontWeight: 700,
                    color: 'var(--status-ok)',
                    background: 'rgba(143, 255, 208, 0.1)',
                    padding: '2px 8px',
                    borderRadius: 'var(--radius-sm)',
                    marginBottom: 'var(--space-2)',
                  }}>{ep.method}</div>
                  <code style={{
                    display: 'block',
                    fontFamily: 'var(--font-mono)',
                    fontSize: 13,
                    color: 'var(--text)',
                    marginBottom: 'var(--space-2)',
                    wordBreak: 'break-all',
                  }}>{ep.path}</code>
                  <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5 }}>{ep.desc}</p>
                </div>
              ))}
            </div>
            <div style={{ textAlign: 'center', marginTop: 'var(--space-6)' }}>
              <a href={trustApiUrl} style={{ color: 'var(--status-ok)', fontSize: 14 }}>Read the Trust API guide →</a>
            </div>
          </Chamber>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer narrow>
          <SectionTitle>FAQ</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <Card>
              <dt style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                Do agents need FunctionFly accounts?
              </dt>
              <dd style={{ color: 'var(--text-dim)', lineHeight: 1.6, fontSize: 14, margin: 0 }}>
                Your product integrates with FunctionFly using API keys and workspace configuration. End-user agents typically talk to your backend, which enforces policy and calls FunctionFly as needed.
              </dd>
            </Card>
            <Card>
              <dt style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                How does this relate to the zero-knowledge vault?
              </dt>
              <dd style={{ color: 'var(--text-dim)', lineHeight: 1.6, fontSize: 14, margin: 0 }}>
                Tools can use secrets without the platform ever seeing plaintext: encryption stays client-side. That pairs well with agent toolchains that must not leak credentials into prompts or logs.
              </dd>
            </Card>
            <Card>
              <dt style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                Where do I start?
              </dt>
              <dd style={{ color: 'var(--text-dim)', lineHeight: 1.6, fontSize: 14, margin: 0 }}>
                Create your workspace to publish or connect functions, then wire the Trust API into your tool-selection path. Use <a href="/trust" style={{ color: 'var(--status-ok)' }}>Trust layer</a> for the verification model and <a href="/pricing" style={{ color: 'var(--status-ok)' }}>Pricing</a> for execution and State Fabric.
              </dd>
            </Card>
          </div>
        </FFContainer>
      </Section>

      <Section>
        <FFContainer>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)',
                fontSize: '36px',
                fontWeight: 700,
                letterSpacing: '-0.005em',
                color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>Ship agents that choose tools with evidence</h2>
              <p style={{ fontSize: 17, color: 'var(--text-dim)', maxWidth: 560, margin: '0 auto var(--space-6)', lineHeight: 1.6 }}>
                Connect manifests, trust scores, and execution-backed signals in one place—then scale with agent plans and State Fabric when workloads grow.
              </p>
              <div style={{ display: 'inline-flex', gap: 'var(--space-3)', flexWrap: 'wrap', justifyContent: 'center' }}>
                <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Start free</SealedButton>
                <FrameButton onClick={() => { window.location.href = '/contact' }}>Talk to us</FrameButton>
              </div>
              <div style={{ marginTop: 'var(--space-6)' }}>
                <StatusPill status="live" label="Trust API live" />
              </div>
            </div>
          </Chamber>
        </FFContainer>
      </Section>
    </>
  )
}

export default ForAgentsPage

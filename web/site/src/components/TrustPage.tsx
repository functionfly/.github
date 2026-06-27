import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { TrustSeal } from './containment/TrustSeal'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'
import { StatusPill } from './containment/StatusPill'
import { CodeBlock } from './containment/CodeBlock'

const Container: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({ children, narrow }) => (
  <div style={{ maxWidth: narrow ? 720 : 1100, margin: '0 auto', padding: '0 var(--space-4)' }}>
    {children}
  </div>
)
const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)
const SectionTitle: React.FC<{ children: React.ReactNode; id?: string; lead?: React.ReactNode }> = ({ children, id, lead }) => (
  <div id={id} style={{ textAlign: 'center', marginBottom: 'var(--space-7)' }}>
    <h2 style={{
      fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
      letterSpacing: '-0.005em', color: 'var(--text)',
      marginBottom: lead ? 'var(--space-4)' : 0,
    }}>{children}</h2>
    {lead && <p style={{ color: 'var(--text-dim)', maxWidth: 640, margin: '0 auto', lineHeight: 1.6 }}>{lead}</p>}
  </div>
)

const TrustPage: React.FC = () => {
  return (
    <>
      <PageGrid />

      <Section>
        <Container>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{
              fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
              textTransform: 'uppercase', color: 'var(--status-ok)',
              marginBottom: 'var(--space-5)',
              display: 'inline-flex', alignItems: 'center', gap: 'var(--space-2)',
            }}>
              <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--status-ok)', boxShadow: '0 0 12px var(--status-ok)' }} />
              Trust, explained
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              The Trust Layer<br />for AI Agents
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              Agents do not need "more tools." They need tools they can trust. FunctionFly turns functions into
              verified, signed, auditable building blocks with execution-backed trust scores.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap', marginBottom: 'var(--space-6)' }}>
              <SealedButton onClick={() => { window.location.hash = 'verification-levels' }}>
                Verification levels
              </SealedButton>
              <FrameButton onClick={() => { window.location.hash = 'trust-api' }}>
                Trust API
              </FrameButton>
            </div>
            <div style={{ display: 'flex', gap: 'var(--space-4)', alignItems: 'center', flexWrap: 'wrap' }}>
              <TrustSeal label="Execution-backed" size="default" />
              <StatusPill status="live" label="L1–L4 verification" />
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(380px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 id="verification-levels" style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Verification levels
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-5)' }}>
                Each level adds deeper assurance before a tool can be considered "trusted."
              </p>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
                {[
                  { k: 'L1', t: 'Format checks', s: 'Validate manifest structure and I/O schema.' },
                  { k: 'L2', t: 'Security scans', s: 'Scan for risky behaviors and validate capability constraints.' },
                  { k: 'L3', t: 'Code review', s: 'Manual and automated review of safety-relevant aspects.' },
                  { k: 'L4', t: 'Platform verified', s: 'Signed and reviewed as official/recommended tooling.' },
                ].map((step) => (
                  <div key={step.k} style={{ display: 'flex', gap: 'var(--space-3)' }}>
                    <div style={{
                      flexShrink: 0, width: 40, height: 40, display: 'flex', alignItems: 'center', justifyContent: 'center',
                      background: 'var(--panel)', border: '1px solid var(--accent)',
                      borderRadius: 'var(--radius)',
                      fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 700, color: 'var(--accent)',
                    }}>{step.k}</div>
                    <div>
                      <div style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 4 }}>{step.t}</div>
                      <div style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5 }}>{step.s}</div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Signing, attestations, revocation
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Trust is portable because verification becomes an immutable record.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-4) 0' }}>
                <li style={{ marginBottom: 8, color: 'var(--text-dim)', fontSize: 14 }}>
                  <strong style={{ color: 'var(--text)' }}>Signing:</strong> Function artifacts include platform signatures.
                </li>
                <li style={{ marginBottom: 8, color: 'var(--text-dim)', fontSize: 14 }}>
                  <strong style={{ color: 'var(--text)' }}>Attestations:</strong> Verifiable records of what was checked and when.
                </li>
                <li style={{ color: 'var(--text-dim)', fontSize: 14 }}>
                  <strong style={{ color: 'var(--text)' }}>Revocation:</strong> If a tool is flagged, it can be downgraded or removed from trusted pools.
                </li>
              </ul>
              <div style={{
                background: 'var(--bg)', border: '1px solid var(--panel-edge)',
                borderRadius: 'var(--radius)', padding: 'var(--space-4)',
              }}>
                <div style={{
                  fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500,
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--status-ok)', marginBottom: 'var(--space-2)',
                }}>What agents see</div>
                <div style={{ fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.5 }}>
                  A trust score and policy-relevant metadata (capabilities, constraints, and verification level).
                </div>
              </div>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { t: 'Execution-backed trust', d: 'Trust scores compound from real execution history so "it ran" and "it ran safely" are both represented.' },
              { t: 'Zero-knowledge vault', d: 'Secrets are encrypted client-side. FunctionFly stores ciphertext only, enabling agent tools without exposing plaintext secrets.' },
              { t: 'Trust API (usage-based)', d: 'Query attestations, trust scores, and revocation state. Bill on usage so agent workloads can adapt trust dynamically.' },
            ].map((c) => (
              <Card key={c.t}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                  {c.t}
                </h3>
                <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, fontSize: 14 }}>{c.d}</p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle id="trust-lifecycle" lead="Trust is not a static label. It starts at publish-time, evolves with execution history, and can be updated through verification, reports, and revocations.">
            Trust score lifecycle
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { k: 1, t: 'Publish', d: 'A function is published with a manifest and a declared I/O contract.' },
              { k: 2, t: 'Verify', d: 'Verification stages validate structure, security, and safety-relevant behavior.' },
              { k: 3, t: 'Sign & attest', d: 'Attestations become immutable records that agents can reason about.' },
              { k: 4, t: 'Execute', d: 'Real execution updates trust score components over time.' },
              { k: 5, t: 'Evolve', d: 'History, reports, and revocation keep trust current for agents and clients.' },
            ].map((s) => (
              <Card key={s.k}>
                <div style={{
                  width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: 'var(--panel)', border: '1px solid var(--accent)',
                  borderRadius: '50%',
                  fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--accent)',
                  marginBottom: 'var(--space-3)',
                }}>{s.k}</div>
                <div style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--text)', marginBottom: 4 }}>{s.t}</div>
                <div style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5 }}>{s.d}</div>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(420px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 id="trust-api" style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Trust API for agents
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Use a scoped API key to read trust scores and trigger verification workflows from your agent runtime.
              </p>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)', marginBottom: 'var(--space-4)' }}>
                {[
                  { k: 'Get trust score', v: 'GET /v1/trust/score/{function_id}' },
                  { k: 'Batch scores', v: 'POST /v1/trust/batch' },
                  { k: 'Get history', v: 'GET /v1/trust/history/{function_id}' },
                  { k: 'Submit verification', v: 'POST /v1/trust/verify' },
                ].map((e) => (
                  <div key={e.v} style={{ display: 'flex', justifyContent: 'space-between', gap: 'var(--space-3)', fontSize: 13 }}>
                    <strong style={{ color: 'var(--text)' }}>{e.k}:</strong>
                    <code style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>{e.v}</code>
                  </div>
                ))}
              </div>
              <div style={{
                background: 'var(--bg)', border: '1px solid var(--panel-edge)',
                borderRadius: 'var(--radius)', padding: 'var(--space-4)',
              }}>
                <div style={{
                  fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500,
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--status-pending)', marginBottom: 'var(--space-2)',
                }}>Scopes matter</div>
                <div style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5 }}>
                  Trust read endpoints require an API key; verification and reporting endpoints require scopes (e.g. <code>verification:request</code>, <code>reports:submit</code>).
                </div>
              </div>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Example: submit verification
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Trigger a verification run for a function version, then query status until the verification completes.
              </p>
              <CodeBlock language="http">
{`POST /v1/trust/verify
Authorization: Bearer API_KEY
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "metadata": {
    "use_case": "data_processing",
    "expected_traffic": "high"
  }
}`}
              </CodeBlock>
              <div style={{ height: 'var(--space-3)' }} />
              <CodeBlock language="json">
{`{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "verification_id": "vfy_abc123def456",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_author": "alice",
  "function_name": "data-processor",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z"
}`}
              </CodeBlock>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle id="partner-tiers" lead="Each tier maps to a verification level, giving your agents the right trust guarantees for the right workload.">
            Partner tiers
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
            {[
              { tier: 'Developer', level: 'L1–L2', price: 'Free', features: ['Format validation (L1)', 'Security scan results (L2)', 'Basic trust scores', '10k API calls/month', 'Community support'], cta: 'Get Started', href: `${APP_DASHBOARD_ORIGIN}/settings#trust-api`, featured: false },
              { tier: 'Startup', level: 'L2–L3', price: '$49', period: '/mo', features: ['Security scans (L2)', 'Code review attestations (L3)', 'Full trust score history', '100k API calls/month', 'Email support'], cta: 'Subscribe', href: `${APP_DASHBOARD_ORIGIN}/settings#trust-api?checkout=startup`, featured: true },
              { tier: 'Business', level: 'L3–L4', price: '$199', period: '/mo', features: ['Code review + attestations (L3)', 'Platform verification (L4)', 'Revocation monitoring', '1M API calls/month', 'Priority support + SLAs'], cta: 'Subscribe', href: `${APP_DASHBOARD_ORIGIN}/settings#trust-api?checkout=business`, featured: false },
              { tier: 'Enterprise', level: 'L4 + Custom', price: 'Custom', period: '', features: ['All Business features', 'Custom verification policies', 'Dedicated trust consultant', 'Unlimited API calls', 'On-premise deployment option'], cta: 'Contact Sales', href: '/contact', featured: false },
            ].map((t) => (
              <Card key={t.tier} style={t.featured ? { borderColor: 'var(--accent)' } : {}}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--space-3)' }}>
                  <div style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)' }}>{t.tier}</div>
                  <div style={{
                    fontFamily: 'var(--font-mono)', fontSize: 11, padding: '2px 8px',
                    border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius-sm)',
                    color: 'var(--status-ok)',
                  }}>{t.level}</div>
                </div>
                <div style={{ fontFamily: 'var(--font-display)', fontSize: 26, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                  {t.price}{t.period && <span style={{ fontSize: 14, color: 'var(--text-dim)' }}>{t.period}</span>}
                </div>
                <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-4) 0' }}>
                  {t.features.map((f) => (
                    <li key={f} style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', color: 'var(--text-dim)', fontSize: 13, marginBottom: 6 }}>
                      <span style={{ color: 'var(--status-ok)' }}>✓</span> {f}
                    </li>
                  ))}
                </ul>
                {t.featured
                  ? <SealedButton onClick={() => { window.location.href = t.href }}>{t.cta}</SealedButton>
                  : <FrameButton onClick={() => { window.location.href = t.href }}>{t.cta}</FrameButton>
                }
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle lead="No other platform combines verification, trust scoring, and zero-knowledge security.">
            Why FunctionFly wins on trust
          </SectionTitle>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="table" style={{ overflowX: 'auto' }}>
              <div className="table__header" style={{
                display: 'grid', gridTemplateColumns: '2fr repeat(4, 1fr)',
                borderBottom: '1px solid var(--panel-edge)',
              }}>
                <div className="table__th">Feature</div>
                <div className="table__th" style={{ color: 'var(--status-ok)' }}>FunctionFly</div>
                <div className="table__th">RapidAPI</div>
                <div className="table__th">Toolhouse</div>
                <div className="table__th">Theagora</div>
              </div>
              {[
                ['Trust scoring', ['check', 'dash', 'dash', 'partial']],
                ['Multi-level verification', ['check', 'dash', 'dash', 'check']],
                ['Zero-knowledge vault', ['check', 'dash', 'dash', 'dash']],
                ['Function execution', ['check', 'dash', 'dash', 'check']],
                ['Agent-native tool discovery', ['check', 'dash', 'check', 'check']],
              ].map(([feature, marks]) => (
                <div key={feature as string} className="table__tr" style={{
                  display: 'grid', gridTemplateColumns: '2fr repeat(4, 1fr)',
                  borderBottom: '1px solid var(--panel-edge)',
                }}>
                  <div className="table__td" style={{ color: 'var(--text)' }}>{feature as string}</div>
                  {(marks as string[]).map((m, i) => (
                    <div key={i} className="table__td" style={{
                      color: i === 0 ? 'var(--status-ok)' : m === 'check' ? 'var(--status-ok)' : 'var(--text-dim)',
                    }}>
                      {m === 'check' && <span>✓</span>}
                      {m === 'dash' && <span>—</span>}
                      {m === 'partial' && <span style={{ fontSize: 11 }}>Partial</span>}
                    </div>
                  ))}
                </div>
              ))}
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
                letterSpacing: '-0.005em', color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>Ready to build with trusted tools?</h2>
              <p style={{ fontSize: 17, color: 'var(--text-dim)', maxWidth: 560, margin: '0 auto var(--space-6)', lineHeight: 1.6 }}>
                Browse trusted functions in the marketplace, or run your own agent trust policy to select tools safely.
              </p>
              <div style={{ display: 'inline-flex', gap: 'var(--space-3)', flexWrap: 'wrap', justifyContent: 'center' }}>
                <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Get Started Free</SealedButton>
                <FrameButton onClick={() => { window.location.href = '/pricing' }}>See trust-aligned pricing</FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default TrustPage

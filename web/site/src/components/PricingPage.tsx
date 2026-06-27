import React, { useState } from 'react'
import { AUTH_ORIGIN } from '../config'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { TrustSeal } from './containment/TrustSeal'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'
import { StatusPill } from './containment/StatusPill'

const Container: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div style={{ maxWidth: 1180, margin: '0 auto', padding: '0 var(--space-4)' }}>{children}</div>
)
const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)
const SectionTitle: React.FC<{ children: React.ReactNode; lead?: React.ReactNode }> = ({ children, lead }) => (
  <div style={{ textAlign: 'center', marginBottom: 'var(--space-7)' }}>
    <h2 style={{
      fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
      letterSpacing: '-0.005em', color: 'var(--text)',
      marginBottom: lead ? 'var(--space-4)' : 0,
    }}>{children}</h2>
    {lead && <p style={{ color: 'var(--text-dim)', maxWidth: 720, margin: '0 auto', lineHeight: 1.6 }}>{lead}</p>}
  </div>
)

interface Tier {
  name: string
  monthly: string
  annual: string
  desc: string
  features: string[]
  cta: string
  href: string
  featured?: boolean
  badge?: string
}

const platformTiers: Tier[] = [
  {
    name: 'Free', monthly: '$0', annual: '$0',
    desc: 'For side projects and experimentation.',
    features: ['25K requests / month', '3 AI agents, 10K calls/mo', '3 published functions', 'L1 verification', 'Community support', '7-day execution logs', '1 State Fabric object'],
    cta: 'Get started', href: `${AUTH_ORIGIN}/signup`,
  },
  {
    name: 'Starter', monthly: '$24', annual: '$19',
    desc: 'For side projects and MVPs.',
    features: ['500K requests / month', '10 AI agents, 100K calls/mo', '10 published functions', 'L1–L2 verification', 'Email support', '30-day execution logs', 'Zero-knowledge Vault (50 secrets)', '3 State Fabric objects', 'Function DNA'],
    cta: 'Start free trial', href: `${AUTH_ORIGIN}/signup?plan=starter`,
  },
  {
    name: 'Professional', monthly: '$79', annual: '$63',
    desc: 'For growing teams and production SaaS.',
    features: ['2.5M requests / month', '100 AI agents, 1M calls/mo', '25 published functions', 'L1–L3 verification', 'Email support', '90-day execution logs', 'Zero-knowledge Vault (500 secrets)', '30-day Time Machine', '10 State Fabric objects + Hot Cache', 'Function DNA', 'Consciousness (basic)'],
    cta: 'Start free trial', href: `${AUTH_ORIGIN}/signup?plan=professional`,
    featured: true, badge: 'Most Popular',
  },
  {
    name: 'Enterprise', monthly: '$299', annual: '$239',
    desc: 'For large-scale apps, compliance, and custom trust policies.',
    features: ['25M requests / month', '500 AI agents, 5M calls/mo', 'Unlimited functions', 'L1–L4 verification', 'Priority + SLA', '1-year execution logs', 'Zero-knowledge Vault (5K secrets)', '90-day Time Machine + reconciliation', 'RBAC + Secret sharing', 'Unlimited State Fabric + all add-ons', 'Function DNA', 'Advanced Consciousness'],
    cta: 'Start free trial', href: `${AUTH_ORIGIN}/signup?plan=enterprise`,
  },
]

const TierCard: React.FC<{ t: Tier; period: 'monthly' | 'annual' }> = ({ t, period }) => (
  <Card style={{
    borderColor: t.featured ? 'var(--accent)' : 'var(--panel-edge)',
    background: t.featured
      ? 'linear-gradient(135deg, rgba(255, 122, 61, 0.08), var(--panel-raised))'
      : 'var(--panel-raised)',
    position: 'relative',
    display: 'flex', flexDirection: 'column', height: '100%',
  }}>
    {t.badge && (
      <div style={{
        position: 'absolute', top: -10, right: 'var(--space-4)',
        background: 'var(--accent)', color: 'var(--text-on-light)',
        fontFamily: 'var(--font-mono)', fontSize: 10, fontWeight: 700,
        letterSpacing: '0.06em', textTransform: 'uppercase',
        padding: '2px 8px', borderRadius: 'var(--radius-sm)',
      }}>{t.badge}</div>
    )}
    <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>{t.name}</h3>
    <div style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
      {period === 'monthly' ? t.monthly : t.annual}
      <span style={{ fontSize: 14, fontWeight: 400, color: 'var(--text-dim)' }}>/mo</span>
    </div>
    <p style={{ color: 'var(--text-dim)', fontSize: 13, lineHeight: 1.5, marginBottom: 'var(--space-4)' }}>{t.desc}</p>
    <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-5) 0', flex: 1 }}>
      {t.features.map((f) => (
        <li key={f} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)', color: 'var(--text-dim)', fontSize: 13, marginBottom: 8 }}>
          <span style={{ color: 'var(--status-ok)', flexShrink: 0, marginTop: 2 }}>✓</span>{f}
        </li>
      ))}
    </ul>
    {t.featured
      ? <SealedButton onClick={() => { window.location.href = t.href }} style={{ width: '100%' }}>{t.cta}</SealedButton>
      : <FrameButton onClick={() => { window.location.href = t.href }} style={{ width: '100%' }}>{t.cta}</FrameButton>
    }
  </Card>
)

const PeriodToggle: React.FC<{ value: 'monthly' | 'annual'; onChange: (v: 'monthly' | 'annual') => void }> = ({ value, onChange }) => (
  <div style={{
    display: 'inline-flex', alignItems: 'center', gap: 0,
    padding: 4, background: 'var(--panel)', border: '1px solid var(--panel-edge)',
    borderRadius: 'var(--radius)', marginBottom: 'var(--space-6)',
  }}>
    {(['monthly', 'annual'] as const).map((p) => (
      <button
        key={p}
        onClick={() => onChange(p)}
        style={{
          padding: 'var(--space-2) var(--space-4)',
          background: value === p ? 'var(--panel-raised)' : 'transparent',
          color: value === p ? 'var(--text)' : 'var(--text-dim)',
          border: 'none', borderRadius: 'var(--radius-sm)',
          fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 500,
          letterSpacing: '0.04em', textTransform: 'uppercase',
          cursor: 'pointer', transition: 'all var(--duration-fast) var(--ease-out)',
        }}
      >
        {p}{p === 'annual' && <span style={{ color: 'var(--status-ok)', marginLeft: 8 }}>−20%</span>}
      </button>
    ))}
  </div>
)

const PricingPage: React.FC = () => {
  const [platformPeriod, setPlatformPeriod] = useState<'monthly' | 'annual'>('monthly')

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
              Infrastructure layer
              <StatusPill status="live" label="Live" />
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Pricing built for <span style={{ color: 'var(--accent)' }}>trust</span>
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              Start free. Paid tiers include AI agents with generous call volumes and concurrency. Upgrade as your agent workloads grow.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap', marginBottom: 'var(--space-6)' }}>
              <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Get started</SealedButton>
              <FrameButton onClick={() => { window.location.href = '/docs' }}>Read the docs</FrameButton>
              <TrustSeal label="Transparent pricing" />
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle lead="Functions, hosting providers, custom domains, vault secrets, and collaboration—priced per workspace.">
            Platform plans
          </SectionTitle>
          <div style={{ textAlign: 'center' }}>
            <PeriodToggle value={platformPeriod} onChange={setPlatformPeriod} />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-5)' }}>
            {platformTiers.map((t) => <TierCard key={t.name} t={t} period={platformPeriod} />)}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Compare plans</SectionTitle>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="table" style={{ overflowX: 'auto' }}>
              <div className="table__header" style={{
                display: 'grid', gridTemplateColumns: '1.4fr repeat(4, 1fr)',
                borderBottom: '1px solid var(--panel-edge)',
              }}>
                <div className="table__th">Feature</div>
                <div className="table__th">Free</div>
                <div className="table__th">Starter</div>
                <div className="table__th" style={{ color: 'var(--accent)' }}>Professional</div>
                <div className="table__th">Enterprise</div>
              </div>
              {[
                ['Requests / month', ['25K', '500K', '2.5M', '25M']],
                ['AI agents included', ['3', '10', '100', '500']],
                ['AI calls / month', ['10K', '100K', '1M', '5M']],
                ['Published functions', ['3', '10', '25', 'Unlimited']],
                ['Verification levels', ['L1', 'L1–L2', 'L1–L3', 'L1–L4']],
                ['Execution logs', ['7 days', '30 days', '90 days', '1 year']],
                ['Time Machine (basic)', ['✓', '✓', '✓', '✓']],
                ['Time Machine (72h replay)', ['—', '✓', '✓', '✓']],
                ['Time Machine (30-day replay)', ['—', '—', '✓', '✓']],
                ['Time Machine (90-day replay)', ['—', '—', '—', '✓']],
                ['Vault secrets', ['25', '50', '500', '5,000']],
                ['State Fabric objects', ['1', '3', '10', 'Unlimited']],
                ['Function DNA', ['—', '✓', '✓', '✓']],
                ['Consciousness (basic)', ['—', '—', '✓', '✓']],
                ['Advanced Consciousness', ['—', '—', '—', '✓']],
                ['Support', ['Community', 'Email', 'Email', 'Priority + SLA']],
              ].map(([feature, vals]) => (
                <div key={feature as string} className="table__tr" style={{
                  display: 'grid', gridTemplateColumns: '1.4fr repeat(4, 1fr)',
                  borderBottom: '1px solid var(--panel-edge)',
                }}>
                  <div className="table__td" style={{ color: 'var(--text)' }}>{feature as string}</div>
                  {(vals as string[]).map((v, i) => (
                    <div key={i} className="table__td" style={{ color: i === 2 ? 'var(--accent)' : 'var(--text-dim)' }}>
                      {v}
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
          <SectionTitle>Frequently asked questions</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            {[
              { q: 'Do prices include tax?', a: 'Listed prices are in USD before applicable taxes. Your invoice will reflect tax rules for your billing address.' },
              { q: 'How does the Trust API and verification billing work?', a: 'Trust scores, attestations, and verification levels are part of the trust layer. Usage-based components are metered separately so you only pay for what agents and policies query.' },
              { q: 'Can we upgrade or downgrade?', a: 'Yes. Move between platform tiers as your traffic and team change; agent plans can be adjusted when your concurrency and call volume need it.' },
              { q: 'What does Enterprise include?', a: 'Volume pricing, custom SLAs, security questionnaires, SSO, and tailored trust policies. Contact sales with your requirements.' },
              { q: 'Is there a free trial for paid plans?', a: 'Yes — all paid plans come with a 14-day free trial. No credit card required to start.' },
            ].map((item) => (
              <Card key={item.q}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {item.q}
                </h3>
                <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6, margin: 0 }}>{item.a}</p>
              </Card>
            ))}
          </div>
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
              }}>
                Ready to build with<br />
                <span style={{ color: 'var(--accent)' }}>trust infrastructure</span>?
              </h2>
              <p style={{ fontSize: 17, color: 'var(--text-dim)', margin: '0 auto var(--space-6)', lineHeight: 1.6 }}>
                Start free today. No credit card required.
              </p>
              <div style={{ display: 'inline-flex', gap: 'var(--space-3)', flexWrap: 'wrap', justifyContent: 'center' }}>
                <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Create free account</SealedButton>
                <FrameButton onClick={() => { window.location.href = '/contact' }}>Talk to sales</FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default PricingPage

import React from 'react'
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
    {lead && <p style={{ color: 'var(--text-dim)', maxWidth: 640, margin: '0 auto', lineHeight: 1.6 }}>{lead}</p>}
  </div>
)

interface PlanCard {
  tier: string
  price: string
  period: string
  desc: string
  features: string[]
  cta: string
  href: string
  featured?: boolean
  badge?: string
}

const plans: PlanCard[] = [
  {
    tier: 'Free', price: '$0', period: '/month',
    desc: 'Perfect for exploring FunctionFly and testing agent integrations.',
    features: ['25,000 requests/month', '3 published functions', '3 AI agents', '10K AI calls/month', 'Basic trust scores', 'L1 verification only', 'Community support', '7-day execution logs', '24h Time Machine', 'Zero-knowledge Vault (25 secrets)', '1 State Fabric object'],
    cta: 'Get started free', href: `${AUTH_ORIGIN}/signup`,
  },
  {
    tier: 'Professional', price: '$79', period: '/month',
    desc: 'For growing teams building production agent applications.',
    features: ['2.5M requests/month', '25 published functions', '100 AI agents included', '1M AI calls/month', 'Advanced trust scores', 'L1-L3 verification', 'Email support', '90-day execution logs', 'Zero-knowledge Vault (500 secrets)', '30-day Time Machine', '10 State Fabric objects + Hot Cache'],
    cta: 'Start free trial', href: `${AUTH_ORIGIN}/signup?plan=professional`,
    featured: true, badge: 'Most Popular',
  },
  {
    tier: 'Enterprise', price: '$299', period: '/month',
    desc: 'For teams requiring advanced trust and collaboration features.',
    features: ['25M requests/month', 'Unlimited published functions', '500 AI agents included', '5M AI calls/month', 'Full trust scores + attestations', 'L1-L4 verification', 'Priority support + SLA', '1-year execution logs', 'Zero-knowledge Vault (5K secrets)', '90-day Time Machine + reconciliation', 'RBAC + Secret sharing', 'Unlimited State Fabric + all add-ons'],
    cta: 'Start free trial', href: `${AUTH_ORIGIN}/signup?plan=enterprise`,
  },
  {
    tier: 'Agent Enterprise', price: '$499', period: '/month',
    desc: 'Unlimited AI scale for organizations with custom security, compliance, and scaling needs.',
    features: ['Unlimited requests', 'Unlimited published functions', 'Unlimited AI agents', 'Unlimited AI calls', 'Full trust infrastructure', 'All verification levels', 'Dedicated support + SLA', '7-year data retention', 'SSO / SAML', 'Custom integrations', 'Audit logs + compliance reports', 'Incident insurance', 'Zero-knowledge Vault (1M secrets)', 'Unlimited State Fabric + all add-ons'],
    cta: 'Contact sales', href: '/contact',
  },
]

const PlanCardEl: React.FC<{ p: PlanCard }> = ({ p }) => (
  <Card style={{
    borderColor: p.featured ? 'var(--accent)' : 'var(--panel-edge)',
    background: p.featured
      ? 'linear-gradient(135deg, rgba(255, 122, 61, 0.08), var(--panel-raised))'
      : 'var(--panel-raised)',
    position: 'relative',
  }}>
    {p.badge && (
      <div style={{
        position: 'absolute', top: -10, right: 'var(--space-4)',
        background: 'var(--accent)', color: 'var(--text-on-light)',
        fontFamily: 'var(--font-mono)', fontSize: 10, fontWeight: 700,
        letterSpacing: '0.06em', textTransform: 'uppercase',
        padding: '2px 8px', borderRadius: 'var(--radius-sm)',
      }}>{p.badge}</div>
    )}
    <div style={{ marginBottom: 'var(--space-5)' }}>
      <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
        {p.tier}
      </h3>
      <div style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
        {p.price}<span style={{ fontSize: 14, fontWeight: 400, color: 'var(--text-dim)' }}>{p.period}</span>
      </div>
      <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.5 }}>{p.desc}</p>
    </div>
    <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-5) 0' }}>
      {p.features.map((f) => (
        <li key={f} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)', color: 'var(--text-dim)', fontSize: 13, marginBottom: 8 }}>
          <span style={{ color: 'var(--status-ok)', flexShrink: 0, marginTop: 2 }}>✓</span> {f}
        </li>
      ))}
    </ul>
    {p.featured
      ? <SealedButton onClick={() => { window.location.href = p.href }} style={{ width: '100%' }}>{p.cta}</SealedButton>
      : <FrameButton onClick={() => { window.location.href = p.href }} style={{ width: '100%' }}>{p.cta}</FrameButton>
    }
  </Card>
)

const PlansPage: React.FC = () => {
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
              Simple, transparent pricing
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Choose the plan<br />that scales with you
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              Start free, pay as you grow. No hidden fees, no surprise bills. Every plan includes trust infrastructure, function registry, and sandboxed execution.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap', alignItems: 'center' }}>
              <TrustSeal label="Execution-backed" size="default" />
              <StatusPill status="live" label="14-day free trial" />
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 'var(--space-5)' }}>
            {plans.map((p) => <PlanCardEl key={p.tier} p={p} />)}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle lead="Enhance your FunctionFly experience with these optional add-ons available on any paid plan.">
            Extend your plan
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>Zero-Knowledge Vault</h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Client-side encrypted secrets storage. Server never sees plaintext. Free includes 25 secrets, Pro up to 500, Team up to 5K.
              </p>
            </Card>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>Dynamic Credentials</h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Auto-rotating database credentials for PostgreSQL and MySQL. Pro includes PostgreSQL, Team adds MySQL support.
              </p>
            </Card>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>Time Machine Replay</h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Replay past function executions with full state. Free: 24h, Starter: 72h, Professional: 30 days, Enterprise: 90 days + live reconciliation.
              </p>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Common questions</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            {[
              { q: 'How do function requests work?', a: 'Each function call counts as one request, regardless of execution time. Failed calls due to trust policy enforcement also count toward your limit.' },
              { q: 'Can I upgrade or downgrade anytime?', a: 'Yes. You can change your plan at any time. Upgrades take effect immediately, and you\'ll be charged a prorated amount. Downgrades take effect at your next billing cycle.' },
              { q: 'What happens if I exceed my request limit?', a: 'We\'ll notify you when you reach 80% of your limit. If you exceed it, functions will return a rate limit error until you upgrade or the cycle resets.' },
              { q: 'Is there a free trial for paid plans?', a: 'Yes, Professional and Enterprise plans come with a 14-day free trial. No credit card required to start.' },
              { q: 'What is the Zero-Knowledge Vault?', a: 'The Vault uses client-side encryption — your passphrase and secrets are never sent to our servers. All encryption happens in your browser using AES-256-GCM. Free includes 25 secrets, Professional up to 500, Team up to 5,000.' },
              { q: 'What verification levels are included?', a: 'Free includes L1 (format checks). Professional adds L2-L3 (security scans and code review). Enterprise adds L4 (platform-verified) plus custom attestation workflows.' },
              { q: 'Do prices include tax?', a: 'Listed prices are in USD and exclude applicable taxes. Your invoice will reflect tax rules for your billing address.' },
            ].map((item) => (
              <Card key={item.q}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {item.q}
                </h3>
                <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6, margin: 0 }}>
                  {item.a}
                </p>
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

export default PlansPage

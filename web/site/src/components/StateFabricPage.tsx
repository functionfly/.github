import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  Card,
} from './sc'
import '../styles/sc-main.css';

const Container: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({ children, narrow }) => (
  <div style={{ maxWidth: narrow ? 720 : 1100, margin: '0 auto', padding: '0 var(--space-4)' }}>
    {children}
  </div>
)
const Section: React.FC<{ children: React.ReactNode; alt?: boolean }> = ({ children, alt }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)

const CheckIcon: React.FC<{ size?: number }> = ({ size = 16 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <polyline points="20 6 9 17 4 12" stroke="var(--status-ok)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const useCases = [
  {
    title: 'Multi-step AI Agents',
    desc: 'Build agents that maintain context across thousands of turns. State Fabric gives your agents memory that survives restarts, scales to millions of operations, and can be replayed for debugging.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm0 18a8 8 0 1 1 8-8 8 8 0 0 1-8 8z"/>
        <path d="M12 6v6l4 2"/>
      </svg>
    ),
  },
  {
    title: 'Long-running Workflows',
    desc: 'Orchestrate workflows that span hours or days. Every step is checkpointed, every state change is auditable, and any step can be replayed or modified mid-flight.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="3" y="3" width="18" height="18" rx="2"/>
        <path d="M3 9h18M9 21V9"/>
      </svg>
    ),
  },
  {
    title: 'Event Sourcing',
    desc: 'Capture every state change as an immutable event. Build read models, enable full audit trails, and replay history to reconstruct state at any point in time.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8"/>
      </svg>
    ),
  },
  {
    title: 'Distributed Caching',
    desc: 'Use State Fabric as a low-latency, consistency-guaranteed cache layer. Reads hit hot data, writes are serialized and replicated, and the cache is always consistent.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
        <path d="M3.27 6.96L12 12.01l8.73-5.05M12 22.08V12"/>
      </svg>
    ),
  },
  {
    title: 'Session Management',
    desc: 'Store user sessions with millisecond latency and zero eventual consistency issues. Sessions survive server restarts, can be shared across services, and replay enables powerful debugging.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
        <circle cx="12" cy="7" r="4"/>
      </svg>
    ),
  },
  {
    title: 'Game State',
    desc: 'Build multiplayer games with authoritative server state. Every player action is ordered, every state change is deterministic, and replays let players review any match frame-by-frame.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <polygon points="5 3 19 12 5 21 5 3"/>
      </svg>
    ),
  },
]

const features = [
  {
    title: 'Hot Cache Tier',
    desc: 'Frequently accessed state stays in memory for sub-millisecond reads. Write paths are serialized and durable, but reads hit cache-first.',
    icon: '⚡',
  },
  {
    title: 'Deterministic Replay',
    desc: 'Replay any execution with bit-for-bit identical results. Debug production issues locally, test edge cases exhaustively, or rebuild state from scratch.',
    icon: '⏪',
  },
  {
    title: 'Snapshot Retention',
    desc: 'State snapshots are retained from 7 days to 365 days depending on your plan. Point-in-time recovery, audit compliance, and historical analysis all covered.',
    icon: '📸',
  },
  {
    title: 'Event Subscription Streams',
    desc: 'Subscribe to state changes in real-time via WebSockets or webhooks. Build reactive UIs, trigger downstream workflows, or stream events to your data warehouse.',
    icon: '📡',
  },
  {
    title: 'Multi-Region Replication',
    desc: 'State is replicated across regions for low-latency access globally and resilience against zone outages. Active-active for reads, active-passive for writes.',
    icon: '🌍',
  },
  {
    title: 'Immutable Audit Logs',
    desc: 'Every state change is logged to an append-only, cryptographically signed audit trail. SOC2 and HIPAA compliant out of the box.',
    icon: '🔒',
  },
]

const howItWorks = [
  {
    step: '01',
    title: 'Define your state schema',
    desc: 'Describe your state as a TypeScript interface. State Fabric handles serialization, validation, and versioning automatically.',
  },
  {
    step: '02',
    title: 'Read and write from any function',
    desc: 'Use the SDK to read, write, and subscribe to state changes. All operations are atomic and consistency-guaranteed.',
  },
  {
    step: '03',
    title: 'Replay when needed',
    desc: 'Debug by replaying any execution locally. Reproduce bugs exactly as they happened in production.',
  },
  {
    step: '04',
    title: 'Scale infinitely',
    desc: 'State Fabric handles sharding, replication, and failover automatically. Your state survives data center fires.',
  },
]

const faqs = [
  {
    q: 'How is State Fabric different from a database?',
    a: 'State Fabric is purpose-built for serverless and AI agent workloads. Unlike a traditional database, every state change is tied to an execution, enabling deterministic replay. Reads are cached in a hot tier for low latency, and the replay model means you never need to write migration scripts for schema changes—you just replay.',
  },
  {
    q: 'Can I use State Fabric with my existing database?',
    a: 'Yes. State Fabric works alongside your existing data layer. Use it for session state, agent memory, workflow checkpoints, and other latency-sensitive state. For persistent storage of record-level data, keep your existing database.',
  },
  {
    q: 'What happens if State Fabric goes down?',
    a: 'State Fabric replicates across multiple availability zones. For Enterprise plans, active-active multi-region replication ensures zero RTO. Your state survives data center failures, and writes are always durably acknowledged before returning.',
  },
  {
    q: 'How does pricing work?',
    a: 'State Fabric is included in all platform plans. Free includes 1 state object, Starter includes 3, Professional includes 10, and Enterprise includes unlimited. Premium add-ons like Hot Cache, Multi-Region Replication, and AI Memory are available as upgrades.',
  },
  {
    q: 'Is State Fabric suitable for financial transactions?',
    a: 'Yes. Enterprise plans include immutable audit logs, multi-region replication, BYOK encryption, and SOC2/HIPAA compliance. The deterministic replay model also provides a complete, auditable history of every state change.',
  },
  {
    q: 'Can I export my state?',
    a: 'Enterprise plans include replay export—dump the full state history to object storage in standard formats. All plans support reading current state via the SDK at any time.',
  },
]

const StateFabricPage: React.FC = () => {
  return (
    <>
      <PageGrid />

      {/* Hero */}
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
              State Fabric
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Durable state for<br />serverless functions
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              State Fabric adds structured, replayable state to your functions. Every state change is logged, every execution can be replayed, and your state survives restarts, scales to millions of operations, and stays consistent across regions.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = `${AUTH_ORIGIN}/signup`}>
                Get started free
              </SealedButton>
              <FrameButton onClick={() => window.location.href = '/docs/state-fabric'}>
                Read the docs
              </FrameButton>
            </div>
          </Chamber>
        </Container>
      </Section>

      {/* Use Cases */}
      <Section alt>
        <Container>
          <div style={{ textAlign: 'center', marginBottom: 'var(--space-8)' }}>
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
              letterSpacing: '-0.005em', color: 'var(--text)',
              marginBottom: 'var(--space-4)',
            }}>
              Built for modern workloads
            </h2>
            <p style={{ color: 'var(--text-dim)', maxWidth: 600, margin: '0 auto' }}>
              From AI agents to long-running workflows, State Fabric handles the state that stateless functions can't.
            </p>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-4)' }}>
            {useCases.map((uc) => (
              <Card key={uc.title}>
                <div style={{
                  width: 48, height: 48, borderRadius: 'var(--radius-md)',
                  background: 'var(--panel-raised)', display: 'flex',
                  alignItems: 'center', justifyContent: 'center',
                  color: 'var(--status-ok)', marginBottom: 'var(--space-4)',
                }}>
                  {uc.icon}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {uc.title}
                </h3>
                <p style={{ fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {uc.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      {/* How It Works */}
      <Section>
        <Container>
          <div style={{ textAlign: 'center', marginBottom: 'var(--space-8)' }}>
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
              letterSpacing: '-0.005em', color: 'var(--text)',
              marginBottom: 'var(--space-4)',
            }}>
              How it works
            </h2>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
            {howItWorks.map((step) => (
              <div key={step.step} style={{ textAlign: 'center' }}>
                <div style={{
                  width: 64, height: 64, borderRadius: '50%',
                  border: '2px solid var(--panel-edge)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontFamily: 'var(--font-mono)', fontSize: 18, fontWeight: 700,
                  color: 'var(--accent)', margin: '0 auto var(--space-4)',
                }}>
                  {step.step}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {step.title}
                </h3>
                <p style={{ fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {step.desc}
                </p>
              </div>
            ))}
          </div>
        </Container>
      </Section>

      {/* Features */}
      <Section alt>
        <Container>
          <div style={{ textAlign: 'center', marginBottom: 'var(--space-8)' }}>
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
              letterSpacing: '-0.005em', color: 'var(--text)',
              marginBottom: 'var(--space-4)',
            }}>
              Enterprise-grade capabilities
            </h2>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-4)' }}>
            {features.map((f) => (
              <div key={f.title} style={{ display: 'flex', gap: 'var(--space-3)' }}>
                <div style={{
                  width: 40, height: 40, borderRadius: 'var(--radius-md)',
                  background: 'var(--panel-raised)', display: 'flex',
                  alignItems: 'center', justifyContent: 'center',
                  fontSize: 18, flexShrink: 0,
                }}>
                  {f.icon}
                </div>
                <div>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-1)' }}>
                    {f.title}
                  </h3>
                  <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                    {f.desc}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </Container>
      </Section>

      {/* FAQ */}
      <Section>
        <Container narrow>
          <div style={{ textAlign: 'center', marginBottom: 'var(--space-8)' }}>
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
              letterSpacing: '-0.005em', color: 'var(--text)',
              marginBottom: 'var(--space-4)',
            }}>
              Frequently asked questions
            </h2>
          </div>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
              {faqs.map((faq) => (
                <div key={faq.q}>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                    {faq.q}
                  </h3>
                  <p style={{ fontSize: 14, color: 'var(--text-dim)', lineHeight: 1.7, margin: 0 }}>
                    {faq.a}
                  </p>
                </div>
              ))}
            </div>
          </Chamber>
        </Container>
      </Section>

      {/* CTA */}
      <Section>
        <Container narrow>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
                letterSpacing: '-0.005em', color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>
                Add state to your functions today
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                State Fabric is included in every platform plan. Start with the free tier—no credit card required.
              </p>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'center', flexWrap: 'wrap' }}>
                <SealedButton onClick={() => window.location.href = `${AUTH_ORIGIN}/signup`}>
                  Get started free
                </SealedButton>
                <FrameButton onClick={() => window.location.href = '/contact'}>
                  Talk to sales
                </FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default StateFabricPage

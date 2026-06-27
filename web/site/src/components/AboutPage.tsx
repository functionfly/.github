import React from 'react'
import { AUTH_ORIGIN } from '../config'
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
  <section style={{ padding: 'var(--space-8) 0', background: alt ? 'var(--bg)' : 'var(--bg)' }}>{children}</section>
)

const ShieldCheckIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const LockIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const ChartIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const BoltIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const CodeIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <polyline points="16 18 22 12 16 6" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
    <polyline points="8 6 2 12 8 18" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const GlobeIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const EyeIcon: React.FC<{ size?: number }> = ({ size = 24 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M1 12s4-8 11-8 11 8 11 8-4 11-8-4-11-11-8-11 8zm10 10a10 10 0 002-10m0 10a10 10 0 002-10" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const steps = [
  { num: '01', title: 'Publish & Verify', desc: 'Publish a function with a manifest describing its behavior, capabilities, and intended use. Submit to automated verification — static analysis, fuzzing, signature checks — and human attestation for higher trust tiers.', icon: <ShieldCheckIcon /> },
  { num: '02', title: 'Sandbox Execution', desc: 'Every function call runs in an isolated WebAssembly sandbox with declared capabilities enforced at runtime. No filesystem access, no network egress, no privilege escalation — unless explicitly granted.', icon: <LockIcon /> },
  { num: '03', title: 'Trust Scores', desc: 'Every execution feeds into a live trust score. Functions that run successfully thousands of times earn higher scores. Agents can set policy thresholds and refuse to call functions below a given tier.', icon: <ChartIcon /> },
  { num: '04', title: 'Zero-Knowledge Vault', desc: 'Store API keys, secrets, and sensitive config in the FunctionFly Vault. Encryption and decryption happen client-side — the server never sees your plaintext or your passphrase.', icon: <BoltIcon /> },
]

const principles = [
  { num: '01', title: 'Deny-by-default capabilities', desc: 'Every capability — network access, filesystem read/write, environment variable exposure — must be declared at publish time and is enforced at runtime by the sandbox.' },
  { num: '02', title: 'Trust is earned, not claimed', desc: 'Trust scores are computed from real execution outcomes. A function that has run 50,000 times with a 99.97% success rate has a measurably different trust profile.' },
  { num: '03', title: 'Multi-level verification', desc: 'Functions can submit to automated checks and human attestation for higher trust tiers. Verification is ongoing, not one-time.' },
  { num: '04', title: 'Observability is not optional', desc: 'Every function call is recorded: input, output, latency, cost, trust score delta, and failure mode.' },
  { num: '05', title: 'Multi-language, no rewrites', desc: 'Write functions in Python, Go, Rust, JavaScript, TypeScript, or WebAssembly. One SDK, every runtime.' },
  { num: '06', title: 'Zero-knowledge by design', desc: 'The server stores only ciphertext, IV, salt, and auth tag — it cannot decrypt your secrets.' },
]

const values = [
  { title: 'Trust is technical', desc: 'We build measurable signals — verification levels, execution history, signed attestations, trust scores — that agents can audit programmatically.', icon: <ShieldCheckIcon size={20} /> },
  { title: 'Safe by default', desc: 'Every function runs with the minimum capabilities it needs — nothing more. No ambient authority, no implicit permissions.', icon: <LockIcon size={20} /> },
  { title: 'Agent-native from day one', desc: 'FunctionFly was designed for agents to discover, evaluate, and call functions — with manifests, trust filters, and policy routing built in.', icon: <GlobeIcon size={20} /> },
  { title: 'Measurable, not magical', desc: 'Trust tiers are earned through verifiable evidence — not self-attestation or marketing language.', icon: <ChartIcon size={20} /> },
  { title: 'Multi-language, no rewrites', desc: 'Write in Python, Go, Rust, JavaScript, TypeScript, or WASM. One SDK, every runtime.', icon: <CodeIcon size={20} /> },
  { title: 'Full observability', desc: 'Execution traces, audit logs, latency breakdowns, cost attribution, and trust score history.', icon: <EyeIcon size={20} /> },
]

const milestones = [
  { year: 'Q1 2026', title: 'Founded', desc: 'FunctionFly incorporated in Wyoming, headquartered in Fort Worth, TX. Began building the trust layer for AI agents.' },
  { year: 'Q2 2026', title: 'Beta Launch', desc: 'Public beta with multi-language SDKs for Python, Go, Rust, JavaScript, and TypeScript. Released the function registry API and Zero-knowledge Vault.' },
  { year: 'Q2 2026', title: 'Today', desc: 'FunctionFly powers trust for agents across production deployments. 5 language runtimes, 4 trust tiers, 12 global regions, billions of function calls processed.', live: true },
]

const AboutPage: React.FC = () => {
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
              textTransform: 'uppercase', color: 'var(--text-faint)',
              marginBottom: 'var(--space-4)',
            }}>
              Founded 2026 — Wyoming / Fort Worth, TX
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Agent tools<br />should be<br />trustworthy.
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              We built FunctionFly so AI agents can call any function and know — not hope — that it's safe, signed, and behaving as advertised. Trust is technical, not metaphorical.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = `${AUTH_ORIGIN}/signup`}>
                Start Building
              </SealedButton>
              <FrameButton onClick={() => window.location.href = '/trust'}>
                How Trust Works
              </FrameButton>
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section alt>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-4)',
          }}>
            Every tool an agent calls is a leap of faith
          </h2>
          <p style={{ fontSize: 15, lineHeight: 1.7, color: 'var(--text-dim)', maxWidth: 800, marginBottom: 'var(--space-4)' }}>
            Today, AI agents call functions — email senders, payment processors, database queries — with no way to verify those functions do what they claim. There's no trust infrastructure: no verification, no attestations, no execution history, no trust scores.
          </p>
          <p style={{ fontSize: 15, lineHeight: 1.7, color: 'var(--text-dim)', maxWidth: 800 }}>
            Existing platforms treat tools as dumb endpoints. Define the API, the agent calls it, whatever comes back is accepted. But in a world where agents act autonomously on your behalf, that gap is not just a technical problem — it's a security and reliability crisis waiting to happen.
          </p>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{
            fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
            textTransform: 'uppercase', color: 'var(--accent)',
            marginBottom: 'var(--space-3)',
          }}>
            The Solution
          </div>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-3)',
          }}>
            FunctionFly is the <span style={{ color: 'var(--accent)' }}>trust layer</span>
          </h2>
          <p style={{ fontSize: 15, color: 'var(--text-dim)', marginBottom: 'var(--space-7)' }}>
            A complete trust infrastructure for AI agent tool usage — from publication to execution to trust score.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-5)' }}>
            {steps.map((step) => (
              <Card key={step.num}>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 24, fontWeight: 500, color: 'var(--accent)', marginBottom: 'var(--space-3)' }}>
                  {step.num}
                </div>
                <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-3)' }}>
                  {step.icon}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {step.title}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {step.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section alt>
        <Container>
          <div style={{
            fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
            textTransform: 'uppercase', color: 'var(--accent)',
            marginBottom: 'var(--space-3)',
          }}>
            How We Build
          </div>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-7)',
          }}>
            The principles behind every decision
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-5)' }}>
            {principles.map((p) => (
              <div key={p.num} style={{ display: 'flex', gap: 'var(--space-4)' }}>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 13, color: 'var(--accent)', flexShrink: 0 }}>
                  {p.num}
                </div>
                <div>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                    {p.title}
                  </h3>
                  <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                    {p.desc}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{
            fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
            textTransform: 'uppercase', color: 'var(--accent)',
            marginBottom: 'var(--space-3)',
          }}>
            Our Values
          </div>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-3)',
          }}>
            What we stand for
          </h2>
          <p style={{ fontSize: 15, color: 'var(--text-dim)', marginBottom: 'var(--space-6)' }}>
            Six beliefs that shape every product decision, every architectural choice, and every line of code we ship.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-4)' }}>
            {values.map((v) => (
              <Card key={v.title}>
                <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-3)' }}>
                  {v.icon}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {v.title}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {v.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section alt>
        <Container>
          <div style={{
            fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
            textTransform: 'uppercase', color: 'var(--accent)',
            marginBottom: 'var(--space-3)',
          }}>
            Milestones
          </div>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-6)',
          }}>
            How we got here
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
            {milestones.map((m, i) => (
              <Card key={m.year}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)', marginBottom: 'var(--space-3)' }}>
                  <span style={{
                    fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500,
                    color: m.live ? 'var(--status-ok)' : 'var(--text-faint)',
                    letterSpacing: '0.06em', textTransform: 'uppercase',
                  }}>
                    {m.year}
                  </span>
                  {m.live && (
                    <span style={{
                      width: 6, height: 6, borderRadius: '50%',
                      background: 'var(--status-ok)',
                      boxShadow: '0 0 8px var(--status-ok)',
                    }} />
                  )}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {m.title}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {m.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container narrow>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{
              fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
              textTransform: 'uppercase', color: 'var(--accent)',
              marginBottom: 'var(--space-3)',
            }}>
              Start Today
            </div>
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
              letterSpacing: '-0.005em', color: 'var(--text)',
              marginBottom: 'var(--space-3)',
            }}>
              Ready to build with<br />
              <span style={{ color: 'var(--accent)' }}>trust infrastructure</span>?
            </h2>
            <p style={{ fontSize: 15, color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-5)' }}>
              Whether you're building agent tooling, evaluating trust infrastructure, or just want to understand how FunctionFly works — get started in minutes.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap', marginBottom: 'var(--space-4)' }}>
              <SealedButton onClick={() => window.location.href = `${AUTH_ORIGIN}/signup`}>
                Create free account
              </SealedButton>
              <FrameButton onClick={() => window.location.href = '/pricing'}>
                View pricing
              </FrameButton>
            </div>
            <p style={{ fontSize: 12, color: 'var(--text-faint)', margin: 0 }}>
              No credit card required. Free tier includes 25K requests/month and 3 functions.
            </p>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default AboutPage

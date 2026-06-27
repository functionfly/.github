import React from 'react'
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
const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)

const UsersIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
    <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const ShieldIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
  </svg>
)

const CodeIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <polyline points="16 18 22 12 16 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <polyline points="8 6 2 12 8 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const MailIcon: React.FC<{ size?: number }> = ({ size = 18 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="2" y="4" width="20" height="16" rx="2" stroke="currentColor" strokeWidth="2"/>
    <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const CareersPage: React.FC = () => {
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
              Join our team
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              We're building the trust layer<br />for AI agents
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              FunctionFly is a small, focused team working on verified functions, trust scores, and zero-knowledge cryptography. If you want to shape how agents choose tools safely, we'd love to hear from you.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = 'mailto:careers@functionfly.com'}>
                Send introduction
              </SealedButton>
              <FrameButton onClick={() => window.location.href = '/'}>
                Learn more
              </FrameButton>
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <CodeIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Engineering
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Build the core platform: function execution, trust scoring, MCP protocol, and the zero-knowledge vault.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <UsersIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Product & Design
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Shape the developer experience and dashboard for publishing, monitoring, and trust management.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ShieldIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Security
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Define verification standards, audit the platform, and ensure the trust layer is trustworthy by design.
              </p>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-3)', textAlign: 'center',
          }}>
            How to apply
          </h2>
          <p style={{ color: 'var(--text-dim)', textAlign: 'center', marginBottom: 'var(--space-7)' }}>
            We're not hiring right now, but we welcome interest from exceptional people.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { step: '1', title: 'Send your introduction', desc: 'Email careers@functionfly.com with a brief intro and links to your work.' },
              { step: '2', title: 'What to include', items: ['Your role interest (engineering, product, design, security, or operations)', '2-3 links to work you\'re proud of', 'Your location/time zone and remote preferences'] },
              { step: '3', title: 'We read every message', desc: 'Response time varies depending on our hiring cycle. We\'ll reach out when roles open.' },
            ].map((item) => (
              <Card key={item.step}>
                <div style={{
                  fontFamily: 'var(--font-mono)', fontSize: 32, fontWeight: 500,
                  color: 'var(--accent)', marginBottom: 'var(--space-3)',
                }}>
                  {item.step}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {item.title}
                </h3>
                {item.desc && (
                  <p style={{ fontSize: 14, color: 'var(--text-dim)', margin: 0, lineHeight: 1.5 }}>
                    {item.desc}
                  </p>
                )}
                {item.items && (
                  <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                    {item.items.map((i) => (
                      <li key={i} style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-1)', lineHeight: 1.5 }}>
                        → {i}
                      </li>
                    ))}
                  </ul>
                )}
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container narrow>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
                letterSpacing: '-0.005em', color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>
                Ready to make agents safer?
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                We'd love to hear from you.
              </p>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'center', flexWrap: 'wrap' }}>
                <SealedButton onClick={() => window.location.href = 'mailto:careers@functionfly.com'}>
                  <span style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <MailIcon /> Send introduction
                  </span>
                </SealedButton>
                <FrameButton onClick={() => window.location.href = '/'}>
                  Learn about FunctionFly
                </FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default CareersPage

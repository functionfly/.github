import React from 'react'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'

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

const LockIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
    <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const ShieldIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const ClockIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
    <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const FileIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
    <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="2"/>
  </svg>
)

const SecurityPage: React.FC = () => {
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
              Security Hub
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Your security is<br />our foundation
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              From encryption to access controls, see how FunctionFly protects your data and how to report vulnerabilities responsibly.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = '/vulnerability'}>
                Report a vulnerability
              </SealedButton>
              <FrameButton onClick={() => window.location.href = '/trust'}>
                Trust center
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
                <LockIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Encryption
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                All data is encrypted in transit with TLS 1.3 and at rest with AES-256. The zero-knowledge vault extends this to client-side encryption.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ShieldIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Access controls
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Role-based access control, multi-factor authentication for admins, and least-privilege principles across all systems.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ClockIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Monitoring
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                24/7 security monitoring, intrusion detection, automated threat response, and SIEM integration for enterprise deployments.
              </p>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Security measures</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { title: 'Encryption at rest', items: ['AES-256 encryption for all stored data', 'Transparent database encryption', 'Encrypted backups with separate key management'] },
              { title: 'Encryption in transit', items: ['TLS 1.3 for all connections', 'Certificate pinning for API endpoints', 'Perfect forward secrecy enabled'] },
              { title: 'Zero-knowledge vault', items: ['Client-side encryption of secrets', 'Platform never sees plaintext', 'You control who can decrypt'] },
              { title: 'Access management', items: ['Role-based access control (RBAC)', 'Multi-factor authentication (MFA)', 'Session management with timeout'] },
              { title: 'Network security', items: ['VPC isolation for production', 'Cloudflare DDoS protection', 'Firewall rules and segmentation'] },
              { title: 'Compliance', items: ['SOC 2 Type II certified', 'GDPR and CCPA compliant', 'COPPA compliant'] },
            ].map((item) => (
              <Card key={item.title}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                  {item.title}
                </h3>
                <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                  {item.items.map((i) => (
                    <li key={i} style={{ marginBottom: 'var(--space-2)', fontSize: 14, color: 'var(--text-dim)', display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)' }}>
                      <span style={{ color: 'var(--status-ok)', flexShrink: 0 }}>✓</span>
                      {i}
                    </li>
                  ))}
                </ul>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Responsible disclosure</SectionTitle>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <p style={{ color: 'var(--text-dim)', lineHeight: 1.7, marginBottom: 'var(--space-5)' }}>
              We are committed to working with the security community to verify, reproduce, and respond to security vulnerabilities in a timely manner.
            </p>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
              How to report
            </h3>
            <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
              Please email <a href="mailto:security@functionfly.com" style={{ color: 'var(--accent)' }}>security@functionfly.com</a> with details about the vulnerability. Include enough information to reproduce the issue — a proof of concept is helpful but not required.
            </p>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
              What we commit to
            </h3>
            <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-4) 0' }}>
              {['Acknowledgment within 24 hours', 'Regular updates on remediation progress', 'Public credit in our hall of fame (with your permission)'].map((item) => (
                <li key={item} style={{ marginBottom: 'var(--space-2)', fontSize: 14, color: 'var(--text-dim)', display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)' }}>
                  <span style={{ color: 'var(--status-ok)', flexShrink: 0 }}>✓</span>
                  {item}
                </li>
              ))}
            </ul>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
              Scope
            </h3>
            <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
              In-scope: production systems, the FunctionFly API, dashboard, and documented public integrations. Out of scope: social engineering, physical security, denial of service attacks against our infrastructure.
            </p>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Security resources</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-4)' }}>
            {[
              { href: '/.well-known/security.txt', icon: <FileIcon size={24} />, title: 'security.txt', desc: 'Our RFC 9116 security contact and policy file' },
              { href: '/trust', icon: <ShieldIcon size={24} />, title: 'Trust Center', desc: 'Verification, attestations, and compliance details' },
              { href: '/privacy', icon: <LockIcon size={24} />, title: 'Privacy Policy', desc: 'How we collect, use, and protect your data' },
              { href: '/compliance', icon: <ClockIcon size={24} />, title: 'Compliance', desc: 'Enterprise security and compliance certifications' },
            ].map((resource) => (
              <a key={resource.href} href={resource.href} style={{
                display: 'flex', alignItems: 'flex-start', gap: 'var(--space-4)',
                padding: 'var(--space-5)', background: 'var(--panel)',
                border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius-lg)',
                textDecoration: 'none', transition: 'border-color var(--duration-fast) var(--ease-out)',
              }}
              onMouseEnter={(e) => e.currentTarget.style.borderColor = 'var(--steel-light)'}
              onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--panel-edge)'}
              >
                <div style={{ color: 'var(--accent)', flexShrink: 0, marginTop: 2 }}>
                  {resource.icon}
                </div>
                <div>
                  <h4 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', margin: '0 0 var(--space-1) 0' }}>
                    {resource.title}
                  </h4>
                  <p style={{ fontSize: 13, color: 'var(--text-dim)', margin: 0, lineHeight: 1.5 }}>
                    {resource.desc}
                  </p>
                </div>
              </a>
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
                Have a security concern?
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                For vulnerabilities or security incidents, email us immediately.
              </p>
              <SealedButton onClick={() => window.location.href = 'mailto:security@functionfly.com'}>
                security@functionfly.com
              </SealedButton>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default SecurityPage

import React from 'react'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'
import { StatusPill } from './containment/StatusPill'

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

const TrustIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
  </svg>
)

const ShieldCheckIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const LockIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
    <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const CompliancePage: React.FC = () => {
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
              Security & Compliance
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Enterprise-grade security<br />and compliance
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              FunctionFly is built with security-first principles. Our platform meets the most demanding compliance requirements for enterprise deployments.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = '/contact'}>
                Contact sales
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
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <TrustIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                SOC 2 Type II
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6 }}>
                Our security controls and practices are audited annually by independent third-party auditors to maintain SOC 2 Type II certification.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ShieldCheckIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                GDPR Compliant
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6 }}>
                We comply with the General Data Protection Regulation for EU users, including data subject rights, legal bases for processing, and international data transfers.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <LockIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                CCPA Compliant
              </h3>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6 }}>
                California residents have rights to know what personal information is collected, opt-out of sale, and request deletion under the California Consumer Privacy Act.
              </p>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle id="architecture" lead="Built from the ground up with security in mind.">
            Security architecture
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Encryption
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  { label: 'Data in transit', value: 'TLS 1.3 for all connections' },
                  { label: 'Data at rest', value: 'AES-256 encryption' },
                  { label: 'Database encryption', value: 'Transparent data encryption' },
                  { label: 'Zero-knowledge vault', value: 'Client-side encryption' },
                ].map((item) => (
                  <li key={item.label} style={{ marginBottom: 'var(--space-3)', fontSize: 14 }}>
                    <strong style={{ color: 'var(--text)' }}>{item.label}:</strong>{' '}
                    <span style={{ color: 'var(--text-dim)' }}>{item.value}</span>
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Access controls
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'Role-based access control (RBAC)',
                  'Multi-factor authentication (MFA)',
                  'Regular access reviews',
                  'Principle of least privilege',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Security monitoring
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  '24/7 security monitoring',
                  'Regular security audits',
                  'Automated threat detection',
                  'SIEM integration',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Breach response
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'Incident response procedures',
                  '72-hour notification',
                  'Comprehensive breach logs',
                  'Post-incident reviews',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Compliance certifications</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { name: 'SOC 2 Type II', desc: 'Annual audit by independent third-party auditors covering security, availability, and confidentiality.', icon: <TrustIcon size={32} /> },
              { name: 'GDPR', desc: 'Full compliance with EU General Data Protection Regulation including data subject rights and DPA agreements.', icon: <ShieldCheckIcon size={32} /> },
              { name: 'CCPA', desc: 'California Consumer Privacy Act compliance with right to know, delete, and opt-out.', icon: <LockIcon size={32} /> },
              { name: 'COPPA', desc: "Children's Online Privacy Protection Act compliant with age verification and parental consent procedures.", icon: <LockIcon size={32} /> },
            ].map((cert) => (
              <Card key={cert.name}>
                <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                  {cert.icon}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                  {cert.name}
                </h3>
                <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                  {cert.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle id="data-residency" lead="Your data is stored and processed with appropriate safeguards.">
            Data residency & transfers
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Data transfer mechanisms
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'Standard Contractual Clauses (SCCs)',
                  'Adequacy decisions',
                  'Binding Corporate Rules',
                  'Explicit consent',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Data hosting locations
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'United States — Primary',
                  'European Union — GDPR-compliant',
                  'Global edge — Cloudflare CDN',
                  'Enterprise — Dedicated deployments',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Enterprise features</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Custom security requirements
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'Private deployments in your cloud',
                  'Customer-managed encryption keys (BYOK)',
                  'Custom SLA with guaranteed uptime',
                  'Dedicated infrastructure options',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                Enterprise support
              </h3>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {[
                  'Dedicated customer success manager',
                  'Priority support with response times',
                  'Custom onboarding and training',
                  'Security questionnaires support',
                ].map((item) => (
                  <li key={item} style={{ marginBottom: 'var(--space-3)', fontSize: 14, color: 'var(--text-dim)' }}>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>
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
                Need a custom compliance solution?
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                Our team can work with you to meet specific security and compliance requirements.
              </p>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'center', flexWrap: 'wrap' }}>
                <SealedButton onClick={() => window.location.href = '/contact'}>
                  Contact sales
                </SealedButton>
                <FrameButton onClick={() => window.location.href = '/trust'}>
                  Trust center
                </FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default CompliancePage

import React from 'react'
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
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

const LayersIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const UsersIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
    <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const ShieldCheckIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const CheckIcon: React.FC<{ size?: number }> = ({ size = 20 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const MailIcon: React.FC<{ size?: number }> = ({ size = 18 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="2" y="4" width="20" height="16" rx="2" stroke="currentColor" strokeWidth="2"/>
    <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const PartnershipsPage: React.FC = () => {
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
              Partnerships
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Build with<br />FunctionFly
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720 }}>
              Join our partner ecosystem and help shape the future of trusted AI agent infrastructure.
            </p>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <LayersIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Technology Partners
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Integrate FunctionFly into your platform. Ideal for AI frameworks, agent runtimes, and developer tools looking to add trusted execution.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {['API access and dedicated support', 'Joint go-to-market activities', 'Technical integration resources'].map((item) => (
                  <li key={item} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)', marginBottom: 'var(--space-2)', fontSize: 14, color: 'var(--text-dim)' }}>
                    <span style={{ color: 'var(--status-ok)', flexShrink: 0 }}>✓</span>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <UsersIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Reseller Partners
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Offer FunctionFly to your customers as part of a broader solution. Perfect for consultancies, system integrators, and SaaS platforms.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {['Volume pricing and margins', 'Sales enablement materials', 'Co-marketing opportunities'].map((item) => (
                  <li key={item} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)', marginBottom: 'var(--space-2)', fontSize: 14, color: 'var(--text-dim)' }}>
                    <span style={{ color: 'var(--status-ok)', flexShrink: 0 }}>✓</span>
                    {item}
                  </li>
                ))}
              </ul>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ShieldCheckIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Trust & Compliance Partners
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6, marginBottom: 'var(--space-4)' }}>
                Work with us on SOC 2, HIPAA, GDPR, and other compliance frameworks. Ideal for security vendors, auditors, and compliance consultants.
              </p>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {['Compliance documentation access', 'Audit coordination support', 'Shared compliance resources'].map((item) => (
                  <li key={item} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)', marginBottom: 'var(--space-2)', fontSize: 14, color: 'var(--text-dim)' }}>
                    <span style={{ color: 'var(--status-ok)', flexShrink: 0 }}>✓</span>
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
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <h2 style={{
              fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Why partner with FunctionFly?
            </h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-5)' }}>
              {[
                { title: 'Growing Market', desc: 'The AI agent infrastructure market is expanding rapidly. Partners benefit from this momentum.' },
                { title: 'Technical Excellence', desc: 'Our trust layer handles security, compliance, and reliability so you can focus on your core product.' },
                { title: 'Dedicated Support', desc: 'Every partner gets a dedicated account team and technical integration support.' },
                { title: 'Market Together', desc: 'Co-marketing campaigns, joint events, and cross-promotion to help you grow.' },
              ].map((benefit) => (
                <div key={benefit.title} style={{ display: 'flex', gap: 'var(--space-3)' }}>
                  <div style={{ color: 'var(--status-ok)', flexShrink: 0, marginTop: 2 }}>
                    <CheckIcon />
                  </div>
                  <div>
                    <h4 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-1)' }}>
                      {benefit.title}
                    </h4>
                    <p style={{ fontSize: 13, color: 'var(--text-dim)', margin: 0, lineHeight: 1.5 }}>
                      {benefit.desc}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </Chamber>
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
                Ready to partner?
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                Reach out to our partnerships team to discuss opportunities.
              </p>
              <SealedButton onClick={() => window.location.href = 'mailto:partnerships@functionfly.com?subject=Partnership Inquiry'}>
                <span style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                  <MailIcon /> partnerships@functionfly.com
                </span>
              </SealedButton>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default PartnershipsPage

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
import { Input } from './containment/Input'

const topicOptions = [
  { value: 'sales', label: 'Sales & Enterprise', email: 'sales@functionfly.com', responseTime: '24h', priority: 'high' as const },
  { value: 'support', label: 'Technical Support', email: 'support@functionfly.com', responseTime: '4h', priority: 'high' as const },
  { value: 'privacy', label: 'Privacy & Data Rights', email: 'privacy@functionfly.com', responseTime: '48h', priority: 'medium' as const },
  { value: 'security', label: 'Security', email: 'security@functionfly.com', responseTime: '24h', priority: 'medium' as const },
  { value: 'legal', label: 'Legal & Compliance', email: 'legal@functionfly.com', responseTime: '72h', priority: 'low' as const },
  { value: 'general', label: 'General & Partnerships', email: 'hello@functionfly.com', responseTime: '48h', priority: 'low' as const },
]

const subjectSuggestions: Record<string, string[]> = {
  sales: ['Enterprise pricing inquiry', 'Volume licensing request', 'Security questionnaire'],
  support: ['Production outage', 'API authentication issue', 'Billing question', 'Account access problem'],
  privacy: ['Data processing inquiry', 'Cookie preferences', 'GDPR request'],
  security: ['Vulnerability report', 'Security concern'],
  legal: ['Contract notice', 'Compliance question', 'Legal correspondence'],
  general: ['Partnership inquiry', 'Press inquiry', 'General question'],
}

const Container: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({ children, narrow }) => (
  <div style={{ maxWidth: narrow ? 560 : 1100, margin: '0 auto', padding: '0 var(--space-4)' }}>
    {children}
  </div>
)

const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>
    {children}
  </section>
)

const ContactPage: React.FC = () => {
  const [selectedTopic, setSelectedTopic] = useState('')
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    if (!selectedTopic || !email || !message) {
      setError('Please fill out all fields')
      return
    }
    const topic = topicOptions.find(t => t.value === selectedTopic)
    if (topic) {
      const subject = encodeURIComponent(message.split('\n')[0] || 'Contact form submission')
      window.location.href = `mailto:${topic.email}?subject=${subject}&body=${encodeURIComponent(message)}`
      setSubmitted(true)
    }
  }

  const selectedTopicData = topicOptions.find(t => t.value === selectedTopic)
  const suggestions = selectedTopic ? subjectSuggestions[selectedTopic] || [] : []

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
              Get in touch
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              We'd love to hear from you
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720 }}>
              We route each request to the right team. Choose the channel that best matches your topic so we can help you quickly.
            </p>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container narrow>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
              Send a message
            </h3>
            <p style={{ color: 'var(--text-dim)', fontSize: 14, marginBottom: 'var(--space-5)' }}>
              Select your topic and we'll route your message to the right team.
            </p>
            <form onSubmit={handleSubmit} noValidate>
              <div style={{ marginBottom: 'var(--space-4)' }}>
                <label htmlFor="topic" style={{ display: 'block', fontSize: 14, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Topic</label>
                <select
                  id="topic"
                  value={selectedTopic}
                  onChange={(e) => setSelectedTopic(e.target.value)}
                  required
                  style={{
                    width: '100%', padding: 'var(--space-3) var(--space-4)',
                    fontSize: 15, color: 'var(--text)',
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--steel)',
                    borderRadius: 'var(--radius)',
                    fontFamily: 'inherit',
                    appearance: 'none',
                    backgroundImage: "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%237c8a8a' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E\")",
                    backgroundRepeat: 'no-repeat',
                    backgroundPosition: 'right var(--space-4) center',
                    paddingRight: 'var(--space-7)',
                  }}
                >
                  <option value="">Select a topic...</option>
                  {topicOptions.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>

              {selectedTopicData && (
                <div style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  flexWrap: 'wrap', gap: 'var(--space-2)',
                  padding: 'var(--space-3) var(--space-4)',
                  background: 'rgba(143, 255, 208, 0.05)',
                  border: '1px solid var(--panel-edge)',
                  borderRadius: 'var(--radius)',
                  marginBottom: 'var(--space-4)',
                  fontSize: 13,
                }}>
                  <span style={{ color: 'var(--status-ok)', fontFamily: 'var(--font-mono)' }}>{selectedTopicData.email}</span>
                  <span style={{ color: 'var(--text-dim)', display: 'inline-flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <circle cx="12" cy="12" r="10" />
                      <polyline points="12 6 12 12 16 14" />
                    </svg>
                    Typically responds in {selectedTopicData.responseTime}
                  </span>
                </div>
              )}

              <div style={{ marginBottom: 'var(--space-4)' }}>
                <label htmlFor="email" style={{ display: 'block', fontSize: 14, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Your email</label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@company.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>

              <div style={{ marginBottom: 'var(--space-4)' }}>
                <label htmlFor="message" style={{ display: 'block', fontSize: 14, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Message</label>
                <textarea
                  id="message"
                  placeholder="Describe your question or issue..."
                  rows={5}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  required
                  style={{
                    width: '100%', padding: 'var(--space-3) var(--space-4)',
                    fontSize: 15, color: 'var(--text)',
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--steel)',
                    borderRadius: 'var(--radius)',
                    fontFamily: 'inherit', resize: 'vertical', minHeight: 120,
                  }}
                />
              </div>

              {suggestions.length > 0 && (
                <div style={{ marginBottom: 'var(--space-4)' }}>
                  <span style={{ display: 'block', fontSize: 12, color: 'var(--text-faint)', marginBottom: 'var(--space-2)' }}>
                    Suggested subject lines:
                  </span>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)' }}>
                    {suggestions.map(s => (
                      <button
                        key={s}
                        type="button"
                        onClick={() => setMessage(prev => prev ? `${s}\n\n${prev}` : s)}
                        style={{
                          padding: '4px 12px', fontSize: 12, color: 'var(--text-dim)',
                          background: 'var(--panel)',
                          border: '1px solid var(--panel-edge)',
                          borderRadius: 'var(--radius)',
                          cursor: 'pointer',
                          fontFamily: 'inherit',
                        }}
                      >{s}</button>
                    ))}
                  </div>
                </div>
              )}

              {error && (
                <div style={{ color: 'var(--status-revoked)', fontSize: 13, marginBottom: 'var(--space-3)' }}>
                  {error}
                </div>
              )}

              <SealedButton type="submit" size="lg">
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <line x1="22" y1="2" x2="11" y2="13" />
                    <polygon points="22 2 15 22 11 13 2 9 22 2" />
                  </svg>
                  Open in email client
                </span>
              </SealedButton>

              {submitted && (
                <div style={{ marginTop: 'var(--space-4)' }}>
                  <StatusPill status="live" label="Mail client opened" />
                </div>
              )}
            </form>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)', textAlign: 'center',
            marginBottom: 'var(--space-7)',
          }}>Other channels</h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--space-5)' }}>
            {topicOptions.map((t) => (
              <Card key={t.value}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)' }}>
                    {t.label}
                  </h3>
                  <span style={{
                    display: 'inline-flex', alignItems: 'center', gap: 4,
                    fontSize: 12, color: t.priority === 'high' ? 'var(--accent)' : 'var(--text-dim)',
                    border: `1px solid ${t.priority === 'high' ? 'var(--accent)' : 'var(--panel-edge)'}`,
                    background: t.priority === 'high' ? 'rgba(255, 122, 61, 0.1)' : 'var(--panel)',
                    padding: '2px 8px', borderRadius: 'var(--radius-sm)',
                  }}>
                    ~{t.responseTime} response
                  </span>
                </div>
                <a href={`mailto:${t.email}`} style={{ fontFamily: 'var(--font-mono)', fontSize: 14, color: 'var(--status-ok)' }}>{t.email}</a>
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
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-5)' }}>
              Self-service resources
            </h3>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 'var(--space-4)' }}>
              {[
                { href: '/docs', title: 'Documentation', desc: 'Guides, API reference, and tutorials' },
                { href: '/trust', title: 'Trust & Security', desc: 'Compliance, certifications, and policies' },
                { href: '/pricing', title: 'Pricing', desc: 'Plans, features, and comparisons' },
              ].map((r) => (
                <a key={r.href} href={r.href} style={{
                  display: 'flex', alignItems: 'center', gap: 'var(--space-4)',
                  padding: 'var(--space-4)',
                  background: 'var(--bg)',
                  border: '1px solid var(--panel-edge)',
                  borderRadius: 'var(--radius)',
                  textDecoration: 'none', color: 'inherit',
                  transition: 'border-color var(--duration-fast) var(--ease-out)',
                }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)', marginBottom: 4 }}>{r.title}</div>
                    <div style={{ fontSize: 13, color: 'var(--text-dim)' }}>{r.desc}</div>
                  </div>
                  <span style={{ color: 'var(--text-faint)' }}>→</span>
                </a>
              ))}
            </div>
            <div style={{ marginTop: 'var(--space-5)', display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <FrameButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/login` }}>Sign in</FrameButton>
              <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Create account</SealedButton>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default ContactPage

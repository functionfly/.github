import React, { useState } from 'react'
import { AUTH_ORIGIN } from '../config'
import './homepage.css'

const topicOptions = [
  { value: 'sales', label: 'Sales & Enterprise', email: 'sales@functionfly.com', responseTime: '24h', priority: 'high' },
  { value: 'support', label: 'Technical Support', email: 'support@functionfly.com', responseTime: '4h', priority: 'high' },
  { value: 'privacy', label: 'Privacy & Data Rights', email: 'privacy@functionfly.com', responseTime: '48h', priority: 'medium' },
  { value: 'security', label: 'Security', email: 'security@functionfly.com', responseTime: '24h', priority: 'medium' },
  { value: 'legal', label: 'Legal & Compliance', email: 'legal@functionfly.com', responseTime: '72h', priority: 'low' },
  { value: 'general', label: 'General & Partnerships', email: 'hello@functionfly.com', responseTime: '48h', priority: 'low' },
]

const subjectSuggestions: Record<string, string[]> = {
  sales: ['Enterprise pricing inquiry', 'Volume licensing request', 'Security questionnaire'],
  support: ['Production outage', 'API authentication issue', 'Billing question', 'Account access problem'],
  privacy: ['Data processing inquiry', 'Cookie preferences', 'GDPR request'],
  security: ['Vulnerability report', 'Security concern'],
  legal: ['Contract notice', 'Compliance question', 'Legal correspondence'],
  general: ['Partnership inquiry', 'Press inquiry', 'General question'],
}

const ContactPage: React.FC = () => {
  const [selectedTopic, setSelectedTopic] = useState('')
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedTopic || !email || !message) return
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
    <div className="ff-homepage">
      <section className="ff-hero-section ff-contact-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Get in touch</span>
          </div>
          <h1 className="ff-hero-headline">
            We'd love to hear from you
          </h1>
          <p className="ff-hero-sub">
            We route each request to the right team. Choose the channel that best matches your topic so we can help you quickly.
          </p>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-contact-form-wrapper">
            <form className="ff-contact-form" onSubmit={handleSubmit}>
              <div className="ff-form-header">
                <h3>Send a message</h3>
                <p>Select your topic and we'll route your message to the right team.</p>
              </div>
              <div className="ff-form-group">
                <label htmlFor="topic">Topic</label>
                <select
                  id="topic"
                  className="ff-select"
                  value={selectedTopic}
                  onChange={(e) => setSelectedTopic(e.target.value)}
                  required
                >
                  <option value="">Select a topic...</option>
                  {topicOptions.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              {selectedTopicData && (
                <div className="ff-topic-meta">
                  <span className="ff-topic-email">{selectedTopicData.email}</span>
                  <span className="ff-topic-response">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    Typically responds in {selectedTopicData.responseTime}
                  </span>
                </div>
              )}
              <div className="ff-form-group">
                <label htmlFor="email">Your email</label>
                <input
                  type="email"
                  id="email"
                  className="ff-input"
                  placeholder="you@company.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="ff-form-group">
                <label htmlFor="message">Message</label>
                <textarea
                  id="message"
                  className="ff-textarea"
                  placeholder="Describe your question or issue..."
                  rows={5}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  required
                />
              </div>
              {suggestions.length > 0 && (
                <div className="ff-subject-suggestions">
                  <span className="ff-suggestions-label">Suggested subject lines:</span>
                  <div className="ff-suggestions-list">
                    {suggestions.map(s => (
                      <button
                        key={s}
                        type="button"
                        className="ff-suggestion-chip"
                        onClick={() => setMessage(prev => prev ? `${s}\n\n${prev}` : s)}
                      >
                        {s}
                      </button>
                    ))}
                  </div>
                </div>
              )}
              <button type="submit" className="ff-btn ff-btn-primary ff-btn-lg ff-form-submit">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <line x1="22" y1="2" x2="11" y2="13" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <polygon points="22 2 15 22 11 13 2 9 22 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Open in email client
              </button>
            </form>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-contact-grid">
            <div className="ff-contact-card ff-contact-card--featured">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>Sales & Enterprise</h3>
                  <span className="ff-response-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~24h response
                  </span>
                </div>
                <p>Volume pricing, custom SLAs, security questionnaires, SSO, and tailored trust policies.</p>
                <a href="mailto:sales@functionfly.com" className="ff-contact-link">sales@functionfly.com</a>
                <p className="ff-contact-hint">See <a href="/pricing">Pricing</a> and <a href="/sla">SLA</a> for published plans.</p>
              </div>
            </div>

            <div className="ff-contact-card ff-contact-card--featured">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                  <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <line x1="12" y1="17" x2="12.01" y2="17" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>Technical Support</h3>
                  <span className="ff-response-badge ff-response-badge--urgent">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~4h response
                  </span>
                </div>
                <p>Account access, billing, API issues, and production-impacting problems.</p>
                <a href="mailto:support@functionfly.com" className="ff-contact-link">support@functionfly.com</a>
                <div className="ff-contact-actions">
                  <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/login`}>Sign in</a>
                  <a className="ff-btn ff-btn-secondary" href={`${AUTH_ORIGIN}/signup`}>Create account</a>
                </div>
              </div>
            </div>

            <div className="ff-contact-card">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>Privacy & Data Rights</h3>
                  <span className="ff-response-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~48h response
                  </span>
                </div>
                <p>Questions about data processing, cookie preferences, or privacy rights requests.</p>
                <a href="mailto:privacy@functionfly.com" className="ff-contact-link">privacy@functionfly.com</a>
                <p className="ff-contact-hint">See our <a href="/privacy">Privacy Policy</a> for details.</p>
              </div>
            </div>

            <div className="ff-contact-card">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>Security</h3>
                  <span className="ff-response-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~24h response
                  </span>
                </div>
                <p>Responsible disclosure of vulnerabilities. Provide enough detail to reproduce.</p>
                <a href="mailto:security@functionfly.com" className="ff-contact-link">security@functionfly.com</a>
              </div>
            </div>

            <div className="ff-contact-card">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
                  <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>Legal & Compliance</h3>
                  <span className="ff-response-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~72h response
                  </span>
                </div>
                <p>Contract notices, compliance questions, and legal correspondence.</p>
                <a href="mailto:legal@functionfly.com" className="ff-contact-link">legal@functionfly.com</a>
                <p className="ff-contact-hint">Copyright notices: <a href="mailto:copyright@functionfly.com">copyright@functionfly.com</a></p>
              </div>
            </div>

            <div className="ff-contact-card">
              <div className="ff-contact-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" stroke="currentColor" strokeWidth="2"/>
                  <polyline points="22,6 12,13 2,6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <div className="ff-contact-content">
                <div className="ff-contact-header">
                  <h3>General & Partnerships</h3>
                  <span className="ff-response-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                      <polyline points="12 6 12 12 16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    </svg>
                    ~48h response
                  </span>
                </div>
                <p>Press, partnerships, or just saying hello.</p>
                <a href="mailto:hello@functionfly.com" className="ff-contact-link">hello@functionfly.com</a>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-info-card">
            <h3>Self-service resources</h3>
            <div className="ff-resource-links">
              <a href="/docs" className="ff-resource-link ff-resource-link--card">
                <div className="ff-resource-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" stroke="currentColor" strokeWidth="2"/>
                    <line x1="8" y1="7" x2="16" y2="7" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <line x1="8" y1="11" x2="14" y2="11" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                </div>
                <div className="ff-resource-text">
                  <span className="ff-resource-title">Documentation</span>
                  <span className="ff-resource-desc">Guides, API reference, and tutorials</span>
                </div>
                <svg className="ff-resource-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <polyline points="9 18 15 12 9 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </a>
              <a href="/trust" className="ff-resource-link ff-resource-link--card">
                <div className="ff-resource-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                    <polyline points="9 12 11 14 15 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div className="ff-resource-text">
                  <span className="ff-resource-title">Trust & Security</span>
                  <span className="ff-resource-desc">Compliance, certifications, and policies</span>
                </div>
                <svg className="ff-resource-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <polyline points="9 18 15 12 9 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </a>
              <a href="/pricing" className="ff-resource-link ff-resource-link--card">
                <div className="ff-resource-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <line x1="12" y1="1" x2="12" y2="23" stroke="currentColor" strokeWidth="2"/>
                    <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                </div>
                <div className="ff-resource-text">
                  <span className="ff-resource-title">Pricing</span>
                  <span className="ff-resource-desc">Plans, features, and comparisons</span>
                </div>
                <svg className="ff-resource-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <polyline points="9 18 15 12 9 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </a>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default ContactPage

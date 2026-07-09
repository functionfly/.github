import React, { useState } from 'react'
import {
  Chamber,
  CornerBrace,
  StatusPill,
  Table,
  AnnotationTag,
} from './containment'
import { SealedButton, FrameButton } from './sc'
import '../styles/sc-main.css';

const ShieldIcon = () => (
  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const LockIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
    <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const DocumentIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
    <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="2"/>
  </svg>
)

const CookieIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
    <circle cx="8" cy="9" r="1.5" fill="currentColor"/>
    <circle cx="15" cy="8" r="1.5" fill="currentColor"/>
    <circle cx="10" cy="14" r="1.5" fill="currentColor"/>
    <circle cx="16" cy="14" r="1.5" fill="currentColor"/>
  </svg>
)

const Accordion: React.FC<{
  title: string;
  defaultOpen?: boolean;
  children: React.ReactNode;
}> = ({ title, defaultOpen = false, children }) => {
  const [isOpen, setIsOpen] = useState(defaultOpen)

  return (
    <Chamber nested className="trust-accordion">
      <button
        className="trust-accordion-header"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
      >
        <h3>{title}</h3>
        <span className={`trust-accordion-icon ${isOpen ? 'is-open' : ''}`} />
      </button>
      <div className={`trust-accordion-content ${isOpen ? 'is-open' : ''}`}>
        {children}
      </div>
    </Chamber>
  )
}

const PrivacyPage: React.FC = () => {
  const subprocessors = [
    { service: 'Stripe', purpose: 'Payment processing', data: 'Billing info, transaction metadata', location: 'US' },
    { service: 'OpenAI', purpose: 'AI completions & embeddings', data: 'Prompts, function metadata', location: 'US' },
    { service: 'Anthropic', purpose: 'AI completions (optional)', data: 'Prompts (if configured)', location: 'US' },
    { service: 'OpenRouter', purpose: 'AI model routing (100+ models)', data: 'Prompts, model responses', location: 'US' },
    { service: 'Cloudflare', purpose: 'CDN, security, DNS, R2 object storage', data: 'IP addresses, request logs, artifacts, backups', location: 'Global' },
    { service: 'Vercel', purpose: 'Frontend hosting', data: 'IP addresses, analytics', location: 'US/EU' },
    { service: 'Resend', purpose: 'Transactional email', data: 'Email addresses, message content', location: 'US' },
    { service: 'Sentry', purpose: 'Error monitoring', data: 'Error logs, stack traces, IP addresses', location: 'US' },
    { service: 'Upstash', purpose: 'Redis caching (serverless)', data: 'Session data, cache entries', location: 'Global' },
    { service: 'Mailchimp', purpose: 'Newsletter & marketing emails', data: 'Email addresses, subscriber status', location: 'US' },
    { service: 'Mixpanel', purpose: 'Product analytics', data: 'User interactions, event properties', location: 'US' },
    { service: 'Sanity', purpose: 'CMS for blog & content', data: 'Blog posts, author information', location: 'US' },
    { service: 'Google Analytics', purpose: 'Website analytics', data: 'Page views, session data', location: 'US' },
    { service: 'Atlas', purpose: 'Agent execution tracing & observability (optional)', data: 'Execution traces, agent prompts, function calls', location: 'US' },
  ]

  return (
    <div>
      {/* Hero Section */}
      <section className="trust-hero">
        <Chamber variant="ribs" className="trust-hero-chamber">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag title="LEGAL / PRIVACY" subtitle="Page 01" />
          <div className="trust-hero-content">
            <div className="trust-hero-eyebrow">
              <span className="trust-pulse-dot" />
              <span>Privacy & Cookie Policy</span>
            </div>
            <h1 className="trust-hero-title">
              Your privacy<br />is our foundation
            </h1>
            <p className="trust-hero-subtitle">
              How we collect, use, and protect your data; cookie practices and your rights.
            </p>
            <div className="trust-hero-actions">
              <a href="#contact">
                <SealedButton>Contact Privacy Team</SealedButton>
              </a>
            </div>
          </div>
          <div className="trust-hero-visual">
            <ShieldIcon />
          </div>
        </Chamber>
      </section>

      {/* Compliance Badges */}
      <section className="trust-section">
        <div className="trust-container">
          <div className="trust-compliance-badges">
            <StatusPill status="live" label="GDPR" />
            <StatusPill status="live" label="CCPA" />
            <StatusPill status="live" label="SOC 2" />
          </div>
        </div>
      </section>

      {/* Cookie Preferences */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Cookie preferences</h2>
          <p className="trust-section-lead">
            Manage your cookie settings in the dashboard. Essential cookies are always active.
          </p>
          <div className="trust-grid-2">
            <Chamber nested className="trust-cookie-card">
              <div className="trust-cookie-header">
                <LockIcon />
                <strong>Necessary cookies</strong>
                <StatusPill status="live" label="Always active" />
              </div>
              <p className="trust-muted">Required for the website to function properly. These cannot be disabled.</p>
              <p className="trust-cookie-meta">Purpose: Session management, authentication, security &nbsp;&middot;&nbsp; Duration: Session / 1 year</p>
            </Chamber>

            <Chamber nested className="trust-cookie-card">
              <div className="trust-cookie-header">
                <CookieIcon />
                <strong>Analytics cookies</strong>
                <StatusPill status="pending" label="Opt-in" />
              </div>
              <p className="trust-muted">Help us understand how visitors interact with our website.</p>
              <p className="trust-cookie-meta">Purpose: Usage analytics, performance monitoring &nbsp;&middot;&nbsp; Duration: Up to 2 years</p>
            </Chamber>

            <Chamber nested className="trust-cookie-card">
              <div className="trust-cookie-header">
                <CookieIcon />
                <strong>Marketing cookies</strong>
                <StatusPill status="pending" label="Opt-in" />
              </div>
              <p className="trust-muted">Used to deliver personalized advertisements.</p>
              <p className="trust-cookie-meta">Purpose: Targeted advertising, retargeting &nbsp;&middot;&nbsp; Duration: Up to 1 year</p>
            </Chamber>

            <Chamber nested className="trust-cookie-card">
              <div className="trust-cookie-header">
                <CookieIcon />
                <strong>Functionality cookies</strong>
                <StatusPill status="pending" label="Opt-in" />
              </div>
              <p className="trust-muted">Enable enhanced functionality and personalization.</p>
              <p className="trust-cookie-meta">Purpose: Language preferences, UI customization &nbsp;&middot;&nbsp; Duration: Up to 1 year</p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Privacy Policy */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Privacy Policy</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Data controller</h3>
              <p className="trust-muted">
                FunctionFly™ LLC (d/b/a FunctionFly™), a Wyoming limited liability company with principal operations in Fort Worth, Texas, United States.
              </p>
              <p className="trust-muted">
                These disclosures supplement our <a href="/terms" className="trust-link">Terms of Service</a>.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>Sensitive data</h3>
              <p className="trust-muted">
                Do not submit PCI data, PHI, or sensitive government-issued identifiers through the Service unless we have explicitly agreed to support that data in writing.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Information We Collect */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Information we collect</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Account & usage data</h3>
              <ul className="trust-list">
                <li>Account and profile information (email, name, company)</li>
                <li>Billing and transaction metadata</li>
                <li>Function and execution data</li>
                <li>Agent and API usage</li>
                <li>Device, browser, and network information</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>Vault & AI data</h3>
              <ul className="trust-list">
                <li><strong>Vault data:</strong> Encrypted ciphertext only. We never see your passphrase.</li>
                <li><strong>AI interactions:</strong> Prompts, messages, and context sent through AI features.</li>
                <li><strong>Support communications:</strong> Messages and attachments you send to us.</li>
              </ul>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Legal Bases */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Legal bases for processing (EEA/UK)</h2>
          <div className="trust-grid-4">
            <Chamber nested className="trust-legal-card">
              <strong>Contract</strong>
              <p className="trust-muted">To provide the services you request</p>
            </Chamber>
            <Chamber nested className="trust-legal-card">
              <strong>Legitimate interests</strong>
              <p className="trust-muted">To secure and improve the Service</p>
            </Chamber>
            <Chamber nested className="trust-legal-card">
              <strong>Legal obligations</strong>
              <p className="trust-muted">To comply with applicable laws</p>
            </Chamber>
            <Chamber nested className="trust-legal-card">
              <strong>Consent</strong>
              <p className="trust-muted">Where required, withdrawable at any time</p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Data Retention */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Data retention</h2>
          <div className="trust-grid-3">
            <Chamber nested className="trust-retention-card">
              <strong>Account data</strong>
              <p className="trust-muted">Active account + 3 years after deletion</p>
            </Chamber>
            <Chamber nested className="trust-retention-card">
              <strong>Usage & analytics</strong>
              <p className="trust-muted">Up to 2 years</p>
            </Chamber>
            <Chamber nested className="trust-retention-card">
              <strong>Execution logs</strong>
              <p className="trust-muted">Varies by type; up to 7 years for financial</p>
            </Chamber>
            <Chamber nested className="trust-retention-card">
              <strong>Support records</strong>
              <p className="trust-muted">3 years</p>
            </Chamber>
            <Chamber nested className="trust-retention-card">
              <strong>Marketing data</strong>
              <p className="trust-muted">1 year after last interaction</p>
            </Chamber>
            <Chamber nested className="trust-retention-card">
              <strong>Longer retention</strong>
              <p className="trust-muted">When required by law or legal proceedings</p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Subprocessor List */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Subprocessor list</h2>
          <p className="trust-section-lead">
            Last reviewed: July 7, 2026. We update this list when adding or changing subprocessors.
          </p>
          <Chamber className="trust-table-chamber">
            <div className="trust-table-wrapper">
              <table className="trust-table">
                <thead>
                  <tr>
                    <th>Service</th>
                    <th>Purpose</th>
                    <th>Data Processed</th>
                    <th>Location</th>
                  </tr>
                </thead>
                <tbody>
                  {subprocessors.map((sp) => (
                    <tr key={sp.service}>
                      <td><strong>{sp.service}</strong></td>
                      <td>{sp.purpose}</td>
                      <td>{sp.data}</td>
                      <td>{sp.location}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Chamber>
        </div>
      </section>

      {/* Your Rights */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Your rights</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>GDPR rights (EU users)</h3>
              <ul className="trust-list">
                <li>Right to access your personal data</li>
                <li>Right to rectification of inaccurate data</li>
                <li>Right to erasure ("right to be forgotten")</li>
                <li>Right to restrict processing</li>
                <li>Right to data portability</li>
                <li>Right to object to processing</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>CCPA rights (California users)</h3>
              <ul className="trust-list">
                <li>Right to know what personal information is collected</li>
                <li>Right to know if personal information is sold or shared</li>
                <li>Right to opt-out of the sale of personal information</li>
                <li>Right to delete personal information</li>
                <li>Right to non-discrimination for exercising CCPA rights</li>
              </ul>
            </Chamber>
          </div>
          <div className="trust-callout" style={{ marginTop: 'var(--space-5)' }}>
            <div className="trust-callout-title">Exercising your rights</div>
            <div className="trust-callout-body">
              Contact us at <a href="mailto:privacy@functionfly.com" className="trust-link">privacy@functionfly.com</a>. We typically respond within 30 to 45 days.
            </div>
          </div>
        </div>
      </section>

      {/* Security */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Security measures</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Encryption</h3>
              <ul className="trust-list">
                <li><strong>Data in transit:</strong> TLS 1.3</li>
                <li><strong>Data at rest:</strong> AES-256</li>
                <li><strong>Database encryption:</strong> Transparent encryption</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>Access controls</h3>
              <ul className="trust-list">
                <li>Role-based access control (RBAC)</li>
                <li>Multi-factor authentication (MFA)</li>
                <li>Regular access reviews</li>
                <li>Principle of least privilege</li>
              </ul>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Zero-Knowledge Vault */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <Chamber className="trust-vault-chamber">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="trust-vault-content">
              <h2>Zero-knowledge vault</h2>
              <p className="trust-muted">
                Secrets you store in the FunctionFly™ Vault are encrypted client-side. We process ciphertext and metadata needed to operate the feature; we do not receive your vault passphrase and cannot decrypt your secrets by design.
              </p>
            </div>
          </Chamber>
        </div>
      </section>

      {/* Contact */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title" id="contact">Contact us</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Privacy inquiries</h3>
              <p className="trust-muted">
                <strong>Legal entity:</strong> FunctionFly™ LLC (d/b/a FunctionFly™), Wyoming (operations: Fort Worth, Texas, United States)
              </p>
              <p className="trust-muted">
                <strong>Email:</strong> <a href="mailto:privacy@functionfly.com" className="trust-link">privacy@functionfly.com</a>
              </p>
              <p className="trust-muted">
                We typically respond within 2 business days.
              </p>
            </Chamber>

            <Chamber className="trust-card" id="report">
              <CornerBrace position="bl" />
              <h3>Report a privacy concern</h3>
              <p className="trust-muted">
                If you believe we've mishandled your data, have a security concern, or need to report a privacy incident, please contact us directly.
              </p>
              <a href="mailto:privacy@functionfly.com?subject=URGENT: Privacy Concern Report">
                <SealedButton>Report Privacy Concern</SealedButton>
              </a>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Footer */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Related pages</h2>
          <div className="trust-grid-4">
            <a href="/terms" className="trust-resource-link">
              <DocumentIcon />
              <span>Terms of Service</span>
            </a>
            <a href="/trust" className="trust-resource-link">
              <ShieldIcon />
              <span>Trust Center</span>
            </a>
            <a href="/compliance" className="trust-resource-link">
              <ShieldIcon />
              <span>Compliance</span>
            </a>
            <a href="/.well-known/security.txt" className="trust-resource-link">
              <LockIcon />
              <span>Security Policy</span>
            </a>
          </div>
          <p className="trust-footer-note">
            Last updated: July 7, 2026
          </p>
        </div>
      </section>
    </div>
  )
}

export default PrivacyPage
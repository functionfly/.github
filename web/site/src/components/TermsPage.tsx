import React from 'react'
import {
  Chamber,
  CornerBrace,
  AnnotationTag,
} from './containment'
import './TrustPage.css'

const DocumentIcon = () => (
  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
    <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="2"/>
    <line x1="8" y1="13" x2="16" y2="13" stroke="currentColor" strokeWidth="2"/>
    <line x1="8" y1="17" x2="16" y2="17" stroke="currentColor" strokeWidth="2"/>
  </svg>
)

const TermsPage: React.FC = () => {
  return (
    <div>
      {/* Hero Section */}
      <section className="trust-hero">
        <Chamber variant="ribs" className="trust-hero-chamber">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag title="LEGAL / TERMS" subtitle="Page 01" />
          <div className="trust-hero-content">
            <div className="trust-hero-eyebrow">
              <span className="trust-pulse-dot" />
              <span>Terms of Service</span>
            </div>
            <h1 className="trust-hero-title">
              Terms of<br />Service
            </h1>
            <p className="trust-hero-subtitle">
              The legal agreement governing your access to and use of FunctionFly services.
            </p>
          </div>
          <div className="trust-hero-visual">
            <DocumentIcon />
          </div>
        </Chamber>
      </section>

      {/* Overview */}
      <section className="trust-section">
        <div className="trust-container">
          <Chamber className="trust-card">
            <CornerBrace position="tr" />
            <p className="trust-muted" style={{ marginBottom: 'var(--space-4)' }}>
              These Terms together with our <a href="/privacy" className="trust-link">Privacy Policy</a> govern your access to and use of FunctionFly™.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Section 1-4 */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>1. Acceptance of terms</h2>
              <p className="trust-muted">
                These Terms are between you and FunctionFly™ LLC. By accessing or using the Service, you agree to be bound by these Terms. If you do not agree, you may not access or use the Service.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>2. The Service</h2>
              <p className="trust-muted">
                FunctionFly™ provides a cloud platform to publish, discover, run, and manage serverless functions and related developer tools.
              </p>
              <ul className="trust-list">
                <li>Function registry, versioning, execution, and multi-runtime deployment</li>
                <li>Dashboard, CLI, SDKs, APIs, and webhooks</li>
                <li>AI agent identities, policies, quotas, and usage-based billing</li>
                <li>Trust, verification, and analytics surfaces</li>
                <li>Secrets vault with client-side encryption</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>3. Accounts and security</h2>
              <p className="trust-muted">You agree to:</p>
              <ul className="trust-list">
                <li>Provide accurate, current, complete information</li>
                <li>Maintain password and API key confidentiality</li>
                <li>Accept responsibility for account activities</li>
                <li>Notify us of unauthorized access promptly</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>4. Acceptable use</h2>
              <p className="trust-muted">You agree not to use the Service to:</p>
              <ul className="trust-list">
                <li>Violate applicable laws or regulations</li>
                <li>Infringe rights or harass, threaten, or defraud</li>
                <li>Distribute malware or attack systems</li>
                <li>Mine cryptocurrency or operate botnets</li>
                <li>Scrape data in violation of our limits</li>
              </ul>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Section 5-8 */}
      <section className="trust-section">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>5. Your content and license</h2>
              <p className="trust-muted">
                "Your Content" means functions, code, configuration, prompts, and other materials you submit. You retain ownership subject to the licenses below.
              </p>
              <ul className="trust-list">
                <li>Grant us a license to host and process Your Content to provide the Service</li>
                <li>Grant end users rights to discover and execute published functions</li>
              </ul>
              <p className="trust-muted">
                You are responsible for Your Content and its consequences.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>6. Third-party services</h2>
              <p className="trust-muted">
                The Service integrates with payment processors (Stripe), cloud providers, email and analytics vendors. Their use of data is described in our Privacy Policy.
              </p>
              <p className="trust-muted">
                You are responsible for your integrations and third-party relationships.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>7. APIs, usage, and charges</h2>
              <p className="trust-muted">
                API access and executions are subject to plan limits, fair-use policies, and quotas. We may throttle use that risks stability or security.
              </p>
              <p className="trust-muted">
                You are responsible for configuring agent policies and budgets. Agents may invoke functions that incur fees.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>8. Secrets vault</h2>
              <p className="trust-muted">
                Vault features may encrypt secrets client-side before they reach us. We do not have your passphrase or the ability to decrypt by design.
              </p>
              <p className="trust-muted">
                You are responsible for safeguarding passphrases and recovery processes.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Section 9-12 */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>9. Trust and verification</h2>
              <p className="trust-muted">
                Trust scores, verification badges, and certificates are provided for transparency. They are not a warranty and are not a substitute for your own testing and risk management.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>10. Intellectual property</h2>
              <p className="trust-muted">
                The Service and its content are owned by FunctionFly™ LLC and protected by IP laws. We reserve all rights not expressly granted.
              </p>
              <p className="trust-muted">
                <strong>DMCA:</strong> Send copyright notices to <a href="mailto:copyright@functionfly.com" className="trust-link">copyright@functionfly.com</a>
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>11. Fees and subscriptions</h2>
              <ul className="trust-list">
                <li>Pay all fees, taxes, and charges shown at checkout</li>
                <li>Subscriptions renew automatically until canceled</li>
                <li>Fees are non-refundable except where required by law</li>
                <li>We may change prices with reasonable notice</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>12. Beta features</h2>
              <p className="trust-muted">
                Experimental or preview features are provided "as is" and may change or end at any time. They may not be covered by SLA commitments.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Section 13-16 */}
      <section className="trust-section">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>13. Service level</h2>
              <p className="trust-muted">
                We strive for high availability but do not guarantee uninterrupted access. See <a href="/sla" className="trust-link">/sla</a> for plan-specific uptime commitments.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>14. Privacy</h2>
              <p className="trust-muted">
                Your use is governed by our <a href="/privacy" className="trust-link">Privacy Policy</a>, describing how we collect, use, and protect personal information.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>15. Termination</h2>
              <p className="trust-muted">
                You may stop using the Service and close your account. We may suspend or terminate access for breach, risk, or if we discontinue the Service.
              </p>
              <p className="trust-muted">
                Sections that should survive will survive termination.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>16. Disclaimer</h2>
              <p className="trust-muted">
                TO THE MAXIMUM EXTENT PERMITTED BY LAW, THE SERVICE IS PROVIDED "AS IS" WITHOUT WARRANTIES OF ANY KIND. WE DO NOT WARRANT THAT IT WILL BE ERROR-FREE OR SUITABLE FOR HIGH-RISK USE.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Section 17-20 */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>17. Limitation of liability</h2>
              <p className="trust-muted">
                FUNCTIONFLY™ SHALL NOT BE LIABLE FOR INDIRECT, INCIDENTAL, SPECIAL, OR CONSEQUENTIAL DAMAGES.
              </p>
              <p className="trust-muted">
                Total liability shall not exceed the greater of (a) fees paid in the prior 12 months, or (b) $10,000.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>18. Indemnification</h2>
              <p className="trust-muted">
                You will defend and indemnify FunctionFly™ from claims arising from Your Content, your violation of Terms, or your use of third-party services.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>19. Changes to terms</h2>
              <p className="trust-muted">
                We may modify these Terms. Continued use after changes constitutes acceptance. We will post updated Terms with a revised "Last updated" date.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>20. Export; government</h2>
              <p className="trust-muted">
                You may not use the Service in violation of export control laws. For U.S. government users, the Service is commercial computer software.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Section 21-23 */}
      <section className="trust-section">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>21. General provisions</h2>
              <ul className="trust-list">
                <li><strong>Assignment:</strong> You may not assign without consent</li>
                <li><strong>Order forms:</strong> Control over these Terms if they conflict</li>
                <li><strong>Entire agreement:</strong> Terms, Privacy Policy, and order forms</li>
                <li><strong>Severability:</strong> Unenforceable provisions do not affect others</li>
                <li><strong>No waiver:</strong> Failure to enforce is not a waiver</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>22. Governing law</h2>
              <p className="trust-muted">
                These Terms are governed by Wyoming law. Any legal action shall be brought exclusively in Wyoming courts, and you consent to personal jurisdiction there.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Contact */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <Chamber className="trust-card" style={{ maxWidth: '720px', margin: '0 auto' }}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <h2>23. Contact</h2>
            <p className="trust-muted">
              FunctionFly™ LLC operates from Fort Worth, Texas, United States. Contracting entity remains FunctionFly™ LLC, a Wyoming limited liability company.
            </p>
            <p className="trust-muted">
              <strong>Email:</strong> <a href="mailto:legal@functionfly.com" className="trust-link">legal@functionfly.com</a>
            </p>
            <p className="trust-muted">
              <a href="/contact" className="trust-link">Contact page</a>
            </p>
            <p className="trust-muted" style={{ marginTop: 'var(--space-5)', fontStyle: 'italic' }}>
              This document is for information about our product policies, not legal advice.
            </p>
            <p className="trust-footer-note">
              Last updated: March 24, 2026
            </p>
          </Chamber>
        </div>
      </section>
    </div>
  )
}

export default TermsPage
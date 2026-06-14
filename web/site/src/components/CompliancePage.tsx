import React from 'react'
import './homepage.css'

const CompliancePage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-compliance-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Security & Compliance</span>
          </div>
          <h1 className="ff-hero-headline">
            Enterprise-grade security<br />and compliance
          </h1>
          <p className="ff-hero-sub">
            FunctionFly is built with security-first principles. Our platform meets the most demanding compliance requirements for enterprise deployments.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href="/contact">Contact sales</a>
            <a className="ff-btn ff-btn-secondary" href="/trust">Trust center</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pillars">
            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>SOC 2 Type II</h2>
              <p>
                Our security controls and practices are audited annually by independent third-party auditors to maintain SOC 2 Type II certification.
              </p>
            </div>

            <div className="ff-pillar-card ff-pillar-card--accent">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>GDPR Compliant</h2>
              <p>
                We comply with the General Data Protection Regulation for EU users, including data subject rights, legal bases for processing, and international data transfers.
              </p>
            </div>

            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>CCPA Compliant</h2>
              <p>
                California residents have rights to know what personal information is collected, opt-out of sale, and request deletion under the California Consumer Privacy Act.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Security architecture</h2>
          <p className="ff-section-lead">
            Built from the ground up with security in mind.
          </p>
          <div className="ff-trust-grid-2">
            <div className="ff-trust-card">
              <h3>Encryption</h3>
              <ul>
                <li><strong>Data in transit:</strong> TLS 1.3 for all connections</li>
                <li><strong>Data at rest:</strong> AES-256 encryption</li>
                <li><strong>Database encryption:</strong> Transparent data encryption for sensitive information</li>
                <li><strong>Zero-knowledge vault:</strong> Client-side encryption, we never see plaintext</li>
              </ul>
            </div>
            <div className="ff-trust-card">
              <h3>Access controls</h3>
              <ul>
                <li><strong>Role-based access control (RBAC)</strong> limiting data access to authorized personnel</li>
                <li><strong>Multi-factor authentication (MFA)</strong> for all administrative access</li>
                <li><strong>Regular access reviews</strong> and automated deprovisioning</li>
                <li><strong>Principle of least privilege</strong> applied to all systems</li>
              </ul>
            </div>
            <div className="ff-trust-card">
              <h3>Security monitoring</h3>
              <ul>
                <li><strong>24/7 security monitoring</strong> and intrusion detection systems</li>
                <li><strong>Regular security audits</strong> and vulnerability assessments</li>
                <li><strong>Automated threat detection</strong> and response systems</li>
                <li><strong>SIEM integration</strong> for enterprise deployments</li>
              </ul>
            </div>
            <div className="ff-trust-card">
              <h3>Breach response</h3>
              <ul>
                <li><strong>Incident response procedures</strong> with immediate containment</li>
                <li><strong>72-hour notification</strong> when legally required</li>
                <li><strong>Comprehensive breach logs</strong> maintained</li>
                <li><strong>Post-incident reviews</strong> to prevent future occurrences</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Compliance certifications</h2>
          <div className="ff-compliance-grid">
            <div className="ff-compliance-card">
              <div className="ff-compliance-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h3>SOC 2 Type II</h3>
              <p>Annual audit by independent third-party auditors covering security, availability, and confidentiality.</p>
            </div>
            <div className="ff-compliance-card">
              <div className="ff-compliance-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                  <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h3>GDPR</h3>
              <p>Full compliance with EU General Data Protection Regulation including data subject rights and DPA agreements.</p>
            </div>
            <div className="ff-compliance-card">
              <div className="ff-compliance-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                </svg>
              </div>
              <h3>CCPA</h3>
              <p>California Consumer Privacy Act compliance with right to know, delete, and opt-out.</p>
            </div>
            <div className="ff-compliance-card">
              <div className="ff-compliance-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <rect x="2" y="7" width="20" height="14" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
                  <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" stroke="currentColor" strokeWidth="2"/>
                </svg>
              </div>
              <h3>COPPA</h3>
              <p>Children's Online Privacy Protection Act compliant with age verification and parental consent procedures.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Data residency & transfers</h2>
          <p className="ff-section-lead">
            Your data is stored and processed with appropriate safeguards.
          </p>
          <div className="ff-trust-grid-2">
            <div className="ff-trust-card">
              <h3>Data transfer mechanisms</h3>
              <ul>
                <li><strong>Standard Contractual Clauses (SCCs)</strong> approved by the European Commission</li>
                <li><strong>Adequacy decisions</strong> for transfers to countries with equivalent privacy protections</li>
                <li><strong>Binding Corporate Rules</strong> for intra-group transfers</li>
                <li><strong>Explicit consent</strong> when required by applicable law</li>
              </ul>
            </div>
            <div className="ff-trust-card">
              <h3>Data hosting locations</h3>
              <ul>
                <li><strong>United States</strong> — Primary deployment region</li>
                <li><strong>European Union</strong> — GDPR-compliant data centers</li>
                <li><strong>Global edge</strong> — Cloudflare CDN for low-latency delivery</li>
                <li><strong>Enterprise options</strong> — Dedicated deployments available</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Enterprise features</h2>
          <div className="ff-trust-grid-2">
            <div className="ff-trust-card">
              <h3>Custom security requirements</h3>
              <ul>
                <li>Private deployments in your cloud environment</li>
                <li>Customer-managed encryption keys (BYOK)</li>
                <li>Custom SLA with guaranteed uptime</li>
                <li>Dedicated infrastructure options</li>
              </ul>
            </div>
            <div className="ff-trust-card">
              <h3>Enterprise support</h3>
              <ul>
                <li>Dedicated customer success manager</li>
                <li>Priority support with guaranteed response times</li>
                <li>Custom onboarding and training</li>
                <li>Security questionnaires and audit support</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Need a custom compliance solution?</h2>
          <p>Our team can work with you to meet specific security and compliance requirements.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href="/contact">Contact sales</a>
            <a className="ff-btn ff-btn-secondary" href="/trust">Trust center</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default CompliancePage

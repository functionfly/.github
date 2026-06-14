import React from 'react'
import './homepage.css'

const SecurityPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-security-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Security Hub</span>
          </div>
          <h1 className="ff-hero-headline">
            Your security is<br />our foundation
          </h1>
          <p className="ff-hero-sub">
            From encryption to access controls, see how FunctionFly protects your data and how to report vulnerabilities responsibly.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href="/vulnerability">Report a vulnerability</a>
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
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Encryption</h2>
              <p>
                All data is encrypted in transit with TLS 1.3 and at rest with AES-256. The zero-knowledge vault extends this to client-side encryption.
              </p>
            </div>

            <div className="ff-pillar-card ff-pillar-card--accent">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Access controls</h2>
              <p>
                Role-based access control, multi-factor authentication for admins, and least-privilege principles across all systems.
              </p>
            </div>

            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                  <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Monitoring</h2>
              <p>
                24/7 security monitoring, intrusion detection, automated threat response, and SIEM integration for enterprise deployments.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Security measures</h2>
          <div className="ff-security-grid">
            <div className="ff-security-card">
              <h3>Encryption at rest</h3>
              <ul>
                <li>AES-256 encryption for all stored data</li>
                <li>Transparent database encryption for sensitive information</li>
                <li>Encrypted backups with separate key management</li>
              </ul>
            </div>
            <div className="ff-security-card">
              <h3>Encryption in transit</h3>
              <ul>
                <li>TLS 1.3 for all connections</li>
                <li>Certificate pinning for API endpoints</li>
                <li>Perfect forward secrecy enabled</li>
              </ul>
            </div>
            <div className="ff-security-card">
              <h3>Zero-knowledge vault</h3>
              <ul>
                <li>Client-side encryption of secrets</li>
                <li>Platform never sees plaintext passphrases</li>
                <li>You control who can decrypt</li>
              </ul>
            </div>
            <div className="ff-security-card">
              <h3>Access management</h3>
              <ul>
                <li>Role-based access control (RBAC)</li>
                <li>Multi-factor authentication (MFA)</li>
                <li>Session management with automatic timeout</li>
              </ul>
            </div>
            <div className="ff-security-card">
              <h3>Network security</h3>
              <ul>
                <li>VPC isolation for production systems</li>
                <li>Cloudflare DDoS protection</li>
                <li>Firewall rules and network segmentation</li>
              </ul>
            </div>
            <div className="ff-security-card">
              <h3>Compliance</h3>
              <ul>
                <li>SOC 2 Type II certified</li>
                <li>GDPR and CCPA compliant</li>
                <li>COPPA compliant</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Responsible disclosure</h2>
          <div className="ff-info-card">
            <p>
              We are committed to working with the security community to verify, reproduce, and respond to security vulnerabilities in a timely manner.
            </p>
            <h3>How to report</h3>
            <p>
              Please email <a href="mailto:security@functionfly.com">security@functionfly.com</a> with details about the vulnerability. Include enough information to reproduce the issue — a proof of concept is helpful but not required.
            </p>
            <h3>What we commit to</h3>
            <ul>
              <li>Acknowledgment within 24 hours</li>
              <li>Regular updates on remediation progress</li>
              <li>Public credit in our hall of fame (with your permission)</li>
            </ul>
            <h3>Scope</h3>
            <p>
              In-scope: production systems, the FunctionFly API, dashboard, and documented public integrations. Out of scope: social engineering, physical security, denial of service attacks against our infrastructure.
            </p>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Security resources</h2>
          <div className="ff-resource-grid">
            <a href="/.well-known/security.txt" className="ff-resource-card">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
                <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="2"/>
              </svg>
              <div>
                <h4>security.txt</h4>
                <p>Our RFC 9116 security contact and policy file</p>
              </div>
            </a>
            <a href="/trust" className="ff-resource-card">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
              </svg>
              <div>
                <h4>Trust Center</h4>
                <p>Verification, attestations, and compliance details</p>
              </div>
            </a>
            <a href="/privacy" className="ff-resource-card">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              </svg>
              <div>
                <h4>Privacy Policy</h4>
                <p>How we collect, use, and protect your data</p>
              </div>
            </a>
            <a href="/compliance" className="ff-resource-card">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
              </svg>
              <div>
                <h4>Compliance</h4>
                <p>Enterprise security and compliance certifications</p>
              </div>
            </a>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Have a security concern?</h2>
          <p>For vulnerabilities or security incidents, email us immediately.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href="mailto:security@functionfly.com">security@functionfly.com</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default SecurityPage

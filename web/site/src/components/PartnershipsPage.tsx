import React from 'react'
import './homepage.css'

const PartnershipsPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-partnerships-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Partnerships</span>
          </div>
          <h1 className="ff-hero-headline">
            Build with<br />FunctionFly
          </h1>
          <p className="ff-hero-sub">
            Join our partner ecosystem and help shape the future of trusted AI agent infrastructure.
          </p>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-partnerships-grid">
            <div className="ff-partnership-card">
              <div className="ff-partnership-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h3>Technology Partners</h3>
              <p>Integrate FunctionFly into your platform. Ideal for AI frameworks, agent runtimes, and developer tools looking to add trusted execution.</p>
              <ul className="ff-partnership-features">
                <li>API access and dedicated support</li>
                <li>Joint go-to-market activities</li>
                <li>Technical integration resources</li>
              </ul>
            </div>

            <div className="ff-partnership-card">
              <div className="ff-partnership-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <path d="M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h3>Reseller Partners</h3>
              <p>Offer FunctionFly to your customers as part of a broader solution. Perfect for consultancies, system integrators, and SaaS platforms.</p>
              <ul className="ff-partnership-features">
                <li>Volume pricing and margins</li>
                <li>Sales enablement materials</li>
                <li>Co-marketing opportunities</li>
              </ul>
            </div>

            <div className="ff-partnership-card">
              <div className="ff-partnership-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <polyline points="9 12 11 14 15 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h3>Trust & Compliance Partners</h3>
              <p>Work with us on SOC 2, HIPAA, GDPR, and other compliance frameworks. Ideal for security vendors, auditors, and compliance consultants.</p>
              <ul className="ff-partnership-features">
                <li>Compliance documentation access</li>
                <li>Audit coordination support</li>
                <li>Shared compliance resources</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-info-card">
            <h3>Why partner with FunctionFly?</h3>
            <div className="ff-benefits-grid">
              <div className="ff-benefit-item">
                <div className="ff-benefit-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div>
                  <h4>Growing Market</h4>
                  <p>The AI agent infrastructure market is expanding rapidly. Partners benefit from this momentum.</p>
                </div>
              </div>
              <div className="ff-benefit-item">
                <div className="ff-benefit-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div>
                  <h4>Technical Excellence</h4>
                  <p>Our trust layer handles security, compliance, and reliability so you can focus on your core product.</p>
                </div>
              </div>
              <div className="ff-benefit-item">
                <div className="ff-benefit-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div>
                  <h4>Dedicated Support</h4>
                  <p>Every partner gets a dedicated account team and technical integration support.</p>
                </div>
              </div>
              <div className="ff-benefit-item">
                <div className="ff-benefit-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div>
                  <h4>Market Together</h4>
                  <p>Co-marketing campaigns, joint events, and cross-promotion to help you grow.</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-partnership-cta">
            <h2>Ready to partner?</h2>
            <p>Reach out to our partnerships team to discuss opportunities.</p>
            <div className="ff-actions">
              <a className="ff-btn ff-btn-primary ff-btn-lg" href="mailto:partnerships@functionfly.com?subject=Partnership Inquiry">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" stroke="currentColor" strokeWidth="2"/>
                  <polyline points="22,6 12,13 2,6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
                partnerships@functionfly.com
              </a>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default PartnershipsPage
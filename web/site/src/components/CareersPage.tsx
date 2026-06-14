import React from 'react'
import './homepage.css'

const CareersPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-careers-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Join our team</span>
          </div>
          <h1 className="ff-hero-headline">
            We're building the trust layer<br />for AI agents
          </h1>
          <p className="ff-hero-sub">
            FunctionFly is a small, focused team working on verified functions, trust scores, and zero-knowledge cryptography. If you want to shape how agents choose tools safely, we'd love to hear from you.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href="mailto:careers@functionfly.com">Send introduction</a>
            <a className="ff-btn ff-btn-secondary" href="/">Learn more</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pillars">
            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                  <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Engineering</h2>
              <p>
                Build the core platform: function execution, trust scoring, MCP protocol, and the zero-knowledge vault.
              </p>
            </div>

            <div className="ff-pillar-card ff-pillar-card--accent">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Product & Design</h2>
              <p>
                Shape the developer experience and dashboard for publishing, monitoring, and trust management.
              </p>
            </div>

            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Security</h2>
              <p>
                Define verification standards, audit the platform, and ensure the trust layer is trustworthy by design.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">How to apply</h2>
          <p className="ff-section-lead">
            We're not hiring right now, but we welcome interest from exceptional people.
          </p>
          <div className="ff-careers-steps">
            <div className="ff-careers-step">
              <div className="ff-careers-step-num">1</div>
              <h3>Send your introduction</h3>
              <p>Email <a href="mailto:careers@functionfly.com">careers@functionfly.com</a> with a brief intro and links to your work.</p>
            </div>
            <div className="ff-careers-step">
              <div className="ff-careers-step-num">2</div>
              <h3>What to include</h3>
              <ul>
                <li>Your role interest (engineering, product, design, security, or operations)</li>
                <li>2-3 links to work you're proud of (GitHub, portfolio, talks, papers)</li>
                <li>Your location/time zone and remote preferences</li>
              </ul>
            </div>
            <div className="ff-careers-step">
              <div className="ff-careers-step-num">3</div>
              <h3>We read every message</h3>
              <p>Response time varies depending on our hiring cycle. We'll reach out when roles open.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Ready to make agents safer?</h2>
          <p>We'd love to hear from you.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href="mailto:careers@functionfly.com">Send introduction</a>
            <a className="ff-btn ff-btn-secondary" href="/">Learn about FunctionFly</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default CareersPage

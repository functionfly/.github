import React from 'react'
import { AUTH_ORIGIN, APP_DASHBOARD_ORIGIN } from '../config'
import './homepage.css'

const PlansPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Simple, transparent pricing</span>
          </div>
          <h1 className="ff-hero-headline">
            Choose the plan<br />that scales with you
          </h1>
          <p className="ff-hero-sub">
            Start free, pay as you grow. No hidden fees, no surprise bills. Every plan includes trust infrastructure, function registry, and sandboxed execution.
          </p>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-plans-grid">
            <div className="ff-plan-card">
              <div className="ff-plan-header">
                <h3 className="ff-plan-name">Free</h3>
                <div className="ff-plan-price">
                  <span className="ff-plan-amount">$0</span>
                  <span className="ff-plan-period">/month</span>
                </div>
                <p className="ff-plan-desc">Perfect for exploring FunctionFly and testing agent integrations.</p>
              </div>
              <ul className="ff-plan-features">
                <li>25,000 requests/month</li>
                <li>3 published functions</li>
                <li>3 AI agents</li>
                <li>10K AI calls/month</li>
                <li>Basic trust scores</li>
                <li>L1 verification only</li>
                <li>Community support</li>
                <li>7-day execution logs</li>
                <li>24h Time Machine</li>
                <li>Zero-knowledge Vault (25 secrets)</li>
              </ul>
              <a className="ff-btn ff-btn-secondary ff-btn-full" href={`${AUTH_ORIGIN}/signup`}>
                Get started free
              </a>
            </div>

            <div className="ff-plan-card ff-plan-card--featured">
              <div className="ff-plan-badge">Most Popular</div>
              <div className="ff-plan-header">
                <h3 className="ff-plan-name">Professional</h3>
                <div className="ff-plan-price">
                  <span className="ff-plan-amount">$79</span>
                  <span className="ff-plan-period">/month</span>
                </div>
                <p className="ff-plan-desc">For growing teams building production agent applications.</p>
              </div>
              <ul className="ff-plan-features">
                <li>2.5M requests/month</li>
                <li>25 published functions</li>
                <li>100 AI agents included</li>
                <li>1M AI calls/month</li>
                <li>Advanced trust scores</li>
                <li>L1-L3 verification</li>
                <li>Email support</li>
                <li>90-day execution logs</li>
                <li>Zero-knowledge Vault (500 secrets)</li>
                <li>30-day Time Machine</li>
              </ul>
              <a className="ff-btn ff-btn-primary ff-btn-full" href={`${AUTH_ORIGIN}/signup?plan=professional`}>
                Start free trial
              </a>
            </div>

            <div className="ff-plan-card">
              <div className="ff-plan-header">
                <h3 className="ff-plan-name">Enterprise</h3>
                <div className="ff-plan-price">
                  <span className="ff-plan-amount">$299</span>
                  <span className="ff-plan-period">/month</span>
                </div>
                <p className="ff-plan-desc">For teams requiring advanced trust and collaboration features.</p>
              </div>
              <ul className="ff-plan-features">
                <li>25M requests/month</li>
                <li>Unlimited published functions</li>
                <li>500 AI agents included</li>
                <li>5M AI calls/month</li>
                <li>Full trust scores + attestations</li>
                <li>L1-L4 verification</li>
                <li>Priority support + SLA</li>
                <li>1-year execution logs</li>
                <li>Zero-knowledge Vault (5K secrets)</li>
                <li>90-day Time Machine + reconciliation</li>
                <li>RBAC + Secret sharing</li>
              </ul>
              <a className="ff-btn ff-btn-secondary ff-btn-full" href={`${AUTH_ORIGIN}/signup?plan=enterprise`}>
                Start free trial
              </a>
            </div>

            <div className="ff-plan-card ff-plan-card--enterprise">
              <div className="ff-plan-header">
                <h3 className="ff-plan-name">Agent Enterprise</h3>
                <div className="ff-plan-price">
                  <span className="ff-plan-amount">$499</span>
                  <span className="ff-plan-period">/month</span>
                </div>
                <p className="ff-plan-desc">Unlimited AI scale for organizations with custom security, compliance, and scaling needs.</p>
              </div>
              <ul className="ff-plan-features">
                <li>Unlimited requests</li>
                <li>Unlimited published functions</li>
                <li>Unlimited AI agents</li>
                <li>Unlimited AI calls</li>
                <li>Full trust infrastructure</li>
                <li>All verification levels</li>
                <li>Dedicated support + SLA</li>
                <li>7-year data retention</li>
                <li>SSO / SAML</li>
                <li>Custom integrations</li>
                <li>Audit logs + compliance reports</li>
                <li>Incident insurance</li>
                <li>Zero-knowledge Vault (1M secrets)</li>
              </ul>
              <a className="ff-btn ff-btn-outline ff-btn-full" href="/contact">
                Contact sales
              </a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Add-ons</span>
            </div>
            <h2 className="ff-section-title">Extend your plan</h2>
            <p className="ff-section-desc">
              Enhance your FunctionFly experience with these optional add-ons available on any paid plan.
            </p>
          </div>

          <div className="ff-addons-grid">
            <div className="ff-addon-card">
              <div className="ff-addon-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
              </div>
              <h3>Zero-Knowledge Vault</h3>
              <p>Client-side encrypted secrets storage. Server never sees plaintext. Free tier includes 25 secrets, Pro up to 500, Team up to 5K.</p>
            </div>

            <div className="ff-addon-card">
              <div className="ff-addon-icon ff-addon-icon--green">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>
              </div>
              <h3>Dynamic Credentials</h3>
              <p>Auto-rotating database credentials for PostgreSQL and MySQL. Pro includes PostgreSQL, Team adds MySQL support.</p>
            </div>

            <div className="ff-addon-card">
              <div className="ff-addon-icon ff-addon-icon--amber">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
              </div>
              <h3>Time Machine Replay</h3>
              <p>Replay past function executions with full state. Free: 24h, Starter: 72h, Professional: 30 days, Enterprise: 90 days + live reconciliation.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>FAQ</span>
            </div>
            <h2 className="ff-section-title">Common questions</h2>
          </div>

          <div className="ff-faq-list">
            <div className="ff-faq-item">
              <h3>How do function requests work?</h3>
              <p>Each function call counts as one request, regardless of execution time. Failed calls due to trust policy enforcement also count toward your limit.</p>
            </div>

            <div className="ff-faq-item">
              <h3>Can I upgrade or downgrade anytime?</h3>
              <p>Yes. You can change your plan at any time. Upgrades take effect immediately, and you'll be charged a prorated amount. Downgrades take effect at your next billing cycle.</p>
            </div>

            <div className="ff-faq-item">
              <h3>What happens if I exceed my request limit?</h3>
              <p>We'll notify you when you reach 80% of your limit. If you exceed it, functions will return a rate limit error until you upgrade or the cycle resets.</p>
            </div>

            <div className="ff-faq-item">
              <h3>Is there a free trial for paid plans?</h3>
              <p>Yes, Professional and Enterprise plans come with a 14-day free trial. No credit card required to start.</p>
            </div>

            <div className="ff-faq-item">
              <h3>What is the Zero-Knowledge Vault?</h3>
              <p>The Vault uses client-side encryption — your passphrase and secrets are never sent to our servers. All encryption happens in your browser using AES-256-GCM. Free includes 25 secrets, Professional up to 500, Team up to 5,000.</p>
            </div>

            <div className="ff-faq-item">
              <h3>What verification levels are included?</h3>
              <p>Free includes L1 (format checks). Professional adds L2-L3 (security scans and code review). Enterprise adds L4 (platform-verified) plus custom attestation workflows.</p>
            </div>

            <div className="ff-faq-item">
              <h3>Do prices include tax?</h3>
              <p>Listed prices are in USD and exclude applicable taxes. Your invoice will reflect tax rules for your billing address.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--cta">
        <div className="ff-container">
          <div className="ff-cta">
            <h2 className="ff-cta-title">
              Ready to build with<br />
              <span className="ff-cta-accent">trust infrastructure</span>?
            </h2>
            <p className="ff-cta-desc">
              Start free today. No credit card required.
            </p>
            <div className="ff-cta-actions">
              <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
                Create free account
              </a>
              <a className="ff-btn ff-btn-outline" href="/contact">
                Talk to sales
              </a>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default PlansPage
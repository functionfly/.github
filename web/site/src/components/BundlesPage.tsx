import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import './homepage.css'
import './pricing.css'

const BundlesPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-statefabric-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Backend-in-a-Box</span>
          </div>
          <h1 className="ff-hero-headline">
            Skip the backend.<br />Ship faster.
          </h1>
          <p className="ff-hero-sub">
            Stop wiring together auth, payments, databases, and email. Get pre-configured bundles with everything you need to launch — and scale when you're ready.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
              Start free — no credit card
            </a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">View pricing</a>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Why use bundles?</h2>
          <div className="ff-addon-grid">
            <div className="ff-addon-card">
              <h4>Hours saved</h4>
              <p>Stop building auth from scratch, integrating Stripe, or debugging email deliverability. Pre-configured and production-ready.</p>
            </div>
            <div className="ff-addon-card">
              <h4>No decision fatigue</h4>
              <p>We picked the best tools for each job. Auth0 for auth, Stripe for payments, SendGrid for email. Just connect your keys and go.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Scale automatically</h4>
              <p>Each bundle scales independently. Hit a limit on one component? Upgrade just that piece without touching the rest.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Works together</h4>
              <p>Auth users flow into your DB. Payments trigger webhooks. Email personalizes by user tier. Everything connected out of the box.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Free until you grow</h4>
              <p>All bundles start free. No credit card, no commitment. Pay nothing until you hit 100 users or $1K MRR.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Swap pieces out</h4>
              <p>Using Auth0 but want to switch to your own auth? No lock-in. Each piece is swappable without rebuilding the whole stack.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pricing-grid">
            <div className="ff-price-card">
              <h3>SaaS Starter</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$29</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For new SaaS products and MVPs.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Pre-configured auth (OAuth, magic links)</li>
                <li><span className="ff-check">✓</span> Stripe payments & subscription management</li>
                <li><span className="ff-check">✓</span> User DB with role-based access</li>
                <li><span className="ff-check">✓</span> Transactional email via SendGrid</li>
                <li><span className="ff-check">✓</span> Usage analytics dashboard</li>
                <li><span className="ff-check">✓</span> Webhook support</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing/bundles?plan=saas`}>Choose SaaS Starter</a>
            </div>

            <div className="ff-price-card ff-price-card--featured">
              <span className="ff-featured-badge">Most Popular</span>
              <h3>AI App</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$39</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For AI-powered applications.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> pgvector for semantic search</li>
                <li><span className="ff-check">✓</span> OpenAI & Anthropic embeddings pipeline</li>
                <li><span className="ff-check">✓</span> Chat history & conversation memory</li>
                <li><span className="ff-check">✓</span> RAG-ready document ingestion</li>
                <li><span className="ff-check">✓</span> Token usage analytics</li>
                <li><span className="ff-check">✓</span> Memory persistence across sessions</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing/bundles?plan=ai-app`}>Choose AI App</a>
            </div>

            <div className="ff-price-card">
              <h3>Marketplace</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$49</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For two-sided marketplaces.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Product listings & search</li>
                <li><span className="ff-check">✓</span> Stripe Connect for multi-vendor payments</li>
                <li><span className="ff-check">✓</span> Real-time messaging between buyers/sellers</li>
                <li><span className="ff-check">✓</span> Push & email notifications</li>
                <li><span className="ff-check">✓</span> Order management & dispute resolution</li>
                <li><span className="ff-check">✓</span> Vendor onboarding & verification</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing/bundles?plan=marketplace`}>Choose Marketplace</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">What's included in every bundle</h2>
          <div className="ff-addon-grid">
            <div className="ff-addon-card">
              <h4>Auth</h4>
              <p>OAuth, magic links, and password auth. MFA, SSO, and SAML available on higher tiers. Session management included.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Database</h4>
              <p>PostgreSQL with automatic backups. User tables, audit logs, and API access pre-wired. Connect your own DB if preferred.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Payments</h4>
              <p>Stripe integration for subscriptions and one-time payments. Automatic invoicing, tax handling, and dunning included.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Email</h4>
              <p>Transactional email via SendGrid. Pre-built templates for onboarding, billing, and notifications. Delivery analytics included.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Analytics</h4>
              <p>Usage dashboards for DAU/MAU, retention, and revenue. Funnel visualization and cohort analysis out of the box.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Webhooks</h4>
              <p>Event-driven architecture with retry logic. Connect any external service or build your own integrations.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Compare bundles</h2>
          <div className="ff-comparison-table">
            <div className="ff-comparison-header">
              <div>Feature</div>
              <div>SaaS Starter</div>
              <div>AI App</div>
              <div>Marketplace</div>
            </div>
            <div className="ff-comparison-row">
              <div>Auth</div>
              <div>OAuth, magic links, MFA</div>
              <div>OAuth, magic links, MFA</div>
              <div>OAuth, magic links, MFA</div>
            </div>
            <div className="ff-comparison-row">
              <div>Database</div>
              <div>User DB + RBAC</div>
              <div>User DB + RBAC</div>
              <div>User DB + RBAC</div>
            </div>
            <div className="ff-comparison-row">
              <div>Payments</div>
              <div>Stripe subscriptions</div>
              <div>Stripe subscriptions</div>
              <div>Stripe Connect (multi-vendor)</div>
            </div>
            <div className="ff-comparison-row">
              <div>Email</div>
              <div>Transactional + templates</div>
              <div>Transactional + templates</div>
              <div>Transactional + templates</div>
            </div>
            <div className="ff-comparison-row">
              <div>Analytics</div>
              <div>Usage + revenue</div>
              <div>Usage + revenue + AI tokens</div>
              <div>Usage + revenue + GMV</div>
            </div>
            <div className="ff-comparison-row">
              <div>Vector / embeddings</div>
              <div>—</div>
              <div>pgvector + OpenAI/Anthropic</div>
              <div>—</div>
            </div>
            <div className="ff-comparison-row">
              <div>Chat / memory</div>
              <div>—</div>
              <div>Conversation memory</div>
              <div>—</div>
            </div>
            <div className="ff-comparison-row">
              <div>Marketplace features</div>
              <div>—</div>
              <div>—</div>
              <div>Listings, messaging, disputes</div>
            </div>
            <div className="ff-comparison-row">
              <div>Webhook support</div>
              <div>Yes</div>
              <div>Yes</div>
              <div>Yes</div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Frequently asked questions</h2>
          <dl className="pricing-faq">
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What does "free until you grow" mean?</dt>
              <dd className="pricing-faq-answer">All bundles start free. You won't be charged until you hit 100 users or $1K MRR — whichever comes first. No credit card required to start.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Can I switch bundles?</dt>
              <dd className="pricing-faq-answer">Yes. If your product evolves from a simple SaaS to an AI-powered app, you can switch bundles. Your data migrates automatically.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What if I need a feature not in my bundle?</dt>
              <dd className="pricing-faq-answer">Each component is swappable. You can add features from other bundles or bring your own integrations. The bundle is a starting point, not a prison.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Can I use my own Stripe/Auth0/SendGrid account?</dt>
              <dd className="pricing-faq-answer">Yes. Bring your own keys for any service. The bundle configures everything — you just need to connect your account.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What happens when I hit the free limits?</dt>
              <dd className="pricing-faq-answer">We'll notify you when you're approaching 100 users or $1K MRR. At that point, you choose whether to upgrade or stay on the free tier with limited functionality.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Is there a setup fee?</dt>
              <dd className="pricing-faq-answer">No. The bundle price is all-inclusive. Connect your external accounts and start building.</dd>
            </div>
          </dl>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Stop building plumbing. Start shipping features.</h2>
          <p>Free to start. No credit card. Scale when you're ready.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Start free — no credit card</a>
            <a className="ff-btn ff-btn-secondary" href="/contact">Talk to sales</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default BundlesPage